package main

import (
	"context"
	"crypto/tls"
	"flag"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/pkg/healthcheck"
	"github.com/vodolaz095/pkg/stopper"
	"github.com/vodolaz095/pkg/tracing"
	"github.com/vodolaz095/pkg/zerologger"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	"github.com/vodolaz095/ldap4gin"
	"github.com/vodolaz095/nginx-ldap-auth/config"
	"github.com/vodolaz095/nginx-ldap-auth/endpoints"
)

var Version = "development"
var Subversion = ""

func main() {
	// load config
	flag.Parse()
	if len(flag.Args()) != 1 {
		log.Fatal().Msgf("please, provide path to config as 1st argument")
	}
	pathToConfig := flag.Args()[0]
	cfg, err := config.LoadFromFile(pathToConfig)
	if err != nil {
		log.Fatal().Err(err).
			Msgf("error loading config from %s: %s", pathToConfig, err)
	}
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(cfg)
	if err != nil {
		log.Fatal().Err(err).
			Msgf("error validating configuration file %s: %s", pathToConfig, err)
	}

	// set logging
	zerologger.Configure(cfg.Log)

	// set main application context
	mainCtx, cancel := stopper.New()
	defer cancel()

	// set tracing
	err = tracing.ConfigureUDP(cfg.Tracing,
		semconv.ServiceName("nginx-ldap-auth"),
		semconv.ServiceVersion(Version),
		semconv.DeploymentEnvironment(cfg.Realm),
	)
	if err != nil {
		log.Fatal().Err(err).Msgf("error configuring tracing: %s", err)
	}

	// set release mode
	if Version != "development" {
		gin.SetMode(gin.ReleaseMode)
	}
	// set openldap authenticator
	authenticator, err := ldap4gin.New(&ldap4gin.Options{
		Debug:            gin.IsDebugging(),
		ConnectionString: cfg.Authenticator.ConnectionString,
		ReadonlyDN:       cfg.Authenticator.ReadonlyDN,
		ReadonlyPasswd:   cfg.Authenticator.ReadonlyPasswd,
		TLS: &tls.Config{
			InsecureSkipVerify: cfg.Authenticator.InsecureTLS,
		},
		StartTLS:      cfg.Authenticator.StartTLS,
		UserBaseTpl:   cfg.Authenticator.UserBaseTpl,
		ExtractGroups: true,
		GroupsOU:      cfg.Authenticator.GroupsOU,
		TTL:           cfg.Authenticator.TTL,
		LogDebugFunc: func(ctx context.Context, format string, data ...any) {
			span := trace.SpanFromContext(ctx)
			logger := log.Debug().CallerSkipFrame(2)
			if span.SpanContext().HasTraceID() {
				logger = logger.Str("trace_id", span.SpanContext().TraceID().String())
			}
			logger.Msgf(format, data...)
		},
	})
	if err != nil {
		log.Fatal().Err(err).Msgf("error configuring authenticator: %s", err)
	}
	log.Info().Msgf("OpenLDAP on %s is ready!", cfg.Authenticator.ConnectionString)
	err = authenticator.Ping(mainCtx)
	if err != nil {
		log.Fatal().Err(err).Msgf("error pinging authenticator: %s", err)
	}
	log.Debug().Msgf("Starting profile under %s prefix...", cfg.WebServer.ProfilePrefix)
	api := endpoints.API{
		Authenticator:                         authenticator,
		Realm:                                 cfg.Realm,
		SubrequestPathForBasicAuthorization:   cfg.WebServer.SubrequestPathForBasicAuthorization,
		SubrequestPathForSessionAuthorization: cfg.WebServer.SubrequestPathForSessionAuthorization,
		ProfilePrefix:                         cfg.WebServer.ProfilePrefix,
		Permissions:                           cfg.Permission,
		Version:                               Version + Subversion,
	}
	hc_supported, err := healthcheck.Ready()
	if err != nil {
		log.Fatal().Err(err).Msgf("error connecting to watchdog: %s", err)
	}

	// start main loop
	eg, ctx := errgroup.WithContext(mainCtx)
	eg.Go(func() error {
		return tracing.Wait(ctx)
	})
	eg.Go(func() error {
		return authenticator.Close()
	})
	eg.Go(func() error {
		return api.StartAuthAPI(ctx, cfg.WebServer)
	})
	eg.Go(func() (err error) {
		if !hc_supported {
			return nil
		}
		err = healthcheck.SetStatus("listening for requests...")
		if err != nil {
			return err
		}
		return healthcheck.StartWatchDog(ctx, []healthcheck.Pinger{
			authenticator,
		})
	})
	err = eg.Wait()
	if err != nil {
		log.Error().Err(err).Msgf("Error stopping application: %s", err)
	} else {
		log.Info().Msg("Application is stopped")
	}
}
