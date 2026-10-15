package handler

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/xuri/excelize/v2"
)

// ═══ فحص المتأخرات — قرار (ع) 10-15 ═══
// «اريد تراجع وياي كل الحجوزات الي يطلع عليهن متأخرة: ليش متأخرات، هل بيهن
// مشكلة، وشكد مدة التأخير». نفس تعريف ماتركس بالضبط (matrix_command_repository)
// حتى الرقم يطابق، ولكل حجز سبب مرجّح من الأدلة، ومشاكل البيانات معلّمة.

type LateAuditHandler struct{ db *sqlx.DB }

func NewLateAuditHandler(db *sqlx.DB) *LateAuditHandler { return &LateAuditHandler{db: db} }

type lateRow struct {
	ID             string     `db:"id" json:"id"`
	Code           string     `db:"code" json:"code"`
	Customer       *string    `db:"customer" json:"customer"`
	Service        *string    `db:"service" json:"service"`
	Division       string     `db:"division" json:"division"`
	Status         string     `db:"status" json:"status"`
	ScheduledAt    time.Time  `db:"scheduledAt" json:"scheduledAt"`
	CompletedAt    *time.Time `db:"completedAt" json:"completedAt"`
	StartedAt      *time.Time `db:"startedAt" json:"startedAt"`
	ArrivedAt      *time.Time `db:"arrivedAt" json:"arrivedAt"`
	DepartedAt     *time.Time `db:"departedAt" json:"departedAt"`
	ContactedAt    *time.Time `db:"confirmationContactedAt" json:"contactedAt"`
	Unstaffed      bool       `db:"unstaffed" json:"unstaffed"`
	PartialCount   int        `db:"partialCount" json:"partialCount"`
	WorkStopReason *string    `db:"workStopReason" json:"workStopReason"`
	PostponeReason *string    `db:"postponeReason" json:"postponeReason"`
	Postpones      int        `db:"postpones" json:"postpones"`
	Reschedules    int        `db:"reschedules" json:"reschedules"`
	LastChangeAt   *time.Time `db:"lastChangeAt" json:"lastChangeAt"`
	OpenMaterials  int        `db:"openMaterials" json:"openMaterials"`
	Leader         *string    `db:"leader" json:"leader"`
	Team           *string    `db:"team" json:"team"`
	Note           *string    `db:"note" json:"note"`
	DaysLate       int        `db:"daysLate" json:"daysLate"`
	Cause          string     `db:"-" json:"cause"`
	CauseDetail    string     `db:"-" json:"causeDetail"`
	DataIssues     []string   `db:"-" json:"dataIssues"`
}

