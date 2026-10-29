package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ═══ «السكرتير» — التخزين والاستعلامات (docs/secretary.md) ═══
// المحادثة وسجل المقترحات والتفضيلات لكل مستخدم (المالك والمدير، كل واحد
// بذاكرته)، واستعلامات القراءة الي تحتاجها أدواته. ما يكتب على بيانات النظام
// إلا عن طريق ExtraTaskRepository.Create بعد «إرسال» من المستخدم.

type SecretaryRepository struct{ db *sqlx.DB }

func NewSecretaryRepository(db *sqlx.DB) *SecretaryRepository { return &SecretaryRepository{db: db} }

// DB للقراءات البسيطة بالمعالج (ملفات كمال).
func (r *SecretaryRepository) DB() *sqlx.DB { return r.db }

var ErrSecretaryActionDone = errors.New("هذا المقترح انرسل أو انلغى من قبل")

// ── المحادثة ──

type SecretaryMessageRow struct {
	ID         string          `db:"id" json:"id"`
	At         time.Time       `db:"createdAt" json:"at"`
	Role       string          `db:"role" json:"role"`
	Text       string          `db:"text" json:"text"`
	Transcript *string         `db:"transcript" json:"transcript"`
	Cards      json.RawMessage `db:"cards" json:"cards"`
	ActionID   *string         `db:"actionId" json:"-"`
	// السجل يرجّع حالة المقترح ويا الرسالة الي طلّعته
	ActionStatus *string `db:"actionStatus" json:"-"`
	ActionResult *string `db:"actionResult" json:"-"`
	HasAudio     bool    `db:"hasAudio" json:"hasAudio"`
}

func (r *SecretaryRepository) SaveMessage(empID, role, text string, transcript *string, cards any, actionID *string) (string, error) {
	return r.SaveMessageIn(nil, empID, role, text, transcript, cards, actionID)
}

