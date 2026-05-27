package main

import (
	"github.com/gin-gonic/gin"

	"github.com/zishan044/url-shortener/internal/auth"
	"github.com/zishan044/url-shortener/internal/cache"
	"github.com/zishan044/url-shortener/internal/config"
	"github.com/zishan044/url-shortener/internal/database"
)

func main() {

	cfg := config.LoadConfig()

	database.Connect(cfg.DatabaseURL)
	defer database.Pool.Close()

	cache.Connect(cfg.RedisURL)
	defer cache.Client.Close()

	authRepo := auth.NewRepository(database.Pool)
	authService := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authService)

	r := gin.Default()
	v1 := r.Group("/api/v1")

	auth.RegisterRoutes(v1, authHandler)

	r.GET("/health", func(c *gin.Context) {

		pgErr := database.Pool.Ping(c.Request.Context())
		rdbErr := cache.Client.Ping(c.Request.Context()).Err()

		var pgStatus, rdbStatus string

		pgStatus = "OK"
		if pgErr != nil {
			pgStatus = "DOWN"
		}

		rdbStatus = "OK"
		if rdbErr != nil {
			rdbStatus = "DOWN"
		}

		c.JSON(200, gin.H{
			"postgres": pgStatus,
			"redis": rdbStatus,
		})
	})

	r.Run()
}