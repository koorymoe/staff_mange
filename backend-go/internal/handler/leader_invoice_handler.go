package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

// LeaderInvoiceHandler يخدم فوترة الليدر: عرض الكتالوج، البحث بمواد الأرشيف
// بالكود، وإنشاء/عرض فواتير الليدر.
type LeaderInvoiceHandler struct {
	service   *service.LeaderInvoiceService
	catalog   *repository.SystemPriceCatalogRepository
	materials *repository.MaterialRepository
}

func NewLeaderInvoiceHandler(
	s *service.LeaderInvoiceService,
	catalog *repository.SystemPriceCatalogRepository,
	materials *repository.MaterialRepository,
) *LeaderInvoiceHandler {
	return &LeaderInvoiceHandler{service: s, catalog: catalog, materials: materials}
}

// GET /api/leader-invoices?employeeId=
func (h *LeaderInvoiceHandler) List(w http.ResponseWriter, r *http.Request) {
	// الليدر يشوف فواتيره هو بس. قبل هيچي كان employeeId ينجي من
	// الرابط — يعني أي ليدر يشيله ويشوف فواتير كل الليدرات، أو
	// يحطّ رقم زميله ويشوف مالته.
	employeeID := r.URL.Query().Get("employeeId")
	if !canReviewInvoices(r) {
		employeeID = middleware.EmployeeIDFromContext(r)
	}
	invoices, err := h.service.List(employeeID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب فواتير الليدر")
		return
	}
	WriteJSON(w, http.StatusOK, invoices)
}

// GET /api/leader-invoices/{id}
func (h *LeaderInvoiceHandler) Get(w http.ResponseWriter, r *http.Request) {
	inv, err := h.service.Get(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	// نفس القاعدة: الليدر ما يفتح فاتورة غيره حتى لو عنده رقمها
	if !canReviewInvoices(r) && inv.EmployeeID != middleware.EmployeeIDFromContext(r) {
		WriteError(w, http.StatusForbidden, "هذي مو فاتورتك")
		return
	}
	WriteJSON(w, http.StatusOK, inv)
}

// canReviewInvoices منو يشوف فواتير كل الليدرات: المحاسب والمراقب
// ومدير النظام والمالك. الليدر نفسه يشوف مالته بس، وباقي الموظفين
// (مصمم، مبيعات...) ما إلهم شغل بهذي الشاشة أصلاً.
func canReviewInvoices(r *http.Request) bool {
	switch middleware.RoleFromContext(r) {
	case "ADMIN", "OWNER", "FINANCE", "MONITOR":
		return true
	}
	return false
}

// POST /api/leader-invoices
func (h *LeaderInvoiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateLeaderInvoiceRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	employeeID := middleware.EmployeeIDFromContext(r)
	inv, err := h.service.Create(employeeID, req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, inv)
}

// POST /api/leader-invoices/estimate — حساب سريع بدون حفظ ولا ربط بحجز.
func (h *LeaderInvoiceHandler) Estimate(w http.ResponseWriter, r *http.Request) {
	var req model.EstimateExecutionCostRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	res, err := h.service.Estimate(req.Items)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	// تفصيل الحساب (المعادلات والحدود الدنيا) أسعار داخلية — للمالك ومدير
	// النظام فقط. الليدر ياخذ المبلغ وبس. نشيلها من الرد نفسه مو من الشاشة
	// بس، وإلا تبقى موجودة بالـJSON لأي أحد يفتح أدوات المطوّر.
	if role := middleware.RoleFromContext(r); role != "ADMIN" && role != "OWNER" {
		res.Breakdown = nil
		res.SystemMinimums = nil
	}
	WriteJSON(w, http.StatusOK, res)
}

// POST /api/leader-invoices/camera-cost — استمارة "حساب تكلفة التنفيذ" الخاصة
// بمنظومة كاميرات المراقبة (شيت مستقل بمعادلة مختلفة عن تكاليف المشروع).
func (h *LeaderInvoiceHandler) CameraCost(w http.ResponseWriter, r *http.Request) {
	var req model.CameraCostRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	res, err := service.CalculateCameraCost(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, res)
}

// GET /api/leader-invoices/camera-cost/options — قوائم نوع المكان ونوع المنظومة
// وأسعار الأعمال الإضافية، حتى الواجهة ما تكتب القيم بنفسها وتختلف عن المحرك.
func (h *LeaderInvoiceHandler) CameraCostOptions(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]any{
		"placeTypes":  service.CameraPlaceTypes,
		"systemTypes": service.CameraSystemTypes,
		"note":        service.CameraCostNote,
	})
}

// PUT /api/leader-invoices/{id}/approve — محصور بـrequireFinance بالراوت.
func (h *LeaderInvoiceHandler) Approve(w http.ResponseWriter, r *http.Request) {
	approverID := middleware.EmployeeIDFromContext(r)
	inv, err := h.service.Approve(r.PathValue("id"), approverID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, inv)
}

// GET /api/system-price-catalog?systemName=
func (h *LeaderInvoiceHandler) ListCatalog(w http.ResponseWriter, r *http.Request) {
	systemName := r.URL.Query().Get("systemName")
	rows, err := h.catalog.List(systemName)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب كتالوج الأسعار")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/materials?code=
func (h *LeaderInvoiceHandler) ListMaterials(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	rows, err := h.materials.List(code)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المواد")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
