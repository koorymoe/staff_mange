package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
)

// ═══ ساعات العمل من البيت ═══
// صلاحية «remote_hours_manage»: المسؤول يضيف ساعات/دقائق/ثواني لموظف بيوم
// معيّن، والنظام يجمع، وبنهاية الشهر ينزّل Excel. الحارس بـmain.go.
type RemoteHoursHandler struct{ repo *repository.RemoteHoursRepository }

func NewRemoteHoursHandler(r *repository.RemoteHoursRepository) *RemoteHoursHandler {
	return &RemoteHoursHandler{repo: r}
}

var baghdad, _ = time.LoadLocation("Asia/Baghdad")

func remoteMonth(r *http.Request) string {
	m := r.URL.Query().Get("month")
	if _, err := time.Parse("2006-01", m); err != nil {
		return time.Now().In(baghdad).Format("2006-01")
	}
	return m
}

// HMS ثواني ← «س:د:ث» (الساعات ممكن تعدّي ٢٤ بمجموع الشهر).
func HMS(sec int) string {
	return fmt.Sprintf("%d:%02d:%02d", sec/3600, (sec%3600)/60, sec%60)
}

// POST /api/remote-hours
func (h *RemoteHoursHandler) Create(w http.ResponseWriter, r *http.Request) {
	var b struct {
		EmployeeID string  `json:"employeeId"`
		Date       string  `json:"date"`
		Hours      int     `json:"hours"`
		Minutes    int     `json:"minutes"`
		Seconds    int     `json:"seconds"`
		Note       *string `json:"note"`
	}
	if err := DecodeJSON(r, &b); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	day, err := time.ParseInLocation("2006-01-02", b.Date, baghdad)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "التاريخ مو صحيح")
		return
	}
	if day.After(time.Now().In(baghdad)) {
		WriteError(w, http.StatusBadRequest, "ما تكدر تسجّل ساعات ليوم بعده ما جا")
		return
	}
	if b.Hours < 0 || b.Minutes < 0 || b.Seconds < 0 || b.Minutes > 59 || b.Seconds > 59 {
		WriteError(w, http.StatusBadRequest, "الدقائق والثواني من 0 لـ59، وماكو أرقام سالبة")
		return
	}
	total := b.Hours*3600 + b.Minutes*60 + b.Seconds
	if total <= 0 {
		WriteError(w, http.StatusBadRequest, "اكتب المدة")
		return
	}
	if !h.repo.EmployeeActive(b.EmployeeID) {
		WriteError(w, http.StatusBadRequest, "الموظف مو موجود أو حسابه موقوف")
		return
	}
	if already := h.repo.DaySeconds(b.EmployeeID, b.Date); already+total > 24*3600 {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("مجموع هاليوم يصير %s — أكثر من ٢٤ ساعة. المسجّل قبل: %s", HMS(already+total), HMS(already)))
		return
	}
	if b.Note != nil {
		n := strings.TrimSpace(*b.Note)
		b.Note = &n
	}
	id, err := h.repo.Create(b.EmployeeID, b.Date, total, b.Note, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]any{"id": id, "seconds": total})
}

