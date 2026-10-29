package service

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// ═══ 🔐 قراءة بيانات المجال للموظف (دازاي 10-29) ═══
// «المفروض كل شي بالنظام الذكاء يكدر يوصله» — بس للموظف بحدود مجاله. query_data
// وdescribe_data ينفتحن إله محصورات بجداول صلاحياته؛ أي جدول برّا المجال أو عمود
// حساس = رفض بالكود. فوكها يبقى دور القراءة وREAD ONLY والمهلة وحد الصفوف.

// matrixDataScope صلاحية ← جداول (البادئة بـ* تطابق كل جدول يبدي بيها).
var matrixDataScope = map[string][]string{
	"vehicle_management":    {"Vehicle*", "FuelCard"},
	"procurement":           {"Procurement*", "ProductRequest", "ProductProcurement", "StockIntake", "Material", "Supplier*", "TechSupplier*"},
	"inventory":             {"StockIntake", "Material", "ToolRequest", "PersonalTool*", "TeamInventory*", "InventoryCheck", "OnDemandTool", "VehicleTool*"},
	"inventory_follow":      {"TeamInventory*", "InventoryCheck", "PersonalTool*"},
	"tool_requests_approve": {"ToolRequest", "PersonalTool*", "OnDemandTool"},
	"suppliers_management":  {"Supplier*", "TechSupplier*"},
	"view_bookings":         {"Booking", "BookingAssignment", "BookingService", "BookingVisit*", "Mission*", "WorkReport", "WorkPhoto"},
	"coordinator":           {"Booking", "BookingAssignment", "BookingService", "BookingVisit*", "Mission*"},
	"crew_management":       {"Booking", "BookingAssignment", "Mission*"},
	"mission_tracking":      {"Booking", "Mission*", "BookingAssignment"},
	"finance":               {"LeaderInvoice*", "Expense*", "Credit*", "RevolvingFund*", "ProjectPayment", "Booking"},
	"finance_audit":         {"LeaderInvoice*", "Expense*", "Credit*", "RevolvingFund*", "ProjectPayment", "Booking"},
	"auditing":              {"LeaderInvoice*", "Booking", "BookingAuditIssue"},
	"monitoring":            {"Attendance*", "LeaveRequest", "DisciplineEvent", "StaffRating", "CrewRating", "ExtraTask", "MatrixScore", "PerformanceReview", "WorkReport", "Booking", "BookingAssignment", "Mission*"},
	"mon_attendance":        {"Attendance*", "LeaveRequest"},
	"mon_staff":             {"StaffRating", "CrewRating", "MatrixScore", "PerformanceReview", "DisciplineEvent"},
	"mon_reports":           {"WorkReport", "BookingProgressReport"},
	"mon_field":             {"Booking", "BookingAssignment", "Mission*", "LocationPing"},
	"staff_management":      {"Attendance*", "LeaveRequest", "ExtraTask", "DisciplineEvent", "EmployeeSkill"},
	"project_management":    {"Project*"},
	"my_projects":           {"Project*"},
	"quality_control":       {"Complaint*", "QualityFollowUp", "QualityIssue", "CustomerRating"},
	"complaints":            {"Complaint*"},
	"gps_system":            {"Gps*"},
	"invoice_gps":           {"Gps*"},
	"invoice_dashcam":       {"Booking", "BookingService"},
	"solar_system":          {"Solar*"},
	"training_manage":       {"Training*"},
}

// للكل: مراجع عامة.
var matrixDataCommon = []string{"Service", "Product", "Department", "Employee"}

// ممنوعة على الموظف حتى لو انطابقت ببادئة.
var matrixDataDeny = regexp.MustCompile(`^(Secretary|Ai|Matrix|Permission|Login|Stealth|Notification|Security|Employee(Permission|Ledger|Letter|Commission|AdminScope))`)

var (
	sqlTableRe     = regexp.MustCompile(`(?i)\b(?:from|join)\s+(?:"([A-Za-z0-9_]+)"|([A-Za-z_][A-Za-z0-9_.]*))`)
	sqlSensitiveRe = regexp.MustCompile(`(?i)(salary|wage|password|passhash|phone|iban|token|secret|national)`)
)

