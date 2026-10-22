package repository

import (
	"strings"

	"github.com/jmoiron/sqlx"
)

// ═══ «🎯 واجباتك اليوم» — كل منصب إله شغل يومي (ع) 10-17 ═══
// «هاي الأهداف مو تارغت يوصله — هاي شغل كل موظف الموجود بالنظام. إذا ما
// سوّاهن أو سوّى جزء وعاف جزء، معناها مو خوش موظف». كل واجب ينحسب من
// الجداول نفسها: شكد سوّى اليوم (Done)، وشكد بعده ينتظر بطابور منصبه (Pending).
// بلا غرامة فلوس — تقييم ماتركس وتنبيه بس.

type Duty struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Hint    string `json:"hint"`
	Done    int    `json:"done"`
	Pending *int   `json:"pending"` // nil = ماكو طابور (يكفي يسوّيه مرة)
	// Late المتأخر أكثر من ٢٤ ساعة بالطابور — الحكم عليه (قرار (ع) 10-22).
	Late   *int   `json:"late,omitempty"`
	Status string `json:"status"` // DONE | PARTIAL | MISSED | INFO
	Link   string `json:"link,omitempty"`
}

type DutiesReport struct {
	EmployeeID string `json:"employeeId"`
	Position   string `json:"position"`
	Duties     []Duty `json:"duties"`
	DoneCount  int    `json:"doneCount"`
	Total      int    `json:"total"`
	Verdict    string `json:"verdict"` // GOOD | PARTIAL | BAD | OFF
	VerdictAr  string `json:"verdictAr"`
	// Off يوم عطلة (الجمعة) أو إجازة معتمدة — الواجبات تبين بس ما تنحسب.
	Off string `json:"off,omitempty"` // FRIDAY | LEAVE
}

type dutyDef struct {
	key, label, hint, link string
	done                   string // $1 = الموظف، اليوم = baghdad_today()
	pending                string // فارغ = ماكو طابور؛ «once» = واجب مرة باليوم
	// late الي بالطابور وفات عليه أكثر من ٢٤ ساعة — قرار (ع) 10-22. الطابور
	// مشترك بين أهل المنصب، فالموظف يتحاسب على المتأخر بس: شي وصل الساعة ١١
	// بالليل، أو شغل ضل لزميله، ما يخلّيه «ناقص». فارغ = الحكم بالطابور كله.
	late string
}

const dutyToday = `baghdad_today()`

// المهلة: الأعمدة بلا منطقة UTC، والي بمنطقة تتقارن ويا now() مباشرة.
const (
	lateNaive = ` < (now() AT TIME ZONE 'UTC') - interval '24 hours'`
	lateTZ    = ` < now() - interval '24 hours'`
)

// فاتورة مكررة لفاتورة معتمدة بنفس اليوم ما تنحسب بطابور المحاسب.
const notTwin = ` AND NOT ` + TwinApprovedSQL

var dutyAttendance = dutyDef{key: "attendance", label: "سجّل حضورك", hint: "أول شي بالدوام", link: "/attendance",
	done:    `SELECT count(*) FROM "Attendance" WHERE "employeeId" = $1 AND baghdad_date("checkIn") = ` + dutyToday,
	pending: "once"}

