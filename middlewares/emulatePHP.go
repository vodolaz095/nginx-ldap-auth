package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// EmulatePHP makes things more funny
func EmulatePHP(router *gin.Engine) {
	router.Use(func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		c.Header("X-Powered-By", "PHP/5.6.14") //https://schd.io/ENp
		c.Header("traceparent", span.SpanContext().TraceID().String())
		c.Header("tracestate", span.SpanContext().TraceState().String())
	})
}
