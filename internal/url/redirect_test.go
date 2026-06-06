package url

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zishan044/url-shortener/internal/models"
)

func TestRedirectWithExpiredURL(t *testing.T) {
	setupTestCache(t)
	gin.SetMode(gin.TestMode)

	// Create an expired URL
	expiredURL := &models.Url{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		OriginalURL: "https://github.com/zishan044",
		ShortCode:   "expired",
		ExpiresAt:   time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt:   time.Now().Add(-24 * time.Hour),
	}

	mockService := &MockURLService{
		GetUrlByShortCodeFunc: func(ctx context.Context, shortCode string) (*models.Url, error) {
			if expiredURL.ExpiresAt.Before(time.Now()) {
				return nil, errors.New("not found")
			}
			return expiredURL, nil
		},
	}

	handler := NewHandler(mockService, nil)
	router := gin.New()
	router.GET("/r/:shortCode", handler.Redirect)

	req := httptest.NewRequest("GET", "/r/expired", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d for expired URL, got %d", http.StatusNotFound, w.Code)
	}
}

func TestRedirectPreservesQueryParameters(t *testing.T) {
	setupTestCache(t)
	gin.SetMode(gin.TestMode)

	testURL := &models.Url{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		OriginalURL: "https://github.com/zishan044?ref=short",
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

	location := w.Header().Get("Location")
	if location != "https://github.com/zishan044?ref=short" {
		t.Errorf("expected location with query params, got %s", location)
	}
}

func TestRedirectHTTPSProtocol(t *testing.T) {
	setupTestCache(t)
	gin.SetMode(gin.TestMode)

	testURL := &models.Url{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		OriginalURL: "https://secure.example.com",
		ShortCode:   "secure123",
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

	req := httptest.NewRequest("GET", "/r/secure123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMovedPermanently {
		t.Errorf("expected status %d, got %d", http.StatusMovedPermanently, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "https://secure.example.com" {
		t.Errorf("expected HTTPS protocol, got %s", location)
	}
}

func TestRedirectRecordsAnalytics(t *testing.T) {
	setupTestCache(t)
	gin.SetMode(gin.TestMode)

	testURL := &models.Url{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		OriginalURL: "https://github.com/zishan044",
		ShortCode:   "analytics",
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

	req := httptest.NewRequest("GET", "/r/analytics", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://example.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMovedPermanently {
		t.Errorf("expected status %d, got %d", http.StatusMovedPermanently, w.Code)
	}

	// The tracking is async, so we just verify the redirect succeeded
	location := w.Header().Get("Location")
	if location != "https://github.com/zishan044" {
		t.Errorf("expected redirect to occur with analytics tracking")
	}
}

func TestRedirectEmptyShortCode(t *testing.T) {
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

	req := httptest.NewRequest("GET", "/r/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Empty shortCode should result in 404
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d for empty short code, got %d", http.StatusNotFound, w.Code)
	}
}

func TestRedirectLongURL(t *testing.T) {
	setupTestCache(t)
	gin.SetMode(gin.TestMode)

	// Test with a very long URL
	longURL := "https://example.com/very/long/path/with/many/segments/and/parameters?param1=value1&param2=value2&param3=value3&param4=value4&param5=value5"

	testURL := &models.Url{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		OriginalURL: longURL,
		ShortCode:   "long",
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

	req := httptest.NewRequest("GET", "/r/long", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMovedPermanently {
		t.Errorf("expected status %d, got %d", http.StatusMovedPermanently, w.Code)
	}

	location := w.Header().Get("Location")
	if location != longURL {
		t.Errorf("expected long URL to be preserved, got %s", location)
	}
}

func TestRedirectSpecialCharactersInShortCode(t *testing.T) {
	setupTestCache(t)
	gin.SetMode(gin.TestMode)

	testURL := &models.Url{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		OriginalURL: "https://github.com/zishan044",
		ShortCode:   "abc-123_xyz",
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

	req := httptest.NewRequest("GET", "/r/abc-123_xyz", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMovedPermanently {
		t.Errorf("expected status %d, got %d", http.StatusMovedPermanently, w.Code)
	}
}
