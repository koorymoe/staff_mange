package service

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/repository"
)

// ═══ 🎯 الماتشينغ — وحدة التدريب (دازاي 11-12) ═══
// «بدل بنك الأسئلة يطلع فدشي اسمه ماتشينغ: اسم الموظف والأشياء الي يحتاج يتدرب عليها».
// مصدرين:
//   - مشكلة بحجز: الليدر قيّم الفني ١–٢، أو كتب «يحتاج تدريب على…»، أو كتب إنه ما يعرف.
//     هذا **إلزامي ومشدّد**: موعد ٧ أيام، ما ينلغي، والتأخير يوصل للمالك ومسؤول التدريب.
//   - نقص مهارة/خبرة: يشتغل على خدمة وما عنده مهاراتها بالنظام. هذا **اقتراح** — ينلغي بسبب.
// مسؤول وحدة التدريب يبدي التدريب، ولمن يخلص يكتب التقرير ويقيّم الموظف (١–٥) — التقرير يوصل للمالك.
// الأدمن والمالك ما يدخلون بالحساب.

const (
	NeedBookingIssue = "BOOKING_ISSUE"
	NeedSkillGap     = "SKILL_GAP"
	NeedManual       = "MANUAL"

	mandatoryDue = 7 * 24 * time.Hour
	// تقييم الليدر الي ينحسب مشكلة
	badRatingMax = 2
)

var (
	ErrNeedNotFound  = errors.New("مو موجود")
	ErrNeedMandatory = errors.New("هذا تدريب إلزامي بسبب مشكلة بحجز — ما ينلغي، لازم يتدرب")
	ErrNeedState     = errors.New("الحالة ما تسمح")
	ErrNeedReport    = errors.New("اكتب التقرير وقيّم الموظف من ١ لـ٥")
)

// كلمات الليدر الي تعني «ما يعرف» — تقييم بيه وحدة منهن = تدريب إلزامي حتى لو النجوم عالية.
var dontKnowWords = []string{"ما يعرف", "مايعرف", "ميعرف", "ما يفتهم", "مايفتهم", "ما يفهم", "يحتاج تدريب", "يحتاج يتدرب", "ما عنده خبرة", "ماعنده خبرة", "ضعيف"}

func NoteSaysUntrained(note string) bool {
	n := strings.TrimSpace(note)
	for _, w := range dontKnowWords {
		if strings.Contains(n, w) {
			return true
		}
	}
	return false
}

type TrainingMatching struct {
	db    *sqlx.DB
	notif *repository.NotificationRepository
	now   func() time.Time
}

func NewTrainingMatching(db *sqlx.DB, notif *repository.NotificationRepository) *TrainingMatching {
	return &TrainingMatching{db: db, notif: notif, now: time.Now}
}

// trackForService أول مسار بالمنهج مربوط بالخدمة (للفني).
func (m *TrainingMatching) trackForService(serviceID string) *string {
	var id string
	if err := m.db.Get(&id, `SELECT id FROM "TrainingCurriculum" WHERE "parentId" IS NULL AND "serviceId" = $1
		ORDER BY (role = 'TECHNICIAN') DESC, ord LIMIT 1`, serviceID); err != nil {
		return nil
	}
	return &id
}

// nodeForTopic أقرب مفردة بالمنهج اسمها يطابق الموضوع — داخل مسار الخدمة إذا موجود.
func (m *TrainingMatching) nodeForTopic(topic string, track *string) *string {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return track
	}
	var id string
	err := m.db.Get(&id, `
		WITH RECURSIVE t AS (
			SELECT id, title FROM "TrainingCurriculum" WHERE $2::text IS NOT NULL AND id = $2
			UNION ALL SELECT c.id, c.title FROM "TrainingCurriculum" c JOIN t ON c."parentId" = t.id)
		SELECT id FROM (
			SELECT id, title FROM t
			UNION ALL SELECT id, title FROM "TrainingCurriculum" WHERE $2::text IS NULL) x
		WHERE strpos(lower($1), lower(title)) > 0 OR strpos(lower(title), lower($1)) > 0
		ORDER BY length(title) DESC LIMIT 1`, topic, track)
	if err != nil {
		return track
	}
	return &id
}

type badRating struct {
	ID            string    `db:"id"`
	TechnicianID  string    `db:"technicianId"`
	TechName      string    `db:"techName"`
	BookingID     string    `db:"bookingId"`
	BookingCode   string    `db:"code"`
	ServiceID     *string   `db:"serviceId"`
	ServiceName   *string   `db:"serviceName"`
	Score         int       `db:"score"`
	Note          *string   `db:"note"`
	NeedsTraining *string   `db:"needsTraining"`
	CreatedAt     time.Time `db:"createdAt"`
}

