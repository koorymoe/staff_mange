package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/repository"
)

// ═══ 🎯 «راقب فلان» — مهام ماتركس (قرار (ع) 10-11) ═══
// المالك يكله «راقب هالموظف هالأسبوع»: كل ليلة ماتركس يكتب سطر يومية
// بالي صار (حساب بحت، مجاني)، وبآخر يوم يكتب خلاصة (هايكو، والأسماء
// مخفية) ويدزها للي طلب المراقبة.

const missionJournalHour = 20

type MatrixMissionService struct {
	db     *sqlx.DB
	aiRepo *repository.AiRepository
	notif  *repository.NotificationRepository
	agent  *MatrixAgent
}

func NewMatrixMissionService(db *sqlx.DB, aiRepo *repository.AiRepository, notif *repository.NotificationRepository, agent *MatrixAgent) *MatrixMissionService {
	return &MatrixMissionService{db: db, aiRepo: aiRepo, notif: notif, agent: agent}
}

type MatrixMission struct {
	ID           string     `db:"id" json:"id"`
	EmployeeID   string     `db:"employeeId" json:"employeeId"`
	EmployeeName string     `db:"employeeName" json:"employeeName"`
	Note         *string    `db:"note" json:"note"`
	StartsOn     string     `db:"startsOn" json:"startsOn"`
	EndsOn       string     `db:"endsOn" json:"endsOn"`
	Active       bool       `db:"active" json:"active"`
	Summary      *string    `db:"summary" json:"summary"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`
	SummaryAt    *time.Time `db:"summaryAt" json:"summaryAt"`
}

type MatrixMissionLog struct {
	Day  string `db:"day" json:"day"`
	Text string `db:"text" json:"text"`
}

const missionSelect = `SELECT m.id, m."employeeId", e.name AS "employeeName", m.note,
	to_char(m."startsOn", 'YYYY-MM-DD') AS "startsOn", to_char(m."endsOn", 'YYYY-MM-DD') AS "endsOn",
	m.active, m.summary, m."createdAt", m."summaryAt"
	FROM "MatrixMission" m JOIN "Employee" e ON e.id = m."employeeId"`

func (s *MatrixMissionService) List(createdBy string) ([]MatrixMission, error) {
	rows := []MatrixMission{}
	err := s.db.Select(&rows, missionSelect+` WHERE m."createdById" = $1 ORDER BY m.active DESC, m."createdAt" DESC LIMIT 50`, createdBy)
	return rows, err
}

func (s *MatrixMissionService) Get(id, createdBy string) (*MatrixMission, []MatrixMissionLog, error) {
	var m MatrixMission
	if err := s.db.Get(&m, missionSelect+` WHERE m.id = $1 AND m."createdById" = $2`, id, createdBy); err != nil {
		return nil, nil, err
	}
	logs := []MatrixMissionLog{}
	err := s.db.Select(&logs, `SELECT to_char(day, 'YYYY-MM-DD') AS day, text FROM "MatrixMissionLog" WHERE "missionId" = $1 ORDER BY day`, id)
	return &m, logs, err
}

func (s *MatrixMissionService) Create(createdBy, employeeID, note string, days int) (*MatrixMission, error) {
	if days < 1 || days > 30 {
		days = 7
	}
	var id string
	err := s.db.Get(&id, `INSERT INTO "MatrixMission" (id, "employeeId", "createdById", note, "startsOn", "endsOn")
		SELECT gen_random_uuid()::text, e.id, $2, NULLIF($3, ''), (now() AT TIME ZONE 'Asia/Baghdad')::date,
			(now() AT TIME ZONE 'Asia/Baghdad')::date + ($4 - 1)
		FROM "Employee" e WHERE e.id = $1 AND e.status = 'ACTIVE' AND e.role NOT IN ('ADMIN', 'OWNER')
		RETURNING id`, employeeID, createdBy, strings.TrimSpace(note), days)
	if err != nil {
		return nil, fmt.Errorf("الموظف مو موجود أو ما ينراقب")
	}
	m, _, err := s.Get(id, createdBy)
	return m, err
}

func (s *MatrixMissionService) Stop(id, createdBy string) error {
	_, err := s.db.Exec(`UPDATE "MatrixMission" SET active = false WHERE id = $1 AND "createdById" = $2`, id, createdBy)
	return err
}

