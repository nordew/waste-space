package v1

import (
	"waste-space/internal/middleware"
	"waste-space/internal/service"
	"waste-space/pkg/auth"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "waste-space/docs"
)

type Handler struct {
	authController *AuthController
	userController *UserController
	tokenService   auth.TokenService
}

func NewHandler(userService service.UserService, tokenService auth.TokenService) *Handler {
	return &Handler{
		authController: NewAuthController(userService),
		userController: NewUserController(userService),
		tokenService:   tokenService,
	}
}

func (h *Handler) InitRoutes(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	{
		h.authController.initAuthRoutes(v1)
		h.userController.initUserRoutes(v1, middleware.Auth(h.tokenService))
	}
}
