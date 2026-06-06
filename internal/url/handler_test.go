package url

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/zishan044/url-shortener/internal/cache"
	"github.com/zishan044/url-shortener/internal/models"
)

func setupTestCache(t *testing.T) {
	t.Helper()

	mr := miniredis.RunT(t)
	prevClient := cache.Client
	cache.Client = redis.NewClient(&redis.Options{Addr: mr.Addr()})

	t.Cleanup(func() {
		_ = cache.Client.Close()
		cache.Client = prevClient
		mr.Close()
	})
}

// MockURLService implements Service interface for testing
type MockURLService struct {
	CreateUrlFunc        func(ctx context.Context, url *models.Url) error
	GetUrlByShortCodeFunc func(ctx context.Context, shortCode string) (*models.Url, error)
}

func (m *MockURLService) CreateUrl(ctx context.Context, url *models.Url) error {
	return m.CreateUrlFunc(ctx, url)
}

func (m *MockURLService) GetUrlByShortCode(ctx context.Context, shortCode string) (*models.Url, error) {
	return m.GetUrlByShortCodeFunc(ctx, shortCode)
}

func TestCreateUrlSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockURLService{
		CreateUrlFunc: func(ctx context.Context, url *models.Url) error {
			url.ID = uuid.New()
			url.ShortCode = "abc123"
			url.CreatedAt = time.Now()
			return nil
		},
	}

	handler := NewHandler(mockService, nil)
	router := gin.New()

	// Add middleware to inject userID
	router.Use(func(c *gin.Context) {
		c.Set("userID", uuid.New())
		c.Next()
	})

	router.POST("/urls", handler.CreateUrl)

	reqBody := CreateUrlRequest{
		OriginalURL: "https://github.com/zishan044",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "URL created successfully" {
		t.Errorf("expected success message, got %v", response["message"])
	}
}

func TestCreateUrlInvalidURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockURLService{}
	handler := NewHandler(mockService, nil)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("userID", uuid.New())
		c.Next()
	})

	router.POST("/urls", handler.CreateUrl)

	reqBody := CreateUrlRequest{
		OriginalURL: "not-a-valid-url",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateUrlMissingSchema(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockURLService{}
	handler := NewHandler(mockService, nil)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("userID", uuid.New())
		c.Next()
	})

	router.POST("/urls", handler.CreateUrl)

	reqBody := CreateUrlRequest{
		OriginalURL: "github.com/zishan044",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateUrlUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockURLService{}
	handler := NewHandler(mockService, nil)
	router := gin.New()

	// No middleware to set userID - should fail
	router.POST("/urls", handler.CreateUrl)

	reqBody := CreateUrlRequest{
		OriginalURL: "https://github.com/zishan044",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestCreateUrlNoOriginalURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockURLService{}
	handler := NewHandler(mockService, nil)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("userID", uuid.New())
		c.Next()
	})

	router.POST("/urls", handler.CreateUrl)

	// Missing original_url field
	reqBody := map[string]interface{}{
		"expires_at": time.Now().Add(24 * time.Hour),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetUrlByShortCodeSuccess(t *testing.T) {
	setupTestCache(t)
	gin.SetMode(gin.TestMode)

	testURL := &models.Url{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		OriginalURL: "https://github.com/zishan044",
		ShortCode:   "abc123",
		CreatedAt:   time.Now(),
	}

	mockService := &MockURLService{
		GetUrlByShortCodeFunc: func(ctx context.Context, shortCode string) (*models.Url, error) {
			return testURL, nil
		},
	}

	handler := NewHandler(mockService, nil)
	router := gin.New()
	router.GET("/urls/:shortCode", handler.GetUrlByShortCode)

	req := httptest.NewRequest("GET", "/urls/abc123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response GetUrlResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.URL.ShortCode != "abc123" {
		t.Errorf("expected short code abc123, got %s", response.URL.ShortCode)
	}
}

func TestGetUrlByShortCodeNotFound(t *testing.T) {
	setupTestCache(t)
	gin.SetMode(gin.TestMode)

	mockService := &MockURLService{
		GetUrlByShortCodeFunc: func(ctx context.Context, shortCode string) (*models.Url, error) {
			return nil, errors.New("not found")
		},
	}

	handler := NewHandler(mockService, nil)
	router := gin.New()
	router.GET("/urls/:shortCode", handler.GetUrlByShortCode)

	req := httptest.NewRequest("GET", "/urls/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestRedirectSuccess(t *testing.T) {
	setupTestCache(t)
	gin.SetMode(gin.TestMode)

	testURL := &models.Url{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		OriginalURL: "https://github.com/zishan044",
		ShortCode:   "abc123",
		CreatedAt:   time.Now(),
	}

	mockService := &MockURLService{
		GetUrlByShortCodeFunc: func(ctx context.Context, shortCode string) (*models.Url, error) {
			return testURL, nil
		},
	}

	handler := NewHandler(mockService, nil)
	router := gin.New()
	router.GET("/r/:shortCode", handler.Redirect)

	req := httptest.NewRequest("GET", "/r/abc123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMovedPermanently {
		t.Errorf("expected status %d, got %d", http.StatusMovedPermanently, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "https://github.com/zishan044" {
		t.Errorf("expected location %s, got %s", "https://github.com/zishan044", location)
	}
}

func TestRedirectNotFound(t *testing.T) {
	setupTestCache(t)
	gin.SetMode(gin.TestMode)

	mockService := &MockURLService{
		GetUrlByShortCodeFunc: func(ctx context.Context, shortCode string) (*models.Url, error) {
			return nil, errors.New("not found")
		},
	}

	handler := NewHandler(mockService, nil)
	router := gin.New()
	router.GET("/r/:shortCode", handler.Redirect)

	req := httptest.NewRequest("GET", "/r/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
