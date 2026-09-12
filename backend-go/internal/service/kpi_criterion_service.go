package service

import (
	"errors"
	"strings"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type KpiCriterionService struct {
	repo *repository.KpiCriterionRepository
}

func NewKpiCriterionService(repo *repository.KpiCriterionRepository) *KpiCriterionService {
	return &KpiCriterionService{repo: repo}
}

func (s *KpiCriterionService) List() ([]model.KpiCriterion, error) {
	return s.repo.List()
}

func (s *KpiCriterionService) Create(req model.CreateKpiCriterionRequest) (*model.KpiCriterion, error) {
	label := strings.TrimSpace(req.Label)
	if label == "" {
		return nil, errors.New("عنوان نقطة الكي بي اي مطلوب")
	}
	return s.repo.Create(label)
}

func (s *KpiCriterionService) Delete(id string) error {
	return s.repo.Delete(id)
}
