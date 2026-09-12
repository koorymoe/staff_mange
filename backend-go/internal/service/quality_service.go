package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type QualityService struct {
	repo *repository.QualityRepository
}

func NewQualityService(repo *repository.QualityRepository) *QualityService {
	return &QualityService{repo: repo}
}

func (s *QualityService) List(category string) ([]model.QualityIssue, error) {
	return s.repo.List(category)
}

func (s *QualityService) Create(req model.CreateQualityIssueRequest, reportedByID string) (*model.QualityIssue, error) {
	if req.Category != "EXECUTION" && req.Category != "OVERSIGHT" {
		return nil, errors.New("نوع المشكلة غير صحيح")
	}
	if req.Title == "" {
		return nil, errors.New("عنوان المشكلة مطلوب")
	}
	return s.repo.Create(req, reportedByID)
}

func (s *QualityService) Update(id string, req model.UpdateQualityIssueRequest) (*model.QualityIssue, error) {
	return s.repo.Update(id, req)
}
