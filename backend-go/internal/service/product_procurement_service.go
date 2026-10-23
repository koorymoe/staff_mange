package service

import (
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ProductProcurementService يربط طلب المنتج بالدوار.
//
// وقت ما أبو الحسابات يجهّز الطلب: ينخصم المبلغ من الدوار، والطلب
// ينوافق عليه وينضاف المنتج للكتالوج (حتى التقني يلكاه بعرض السعر)،
// بس سطر التجهيز يبقى معلّق لحد ما المبلغ يرجع للدوار.
type ProductProcurementService struct {
	repo     *repository.ProductProcurementRepository
	requests *ProductRequestService
}

func NewProductProcurementService(repo *repository.ProductProcurementRepository, requests *ProductRequestService) *ProductProcurementService {
	return &ProductProcurementService{repo: repo, requests: requests}
}

func (s *ProductProcurementService) List(status string) ([]model.ProductProcurement, error) {
	return s.repo.List(status)
}

func (s *ProductProcurementService) Fulfill(requestID string, req model.FulfillProductRequest, byID string) (*model.ProductProcurement, error) {
	p, err := s.repo.Fulfill(requestID, req, byID)
	if err != nil {
		return nil, err
	}
	// الصرف صار فعلاً؛ فشل إضافة المنتج للكتالوج ما يلغي التجهيز —
	// المدير يقدر يضيفه يدوي، والفلوس مسجّلة على أي حال.
	if s.requests != nil {
		_, _ = s.requests.Approve(requestID, byID)
	}
	return p, nil
}

func (s *ProductProcurementService) Settle(id string, req model.SettleProcurementRequest, byID string) (*model.ProductProcurement, error) {
	return s.repo.Settle(id, req.Note, byID)
}
