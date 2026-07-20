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

func (s *ComplaintService) StatsByCustomer() ([]model.ComplaintCustomerStat, error) {
	return s.repo.StatsByCustomer()
}

func (s *ComplaintService) Create(req model.CreateComplaintRequest) (*model.Complaint, error) {
	if req.CustomerID == "" || req.CreatedByEmployeeID == "" {
		return nil, errors.New("customerId و createdByEmployeeId مطلوبين")
	}
	if req.Type == "" {
		req.Type = "OTHER"
	}
	if _, ok := model.ComplaintTypeLabels[req.Type]; !ok {
		return nil, errors.New("نوع شكوى غير معروف")
	}
	return s.repo.Create(req.CustomerID, req.BookingID, req.Type, req.Description, req.CreatedByEmployeeID, req.RelatedEmployeeID)
}

func (s *ComplaintService) Update(id string, req model.UpdateComplaintRequest) (*model.Complaint, error) {
	return s.repo.Update(id, req.Status, req.AssignedToEmployeeID, req.Resolution)
}

func (s *ComplaintService) Resolve(id string, req model.ResolveComplaintRequest) (*model.Complaint, error) {
	return s.repo.Resolve(id, req.Resolution)
}
