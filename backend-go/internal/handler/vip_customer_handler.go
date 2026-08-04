package handler

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

type VipCustomerHandler struct {
	repo      *repository.VipCustomerRepository
	customers *service.CustomerService
}

func NewVipCustomerHandler(repo *repository.VipCustomerRepository, customers *service.CustomerService) *VipCustomerHandler {
	return &VipCustomerHandler{repo: repo, customers: customers}
}

// GET /api/vip-customers — التفاصيل الكاملة، لمدير النظام فقط (محمي بالراوت).
func (h *VipCustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قائمة الشخصيات المهمة")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

// GET /api/vip-customers/ids — معرّفات الزبائن المعلّمين فقط، لأي موظف مسجل دخول
// (حتى تبيّن الواجهة الزر مضغوط أو لا، بدون كشف التفاصيل).
func (h *VipCustomerHandler) ListIDs(w http.ResponseWriter, r *http.Request) {
	ids, err := h.repo.ListCustomerIDs()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب القائمة")
		return
	}
	WriteJSON(w, http.StatusOK, ids)
}

// POST /api/vip-customers — أي موظف مسجل دخول يقدر يعلّم زبون بضغطة زر.
func (h *VipCustomerHandler) Mark(w http.ResponseWriter, r *http.Request) {
	var req model.MarkVipCustomerRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	// الإضافة اليدوية: الإداري يدز الرقم بس، والنظام يطلع الزبون بمعلوماته.
	// نحل الرقم للزبون هنا بالسيرفر مو بالواجهة، حتى ما ينضاف زبون وهمي
	// بإرسال customerId مصنوع من العميل.
	if strings.TrimSpace(req.CustomerID) == "" && strings.TrimSpace(req.Phone) != "" {
		c, err := h.customers.Lookup(strings.TrimSpace(req.Phone))
		if err != nil || c == nil {
			WriteError(w, http.StatusNotFound, "ماكو زبون بهذا الرقم — تأكد من الرقم أو سجّله كزبون أول")
			return
		}
		req.CustomerID = c.ID
	}
	if strings.TrimSpace(req.CustomerID) == "" {
		WriteError(w, http.StatusBadRequest, "الزبون مطلوب")
		return
	}
	saved, err := h.repo.Mark(uuid.NewString(), req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر تعليم الزبون كشخصية مهمة")
		return
	}
	WriteJSON(w, http.StatusCreated, saved)
}

// DELETE /api/vip-customers/{customerId} — إزالة التعليم، لمدير النظام فقط.
func (h *VipCustomerHandler) Unmark(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Unmark(r.PathValue("customerId")); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر إزالة التعليم")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
