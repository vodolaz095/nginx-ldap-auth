package endpoints

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/ldap4gin"
	"go.opentelemetry.io/otel/attribute"
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
		origin := c.GetHeader("X-Original-URI")
		if origin != "" {
			span.SetAttributes(attribute.String("original_uri", origin))
		} else {
			span.AddEvent("header X-Original-URI is missing")
			c.String(http.StatusBadRequest, "header X-Original-URI is missing")
			return
		}
		user, err := api.Authenticator.Extract(c)
		if err != nil {
			if errors.Is(err, ldap4gin.ErrUnauthorized) {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			log.Error().Err(err).Msgf("Authenticator error: %S", err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		err = api.checkPermissions(c.Request.Context(), c.Request.Host, origin, user)
		if err != nil {
			if errors.Is(err, errAccessDenied) {
				c.String(http.StatusForbidden, "Forbidden: %s", err)
				return
			}
			log.Error().Err(err).Msgf("checking permissions: %s", err)
			return
		}
		log.Debug().
			Str("trace_id", span.SpanContext().TraceID().String()).
			Msgf("User %s is allowed to %s on hostname %s",
				user.String(), origin, c.Request.Host)
		c.String(http.StatusOK, "Welcome, %s!", user.String())
	})
}
