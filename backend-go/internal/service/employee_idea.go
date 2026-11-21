package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ 💡 عندي فكرة (دازاي 11-21 — أفكار ٤، كايزن) ═══
// الموظف يقترح تحسين لشغله. مدير إدارته (أو نائبه) يرد خلال ٤٨ ساعة: نطبقها / نجربها / لا وليش.
// الي ما ينرد ٤٨ ساعة يصعد للمالك. الفكرة الي «نطبقها» تنحسب للموظف شغل إضافي بتقييمه.

var IdeaCategories = map[string]string{"TIME": "⏱️ وقت", "MONEY": "💰 فلوس", "QUALITY": "⭐ جودة", "SAFETY": "🦺 سلامة", "OTHER": "💡 ثاني"}
var IdeaStatuses = map[string]string{"ACCEPTED": "✅ نطبقها", "TRYING": "🧪 نجربها", "REJECTED": "❌ لا"}

const IdeaReplyHours = 48

var ErrIdeaInvalid = errors.New("اكتب فكرتك بوضوح (١٠ حروف على الأقل)")

type EmployeeIdea struct {
	ID         string     `db:"id" json:"id"`
	EmployeeID string     `db:"employeeId" json:"employeeId"`
	Employee   string     `db:"employee" json:"employee"`
	Text       string     `db:"text" json:"text"`
	Category   string     `db:"category" json:"category"`
	Status     string     `db:"status" json:"status"`
	Reply      *string    `db:"reply" json:"reply"`
	DecidedBy  *string    `db:"decidedBy" json:"decidedBy"`
	DecidedAt  *time.Time `db:"decidedAt" json:"decidedAt"`
	Late       bool       `db:"late" json:"late"` // عدّت ٤٨ ساعة بلا رد
	CreatedAt  time.Time  `db:"createdAt" json:"createdAt"`
}

type IdeaService struct {
	db    *sqlx.DB
	notif interface {
		Create(employeeID, typ, msg string) error
	}
}

func NewIdeaService(db *sqlx.DB, notif interface {
	Create(employeeID, typ, msg string) error
}) *IdeaService {
	return &IdeaService{db: db, notif: notif}
}

const ideaSelect = `SELECT i.id, i."employeeId", e.name AS employee, i.text, i.category, i.status, i.reply,
	d.name AS "decidedBy", i."decidedAt", (i.status = 'NEW' AND i."createdAt" < now() - interval '48 hours') AS late, i."createdAt"
	FROM "EmployeeIdea" i JOIN "Employee" e ON e.id = i."employeeId" LEFT JOIN "Employee" d ON d.id = i."decidedById"`

// Submit فكرة جديدة من الموظف.
func (s *IdeaService) Submit(me, text, category string) error {
	text = strings.TrimSpace(text)
	if len([]rune(text)) < 10 {
		return ErrIdeaInvalid
	}
	if _, ok := IdeaCategories[category]; !ok {
		category = "OTHER"
	}
	_, err := s.db.Exec(`INSERT INTO "EmployeeIdea" (id, "employeeId", text, category) VALUES (gen_random_uuid()::text, $1, $2, $3)`, me, text, category)
	return err
}

// Mine أفكار الموظف وردودها.
func (s *IdeaService) Mine(me string) ([]EmployeeIdea, error) {
	out := []EmployeeIdea{}
	err := s.db.Select(&out, ideaSelect+` WHERE i."employeeId" = $1 ORDER BY i."createdAt" DESC LIMIT 100`, me)
	return out, err
}

// Queue الأفكار الي تنتظر ردي (مدير/نائب الإدارة لمنتسبيها، المالك/المدير الكل) + آخر الي انرد عليها.
func (s *IdeaService) Queue(me, role string) ([]EmployeeIdea, error) {
	out := []EmployeeIdea{}
	err := s.db.Select(&out, ideaSelect+` WHERE i."createdAt" > now() - interval '60 days' AND `+reviewScope(role, `i."employeeId"`)+`
		ORDER BY (i.status = 'NEW') DESC, i."createdAt" DESC LIMIT 200`, me)
	return out, err
}

