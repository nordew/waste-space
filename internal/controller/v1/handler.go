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
	authController     *AuthController
	userController     *UserController
	dumpsterController *DumpsterController
	tokenService       auth.TokenService
}

func NewHandler(userService service.UserService, dumpsterService service.DumpsterService, tokenService auth.TokenService) *Handler {
	return &Handler{
		authController:     NewAuthController(userService),
		userController:     NewUserController(userService),
		dumpsterController: NewDumpsterController(dumpsterService),
		tokenService:       tokenService,
	}
}

func (h *Handler) InitRoutes(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authMW := middleware.Auth(h.tokenService)

	v1 := router.Group("/api/v1")
	{
		h.authController.initAuthRoutes(v1)
		h.userController.initUserRoutes(v1, authMW)
		h.dumpsterController.initDumpsterRoutes(v1, authMW)
	}
}
