package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type DesignAssetHandler struct {
	repo      *repository.DesignAssetRepository
	employees *repository.EmployeeRepository
}

func NewDesignAssetHandler(r *repository.DesignAssetRepository, employees *repository.EmployeeRepository) *DesignAssetHandler {
	return &DesignAssetHandler{repo: r, employees: employees}
}

// GET /api/design-assets?archived=1
func (h *DesignAssetHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.List(r.URL.Query().Get("archived") == "1")
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المعرض")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/design-assets/categories — التصنيفات من الخادم، حتى
// التسمية وحدة بالمكانين ولا تنكتب مرتين.
func (h *DesignAssetHandler) Categories(w http.ResponseWriter, _ *http.Request) {
	WriteJSON(w, http.StatusOK, model.DesignCategoryLabels)
}

// POST /api/design-assets
//
// ⚠️ الرفع نفسه يمرّ على `POST /api/files` الموجود — هذا المسار
// يخزّن **المفتاح** الراجع منه بس. صفر منطق تخزين جديد.
func (h *DesignAssetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateDesignAssetRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	byID := middleware.EmployeeIDFromContext(r)
	byName := ""
	if e, err := h.employees.FindByID(byID); err == nil && e != nil {
		byName = e.Name
	}
	row, err := h.repo.Create(req, byID, byName)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, row)
}

// PUT /api/design-assets/{id}/archive — أرشفة أو إرجاع.
func (h *DesignAssetHandler) SetArchived(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Archived bool `json:"archived"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	row, err := h.repo.SetArchived(r.PathValue("id"), req.Archived)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر تحديث التصميم")
		return
	}
	WriteJSON(w, http.StatusOK, row)
}
