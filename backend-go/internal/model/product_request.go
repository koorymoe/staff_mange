package model

import "time"

// ProductRequest اقتراح منتج جديد يُضاف لكتالوج النظام (وحدة التقنيين → إدارة
// المنتجات) — مختلف عن ProcurementRequest (طلب توفير/شراء لكمية موجودة أصلاً):
// هذا اقتراح "نضيف هذا المنتج للكتالوج" بكامل مواصفاته، يراجعه المدير ويوافق
// أو يرفض. يقدر يفتحه: المدير، أي تقني (content_technician)، أو مسؤول المشتريات.
type ProductRequest struct {
	ID            string     `db:"id" json:"id"`
	RequestedByID string     `db:"requestedById" json:"-"`
	ProductName   string     `db:"productName" json:"productName"`
	Specs         *string    `db:"specs" json:"specs"`       // المواصفات
	Source        *string    `db:"source" json:"source"`     // المصدر
	Model         *string    `db:"model" json:"model"`       // الموديل
	Category      *string    `db:"category" json:"category"` // التصنيف/الدورة
	Price         *float64   `db:"price" json:"price"`
	Status        string     `db:"status" json:"status"` // PENDING | APPROVED | REJECTED
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`
	ResolvedAt    *time.Time `db:"resolvedAt" json:"resolvedAt"`
	ResolvedByID  *string    `db:"resolvedById" json:"-"`

	RequestedBy *EmployeeBrief `db:"-" json:"requestedBy"`
	ResolvedBy  *EmployeeBrief `db:"-" json:"resolvedBy"`
}

type CreateProductProposalRequest struct {
	ProductName string   `json:"productName"`
	Specs       *string  `json:"specs"`
	Source      *string  `json:"source"`
	Model       *string  `json:"model"`
	Category    *string  `json:"category"`
	Price       *float64 `json:"price"`
}

const (
	ProductRequestPending  = "PENDING"
	ProductRequestApproved = "APPROVED"
	ProductRequestRejected = "REJECTED"
)
