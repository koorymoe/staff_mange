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
	// ملاحظة قصيرة للبند («على باب المكتب الداخلي»). ما تدخل بالحساب أبداً —
	// بس الفني يحتاجها لما يوصل الموقع، والمحاسب يحتاجها لما يدقّق سطر
	// ما يفهم ليش سعره هيج. تنحفظ بنفس عمود JSON مثل بقية الحقول، فما
	// تحتاج هجرة ولا تكسر الفواتير القديمة (تنقرأ فارغة).
	Notes string `json:"notes"`
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
	ID               string  `db:"id" json:"id"`
	BookingID        *string `db:"bookingId" json:"bookingId"`
	EmployeeID       string  `db:"employeeId" json:"employeeId"`
	CustomerName     *string `db:"customerName" json:"customerName"`
	CustomerPhone    *string `db:"customerPhone" json:"customerPhone"`
	CustomerAddress  *string `db:"customerAddress" json:"customerAddress"`
	SystemsJSON      string  `db:"systems" json:"-"`
	ItemsJSON        string  `db:"items" json:"-"`
	TotalDeviceCount int     `db:"totalDeviceCount" json:"totalDeviceCount"`
	ExecutionCost    float64 `db:"executionCost" json:"executionCost"`
	MaterialsTotal   float64 `db:"materialsTotal" json:"materialsTotal"`
	DiscountValue    float64 `db:"discountValue" json:"discountValue"`
	NetTotal         float64 `db:"netTotal" json:"netTotal"`
	// ── الشغل المجاني ──
	// ⚠️ أعمدة بالجدول → لازم حقول هنا (الجلب SELECT *).
	IsFree         bool    `db:"isFree" json:"isFree"`
	FreeReasonID   *string `db:"freeReasonId" json:"freeReasonId"`
	FreeReasonNote *string `db:"freeReasonNote" json:"freeReasonNote"`
	// اسم السبب للعرض — ينجلب بالربط مو من الجدول
	FreeReasonLabel      *string    `db:"-" json:"freeReasonLabel"`
	// ═══ الكلفة اليدوية ═══
	//
	// ⚠️ `PricingMode` يميّز الفاتورة **للأبد**: بلاه تندسّ الفاتورة
	// اليدوية بين فواتير الجدول، وما يعرف المحاسب ليش سعرها هيج.
	// ⚠️ أعمدة بالجدول → لازم حقول هنا (الجلب SELECT *).
	PricingMode     string  `db:"pricingMode" json:"pricingMode"`
	ManualWork      *string `db:"manualWork" json:"manualWork,omitempty"`
	ManualPriceNote *string `db:"manualPriceNote" json:"manualPriceNote,omitempty"`

	AccountingCode       string     `db:"accountingCode" json:"accountingCode"`
	Status               string     `db:"status" json:"status"` // SUBMITTED | APPROVED
	CreatedAt            time.Time  `db:"createdAt" json:"createdAt"`
	ApprovedByEmployeeID *string    `db:"approvedByEmployeeId" json:"approvedByEmployeeId"`
	ApprovedAt           *time.Time `db:"approvedAt" json:"approvedAt"`
	// رقم الفاتورة المحاسبية الصادرة من نظام المحاسب الخارجي، وتاريخ
	// ربطه. إجباري وقت الاعتماد — بدونه ما ينربط الخيط بين النظامين.
	// ⚠️ أعمدة بالجدول → لازم حقول هنا (SELECT *).
	ExternalInvoiceNumber *string    `db:"externalInvoiceNumber" json:"externalInvoiceNumber"`
	ExternalInvoiceAt     *time.Time `db:"externalInvoiceAt" json:"externalInvoiceAt"`
	// تعديل المحاسب على المبالغ: سببه ووقته. ⚠️ أعمدة بالجدول.
	AdjustedReason *string `db:"adjustedReason" json:"adjustedReason"`

	// ═══ مراجعة المراقب ═══
	//
	// ⚠️ الحالة تنشتق من الطوابع: `monitorRequestedAt` موجود و
	// `monitorDecidedAt` فارغ = الفاتورة **عند المراقب الآن**. عمود
	// «مرحلة» منفصل يعني حقيقتين لنفس الشي.
	MonitorRequestedAt   *time.Time `db:"monitorRequestedAt" json:"monitorRequestedAt,omitempty"`
	MonitorRequestedByID *string    `db:"monitorRequestedById" json:"monitorRequestedById,omitempty"`
	MonitorRequestNote   *string    `db:"monitorRequestNote" json:"monitorRequestNote,omitempty"`
	MonitorDecidedAt     *time.Time `db:"monitorDecidedAt" json:"monitorDecidedAt,omitempty"`
	MonitorDecidedByID   *string    `db:"monitorDecidedById" json:"monitorDecidedById,omitempty"`
	MonitorVerdict       *string    `db:"monitorVerdict" json:"monitorVerdict,omitempty"`
	MonitorNote          *string    `db:"monitorNote" json:"monitorNote,omitempty"`

	// ═══ إرجاع المالك للمحاسب ═══
	ReturnedAt    *time.Time `db:"returnedAt" json:"returnedAt,omitempty"`
	ReturnedByID  *string    `db:"returnedById" json:"returnedById,omitempty"`
	ReturnReason  *string    `db:"returnReason" json:"returnReason,omitempty"`
	ReturnedCount int        `db:"returnedCount" json:"returnedCount"`

	// أسماء للعرض — ما تنخزن
	MonitorRequestedByName *string `db:"-" json:"monitorRequestedByName,omitempty"`
	MonitorDecidedByName   *string `db:"-" json:"monitorDecidedByName,omitempty"`
	ReturnedByName         *string `db:"-" json:"returnedByName,omitempty"`

	// ═══ التدقيق ═══
	// حكم المحاسب قبل الاعتماد: مطابق / غير مطابق / خطأ بالسعر.
	// ⚠️ أعمدة بالجدول → لازم حقول هنا (SELECT *).
	AuditVerdict  *string    `db:"auditVerdict" json:"auditVerdict,omitempty"`
	AuditNote     *string    `db:"auditNote" json:"auditNote,omitempty"`
	AuditedByID   *string    `db:"auditedById" json:"auditedById,omitempty"`
	AuditedAt     *time.Time `db:"auditedAt" json:"auditedAt,omitempty"`
	AuditedAmount *float64   `db:"auditedAmount" json:"auditedAmount,omitempty"`

	// ═══ سحب الاعتماد ═══
	// «لازم تخليلي خيار أكدر أرجعله الفواتير الما معتمدة».
	// ⚠️ ما ينمحي شي: الرقم القديم وسبب السحب يبقون بالسجل.
	RevokedAt    *time.Time `db:"revokedAt" json:"revokedAt,omitempty"`
	RevokedByID  *string    `db:"revokedById" json:"revokedById,omitempty"`
	RevokeReason *string    `db:"revokeReason" json:"revokeReason,omitempty"`
	RevokedCount int        `db:"revokedCount" json:"revokedCount"`

	AuditedByName *string    `db:"-" json:"auditedByName,omitempty"`
	RevokedByName *string    `db:"-" json:"revokedByName,omitempty"`
	AdjustedAt    *time.Time `db:"adjustedAt" json:"adjustedAt"`

	// تفاصيل يحتاجها المحاسب: منو الليدر الي رفعها، ومنو اعتمدها،
	// وأي حجز تخص. تنعبّى بالتهيئة مو من الجدول.
	EmployeeName   string  `db:"-" json:"employeeName"`
	EmployeeRole   string  `db:"-" json:"employeeRole"`
	EmployeePhone  *string `db:"-" json:"employeePhone"`
	ApprovedByName *string `db:"-" json:"approvedByName"`
	BookingCode    *string `db:"-" json:"bookingCode"`
	// ملخص الحجز المرتبط — العنوان والخدمة والمبالغ والكادر المنفّذ
	Booking *Booking `db:"-" json:"booking"`

	Systems   []string                    `db:"-" json:"systems"`
	Items     []ExecutionCostItem         `db:"-" json:"items"`
	Materials []LeaderInvoiceMaterialItem `db:"-" json:"materials"`
}

