package model

import (
	"time"
	"waste-space/internal/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Dumpster struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OwnerID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"ownerId" validate:"required"`
	Owner       *User          `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title" validate:"required,min=5,max=255"`
	Description string         `gorm:"type:text" json:"description"`
	Location    string         `gorm:"type:varchar(255);not null" json:"location" validate:"required"`
	Latitude    float64        `gorm:"type:decimal(10,8);not null" json:"latitude" validate:"required,latitude"`
	Longitude   float64        `gorm:"type:decimal(11,8);not null" json:"longitude" validate:"required,longitude"`
	Address     string         `gorm:"type:varchar(255);not null" json:"address" validate:"required"`
	City        string         `gorm:"type:varchar(100);not null" json:"city" validate:"required"`
	State       string         `gorm:"type:varchar(50);not null" json:"state" validate:"required"`
	ZipCode     string         `gorm:"type:varchar(10);not null" json:"zipCode" validate:"required"`
	PricePerDay float64        `gorm:"type:decimal(10,2);not null" json:"pricePerDay" validate:"required,gt=0"`
	Size        DumpsterSize   `gorm:"type:varchar(20);not null" json:"size" validate:"required,oneof=small medium large extraLarge"`
	IsAvailable bool           `gorm:"default:true;not null" json:"isAvailable"`
	Rating      float64        `gorm:"type:decimal(3,2);default:0.0" json:"rating" validate:"gte=0,lte=5"`
	ReviewCount int            `gorm:"default:0" json:"reviewCount"`
	Capacity    string         `gorm:"type:varchar(50)" json:"capacity"`
	Weight      string         `gorm:"type:varchar(50)" json:"weight"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;not null" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime;not null" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type DumpsterSize string

const (
	DumpsterSizeSmall      DumpsterSize = "small"
	DumpsterSizeMedium     DumpsterSize = "medium"
	DumpsterSizeLarge      DumpsterSize = "large"
	DumpsterSizeExtraLarge DumpsterSize = "extraLarge"
)

func NewDumpsterFromDTO(ownerID uuid.UUID, req dto.CreateDumpsterRequest) *Dumpster {
	return &Dumpster{
		OwnerID:     ownerID,
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		ZipCode:     req.ZipCode,
		PricePerDay: req.PricePerDay,
		Size:        DumpsterSize(req.Size),
		Capacity:    req.Capacity,
		Weight:      req.Weight,
	}
}

func (d *Dumpster) ToResponse() dto.DumpsterResponse {
	resp := dto.DumpsterResponse{
		ID:          d.ID.String(),
		OwnerID:     d.OwnerID.String(),
		Title:       d.Title,
		Description: d.Description,
		Location:    d.Location,
		Latitude:    d.Latitude,
		Longitude:   d.Longitude,
		Address:     d.Address,
		City:        d.City,
		State:       d.State,
		ZipCode:     d.ZipCode,
		PricePerDay: d.PricePerDay,
		Size:        string(d.Size),
		IsAvailable: d.IsAvailable,
		Rating:      d.Rating,
		ReviewCount: d.ReviewCount,
		Capacity:    d.Capacity,
		Weight:      d.Weight,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}

	if d.Owner != nil {
		ownerResp := d.Owner.ToResponse()
		resp.Owner = &ownerResp
	}

	return resp
}
