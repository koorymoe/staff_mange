package service

import (
	"errors"

	"github.com/google/uuid"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type ServiceCatalogService struct {
	repo *repository.ServiceRepository
}

func NewServiceCatalogService(repo *repository.ServiceRepository) *ServiceCatalogService {
	return &ServiceCatalogService{repo: repo}
}

func (s *ServiceCatalogService) List() ([]model.Service, error) {
	return s.repo.List()
}

func (s *ServiceCatalogService) Create(req model.CreateServiceRequest) (*model.Service, error) {
	if req.Name == "" {
		return nil, errors.New("اسم الخدمة مطلوب")
	}

	svc := &model.Service{ID: uuid.NewString(), Name: req.Name, Category: req.Category}
	if err := s.repo.Create(svc); err != nil {
		return nil, err
	}
	svc.Skills = []model.Skill{}
	return svc, nil
}
