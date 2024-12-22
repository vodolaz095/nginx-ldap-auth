package endpoints

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (api *API) injectLogout() {
	api.engine.GET("/auth/logout", func(c *gin.Context) {
		c.Header("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%q\", charset=\"UTF-8\"", api.Realm))
		err := api.Authenticator.Logout(c)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.AbortWithStatus(http.StatusUnauthorized)
	})
}