const lateAuditSQL = `
	SELECT b.id, b.code::text AS code, c.name AS customer, s.name AS service, COALESCE(s.division, 'ENGINEERING') AS division,
		b.status::text AS status, b."scheduledAt", b."completedAt", b."startedAt", b."arrivedAt", b."departedAt", b."confirmationContactedAt",
		NOT EXISTS (SELECT 1 FROM "BookingAssignment" a WHERE a."bookingId" = b.id) AS unstaffed,
		b."partialCount", b."workStopReason", b."postponeReason",
		(SELECT count(*) FROM "ScheduleChangeLog" l WHERE l."bookingId" = b.id AND l.kind = 'POSTPONE')::int AS postpones,
		(SELECT count(*) FROM "ScheduleChangeLog" l WHERE l."bookingId" = b.id)::int AS reschedules,
		(SELECT max(l."createdAt") FROM "ScheduleChangeLog" l WHERE l."bookingId" = b.id) AS "lastChangeAt",
		(SELECT count(*) FROM "ProcurementRequest" p WHERE p."bookingId" = b.id AND p.status::text NOT IN ('FULFILLED', 'REJECTED', 'CANCELLED'))::int AS "openMaterials",
		COALESCE((SELECT e.name FROM "Employee" e WHERE e.id = b."projectSupervisorId"),
			(SELECT e.name FROM "Mission" m JOIN "Employee" e ON e.id = m."leaderId" WHERE m."bookingId" = b.id ORDER BY m."assignedAt" DESC NULLS LAST LIMIT 1)) AS leader,
		(SELECT string_agg(e.name, '، ') FROM "BookingAssignment" a JOIN "Employee" e ON e.id = a."employeeId" WHERE a."bookingId" = b.id) AS team,
		(SELECT NULLIF(btrim(me.note), '') FROM "Mission" m JOIN "MissionEvent" me ON me."missionId" = m.id
			WHERE m."bookingId" = b.id AND me.note IS NOT NULL AND btrim(me.note) <> '' ORDER BY me."createdAt" DESC LIMIT 1) AS note,
		(COALESCE(baghdad_date(b."completedAt"), baghdad_today()) - baghdad_date(b."scheduledAt"))::int AS "daysLate"
	FROM "Booking" b
	LEFT JOIN "Customer" c ON c.id = b."customerId"
	LEFT JOIN "Service" s ON s.id = b."serviceId"
	WHERE baghdad_date(b."scheduledAt") BETWEEN $1::date AND $2::date
	  AND b.status::text <> 'CANCELLED' AND b."archivedAt" IS NULL AND b."bookingType" IS DISTINCT FROM 'INTERNAL'
	  AND ((b.status = 'COMPLETED' AND baghdad_date(b."completedAt") > baghdad_date(b."scheduledAt"))
	       OR (b.status::text NOT IN ('COMPLETED','CANCELLED') AND baghdad_date(b."scheduledAt") < baghdad_today()))
	ORDER BY "daysLate" DESC, b."scheduledAt"`

// classifyLate السبب المرجّح من الأدلة — بالترتيب: الأقوى دليلاً أول.
func classifyLate(r *lateRow, loc *time.Location) {
	day := func(t time.Time) string { return t.In(loc).Format("2006-01-02") }
	sched := day(r.ScheduledAt)
	switch {
	case r.Postpones > 0 || (r.PostponeReason != nil && *r.PostponeReason != ""):
		r.Cause = "POSTPONED"
		if r.PostponeReason != nil {
			r.CauseDetail = *r.PostponeReason
		}
	case r.Unstaffed:
		r.Cause = "NO_CREW"
	case r.Status == "PENDING" && r.ContactedAt == nil:
		r.Cause = "NOT_CONFIRMED"
	case r.OpenMaterials > 0:
		r.Cause = "MATERIALS"
		r.CauseDetail = strconv.Itoa(r.OpenMaterials) + " طلب مواد مفتوح"
	case r.WorkStopReason != nil && *r.WorkStopReason != "":
		r.Cause = "WORK_STOPPED"
		r.CauseDetail = *r.WorkStopReason
	case r.PartialCount > 0:
		r.Cause = "MULTI_DAY"
		r.CauseDetail = strconv.Itoa(r.PartialCount) + " إنجاز جزئي"
	case (r.DepartedAt != nil && day(*r.DepartedAt) > sched) || (r.ArrivedAt != nil && day(*r.ArrivedAt) > sched) || (r.StartedAt != nil && day(*r.StartedAt) > sched):
		r.Cause = "LATE_START"
		r.CauseDetail = "الشغل بدا بعد يوم موعده"
	case r.Status != "COMPLETED" && r.StartedAt == nil && r.ArrivedAt == nil:
		r.Cause = "NOT_STARTED"
		r.CauseDetail = "موعده فات وما انطلق ولا بدا"
	default:
		r.Cause = "UNKNOWN"
	}
	if r.Note != nil && r.CauseDetail == "" {
		r.CauseDetail = *r.Note
	}
	// مشاكل بالبيانات — حجوزات ممكن ما لازم تنحسب متأخرة
	r.DataIssues = []string{}
	if r.CompletedAt != nil && r.LastChangeAt != nil && r.LastChangeAt.After(*r.CompletedAt) {
		r.DataIssues = append(r.DataIssues, "الموعد انغيّر بعد الإنجاز")
	}
	if r.CompletedAt != nil && r.StartedAt != nil && r.CompletedAt.Before(*r.StartedAt) {
		r.DataIssues = append(r.DataIssues, "الإنجاز مسجّل قبل البدء")
	}
	if r.Status == "COMPLETED" && r.StartedAt == nil && r.ArrivedAt == nil {
		r.DataIssues = append(r.DataIssues, "انسجّل منجز بلا ولا مرحلة (وصول/بدء)")
	}
	if r.Reschedules > 0 && r.Postpones == 0 && r.DaysLate > 0 && r.CompletedAt != nil && r.LastChangeAt != nil && r.LastChangeAt.Before(r.ScheduledAt) {
		r.DataIssues = append(r.DataIssues, "انعاد جدولته — تأكد إن الموعد الأخير هو المعتمد")
	}
}

