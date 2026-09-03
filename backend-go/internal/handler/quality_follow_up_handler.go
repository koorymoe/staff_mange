package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

type QualityFollowUpHandler struct {
	service     *service.QualityFollowUpService
	permissions *repository.PermissionRepository
}

func NewQualityFollowUpHandler(s *service.QualityFollowUpService, p *repository.PermissionRepository) *QualityFollowUpHandler {
	return &QualityFollowUpHandler{service: s, permissions: p}
}

// GET /api/quality-follow-ups

// blockMonitorWrite يمنع المراقب المدقق (أو مو موظف وصل الشاشة
// بصلاحية monitoring/auditing) من مسارات الكتابة.
//
// ⚠️⚠️ المراقب ياخذ صلاحية `quality_control` افتراضياً بدوره، وكل
// مسارات الجودة عليها نفس الحارس — فچان **يقدر يتصل بالزبون
// ويسجّل الحكم والكشف** مثل مهندس الجودة بالضبط.
//
// وصاحب النظام قرر: المراقب **يشوف كلشي ويدقّق شغل المهندس، بس
// ما يتصل بالزبون**. لأنه لو اتصل وحكم، يصير يدقّق شغلاً سوّاه
// هو — نفس الخلط الي شلناه بالتدقيق اليومي وبفواتير الليدر.
//
// ⚠️⚠️ ونفس القيد لازم ينطبق على موظف وصل الشاشة بمنح صلاحية
// `monitoring`/`auditing` فردياً (مثل ليدر انمنحها ليوم واحد) —
// وإلا صار أوسع صلاحية من المراقب الحقيقي نفسه.
//
// ⚠️ ورفض صريح مو تسجيل مخالفة: الأزرار چانت معروضة إله، فضغطه
// عليها مو محاولة تجاوز.
func (h *QualityFollowUpHandler) blockMonitorWrite(w http.ResponseWriter, r *http.Request) bool {
	role := middleware.RoleFromContext(r)
	if role == "ADMIN" || role == "OWNER" || role == "QUALITY_ENGINEER" {
		return false
	}
	blocked := role == "MONITOR"
	if !blocked && h.permissions != nil {
		if perms, err := h.permissions.ListForEmployee(middleware.EmployeeIDFromContext(r)); err == nil {
			for _, p := range perms {
				if p.Name == "monitoring" || p.Name == "auditing" {
					blocked = true
					break
				}
			}
		}
	}
	if blocked {
		WriteError(w, http.StatusForbidden,
			"المراقب يشوف ويدقّق — التواصل مع الزبون وتسجيل الحكم لمهندس الجودة")
		return true
	}
	return false
}

func (h *QualityFollowUpHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب متابعات الجودة")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

// PUT /api/quality-follow-ups/{id}
func (h *QualityFollowUpHandler) Update(w http.ResponseWriter, r *http.Request) {
	if h.blockMonitorWrite(w, r) {
		return
	}
	var req model.UpdateQualityFollowUpRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	item, err := h.service.Update(r.PathValue("id"), middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

// POST /api/quality-follow-ups/{id}/verdict — حكم مهندس الجودة.
//
// تقرير إيجابي: ينسجّل وبس.
// تقرير سلبي: تنخصم نقطة «شكوى الزبائن» من الليدر فوراً — إلا إذا
// طلب كشف، وقتها الغرامة تنتظر النتيجة (الزبون ممكن يكون يجذب).
func (h *QualityFollowUpHandler) Verdict(w http.ResponseWriter, r *http.Request) {
	if h.blockMonitorWrite(w, r) {
		return
	}
	var req model.QualityVerdictRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	if req.ReportType != "POSITIVE" && req.ReportType != "NEGATIVE" {
		WriteError(w, http.StatusBadRequest, "لازم تحدد التقرير: إيجابي أو سلبي")
		return
	}
	// التقرير السلبي يخصم من موظف فعلياً — ما يصير ينزل بلا سبب مكتوب.
	// الإيجابي ما يحتاج، لأنه ما يترتب عليه شي.
	if req.ReportType == "NEGATIVE" && TextLen(req.Notes) < 5 {
		WriteError(w, http.StatusBadRequest, "اكتب شنو كالك الزبون — التقرير السلبي يخصم نقطة من الليدر")
		return
	}
	items, err := h.service.Verdict(r.PathValue("id"), middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

// POST /api/quality-follow-ups/{id}/inspect — نتيجة الكشف الميداني.
func (h *QualityFollowUpHandler) Inspect(w http.ResponseWriter, r *http.Request) {
	if h.blockMonitorWrite(w, r) {
		return
	}
	var req model.QualityInspectionRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	if req.Result != "CUSTOMER_RIGHT" && req.Result != "CUSTOMER_WRONG" {
		WriteError(w, http.StatusBadRequest, "لازم تحدد نتيجة الكشف")
		return
	}
	if TextLen(req.Notes) < 5 {
		WriteError(w, http.StatusBadRequest, "اكتب شنو شفت بالكشف — هذا الي يعتمد عليه القرار")
		return
	}
	items, err := h.service.Inspect(r.PathValue("id"), middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, items)
}