// dayFacts سطر يومية لموظف بيوم (حساب بحت).
func (s *MatrixMissionService) dayFacts(empID, day string) string {
	var f struct {
		Done    int     `db:"done"`
		Late    int     `db:"late"`
		Open    int     `db:"open"`
		Inv     int     `db:"inv"`
		CheckIn *string `db:"checkin"`
	}
	err := s.db.Get(&f, `SELECT
		(SELECT count(DISTINCT b.id) FROM "BookingAssignment" ba JOIN "Booking" b ON b.id = ba."bookingId"
			WHERE ba."employeeId" = $1 AND b.status = 'COMPLETED' AND (b."completedAt" AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Baghdad')::date = $2::date) AS done,
		(SELECT count(DISTINCT b.id) FROM "BookingAssignment" ba JOIN "Booking" b ON b.id = ba."bookingId"
			WHERE ba."employeeId" = $1 AND b.status = 'COMPLETED' AND (b."completedAt" AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Baghdad')::date = $2::date
			AND b."scheduledAt" IS NOT NULL AND b."completedAt" > b."scheduledAt" + interval '4 hours') AS late,
		(SELECT count(DISTINCT b.id) FROM "BookingAssignment" ba JOIN "Booking" b ON b.id = ba."bookingId"
			WHERE ba."employeeId" = $1 AND b.status IN ('CONFIRMED', 'PENDING', 'PARTIAL') AND b."archivedAt" IS NULL) AS open,
		(SELECT count(*) FROM "LeaderInvoice" li WHERE li."employeeId" = $1
			AND (li."createdAt" AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Baghdad')::date = $2::date) AS inv,
		(SELECT to_char(min(a."checkIn") AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Baghdad', 'HH24:MI') FROM "Attendance" a WHERE a."employeeId" = $1 AND a.date::date = $2::date) AS checkin`,
		empID, day)
	if err != nil {
		log.Printf("mission facts: %v", err)
		return "تعذر جمع بيانات اليوم"
	}
	in := "ما سجّل حضور"
	if f.CheckIn != nil {
		in = "حضر " + *f.CheckIn
	}
	return fmt.Sprintf("%s · خلّص %d شغلة (متأخرة %d) · عنده %d مفتوحة · قدّم %d فاتورة", in, f.Done, f.Late, f.Open, f.Inv)
}

// RunIfDue كل ليلة بعد ٨: سطر اليومية لكل مهمة فعّالة، والخلاصة لمن تنتهي.
func (s *MatrixMissionService) RunIfDue() error {
	now := time.Now().In(debriefLoc)
	if now.Hour() < missionJournalHour {
		return nil
	}
	day := now.Format("2006-01-02")
	ms := []MatrixMission{}
	if err := s.db.Select(&ms, missionSelect+` WHERE m.active AND m."startsOn" <= $1::date`, day); err != nil {
		return err
	}
	for _, m := range ms {
		_, _ = s.db.Exec(`INSERT INTO "MatrixMissionLog" (id, "missionId", day, text) VALUES (gen_random_uuid()::text, $1, $2::date, $3)
			ON CONFLICT ("missionId", day) DO NOTHING`, m.ID, day, s.dayFacts(m.EmployeeID, day))
		if day >= m.EndsOn {
			s.finish(m)
		}
	}
	return nil
}

func (s *MatrixMissionService) finish(m MatrixMission) {
	logs := []MatrixMissionLog{}
	_ = s.db.Select(&logs, `SELECT to_char(day, 'YYYY-MM-DD') AS day, text FROM "MatrixMissionLog" WHERE "missionId" = $1 ORDER BY day`, m.ID)
	var b strings.Builder
	for _, l := range logs {
		fmt.Fprintf(&b, "%s: %s\n", l.Day, l.Text)
	}
	summary := "يومية المراقبة:\n" + b.String()
	if s.agent != nil && s.agent.Enabled("MISSION") {
		note := ""
		if m.Note != nil {
			note = "سبب المراقبة: " + *m.Note + "\n"
		}
		task := fmt.Sprintf("راقبت الموظف %s من %s لحد %s.\n%sهاي يوميته:\n%s\nاكتب خلاصة قصيرة (٥ أسطر بالكثير): شلون جان أداؤه، شنو الزين، شنو المقلق، وشنو تنصح المالك يسوي. افتح employee_profile إذا تحتاج.",
			m.EmployeeName, m.StartsOn, m.EndsOn, note, b.String())
		if run, err := s.agent.Run("MISSION", chatSystemPrompt(), task); err == nil && strings.TrimSpace(run.Answer) != "" {
			summary = run.Answer
		}
	}
	_, _ = s.db.Exec(`UPDATE "MatrixMission" SET active = false, summary = $2, "summaryAt" = now() WHERE id = $1`, m.ID, summary)
	var creator string
	if err := s.db.Get(&creator, `SELECT "createdById" FROM "MatrixMission" WHERE id = $1`, m.ID); err == nil {
		_ = s.notif.Create(creator, "AI_MISSION", "🎯 ماتركس خلّص مراقبة "+m.EmployeeName+" — الخلاصة جاهزة بتبويب «مهام ماتركس».")
	}
}

// chatSystemPrompt الوصف العام لماتركس (مختصر) للمهام.
func chatSystemPrompt() string {
	return `أنت «ماتركس»، الذكاء الاصطناعي لشركة الأماني بالعراق. تكتب للمالك باللهجة العراقية، مختصر وواضح، بالأرقام الي شفتها بس.
- الموظفين يجوك كرموز «موظف#n» — استعمل نفس الرمز.
- لا تقترح غرامة أو خصم فلوس، ولا تخمّن أسباب شخصية.`
}
