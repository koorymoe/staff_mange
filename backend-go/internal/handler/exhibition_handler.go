package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type ExhibitionHandler struct {
	service *service.ExhibitionService
}

func NewExhibitionHandler(s *service.ExhibitionService) *ExhibitionHandler {
	return &ExhibitionHandler{service: s}
}

func (h *ExhibitionHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المعارض")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *ExhibitionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateExhibitionRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات المعرض غير صحيحة")
		return
	}
	item, err := h.service.Create(req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, item)
}

func (h *ExhibitionHandler) Nominate(w http.ResponseWriter, r *http.Request) {
	var req model.NominateExhibitionRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	item, err := h.service.Nominate(r.PathValue("id"), req.EmployeeIDs)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر حفظ الترشيح")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *ExhibitionHandler) AddPhotos(w http.ResponseWriter, r *http.Request) {
	var req model.AddExhibitionPhotosRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	item, err := h.service.AddPhotos(r.PathValue("id"), req.PhotoUrls)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر حفظ الصور")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *ExhibitionHandler) SetFindings(w http.ResponseWriter, r *http.Request) {
	var req model.SetExhibitionFindingsRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	item, err := h.service.SetFindings(r.PathValue("id"), req.KeyFindings)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *ExhibitionHandler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GenerateVisitReport(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *ExhibitionHandler) Archive(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Archive(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الأرشفة")
		return
	}
	WriteJSON(w, http.StatusOK, item)
}
