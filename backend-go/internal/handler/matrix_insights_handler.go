package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

// MatrixInsightsHandler مسارات ماتركس (كلها GET — قراءة واقتراح بس).
type MatrixInsightsHandler struct {
	insights *service.MatrixInsightsService
	weekly   *service.WeeklyReportService
	perms    *repository.PermissionRepository
}

func NewMatrixInsightsHandler(insights *service.MatrixInsightsService, weekly *service.WeeklyReportService, perms *repository.PermissionRepository) *MatrixInsightsHandler {
	return &MatrixInsightsHandler{insights: insights, weekly: weekly, perms: perms}
}

// GET /api/ai/crew-recommendation?bookingId=
func (h *MatrixInsightsHandler) CrewRecommendation(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("bookingId"))
	if id == "" {
		WriteError(w, http.StatusBadRequest, "bookingId مطلوب")
		return
	}
	out, err := h.insights.CrewRecommendation(id)
	if errors.Is(err, sql.ErrNoRows) {
		WriteError(w, http.StatusNotFound, "الحجز مو موجود")
		return
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حساب توصية الكادر")
		return
	}
	WriteJSON(w, http.StatusOK, out)
}

// GET /api/ai/new-employee-curves
func (h *MatrixInsightsHandler) NewEmployeeCurves(w http.ResponseWriter, r *http.Request) {
	out, err := h.insights.NewEmployeeCurves()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حساب منحنى الموظفين الجدد")
		return
	}
	WriteJSON(w, http.StatusOK, out)
}

// GET /api/ai/stock-forecast
func (h *MatrixInsightsHandler) StockForecast(w http.ResponseWriter, r *http.Request) {
	out, err := h.insights.StockForecast()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حساب استهلاك المواد")
		return
	}
	WriteJSON(w, http.StatusOK, out)
}

// GET /api/ai/replacement-suggestions — الحارس يسمح لـADMIN/OWNER أو
// it_assets أو vehicle_management، وهنا نفلتر: كل واحد يشوف قسمه بس.
func (h *MatrixInsightsHandler) ReplacementSuggestions(w http.ResponseWriter, r *http.Request) {
	role, _ := r.Context().Value(middleware.ContextRole).(string)
	it, veh := true, true
	if role != "ADMIN" && role != "OWNER" {
		empID, _ := r.Context().Value(middleware.ContextEmployeeID).(string)
		perms, err := h.perms.ListForEmployee(empID)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "تعذر قراءة الصلاحيات")
			return
		}
		it, veh = false, false
		for _, p := range perms {
			switch p.Name {
			case "it_assets":
				it = true
			case "vehicle_management":
				veh = true
			}
		}
	}
	out, err := h.insights.Replacements(it, veh)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حساب اقتراحات الاستبدال")
		return
	}
	WriteJSON(w, http.StatusOK, out)
}

// GET /api/ai/weekly-report?week=YYYY-Www
func (h *MatrixInsightsHandler) WeeklyReport(w http.ResponseWriter, r *http.Request) {
	out, err := h.weekly.Report(strings.TrimSpace(r.URL.Query().Get("week")))
	if err != nil {
		if strings.Contains(err.Error(), "أسبوع") {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, "تعذر حساب التقرير الأسبوعي")
		return
	}
	WriteJSON(w, http.StatusOK, out)
}
