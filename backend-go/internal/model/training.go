package model

import "time"

type TrainingMaterial struct {
	ID        string    `db:"id" json:"id"`
	ServiceID string    `db:"serviceId" json:"serviceId"`
	Title     string    `db:"title" json:"title"`
	URL       string    `db:"url" json:"url"`
	Type      string    `db:"type" json:"type"`
	Order     int       `db:"order" json:"order"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`

	Service *Service `db:"-" json:"service"`
}

type CreateTrainingMaterialRequest struct {
	ServiceID string  `json:"serviceId"`
	Title     string  `json:"title"`
	URL       string  `json:"url"`
	Type      *string `json:"type"`
	Order     *int    `json:"order"`
}

type UpdateTrainingMaterialRequest struct {
	Title *string `json:"title"`
	URL   *string `json:"url"`
	Type  *string `json:"type"`
	Order *int    `json:"order"`
}

type EmployeeTrainingAssignment struct {
	ID         string    `db:"id" json:"id"`
	EmployeeID string    `db:"employeeId" json:"employeeId"`
	ServiceID  string    `db:"serviceId" json:"serviceId"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
}

type SetTrainingAssignmentsRequest struct {
	ServiceIDs []string `json:"serviceIds"`
}

type MaterialsMineResponse struct {
	Services  []Service          `json:"services"`
	Materials []TrainingMaterial `json:"materials"`
}
