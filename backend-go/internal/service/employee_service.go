package service

import (
	"errors"

	"github.com/google/uuid"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type EmployeeService struct {
	repo *repository.EmployeeRepository
}

func NewEmployeeService(repo *repository.EmployeeRepository) *EmployeeService {
	return &EmployeeService{repo: repo}
}

func (s *EmployeeService) List() ([]model.Employee, error) {
	return s.repo.List()
}

func (s *EmployeeService) Get(id string) (*model.Employee, error) {
	return s.repo.FindByID(id)
}

func (s *EmployeeService) Create(req model.CreateEmployeeRequest) (*model.Employee, error) {
	if req.Name == "" {
		return nil, errors.New("الاسم مطلوب")
	}

	role := "TECHNICIAN"
	if req.Role != nil && *req.Role != "" {
		role = *req.Role
	}
	shift := "MORNING"
	if req.Shift != nil && *req.Shift != "" {
		shift = *req.Shift
	}

	employee := &model.Employee{
		ID:          uuid.NewString(),
		Name:        req.Name,
		Certificate: req.Certificate,
		Position:    req.Position,
		Phone:       req.Phone,
		Username:    req.Username,
		JobTitle:    req.JobTitle,
		Salary:      req.Salary,
		Shift:       &shift,
		ShiftStart:  req.ShiftStart,
		ShiftEnd:    req.ShiftEnd,
		Role:        role,
		Skills:      []model.EmployeeSkillDetail{},
	}

	if req.Password != nil && *req.Password != "" {
		hashed, err := HashPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		employee.Password = &hashed
	}

	if err := s.repo.Create(employee); err != nil {
		return nil, err
	}
	return employee, nil
}

func (s *EmployeeService) Update(id string, req model.UpdateEmployeeRequest) (*model.Employee, error) {
	employee, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		employee.Name = *req.Name
	}
	if req.Certificate != nil {
		employee.Certificate = req.Certificate
	}
	if req.Position != nil {
		employee.Position = req.Position
	}
	if req.Phone != nil {
		employee.Phone = req.Phone
	}
	if req.Status != nil {
		employee.Status = *req.Status
	}
	if req.Role != nil {
		employee.Role = *req.Role
	}
	if req.OnDuty != nil {
		employee.OnDuty = *req.OnDuty
	}
	if req.Username != nil {
		employee.Username = req.Username
	}
	if req.HasDrivingLicense != nil {
		employee.HasDrivingLicense = *req.HasDrivingLicense
	}
	if req.HasSafetyCertificate != nil {
		employee.HasSafetyCertificate = *req.HasSafetyCertificate
	}
	if req.IsLeader != nil {
		employee.IsLeader = *req.IsLeader
	}
	if req.IsTrainee != nil {
		employee.IsTrainee = *req.IsTrainee
	}
	if req.Salary != nil {
		employee.Salary = req.Salary
	}
	if req.Shift != nil {
		employee.Shift = req.Shift
	}
	if req.ShiftStart != nil {
		employee.ShiftStart = req.ShiftStart
	}
	if req.ShiftEnd != nil {
		employee.ShiftEnd = req.ShiftEnd
	}
	if req.MonthlyLeaves != nil {
		employee.MonthlyLeaves = *req.MonthlyLeaves
	}
	if req.JobTitle != nil {
		employee.JobTitle = req.JobTitle
	}

	if req.Password != nil && *req.Password != "" {
		hashed, err := HashPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		employee.Password = &hashed
	} else {
		empty := ""
		employee.Password = &empty
	}

	if err := s.repo.Update(employee); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *EmployeeService) Supervisors() ([]model.Employee, error) {
	return s.repo.Supervisors()
}

func (s *EmployeeService) Match(serviceID string) ([]model.Employee, error) {
	if serviceID == "" {
		return nil, errors.New("serviceId is required")
	}
	return s.repo.MatchForService(serviceID)
}

func (s *EmployeeService) SetSkills(employeeID string, req model.SetEmployeeSkillsRequest) (*model.Employee, error) {
	if err := s.repo.SetSkills(employeeID, req.Skills); err != nil {
		return nil, err
	}
	employee, err := s.repo.FindByID(employeeID)
	if err != nil {
		return nil, err
	}
	skills, err := s.repo.SkillsForEmployee(employeeID)
	if err != nil {
		return nil, err
	}
	employee.Skills = skills
	return employee, nil
}
