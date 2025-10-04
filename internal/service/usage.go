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

type UsageService interface {
	StartUsage(ctx context.Context, userID, dumpsterID string, req dto.StartUsageRequest) (*dto.UsageResponse, error)
	EndUsage(ctx context.Context, userID, id string, req dto.EndUsageRequest) (*dto.UsageResponse, error)
	GetByID(ctx context.Context, id string) (*dto.UsageResponse, error)
	GetByDumpsterID(ctx context.Context, dumpsterID string, req dto.UsageListRequest) (*dto.UsageListResponse, error)
	GetByUserID(ctx context.Context, userID string, req dto.UsageListRequest) (*dto.UsageListResponse, error)
	GetStats(ctx context.Context, dumpsterID, userID *string) (*dto.UsageStatsResponse, error)
	List(ctx context.Context, req dto.UsageListRequest) (*dto.UsageListResponse, error)
	Delete(ctx context.Context, id string) error
}

type usageService struct {
	usageRepo    repository.UsageRepository
	dumpsterRepo repository.DumpsterRepository
	logger       *zap.Logger
}

func NewUsageService(
	usageRepo repository.UsageRepository,
	dumpsterRepo repository.DumpsterRepository,
	logger *zap.Logger) UsageService {
	return &usageService{
		usageRepo:    usageRepo,
		dumpsterRepo: dumpsterRepo,
		logger:       logger,
	}
}

func (s *usageService) StartUsage(
	ctx context.Context,
	userID, dumpsterID string,
	req dto.StartUsageRequest) (*dto.UsageResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.BadRequest("invalid user ID")
	}

	dumpsterUUID, err := uuid.Parse(dumpsterID)
	if err != nil {
		return nil, apperrors.BadRequest("invalid dumpster ID")
	}

	dumpster, err := s.dumpsterRepo.GetByID(ctx, dumpsterUUID)
	if err != nil {
		return nil, err
	}

	if !dumpster.IsAvailable {
		return nil, apperrors.BadRequest("dumpster is not available")
	}

	activeUsage, err := s.usageRepo.GetActiveUsageByUserAndDumpster(ctx, userUUID, dumpsterUUID)
	if err != nil {
		s.logger.Error("failed to check active usage", zap.String("userId", userID), zap.String("dumpsterId", dumpsterID), zap.Error(err))
		return nil, err
	}
	if activeUsage != nil {
		return nil, apperrors.BadRequest("you already have an active usage session for this dumpster")
	}

	usage := model.NewDumpsterUsageFromDTO(userUUID, dumpsterUUID, req)

	if err := s.usageRepo.Create(ctx, usage); err != nil {
		s.logger.Error("failed to create usage", zap.String("userId", userID), zap.String("dumpsterId", dumpsterID), zap.Error(err))
		return nil, err
	}

	response := usage.ToResponse()
	return &response, nil
}

func (s *usageService) EndUsage(
	ctx context.Context,
	userID, id string,
	req dto.EndUsageRequest) (*dto.UsageResponse, error) {
	usageID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.BadRequest("invalid usage ID")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.BadRequest("invalid user ID")
	}

	usage, err := s.usageRepo.GetByID(ctx, usageID)
	if err != nil {
		return nil, err
	}

	if usage.UserID != userUUID {
		return nil, apperrors.Forbidden("you don't have permission to end this usage session")
	}

	if usage.Status != model.UsageStatusActive {
		return nil, apperrors.BadRequest("usage session is not active")
	}

	if req.EndTime.Before(usage.StartTime) {
		return nil, apperrors.BadRequest("end time must be after start time")
	}

	usage.EndTime = &req.EndTime
	duration := int(req.EndTime.Sub(usage.StartTime).Minutes())
	usage.DurationMinutes = &duration

	dumpster, err := s.dumpsterRepo.GetByID(ctx, usage.DumpsterID)
	if err != nil {
		s.logger.Error("failed to get dumpster for cost calculation", zap.String("dumpsterId", usage.DumpsterID.String()), zap.Error(err))
		return nil, err
	}

	totalCost := s.calculateCost(dumpster.PricePerDay, duration)
	usage.TotalCost = &totalCost
	usage.Status = model.UsageStatusCompleted

	if req.Notes != "" {
		usage.Notes = req.Notes
	}

	if err := s.usageRepo.Update(ctx, usage); err != nil {
		s.logger.Error("failed to update usage", zap.String("usageId", id), zap.Error(err))
		return nil, err
	}

	response := usage.ToResponse()
	return &response, nil
}

