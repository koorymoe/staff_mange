package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// أسماء صلاحيات عروض الأسعار الثلاث (مستويات متدرجة، ما تنمنح مع بعض عادةً):
//   - quotation_create: إضافة عرض سعر جديد فقط — ما يشوف أي عروض سابقة إطلاقاً.
//   - quotation_edit_own: إضافة وتعديل — يشوف ويعدّل بس عروضه هو.
//   - quotation_manage_all: إضافة وتعديل واطلاع — يشوف ويعدّل كل عروض الأسعار
//     بغض النظر عن مين سواها.
//
// "quotation_system" القديمة تبقى تعمل كمرادف لـquotation_manage_all حتى ما
// ننكسر على أي موظف كانت ممنوحة له سابقاً قبل هذا التقسيم.
const (
	QuotationPermCreate    = "quotation_create"
	QuotationPermEditOwn   = "quotation_edit_own"
	QuotationPermManageAll = "quotation_manage_all"
	quotationPermLegacy    = "quotation_system"
)

type QuotationService struct {
	repo        *repository.QuotationRepository
	permissions *repository.PermissionRepository
}

func NewQuotationService(repo *repository.QuotationRepository, permissions *repository.PermissionRepository) *QuotationService {
	return &QuotationService{repo: repo, permissions: permissions}
}

func (s *QuotationService) actorPerms(actorID string) (hasManageAll, hasEditOwn bool, err error) {
	perms, err := s.permissions.ListForEmployee(actorID)
	if err != nil {
		return false, false, err
	}
	for _, p := range perms {
		if p.Name == QuotationPermManageAll || p.Name == quotationPermLegacy {
			hasManageAll = true
		}
		if p.Name == QuotationPermEditOwn {
			hasEditOwn = true
		}
	}
	return hasManageAll, hasEditOwn, nil
}

// List يرجّع عروض الأسعار حسب مستوى صلاحية الموظف: الأدمن/manage_all يشوف
// الكل (مع فلتر بحث اختياري)، edit_own يشوف بس عروضه، وأي مستوى أوطى (إضافة
// فقط، أو بدون صلاحية) يرجّعله قائمة فاضية عمداً — "ماريدهم يشوفون عروض
// الأسعار القديمة".
func (s *QuotationService) List(actorID, role, search string) ([]model.Quotation, error) {
	all, err := s.repo.List(search)
	if err != nil {
		return nil, err
	}
	if role == "ADMIN" || role == "OWNER" {
		return all, nil
	}

	hasManageAll, hasEditOwn, err := s.actorPerms(actorID)
	if err != nil {
		return nil, err
	}
	if hasManageAll {
		return all, nil
	}
	if hasEditOwn {
		mine := make([]model.Quotation, 0, len(all))
		for _, q := range all {
			if q.CreatedByEmployeeID == actorID {
				mine = append(mine, q)
			}
		}
		return mine, nil
	}
	return []model.Quotation{}, nil
}

func (s *QuotationService) Get(id, actorID, role string) (*model.Quotation, error) {
	q, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if q == nil {
		return nil, errors.New("Quotation not found")
	}
	if role == "ADMIN" || role == "OWNER" {
		return q, nil
	}
	hasManageAll, hasEditOwn, err := s.actorPerms(actorID)
	if err != nil {
		return nil, err
	}
	if hasManageAll || (hasEditOwn && q.CreatedByEmployeeID == actorID) {
		return q, nil
	}
	return nil, errors.New("لا تملك صلاحية الاطلاع على هذا العرض")
}

func (s *QuotationService) Create(req model.CreateQuotationRequest) (*model.Quotation, error) {
	if req.CustomerName == "" || req.CreatedByEmployeeID == "" {
		return nil, errors.New("customerName and createdByEmployeeId are required")
	}
	for _, item := range req.Items {
		if item.UnitPrice < 0 {
			return nil, errors.New("سعر الوحدة ما يصير يكون بالسالب")
		}
	}
	return s.repo.Create(req)
}

// Update يتحقق: manage_all/أدمن يعدّل أي عرض، edit_own يعدّل بس عروضه هو
// (غيرها مرفوض صراحة)، وأي مستوى أوطى ما يقدر يعدّل إطلاقاً.
func (s *QuotationService) Update(id, actorID, role string, req model.UpdateQuotationRequest) (*model.Quotation, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("Quotation not found")
	}
	if role != "ADMIN" && role != "OWNER" {
		hasManageAll, hasEditOwn, err := s.actorPerms(actorID)
		if err != nil {
			return nil, err
		}
		if !hasManageAll {
			if !hasEditOwn || existing.CreatedByEmployeeID != actorID {
				return nil, errors.New("تقدر تعدّل بس عروض الأسعار الي سويتها إنت")
			}
		}
	}
	return s.repo.Update(id, req)
}

func (s *QuotationService) Delete(id string) error {
	return s.repo.Delete(id)
}