// Generate يقرا التقييمات والمهارات ويطلّع الاحتياجات الجديدة (ما يكرر). يرجّع عدد الإلزامي والاقتراح.
func (m *TrainingMatching) Generate() (mandatory, suggested int, err error) {
	rows := []badRating{}
	err = m.db.Select(&rows, `
		SELECT r.id, r."technicianId", e.name AS "techName", r."bookingId", b.code, b."serviceId", s.name AS "serviceName",
		       r.score, r.note, r."needsTraining", r."createdAt"
		  FROM "CrewRating" r
		  JOIN "Employee" e ON e.id = r."technicianId" AND e.role NOT IN ('ADMIN','OWNER')
		  JOIN "Booking" b ON b.id = r."bookingId"
		  LEFT JOIN "Service" s ON s.id = b."serviceId"
		 WHERE r."createdAt" > now() - interval '60 days'
		   AND NOT EXISTS (SELECT 1 FROM "TrainingMatch" n WHERE n."sourceRatingId" = r.id)
		   AND (r.score <= $1 OR COALESCE(trim(r."needsTraining"), '') <> '' OR r.note IS NOT NULL)`, badRatingMax)
	if err != nil {
		return 0, 0, err
	}
	for _, r := range rows {
		note := ""
		if r.Note != nil {
			note = strings.TrimSpace(*r.Note)
		}
		want := ""
		if r.NeedsTraining != nil {
			want = strings.TrimSpace(*r.NeedsTraining)
		}
		if r.Score > badRatingMax && want == "" && !NoteSaysUntrained(note) {
			continue
		}
		svc := "الخدمة"
		if r.ServiceName != nil && *r.ServiceName != "" {
			svc = *r.ServiceName
		}
		topic := want
		if topic == "" {
			topic = svc
		}
		var track *string
		if r.ServiceID != nil {
			track = m.trackForService(*r.ServiceID)
		}
		cur := m.nodeForTopic(topic, track)
		reason := fmt.Sprintf("الليدر قيّمه %d/5 بحجز %s (%s)", r.Score, r.BookingCode, svc)
		if note != "" {
			reason += " — «" + note + "»"
		}
		due := r.CreatedAt.Add(mandatoryDue)
		res, e := m.db.Exec(`INSERT INTO "TrainingMatch" ("employeeId", kind, "curriculumId", topic, reason, "sourceBookingId", "sourceRatingId", "dueAt")
			VALUES ($1,'BOOKING_ISSUE',$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING`, r.TechnicianID, cur, topic, reason, r.BookingID, r.ID, due)
		if e != nil {
			log.Printf("training need (rating %s): %v", r.ID, e)
			continue
		}
		if n, _ := res.RowsAffected(); n == 1 {
			mandatory++
			_ = m.notif.CreateForPermission("training_manage", "TRAINING_NEED",
				fmt.Sprintf("🎯 تدريب إلزامي: %s يحتاج يتدرب على «%s» — %s. لازم يخلص خلال ٧ أيام (من «الماتشينغ» بوحدة التدريب).", r.TechName, topic, reason))
		}
	}

	// نقص مهارة: فني اشتغل على خدمة بآخر ٩٠ يوم وما عنده مهارة من مهاراتها بالنظام — اقتراح.
	res, err := m.db.Exec(`
		INSERT INTO "TrainingMatch" ("employeeId", kind, "curriculumId", topic, reason, "skillId")
		SELECT DISTINCT ON (ba."employeeId", sk.id) ba."employeeId", 'SKILL_GAP',
		       (SELECT c.id FROM "TrainingCurriculum" c WHERE c."parentId" IS NULL AND c."serviceId" = sk."serviceId" ORDER BY c.ord LIMIT 1),
		       sk.name, 'اشتغل على ' || s.name || ' وما عنده مهارة «' || sk.name || '» بالنظام', sk.id
		  FROM "BookingAssignment" ba
		  JOIN "Employee" e ON e.id = ba."employeeId" AND e.role NOT IN ('ADMIN','OWNER')
		  JOIN "Booking" b ON b.id = ba."bookingId" AND b."completedAt" > now() - interval '90 days'
		  JOIN "Skill" sk ON sk."serviceId" = b."serviceId"
		  JOIN "Service" s ON s.id = sk."serviceId"
		 WHERE NOT EXISTS (SELECT 1 FROM "EmployeeSkill" es WHERE es."employeeId" = ba."employeeId" AND es."skillId" = sk.id AND es."canPerform")
		ON CONFLICT DO NOTHING`)
	if err != nil {
		return mandatory, 0, err
	}
	n, _ := res.RowsAffected()
	return mandatory, int(n), nil
}

