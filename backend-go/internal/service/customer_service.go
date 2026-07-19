package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type CustomerService struct {
	repo *repository.CustomerRepository
}

func NewCustomerService(repo *repository.CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) List() ([]model.CustomerResponse, error) {
	customers, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	out := make([]model.CustomerResponse, len(customers))
	for i, c := range customers {
		out[i] = c.ToResponse()
	}
	return out, nil
}

func (s *CustomerService) Lookup(phone string) (*model.CustomerResponse, error) {
	c, err := s.repo.FindByPhone(phone)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, nil
	}
	resp := c.ToResponse()
	count, err := s.repo.CountBookings(c.ID)
	if err == nil {
		resp.PreviousBookingsCount = &count
	}
	return &resp, nil
}

// FindOrCreate يعيد الزبون الموجود بنفس رقم الهاتف، أو ينشئ واحداً جديداً
func (s *CustomerService) FindOrCreate(req model.CreateCustomerRequest) (*model.CustomerResponse, error) {
	if req.Name == "" || req.Phone == "" {
		return nil, errors.New("الاسم ورقم الهاتف مطلوبان")
	}

	existing, err := s.repo.FindByPhone(req.Phone)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		resp := existing.ToResponse()
		existed := true
		resp.Existed = &existed
		return &resp, nil
	}

	created, err := s.repo.Create(req.Name, req.Phone, req.Location, req.MapLatitude, req.MapLongitude)
	if err != nil {
		return nil, err
	}
	resp := created.ToResponse()
	existed := false
	resp.Existed = &existed
	return &resp, nil
}
