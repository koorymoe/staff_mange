package handler

import (
	"net/http"

	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
	"staffmange-api/internal/storage"
)

// VoiceTaskHandler مهمة بالفويس: المدير يسجّل، ويرجعله مسودة مهمة يعدّلها
// ويوجّهها بنفسه من نفس شاشة المهام الإضافية. الصوت ما ينحفظ.
type VoiceTaskHandler struct {
	voice     *service.VoiceAnalysisService
	employees *repository.EmployeeRepository
}

func NewVoiceTaskHandler(v *service.VoiceAnalysisService, e *repository.EmployeeRepository) *VoiceTaskHandler {
	return &VoiceTaskHandler{voice: v, employees: e}
}

// POST /api/extra-tasks/voice-draft — حقل «file».
func (h *VoiceTaskHandler) Draft(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxVoiceBytes+64<<10)
	if err := r.ParseMultipartForm(maxVoiceBytes); err != nil {
		WriteError(w, http.StatusBadRequest, "الفويس أطول من المسموح (٣ دقايق)")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "ماكو فويس بالطلب")
		return
	}
	defer file.Close()
	data, err := storage.ReadLimited(file, maxVoiceBytes)
	if err != nil {
		WriteError(w, http.StatusRequestEntityTooLarge, err.Error())
		return
	}
	ct := sniffAudio(data)
	if ct == "" {
		WriteError(w, http.StatusBadRequest, "صيغة الصوت مو مدعومة")
		return
	}
	list, err := h.employees.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الموظفين")
		return
	}
	emps := make([]service.VoiceTaskEmployee, 0, len(list))
	for _, e := range list {
		emps = append(emps, service.VoiceTaskEmployee{ID: e.ID, Name: e.Name})
	}
	d, err := h.voice.TaskDraft(r.Context(), data, ct, emps)
	if err != nil {
		WriteError(w, http.StatusBadGateway, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, d)
}
