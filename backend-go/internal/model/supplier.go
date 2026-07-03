package model

import (
	"time"

	"github.com/lib/pq"
)

type SupplierSpecialty struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Order     int       `db:"order" json:"order"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

type CreateSupplierSpecialtyRequest struct {
	Name string `json:"name"`
}

type Supplier struct {
	ID                 string         `db:"id" json:"id"`
	CompanyName        string         `db:"companyName" json:"companyName"`
	OwnerName          string         `db:"ownerName" json:"ownerName"`
	Phone              string         `db:"phone" json:"phone"`
	Lat                *float64       `db:"lat" json:"lat"`
	Lng                *float64       `db:"lng" json:"lng"`
	IsMaterialSupplier bool           `db:"isMaterialSupplier" json:"isMaterialSupplier"`
	IsContractor       bool           `db:"isContractor" json:"isContractor"`
	TraderTypes        pq.StringArray `db:"traderTypes" json:"traderTypes"`
	Notes              *string        `db:"notes" json:"notes"`
	CreatedAt          time.Time      `db:"createdAt" json:"createdAt"`
	UpdatedAt          time.Time      `db:"updatedAt" json:"updatedAt"`

	Specialties []SupplierSpecialty `db:"-" json:"specialties"`
	AvgRating   float64             `db:"-" json:"avgRating"`
	RatingCount int                 `db:"-" json:"ratingCount"`
}

type SupplierRating struct {
	ID          string    `db:"id" json:"id"`
	SupplierID  string    `db:"supplierId" json:"supplierId"`
	Value       int       `db:"value" json:"value"`
	Note        *string   `db:"note" json:"note"`
	RatedByID   string    `db:"ratedById" json:"ratedById"`
	RatedByName string    `db:"ratedByName" json:"ratedByName"`
	CreatedAt   time.Time `db:"createdAt" json:"createdAt"`
}

type UpsertSupplierRequest struct {
	CompanyName        string   `json:"companyName"`
	OwnerName          string   `json:"ownerName"`
	Phone              string   `json:"phone"`
	Lat                *float64 `json:"lat"`
	Lng                *float64 `json:"lng"`
	IsMaterialSupplier bool     `json:"isMaterialSupplier"`
	IsContractor       bool     `json:"isContractor"`
	TraderTypes        []string `json:"traderTypes"`
	Notes              *string  `json:"notes"`
	SpecialtyIDs       []string `json:"specialtyIds"`
}

type RateSupplierRequest struct {
	Value       *int    `json:"value"`
	Note        *string `json:"note"`
	RatedByID   string  `json:"ratedById"`
	RatedByName string  `json:"ratedByName"`
}
