package handler

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// ═══ 🧠 تحليل ماتركس للتقرير — القراءة ═══
// المراقب: نفس صلاحية تقارير الموظفين. الموظف: تقريره هو بس (غيره = 404).

type reportAnalysisOut struct {
	Enabled  bool                         `json:"enabled"`
	Analysis *service.DailyReportAnalysis `json:"analysis"`
}

func (h *WorkGateHandler) writeAnalysis(w http.ResponseWriter, reportID string) {
	a, err := service.GetReportAnalysis(h.db, reportID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("[report-analysis] get %s: %v", reportID, err)
	}
	WriteJSON(w, http.StatusOK, reportAnalysisOut{Enabled: service.ReportAnalysisEnabled(), Analysis: a})
}

// GET /api/daily-reports/{id}/analysis
func (h *WorkGateHandler) ReportAnalysis(w http.ResponseWriter, r *http.Request) {
	var ok bool
	if err := h.db.Get(&ok, `SELECT true FROM "DailyWorkReport" WHERE id = $1`, r.PathValue("id")); err != nil {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	h.writeAnalysis(w, r.PathValue("id"))
}

// GET /api/me/daily-reports/{id}/analysis — تقريري أنا بس
func (h *WorkGateHandler) MyReportAnalysis(w http.ResponseWriter, r *http.Request) {
	var ok bool
	if err := h.db.Get(&ok, `SELECT true FROM "DailyWorkReport" WHERE id = $1 AND "employeeId" = $2`,
		r.PathValue("id"), middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusNotFound, "مو موجود")
		return
	}
	h.writeAnalysis(w, r.PathValue("id"))
}
