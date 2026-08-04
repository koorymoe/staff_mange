package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) List() ([]model.Product, error) {
	return s.repo.List()
}

func (s *ProductService) Create(req model.CreateProductRequest) (*model.Product, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.DefaultPrice != nil && *req.DefaultPrice < 0 {
		return nil, errors.New("السعر الافتراضي ما يصير يكون بالسالب")
	}
	if req.WholesalePrice != nil && *req.WholesalePrice < 0 {
		return nil, errors.New("سعر الجملة ما يصير يكون بالسالب")
	}
	if req.Availability != nil && !model.ValidProductAvailability(*req.Availability) {
		return nil, errors.New("حالة التوفر غير معروفة")
	}
	return s.repo.Create(req)
}

func (s *ProductService) Update(id string, req model.UpdateProductRequest) (*model.Product, error) {
	if req.DefaultPrice != nil && *req.DefaultPrice < 0 {
		return nil, errors.New("السعر الافتراضي ما يصير يكون بالسالب")
	}
	if req.WholesalePrice != nil && *req.WholesalePrice < 0 {
		return nil, errors.New("سعر الجملة ما يصير يكون بالسالب")
	}
	if req.Availability != nil && !model.ValidProductAvailability(*req.Availability) {
		return nil, errors.New("حالة التوفر غير معروفة")
	}
	return s.repo.Update(id, req)
}

func (s *ProductService) Delete(id string) error {
	return s.repo.Delete(id)
}
