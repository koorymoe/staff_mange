package service

import (
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type PermissionService struct {
	permissions *repository.PermissionRepository
	employees   *repository.EmployeeRepository
	seeded      bool
}

func NewPermissionService(permissions *repository.PermissionRepository, employees *repository.EmployeeRepository) *PermissionService {
	return &PermissionService{permissions: permissions, employees: employees}
}

func (s *PermissionService) ensureSeeded() error {
	if s.seeded {
		return nil
	}
	if err := s.permissions.EnsureSeeded(); err != nil {
		return err
	}
	s.seeded = true
	return nil
}

func (s *PermissionService) ListAll() ([]model.Permission, error) {
	if err := s.ensureSeeded(); err != nil {
		return nil, err
	}
	return s.permissions.ListAll()
}

func (s *PermissionService) ListForEmployee(employeeID string) ([]model.Permission, error) {
	return s.permissions.ListForEmployee(employeeID)
}

func (s *PermissionService) SetForEmployee(employeeID string, permissionIDs []string) ([]model.Permission, error) {
	if err := s.permissions.ReplaceForEmployee(employeeID, permissionIDs); err != nil {
		return nil, err
	}
	return s.permissions.ListForEmployee(employeeID)
}

func (s *PermissionService) ApplyDefaults(employeeID string) ([]model.Permission, error) {
	employee, err := s.employees.FindByID(employeeID)
	if err != nil {
		return nil, err
	}

	defaults := model.RoleDefaultPermissions[employee.Role]
	if len(defaults) == 0 {
		return []model.Permission{}, nil
	}

	if err := s.ensureSeeded(); err != nil {
		return nil, err
	}

	all, err := s.permissions.ListAll()
	if err != nil {
		return nil, err
	}

	defaultSet := make(map[string]bool, len(defaults))
	for _, name := range defaults {
		defaultSet[name] = true
	}

	var defaultIDs []string
	for _, p := range all {
		if defaultSet[p.Name] {
			defaultIDs = append(defaultIDs, p.ID)
		}
	}

	if err := s.permissions.AddMissingForEmployee(employeeID, defaultIDs); err != nil {
		return nil, err
	}

	return s.permissions.ListForEmployee(employeeID)
}

func (s *PermissionService) RoleDefaults() map[string][]string {
	return model.RoleDefaultPermissions
}
