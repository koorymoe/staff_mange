package service

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type ProcurementService struct {
	repo        *repository.ProcurementRepository
	permissions *repository.PermissionRepository
	// monitor: صندوق المراقب. اختياري.
	monitor MonitorFeed
	// إشعار الطالب بقرار طلبه. اختياري.
	notifications *repository.NotificationRepository
}

// SetMonitorFeed يربط صندوق المراقب بعد البناء.
func (s *ProcurementService) SetMonitorFeed(m MonitorFeed) { s.monitor = m }

// ⚠️ نفس فجوة طلبات الأدوات: الطالب ما يعرف شنو صار بطلبه.
func (s *ProcurementService) SetNotificationRepository(n *repository.NotificationRepository) {
	s.notifications = n
}

func (s *ProcurementService) notifyRequester(employeeID, notifType, message string) {
	if s.notifications == nil || employeeID == "" {
		return
	}
	if err := s.notifications.Create(employeeID, notifType, message); err != nil {
		log.Printf("[procurement] تعذر إرسال إشعار القرار للموظف %s: %v", employeeID, err)
	}
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
	case model.RequestTypeManualSupply:
		return "procurement_manual", nil
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
	saved, err := s.repo.UpdateStatus(id, req.Status)
	if err != nil {
		return nil, err
	}
	if saved != nil {
		switch req.Status {
		case "REJECTED":
			s.notifyRequester(saved.RequestedByID, "procurement_decision", "❌ انرفض طلب المواد مالتك ("+saved.Code+")")
		case "APPROVED":
			s.notifyRequester(saved.RequestedByID, "procurement_decision", "✅ انوافق على طلب المواد مالتك ("+saved.Code+")")
		}
	}
	return saved, nil
}

// Fulfill أبو الكميات يجهّز الطلب. المورد إلزامي: بدونه ما نعرف من وين
// انجابت المادة، والسعر يبقى بلا سند نحاسب عليه.
func (s *ProcurementService) Fulfill(id string, req model.FulfillProcurementRequestRequest) (*model.ProcurementRequest, error) {
	if req.SupplierID == nil || *req.SupplierID == "" {
		return nil, errors.New("لازم تحدد المورد الي انجابت منه المادة")
	}
	saved, err := s.repo.Fulfill(id, req)
	if err != nil {
		return nil, err
	}
	// ⚠️ لحظة التجهيز هي لحظة **صرف الفلوس** — بعدها المادة انشترت
	// وما ينفع تراجع. هاي الي يحتاج المراقب يشوفها من إداري الكميات،
	// مو الطلب لما ينفتح.
	if s.monitor != nil && saved != nil {
		cost := "بلا مبلغ"
		if saved.TotalCost != nil {
			cost = fmt.Sprintf("%.0f د.ع", *saved.TotalCost)
		}
		// الاسم ممكن ما ينجلب مع الحفظ — ما نطبع «المورد: » فاضية
		supplier := ""
		if saved.Supplier != nil && saved.Supplier.Name != "" {
			supplier = " • المورد: " + saved.Supplier.Name
		}
		names := make([]string, 0, len(saved.Items))
		for _, it := range saved.Items {
			names = append(names, fmt.Sprintf("%s ×%d", it.ProductName, it.Quantity))
		}
		s.monitor.Stage(model.MonitorStageProcurementFulfilled, "PROCUREMENT", saved.ID,
			"طلب مواد "+saved.Code,
			fmt.Sprintf("الكلفة: %s%s • %s", cost, supplier, strings.Join(names, "، ")),
			"PROCUREMENT_ADMIN", saved.FulfilledByID)
	}
	return saved, nil
}
