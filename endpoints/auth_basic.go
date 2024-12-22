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
	api.engine.GET("/auth/basic", func(c *gin.Context) {
		user, err := api.Authenticator.Extract(c)
		if err != nil {
			if errors.Is(err, ldap4gin.ErrUnauthorized) {
				username, password, ok := c.Request.BasicAuth()
				if !ok {
					c.Header("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%q\", charset=\"UTF-8\"", api.Realm))
					c.AbortWithStatus(http.StatusUnauthorized)
					return
				}
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
				c.String(http.StatusOK, "Welcome, %s!", username)
				return
			}
			log.Error().Err(err).Msgf("Authenticator error: %S", err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		log.Debug().Msgf("Welcome, %s!", user.String())
		c.String(http.StatusOK, "Welcome, %s!", user.String())
	})
}
