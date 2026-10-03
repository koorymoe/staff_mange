package handler

import (
	"errors"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// مركز قيادة ماتركس — كلها ADMIN/OWNER.
type MatrixCommandHandler struct{ svc *service.MatrixCommandService }

func NewMatrixCommandHandler(s *service.MatrixCommandService) *MatrixCommandHandler {
	return &MatrixCommandHandler{svc: s}
}

// GET /api/ai/feed
func (h *MatrixCommandHandler) Feed(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.Feed()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب البث")
		return
	}
	WriteJSON(w, http.StatusOK, v)
}

// GET /api/ai/daily-trend
func (h *MatrixCommandHandler) Trend(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.Trend()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حساب الاتجاه")
		return
	}
	WriteJSON(w, http.StatusOK, v)
}

// GET /api/ai/late-focus
func (h *MatrixCommandHandler) LateFocus(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.LateFocus()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حساب التأخير")
		return
	}
	WriteJSON(w, http.StatusOK, v)
}

// POST /api/ai/ask {question}
func (h *MatrixCommandHandler) Ask(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Question string `json:"question"`
	}
	if err := DecodeJSON(r, &body); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	v, err := h.svc.Ask(middleware.EmployeeIDFromContext(r), body.Question)
	if errors.Is(err, service.ErrAskLimit) {
		WriteError(w, http.StatusTooManyRequests, err.Error())
		return
	}
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, v)
}
