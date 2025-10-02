package dto

import (
	"time"
)

type CreateDumpsterRequest struct {
	Title       string  `json:"title" validate:"required,min=5,max=255"`
	Description string  `json:"description"`
	Location    string  `json:"location" validate:"required"`
	Latitude    float64 `json:"latitude" validate:"required,latitude"`
	Longitude   float64 `json:"longitude" validate:"required,longitude"`
	Address     string  `json:"address" validate:"required"`
	City        string  `json:"city" validate:"required"`
	State       string  `json:"state" validate:"required"`
	ZipCode     string  `json:"zipCode" validate:"required"`
	PricePerDay float64 `json:"pricePerDay" validate:"required,gt=0"`
	Size        string  `json:"size" validate:"required,oneof=small medium large extraLarge"`
	Capacity    string  `json:"capacity"`
	Weight      string  `json:"weight"`
}

type UpdateDumpsterRequest struct {
	Title       *string  `json:"title,omitempty" validate:"omitempty,min=5,max=255"`
	Description *string  `json:"description,omitempty"`
	Location    *string  `json:"location,omitempty"`
	Latitude    *float64 `json:"latitude,omitempty" validate:"omitempty,latitude"`
	Longitude   *float64 `json:"longitude,omitempty" validate:"omitempty,longitude"`
	Address     *string  `json:"address,omitempty"`
	City        *string  `json:"city,omitempty"`
	State       *string  `json:"state,omitempty"`
	ZipCode     *string  `json:"zipCode,omitempty"`
	PricePerDay *float64 `json:"pricePerDay,omitempty" validate:"omitempty,gt=0"`
	Size        *string  `json:"size,omitempty" validate:"omitempty,oneof=small medium large extraLarge"`
	IsAvailable *bool    `json:"isAvailable,omitempty"`
	Capacity    *string  `json:"capacity,omitempty"`
	Weight      *string  `json:"weight,omitempty"`
}

type DumpsterResponse struct {
	ID          string        `json:"id"`
	OwnerID     string        `json:"ownerId"`
	Owner       *UserResponse `json:"owner,omitempty"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Location    string        `json:"location"`
	Latitude    float64       `json:"latitude"`
	Longitude   float64       `json:"longitude"`
	Address     string        `json:"address"`
	City        string        `json:"city"`
	State       string        `json:"state"`
	ZipCode     string        `json:"zipCode"`
	PricePerDay float64       `json:"pricePerDay"`
	Size        string        `json:"size"`
	IsAvailable bool          `json:"isAvailable"`
	Rating      float64       `json:"rating"`
	ReviewCount int           `json:"reviewCount"`
	Capacity    string        `json:"capacity"`
	Weight      string        `json:"weight"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
}

type DumpsterSearchRequest struct {
	City        string   `json:"city"`
	State       string   `json:"state"`
	ZipCode     string   `json:"zipCode"`
	MinPrice    *float64 `json:"minPrice,omitempty" validate:"omitempty,gte=0"`
	MaxPrice    *float64 `json:"maxPrice,omitempty" validate:"omitempty,gte=0"`
	Size        string   `json:"size" validate:"omitempty,oneof=small medium large extraLarge"`
	IsAvailable *bool    `json:"isAvailable,omitempty"`
	Latitude    *float64 `json:"latitude,omitempty" validate:"omitempty,latitude"`
	Longitude   *float64 `json:"longitude,omitempty" validate:"omitempty,longitude"`
	RadiusKm    *float64 `json:"radiusKm,omitempty" validate:"omitempty,gt=0"`
	Limit       int      `json:"limit" validate:"omitempty,min=1,max=100"`
	Offset      int      `json:"offset" validate:"omitempty,min=0"`
}
