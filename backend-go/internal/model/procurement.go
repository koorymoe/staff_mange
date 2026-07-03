package model

import "time"

type ProcurementRequest struct {
	ID               string     `db:"id" json:"id"`
	Code             string     `db:"code" json:"code"`
	RequestedByID    string     `db:"requestedById" json:"-"`
	BookingID        *string    `db:"bookingId" json:"-"`
	Notes            *string    `db:"notes" json:"notes"`
	Status           string     `db:"status" json:"status"`
	FulfilledByID    *string    `db:"fulfilledById" json:"-"`
	TotalCost        *float64   `db:"totalCost" json:"totalCost"`
	FulfillmentNotes *string    `db:"fulfillmentNotes" json:"fulfillmentNotes"`
	CreatedAt        time.Time  `db:"createdAt" json:"createdAt"`
	FulfilledAt      *time.Time `db:"fulfilledAt" json:"fulfilledAt"`

	Items       []ProcurementItem        `db:"-" json:"items"`
	RequestedBy *EmployeeBrief           `db:"-" json:"requestedBy"`
	FulfilledBy *EmployeeBrief           `db:"-" json:"fulfilledBy"`
	Booking     *ProcurementBookingBrief `db:"-" json:"booking"`
}

// ProcurementBookingBrief يماثل include: { booking: { include: { customer: { select ... } } } } بالباك إند القديم
type ProcurementBookingBrief struct {
	ID       string         `db:"id" json:"id"`
	Customer *CustomerBrief `db:"-" json:"customer"`
}

type CustomerBrief struct {
	ID    string `db:"id" json:"id"`
	Name  string `db:"name" json:"name"`
	Phone string `db:"phone" json:"phone"`
}

type ProcurementItem struct {
	ID          string   `db:"id" json:"id"`
	RequestID   string   `db:"requestId" json:"requestId"`
	ProductName string   `db:"productName" json:"productName"`
	Quantity    int      `db:"quantity" json:"quantity"`
	UnitPrice   *float64 `db:"unitPrice" json:"unitPrice"`
	TotalPrice  *float64 `db:"totalPrice" json:"totalPrice"`
	Fulfilled   bool     `db:"fulfilled" json:"fulfilled"`
}

type CreateProcurementItemRequest struct {
	ProductName string   `json:"productName"`
	Quantity    int      `json:"quantity"`
	UnitPrice   *float64 `json:"unitPrice"`
	TotalPrice  *float64 `json:"totalPrice"`
}

type CreateProcurementRequestRequest struct {
	RequestedByID string                         `json:"requestedById"`
	BookingID     *string                        `json:"bookingId"`
	Notes         *string                        `json:"notes"`
	Items         []CreateProcurementItemRequest `json:"items"`
}

type UpdateProcurementStatusRequest struct {
	Status string `json:"status"`
}

type FulfillProcurementItemRequest struct {
	ID         string   `json:"id"`
	UnitPrice  *float64 `json:"unitPrice"`
	TotalPrice *float64 `json:"totalPrice"`
	Fulfilled  *bool    `json:"fulfilled"`
}

type FulfillProcurementRequestRequest struct {
	FulfilledByID    string                          `json:"fulfilledById"`
	TotalCost        *float64                        `json:"totalCost"`
	FulfillmentNotes *string                         `json:"fulfillmentNotes"`
	Items            []FulfillProcurementItemRequest `json:"items"`
}

type ProcurementStats struct {
	TotalSpent     float64            `json:"totalSpent"`
	TotalItems     int                `json:"totalItems"`
	PendingCount   int                `json:"pendingCount"`
	MonthlySpent   float64            `json:"monthlySpent"`
	FulfilledCount int                `json:"fulfilledCount"`
	ByMonth        map[string]float64 `json:"byMonth"`
}
