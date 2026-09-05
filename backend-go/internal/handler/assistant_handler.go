package handler

import (
	"net/http"
	"strconv"
	"time"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

type AssistantHandler struct {
	service       *service.AssistantService
	conversations *repository.AssistantConversationRepository
}

func NewAssistantHandler(s *service.AssistantService, conversations *repository.AssistantConversationRepository) *AssistantHandler {
	return &AssistantHandler{service: s, conversations: conversations}
}

type AskAssistantRequest struct {
	Message string `json:"message"`
}

// POST /api/assistant/ask — أي موظف مسجل دخول يسأل عن بياناته الشخصية بس
func (h *AssistantHandler) Ask(w http.ResponseWriter, r *http.Request) {
	var req AskAssistantRequest
	if err := DecodeJSON(r, &req); err != nil || req.Message == "" {
		WriteError(w, http.StatusBadRequest, "اكتب سؤالك أول")
		return
	}
	reply, err := h.service.Ask(middleware.EmployeeIDFromContext(r), req.Message)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"reply": reply})
}

type ManagerChatRequest struct {
	Message string             `json:"message"`
	History []service.ChatTurn `json:"history"`
}

// POST /api/assistant/manager-chat — محادثة حرة للمراقب/الأدمن، يسأل عن أي موظف بالاسم
func (h *AssistantHandler) ManagerChat(w http.ResponseWriter, r *http.Request) {
	var req ManagerChatRequest
	if err := DecodeJSON(r, &req); err != nil || req.Message == "" {
		WriteError(w, http.StatusBadRequest, "اكتب سؤالك أول")
		return
	}
	reply, err := h.service.ManagerChat(middleware.EmployeeIDFromContext(r), req.Message, req.History)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"reply": reply})
}

// GET /api/assistant/conversations — حصري للمالك: قائمة كل محادثات الموظفين
// مع المساعد الذكي، مع فلترة اختيارية بالموظف والتاريخ وتقسيم صفحات (limit/offset).
func (h *AssistantHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := repository.AssistantConversationFilter{
		EmployeeID: q.Get("employeeId"),
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Offset = n
		}
	}
	if v := q.Get("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.From = &t
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			end := t.Add(24*time.Hour - time.Second)
			filter.To = &end
		}
	}

	rows, total, err := h.conversations.ListAll(filter)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المحادثات")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"conversations": rows, "total": total})
}

// GET /api/assistant/conversations/employees — حصري للمالك: قائمة الموظفين
// المميزين اللي عندهم محادثة وحدة عالأقل، لتعبئة قائمة الفلتر بالواجهة.
func (h *AssistantHandler) ListConversationEmployees(w http.ResponseWriter, r *http.Request) {
	rows, err := h.conversations.ListEmployeesWithConversations()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب قائمة الموظفين")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"employees": rows})
}
