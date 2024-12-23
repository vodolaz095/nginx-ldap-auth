package endpoints

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/ldap4gin"
)

func (api *API) injectSession() {
	api.engine.GET("/auth/session", func(c *gin.Context) {
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
		log.Debug().Msgf("Welcome, %s!", user.String())
		c.String(http.StatusOK, "Welcome, %s!", user.String())
	})
}
