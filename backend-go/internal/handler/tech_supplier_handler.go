package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
)

// ═══ موردين التقنيين — قرار (ع) 10-06 ═══
// التقني يضيف، والمدير يختار لكل تقني شنو يطلعله. المضاف ما يطلع لصاحبه
// إلا إذا المدير اختاره إله.
type TechSupplierHandler struct{ repo *repository.TechSupplierRepository }

func NewTechSupplierHandler(repo *repository.TechSupplierRepository) *TechSupplierHandler {
	return &TechSupplierHandler{repo: repo}
}

func (h *TechSupplierHandler) guard(w http.ResponseWriter, r *http.Request) bool {
	if !h.repo.CanUse(middleware.EmployeeIDFromContext(r)) {
		WriteError(w, http.StatusForbidden, "موردين التقنيين للتقنيين ومسؤولي الخدمات")
		return false
	}
	return true
}

// GET /api/tech-suppliers/can — حتى الواجهة تعرف تطلّع الخانة.
func (h *TechSupplierHandler) Can(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]bool{"can": h.repo.CanUse(middleware.EmployeeIDFromContext(r))})
}

// GET /api/tech-suppliers/mine
func (h *TechSupplierHandler) Mine(w http.ResponseWriter, r *http.Request) {
	if !h.guard(w, r) {
		return
	}
	rows, err := h.repo.Mine(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

func readTechSupplier(w http.ResponseWriter, r *http.Request) (*repository.TechSupplierIn, bool) {
	var in repository.TechSupplierIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return nil, false
	}
	in.CompanyName, in.Phone = strings.TrimSpace(in.CompanyName), strings.TrimSpace(in.Phone)
	if in.CompanyName == "" || in.Phone == "" {
		WriteError(w, http.StatusBadRequest, "اسم المورّد ورقمه إجباريات")
		return nil, false
	}
	return &in, true
}

// POST /api/tech-suppliers — التقني يضيف (ما يطلعله لحد ما المدير يختاره).
func (h *TechSupplierHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !h.guard(w, r) {
		return
	}
	in, ok := readTechSupplier(w, r)
	if !ok {
		return
	}
	id, err := h.repo.Create(*in, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الإضافة")
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]string{"id": id, "message": "انضاف المورّد ✓ — يطلعلك بقائمتك بعد ما المدير يختاره إلك."})
}

// ── المدير ──

// GET /api/tech-suppliers — كل الموردين ويا منو مختارلهم.
func (h *TechSupplierHandler) All(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.All()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	people, err := h.repo.Eligible()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"suppliers": rows, "people": people})
}

// PUT /api/tech-suppliers/{id}
func (h *TechSupplierHandler) Update(w http.ResponseWriter, r *http.Request) {
	in, ok := readTechSupplier(w, r)
	if !ok {
		return
	}
	if err := h.repo.Update(r.PathValue("id"), *in); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// DELETE /api/tech-suppliers/{id}
func (h *TechSupplierHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Delete(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// PUT /api/tech-suppliers/access/{employeeId} — {supplierIds: [...]}
func (h *TechSupplierHandler) SetAccess(w http.ResponseWriter, r *http.Request) {
	var in struct {
		SupplierIDs []string `json:"supplierIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if err := h.repo.SetForEmployee(r.PathValue("employeeId"), in.SupplierIDs); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
