package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/zishan044/url-shortener/internal/models"
)

// MockAuthService implements Service interface for testing
type MockAuthService struct {
	RegisterFunc func(ctx context.Context, username, email, password string) (*models.User, error)
	LoginFunc    func(ctx context.Context, email, password string) (string, error)
}

func (m *MockAuthService) Register(ctx context.Context, username, email, password string) (*models.User, error) {
	return m.RegisterFunc(ctx, username, email, password)
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (string, error) {
	return m.LoginFunc(ctx, email, password)
}

func TestRegisterSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockAuthService{
		RegisterFunc: func(ctx context.Context, username, email, password string) (*models.User, error) {
			return &models.User{
				ID:       uuid.New(),
				Username: username,
				Email:    email,
			}, nil
		},
	}

	handler := NewHandler(mockService)
	router := gin.New()
	router.POST("/register", handler.Register)

	reqBody := RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "User registered successfully" {
		t.Errorf("expected success message, got %v", response["message"])
	}
}

func TestRegisterInvalidEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockAuthService{}
	handler := NewHandler(mockService)
	router := gin.New()
	router.POST("/register", handler.Register)

	reqBody := RegisterRequest{
		Username: "testuser",
		Email:    "invalid-email",
		Password: "password123",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRegisterWeakPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockAuthService{}
	handler := NewHandler(mockService)
	router := gin.New()
	router.POST("/register", handler.Register)

	reqBody := RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "weak",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLoginSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockAuthService{
		LoginFunc: func(ctx context.Context, email, password string) (string, error) {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"email": email,
			})
			return token.SignedString([]byte("test-secret"))
		},
	}

	handler := NewHandler(mockService)
	router := gin.New()
	router.POST("/login", handler.Login)

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response LoginResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.TokenType != "Bearer" {
		t.Errorf("expected token_type Bearer, got %s", response.TokenType)
	}

	if response.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestLoginInvalidEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockAuthService{}
	handler := NewHandler(mockService)
	router := gin.New()
	router.POST("/login", handler.Login)

	reqBody := LoginRequest{
		Email:    "invalid-email",
		Password: "password123",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLoginEmptyPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &MockAuthService{}
	handler := NewHandler(mockService)
	router := gin.New()
	router.POST("/login", handler.Login)

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
