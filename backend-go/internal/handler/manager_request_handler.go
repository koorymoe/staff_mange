package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
)

// ═══ رفع طلب للمدير (ع) 11-04 ═══
type ManagerRequestHandler struct {
	db   *sqlx.DB
	repo *repository.ManagerRequestRepository
}

func NewManagerRequestHandler(db *sqlx.DB, repo *repository.ManagerRequestRepository) *ManagerRequestHandler {
	return &ManagerRequestHandler{db: db, repo: repo}
}

// POST /api/manager-requests — أي موظف. attachInventory يرفق أرقام المخازن من الخادم
// (مو من الواجهة) حتى الرقم الي يوصل المدير حقيقي.
func (h *ManagerRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title           string `json:"title"`
		Details         string `json:"details"`
		Priority        string `json:"priority"`
		AttachInventory bool   `json:"attachInventory"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	ctx := ""
	if in.AttachInventory {
		ctx = repository.InventoryContext(h.db)
	}
	id, err := h.repo.Create(middleware.EmployeeIDFromContext(r), in.Title, in.Details, in.Priority, ctx, "SCREEN")
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// GET /api/manager-requests/inventory-context — معاينة الأرقام قبل الإرسال.
func (h *ManagerRequestHandler) InventoryContext(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"context": repository.InventoryContext(h.db)})
}

// GET /api/manager-requests/mine
func (h *ManagerRequestHandler) Mine(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.Mine(middleware.EmployeeIDFromContext(r))
	if err != nil {
		log.Printf("manager requests mine: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب طلباتك")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/manager-requests — المدير والمالك.
func (h *ManagerRequestHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.All()
	if err != nil {
		log.Printf("manager requests: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الطلبات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/manager-requests/{id}/handle
func (h *ManagerRequestHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Status string `json:"status"`
		Reply  string `json:"reply"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if err := h.repo.Handle(r.PathValue("id"), middleware.EmployeeIDFromContext(r), in.Status, in.Reply); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
