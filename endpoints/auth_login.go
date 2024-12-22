package endpoints

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// loginForm is used for authorization via username+password
type loginForm struct {
	Username string `json:"username" form:"username"  binding:"required"`
	Password string `json:"password" form:"password"  binding:"required"`
}

func (api *API) injecSessionAuth() {
	api.engine.GET("/auth/login", func(c *gin.Context) {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"title": "Authorization is required",
		})
	})
	api.engine.POST("/auth/login", func(c *gin.Context) {
		c.String(http.StatusOK, "login form")
	})
}
