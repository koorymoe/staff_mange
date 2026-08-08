package service

import (
	"errors"
	"strings"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type QualityFollowUpService struct {
	repo          *repository.QualityFollowUpRepository
	notifications *repository.NotificationRepository
}

func NewQualityFollowUpService(repo *repository.QualityFollowUpRepository, notifications *repository.NotificationRepository) *QualityFollowUpService {
	return &QualityFollowUpService{repo: repo, notifications: notifications}
}

// Verdict حكم مهندس الجودة. الليدر ينوصله إشعار بالخصم وسببه —
// الغرامة الي ما يدري بيها ما تصلّح شي، بس تخلق زعل بلا مصدر واضح.
func (s *QualityFollowUpService) Verdict(id, byEmployeeID string, req model.QualityVerdictRequest) ([]model.QualityFollowUp, error) {
	penalized, err := s.repo.Verdict(id, byEmployeeID, req)
	if err != nil {
		return nil, err
	}
	s.notifyPenalty(penalized, "شكوى زبون بعد متابعة الجودة", req.Notes)
	return s.repo.List()
}

// Inspect نتيجة الكشف الميداني على شكوى موقوفة.
func (s *QualityFollowUpService) Inspect(id, byEmployeeID string, req model.QualityInspectionRequest) ([]model.QualityFollowUp, error) {
	penalized, err := s.repo.Inspect(id, byEmployeeID, req)
	if err != nil {
		return nil, err
	}
	s.notifyPenalty(penalized, "الكشف أكّد شكوى الزبون", req.Notes)
	return s.repo.List()
}

func (s *QualityFollowUpService) notifyPenalty(employeeID, headline, notes string) {
	if employeeID == "" || s.notifications == nil {
		return
	}
	_ = s.notifications.Create(employeeID, "quality_penalty",
		"⚠️ انخصمت نقطة «"+model.QualityKpiCriterion+"» — "+headline+": "+strings.TrimSpace(notes))
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
