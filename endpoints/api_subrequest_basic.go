package endpoints

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/ldap4gin"
	"github.com/vodolaz095/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

func (api *API) injectBasicAuth() {
	if api.SubrequestPathForBasicAuthorization == "" {
		log.Debug().Msgf("subrequest for basic authorization is disabled")
		return
	}
	log.Debug().Msgf("subrequest for basic authorization is enabled for %s", api.SubrequestPathForBasicAuthorization)

	api.engine.GET(api.SubrequestPathForBasicAuthorization, func(c *gin.Context) {
		var err error
		var user *ldap4gin.User
		var ok bool
		span := trace.SpanFromContext(c.Request.Context())
		span.SetName("subrequest_basic")
		tracing.AttachCodeLocationToSpan(span)
		span.SetAttributes(semconv.NetHostName(c.Request.Host))
		span.SetAttributes(semconv.ServerAddress(c.Request.Host))
		origin := c.GetHeader("X-Original-URI")
		if origin != "" {
			span.SetAttributes(attribute.String("original_uri", origin))
		} else {
			span.AddEvent("header X-Original-URI is missing")
			c.String(http.StatusBadRequest, "header X-Original-URI is missing")
			return
		}
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\", charset=\"UTF-8\"", api.Realm))
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		logger := log.With().
			Str("hostname", c.Request.Host).
			IPAddr("client_ip", net.ParseIP(c.ClientIP())).
			Str("original_uri", origin).
			Str("username", username).
			Str("trace_id", span.SpanContext().TraceID().String()).
			Logger()

		key := GetMD5Hash(c.Request.Host, username, password)
		user, ok = api.authCache.Get(key)
		if !ok {
			span.AddEvent("cache miss")
			span.SetAttributes(attribute.Bool("cache_hit", false))

			err = api.Authenticator.Authorize(c, username, password)
			if err != nil {
				if errors.Is(err, ldap4gin.ErrInvalidCredentials) {
					c.Header("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\", charset=\"UTF-8\"", api.Realm))
					c.AbortWithStatus(http.StatusUnauthorized)
					return
				}
				logger.Err(err).
					Bool("cache_hit", ok).
					Msgf("Error checking username and password from basic challenge: %s", err)
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			user, err = api.Authenticator.Extract(c)
			if err != nil {
				logger.Err(err).
					Bool("cache_hit", ok).
					Msgf("Extracting user from metadata: %s", err)
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			api.authCache.Add(key, user)
		} else {
			span.AddEvent("cache hit")
			span.SetAttributes(attribute.Bool("cache_hit", true))
		}
		err = api.checkPermissions(c.Request.Context(), c.Request.Host, origin, user)
		if err != nil {
			if errors.Is(err, errAccessDenied) {
				c.String(http.StatusForbidden, "Forbidden: %s", err)
				logger.Debug().
					Bool("cache_hit", ok).
					Msgf("user %s cannot access %s%s", user.UID, c.Request.Host, origin)

				return
			}
			logger.Error().Err(err).
				Bool("cache_hit", ok).
				Msgf("checking permissions: %s", err)

			return
		}
		logger.Debug().
			Bool("cache_hit", ok).
			Msgf("User %s is allowed to %s on hostname %s",
				user.String(), origin, c.Request.Host)

		c.String(http.StatusOK, "Welcome, %s!", username)
		return
	})
}
