package service

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/safeguard"
)

// ═══ تقرير اتصال التقني (ك 11-23، طلب زينا) ═══
// موظف الحجوزات يحوّل الحجز (أو استفسار الطاقة الشمسية) للتقني، والتقني لازم
// يضغط «تم الاتصال» ويكتب ملاحظة (اتصل وما رد…). لحد ما يرفعه:
//   - رقم أحمر على البند بالقائمة (حجوزاتي المسلّمة / استفسارات الطاقة الشمسية).
//   - تذكير كل ٤ ساعات.
//   - إذا عبرت ٢٤ ساعة بلا تقرير: تنخصم نقطة وحدة من نقاط الكي بي آي الثمانية
//     بسبب «عدم إكمال المهام الموجهة إليه» — نقطة بالقاعدة وبدليل، بلا مبلغ فلوس.
// والتقارير والخصم يوصلن للمراقب، ومدير العمليات ونائبه، ومدير الإدارة ونائبه.

// TechContactKpiReason السبب المكتوب بالخصم — نصاً من طلب زينا.
const TechContactKpiReason = "عدم إكمال المهام الموجهة إليه"

const techContactHours = 24

// TechContactAutoTag علامة الخصم التلقائي بالسبب
const TechContactAutoTag = "🤖 خصم تلقائي من النظام"

type TechContactWatch struct {
	db    *sqlx.DB
	notif *repository.NotificationRepository
}

func NewTechContactWatch(db *sqlx.DB, notif *repository.NotificationRepository) *TechContactWatch {
	return &TechContactWatch{db: db, notif: notif}
}

func (w *TechContactWatch) Start() {
	safeguard.Loop("تقرير اتصال التقني", 3*time.Minute, time.Hour, func() {
		safeguard.Run("تذكير اتصال التقني", w.remind)
		safeguard.Run("خصم اتصال التقني", w.deduct)
	})
}

type pendingContact struct {
	Kind   string    `db:"kind"` // booking | solar
	ID     string    `db:"id"`
	Ref    string    `db:"ref"`
	TechID string    `db:"techId"`
	Tech   string    `db:"tech"`
	Since  time.Time `db:"since"`
}

// onLeaveSQL (ل) 11-24 دازاي «لازم النظام يشوف جدول الإجازات»: إجازة معتمدة تغطي يوم من الفترة [من، لـ]
// (بتوقيت بغداد). التقني الي بإجازة ما يوصله تذكير، وما ينخصم عليه إذا إجازته كانت بمهلة الـ٢٤ ساعة.
func onLeaveSQL(emp, from, to string) string {
	return `EXISTS (SELECT 1 FROM "LeaveRequest" lv WHERE lv."employeeId" = ` + emp + ` AND lv.status = 'APPROVED'
		AND lv."startDate" <= baghdad_date(` + to + `) AND lv."endDate" >= baghdad_date(` + from + `))`
}

// pendingSQL كل شي محوّل لتقني وبعده بلا تقرير اتصال (ومفتوح).
const pendingSQL = `
	SELECT 'booking' AS kind, b.id, b.code AS ref, b."handoverToId" AS "techId", e.name AS tech, b."handoverAt" AS since
	FROM "Booking" b JOIN "Employee" e ON e.id = b."handoverToId"
	WHERE b."handoverToId" IS NOT NULL AND b."techContactedAt" IS NULL AND b."techContactKpiAt" IS NULL
	  AND b."handoverAt" IS NOT NULL AND b.status::text NOT IN ('COMPLETED', 'CANCELLED') AND b."archivedAt" IS NULL
	  -- (ل) 11-24 دازاي: الحجز الي بدا شغله (وصلوا أو بدوا) ما ينحاسب عليه التقني حتى لو نسى «تم الاتصال»
	  AND b."arrivedAt" IS NULL AND b."startedAt" IS NULL
	UNION ALL
	SELECT 'solar', q.id, '#' || q.number::text, q."techId", e.name, q."techClaimedAt"
	FROM "SolarInquiry" q JOIN "Employee" e ON e.id = q."techId"
	WHERE q."techId" IS NOT NULL AND q."techOutcome" IS NULL AND q."qualityAt" IS NULL
	  AND q."techContactKpiAt" IS NULL AND q."techClaimedAt" IS NOT NULL
	  -- (ك 11-25) الي جاوبه موظف بالمكالمة ما ينحاسب عليه التقني بالاتصال
	  AND q."answeredAt" IS NULL`

func refLabel(p pendingContact) string {
	if p.Kind == "solar" {
		return "استفسار الطاقة الشمسية " + p.Ref
	}
	return "الحجز " + p.Ref
}

// remind تذكير كل ٤ ساعات لحد الـ٢٤ (مرة وحدة بكل ٤ ساعات حتى لو الدورة تأخرت).
func (w *TechContactWatch) remind() {
	var rows []pendingContact
	if err := w.db.Select(&rows, `SELECT * FROM (`+pendingSQL+`) p
		WHERE p.since < now() - interval '4 hours' AND p.since > now() - interval '24 hours'
		  AND NOT `+onLeaveSQL(`p."techId"`, `now()`, `now()`)); err != nil {
		log.Printf("[tech-contact] remind: %v", err)
		return
	}
	for _, p := range rows {
		var recent bool
		_ = w.db.Get(&recent, `SELECT EXISTS (SELECT 1 FROM "Notification" WHERE "employeeId" = $1 AND type = 'tech_contact_reminder'
			AND strpos(message, $2) > 0 AND "createdAt" > now() - interval '235 minutes')`, p.TechID, refLabel(p)+" —")
		if recent {
			continue
		}
		left := techContactHours - int(time.Since(p.Since).Hours())
		if left < 1 {
			left = 1
		}
		_ = w.notif.Create(p.TechID, "tech_contact_reminder",
			fmt.Sprintf("⏰ عندك اتصال لازم تسويه: %s — اتصل بالزبون واضغط «تم الاتصال» واكتب ملاحظتك. باقي %d ساعة وبعدها تنخصم نقطة كي بي آي.", refLabel(p), left))
	}
}

