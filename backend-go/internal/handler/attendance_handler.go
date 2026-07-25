package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/xuri/excelize/v2"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

type AttendanceHandler struct {
	service        *service.AttendanceService
	permissionRepo *repository.PermissionRepository
}

func NewAttendanceHandler(s *service.AttendanceService, permissionRepo *repository.PermissionRepository) *AttendanceHandler {
	return &AttendanceHandler{service: s, permissionRepo: permissionRepo}
}

// canViewEmployee يسمح للموظف نفسه، أو لأي أدمن/مراقب/مالك، أو لأي حد عنده
// صلاحية "monitoring" المخصصة، بمشاهدة أو تصدير سجل حضور موظف معيّن.
func (h *AttendanceHandler) canViewEmployee(r *http.Request, employeeID string) bool {
	requesterID := middleware.EmployeeIDFromContext(r)
	if requesterID == employeeID {
		return true
	}
	role := middleware.RoleFromContext(r)
	if role == "ADMIN" || role == "MONITOR" || role == "OWNER" {
		return true
	}
	if h.permissionRepo != nil {
		perms, err := h.permissionRepo.ListForEmployee(requesterID)
		if err == nil {
			for _, p := range perms {
				if p.Name == "monitoring" {
					return true
				}
			}
		}
	}
	return false
}

// POST /api/attendance/checkin
func (h *AttendanceHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
	employeeID := middleware.EmployeeIDFromContext(r)
	rec, err := h.service.CheckIn(employeeID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, rec)
}

// POST /api/attendance/checkout
func (h *AttendanceHandler) CheckOut(w http.ResponseWriter, r *http.Request) {
	employeeID := middleware.EmployeeIDFromContext(r)
	rec, err := h.service.CheckOut(employeeID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, rec)
}

// GET /api/attendance/mine
func (h *AttendanceHandler) Mine(w http.ResponseWriter, r *http.Request) {
	employeeID := middleware.EmployeeIDFromContext(r)
	rec, err := h.service.Mine(employeeID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب سجل الحضور")
		return
	}
	WriteJSON(w, http.StatusOK, rec)
}

// GET /api/attendance/open — الجلسة المفتوحة حالياً للموظف (لو موجودة)، مع مجموع
// دقائق اليوم عبر كل الجلسات (المؤكدة + الجارية).
func (h *AttendanceHandler) OpenSession(w http.ResponseWriter, r *http.Request) {
	employeeID := middleware.EmployeeIDFromContext(r)
	open, err := h.service.OpenSession(employeeID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب حالة الحضور")
		return
	}
	sessions, totalMinutes, isOpen, err := h.service.TodaySessions(employeeID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب جلسات اليوم")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"open":         open,
		"sessions":     sessions,
		"totalMinutes": totalMinutes,
		"isOpen":       isOpen,
	})
}

// GET /api/attendance/today
func (h *AttendanceHandler) Today(w http.ResponseWriter, r *http.Request) {
	records, err := h.service.Today()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب حضور اليوم")
		return
	}
	WriteJSON(w, http.StatusOK, records)
}

// GET /api/attendance/today-summary — ملخص مجمّع لكل موظف عنده حضور اليوم
// (عدد الجلسات، أول دخول، آخر خروج، مجموع الدقائق، هل نشط الآن).
func (h *AttendanceHandler) TodaySummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.service.TodaySummary()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب ملخص حضور اليوم")
		return
	}
	WriteJSON(w, http.StatusOK, summary)
}

// GET /api/attendance/employee/{id}?month=YYYY-MM
func (h *AttendanceHandler) MonthlyReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.service.MonthlyReport(r.PathValue("id"), r.URL.Query().Get("month"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, report)
}

// PUT /api/attendance/{id}
func (h *AttendanceHandler) Correct(w http.ResponseWriter, r *http.Request) {
	var req model.SetAttendanceCorrectionRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	rec, err := h.service.Correct(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, rec)
}

func formatTimeOrDash(t *time.Time) string {
	if t == nil {
		return "—"
	}
	return t.Local().Format("15:04")
}