// MatrixDataScope الجداول المسموحة لموظف (nil = ماكو مجال).
func MatrixDataScope(perms []string) []string {
	seen := map[string]bool{}
	out := []string{}
	add := func(t string) {
		if !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	any := false
	for _, p := range perms {
		for _, t := range matrixDataScope[p] {
			add(t)
			any = true
		}
	}
	if !any {
		return nil
	}
	for _, t := range matrixDataCommon {
		add(t)
	}
	sort.Strings(out)
	return out
}

func tableAllowed(t string, scope []string) bool {
	if matrixDataDeny.MatchString(t) && t != "Employee" {
		return false
	}
	for _, s := range scope {
		if strings.HasSuffix(s, "*") {
			if strings.HasPrefix(t, strings.TrimSuffix(s, "*")) {
				return true
			}
		} else if t == s {
			return true
		}
	}
	return false
}

// AllowedQuery يتأكد إن الاستعلام يقرا بس من جداول المجال، وبلا أعمدة حساسة.
func AllowedQuery(q string, scope []string) error {
	if sqlSensitiveRe.MatchString(q) {
		return fmt.Errorf("الاستعلام يطلب معلومة حساسة (راتب، هاتف، كلمة سر…) — مو ضمن صلاحياته")
	}
	if regexp.MustCompile(`(?i)information_schema|pg_`).MatchString(q) {
		return fmt.Errorf("استعمل describe_data بدل جداول النظام")
	}
	ctes := map[string]bool{}
	for _, m := range regexp.MustCompile(`(?i)(?:with|,)\s+"?([A-Za-z_][A-Za-z0-9_]*)"?\s+as\s*\(`).FindAllStringSubmatch(q, -1) {
		ctes[strings.ToLower(m[1])] = true
	}
	found := false
	for _, m := range sqlTableRe.FindAllStringSubmatch(q, -1) {
		t := m[1]
		if t == "" {
			t = m[2]
		}
		if ctes[strings.ToLower(t)] || strings.HasPrefix(strings.ToLower(t), "lateral") || strings.HasPrefix(strings.ToLower(t), "generate_series") || strings.HasPrefix(strings.ToLower(t), "unnest") {
			continue
		}
		found = true
		if !tableAllowed(t, scope) {
			return fmt.Errorf("جدول «%s» برّا مجال صلاحياته — گوله بأدب إنه مو ضمن صلاحياتك", t)
		}
	}
	if !found {
		return fmt.Errorf("ما لگيت الجدول بالاستعلام — حط اسم الجدول بين \"\"")
	}
	return nil
}

// ScopedDataTools نسخة query_data/describe_data محصورة بمجال موظف.
func (a *MatrixAgent) ScopedDataTools(scope []string) []agentTool {
	var q, d agentTool
	for _, t := range a.tools {
		switch t.Name {
		case "query_data":
			q = t
		case "describe_data":
			d = t
		}
	}
	if q.Run == nil || d.Run == nil || len(scope) == 0 {
		return nil
	}
	qOrig, dOrig := q.Run, d.Run
	q.Desc = "اقرا بيانات **مجالك بس** بـSELECT: " + strings.Join(scope, "، ") + ". " + q.Desc
	q.Run = func(in map[string]any, m *nameMask) (string, error) {
		if err := AllowedQuery(strArg(in, "sql"), scope); err != nil {
			return "", err
		}
		return qOrig(in, m)
	}
	d.Desc = "خريطة جداول مجاله وأعمدتها (بلا topic = أسماء الجداول المسموحة)."
	d.Run = func(in map[string]any, m *nameMask) (string, error) {
		topic := strings.Trim(strArg(in, "topic"), `" `)
		if topic == "" {
			return "الجداول المسموحة إله (البادئة بـ* = كل جدول يبدي بيها): " + strings.Join(scope, "، ") + "\nاطلب topic باسم جدول حتى تشوف أعمدته.", nil
		}
		out, err := dOrig(in, m)
		if err != nil {
			return "", err
		}
		var keep []string
		for _, line := range strings.Split(out, "\n") {
			if strings.HasPrefix(line, `"`) {
				name := strings.SplitN(strings.TrimPrefix(line, `"`), `"`, 2)[0]
				if !tableAllowed(name, scope) {
					continue
				}
				keep = append(keep, sqlSensitiveRe.ReplaceAllString(line, "—"))
				continue
			}
			if strings.HasPrefix(line, "  ↳") && len(keep) == 0 {
				continue
			}
			keep = append(keep, line)
		}
		return strings.Join(keep, "\n"), nil
	}
	return []agentTool{q, d}
}
