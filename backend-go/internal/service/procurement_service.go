package service

import (
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type ProcurementService struct {
	repo *repository.ProcurementRepository
}

func NewProcurementService(repo *repository.ProcurementRepository) *ProcurementService {
	return &ProcurementService{repo: repo}
}

func (s *ProcurementService) List() ([]model.ProcurementRequest, error) {
	return s.repo.List()
}

func (s *ProcurementService) Stats() (*model.ProcurementStats, error) {
	return s.repo.Stats()
}

func (s *ProcurementService) Create(req model.CreateProcurementRequestRequest) (*model.ProcurementRequest, error) {
	return s.repo.Create(req)
}

func (s *ProcurementService) UpdateStatus(id string, req model.UpdateProcurementStatusRequest) (*model.ProcurementRequest, error) {
	return s.repo.UpdateStatus(id, req.Status)
}

func (s *ProcurementService) Fulfill(id string, req model.FulfillProcurementRequestRequest) (*model.ProcurementRequest, error) {
	return s.repo.Fulfill(id, req)
}
