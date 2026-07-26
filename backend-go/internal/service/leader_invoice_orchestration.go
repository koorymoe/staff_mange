package service

import (
	"fmt"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// LeaderInvoiceService ينسّق حساب فاتورة الليدر كاملة: تكاليف التنفيذ (بالكتالوج)
// + بنود المواد (بالأرشيف أو يدوي) + الخصم + كود المحاسبة، ثم يحفظها.
type LeaderInvoiceService struct {
	invoices  *repository.LeaderInvoiceRepository
	catalog   *repository.SystemPriceCatalogRepository
	materials *repository.MaterialRepository
}

func NewLeaderInvoiceService(
	invoices *repository.LeaderInvoiceRepository,
	catalog *repository.SystemPriceCatalogRepository,
	materials *repository.MaterialRepository,
) *LeaderInvoiceService {
	return &LeaderInvoiceService{invoices: invoices, catalog: catalog, materials: materials}
}

// Create يبني ويحفظ فاتورة ليدر جديدة من طلب الإنشاء، بحساب تكاليف التنفيذ
// ومجموع المواد صرفياً بالسيرفر (مو من الواجهة) — نفس منطق شيت "تكاليف المشروع".
func (s *LeaderInvoiceService) Create(employeeID string, req model.CreateLeaderInvoiceRequest) (*model.LeaderInvoice, error) {
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("يجب اختيار بند تنفيذ واحد على الأقل")
	}
	if len(req.Systems) == 0 {
		return nil, fmt.Errorf("يجب اختيار منظومة واحدة على الأقل")
	}
	if len(req.Systems) > 3 {
		return nil, fmt.Errorf("لا يمكن اختيار أكثر من 3 منظومات بالفاتورة الواحدة")
	}

	catalog, err := s.catalog.All()
	if err != nil {
		return nil, err
	}

	totalDeviceCount := 0
	for _, item := range req.Items {
		totalDeviceCount += item.Count
	}

	executionCost, err := CalculateExecutionCost(req.Items, catalog, totalDeviceCount)
	if err != nil {
		return nil, err
	}

	var materialLines []model.LeaderInvoiceMaterialItem
	var materialsTotal float64
	for _, line := range req.Materials {
		if line.Quantity <= 0 {
			return nil, fmt.Errorf("كمية المادة يجب أن تكون أكبر من صفر")
		}
		var materialID *string
		name := ""
		unitPrice := 0.0
		profit := 0.0
		if line.MaterialCode != nil && *line.MaterialCode != "" {
			mat, err := s.materials.GetByCode(*line.MaterialCode)
			if err != nil {
				return nil, err
			}
			if mat == nil {
				return nil, fmt.Errorf("لا توجد مادة بالكود %s", *line.MaterialCode)
			}
			materialID = &mat.ID
			name = mat.Name
			unitPrice = mat.SellPrice
			profit = mat.ProfitPerUnit
		} else if line.Name != nil && *line.Name != "" {
			name = *line.Name
			if line.UnitPrice != nil {
				unitPrice = *line.UnitPrice
			}
			if line.ProfitPerUnit != nil {
				profit = *line.ProfitPerUnit
			}
		} else {
			return nil, fmt.Errorf("كل بند مادة يحتاج كود أو اسم")
		}

		lineTotal := unitPrice * line.Quantity
		materialsTotal += lineTotal
		materialLines = append(materialLines, model.LeaderInvoiceMaterialItem{
			MaterialID:    materialID,
			Name:          name,
			Quantity:      line.Quantity,
			UnitPrice:     unitPrice,
			ProfitPerUnit: profit,
			LineTotal:     lineTotal,
		})
	}

	discount := req.DiscountValue
	if discount < 0 {
		discount = 0
	}
	netTotal := float64(executionCost) + materialsTotal - discount
	if netTotal < 0 {
		netTotal = 0
	}

	inv := &model.LeaderInvoice{
		BookingID:        req.BookingID,
		EmployeeID:       employeeID,
		CustomerName:     req.CustomerName,
		CustomerPhone:    req.CustomerPhone,
		CustomerAddress:  req.CustomerAddress,
		Systems:          req.Systems,
		Items:            req.Items,
		TotalDeviceCount: totalDeviceCount,
		ExecutionCost:    float64(executionCost),
		MaterialsTotal:   materialsTotal,
		DiscountValue:    discount,
		NetTotal:         netTotal,
		Status:           "SUBMITTED",
	}

	saved, err := s.invoices.Create(inv, materialLines)
	if err != nil {
		return nil, err
	}
	return saved, nil
}

func (s *LeaderInvoiceService) Get(id string) (*model.LeaderInvoice, error) {
	inv, err := s.invoices.GetByID(id)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, fmt.Errorf("الفاتورة غير موجودة")
	}
	return inv, nil
}

func (s *LeaderInvoiceService) List(employeeID string) ([]model.LeaderInvoice, error) {
	return s.invoices.List(employeeID)
}
