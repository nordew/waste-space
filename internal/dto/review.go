package dto

import "time"

type CreateReviewRequest struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment" validate:"omitempty,max=1000"`
}

type UpdateReviewRequest struct {
	Rating  *int    `json:"rating,omitempty" validate:"omitempty,min=1,max=5"`
	Comment *string `json:"comment,omitempty" validate:"omitempty,max=1000"`
}

type ReviewResponse struct {
	ID         string        `json:"id"`
	DumpsterID string        `json:"dumpsterId"`
	UserID     string        `json:"userId"`
	User       *UserResponse `json:"user,omitempty"`
	Rating     int           `json:"rating"`
	Comment    string        `json:"comment"`
	CreatedAt  time.Time     `json:"createdAt"`
	UpdatedAt  time.Time     `json:"updatedAt"`
}

type ReviewListRequest struct {
	Page  int `form:"page" validate:"omitempty,min=1"`
	Limit int `form:"limit" validate:"omitempty,min=1,max=100"`
}

type ReviewListResponse struct {
	Reviews    []ReviewResponse `json:"reviews"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"totalPages"`
}
