package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ «اتخاذ إجراء» على صف بصندوق المراقب (أحكام ماتركس) ═══
// (ع): تنبيه الموظف · طلب تبرير · إنشاء متابعة · إحالة للمسؤول.
// ماكو غرامة ولا نقاط هنا — رسائل ومهمة متابعة بس، وكلها تنسجل بسجل النشاط.
type MonitorActionHandler struct {
	reviews *repository.MonitorReviewRepository
	notif   *repository.NotificationRepository
	tasks   *repository.ExtraTaskRepository
}

func NewMonitorActionHandler(r *repository.MonitorReviewRepository, n *repository.NotificationRepository, t *repository.ExtraTaskRepository) *MonitorActionHandler {
	return &MonitorActionHandler{reviews: r, notif: n, tasks: t}
}

// POST /api/monitor-reviews/{id}/action {action: NOTIFY|JUSTIFY|FOLLOW_UP|ESCALATE, note}
func (h *MonitorActionHandler) Act(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Action string `json:"action"`
		Note   string `json:"note"`
	}
	if err := DecodeJSON(r, &b); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	row, err := h.reviews.Get(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, "الصف مو موجود")
		return
	}
	// الموظف المعني: من تفاصيل التأخر، وإلا صاحب الشغل.
	empID, empName, what := "", "", row.Title
	if row.Late != nil && row.Late.EmployeeID != nil {
		empID = *row.Late.EmployeeID
		if row.Late.EmployeeName != nil {
			empName = *row.Late.EmployeeName
		}
		if row.Late.BookingCode != nil {
			what += " — الحجز " + *row.Late.BookingCode
		}
		if row.Late.MinutesLate != nil {
			what += fmt.Sprintf(" (تأخر %d دقيقة)", *row.Late.MinutesLate)
		}
	} else if row.OwnerEmployeeID != nil {
		empID = *row.OwnerEmployeeID
	}
	note := strings.TrimSpace(b.Note)
	extra := ""
	if note != "" {
		extra = " — ملاحظة المراقب: " + note
	}
	me := middleware.EmployeeIDFromContext(r)
	needEmp := func() bool {
		if empID == "" {
			WriteError(w, http.StatusBadRequest, "ما معروف منو الموظف المعني بهذا الصف")
			return false
		}
		return true
	}
	var done string
	switch b.Action {
	case "NOTIFY":
		if !needEmp() {
			return
		}
		_ = h.notif.Create(empID, "monitor_notice", "🔔 تنبيه من المراقب: "+what+extra)
		done = "انبعث تنبيه للموظف"
	case "JUSTIFY":
		if !needEmp() {
			return
		}
		_ = h.notif.Create(empID, "monitor_justify", "📝 المراقب يطلب تبرير: "+what+". اكتب السبب لمسؤولك أو بملاحظة المهمة."+extra)
		done = "انطلب تبرير من الموظف"
	case "FOLLOW_UP":
		if !needEmp() {
			return
		}
		due := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
		desc := "متابعة من صندوق المراقب: " + what + extra
		if _, err := h.tasks.Create(model.CreateExtraTaskRequest{Title: "متابعة: " + row.Title, Description: &desc,
			AssignedToID: empID, Priority: model.ExtraTaskUrgent, DueAt: &due}, me); err != nil {
			WriteError(w, http.StatusBadRequest, "تعذر إنشاء المتابعة: "+err.Error())
			return
		}
		_ = h.notif.Create(empID, "extra_task", "📌 مهمة متابعة من المراقب: "+row.Title+" — موعدها باچر.")
		done = "انعملت مهمة متابعة للموظف (موعدها باچر)"
	case "ESCALATE":
		who := what
		if empName != "" {
			who = empName + ": " + what
		}
		_ = h.notif.CreateForRolesOrPermission([]string{"OWNER", "ADMIN"}, "", "monitor_escalation", "⬆️ إحالة من المراقب — "+who+extra)
		done = "انحالت للمدير والمالك"
	default:
		WriteError(w, http.StatusBadRequest, "إجراء غير معروف")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"done": done})
}
