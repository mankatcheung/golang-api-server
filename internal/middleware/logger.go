package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-api-server/internal/handler"
)

// RequestLogger returns a Gin middleware that logs each HTTP request as a
// structured slog record after the handler chain completes.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		reqID, _ := c.Get("request_id")
		slog.Info("http request",
			"method",     c.Request.Method,
			"path",       c.FullPath(),
			"status",     c.Writer.Status(),
			"latency",    time.Since(start).String(),
			"ip",         c.ClientIP(),
			"user_id",    handler.GetUserID(c),
			"request_id", reqID,
		)
	}
}