var lateCauseLabels = map[string]string{
	"POSTPONED": "🚫 الزبون أجّل", "NO_CREW": "👥 بلا كادر", "NOT_CONFIRMED": "📞 ما انثبّت ويّا الزبون",
	"MATERIALS": "📦 نقص مواد", "WORK_STOPPED": "⛔ توقف العمل", "MULTI_DAY": "🧩 شغل على أيام",
	"LATE_START": "🐢 انطلق متأخر", "NOT_STARTED": "⏳ ما انطلق", "UNKNOWN": "❔ ماكو دليل",
}

// GET /api/late-audit?from=YYYY-MM-DD&to=YYYY-MM-DD[&format=xlsx]
func (h *LateAuditHandler) List(w http.ResponseWriter, r *http.Request) {
	loc := baghdadLoc()
	q := r.URL.Query()
	to := q.Get("to")
	if to == "" {
		to = time.Now().In(loc).Format("2006-01-02")
	}
	from := q.Get("from")
	if from == "" {
		from = time.Now().In(loc).AddDate(0, 0, -30).Format("2006-01-02")
	}
	rows := []lateRow{}
	if err := h.db.Select(&rows, lateAuditSQL, from, to); err != nil {
		log.Printf("[late-audit] %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر الجلب")
		return
	}
	var total int
	_ = h.db.Get(&total, `SELECT count(*) FROM "Booking" b WHERE baghdad_date(b."scheduledAt") BETWEEN $1::date AND $2::date
		AND b.status::text <> 'CANCELLED' AND b."archivedAt" IS NULL AND b."bookingType" IS DISTINCT FROM 'INTERNAL'`, from, to)
	for i := range rows {
		classifyLate(&rows[i], loc)
	}
	if q.Get("format") != "xlsx" {
		WriteJSON(w, http.StatusOK, map[string]any{"from": from, "to": to, "total": total, "rows": rows, "causes": lateCauseLabels})
		return
	}
	f := excelize.NewFile()
	sh := "المتأخرات"
	_ = f.SetSheetName("Sheet1", sh)
	rtl := true
	_ = f.SetSheetView(sh, 0, &excelize.ViewOptions{RightToLeft: &rtl})
	head := []string{"الحجز", "الزبون", "الخدمة", "القسم", "الحالة", "الموعد", "الإنجاز", "أيام التأخير", "السبب", "التفصيل", "الليدر", "الكادر", "مشاكل بالبيانات"}
	for i, t := range head {
		c, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sh, c, t)
	}
	s := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}
	for ri, x := range rows {
		comp := ""
		if x.CompletedAt != nil {
			comp = x.CompletedAt.In(loc).Format("2006-01-02")
		}
		div := "هندسي"
		if x.Division == "DECOR" {
			div = "ديكور"
		}
		issues := ""
		for j, d := range x.DataIssues {
			if j > 0 {
				issues += "، "
			}
			issues += d
		}
		vals := []any{x.Code, s(x.Customer), s(x.Service), div, x.Status, x.ScheduledAt.In(loc).Format("2006-01-02"), comp, x.DaysLate,
			lateCauseLabels[x.Cause], x.CauseDetail, s(x.Leader), s(x.Team), issues}
		for ci, v := range vals {
			c, _ := excelize.CoordinatesToCellName(ci+1, ri+2)
			_ = f.SetCellValue(sh, c, v)
		}
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape("المتأخرات-"+from+"-"+to+".xlsx"))
	_ = f.Write(w)
}