// SaveMessageIn يحفظ رسالة بمحادثة (convID nil = بلا محادثة — للتوافق القديم).
func (r *SecretaryRepository) SaveMessageIn(convID *string, empID, role, text string, transcript *string, cards any, actionID *string) (string, error) {
	raw, err := json.Marshal(cards)
	if err != nil || cards == nil || string(raw) == "null" {
		raw = []byte("[]")
	}
	id := uuid.NewString()
	_, err = r.db.Exec(`INSERT INTO "SecretaryMessage" (id, "employeeId", role, text, transcript, cards, "actionId", "conversationId")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, id, empID, role, text, transcript, raw, actionID, convID)
	if err == nil && convID != nil {
		_, _ = r.db.Exec(`UPDATE "SecretaryConversation" SET "updatedAt" = now() WHERE id = $1`, *convID)
	}
	return id, err
}

// ── 💬 محادثات ايريس المتعددة + الذاكرة (دازاي 10-23) ──

type SecretaryConversation struct {
	ID             string     `db:"id" json:"id"`
	Title          string     `db:"title" json:"title"`
	Summary        string     `db:"summary" json:"-"`
	SummarizedUpTo *time.Time `db:"summarizedUpTo" json:"-"`
	Messages       int        `db:"messages" json:"messages"`
	CreatedAt      time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time  `db:"updatedAt" json:"updatedAt"`
}

const secConvCols = `c.id, c.title, c.summary, c."summarizedUpTo", c."createdAt", c."updatedAt",
	(SELECT COUNT(*) FROM "SecretaryMessage" m WHERE m."conversationId" = c.id)::int AS messages`

func (r *SecretaryRepository) Conversations(empID string) ([]SecretaryConversation, error) {
	rows := []SecretaryConversation{}
	err := r.db.Select(&rows, `SELECT `+secConvCols+` FROM "SecretaryConversation" c
		WHERE c."employeeId" = $1 ORDER BY c."updatedAt" DESC LIMIT 100`, empID)
	return rows, err
}

// Conversation محادثته هو بس.
func (r *SecretaryRepository) Conversation(empID, id string) (*SecretaryConversation, error) {
	var c SecretaryConversation
	if err := r.db.Get(&c, `SELECT `+secConvCols+` FROM "SecretaryConversation" c WHERE c.id = $1 AND c."employeeId" = $2`, id, empID); err != nil {
		return nil, errors.New("المحادثة مو موجودة")
	}
	return &c, nil
}

func (r *SecretaryRepository) CreateConversation(empID, title string) (*SecretaryConversation, error) {
	if strings.TrimSpace(title) == "" {
		title = "محادثة جديدة"
	}
	id := uuid.NewString()
	if _, err := r.db.Exec(`INSERT INTO "SecretaryConversation" (id, "employeeId", title) VALUES ($1, $2, $3)`, id, empID, title); err != nil {
		return nil, err
	}
	return r.Conversation(empID, id)
}

// LatestConversation آخر محادثة — وإذا ماكو يفتح وحدة جديدة.
func (r *SecretaryRepository) LatestConversation(empID string) (*SecretaryConversation, error) {
	var id string
	if err := r.db.Get(&id, `SELECT id FROM "SecretaryConversation" WHERE "employeeId" = $1 ORDER BY "updatedAt" DESC LIMIT 1`, empID); err == nil {
		return r.Conversation(empID, id)
	}
	return r.CreateConversation(empID, "")
}

func (r *SecretaryRepository) RenameConversation(empID, id, title string) error {
	res, err := r.db.Exec(`UPDATE "SecretaryConversation" SET title = $3 WHERE id = $1 AND "employeeId" = $2`, id, empID, title)
	return noteHit(res, err)
}

func (r *SecretaryRepository) DeleteConversation(empID, id string) error {
	res, err := r.db.Exec(`DELETE FROM "SecretaryConversation" WHERE id = $1 AND "employeeId" = $2`, id, empID)
	return noteHit(res, err)
}

// ConversationHistory آخر رسائل محادثة (الأحدث آخراً).
func (r *SecretaryRepository) ConversationHistory(convID string, limit int) ([]SecretaryMessageRow, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows := []SecretaryMessageRow{}
	err := r.db.Select(&rows, `SELECT * FROM (
		SELECT m.id, m."createdAt", m.role, m.text, m.transcript, m.cards, m."actionId",
		       a.status AS "actionStatus", a.result AS "actionResult",
		       EXISTS (SELECT 1 FROM "SecretaryAudio" au WHERE au."messageId" = m.id) AS "hasAudio"
		FROM "SecretaryMessage" m LEFT JOIN "SecretaryAction" a ON a.id = m."actionId"
		WHERE m."conversationId" = $1 ORDER BY m."createdAt" DESC LIMIT $2) x
		ORDER BY "createdAt" ASC`, convID, limit)
	return rows, err
}

// ToSummarize الرسائل الي بعدها ما دخلت الذاكرة، ما عدا آخر keep — إذا صارن أكثر من threshold.
func (r *SecretaryRepository) ToSummarize(c *SecretaryConversation, keep, threshold int) ([]SecretaryMessageRow, error) {
	rows := []SecretaryMessageRow{}
	err := r.db.Select(&rows, `SELECT m.id, m."createdAt", m.role, m.text, m.transcript, m.cards, m."actionId",
		NULL::text AS "actionStatus", NULL::text AS "actionResult"
		FROM "SecretaryMessage" m WHERE m."conversationId" = $1
		  AND ($2::timestamptz IS NULL OR m."createdAt" > $2)
		ORDER BY m."createdAt"`, c.ID, c.SummarizedUpTo)
	if err != nil || len(rows)-keep < threshold {
		return nil, err
	}
	return rows[:len(rows)-keep], nil
}

func (r *SecretaryRepository) SetMemory(convID, summary string, upTo time.Time) error {
	_, err := r.db.Exec(`UPDATE "SecretaryConversation" SET summary = $2, "summarizedUpTo" = $3 WHERE id = $1`, convID, summary, upTo)
	return err
}

// TitleIfNew يسمّي المحادثة من أول رسالة إذا بعدها «محادثة جديدة».
func (r *SecretaryRepository) TitleIfNew(convID, text string) {
	t := strings.TrimSpace(strings.ReplaceAll(text, "\n", " "))
	if rs := []rune(t); len(rs) > 40 {
		t = string(rs[:40]) + "…"
	}
	if t == "" {
		return
	}
	_, _ = r.db.Exec(`UPDATE "SecretaryConversation" SET title = $2 WHERE id = $1 AND title = 'محادثة جديدة'`, convID, t)
}

// History آخر الرسائل (الأحدث آخراً) ويا حالة المقترح إذا اكو.
func (r *SecretaryRepository) History(empID string, limit int) ([]SecretaryMessageRow, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows := []SecretaryMessageRow{}
	err := r.db.Select(&rows, `SELECT * FROM (
		SELECT m.id, m."createdAt", m.role, m.text, m.transcript, m.cards, m."actionId",
		       a.status AS "actionStatus", a.result AS "actionResult"
		FROM "SecretaryMessage" m LEFT JOIN "SecretaryAction" a ON a.id = m."actionId"
		WHERE m."employeeId" = $1 ORDER BY m."createdAt" DESC LIMIT $2) x
		ORDER BY "createdAt" ASC`, empID, limit)
	return rows, err
}

// Prefs التفضيلات المعتمدة بس.
func (r *SecretaryRepository) Prefs(empID string) []string {
	out := []string{}
	_ = r.db.Select(&out, `SELECT text FROM "SecretaryPref" WHERE "employeeId" = $1 AND active ORDER BY "createdAt"`, empID)
	return out
}

// ── المقترحات (السجل) ──

type SecretaryActionRow struct {
	ID        string          `db:"id"`
	EmpID     string          `db:"employeeId"`
	Kind      string          `db:"kind"`
	Request   string          `db:"request"`
	Summary   string          `db:"summary"`
	Fields    json.RawMessage `db:"fields"`
	Status    string          `db:"status"`
	Result    *string         `db:"result"`
	CreatedAt time.Time       `db:"createdAt"`
	DecidedAt *time.Time      `db:"decidedAt"`
}

func (r *SecretaryRepository) CreateAction(empID, kind, request, summary string, fields any) (string, error) {
	raw, err := json.Marshal(fields)
	if err != nil {
		return "", err
	}
	id := uuid.NewString()
	_, err = r.db.Exec(`INSERT INTO "SecretaryAction" (id, "employeeId", kind, request, summary, fields)
		VALUES ($1, $2, $3, $4, $5, $6)`, id, empID, kind, request, summary, raw)
	return id, err
}

// Action مقترح المستخدم نفسه بس — مقترح غيره «مو موجود».
func (r *SecretaryRepository) Action(id, empID string) (*SecretaryActionRow, error) {
	var a SecretaryActionRow
	if err := r.db.Get(&a, `SELECT id, "employeeId", kind, request, summary, fields, status, result, "createdAt", "decidedAt"
		FROM "SecretaryAction" WHERE id = $1 AND "employeeId" = $2`, id, empID); err != nil {
		return nil, err
	}
	return &a, nil
}

// Decide يقفل المقترح مرة وحدة: الثانية ErrSecretaryActionDone.
func (r *SecretaryRepository) Decide(id, empID, status, result string, fields any) error {
	var raw any
	if fields != nil {
		b, err := json.Marshal(fields)
		if err != nil {
			return err
		}
		raw = b
	}
	res, err := r.db.Exec(`UPDATE "SecretaryAction" SET status = $3, result = NULLIF($4, ''),
		fields = COALESCE($5::jsonb, fields), "decidedAt" = now()
		WHERE id = $1 AND "employeeId" = $2 AND status = 'PROPOSED'`, id, empID, status, result, raw)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrSecretaryActionDone
	}
	return nil
}

// ── استعلامات الأدوات ──

type SecretaryBookingHit struct {
	ID          string     `db:"id"`
	Code        string     `db:"code"`
	Status      string     `db:"status"`
	Customer    *string    `db:"customer"`
	Phone       *string    `db:"phone"`
	Service     *string    `db:"service"`
	ScheduledAt *time.Time `db:"scheduledAt"`
	UpdatedAt   time.Time  `db:"updatedAt"`
	Crew        *string    `db:"crew"`
	Station     string     `db:"station"`
}

type SecretaryProjectHit struct {
	ID        string    `db:"id"`
	Code      *string   `db:"code"`
	Name      string    `db:"name"`
	Stage     string    `db:"stage"`
	Owner     *string   `db:"owner"`
	UpdatedAt time.Time `db:"updatedAt"`
}

type SecretaryCustomerHit struct {
	ID       string  `db:"id"`
	Code     int     `db:"customerCode"`
	Name     string  `db:"name"`
	Phone    *string `db:"phone"`
	Bookings int     `db:"bookings"`
}

type SecretaryEmployeeHit struct {
	ID       string  `db:"id"`
	Name     string  `db:"name"`
	Role     string  `db:"role"`
	Phone    *string `db:"phone"`
	IsLeader bool    `db:"isLeader"`
}

type SecretaryFind struct {
	Bookings  []SecretaryBookingHit
	Projects  []SecretaryProjectHit
	Customers []SecretaryCustomerHit
	Employees []SecretaryEmployeeHit
}

// Find بحث واحد بالاسم أو الكود أو الهاتف بكل الأنواع (٥ لكل نوع).
func (r *SecretaryRepository) Find(q string) (*SecretaryFind, error) {
	q = strings.TrimSpace(q)
	out := &SecretaryFind{}
	if len([]rune(q)) < 2 {
		return out, nil
	}
	like := "%" + q + "%"
	digits := strings.Map(func(c rune) rune {
		if c >= '0' && c <= '9' {
			return c
		}
		return -1
	}, q)
	phone := ""
	if len(digits) >= 6 {
		phone = "%" + digits[len(digits)-6:] + "%"
	}
	if err := r.db.Select(&out.Bookings, `
		SELECT b.id, b.code, b.status::text AS status, c.name AS customer, c.phone, s.name AS service,
		       b."scheduledAt", b."updatedAt",
		       (SELECT string_agg(e.name, '، ' ORDER BY e.name) FROM "BookingAssignment" ba
		          JOIN "Employee" e ON e.id = ba."employeeId" WHERE ba."bookingId" = b.id) AS crew,
		       `+stationCaseSQL()+` AS station
		FROM "Booking" b
		LEFT JOIN "Customer" c ON c.id = b."customerId"
		LEFT JOIN "Service" s ON s.id = b."serviceId"
		WHERE b."archivedAt" IS NULL AND (b.code ILIKE $1 OR c.name ILIKE $1 OR ($2 <> '' AND c.phone LIKE $2))
		ORDER BY (b.code = $3) DESC, b."updatedAt" DESC LIMIT 5`, like, phone, q); err != nil {
		return nil, err
	}
	if err := r.db.Select(&out.Projects, `
		SELECT p.id, p.code, p.name, p.stage, e.name AS owner, p."updatedAt"
		FROM "Project" p
		LEFT JOIN "Employee" e ON e.id = COALESCE(p."delegatedToEmployeeId", p."responsibleEmployeeId")
		WHERE p.name ILIKE $1 OR p.code ILIKE $1 OR ($2 <> '' AND p.phone LIKE $2)
		ORDER BY p."updatedAt" DESC LIMIT 5`, like, phone); err != nil {
		return nil, err
	}
	if err := r.db.Select(&out.Customers, `
		SELECT c.id, c."customerCode", c.name, c.phone,
		       (SELECT count(*) FROM "Booking" b WHERE b."customerId" = c.id)::int AS bookings
		FROM "Customer" c
		WHERE c.name ILIKE $1 OR ($2 <> '' AND c.phone LIKE $2) OR c."customerCode"::text = $3
		ORDER BY bookings DESC LIMIT 5`, like, phone, q); err != nil {
		return nil, err
	}
	if err := r.db.Select(&out.Employees, `
		SELECT id, name, role::text AS role, phone, "isLeader"
		FROM "Employee" WHERE status = 'ACTIVE' AND name ILIKE $1
		ORDER BY name LIMIT 5`, like); err != nil {
		return nil, err
	}
	return out, nil
}

type SecretaryStep struct {
	Label string    `db:"label"`
	At    time.Time `db:"at"`
	By    *string   `db:"by"`
}

// BookingTimeline مراحل حجز بالترتيب (بالكود أو المعرّف).
func (r *SecretaryRepository) BookingTimeline(ref string) (*SecretaryBookingHit, []SecretaryStep, error) {
	var b SecretaryBookingHit
	err := r.db.Get(&b, `
		SELECT b.id, b.code, b.status::text AS status, c.name AS customer, c.phone, s.name AS service,
		       b."scheduledAt", b."updatedAt",
		       (SELECT string_agg(e.name, '، ' ORDER BY e.name) FROM "BookingAssignment" ba
		          JOIN "Employee" e ON e.id = ba."employeeId" WHERE ba."bookingId" = b.id) AS crew,
		       `+stationCaseSQL()+` AS station
		FROM "Booking" b
		LEFT JOIN "Customer" c ON c.id = b."customerId"
		LEFT JOIN "Service" s ON s.id = b."serviceId"
		WHERE b.code = $1 OR b.id = $1 LIMIT 1`, strings.TrimSpace(ref))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	steps := []SecretaryStep{}
	err = r.db.Select(&steps, `SELECT label, at, by FROM (
		SELECT 'انسجّل' AS label, b."createdAt" AS at, (SELECT name FROM "Employee" WHERE id = b."createdById") AS by FROM "Booking" b WHERE b.id = $1
		UNION ALL SELECT 'انتصل بالزبون', b."confirmationContactedAt", (SELECT name FROM "Employee" WHERE id = b."confirmationContactedById") FROM "Booking" b WHERE b.id = $1
		UNION ALL SELECT 'انثبّت', b."confirmedAt", b."confirmedByName" FROM "Booking" b WHERE b.id = $1
		UNION ALL SELECT CASE p.stage WHEN 'ACCEPTED' THEN 'استلم' WHEN 'MATERIALS' THEN 'المواد جاهزة' WHEN 'DEPARTED' THEN 'انطلق'
		                 WHEN 'ARRIVED' THEN 'وصل' WHEN 'STARTED' THEN 'بدأ العمل' WHEN 'COMPLETED' THEN 'أنجز' ELSE p.stage END,
		       p."pressedAt", e.name
		  FROM "BookingStageProof" p LEFT JOIN "Employee" e ON e.id = p."employeeId" WHERE p."bookingId" = $1
		UNION ALL SELECT 'انجز', b."completedAt", NULL FROM "Booking" b WHERE b.id = $1
	) x WHERE at IS NOT NULL ORDER BY at`, b.ID)
	return &b, steps, err
}

type SecretaryAttendanceRow struct {
	Name     string     `db:"name"`
	Role     string     `db:"role"`
	CheckIn  *time.Time `db:"checkIn"`
	CheckOut *time.Time `db:"checkOut"`
	Shift    *string    `db:"shift"`
}

// AttendanceToday كل الموظفين الفعّالين وحضورهم اليوم (بلا المدير والمالك).
func (r *SecretaryRepository) AttendanceToday() ([]SecretaryAttendanceRow, error) {
	rows := []SecretaryAttendanceRow{}
	err := r.db.Select(&rows, `
		SELECT e.name, e.role::text AS role, a."checkIn", a."checkOut",
		       NULLIF(concat_ws('–', e."shiftStart", e."shiftEnd"), '') AS shift
		FROM "Employee" e
		LEFT JOIN LATERAL (SELECT "checkIn", "checkOut" FROM "Attendance"
		     WHERE "employeeId" = e.id AND baghdad_date("checkIn") = baghdad_today()
		     ORDER BY "checkIn" DESC LIMIT 1) a ON true
		WHERE e.status = 'ACTIVE' AND e.role::text NOT IN ('ADMIN', 'OWNER')
		ORDER BY (a."checkIn" IS NULL), e.name`)
	return rows, err
}

// SecretaryPending شي ينتظر قرار المالك أو المدير.
type SecretaryPending struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	Question  string    `json:"question"`
	Why       string    `json:"why"`
	Impact    string    `json:"impact"`
	Affects   string    `json:"affects"`
	Link      string    `json:"link"`
	CreatedAt time.Time `json:"createdAt"`
}

// Pending طلبات حذف الحجوزات، والإجازات الي مسارها الإدارة.
func (r *SecretaryRepository) Pending() ([]SecretaryPending, error) {
	type row struct {
		ID        string    `db:"id"`
		Kind      string    `db:"kind"`
		Who       string    `db:"who"`
		Ref       string    `db:"ref"`
		Note      string    `db:"note"`
		CreatedAt time.Time `db:"createdAt"`
	}
	rows := []row{}
	err := r.db.Select(&rows, `
		SELECT 'delete:' || d.id AS id, 'delete_request' AS kind, COALESCE(e.name, '') AS who,
		       COALESCE(b.code, '') AS ref, d.reason AS note, d."createdAt"
		FROM "BookingDeleteRequest" d
		JOIN "Booking" b ON b.id = d."bookingId"
		LEFT JOIN "Employee" e ON e.id = d."requestedById"
		WHERE d.status = 'PENDING'
		UNION ALL
		SELECT 'leave:' || l.id, 'leave', COALESCE(e.name, ''),
		       to_char(l."startDate", 'YYYY-MM-DD') || CASE WHEN l."endDate" <> l."startDate"
		         THEN ' ← ' || to_char(l."endDate", 'YYYY-MM-DD') ELSE '' END,
		       COALESCE(l.reason, ''), l."createdAt"
		FROM "LeaveRequest" l
		LEFT JOIN "Employee" e ON e.id = l."employeeId"
		WHERE l.status IN ('PENDING', 'PRELIMINARY') AND l.route = 'ADMIN'
		ORDER BY "createdAt" LIMIT 30`)
	if err != nil {
		return nil, err
	}
	out := make([]SecretaryPending, 0, len(rows))
	for _, x := range rows {
		p := SecretaryPending{ID: x.ID, Kind: x.Kind, CreatedAt: x.CreatedAt}
		switch x.Kind {
		case "delete_request":
			p.Question = "حذف الحجز " + x.Ref + " — توافق؟"
			p.Why = x.Who + " طلب الحذف: " + x.Note
			p.Impact = "الموافقة تحذف الحجز، والرفض يرجّعه للشغل"
			p.Affects = "الحجز " + x.Ref
			p.Link = "/booking-delete-requests"
		case "leave":
			p.Question = "إجازة " + x.Who + " (" + x.Ref + ") — توافق؟"
			p.Why = "مسارها الإدارة"
			if x.Note != "" {
				p.Why += ": " + x.Note
			}
			p.Impact = "الموافقة تشيله من التكليف بهالأيام"
			p.Affects = x.Who
			p.Link = "/leaves"
		}
		out = append(out, p)
	}
	return out, nil
}

// SecretaryToday أرقام اليوم الخام للملخص — نفس تعريفات اللوحة (DrillCountSQL).
type SecretaryToday struct {
	BookingsToday  int `db:"bookingsToday" json:"bookingsToday"`
	CompletedToday int `db:"completedToday" json:"completedToday"`
	InField        int `db:"inField" json:"inField"`
	Overdue        int `db:"overdue" json:"overdue"`
	NeedsCoord     int `db:"needsCoord" json:"needsCoord"`
	NeedsPaper     int `db:"needsPaper" json:"needsPaper"`
	OpenComplaints int `db:"openComplaints" json:"openComplaints"`
	PresentNow     int `db:"presentNow" json:"presentNow"`
	ActiveStaff    int `db:"activeStaff" json:"activeStaff"`
	OpenTasks      int `db:"openTasks" json:"openTasks"`
}

func (r *SecretaryRepository) Today() (*SecretaryToday, error) {
	var t SecretaryToday
	err := r.db.Get(&t, `
		SELECT `+DrillCountSQL("today_scheduled")+` AS "bookingsToday",
		       `+DrillCountSQL("completed_today")+` AS "completedToday",
		       `+DrillCountSQL("in_field")+` AS "inField",
		       `+DrillCountSQL("overdue")+` AS overdue,
		       `+DrillCountSQL("needs_coord")+` AS "needsCoord",
		       `+DrillCountSQL("needs_paper")+` AS "needsPaper",
		       (SELECT count(*) FROM "Complaint" WHERE status::text IN ('NEW', 'IN_PROGRESS'))::int AS "openComplaints",
		       (SELECT count(DISTINCT a."employeeId") FROM "Attendance" a JOIN "Employee" e ON e.id = a."employeeId"
		         WHERE baghdad_date(a."checkIn") = baghdad_today() AND a."checkOut" IS NULL
		           AND e.role::text NOT IN ('ADMIN', 'OWNER'))::int AS "presentNow",
		       (SELECT count(*) FROM "Employee" WHERE status = 'ACTIVE' AND role::text NOT IN ('ADMIN', 'OWNER'))::int AS "activeStaff",
		       (SELECT count(*) FROM "ExtraTask" WHERE status IN ('NEW', 'IN_PROGRESS'))::int AS "openTasks"
		FROM "Booking" b WHERE `+drillBase())
	return &t, err
}

// ActiveEmployees للمطابقة بالاسم بمسودة المهمة.
func (r *SecretaryRepository) ActiveEmployees() ([]SecretaryEmployeeHit, error) {
	rows := []SecretaryEmployeeHit{}
	err := r.db.Select(&rows, `SELECT id, name, role::text AS role, phone, "isLeader" FROM "Employee"
		WHERE status = 'ACTIVE' AND role::text NOT IN ('OWNER') ORDER BY name`)
	return rows, err
}

// SetResult نتيجة التنفيذ بعد ما انقفل المقترح SENT.
func (r *SecretaryRepository) SetResult(id, empID, result string) error {
	_, err := r.db.Exec(`UPDATE "SecretaryAction" SET result = $3 WHERE id = $1 AND "employeeId" = $2`, id, empID, result)
	return err
}

// MarkFailed إذا انقفل SENT والتنفيذ فشل — السجل يگول الحقيقة.
func (r *SecretaryRepository) MarkFailed(id, empID, result string) error {
	_, err := r.db.Exec(`UPDATE "SecretaryAction" SET status = 'FAILED', result = $3 WHERE id = $1 AND "employeeId" = $2`, id, empID, result)
	return err
}

// ── «مذكراتي» و«جدول أعمالي» (دازاي 10-23) — كل مستخدم ملاحظاته بس ──

type SecretaryNote struct {
	ID        string     `db:"id" json:"id"`
	Kind      string     `db:"kind" json:"kind"` // DIARY | AGENDA
	Body      string     `db:"body" json:"body"`
	At        *time.Time `db:"at" json:"at"`
	DoneAt    *time.Time `db:"doneAt" json:"doneAt"`
	CreatedAt time.Time  `db:"createdAt" json:"createdAt"`
}

const secNoteCols = `id, kind, body, "at", "doneAt", "createdAt"`

func (r *SecretaryRepository) AddNote(owner, kind, body string, at *time.Time) (*SecretaryNote, error) {
	var n SecretaryNote
	err := r.db.Get(&n, `INSERT INTO "SecretaryNote" (id, "ownerId", kind, body, "at")
		VALUES ($1,$2,$3,$4,$5) RETURNING `+secNoteCols, uuid.NewString(), owner, kind, body, at)
	return &n, err
}

// Notes ملاحظات نوع معيّن، الأحدث أول. q اختياري (بحث بالنص).
func (r *SecretaryRepository) Notes(owner, kind, q string, limit int) ([]SecretaryNote, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows := []SecretaryNote{}
	err := r.db.Select(&rows, `SELECT `+secNoteCols+` FROM "SecretaryNote"
		WHERE "ownerId" = $1 AND ($2 = '' OR kind = $2) AND ($3 = '' OR body ILIKE '%' || $3 || '%')
		ORDER BY "createdAt" DESC LIMIT $4`, owner, kind, strings.TrimSpace(q), limit)
	return rows, err
}

func (r *SecretaryRepository) SetNoteDone(owner, id string, done bool) error {
	res, err := r.db.Exec(`UPDATE "SecretaryNote" SET "doneAt" = CASE WHEN $3 THEN now() END
		WHERE id = $1 AND "ownerId" = $2`, id, owner, done)
	return noteHit(res, err)
}

func (r *SecretaryRepository) DeleteNote(owner, id string) error {
	res, err := r.db.Exec(`DELETE FROM "SecretaryNote" WHERE id = $1 AND "ownerId" = $2`, id, owner)
	return noteHit(res, err)
}

func noteHit(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("الملاحظة مو موجودة")
	}
	return nil
}

// AgendaItem بند بـ«جدول أعمالي»: مقابلة/ضيف انحدد موعده، أو شي سجّله المستخدم.
type AgendaItem struct {
	ID     string     `db:"id" json:"id"`
	Kind   string     `db:"kind" json:"kind"` // meeting | guest | mine
	Title  string     `db:"title" json:"title"`
	Who    string     `db:"who" json:"who"`
	Note   string     `db:"note" json:"note"`
	At     *time.Time `db:"at" json:"at"`
	DoneAt *time.Time `db:"doneAt" json:"doneAt"`
}

// Agenda من بداية اليوم (بغداد) وبعد، ويا بنوده الي ما إلها وقت ولسه ما خلصت.
func (r *SecretaryRepository) Agenda(owner string) ([]AgendaItem, error) {
	rows := []AgendaItem{}
	err := r.db.Select(&rows, `
		SELECT m.id, lower(m.kind) AS kind,
		       CASE WHEN m.kind = 'GUEST' THEN 'ضيف: ' || COALESCE(m."guestName", '') ELSE 'مقابلة' END AS title,
		       COALESCE(e.name, '') AS who, COALESCE(m.reason, '') AS note, m."scheduledAt" AS "at", NULL::timestamptz AS "doneAt"
		FROM "MeetingRequest" m LEFT JOIN "Employee" e ON e.id = m."employeeId"
		WHERE m.status = 'SCHEDULED' AND baghdad_date(m."scheduledAt") >= baghdad_today()
		UNION ALL
		SELECT n.id, 'mine', n.body, '', '', n."at", n."doneAt"
		FROM "SecretaryNote" n
		WHERE n."ownerId" = $1 AND n.kind = 'AGENDA'
		  AND ((n."at" IS NOT NULL AND baghdad_date(n."at") >= baghdad_today()) OR (n."at" IS NULL AND n."doneAt" IS NULL))
		ORDER BY "at" NULLS LAST LIMIT 100`, owner)
	return rows, err
}

// SaveAudio يحفظ بصمة المالك ويا رسالته (تنمسح ويا الرسالة).
func (r *SecretaryRepository) SaveAudio(messageID, ct string, data []byte) {
	_, _ = r.db.Exec(`INSERT INTO "SecretaryAudio" ("messageId", "contentType", data) VALUES ($1, $2, $3)`, messageID, ct, data)
}

// Audio بصمة رسالة — لصاحبها بس.
func (r *SecretaryRepository) Audio(empID, messageID string) (string, []byte, error) {
	var row struct {
		CT   string `db:"contentType"`
		Data []byte `db:"data"`
	}
	err := r.db.Get(&row, `SELECT au."contentType", au.data FROM "SecretaryAudio" au
		JOIN "SecretaryMessage" m ON m.id = au."messageId" WHERE au."messageId" = $1 AND m."employeeId" = $2`, messageID, empID)
	return row.CT, row.Data, err
}
