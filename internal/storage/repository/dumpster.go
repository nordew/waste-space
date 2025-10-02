package repository

import (
	"context"
	"errors"
	"waste-space/internal/model"
	apperrors "waste-space/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DumpsterSearchFilters struct {
	City        string
	State       string
	ZipCode     string
	MinPrice    *float64
	MaxPrice    *float64
	Size        string
	IsAvailable *bool
	Limit       int
	Offset      int
}

type DumpsterRepository interface {
	Create(ctx context.Context, dumpster *model.Dumpster) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Dumpster, error)
	GetByIDWithOwner(ctx context.Context, id uuid.UUID) (*model.Dumpster, error)
	GetByOwnerID(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*model.Dumpster, error)
	Update(ctx context.Context, dumpster *model.Dumpster) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*model.Dumpster, error)
	Search(ctx context.Context, filters DumpsterSearchFilters) ([]*model.Dumpster, error)
	Count(ctx context.Context) (int64, error)
	UpdateAvailability(ctx context.Context, id uuid.UUID, isAvailable bool) error
	UpdateRating(ctx context.Context, id uuid.UUID, rating float64, reviewCount int) error
}

type dumpsterRepository struct {
	db *gorm.DB
}

func NewDumpsterRepository(db *gorm.DB) DumpsterRepository {
	return &dumpsterRepository{db: db}
}

func (r *dumpsterRepository) Create(ctx context.Context, dumpster *model.Dumpster) error {
	result := r.db.WithContext(ctx).Create(dumpster)
	if result.Error != nil {
		return apperrors.Internal("failed to create dumpster", result.Error)
	}

	return nil
}

func (r *dumpsterRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Dumpster, error) {
	var dumpster model.Dumpster
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&dumpster)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("dumpster not found")
		}
		return nil, apperrors.Internal("failed to get dumpster", result.Error)
	}

	return &dumpster, nil
}

func (r *dumpsterRepository) GetByIDWithOwner(ctx context.Context, id uuid.UUID) (*model.Dumpster, error) {
	var dumpster model.Dumpster
	result := r.db.WithContext(ctx).Preload("Owner").Where("id = ?", id).First(&dumpster)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("dumpster not found")
		}
		return nil, apperrors.Internal("failed to get dumpster", result.Error)
	}

	return &dumpster, nil
}

func (r *dumpsterRepository) GetByOwnerID(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*model.Dumpster, error) {
	var dumpsters []*model.Dumpster
	result := r.db.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dumpsters)
	if result.Error != nil {
		return nil, apperrors.Internal("failed to get dumpsters by owner", result.Error)
	}

	return dumpsters, nil
}

func (r *dumpsterRepository) Update(ctx context.Context, dumpster *model.Dumpster) error {
	result := r.db.WithContext(ctx).Save(dumpster)
	if result.Error != nil {
		return apperrors.Internal("failed to update dumpster", result.Error)
	}

	if result.RowsAffected == 0 {
		return apperrors.NotFound("dumpster not found")
	}

	return nil
}

func (r *dumpsterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Dumpster{}, id)
	if result.Error != nil {
		return apperrors.Internal("failed to delete dumpster", result.Error)
	}

	if result.RowsAffected == 0 {
		return apperrors.NotFound("dumpster not found")
	}

	return nil
}

func (r *dumpsterRepository) List(ctx context.Context, limit, offset int) ([]*model.Dumpster, error) {
	var dumpsters []*model.Dumpster
	result := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dumpsters)
	if result.Error != nil {
		return nil, apperrors.Internal("failed to list dumpsters", result.Error)
	}

	return dumpsters, nil
}

func (r *dumpsterRepository) Search(ctx context.Context, filters DumpsterSearchFilters) ([]*model.Dumpster, error) {
	var dumpsters []*model.Dumpster
	query := r.db.WithContext(ctx)

	if filters.City != "" {
		query = query.Where("city = ?", filters.City)
	}

	if filters.State != "" {
		query = query.Where("state = ?", filters.State)
	}

	if filters.ZipCode != "" {
		query = query.Where("zip_code = ?", filters.ZipCode)
	}

	if filters.MinPrice != nil {
		query = query.Where("price_per_day >= ?", *filters.MinPrice)
	}

	if filters.MaxPrice != nil {
		query = query.Where("price_per_day <= ?", *filters.MaxPrice)
	}

	if filters.Size != "" {
		query = query.Where("size = ?", filters.Size)
	}

	if filters.IsAvailable != nil {
		query = query.Where("is_available = ?", *filters.IsAvailable)
	}

	result := query.
		Order("created_at DESC").
		Limit(filters.Limit).
		Offset(filters.Offset).
		Find(&dumpsters)
	if result.Error != nil {
		return nil, apperrors.Internal("failed to search dumpsters", result.Error)
	}

	return dumpsters, nil
}

func (r *dumpsterRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&model.Dumpster{}).Count(&count)
	if result.Error != nil {
		return 0, apperrors.Internal("failed to count dumpsters", result.Error)
	}

	return count, nil
}

func (r *dumpsterRepository) UpdateAvailability(ctx context.Context, id uuid.UUID, isAvailable bool) error {
	result := r.db.WithContext(ctx).
		Model(&model.Dumpster{}).
		Where("id = ?", id).
		Update("is_available", isAvailable)
	if result.Error != nil {
		return apperrors.Internal("failed to update availability", result.Error)
	}

	if result.RowsAffected == 0 {
		return apperrors.NotFound("dumpster not found")
	}

	return nil
}

func (r *dumpsterRepository) UpdateRating(ctx context.Context, id uuid.UUID, rating float64, reviewCount int) error {
	result := r.db.WithContext(ctx).
		Model(&model.Dumpster{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"rating":       rating,
			"review_count": reviewCount,
		})
	if result.Error != nil {
		return apperrors.Internal("failed to update rating", result.Error)
	}

	if result.RowsAffected == 0 {
		return apperrors.NotFound("dumpster not found")
	}

	return nil
}