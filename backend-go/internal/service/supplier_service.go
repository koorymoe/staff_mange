package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type SupplierService struct {
	repo *repository.SupplierRepository
}

func NewSupplierService(repo *repository.SupplierRepository) *SupplierService {
	return &SupplierService{repo: repo}
}

func (s *SupplierService) ListSpecialties() ([]model.SupplierSpecialty, error) {
	return s.repo.ListSpecialties()
}

func (s *SupplierService) CreateSpecialty(req model.CreateSupplierSpecialtyRequest) (*model.SupplierSpecialty, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	exists, err := s.repo.SpecialtyExists(req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("التخصص موجود مسبقاً")
	}
	count, err := s.repo.CountSpecialties()
	if err != nil {
		return nil, err
	}
	return s.repo.CreateSpecialty(req.Name, count)
}

func (s *SupplierService) DeleteSpecialty(id string) error {
	return s.repo.DeleteSpecialty(id)
}

func (s *SupplierService) List() ([]model.Supplier, error) {
	return s.repo.List()
}

func (s *SupplierService) Create(req model.UpsertSupplierRequest) (*model.Supplier, error) {
	return s.repo.Create(req)
}

func (s *SupplierService) Update(id string, req model.UpsertSupplierRequest) (*model.Supplier, error) {
	return s.repo.Update(id, req)
}

func (s *SupplierService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *SupplierService) Rate(supplierID string, req model.RateSupplierRequest) (*model.SupplierRating, error) {
	if req.Value == nil || *req.Value < 1 || *req.Value > 5 {
		return nil, errors.New("التقييم يجب أن يكون بين 1 و 5")
	}
	return s.repo.Rate(supplierID, *req.Value, req.Note, req.RatedByID, req.RatedByName)
}
