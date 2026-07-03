package model

import "time"

type GpsCustomer struct {
	ID                    string    `db:"id" json:"id"`
	FullName              string    `db:"fullName" json:"fullName"`
	FatherName            *string   `db:"fatherName" json:"fatherName"`
	GrandfatherName       *string   `db:"grandfatherName" json:"grandfatherName"`
	Phone                 string    `db:"phone" json:"phone"`
	Address               *string   `db:"address" json:"address"`
	Governorate           *string   `db:"governorate" json:"governorate"`
	IDCardFrontURL        *string   `db:"idCardFrontUrl" json:"idCardFrontUrl"`
	IDCardBackURL         *string   `db:"idCardBackUrl" json:"idCardBackUrl"`
	ResidenceCardFrontURL *string   `db:"residenceCardFrontUrl" json:"residenceCardFrontUrl"`
	ResidenceCardBackURL  *string   `db:"residenceCardBackUrl" json:"residenceCardBackUrl"`
	CreatedAt             time.Time `db:"createdAt" json:"createdAt"`
}

type UpsertGpsCustomerRequest struct {
	FullName              *string `json:"fullName"`
	FatherName            *string `json:"fatherName"`
	GrandfatherName       *string `json:"grandfatherName"`
	Phone                 *string `json:"phone"`
	Address               *string `json:"address"`
	Governorate           *string `json:"governorate"`
	IDCardFrontURL        *string `json:"idCardFrontUrl"`
	IDCardBackURL         *string `json:"idCardBackUrl"`
	ResidenceCardFrontURL *string `json:"residenceCardFrontUrl"`
	ResidenceCardBackURL  *string `json:"residenceCardBackUrl"`
}

