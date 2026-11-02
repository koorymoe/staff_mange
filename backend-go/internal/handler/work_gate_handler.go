package handler

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
	"staffmange-api/internal/timeutil"
)

// ═══ 🧠 بوابة الالتزام (دازاي 11-02 — تعديلات ٩.١) ═══
// ١. استمارة معلومات الموظف: مرة وحدة. الموظف يكتب والإدارة تطابق — إذا الراتب
//    أو نهاية العقد يختلف عن ملفه، تنبيه للإدارة. (الراتب ما يطلع لأي ذكاء خارجي.)
// ٢. تقرير الإنجاز اليومي: كل يوم داوم بيه (عدا الإجازة) لازم إله تقرير. الي فاته،
//    أول ما يفتح النظام تنقفل الشاشات لحد ما يكتبه — وينأشّر «متأخر».
// ماتركس يقرا التقارير ويطابقها بالشغل المسجّل (٩.٤).

type WorkGateHandler struct {
	db    *sqlx.DB
	notif *repository.NotificationRepository
	duty  *service.RatingDuty
}

func NewWorkGateHandler(db *sqlx.DB, n *repository.NotificationRepository) *WorkGateHandler {
	return &WorkGateHandler{db: db, notif: n, duty: service.NewRatingDuty(db)}
}

type workGate struct {
	Survey      bool    `json:"survey"`      // لازم يعبّي الاستمارة
	MissingDay  *string `json:"missingDay"`  // يوم داوم بيه وما كتب تقريره (YYYY-MM-DD)
	TodayDone   bool    `json:"todayDone"`   // كتب تقرير اليوم
	JobTitle    string  `json:"jobTitle"`    // للتعبئة المسبقة
	ReportSince string  `json:"reportSince"` // من وين يبدي الإلزام
	// ٩.٢: تقييمات فاتته (أقدم يوم خلال ٣ أيام) — تنقفل الشاشات لحد ما يعوّض ويبرّر
	RatingDuty *ratingDutyGate `json:"ratingDuty"`
}

type ratingDutyGate struct {
	Day    string             `json:"day"`
	Items  []service.DutyItem `json:"items"`
	Labels map[string]string  `json:"labels"`
}

func gateExempt(r *http.Request) bool {
	return middleware.IsStealth(r) || model.IsAdminRole(middleware.RoleFromContext(r))
}

func companyToday() time.Time {
	n := time.Now().In(timeutil.CompanyLocation())
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, timeutil.CompanyLocation())
}

// missingReportDay آخر يوم (قبل اليوم، خلال ٧ أيام، ومن يوم تشغيل الميزة) داوم
// بيه الموظف، وما عنده إجازة معتمدة، وما كتب تقريره.
func (h *WorkGateHandler) missingReportDay(emp string) (*string, string) {
	var since string
	_ = h.db.Get(&since, `SELECT value FROM "OrgGovernorState" WHERE key = 'daily_report_since'`)
	var day *string
	_ = h.db.Get(&day, `
		SELECT to_char(max(a.date::date), 'YYYY-MM-DD') FROM "Attendance" a
		WHERE a."employeeId" = $1
		  AND a.date::date < $2::date AND a.date::date >= $2::date - 7
		  AND ($3 = '' OR a.date::date >= $3::date)
		  AND NOT EXISTS (SELECT 1 FROM "DailyWorkReport" d WHERE d."employeeId" = $1 AND d.day = a.date::date)
		  AND NOT EXISTS (SELECT 1 FROM "LeaveRequest" l WHERE l."employeeId" = $1 AND l.status = 'APPROVED'
		                  AND a.date::date BETWEEN l."startDate" AND l."endDate")`,
		emp, companyToday().Format("2006-01-02"), since)
	return day, since
}

// GET /api/me/work-gate
func (h *WorkGateHandler) Gate(w http.ResponseWriter, r *http.Request) {
	g := workGate{}
	if gateExempt(r) {
		WriteJSON(w, http.StatusOK, g)
		return
	}
	emp := middleware.EmployeeIDFromContext(r)
	_ = h.db.Get(&g.Survey, `SELECT NOT EXISTS (SELECT 1 FROM "EmployeeProfileSurvey" WHERE "employeeId" = $1)`, emp)
	_ = h.db.Get(&g.JobTitle, `SELECT COALESCE("jobTitle", position, '') FROM "Employee" WHERE id = $1`, emp)
	g.MissingDay, g.ReportSince = h.missingReportDay(emp)
	_ = h.db.Get(&g.TodayDone, `SELECT EXISTS (SELECT 1 FROM "DailyWorkReport" WHERE "employeeId" = $1 AND day = $2::date)`, emp, companyToday().Format("2006-01-02"))
	if day, items, err := h.duty.Pending(emp); err != nil {
		log.Printf("[work-gate] rating duty %s: %v", emp, err) // ما ينبلع بصمت
	} else if len(items) > 0 {
		g.RatingDuty = &ratingDutyGate{Day: day, Items: items, Labels: service.DutyStageLabels}
	}
	WriteJSON(w, http.StatusOK, g)
}

