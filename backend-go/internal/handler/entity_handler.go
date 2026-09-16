package handler

import (
	"log"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// EntityHandler الكيان — تقرير الموظف الحي وتوليد شخصيته.
type EntityHandler struct {
	service *service.EntityService
}

func NewEntityHandler(s *service.EntityService) *EntityHandler {
	return &EntityHandler{service: s}
}

// GET /api/entity/briefing — تقرير الكيان لصاحب التوكن.
//
// ⚠️ ماكو معامل `employeeId` بقصد: الهوية تجي من التوكن وحده. لو
// انقبلت من الطلب، أي موظف يقدر يقرا غرامات زميله ورصيده بتبديل
// رقم بالرابط.
func (h *EntityHandler) Briefing(w http.ResponseWriter, r *http.Request) {
	b, err := h.service.Briefing(middleware.EmployeeIDFromContext(r))
	if err != nil {
		log.Printf("entity briefing: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب تقرير الكيان")
		return
	}
	WriteJSON(w, http.StatusOK, b)
}

// GET /api/entity/character/me — شخصيتي أنا (أو فراغ لو ما انولدت).
func (h *EntityHandler) MyCharacter(w http.ResponseWriter, r *http.Request) {
	ch, err := h.service.Character(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الشخصية")
		return
	}
	WriteJSON(w, http.StatusOK, ch)
}

// POST /api/entity/character/{employeeId}/generate — توليد/إعادة توليد.
//
// ⚠️ التوليد ينادي مولّد صور خارجي ثلاث مرات، ويمر بحارس المالك/المدير
// بالمسار — مو متروك لكل موظف: سقف المزوّد اليومي محدود، وأي موظف
// يقدر يحرقه بضغطات متكررة.
func (h *EntityHandler) GenerateCharacter(w http.ResponseWriter, r *http.Request) {
	ch, err := h.service.GenerateCharacter(r.PathValue("employeeId"), middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, ch)
}
