package service

import (
	"errors"

	"github.com/google/uuid"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type AttendanceIconRequestService struct {
	repo      *repository.AttendanceIconRequestRepository
	employees *repository.EmployeeRepository
}

func NewAttendanceIconRequestService(repo *repository.AttendanceIconRequestRepository, employees *repository.EmployeeRepository) *AttendanceIconRequestService {
	return &AttendanceIconRequestService{repo: repo, employees: employees}
}

func (s *AttendanceIconRequestService) ListPending() ([]model.AttendanceIconRequest, error) {
	return s.repo.ListPending()
}

func (s *AttendanceIconRequestService) Create(employeeID, requestedIcon string) (*model.AttendanceIconRequest, error) {
	if requestedIcon == "" {
		return nil, errors.New("يرجى اختيار رمز")
	}
	return s.repo.Create(uuid.NewString(), employeeID, requestedIcon)
}

func (s *AttendanceIconRequestService) Approve(id, resolvedByID string) error {
	req, err := s.repo.FindByID(id)
	if err != nil || req == nil {
		return errors.New("الطلب غير موجود")
	}
	if err := s.employees.SetAttendanceIcon(req.EmployeeID, req.RequestedIcon); err != nil {
		return err
	}
	return s.repo.Resolve(id, "APPROVED", resolvedByID)
}

func (s *AttendanceIconRequestService) Reject(id, resolvedByID string) error {
	return s.repo.Resolve(id, "REJECTED", resolvedByID)
}
