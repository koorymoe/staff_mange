package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type KpiService struct {
	repo *repository.KpiRepository
}

func NewKpiService(repo *repository.KpiRepository) *KpiService {
	return &KpiService{repo: repo}
}

func (s *KpiService) List() ([]model.KpiEvaluation, error) {
	return s.repo.List()
}

func (s *KpiService) ListForEmployee(employeeID string) ([]model.KpiEvaluation, error) {
	return s.repo.ListForEmployee(employeeID)
}

func (s *KpiService) Create(req model.CreateKpiEvaluationRequest) (*model.KpiEvaluation, error) {
	if req.EmployeeID == "" || req.EvaluatorID == "" || req.Points == nil {
		return nil, errors.New("employeeId, evaluatorId, and points are required")
	}
	deductionAmount := float64(*req.Points) * 10000
	return s.repo.Create(req.EmployeeID, req.EvaluatorID, *req.Points, req.Reason, deductionAmount)
}

func (s *KpiService) Delete(id string) error {
	return s.repo.Delete(id)
}
