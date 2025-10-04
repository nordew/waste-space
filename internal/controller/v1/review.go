package v1

import (
	"net/http"
	"waste-space/internal/dto"
	"waste-space/internal/middleware"
	"waste-space/internal/service"
	apperrors "waste-space/pkg/errors"

	"github.com/gin-gonic/gin"
)

type ReviewController struct {
	reviewService service.ReviewService
}

func NewReviewController(reviewService service.ReviewService) *ReviewController {
	return &ReviewController{
		reviewService: reviewService,
	}
}

func (c *ReviewController) initReviewRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	reviews := rg.Group("/reviews")
	{
		reviews.GET("/:id", c.getByID)

		reviews.Use(authMiddleware)
		{
			reviews.GET("/user/:userId", c.getUserReviews)
		}
	}

	dumpsters := rg.Group("/dumpsters/:id")
	{
		dumpsters.GET("/reviews", c.getDumpsterReviews)

		dumpsters.Use(authMiddleware)
		{
			dumpsters.POST("/reviews", c.create)
			dumpsters.PUT("/reviews/:reviewId", c.update)
			dumpsters.DELETE("/reviews/:reviewId", c.delete)
		}
	}
}

// @Summary Get review by ID
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path string true "Review ID"
// @Success 200 {object} dto.ReviewResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/reviews/{id} [get]
func (c *ReviewController) getByID(ctx *gin.Context) {
	id := ctx.Param("id")

	response, err := c.reviewService.GetByID(ctx.Request.Context(), id)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Create review for dumpster
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Dumpster ID"
// @Param request body dto.CreateReviewRequest true "Review data"
// @Success 201 {object} dto.ReviewResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/dumpsters/{id}/reviews [post]
func (c *ReviewController) create(ctx *gin.Context) {
	userID, ok := c.getUserIDFromContext(ctx)
	if !ok {
		return
	}

	dumpsterID := ctx.Param("id")

	var req dto.CreateReviewRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.reviewService.Create(ctx.Request.Context(), userID, dumpsterID, req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

// @Summary Update review
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Dumpster ID"
// @Param reviewId path string true "Review ID"
// @Param request body dto.UpdateReviewRequest true "Review update data"
// @Success 200 {object} dto.ReviewResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/dumpsters/{id}/reviews/{reviewId} [put]
func (c *ReviewController) update(ctx *gin.Context) {
	userID, ok := c.getUserIDFromContext(ctx)
	if !ok {
		return
	}

	id := ctx.Param("reviewId")

	var req dto.UpdateReviewRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.reviewService.Update(ctx.Request.Context(), userID, id, req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Delete review
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Dumpster ID"
// @Param reviewId path string true "Review ID"
// @Success 204
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/dumpsters/{id}/reviews/{reviewId} [delete]
func (c *ReviewController) delete(ctx *gin.Context) {
	userID, ok := c.getUserIDFromContext(ctx)
	if !ok {
		return
	}

	id := ctx.Param("reviewId")

	if err := c.reviewService.Delete(ctx.Request.Context(), userID, id); err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// @Summary Get reviews for dumpster
// @Tags reviews
// @Accept json
// @Produce json
// @Param id path string true "Dumpster ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.ReviewListResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/dumpsters/{id}/reviews [get]
func (c *ReviewController) getDumpsterReviews(ctx *gin.Context) {
	dumpsterID := ctx.Param("id")

	var req dto.ReviewListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.reviewService.GetByDumpsterID(ctx.Request.Context(), dumpsterID, req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// @Summary Get reviews by user
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param userId path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.ReviewListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/reviews/user/{userId} [get]
func (c *ReviewController) getUserReviews(ctx *gin.Context) {
	userID := ctx.Param("userId")

	var req dto.ReviewListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		handleError(ctx, apperrors.BadRequest(err.Error()))
		return
	}

	response, err := c.reviewService.GetByUserID(ctx.Request.Context(), userID, req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *ReviewController) getUserIDFromContext(ctx *gin.Context) (string, bool) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		handleError(ctx, apperrors.Unauthorized("unauthorized"))
		return "", false
	}
	return userID.String(), true
}