// deduct بعد ٢٤ ساعة بلا تقرير: نقطة وحدة، مرة وحدة لكل حجز/استفسار.
func (w *TechContactWatch) deduct() {
	var rows []pendingContact
	if err := w.db.Select(&rows, `SELECT * FROM (`+pendingSQL+`) p WHERE p.since < now() - interval '24 hours'
		-- (ك 11-23) القاعدة تبدي من يوم انطبقت: الي انحوّل قبلها ما ينحاسب عليه بأثر رجعي
		AND p.since >= (SELECT "appliedAt" FROM "SchemaMigration" WHERE version = '0522_tech_contact_report')
		AND NOT `+onLeaveSQL(`p."techId"`, `p.since`, `p.since + interval '24 hours'`)); err != nil {
		log.Printf("[tech-contact] deduct: %v", err)
		return
	}
	if len(rows) == 0 {
		return
	}
	var author string
	if err := w.db.Get(&author, `SELECT id FROM "Employee" WHERE role IN ('OWNER', 'ADMIN') AND status = 'ACTIVE'
		ORDER BY CASE role WHEN 'OWNER' THEN 0 ELSE 1 END LIMIT 1`); err != nil {
		log.Printf("[tech-contact] author: %v", err)
		return
	}
	for _, p := range rows {
		// (ك 11-23) يبدي باسم المعيار حتى ينحسب من نقطة «تنفيذ المهام الموكلة إليه» — نفس خصم تقصير التثبيت
		// (ل) 11-24 دازاي: ينكتب صراحةً إنه خصم تلقائي من النظام (مو المالك خصم بيده). الجدول يحتاج مقيّم،
		// فيبقى باسم المالك تقنياً، بس السبب يوضح. واسم المعيار يبقى بالأول حتى ينحسب من نقطته.
		reason := fmt.Sprintf("%s — %s: ما رفع تقرير اتصال %s خلال %d ساعة من التحويل · %s", w.taskCriterion(), TechContactKpiReason, refLabel(p), techContactHours, TechContactAutoTag)
		applied, err := w.apply(p, author, reason)
		if err != nil {
			log.Printf("[tech-contact] apply %s: %v", p.ID, err)
			continue
		}
		if !applied {
			continue // انخصمت قبل — ما نعيد الإشعارات
		}
		_ = w.notif.Create(p.TechID, "kpi_deduction", "🔻 انخصمت نقطة كي بي آي — "+reason+". اتصل بالزبون ورفع التقرير هسه.")
		_ = w.notif.CreateForTechContactWatchers("tech_contact_kpi", fmt.Sprintf("🔻 %s: انخصمت نقطة كي بي آي — %s", p.Tech, reason), p.TechID)
	}
}

// taskCriterion اسم معيار «تنفيذ المهام» كما هو بالقاعدة (من النقاط الثمانية).
func (w *TechContactWatch) taskCriterion() string {
	var l string
	if err := w.db.Get(&l, `SELECT label FROM "KpiCriterion" WHERE label LIKE '%تنفيذ المهام%' OR label LIKE '%إتمام المهام%' OR label LIKE '%اتمام المهام%' ORDER BY "createdAt" LIMIT 1`); err != nil || l == "" {
		return "تنفيذ المهام الموكلة إليه"
	}
	return l
}

func (w *TechContactWatch) apply(p pendingContact, author, reason string) (bool, error) {
	tx, err := w.db.Beginx()
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()
	table := `"Booking"`
	if p.Kind == "solar" {
		table = `"SolarInquiry"`
	}
	res, err := tx.Exec(`UPDATE `+table+` SET "techContactKpiAt" = now() WHERE id = $1 AND "techContactKpiAt" IS NULL`, p.ID)
	if err != nil {
		return false, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return false, nil // انخصمت قبل
	}
	if _, err := tx.Exec(`INSERT INTO "KpiEvaluation" (id, "employeeId", "evaluatorId", points, reason, "deductionAmount")
		VALUES ($1, $2, $3, -1, $4, 0)`, uuid.NewString(), p.TechID, author, reason); err != nil {
		return false, err
	}
	return true, tx.Commit()
}

// TechContactBadges أرقام القائمة: الي محوّل إله وبعده بلا تقرير اتصال.
func TechContactBadges(db *sqlx.DB, me string) (bookings, solar int) {
	_ = db.Get(&bookings, `SELECT count(*) FROM "Booking" b WHERE b."handoverToId" = $1 AND b."techContactedAt" IS NULL
		AND b.status::text NOT IN ('COMPLETED', 'CANCELLED') AND b."archivedAt" IS NULL`, me)
	_ = db.Get(&solar, `SELECT count(*) FROM "SolarInquiry" q WHERE q."techOutcome" IS NULL AND q."qualityAt" IS NULL AND q."answeredAt" IS NULL
		AND (q."techId" = $1 OR (q."techId" IS NULL AND `+model.HasPermSQL(`$1`, `= 'solar_inquiry_tech'`)+`))`, me)
	return
}
