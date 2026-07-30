package model

import (
	"fmt"
	"time"
)

// SystemPriceCatalog صف واحد من كتالوج أسعار المنظومات الثمانية (تركيب/تسليك/برمجة) —
// نفس بيانات شيت "تكاليف المشروع" الي كان الليدر يستخدمه بجوجل شيت.
type SystemPriceCatalog struct {
	ID         string    `db:"id" json:"id"`
	SystemName string    `db:"systemName" json:"systemName"`
	ItemName   string    `db:"itemName" json:"itemName"`
	Category   string    `db:"category" json:"category"` // install | wiring | programming
	Value      float64   `db:"value" json:"value"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
}

// Material مادة من أرشيف "مواد الشد" — لها كود فريد يستخدم بالفاتورة للبحث
// السريع (مثل VLOOKUP بالشيت القديم).
type Material struct {
	ID             string    `db:"id" json:"id"`
	Name           string    `db:"name" json:"name"`
	Code           string    `db:"code" json:"code"`
	SellPrice      float64   `db:"sellPrice" json:"sellPrice"`
	WholesalePrice float64   `db:"wholesalePrice" json:"wholesalePrice"` // سعر الجملة — profitPerUnit = sellPrice - wholesalePrice
	ProfitPerUnit  float64   `db:"profitPerUnit" json:"profitPerUnit"`
	CreatedAt      time.Time `db:"createdAt" json:"createdAt"`
}

// ExecutionCostItem بند تنفيذ واحد (منظومة + عنصر) يدخل بحساب تكاليف التنفيذ.
type ExecutionCostItem struct {
	SystemName        string `json:"systemName"`
	ItemName          string `json:"itemName"`
	Count             int    `json:"count"`
	HeightMeters      int    `json:"heightMeters"`      // لحساب مضاعف الارتفاع
	WiringItemName    string `json:"wiringItemName"`    // اسم نوع التسليك المختار (اختياري)
	CableLengthMeters int    `json:"cableLengthMeters"` // طول الكيبل بالمتر (اختياري)
	ProgrammingItem   string `json:"programmingItem"`   // اسم خدمة البرمجة المختارة (اختياري)
}

// LeaderInvoiceMaterialItem بند مادة واحد بفاتورة الليدر (من الأرشيف بالكود أو يدوي).
type LeaderInvoiceMaterialItem struct {
	ID              string    `db:"id" json:"id"`
	LeaderInvoiceID string    `db:"leaderInvoiceId" json:"leaderInvoiceId"`
	MaterialID      *string   `db:"materialId" json:"materialId"`
	Name            string    `db:"name" json:"name"`
	Quantity        float64   `db:"quantity" json:"quantity"`
	UnitPrice       float64   `db:"unitPrice" json:"unitPrice"`
	ProfitPerUnit   float64   `db:"profitPerUnit" json:"profitPerUnit"`
	LineTotal       float64   `db:"lineTotal" json:"lineTotal"`
	CreatedAt       time.Time `db:"createdAt" json:"createdAt"`
}

// LeaderInvoice فاتورة الليدر الكاملة — منظومات مختارة (حتى 3)، بنود تنفيذ،
// مواد، خصم، والمجموع الصافي. مبنية لتحل محل شيت جوجل بنفس منطق الحساب تماماً.
type LeaderInvoice struct {
	ID                   string     `db:"id" json:"id"`
	BookingID            *string    `db:"bookingId" json:"bookingId"`
	EmployeeID           string     `db:"employeeId" json:"employeeId"`
	CustomerName         *string    `db:"customerName" json:"customerName"`
	CustomerPhone        *string    `db:"customerPhone" json:"customerPhone"`
	CustomerAddress      *string    `db:"customerAddress" json:"customerAddress"`
	SystemsJSON          string     `db:"systems" json:"-"`
	ItemsJSON            string     `db:"items" json:"-"`
	TotalDeviceCount     int        `db:"totalDeviceCount" json:"totalDeviceCount"`
	ExecutionCost        float64    `db:"executionCost" json:"executionCost"`
	MaterialsTotal       float64    `db:"materialsTotal" json:"materialsTotal"`
	DiscountValue        float64    `db:"discountValue" json:"discountValue"`
	NetTotal             float64    `db:"netTotal" json:"netTotal"`
	AccountingCode       string     `db:"accountingCode" json:"accountingCode"`
	Status               string     `db:"status" json:"status"` // SUBMITTED | APPROVED
	CreatedAt            time.Time  `db:"createdAt" json:"createdAt"`
	ApprovedByEmployeeID *string    `db:"approvedByEmployeeId" json:"approvedByEmployeeId"`
	ApprovedAt           *time.Time `db:"approvedAt" json:"approvedAt"`

	Systems   []string                    `db:"-" json:"systems"`
	Items     []ExecutionCostItem         `db:"-" json:"items"`
	Materials []LeaderInvoiceMaterialItem `db:"-" json:"materials"`
}

// CreateLeaderInvoiceRequest جسم طلب إنشاء فاتورة ليدر جديدة.
type CreateLeaderInvoiceRequest struct {
	BookingID       *string                     `json:"bookingId"`
	CustomerName    *string                     `json:"customerName"`
	CustomerPhone   *string                     `json:"customerPhone"`
	CustomerAddress *string                     `json:"customerAddress"`
	Systems         []string                    `json:"systems"` // حتى 3 منظومات
	Items           []ExecutionCostItem         `json:"items"`
	Materials       []CreateMaterialLineRequest `json:"materials"`
	DiscountValue   float64                     `json:"discountValue"`
}

// EstimateExecutionCostRequest طلب "حساب كلفة" سريع بدون ربط بحجز ولا حفظ —
// يستخدمه الليدر لما زبون يستفسر عن سعر تقريبي، بنفس محرك الحساب بالضبط.
type EstimateExecutionCostRequest struct {
	Items []ExecutionCostItem `json:"items"`
}

// EstimateExecutionCostResponse نتيجة الحساب السريع فقط، بدون أي حفظ بقاعدة البيانات.
type EstimateExecutionCostResponse struct {
	ExecutionCost    int64 `json:"executionCost"`
	TotalDeviceCount int   `json:"totalDeviceCount"`
}

// CreateMaterialLineRequest بند مادة واحد ضمن طلب إنشاء الفاتورة — إما materialCode
// (بحث بالأرشيف) أو name/unitPrice يدوي لمادة مو موجودة بالكود.
type CreateMaterialLineRequest struct {
	MaterialCode  *string  `json:"materialCode"`
	Name          *string  `json:"name"`
	Quantity      float64  `json:"quantity"`
	UnitPrice     *float64 `json:"unitPrice"`
	ProfitPerUnit *float64 `json:"profitPerUnit"`
}

// GenerateAccountingCode يبني كوداً محاسبياً فريداً وقابل للتتبع لفاتورة الليدر —
// لا يطابق صيغة REGEXREPLACE المعقّدة بالشيت الأصلي (غير ضرورية هنا، فقط لازم
// يكون فريد وقابل للتتبع)، بصيغة: LDR-YYYYMMDD-<آخر 6 محارف من معرف الفاتورة>.
func GenerateAccountingCode(invoiceID string, createdAt time.Time) string {
	suffix := invoiceID
	if len(suffix) > 6 {
		suffix = suffix[len(suffix)-6:]
	}
	return fmt.Sprintf("LDR-%s-%s", createdAt.Format("20060102"), suffix)
}
