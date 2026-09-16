package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"staffmange-api/internal/service"
)

// EmployeeStatsHandler يخدم صفحة إحصائيات الموظفين الشهرية (OWNER/ADMIN حصراً،
// مقيّدة بالبوابة على مستوى main.go بنفس نمط requireAdmin المستخدم بباقي
// الصفحات الحساسة).
type EmployeeStatsHandler struct {
	service *service.EmployeeMonthlyStatsService
}

func NewEmployeeStatsHandler(s *service.EmployeeMonthlyStatsService) *EmployeeStatsHandler {
	return &EmployeeStatsHandler{service: s}
}

func monthParam(r *http.Request) string {
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	return month
}

// GET /api/employee-stats/monthly?month=2026-07
func (h *EmployeeStatsHandler) Monthly(w http.ResponseWriter, r *http.Request) {
	month := monthParam(r)
	stats, err := h.service.Monthly(month)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}

// GET /api/employee-stats/range?from=2026-07-01&to=2026-07-07
func (h *EmployeeStatsHandler) Range(w http.ResponseWriter, r *http.Request) {
	from, to := r.URL.Query().Get("from"), r.URL.Query().Get("to")
	stats, err := h.service.Range(from, to)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}

// GET /api/employee-stats/curve/{employeeId}?months=6&month=2026-07
func (h *EmployeeStatsHandler) Curve(w http.ResponseWriter, r *http.Request) {
	months, _ := strconv.Atoi(r.URL.Query().Get("months"))
	curve, err := h.service.Curve(r.PathValue("employeeId"), months, r.URL.Query().Get("month"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, curve)
}

// GET /api/employee-stats/monthly/export?month=2026-07
func (h *EmployeeStatsHandler) MonthlyExport(w http.ResponseWriter, r *http.Request) {
	month := monthParam(r)
	stats, err := h.service.Monthly(month)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	f := excelize.NewFile()
	sheet := "إحصائيات الموظفين"
	f.SetSheetName("Sheet1", sheet)
	// ⚠️ **الإكسل چان يصدّر ١٠ أعمدة والشاشة تعرض ١٨** — الصيانة
	// المجانية وأعمال داخل الشركة وعدد الخدمات ونقاط الكي بي اي
	// الكاملة كانوا **مفقودين من الملف**. و(ع) يطبع هذا الملف
	// ويقارنه بورقته، فملف ناقص يعني قرار على نصف صورة.
	//
	// الترتيب **نفس ترتيب الشاشة** بالضبط — عمود ينزاح بالملف عن
	// الشاشة يخلي المقارنة تكذب بلا ما أحد يلاحظ.
	headers := []string{
		"الموظف", "الدور", "عدد الخدمات التي يعرفها", "سرعة العمل",
		"نظافة السيارة", "عدد تقييمات السيارة", "الشكاوى", "عدد المبيعات",
		"الحجوزات المكتملة", "كل الحجوزات المسندة", "حجوزات الصيانة",
		"صيانات مجانية", "أعمال داخل الشركة", "أنواع الأعمال",
		"نقاط الكي بي اي", "تقييم يدوي", "قيمة النقاط اليدوية",
		"المبالغ الي دخّلها", "راتبه", "تغطية الراتب %", "إجمالي العمولة",
	}
	for i, hd := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, hd)
	}
	setExcelHeaderStyle(f, sheet, "U")

	bodyStyle, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{Horizontal: "right", WrapText: true}})

	for i, s := range stats {
		row := i + 2
		workSpeed := "—"
		if s.WorkSpeedScore != nil {
			workSpeed = fmt.Sprintf("%.2f", *s.WorkSpeedScore)
		}
		cleanliness := "—"
		if s.VehicleCleanlinessScore != nil {
			cleanliness = fmt.Sprintf("%.2f", *s.VehicleCleanlinessScore)
		}
		// ⚠️ **«—» مو «٠»** بالملف مثل الشاشة: راتب مو مكتوب يعني
		// ماكو مقارنة، ونسبة على راتب صفر لا نهاية.
		salary := "—"
		if s.Salary != nil {
			salary = fmt.Sprintf("%.0f", *s.Salary)
		}
		coverage := "—"
		if s.SalaryCoverage != nil {
			coverage = fmt.Sprintf("%.1f", *s.SalaryCoverage)
		}
		workTypes := "—"
		if len(s.InHouseWorkTypes) > 0 {
			workTypes = strings.Join(s.InHouseWorkTypes, " · ")
		}
		values := []interface{}{
			s.EmployeeName, s.Role, s.ServicesKnownCount, workSpeed,
			cleanliness, s.VehicleRatingsCount, s.ComplaintsCount, s.SalesCount,
			s.CompletedBookingsCount, s.TotalBookingsCount, s.MaintenanceBookingsCount,
			s.FreeMaintenanceCount, s.InHouseWorksCount, workTypes,
			s.SmartKpiPoints, s.KpiPoints, fmt.Sprintf("%.0f", s.KpiPointsValue),
			fmt.Sprintf("%.0f", s.RevenueBrought), salary, coverage,
			fmt.Sprintf("%.2f", s.TotalCommission),
		}
		for c, v := range values {
			cell, _ := excelize.CoordinatesToCellName(c+1, row)
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, bodyStyle)
		}
	}
	_ = f.SetColWidth(sheet, "A", "A", 22)
	_ = f.SetColWidth(sheet, "B", "U", 16)
	_ = f.SetColWidth(sheet, "N", "N", 30)
	_ = f.SetColWidth(sheet, "R", "T", 18)

	filename := fmt.Sprintf("employee-stats-%s.xlsx", month)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, url.QueryEscape(filename)))
	if err := f.Write(w); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر إنشاء ملف الإكسل")
		return
	}
}
