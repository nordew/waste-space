package dto

import "time"

type CreateUserRequest struct {
	FirstName   string    `json:"firstName" validate:"required,min=2,max=100"`
	LastName    string    `json:"lastName" validate:"required,min=2,max=100"`
	Email       string    `json:"email" validate:"required,email"`
	Password    string    `json:"password" validate:"required,min=8,max=72"`
	PhoneNumber string    `json:"phoneNumber" validate:"required,e164"`
	DateOfBirth time.Time `json:"dateOfBirth" validate:"required"`
	Address     string    `json:"address" validate:"required"`
	City        string    `json:"city" validate:"required"`
	State       string    `json:"state" validate:"omitempty,len=2"`
	ZipCode     string    `json:"zipCode" validate:"required,numeric"`
}

type UpdateUserRequest struct {
	FirstName   *string    `json:"firstName,omitempty" validate:"omitempty,min=2,max=100"`
	LastName    *string    `json:"lastName,omitempty" validate:"omitempty,min=2,max=100"`
	PhoneNumber *string    `json:"phoneNumber,omitempty" validate:"omitempty,e164"`
	DateOfBirth *time.Time `json:"dateOfBirth,omitempty"`
	Address     *string    `json:"address,omitempty"`
	City        *string    `json:"city,omitempty"`
	State       *string    `json:"state,omitempty" validate:"omitempty,len=2"`
	ZipCode     *string    `json:"zipCode,omitempty" validate:"omitempty,numeric"`
}

type UserResponse struct {
	ID              string     `json:"id"`
	FirstName       string     `json:"firstName"`
	LastName        string     `json:"lastName"`
	Email           string     `json:"email"`
	PhoneNumber     string     `json:"phoneNumber"`
	DateOfBirth     time.Time  `json:"dateOfBirth"`
	Address         string     `json:"address"`
	City            string     `json:"city"`
	State           string     `json:"state"`
	ZipCode         string     `json:"zipCode"`
	IsEmailVerified bool       `json:"isEmailVerified"`
	IsPhoneVerified bool       `json:"isPhoneVerified"`
	IsActive        bool       `json:"isActive"`
	LastLoginAt     *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}
