package model

import "time"

type WorkReport struct {
	ID            string    `db:"id" json:"id"`
	BookingID     string    `db:"bookingId" json:"bookingId"`
	EmployeeID    string    `db:"employeeId" json:"employeeId"`
	WorkStatus    string    `db:"workStatus" json:"workStatus"` // COMPLETED | STOPPED
	Events        *string   `db:"events" json:"events"`
	ExtraRequests *string   `db:"extraRequests" json:"extraRequests"`
	CleanedSite   bool      `db:"cleanedSite" json:"cleanedSite"`
	GaveInfo      bool      `db:"gaveInfo" json:"gaveInfo"`
	TookPhotos    bool      `db:"tookPhotos" json:"tookPhotos"`
	StopReason    *string   `db:"stopReason" json:"stopReason"`
	Notes         *string   `db:"notes" json:"notes"`
	CreatedAt     time.Time `db:"createdAt" json:"createdAt"`

	Employee *EmployeeBrief          `db:"-" json:"employee"`
	Booking  *WorkReportBookingBrief `db:"-" json:"booking"`
}

type WorkReportBookingBrief struct {
	ID   string `db:"id" json:"id"`
	Code string `db:"code" json:"code"`
	// اسم الزبون ورقم هاتفه — الهاتف لازم بالتقرير حتى مهندس الجودة يقدر
	// يتصل بالزبون مباشرة من التقرير بدون ما يدور عليه بشاشة ثانية.
	CustomerName  string  `db:"customerName" json:"customerName"`
	CustomerPhone *string `db:"customerPhone" json:"customerPhone"`
}

type CreateWorkReportRequest struct {
	BookingID     string  `json:"bookingId"`
	WorkStatus    string  `json:"workStatus"`
	Events        *string `json:"events"`
	ExtraRequests *string `json:"extraRequests"`
	CleanedSite   bool    `json:"cleanedSite"`
	GaveInfo      bool    `json:"gaveInfo"`
	TookPhotos    bool    `json:"tookPhotos"`
	StopReason    *string `json:"stopReason"`
	Notes         *string `json:"notes"`
}
