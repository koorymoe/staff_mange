package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

// StoryHandler قصص الكيان — كلها **لصاحب الجلسة حصراً**.
//
// ⚠️ ماكو ولا مسار ياخذ `employeeId` من الرابط. الهوية تجي من التوكن
// وحده — نفس درس `/training/materials/mine` الي چان اسمه «مالتي»
// وياخذ الرقم من الرابط، فأي أحد يبدّله ويقرا مال غيره.
type StoryHandler struct{ service *service.StoryService }

func NewStoryHandler(s *service.StoryService) *StoryHandler { return &StoryHandler{service: s} }

// GET /api/stories/next — القصة الي دورها الآن، ومعها مشهدها.
//
// ⚠️ **وحدة بس**: قصة جسدية وحدة تلعب بالوقت، والباقي ينتظر بالطابور.
func (h *StoryHandler) Next(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	story, err := h.service.Next(me)
	if err != nil {
		log.Printf("story next: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب القصة")
		return
	}
	pending, err := h.service.PendingCount(me)
	if err != nil {
		log.Printf("story pending count: %v", err)
	}
	WriteJSON(w, http.StatusOK, map[string]any{"story": story, "pending": pending})
}

// GET /api/stories/mine — صندوق القصص (سجل مقروء بعد ما ينتهي المشهد).
func (h *StoryHandler) Mine(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	rows, err := h.service.Inbox(middleware.EmployeeIDFromContext(r), limit)
	if err != nil {
		log.Printf("story inbox: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الصندوق")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// advanceRequest المرحلة الجديدة وآخر خطوة انعرضت (checkpoint).
type advanceRequest struct {
	Status string `json:"status"`
	Step   int    `json:"step"`
}

// allowedAdvance المراحل الي تقدر الواجهة تعلنها.
//
// ⚠️ **`ACKNOWLEDGED` إقرار الموظف نفسه** بضغطة، مو استنتاج من إن
// المشهد انعرض. و«خرج من الشاشة» **ما ينحسب قراءة** — ولذلك ماكو
// مرحلة تنكتب تلقائياً من طرف المرسِل إطلاقاً.
var allowedAdvance = map[string]bool{
	model.StoryStatusDelivered:    true,
	model.StoryStatusPlaying:      true,
	model.StoryStatusSeen:         true,
	model.StoryStatusOpened:       true,
	model.StoryStatusAcknowledged: true,
}

// POST /api/stories/{id}/advance — تقدّم القصة مرحلة.
func (h *StoryHandler) Advance(w http.ResponseWriter, r *http.Request) {
	var req advanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	if !allowedAdvance[req.Status] {
		WriteError(w, http.StatusBadRequest, "مرحلة غير مسموحة: "+req.Status)
		return
	}
	if err := h.service.Advance(r.PathValue("id"), middleware.EmployeeIDFromContext(r), req.Status, req.Step); err != nil {
		log.Printf("story advance: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر تحديث القصة")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}
