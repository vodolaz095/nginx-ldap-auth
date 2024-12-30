package endpoints

import (
	"errors"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/ldap4gin"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

func (api *API) injectSessionSubrequest() {
	if api.SubrequestPathForSessionAuthorization == "" {
		log.Debug().Msgf("session subrequest authorization is disabled")
		return
	}
	log.Debug().Msgf("session subrequest authorization is enabled for %s", api.SubrequestPathForBasicAuthorization)
	api.engine.GET(api.SubrequestPathForSessionAuthorization, func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		span.SetName("subrequest_session")
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
		logger := log.With().
			Str("hostname", c.Request.Host).
			IPAddr("client_ip", net.ParseIP(c.ClientIP())).
			Str("original_uri", origin).
			Str("trace_id", span.SpanContext().TraceID().String()).
			Logger()

		user, err := api.Authenticator.Extract(c)
		if err != nil {
			if errors.Is(err, ldap4gin.ErrUnauthorized) {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			logger.Error().Err(err).
				Msgf("Authenticator error: %S", err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		err = api.checkPermissions(c.Request.Context(), c.Request.Host, origin, user)
		if err != nil {
			if errors.Is(err, errAccessDenied) {
				c.String(http.StatusForbidden, "Forbidden: %s", err)
				return
			}
			logger.Error().Err(err).
				Str("username", user.UID).
				Msgf("checking permissions: %s", err)
			return
		}
		logger.Debug().
			Str("username", user.UID).
			Msgf("User %s is allowed to %s on hostname %s",
				user.String(), origin, c.Request.Host)
		c.String(http.StatusOK, "Welcome, %s!", user.String())
	})
}
