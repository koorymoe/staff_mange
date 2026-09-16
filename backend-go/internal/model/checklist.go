package model

import (
	"time"

	"github.com/lib/pq"
)

type ProjectChecklist struct {
	ID          string         `db:"id" json:"id"`
	ProjectID   *string        `db:"projectId" json:"projectId"`
	Title       string         `db:"title" json:"title"`
	CreatedByID string         `db:"createdById" json:"-"`
	PhotoUrls   pq.StringArray `db:"photoUrls" json:"photoUrls"`
	CreatedAt   time.Time      `db:"createdAt" json:"createdAt"`

	Project   *Project       `db:"-" json:"project"`
	CreatedBy *EmployeeBrief `db:"-" json:"createdBy"`
}

type CreateChecklistRequest struct {
	ProjectID *string `json:"projectId"`
	Title     string  `json:"title"`
}

type AddChecklistPhotosRequest struct {
	PhotoUrls []string `json:"photoUrls"`
}
