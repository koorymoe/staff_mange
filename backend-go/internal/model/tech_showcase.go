package model

import (
	"time"

	"github.com/lib/pq"
)

type TechShowcaseItem struct {
	ID          string         `db:"id" json:"id"`
	EmployeeID  string         `db:"employeeId" json:"-"`
	Title       string         `db:"title" json:"title"`
	Description *string        `db:"description" json:"description"`
	MediaUrls   pq.StringArray `db:"mediaUrls" json:"mediaUrls"`
	CreatedAt   time.Time      `db:"createdAt" json:"createdAt"`

	Employee *EmployeeBrief `db:"-" json:"employee"`
}

type CreateTechShowcaseItemRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
}

type AddTechShowcaseMediaRequest struct {
	MediaUrls []string `json:"mediaUrls"`
}
