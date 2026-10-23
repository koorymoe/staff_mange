package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ═══ «الرقم = القائمة» — قرار (ع) 10-16 ═══
//
// المدير لگى بطاقات اللوحة الرئيسية تگول «٥ حجوزات اليوم» ومن يضغط
// تفتحله شاشة بيها مئات المهام، و«منجزة وناقصها ورق ٨٦» تفتح تبويباً
// فارغاً. السبب: الرقم ينحسب بشرط، والرابط يفتح شاشة عامة بشرط ثاني.
//
// الحل الجذري: **تعريف واحد لكل رقم** (DrillDefs). نفس الشرط يعدّ
// البطاقة (DrillCountSQL) ويطلّع قائمتها (Drill). فلو أحد غيّر الشرط
// يتغيّر بالمكانين سوه، وما يرجع يصير رقم بلا قائمة.

// DrillDef تعريف رقم واحد من أرقام اللوحة.
type DrillDef struct {
	Key     string
	Title   string
	Explain string // سطر عراقي: شنو هذا الرقم ومن يا فترة
	Where   string // شرط على الحجز b (فوق شرط «المحسوب» العام)
	AgeFrom string // من وين ينحسب «صارله» — عمود على b
}

// الشرط العام لكل رقم: مو محذوف ولا مطلوب حذفه.
func drillBase() string { return BookingCountableSQL("b") }

const openMissionSQL = `EXISTS (SELECT 1 FROM "Mission" m WHERE m."bookingId" = b.id
	AND (` + missionStageExpr + `) NOT IN ('COMPLETED','STOPPED'))`

// ═══ «المتأخرات» — تعريف (ع) 10-23 ═══
// «تصنيف كلمة متأخر هو شنو؟» — التأخير لكل مرحلة من لحظتها، مو من موعد الحجز وبس:
//   ١ تواصل: حجز جديد فات عليه ساعتين وما أحد حچى ويا الزبون.
//   ٢ انطلاق: الموعد فات بنص ساعة والكادر ما طلع («الليدر مجهزليه بالأربع ويطلع بالخمس»).
//   ٣ تنفيذ: الشغل بدا وطوّل أكثر من ٢٤ ساعة وما خلص.
//   ٤ ورق: من «تم الإنجاز» يبدي يحسب — فات ٢٤ ساعة وناقصه فاتورة أو تقرير.
// و«بانتظار قرار الزبون» **مو متأخر** (محطته وحدها، ويتأرشف استفسار بعد ٥ أيام).
// الحجز الي موعده بعدين **مو متأخر** لحد ما يجي موعده.
var (
	lateContactSQL = `(` + bucketCondition("pending") + `) AND b."confirmationContactedAt" IS NULL
		AND b."createdAt" < (now() AT TIME ZONE 'UTC') - interval '2 hours'`
	lateDepartSQL = `(` + bucketCondition("assigned") + `) AND NOT ` + startedSQL + ` AND b."departedAt" IS NULL
		AND b."scheduledAt" IS NOT NULL AND b."scheduledAt" < (now() AT TIME ZONE 'UTC') - interval '30 minutes'`
	lateExecSQL = `(` + bucketCondition("assigned") + `) AND ` + startedSQL + `
		AND COALESCE(b."startedAt", b."arrivedAt", b."departedAt") < (now() AT TIME ZONE 'UTC') - interval '24 hours'`
	latePaperSQL = `(` + bucketCondition("needs_paper") + `)
		AND b."completedAt" < (now() AT TIME ZONE 'UTC') - interval '24 hours'
		AND b."completedAt" > (now() AT TIME ZONE 'UTC') - interval '60 days'`
	overdueSQL = `((` + lateContactSQL + `) OR (` + lateDepartSQL + `) OR (` + lateExecSQL + `) OR (` + latePaperSQL + `))`
	// overdueKindSQL نوع التأخير — يطلع بعمود «وين واكف» بصفحة المتأخرات.
	overdueKindSQL = `CASE
		WHEN ` + lateContactSQL + ` THEN '📞 تأخير تواصل — ما أحد حچى ويا الزبون'
		WHEN ` + lateDepartSQL + ` THEN '🚚 الكادر ما طلع — الموعد فات'
		WHEN ` + lateExecSQL + ` THEN '⏳ التنفيذ طوّل أكثر من ٢٤ ساعة'
		WHEN ` + latePaperSQL + ` THEN '🧾 خلص وناقصه ' || concat_ws(' و',
			CASE WHEN NOT EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id AND li."revokedAt" IS NULL) THEN 'فاتورة' END,
			CASE WHEN NOT EXISTS (SELECT 1 FROM "WorkReport" w WHERE w."bookingId" = b.id) THEN 'تقرير' END)
		END`
	// overdueSinceSQL من يمته يحسب التأخير — حسب نوعه.
	overdueSinceSQL = `CASE
		WHEN ` + lateContactSQL + ` THEN b."createdAt"
		WHEN ` + lateDepartSQL + ` THEN b."scheduledAt"
		WHEN ` + lateExecSQL + ` THEN COALESCE(b."startedAt", b."arrivedAt", b."departedAt")
		ELSE b."completedAt" END`
)

