package service

import (
	"errors"
	"strings"

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

func (s *ServiceCatalogService) CreateSkill(serviceID string, req model.CreateSkillRequest) (*model.Skill, error) {
	if req.Name == "" {
		return nil, errors.New("اسم المهارة مطلوب")
	}
	sk := &model.Skill{ID: uuid.NewString(), Name: req.Name, ServiceID: serviceID}
	if err := s.repo.CreateSkill(sk); err != nil {
		return nil, err
	}
	return sk, nil
}

func (s *ServiceCatalogService) Delete(id string) error {
	err := s.repo.Delete(id)
	if err != nil && strings.Contains(err.Error(), "violates foreign key constraint") {
		return errors.New("لا يمكن حذف هذه الخدمة لوجود حجوزات مرتبطة بها")
	}
	return err
}
