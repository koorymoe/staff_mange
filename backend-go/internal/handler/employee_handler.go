package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

// canSeeSalaries يحدد مين يشوف رواتب باقي الموظفين — الرواتب بيانات حساسة، مو كل
// موظف مسجل دخول يحتاج يشوفها لبقية الكادر.
func canSeeSalaries(role string) bool {
	switch role {
	// OWNER كان ناقص هنا — المالك ما كان يشوف رواتب كادره إطلاقاً
	case "OWNER", "ADMIN", "HR_COORDINATOR", "MONITOR", "FINANCE":
		return true
	default:
		return false
	}
}

type EmployeeHandler struct {
	service *service.EmployeeService
}

func NewEmployeeHandler(s *service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{service: s}
}

// GET /api/v1/employees
func (h *EmployeeHandler) List(w http.ResponseWriter, r *http.Request) {
	employees, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قائمة الموظفين")
		return
	}
	// نبني نسخة مقيّدة حسب دور الطالب — الحقول الي ما تخصه تنشال من الـJSON
	// بالكامل، مو تنرجع null (شوف employee_view.go)
	WriteJSON(w, http.StatusOK, ViewEmployees(employees, middleware.RoleFromContext(r), middleware.EmployeeIDFromContext(r)))
}

// GET /api/v1/employees/archived — الأدمن/المالك فقط، يشوفون الموظفين المؤرشفين والمحذوفين
func (h *EmployeeHandler) ListArchived(w http.ResponseWriter, r *http.Request) {
	employees, err := h.service.ListArchived()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الموظفين المؤرشفين")
		return
	}
	WriteJSON(w, http.StatusOK, employees)
}

// GET /api/v1/employees/{id}
func (h *EmployeeHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/employees/")
	employee, err := h.service.Get(id)
	if errors.Is(err, sql.ErrNoRows) {
		WriteError(w, http.StatusNotFound, "الموظف غير موجود")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب بيانات الموظف")
		return
	}
	WriteJSON(w, http.StatusOK, ViewEmployee(employee, middleware.RoleFromContext(r), middleware.EmployeeIDFromContext(r)))
}

// (إنشاء الموظف انتقل لـEmployeeAdminHandler.Create — المالك أو صلاحية employee_create.)

// PUT /api/v1/employees/{id}
// PUT /api/employees/{id}/identity — {name?, role?}
//
// صلاحية «تعديل اسم الموظف ودوره» (edit_employee_identity): (ع) يريد
// ينطيها لموظف بدون ما يصير مدير نظام. ⚠️ الي عنده الصلاحية بس (مو
// مدير) ما يلمس حسابات المدير/المالك، وما يمنح دور مدير أو مالك،
// وما يغيّر دوره هو — وإلا الصلاحية تصير طريق للترقية الذاتية.
func (h *EmployeeHandler) UpdateIdentity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Name *string `json:"name"`
		Role *string `json:"role"`
	}
	if err := DecodeJSON(r, &body); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	if body.Name != nil {
		n := strings.TrimSpace(*body.Name)
		if n == "" {
			WriteError(w, http.StatusBadRequest, "الاسم ما يكون فارغ")
			return
		}
		body.Name = &n
	}
	role := middleware.RoleFromContext(r)
	if role != "ADMIN" && role != "OWNER" {
		target, err := h.service.Get(id)
		if err != nil || target == nil {
			WriteError(w, http.StatusNotFound, "الموظف غير موجود")
			return
		}
		if target.Role == "ADMIN" || target.Role == "OWNER" {
			WriteError(w, http.StatusForbidden, "حساب المدير والمالك يعدّله المدير بس")
			return
		}
		if body.Role != nil {
			if *body.Role == "ADMIN" || *body.Role == "OWNER" {
				WriteError(w, http.StatusForbidden, "منح دور المدير أو المالك للمدير بس")
				return
			}
			if id == middleware.EmployeeIDFromContext(r) {
				WriteError(w, http.StatusForbidden, "ما تكدر تغيّر دورك إنت")
				return
			}
		}
	}
	employee, err := h.service.Update(id, model.UpdateEmployeeRequest{Name: body.Name, Role: body.Role})
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, employee)
}

