package model

import (
	"fmt"
	"time"
)

type Customer struct {
	ID           string    `db:"id" json:"id"`
	CustomerCode int       `db:"customerCode" json:"customerCode"`
	Name         string    `db:"name" json:"name"`
	Phone        string    `db:"phone" json:"phone"`
	Location     *string   `db:"location" json:"location"`
	CreatedAt    time.Time `db:"createdAt" json:"createdAt"`
}

// CustomerResponse يضيف حقل "code" المنسّق (CUST-00001) مثل الباك إند القديم بالضبط
type CustomerResponse struct {
	Customer
	Code    string `json:"code"`
	Existed *bool  `json:"existed,omitempty"`
}

func (c Customer) FormatCode() string {
	return fmt.Sprintf("CUST-%05d", c.CustomerCode)
}

func (c Customer) ToResponse() CustomerResponse {
	return CustomerResponse{Customer: c, Code: c.FormatCode()}
}

type CreateCustomerRequest struct {
	Name     string  `json:"name"`
	Phone    string  `json:"phone"`
	Location *string `json:"location"`
}
