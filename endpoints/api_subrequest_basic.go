package endpoints

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/ldap4gin"
	"go.opentelemetry.io/otel/attribute"
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
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%q\", charset=\"UTF-8\"", api.Realm))
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// TODO - in memory cache!
		err = api.Authenticator.Authorize(c, username, password)
		if err != nil {
			if errors.Is(err, ldap4gin.ErrInvalidCredentials) {
				c.Header("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%q\", charset=\"UTF-8\"", api.Realm))
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			log.Error().Err(err).Msgf("Error checking username and password from basic challenge: %s", err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		user, err := api.Authenticator.Extract(c)
		if err != nil {
			log.Error().Err(err).Msgf("Extracting user from metadata: %s", err)
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
		c.String(http.StatusOK, "Welcome, %s!", username)
		return
	})
}
