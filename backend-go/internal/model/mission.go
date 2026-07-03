package model

import (
	"time"

	"github.com/lib/pq"
)

type Mission struct {
	ID               string         `db:"id" json:"id"`
	Code             string         `db:"code" json:"code"`
	BookingID        string         `db:"bookingId" json:"bookingId"`
	Stage            string         `db:"stage" json:"stage"`
	LeaderID         string         `db:"leaderId" json:"leaderId"`
	MemberIDs        pq.StringArray `db:"memberIds" json:"memberIds"`
	CustomerLat      *float64       `db:"customerLat" json:"customerLat"`
	CustomerLng      *float64       `db:"customerLng" json:"customerLng"`
	CustomerAddress  *string        `db:"customerAddress" json:"customerAddress"`
	AssignedAt       time.Time      `db:"assignedAt" json:"assignedAt"`
	MaterialsReadyAt *time.Time     `db:"materialsReadyAt" json:"materialsReadyAt"`
	DepartedAt       *time.Time     `db:"departedAt" json:"departedAt"`
	ArrivedAt        *time.Time     `db:"arrivedAt" json:"arrivedAt"`
	WorkStartedAt    *time.Time     `db:"workStartedAt" json:"workStartedAt"`
	CompletedAt      *time.Time     `db:"completedAt" json:"completedAt"`
	StoppedAt        *time.Time     `db:"stoppedAt" json:"stoppedAt"`
	DepartureLat     *float64       `db:"departureLat" json:"departureLat"`
	DepartureLng     *float64       `db:"departureLng" json:"departureLng"`
	ArrivalLat       *float64       `db:"arrivalLat" json:"arrivalLat"`
	ArrivalLng       *float64       `db:"arrivalLng" json:"arrivalLng"`
	EstimatedMinutes *int           `db:"estimatedMinutes" json:"estimatedMinutes"`
	ActualMinutes    *int           `db:"actualMinutes" json:"actualMinutes"`
	DistanceKm       *float64       `db:"distanceKm" json:"distanceKm"`
	StopReason       *string        `db:"stopReason" json:"stopReason"`
	Notes            *string        `db:"notes" json:"notes"`
	CreatedAt        time.Time      `db:"createdAt" json:"createdAt"`
	UpdatedAt        time.Time      `db:"updatedAt" json:"updatedAt"`

	Booking *Booking        `db:"-" json:"booking,omitempty"`
	Events  []MissionEvent  `db:"-" json:"events,omitempty"`
	Leader  *EmployeeBrief  `db:"-" json:"leader,omitempty"`
	Members []EmployeeBrief `db:"-" json:"members,omitempty"`
}

type MissionEvent struct {
	ID         string    `db:"id" json:"id"`
	MissionID  string    `db:"missionId" json:"missionId"`
	EmployeeID string    `db:"employeeId" json:"employeeId"`
	Action     string    `db:"action" json:"action"`
	Lat        *float64  `db:"lat" json:"lat"`
	Lng        *float64  `db:"lng" json:"lng"`
	Note       *string   `db:"note" json:"note"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
}

type CreateMissionRequest struct {
	BookingID       string   `json:"bookingId"`
	LeaderID        string   `json:"leaderId"`
	MemberIDs       []string `json:"memberIds"`
	CustomerLat     *string  `json:"customerLat"`
	CustomerLng     *string  `json:"customerLng"`
	CustomerAddress *string  `json:"customerAddress"`
}

type UpdateMissionStageRequest struct {
	Stage            string  `json:"stage"`
	EmployeeID       string  `json:"employeeId"`
	Lat              *string `json:"lat"`
	Lng              *string `json:"lng"`
	Note             *string `json:"note"`
	EstimatedMinutes *string `json:"estimatedMinutes"`
	DistanceKm       *string `json:"distanceKm"`
	StopReason       *string `json:"stopReason"`
}

type MissionMonitorResponse struct {
	Missions []Mission           `json:"missions"`
	Stats    MissionMonitorStats `json:"stats"`
}

type MissionMonitorStats struct {
	Total     int `json:"total"`
	Assigned  int `json:"assigned"`
	Preparing int `json:"preparing"`
	EnRoute   int `json:"enRoute"`
	Arrived   int `json:"arrived"`
	Working   int `json:"working"`
}

type MissionPerformanceReport struct {
	Employee          EmployeeBrief `json:"employee"`
	TotalMissions     int           `json:"totalMissions"`
	Completed         int           `json:"completed"`
	Stopped           int           `json:"stopped"`
	OnTime            int           `json:"onTime"`
	Late              int           `json:"late"`
	AvgDelayMinutes   int           `json:"avgDelayMinutes"`
	CompliancePercent int           `json:"compliancePercent"`
}
