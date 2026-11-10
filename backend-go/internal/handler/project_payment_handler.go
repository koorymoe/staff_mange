package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"staffmange-api/internal/authz"
	"strings"
	"time"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
)

// ═══ دفعات المشاريع — قرار (ع) 10-06 ═══
// يسجّل: المحاسب، ومدير المشاريع، ومشرف المشروع نفسه.
// يأكد ويلغي: المحاسب (والمدير/المالك). وقيمة العقد: إدارة المشاريع والمحاسب.
type ProjectPaymentHandler struct {
	repo  *repository.ProjectPaymentRepository
	perms *repository.PermissionRepository
}

func NewProjectPaymentHandler(repo *repository.ProjectPaymentRepository, perms *repository.PermissionRepository) *ProjectPaymentHandler {
	return &ProjectPaymentHandler{repo: repo, perms: perms}
}

func (h *ProjectPaymentHandler) isFinance(r *http.Request) bool {
	return authz.SeeAll(r, h.perms, authz.ProjectPaymentsFinance) // (11-04) محرك الصلاحيات
}

func (h *ProjectPaymentHandler) canRecord(r *http.Request, projectID string) bool {
	// (11-04 — المرحلة ٤) القاعدة بالمحرك + الملكية (مشرف المشروع)
	return authz.Can(r, h.perms, authz.ProjectContractValue, h.repo.IsSupervisor(projectID, middleware.EmployeeIDFromContext(r)))
}

func (h *ProjectPaymentHandler) canValue(r *http.Request) bool {
	return authz.Can(r, h.perms, authz.ProjectContractValue)
}

type projectPaymentsOut struct {
	Money     *repository.ProjectMoney    `json:"money"`
	Payments  []repository.ProjectPayment `json:"payments"`
	CanVerify bool                        `json:"canVerify"`
	CanValue  bool                        `json:"canValue"`
}

func (h *ProjectPaymentHandler) out(w http.ResponseWriter, r *http.Request, projectID string) {
	m, err := h.repo.Money(projectID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "المشروع مو موجود")
		return
	}
	list, err := h.repo.List(projectID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, projectPaymentsOut{Money: m, Payments: list, CanVerify: h.isFinance(r),
		CanValue: h.canValue(r)})
}

// GET /api/projects/{id}/payments
func (h *ProjectPaymentHandler) List(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !h.canRecord(r, id) {
		authz.Deny(w, r, authz.ProjectContractValue)
		return
	}
	h.out(w, r, id)
}

// POST /api/projects/{id}/payments
func (h *ProjectPaymentHandler) Add(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !h.canRecord(r, id) {
		authz.Deny(w, r, authz.ProjectContractValue)
		return
	}
	var in repository.ProjectPaymentIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Amount <= 0 {
		WriteError(w, http.StatusBadRequest, "اكتب مبلغ الدفعة")
		return
	}
	if _, err := time.Parse("2006-01-02", in.PaidAt); err != nil {
		WriteError(w, http.StatusBadRequest, "حدد تاريخ الدفعة")
		return
	}
	switch in.Method {
	case "CASH", "TRANSFER", "CHEQUE":
	default:
		in.Method = "CASH"
	}
	// المحاسب الي يسجّل بنفسه = مأكدة؛ غيره تنتظر تأكيد المحاسب.
	if _, err := h.repo.Add(id, middleware.EmployeeIDFromContext(r), in, h.isFinance(r)); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر تسجيل الدفعة")
		return
	}
	h.out(w, r, id)
}

// POST /api/project-payments/{id}/verify
func (h *ProjectPaymentHandler) Verify(w http.ResponseWriter, r *http.Request) {
	if !h.isFinance(r) {
		authz.Deny(w, r, authz.ProjectPaymentsFinance)
		return
	}
	pid := h.repo.ProjectOf(r.PathValue("id"))
	if err := h.repo.Verify(r.PathValue("id"), middleware.EmployeeIDFromContext(r)); err != nil || pid == "" {
		WriteError(w, http.StatusBadRequest, paymentErrMsg(err))
		return
	}
	h.out(w, r, pid)
}

// POST /api/project-payments/{id}/cancel
func (h *ProjectPaymentHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	if !h.isFinance(r) {
		authz.Deny(w, r, authz.ProjectPaymentsFinance)
		return
	}
	var in struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if len([]rune(strings.TrimSpace(in.Reason))) < 3 {
		WriteError(w, http.StatusBadRequest, "اكتب سبب الإلغاء")
		return
	}
	pid := h.repo.ProjectOf(r.PathValue("id"))
	if err := h.repo.Cancel(r.PathValue("id"), middleware.EmployeeIDFromContext(r), strings.TrimSpace(in.Reason)); err != nil || pid == "" {
		WriteError(w, http.StatusBadRequest, paymentErrMsg(err))
		return
	}
	h.out(w, r, pid)
}

// PUT /api/projects/{id}/contract-value
func (h *ProjectPaymentHandler) SetValue(w http.ResponseWriter, r *http.Request) {
	if !h.canValue(r) {
		authz.Deny(w, r, authz.ProjectContractValue)
		return
	}
	var in struct {
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Amount < 0 {
		WriteError(w, http.StatusBadRequest, "اكتب قيمة العقد")
		return
	}
	id := r.PathValue("id")
	if err := h.repo.SetContractValue(id, middleware.EmployeeIDFromContext(r), in.Amount); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر")
		return
	}
	h.out(w, r, id)
}

// GET /api/project-payments/overview — كل المشاريع بفلوسها (المحاسب وإدارة المشاريع).
func (h *ProjectPaymentHandler) Overview(w http.ResponseWriter, r *http.Request) {
	if !authz.Can(r, h.perms, authz.ProjectPaymentsOverview) {
		authz.Deny(w, r, authz.ProjectPaymentsOverview)
		return
	}
	rows, err := h.repo.AllMoney()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// ═══ سجل الدفعات + تصحيح المبلغ — للمالك بس (ع 10-13) ═══

// GET /api/project-payments/audit?by=&since=
func (h *ProjectPaymentHandler) Audit(w http.ResponseWriter, r *http.Request) {
	if !authz.Can(r, h.perms, authz.ProjectPaymentsAudit) {
		authz.Deny(w, r, authz.ProjectPaymentsAudit)
		return
	}
	rows, err := h.repo.Audit(r.URL.Query().Get("by"), r.URL.Query().Get("since"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب السجل")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/project-payments/{id}/amount {amount, reason}
func (h *ProjectPaymentHandler) SetAmount(w http.ResponseWriter, r *http.Request) {
	if !authz.Can(r, h.perms, authz.ProjectPaymentsAudit) {
		authz.Deny(w, r, authz.ProjectPaymentsAudit)
		return
	}
	var in struct {
		Amount float64 `json:"amount"`
		Reason string  `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Amount < 0 {
		WriteError(w, http.StatusBadRequest, "المبلغ مو مفهوم")
		return
	}
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		reason = "تصحيح تكرار"
	}
	if err := h.repo.SetAmount(r.PathValue("id"), middleware.EmployeeIDFromContext(r), in.Amount, reason); err != nil {
		WriteError(w, http.StatusBadRequest, paymentErrMsg(err))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func paymentErrMsg(err error) string {
	if errors.Is(err, repository.ErrPaymentUnchanged) {
		return err.Error()
	}
	return "تعذر"
}
