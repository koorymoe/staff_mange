package handler

import (
	"fmt"
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

type PermissionHandler struct {
	lockout *repository.SecurityLockoutRepository
	service *service.PermissionService
}

func NewPermissionHandler(s *service.PermissionService, lockout *repository.SecurityLockoutRepository) *PermissionHandler {
	return &PermissionHandler{service: s, lockout: lockout}
}

// GET /api/v1/permissions
func (h *PermissionHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	perms, err := h.service.ListAll()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قائمة الصلاحيات")
		return
	}
	WriteJSON(w, http.StatusOK, perms)
}

// GET /api/v1/permissions/employee/{id}
func (h *PermissionHandler) ListForEmployee(w http.ResponseWriter, r *http.Request) {
	employeeID := extractID(r.URL.Path, "/api/permissions/employee/")
	perms, err := h.service.ListForEmployee(employeeID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب صلاحيات الموظف")
		return
	}
	WriteJSON(w, http.StatusOK, perms)
}

// PUT /api/v1/permissions/employee/{id} — ADMIN فقط (يُطبَّق بالـ middleware بالراوتر)
// GET /api/permissions/employees?permission=project_management&roles=PROJECT_MANAGER,ADMIN
func (h *PermissionHandler) EmployeesWithPermission(w http.ResponseWriter, r *http.Request) {
	permission := r.URL.Query().Get("permission")
	if permission == "" {
		WriteError(w, http.StatusBadRequest, "اسم الصلاحية مطلوب")
		return
	}
	var roles []string
	if raw := r.URL.Query().Get("roles"); raw != "" {
		roles = strings.Split(raw, ",")
	}
	employees, err := h.service.EmployeesWithPermission(permission, roles)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الموظفين")
		return
	}
	WriteJSON(w, http.StatusOK, employees)
}

func (h *PermissionHandler) SetForEmployee(w http.ResponseWriter, r *http.Request) {
	employeeID := extractID(r.URL.Path, "/api/permissions/employee/")

	var req model.SetPermissionsRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}

	perms, err := h.service.SetForEmployee(employeeID, req.PermissionIDs)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر تحديث صلاحيات الموظف")
		return
	}
	// منح/سحب الصلاحيات حدث حسّاس — لازم يبقى مؤرشف ويطلع بلوحة المالك:
	// منو غيّر صلاحيات منو، ومتى، ومن أي جهاز.
	if h.lockout != nil {
		names := make([]string, 0, len(perms))
		for _, p := range perms {
			names = append(names, p.Name)
		}
		by := middleware.EmployeeIDFromContext(r)
		_ = h.lockout.LogEvent(&employeeID, "", "PERMISSIONS_CHANGED",
			fmt.Sprintf("غيّرها الموظف %s — الصلاحيات الحالية: %s", by, strings.Join(names, ", ")),
			clientIP(r), r.UserAgent())
	}
	WriteJSON(w, http.StatusOK, perms)
}

// POST /api/v1/permissions/employee/{id}/apply-defaults — ADMIN فقط
func (h *PermissionHandler) ApplyDefaults(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/permissions/employee/")
	employeeID := strings.TrimSuffix(path, "/apply-defaults")

	perms, err := h.service.ApplyDefaults(employeeID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "الموظف غير موجود")
		return
	}
	WriteJSON(w, http.StatusOK, perms)
}

// GET /api/v1/permissions/role-defaults
func (h *PermissionHandler) RoleDefaults(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, h.service.RoleDefaults())
}
