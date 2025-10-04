package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"waste-space/internal/dto"
	"waste-space/internal/model"
	"waste-space/internal/storage/repository"
	apperrors "waste-space/pkg/errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DumpsterService interface {
	Create(ctx context.Context, ownerID string, req dto.CreateDumpsterRequest) (*dto.DumpsterResponse, error)
	GetByID(ctx context.Context, id string) (*dto.DumpsterResponse, error)
	Update(ctx context.Context, ownerID, id string, req dto.UpdateDumpsterRequest) (*dto.DumpsterResponse, error)
	Delete(ctx context.Context, ownerID, id string) error
	List(ctx context.Context, req dto.DumpsterListRequest) (*dto.DumpsterListResponse, error)
	Search(ctx context.Context, req dto.DumpsterSearchRequest) (*dto.DumpsterListResponse, error)
	FindNearby(ctx context.Context, req dto.NearbyDumpstersRequest) ([]dto.DumpsterResponse, error)
	CheckAvailability(ctx context.Context, id string) (*dto.AvailabilityResponse, error)
	BookDumpster(ctx context.Context, userID, dumpsterID string, req dto.BookDumpsterRequest) (*dto.BookingResponse, error)
}

type dumpsterService struct {
	dumpsterRepo repository.DumpsterRepository
	logger       *zap.Logger
}

func NewDumpsterService(
	dumpsterRepo repository.DumpsterRepository,
	logger *zap.Logger) DumpsterService {
	return &dumpsterService{
		dumpsterRepo: dumpsterRepo,
		logger:       logger,
	}
}

func (s *dumpsterService) Create(
	ctx context.Context,
	ownerID string,
	req dto.CreateDumpsterRequest) (*dto.DumpsterResponse, error) {
	ownerUUID, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, apperrors.BadRequest("invalid owner ID")
	}

	dumpster := model.NewDumpsterFromDTO(ownerUUID, req)

	if err := s.dumpsterRepo.Create(ctx, dumpster); err != nil {
		s.logger.Error("failed to create dumpster", zap.String("ownerId", ownerID), zap.Error(err))
		return nil, err
	}

	response := dumpster.ToResponse()
	return &response, nil
}

func (s *dumpsterService) GetByID(ctx context.Context, id string) (*dto.DumpsterResponse, error) {
	dumpsterID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.BadRequest("invalid dumpster ID")
	}

	dumpster, err := s.dumpsterRepo.GetByID(ctx, dumpsterID)
	if err != nil {
		return nil, err
	}

	response := dumpster.ToResponse()
	return &response, nil
}

func (s *dumpsterService) Update(
	ctx context.Context,
	ownerID, id string,
	req dto.UpdateDumpsterRequest) (*dto.DumpsterResponse, error) {
	dumpsterID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.BadRequest("invalid dumpster ID")
	}

	ownerUUID, err := uuid.Parse(ownerID)
	if err != nil {
		return nil, apperrors.BadRequest("invalid owner ID")
	}

	dumpster, err := s.dumpsterRepo.GetByID(ctx, dumpsterID)
	if err != nil {
		return nil, err
	}

	if dumpster.OwnerID != ownerUUID {
		return nil, apperrors.Forbidden("you don't have permission to update this dumpster")
	}

	s.applyDumpsterUpdates(dumpster, req)

	if err := s.dumpsterRepo.Update(ctx, dumpster); err != nil {
		s.logger.Error("failed to update dumpster", zap.String("dumpsterId", id), zap.Error(err))
		return nil, err
	}

	response := dumpster.ToResponse()
	return &response, nil
}

func (s *dumpsterService) Delete(ctx context.Context, ownerID, id string) error {
	dumpsterID, err := uuid.Parse(id)
	if err != nil {
		return apperrors.BadRequest("invalid dumpster ID")
	}

	ownerUUID, err := uuid.Parse(ownerID)
	if err != nil {
		return apperrors.BadRequest("invalid owner ID")
	}

	dumpster, err := s.dumpsterRepo.GetByID(ctx, dumpsterID)
	if err != nil {
		return err
	}

	if dumpster.OwnerID != ownerUUID {
		return apperrors.Forbidden("you don't have permission to delete this dumpster")
	}

	return s.dumpsterRepo.Delete(ctx, dumpsterID)
}

func (s *dumpsterService) List(ctx context.Context, req dto.DumpsterListRequest) (*dto.DumpsterListResponse, error) {
	if req.Location != "" {
		coords := s.parseLocation(req.Location)
		if len(coords) == 2 {
			nearbyReq := dto.NearbyDumpstersRequest{
				Latitude:    coords[0],
				Longitude:   coords[1],
				MaxDistance: req.MaxDistance,
				Limit:       req.Limit,
			}
			dumpsters, err := s.dumpsterRepo.FindNearby(ctx, nearbyReq)
			if err != nil {
				s.logger.Error("failed to find nearby dumpsters", zap.Error(err))
				return nil, err
			}
			return s.buildDumpsterListResponse(dumpsters, int64(len(dumpsters)), req.Page, req.Limit), nil
		}
	}

	dumpsters, total, err := s.dumpsterRepo.List(ctx, req)
	if err != nil {
		s.logger.Error("failed to list dumpsters", zap.Error(err))
		return nil, err
	}

	return s.buildDumpsterListResponse(dumpsters, total, req.Page, req.Limit), nil
}

