package model

import (
	"fmt"
	"time"
)

type Customer struct {
	ID           string  `db:"id" json:"id"`
	CustomerCode int     `db:"customerCode" json:"customerCode"`
	Name         string  `db:"name" json:"name"`
	Phone        string  `db:"phone" json:"phone"`
	Location     *string `db:"location" json:"location"`
	// منصب الزبون — يظهر بالشخصيات المهمة. لازم يكون موجود هنا وإلا
	// SELECT * ينكسر بـ"missing destination name position" (نفس الفخ الي
	// وكعنا بيه مرتين قبل: عمود بقاعدة البيانات بلا حقل بالستركت).
	Position     *string   `db:"position" json:"position"`
	MapLatitude  *float64  `db:"mapLatitude" json:"mapLatitude"`
	MapLongitude *float64  `db:"mapLongitude" json:"mapLongitude"`
	CreatedAt    time.Time `db:"createdAt" json:"createdAt"`
}

// CustomerResponse يضيف حقل "code" المنسّق (CUST-00001) مثل الباك إند القديم بالضبط،
// و"services" — قائمة أسماء الخدمات الي طلبها هذا الزبون (جي بي اس، كاميرات...) —
// نفس الزبون بنفس الكود يمكن يكون عنده أكثر من خدمة.
type CustomerResponse struct {
	Customer
	Code                  string   `json:"code"`
	Services              []string `json:"services"`
	Existed               *bool    `json:"existed,omitempty"`
	PreviousBookingsCount *int     `json:"previousBookingsCount,omitempty"`
}

func (c Customer) FormatCode() string {
	return fmt.Sprintf("CUST-%05d", c.CustomerCode)
}

func (c Customer) ToResponse() CustomerResponse {
	return CustomerResponse{Customer: c, Code: c.FormatCode(), Services: []string{}}
}

type CreateCustomerRequest struct {
	Name         string   `json:"name"`
	Phone        string   `json:"phone"`
	Location     *string  `json:"location"`
	MapLatitude  *float64 `json:"mapLatitude"`
	MapLongitude *float64 `json:"mapLongitude"`
}

type UpdateCustomerRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// CustomerGpsInfo معلومات إضافية خاصة بزبائن الجي بي اس فقط (مستوردة من نظام التتبع القديم)
type CustomerGpsInfo struct {
	ID              string     `db:"id" json:"id"`
	CustomerID      string     `db:"customerId" json:"customerId"`
	GpsNumber       *string    `db:"gpsNumber" json:"gpsNumber"`
	DeviceID        *string    `db:"deviceId" json:"deviceId"`
	SubscriptionEnd *time.Time `db:"subscriptionEnd" json:"subscriptionEnd"`
}

// CustomerGpsResponse زبون + معلوماته الخاصة بالجي بي اس — لعرض "قائمة زبائن الجي بي اس" وحدهم
type CustomerGpsResponse struct {
	CustomerResponse
	GpsNumber       *string    `json:"gpsNumber"`
	DeviceID        *string    `json:"deviceId"`
	SubscriptionEnd *time.Time `json:"subscriptionEnd"`
}
