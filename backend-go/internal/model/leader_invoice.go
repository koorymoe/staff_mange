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
	SystemName         string `json:"systemName"`
	ItemName           string `json:"itemName"`
	Count              int    `json:"count"`
	HeightMeters       int    `json:"heightMeters"`       // لحساب مضاعف الارتفاع
	WiringItemName     string `json:"wiringItemName"`     // اسم نوع التسليك المختار (اختياري)
	WiringHeightMeters int    `json:"wiringHeightMeters"` // ارتفاع التسليك — قاعدته ثنائية (>=5 متر = ضعف)
	CableLengthMeters  int    `json:"cableLengthMeters"`  // طول الكيبل بالمتر (اختياري)
	ProgrammingItem    string `json:"programmingItem"`    // اسم خدمة البرمجة المختارة (اختياري)
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

// ExecutionCostBreakdownLine تفصيل حساب بند واحد — يُعرض بالواجهة حتى الليدر
// يشوف من وين طلع كل رقم بدل ما يثق برقم أعمى.
type ExecutionCostBreakdownLine struct {
	SystemName string `json:"systemName"`
	ItemName   string `json:"itemName"`
	Count      int    `json:"count"`

	UnitInstallPrice float64 `json:"unitInstallPrice"`
	HeightMeters     int     `json:"heightMeters"`
	HeightMultiplier float64 `json:"heightMultiplier"`
	InstallTotal     float64 `json:"installTotal"`

	WiringItemName      string  `json:"wiringItemName"`
	WiringMultiplier    float64 `json:"wiringMultiplier"`
	WiringHeightMeters  int     `json:"wiringHeightMeters"`
	WiringHeightWeight  float64 `json:"wiringHeightWeight"`
	CableLengthMeters   int     `json:"cableLengthMeters"`
	WiringPricePerMeter float64 `json:"wiringPricePerMeter"`
	WiringByDeviceCount float64 `json:"wiringByDeviceCount"`
	WiringByCableLength float64 `json:"wiringByCableLength"`
	WiringBasis         string  `json:"wiringBasis"`
	WiringTotal         float64 `json:"wiringTotal"`

	ProgrammingItem  string  `json:"programmingItem"`
	ProgrammingTotal float64 `json:"programmingTotal"`

	LineTotal float64 `json:"lineTotal"`
}

// ExecutionCostSystemMinimum تفصيل تطبيق الحدود الدنيا لمنظومة واحدة — الشيت
// يطبّقها لكل منظومة على حدة (صفَّي G59 و R59 بكل بلوك منظومة).
type ExecutionCostSystemMinimum struct {
	SystemName string `json:"systemName"`

	DeviceCount             int     `json:"deviceCount"`
	InstallWiringCalculated float64 `json:"installWiringCalculated"`
	InstallMinimumPerDevice float64 `json:"installMinimumPerDevice"`
	InstallMinimumTotal     float64 `json:"installMinimumTotal"`
	InstallApplied          float64 `json:"installApplied"`
	InstallFloorUsed        bool    `json:"installFloorUsed"`

	ProgrammingCount      int     `json:"programmingCount"`
	ProgrammingCalculated float64 `json:"programmingCalculated"`
	ProgrammingMinimum    float64 `json:"programmingMinimum"`
	ProgrammingApplied    float64 `json:"programmingApplied"`
	ProgrammingFloorUsed  bool    `json:"programmingFloorUsed"`
}

// EstimateExecutionCostResponse نتيجة الحساب السريع فقط، بدون أي حفظ بقاعدة البيانات.
type EstimateExecutionCostResponse struct {
	ExecutionCost    int64                        `json:"executionCost"`
	TotalDeviceCount int                          `json:"totalDeviceCount"`
	Breakdown        []ExecutionCostBreakdownLine `json:"breakdown"`
	SystemMinimums   []ExecutionCostSystemMinimum `json:"systemMinimums"`
}

// ── شيت "حساب تكلفة التنفيذ" (منظومة كاميرات المراقبة) ──
// شيت مستقل تماماً عن "تكاليف المشروع" وله معادلة مختلفة: أربع طبقات ضرب
// متتالية على سعر أساس مأخوذ من شريحة طول الكيبل، ثم أعمال إضافية وخصم.

// CameraCostRow صف كاميرا واحدة بالاستمارة (الشيت يسمح 10 صفوف).
type CameraCostRow struct {
	NormalCableMeters float64 `json:"normalCableMeters"` // طول الكيبل عادي
	VipCableMeters    float64 `json:"vipCableMeters"`    // طول الكيبل VIP (يُضرب 1.2)
	HeightAbove3m     bool    `json:"heightAbove3m"`     // ارتفاع الكاميرا أعلى من 3 متر
}

// CameraCostExtras الأعمال الإضافية بأسفل الاستمارة.
type CameraCostExtras struct {
	ScreenLarge43Count int     `json:"screenLarge43Count"`   // تثبيت شاشة 43 وأكبر × 15000
	ScreenSmall43Count int     `json:"screenSmall43Count"`   // تثبيت شاشة أصغر من 43 × 7500
	RackCount          int     `json:"rackCount"`            // تثبيت راك × 15000
	BoardCount         int     `json:"boardCount"`           // تثبيت بورد × 7500
	VipInternetMeters  int     `json:"vipInternetMeters"`    // مد كيبل انترنيت VIP × 400
	NormalInternetM    int     `json:"normalInternetMeters"` // مد كيبل انترنيت عادي × 200
	ProgrammingAmount  float64 `json:"programmingAmount"`    // برمجة — مبلغ يُدخل يدوي
	OtherAmount        float64 `json:"otherAmount"`          // غيرها — مبلغ يُدخل يدوي
}

// CameraCostRequest طلب حساب استمارة كاميرات المراقبة.
type CameraCostRequest struct {
	PlaceType  string           `json:"placeType"`  // منزل سكني / محل تجاري / مدرسة او شركة / مصنع او معمل
	SystemType string           `json:"systemType"` // ANLOGE / IP
	Rows       []CameraCostRow  `json:"rows"`
	Extras     CameraCostExtras `json:"extras"`
	Discount   float64          `json:"discount"` // مقدار الخصم
}

// CameraCostRowResult تفصيل حساب صف كاميرا واحد (كل طبقة على حدة).
type CameraCostRowResult struct {
	Index            int     `json:"index"`
	BasePrice        float64 `json:"basePrice"`        // J: شريحة العادي + 1.2 × شريحة VIP
	PlaceMultiplier  float64 `json:"placeMultiplier"`  // K
	AfterPlace       float64 `json:"afterPlace"`       // K
	SystemMultiplier float64 `json:"systemMultiplier"` // L
	AfterSystem      float64 `json:"afterSystem"`      // L
	HeightMultiplier float64 `json:"heightMultiplier"` // M
	Total            float64 `json:"total"`            // M
	CountsAsCamera   bool    `json:"countsAsCamera"`   // I: 1 إذا اكو طول كيبل
}

// CameraCostResponse نتيجة حساب استمارة الكاميرات كاملة.
type CameraCostResponse struct {
	Rows         []CameraCostRowResult `json:"rows"`
	CameraCount  int                   `json:"cameraCount"`  // H3
	CamerasTotal float64               `json:"camerasTotal"` // M18
	ExtrasTotal  float64               `json:"extrasTotal"`  // مجموع الأعمال الإضافية
	Discount     float64               `json:"discount"`
	FinalAmount  float64               `json:"finalAmount"` // D18
	Note         string                `json:"note"`
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