const needsCoordSQL = `b.status = 'CONFIRMED' AND (b."scheduledAt" IS NULL OR NOT ` + hasCrewSQL + `)`

const weekSQL = `baghdad_date(b."createdAt") BETWEEN baghdad_today() - 6 AND baghdad_today()`

var drillDefs = []DrillDef{
	{Key: "today_scheduled", Title: "حجوزات اليوم", AgeFrom: `"scheduledAt"`,
		Explain: "الحجوزات الي موعدها اليوم (بتوقيت بغداد) — شكد ما جانت حالتها.",
		Where:   `baghdad_date(b."scheduledAt") = baghdad_today()`},
	{Key: "new_today", Title: "حجوزات وصلت اليوم", AgeFrom: `"createdAt"`,
		Explain: "الحجوزات الي انسجّلت بالنظام اليوم — حجز جديد دخل، مو موعده اليوم.",
		Where:   `baghdad_date(b."createdAt") = baghdad_today()`},
	{Key: "completed_today", Title: "خلّصوا اليوم", AgeFrom: `"completedAt"`,
		Explain: "الحجوزات الي تأشّرت «منجز» اليوم.",
		Where:   `baghdad_date(b."completedAt") = baghdad_today()`},
	{Key: "in_field", Title: "بالميدان هسه", AgeFrom: `"startedAt"`,
		Explain: "الكادر انطلق أو وصل أو بدا الشغل اليوم، وبعده ما خلص ولا توقف.",
		Where:   InFieldSQL},
	{Key: "open_missions", Title: "المهام المفتوحة", AgeFrom: `"scheduledAt"`,
		Explain: "حجوزات عليها مهمة ميدانية بعدها ما خلصت ولا توقفت — من أول ما تنسند لحد الإنجاز. (كل الفترات، مو اليوم بس.)",
		Where:   openMissionSQL},
	{Key: "overdue", Title: "المتأخرات", AgeFrom: `"scheduledAt"`,
		Explain: "كل مرحلة تتأخر من لحظتها: حجز جديد بلا تواصل +ساعتين · الموعد فات والكادر ما طلع · التنفيذ طوّل +٢٤ ساعة · خلص وناقصه فاتورة أو تقرير +٢٤ ساعة. بانتظار قرار الزبون مو متأخر.",
		Where:   overdueSQL},
	{Key: "needs_coord", Title: "حجوزات تنتظر تنسيق", AgeFrom: `"confirmedAt"`,
		Explain: "حجوزات حالتها «مثبّتة» بس ناقصها موعد أو كادر. عمود «وين واكف» يگلك ليش ما تنسّقت (الزبون ما رد، عند المشاريع، مؤجل…). المسؤول: الإداري المنسّق.",
		Where:   needsCoordSQL},
	{Key: "needs_contact", Title: "ينتظرون تواصل وتثبيت", AgeFrom: `"createdAt"`,
		Explain: "انسجّلت وما أحد ثبّتها ويا الزبون بعد — محطة «بانتظار التثبيت» بالحجوزات.",
		Where:   bucketCondition("pending")},
	{Key: "needs_crew", Title: "ينتظرون كادراً وموعداً", AgeFrom: `"confirmedAt"`,
		Explain: "انثبّتت ويا الزبون وما انحطلها كادر — محطة «تم التثبيت بحاجة لكادر». (الي الزبون ما رد أو مؤجلة تطلع بمحطتها، مو هنا.)",
		Where:   bucketCondition("confirmed")},
	{Key: "needs_paper", Title: "منجزة وناقصها ورق", AgeFrom: `"completedAt"`,
		Explain: "الشغل خلص بس ناقص الفاتورة أو التقرير — هاي الي تجيب الغرامات. (كل الفترات.)",
		Where:   bucketCondition("needs_paper")},
	{Key: "needs_finish", Title: "تحتاج موعد إكمال", AgeFrom: `"lastPartialAt"`,
		Explain: "الكادر طلع وأنجز جزء، والشغل ما كمل — يحتاج موعد ثاني.",
		Where:   `b.status = 'PARTIAL'`},
	{Key: "week", Title: "حجوزات آخر ٧ أيام", AgeFrom: `"createdAt"`,
		Explain: "كل حجز انسجّل خلال آخر ٧ أيام (اليوم و٦ أيام قبله). نسبة «نجاح الأسبوع» = المنجز منها ÷ كلها.",
		Where:   weekSQL},
	{Key: "unverified", Title: "مبالغ تنتظر التدقيق", AgeFrom: `"completedAt"`,
		Explain: "حجوزات منجزة (مو شغل داخلي) المحاسب بعده ما دقّق المبلغ المستلم — من كل الأيام، مو اليوم بس.",
		Where:   `b.status = 'COMPLETED' AND NOT b."amountVerified" AND b."bookingType" <> 'INTERNAL'`},
	{Key: "verified", Title: "مبالغ مدقّقة", AgeFrom: `"completedAt"`,
		Explain: "حجوزات منجزة (مو شغل داخلي) المحاسب دقّق مبلغها — من كل الأيام.",
		Where:   `b.status = 'COMPLETED' AND b."amountVerified" AND b."bookingType" <> 'INTERNAL'`},
	{Key: "in_progress", Title: "قيد التنفيذ", AgeFrom: `"startedAt"`,
		Explain: "حجوزات حالتها «قيد التنفيذ» — الكادر باشر وبعده ما خلص.",
		Where:   `b.status = 'IN_PROGRESS'`},
	{Key: "inhouse_month", Title: "شغل داخل الشركة — هذا الشهر", AgeFrom: `"completedAt"`,
		Explain: "حجوزات انخلصت هذا الشهر والشغل صار جوّا الشركة (الزبون جاب الجهاز).",
		Where:   `b.status = 'COMPLETED' AND b."workLocation" = 'IN_HOUSE' AND to_char(baghdad_date(b."completedAt"), 'YYYY-MM') = to_char(baghdad_today(), 'YYYY-MM')`},
	{Key: "onsite_month", Title: "شغل بالموقع — هذا الشهر", AgeFrom: `"completedAt"`,
		Explain: "حجوزات انخلصت هذا الشهر والكادر راح لموقع الزبون.",
		Where:   `b.status = 'COMPLETED' AND b."workLocation" <> 'IN_HOUSE' AND to_char(baghdad_date(b."completedAt"), 'YYYY-MM') = to_char(baghdad_today(), 'YYYY-MM')`},
}

