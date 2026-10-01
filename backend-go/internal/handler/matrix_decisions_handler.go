package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// MatrixDecisionsHandler صندوق قرارات ماتركس — للمالك ومدير النظام بس.
type MatrixDecisionsHandler struct {
	svc *service.MatrixAutopilotService
}

func NewMatrixDecisionsHandler(svc *service.MatrixAutopilotService) *MatrixDecisionsHandler {
	return &MatrixDecisionsHandler{svc: svc}
}

// GET /api/ai/decisions?day=YYYY-MM-DD
func (h *MatrixDecisionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	box, err := h.svc.Decisions(r.URL.Query().Get("day"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب صندوق القرارات")
		return
	}
	WriteJSON(w, http.StatusOK, box)
}

// POST /api/ai/actions/{id}/undo — «لا تسوي هذا»: يعلّم الفعل مرفوض،
// وماتركس ما يعيده على نفس الشي ٣٠ يوم.
func (h *MatrixDecisionsHandler) Undo(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Undo(r.PathValue("id"), middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/ai/actions/kinds/{kind}/resume — يرجّع نوع وقّفه ماتركس.
func (h *MatrixDecisionsHandler) Resume(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Resume(r.PathValue("kind")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/ai/my-watch — حالة عين ماتركس للموظف نفسه (كل موظف يشوف حالته بس).
func (h *MatrixDecisionsHandler) MyWatch(w http.ResponseWriter, r *http.Request) {
	st, err := h.svc.WatchState(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب حالة العين")
		return
	}
	WriteJSON(w, http.StatusOK, st)
}
