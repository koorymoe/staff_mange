package handler

import (
	"encoding/json"
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
	// permissions: فحص صلاحية نوع فاتورة الخدمة (جي بي اس/داش كام).
	// اختياري — بدونه المالك والمدير بس يفوترون.
	permissions *repository.PermissionRepository
}

func NewLeaderInvoiceHandler(
	s *service.LeaderInvoiceService,
	catalog *repository.SystemPriceCatalogRepository,
	materials *repository.MaterialRepository,
) *LeaderInvoiceHandler {
	return &LeaderInvoiceHandler{service: s, catalog: catalog, materials: materials}
}

// SetPermissions يربط مستودع الصلاحيات بعد البناء.
func (h *LeaderInvoiceHandler) SetPermissions(p *repository.PermissionRepository) { h.permissions = p }

// GET /api/leader-invoices?employeeId=
func (h *LeaderInvoiceHandler) List(w http.ResponseWriter, r *http.Request) {
	// الليدر يشوف فواتيره هو بس. قبل هيچي كان employeeId ينجي من
	// الرابط — يعني أي ليدر يشيله ويشوف فواتير كل الليدرات، أو
	// يحطّ رقم زميله ويشوف مالته.
	employeeID := r.URL.Query().Get("employeeId")
	if !canReviewInvoices(r) {
		employeeID = middleware.EmployeeIDFromContext(r)
	}
	// ⚠️ المرحلة معامل اختياري: بلاها ترجع الكل — فأي نداء قديم يشتغل
	// مثل ما چان بالضبط.
	invoices, err := h.service.ListByStage(employeeID, r.URL.Query().Get("stage"))
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

// POST /api/leader-invoices/service — فاتورة جي بي اس أو داش كام
// بسعر يكتبه مسؤول الخدمة.
//
// ⚠️⚠️ **الصلاحية تنفحص حسب النوع هنا، مو بالمسار**: حارس المسار
// يگدر يفحص «وحدة من الاثنتين» بس، ولو وقفنا عنده يصير مسؤول
// الجي بي اس يفوتر داش كام. النوع يجي بالطلب، فالفحص لازم يكون
// بعد ما ينقرا.
func (h *LeaderInvoiceHandler) CreateServiceInvoice(w http.ResponseWriter, r *http.Request) {
	var req model.CreateServiceInvoiceRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	perm, ok := model.ServiceInvoicePermission[req.Kind]
	if !ok {
		WriteError(w, http.StatusBadRequest, "نوع الفاتورة لازم يكون جي بي اس أو داش كام")
		return
	}
	employeeID := middleware.EmployeeIDFromContext(r)
	role := middleware.RoleFromContext(r)
	allowed := role == "ADMIN" || role == "OWNER"
	if !allowed && h.permissions != nil {
		if has, err := h.permissions.HasPermission(employeeID, perm); err == nil && has {
			allowed = true
		}
	}
	if !allowed {
		WriteError(w, http.StatusForbidden, "ما عندك صلاحية «"+model.ServiceInvoiceKindLabel[req.Kind]+"»")
		return
	}
	inv, err := h.service.CreateServiceInvoice(employeeID, req)
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
	var req model.ApproveLeaderInvoiceRequest
	// الجسم اختياري بالشكل بس الرقم إجباري بالمنطق — الخدمة ترفض الفاضي
	_ = json.NewDecoder(r.Body).Decode(&req)
	inv, err := h.service.Approve(r.PathValue("id"), approverID, req.ExternalInvoiceNumber)
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

// GET /api/leader-invoices/by-number?number= — يلكه الفاتورة برقم
// المحاسب الخارجي. هذا سبب أرشفة الرقم أصلاً.
func (h *LeaderInvoiceHandler) FindByExternalNumber(w http.ResponseWriter, r *http.Request) {
	inv, err := h.service.FindByExternalNumber(r.URL.Query().Get("number"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if inv == nil {
		WriteError(w, http.StatusNotFound, "ماكو فاتورة بهذا الرقم")
		return
	}
	WriteJSON(w, http.StatusOK, inv)
}

// PUT /api/leader-invoices/{id}/external-number — ربط رقم فاتورة
// محاسبية بفاتورة معتمدة أصلاً (الفواتير القديمة).
func (h *LeaderInvoiceHandler) SetExternalNumber(w http.ResponseWriter, r *http.Request) {
	var req model.SetExternalNumberRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	inv, err := h.service.SetExternalNumber(r.PathValue("id"), req.ExternalInvoiceNumber)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, inv)
}

// PUT /api/leader-invoices/{id}/adjust — تعديل المحاسب على المبالغ.
func (h *LeaderInvoiceHandler) Adjust(w http.ResponseWriter, r *http.Request) {
	var req model.AdjustLeaderInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	inv, err := h.service.AdjustAmounts(r.PathValue("id"), req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, inv)
}

// GET /api/free-work-reasons — قائمة أسباب الشغل المجاني.
func (h *LeaderInvoiceHandler) FreeReasons(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.FreeReasons()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب أسباب المجانية")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}


// GET /api/leader-invoices/{id}/adjustments — سجل تعديلات المحاسب.
//
// «شنو كان المبلغ وشنو صار ومنو غيّره وليش». قبل، هاي المعلومة ما
// كانت موجودة أصلاً — التعديل يمحي الأصل.
func (h *LeaderInvoiceHandler) Adjustments(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.Adjustments(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// ═══ التدقيق ═══

// PUT /api/leader-invoices/{id}/audit — حكم المحاسب قبل الاعتماد.
func (h *LeaderInvoiceHandler) SetAuditVerdict(w http.ResponseWriter, r *http.Request) {
	var req model.AuditVerdictRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	inv, err := h.service.SetAuditVerdict(r.PathValue("id"), req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, inv)
}

// PUT /api/leader-invoices/{id}/revoke — سحب اعتماد انصار بالغلط.
func (h *LeaderInvoiceHandler) RevokeApproval(w http.ResponseWriter, r *http.Request) {
	var req model.RevokeApprovalRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	inv, err := h.service.RevokeApproval(r.PathValue("id"), req.Reason, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, inv)
}

// PUT /api/leader-invoices/{id}/monitor-request — المحاسب يرسلها للمراقب.
func (h *LeaderInvoiceHandler) RequestMonitorReview(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Note string `json:"note"`
	}
	_ = DecodeJSON(r, &req)
	inv, err := h.service.RequestMonitorReview(r.PathValue("id"), middleware.EmployeeIDFromContext(r), req.Note)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, inv)
}

// PUT /api/leader-invoices/{id}/monitor-decide — المراقب يبتّ.
func (h *LeaderInvoiceHandler) DecideMonitorReview(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Verdict string `json:"verdict"`
		Note    string `json:"note"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	inv, err := h.service.DecideMonitorReview(r.PathValue("id"), req.Verdict, req.Note, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, inv)
}

// PUT /api/leader-invoices/{id}/return — المالك يرجّعها للمحاسب.
func (h *LeaderInvoiceHandler) ReturnToAccountant(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Reason string `json:"reason"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	inv, err := h.service.ReturnToAccountant(r.PathValue("id"), middleware.EmployeeIDFromContext(r), req.Reason)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, inv)
}

// GET /api/leader-invoices/approved-without-number — الفجوة الي كلّفت.
func (h *LeaderInvoiceHandler) ApprovedWithoutNumber(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.ListApprovedWithoutNumber()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الجلب")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
