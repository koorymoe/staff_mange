package service

import (
	"errors"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type KpiService struct {
	repo      *repository.KpiRepository
	employees *repository.EmployeeRepository
}

func NewKpiService(repo *repository.KpiRepository, employees *repository.EmployeeRepository) *KpiService {
	return &KpiService{repo: repo, employees: employees}
}

func (s *KpiService) List() ([]model.KpiEvaluation, error) {
	return s.repo.List()
}

func (s *KpiService) ListForEmployee(employeeID string) ([]model.KpiEvaluation, error) {
	return s.repo.ListForEmployee(employeeID)
}

// Create ينشئ تقييماً جديداً. إذا كان التقييم سلبياً، يرجّع الموظف تلقائياً لوضع
// "متدرب" (isTrainee) حتى يكمل التدريب المطلوب قبل ما يرجع لباقي شاشات النظام.
func (s *KpiService) Create(req model.CreateKpiEvaluationRequest) (*model.KpiEvaluation, error) {
	if req.EmployeeID == "" || req.EvaluatorID == "" || req.Points == nil {
		return nil, errors.New("employeeId, evaluatorId, and points are required")
	}
	deductionAmount := float64(*req.Points) * 10000
	eval, err := s.repo.Create(req.EmployeeID, req.EvaluatorID, *req.Points, req.Reason, deductionAmount)
	if err != nil {
		return nil, err
	}
	if *req.Points < 0 {
		_ = s.employees.SetTrainee(req.EmployeeID, true)
	}
	return eval, nil
}

// CompleteTraining يخرج الموظف من وضع "متدرب" ويسجل تقييماً إيجابياً يعكس
// اجتيازه للتدريب، بدل ما يظل التقييم السلبي وحده بسجله.
func (s *KpiService) CompleteTraining(employeeID, evaluatorID string) (*model.KpiEvaluation, error) {
	if err := s.employees.SetTrainee(employeeID, false); err != nil {
		return nil, err
	}
	points := 1
	deductionAmount := float64(points) * 10000
	return s.repo.Create(employeeID, evaluatorID, points, "إكمال التدريب بنجاح", deductionAmount)
}

func (s *KpiService) Delete(id string) error {
	return s.repo.Delete(id)
}

// RoleLeaderboard يرجع ترتيب موظفي دور معيّن أسبوعياً وشهرياً معاً
func (s *KpiService) RoleLeaderboard(role string) (*model.RoleKpiLeaderboard, error) {
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7).Format("2006-01-02")
	monthAgo := now.AddDate(0, -1, 0).Format("2006-01-02")

	weekly, err := s.repo.RoleLeaderboard(role, weekAgo)
	if err != nil {
		return nil, err
	}
	monthly, err := s.repo.RoleLeaderboard(role, monthAgo)
	if err != nil {
		return nil, err
	}
	return &model.RoleKpiLeaderboard{Role: role, Weekly: weekly, Monthly: monthly}, nil
}
