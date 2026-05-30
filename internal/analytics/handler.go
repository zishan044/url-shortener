package analytics

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetURLAnalytics(c *gin.Context) {
	rawID := c.Param("id")
	if rawID == "" {
		rawID = c.Param("shortCode")
	}

	urlID, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "invalid_url_id",
				"message": "url id must be a valid UUID",
			},
		})
		return
	}

	response, err := h.service.GetURLAnalytics(c.Request.Context(), urlID)
	if err != nil {
		if errors.Is(err, ErrURLNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "url_not_found",
					"message": "url not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "analytics_unavailable",
				"message": "failed to load analytics",
			},
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
