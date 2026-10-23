package service

import (
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type NotificationService struct {
	repo *repository.NotificationRepository
}

func NewNotificationService(repo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) ListForEmployee(employeeID string) ([]model.Notification, error) {
	return s.repo.ListForEmployee(employeeID, 50)
}

func (s *NotificationService) UnreadCount(employeeID string) (int, error) {
	return s.repo.UnreadCount(employeeID)
}

func (s *NotificationService) MarkRead(id, employeeID string) error {
	return s.repo.MarkRead(id, employeeID)
}

func (s *NotificationService) MarkAllRead(employeeID string) error {
	return s.repo.MarkAllRead(employeeID)
}
