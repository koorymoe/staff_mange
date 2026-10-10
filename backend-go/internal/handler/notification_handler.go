package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

type NotificationHandler struct {
	service *service.NotificationService
}

func NewNotificationHandler(s *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: s}
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	employeeID := middleware.EmployeeIDFromContext(r)
	notifications, err := h.service.ListForEmployee(employeeID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الإشعارات")
		return
	}
	unread, _ := h.service.UnreadCount(employeeID)
	WriteJSON(w, http.StatusOK, map[string]any{
		"notifications": notifications,
		"unreadCount":   unread,
	})
}

func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	employeeID := middleware.EmployeeIDFromContext(r)
	if err := h.service.MarkRead(r.PathValue("id"), employeeID); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر تحديث الإشعار")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	employeeID := middleware.EmployeeIDFromContext(r)
	if err := h.service.MarkAllRead(employeeID); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر تحديث الإشعارات")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}
