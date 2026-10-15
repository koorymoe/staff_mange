package handler

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/repository"
)

// ═══ فحص دقة ماتركس — قرار (ع) 10-15 ═══
// «اريد فحص لصحة ودقة تدقيقات ماتركس وحساباته». كل فحص يشتغل على البيانات
// الحقيقية ويطلع ✓ أو ✗ ويّا عدد الحالات وأمثلة تودّي للمصدر. يشتغل وقت فتح
// الشاشة، وكل يوم بالخلفية — وإذا فشل فحص يوصل إشعار للمالك.

type MatrixSelfCheckHandler struct {
	db     *sqlx.DB
	notify func(string)
}

func NewMatrixSelfCheckHandler(db *sqlx.DB, notify func(string)) *MatrixSelfCheckHandler {
	return &MatrixSelfCheckHandler{db: db, notify: notify}
}

type selfCheck struct {
	Key      string   `json:"key"`
	Area     string   `json:"area"`
	Title    string   `json:"title"`
	Why      string   `json:"why"`
	Severity string   `json:"severity"` // ERROR = حساب غلط · WARN = بيانات تحتاج مراجعة
	Count    int      `json:"count"`
	OK       bool     `json:"ok"`
	Examples []string `json:"examples"`
	Link     string   `json:"link,omitempty"`
}

type checkDef struct {
	key, area, title, why, severity, link string
	// SQL يرجّع صفوف (label) — كل صف حالة غلط
	sql string
}

