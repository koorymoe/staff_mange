package handler

import (
	"net/http"
	"slices"
	"strings"

	"github.com/lib/pq"

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
	scopeRepo = r
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
		if t, err := h.service.Get(id); err == nil && t != nil && !h.inScope(r, t.Role) {
			WriteError(w, http.StatusForbidden, "ما مسموحلك تعدّل موظف بهذا الدور — المالك يحدد الأدوار الي تعدّلها")
			return
		}
		if req.Role != nil && !h.inScope(r, *req.Role) {
			WriteError(w, http.StatusForbidden, "ما مسموحلك تحوّل الموظف لهذا الدور")
			return
		}
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
		newRole := "EMPLOYEE"
		if req.Role != nil && *req.Role != "" {
			newRole = *req.Role
		}
		if !h.inScope(r, newRole) {
			WriteError(w, http.StatusForbidden, "ما مسموحلك تضيف موظف بهذا الدور — المالك يحدد الأدوار الي تضيفها")
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

// ── نطاق الأدوار للمفوَّض (قرار (ع) 10-11) ──
// «الموظف الي انطيه صلاحية يضيف موظف، اني احدد شنو الأدوار الي يضيفهم».
// بلا نطاق محدد = ما يضيف ولا يعدّل أحد (الافتراضي مقفول، مو مفتوح).

// scopeRepo يتعبّى من NewEmployeeAdminHandler — حتى معالج «الاسم والدور»
// (EmployeeHandler) يفحص نفس النطاق.
var scopeRepo *repository.EmployeeRepository

func (h *EmployeeAdminHandler) scopeOf(empID string) []string { return adminScopeOf(empID) }

func adminScopeOf(empID string) []string {
	var roles pq.StringArray
	if scopeRepo != nil {
		_ = scopeRepo.DB().Get(&roles, `SELECT roles FROM "EmployeeAdminScope" WHERE "employeeId" = $1`, empID)
	}
	return roles
}

func (h *EmployeeAdminHandler) inScope(r *http.Request, role string) bool { return roleInAdminScope(r, role) }

func roleInAdminScope(r *http.Request, role string) bool {
	if isTop(middleware.RoleFromContext(r)) {
		return true
	}
	return slices.Contains(adminScopeOf(middleware.EmployeeIDFromContext(r)), role)
}

// GET /api/employees/{id}/admin-scope — المالك والمدير
func (h *EmployeeAdminHandler) GetScope(w http.ResponseWriter, r *http.Request) {
	roles := h.scopeOf(r.PathValue("id"))
	if roles == nil {
		roles = []string{}
	}
	WriteJSON(w, http.StatusOK, map[string][]string{"roles": roles})
}

// GET /api/employees/my-admin-scope — الأدوار المسموحة إلي (حتى القائمة تبين بس هي)
func (h *EmployeeAdminHandler) MyScope(w http.ResponseWriter, r *http.Request) {
	if isTop(middleware.RoleFromContext(r)) {
		WriteJSON(w, http.StatusOK, map[string]any{"all": true, "roles": []string{}})
		return
	}
	roles := h.scopeOf(middleware.EmployeeIDFromContext(r))
	if roles == nil {
		roles = []string{}
	}
	WriteJSON(w, http.StatusOK, map[string]any{"all": false, "roles": roles})
}

// PUT /api/employees/{id}/admin-scope {roles} — المالك والمدير
func (h *EmployeeAdminHandler) SetScope(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Roles []string `json:"roles"`
	}
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	clean := []string{}
	for _, x := range in.Roles {
		x = strings.TrimSpace(x)
		if x != "" && !isTop(x) && !slices.Contains(clean, x) {
			clean = append(clean, x)
		}
	}
	if _, err := h.repo.DB().Exec(`INSERT INTO "EmployeeAdminScope" ("employeeId", roles, "updatedById") VALUES ($1, $2, $3)
		ON CONFLICT ("employeeId") DO UPDATE SET roles = EXCLUDED.roles, "updatedById" = EXCLUDED."updatedById", "updatedAt" = now()`,
		r.PathValue("id"), pq.StringArray(clean), middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string][]string{"roles": clean})
}
