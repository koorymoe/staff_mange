package handler

import (
	"net/http"
	"strconv"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
	"staffmange-api/internal/storage"
)

// AchievementHandler الإنجازات — تقرير يومي حر من أي موظف بأي دور.
//
// ⚠️ `Create` بلا حراسة دور (أي موظف مسجّل دخول)؛ `List`/`Review`
// محميين بـ`requireMonitor` بمسارات `main.go` — المدير والمالك والمراقب (ع) 10-23.
type AchievementHandler struct {
	service  *service.AchievementService
	voices   *repository.AchievementVoiceRepository
	store    storage.Store
	analyzer *service.VoiceAnalysisService
	notif    interface {
		Create(employeeID, kind, message string) error
	}
}

// SetNotifier الموظف ينبعثله إشعار بالرد على إنجازه.
func (h *AchievementHandler) SetNotifier(n interface {
	Create(employeeID, kind, message string) error
}) {
	h.notif = n
}

// reviewerIsMonitor المراقب (مو المدير/المالك) — ما يراجع إنجازات الإدارة ولا إنجازه.
func reviewerIsMonitor(r *http.Request) bool { return middleware.RoleFromContext(r) != "ADMIN" }

func isManagementRole(role string) bool { return role == "ADMIN" || role == "OWNER" }

func NewAchievementHandler(s *service.AchievementService) *AchievementHandler {
	return &AchievementHandler{service: s}
}

func (h *AchievementHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAchievementRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	a, err := h.service.Create(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, a)
}

// GET /api/achievements?employeeId=&day=&limit=
func (h *AchievementHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	rows, err := h.service.List(q.Get("employeeId"), q.Get("day"), limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الإنجازات")
		return
	}
	if reviewerIsMonitor(r) {
		me := middleware.EmployeeIDFromContext(r)
		kept := rows[:0]
		for _, a := range rows {
			if !isManagementRole(a.Role) && a.EmployeeID != me {
				kept = append(kept, a)
			}
		}
		rows = kept
	}
	h.attachVoices(rows)
	for i := range rows {
		if rows[i].ReviewStatus == model.AchievementReviewPending {
			rows[i].SuggestedStatus, rows[i].SuggestedReply = service.SuggestAchievementReply(rows[i])
		}
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/achievements/{id}/review
func (h *AchievementHandler) Review(w http.ResponseWriter, r *http.Request) {
	var req model.ReviewAchievementRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صالحة")
		return
	}
	me := middleware.EmployeeIDFromContext(r)
	cur, err := h.service.Repo().FindByID(r.PathValue("id"))
	if err != nil || cur == nil {
		WriteError(w, http.StatusNotFound, "الإنجاز مو موجود")
		return
	}
	if cur.EmployeeID == me || (reviewerIsMonitor(r) && isManagementRole(cur.Role)) {
		WriteError(w, http.StatusBadRequest, "ما تكدر تراجع هذا الإنجاز")
		return
	}
	a, err := h.service.Review(cur.ID, me, req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.notif != nil {
		msg := "📝 انراجع إنجازك اليومي ✓ زين"
		if a.ReviewStatus == model.AchievementReviewNeeds {
			msg = "📝 انراجع إنجازك اليومي — يحتاج مراجعة"
		}
		if note := strings.TrimSpace(req.Note); note != "" {
			msg += ": " + note
		}
		_ = h.notif.Create(a.EmployeeID, "achievement_review", msg)
	}
	WriteJSON(w, http.StatusOK, a)
}

// GET /api/achievements/mine — الموظف يشوف إنجازاته هو بس. چانت شاشة
// «إنجازاتي» تنادي قائمة المدير، فكل موظف ينرفض وما يشوف تاريخه.
func (h *AchievementHandler) Mine(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	rows, err := h.service.List(middleware.EmployeeIDFromContext(r), "", limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الإنجازات")
		return
	}
	h.attachVoices(rows)
	WriteJSON(w, http.StatusOK, rows)
}