// GET /api/remote-hours?employeeId=&month=
func (h *RemoteHoursHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.List(r.URL.Query().Get("employeeId"), remoteMonth(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الساعات")
		return
	}
	total := 0
	for _, x := range rows {
		total += x.Seconds
	}
	WriteJSON(w, http.StatusOK, map[string]any{"entries": rows, "totalSeconds": total})
}

// GET /api/remote-hours/summary?month=
func (h *RemoteHoursHandler) Summary(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.Summary(remoteMonth(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الملخص")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// DELETE /api/remote-hours/{id} — للي سجّلها أو للمدير.
func (h *RemoteHoursHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	by, ok := h.repo.AddedBy(id)
	if !ok {
		WriteError(w, http.StatusNotFound, "السجل مو موجود")
		return
	}
	role := middleware.RoleFromContext(r)
	if role != "ADMIN" && role != "OWNER" && by != middleware.EmployeeIDFromContext(r) {
		WriteError(w, http.StatusForbidden, "تحذف بس الساعات الي إنت سجّلتها")
		return
	}
	if err := h.repo.Delete(id); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحذف")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/remote-hours/export?month=&employeeId=
func (h *RemoteHoursHandler) Export(w http.ResponseWriter, r *http.Request) {
	month := remoteMonth(r)
	empID := r.URL.Query().Get("employeeId")
	rows, err := h.repo.List(empID, month)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الساعات")
		return
	}
	f := excelize.NewFile()
	body, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{Horizontal: "right", WrapText: true}})
	put := func(sheet string, row int, vals ...any) {
		for c, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(c+1, row)
			_ = f.SetCellValue(sheet, cell, v)
			if row > 1 {
				_ = f.SetCellStyle(sheet, cell, cell, body)
			}
		}
	}

	// الملخص
	sum := "الملخص"
	f.SetSheetName("Sheet1", sum)
	_ = f.SetSheetView(sum, 0, &excelize.ViewOptions{RightToLeft: boolPtr(true)})
	put(sum, 1, "الموظف", "عدد الأيام", "المجموع (س:د:ث)", "المجموع بالساعات")
	setExcelHeaderStyle(f, sum, "D")
	type agg struct {
		name string
		days map[string]bool
		sec  int
	}
	order := []string{}
	byEmp := map[string]*agg{}
	grand := 0
	for _, x := range rows {
		a, ok := byEmp[x.EmployeeID]
		if !ok {
			a = &agg{name: x.EmployeeName, days: map[string]bool{}}
			byEmp[x.EmployeeID] = a
			order = append(order, x.EmployeeID)
		}
		a.days[x.WorkDate.Format("2006-01-02")] = true
		a.sec += x.Seconds
		grand += x.Seconds
	}
	i := 2
	for _, id := range order {
		a := byEmp[id]
		put(sum, i, a.name, len(a.days), HMS(a.sec), fmt.Sprintf("%.2f", float64(a.sec)/3600))
		i++
	}
	put(sum, i+1, "المجموع الكلي", "", HMS(grand), fmt.Sprintf("%.2f", float64(grand)/3600))
	_ = f.SetColWidth(sum, "A", "A", 28)
	_ = f.SetColWidth(sum, "B", "D", 18)

	// التفاصيل
	det := "التفاصيل"
	_, _ = f.NewSheet(det)
	_ = f.SetSheetView(det, 0, &excelize.ViewOptions{RightToLeft: boolPtr(true)})
	put(det, 1, "الموظف", "التاريخ", "المدة (س:د:ث)", "ملاحظة", "سجّلها", "وقت التسجيل")
	setExcelHeaderStyle(f, det, "F")
	for j, x := range rows {
		note := ""
		if x.Note != nil {
			note = *x.Note
		}
		put(det, j+2, x.EmployeeName, x.WorkDate.Format("2006-01-02"), HMS(x.Seconds), note, x.AddedBy, x.CreatedAt.In(baghdad).Format("2006-01-02 15:04"))
	}
	_ = f.SetColWidth(det, "A", "A", 28)
	_ = f.SetColWidth(det, "B", "C", 16)
	_ = f.SetColWidth(det, "D", "D", 34)
	_ = f.SetColWidth(det, "E", "F", 20)

	name := fmt.Sprintf("remote-hours-%s.xlsx", month)
	if empID != "" && len(rows) > 0 {
		name = fmt.Sprintf("remote-hours-%s-%s.xlsx", rows[0].EmployeeName, month)
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename*=UTF-8''%s`, url.PathEscape(name)))
	if err := f.Write(w); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر إنشاء ملف الإكسل")
	}
}

func boolPtr(b bool) *bool { return &b }
