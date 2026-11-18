package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// ═══ تأكيد السلبي بالتقييم (دازاي 11-18) ═══
// «التقييم السيء… يجي المراقب او منو بلهيكل الجديد… مدير الادارة او نائب الادارة
// يتأكد انو صدك هوه غلطان». أي نقطة سلبية ما تنحسب إلا بعد ما يأكدها:
//   - تقييم الليدر السيء (١–٢ نجمة) للفني ← مدير إدارة الفني أو نائبه
//   - المشكلة الوظيفية ← نفس المسؤول يحدد منو الغلطان (أو ماكو غلطان)
//   - نقص العدّة بعد الشغل ← إداري الكميات يگول «مبرَّر» (ScoreJustify)
// الموظف الي ماله مكان بالهيكل: المالك أو مدير النظام يأكدون.

var ErrReviewHidden = errors.New("مو موجود")

// deptHeadSQL: $1 مدير/نائب الإدارة (مستوى الإدارة، مو الوحدة) الي الموظف (target) بيها.
// النائب ما يأكد على مديره.
func deptHeadSQL(target string) string {
	return `EXISTS (SELECT 1 FROM "OrgAssignment" ap
		JOIN "OrgAssignment" em ON em."divisionId" = ap."divisionId" AND em."endedAt" IS NULL AND em."employeeId" = ` + target + `
		WHERE ap."employeeId" = $1 AND ap."endedAt" IS NULL AND ap."unitId" IS NULL AND ap.position IN ('DIRECTOR','DEPUTY')
		  AND ap."employeeId" <> em."employeeId"
		  AND NOT (ap.position = 'DEPUTY' AND em.position = 'DIRECTOR' AND em."unitId" IS NULL))`
}

// reviewScope: المالك/المدير يشوفون الكل عدا الي يخصهم؛ غيرهم بس منتسبي إدارتهم.
func reviewScope(role, target string) string {
	if model.IsAdminRole(role) {
		return target + ` <> $1`
	}
	return deptHeadSQL(target)
}