var dutiesByPosition = map[string][]dutyDef{
	"MONITOR": {
		dutyAttendance,
		{key: "inbox", label: "دقّق صندوق المراقب", hint: "الفواتير بعد تدقيق المحاسب، والحجوزات، وأحكام ماتركس", link: "/monitor-inbox",
			done:    `SELECT count(*) FROM "MonitorReview" WHERE "reviewedById" = $1 AND baghdad_date("reviewedAt") = ` + dutyToday,
			pending: `SELECT count(*) FROM "MonitorReview" WHERE status = 'PENDING' AND ` + awaitingAccountant(`"MonitorReview"`),
			late:    `SELECT count(*) FROM "MonitorReview" WHERE status = 'PENDING' AND ` + awaitingAccountant(`"MonitorReview"`) + ` AND "createdAt"` + lateTZ},
		{key: "workReports", label: "دقّق تقارير الحجوزات", hint: "«✅ تم التدقيق» على كل تقرير", link: "/reports-desk",
			done:    `SELECT count(*) FROM "WorkReport" WHERE "reviewedById" = $1 AND baghdad_date("reviewedAt") = ` + dutyToday,
			pending: `SELECT count(*) FROM "WorkReport" WHERE "reviewedAt" IS NULL`,
			late:    `SELECT count(*) FROM "WorkReport" WHERE "reviewedAt" IS NULL AND "createdAt"` + lateNaive},
		{key: "dailyReports", label: "دقّق تقارير المشاريع اليومية", link: "/reports-desk",
			done:    `SELECT count(*) FROM "BookingProgressReport" WHERE "reviewedById" = $1 AND baghdad_date("reviewedAt") = ` + dutyToday,
			pending: `SELECT count(*) FROM "BookingProgressReport" WHERE "reviewedAt" IS NULL`,
			late:    `SELECT count(*) FROM "BookingProgressReport" WHERE "reviewedAt" IS NULL AND "createdAt"` + lateTZ},
		{key: "justifications", label: "احكم على التبريرات الي وصلت", link: "/monitor-inbox",
			done:    `SELECT count(*) FROM "MonitorJustification" WHERE "monitorId" = $1 AND baghdad_date("decidedAt") = ` + dutyToday,
			pending: `SELECT count(*) FROM "MonitorJustification" WHERE "answeredAt" IS NOT NULL AND verdict IS NULL`,
			late:    `SELECT count(*) FROM "MonitorJustification" WHERE "answeredAt" IS NOT NULL AND verdict IS NULL AND "answeredAt"` + lateNaive},
	},
	"FINANCE": {
		dutyAttendance,
		{key: "audit", label: "دقّق فواتير الليدرية", hint: "خلال ٢٤ ساعة من وصولها", link: "/accountant-desk",
			done:    `SELECT count(*) FROM "LeaderInvoice" WHERE "auditedById" = $1 AND baghdad_date("auditedAt") = ` + dutyToday,
			pending: `SELECT count(*) FROM "LeaderInvoice" WHERE "auditedAt" IS NULL AND status <> 'APPROVED' AND NOT "isFree"` + notTwin,
			late:    `SELECT count(*) FROM "LeaderInvoice" WHERE "auditedAt" IS NULL AND status <> 'APPROVED' AND NOT "isFree"` + notTwin + ` AND "createdAt"` + lateNaive},
		{key: "approve", label: "اعتمد الفواتير المدقّقة برقمها المحاسبي", link: "/accountant-desk",
			done:    `SELECT count(*) FROM "LeaderInvoice" WHERE "approvedByEmployeeId" = $1 AND baghdad_date("approvedAt") = ` + dutyToday,
			pending: `SELECT count(*) FROM "LeaderInvoice" WHERE "auditedAt" IS NOT NULL AND status <> 'APPROVED' AND NOT "isFree"` + notTwin,
			late:    `SELECT count(*) FROM "LeaderInvoice" WHERE "auditedAt" IS NOT NULL AND status <> 'APPROVED' AND NOT "isFree"` + notTwin + ` AND "auditedAt"` + lateTZ},
	},
	"HR_COORDINATOR": {
		dutyAttendance,
		{key: "contact", label: "تواصل ويا الزبائن وثبّت الحجوزات الجديدة", link: "/coordinator",
			done:    `SELECT count(*) FROM "Booking" WHERE "confirmationContactedById" = $1 AND baghdad_date("confirmationContactedAt") = ` + dutyToday,
			pending: `SELECT ` + "%needs_contact%" + ` FROM "Booking" b WHERE ` + "%base%",
			late:    `SELECT ` + "%needs_contact_late%" + ` FROM "Booking" b WHERE ` + "%base%"},
		{key: "crew", label: "كلّف كوادر للحجوزات المثبّتة", link: "/coordinator",
			done:    `SELECT count(*) FROM "BookingAssignment" WHERE "assignedById" = $1 AND baghdad_date("createdAt") = ` + dutyToday,
			pending: `SELECT ` + "%needs_crew%" + ` FROM "Booking" b WHERE ` + "%base%",
			late:    `SELECT ` + "%needs_crew_late%" + ` FROM "Booking" b WHERE ` + "%base%"},
		{key: "rateLeaders", label: "قيّم الليدرية", link: "/staff-score",
			done: `SELECT count(*) FROM "StaffRating" WHERE "raterId" = $1 AND baghdad_date("createdAt") = ` + dutyToday},
	},
	"QUALITY_ENGINEER": {
		dutyAttendance,
		{key: "quality", label: "تواصل ويا زبائن متابعة الجودة", hint: "تأكد الشغل تم والزبون راضي", link: "/quality-desk",
			done:    `SELECT count(*) FROM "QualityFollowUp" WHERE "contactedByEmployeeId" = $1 AND baghdad_date("contactedAt") = ` + dutyToday,
			pending: `SELECT count(*) FROM "QualityFollowUp" WHERE status = 'PENDING'`,
			late:    `SELECT count(*) FROM "QualityFollowUp" WHERE status = 'PENDING' AND "createdAt"` + lateNaive},
		{key: "complaints", label: "تابع الشكاوى المفتوحة", link: "/quality-desk",
			done:    `SELECT count(DISTINCT "complaintId") FROM "ComplaintEvent" WHERE "byEmployeeId" = $1 AND baghdad_date("createdAt") = ` + dutyToday,
			pending: `SELECT count(*) FROM "Complaint" WHERE status::text IN ('NEW','IN_PROGRESS')`,
			late:    `SELECT count(*) FROM "Complaint" WHERE status::text IN ('NEW','IN_PROGRESS') AND "createdAt"` + lateNaive},
	},
	"LEADER": {
		dutyAttendance,
		{key: "inventory", label: "جرد عدّتك", link: "/inventory",
			done:    `SELECT count(*) FROM "InventoryCheck" WHERE "employeeId" = $1 AND baghdad_date("checkedAt") = ` + dutyToday,
			pending: "once"},
		{key: "stages", label: "سجّل مراحل حجوزاتك (انطلاق، وصول، بدء، إنجاز)", link: "/my-tasks",
			done: `SELECT count(*) FROM "BookingStageProof" WHERE "employeeId" = $1 AND baghdad_date("pressedAt") = ` + dutyToday},
		{key: "reports", label: "ارفع تقرير العمل لكل حجز خلص", link: "/work-reports",
			done: `SELECT count(*) FROM "WorkReport" WHERE "employeeId" = $1 AND baghdad_date("createdAt") = ` + dutyToday,
			pending: `SELECT count(*) FROM "Booking" b WHERE b.status = 'COMPLETED' AND b."completedAt" > now() - interval '7 days' AND %worked%
			   AND NOT EXISTS (SELECT 1 FROM "WorkReport" w WHERE w."bookingId" = b.id)`,
			late: `SELECT count(*) FROM "Booking" b WHERE b.status = 'COMPLETED' AND b."completedAt" > now() - interval '7 days' AND %worked%
			   AND b."completedAt"` + lateNaive + ` AND NOT EXISTS (SELECT 1 FROM "WorkReport" w WHERE w."bookingId" = b.id)`},
		{key: "invoices", label: "سوّي فاتورة لكل حجز خلص", link: "/leader-invoices/new",
			done: `SELECT count(*) FROM "LeaderInvoice" WHERE "employeeId" = $1 AND baghdad_date("createdAt") = ` + dutyToday,
			pending: `SELECT count(*) FROM "Booking" b WHERE b.status = 'COMPLETED' AND b."completedAt" > now() - interval '7 days' AND %worked%
			   AND NOT EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id)`,
			late: `SELECT count(*) FROM "Booking" b WHERE b.status = 'COMPLETED' AND b."completedAt" > now() - interval '7 days' AND %worked%
			   AND b."completedAt"` + lateNaive + ` AND NOT EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id)`},
		{key: "crewRating", label: "قيّم كادرك", link: "/my-tasks",
			done: `SELECT count(*) FROM "CrewRating" WHERE "leaderId" = $1 AND baghdad_date("createdAt") = ` + dutyToday},
	},
	"TECHNICIAN": {
		dutyAttendance,
		{key: "inventory", label: "جرد عدّتك", link: "/inventory",
			done:    `SELECT count(*) FROM "InventoryCheck" WHERE "employeeId" = $1 AND baghdad_date("checkedAt") = ` + dutyToday,
			pending: "once"},
		{key: "stages", label: "سجّل مراحل حجوزاتك", link: "/my-tasks",
			done: `SELECT count(*) FROM "BookingStageProof" WHERE "employeeId" = $1 AND baghdad_date("pressedAt") = ` + dutyToday},
	},
}

