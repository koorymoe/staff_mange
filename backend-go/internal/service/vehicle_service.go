package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type VehicleService struct {
	repo *repository.VehicleRepository
}

func NewVehicleService(repo *repository.VehicleRepository) *VehicleService {
	return &VehicleService{repo: repo}
}

func (s *VehicleService) List() ([]model.Vehicle, error) { return s.repo.List() }

func (s *VehicleService) Create(req model.CreateVehicleRequest) (*model.Vehicle, error) {
	if req.Name == "" || req.PlateNumber == "" {
		return nil, errors.New("اسم السيارة ورقم اللوحة مطلوبان")
	}
	return s.repo.Create(req)
}

func (s *VehicleService) ListLogs(vehicleID string) ([]model.VehicleLog, error) {
	return s.repo.ListLogs(vehicleID)
}

func (s *VehicleService) CreateLog(vehicleID string, req model.CreateVehicleLogRequest, recordedByID string) (*model.VehicleLog, error) {
	if req.Type != "FUEL" && req.Type != "CLEANING" && req.Type != "OIL_CHANGE" {
		return nil, errors.New("نوع السجل غير صحيح")
	}
	return s.repo.CreateLog(vehicleID, req, recordedByID)
}

func (s *VehicleService) ListIncidents(vehicleID string) ([]model.VehicleIncident, error) {
	return s.repo.ListIncidents(vehicleID)
}

func (s *VehicleService) CreateIncident(vehicleID string, req model.CreateVehicleIncidentRequest, reportedByID string) (*model.VehicleIncident, error) {
	if req.Type != "FAULT" && req.Type != "DAMAGE" {
		return nil, errors.New("نوع الحادثة غير صحيح")
	}
	if req.Description == "" {
		return nil, errors.New("وصف العطل/الضرر مطلوب")
	}
	return s.repo.CreateIncident(vehicleID, req, reportedByID)
}

func (s *VehicleService) UpdateIncident(id string, req model.UpdateVehicleIncidentRequest) (*model.VehicleIncident, error) {
	return s.repo.UpdateIncident(id, req)
}

func (s *VehicleService) ListMonthlyStatus(vehicleID string) ([]model.VehicleMonthlyStatus, error) {
	return s.repo.ListMonthlyStatus(vehicleID)
}

func (s *VehicleService) SetMonthlyStatus(vehicleID string, req model.SetVehicleMonthlyStatusRequest, recordedByID string) (*model.VehicleMonthlyStatus, error) {
	if req.Month == "" {
		return nil, errors.New("الشهر مطلوب")
	}
	return s.repo.SetMonthlyStatus(vehicleID, req, recordedByID)
}
