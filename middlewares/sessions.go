package middlewares

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/vodolaz095/nginx-ldap-auth/config"
)

// UseCookieSession uses secure cookies
func UseCookieSession(router *gin.Engine, cfg config.WebServer) {
	sessionStore := cookie.NewStore([]byte(cfg.SessionSecret))
	sessionStore.Options(sessions.Options{
		Path:     "/",
		MaxAge:   int(cfg.SessionMaxAgeInSeconds.Seconds()),
		HttpOnly: true,
		Secure:   !gin.IsDebugging(),
		SameSite: http.SameSiteStrictMode,
	})
	router.Use(sessions.Sessions("PHPSESSID", sessionStore))
}
