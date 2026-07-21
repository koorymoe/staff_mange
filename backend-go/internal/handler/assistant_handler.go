package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

type AssistantHandler struct {
	service *service.AssistantService
}

func NewAssistantHandler(s *service.AssistantService) *AssistantHandler {
	return &AssistantHandler{service: s}
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
	reply, err := h.service.ManagerChat(req.Message, req.History)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"reply": reply})
}
