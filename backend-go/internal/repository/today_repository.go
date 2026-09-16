package repository

import (
	"github.com/jmoiron/sqlx"
)

// ═══ لوحة اليوم ═══
//
// «خل نضيف ملاعيب وترتيبات وزينة»: لوحة «شغلي اليوم» + «شنو صاير
// اليوم» + لوحة الشرف + مخطط الحركة.
//
// الإداري يفتح النظام الصبح ويسأل نفس الأسئلة كل يوم: شكد حجز اليوم؟
// منو بالميدان هسه؟ شكد خلّصوا؟ شنو ينتظرني؟ وچان لازم يفتح أربع
// شاشات حتى يجاوبهن.
//
// ⚠️ الأرقام كلها بنداء **واحد**: أربع بطاقات + لوحة شرف + مخطط
// أربعتعش يوم = ست نداءات لو انفصلن، بكل فتحة صفحة رئيسية ولكل
// موظف. والصفحة الرئيسية أكثر شاشة تنفتح بالنظام.
type TodayRepository struct {
	db *sqlx.DB
}

func NewTodayRepository(db *sqlx.DB) *TodayRepository {
	return &TodayRepository{db: db}
}

// TopCrewMember صف بلوحة الشرف — الطلعات هي المقياس (مو الحجوزات).
type TopCrewMember struct {
	EmployeeID string  `db:"employeeId" json:"employeeId"`
	Name       string  `db:"name" json:"name"`
	PhotoURL   *string `db:"photoUrl" json:"photoUrl"`
	Visits     int     `db:"visits" json:"visits"`
	Done       int     `db:"done" json:"done"`
}

// DayPoint نقطة بمخطط آخر أربعتعش يوم.
type DayPoint struct {
	Day   string `db:"day" json:"day"`
	Count int    `db:"count" json:"count"`
}

// TodayBoard كل الي تعرضه لوحة اليوم.
type TodayBoard struct {
	// شنو صاير اليوم
	BookingsToday  int `json:"bookingsToday"`
	InField        int `json:"inField"`
	CompletedToday int `json:"completedToday"`
	NewToday       int `json:"newToday"`
	// شغلي اليوم — الي ينتظر تصرّف
	NeedsContact int `json:"needsContact"`
	NeedsCrew    int `json:"needsCrew"`
	NeedsPaper   int `json:"needsPaper"`
	NeedsFinish  int `json:"needsFinish"`

	// ⚠️ نافذة ٧ أيام متدحرجة تنتهي اليوم — تتفادى نقاش «أول يوم
	// بالأسبوع» (أحد أو اثنين)، ونفس أسلوب `Last14` بالملف.
	WeekTotal     int `json:"weekTotal"`
	WeekCompleted int `json:"weekCompleted"`

	TopCrew []TopCrewMember `json:"topCrew"`
	Last14  []DayPoint      `json:"last14"`
}