var selfChecks = []checkDef{
	{"score_cancelled", "التقييم", "نقاط على حجوزات ملغية أو مؤرشفة", "حجز انلغى ما لازم يحاسب أحد", "ERROR", "/matrix/score",
		`SELECT 'حجز ' || b.code || ' · ' || e.name || ' · ' || m.rule FROM "MatrixScore" m JOIN "Booking" b ON b.id = m."sourceId" JOIN "Employee" e ON e.id = m."employeeId"
		 WHERE m.source = 'BOOKING' AND m."cancelledAt" IS NULL AND (b.status::text = 'CANCELLED' OR b."archivedAt" IS NOT NULL)`},
	{"score_admin", "التقييم", "نقاط على المالك أو المدير", "المالك والمدير ما ينقيّمون", "ERROR", "",
		`SELECT e.name || ' · ' || m.rule FROM "MatrixScore" m JOIN "Employee" e ON e.id = m."employeeId"
		 WHERE m."cancelledAt" IS NULL AND e.role::text IN ('ADMIN', 'OWNER')`},
	{"score_no_reason", "التقييم", "نقاط بلا سبب مكتوب", "كل نقطة لازم يبين سببها للموظف", "ERROR", "",
		`SELECT e.name || ' · ' || m.rule FROM "MatrixScore" m JOIN "Employee" e ON e.id = m."employeeId" WHERE m."cancelledAt" IS NULL AND btrim(m.reason) = ''`},
	{"score_inactive", "التقييم", "نقاط جديدة لموظف مو فعّال", "الموظف الموقوف ما ينحسب", "WARN", "",
		`SELECT DISTINCT e.name FROM "MatrixScore" m JOIN "Employee" e ON e.id = m."employeeId"
		 WHERE m."cancelledAt" IS NULL AND m.at > now() - interval '30 days' AND e.status::text <> 'ACTIVE'`},
	{"score_range", "التقييم", "نقاط خارج المدى (أكثر من الحد أو بالسالب)", "النقطة لازم بين صفر والحد الأعلى", "ERROR", "",
		`SELECT e.name || ' · ' || m.rule || ' = ' || m.points || '/' || m."maxPoints" FROM "MatrixScore" m JOIN "Employee" e ON e.id = m."employeeId"
		 WHERE m."cancelledAt" IS NULL AND (m.points < 0 OR m.points > m."maxPoints")`},
	{"stage_order", "المراحل", "مراحل بترتيب مستحيل", "الإنجاز قبل البدء، أو البدء قبل الوصول، أو الوصول قبل الانطلاق", "WARN", "/bookings",
		`SELECT 'حجز ' || b.code FROM "Booking" b WHERE b."archivedAt" IS NULL AND (
		   (b."completedAt" IS NOT NULL AND b."startedAt" IS NOT NULL AND b."completedAt" < b."startedAt" - interval '1 minute')
		   OR (b."startedAt" IS NOT NULL AND b."arrivedAt" IS NOT NULL AND b."startedAt" < b."arrivedAt" - interval '1 minute')
		   OR (b."arrivedAt" IS NOT NULL AND b."departedAt" IS NOT NULL AND b."arrivedAt" < b."departedAt" - interval '1 minute'))`},
	{"duration_outlier", "المراحل", "مدة شغل مستبعدة من المتوسطات (آخر ٩٠ يوم)", "تسجيل بنفس اللحظة أو تسجيل جماعي أو أطول من ١٦ ساعة — ماتركس يستبعدها حتى ما تخرّب المتوسط", "WARN", "/bookings",
		`SELECT 'حجز ' || b.code || ' · ' || round(EXTRACT(EPOCH FROM (b."completedAt" - b."startedAt")) / 60) || ' دقيقة' FROM "Booking" b
		 WHERE b.status::text = 'COMPLETED' AND b."startedAt" IS NOT NULL AND b."completedAt" > now() - interval '90 days'
		   AND booking_work_minutes(b.id, b."startedAt", b."completedAt", b."autoArrivedAt", b."autoCompletedAt") IS NULL`},
	{"service_median", "المتوسطات", "خدمات وسيط مدتها أقل من ١٥ دقيقة", "وسيط قليل يعني تسجيل بنفس اللحظة — يخلّي سرعة الموظفين تطلع غلط", "WARN", "",
		`SELECT s.name || ' · ' || round(percentile_cont(0.5) WITHIN GROUP (ORDER BY booking_work_minutes(b.id, b."startedAt", b."completedAt", b."autoArrivedAt", b."autoCompletedAt"))) || ' دقيقة'
		 FROM "Booking" b JOIN "Service" s ON s.id = b."serviceId" WHERE b."completedAt" > now() - interval '180 days'
		 GROUP BY s.name HAVING count(booking_work_minutes(b.id, b."startedAt", b."completedAt", b."autoArrivedAt", b."autoCompletedAt")) >= 3
		   AND percentile_cont(0.5) WITHIN GROUP (ORDER BY booking_work_minutes(b.id, b."startedAt", b."completedAt", b."autoArrivedAt", b."autoCompletedAt")) < 15`},
	{"late_data", "المتأخرات", "حجوزات محسوبة متأخرة وبياناتها مشكوكة (آخر ٣٠ يوم)", "الموعد انغيّر بعد الإنجاز، أو انسجّل منجز بلا ولا مرحلة — ممكن تنحسب متأخرة وهي مو", "WARN", "/late-audit",
		`SELECT 'حجز ' || b.code FROM "Booking" b WHERE baghdad_date(b."scheduledAt") >= baghdad_today() - 30
		   AND b.status::text = 'COMPLETED' AND b."archivedAt" IS NULL AND baghdad_date(b."completedAt") > baghdad_date(b."scheduledAt")
		   AND (EXISTS (SELECT 1 FROM "ScheduleChangeLog" l WHERE l."bookingId" = b.id AND l."createdAt" > b."completedAt")
		        OR (b."startedAt" IS NULL AND b."arrivedAt" IS NULL))`},
	{"proof_far", "أدلة المراحل", "ضغطات معلّمة «بعيد» ومسافتها أقل من كيلو", "حكم «بعيد» لازم يطابق المسافة المحسوبة", "ERROR", "/field-map",
		`SELECT 'حجز ' || b.code || ' · ' || round(p."distanceM") || ' م' FROM "BookingStageProof" p JOIN "Booking" b ON b.id = p."bookingId"
		 WHERE p.verdict = 'FAR' AND p."distanceM" IS NOT NULL AND p."distanceM" <= 1000`},
	{"purchase_loss", "المشتريات", "تنبيهات خسارة حسابها غلط", "الخسارة = (السعر − المرجع) × الكمية", "ERROR", "/procurement",
		`SELECT a."productName" || ' · مسجّل ' || round(a."lossIqd") || ' والصحيح ' || round((a."unitPrice" - a."refPrice") * a.quantity)
		 FROM "ProcurementPriceAlert" a WHERE a.kind <> 'QUOTE' AND abs(a."lossIqd" - (a."unitPrice" - a."refPrice") * a.quantity) > 1`},
	{"standard_missing", "المشتريات", "مواد انشترت بآخر ٩٠ يوم بلا سعر ستاندرد", "بلا سعر مرجعي ماتركس ما يكدر يحكم إذا الشراء غالي", "WARN", "/procurement",
		`SELECT DISTINCT p.product FROM "PurchaseLine" p WHERE p.at > now() - interval '90 days'
		 AND NOT EXISTS (SELECT 1 FROM "StandardPrice" s WHERE s.key = p.key)`},
	{"forecast_unjudged", "التوقعات", "أشهر خلصت وبيها توقع ما انحكم عليه", "كل شهر خلص لازم ينقارن توقعه بالنتيجة", "ERROR", "/",
		`SELECT DISTINCT l.month FROM "MatrixForecastLog" l WHERE l.month < to_char(baghdad_today(), 'YYYY-MM')
		 AND NOT EXISTS (SELECT 1 FROM "MatrixForecastOutcome" o WHERE o.month = l.month)`},
	{"forecast_math", "التوقعات", "أحكام توقع حسابها غلط", "الخطأ٪ = (التوقع − الحقيقة) ÷ الحقيقة، والحكم صح إذا ±١٠٪", "ERROR", "/",
		`SELECT o.month || ' · ' || o.kind FROM "MatrixForecastOutcome" o WHERE o.actual > 0 AND (
		   abs(o."errorPct" - (o.forecast - o.actual) / o.actual * 100) > 0.5
		   OR (o.verdict = 'CORRECT') <> (abs((o.forecast - o.actual) / o.actual * 100) <= 10))`},
}

