package endpoints

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/ldap4gin"
)

func (api *API) injectBasicAuth() {
	if api.SubrequestPathForBasicAuthorization == "" {
		log.Debug().Msgf("subrequest for basic authorization is disabled")
		return
	}
	log.Debug().Msgf("subrequest for basic authorization is enabled for %s", api.SubrequestPathForBasicAuthorization)

	api.engine.GET(api.SubrequestPathForBasicAuthorization, func(c *gin.Context) {
		var err error
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

		err = api.checkPermissions(c.Request.Context(), c.Request.Host, c.Request.RequestURI, user)
		if err != nil {
			if errors.Is(err, errAccessDenied) {
				c.String(http.StatusForbidden, "Forbidden: %s", err)
				return
			}
			log.Error().Err(err).Msgf("checking permissions: %s", err)
			return
		}

		c.String(http.StatusOK, "Welcome, %s!", username)
		return
	})
}
