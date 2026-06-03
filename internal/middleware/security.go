package middleware

import (
	"github.com/gin-gonic/gin"
)

func SecureHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")

		c.Header("X-Content-Type-Options", "nosniff")

		c.Header("X-XSS-Protection", "1; mode=block")

		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';")

		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()")

		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		c.Header("Server", "")

		c.Next()
	}
}
