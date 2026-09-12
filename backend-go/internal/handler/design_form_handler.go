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

func (h *DesignFormHandler) ListForms(w http.ResponseWriter, r *http.Request) {
	forms, err := h.service.ListForms()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الاستمارات")
		return
	}
	WriteJSON(w, http.StatusOK, forms)
}

func (h *DesignFormHandler) CreateForm(w http.ResponseWriter, r *http.Request) {
	var req model.CreateDesignFormRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الاستمارة غير صحيحة")
		return
	}
	form, err := h.service.CreateForm(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, form)
}

func (h *DesignFormHandler) DeleteForm(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteForm(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر حذف الاستمارة")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *DesignFormHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.PathValue("formId"))
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
	req.FormID = r.PathValue("formId")
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

// Public: الوصول عبر الرابط العام (publicToken) بدون تسجيل دخول — نرجّع بس
// اسم الاستمارة وأسئلتها، بدون أي معلومة ثانية عن النظام.
func (h *DesignFormHandler) PublicGet(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	form, err := h.service.GetFormByToken(token)
	if err != nil {
		WriteError(w, http.StatusNotFound, "الاستمارة غير موجودة")
		return
	}
	questions, err := h.service.List(form.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الاستمارة")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{
		"name":      form.Name,
		"questions": questions,
	})
}

func (h *DesignFormHandler) PublicSubmit(w http.ResponseWriter, r *http.Request) {
	var req model.SubmitDesignFormRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	sub, err := h.service.Submit(r.PathValue("token"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, sub)
}

func (h *DesignFormHandler) ListSubmissions(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListSubmissions(r.PathValue("formId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الأجوبة")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}
