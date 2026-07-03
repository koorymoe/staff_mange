package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type QuotationService struct {
	repo *repository.QuotationRepository
}

func NewQuotationService(repo *repository.QuotationRepository) *QuotationService {
	return &QuotationService{repo: repo}
}

func (s *QuotationService) List() ([]model.Quotation, error) {
	return s.repo.List()
}

func (s *QuotationService) Get(id string) (*model.Quotation, error) {
	q, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if q == nil {
		return nil, errors.New("Quotation not found")
	}
	return q, nil
}

func (s *QuotationService) Create(req model.CreateQuotationRequest) (*model.Quotation, error) {
	if req.CustomerName == "" || req.CreatedByEmployeeID == "" {
		return nil, errors.New("customerName and createdByEmployeeId are required")
	}
	return s.repo.Create(req)
}

func (s *QuotationService) Update(id string, req model.UpdateQuotationRequest) (*model.Quotation, error) {
	return s.repo.Update(id, req)
}

func (s *QuotationService) Delete(id string) error {
	return s.repo.Delete(id)
}
