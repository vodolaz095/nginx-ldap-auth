package endpoints

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/ldap4gin"
	"github.com/vodolaz095/nginx-ldap-auth/middlewares"
	"github.com/vodolaz095/nginx-ldap-auth/public"
)

// loginForm is used for authorization via username+password
type loginForm struct {
	CSRF     string `form:"_csrf" binding:"required"`
	Username string `form:"username"  binding:"required"`
	Password string `form:"password"  binding:"required"`
}

func (api *API) injectLoginForm() {
	// load static files
	fs := http.FS(public.Assets)
	api.engine.StaticFS(api.ProfilePrefix+"/assets/", fs)
	api.engine.GET(api.ProfilePrefix+"/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("favicon.ico", fs)
	})
	api.engine.GET(api.ProfilePrefix+"/", func(c *gin.Context) {
		session := sessions.Default(c)
		flashes := session.Flashes()
		defer session.Save()

		csrf, ok := c.Get("csrf")
		if !ok {
			csrf = "error!"
		}
		user, err := api.Authenticator.Extract(c)
		if err != nil {
			if errors.Is(err, ldap4gin.ErrUnauthorized) {
				c.HTML(http.StatusUnauthorized, "login.html", gin.H{
					"title":         "Authorization is required",
					"csrf":          csrf.(string),
					"realm":         api.Realm,
					"flashes":       flashes,
					"profilePrefix": template.HTMLAttr(api.ProfilePrefix),
				})
				return
			}
			log.Error().Err(err).Msgf("Error checking session for /auth/login: %s", err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.HTML(http.StatusOK, "profile.html", gin.H{
			"title":         fmt.Sprintf("Welcome, %s!", user.String()),
			"realm":         api.Realm,
			"placesAllowed": api.listAllowed(c.Request.Host, user),
			"user":          user,
			"flashes":       flashes,
			"profilePrefix": template.HTMLAttr(api.ProfilePrefix),
		})
	})

	api.engine.GET(api.ProfilePrefix+"/logout", func(c *gin.Context) {
		err := api.Authenticator.Logout(c)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Redirect(http.StatusFound, api.ProfilePrefix+"/")
	})

	api.engine.Use(middlewares.CheckCSRF)
	api.engine.POST(api.ProfilePrefix+"/login", func(c *gin.Context) {
		session := sessions.Default(c)
		defer session.Save()

		var bdy loginForm
		err := c.Bind(&bdy)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		err = api.Authenticator.Authorize(c, bdy.Username, bdy.Password)
		if err != nil {
			session.AddFlash(fmt.Sprintf("Authorization error: %s", err.Error()))
			log.Error().Err(err).Msgf("error authorizing %s: %s", bdy.Username, err)
		} else {
			session.AddFlash(fmt.Sprintf("Welcome, %s!", bdy.Username))
		}
		c.Redirect(http.StatusFound, api.ProfilePrefix+"/")
	})

	api.engine.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusFound, api.ProfilePrefix+"/")
	})
}
