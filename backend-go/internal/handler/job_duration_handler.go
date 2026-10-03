package handler

import (
	"net/http"
	"strconv"

	"staffmange-api/internal/service"
)

// JobDurationHandler يخدم تقدير مدة العمل المتعلَّم (learned baseline) — للمنسق
// حتى يشوف تقدير واقعي قبل تثبيت موعد/فريق، أو رسالة واضحة أن البيانات غير كافية بعد.
type JobDurationHandler struct {
	estimator *service.JobDurationEstimatorService
}

func NewJobDurationHandler(estimator *service.JobDurationEstimatorService) *JobDurationHandler {
	return &JobDurationHandler{estimator: estimator}
}

// GET /api/job-duration-estimate?systemName=...&jobType=...&itemCount=...&crewSize=...
func (h *JobDurationHandler) Estimate(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	systemName := q.Get("systemName")
	jobType := q.Get("jobType")
	itemCount, _ := strconv.Atoi(q.Get("itemCount"))
	crewSize, _ := strconv.Atoi(q.Get("crewSize"))

	if systemName == "" || jobType == "" {
		WriteError(w, http.StatusBadRequest, "systemName و jobType مطلوبان")
		return
	}

	expected, sampleCount, err := h.estimator.EstimateExpectedMinutes(systemName, jobType, itemCount, crewSize)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حساب التقدير")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"expectedMinutes": expected, // null لو البيانات غير كافية بعد
		"sampleCount":     sampleCount,
		"minSamples":      service.MinSamplesForEstimate,
	})
}
