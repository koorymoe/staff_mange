package service

import (
	"strings"
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
	if model.IsAdminRole(role) {
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
	if model.IsAdminRole(role) {
		return q, nil
	}
	hasManageAll, hasEditOwn, err := s.actorPerms(actorID)
	if err != nil {
		return nil, err
	}
	if hasManageAll || (hasEditOwn && q.CreatedByEmployeeID == actorID) {
		return q, nil
	}
	return nil, ErrQuotationHidden
}

// ErrQuotationHidden (11-04 — المبدأ ٦و) عرض موجود بس مو إلك: المعالج يرجّع
// نفس «مو موجود» ويسجّل الرفض — ما نكشف إن العرض موجود.
var ErrQuotationHidden = errors.New("عرض السعر غير موجود")

// Create (11-07): actorID/role من الجلسة — المنشئ ما ينوخذ من جسم الطلب (إلا الأدمن)،
// و«المشروع عنده عرض» يمر بنفس فحص الاطلاع مال Get (چان يرجّع عرض أي مشروع لأي أحد).
func (s *QuotationService) Create(req model.CreateQuotationRequest, actorID, role string) (*model.Quotation, error) {
	if !model.IsAdminRole(role) || req.CreatedByEmployeeID == "" {
		req.CreatedByEmployeeID = actorID
	}
	if req.CustomerName == "" || req.CreatedByEmployeeID == "" {
		return nil, errors.New("customerName and createdByEmployeeId are required")
	}
	if err := validateQuotationAmounts(req.Items, req.DiscountPercent); err != nil {
		return nil, err
	}
	// عرض واحد لكل مشروع: إذا المشروع عنده عرض نرجّعه بدل ما نسوي ثاني.
	if req.ProjectID != nil && *req.ProjectID != "" {
		existing, err := s.repo.FindByProject(*req.ProjectID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return s.Get(existing.ID, actorID, role)
		}
	}
	return s.repo.Create(req)
}

// ByProject يرجّع عرض المشروع (nil إذا ما عنده) بنفس فحص الاطلاع مال Get.
func (s *QuotationService) ByProject(projectID, actorID, role string) (*model.Quotation, error) {
	q, err := s.repo.FindByProject(projectID)
	if err != nil || q == nil {
		return nil, err
	}
	return s.Get(q.ID, actorID, role)
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
	if err := validateQuotationAmounts(req.Items, req.DiscountPercent); err != nil {
		return nil, err
	}
	// actorID ينمرّر حتى النسخة المؤرشفة تعرف منو عدّل — «منو بدّل
	// السعر» سؤال يجي متأخر ولازم إله جواب.
	return s.repo.Update(id, req, actorID)
}

func (s *QuotationService) Delete(id string) error {
	return s.repo.Delete(id)
}

// Versions النسخ المؤرشفة لعرض — نفس صلاحية قراءة العرض نفسه.
func (s *QuotationService) Versions(id string) ([]model.QuotationVersion, error) {
	return s.repo.Versions(id)
}

// ListPage نفس صلاحيات List، بس صفحة بصفحة وبلا بنود ولا صور.
func (s *QuotationService) ListPage(actorID, role, search string, page, limit int) (*repository.QuotationPage, error) {
	owner := ""
	if role != "ADMIN" && role != "OWNER" {
		hasManageAll, hasEditOwn, err := s.actorPerms(actorID)
		if err != nil {
			return nil, err
		}
		switch {
		case hasManageAll:
		case hasEditOwn:
			owner = actorID
		default:
			return &repository.QuotationPage{Items: []model.Quotation{}, Page: page}, nil
		}
	}
	return s.repo.ListPage(search, owner, page, limit)
}

// validateQuotationAmounts (11-07): الخصم بين ٠ و١٠٠، والكمية والسعر مو سالبين — بالإنشاء والتعديل.
// چان التعديل بلا أي فحص، والخصم بلا حد: خصم ١٥٠٪ يطلّع صافي سالب يدخل بإحصائيات المشاريع.
func validateQuotationAmounts(items []model.QuotationItemInput, discount *float64) error {
	if discount != nil && (*discount < 0 || *discount > 100) {
		return errors.New("نسبة الخصم لازم بين ٠ و ١٠٠")
	}
	for _, it := range items {
		if it.UnitPrice < 0 {
			return errors.New("سعر الوحدة ما يصير يكون بالسالب")
		}
		if it.Quantity < 0 {
			return errors.New("الكمية ما تصير سالبة")
		}
	}
	return nil
}

// ═══ بلاغ #33: ربط عرض موجود بمشروع ═══

var (
	ErrLinkTaken = errors.New("المشروع عنده عرض، أو العرض مربوط بمشروع ثاني")
)

// linkScope: "" = يشوف كل العروض، id = عروضه هو بس، ErrQuotationHidden = ماعنده عروض.
func (s *QuotationService) linkScope(actorID, role string) (string, error) {
	if model.IsAdminRole(role) {
		return "", nil
	}
	all, own, err := s.actorPerms(actorID)
	if err != nil {
		return "", err
	}
	if all {
		return "", nil
	}
	if own {
		return actorID, nil
	}
	return "", ErrQuotationHidden
}

// LinkCandidates عروض بلا مشروع تنفع لهذا المشروع (نفس الزبون أولاً) + بحث.
func (s *QuotationService) LinkCandidates(projectID, q, actorID, role string) ([]repository.LinkCandidate, error) {
	scope, err := s.linkScope(actorID, role)
	if err != nil {
		return nil, err
	}
	name, phone, ok := s.repo.ProjectBasics(projectID)
	if !ok {
		return nil, ErrQuotationHidden
	}
	return s.repo.LinkCandidates(name, phone, strings.TrimSpace(q), scope)
}

// LinkToProject يربط عرض موجود (يشوفه الموظف) بمشروع ما عنده عرض.
func (s *QuotationService) LinkToProject(projectID, quotationID, actorID, role string) (*model.Quotation, error) {
	if _, err := s.linkScope(actorID, role); err != nil {
		return nil, err
	}
	if _, _, ok := s.repo.ProjectBasics(projectID); !ok {
		return nil, ErrQuotationHidden
	}
	if _, err := s.Get(quotationID, actorID, role); err != nil {
		return nil, ErrQuotationHidden
	}
	ok, err := s.repo.LinkToProject(quotationID, projectID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrLinkTaken
	}
	return s.Get(quotationID, actorID, role)
}
