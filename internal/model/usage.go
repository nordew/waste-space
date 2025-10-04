package model

import (
	"time"
	"waste-space/internal/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DumpsterUsage struct {
	ID              uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DumpsterID      uuid.UUID       `gorm:"type:uuid;not null;index" json:"dumpsterId" validate:"required"`
	Dumpster        *Dumpster       `gorm:"foreignKey:DumpsterID" json:"dumpster,omitempty"`
	UserID          uuid.UUID       `gorm:"type:uuid;not null;index" json:"userId" validate:"required"`
	User            *User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	StartTime       time.Time       `gorm:"not null;index" json:"startTime" validate:"required"`
	EndTime         *time.Time      `json:"endTime"`
	DurationMinutes *int            `json:"durationMinutes"`
	TotalCost       *float64        `gorm:"type:decimal(10,2)" json:"totalCost"`
	Status          UsageStatus     `gorm:"type:varchar(20);not null;default:'active';index" json:"status" validate:"required,oneof=active completed cancelled"`
	Notes           string          `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time       `gorm:"autoCreateTime;not null" json:"createdAt"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime;not null" json:"updatedAt"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"-"`
}

type UsageStatus string

const (
	UsageStatusActive    UsageStatus = "active"
	UsageStatusCompleted UsageStatus = "completed"
	UsageStatusCancelled UsageStatus = "cancelled"
)

func NewDumpsterUsageFromDTO(
	userID, dumpsterID uuid.UUID,
	req dto.StartUsageRequest) *DumpsterUsage {
	return &DumpsterUsage{
		UserID:     userID,
		DumpsterID: dumpsterID,
		StartTime:  req.StartTime,
		Status:     UsageStatusActive,
		Notes:      req.Notes,
	}
}

func (u *DumpsterUsage) ToResponse() dto.UsageResponse {
	resp := dto.UsageResponse{
		ID:         u.ID.String(),
		DumpsterID: u.DumpsterID.String(),
		UserID:     u.UserID.String(),
		StartTime:  u.StartTime,
		Status:     string(u.Status),
		Notes:      u.Notes,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}

	if u.EndTime != nil {
		resp.EndTime = u.EndTime
	}

	if u.DurationMinutes != nil {
		resp.DurationMinutes = u.DurationMinutes
	}

	if u.TotalCost != nil {
		resp.TotalCost = u.TotalCost
	}

	if u.User != nil {
		userResp := u.User.ToResponse()
		resp.User = &userResp
	}

	if u.Dumpster != nil {
		dumpsterResp := u.Dumpster.ToResponse()
		resp.Dumpster = &dumpsterResp
	}

	return resp
}
