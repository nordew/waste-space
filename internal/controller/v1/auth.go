package v1

import (
	"net/http"
	"waste-space/internal/dto"
	"waste-space/internal/service"
	apperrors "waste-space/pkg/errors"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	userService service.UserService
}

func NewAuthController(userService service.UserService) *AuthController {
	return &AuthController{
		userService: userService,
	}
}

func (c *AuthController) initAuthRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", c.register)
		auth.POST("/login", c.login)
		auth.POST("/refresh", c.refreshToken)
	}
}

// @Summary Register new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.CreateUserRequest true "User registration data"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/auth/register [post]
func (c *AuthController) register(ctx *gin.Context) {
	var req dto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.userService.Register(ctx.Request.Context(), req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// @Summary Login user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/login [post]
func (c *AuthController) login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.userService.Login(ctx.Request.Context(), req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dto.RefreshTokenResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/refresh [post]
func (c *AuthController) refreshToken(ctx *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.userService.RefreshToken(ctx.Request.Context(), req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func handleError(ctx *gin.Context, err error) {
	status := apperrors.GetHTTPStatus(err)
	ctx.JSON(status, gin.H{"error": err.Error()})
}