// CreateLeaderInvoiceRequest جسم طلب إنشاء فاتورة ليدر جديدة.
type CreateLeaderInvoiceRequest struct {
	BookingID *string `json:"bookingId"`
	// الشغل المجاني: الفاتورة تنسوّى بصفر بسبب من القائمة
	IsFree          bool                        `json:"isFree"`
	FreeReasonID    *string                     `json:"freeReasonId"`
	FreeReasonNote  *string                     `json:"freeReasonNote"`
	CustomerName    *string                     `json:"customerName"`
	CustomerPhone   *string                     `json:"customerPhone"`
	CustomerAddress *string                     `json:"customerAddress"`
	Systems         []string                    `json:"systems"` // حتى 3 منظومات
	Items           []ExecutionCostItem         `json:"items"`
	Materials       []CreateMaterialLineRequest `json:"materials"`
	DiscountValue   float64                     `json:"discountValue"`
}

// ═══ فاتورة خدمة بسعر يدوي (جي بي اس / داش كام) ═══
//
// «هذا ما يرادله تيم وليدر، يرادله فني واحد فقط… والي يسوّي الفاتورة
// هو مسؤول الخدمة نفسها، واني أخلي السعر بكيفي».
//
// ⚠️ **نوع منفصل مو علم على فاتورة الليدر**: فاتورة الليدر تلزم
// منظومات وبنود تنفيذ وسعرها ينحسب بالسيرفر — وهذا مقصود ويبقى.
// خلط الاثنين بمسار واحد يفتح باب «فاتورة بسعر بالإيد» لكل أحد،
// ويفرّغ حساب الكلفة من معناه.
type CreateServiceInvoiceRequest struct {
	BookingID       *string  `json:"bookingId"`
	Kind            string   `json:"kind"` // GPS | DASHCAM
	CustomerName    *string  `json:"customerName"`
	CustomerPhone   *string  `json:"customerPhone"`
	CustomerAddress *string  `json:"customerAddress"`
	// السعر يحطّه مسؤول الخدمة — ماكو جدول كلفة لهذي الخدمات.
	Price float64 `json:"price"`
	Note  *string `json:"note"`
}

