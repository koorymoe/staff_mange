package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type DeviceMaintenanceService struct {
	repo         *repository.DeviceMaintenanceRepository
	customerRepo *repository.CustomerRepository
}

func NewDeviceMaintenanceService(repo *repository.DeviceMaintenanceRepository, customerRepo *repository.CustomerRepository) *DeviceMaintenanceService {
	return &DeviceMaintenanceService{repo: repo, customerRepo: customerRepo}
}

func (s *DeviceMaintenanceService) Create(employeeID string, req model.CreateDeviceMaintenanceTicketRequest) (*model.DeviceMaintenanceTicket, error) {
	if req.DeviceTypeName == "" || req.Problem == "" {
		return nil, errors.New("نوع/اسم الجهاز والمشكلة مطلوبان")
	}
	customer, err := s.customerRepo.FindByCode(req.CustomerCode)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, errors.New("لا يوجد زبون بهذا الكود")
	}
	return s.repo.Create(employeeID, customer.ID, req)
}

func (s *DeviceMaintenanceService) List() ([]model.DeviceMaintenanceTicket, error) {
	return s.repo.List()
}

func (s *DeviceMaintenanceService) Update(id string, req model.UpdateDeviceMaintenanceTicketRequest) (*model.DeviceMaintenanceTicket, error) {
	return s.repo.Update(id, req)
}
