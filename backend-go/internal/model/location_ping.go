package model

import "time"

// LocationPing نقطة موقع حية يرسلها متصفح الفني (أو أي موظف ميداني) وهو ماشي — سلسلة
// النقاط تكون مسار حركته الفعلي، تُعرض على الخريطة بمتابعة الفرق الميدانية.
type LocationPing struct {
	ID         string    `db:"id" json:"id"`
	EmployeeID string    `db:"employeeId" json:"employeeId"`
	BookingID  *string   `db:"bookingId" json:"bookingId"`
	Latitude   float64   `db:"latitude" json:"latitude"`
	Longitude  float64   `db:"longitude" json:"longitude"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
}

type CreateLocationPingRequest struct {
	BookingID *string `json:"bookingId"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
