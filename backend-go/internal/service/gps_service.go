package service

import (
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type GpsService struct {
	repo *repository.GpsRepository
}

func NewGpsService(repo *repository.GpsRepository) *GpsService {
	return &GpsService{repo: repo}
}

func (s *GpsService) ListCustomers() ([]model.GpsCustomer, error) {
	return s.repo.ListCustomers()
}

func (s *GpsService) CreateCustomer(req model.UpsertGpsCustomerRequest) (*model.GpsCustomer, error) {
	return s.repo.CreateCustomer(req)
}

func (s *GpsService) UpdateCustomer(id string, req model.UpsertGpsCustomerRequest) (*model.GpsCustomer, error) {
	return s.repo.UpdateCustomer(id, req)
}

func (s *GpsService) ListSims() ([]model.SimCard, error) {
	return s.repo.ListSims()
}

func (s *GpsService) CreateSim(req model.UpsertSimCardRequest) (*model.SimCard, error) {
	return s.repo.CreateSim(req)
}

func (s *GpsService) UpdateSim(id string, req model.UpsertSimCardRequest) (*model.SimCard, error) {
	return s.repo.UpdateSim(id, req)
}

func (s *GpsService) ListDevices() ([]model.GpsDeviceRequest, error) {
	return s.repo.ListDevices()
}

func (s *GpsService) CreateDevice(req model.UpsertGpsDeviceRequest) (*model.GpsDeviceRequest, error) {
	return s.repo.CreateDevice(req)
}

func (s *GpsService) UpdateDevice(id string, req model.UpsertGpsDeviceRequest) (*model.GpsDeviceRequest, error) {
	return s.repo.UpdateDevice(id, req)
}

func (s *GpsService) ListRenewals() ([]model.GpsRenewalRequest, error) {
	return s.repo.ListRenewals()
}

func (s *GpsService) CreateRenewal(req model.UpsertGpsRenewalRequest) (*model.GpsRenewalRequest, error) {
	return s.repo.CreateRenewal(req)
}

func (s *GpsService) UpdateRenewal(id string, req model.UpsertGpsRenewalRequest) (*model.GpsRenewalRequest, error) {
	return s.repo.UpdateRenewal(id, req)
}

func (s *GpsService) ListMaintenance() ([]model.GpsMaintenanceRequest, error) {
	return s.repo.ListMaintenance()
}

func (s *GpsService) CreateMaintenance(req model.UpsertGpsMaintenanceRequest) (*model.GpsMaintenanceRequest, error) {
	return s.repo.CreateMaintenance(req)
}

func (s *GpsService) UpdateMaintenance(id string, req model.UpsertGpsMaintenanceRequest) (*model.GpsMaintenanceRequest, error) {
	return s.repo.UpdateMaintenance(id, req)
}

func (s *GpsService) ListSettings() ([]model.GpsSubscriptionPrice, error) {
	return s.repo.ListSettings()
}

func (s *GpsService) UpsertSetting(req model.UpsertGpsSubscriptionPriceRequest) (*model.GpsSubscriptionPrice, error) {
	id := "default"
	if req.ID != nil && *req.ID != "" {
		id = *req.ID
	}
	return s.repo.UpsertSetting(id, req.SubscriptionType, req.Price)
}

func (s *GpsService) Stats() (*model.GpsStats, error) {
	return s.repo.Stats()
}
