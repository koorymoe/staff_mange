package handler

import (
	"log"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type RevolvingFundHandler struct {
	repo *repository.RevolvingFundRepository
}

func NewRevolvingFundHandler(r *repository.RevolvingFundRepository) *RevolvingFundHandler {
	return &RevolvingFundHandler{repo: r}
}

func actor(r *http.Request) *string {
	if id := middleware.EmployeeIDFromContext(r); id != "" {
		return &id
	}
	return nil
}

// GET /api/funds — الدوارات وأرصدتها
func (h *RevolvingFundHandler) ListFunds(w http.ResponseWriter, r *http.Request) {
	funds, err := h.repo.ListFunds()
	if err != nil {
		log.Printf("list funds: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الدوارات")
		return
	}
	WriteJSON(w, http.StatusOK, funds)
}

// PUT /api/funds/{id} — تعديل الدوار (المبلغ قابل للتغيير)
func (h *RevolvingFundHandler) UpdateFund(w http.ResponseWriter, r *http.Request) {
	var req model.UpsertFundRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	f, err := h.repo.UpdateFund(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر تعديل الدوار")
		return
	}
	WriteJSON(w, http.StatusOK, f)
}

// POST /api/funds/{id}/topup — تغذية الدوار
func (h *RevolvingFundHandler) Topup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Amount float64 `json:"amount"`
		Notes  *string `json:"notes"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	if err := h.repo.Topup(r.PathValue("id"), req.Amount, req.Notes, actor(r)); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/funds/disburse — المحاسب يسلّم موظف مبلغ
func (h *RevolvingFundHandler) Disburse(w http.ResponseWriter, r *http.Request) {
	var req model.DisburseRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	t, err := h.repo.Disburse(req, actor(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, t)
}

// GET /api/funds/balances — كل الموظفين الساحبين وشكد مطلوب من كل واحد
func (h *RevolvingFundHandler) Balances(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.EmployeeBalances()
	if err != nil {
		log.Printf("fund balances: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب أرصدة الموظفين")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/funds/my-balance — رصيد الموظف الحالي (يظهر بلوحته)
func (h *RevolvingFundHandler) MyBalance(w http.ResponseWriter, r *http.Request) {
	b, err := h.repo.EmployeeBalance(middleware.EmployeeIDFromContext(r))
	if err != nil {
		log.Printf("my fund balance: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب رصيدك")
		return
	}
	WriteJSON(w, http.StatusOK, b)
}

// GET /api/funds/my-transactions — حركات الموظف الحالي
func (h *RevolvingFundHandler) MyTxns(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListTxns(middleware.EmployeeIDFromContext(r), "")
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب حركاتك")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/funds/transactions?employeeId=&status= — للمحاسب
func (h *RevolvingFundHandler) Txns(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListTxns(r.URL.Query().Get("employeeId"), r.URL.Query().Get("status"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الحركات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/funds/settlements — الموظف يرفع تسوية (صرف + مرتجع + وصل)
func (h *RevolvingFundHandler) SubmitSettlement(w http.ResponseWriter, r *http.Request) {
	var req model.SettlementRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	t, err := h.repo.SubmitSettlement(middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, t)
}

// PUT /api/funds/settlements/{id}/review — المحاسب يدقّق ويوافق أو يرفض
func (h *RevolvingFundHandler) ReviewSettlement(w http.ResponseWriter, r *http.Request) {
	var req model.ReviewSettlementRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	t, err := h.repo.ReviewSettlement(r.PathValue("id"), req, actor(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, t)
}