func (s *dumpsterService) Search(ctx context.Context, req dto.DumpsterSearchRequest) (*dto.DumpsterListResponse, error) {
	dumpsters, total, err := s.dumpsterRepo.Search(ctx, req)
	if err != nil {
		s.logger.Error("failed to search dumpsters", zap.Error(err))
		return nil, err
	}

	return s.buildDumpsterListResponse(dumpsters, total, req.Page, req.Limit), nil
}

func (s *dumpsterService) FindNearby(ctx context.Context, req dto.NearbyDumpstersRequest) ([]dto.DumpsterResponse, error) {
	dumpsters, err := s.dumpsterRepo.FindNearby(ctx, req)
	if err != nil {
		s.logger.Error("failed to find nearby dumpsters", zap.Error(err))
		return nil, err
	}

	responses := make([]dto.DumpsterResponse, len(dumpsters))
	for i, dumpster := range dumpsters {
		responses[i] = dumpster.ToResponse()
	}

	return responses, nil
}

func (s *dumpsterService) CheckAvailability(ctx context.Context, id string) (*dto.AvailabilityResponse, error) {
	dumpsterID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.BadRequest("invalid dumpster ID")
	}

	dumpster, err := s.dumpsterRepo.GetByID(ctx, dumpsterID)
	if err != nil {
		return nil, err
	}

	message := ""
	if !dumpster.IsAvailable {
		message = "Dumpster is currently unavailable"
	}

	return &dto.AvailabilityResponse{
		DumpsterID:  id,
		IsAvailable: dumpster.IsAvailable,
		Message:     message,
	}, nil
}

func (s *dumpsterService) BookDumpster(
	ctx context.Context,
	userID, dumpsterID string,
	req dto.BookDumpsterRequest) (*dto.BookingResponse, error) {
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

	days := req.EndDate.Sub(req.StartDate).Hours() / 24
	if days <= 0 {
		return nil, apperrors.BadRequest("end date must be after start date")
	}

	totalPrice := dumpster.PricePerDay * days

	return &dto.BookingResponse{
		ID:         uuid.New().String(),
		DumpsterID: dumpsterID,
		UserID:     userID,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		TotalPrice: totalPrice,
		Status:     "pending",
		CreatedAt:  req.StartDate,
	}, nil
}

func (s *dumpsterService) applyDumpsterUpdates(dumpster *model.Dumpster, req dto.UpdateDumpsterRequest) {
	if req.Title != nil {
		dumpster.Title = *req.Title
	}
	if req.Description != nil {
		dumpster.Description = *req.Description
	}
	if req.Location != nil {
		dumpster.Location = *req.Location
	}
	if req.Latitude != nil {
		dumpster.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		dumpster.Longitude = *req.Longitude
	}
	if req.Address != nil {
		dumpster.Address = *req.Address
	}
	if req.City != nil {
		dumpster.City = *req.City
	}
	if req.State != nil {
		dumpster.State = *req.State
	}
	if req.ZipCode != nil {
		dumpster.ZipCode = *req.ZipCode
	}
	if req.PricePerDay != nil {
		dumpster.PricePerDay = *req.PricePerDay
	}
	if req.Size != nil {
		dumpster.Size = model.DumpsterSize(*req.Size)
	}
	if req.IsAvailable != nil {
		dumpster.IsAvailable = *req.IsAvailable
	}
	if req.Capacity != nil {
		dumpster.Capacity = *req.Capacity
	}
	if req.Weight != nil {
		dumpster.Weight = *req.Weight
	}
}

func (s *dumpsterService) parseLocation(location string) []float64 {
	parts := strings.Split(location, ",")
	if len(parts) != 2 {
		return nil
	}

	var lat, lng float64
	if _, err := fmt.Sscanf(strings.TrimSpace(parts[0]), "%f", &lat); err != nil {
		return nil
	}
	if _, err := fmt.Sscanf(strings.TrimSpace(parts[1]), "%f", &lng); err != nil {
		return nil
	}

	return []float64{lat, lng}
}

func (s *dumpsterService) buildDumpsterListResponse(
	dumpsters []*model.Dumpster,
	total int64,
	page, limit int) *dto.DumpsterListResponse {
	page = max(page, 1)
	limit = max(limit, 1)

	responses := make([]dto.DumpsterResponse, len(dumpsters))
	for i, dumpster := range dumpsters {
		responses[i] = dumpster.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &dto.DumpsterListResponse{
		Dumpsters:  responses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}
