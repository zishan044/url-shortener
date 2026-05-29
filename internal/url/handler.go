package url

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zishan044/url-shortener/internal/cache"
	"github.com/zishan044/url-shortener/internal/models"
	"github.com/zishan044/url-shortener/internal/queue"
)

type Handler struct {
	service   Service
	publisher *queue.Publisher
}

func NewHandler(service Service, publisher *queue.Publisher) *Handler {
	return &Handler{service: service, publisher: publisher}
}

type CreateUrlRequest struct {
	OriginalURL string    `json:"original_url" binding:"required,url"`
	ExpiresAt   time.Time `json:"expires_at"`
}

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

	url := &models.Url{
		UserID:      userID.(uuid.UUID),
		OriginalURL: req.OriginalURL,
		ExpiresAt:   req.ExpiresAt,
	}

	if err := url.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.CreateUrl(c.Request.Context(), url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "URL created successfully",
		"url":     url,
	})
}

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
			// Log error but don't fail the request
		}
	}

	h.trackClickAsync(c, url)

	c.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}

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


