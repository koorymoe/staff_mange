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

	// الشعبة مو تفصيل شكلي: هي الي تحدد مهارات هذي الخدمة تنعرض لأي كادر.
	// نتحقق منها هنا حتى ما تنحفظ قيمة غلط تخلي الخدمة ما تطلع لولا شعبة.
	division := model.DivisionEngineering
	if req.Division != nil && *req.Division != "" {
		if *req.Division != model.DivisionEngineering && *req.Division != model.DivisionDecor {
			return nil, errors.New("شعبة الخدمة لازم تكون هندسية أو ديكور")
		}
		division = *req.Division
	}

	svc := &model.Service{ID: uuid.NewString(), Name: req.Name, Category: req.Category, Division: division}
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