type ScoreReviewItem struct {
	Kind      string    `db:"kind" json:"kind"` // LEADER_RATING | ISSUE
	ID        string    `db:"id" json:"id"`
	Employee  string    `db:"employee" json:"employee"`
	Other     *string   `db:"other" json:"other"` // الليدر الي قيّم، أو الطرف الثاني
	OtherID   *string   `db:"otherId" json:"otherId"`
	EmpID     string    `db:"empId" json:"empId"`
	Score     *int      `db:"score" json:"score"`
	Title     string    `db:"title" json:"title"`
	Note      *string   `db:"note" json:"note"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

type ScoreReviewService struct {
	db    *sqlx.DB
	notif interface {
		Create(employeeID, typ, msg string) error
	}
}

func NewScoreReviewService(db *sqlx.DB, notif interface {
	Create(employeeID, typ, msg string) error
}) *ScoreReviewService {
	return &ScoreReviewService{db: db, notif: notif}
}

// Queue السلبي الي ينتظر تأكيدي (آخر ٦٠ يوم).
func (s *ScoreReviewService) Queue(me, role string) ([]ScoreReviewItem, error) {
	out := []ScoreReviewItem{}
	q := `SELECT 'LEADER_RATING' AS kind, c.id, t.name AS employee, l.name AS other, c."leaderId" AS "otherId", c."technicianId" AS "empId",
			c.score, 'حجز ' || COALESCE(b.code, '') AS title, c.note, c."createdAt"
		FROM "CrewRating" c JOIN "Employee" t ON t.id = c."technicianId" JOIN "Employee" l ON l.id = c."leaderId"
		LEFT JOIN "Booking" b ON b.id = c."bookingId"
		WHERE c.score <= 2 AND c."reviewStatus" IS NULL AND c."createdAt" > now() - interval '60 days'
		  AND c."leaderId" <> $1 AND ` + reviewScope(role, `c."technicianId"`) + `
		UNION ALL
		SELECT 'ISSUE', i.id, a.name, pb.name, i."partyBId", i."partyAId", NULL::int, i.title, i.description, i."createdAt"
		FROM "WorkplaceIssue" i JOIN "Employee" a ON a.id = i."partyAId" LEFT JOIN "Employee" pb ON pb.id = i."partyBId"
		WHERE NOT i."faultDecided" AND i."createdAt" > now() - interval '60 days'
		  AND $1 <> i."partyAId" AND $1 IS DISTINCT FROM i."partyBId"
		  AND (` + reviewScope(role, `i."partyAId"`) + ` OR (i."partyBId" IS NOT NULL AND ` + reviewScope(role, `i."partyBId"`) + `))
		ORDER BY "createdAt" DESC LIMIT 200`
	err := s.db.Select(&out, q, me)
	return out, err
}

// DecideRating يأكد التقييم السيء (ينحسب) أو يرفضه (ما ينحسب).
func (s *ScoreReviewService) DecideRating(me, role, id string, confirm bool) error {
	st := "REJECTED"
	if confirm {
		st = "CONFIRMED"
	}
	res, err := s.db.Exec(`UPDATE "CrewRating" c SET "reviewStatus" = $3, "reviewedById" = $1, "reviewedAt" = now()
		WHERE c.id = $2 AND c.score <= 2 AND c."reviewStatus" IS NULL AND c."leaderId" <> $1 AND `+reviewScope(role, `c."technicianId"`), me, id, st)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return ErrReviewHidden
	}
	InvalidateBoard()
	return nil
}

// DecideIssue يحدد منو الغلطان بالمشكلة: fault = أحد الطرفين، أو "" = ماكو غلطان.
func (s *ScoreReviewService) DecideIssue(me, role, id, fault string) error {
	fault = strings.TrimSpace(fault)
	var faultArg any
	if fault != "" {
		faultArg = fault
	}
	res, err := s.db.Exec(`UPDATE "WorkplaceIssue" i SET "faultDecided" = true, "faultEmployeeId" = $3, "faultById" = $1, "faultAt" = now(), "updatedAt" = now()
		WHERE i.id = $2 AND NOT i."faultDecided" AND $1 <> i."partyAId" AND $1 IS DISTINCT FROM i."partyBId"
		  AND ($3::text IS NULL OR $3 = i."partyAId" OR $3 = i."partyBId")
		  AND (`+reviewScope(role, `i."partyAId"`)+` OR (i."partyBId" IS NOT NULL AND `+reviewScope(role, `i."partyBId"`)+`))`, me, id, faultArg)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return ErrReviewHidden
	}
	InvalidateBoard()
	return nil
}

// ═══ مبرّر نقص العدّة بعد الشغل — إداري الكميات ═══

type AfterShortage struct {
	BookingID  string     `db:"bookingId" json:"bookingId"`
	Code       *string    `db:"code" json:"code"`
	EmployeeID string     `db:"employeeId" json:"employeeId"`
	Employee   string     `db:"employee" json:"employee"`
	Missing    *string    `db:"missingItems" json:"missingItems"`
	CheckedAt  time.Time  `db:"checkedAt" json:"checkedAt"`
	Justified  *time.Time `db:"justifiedAt" json:"justifiedAt"`
}

// AfterShortages نواقص العدّة بعد الحجوزات (آخر ٣٠ يوم).
func (s *ScoreReviewService) AfterShortages() ([]AfterShortage, error) {
	out := []AfterShortage{}
	err := s.db.Select(&out, `SELECT i."bookingId", b.code, i."employeeId", e.name AS employee, i."missingItems", i."checkedAt", i."justifiedAt"
		FROM "BookingAfterInventory" i JOIN "Employee" e ON e.id = i."employeeId" LEFT JOIN "Booking" b ON b.id = i."bookingId"
		WHERE NOT i.complete AND i."checkedAt" > now() - interval '30 days'
		ORDER BY (i."justifiedAt" IS NULL) DESC, i."checkedAt" DESC LIMIT 300`)
	return out, err
}

// Justify «مبرَّر»: النقص ما ينحسب على الفني. ما يگدر يبرر نقصه هو.
func (s *ScoreReviewService) Justify(me, bookingID, employeeID string) error {
	res, err := s.db.Exec(`UPDATE "BookingAfterInventory" SET "justifiedById" = $1, "justifiedAt" = now()
		WHERE "bookingId" = $2 AND "employeeId" = $3 AND NOT complete AND "justifiedAt" IS NULL AND "employeeId" <> $1`, me, bookingID, employeeID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return ErrReviewHidden
	}
	InvalidateBoard()
	return nil
}

// NotifyDaily (مرة باليوم بعد ١١ بغداد) ينبّه كل مدير/نائب إدارة بعدد السلبي الي ينتظر تأكيده،
// والي ماله مسؤول بالهيكل يروح للمدير. claim يحجز اليوم حتى ما يتكرر.
func (s *ScoreReviewService) NotifyDaily(now time.Time, claim func(key, day string) (bool, error), notifyAdmins func(msg string)) error {
	now = now.In(debriefLoc)
	if now.Hour() < 11 || s.notif == nil {
		return nil
	}
	type row struct {
		Target string `db:"target"`
	}
	targets := []row{}
	if err := s.db.Select(&targets, `
		SELECT c."technicianId" AS target FROM "CrewRating" c
		WHERE c.score <= 2 AND c."reviewStatus" IS NULL AND c."createdAt" > now() - interval '60 days'
		UNION ALL
		SELECT i."partyAId" FROM "WorkplaceIssue" i WHERE NOT i."faultDecided" AND i."createdAt" > now() - interval '60 days'`); err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil
	}
	ok, err := claim("SCORE_REVIEW_DAILY", now.Format("2006-01-02"))
	if err != nil || !ok {
		return err
	}
	count := map[string]int{}
	orphans := 0
	for _, t := range targets {
		heads := []string{}
		_ = s.db.Select(&heads, `SELECT DISTINCT ap."employeeId" FROM "OrgAssignment" ap
			JOIN "OrgAssignment" em ON em."divisionId" = ap."divisionId" AND em."endedAt" IS NULL AND em."employeeId" = $1
			WHERE ap."endedAt" IS NULL AND ap."unitId" IS NULL AND ap.position IN ('DIRECTOR','DEPUTY') AND ap."employeeId" <> $1
			  AND NOT (ap.position = 'DEPUTY' AND em.position = 'DIRECTOR' AND em."unitId" IS NULL)`, t.Target)
		if len(heads) == 0 {
			orphans++
		}
		for _, h := range heads {
			count[h]++
		}
	}
	for id, n := range count {
		_ = s.notif.Create(id, "SCORE_REVIEW", fmt.Sprintf("⚖️ عندك %d تقييم سلبي ينتظر تأكيدك (تقييم ليدر سيء أو مشكلة وظيفية) — «تأكيد التقييمات». الي ما تأكده ما ينحسب.", n))
	}
	if orphans > 0 && notifyAdmins != nil {
		notifyAdmins(fmt.Sprintf("⚖️ %d تقييم سلبي ينتظر تأكيد لموظفين ما إلهم مدير إدارة بالهيكل — «تأكيد التقييمات».", orphans))
	}
	return nil
}
