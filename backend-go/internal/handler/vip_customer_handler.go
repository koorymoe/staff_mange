package handler

import (
	"log"
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
		phone := strings.TrimSpace(req.Phone)
		c, err := h.customers.Lookup(phone)
		if err == nil && c != nil {
			req.CustomerID = c.ID
		} else if strings.TrimSpace(req.Name) != "" {
			// شخصية مهمة مو زبون عدنا: ننشئ سجل جديد بمعلوماته حتى
			// تنحفظ وترجع تنلكه بالرقم بأي وقت.
			created, cerr := h.customers.FindOrCreate(model.CreateCustomerRequest{
				Name:         strings.TrimSpace(req.Name),
				Phone:        phone,
				Location:     req.Location,
				MapLatitude:  req.MapLatitude,
				MapLongitude: req.MapLongitude,
			})
			if cerr != nil || created == nil {
				WriteError(w, http.StatusBadRequest, "تعذر حفظ معلومات الشخصية")
				return
			}
			req.CustomerID = created.ID
			if req.LocationURL != nil && *req.LocationURL != "" {
				_, _ = h.customers.Update(created.ID, model.UpdateCustomerRequest{
					Name: created.Name, Phone: created.Phone, LocationURL: req.LocationURL,
				})
			}
		} else {
			WriteError(w, http.StatusNotFound, "ماكو زبون بهذا الرقم — اكتب اسمه حتى نضيفه كشخصية مهمة جديدة")
			return
		}
	}
	if strings.TrimSpace(req.CustomerID) == "" {
		WriteError(w, http.StatusBadRequest, "الزبون مطلوب")
		return
	}
	saved, err := h.repo.Mark(uuid.NewString(), req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		log.Printf("mark vip: %v", err)
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
