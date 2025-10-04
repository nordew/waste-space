package service

import (
	"context"
	"math"
	"waste-space/internal/dto"
	"waste-space/internal/model"
	"waste-space/internal/storage/repository"
	apperrors "waste-space/pkg/errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ReviewService interface {
	Create(ctx context.Context, userID, dumpsterID string, req dto.CreateReviewRequest) (*dto.ReviewResponse, error)
	GetByID(ctx context.Context, id string) (*dto.ReviewResponse, error)
	Update(ctx context.Context, userID, id string, req dto.UpdateReviewRequest) (*dto.ReviewResponse, error)
	Delete(ctx context.Context, userID, id string) error
	GetByDumpsterID(ctx context.Context, dumpsterID string, req dto.ReviewListRequest) (*dto.ReviewListResponse, error)
	GetByUserID(ctx context.Context, userID string, req dto.ReviewListRequest) (*dto.ReviewListResponse, error)
}

type reviewService struct {
	reviewRepo   repository.ReviewRepository
	dumpsterRepo repository.DumpsterRepository
	logger       *zap.Logger
}

func NewReviewService(
	reviewRepo repository.ReviewRepository,
	dumpsterRepo repository.DumpsterRepository,
	logger *zap.Logger) ReviewService {
	return &reviewService{
		reviewRepo:   reviewRepo,
		dumpsterRepo: dumpsterRepo,
		logger:       logger,
	}
}

func (s *reviewService) Create(
	ctx context.Context,
	userID, dumpsterID string,
	req dto.CreateReviewRequest) (*dto.ReviewResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.BadRequest("invalid user ID")
	}

	dumpsterUUID, err := uuid.Parse(dumpsterID)
	if err != nil {
		return nil, apperrors.BadRequest("invalid dumpster ID")
	}

	if _, err := s.dumpsterRepo.GetByID(ctx, dumpsterUUID); err != nil {
		return nil, err
	}

	existingReview, err := s.reviewRepo.GetByUserAndDumpster(ctx, userUUID, dumpsterUUID)
	if err != nil {
		s.logger.Error("failed to check existing review", zap.String("userId", userID), zap.String("dumpsterId", dumpsterID), zap.Error(err))
		return nil, err
	}
	if existingReview != nil {
		return nil, apperrors.BadRequest("you have already reviewed this dumpster")
	}

	review := model.NewReviewFromDTO(userUUID, dumpsterUUID, req)

	if err := s.reviewRepo.Create(ctx, review); err != nil {
		s.logger.Error("failed to create review", zap.String("userId", userID), zap.String("dumpsterId", dumpsterID), zap.Error(err))
		return nil, err
	}

	if err := s.updateDumpsterRating(ctx, dumpsterUUID); err != nil {
		s.logger.Error("failed to update dumpster rating", zap.String("dumpsterId", dumpsterID), zap.Error(err))
		return nil, err
	}

	response := review.ToResponse()
	return &response, nil
}

func (s *reviewService) GetByID(ctx context.Context, id string) (*dto.ReviewResponse, error) {
	reviewID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.BadRequest("invalid review ID")
	}

	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	response := review.ToResponse()
	return &response, nil
}

func (s *reviewService) Update(
	ctx context.Context,
	userID, id string,
	req dto.UpdateReviewRequest) (*dto.ReviewResponse, error) {
	reviewID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.BadRequest("invalid review ID")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.BadRequest("invalid user ID")
	}

	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return nil, err
	}

	if review.UserID != userUUID {
		return nil, apperrors.Forbidden("you don't have permission to update this review")
	}

	s.applyReviewUpdates(review, req)

	if err := s.reviewRepo.Update(ctx, review); err != nil {
		s.logger.Error("failed to update review", zap.String("reviewId", id), zap.Error(err))
		return nil, err
	}

	if err := s.updateDumpsterRating(ctx, review.DumpsterID); err != nil {
		s.logger.Error("failed to update dumpster rating after review update", zap.String("dumpsterId", review.DumpsterID.String()), zap.Error(err))
		return nil, err
	}

	response := review.ToResponse()
	return &response, nil
}