func (s *usageService) GetByID(ctx context.Context, id string) (*dto.UsageResponse, error) {
	usageID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.BadRequest("invalid usage ID")
	}

	usage, err := s.usageRepo.GetByID(ctx, usageID)
	if err != nil {
		return nil, err
	}

	response := usage.ToResponse()
	return &response, nil
}

func (s *usageService) GetByDumpsterID(
	ctx context.Context,
	dumpsterID string,
	req dto.UsageListRequest) (*dto.UsageListResponse, error) {
	dumpsterUUID, err := uuid.Parse(dumpsterID)
	if err != nil {
		return nil, apperrors.BadRequest("invalid dumpster ID")
	}

	usages, total, err := s.usageRepo.GetByDumpsterID(ctx, dumpsterUUID, req)
	if err != nil {
		s.logger.Error("failed to get usages by dumpster", zap.String("dumpsterId", dumpsterID), zap.Error(err))
		return nil, err
	}

	return s.buildUsageListResponse(usages, total, req.Page, req.Limit), nil
}

func (s *usageService) GetByUserID(
	ctx context.Context,
	userID string,
	req dto.UsageListRequest) (*dto.UsageListResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.BadRequest("invalid user ID")
	}

	usages, total, err := s.usageRepo.GetByUserID(ctx, userUUID, req)
	if err != nil {
		s.logger.Error("failed to get usages by user", zap.String("userId", userID), zap.Error(err))
		return nil, err
	}

	return s.buildUsageListResponse(usages, total, req.Page, req.Limit), nil
}

func (s *usageService) GetStats(
	ctx context.Context,
	dumpsterID, userID *string) (*dto.UsageStatsResponse, error) {
	var dumpsterUUID *uuid.UUID
	var userUUID *uuid.UUID

	if dumpsterID != nil {
		parsed, err := uuid.Parse(*dumpsterID)
		if err != nil {
			return nil, apperrors.BadRequest("invalid dumpster ID")
		}
		dumpsterUUID = &parsed
	}

	if userID != nil {
		parsed, err := uuid.Parse(*userID)
		if err != nil {
			return nil, apperrors.BadRequest("invalid user ID")
		}
		userUUID = &parsed
	}

	stats, err := s.usageRepo.GetStats(ctx, dumpsterUUID, userUUID)
	if err != nil {
		s.logger.Error("failed to get usage stats", zap.Error(err))
		return nil, err
	}

	return stats, nil
}

func (s *usageService) List(
	ctx context.Context,
	req dto.UsageListRequest) (*dto.UsageListResponse, error) {
	usages, total, err := s.usageRepo.List(ctx, req)
	if err != nil {
		s.logger.Error("failed to list usages", zap.Error(err))
		return nil, err
	}

	return s.buildUsageListResponse(usages, total, req.Page, req.Limit), nil
}

func (s *usageService) Delete(ctx context.Context, id string) error {
	usageID, err := uuid.Parse(id)
	if err != nil {
		return apperrors.BadRequest("invalid usage ID")
	}

	if err := s.usageRepo.Delete(ctx, usageID); err != nil {
		s.logger.Error("failed to delete usage", zap.String("usageId", id), zap.Error(err))
		return err
	}

	return nil
}

func (s *usageService) calculateCost(pricePerDay float64, durationMinutes int) float64 {
	minutesPerDay := 24.0 * 60.0
	return (pricePerDay / minutesPerDay) * float64(durationMinutes)
}

func (s *usageService) buildUsageListResponse(
	usages []*model.DumpsterUsage,
	total int64,
	page, limit int) *dto.UsageListResponse {
	page = max(page, 1)
	limit = max(limit, 1)

	responses := make([]dto.UsageResponse, len(usages))
	for i, usage := range usages {
		responses[i] = usage.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &dto.UsageListResponse{
		Usages:     responses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}
