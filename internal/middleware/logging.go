package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		clientIP := getClientIP(c)

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		level := slog.LevelInfo
		if statusCode >= 400 && statusCode < 500 {
			level = slog.LevelWarn
		} else if statusCode >= 500 {
			level = slog.LevelError
		}

		logger.Log(c.Request.Context(), level, "http request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("query", c.Request.URL.RawQuery),
			slog.Int("status", statusCode),
			slog.String("client_ip", clientIP),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.Duration("duration", duration),
			slog.Int("bytes_written", c.Writer.Size()),
		)
	}
}
