package handler

import (
	"net/http"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type DepartmentHandler struct {
	repo *repository.DepartmentRepository
}

func NewDepartmentHandler(r *repository.DepartmentRepository) *DepartmentHandler {
	return &DepartmentHandler{repo: r}
}

// GET /api/departments?all=1
//
// القراءة لأي موظف مسجَّل: منتقي «حجز داخل الشركة» يحتاجها، وحصرها
// بصلاحية يخلّي المنتقي فارغاً لمن يفتحه.
// ⚠️ `all=1` (الفعّال وغير الفعّال) للشاشة الإدارية بس — المنتقي
// ياخذ الفعّال حتى ما ينحجز لقسم انلغى.
func (h *DepartmentHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.List(r.URL.Query().Get("all") == "1")
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الأقسام")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/departments — حارس المسار: مالك/مدير أو `departments_manage`.
func (h *DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.SaveDepartmentRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	row, err := h.repo.CreateDepartment(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, row)
}

// PUT /api/departments/{id}
func (h *DepartmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.SaveDepartmentRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	row, err := h.repo.UpdateDepartment(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر تعديل القسم")
		return
	}
	WriteJSON(w, http.StatusOK, row)
}

// POST /api/department-heads
func (h *DepartmentHandler) CreateHead(w http.ResponseWriter, r *http.Request) {
	var req model.SaveDepartmentHeadRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	row, err := h.repo.CreateHead(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, row)
}

// PUT /api/department-heads/{id}
func (h *DepartmentHandler) UpdateHead(w http.ResponseWriter, r *http.Request) {
	var req model.SaveDepartmentHeadRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	row, err := h.repo.UpdateHead(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر تعديل المسؤول")
		return
	}
	WriteJSON(w, http.StatusOK, row)
}
