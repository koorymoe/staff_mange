package model

import "time"

// طلب حذف حجز.
//
// الحجوزات التجريبية والملغاة لازم تنشال، بس الحذف ما يترد — فالإداري
// يطلب، والمراقب أو مدير النظام يوافق أو يرفض. بلا هالدورة، أي غلطة
// تروح بيانات ما ترجع.

const (
	BookingDeleteStatusPending  = "PENDING"
	BookingDeleteStatusApproved = "APPROVED"
	BookingDeleteStatusRejected = "REJECTED"
)

var BookingDeleteStatusLabels = map[string]string{
	BookingDeleteStatusPending:  "بانتظار الموافقة",
	BookingDeleteStatusApproved: "انحذف",
	BookingDeleteStatusRejected: "انرفض الطلب",
}

type BookingDeleteRequest struct {
	ID            string     `db:"id" json:"id"`
	BookingID     string     `db:"bookingId" json:"bookingId"`
	RequestedByID string     `db:"requestedById" json:"requestedById"`
	Reason        string     `db:"reason" json:"reason"`
	Status        string     `db:"status" json:"status"`
	DecidedByID   *string    `db:"decidedById" json:"decidedById"`
	DecidedAt     *time.Time `db:"decidedAt" json:"decidedAt"`
	DecisionNote  *string    `db:"decisionNote" json:"decisionNote"`
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`

	BookingCode     string  `db:"bookingCode" json:"bookingCode"`
	CustomerName    string  `db:"customerName" json:"customerName"`
	BookingStatus   string  `db:"bookingStatus" json:"bookingStatus"`
	RequestedByName string  `db:"requestedByName" json:"requestedByName"`
	DecidedByName   *string `db:"decidedByName" json:"decidedByName"`

	StatusLabel string `db:"-" json:"statusLabel"`
}

type CreateBookingDeleteRequest struct {
	Reason string `json:"reason"`
}

type DecideBookingDeleteRequest struct {
	Approve bool    `json:"approve"`
	Note    *string `json:"note"`
}