var positionAr = map[string]string{
	"MONITOR": "مراقب", "FINANCE": "محاسب", "HR_COORDINATOR": "إداري الحجوزات", "QUALITY_ENGINEER": "مهندس جودة",
	"LEADER": "تيم ليدر", "TECHNICIAN": "فني", "OTHER": "موظف",
}

type DutiesRepository struct{ db *sqlx.DB }

func NewDutiesRepository(db *sqlx.DB) *DutiesRepository { return &DutiesRepository{db: db} }

// position منصب الموظف للواجبات — صاحب صلاحية monitoring مراقب (نفس قاعدة النظام).
func (r *DutiesRepository) position(id string) string {
	var e struct {
		Role     string `db:"role"`
		IsLeader bool   `db:"isLeader"`
		Monitor  bool   `db:"monitor"`
		Coord    bool   `db:"coord"`
	}
	if err := r.db.Get(&e, `SELECT role::text AS role, "isLeader",
		EXISTS (SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId" WHERE ep."employeeId" = $1 AND p.name = 'monitoring') AS monitor,
		EXISTS (SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId" WHERE ep."employeeId" = $1 AND p.name = 'coordinator') AS coord
		FROM "Employee" WHERE id = $1`, id); err != nil {
		return "OTHER"
	}
	switch {
	case e.Role == "ADMIN" || e.Role == "OWNER":
		// المالك والمدير ما ينقيّمون — الدور بالقاعدة OWNER (ينقلب ADMIN بالجلسة
		// بس)، فچان المالك يطلع «موظف» وينحاسب على الحضور.
		return "ADMIN"
	case e.Role == "MONITOR" || (e.Monitor && e.Role != "QUALITY_ENGINEER"):
		return "MONITOR"
	case (e.Role == "TECHNICIAN" || e.Role == "TECHNICAL") && e.IsLeader:
		return "LEADER"
	case e.Role == "TECHNICIAN" || e.Role == "TECHNICAL":
		return "TECHNICIAN"
	case e.Role == "HR_COORDINATOR" || e.Coord:
		return "HR_COORDINATOR"
	case e.Role == "FINANCE" || e.Role == "QUALITY_ENGINEER":
		return e.Role
	}
	return "OTHER"
}

