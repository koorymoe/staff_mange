package model

import "time"

// BookingSurveyReport نتائج زيارة معاينة («كشف») — بديل خفيف عن
// WorkReport لحجوزات بلا فاتورة ولا شغل فعلي، بس تحتاج توثيق شنو
// شاف الكادر بالموقع.
type BookingSurveyReport struct {
	ID            string    `db:"id" json:"id"`
	BookingID     string    `db:"bookingId" json:"bookingId"`
	EmployeeID    string    `db:"employeeId" json:"employeeId"`
	CustomerWants string    `db:"customerWants" json:"customerWants"`
	SiteAreaSqm   *float64  `db:"siteAreaSqm" json:"siteAreaSqm,omitempty"`
	SiteDetails   *string   `db:"siteDetails" json:"siteDetails,omitempty"`
	OtherNotes    *string   `db:"otherNotes" json:"otherNotes,omitempty"`
	CreatedAt     time.Time `db:"createdAt" json:"createdAt"`

	Employee *EmployeeBrief `db:"-" json:"employee,omitempty"`
}

type CreateBookingSurveyReportRequest struct {
	BookingID     string   `json:"bookingId"`
	CustomerWants string   `json:"customerWants"`
	SiteAreaSqm   *float64 `json:"siteAreaSqm"`
	SiteDetails   *string  `json:"siteDetails"`
	OtherNotes    *string  `json:"otherNotes"`
}
