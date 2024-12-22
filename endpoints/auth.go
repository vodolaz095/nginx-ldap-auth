package endpoints

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/ldap4gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/vodolaz095/nginx-ldap-auth/config"
	"github.com/vodolaz095/nginx-ldap-auth/middlewares"
	"github.com/vodolaz095/nginx-ldap-auth/public"
)

type API struct {
	Authenticator *ldap4gin.Authenticator
	Realm         string
	engine        *gin.Engine
}

func (api *API) StartAuthAPI(ctx context.Context, cfg config.WebServer) (err error) {
	api.engine = gin.New()
	api.engine.Use(gin.Recovery())

	err = injectTemplates(api.engine)
	if err != nil {
		return fmt.Errorf("error injecting templates: %w", err)
	}

	api.engine.Use(otelgin.Middleware("nginx-ldap-auth-api",
		otelgin.WithSpanNameFormatter(func(r *http.Request) string {
			return r.Method + " " + r.URL.Path
		})),
	)
	middlewares.UseCookieSession(api.engine, cfg)
	middlewares.EmulatePHP(api.engine)
	middlewares.UseCSRF(api.engine)

	api.engine.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// todo TRACE-ID
		log.Debug().Msgf("[%s] - \"%s %s %s\" -> code=%d lat=%s size=%d / \"%s\"",
			param.ClientIP,
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.BodySize,
			param.Request.UserAgent(),
		)
		return ""
	}))

	// load static files
	fs := http.FS(public.Assets)
	api.engine.StaticFS("/assets/", fs)
	api.engine.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("favicon.ico", fs)
	})
	api.engine.GET("/robots.txt", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})

	// HTTP request handlers
	api.injectBasicAuth()
	api.injecSessionAuth()
	api.injectLogout()
	api.injectWhoAmI()

	// starting listener
	listener, err := net.Listen(cfg.Network, cfg.Listen)
	if err != nil {
		return fmt.Errorf("error starting listener on %s:%s - %w", cfg.Network, cfg.Listen, err)
	}
	go func() {
		<-ctx.Done()
		log.Debug().Msgf("Closing HTTP server on %s %s...", cfg.Network, cfg.Listen)
		listener.Close()
	}()
	err = api.engine.RunListener(listener)
	if strings.Contains(err.Error(), "use of closed network connection") {
		return nil
	}
	return nil
}
