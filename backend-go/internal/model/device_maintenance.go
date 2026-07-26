package model

import "time"

// DeviceMaintenanceTicket يمثل تذكرة صيانة جهاز عام (كاميرا/إنذار/بصمة/أي جهاز)
// جايه من الزبون — مقابل شيت "صيانة الاجهزة"، منفصل تماماً عن صيانة اشتراكات
// الـGPS (GpsMaintenance*) الي هي خاصة بأجهزة تتبع مباعة للزبائن.
type DeviceMaintenanceTicket struct {
	ID                 string     `db:"id" json:"id"`
	AppointmentDate    *time.Time `db:"appointmentDate" json:"appointmentDate"`
	CustomerID         string     `db:"customerId" json:"customerId"`
	DeviceTypeName     string     `db:"deviceTypeName" json:"deviceTypeName"`
	Problem            string     `db:"problem" json:"problem"`
	DeviceSerial       *string    `db:"deviceSerial" json:"deviceSerial"`
	ReceivedAt         *time.Time `db:"receivedAt" json:"receivedAt"`
	DeliveredAt        *time.Time `db:"deliveredAt" json:"deliveredAt"`
	InvoiceNumber      string     `db:"invoiceNumber" json:"invoiceNumber"`
	EmployeeID         string     `db:"employeeId" json:"employeeId"`
	CreatedAt          time.Time  `db:"createdAt" json:"createdAt"`

	Customer *CustomerResponse `db:"-" json:"customer"`
	Employee *EmployeeBrief    `db:"-" json:"employee"`
	Status   string            `db:"-" json:"status"`
}

// DeriveStatus يحسب حالة التذكرة من طابعي الاستلام والتسليم الزمنيين:
// لا استلام = NEW (جديدة، بانتظار استلام الجهاز من الزبون)
// استلام بدون تسليم = IN_PROGRESS (الجهاز عندنا قيد الصيانة)
// استلام + تسليم = DELIVERED (تم تسليم الجهاز للزبون بعد الصيانة)
func (t DeviceMaintenanceTicket) DeriveStatus() string {
	switch {
	case t.DeliveredAt != nil:
		return "DELIVERED"
	case t.ReceivedAt != nil:
		return "IN_PROGRESS"
	default:
		return "NEW"
	}
}

type CreateDeviceMaintenanceTicketRequest struct {
	AppointmentDate *time.Time `json:"appointmentDate"`
	CustomerCode    int        `json:"customerCode"`
	DeviceTypeName  string     `json:"deviceTypeName"`
	Problem         string     `json:"problem"`
	DeviceSerial    *string    `json:"deviceSerial"`
}

type UpdateDeviceMaintenanceTicketRequest struct {
	AppointmentDate *time.Time `json:"appointmentDate"`
	DeviceTypeName  *string    `json:"deviceTypeName"`
	Problem         *string    `json:"problem"`
	DeviceSerial    *string    `json:"deviceSerial"`
	MarkReceived    bool       `json:"markReceived"`
	MarkDelivered   bool       `json:"markDelivered"`
}
