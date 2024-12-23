package middlewares

import (
	"crypto/rand"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func generateRandomString(n int) (string, error) {
	bytes, err := generateRandomBytes(n)
	if err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}

// UseCSRF is middleware for preventing cross site request forgery
func UseCSRF(router *gin.Engine) {
	router.Use(func(c *gin.Context) {
		session := sessions.Default(c)
		csrf := session.Get("csrf")
		if csrf != nil {
			c.Set("csrf", csrf)
			c.Next()
			return
		}
		newCsrf, err := generateRandomString(10)
		if err != nil {
			panic(err)
		}
		c.Set("csrf", newCsrf)
		session.Set("csrf", newCsrf)
		err = session.Save()
		if err != nil {
			panic(err)
		}
		c.Next()
	})
}

// CheckCSRF checks csrf token
func CheckCSRF(c *gin.Context) {
	span := trace.SpanFromContext(c.Request.Context())
	switch c.Request.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		session := sessions.Default(c)
		span.AddEvent("Validating CSRF token from session...")
		actual := c.PostForm("_csrf")
		expected := session.Get("csrf").(string)

		if actual != expected {
			span.AddEvent("CSRF token mismatch")
			c.String(http.StatusBadRequest, "csrf mismatch")
			c.Abort()
			return
		}
		span.AddEvent("CSRF token validated")
		c.Next()
	default:
		c.Next()
		return
	}
}