// أنماط التسعير.
const (
	// PricingModeCatalog السعر انحسب من جدول الكلفة — الوضع الطبيعي.
	PricingModeCatalog = "CATALOG"
	// PricingModeManual السعر كتبه صاحب الصلاحية بإيده.
	PricingModeManual = "MANUAL"
)

// PricingModeLabel التسمية العربية — بيانات لا شرط مكرَّر بكل شاشة.
var PricingModeLabel = map[string]string{
	PricingModeCatalog: "من جدول الكلفة",
	PricingModeManual:  "كلفة يدوية",
}

// PermissionInvoiceManual الصلاحية الي تخوّل الفاتورة اليدوية.
//
// ⚠️ **معزولة وتنمنح فرد-فرد**: سعر حر يشيل الحارس الوحيد على
// التسعير، فما تنضاف لأي دور افتراضي — ولا حتى لدور الليدر.
const PermissionInvoiceManual = "invoice_manual"

// ManualWorkMinRunes أقصر وصف مقبول لشنو انعمل.
//
// ⚠️ **الوصف مو تزيين**: هو **المرجع الوحيد** الي يقدر المحاسب
// يدقّق بيه سعراً ما جا من جدول. «صيانة» ما تنفع مرجعاً بعد شهر.
const ManualWorkMinRunes = 10

// CreateManualInvoiceRequest فاتورة بكلفة يدوية: الليدر يكتب شنو
// اشتغل للزبون، ويحطّ السعر بنفسه.
type CreateManualInvoiceRequest struct {
	BookingID       *string  `json:"bookingId"`
	CustomerName    *string  `json:"customerName"`
	CustomerPhone   *string  `json:"customerPhone"`
	CustomerAddress *string  `json:"customerAddress"`
	// Work شنو انعمل للزبون بالضبط — إجباري.
	Work string `json:"work"`
	// Price السعر الي يريده — إجباري وأكبر من صفر.
	Price float64 `json:"price"`
	// Systems المنظومات المشمولة (اختيارية) — للعرض والتصنيف بس،
	// ما تدخل بأي حساب.
	Systems []string `json:"systems"`
	Note    *string  `json:"note"`
}

const (
	ServiceInvoiceGps     = "GPS"
	ServiceInvoiceDashcam = "DASHCAM"
)

// ServiceInvoiceKindLabel التسمية العربية — بيانات لا شرط مكرَّر.
var ServiceInvoiceKindLabel = map[string]string{
	ServiceInvoiceGps:     "فاتورة الجي بي اس",
	ServiceInvoiceDashcam: "فاتورة الداش كام",
}

