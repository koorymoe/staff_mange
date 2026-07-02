package model

import "time"

type Skill struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	ServiceID string    `db:"serviceId" json:"serviceId"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

type Service struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Category  *string   `db:"category" json:"category"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
	Skills    []Skill   `db:"-" json:"skills"`
}

type CreateServiceRequest struct {
	Name     string  `json:"name"`
	Category *string `json:"category"`
}