// POST /api/me/profile-survey
func (h *WorkGateHandler) SubmitSurvey(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Major         string   `json:"major"`
		Degree        string   `json:"degree"`
		SelfSalaryIQD *float64 `json:"selfSalaryIQD"`
		MonthlyHours  *int     `json:"monthlyHours"`
		CurrentJob    string   `json:"currentJob"`
		CompanyRating int      `json:"companyRating"`
		CompanyNote   string   `json:"companyNote"`
		ContractEnd   string   `json:"contractEnd"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	t := strings.TrimSpace
	switch {
	case len([]rune(t(in.Major))) < 2 || len([]rune(t(in.Degree))) < 2:
		WriteError(w, http.StatusBadRequest, "اكتب تخصصك الدراسي وشهادتك")
		return
	case len([]rune(t(in.CurrentJob))) < 2:
		WriteError(w, http.StatusBadRequest, "اكتب وظيفتك الحالية")
		return
	case in.CompanyRating < 1 || in.CompanyRating > 5:
		WriteError(w, http.StatusBadRequest, "قيّم العمل ويا الأماني من ١ لـ ٥")
		return
	case in.MonthlyHours != nil && (*in.MonthlyHours < 1 || *in.MonthlyHours > 400):
		WriteError(w, http.StatusBadRequest, "الساعات بالشهر بين ١ و٤٠٠")
		return
	case in.SelfSalaryIQD != nil && *in.SelfSalaryIQD < 0:
		WriteError(w, http.StatusBadRequest, "الراتب ما يكون سالب")
		return
	}
	var contract *string
	if c := t(in.ContractEnd); c != "" {
		if _, err := time.Parse("2006-01-02", c); err != nil {
			WriteError(w, http.StatusBadRequest, "تاريخ نهاية العقد لازم يوم/شهر/سنة")
			return
		}
		contract = &c
	}
	emp := middleware.EmployeeIDFromContext(r)
	// المطابقة ويا ملف الموظف (الي تعبّيه الإدارة)
	var file struct {
		Name     string     `db:"name"`
		Salary   *float64   `db:"salary"`
		Contract *time.Time `db:"contractEndDate"`
	}
	_ = h.db.Get(&file, `SELECT name, salary, "contractEndDate" FROM "Employee" WHERE id = $1`, emp)
	diffs := []string{}
	if in.SelfSalaryIQD != nil && file.Salary != nil && math.Abs(*in.SelfSalaryIQD-*file.Salary) >= 1 {
		diffs = append(diffs, fmt.Sprintf("الراتب: كتب %.0f والملف %.0f", *in.SelfSalaryIQD, *file.Salary))
	}
	if contract != nil && file.Contract != nil && *contract != file.Contract.Format("2006-01-02") {
		diffs = append(diffs, fmt.Sprintf("نهاية العقد: كتب %s والملف %s", *contract, file.Contract.Format("2006-01-02")))
	}
	var mismatch *string
	if len(diffs) > 0 {
		m := strings.Join(diffs, " · ")
		mismatch = &m
	}
	res, err := h.db.Exec(`INSERT INTO "EmployeeProfileSurvey" ("employeeId", major, degree, "selfSalaryIQD", "monthlyHours", "currentJob",
			"companyRating", "companyNote", "contractEnd", mismatch)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::date, $10) ON CONFLICT ("employeeId") DO NOTHING`,
		emp, t(in.Major), t(in.Degree), in.SelfSalaryIQD, in.MonthlyHours, t(in.CurrentJob), in.CompanyRating, t(in.CompanyNote), contract, mismatch)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الحفظ")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		WriteError(w, http.StatusConflict, "عبّيت الاستمارة من قبل — التعديل يصير من إدارة الكوادر")
		return
	}
	if mismatch != nil && h.notif != nil {
		_ = h.notif.CreateForRolesOrPermission([]string{"HR_COORDINATOR"}, "staff_management", "profile_survey_mismatch",
			"🧾 استمارة "+file.Name+" ما تطابق ملفه: "+*mismatch+" — راجعها بشاشة «تقارير الموظفين»")
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/me/daily-report {day, body}
func (h *WorkGateHandler) SubmitDaily(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Day  string `json:"day"`
		Body string `json:"body"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	body := strings.TrimSpace(in.Body)
	if len([]rune(body)) < 20 {
		WriteError(w, http.StatusBadRequest, "اكتب شنو سويت بالتفصيل (٢٠ حرف على الأقل)")
		return
	}
	today := companyToday()
	day, err := time.ParseInLocation("2006-01-02", in.Day, timeutil.CompanyLocation())
	if err != nil || day.After(today) || day.Before(today.AddDate(0, 0, -7)) {
		WriteError(w, http.StatusBadRequest, "اليوم لازم خلال آخر أسبوع")
		return
	}
	emp := middleware.EmployeeIDFromContext(r)
	late := day.Before(today)
	if _, err := h.db.Exec(`INSERT INTO "DailyWorkReport" ("employeeId", day, body, late) VALUES ($1, $2::date, $3, $4)
		ON CONFLICT ("employeeId", day) DO UPDATE SET body = EXCLUDED.body`, emp, in.Day, body, late); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true, "late": late})
}

