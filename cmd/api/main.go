package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zishan044/url-shortener/internal/analytics"
	"github.com/zishan044/url-shortener/internal/auth"
	"github.com/zishan044/url-shortener/internal/cache"
	"github.com/zishan044/url-shortener/internal/config"
	"github.com/zishan044/url-shortener/internal/database"
	"github.com/zishan044/url-shortener/internal/queue"
	"github.com/zishan044/url-shortener/internal/url"
)

func main() {

	cfg := config.LoadConfig()

	database.Connect(cfg.DatabaseURL)
	defer database.Pool.Close()

	cache.Connect(cfg.RedisURL)
	defer cache.Client.Close()

	publisher, err := queue.NewPublisher(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer publisher.Close()

	analyticsRepo := analytics.NewRepository(database.Pool)
	analyticsService := analytics.NewService(analyticsRepo)
	analyticsHandler := analytics.NewHandler(analyticsService)
	aggregationJob := analytics.NewAggregationJob(analyticsRepo, 5*time.Minute)

	worker, err := analytics.NewWorker(cfg.RabbitMQURL, analyticsRepo)
	if err != nil {
		log.Fatalf("Failed to create analytics worker: %v", err)
	}

	workerCtx, workerCancel := context.WithCancel(context.Background())
	go func() {
		if err := worker.Start(workerCtx); err != nil {
			log.Printf("Worker error: %v", err)
		}
	}()
	go aggregationJob.Start(workerCtx)

	authRepo := auth.NewRepository(database.Pool)
	authService := auth.NewService(authRepo, cfg.JWTSecret)
	authHandler := auth.NewHandler(authService)

	urlRepo := url.NewRepository(database.Pool)
	urlService := url.NewService(urlRepo)
	urlHandler := url.NewHandler(urlService, publisher)

	r := gin.Default()
	v1 := r.Group("/api/v1")

	auth.RegisterRoutes(v1, authHandler)
	analytics.RegisterRoutes(v1, analyticsHandler)
	url.RegisterRoutes(v1, urlHandler, cfg.JWTSecret)

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
			"redis":    rdbStatus,
		})
	})

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down gracefully...")
		workerCancel()
		os.Exit(0)
	}()

	r.Run()
}
