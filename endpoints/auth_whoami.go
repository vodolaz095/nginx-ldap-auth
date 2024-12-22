package endpoints

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/ldap4gin"
)

func (api *API) injectWhoAmI() {
	api.engine.GET("/auth/whoami", func(c *gin.Context) {
		user, err := api.Authenticator.Extract(c)
		if err != nil {
			if errors.Is(err, ldap4gin.ErrUnauthorized) {
				c.String(http.StatusUnauthorized, "anomimus")
				return
			}
			log.Error().Err(err).Msgf("Error rendering whoami: %s", err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.String(http.StatusOK, "Welcome, %s!", user.String())
	})
}
