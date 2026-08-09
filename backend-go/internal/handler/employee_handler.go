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

// POST /api/v1/employees
func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateEmployeeRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}

	employee, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, employee)
}

// PUT /api/v1/employees/{id}
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
