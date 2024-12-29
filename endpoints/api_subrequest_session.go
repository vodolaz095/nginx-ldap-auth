package endpoints

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/ldap4gin"
)

func (api *API) injectSessionSubrequest() {
	if api.SubrequestPathForSessionAuthorization == "" {
		log.Debug().Msgf("session subrequest authorization is disabled")
		return
	}
	log.Debug().Msgf("session subrequest authorization is enabled for %s", api.SubrequestPathForBasicAuthorization)
	api.engine.GET(api.SubrequestPathForSessionAuthorization, func(c *gin.Context) {
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
		err = api.checkPermissions(c.Request.Context(), c.Request.Host, c.Request.RequestURI, user)
		if err != nil {
			if errors.Is(err, errAccessDenied) {
				c.String(http.StatusForbidden, "Forbidden: %s", err)
				return
			}
			log.Error().Err(err).Msgf("checking permissions: %s", err)
			return
		}

		log.Debug().Msgf("Welcome, %s!", user.String())
		c.String(http.StatusOK, "Welcome, %s!", user.String())
	})
}
