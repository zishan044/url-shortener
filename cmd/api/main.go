package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zishan044/url-shortener/internal/analytics"
	"github.com/zishan044/url-shortener/internal/auth"
	"github.com/zishan044/url-shortener/internal/cache"
	"github.com/zishan044/url-shortener/internal/config"
	"github.com/zishan044/url-shortener/internal/database"
	"github.com/zishan044/url-shortener/internal/middleware"
	"github.com/zishan044/url-shortener/internal/queue"
	"github.com/zishan044/url-shortener/internal/url"
)

func main() {
	cfg := config.LoadConfig()


	var logLevel slog.Level
	if cfg.Environment == "production" {
		logLevel = slog.LevelInfo
	} else {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	slog.SetDefault(logger)


	database.Connect(cfg.DatabaseURL)
	defer database.Pool.Close()

	cache.Connect(cfg.RedisURL)
	defer cache.Client.Close()

	publisher, err := queue.NewPublisher(cfg.RabbitMQURL)
	if err != nil {
		logger.Error("failed to connect to RabbitMQ", slog.Any("error", err))
		os.Exit(1)
	}
	defer publisher.Close()


	analyticsRepo := analytics.NewRepository(database.Pool)
	analyticsService := analytics.NewService(analyticsRepo)
	analyticsHandler := analytics.NewHandler(analyticsService)
	aggregationJob := analytics.NewAggregationJob(analyticsRepo, 5*time.Minute)

	worker, err := analytics.NewWorker(cfg.RabbitMQURL, analyticsRepo)
	if err != nil {
		logger.Error("failed to create analytics worker", slog.Any("error", err))
		os.Exit(1)
	}

	workerCtx, workerCancel := context.WithCancel(context.Background())
	go func() {
		if err := worker.Start(workerCtx); err != nil {
			logger.Error("worker error", slog.Any("error", err))
		}
	}()
	go aggregationJob.Start(workerCtx)


	authRepo := auth.NewRepository(database.Pool)
	authService := auth.NewService(authRepo, cfg.JWTSecret, cfg.JWTExpiry)
	authHandler := auth.NewHandler(authService)


	urlRepo := url.NewRepository(database.Pool)
	urlService := url.NewService(urlRepo)
	urlHandler := url.NewHandler(urlService, publisher)


	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()


	r.Use(middleware.RecoveryMiddleware(logger))
	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(middleware.SecureHeadersMiddleware())
	r.Use(middleware.CORSMiddleware(parseOrigins(cfg.AllowedOrigins)))
	r.Use(middleware.NewRateLimiter(cache.Client, cfg.RateLimitRequests, cfg.RateLimitWindow).Middleware())


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

		c.JSON(http.StatusOK, gin.H{
			"postgres": pgStatus,
			"redis":    rdbStatus,
			"time":     time.Now().Unix(),
		})
	})

	// API routes with timeout
	v1 := r.Group("/api/v1")
	v1.Use(timeoutMiddleware(cfg.RequestTimeout))

	auth.RegisterRoutes(v1, authHandler)
	analytics.RegisterRoutes(v1, analyticsHandler)
	url.RegisterRoutes(v1, urlHandler, cfg.JWTSecret)


	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  cfg.RequestTimeout,
		WriteTimeout: cfg.RequestTimeout,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		logger.Info("shutting down gracefully...")
		workerCancel()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("server shutdown error", slog.Any("error", err))
		}
	}()

	logger.Info("starting server", slog.String("port", cfg.Port), slog.String("environment", cfg.Environment))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server error", slog.Any("error", err))
		os.Exit(1)
	}
}

func parseOrigins(origins []string) []string {
	if len(origins) == 0 {
		return []string{"*"}
	}

	var parsed []string
	for _, origin := range origins {
		for _, o := range strings.Split(origin, ",") {
			parsed = append(parsed, strings.TrimSpace(o))
		}
	}
	return parsed
}

func timeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
