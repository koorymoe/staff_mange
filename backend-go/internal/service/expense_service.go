package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type ExpenseService struct {
	repo *repository.ExpenseRepository
}

func NewExpenseService(repo *repository.ExpenseRepository) *ExpenseService {
	return &ExpenseService{repo: repo}
}

func (s *ExpenseService) List(employeeID string) ([]model.Expense, error) {
	return s.repo.List(employeeID)
}

func (s *ExpenseService) Create(req model.CreateExpenseRequest) (*model.Expense, error) {
	if req.EmployeeID == "" || req.Amount == nil {
		return nil, errors.New("employeeId and amount are required")
	}
	if *req.Amount < 0 {
		return nil, errors.New("مبلغ المصروف ما يصير يكون بالسالب")
	}
	return s.repo.Create(req.EmployeeID, *req.Amount, req.Description)
}

func (s *ExpenseService) UpdateStatus(id string, req model.UpdateExpenseStatusRequest) (*model.Expense, error) {
	if req.Status != "APPROVED" && req.Status != "REJECTED" {
		return nil, errors.New("status must be APPROVED or REJECTED")
	}
	return s.repo.UpdateStatus(id, req.Status)
}
