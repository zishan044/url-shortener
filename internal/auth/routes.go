package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, handler *Handler) {
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", handler.Register)
		authGroup.POST("/login", handler.Login)
	}
}