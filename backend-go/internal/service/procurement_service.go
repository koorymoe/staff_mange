package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type ProcurementService struct {
	repo        *repository.ProcurementRepository
	permissions *repository.PermissionRepository
}

func NewProcurementService(repo *repository.ProcurementRepository, permissions *repository.PermissionRepository) *ProcurementService {
	return &ProcurementService{repo: repo, permissions: permissions}
}

func (s *ProcurementService) List() ([]model.ProcurementRequest, error) {
	return s.repo.List()
}

func (s *ProcurementService) Stats() (*model.ProcurementStats, error) {
	return s.repo.Stats()
}

// ErrForbiddenRequestType يُرجَع لما الموظف ما يملك صلاحية نوع الطلب المطلوب —
// الهاندلر يترجمها لـ 403 بدل 400 العادية.
var ErrForbiddenRequestType = errors.New("لا تملك صلاحية تقديم هذا النوع من الطلبات")

// permissionForRequestType يحدد اسم الصلاحية المطلوبة حسب نوع الطلب — كل نوع
// له صلاحية مستقلة يمنحها الأدمن يدوياً، ما تنجر تلقائياً مع أي دور.
func permissionForRequestType(requestType string) (string, error) {
	switch requestType {
	case model.RequestTypePersonalSupply:
		return "procurement_personal", nil
	case model.RequestTypeCustomerProduct:
		return "procurement_customer", nil
	default:
		return "", errors.New("نوع الطلب غير معروف")
	}
}

// Create ينشئ طلب مشتريات جديد بعد التحقق من صلاحية الموظف الخاصة بنوع الطلب
// (procurement_personal / procurement_customer) — ADMIN و OWNER يتخطون الفحص.
func (s *ProcurementService) Create(employeeID, role string, req model.CreateProcurementRequestRequest) (*model.ProcurementRequest, error) {
	if req.RequestType == "" {
		req.RequestType = model.RequestTypeCustomerProduct
	}
	permName, err := permissionForRequestType(req.RequestType)
	if err != nil {
		return nil, err
	}

	if role != "ADMIN" && role != "OWNER" {
		has, err := s.permissions.HasPermission(employeeID, permName)
		if err != nil {
			return nil, err
		}
		if !has {
			return nil, ErrForbiddenRequestType
		}
	}

	// صاحب الطلب ينتحدد من الجلسة مو من جسم الطلب. قبل هيچي كان
	// requestedById ينجي من العميل — يعني أي موظف يقدر يرفع طلب
	// باسم زميله، والحساب يروح على غيره.
	req.RequestedByID = employeeID
	return s.repo.Create(req)
}

func (s *ProcurementService) UpdateStatus(id string, req model.UpdateProcurementStatusRequest) (*model.ProcurementRequest, error) {
	return s.repo.UpdateStatus(id, req.Status)
}

func (s *ProcurementService) Fulfill(id string, req model.FulfillProcurementRequestRequest) (*model.ProcurementRequest, error) {
	return s.repo.Fulfill(id, req)
}
