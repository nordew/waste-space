package repository

import (
	"context"
	"errors"
	"fmt"
	"waste-space/internal/dto"
	"waste-space/internal/model"
	apperrors "waste-space/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	defaultPageSize       = 20
	maxPageSize           = 100
	defaultNearbyDistance = 25.0
	earthRadiusKm         = 6371.0
)

type DumpsterRepository interface {
	Create(ctx context.Context, dumpster *model.Dumpster) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Dumpster, error)
	Update(ctx context.Context, dumpster *model.Dumpster) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req dto.DumpsterListRequest) ([]*model.Dumpster, int64, error)
	Search(ctx context.Context, req dto.DumpsterSearchRequest) ([]*model.Dumpster, int64, error)
	FindNearby(ctx context.Context, req dto.NearbyDumpstersRequest) ([]*model.Dumpster, error)
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
	result := r.db.WithContext(ctx).Preload("Owner").Where("id = ?", id).First(&dumpster)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("dumpster not found")
		}
		return nil, apperrors.Internal("failed to get dumpster", result.Error)
	}
	return &dumpster, nil
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

func (r *dumpsterRepository) List(
	ctx context.Context,
	req dto.DumpsterListRequest) ([]*model.Dumpster, int64, error) {
	var dumpsters []*model.Dumpster
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Dumpster{}).Preload("Owner")

	if req.MaxPrice != nil {
		query = query.Where("price_per_day <= ?", *req.MaxPrice)
	}

	if req.Size != "" {
		query = query.Where("size = ?", req.Size)
	}

	if req.AvailableNow != nil && *req.AvailableNow {
		query = query.Where("is_available = ?", true)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Internal("failed to count dumpsters", err)
	}

	page := max(req.Page, 1)
	limit := max(req.Limit, defaultPageSize)
	if limit > maxPageSize {
		limit = maxPageSize
	}

	offset := (page - 1) * limit

	sortBy := "created_at DESC"
	switch req.SortBy {
	case "price":
		sortBy = "price_per_day ASC"
	case "rating":
		sortBy = "rating DESC"
	case "availability":
		sortBy = "is_available DESC, created_at DESC"
	}

	query = query.Order(sortBy).Limit(limit).Offset(offset)

	if err := query.Find(&dumpsters).Error; err != nil {
		return nil, 0, apperrors.Internal("failed to list dumpsters", err)
	}

	return dumpsters, total, nil
}

func (r *dumpsterRepository) Search(
	ctx context.Context,
	req dto.DumpsterSearchRequest) ([]*model.Dumpster, int64, error) {
	var dumpsters []*model.Dumpster
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Dumpster{}).Preload("Owner")

	if req.Query != "" {
		searchPattern := "%" + req.Query + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ? OR location ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if req.City != "" {
		query = query.Where("city ILIKE ?", "%"+req.City+"%")
	}

	if req.State != "" {
		query = query.Where("state = ?", req.State)
	}

	if req.ZipCode != "" {
		query = query.Where("zip_code = ?", req.ZipCode)
	}

	if req.MinPrice != nil {
		query = query.Where("price_per_day >= ?", *req.MinPrice)
	}

	if req.MaxPrice != nil {
		query = query.Where("price_per_day <= ?", *req.MaxPrice)
	}

	if req.Size != "" {
		query = query.Where("size = ?", req.Size)
	}

	if req.IsAvailable != nil {
		query = query.Where("is_available = ?", *req.IsAvailable)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Internal("failed to count search results", err)
	}

	page := max(req.Page, 1)
	limit := max(req.Limit, defaultPageSize)
	if limit > maxPageSize {
		limit = maxPageSize
	}

	offset := (page - 1) * limit

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&dumpsters).Error; err != nil {
		return nil, 0, apperrors.Internal("failed to search dumpsters", err)
	}

	return dumpsters, total, nil
}

func (r *dumpsterRepository) FindNearby(
	ctx context.Context,
	req dto.NearbyDumpstersRequest) ([]*model.Dumpster, error) {
	var dumpsters []*model.Dumpster

	maxDistance := defaultNearbyDistance
	if req.MaxDistance != nil {
		maxDistance = *req.MaxDistance
	}

	limit := max(req.Limit, defaultPageSize)

	query := fmt.Sprintf(`
		SELECT * FROM (
			SELECT *,
			(%f * acos(cos(radians(%f)) * cos(radians(latitude)) *
			cos(radians(longitude) - radians(%f)) +
			sin(radians(%f)) * sin(radians(latitude)))) AS distance
			FROM dumpsters
			WHERE deleted_at IS NULL
		) AS dumpsters_with_distance
		WHERE distance < %f
		ORDER BY distance
		LIMIT %d
	`, earthRadiusKm,
		req.Latitude,
		req.Longitude,
		req.Latitude,
		maxDistance,
		limit)

	if err := r.db.WithContext(ctx).
		Preload("Owner").
		Raw(query).
		Scan(&dumpsters).Error; err != nil {
		return nil, apperrors.Internal("failed to find nearby dumpsters", err)
	}

	return dumpsters, nil
}
