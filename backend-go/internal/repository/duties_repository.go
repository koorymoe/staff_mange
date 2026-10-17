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
	Status  string `json:"status"`  // DONE | PARTIAL | MISSED | INFO
	Link    string `json:"link,omitempty"`
}

type DutiesReport struct {
	EmployeeID string `json:"employeeId"`
	Position   string `json:"position"`
	Duties     []Duty `json:"duties"`
	DoneCount  int    `json:"doneCount"`
	Total      int    `json:"total"`
	Verdict    string `json:"verdict"` // GOOD | PARTIAL | BAD
	VerdictAr  string `json:"verdictAr"`
}

type dutyDef struct {
	key, label, hint, link string
	done                   string // $1 = الموظف، اليوم = baghdad_today()
	pending                string // فارغ = ماكو طابور؛ «once» = واجب مرة باليوم
}

const dutyToday = `baghdad_today()`

var dutyAttendance = dutyDef{"attendance", "سجّل حضورك", "أول شي بالدوام", "/attendance",
	`SELECT count(*) FROM "Attendance" WHERE "employeeId" = $1 AND baghdad_date("checkIn") = ` + dutyToday, "once"}

var dutiesByPosition = map[string][]dutyDef{
	"MONITOR": {
		dutyAttendance,
		{"inbox", "دقّق صندوق المراقب", "الفواتير بعد تدقيق المحاسب، والحجوزات، وأحكام ماتركس", "/monitor-inbox",
			`SELECT count(*) FROM "MonitorReview" WHERE "reviewedById" = $1 AND baghdad_date("reviewedAt") = ` + dutyToday,
			`SELECT count(*) FROM "MonitorReview" WHERE status = 'PENDING' AND ` + awaitingAccountant(`"MonitorReview"`)},
		{"workReports", "دقّق تقارير الحجوزات", "«✅ تم التدقيق» على كل تقرير", "/reports-desk",
			`SELECT count(*) FROM "WorkReport" WHERE "reviewedById" = $1 AND baghdad_date("reviewedAt") = ` + dutyToday,
			`SELECT count(*) FROM "WorkReport" WHERE "reviewedAt" IS NULL`},
		{"dailyReports", "دقّق تقارير المشاريع اليومية", "", "/reports-desk",
			`SELECT count(*) FROM "BookingProgressReport" WHERE "reviewedById" = $1 AND baghdad_date("reviewedAt") = ` + dutyToday,
			`SELECT count(*) FROM "BookingProgressReport" WHERE "reviewedAt" IS NULL`},
		{"justifications", "احكم على التبريرات الي وصلت", "", "/monitor-inbox",
			`SELECT count(*) FROM "MonitorJustification" WHERE "monitorId" = $1 AND baghdad_date("decidedAt") = ` + dutyToday,
			`SELECT count(*) FROM "MonitorJustification" WHERE "answeredAt" IS NOT NULL AND verdict IS NULL`},
	},
	"FINANCE": {
		dutyAttendance,
		{"audit", "دقّق فواتير الليدرية", "خلال ٢٤ ساعة من وصولها", "/accountant-desk",
			`SELECT count(*) FROM "LeaderInvoice" WHERE "auditedById" = $1 AND baghdad_date("auditedAt") = ` + dutyToday,
			`SELECT count(*) FROM "LeaderInvoice" WHERE "auditedAt" IS NULL AND status <> 'APPROVED' AND NOT "isFree"`},
		{"approve", "اعتمد الفواتير المدقّقة برقمها المحاسبي", "", "/accountant-desk",
			`SELECT count(*) FROM "LeaderInvoice" WHERE "approvedByEmployeeId" = $1 AND baghdad_date("approvedAt") = ` + dutyToday,
			`SELECT count(*) FROM "LeaderInvoice" WHERE "auditedAt" IS NOT NULL AND status <> 'APPROVED' AND NOT "isFree"`},
	},
	"HR_COORDINATOR": {
		dutyAttendance,
		{"contact", "تواصل ويا الزبائن وثبّت الحجوزات الجديدة", "", "/coordinator",
			`SELECT count(*) FROM "Booking" WHERE "confirmationContactedById" = $1 AND baghdad_date("confirmationContactedAt") = ` + dutyToday,
			`SELECT ` + "%needs_contact%" + ` FROM "Booking" b WHERE ` + "%base%"},
		{"crew", "كلّف كوادر للحجوزات المثبّتة", "", "/coordinator",
			`SELECT count(*) FROM "BookingAssignment" WHERE "assignedById" = $1 AND baghdad_date("createdAt") = ` + dutyToday,
			`SELECT ` + "%needs_crew%" + ` FROM "Booking" b WHERE ` + "%base%"},
		{"rateLeaders", "قيّم الليدرية", "", "/staff-score",
			`SELECT count(*) FROM "StaffRating" WHERE "raterId" = $1 AND baghdad_date("createdAt") = ` + dutyToday, ""},
	},
	"QUALITY_ENGINEER": {
		dutyAttendance,
		{"quality", "تواصل ويا زبائن متابعة الجودة", "تأكد الشغل تم والزبون راضي", "/quality-desk",
			`SELECT count(*) FROM "QualityFollowUp" WHERE "contactedByEmployeeId" = $1 AND baghdad_date("contactedAt") = ` + dutyToday,
			`SELECT count(*) FROM "QualityFollowUp" WHERE status = 'PENDING'`},
		{"complaints", "تابع الشكاوى المفتوحة", "", "/quality-desk",
			`SELECT count(DISTINCT "complaintId") FROM "ComplaintEvent" WHERE "byEmployeeId" = $1 AND baghdad_date("createdAt") = ` + dutyToday,
			`SELECT count(*) FROM "Complaint" WHERE status::text IN ('NEW','IN_PROGRESS')`},
	},
	"LEADER": {
		dutyAttendance,
		{"inventory", "جرد عدّتك", "", "/inventory",
			`SELECT count(*) FROM "InventoryCheck" WHERE "employeeId" = $1 AND baghdad_date("checkedAt") = ` + dutyToday, "once"},
		{"stages", "سجّل مراحل حجوزاتك (انطلاق، وصول، بدء، إنجاز)", "", "/my-tasks",
			`SELECT count(*) FROM "BookingStageProof" WHERE "employeeId" = $1 AND baghdad_date("pressedAt") = ` + dutyToday, ""},
		{"reports", "ارفع تقرير العمل لكل حجز خلص", "", "/work-reports",
			`SELECT count(*) FROM "WorkReport" WHERE "employeeId" = $1 AND baghdad_date("createdAt") = ` + dutyToday,
			`SELECT count(*) FROM "Booking" b WHERE b.status = 'COMPLETED' AND b."completedAt" > now() - interval '7 days' AND %worked%
			   AND NOT EXISTS (SELECT 1 FROM "WorkReport" w WHERE w."bookingId" = b.id)`},
		{"invoices", "سوّي فاتورة لكل حجز خلص", "", "/leader-invoices/new",
			`SELECT count(*) FROM "LeaderInvoice" WHERE "employeeId" = $1 AND baghdad_date("createdAt") = ` + dutyToday,
			`SELECT count(*) FROM "Booking" b WHERE b.status = 'COMPLETED' AND b."completedAt" > now() - interval '7 days' AND %worked%
			   AND NOT EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id)`},
		{"crewRating", "قيّم كادرك", "", "/my-tasks",
			`SELECT count(*) FROM "CrewRating" WHERE "leaderId" = $1 AND baghdad_date("createdAt") = ` + dutyToday, ""},
	},
	"TECHNICIAN": {
		dutyAttendance,
		{"inventory", "جرد عدّتك", "", "/inventory",
			`SELECT count(*) FROM "InventoryCheck" WHERE "employeeId" = $1 AND baghdad_date("checkedAt") = ` + dutyToday, "once"},
		{"stages", "سجّل مراحل حجوزاتك", "", "/my-tasks",
			`SELECT count(*) FROM "BookingStageProof" WHERE "employeeId" = $1 AND baghdad_date("pressedAt") = ` + dutyToday, ""},
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
	case e.Role == "ADMIN":
		return "ADMIN" // المالك/المدير ما ينقيّم
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
	rep := DutiesReport{EmployeeID: employeeID, Position: positionAr[pos], Duties: []Duty{}}
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
		if du.Pending != nil {
			switch {
			case *du.Pending == 0:
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
	switch {
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
