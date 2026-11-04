package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
)

// MeetingRequestHandler «اطلب مقابلة المدير» — أي موظف. الطلب يروح لتذكيرات
// السكرتير (المالك والمدير)، والموظف يشوف حالة طلباته بس.
type MeetingRequestHandler struct {
	repo          *repository.MeetingRequestRepository
	notifications *repository.NotificationRepository
}

func NewMeetingRequestHandler(r *repository.MeetingRequestRepository, n *repository.NotificationRepository) *MeetingRequestHandler {
	return &MeetingRequestHandler{repo: r, notifications: n}
}

// حد الطلبات المفتوحة للموظف — حتى ما تنغرق التذكيرات.
const maxOpenMeetingRequests = 3

// POST /api/meeting-requests
func (h *MeetingRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	var in struct {
		Kind        string `json:"kind"`
		Reason      string `json:"reason"`
		GuestName   string `json:"guestName"`
		PreferredAt string `json:"preferredAt"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	in.Reason, in.GuestName = strings.TrimSpace(in.Reason), strings.TrimSpace(in.GuestName)
	if in.Kind != "MEETING" && in.Kind != "GUEST" {
		WriteError(w, http.StatusBadRequest, "نوع الطلب لازم مقابلة أو ضيف")
		return
	}
	if n := len([]rune(in.Reason)); n < 3 || n > 500 {
		WriteError(w, http.StatusBadRequest, "اكتب سبب المقابلة (٣ إلى ٥٠٠ حرف)")
		return
	}
	var guest *string
	if in.Kind == "GUEST" {
		if n := len([]rune(in.GuestName)); n < 2 || n > 80 {
			WriteError(w, http.StatusBadRequest, "اكتب اسم الضيف")
			return
		}
		guest = &in.GuestName
	}
	var preferred *time.Time
	if in.PreferredAt != "" {
		t, err := time.ParseInLocation("2006-01-02T15:04", in.PreferredAt, secLoc)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "صيغة الوقت غلط")
			return
		}
		if t.Before(time.Now().Add(-time.Hour)) {
			WriteError(w, http.StatusBadRequest, "الوقت المناسب لازم يكون بالجاي")
			return
		}
		preferred = &t
	}
	if n, err := h.repo.OpenCount(me); err == nil && n >= maxOpenMeetingRequests {
		WriteError(w, http.StatusConflict, "عندك ٣ طلبات تنتظر — انتظر الرد عليهن أول")
		return
	}
	m, err := h.repo.Create(me, in.Kind, in.Reason, guest, preferred)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ الطلب")
		return
	}
	msg := "🗓️ " + m.Who + " يطلب مقابلة"
	if m.Kind == "GUEST" {
		msg = "🗓️ " + m.Who + " يريد يجيب ضيف (" + in.GuestName + ")"
	}
	_ = h.notifications.CreateForRolesOrPermission([]string{"OWNER", "ADMIN"}, "", "meeting_request", msg+" — بتذكيرات السكرتير")
	WriteJSON(w, http.StatusCreated, m)
}

// GET /api/meeting-requests/mine
func (h *MeetingRequestHandler) Mine(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.Mine(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الطلبات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
