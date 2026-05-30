package analytics

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, handler *Handler) {
	r.GET("/urls/:shortCode/analytics", handler.GetURLAnalytics)
}