func (r *DutiesRepository) fill(sql string) string {
	out := sql
	for k, v := range map[string]string{
		"%needs_contact%": DrillCountSQL("needs_contact"), "%needs_crew%": DrillCountSQL("needs_crew"),
		"%needs_contact_late%": DrillLateSQL("needs_contact"), "%needs_crew_late%": DrillLateSQL("needs_crew"),
		"%base%": drillBase(), "%worked%": BookingWorkedBySQL("b", "$1"),
	} {
		out = strings.ReplaceAll(out, k, v)
	}
	return out
}

// Today واجبات الموظف اليوم وشكد سوّى منها.
func (r *DutiesRepository) Today(employeeID string) DutiesReport {
	pos := r.position(employeeID)
	if pos == "ADMIN" {
		return DutiesReport{EmployeeID: employeeID, Position: "ADMIN", Duties: []Duty{}}
	}
	defs := dutiesByPosition[pos]
	if len(defs) == 0 {
		defs = []dutyDef{dutyAttendance}
	}
	rep := DutiesReport{EmployeeID: employeeID, Position: positionAr[pos], Duties: []Duty{}, Off: r.dayOff(employeeID)}
	for _, d := range defs {
		du := Duty{Key: d.key, Label: d.label, Hint: d.hint, Link: d.link}
		_ = r.db.Get(&du.Done, r.fill(d.done), employeeID)
		switch d.pending {
		case "":
			du.Status = "INFO"
			if du.Done > 0 {
				du.Status = "DONE"
			}
		case "once":
			p := 0
			if du.Done == 0 {
				p = 1
			}
			du.Pending = &p
		default:
			var p int
			q := r.fill(d.pending)
			var err error
			if strings.Contains(q, "$1") {
				err = r.db.Get(&p, q, employeeID)
			} else {
				err = r.db.Get(&p, q)
			}
			if err == nil {
				du.Pending = &p
			}
		}
		judge := du.Pending
		if d.late != "" && du.Pending != nil {
			var l int
			q := r.fill(d.late)
			args := []any{}
			if strings.Contains(q, "$1") {
				args = append(args, employeeID)
			}
			if err := r.db.Get(&l, q, args...); err == nil {
				du.Late, judge = &l, &l
			}
		}
		if judge != nil {
			switch {
			case *judge == 0:
				du.Status = "DONE"
			case du.Done > 0:
				du.Status = "PARTIAL"
			default:
				du.Status = "MISSED"
			}
			rep.Total++
			if du.Status == "DONE" {
				rep.DoneCount++
			}
		}
		rep.Duties = append(rep.Duties, du)
	}
	// (ع) 10-19: الشغل الموجّه إله بإسمه (مشاريع، حجوزات مرحّلة، مهام، شكاوى).
	for _, du := range r.personal(employeeID) {
		if du.Pending != nil {
			rep.Total++
			if du.Status == "DONE" {
				rep.DoneCount++
			}
		}
		rep.Duties = append(rep.Duties, du)
	}
	switch {
	case rep.Off != "":
		rep.Total, rep.DoneCount = 0, 0
		rep.Verdict, rep.VerdictAr = "OFF", map[string]string{"FRIDAY": "جمعة — عطلة", "LEAVE": "إجازة معتمدة"}[rep.Off]
	case rep.Total == 0 || rep.DoneCount == rep.Total:
		rep.Verdict, rep.VerdictAr = "GOOD", "خوش — خلّص واجباته"
	case rep.DoneCount == 0 && !anyPartial(rep.Duties):
		rep.Verdict, rep.VerdictAr = "BAD", "ما سوّى واجباته"
	default:
		rep.Verdict, rep.VerdictAr = "PARTIAL", "سوّى جزء وعاف جزء"
	}
	return rep
}

func anyPartial(ds []Duty) bool {
	for _, d := range ds {
		if d.Status == "PARTIAL" {
			return true
		}
	}
	return false
}

// dayOff عطلة الموظف اليوم: الجمعة للكل (قرار (ع) 10-22)، أو إجازة معتمدة تغطي اليوم.
func (r *DutiesRepository) dayOff(employeeID string) string {
	var off string
	_ = r.db.Get(&off, `SELECT CASE
		WHEN extract(isodow FROM baghdad_today()) = 5 THEN 'FRIDAY'
		WHEN EXISTS (SELECT 1 FROM "LeaveRequest" l WHERE l."employeeId" = $1 AND l.status = 'APPROVED'
		  AND baghdad_today() BETWEEN l."startDate" AND l."endDate") THEN 'LEAVE'
		ELSE '' END`, employeeID)
	return off
}