// PUT /api/employees/{id}/profile — حقول الملف بس لصاحب صلاحية «تعديل
// ملف الموظف». 🔴 چانت الشاشة تعرض الحقول وتحفظها على مسار المدير،
// فكل حفظ ينرفض وينحسب مخالفة — وخمس مخالفات تقفل حساب الموظف.
// ⚠️ قائمة بيضاء: الاسم والدور وبيانات الدخول والحالة والصورة ما تمر.
func (h *EmployeeHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateEmployeeRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	req.Name, req.Role, req.Status, req.Username, req.Password, req.PhotoURL, req.SecondaryRoles = nil, nil, nil, nil, nil, nil, nil
	id := r.PathValue("id")
	if role := middleware.RoleFromContext(r); role != "ADMIN" && role != "OWNER" {
		if target, err := h.service.Get(id); err != nil || target == nil {
			WriteError(w, http.StatusNotFound, "الموظف غير موجود")
			return
		} else if target.Role == "ADMIN" || target.Role == "OWNER" {
			WriteError(w, http.StatusForbidden, "ملف المدير والمالك يعدّله المدير بس")
			return
		}
	}
	employee, err := h.service.Update(id, req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, employee)
}

func (h *EmployeeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/employees/")

	var req model.UpdateEmployeeRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}

	// بيانات الدخول = فتح حساب بشكل ثاني.
	//
	// المالك طلب إن فتح الحسابات إله وحده، ومنع POST /api/employees
	// لحاله ما يسدّ الباب: مدير النظام يكدر يحط اسم مستخدم وباسورد على
	// **أي** حساب موجود ويدخل بيه — يعني يستولي على حساب المالك نفسه.
	// لهذا نفس القيد على تغيير اسم المستخدم أو الباسورد من هذا المسار.
	//
	// بقية الحقول (الراتب، الدور، المهارات...) تبقى لمدير النظام مثل
	// ما كانت — القيد على بيانات الدخول بس.
	if req.Username != nil || req.Password != nil {
		if role, _ := r.Context().Value(middleware.ContextRole).(string); role != "OWNER" {
			WriteError(w, http.StatusForbidden, "تغيير بيانات الدخول للمالك وحده")
			return
		}
	}

	// ⚠️ نفس نمط بيانات الدخول أعلاه بالضبط: صاحب النظام قرر إن صورة
	// الموظف يحطّها المالك بس — لا ADMIN. والواجهة تخفي زر الرفع عن
	// ADMIN أصلاً، بس المنع الحقيقي هنا لأن إخفاء الزر لحاله ما يمنع
	// نداءً مباشراً بيد أي حساب ADMIN.
	if req.PhotoURL != nil {
		if role, _ := r.Context().Value(middleware.ContextRole).(string); role != "OWNER" {
			WriteError(w, http.StatusForbidden, "تغيير صورة الموظف للمالك وحده")
			return
		}
	}

	employee, err := h.service.Update(id, req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, employee)
}

func extractID(path, prefix string) string {
	return strings.TrimSuffix(strings.TrimPrefix(path, prefix), "/")
}

// GET /api/v1/employees/supervisors
func (h *EmployeeHandler) Supervisors(w http.ResponseWriter, r *http.Request) {
	employees, err := h.service.Supervisors()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المشرفين")
		return
	}
	WriteJSON(w, http.StatusOK, employees)
}

// GET /api/v1/employees/match?serviceId=
func (h *EmployeeHandler) Match(w http.ResponseWriter, r *http.Request) {
	employees, err := h.service.Match(r.URL.Query().Get("serviceId"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, employees)
}

// PUT /api/v1/employees/{id}/skills
func (h *EmployeeHandler) SetSkills(w http.ResponseWriter, r *http.Request) {
	var req model.SetEmployeeSkillsRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	employee, err := h.service.SetSkills(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, employee)
}

// LinkHistoricalRecords يربط سجلات تاريخية (حجوزات/شكاوى مستوردة بالاسم) بحساب موظف
// حالي بنفس الاسم — للموظفين القدامى الي رجعوا للشركة وصار عندهم حساب من جديد.
func (h *EmployeeHandler) LinkHistoricalRecords(w http.ResponseWriter, r *http.Request) {
	bookingsLinked, complaintsLinked, err := h.service.LinkHistoricalRecords(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]int{
		"bookingsLinked":   bookingsLinked,
		"complaintsLinked": complaintsLinked,
	})
}
