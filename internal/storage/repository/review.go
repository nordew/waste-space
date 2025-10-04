package repository

import (
	"context"
	"errors"
	"waste-space/internal/dto"
	"waste-space/internal/model"
	apperrors "waste-space/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(ctx context.Context, review *model.Review) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Review, error)
	Update(ctx context.Context, review *model.Review) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByDumpsterID(ctx context.Context, dumpsterID uuid.UUID, req dto.ReviewListRequest) ([]*model.Review, int64, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, req dto.ReviewListRequest) ([]*model.Review, int64, error)
	GetByUserAndDumpster(ctx context.Context, userID, dumpsterID uuid.UUID) (*model.Review, error)
	GetAverageRating(ctx context.Context, dumpsterID uuid.UUID) (float64, error)
	GetReviewCount(ctx context.Context, dumpsterID uuid.UUID) (int, error)
}

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(ctx context.Context, review *model.Review) error {
	result := r.db.WithContext(ctx).Create(review)
	if result.Error != nil {
		return apperrors.Internal("failed to create review", result.Error)
	}
	return nil
}

func (r *reviewRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Review, error) {
	var review model.Review
	result := r.db.WithContext(ctx).Preload("User").Preload("Dumpster").Where("id = ?", id).First(&review)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("review not found")
		}
		return nil, apperrors.Internal("failed to get review", result.Error)
	}
	return &review, nil
}

func (r *reviewRepository) Update(ctx context.Context, review *model.Review) error {
	result := r.db.WithContext(ctx).Save(review)
	if result.Error != nil {
		return apperrors.Internal("failed to update review", result.Error)
	}

	if result.RowsAffected == 0 {
		return apperrors.NotFound("review not found")
	}

	return nil
}

func (r *reviewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Review{}, id)
	if result.Error != nil {
		return apperrors.Internal("failed to delete review", result.Error)
	}

	if result.RowsAffected == 0 {
		return apperrors.NotFound("review not found")
	}

	return nil
}

func (r *reviewRepository) GetByDumpsterID(
	ctx context.Context,
	dumpsterID uuid.UUID,
	req dto.ReviewListRequest) ([]*model.Review, int64, error) {
	var reviews []*model.Review
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Review{}).Preload("User").Where("dumpster_id = ?", dumpsterID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Internal("failed to count reviews", err)
	}

	page := max(req.Page, 1)
	limit := max(req.Limit, defaultPageSize)
	if limit > maxPageSize {
		limit = maxPageSize
	}

	offset := (page - 1) * limit

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&reviews).Error; err != nil {
		return nil, 0, apperrors.Internal("failed to get reviews", err)
	}

	return reviews, total, nil
}

func (r *reviewRepository) GetByUserID(
	ctx context.Context,
	userID uuid.UUID,
	req dto.ReviewListRequest) ([]*model.Review, int64, error) {
	var reviews []*model.Review
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Review{}).Preload("Dumpster").Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Internal("failed to count reviews", err)
	}

	page := max(req.Page, 1)
	limit := max(req.Limit, defaultPageSize)
	if limit > maxPageSize {
		limit = maxPageSize
	}

	offset := (page - 1) * limit

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&reviews).Error; err != nil {
		return nil, 0, apperrors.Internal("failed to get reviews", err)
	}

	return reviews, total, nil
}

func (r *reviewRepository) GetByUserAndDumpster(
	ctx context.Context,
	userID, dumpsterID uuid.UUID) (*model.Review, error) {
	var review model.Review
	result := r.db.WithContext(ctx).Where("user_id = ? AND dumpster_id = ?", userID, dumpsterID).First(&review)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.Internal("failed to get review", result.Error)
	}
	return &review, nil
}

func (r *reviewRepository) GetAverageRating(ctx context.Context, dumpsterID uuid.UUID) (float64, error) {
	var avgRating float64
	result := r.db.WithContext(ctx).Model(&model.Review{}).Where("dumpster_id = ?", dumpsterID).Select("COALESCE(AVG(rating), 0)").Scan(&avgRating)
	if result.Error != nil {
		return 0, apperrors.Internal("failed to calculate average rating", result.Error)
	}
	return avgRating, nil
}

func (r *reviewRepository) GetReviewCount(ctx context.Context, dumpsterID uuid.UUID) (int, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&model.Review{}).Where("dumpster_id = ?", dumpsterID).Count(&count)
	if result.Error != nil {
		return 0, apperrors.Internal("failed to count reviews", result.Error)
	}
	return int(count), nil
}