func (r *TodayRepository) Board() (*TodayBoard, error) {
	b := &TodayBoard{TopCrew: []TopCrewMember{}, Last14: []DayPoint{}}

	// ⚠️ «اليوم» بتوقيت بغداد مو بتوقيت السيرفر: الفرق يخلّي حجوزات
	// الليل تنحسب على يوم باچر، والإداري يشوف رقماً ما يطابق الي
	// قدّامه.
	row := struct {
		BookingsToday  int `db:"bookingsToday"`
		InField        int `db:"inField"`
		CompletedToday int `db:"completedToday"`
		NewToday       int `db:"newToday"`
		NeedsContact   int `db:"needsContact"`
		NeedsCrew      int `db:"needsCrew"`
		NeedsPaper     int `db:"needsPaper"`
		NeedsFinish    int `db:"needsFinish"`
		WeekTotal      int `db:"weekTotal"`
		WeekCompleted  int `db:"weekCompleted"`
	}{}
	err := r.db.Get(&row, `
		SELECT
		  COUNT(*) FILTER (WHERE baghdad_date(b."scheduledAt") = baghdad_date(now())) AS "bookingsToday",
		  COUNT(*) FILTER (WHERE b.status = 'IN_PROGRESS' OR (b."startedAt" IS NOT NULL AND b."completedAt" IS NULL)) AS "inField",
		  COUNT(*) FILTER (WHERE baghdad_date(b."completedAt") = baghdad_date(now())) AS "completedToday",
		  COUNT(*) FILTER (WHERE baghdad_date(b."createdAt") = baghdad_date(now())) AS "newToday",
		  COUNT(*) FILTER (WHERE `+bucketCondition("pending")+`) AS "needsContact",
		  COUNT(*) FILTER (WHERE `+bucketCondition("confirmed")+`) AS "needsCrew",
		  -- منجز وناقصه فاتورة أو تقرير: هذا الي يجيب الغرامات
		  COUNT(*) FILTER (WHERE b.status = 'COMPLETED' AND b."settledLegacyAt" IS NULL
		                     AND (NOT `+hasInvoiceSQL+` OR NOT `+hasReportSQL+`)) AS "needsPaper",
		  COUNT(*) FILTER (WHERE b.status = 'PARTIAL') AS "needsFinish",
		  COUNT(*) FILTER (WHERE baghdad_date(b."createdAt")
		             BETWEEN baghdad_date(now()) - interval '6 days' AND baghdad_date(now())) AS "weekTotal",
		  COUNT(*) FILTER (WHERE b.status = 'COMPLETED' AND baghdad_date(b."createdAt")
		             BETWEEN baghdad_date(now()) - interval '6 days' AND baghdad_date(now())) AS "weekCompleted"
		FROM "Booking" b WHERE b."archivedAt" IS NULL` + NotDeletePendingSQL("b"))
	if err != nil {
		return nil, err
	}
	b.BookingsToday, b.InField, b.CompletedToday, b.NewToday = row.BookingsToday, row.InField, row.CompletedToday, row.NewToday
	b.NeedsContact, b.NeedsCrew, b.NeedsPaper, b.NeedsFinish = row.NeedsContact, row.NeedsCrew, row.NeedsPaper, row.NeedsFinish
	b.WeekTotal, b.WeekCompleted = row.WeekTotal, row.WeekCompleted

	// ═══ لوحة الشرف ═══
	// ⚠️ المقياس **الطلعات** مو الحجوزات: هذا الي اتفقنا عليه لمن
	// صلّحنا الإنتاجية — الحجز الي ياخذ أربع أيام أربع طلعات، وكل
	// طلعة إلها كادرها.
	_ = r.db.Select(&b.TopCrew, `
		SELECT e.id AS "employeeId", e.name, e."photoUrl",
		       COUNT(*) AS visits,
		       COUNT(*) FILTER (WHERE v.outcome = 'DONE') AS done
		FROM "BookingVisitCrew" vc
		JOIN "BookingVisit" v ON v.id = vc."visitId"
		JOIN "Employee" e ON e.id = vc."employeeId"
		WHERE to_char(v."occurredAt", 'YYYY-MM') = to_char(now(), 'YYYY-MM')
		  AND e.status = 'ACTIVE'
		GROUP BY e.id, e.name, e."photoUrl"
		ORDER BY visits DESC, done DESC
		LIMIT 3`)

	// ═══ مخطط آخر أربعتعش يوم ═══
	// ⚠️ `generate_series` يخلّي اليوم الفاضي يطلع صفراً بدل ما ينشال
	// من المخطط — الخط الي يقفز فوگ الأيام الفاضية يكذب على العين.
	_ = r.db.Select(&b.Last14, `
		SELECT to_char(d.day, 'YYYY-MM-DD') AS day,
		       COUNT(b.id) AS count
		FROM generate_series(baghdad_date(now()) - interval '13 days', baghdad_date(now()), interval '1 day') AS d(day)
		LEFT JOIN "Booking" b ON b."archivedAt" IS NULL
		     AND NOT ` + BookingDeletePendingSQL("b") + `
		     AND baghdad_date(b."createdAt") = d.day::date
		GROUP BY d.day ORDER BY d.day`)

	return b, nil
}
