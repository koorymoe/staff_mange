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

// Search بحث الزبائن بالسيرفر مع حد أقصى.
//
// ⚠️ لازم يمر بنفس تحويل List بالضبط (ToResponse + خدمات الزبون):
// الواجهة تعتمد على c.code وc.services، ولو رجّعنا الصف الخام تنكسر
// الشاشة. (صارت فعلاً وانلكت بفحص المتصفح.)
func (s *CustomerService) Search(search string, limit int) ([]model.CustomerResponse, error) {
	customers, err := s.repo.Search(search, limit)
	if err != nil {
		return nil, err
	}
	tags, err := s.repo.ServiceTagsByCustomer()
	if err != nil {
		return nil, err
	}
	out := make([]model.CustomerResponse, len(customers))
	for i, c := range customers {
		resp := c.ToResponse()
		if svc, ok := tags[c.ID]; ok {
			resp.Services = svc
		}
		out[i] = resp
	}
	return out, nil
}

// Get زبون واحد بالـ id (بند ٧م: بدل تنزيل كل الزبائن حتى نلگى واحد).
func (s *CustomerService) Get(id string) (*model.CustomerResponse, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	resp := c.ToResponse()
	return &resp, nil
}

// CustomerPage — بند ٧م: صفحة الزبائن ويا أرقام الرأس من الخادم.
type CustomerPage struct {
	Rows         []model.CustomerGpsResponse `json:"rows"`
	Total        int                         `json:"total"`
	All          int                         `json:"all"`
	NewThisMonth int                         `json:"newThisMonth"`
	VIP          int                         `json:"vip"`
	GPSCount     int                         `json:"gpsCount"`
	Locations    []string                    `json:"locations"`
}

func (s *CustomerService) Page(q repository.CustomerPageQuery) (*CustomerPage, error) {
	res, err := s.repo.Page(q)
	if err != nil {
		return nil, err
	}
	tags, err := s.repo.ServiceTagsByCustomer()
	if err != nil {
		return nil, err
	}
	out := &CustomerPage{Rows: make([]model.CustomerGpsResponse, len(res.Rows)), Total: res.Total, All: res.All,
		NewThisMonth: res.NewThisMonth, VIP: res.VIP, GPSCount: res.GPSCount, Locations: res.Locations}
	for i, r := range res.Rows {
		resp := r.Customer.ToResponse()
		if svc, ok := tags[r.ID]; ok {
			resp.Services = svc
		}
		out.Rows[i] = model.CustomerGpsResponse{CustomerResponse: resp, GpsNumber: r.GpsNumber, DeviceID: r.DeviceID, SubscriptionEnd: r.SubscriptionEnd}
	}
	return out, nil
}

func (s *CustomerService) ListGpsCustomers() ([]model.CustomerGpsResponse, error) {
	return s.repo.ListGpsCustomers()
}

func (s *CustomerService) Update(id string, req model.UpdateCustomerRequest) (*model.CustomerResponse, error) {
	if req.Name == "" || req.Phone == "" {
		return nil, errors.New("الاسم ورقم الهاتف مطلوبان")
	}
	c, err := s.repo.Update(id, req)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, nil
	}
	resp := c.ToResponse()
	return &resp, nil
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
	if open, err := s.repo.OpenBookings(c.ID); err == nil {
		resp.OpenBookings = open
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
