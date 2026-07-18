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
	case "ADMIN", "HR_COORDINATOR", "MONITOR", "FINANCE":
		return true
	default:
		return false
	}
}

// redactSalaries يشيل الراتب من أي موظف غير الشخص نفسه إذا الطالب ما عنده صلاحية يشوف الرواتب
func redactSalaries(employees []model.Employee, role, selfID string) {
	if canSeeSalaries(role) {
		return
	}
	for i := range employees {
		if employees[i].ID != selfID {
			employees[i].Salary = nil
		}
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
	redactSalaries(employees, middleware.RoleFromContext(r), middleware.EmployeeIDFromContext(r))
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
	if !canSeeSalaries(middleware.RoleFromContext(r)) && employee.ID != middleware.EmployeeIDFromContext(r) {
		employee.Salary = nil
	}
	WriteJSON(w, http.StatusOK, employee)
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