type dailyReportRow struct {
	ID           string    `db:"id" json:"id"`
	EmployeeID   string    `db:"employeeId" json:"employeeId"`
	EmployeeName string    `db:"employeeName" json:"employeeName"`
	Day          string    `db:"day" json:"day"`
	Body         string    `db:"body" json:"body"`
	Late         bool      `db:"late" json:"late"`
	CreatedAt    time.Time `db:"createdAt" json:"createdAt"`
}

type missingRow struct {
	EmployeeID   string `db:"employeeId" json:"employeeId"`
	EmployeeName string `db:"employeeName" json:"employeeName"`
}

// GET /api/daily-reports?day=YYYY-MM-DD — التقارير + منو داوم وما كتب
func (h *WorkGateHandler) ListDaily(w http.ResponseWriter, r *http.Request) {
	day := r.URL.Query().Get("day")
	if _, err := time.Parse("2006-01-02", day); err != nil {
		day = companyToday().AddDate(0, 0, -1).Format("2006-01-02")
	}
	rows := []dailyReportRow{}
	_ = h.db.Select(&rows, `SELECT d.id, d."employeeId", e.name AS "employeeName", to_char(d.day, 'YYYY-MM-DD') AS day, d.body, d.late, d."createdAt"
		FROM "DailyWorkReport" d JOIN "Employee" e ON e.id = d."employeeId" WHERE d.day = $1::date ORDER BY e.name`, day)
	missing := []missingRow{}
	_ = h.db.Select(&missing, `SELECT DISTINCT e.id AS "employeeId", e.name AS "employeeName" FROM "Attendance" a JOIN "Employee" e ON e.id = a."employeeId"
		WHERE a.date::date = $1::date AND e.role::text NOT IN ('ADMIN', 'OWNER')
		  AND NOT EXISTS (SELECT 1 FROM "DailyWorkReport" d WHERE d."employeeId" = e.id AND d.day = $1::date)
		  AND NOT EXISTS (SELECT 1 FROM "LeaveRequest" l WHERE l."employeeId" = e.id AND l.status = 'APPROVED' AND $1::date BETWEEN l."startDate" AND l."endDate")
		ORDER BY 2`, day)
	WriteJSON(w, http.StatusOK, map[string]any{"day": day, "reports": rows, "missing": missing})
}

type surveyRow struct {
	EmployeeID    string     `db:"employeeId" json:"employeeId"`
	EmployeeName  string     `db:"employeeName" json:"employeeName"`
	Major         string     `db:"major" json:"major"`
	Degree        string     `db:"degree" json:"degree"`
	SelfSalaryIQD *float64   `db:"selfSalaryIQD" json:"selfSalaryIQD"`
	FileSalary    *float64   `db:"fileSalary" json:"fileSalary"`
	MonthlyHours  *int       `db:"monthlyHours" json:"monthlyHours"`
	CurrentJob    string     `db:"currentJob" json:"currentJob"`
	CompanyRating int        `db:"companyRating" json:"companyRating"`
	CompanyNote   string     `db:"companyNote" json:"companyNote"`
	ContractEnd   *string    `db:"contractEnd" json:"contractEnd"`
	FileContract  *string    `db:"fileContract" json:"fileContract"`
	Mismatch      *string    `db:"mismatch" json:"mismatch"`
	SubmittedAt   time.Time  `db:"submittedAt" json:"submittedAt"`
	ReviewedAt    *time.Time `db:"reviewedAt" json:"reviewedAt"`
}