// Decide الرد على الفكرة. «نطبقها» = شغل إضافي بتقييم صاحبها.
func (s *IdeaService) Decide(me, role, id, status, reply string) error {
	if _, ok := IdeaStatuses[status]; !ok {
		return ErrReviewHidden
	}
	reply = strings.TrimSpace(reply)
	if status == "REJECTED" && reply == "" {
		return errors.New("اكتب ليش لا — الموظف لازم يعرف السبب")
	}
	var replyArg any
	if reply != "" {
		replyArg = reply
	}
	var row struct {
		Emp  string `db:"employeeId"`
		Text string `db:"text"`
	}
	err := s.db.Get(&row, `UPDATE "EmployeeIdea" i SET status = $3, reply = $4, "decidedById" = $1, "decidedAt" = now()
		WHERE i.id = $2 AND i.status IN ('NEW','TRYING') AND `+reviewScope(role, `i."employeeId"`)+` RETURNING i."employeeId", i.text`, me, id, status, replyArg)
	if err != nil {
		return ErrReviewHidden
	}
	if status == "ACCEPTED" {
		short := []rune(row.Text)
		if len(short) > 40 {
			short = append(short[:40], '…')
		}
		label := "💡 فكرة انطبقت: " + string(short)
		_, _ = s.db.Exec(`INSERT INTO "MatrixScore" ("employeeId", source, "sourceId", "sourceLabel", rule, points, "maxPoints", reason, "at")
			VALUES ($1, 'IDEA', $2, $3, 'IDEA', 2, 2, $3, now()) ON CONFLICT DO NOTHING`, row.Emp, id, label)
		InvalidateBoard()
	}
	if s.notif != nil {
		msg := fmt.Sprintf("💡 فكرتك انرد عليها: %s", IdeaStatuses[status])
		if reply != "" {
			msg += " — " + reply
		}
		_ = s.notif.Create(row.Emp, "IDEA", msg)
	}
	return nil
}

// NotifyDaily: كل مدير/نائب إدارة بعدد الأفكار الي تنتظر رده، والي عدّت ٤٨ ساعة تصعد للمالك (مرة وحدة).
func (s *IdeaService) NotifyDaily(now time.Time, claim func(key, day string) (bool, error), notifyOwners func(msg string)) error {
	now = now.In(debriefLoc)
	if now.Hour() < 10 || s.notif == nil {
		return nil
	}
	ok, err := claim("IDEA_DAILY", now.Format("2006-01-02"))
	if err != nil || !ok {
		return err
	}
	var heads []struct {
		ID string `db:"id"`
		N  int    `db:"n"`
	}
	_ = s.db.Select(&heads, `SELECT ap."employeeId" AS id, COUNT(DISTINCT i.id) AS n FROM "EmployeeIdea" i
		JOIN "OrgAssignment" em ON em."employeeId" = i."employeeId" AND em."endedAt" IS NULL
		JOIN "OrgAssignment" ap ON ap."divisionId" = em."divisionId" AND ap."endedAt" IS NULL AND ap."unitId" IS NULL
		  AND ap.position IN ('DIRECTOR','DEPUTY') AND ap."employeeId" <> i."employeeId"
		WHERE i.status = 'NEW' GROUP BY 1`)
	for _, h := range heads {
		_ = s.notif.Create(h.ID, "IDEA", fmt.Sprintf("💡 عندك %d فكرة من موظفينك تنتظر ردك — «أفكار الموظفين». الرد خلال ٤٨ ساعة وإلا تروح للمالك.", h.N))
	}
	var late int
	_ = s.db.Get(&late, `WITH up AS (UPDATE "EmployeeIdea" SET "escalatedAt" = now()
		WHERE status = 'NEW' AND "escalatedAt" IS NULL AND "createdAt" < now() - interval '48 hours' RETURNING 1) SELECT COUNT(*) FROM up`)
	if late > 0 && notifyOwners != nil {
		notifyOwners(fmt.Sprintf("💡 %d فكرة موظف عدّت ٤٨ ساعة وما انرد عليها — «أفكار الموظفين».", late))
	}
	return nil
}
