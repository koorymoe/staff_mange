package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type QualityFollowUpHandler struct {
	service *service.QualityFollowUpService
}

func NewQualityFollowUpHandler(s *service.QualityFollowUpService) *QualityFollowUpHandler {
	return &QualityFollowUpHandler{service: s}
}

// GET /api/quality-follow-ups
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
