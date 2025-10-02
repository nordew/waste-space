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

type DumpsterListRequest struct {
	Page          int      `form:"page" validate:"omitempty,min=1"`
	Limit         int      `form:"limit" validate:"omitempty,min=1,max=100"`
	SortBy        string   `form:"sortBy" validate:"omitempty,oneof=price distance rating availability"`
	Location      string   `form:"location"`
	MaxPrice      *float64 `form:"maxPrice" validate:"omitempty,gt=0"`
	Size          string   `form:"size" validate:"omitempty,oneof=small medium large extraLarge"`
	AvailableNow  *bool    `form:"availableNow"`
	MaxDistance   *float64 `form:"maxDistance" validate:"omitempty,gt=0"`
}

type DumpsterSearchRequest struct {
	Query       string   `form:"q"`
	City        string   `form:"city"`
	State       string   `form:"state"`
	ZipCode     string   `form:"zipCode"`
	MinPrice    *float64 `form:"minPrice" validate:"omitempty,gte=0"`
	MaxPrice    *float64 `form:"maxPrice" validate:"omitempty,gte=0"`
	Size        string   `form:"size" validate:"omitempty,oneof=small medium large extraLarge"`
	IsAvailable *bool    `form:"isAvailable"`
	Page        int      `form:"page" validate:"omitempty,min=1"`
	Limit       int      `form:"limit" validate:"omitempty,min=1,max=100"`
}

type NearbyDumpstersRequest struct {
	Latitude    float64  `form:"lat" validate:"required,latitude"`
	Longitude   float64  `form:"lng" validate:"required,longitude"`
	MaxDistance *float64 `form:"maxDistance" validate:"omitempty,gt=0"`
	Limit       int      `form:"limit" validate:"omitempty,min=1,max=100"`
}

type BookDumpsterRequest struct {
	StartDate time.Time `json:"startDate" validate:"required"`
	EndDate   time.Time `json:"endDate" validate:"required,gtfield=StartDate"`
}

type BookingResponse struct {
	ID          string    `json:"id"`
	DumpsterID  string    `json:"dumpsterId"`
	UserID      string    `json:"userId"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	TotalPrice  float64   `json:"totalPrice"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

type AvailabilityResponse struct {
	DumpsterID  string `json:"dumpsterId"`
	IsAvailable bool   `json:"isAvailable"`
	Message     string `json:"message,omitempty"`
}

type DumpsterListResponse struct {
	Dumpsters  []DumpsterResponse `json:"dumpsters"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"totalPages"`
}
