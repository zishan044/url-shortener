package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				clientIP := getClientIP(c)
				
				logger.Error("panic recovered",
					slog.String("error", fmt.Sprintf("%v", err)),
					slog.String("client_ip", clientIP),
					slog.String("path", c.Request.URL.Path),
					slog.String("method", c.Request.Method),
					slog.String("stack", string(debug.Stack())),
				)

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}
