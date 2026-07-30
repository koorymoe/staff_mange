package handler

import (
	"net/http"

	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type DesignFormHandler struct {
	service *service.DesignFormService
}

func NewDesignFormHandler(s *service.DesignFormService) *DesignFormHandler {
	return &DesignFormHandler{service: s}
}

func (h *DesignFormHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الأسئلة")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *DesignFormHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateDesignFormQuestionRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات السؤال غير صحيحة")
		return
	}
	item, err := h.service.Create(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, item)
}

func (h *DesignFormHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateDesignFormQuestionRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	item, err := h.service.Update(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *DesignFormHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر حذف السؤال")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *DesignFormHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	var req model.ReorderDesignFormQuestionsRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if err := h.service.Reorder(req.QuestionIDs); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر إعادة الترتيب")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