// ServiceInvoicePermission الصلاحية الي تخوّل كل نوع — «صلاحيتين
// أنطيها للشخص الي يعجبني»، فكل خدمة معزولة بصلاحيتها.
var ServiceInvoicePermission = map[string]string{
	ServiceInvoiceGps:     "invoice_gps",
	ServiceInvoiceDashcam: "invoice_dashcam",
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
	ScreenLarge43Count  int     `json:"screenLarge43Count"`   // تثبيت شاشة 43 وأكبر × 15000
	ScreenSmall43Count  int     `json:"screenSmall43Count"`   // تثبيت شاشة أصغر من 43 × 7500
	RackCount           int     `json:"rackCount"`            // تثبيت راك × 15000
	BoardCount          int     `json:"boardCount"`           // تثبيت بورد × 7500
	IpCameraChangeCount int     `json:"ipCameraChangeCount"`  // تغيير أو إضافة IP كاميرا × 15000
	VipInternetMeters   int     `json:"vipInternetMeters"`    // مد كيبل انترنيت VIP × 400
	NormalInternetM     int     `json:"normalInternetMeters"` // مد كيبل انترنيت عادي × 200
	ProgrammingAmount   float64 `json:"programmingAmount"`    // برمجة — مبلغ يُدخل يدوي
	OtherAmount         float64 `json:"otherAmount"`          // غيرها — مبلغ يُدخل يدوي
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

// ApproveLeaderInvoiceRequest اعتماد فاتورة الليدر — رقم الفاتورة
// المحاسبية إجباري.
type ApproveLeaderInvoiceRequest struct {
	ExternalInvoiceNumber string `json:"externalInvoiceNumber"`
}

// SetExternalNumberRequest ربط رقم فاتورة محاسبية بفاتورة معتمدة أصلاً.
type SetExternalNumberRequest struct {
	ExternalInvoiceNumber string `json:"externalInvoiceNumber"`
}

// AdjustLeaderInvoiceRequest تعديل المحاسب على مبالغ الفاتورة.
type AdjustLeaderInvoiceRequest struct {
	ExecutionCost  float64 `json:"executionCost"`
	MaterialsTotal float64 `json:"materialsTotal"`
	DiscountValue  float64 `json:"discountValue"`
	Reason         string  `json:"reason"`
}

// FreeWorkReason سبب من قائمة الشغل المجاني.
//
// قائمة مو نص حر: النص الحر ما ينجمّع ولا ينحسب. لما يكون من قائمة
// نقدر نجاوب «شكد كلّفنا الضمان هالسنة؟» بسؤال واحد.
type FreeWorkReason struct {
	ID        string `db:"id" json:"id"`
	Label     string `db:"label" json:"label"`
	SortOrder int    `db:"sortOrder" json:"sortOrder"`
	Active    bool   `db:"active" json:"active"`
	NeedsNote bool   `db:"needsNote" json:"needsNote"`
}

// LeaderInvoiceAdjustment سطر واحد بسجل تعديلات المحاسب على الفاتورة.
//
// نحفظ المبالغ الأربعة **قبل وبعد** حتى لو ما انتغيّر إلا واحد —
// المقارنة لازم تكون كاملة، والعرض هو الي يخفي الي ما انتغيّر.
type LeaderInvoiceAdjustment struct {
	ID        string `db:"id" json:"id"`
	InvoiceID string `db:"invoiceId" json:"invoiceId"`

	OldExecutionCost  float64 `db:"oldExecutionCost" json:"oldExecutionCost"`
	NewExecutionCost  float64 `db:"newExecutionCost" json:"newExecutionCost"`
	OldMaterialsTotal float64 `db:"oldMaterialsTotal" json:"oldMaterialsTotal"`
	NewMaterialsTotal float64 `db:"newMaterialsTotal" json:"newMaterialsTotal"`
	OldDiscountValue  float64 `db:"oldDiscountValue" json:"oldDiscountValue"`
	NewDiscountValue  float64 `db:"newDiscountValue" json:"newDiscountValue"`
	OldNetTotal       float64 `db:"oldNetTotal" json:"oldNetTotal"`
	NewNetTotal       float64 `db:"newNetTotal" json:"newNetTotal"`

	Reason       string    `db:"reason" json:"reason"`
	AdjustedByID *string   `db:"adjustedById" json:"adjustedById"`
	CreatedAt    time.Time `db:"createdAt" json:"createdAt"`

	// اسم المحاسب — ينجي بالربط مو من الجدول
	AdjustedByName *string `db:"adjustedByName" json:"adjustedByName"`
}

// ═══ أحكام التدقيق ═══
//
// «بالتدقيق: مطابق / غير مطابق / خطأ بالسعر».
//
//	مطابق      = سعر الفاتورة نفسه المبلغ الي داخل
//	غير مطابق  = الفاتورة سعرها يختلف عن المبلغ الداخل
//	خطأ بالسعر = الموظف غلط، جاب أعلى من الفاتورة أو أوطى
const (
	AuditVerdictMatched    = "MATCHED"
	AuditVerdictMismatch   = "MISMATCH"
	AuditVerdictPriceError = "PRICE_ERROR"
)

func AuditVerdictLabel(v string) string {
	switch v {
	case AuditVerdictMatched:
		return "مطابق"
	case AuditVerdictMismatch:
		return "غير مطابق"
	case AuditVerdictPriceError:
		return "خطأ بالسعر"
	}
	return v
}

func ValidAuditVerdict(v string) bool {
	return v == AuditVerdictMatched || v == AuditVerdictMismatch || v == AuditVerdictPriceError
}

// AuditVerdictRequest حكم المحاسب على الفاتورة.
//
// ⚠️ الملاحظة إجبارية بغير المطابق وبخطأ السعر: «غير مطابق» بلا شرح
// ما تفيد لا المراقب ولا المدير — والاثنين يقرونها.
type AuditVerdictRequest struct {
	Verdict string   `json:"verdict"`
	Note    string   `json:"note"`
	Amount  *float64 `json:"auditedAmount"`
}

// RevokeApprovalRequest سحب اعتماد فاتورة انعتمدت بالغلط.
type RevokeApprovalRequest struct {
	Reason string `json:"reason"`
}
