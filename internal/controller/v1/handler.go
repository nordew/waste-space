package v1

import (
	"waste-space/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "waste-space/docs"
)

type Handler struct {
	authController *AuthController
}

func NewHandler(userService service.UserService) *Handler {
	return &Handler{
		authController: NewAuthController(userService),
	}
}

func (h *Handler) InitRoutes(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	{
		h.authController.initAuthRoutes(v1)
	}
}
