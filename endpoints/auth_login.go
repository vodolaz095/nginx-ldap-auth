package endpoints

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/ldap4gin"
	"github.com/vodolaz095/nginx-ldap-auth/middlewares"
)

// loginForm is used for authorization via username+password
type loginForm struct {
	CSRF     string `form:"_csrf" binding:"required"`
	Username string `form:"username"  binding:"required"`
	Password string `form:"password"  binding:"required"`
}

func (api *API) injectLoginForm() {
	api.engine.GET("/auth/login", func(c *gin.Context) {
		csrf, ok := c.Get("csrf")
		if !ok {
			csrf = "error!"
		}
		user, err := api.Authenticator.Extract(c)
		if err != nil {
			if errors.Is(err, ldap4gin.ErrUnauthorized) {
				c.HTML(http.StatusUnauthorized, "login.html", gin.H{
					"title": "Authorization is required",
					"csrf":  csrf.(string),
				})
				return
			}
			log.Error().Err(err).Msgf("Error checking session for /auth/login: %s", err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.HTML(http.StatusOK, "profile.html", gin.H{
			"title": fmt.Sprintf("Welcome, %s!", user.String()),
			"user":  user,
		})
	})

	api.engine.GET("/auth/logout", func(c *gin.Context) {
		err := api.Authenticator.Logout(c)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Redirect(http.StatusFound, "/auth/login")
	})

	api.engine.GET("/auth/whoami", func(c *gin.Context) {
		user, err := api.Authenticator.Extract(c)
		if err != nil {
			if errors.Is(err, ldap4gin.ErrUnauthorized) {
				c.Redirect(http.StatusFound, "/auth/login")
				return
			}
			log.Error().Err(err).Msgf("Error rendering whoami: %s", err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.HTML(http.StatusOK, "profile.html", gin.H{
			"title": fmt.Sprintf("Welcome, %s!", user.String()),
			"user":  user,
		})
	})

	api.engine.Use(middlewares.CheckCSRF)
	api.engine.POST("/auth/login", func(c *gin.Context) {
		var bdy loginForm
		err := c.Bind(&bdy)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		err = api.Authenticator.Authorize(c, bdy.Username, bdy.Password)
		if err != nil {
			log.Error().Err(err).Msgf("Error rendering whoami: %s", err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Redirect(http.StatusFound, "/auth/login")
	})
}
