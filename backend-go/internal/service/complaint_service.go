package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type ComplaintService struct {
	repo *repository.ComplaintRepository
}

func NewComplaintService(repo *repository.ComplaintRepository) *ComplaintService {
	return &ComplaintService{repo: repo}
}

func (s *ComplaintService) List() ([]model.Complaint, error) {
	return s.repo.List()
}

func (s *ComplaintService) Create(req model.CreateComplaintRequest) (*model.Complaint, error) {
	if req.CustomerID == "" || req.Description == "" || req.CreatedByEmployeeID == "" {
		return nil, errors.New("customerId, description, and createdByEmployeeId are required")
	}
	return s.repo.Create(req.CustomerID, req.BookingID, req.Description, req.CreatedByEmployeeID)
}

func (s *ComplaintService) Update(id string, req model.UpdateComplaintRequest) (*model.Complaint, error) {
	return s.repo.Update(id, req.Status, req.AssignedToEmployeeID, req.Resolution)
}

func (s *ComplaintService) Resolve(id string, req model.ResolveComplaintRequest) (*model.Complaint, error) {
	return s.repo.Resolve(id, req.Resolution)
}
