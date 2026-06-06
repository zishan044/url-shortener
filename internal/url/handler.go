package url

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zishan044/url-shortener/internal/cache"
	"github.com/zishan044/url-shortener/internal/models"
	"github.com/zishan044/url-shortener/internal/queue"
	"github.com/zishan044/url-shortener/internal/validators"
)

type Handler struct {
	service   Service
	publisher *queue.Publisher
}

func NewHandler(service Service, publisher *queue.Publisher) *Handler {
	return &Handler{service: service, publisher: publisher}
}

type CreateUrlRequest struct {
	OriginalURL string    `json:"original_url" binding:"required,url" example:"https://github.com/zishan044"`
	ExpiresAt   time.Time `json:"expires_at" example:"2025-12-31T23:59:59Z"`
}

type CreateUrlResponse struct {
	Message string       `json:"message"`
	URL     models.Url `json:"url"`
}

type GetUrlResponse struct {
	URL models.Url `json:"url"`
}

// CreateUrl godoc
// @Summary Create a short URL
// @Description Create a new shortened URL for authenticated users
// @Tags urls
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateUrlRequest true "Create URL request"
// @Success 201 {object} CreateUrlResponse "URL created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /urls/ [post]
func (h *Handler) CreateUrl(c *gin.Context) {
	var req CreateUrlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	sanitizedURL, err := validators.SanitizeURL(req.OriginalURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url := &models.Url{
		UserID:      userID.(uuid.UUID),
		OriginalURL: sanitizedURL,
		ExpiresAt:   req.ExpiresAt,
	}

	if err := url.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.CreateUrl(c.Request.Context(), url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "URL created successfully",
		"url":     url,
	})
}

// GetUrlByShortCode godoc
// @Summary Get URL by short code
// @Description Retrieve the original URL and associated metadata for a short code
// @Tags urls
// @Produce json
// @Param shortCode path string true "Short code"
// @Success 200 {object} GetUrlResponse "URL retrieved successfully"
// @Failure 404 {object} map[string]interface{} "URL not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /urls/{shortCode} [get]
func (h *Handler) GetUrlByShortCode(c *gin.Context) {
	shortCode := c.Param("shortCode")
	cacheKey := "url:" + shortCode

	cachedUrl, err := cache.Get[models.Url](c.Request.Context(), cacheKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cache error"})
		return
	}

	var url *models.Url
	if cachedUrl != nil {
		url = cachedUrl
	} else {
		url, err = h.service.GetUrlByShortCode(c.Request.Context(), shortCode)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}

		if err := cache.Set(c.Request.Context(), cacheKey, *url, 1*time.Hour); err != nil {
		}
	}

	h.trackClickAsync(c, url)

	c.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}

// Redirect godoc
// @Summary Redirect to original URL
// @Description Redirect to the original URL and record analytics
// @Tags urls
// @Param shortCode path string true "Short code"
// @Success 301 "Redirect to original URL"
// @Failure 404 {object} map[string]interface{} "URL not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /r/{shortCode} [get]
func (h *Handler) Redirect(c *gin.Context) {
	shortCode := c.Param("shortCode")
	cacheKey := "url:" + shortCode

	cachedUrl, err := cache.Get[models.Url](c.Request.Context(), cacheKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cache error"})
		return
	}

	var url *models.Url
	if cachedUrl != nil {
		url = cachedUrl
	} else {
		url, err = h.service.GetUrlByShortCode(c.Request.Context(), shortCode)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
			return
		}

		if err := cache.Set(c.Request.Context(), cacheKey, *url, 1*time.Hour); err != nil {

		}
	}

	
	h.trackClickAsync(c, url)

	c.Redirect(http.StatusMovedPermanently, url.OriginalURL)
}

func (h *Handler) trackClickAsync(c *gin.Context, url *models.Url) {
	if h.publisher == nil {
		return
	}

	go func() {
		click := &models.Click{
			ID:        uuid.New(),
			URLID:     url.ID,
			Timestamp: time.Now(),
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Referrer:  c.Request.Referer(),
		}

		if err := h.publisher.PublishClick(c.Request.Context(), click); err != nil {
			// Log error but don't fail the request
		}
	}()
}


