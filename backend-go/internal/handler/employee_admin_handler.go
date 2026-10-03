package handler

import (
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

// ═══ صلاحيات إدارة الموظفين الي تنطى لأي أحد ═══
// (ع): تعديل كل البيانات، إيقاف/تفعيل الحساب، إضافة موظف، جدول الدوام.
//
// ⚠️ نفس حماية «تعديل الاسم والدور» الموجودة: الي عنده الصلاحية (مو
// مدير) ما يلمس حساب مدير/مالك، وما يمنح دور مدير/مالك، وما يرقّي نفسه.
// وكل منع يصير هنا بالمعالج (مو بالحارس) فما ينحسب مخالفة على الموظف.

type EmployeeAdminHandler struct {
	service *service.EmployeeService
	perms   *repository.PermissionRepository
	repo    *repository.EmployeeRepository
}

func NewEmployeeAdminHandler(s *service.EmployeeService, p *repository.PermissionRepository, r *repository.EmployeeRepository) *EmployeeAdminHandler {
	return &EmployeeAdminHandler{service: s, perms: p, repo: r}
}

func isTop(role string) bool { return role == "ADMIN" || role == "OWNER" }

// guardTarget: يرجّع false ويكتب الخطأ لو الهدف حساب مدير/مالك والطالب مو مدير.
func (h *EmployeeAdminHandler) guardTarget(w http.ResponseWriter, r *http.Request, id string) bool {
	if isTop(middleware.RoleFromContext(r)) {
		return true
	}
	t, err := h.service.Get(id)
	if err != nil || t == nil {
		WriteError(w, http.StatusNotFound, "الموظف غير موجود")
		return false
	}
	if isTop(t.Role) {
		WriteError(w, http.StatusForbidden, "حساب المدير والمالك يعدّله المدير بس")
		return false
	}
	return true
}

// PUT /api/employees/{id}/all — كل بيانات الموظف (قائمة بيضاء).
func (h *EmployeeAdminHandler) UpdateAll(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req model.UpdateEmployeeRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	// بيانات الدخول والصورة والحالة تبقى بمساراتها (المالك / الإيقاف).
	req.Username, req.Password, req.PhotoURL, req.Status = nil, nil, nil, nil
	if req.Name != nil {
		n := strings.TrimSpace(*req.Name)
		if n == "" {
			WriteError(w, http.StatusBadRequest, "الاسم ما يكون فارغ")
			return
		}
		req.Name = &n
	}
	if !h.guardTarget(w, r, id) {
		return
	}
	if !isTop(middleware.RoleFromContext(r)) {
		if req.Role != nil && isTop(*req.Role) {
			WriteError(w, http.StatusForbidden, "منح دور المدير أو المالك للمدير بس")
			return
		}
		if req.Role != nil && id == middleware.EmployeeIDFromContext(r) {
			WriteError(w, http.StatusForbidden, "ما تكدر تغيّر دورك إنت")
			return
		}
		req.SecondaryRoles = nil
	}
	emp, err := h.service.Update(id, req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, emp)
}

// PUT /api/employees/{id}/suspend — {reason}: إيقاف بلا حذف.
func (h *EmployeeAdminHandler) Suspend(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Reason string `json:"reason"`
	}
	_ = DecodeJSON(r, &body)
	if strings.TrimSpace(body.Reason) == "" {
		WriteError(w, http.StatusBadRequest, "اكتب سبب الإيقاف (مثلاً: ترك الشركة)")
		return
	}
	if id == middleware.EmployeeIDFromContext(r) {
		WriteError(w, http.StatusForbidden, "ما تكدر توقف حسابك إنت")
		return
	}
	if !h.guardTarget(w, r, id) {
		return
	}
	if err := h.repo.SetSuspended(id, true, strings.TrimSpace(body.Reason), middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// PUT /api/employees/{id}/reactivate — يرجّع الحساب.
func (h *EmployeeAdminHandler) Reactivate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !h.guardTarget(w, r, id) {
		return
	}
	if err := h.repo.SetSuspended(id, false, "", middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/employees/suspended — الموقوفين (لصاحب صلاحية الإيقاف).
func (h *EmployeeAdminHandler) ListSuspended(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListSuspended()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الجلب")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/employees — المالك، أو صاحب «إضافة موظف» (بلا أدوار عليا).
func (h *EmployeeAdminHandler) Create(w http.ResponseWriter, r *http.Request) {
	role := middleware.RoleFromContext(r)
	if role != "OWNER" {
		ok, _ := h.perms.HasPermission(middleware.EmployeeIDFromContext(r), "employee_create")
		if !ok {
			WriteError(w, http.StatusForbidden, "إضافة الموظفين للمالك أو لصاحب صلاحية «إضافة موظف جديد»")
			return
		}
	}
	var req model.CreateEmployeeRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	if role != "OWNER" {
		if req.Role != nil && isTop(*req.Role) {
			WriteError(w, http.StatusForbidden, "حساب مدير أو مالك يفتحه المالك بس")
			return
		}
		req.SecondaryRoles = nil
	}
	emp, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, emp)
}

// PUT /api/employees/{id}/schedule — {shift, shiftStart, shiftEnd}
func (h *EmployeeAdminHandler) Schedule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Shift      *string `json:"shift"`
		ShiftStart *string `json:"shiftStart"`
		ShiftEnd   *string `json:"shiftEnd"`
	}
	if err := DecodeJSON(r, &body); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	for _, t := range []*string{body.ShiftStart, body.ShiftEnd} {
		if t != nil && *t != "" && !validClock(*t) {
			WriteError(w, http.StatusBadRequest, "الوقت لازم يكون بصيغة 08:00")
			return
		}
	}
	if !h.guardTarget(w, r, id) {
		return
	}
	emp, err := h.service.Update(id, model.UpdateEmployeeRequest{Shift: body.Shift, ShiftStart: body.ShiftStart, ShiftEnd: body.ShiftEnd})
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, emp)
}

// GET /api/work-schedule — جدول كل الموظفين الفعّالين وحضورهم اليوم.
func (h *EmployeeAdminHandler) WorkSchedule(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.WorkSchedule()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الجدول")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

func validClock(s string) bool {
	if len(s) != 5 || s[2] != ':' {
		return false
	}
	h := int(s[0]-'0')*10 + int(s[1]-'0')
	m := int(s[3]-'0')*10 + int(s[4]-'0')
	return s[0] >= '0' && s[0] <= '2' && s[1] >= '0' && s[1] <= '9' && s[3] >= '0' && s[3] <= '5' && s[4] >= '0' && s[4] <= '9' && h < 24 && m < 60
}
