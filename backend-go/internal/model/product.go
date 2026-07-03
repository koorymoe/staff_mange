package model

import "time"

type Product struct {
	ID           string    `db:"id" json:"id"`
	Name         string    `db:"name" json:"name"`
	Unit         *string   `db:"unit" json:"unit"`
	DefaultPrice *float64  `db:"defaultPrice" json:"defaultPrice"`
	ImageBase64  *string   `db:"imageBase64" json:"imageBase64"`
	CreatedAt    time.Time `db:"createdAt" json:"createdAt"`
}

type CreateProductRequest struct {
	Name         string   `json:"name"`
	Unit         *string  `json:"unit"`
	DefaultPrice *float64 `json:"defaultPrice"`
	ImageBase64  *string  `json:"imageBase64"`
}

type UpdateProductRequest struct {
	Name         *string  `json:"name"`
	Unit         *string  `json:"unit"`
	DefaultPrice *float64 `json:"defaultPrice"`
	ImageBase64  *string  `json:"imageBase64"`
}
