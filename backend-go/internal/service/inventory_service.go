package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type InventoryService struct {
	repo *repository.InventoryRepository
}

func NewInventoryService(repo *repository.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

func (s *InventoryService) CreateInventoryCheck(employeeID string, req model.CreateInventoryCheckRequest) (*model.InventoryCheck, error) {
	return s.repo.CreateInventoryCheck(employeeID, req)
}

func (s *InventoryService) TodaysInventoryChecks() ([]model.InventoryCheck, error) {
	return s.repo.TodaysInventoryChecks()
}

func (s *InventoryService) ResolveInventoryCheck(id string, resolvedByID string) (*model.InventoryCheck, error) {
	return s.repo.ResolveInventoryCheck(id, resolvedByID)
}

func (s *InventoryService) ListPersonalTools(employeeID string) ([]model.PersonalTool, error) {
	return s.repo.ListPersonalTools(employeeID)
}

func (s *InventoryService) CreatePersonalTool(req model.CreatePersonalToolRequest) (*model.PersonalTool, error) {
	if req.EmployeeID == "" || req.Name == "" || req.Barcode == "" {
		return nil, errors.New("employeeId, name and barcode are required")
	}
	return s.repo.CreatePersonalTool(req.EmployeeID, req.Name, req.Barcode)
}

func (s *InventoryService) UpdatePersonalTool(id string, req model.UpdatePersonalToolRequest) (*model.PersonalTool, error) {
	return s.repo.UpdatePersonalTool(id, req.Status, req.CheckedOut)
}

func (s *InventoryService) DeletePersonalTool(id string) error {
	return s.repo.DeletePersonalTool(id)
}

func (s *InventoryService) ListVehicleTools(vehicleID string) ([]model.VehicleTool, error) {
	return s.repo.ListVehicleTools(vehicleID)
}

func (s *InventoryService) CreateVehicleTool(req model.CreateVehicleToolRequest) (*model.VehicleTool, error) {
	if req.Name == "" || req.Barcode == "" || req.VehicleID == "" {
		return nil, errors.New("name, barcode and vehicleId are required")
	}
	return s.repo.CreateVehicleTool(req.Name, req.Barcode, req.VehicleID)
}

func (s *InventoryService) UpdateVehicleTool(id string, req model.UpdateVehicleToolRequest) (*model.VehicleTool, error) {
	return s.repo.UpdateVehicleTool(id, req.Name, req.Barcode, req.Status)
}

func (s *InventoryService) DeleteVehicleTool(id string) error {
	return s.repo.DeleteVehicleTool(id)
}

func (s *InventoryService) ListOnDemandTools() ([]model.OnDemandTool, error) {
	return s.repo.ListOnDemandTools()
}

func (s *InventoryService) CreateOnDemandTool(req model.CreateOnDemandToolRequest) (*model.OnDemandTool, error) {
	if req.Name == "" || req.Barcode == "" {
		return nil, errors.New("name and barcode are required")
	}
	return s.repo.CreateOnDemandTool(req.Name, req.Barcode, req.TotalQuantity, req.AvailableQuantity)
}

func (s *InventoryService) UpdateOnDemandTool(id string, req model.UpdateOnDemandToolRequest) (*model.OnDemandTool, error) {
	return s.repo.UpdateOnDemandTool(id, req.Name, req.Barcode, req.Status, req.TotalQuantity, req.AvailableQuantity)
}

func (s *InventoryService) ListToolRequests(employeeID string) ([]model.ToolRequest, error) {
	return s.repo.ListToolRequests(employeeID)
}

func (s *InventoryService) CreateToolRequest(req model.CreateToolRequestRequest) (*model.ToolRequest, error) {
	if req.EmployeeID == "" || req.ToolID == "" {
		return nil, errors.New("employeeId and toolId are required")
	}
	return s.repo.CreateToolRequest(req.EmployeeID, req.ToolID)
}

func (s *InventoryService) ApproveToolRequest(id string, req model.ApproveToolRequestRequest) (*model.ToolRequest, error) {
	if req.ApprovedByID == "" {
		return nil, errors.New("approvedById is required")
	}
	return s.repo.ApproveToolRequest(id, req.ApprovedByID)
}

func (s *InventoryService) RejectToolRequest(id string) (*model.ToolRequest, error) {
	return s.repo.RejectToolRequest(id)
}

func (s *InventoryService) ReturnToolRequest(id string) (*model.ToolRequest, error) {
	return s.repo.ReturnToolRequest(id)
}
