package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type TrainingService struct {
	repo *repository.TrainingRepository
}

func NewTrainingService(repo *repository.TrainingRepository) *TrainingService {
	return &TrainingService{repo: repo}
}

func (s *TrainingService) MaterialsMine(employeeID string) (*model.MaterialsMineResponse, error) {
	services, err := s.repo.AssignedServices(employeeID)
	if err != nil {
		return nil, err
	}
	serviceIDs, err := s.repo.AssignedServiceIDs(employeeID)
	if err != nil {
		return nil, err
	}
	materials, err := s.repo.MaterialsForServices(serviceIDs)
	if err != nil {
		return nil, err
	}
	return &model.MaterialsMineResponse{Services: services, Materials: materials}, nil
}

func (s *TrainingService) Assignments(employeeID string) ([]model.Service, error) {
	return s.repo.AssignedServices(employeeID)
}

func (s *TrainingService) SetAssignments(employeeID string, req model.SetTrainingAssignmentsRequest) ([]model.Service, error) {
	if req.ServiceIDs == nil {
		return nil, errors.New("serviceIds must be an array")
	}
	if err := s.repo.SetAssignments(employeeID, req.ServiceIDs); err != nil {
		return nil, err
	}
	return s.repo.AssignedServices(employeeID)
}

func (s *TrainingService) ListMaterials(serviceID string) ([]model.TrainingMaterial, error) {
	return s.repo.ListMaterials(serviceID)
}

func (s *TrainingService) CreateMaterial(req model.CreateTrainingMaterialRequest) (*model.TrainingMaterial, error) {
	if req.ServiceID == "" || req.Title == "" || req.URL == "" {
		return nil, errors.New("serviceId, title and url are required")
	}
	materialType := "VIDEO"
	if req.Type != nil {
		materialType = *req.Type
	}
	order := 0
	if req.Order != nil {
		order = *req.Order
	}
	return s.repo.CreateMaterial(req.ServiceID, req.Title, req.URL, materialType, order)
}

func (s *TrainingService) UpdateMaterial(id string, req model.UpdateTrainingMaterialRequest) (*model.TrainingMaterial, error) {
	return s.repo.UpdateMaterial(id, req.Title, req.URL, req.Type, req.Order)
}

func (s *TrainingService) DeleteMaterial(id string) error {
	return s.repo.DeleteMaterial(id)
}