type SimCard struct {
	ID         string    `db:"id" json:"id"`
	SimNumber  string    `db:"simNumber" json:"simNumber"`
	ICCID      *string   `db:"iccid" json:"iccid"`
	Operator   string    `db:"operator" json:"operator"`
	Status     string    `db:"status" json:"status"`
	CustomerID *string   `db:"customerId" json:"customerId"`
	Notes      *string   `db:"notes" json:"notes"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
}

type UpsertSimCardRequest struct {
	SimNumber  *string `json:"simNumber"`
	ICCID      *string `json:"iccid"`
	Operator   *string `json:"operator"`
	Status     *string `json:"status"`
	CustomerID *string `json:"customerId"`
	Notes      *string `json:"notes"`
}

type GpsDeviceRequest struct {
	ID                  string     `db:"id" json:"id"`
	CustomerID          string     `db:"customerId" json:"-"`
	EmployeeID          string     `db:"employeeId" json:"-"`
	AdminID             *string    `db:"adminId" json:"adminId"`
	PurchaseType        string     `db:"purchaseType" json:"purchaseType"`
	SubscriptionType    string     `db:"subscriptionType" json:"subscriptionType"`
	SubscriptionStart   *time.Time `db:"subscriptionStart" json:"subscriptionStart"`
	SubscriptionEnd     *time.Time `db:"subscriptionEnd" json:"subscriptionEnd"`
	SubscriptionStatus  string     `db:"subscriptionStatus" json:"subscriptionStatus"`
	Status              string     `db:"status" json:"status"`
	SimCardID           *string    `db:"simCardId" json:"-"`
	Notes               *string    `db:"notes" json:"notes"`
	IsChecked           bool       `db:"isChecked" json:"isChecked"`
	IsActivated         bool       `db:"isActivated" json:"isActivated"`
	IsDelivered         bool       `db:"isDelivered" json:"isDelivered"`
	InvoicePhotoURL     *string    `db:"invoicePhotoUrl" json:"invoicePhotoUrl"`
	GpsNumber           *string    `db:"gpsNumber" json:"gpsNumber"`
	ResidenceCardNumber *string    `db:"residenceCardNumber" json:"residenceCardNumber"`
	ActivationDate      *time.Time `db:"activationDate" json:"activationDate"`
	DeliveredAt         *time.Time `db:"deliveredAt" json:"deliveredAt"`
	CreatedAt           time.Time  `db:"createdAt" json:"createdAt"`

	Customer *GpsCustomer   `db:"-" json:"customer"`
	Employee *EmployeeBrief `db:"-" json:"employee"`
	SimCard  *SimCard       `db:"-" json:"simCard"`
}

type UpsertGpsDeviceRequest struct {
	CustomerID          *string    `json:"customerId"`
	EmployeeID          *string    `json:"employeeId"`
	AdminID             *string    `json:"adminId"`
	PurchaseType        *string    `json:"purchaseType"`
	SubscriptionType    *string    `json:"subscriptionType"`
	SubscriptionStart   *time.Time `json:"subscriptionStart"`
	SubscriptionEnd     *time.Time `json:"subscriptionEnd"`
	SubscriptionStatus  *string    `json:"subscriptionStatus"`
	Status              *string    `json:"status"`
	SimCardID           *string    `json:"simCardId"`
	Notes               *string    `json:"notes"`
	IsChecked           *bool      `json:"isChecked"`
	IsActivated         *bool      `json:"isActivated"`
	IsDelivered         *bool      `json:"isDelivered"`
	InvoicePhotoURL     *string    `json:"invoicePhotoUrl"`
	GpsNumber           *string    `json:"gpsNumber"`
	ResidenceCardNumber *string    `json:"residenceCardNumber"`
	ActivationDate      *time.Time `json:"activationDate"`
	DeliveredAt         *time.Time `json:"deliveredAt"`
}

type GpsRenewalRequest struct {
	ID               string     `db:"id" json:"id"`
	CustomerID       string     `db:"customerId" json:"-"`
	DeviceRequestID  string     `db:"deviceRequestId" json:"-"`
	EmployeeID       string     `db:"employeeId" json:"-"`
	AdminID          *string    `db:"adminId" json:"adminId"`
	SubscriptionType string     `db:"subscriptionType" json:"subscriptionType"`
	NewEndDate       *time.Time `db:"newEndDate" json:"newEndDate"`
	Status           string     `db:"status" json:"status"`
	CreatedAt        time.Time  `db:"createdAt" json:"createdAt"`

	Customer      *GpsCustomer      `db:"-" json:"customer"`
	DeviceRequest *GpsDeviceRequest `db:"-" json:"deviceRequest"`
}

type UpsertGpsRenewalRequest struct {
	CustomerID       *string    `json:"customerId"`
	DeviceRequestID  *string    `json:"deviceRequestId"`
	EmployeeID       *string    `json:"employeeId"`
	AdminID          *string    `json:"adminId"`
	SubscriptionType *string    `json:"subscriptionType"`
	NewEndDate       *time.Time `json:"newEndDate"`
	Status           *string    `json:"status"`
}

type GpsMaintenanceRequest struct {
	ID                 string     `db:"id" json:"id"`
	CustomerID         string     `db:"customerId" json:"-"`
	EmployeeID         string     `db:"employeeId" json:"-"`
	AdminID            *string    `db:"adminId" json:"adminId"`
	ProblemDescription string     `db:"problemDescription" json:"problemDescription"`
	Status             string     `db:"status" json:"status"`
	AdminNotes         *string    `db:"adminNotes" json:"adminNotes"`
	CreatedAt          time.Time  `db:"createdAt" json:"createdAt"`
	ResolvedAt         *time.Time `db:"resolvedAt" json:"resolvedAt"`

	Customer *GpsCustomer   `db:"-" json:"customer"`
	Employee *EmployeeBrief `db:"-" json:"employee"`
}

type UpsertGpsMaintenanceRequest struct {
	CustomerID         *string    `json:"customerId"`
	EmployeeID         *string    `json:"employeeId"`
	AdminID            *string    `json:"adminId"`
	ProblemDescription *string    `json:"problemDescription"`
	Status             *string    `json:"status"`
	AdminNotes         *string    `json:"adminNotes"`
	ResolvedAt         *time.Time `json:"resolvedAt"`
}

type GpsSubscriptionPrice struct {
	ID               string    `db:"id" json:"id"`
	SubscriptionType string    `db:"subscriptionType" json:"subscriptionType"`
	Price            float64   `db:"price" json:"price"`
	CreatedAt        time.Time `db:"createdAt" json:"createdAt"`
}

type UpsertGpsSubscriptionPriceRequest struct {
	ID               *string `json:"id"`
	SubscriptionType string  `json:"subscriptionType"`
	Price            float64 `json:"price"`
}

type GpsStats struct {
	DevicesByStatus []GpsStatusCount `json:"devicesByStatus"`
	TotalDevices    int              `json:"totalDevices"`
	TotalCustomers  int              `json:"totalCustomers"`
	TotalSims       int              `json:"totalSims"`
	AvailableSims   int              `json:"availableSims"`
	InUseSims       int              `json:"inUseSims"`
}

type GpsStatusCount struct {
	Status string `db:"status" json:"status"`
	Count  int    `db:"count" json:"count"`
}
