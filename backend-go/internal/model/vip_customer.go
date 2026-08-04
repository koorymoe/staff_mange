package model

import "time"

// VipCustomer تعليم زبون كـ"شخصية مهمة" — أي موظف يقدر يعلّمه بضغطة زر لحظة
// ما يثبّت حجزه أو يتعامل وياه، والسجل يحفظ مين علّمه وشنو كان طلب الزبون.
// القائمة تُعرض لمدير النظام فقط.
type VipCustomer struct {
	ID               string  `db:"id" json:"id"`
	CustomerID       string  `db:"customerId" json:"customerId"`
	BookingID        *string `db:"bookingId" json:"bookingId"`
	ProjectID        *string `db:"projectId" json:"projectId"`
	RequestSummary   *string `db:"requestSummary" json:"requestSummary"`
	CustomerPosition *string `db:"customerPosition" json:"customerPosition"`
	// Source من وين انرحّل: MANUAL (موظف علّمه) أو BOOKING أو PROJECT
	Source             string    `db:"source" json:"source"`
	Note               *string   `db:"note" json:"note"`
	MarkedByEmployeeID string    `db:"markedByEmployeeId" json:"markedByEmployeeId"`
	CreatedAt          time.Time `db:"createdAt" json:"createdAt"`

	// حقول معروضة تُملأ بالاستعلام نفسه (JOIN) — بدون استعلام إضافي لكل صف.
	CustomerName  string  `db:"customerName" json:"customerName"`
	CustomerPhone string  `db:"customerPhone" json:"customerPhone"`
	BookingCode   *string `db:"bookingCode" json:"bookingCode"`
	ProjectCode   *string `db:"projectCode" json:"projectCode"`
	ProjectName   *string `db:"projectName" json:"projectName"`
	SourceLabel   string  `db:"-" json:"sourceLabel"`
	MarkedByName  string  `db:"markedByName" json:"markedByName"`
}

type MarkVipCustomerRequest struct {
	CustomerID string `json:"customerId"`
	// Phone بديل عن CustomerID بالإضافة اليدوية: الإداري يدز الرقم بس
	// والسيرفر يطلع الزبون منه.
	Phone            string  `json:"phone"`
	BookingID        *string `json:"bookingId"`
	ProjectID        *string `json:"projectId"`
	RequestSummary   *string `json:"requestSummary"`
	CustomerPosition *string `json:"customerPosition"`
	Note             *string `json:"note"`
}

// مصادر ترحيل الشخصية المهمة
const (
	VipSourceManual  = "MANUAL"
	VipSourceBooking = "BOOKING"
	VipSourceProject = "PROJECT"
)

var VipSourceLabels = map[string]string{
	VipSourceManual:  "علّمه موظف",
	VipSourceBooking: "مرحّل من حجز",
	VipSourceProject: "مرحّل من مشروع",
}