// NotifyOverdue الإلزامي الي فات موعده وما خلص — مرة وحدة للمالك ومسؤول التدريب.
func (m *TrainingMatching) NotifyOverdue() (int, error) {
	rows := []struct {
		ID    string `db:"id"`
		Name  string `db:"name"`
		Topic string `db:"topic"`
	}{}
	err := m.db.Select(&rows, `
		UPDATE "TrainingMatch" n SET "overdueNotifiedAt" = now()
		  FROM "Employee" e
		 WHERE e.id = n."employeeId" AND n.kind = 'BOOKING_ISSUE' AND n.status IN ('OPEN','IN_TRAINING')
		   AND n."dueAt" < now() AND n."overdueNotifiedAt" IS NULL
		RETURNING n.id, e.name, n.topic`)
	if err != nil {
		return 0, err
	}
	for _, r := range rows {
		msg := fmt.Sprintf("⏰ تدريب إلزامي متأخر: %s بعده ما خلص تدريبه على «%s» وفات موعده.", r.Name, r.Topic)
		_ = m.notif.CreateForPermission("training_manage", "TRAINING_NEED", msg)
		_ = m.notif.CreateForRole("OWNER", "TRAINING_NEED", msg)
	}
	return len(rows), nil
}

// Run دورة كاملة (للمؤقت).
func (m *TrainingMatching) Run() {
	if a, b, err := m.Generate(); err != nil {
		log.Printf("training matching: %v", err)
	} else if a+b > 0 {
		log.Printf("🎯 الماتشينغ: %d إلزامي، %d اقتراح", a, b)
	}
	if _, err := m.NotifyOverdue(); err != nil {
		log.Printf("training overdue: %v", err)
	}
}

type TrainingNeedRow struct {
	ID           string     `db:"id" json:"id"`
	EmployeeID   string     `db:"employeeId" json:"employeeId"`
	EmployeeName string     `db:"employeeName" json:"employeeName"`
	EmployeeRole string     `db:"employeeRole" json:"employeeRole"`
	Kind         string     `db:"kind" json:"kind"`
	CurriculumID *string    `db:"curriculumId" json:"curriculumId"`
	Path         *string    `db:"path" json:"path"`
	Topic        string     `db:"topic" json:"topic"`
	Reason       *string    `db:"reason" json:"reason"`
	BookingID    *string    `db:"sourceBookingId" json:"bookingId"`
	Status       string     `db:"status" json:"status"`
	DueAt        *time.Time `db:"dueAt" json:"dueAt"`
	DismissNote  *string    `db:"dismissNote" json:"dismissNote"`
	StartedAt    *time.Time `db:"startedAt" json:"startedAt"`
	DoneAt       *time.Time `db:"doneAt" json:"doneAt"`
	ReportText   *string    `db:"reportText" json:"reportText"`
	ReportScore  *int       `db:"reportScore" json:"reportScore"`
	ReportedBy   *string    `db:"reportedBy" json:"reportedBy"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`
}

const needSelect = `
	WITH RECURSIVE p AS (
		SELECT id, id AS leaf, title::text AS path, "parentId" FROM "TrainingCurriculum"
		UNION ALL SELECT p.id, p.leaf, c.title || ' ← ' || p.path, c."parentId" FROM p JOIN "TrainingCurriculum" c ON c.id = p."parentId")
	SELECT n.id, n."employeeId", e.name AS "employeeName", e.role AS "employeeRole", n.kind, n."curriculumId",
	       (SELECT path FROM p WHERE p.leaf = n."curriculumId" AND p."parentId" IS NULL LIMIT 1) AS path,
	       n.topic, n.reason, n."sourceBookingId", n.status, n."dueAt", n."dismissNote", n."startedAt", n."doneAt",
	       n."reportText", n."reportScore", rb.name AS "reportedBy", n."createdAt"
	  FROM "TrainingMatch" n
	  JOIN "Employee" e ON e.id = n."employeeId"
	  LEFT JOIN "Employee" rb ON rb.id = n."reportedById"`

// List status: open = مفتوح وبالتدريب، done = التقارير، all.
func (m *TrainingMatching) List(status string) ([]TrainingNeedRow, error) {
	where := `n.status IN ('OPEN','IN_TRAINING')`
	order := `(n.kind = 'BOOKING_ISSUE') DESC, n."dueAt" NULLS LAST, e.name`
	switch status {
	case "done":
		where, order = `n.status = 'DONE'`, `n."doneAt" DESC`
	case "dismissed":
		where, order = `n.status = 'DISMISSED'`, `n."updatedAt" DESC`
	}
	rows := []TrainingNeedRow{}
	err := m.db.Select(&rows, needSelect+` WHERE `+where+` ORDER BY `+order+` LIMIT 500`)
	return rows, err
}

