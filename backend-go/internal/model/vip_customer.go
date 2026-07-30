package model

import "time"

// VipCustomer تعليم زبون كـ"شخصية مهمة" — أي موظف يقدر يعلّمه بضغطة زر لحظة
// ما يثبّت حجزه أو يتعامل وياه، والسجل يحفظ مين علّمه وشنو كان طلب الزبون.
// القائمة تُعرض لمدير النظام فقط.
type VipCustomer struct {
	ID                 string    `db:"id" json:"id"`
	CustomerID         string    `db:"customerId" json:"customerId"`
	BookingID          *string   `db:"bookingId" json:"bookingId"`
	RequestSummary     *string   `db:"requestSummary" json:"requestSummary"`
	Note               *string   `db:"note" json:"note"`
	MarkedByEmployeeID string    `db:"markedByEmployeeId" json:"markedByEmployeeId"`
	CreatedAt          time.Time `db:"createdAt" json:"createdAt"`

	// حقول معروضة تُملأ بالاستعلام نفسه (JOIN) — بدون استعلام إضافي لكل صف.
	CustomerName  string  `db:"customerName" json:"customerName"`
	CustomerPhone string  `db:"customerPhone" json:"customerPhone"`
	BookingCode   *string `db:"bookingCode" json:"bookingCode"`
	MarkedByName  string  `db:"markedByName" json:"markedByName"`
}

type MarkVipCustomerRequest struct {
	CustomerID     string  `json:"customerId"`
	BookingID      *string `json:"bookingId"`
	RequestSummary *string `json:"requestSummary"`
	Note           *string `json:"note"`
}
