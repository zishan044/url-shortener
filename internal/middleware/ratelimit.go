package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client      *redis.Client
	maxRequests int
	window      time.Duration
}

func NewRateLimiter(client *redis.Client, maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		client:      client,
		maxRequests: maxRequests,
		window:      window,
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := getClientIP(c)
		key := fmt.Sprintf("rate_limit:%s", clientIP)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		count, err := rl.client.Get(ctx, key).Int64()
		if err != nil && err != redis.Nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limiter error"})
			c.Abort()
			return
		}

		if count >= int64(rl.maxRequests) {
			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.maxRequests))
			c.Header("X-RateLimit-Remaining", "0")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"retry_after": rl.window.Seconds(),
			})
			c.Abort()
			return
		}

		pipe := rl.client.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, rl.window)
		_, err = pipe.Exec(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limiter error"})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.maxRequests))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(int64(rl.maxRequests)-count-1, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(rl.window).Unix(), 10))

		c.Next()
	}
}

func getClientIP(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Try X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}