// DrillDefs كل التعريفات (للاختبار وللواجهة).
func DrillDefs() []DrillDef { return drillDefs }

// DrillByKey يرجّع التعريف أو nil.
func DrillByKey(key string) *DrillDef {
	for i := range drillDefs {
		if drillDefs[i].Key == key {
			return &drillDefs[i]
		}
	}
	return nil
}

// DrillCountSQL تعبير COUNT للبطاقة — نفس شرط القائمة بالضبط.
func DrillCountSQL(key string) string {
	d := DrillByKey(key)
	if d == nil {
		panic("drill: unknown key " + key)
	}
	return `COUNT(*) FILTER (WHERE ` + d.Where + `)`
}

// DrillLateSQL نفس DrillCountSQL بس الي فات عليه أكثر من ٢٤ ساعة من عمره
// (AgeFrom) — للواجبات: الموظف يتحاسب على المتأخر، مو على الي وصل هسه.
func DrillLateSQL(key string) string {
	d := DrillByKey(key)
	if d == nil {
		panic("drill: unknown key " + key)
	}
	return `COUNT(*) FILTER (WHERE (` + d.Where + `) AND b.` + d.AgeFrom + ` < (now() AT TIME ZONE 'UTC') - interval '24 hours')`
}

// crewSQL الكادر كامل (ع 10-23: «لازم يطلع الليدر مال الحجز مع الفني»): المكلّفين،
// وليدر المهمة وأعضاؤها، ومشرف المشروع — والليدر أول بنجمة.
const crewSQL = `(SELECT string_agg(CASE WHEN e."isLeader" THEN '⭐ ' || e.name ELSE e.name END, '، '
		ORDER BY e."isLeader" DESC, e.name)
	FROM "Employee" e WHERE e.id IN (
		SELECT ba."employeeId" FROM "BookingAssignment" ba WHERE ba."bookingId" = b.id
		UNION SELECT m."leaderId" FROM "Mission" m WHERE m."bookingId" = b.id AND m."leaderId" IS NOT NULL
		UNION SELECT unnest(m."memberIds") FROM "Mission" m WHERE m."bookingId" = b.id
		UNION SELECT b."projectSupervisorId" WHERE b."projectSupervisorId" IS NOT NULL))`

// DrillRow صف بالقائمة.
type DrillRow struct {
	ID          string  `db:"id" json:"id"`
	Code        string  `db:"code" json:"code"`
	Customer    *string `db:"customer" json:"customer"`
	Phone       *string `db:"phone" json:"phone"`
	Service     *string `db:"service" json:"service"`
	Status      string  `db:"status" json:"status"`
	ScheduledAt *string `db:"scheduledAt" json:"scheduledAt"`
	CreatedAt   string  `db:"createdAt" json:"createdAt"`
	Crew        *string `db:"crew" json:"crew"`
	Responsible *string `db:"responsible" json:"responsible"`
	AgeDays     *int    `db:"ageDays" json:"ageDays"`
	Station     string  `db:"station" json:"station"`
	Missing     *string `db:"missing" json:"missing"`
}

