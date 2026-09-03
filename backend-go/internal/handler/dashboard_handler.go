package handler

import (
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// DashboardHandler أرقام اللوحة الرئيسية.
//
// قبله كانت اللوحة تسحب كل الموظفين وكل العملاء وكل الحجوزات (ميغابايتات)
// عشان تعرض أربع أرقام. هذا يعني إن أي موظف يفتح الصفحة الرئيسية ينزل
// على جهازه أرشيف الشركة كامل ويقدر يقراه من F12.
//
// هذا المسار يرجّع الأرقام بس — بدون ولا سطر بيانات.
type DashboardHandler struct {
	db *sqlx.DB
}

func NewDashboardHandler(db *sqlx.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

type dashboardSummary struct {
	EmployeeCount  int `db:"employeeCount" json:"employeeCount"`
	CustomerCount  int `db:"customerCount" json:"customerCount"`
	BookingCount   int `db:"bookingCount" json:"bookingCount"`
	GpsDeviceCount int `db:"gpsDeviceCount" json:"gpsDeviceCount"`
}

// GET /api/dashboard/summary
func (h *DashboardHandler) Summary(w http.ResponseWriter, r *http.Request) {
	var s dashboardSummary
	err := h.db.Get(&s, `
		SELECT
			(SELECT COUNT(*) FROM "Employee" WHERE status = 'ACTIVE')  AS "employeeCount",
			(SELECT COUNT(*) FROM "Customer")                          AS "customerCount",
			(SELECT COUNT(*) FROM "Booking")                           AS "bookingCount",
			(SELECT COUNT(*) FROM "GpsDeviceRequest")                  AS "gpsDeviceCount"`)
	if err != nil {
		log.Printf("dashboard summary: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب ملخص اللوحة")
		return
	}
	WriteJSON(w, http.StatusOK, s)
}

// ═══ ملخّص المحاسب والمراقب ═══
//
// لوحة المحاسب ولوحة المراقب جانت تنزّل **كل الحجوزات المنجزة من يوم ما
// انفتح النظام** — بكل زبائنها وخدماتها وتعييناتها وسلة موادها — بس
// حتى تجمع تسع أرقام بالمتصفح. يعني كل ما تخلص شغلة جديدة، صفحتهم
// الرئيسية تصير أثقل، والثقل ما يوكف لأن الأرشيف ما ينقص.
//
// الجمع مكانه قاعدة البيانات. هذا المسار يرجّع نفس الأرقام بالضبط —
// مجاميع كاملة مو مقصوصة — بعشرات البايتات بدل ميغابايتات.
type financeSummary struct {
	CompletedCount    int     `db:"completedCount" json:"completedCount"`
	UnverifiedCount   int     `db:"unverifiedCount" json:"unverifiedCount"`
	VerifiedCount     int     `db:"verifiedCount" json:"verifiedCount"`
	TodayCompleted    int     `db:"todayCompleted" json:"todayCompleted"`
	TodayCreated      int     `db:"todayCreated" json:"todayCreated"`
	PendingCount      int     `db:"pendingCount" json:"pendingCount"`
	ConfirmedCount    int     `db:"confirmedCount" json:"confirmedCount"`
	InProgressCount   int     `db:"inProgressCount" json:"inProgressCount"`
	ActiveCrewCount   int     `db:"activeCrewCount" json:"activeCrewCount"`
	TotalCollected    float64 `db:"totalCollected" json:"totalCollected"`
	UnverifiedAmount  float64 `db:"unverifiedAmount" json:"unverifiedAmount"`
	VerifiedAmount    float64 `db:"verifiedAmount" json:"verifiedAmount"`
	TotalQuoted       float64 `db:"totalQuoted" json:"totalQuoted"`
	TotalCartValue    float64 `db:"totalCartValue" json:"totalCartValue"`
	PendingExpenses   int     `db:"pendingExpenses" json:"pendingExpenses"`
	ApprovedExpenses  int     `db:"approvedExpenses" json:"approvedExpenses"`
	TotalExpenseValue float64 `db:"totalExpenseValue" json:"totalExpenseValue"`
}

// GET /api/dashboard/finance-summary
func (h *DashboardHandler) FinanceSummary(w http.ResponseWriter, r *http.Request) {
	var s financeSummary
	err := h.db.Get(&s, `
		SELECT
			COUNT(*) FILTER (WHERE b.status = 'COMPLETED')                              AS "completedCount",
			COUNT(*) FILTER (WHERE b.status = 'COMPLETED' AND NOT b."amountVerified")   AS "unverifiedCount",
			COUNT(*) FILTER (WHERE b.status = 'COMPLETED' AND b."amountVerified")       AS "verifiedCount",
			COUNT(*) FILTER (WHERE b.status = 'COMPLETED'
			-- ⚠️ چان ::date = CURRENT_DATE — يعني «اليوم» بنظر الخادم
			-- مو بنظر بغداد. حجز ينخلص ١٢:٣٠ ليلاً بغداد چان ينحسب
			-- **أمس** هنا و**اليوم** بشاشة التدقيق اليومي (الي تستعمل
			-- baghdad_date)، فالشاشتان تختلفان على نفس الحجز.
			                 AND baghdad_date(b."completedAt") = baghdad_today())       AS "todayCompleted",
			-- «انفتحن اليوم»: چانت تنعدّ بالواجهة من قائمة مقصوصة عند
			-- ٢٠٠ حجز مكتمل، فأي حجز اليوم برّا أحدث ٢٠٠ ما ينعدّ
			-- والرقم يطلع أقل من الحقيقة بلا ما يبيّن.
			COUNT(*) FILTER (WHERE baghdad_date(b."createdAt") = baghdad_today())       AS "todayCreated",
			COUNT(*) FILTER (WHERE b.status = 'PENDING')                                AS "pendingCount",
			COUNT(*) FILTER (WHERE b.status = 'CONFIRMED')                              AS "confirmedCount",
			COUNT(*) FILTER (WHERE b.status = 'IN_PROGRESS')                            AS "inProgressCount",
			(SELECT COUNT(DISTINCT ba."employeeId") FROM "BookingAssignment" ba
			  JOIN "Booking" b2 ON b2.id = ba."bookingId"
			  WHERE b2.status = 'IN_PROGRESS')                                          AS "activeCrewCount",
			COALESCE(SUM(COALESCE(b."amountCollected",0) + COALESCE(b."advancePaid",0))
			         FILTER (WHERE b.status = 'COMPLETED'), 0)                          AS "totalCollected",
			-- ⚠️ «غير مدققة»/«مدققة» بنفس نطاق totalCollected أعلاه
			-- (COMPLETED فعلاً، شامل المقدم) — لا تحسب Booking.amountCollected
			-- الخام بلا فلترة حالة (هذاك مصدر رقم مختلف بـstats_repository.go
			-- كان يعطي رقماً غير هذا لنفس المفهوم).
			COALESCE(SUM(COALESCE(b."amountCollected",0) + COALESCE(b."advancePaid",0))
			         FILTER (WHERE b.status = 'COMPLETED' AND NOT b."amountVerified"), 0) AS "unverifiedAmount",
			COALESCE(SUM(COALESCE(b."amountCollected",0) + COALESCE(b."advancePaid",0))
			         FILTER (WHERE b.status = 'COMPLETED' AND b."amountVerified"), 0)     AS "verifiedAmount",
			COALESCE(SUM(COALESCE(b."quotedPrice",0))
			         FILTER (WHERE b.status = 'COMPLETED'), 0)                          AS "totalQuoted",
			(SELECT COALESCE(SUM(ci."totalPrice"), 0) FROM "CartItem" ci
			  JOIN "Booking" b3 ON b3.id = ci."bookingId"
			  WHERE b3.status = 'COMPLETED')                                            AS "totalCartValue",
			(SELECT COUNT(*) FROM "Expense" WHERE status = 'PENDING')                   AS "pendingExpenses",
			(SELECT COUNT(*) FROM "Expense" WHERE status = 'APPROVED')                  AS "approvedExpenses",
			(SELECT COALESCE(SUM(amount),0) FROM "Expense" WHERE status = 'APPROVED')   AS "totalExpenseValue"
		FROM "Booking" b`)
	if err != nil {
		log.Printf("finance summary: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الملخص المالي")
		return
	}
	WriteJSON(w, http.StatusOK, s)
}

// ═══ نبض اليوم — الواجهة الرئيسية للإداري ═══
//
// الشاشة الرئيسية جانت أربع أزرار وبس. الإداري يفتح النظام الصبح وما
// يشوف ولا رقم: كم حجز عدنا اليوم؟ منو بالميدان هسه؟ أكو شكوى جديدة؟
// فيضطر يفتح أربع شاشات واحدة واحدة حتى يعرف وضع يومه.
//
// ⚠️ كلها مجاميع بقاعدة البيانات — مو تنزيل الحجوزات للمتصفح وجمعها
// هناك. نفس درس ملخّص المحاسب: الجمع بالمتصفح يثقل كل يوم مع تراكم
// الأرشيف، وما يوكف لأن الأرشيف ما ينقص.
//
// ⚠️ «حجوزات اليوم» = المجدولة اليوم فعلاً (scheduledAt)، مو COUNT(*)
// على كل الجدول. اللافتة القديمة جانت تعرض المجموع التاريخي وتسميه
// «حجوزات اليوم» — رقم أكبر بمرات من الحقيقة.
type todayPulse struct {
	TodayBookings     int `db:"todayBookings"  json:"todayBookings"`
	YesterdayBookings int `db:"yesterdayBookings" json:"yesterdayBookings"`
	OpenMissions      int `db:"openMissions"   json:"openMissions"`
	NewComplaints     int `db:"newComplaints"  json:"newComplaints"`
	NeedsCoordination int `db:"needsCoordination" json:"needsCoordination"`
	CrewInField       int `db:"crewInField"    json:"crewInField"`
	OverdueMissions   int `db:"overdueMissions" json:"overdueMissions"`
}

// GET /api/dashboard/today-pulse
func (h *DashboardHandler) TodayPulse(w http.ResponseWriter, r *http.Request) {
	var s todayPulse
	err := h.db.Get(&s, `
		SELECT
			(SELECT COUNT(*) FROM "Booking"
			  WHERE "archivedAt" IS NULL AND "scheduledAt"::date = CURRENT_DATE)     AS "todayBookings",
			-- أمس للمقارنة: رقم بلا مرجع ما يگول شي. ١٢ حجز زين لو خبل؟
			(SELECT COUNT(*) FROM "Booking"
			  WHERE "archivedAt" IS NULL
			    AND "scheduledAt"::date = CURRENT_DATE - 1)                          AS "yesterdayBookings",
			-- المهام المفتوحة: الي لسه بالميدان، مو المنجزة ولا المتوقفة
			(SELECT COUNT(*) FROM "Mission"
			  WHERE stage NOT IN ('COMPLETED','STOPPED'))                            AS "openMissions",
			(SELECT COUNT(*) FROM "Complaint" WHERE status = 'NEW')                  AS "newComplaints",
			-- مثبّت بس بلا موعد أو بلا كادر — هذا الي يحتاج شغل الإداري
			(SELECT COUNT(*) FROM "Booking" b
			  WHERE b."archivedAt" IS NULL AND b.status = 'CONFIRMED'
			    AND (b."scheduledAt" IS NULL
			         OR NOT EXISTS (SELECT 1 FROM "BookingAssignment" ba
			                         WHERE ba."bookingId" = b.id)))                  AS "needsCoordination",
			-- كوادر بالميدان هسه: طلعوا وما وصلوا/خلصوا
			(SELECT COUNT(DISTINCT m."leaderId") FROM "Mission" m
			  WHERE m.stage IN ('EN_ROUTE','ARRIVED','WORK_STARTED'))                AS "crewInField",
			-- تأخر ميداني: مهمة مفتوحة من أكثر من يوم بلا إنجاز
			(SELECT COUNT(*) FROM "Mission"
			  WHERE stage NOT IN ('COMPLETED','STOPPED')
			    AND "assignedAt" < now() - interval '24 hours')                      AS "overdueMissions"`)
	if err != nil {
		log.Printf("today pulse: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب نبض اليوم")
		return
	}
	WriteJSON(w, http.StatusOK, s)
}
