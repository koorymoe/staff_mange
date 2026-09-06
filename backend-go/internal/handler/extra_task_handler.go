package handler

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// storyEmitter محرّك القصص — واجهة صغيرة: المعالج يعرف Emit بس.
type storyEmitter interface {
	Emit(req model.EmitStoryRequest)
}

// ExtraTaskHandler المهام الإضافية — المدير يوجّه شغل لموظف.
type ExtraTaskHandler struct {
	repo          *repository.ExtraTaskRepository
	notifications *repository.NotificationRepository
	stories       storyEmitter
}

func NewExtraTaskHandler(r *repository.ExtraTaskRepository, n *repository.NotificationRepository) *ExtraTaskHandler {
	return &ExtraTaskHandler{repo: r, notifications: n}
}

// SetStories يركّب محرّك القصص بعد الإنشاء — اختياري، والتوجيه يشتغل بدونه.
func (h *ExtraTaskHandler) SetStories(e storyEmitter) { h.stories = e }

// POST /api/extra-tasks — توجيه مهمة (للمدير).
func (h *ExtraTaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateExtraTaskRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	task, err := h.repo.Create(req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	// ⚠️ الإشعار هو الي يخلي المهمة «توصل» فعلاً. بدونه الموظف لازم
	// يفتح الشاشة بنفسه ويتصفّح — يعني نفس مشكلة التلفون بشكل ثاني.
	// وفشله ما يرجّع خطأ: المهمة انحفظت، وإفشال الطلب يخلي المدير
	// يعيد التوجيه فتنكرر المهمة.
	if h.notifications != nil {
		urgent := ""
		if task.Priority == model.ExtraTaskUrgent {
			urgent = "🔴 مستعجلة — "
		}
		_ = h.notifications.Create(task.AssignedToID, "extra_task",
			fmt.Sprintf("📋 %sمهمة جديدة موجّهة إلك: %s", urgent, task.Title))
	}

	// ⚠️ القصة **بعد** ما تنحفظ المهمة، ومرسِلها هو الموجِّه نفسه —
	// نفس قرار (ع): الموظف يعرف منو دزّها إله، مو «رسالة من النظام».
	if h.stories != nil && task != nil {
		reason := "بلا تفاصيل إضافية"
		if task.Description != nil && strings.TrimSpace(*task.Description) != "" {
			reason = *task.Description
		}
		h.stories.Emit(model.EmitStoryRequest{
			EventID:     task.ID,
			EventKind:   model.StoryEventAdminMessage,
			SenderID:    task.AssignedByID,
			RecipientID: task.AssignedToID,
			Payload: map[string]any{
				"title":  "مهمة موجّهة إلك: " + task.Title,
				"reason": reason,
				"link":   "/my-extra-tasks",
			},
		})
	}
	WriteJSON(w, http.StatusCreated, task)
}

// GET /api/extra-tasks/mine?includeDone=1 — مهام الموظف نفسه.
//
// ⚠️ المعرّف من التوكن مو من الطلب: بدونه أي موظف يقرا مهام غيره
// بتبديل رقم بالرابط.
func (h *ExtraTaskHandler) Mine(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListForEmployee(
		middleware.EmployeeIDFromContext(r),
		r.URL.Query().Get("includeDone") == "1")
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب مهامك")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/extra-tasks?status=&assigneeId= — للمدير.
func (h *ExtraTaskHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListAll(r.URL.Query().Get("status"), r.URL.Query().Get("assigneeId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المهام")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/extra-tasks/mine/count — عدّاد الشارة بالقائمة.
func (h *ExtraTaskHandler) MyOpenCount(w http.ResponseWriter, r *http.Request) {
	n, err := h.repo.OpenCountForEmployee(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر العدّ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]int{"count": n})
}

// PUT /api/extra-tasks/{id}/seen
func (h *ExtraTaskHandler) MarkSeen(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.MarkSeen(r.PathValue("id"), middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// PUT /api/extra-tasks/{id}/start
func (h *ExtraTaskHandler) Start(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.repo.Start(id, middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.respondTask(w, id)
}

// PUT /api/extra-tasks/{id}/complete
func (h *ExtraTaskHandler) Complete(w http.ResponseWriter, r *http.Request) {
	var req model.CompleteExtraTaskRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	note := strings.TrimSpace(req.DoneNote)
	// ⚠️ العدّ بالحروف مو بالبايتات — الحرف العربي بايتين وأكثر.
	if utf8.RuneCountInString(note) < 5 {
		WriteError(w, http.StatusBadRequest, "اكتب شنو سويت بالضبط — «تم» ما تكفي")
		return
	}
	id := r.PathValue("id")
	employeeID := middleware.EmployeeIDFromContext(r)
	if err := h.repo.Complete(id, employeeID, note); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	// نبلّغ الي وجّهها إنها انخلصت — هو الي ينتظر النتيجة.
	if task, err := h.repo.FindByID(id); err == nil && task != nil &&
		task.AssignedByID != nil && h.notifications != nil {
		who := ""
		if task.AssignedToName != nil {
			who = *task.AssignedToName
		}
		_ = h.notifications.Create(*task.AssignedByID, "extra_task_done",
			fmt.Sprintf("✅ %s خلّص المهمة: %s", who, task.Title))
	}
	h.respondTask(w, id)
}

// PUT /api/extra-tasks/{id}/cancel — المدير يلغي بسبب.
func (h *ExtraTaskHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Reason string `json:"reason"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	reason := strings.TrimSpace(req.Reason)
	if utf8.RuneCountInString(reason) < 3 {
		WriteError(w, http.StatusBadRequest, "اكتب سبب الإلغاء")
		return
	}
	id := r.PathValue("id")
	if err := h.repo.Cancel(id, reason); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	// الموظف لازم يعرف إنها انلغت — وإلا يشتغل بشي ما عاد مطلوب.
	if task, err := h.repo.FindByID(id); err == nil && task != nil && h.notifications != nil {
		_ = h.notifications.Create(task.AssignedToID, "extra_task_cancelled",
			fmt.Sprintf("✖️ انلغت المهمة «%s» — السبب: %s", task.Title, reason))
	}
	h.respondTask(w, id)
}

func (h *ExtraTaskHandler) respondTask(w http.ResponseWriter, id string) {
	task, err := h.repo.FindByID(id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المهمة")
		return
	}
	WriteJSON(w, http.StatusOK, task)
}
