package repository

import (
	"errors"
	"fmt"
	"log"
	"staffmange-api/internal/errlog"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ رفع طلب للمدير (طلب (ع) 11-04، من اقتراح ماتركس) ═══
// «علي اركان (إداري الكميات) سأل إذا يكدر يوجّه مهمة للمدير وما عنده طريقة».
// أي موظف يرفع طلب بعنوان وتفاصيل وأولوية، والأرقام الي يشوفها (طلب متأخر، نقص
// جرد…) تنرفق ويا الطلب حتى المدير يقرر بلا ما يسأل. الطلب يبقى مسجّل لحد ما ينحسم.

type ManagerRequest struct {
	ID            string     `db:"id" json:"id"`
	RequesterID   string     `db:"requesterId" json:"requesterId"`
	RequesterName string     `db:"requesterName" json:"requesterName"`
	Title         string     `db:"title" json:"title"`
	Details       string     `db:"details" json:"details"`
	Priority      string     `db:"priority" json:"priority"`
	Context       *string    `db:"context" json:"context"`
	Source        string     `db:"source" json:"source"`
	Status        string     `db:"status" json:"status"`
	Reply         *string    `db:"reply" json:"reply"`
	HandledByName *string    `db:"handledByName" json:"handledByName"`
	HandledAt     *time.Time `db:"handledAt" json:"handledAt"`
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`
}

type ManagerRequestRepository struct {
	db    *sqlx.DB
	notif *NotificationRepository
}

func NewManagerRequestRepository(db *sqlx.DB, notif *NotificationRepository) *ManagerRequestRepository {
	return &ManagerRequestRepository{db: db, notif: notif}
}

var managerRequestPriorities = map[string]string{"NORMAL": "عادي", "HIGH": "مهم", "URGENT": "🚨 عاجل"}

const managerRequestSelect = `SELECT m.id, m."requesterId", COALESCE(e.name, '—') AS "requesterName", m.title, m.details,
	m.priority, m.context, m.source, m.status, m.reply, h.name AS "handledByName", m."handledAt", m."createdAt"
	FROM "ManagerRequest" m LEFT JOIN "Employee" e ON e.id = m."requesterId" LEFT JOIN "Employee" h ON h.id = m."handledById"`

// Create يسجّل الطلب وينبّه المدير والمالك.
func (r *ManagerRequestRepository) Create(requesterID, title, details, priority, context, source string) (string, error) {
	title, details = strings.TrimSpace(title), strings.TrimSpace(details)
	if len([]rune(title)) < 3 {
		return "", errors.New("اكتب عنوان للطلب")
	}
	if len([]rune(title)) > 200 || len([]rune(details)) > 4000 || len([]rune(context)) > 4000 {
		return "", errors.New("الطلب طويل هواية — اختصره")
	}
	if _, ok := managerRequestPriorities[priority]; !ok {
		priority = "NORMAL"
	}
	if source != "MATRIX" {
		source = "SCREEN"
	}
	var id string
	if err := r.db.Get(&id, `INSERT INTO "ManagerRequest" (id, "requesterId", title, details, priority, context, source)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, NULLIF($5, ''), $6) RETURNING id`,
		requesterID, title, details, priority, strings.TrimSpace(context), source); err != nil {
		return "", err
	}
	if r.notif != nil {
		var name string
		errlog.Err(r.db.Get(&name, `SELECT name FROM "Employee" WHERE id = $1`, requesterID))
		msg := fmt.Sprintf("📨 طلب للمدير (%s) من %s: %s", managerRequestPriorities[priority], name, title)
		_ = r.notif.CreateForRole("ADMIN", "MANAGER_REQUEST", msg)
		_ = r.notif.CreateForRole("OWNER", "MANAGER_REQUEST", msg)
	}
	return id, nil
}

// Mine طلبات الموظف نفسه (آخر ٥٠).
func (r *ManagerRequestRepository) Mine(requesterID string) ([]ManagerRequest, error) {
	rows := []ManagerRequest{}
	err := r.db.Select(&rows, managerRequestSelect+` WHERE m."requesterId" = $1 ORDER BY m."createdAt" DESC LIMIT 50`, requesterID)
	return rows, err
}

// All للمدير: المفتوح أول (العاجل فوق)، وبعده المحسوم آخر ٦٠ يوم.
func (r *ManagerRequestRepository) All() ([]ManagerRequest, error) {
	rows := []ManagerRequest{}
	err := r.db.Select(&rows, managerRequestSelect+`
		WHERE m.status = 'OPEN' OR m."handledAt" > now() - interval '60 days'
		ORDER BY (m.status = 'OPEN') DESC,
		         CASE m.priority WHEN 'URGENT' THEN 0 WHEN 'HIGH' THEN 1 ELSE 2 END, m."createdAt"
		LIMIT 300`)
	return rows, err
}

// Handle المدير يحسم: DONE أو REJECTED، ويا رد يوصل للموظف.
func (r *ManagerRequestRepository) Handle(id, byID, status, reply string) error {
	if status != "DONE" && status != "REJECTED" {
		return errors.New("الحالة لازم «انعمل» أو «مرفوض»")
	}
	reply = strings.TrimSpace(reply)
	if status == "REJECTED" && len([]rune(reply)) < 3 {
		return errors.New("اكتب سبب الرفض حتى يوصل للموظف")
	}
	var row struct {
		RequesterID string `db:"requesterId"`
		Title       string `db:"title"`
	}
	if err := r.db.Get(&row, `UPDATE "ManagerRequest" SET status = $2, reply = NULLIF($3, ''), "handledById" = $4, "handledAt" = now()
		WHERE id = $1 AND status = 'OPEN' RETURNING "requesterId", title`, id, status, reply, byID); err != nil {
		return errors.New("الطلب مو موجود أو انحسم قبل")
	}
	if r.notif != nil {
		word := "✅ انعمل"
		if status == "REJECTED" {
			word = "❌ انرفض"
		}
		msg := fmt.Sprintf("%s طلبك للمدير: %s", word, row.Title)
		if reply != "" {
			msg += " — " + reply
		}
		_ = r.notif.Create(row.RequesterID, "MANAGER_REQUEST", msg)
	}
	return nil
}

// InventoryContext أرقام المخازن الي تنرفق ويا الطلب (من نفس مصادر procurement_watch).
func InventoryContext(db *sqlx.DB) string {
	s := NewMatrixScoreRepository(db)
	var b strings.Builder
	old := func(title string, rows []ProcDecision, hours float64) {
		n, oldest := 0, 0.0
		for _, x := range rows {
			if h := time.Since(x.CreatedAt).Hours(); h >= hours {
				n++
				if h > oldest {
					oldest = h
				}
			}
		}
		fmt.Fprintf(&b, "• %s: %d معلّق (منها %d أكثر من %.0f ساعة", title, len(rows), n, hours)
		if n > 0 {
			fmt.Fprintf(&b, "، أقدمها %.0f يوم", oldest/24)
		}
		b.WriteString(")\n")
	}
	if m, err := s.MaterialDecisions(true); err == nil {
		old("طلبات المواد", m, 48)
	}
	if t, err := s.ToolDecisions(true); err == nil {
		old("طلبات الأدوات", t, 48)
	}
	var sh struct {
		N      int     `db:"n"`
		Oldest float64 `db:"oldest"`
	}
	if err := db.Get(&sh, `SELECT count(*) AS n, COALESCE(EXTRACT(EPOCH FROM now() - min("checkedAt")) / 86400, 0) AS oldest
		FROM "InventoryCheck" WHERE NOT complete AND NOT resolved`); err == nil {
		fmt.Fprintf(&b, "• نواقص جرد مفتوحة: %d (أقدمها %.0f يوم)\n", sh.N, sh.Oldest)
	}
	if u, err := s.VehiclesUnratedToday(); err == nil {
		fmt.Fprintf(&b, "• سيارات ما انقيّمت اليوم: %d\n", len(u))
	}
	if f, err := s.FleetOverdue(); err == nil && len(f) > 0 {
		fmt.Fprintf(&b, "• صيانة/وثائق فات موعدها: %s\n", strings.Join(f, "؛ "))
	}
	return strings.TrimSpace(b.String())
}

// RemindStale (ع) 11-08: الطلب المفتوح أكثر من ٤٨ ساعة بلا رد يرجع يوصل للمدير والمالك،
// وبعدها كل ٤٨ ساعة لحد ما ينحسم. يأشّر «ذكّرت» بنفس الاستعلام حتى ما يتكرر.
func (r *ManagerRequestRepository) RemindStale() (int, error) {
	rows := []struct {
		Title string    `db:"title"`
		Name  string    `db:"name"`
		Prio  string    `db:"priority"`
		At    time.Time `db:"createdAt"`
	}{}
	if err := r.db.Select(&rows, `UPDATE "ManagerRequest" m SET "remindedAt" = now()
		FROM "Employee" e
		WHERE e.id = m."requesterId" AND m.status = 'OPEN'
		  AND m."createdAt" < now() - interval '48 hours'
		  AND (m."remindedAt" IS NULL OR m."remindedAt" < now() - interval '48 hours')
		RETURNING m.title, e.name, m.priority, m."createdAt"`); err != nil {
		return 0, err
	}
	if r.notif != nil {
		for _, x := range rows {
			days := int(time.Since(x.At).Hours() / 24)
			msg := fmt.Sprintf("⏰ طلب للمدير (%s) من %s بعده بلا رد من %d يوم: %s — من «📨 طلبات للمدير».", managerRequestPriorities[x.Prio], x.Name, days, x.Title)
			_ = r.notif.CreateForRole("ADMIN", "MANAGER_REQUEST", msg)
			_ = r.notif.CreateForRole("OWNER", "MANAGER_REQUEST", msg)
		}
	}
	return len(rows), nil
}

// StartReminder يفحص كل ساعة.
func (r *ManagerRequestRepository) StartReminder() {
	go func() {
		time.Sleep(5 * time.Minute)
		for {
			if n, err := r.RemindStale(); err == nil && n > 0 {
				log.Printf("[manager-requests] ذكّرت بـ%d طلب بلا رد", n)
			}
			time.Sleep(time.Hour)
		}
	}()
}
