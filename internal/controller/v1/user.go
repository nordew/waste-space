package v1

import (
	"net/http"
	"waste-space/internal/dto"
	"waste-space/internal/middleware"
	"waste-space/internal/service"
	apperrors "waste-space/pkg/errors"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService service.UserService
}

func NewUserController(userService service.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (c *UserController) initUserRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	users := rg.Group("/users")
	users.Use(authMiddleware)
	{
		users.GET("/me", c.getMe)
		users.PUT("/me", c.updateMe)
		users.PATCH("/me/email", c.updateEmail)
		users.PATCH("/me/phone", c.updatePhone)
		users.PATCH("/me/password", c.updatePassword)
		users.DELETE("/me", c.deleteMe)
		users.GET("/:id", c.getByID)
	}
}

// @Summary Get current user profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/users/me [get]
func (c *UserController) getMe(ctx *gin.Context) {
	userID, ok := c.getUserIDFromContext(ctx)
	if !ok {
		return
	}

	response, err := c.userService.GetMe(ctx.Request.Context(), userID)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Update current user profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateUserRequest true "User update data"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/users/me [put]
func (c *UserController) updateMe(ctx *gin.Context) {
	userID, ok := c.getUserIDFromContext(ctx)
	if !ok {
		return
	}

	var req dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.userService.UpdateMe(ctx.Request.Context(), userID, req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Update current user email
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateEmailRequest true "Email update data"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/users/me/email [patch]
func (c *UserController) updateEmail(ctx *gin.Context) {
	userID, ok := c.getUserIDFromContext(ctx)
	if !ok {
		return
	}

	var req dto.UpdateEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.userService.UpdateEmail(ctx.Request.Context(), userID, req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Update current user phone number
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdatePhoneRequest true "Phone update data"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/users/me/phone [patch]
func (c *UserController) updatePhone(ctx *gin.Context) {
	userID, ok := c.getUserIDFromContext(ctx)
	if !ok {
		return
	}

	var req dto.UpdatePhoneRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.userService.UpdatePhone(ctx.Request.Context(), userID, req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Update current user password
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdatePasswordRequest true "Password update data"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/users/me/password [patch]
func (c *UserController) updatePassword(ctx *gin.Context) {
	userID, ok := c.getUserIDFromContext(ctx)
	if !ok {
		return
	}

	var req dto.UpdatePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	if err := c.userService.UpdatePassword(ctx.Request.Context(), userID, req); err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// @Summary Delete current user account
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 204
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/users/me [delete]
func (c *UserController) deleteMe(ctx *gin.Context) {
	userID, ok := c.getUserIDFromContext(ctx)
	if !ok {
		return
	}

	if err := c.userService.DeleteMe(ctx.Request.Context(), userID); err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// @Summary Get user by ID
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/users/{id} [get]
func (c *UserController) getByID(ctx *gin.Context) {
	id := ctx.Param("id")

	response, err := c.userService.GetByID(ctx.Request.Context(), id)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *UserController) getUserIDFromContext(ctx *gin.Context) (string, bool) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		handleError(ctx, apperrors.Unauthorized("unauthorized"))
		return "", false
	}
	return userID.String(), true
}