// DrillResult القائمة ويا عنوانها وشرحها.
type DrillResult struct {
	Key     string     `json:"key"`
	Title   string     `json:"title"`
	Explain string     `json:"explain"`
	Count   int        `json:"count"`
	Rows    []DrillRow `json:"rows"`
	Extra   any        `json:"extra,omitempty"`
}

// stationCaseSQL «وين واكف الحجز» — نفس محطات شاشة الحجوزات.
func stationCaseSQL() string {
	return `CASE
		WHEN ` + deletePendingSQL + ` THEN 'ينتظر قرار حذف'
		WHEN ` + atProjectsSQL + ` THEN 'عند إدارة المشاريع'
		WHEN ` + atTechSQL + ` THEN 'مرحّل للتقني'
		WHEN b.status = 'COMPLETED' THEN 'منجز'
		WHEN b.status = 'CANCELLED' THEN 'ملغى'
		WHEN b.status = 'PARTIAL' THEN 'إنجاز جزئي'
		WHEN b."workStoppedAt" IS NOT NULL THEN 'متوقف'
		WHEN b."waitingSince" IS NOT NULL AND b."waitingKind" = 'CUSTOMER_DECISION' THEN 'ينتظر قرار الزبون'
		WHEN b."waitingSince" IS NOT NULL THEN 'الزبون ما رد'
		WHEN b."awaitingReschedule" THEN 'مؤجل — ينتظر موعد'
		WHEN ` + startedSQL + ` THEN 'بالتنفيذ'
		WHEN ` + hasCrewSQL + ` THEN 'مكلّف — عليه كادر'
		WHEN b."confirmedAt" IS NOT NULL THEN 'مثبّت — ينتظر كادر'
		ELSE 'بانتظار التثبيت' END`
}

// Drill يطلّع قائمة رقم واحد — نفس شرط عدّه.
func Drill(db *sqlx.DB, key string, limit int) (*DrillResult, error) {
	d := DrillByKey(key)
	if d == nil {
		return nil, fmt.Errorf("unknown drill %q", key)
	}
	res := &DrillResult{Key: d.Key, Title: d.Title, Explain: d.Explain, Rows: []DrillRow{}}
	where := drillBase() + ` AND (` + d.Where + `)`
	if err := db.Get(&res.Count, `SELECT COUNT(*) FROM "Booking" b WHERE `+where); err != nil {
		return nil, err
	}
	ageExpr, stationExpr, orderExpr := `b.`+d.AgeFrom, stationCaseSQL(), `b.`+d.AgeFrom
	if d.Key == "overdue" {
		ageExpr, stationExpr, orderExpr = overdueSinceSQL, overdueKindSQL, overdueSinceSQL
	}
	q := `
		SELECT b.id, b.code, c.name AS customer, c.phone, s.name AS service, b.status::text AS status,
		       to_char(b."scheduledAt" AT TIME ZONE 'Asia/Baghdad', 'YYYY-MM-DD HH24:MI') AS "scheduledAt",
		       to_char(b."createdAt" AT TIME ZONE 'Asia/Baghdad', 'YYYY-MM-DD HH24:MI') AS "createdAt",
		       ` + crewSQL + ` AS crew,
		       COALESCE(b."confirmedByName", (SELECT e.name FROM "Employee" e WHERE e.id = b."createdById")) AS responsible,
		       (baghdad_today() - baghdad_date(` + ageExpr + `))::int AS "ageDays",
		       ` + stationExpr + ` AS station,
		       NULLIF(concat_ws(' + ',
		         CASE WHEN b."scheduledAt" IS NULL THEN 'بلا موعد' END,
		         CASE WHEN NOT ` + hasCrewSQL + ` THEN 'بلا كادر' END), '') AS missing
		FROM "Booking" b
		LEFT JOIN "Customer" c ON c.id = b."customerId"
		LEFT JOIN "Service" s ON s.id = b."serviceId"
		WHERE ` + where + `
		ORDER BY ` + orderExpr + ` ASC NULLS LAST
		LIMIT $1`
	if err := db.Select(&res.Rows, q, limit); err != nil {
		return nil, err
	}
	// «نجاح الأسبوع»: المنجز منها — حتى الصفحة تبيّن الحساب بالأرقام.
	if key == "week" {
		var done int
		if err := db.Get(&done, `SELECT COUNT(*) FROM "Booking" b WHERE `+where+` AND b.status = 'COMPLETED'`); err != nil {
			return nil, err
		}
		res.Extra = map[string]int{"completed": done, "total": res.Count}
	}
	return res, nil
}
