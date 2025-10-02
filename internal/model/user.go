package model

import (
	"time"
	"waste-space/internal/dto"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FirstName       string         `gorm:"type:varchar(100);not null" json:"firstName" validate:"required,min=2,max=100"`
	LastName        string         `gorm:"type:varchar(100);not null" json:"lastName" validate:"required,min=2,max=100"`
	Email           string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email" validate:"required,email"`
	PasswordHash    string         `gorm:"type:varchar(255);not null" json:"-"`
	PhoneNumber     string         `gorm:"type:varchar(20);not null" json:"phoneNumber" validate:"required,e164"`
	DateOfBirth     time.Time      `gorm:"type:date;not null" json:"dateOfBirth" validate:"required"`
	Address         string         `gorm:"type:varchar(255);not null" json:"address" validate:"required"`
	City            string         `gorm:"type:varchar(100);not null" json:"city" validate:"required"`
	State           string         `gorm:"type:varchar(50)" json:"state" validate:"omitempty,len=2"`
	ZipCode         string         `gorm:"type:varchar(10);not null" json:"zipCode" validate:"required,numeric"`
	IsEmailVerified bool           `gorm:"default:false;not null" json:"isEmailVerified"`
	IsPhoneVerified bool           `gorm:"default:false;not null" json:"isPhoneVerified"`
	IsActive        bool           `gorm:"default:true;not null" json:"isActive"`
	LastLoginAt     *time.Time     `gorm:"type:timestamp" json:"lastLoginAt,omitempty"`
	CreatedAt       time.Time      `gorm:"autoCreateTime;not null" json:"createdAt"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime;not null" json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"` // Soft delete
}

func NewUserFromDTO(req dto.CreateUserRequest) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		PhoneNumber:  req.PhoneNumber,
		DateOfBirth:  req.DateOfBirth,
		Address:      req.Address,
		City:         req.City,
		State:        req.State,
		ZipCode:      req.ZipCode,
	}, nil
}

func (u *User) ToResponse() dto.UserResponse {
	return dto.UserResponse{
		ID:              u.ID.String(),
		FirstName:       u.FirstName,
		LastName:        u.LastName,
		Email:           u.Email,
		PhoneNumber:     u.PhoneNumber,
		DateOfBirth:     u.DateOfBirth,
		Address:         u.Address,
		City:            u.City,
		State:           u.State,
		ZipCode:         u.ZipCode,
		IsEmailVerified: u.IsEmailVerified,
		IsPhoneVerified: u.IsPhoneVerified,
		IsActive:        u.IsActive,
		LastLoginAt:     u.LastLoginAt,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}
