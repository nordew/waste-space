package model

import (
	"time"
	"waste-space/internal/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Review struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DumpsterID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"dumpsterId" validate:"required"`
	Dumpster    *Dumpster      `gorm:"foreignKey:DumpsterID" json:"dumpster,omitempty"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"userId" validate:"required"`
	User        *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Rating      int            `gorm:"not null" json:"rating" validate:"required,min=1,max=5"`
	Comment     string         `gorm:"type:text" json:"comment"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;not null" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime;not null" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func NewReviewFromDTO(userID, dumpsterID uuid.UUID, req dto.CreateReviewRequest) *Review {
	return &Review{
		UserID:     userID,
		DumpsterID: dumpsterID,
		Rating:     req.Rating,
		Comment:    req.Comment,
	}
}

func (r *Review) ToResponse() dto.ReviewResponse {
	resp := dto.ReviewResponse{
		ID:         r.ID.String(),
		DumpsterID: r.DumpsterID.String(),
		UserID:     r.UserID.String(),
		Rating:     r.Rating,
		Comment:    r.Comment,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}

	if r.User != nil {
		userResp := r.User.ToResponse()
		resp.User = &userResp
	}

	return resp
}