func (h *MatrixSelfCheckHandler) run() []selfCheck {
	out := make([]selfCheck, 0, len(selfChecks))
	for _, c := range selfChecks {
		rows := []string{}
		r := selfCheck{Key: c.key, Area: c.area, Title: c.title, Why: c.why, Severity: c.severity, Link: c.link, Examples: []string{}}
		if err := h.db.Select(&rows, c.sql); err != nil {
			log.Printf("[selfcheck] %s: %v", c.key, err)
			r.Examples = []string{"تعذر تشغيل الفحص: " + err.Error()}
			r.Count = -1
			out = append(out, r)
			continue
		}
		r.Count, r.OK = len(rows), len(rows) == 0
		if len(rows) > 8 {
			rows = rows[:8]
		}
		r.Examples = rows
		out = append(out, r)
	}
	return out
}

// GET /api/matrix-selfcheck
func (h *MatrixSelfCheckHandler) Get(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]any{"checks": h.run(), "at": time.Now()})
}

// RunDaily كل يوم الساعة ٩ الصبح: إذا أكو فحص «حساب غلط» فشل → إشعار للمالك.
func (h *MatrixSelfCheckHandler) RunDaily() {
	last := ""
	for {
		now := time.Now().In(baghdadLoc())
		if d := now.Format("2006-01-02"); now.Hour() >= 9 && d != last {
			last = d
			bad := []string{}
			for _, c := range h.run() {
				if !c.OK && c.Severity == "ERROR" {
					bad = append(bad, fmt.Sprintf("%s (%d)", c.Title, c.Count))
				}
			}
			if len(bad) > 0 && h.notify != nil {
				h.notify("🧪 فحص دقة ماتركس: " + strings.Join(bad, " · ") + " — افتح «فحص دقة ماتركس» حتى تشوف الحالات")
			}
		}
		time.Sleep(30 * time.Minute)
	}
}

// (ع) 10-15: «ميصير نفس الحجز يكون بمكانين بالسلسلة بنفس الوقت ابد ابد».
// يبني الفحص من نفس شروط المحطات بالضبط — أي حجز يطابق أكثر من محطة يطلع.
func init() {
	st := [][2]string{
		{"بانتظار التثبيت", repository.BucketCondition("pending")},
		{"تم التثبيت", repository.BucketCondition("confirmed")},
		{"مكلّف", repository.BucketCondition("assigned")},
		{"تحتاج إكمال", repository.BucketCondition("partial")},
		{"تم الإنجاز", repository.BucketCondition("done")},
		{"عند المشاريع", repository.BucketCondition("at_projects")},
		{"عند التقني", repository.BucketCondition("at_tech")},
		{"بانتظار الحذف", repository.BucketCondition("delete_pending")},
		{"ما وصلت للتنفيذ", `(NOT ` + repository.StationElsewhereSQL + ` AND (b.status = 'CANCELLED' OR b."awaitingReschedule" OR (b."waitingSince" IS NOT NULL AND COALESCE(b."waitingKind", 'NO_ANSWER') <> 'CUSTOMER_DECISION')))`},
		{"بانتظار موافقة الزبون", `(NOT ` + repository.StationElsewhereSQL + ` AND b.status <> 'CANCELLED' AND b."waitingSince" IS NOT NULL AND b."waitingKind" = 'CUSTOMER_DECISION')`},
	}
	parts := []string{}
	for _, x := range st {
		parts = append(parts, `CASE WHEN `+x[1]+` THEN '`+x[0]+`' END`)
	}
	selfChecks = append([]checkDef{{"double_station", "محطات الحجوزات", "حجز ظاهر بمحطتين بنفس الوقت",
		"كل حجز لازم يكون بمحطة وحدة بالسلسلة — مو مكلّف وبانتظار الحذف، ولا قيد الإنجاز وبانتظار الزبون", "ERROR", "/bookings",
		`SELECT 'حجز ' || code || ' · ' || array_to_string(st, ' + ') FROM (
			SELECT b.code, array_remove(ARRAY[` + strings.Join(parts, ", ") + `], NULL) AS st FROM "Booking" b WHERE b."archivedAt" IS NULL) x
		 WHERE cardinality(st) > 1`}}, selfChecks...)
}