func (m *TrainingMatching) get(id string) (*TrainingNeedRow, error) {
	var r TrainingNeedRow
	if err := m.db.Get(&r, needSelect+` WHERE n.id = $1`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNeedNotFound
		}
		return nil, err
	}
	return &r, nil
}

// Add مسؤول التدريب يضيف احتياج بإيده (إلزامي بموعد ٧ أيام).
func (m *TrainingMatching) Add(actor, employeeID, topic string, curriculumID *string, reason string) (*TrainingNeedRow, error) {
	topic = strings.TrimSpace(topic)
	if topic == "" || employeeID == "" {
		return nil, ErrNeedState
	}
	var role string
	if err := m.db.Get(&role, `SELECT role FROM "Employee" WHERE id = $1`, employeeID); err != nil || role == "ADMIN" || role == "OWNER" {
		return nil, ErrNeedNotFound
	}
	if curriculumID != nil && *curriculumID == "" {
		curriculumID = nil
	}
	var id string
	err := m.db.Get(&id, `INSERT INTO "TrainingMatch" ("employeeId", kind, "curriculumId", topic, reason, "dueAt", "createdById")
		VALUES ($1,'MANUAL',$2,$3,NULLIF($4,''),$5,$6) RETURNING id`, employeeID, curriculumID, topic, strings.TrimSpace(reason), m.now().Add(mandatoryDue), actor)
	if err != nil {
		return nil, err
	}
	return m.get(id)
}

// Start مسؤول التدريب يبدي تدريب الموظف — يوصله إشعار.
func (m *TrainingMatching) Start(id string) (*TrainingNeedRow, error) {
	r, err := m.get(id)
	if err != nil {
		return nil, err
	}
	if r.Status != "OPEN" {
		return nil, ErrNeedState
	}
	m.db.MustExec(`UPDATE "TrainingMatch" SET status = 'IN_TRAINING', "startedAt" = now(), "updatedAt" = now() WHERE id = $1`, id)
	_ = m.notif.Create(r.EmployeeID, "TRAINING_NEED", fmt.Sprintf("📚 وحدة التدريب: بديت تدريبك على «%s».", r.Topic))
	return m.get(id)
}

// Complete مسؤول التدريب يكتب التقرير ويقيّم — التقرير يوصل للمالك.
func (m *TrainingMatching) Complete(actor, id, report string, score int) (*TrainingNeedRow, error) {
	report = strings.TrimSpace(report)
	if report == "" || score < 1 || score > 5 {
		return nil, ErrNeedReport
	}
	r, err := m.get(id)
	if err != nil {
		return nil, err
	}
	if r.Status != "OPEN" && r.Status != "IN_TRAINING" {
		return nil, ErrNeedState
	}
	m.db.MustExec(`UPDATE "TrainingMatch" SET status = 'DONE', "doneAt" = now(), "startedAt" = COALESCE("startedAt", now()),
		"reportText" = $2, "reportScore" = $3, "reportedById" = $4, "updatedAt" = now() WHERE id = $1`, id, report, score, actor)
	kind := "اقتراح"
	if r.Kind != NeedSkillGap {
		kind = "إلزامي"
	}
	_ = m.notif.CreateForRole("OWNER", "TRAINING_REPORT",
		fmt.Sprintf("📋 تقرير وحدة التدريب: %s خلص تدريبه (%s) على «%s» — التقييم %d/5. التفاصيل بـ«تقارير التدريب».", r.EmployeeName, kind, r.Topic, score))
	return m.get(id)
}

// Dismiss يلغي اقتراح (نقص مهارة) بسبب. الإلزامي ما ينلغي.
func (m *TrainingMatching) Dismiss(id, note string) (*TrainingNeedRow, error) {
	note = strings.TrimSpace(note)
	r, err := m.get(id)
	if err != nil {
		return nil, err
	}
	if r.Kind != NeedSkillGap {
		return nil, ErrNeedMandatory
	}
	if r.Status != "OPEN" || note == "" {
		return nil, ErrNeedState
	}
	m.db.MustExec(`UPDATE "TrainingMatch" SET status = 'DISMISSED', "dismissNote" = $2, "updatedAt" = now() WHERE id = $1`, id, note)
	return m.get(id)
}
