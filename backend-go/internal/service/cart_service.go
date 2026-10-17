package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type CartService struct {
	repo *repository.CartRepository
}

func NewCartService(repo *repository.CartRepository) *CartService {
	return &CartService{repo: repo}
}

func (s *CartService) ListForBooking(bookingID string) ([]model.CartItem, error) {
	return s.repo.ListForBooking(bookingID)
}

func (s *CartService) Create(bookingID string, req model.CreateCartItemRequest) (*model.CartItem, error) {
	if req.ProductName == "" || req.Quantity == nil || req.UnitPrice == nil {
		return nil, errors.New("productName, quantity, and unitPrice are required")
	}
	return s.repo.Create(bookingID, req.ProductName, *req.Quantity, *req.UnitPrice, req.Notes)
}

func (s *CartService) Update(id string, req model.UpdateCartItemRequest) (*model.CartItem, error) {
	return s.repo.Update(id, req)
}

func (s *CartService) Delete(id string) error {
	return s.repo.Delete(id)
}