func (s *reviewService) Delete(ctx context.Context, userID, id string) error {
	reviewID, err := uuid.Parse(id)
	if err != nil {
		return apperrors.BadRequest("invalid review ID")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.BadRequest("invalid user ID")
	}

	review, err := s.reviewRepo.GetByID(ctx, reviewID)
	if err != nil {
		return err
	}

	if review.UserID != userUUID {
		return apperrors.Forbidden("you don't have permission to delete this review")
	}

	dumpsterID := review.DumpsterID

	if err := s.reviewRepo.Delete(ctx, reviewID); err != nil {
		s.logger.Error("failed to delete review", zap.String("reviewId", id), zap.Error(err))
		return err
	}

	if err := s.updateDumpsterRating(ctx, dumpsterID); err != nil {
		s.logger.Error("failed to update dumpster rating after review deletion", zap.String("dumpsterId", dumpsterID.String()), zap.Error(err))
		return err
	}

	return nil
}

func (s *reviewService) GetByDumpsterID(
	ctx context.Context,
	dumpsterID string,
	req dto.ReviewListRequest) (*dto.ReviewListResponse, error) {
	dumpsterUUID, err := uuid.Parse(dumpsterID)
	if err != nil {
		return nil, apperrors.BadRequest("invalid dumpster ID")
	}

	reviews, total, err := s.reviewRepo.GetByDumpsterID(ctx, dumpsterUUID, req)
	if err != nil {
		s.logger.Error("failed to get reviews by dumpster", zap.String("dumpsterId", dumpsterID), zap.Error(err))
		return nil, err
	}

	return s.buildReviewListResponse(reviews, total, req.Page, req.Limit), nil
}

func (s *reviewService) GetByUserID(
	ctx context.Context,
	userID string,
	req dto.ReviewListRequest) (*dto.ReviewListResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.BadRequest("invalid user ID")
	}

	reviews, total, err := s.reviewRepo.GetByUserID(ctx, userUUID, req)
	if err != nil {
		s.logger.Error("failed to get reviews by user", zap.String("userId", userID), zap.Error(err))
		return nil, err
	}

	return s.buildReviewListResponse(reviews, total, req.Page, req.Limit), nil
}

func (s *reviewService) applyReviewUpdates(review *model.Review, req dto.UpdateReviewRequest) {
	if req.Rating != nil {
		review.Rating = *req.Rating
	}
	if req.Comment != nil {
		review.Comment = *req.Comment
	}
}

func (s *reviewService) updateDumpsterRating(ctx context.Context, dumpsterID uuid.UUID) error {
	avgRating, err := s.reviewRepo.GetAverageRating(ctx, dumpsterID)
	if err != nil {
		s.logger.Error("failed to get average rating", zap.String("dumpsterId", dumpsterID.String()), zap.Error(err))
		return err
	}

	reviewCount, err := s.reviewRepo.GetReviewCount(ctx, dumpsterID)
	if err != nil {
		s.logger.Error("failed to get review count", zap.String("dumpsterId", dumpsterID.String()), zap.Error(err))
		return err
	}

	dumpster, err := s.dumpsterRepo.GetByID(ctx, dumpsterID)
	if err != nil {
		s.logger.Error("failed to get dumpster for rating update", zap.String("dumpsterId", dumpsterID.String()), zap.Error(err))
		return err
	}

	dumpster.Rating = avgRating
	dumpster.ReviewCount = reviewCount

	if err := s.dumpsterRepo.Update(ctx, dumpster); err != nil {
		s.logger.Error("failed to save updated dumpster rating", zap.String("dumpsterId", dumpsterID.String()), zap.Error(err))
		return err
	}

	return nil
}

func (s *reviewService) buildReviewListResponse(
	reviews []*model.Review,
	total int64,
	page, limit int) *dto.ReviewListResponse {
	page = max(page, 1)
	limit = max(limit, 1)

	responses := make([]dto.ReviewResponse, len(reviews))
	for i, review := range reviews {
		responses[i] = review.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &dto.ReviewListResponse{
		Reviews:    responses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}