func sessionsJoined(sessions []model.Attendance, checkIn bool) string {
	out := ""
	for i, s := range sessions {
		if i > 0 {
			out += "\n"
		}
		if checkIn {
			out += s.CheckIn.Local().Format("15:04")
		} else {
			out += formatTimeOrDash(s.CheckOut)
		}
	}
	if out == "" {
		out = "—"
	}
	return out
}

func setExcelHeaderStyle(f *excelize.File, sheet string, lastCol string) {
	style, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "right", WrapText: true},
	})
	_ = f.SetCellStyle(sheet, "A1", lastCol+"1", style)
}

// GET /api/attendance/export/employee/{id}?month=YYYY-MM
func (h *AttendanceHandler) ExportEmployeeMonth(w http.ResponseWriter, r *http.Request) {
	employeeID := r.PathValue("id")
	if !h.canViewEmployee(r, employeeID) {
		WriteError(w, http.StatusForbidden, "لا تملك صلاحية الوصول لهذا السجل")
		return
	}
	report, err := h.service.MonthlyReport(employeeID, r.URL.Query().Get("month"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	f := excelize.NewFile()
	sheet := "الحضور"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"التاريخ", "وقت الدخول", "وقت الخروج", "مجموع الساعات"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	setExcelHeaderStyle(f, sheet, "D")

	bodyStyle, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{Horizontal: "right", WrapText: true}})

	for i, day := range report.Days {
		row := i + 2
		hours := fmt.Sprintf("%.2f", float64(day.TotalMinutes)/60)
		values := []interface{}{
			day.Date,
			sessionsJoined(day.Sessions, true),
			sessionsJoined(day.Sessions, false),
			hours,
		}
		for c, v := range values {
			cell, _ := excelize.CoordinatesToCellName(c+1, row)
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, bodyStyle)
		}
	}
	_ = f.SetColWidth(sheet, "A", "A", 14)
	_ = f.SetColWidth(sheet, "B", "C", 22)
	_ = f.SetColWidth(sheet, "D", "D", 14)

	filename := fmt.Sprintf("attendance-%s-%s.xlsx", employeeID, report.Month)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, url.QueryEscape(filename)))
	if err := f.Write(w); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر إنشاء ملف الإكسل")
		return
	}
}

// GET /api/attendance/export/today?date=YYYY-MM-DD
func (h *AttendanceHandler) ExportToday(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	var summary []model.EmployeeDailyAttendanceSummary
	var err error
	if date == "" {
		summary, err = h.service.TodaySummary()
		date = time.Now().Format("2006-01-02")
	} else {
		summary, err = h.service.DaySummary(date)
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب بيانات الحضور")
		return
	}

	f := excelize.NewFile()
	sheet := "حضور اليوم"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"الموظف", "وقت الدخول الأول", "وقت الخروج الأخير", "عدد الجلسات", "مجموع الساعات", "الحالة"}
	for i, hd := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, hd)
	}
	setExcelHeaderStyle(f, sheet, "F")

	bodyStyle, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{Horizontal: "right", WrapText: true}})

	for i, row := range summary {
		r := i + 2
		name := row.EmployeeID
		if row.Employee != nil {
			name = row.Employee.Name
		}
		status := "منتهي"
		if row.CurrentlyActive {
			status = "نشط الآن"
		}
		hours := fmt.Sprintf("%.2f", float64(row.TotalMinutes)/60)
		values := []interface{}{
			name,
			row.FirstCheckIn.Local().Format("15:04"),
			formatTimeOrDash(row.LastCheckOut),
			row.SessionsCount,
			hours,
			status,
		}
		for c, v := range values {
			cell, _ := excelize.CoordinatesToCellName(c+1, r)
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, bodyStyle)
		}
	}
	_ = f.SetColWidth(sheet, "A", "A", 20)
	_ = f.SetColWidth(sheet, "B", "C", 18)
	_ = f.SetColWidth(sheet, "D", "F", 14)

	filename := fmt.Sprintf("attendance-today-%s.xlsx", date)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, url.QueryEscape(filename)))
	if err := f.Write(w); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر إنشاء ملف الإكسل")
		return
	}
}
