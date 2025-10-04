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

type UsageRepository interface {
	Create(ctx context.Context, usage *model.DumpsterUsage) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.DumpsterUsage, error)
	Update(ctx context.Context, usage *model.DumpsterUsage) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByDumpsterID(ctx context.Context, dumpsterID uuid.UUID, req dto.UsageListRequest) ([]*model.DumpsterUsage, int64, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, req dto.UsageListRequest) ([]*model.DumpsterUsage, int64, error)
	GetActiveUsageByUserAndDumpster(ctx context.Context, userID, dumpsterID uuid.UUID) (*model.DumpsterUsage, error)
	GetStats(ctx context.Context, dumpsterID *uuid.UUID, userID *uuid.UUID) (*dto.UsageStatsResponse, error)
	List(ctx context.Context, req dto.UsageListRequest) ([]*model.DumpsterUsage, int64, error)
}

type usageRepository struct {
	db *gorm.DB
}

func NewUsageRepository(db *gorm.DB) UsageRepository {
	return &usageRepository{db: db}
}

func (r *usageRepository) Create(ctx context.Context, usage *model.DumpsterUsage) error {
	result := r.db.WithContext(ctx).Create(usage)
	if result.Error != nil {
		return apperrors.Internal("failed to create usage", result.Error)
	}
	return nil
}

func (r *usageRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.DumpsterUsage, error) {
	var usage model.DumpsterUsage
	result := r.db.WithContext(ctx).Preload("User").Preload("Dumpster").Where("id = ?", id).First(&usage)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("usage not found")
		}
		return nil, apperrors.Internal("failed to get usage", result.Error)
	}
	return &usage, nil
}

func (r *usageRepository) Update(ctx context.Context, usage *model.DumpsterUsage) error {
	result := r.db.WithContext(ctx).Save(usage)
	if result.Error != nil {
		return apperrors.Internal("failed to update usage", result.Error)
	}

	if result.RowsAffected == 0 {
		return apperrors.NotFound("usage not found")
	}

	return nil
}

func (r *usageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.DumpsterUsage{}, id)
	if result.Error != nil {
		return apperrors.Internal("failed to delete usage", result.Error)
	}

	if result.RowsAffected == 0 {
		return apperrors.NotFound("usage not found")
	}

	return nil
}

func (r *usageRepository) GetByDumpsterID(
	ctx context.Context,
	dumpsterID uuid.UUID,
	req dto.UsageListRequest) ([]*model.DumpsterUsage, int64, error) {
	var usages []*model.DumpsterUsage
	var total int64

	query := r.db.WithContext(ctx).Model(&model.DumpsterUsage{}).Preload("User").Where("dumpster_id = ?", dumpsterID)

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Internal("failed to count usages", err)
	}

	page := max(req.Page, 1)
	limit := max(req.Limit, defaultPageSize)
	if limit > maxPageSize {
		limit = maxPageSize
	}

	offset := (page - 1) * limit

	if err := query.Order("start_time DESC").Limit(limit).Offset(offset).Find(&usages).Error; err != nil {
		return nil, 0, apperrors.Internal("failed to get usages", err)
	}

	return usages, total, nil
}

func (r *usageRepository) GetByUserID(
	ctx context.Context,
	userID uuid.UUID,
	req dto.UsageListRequest) ([]*model.DumpsterUsage, int64, error) {
	var usages []*model.DumpsterUsage
	var total int64

	query := r.db.WithContext(ctx).Model(&model.DumpsterUsage{}).Preload("Dumpster").Where("user_id = ?", userID)

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Internal("failed to count usages", err)
	}

	page := max(req.Page, 1)
	limit := max(req.Limit, defaultPageSize)
	if limit > maxPageSize {
		limit = maxPageSize
	}

	offset := (page - 1) * limit

	if err := query.Order("start_time DESC").Limit(limit).Offset(offset).Find(&usages).Error; err != nil {
		return nil, 0, apperrors.Internal("failed to get usages", err)
	}

	return usages, total, nil
}

func (r *usageRepository) GetActiveUsageByUserAndDumpster(
	ctx context.Context,
	userID, dumpsterID uuid.UUID) (*model.DumpsterUsage, error) {
	var usage model.DumpsterUsage
	result := r.db.WithContext(ctx).Where("user_id = ? AND dumpster_id = ? AND status = ?", userID, dumpsterID, model.UsageStatusActive).First(&usage)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperrors.Internal("failed to get usage", result.Error)
	}
	return &usage, nil
}

func (r *usageRepository) GetStats(
	ctx context.Context,
	dumpsterID *uuid.UUID,
	userID *uuid.UUID) (*dto.UsageStatsResponse, error) {
	var stats dto.UsageStatsResponse

	query := r.db.WithContext(ctx).Model(&model.DumpsterUsage{})

	if dumpsterID != nil {
		query = query.Where("dumpster_id = ?", *dumpsterID)
	}

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	if err := query.Count(&stats.TotalUsages).Error; err != nil {
		return nil, apperrors.Internal("failed to count total usages", err)
	}

	if err := query.Where("status = ?", model.UsageStatusActive).Count(&stats.ActiveUsages).Error; err != nil {
		return nil, apperrors.Internal("failed to count active usages", err)
	}

	if err := query.Where("status = ?", model.UsageStatusCompleted).Count(&stats.CompletedUsages).Error; err != nil {
		return nil, apperrors.Internal("failed to count completed usages", err)
	}

	var totalMinutes *int64
	if err := query.Select("COALESCE(SUM(duration_minutes), 0)").Scan(&totalMinutes).Error; err != nil {
		return nil, apperrors.Internal("failed to calculate total minutes", err)
	}
	if totalMinutes != nil {
		stats.TotalMinutes = *totalMinutes
	}

	var totalRevenue *float64
	if err := query.Select("COALESCE(SUM(total_cost), 0)").Scan(&totalRevenue).Error; err != nil {
		return nil, apperrors.Internal("failed to calculate total revenue", err)
	}
	if totalRevenue != nil {
		stats.TotalRevenue = *totalRevenue
	}

	return &stats, nil
}

func (r *usageRepository) List(
	ctx context.Context,
	req dto.UsageListRequest) ([]*model.DumpsterUsage, int64, error) {
	var usages []*model.DumpsterUsage
	var total int64

	query := r.db.WithContext(ctx).Model(&model.DumpsterUsage{}).Preload("User").Preload("Dumpster")

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.DumpsterID != "" {
		dumpsterID, err := uuid.Parse(req.DumpsterID)
		if err != nil {
			return nil, 0, apperrors.BadRequest("invalid dumpster id")
		}
		query = query.Where("dumpster_id = ?", dumpsterID)
	}

	if req.UserID != "" {
		userID, err := uuid.Parse(req.UserID)
		if err != nil {
			return nil, 0, apperrors.BadRequest("invalid user id")
		}
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Internal("failed to count usages", err)
	}

	page := max(req.Page, 1)
	limit := max(req.Limit, defaultPageSize)
	if limit > maxPageSize {
		limit = maxPageSize
	}

	offset := (page - 1) * limit

	if err := query.Order("start_time DESC").Limit(limit).Offset(offset).Find(&usages).Error; err != nil {
		return nil, 0, apperrors.Internal("failed to get usages", err)
	}

	return usages, total, nil
}