// GET /api/profile-surveys — الاستمارات + منو بعده ما عبّى (إدارة الكوادر)
func (h *WorkGateHandler) ListSurveys(w http.ResponseWriter, r *http.Request) {
	rows := []surveyRow{}
	_ = h.db.Select(&rows, `SELECT s."employeeId", e.name AS "employeeName", s.major, s.degree, s."selfSalaryIQD", e.salary AS "fileSalary",
			s."monthlyHours", s."currentJob", s."companyRating", s."companyNote", to_char(s."contractEnd", 'YYYY-MM-DD') AS "contractEnd",
			to_char(e."contractEndDate", 'YYYY-MM-DD') AS "fileContract", s.mismatch, s."submittedAt", s."reviewedAt"
		FROM "EmployeeProfileSurvey" s JOIN "Employee" e ON e.id = s."employeeId"
		ORDER BY (s.mismatch IS NOT NULL AND s."reviewedAt" IS NULL) DESC, s."submittedAt" DESC`)
	pending := []missingRow{}
	_ = h.db.Select(&pending, `SELECT e.id AS "employeeId", e.name AS "employeeName" FROM "Employee" e
		WHERE e.status = 'ACTIVE' AND e.role::text NOT IN ('ADMIN', 'OWNER')
		  AND NOT EXISTS (SELECT 1 FROM "EmployeeProfileSurvey" s WHERE s."employeeId" = e.id) ORDER BY e.name`)
	WriteJSON(w, http.StatusOK, map[string]any{"surveys": rows, "pending": pending})
}

// POST /api/profile-surveys/{id}/review — الإدارة طابقت (والتصحيح بملف الموظف نفسه)
func (h *WorkGateHandler) ReviewSurvey(w http.ResponseWriter, r *http.Request) {
	res, err := h.db.Exec(`UPDATE "EmployeeProfileSurvey" SET "reviewedById" = $2, "reviewedAt" = now() WHERE "employeeId" = $1`,
		r.PathValue("id"), middleware.EmployeeIDFromContext(r))
	if n, _ := res.RowsAffected(); err != nil || n == 0 {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/me/rating-duty?day= — الي عليه يقيّمه (اليوم افتراضياً)
func (h *WorkGateHandler) MyDuty(w http.ResponseWriter, r *http.Request) {
	day := r.URL.Query().Get("day")
	if _, err := time.Parse("2006-01-02", day); err != nil {
		day = companyToday().Format("2006-01-02")
	}
	items, err := h.duty.Owed(middleware.EmployeeIDFromContext(r), day)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"day": day, "items": items, "labels": service.DutyStageLabels})
}

// POST /api/me/rating-duty {day, reason, ratings[]}
func (h *WorkGateHandler) SubmitDuty(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Day     string               `json:"day"`
		Reason  string               `json:"reason"`
		Ratings []service.DutyRating `json:"ratings"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	if _, err := time.Parse("2006-01-02", in.Day); err != nil {
		WriteError(w, http.StatusBadRequest, "اليوم غلط")
		return
	}
	if err := h.duty.Compensate(middleware.EmployeeIDFromContext(r), in.Day, in.Reason, in.Ratings); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

type negligenceRow struct {
	ID            string     `db:"id" json:"id"`
	EmployeeName  string     `db:"employeeName" json:"employeeName"`
	Day           string     `db:"day" json:"day"`
	Kind          string     `db:"kind" json:"kind"`
	Detail        string     `db:"detail" json:"detail"`
	Reason        string     `db:"reason" json:"reason"`
	CompensatedAt *time.Time `db:"compensatedAt" json:"compensatedAt"`
}

// GET /api/negligence — سجل الإهمال (آخر ٦٠ يوم)
func (h *WorkGateHandler) Negligence(w http.ResponseWriter, r *http.Request) {
	rows := []negligenceRow{}
	_ = h.db.Select(&rows, `SELECT n.id, e.name AS "employeeName", to_char(n.day, 'YYYY-MM-DD') AS day, n.kind, n.detail, n.reason, n."compensatedAt"
		FROM "EmployeeNegligence" n JOIN "Employee" e ON e.id = n."employeeId"
		WHERE n.day >= current_date - 60 ORDER BY n.day DESC, e.name LIMIT 300`)
	WriteJSON(w, http.StatusOK, rows)
}

// Sweep للحلقة الساعية.
func (h *WorkGateHandler) Sweep() (int, error) { return h.duty.Sweep() }
