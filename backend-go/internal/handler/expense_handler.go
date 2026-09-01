package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

type ExpenseHandler struct {
	service *service.ExpenseService
	perms   *repository.PermissionRepository
}

func NewExpenseHandler(s *service.ExpenseService, perms *repository.PermissionRepository) *ExpenseHandler {
	return &ExpenseHandler{service: s, perms: perms}
}

// canSeeAllExpenses منو يشوف مصاريف الشركة كلها؟
//
// ⚠️⚠️ **ثغرة چانت مفتوحة**: المسار عليه `requireAuth` بس، والمعامل
// `employeeId` يجي من الرابط — فأي موظف يشيله ويقرا **مصاريف الشركة
// كلها**: مبالغ زملائه وأوصافها.
//
// ⚠️ وقائمة السماح مبنية على **الشاشات الشغّالة اليوم** بالضبط
// (تدقيق الحسابات ومراجعة المصاريف)، حتى ما ينكسر ولا موظف: الي
// چان يشوف يبقى يشوف، والي ما إله يشوف مصاريفه هو.
func (h *ExpenseHandler) canSeeAll(r *http.Request) bool {
	switch middleware.RoleFromContext(r) {
	case "ADMIN", "OWNER", "FINANCE", "MONITOR":
		return true
	}
	if h.perms == nil {
		return false
	}
	rows, err := h.perms.ListForEmployee(middleware.EmployeeIDFromContext(r))
	if err != nil {
		return false
	}
	for _, p := range rows {
		if p.Name == "expenses_manage" || p.Name == "finance" {
			return true
		}
	}
	return false
}

// GET /api/v1/expenses?employeeId=
func (h *ExpenseHandler) List(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employeeId")
	// ⚠️ الي ما عنده صلاحية مالية يشوف **مصاريفه هو** مهما كتب
	// بالرابط — الحصر بالخادم مو بالواجهة.
	if !h.canSeeAll(r) {
		employeeID = middleware.EmployeeIDFromContext(r)
	}
	expenses, err := h.service.List(employeeID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المصاريف")
		return
	}
	WriteJSON(w, http.StatusOK, expenses)
}

// POST /api/v1/expenses
func (h *ExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateExpenseRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	// المصروف دايماً باسم صاحب الجلسة الحالي — ما نثق بـ employeeId المرسل
	// بالطلب حتى ما يقدر موظف يسجل مصروف باسم زميله.
	req.EmployeeID = middleware.EmployeeIDFromContext(r)
	expense, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, expense)
}

// PUT /api/v1/expenses/{id}/status
func (h *ExpenseHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateExpenseStatusRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	expense, err := h.service.UpdateStatus(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, expense)
}
