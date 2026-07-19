package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type QualityFollowUpService struct {
	repo *repository.QualityFollowUpRepository
}

func NewQualityFollowUpService(repo *repository.QualityFollowUpRepository) *QualityFollowUpService {
	return &QualityFollowUpService{repo: repo}
}

func (s *QualityFollowUpService) List() ([]model.QualityFollowUp, error) {
	return s.repo.List()
}

var validQualityFollowUpStatuses = map[string]bool{
	"PENDING":         true,
	"CONTACTED_OK":    true,
	"CONTACTED_ISSUE": true,
	"CONVERTED":       true,
	"CLOSED":          true,
}

func (s *QualityFollowUpService) Update(id, contactedByEmployeeID string, req model.UpdateQualityFollowUpRequest) (*model.QualityFollowUp, error) {
	if !validQualityFollowUpStatuses[req.Status] {
		return nil, errors.New("حالة متابعة غير معروفة")
	}
	return s.repo.Update(id, req.Status, contactedByEmployeeID, req.ContactNotes)
}
