package dto

import (
	"time"
)

type StartUsageRequest struct {
	StartTime time.Time `json:"startTime" validate:"required"`
	Notes     string    `json:"notes"`
}

type EndUsageRequest struct {
	EndTime time.Time `json:"endTime" validate:"required"`
	Notes   string    `json:"notes"`
}

type UsageResponse struct {
	ID              string           `json:"id"`
	DumpsterID      string           `json:"dumpsterId"`
	Dumpster        *DumpsterResponse `json:"dumpster,omitempty"`
	UserID          string           `json:"userId"`
	User            *UserResponse    `json:"user,omitempty"`
	StartTime       time.Time        `json:"startTime"`
	EndTime         *time.Time       `json:"endTime,omitempty"`
	DurationMinutes *int             `json:"durationMinutes,omitempty"`
	TotalCost       *float64         `json:"totalCost,omitempty"`
	Status          string           `json:"status"`
	Notes           string           `json:"notes"`
	CreatedAt       time.Time        `json:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt"`
}

type UsageListResponse struct {
	Usages     []UsageResponse `json:"usages"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"totalPages"`
}

type UsageStatsResponse struct {
	TotalUsages     int64   `json:"totalUsages"`
	ActiveUsages    int64   `json:"activeUsages"`
	CompletedUsages int64   `json:"completedUsages"`
	TotalMinutes    int64   `json:"totalMinutes"`
	TotalRevenue    float64 `json:"totalRevenue"`
}

type UsageListRequest struct {
	Page       int    `form:"page" validate:"omitempty,min=1"`
	Limit      int    `form:"limit" validate:"omitempty,min=1,max=100"`
	Status     string `form:"status" validate:"omitempty,oneof=active completed cancelled"`
	DumpsterID string `form:"dumpsterId"`
	UserID     string `form:"userId"`
}
