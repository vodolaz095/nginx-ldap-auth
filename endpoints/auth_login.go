package endpoints

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/ldap4gin"
)

// loginForm is used for authorization via username+password
type loginForm struct {
	Username string `json:"username" form:"username"  binding:"required"`
	Password string `json:"password" form:"password"  binding:"required"`
}

func (api *API) injectLoginForm() {
	api.engine.GET("/auth/login", func(c *gin.Context) {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"title": "Authorization is required",
		})
	})

	api.engine.POST("/auth/login", func(c *gin.Context) {
		c.String(http.StatusOK, "login form")
	})

	api.engine.GET("/auth/logout", func(c *gin.Context) {
		c.Header("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%q\", charset=\"UTF-8\"", api.Realm))
		err := api.Authenticator.Logout(c)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.AbortWithStatus(http.StatusUnauthorized)
	})

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
