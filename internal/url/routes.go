package url

import (
	"github.com/gin-gonic/gin"
	"github.com/zishan044/url-shortener/internal/middleware"
)

func RegisterRoutes(r *gin.RouterGroup, handler *Handler, jwtSecret string) {
	urlGroup := r.Group("/urls")
	{
		urlGroup.POST("/", middleware.AuthMiddleware(jwtSecret), handler.CreateUrl)
		urlGroup.GET("/:shortCode", handler.GetUrlByShortCode)
	}
}