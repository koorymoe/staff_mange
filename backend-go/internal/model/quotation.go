package model

import "time"

type Quotation struct {
	ID                  string    `db:"id" json:"id"`
	QuotationNumber     string    `db:"quotationNumber" json:"quotationNumber"`
	CustomerName        string    `db:"customerName" json:"customerName"`
	CustomerPhone       *string   `db:"customerPhone" json:"customerPhone"`
	CustomerAddress     *string   `db:"customerAddress" json:"customerAddress"`
	ProjectName         *string   `db:"projectName" json:"projectName"`
	GrandTotal          float64   `db:"grandTotal" json:"grandTotal"`
	DiscountPercent     float64   `db:"discountPercent" json:"discountPercent"`
	DiscountValue       float64   `db:"discountValue" json:"discountValue"`
	NetTotal            float64   `db:"netTotal" json:"netTotal"`
	Notes               *string   `db:"notes" json:"notes"`
	Duration            *string   `db:"duration" json:"duration"`
	Status              string    `db:"status" json:"status"`
	CreatedByEmployeeID string    `db:"createdByEmployeeId" json:"-"`
	CreatedAt           time.Time `db:"createdAt" json:"createdAt"`

	Items             []QuotationItem `db:"-" json:"items"`
	CreatedByEmployee *EmployeeBrief  `db:"-" json:"createdByEmployee"`
}

type QuotationItem struct {
	ID          string  `db:"id" json:"id"`
	QuotationID string  `db:"quotationId" json:"quotationId"`
	ProductName string  `db:"productName" json:"productName"`
	Unit        *string `db:"unit" json:"unit"`
	Quantity    int     `db:"quantity" json:"quantity"`
	UnitPrice   float64 `db:"unitPrice" json:"unitPrice"`
	TotalPrice  float64 `db:"totalPrice" json:"totalPrice"`
}

type QuotationItemInput struct {
	ProductName string  `json:"productName"`
	Unit        *string `json:"unit"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
}

type CreateQuotationRequest struct {
	CustomerName        string               `json:"customerName"`
	CustomerPhone       *string              `json:"customerPhone"`
	CustomerAddress     *string              `json:"customerAddress"`
	ProjectName         *string              `json:"projectName"`
	Items               []QuotationItemInput `json:"items"`
	DiscountPercent     *float64             `json:"discountPercent"`
	Notes               *string              `json:"notes"`
	Duration            *string              `json:"duration"`
	CreatedByEmployeeID string               `json:"createdByEmployeeId"`
}

type UpdateQuotationRequest struct {
	CustomerName    *string              `json:"customerName"`
	CustomerPhone   *string              `json:"customerPhone"`
	CustomerAddress *string              `json:"customerAddress"`
	ProjectName     *string              `json:"projectName"`
	DiscountPercent *float64             `json:"discountPercent"`
	Notes           *string              `json:"notes"`
	Duration        *string              `json:"duration"`
	Status          *string              `json:"status"`
	Items           []QuotationItemInput `json:"items"`
}
