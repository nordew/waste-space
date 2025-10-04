package v1

import (
	"net/http"
	"waste-space/internal/dto"
	"waste-space/internal/middleware"
	"waste-space/internal/service"
	apperrors "waste-space/pkg/errors"

	"github.com/gin-gonic/gin"
)

type UsageController struct {
	usageService service.UsageService
}

func NewUsageController(usageService service.UsageService) *UsageController {
	return &UsageController{
		usageService: usageService,
	}
}

func (c *UsageController) initUsageRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	usages := rg.Group("/usages")
	usages.Use(authMiddleware)
	{
		usages.GET("/:id", c.getByID)
		usages.GET("", c.list)
		usages.GET("/stats", c.getStats)
		usages.GET("/user/:userId", c.getUserUsages)
		usages.DELETE("/:id", c.delete)
	}

	dumpsters := rg.Group("/dumpsters/:id")
	dumpsters.Use(authMiddleware)
	{
		dumpsters.POST("/usages/start", c.startUsage)
		dumpsters.PUT("/usages/:usageId/end", c.endUsage)
		dumpsters.GET("/usages", c.getDumpsterUsages)
	}
}

// @Summary Start dumpster usage
// @Tags usages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Dumpster ID"
// @Param request body dto.StartUsageRequest true "Usage start data"
// @Success 201 {object} dto.UsageResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/dumpsters/{id}/usages/start [post]
func (c *UsageController) startUsage(ctx *gin.Context) {
	userID, ok := c.getUserIDFromContext(ctx)
	if !ok {
		return
	}

	dumpsterID := ctx.Param("id")

	var req dto.StartUsageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.usageService.StartUsage(ctx.Request.Context(), userID, dumpsterID, req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// @Summary End dumpster usage
// @Tags usages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Dumpster ID"
// @Param usageId path string true "Usage ID"
// @Param request body dto.EndUsageRequest true "Usage end data"
// @Success 200 {object} dto.UsageResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/dumpsters/{id}/usages/{usageId}/end [put]
func (c *UsageController) endUsage(ctx *gin.Context) {
	userID, ok := c.getUserIDFromContext(ctx)
	if !ok {
		return
	}

	usageID := ctx.Param("usageId")

	var req dto.EndUsageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.usageService.EndUsage(ctx.Request.Context(), userID, usageID, req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Get usage by ID
// @Tags usages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Usage ID"
// @Success 200 {object} dto.UsageResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/usages/{id} [get]
func (c *UsageController) getByID(ctx *gin.Context) {
	id := ctx.Param("id")

	response, err := c.usageService.GetByID(ctx.Request.Context(), id)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Get usages for dumpster
// @Tags usages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Dumpster ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Filter by status (active, completed, cancelled)"
// @Success 200 {object} dto.UsageListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/dumpsters/{id}/usages [get]
func (c *UsageController) getDumpsterUsages(ctx *gin.Context) {
	dumpsterID := ctx.Param("id")

	var req dto.UsageListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.usageService.GetByDumpsterID(ctx.Request.Context(), dumpsterID, req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Get usages by user
// @Tags usages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Filter by status (active, completed, cancelled)"
// @Success 200 {object} dto.UsageListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/usages/user/{userId} [get]
func (c *UsageController) getUserUsages(ctx *gin.Context) {
	userID := ctx.Param("userId")

	var req dto.UsageListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.usageService.GetByUserID(ctx.Request.Context(), userID, req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary List all usages with filters
// @Tags usages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Filter by status (active, completed, cancelled)"
// @Param dumpsterId query string false "Filter by dumpster ID"
// @Param userId query string false "Filter by user ID"
// @Success 200 {object} dto.UsageListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/usages [get]
func (c *UsageController) list(ctx *gin.Context) {
	var req dto.UsageListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.usageService.List(ctx.Request.Context(), req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Get usage statistics
// @Tags usages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param dumpsterId query string false "Filter by dumpster ID"
// @Param userId query string false "Filter by user ID"
// @Success 200 {object} dto.UsageStatsResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/usages/stats [get]
func (c *UsageController) getStats(ctx *gin.Context) {
	dumpsterID := ctx.Query("dumpsterId")
	userID := ctx.Query("userId")

	var dumpsterIDPtr *string
	var userIDPtr *string

	if dumpsterID != "" {
		dumpsterIDPtr = &dumpsterID
	}

	if userID != "" {
		userIDPtr = &userID
	}

	response, err := c.usageService.GetStats(ctx.Request.Context(), dumpsterIDPtr, userIDPtr)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Delete usage
// @Tags usages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Usage ID"
// @Success 204
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/usages/{id} [delete]
func (c *UsageController) delete(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.usageService.Delete(ctx.Request.Context(), id); err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func (c *UsageController) getUserIDFromContext(ctx *gin.Context) (string, bool) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		handleError(ctx, apperrors.Unauthorized("unauthorized"))
		return "", false
	}
	return userID.String(), true
}
