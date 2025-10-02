package v1

import (
	"net/http"
	"waste-space/internal/dto"
	"waste-space/internal/middleware"
	"waste-space/internal/service"
	apperrors "waste-space/pkg/errors"

	"github.com/gin-gonic/gin"
)

type DumpsterController struct {
	dumpsterService service.DumpsterService
}

func NewDumpsterController(dumpsterService service.DumpsterService) *DumpsterController {
	return &DumpsterController{
		dumpsterService: dumpsterService,
	}
}

func (c *DumpsterController) initDumpsterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	dumpsters := rg.Group("/dumpsters")
	{
		dumpsters.GET("", c.list)
		dumpsters.GET("/search", c.search)
		dumpsters.GET("/nearby", c.nearby)
		dumpsters.GET("/:id", c.getByID)
		dumpsters.GET("/:id/availability", c.checkAvailability)

		dumpsters.Use(authMiddleware)
		{
			dumpsters.POST("", c.create)
			dumpsters.PUT("/:id", c.update)
			dumpsters.DELETE("/:id", c.delete)
			dumpsters.POST("/:id/book", c.book)
		}
	}
}

// @Summary List dumpsters
// @Tags dumpsters
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param sortBy query string false "Sort by: price|distance|rating|availability"
// @Param location query string false "Coordinates lat,lng"
// @Param maxPrice query number false "Maximum price per day"
// @Param size query string false "Size: small|medium|large|extraLarge"
// @Param availableNow query boolean false "Available now"
// @Param maxDistance query number false "Maximum distance in km"
// @Success 200 {object} dto.DumpsterListResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/dumpsters [get]
func (c *DumpsterController) list(ctx *gin.Context) {
	var req dto.DumpsterListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.dumpsterService.List(ctx.Request.Context(), req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Get dumpster by ID
// @Tags dumpsters
// @Accept json
// @Produce json
// @Param id path string true "Dumpster ID"
// @Success 200 {object} dto.DumpsterResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/dumpsters/{id} [get]
func (c *DumpsterController) getByID(ctx *gin.Context) {
	id := ctx.Param("id")

	response, err := c.dumpsterService.GetByID(ctx.Request.Context(), id)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Create dumpster
// @Tags dumpsters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateDumpsterRequest true "Dumpster data"
// @Success 201 {object} dto.DumpsterResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/dumpsters [post]
func (c *DumpsterController) create(ctx *gin.Context) {
	userID, ok := c.getUserIDFromContext(ctx)
	if !ok {
		return
	}

	var req dto.CreateDumpsterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.dumpsterService.Create(ctx.Request.Context(), userID, req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// @Summary Update dumpster
// @Tags dumpsters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Dumpster ID"
// @Param request body dto.UpdateDumpsterRequest true "Dumpster update data"
// @Success 200 {object} dto.DumpsterResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/dumpsters/{id} [put]
func (c *DumpsterController) update(ctx *gin.Context) {
	userID, ok := c.getUserIDFromContext(ctx)
	if !ok {
		return
	}

	id := ctx.Param("id")

	var req dto.UpdateDumpsterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.dumpsterService.Update(ctx.Request.Context(), userID, id, req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Delete dumpster
// @Tags dumpsters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Dumpster ID"
// @Success 204
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/dumpsters/{id} [delete]
func (c *DumpsterController) delete(ctx *gin.Context) {
	userID, ok := c.getUserIDFromContext(ctx)
	if !ok {
		return
	}

	id := ctx.Param("id")

	if err := c.dumpsterService.Delete(ctx.Request.Context(), userID, id); err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// @Summary Search dumpsters
// @Tags dumpsters
// @Accept json
// @Produce json
// @Param q query string false "Search query"
// @Param city query string false "City"
// @Param state query string false "State"
// @Param zipCode query string false "Zip code"
// @Param minPrice query number false "Minimum price"
// @Param maxPrice query number false "Maximum price"
// @Param size query string false "Size: small|medium|large|extraLarge"
// @Param isAvailable query boolean false "Available"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.DumpsterListResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/dumpsters/search [get]
func (c *DumpsterController) search(ctx *gin.Context) {
	var req dto.DumpsterSearchRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.dumpsterService.Search(ctx.Request.Context(), req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Find nearby dumpsters
// @Tags dumpsters
// @Accept json
// @Produce json
// @Param lat query number true "Latitude"
// @Param lng query number true "Longitude"
// @Param maxDistance query number false "Maximum distance in km" default(25)
// @Param limit query int false "Maximum results" default(20)
// @Success 200 {array} dto.DumpsterResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/dumpsters/nearby [get]
func (c *DumpsterController) nearby(ctx *gin.Context) {
	var req dto.NearbyDumpstersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.dumpsterService.FindNearby(ctx.Request.Context(), req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Book dumpster
// @Tags dumpsters
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Dumpster ID"
// @Param request body dto.BookDumpsterRequest true "Booking data"
// @Success 201 {object} dto.BookingResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/dumpsters/{id}/book [post]
func (c *DumpsterController) book(ctx *gin.Context) {
	userID, ok := c.getUserIDFromContext(ctx)
	if !ok {
		return
	}

	id := ctx.Param("id")

	var req dto.BookDumpsterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.dumpsterService.BookDumpster(ctx.Request.Context(), userID, id, req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// @Summary Check dumpster availability
// @Tags dumpsters
// @Accept json
// @Produce json
// @Param id path string true "Dumpster ID"
// @Success 200 {object} dto.AvailabilityResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/dumpsters/{id}/availability [get]
func (c *DumpsterController) checkAvailability(ctx *gin.Context) {
	id := ctx.Param("id")

	response, err := c.dumpsterService.CheckAvailability(ctx.Request.Context(), id)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *DumpsterController) getUserIDFromContext(ctx *gin.Context) (string, bool) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		handleError(ctx, apperrors.Unauthorized("unauthorized"))
		return "", false
	}
	return userID.String(), true
}
