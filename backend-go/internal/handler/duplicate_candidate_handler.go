package handler

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

type DuplicateCandidateHandler struct {
	service   *service.DuplicateCandidateService
	repo      *repository.DuplicateCandidateRepository
	deletes   *repository.BookingDeleteRequestRepository
	notif     *repository.NotificationRepository
	employees *repository.EmployeeRepository
}

func NewDuplicateCandidateHandler(s *service.DuplicateCandidateService) *DuplicateCandidateHandler {
	return &DuplicateCandidateHandler{service: s}
}

// GET /api/duplicate-candidates?kind=&status=
func (h *DuplicateCandidateHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rows, err := h.service.List(q.Get("kind"), q.Get("status"))
	if err != nil {
		log.Printf("list duplicate candidates: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قائمة التكرار")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/duplicate-candidates/{id}/dismiss
func (h *DuplicateCandidateHandler) Dismiss(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Dismiss(r.PathValue("id"), middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ═══ حلول التكرار — (ع): «أريد موظف يشوف التكرارات وياخذ حلول» ═══

// SetResolvers يربط طلبات الحذف والإشعارات (من main).
func (h *DuplicateCandidateHandler) SetResolvers(repo *repository.DuplicateCandidateRepository,
	deletes *repository.BookingDeleteRequestRepository, notif *repository.NotificationRepository, employees *repository.EmployeeRepository) {
	h.repo, h.deletes, h.notif, h.employees = repo, deletes, notif, employees
}

var mergeTableLabels = map[string]string{
	"Booking": "حجوزات", "Complaint": "شكاوى", "QualityFollowUp": "متابعات جودة", "SolarInstallation": "منظومات شمسية",
	"DeviceMaintenanceTicket": "تذاكر صيانة", "GpsDeviceRequest": "أجهزة جي بي اس", "GpsRenewalRequest": "طلبات تجديد",
	"GpsRenewalFollowUp": "متابعات تجديد", "GpsMaintenanceRequest": "طلبات صيانة جي بي اس", "SimCard": "شرائح",
	"CustomerServiceTag": "علامات خدمة",
}

func labelCounts[T int | int64](m map[string]T) []map[string]any {
	out := []map[string]any{}
	for t, n := range m {
		l := mergeTableLabels[t]
		if l == "" {
			l = t
		}
		out = append(out, map[string]any{"label": l, "count": n})
	}
	return out
}

// pairCustomers يتأكد keepId واحد من الزوج ويرجّع (الأصلي، المكرر).
func (h *DuplicateCandidateHandler) pairCustomers(id, keepID string) (*model.DuplicateCandidate, string, error) {
	c, err := h.repo.Get(id)
	if err != nil || c.Kind != "CUSTOMER" {
		return nil, "", fmt.Errorf("الزوج مو موجود أو مو زبائن")
	}
	if c.Status != "PENDING" {
		return nil, "", fmt.Errorf("هالزوج انحسم من قبل")
	}
	switch keepID {
	case c.EntityAID:
		return c, c.EntityBID, nil
	case c.EntityBID:
		return c, c.EntityAID, nil
	}
	return nil, "", fmt.Errorf("اختار الزبون الأصلي من الزوج نفسه")
}

// GET /api/duplicate-candidates/{id}/merge-preview?keep=
func (h *DuplicateCandidateHandler) MergePreview(w http.ResponseWriter, r *http.Request) {
	_, drop, err := h.pairCustomers(r.PathValue("id"), r.URL.Query().Get("keep"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"moves": labelCounts(h.repo.MergeCounts(drop))})
}

// POST /api/duplicate-candidates/{id}/merge {keepId}
func (h *DuplicateCandidateHandler) Merge(w http.ResponseWriter, r *http.Request) {
	var b struct {
		KeepID string `json:"keepId"`
	}
	if err := DecodeJSON(r, &b); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	c, drop, err := h.pairCustomers(r.PathValue("id"), b.KeepID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	keepB, dropB := c.CustomerA, c.CustomerB
	if b.KeepID == c.EntityBID {
		keepB, dropB = c.CustomerB, c.CustomerA
	}
	moved, err := h.repo.MergeCustomers(b.KeepID, drop)
	if err != nil {
		log.Printf("merge customers %s→%s: %v", drop, b.KeepID, err)
		WriteError(w, http.StatusInternalServerError, "تعذّر الدمج — ما تغيّر شي")
		return
	}
	me := middleware.EmployeeIDFromContext(r)
	code := func(x *model.DuplicateCustomerBrief) string {
		if x == nil {
			return "?"
		}
		return fmt.Sprintf("CUST-%05d", x.CustomerCode)
	}
	parts := []string{}
	for _, m := range labelCounts(moved) {
		parts = append(parts, fmt.Sprintf("%v %v", m["count"], m["label"]))
	}
	what := "ماكو شي ينتقل"
	if len(parts) > 0 {
		what = strings.Join(parts, "، ")
	}
	dropDesc := code(dropB)
	if dropB != nil {
		dropDesc += " «" + dropB.Name + "» " + dropB.Phone
	}
	note := fmt.Sprintf("اندمج %s بـ%s — انتقل: %s.", dropDesc, code(keepB), what)
	_ = h.repo.Resolve(c.ID, me, "MERGED", note)
	if h.notif != nil {
		name := ""
		if e, err := h.employees.FindByID(me); err == nil && e != nil {
			name = e.Name
		}
		_ = h.notif.CreateForRolesOrPermission([]string{"OWNER", "ADMIN"}, "", "DUPLICATE",
			fmt.Sprintf("🔀 %s دمج زبونين بتدقيق التكرار: %s", name, note))
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "note": note})
}

// POST /api/duplicate-candidates/{id}/request-delete {bookingId}
// طلب حذف عادي ينتظر موافقة المراقب/المدير — ما ينحذف شي بيد موظف واحد.
func (h *DuplicateCandidateHandler) RequestDelete(w http.ResponseWriter, r *http.Request) {
	var b struct {
		BookingID string `json:"bookingId"`
	}
	if err := DecodeJSON(r, &b); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	c, err := h.repo.Get(r.PathValue("id"))
	if err != nil || c.Kind != "BOOKING" || c.Status != "PENDING" {
		WriteError(w, http.StatusBadRequest, "الزوج مو موجود أو انحسم من قبل")
		return
	}
	other := c.BookingA
	if b.BookingID == c.EntityAID {
		other = c.BookingB
	} else if b.BookingID != c.EntityBID {
		WriteError(w, http.StatusBadRequest, "اختار الحجز المكرر من الزوج نفسه")
		return
	}
	otherCode := "?"
	if other != nil {
		otherCode = other.Code
	}
	me := middleware.EmployeeIDFromContext(r)
	reason := "تكرار مع " + otherCode + " (تدقيق التكرار)"
	out, err := h.deletes.Create(b.BookingID, me, reason, model.BookingDeleteChannelCallCenter, model.BookingDeleteTypeRecurringDuplicate)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	_ = h.repo.Resolve(c.ID, me, "DELETE_REQUESTED", "انطلب حذف "+out.BookingCode+" — "+reason+". ينتظر موافقة المراقب أو المدير.")
	if h.notif != nil {
		msg := "🗑️ طلب حذف الحجز " + out.BookingCode + " من " + out.RequestedByName + " — السبب: " + reason
		_ = h.notif.CreateForRole("MONITOR", "booking_delete_request", msg)
		_ = h.notif.CreateForRole("ADMIN", "booking_delete_request", msg)
		_ = h.notif.CreateForRole("OWNER", "booking_delete_request", msg)
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true, "bookingCode": out.BookingCode})
}

// GET /api/duplicate-candidates/report — ملخص ماتركس: كم معلّق، ومنو ينتج التكرار.
func (h *DuplicateCandidateHandler) Report(w http.ResponseWriter, r *http.Request) {
	rep, err := h.repo.Report()
	if err != nil {
		log.Printf("duplicate report: %v", err)
	}
	WriteJSON(w, http.StatusOK, rep)
}
