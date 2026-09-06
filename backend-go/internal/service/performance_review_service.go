package service

import (
	"errors"
	"fmt"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type PerformanceReviewService struct {
	repo      *repository.PerformanceReviewRepository
	employees *repository.EmployeeRepository
	bookings  *repository.BookingRepository
	// إشعار الإدارة ببلاغات المخالفة والالتزام. اختياري — بدونه
	// التقييم ينسجّل عادي بس ماكو منو ينتبه للبلاغ.
	notifications *repository.NotificationRepository
	permissions   *repository.PermissionRepository
}

func NewPerformanceReviewService(
	repo *repository.PerformanceReviewRepository,
	employees *repository.EmployeeRepository,
	bookings *repository.BookingRepository,
	notifications *repository.NotificationRepository,
	permissions *repository.PermissionRepository,
) *PerformanceReviewService {
	return &PerformanceReviewService{repo: repo, employees: employees, bookings: bookings, notifications: notifications, permissions: permissions}
}

// hasMonitorAccess: نفس صلاحيات ADMIN/OWNER/MONITOR بالتقييم — أو
// موظف انمنح monitoring/auditing فردياً (مثل ليدر ليوم واحد) بلا
// ما يتحول دوره فعلياً. نفس مبدأ RequireRoleOrAnyPermission بالخادم.
func (s *PerformanceReviewService) hasMonitorAccess(e *model.Employee) bool {
	if e.Role == "ADMIN" || e.Role == "OWNER" || e.Role == "MONITOR" {
		return true
	}
	if s.permissions == nil {
		return false
	}
	perms, err := s.permissions.ListForEmployee(e.ID)
	if err != nil {
		return false
	}
	for _, p := range perms {
		if p.Name == "monitoring" || p.Name == "auditing" {
			return true
		}
	}
	return false
}

// RatableEmployees يرجّع الموظفين الي المستخدم الحالي يقدر يقيّمهم حسب السلسلة
// الهرمية — الأدمن/المالك/المراقب يشوفون كل الموظفين، إداري الكوادر يشوف
// الليدرية والفنيين، أما التيم ليدر فيشوف بس فنيي حجوزاته الفعليين (زملاءه
// بالحجز، مو كل فنيي النظام).
func (s *PerformanceReviewService) RatableEmployees(evaluatorID string) ([]model.EmployeeBrief, error) {
	evaluator, err := s.employees.FindByID(evaluatorID)
	if err != nil || evaluator == nil {
		return nil, errors.New("تعذر تحديد هوية المقيّم")
	}

	// ⚠️ نفس فرع `authorizeReview` بالضبط — ADMIN/OWNER/MONITOR يقيّمون
	// أي أحد بلا قيد. كانت هذي الحالة تسقط للفرع الأخير (قائمة فارغة)
	// لعدم تطابقها مع أي شرط هنا — يعني المراقب والمالك يفتحون شاشة
	// التقييم ويشوفون قائمة فاضية رغم إن الخادم يقبل تقييمهم لأي أحد.
	if s.hasMonitorAccess(evaluator) {
		all, err := s.employees.List()
		if err != nil {
			return nil, err
		}
		result := make([]model.EmployeeBrief, 0, len(all))
		for _, e := range all {
			if e.ID != evaluatorID {
				result = append(result, model.EmployeeBrief{ID: e.ID, Name: e.Name})
			}
		}
		return result, nil
	}

	// ⚠️ نفس فرع `authorizeReview` — أبو الكوادر يقيّم الليدرية
	// والفنيين العاديين سوا.
	if evaluator.Role == "HR_COORDINATOR" {
		all, err := s.employees.List()
		if err != nil {
			return nil, err
		}
		result := make([]model.EmployeeBrief, 0, len(all))
		for _, e := range all {
			if e.IsLeader || e.Role == "TECHNICIAN" {
				result = append(result, model.EmployeeBrief{ID: e.ID, Name: e.Name})
			}
		}
		return result, nil
	}

	if evaluator.IsLeader {
		mates, err := s.bookings.CrewMatesForEmployee(evaluatorID)
		if err != nil {
			return nil, err
		}
		// نرجّع بس الفنيين العاديين (مو ليدرات ثانية) من بين زملاء الحجز —
		// نفس قيد authorizeReview بالضبط.
		result := make([]model.EmployeeBrief, 0, len(mates))
		for _, m := range mates {
			emp, err := s.employees.FindByID(m.ID)
			if err == nil && emp != nil && emp.Role == "TECHNICIAN" && !emp.IsLeader {
				result = append(result, m)
			}
		}
		return result, nil
	}

	return []model.EmployeeBrief{}, nil
}

// Create يسجل تقييم أداء بعد التحقق من السلسلة الهرمية:
//   - الأدمن يقيّم أي أحد
//   - إداري الكوادر (HR_COORDINATOR) يقيّم التيم ليدرات بس
//   - أي تيم ليدر يقيّم فنيي حجوزاته الفعليين بس (زملاءه بالحجز، مو كل فنيي
//     النظام) — يتحقق عبر BookingAssignment، مو مجرد الدور.
//
// تقييم سلبي يرجّع الموظف لوضع "متدرب" (نفس آلية قفل التدريب المستخدمة بـKPI).
func (s *PerformanceReviewService) Create(evaluatorID string, req model.CreatePerformanceReviewRequest) (*model.PerformanceReview, error) {
	if req.EmployeeID == "" || req.Reason == "" {
		return nil, errors.New("الموظف وسبب التقييم مطلوبين")
	}
	// NEGATIVE القديمة تنقبل وتنتحوّل — نسخة واجهة قديمة ما تصير
	// تفشل تقييم كتبه الليدر فعلاً.
	if req.Rating == "NEGATIVE" {
		req.Rating = model.ReviewNeedsTraining
	}
	if !model.ValidReviewRating(req.Rating) {
		return nil, errors.New("نوع التقييم غير معروف")
	}

	evaluator, err := s.employees.FindByID(evaluatorID)
	if err != nil {
		return nil, errors.New("تعذر تحديد هوية المقيّم")
	}
	target, err := s.employees.FindByID(req.EmployeeID)
	if err != nil {
		return nil, errors.New("الموظف غير موجود")
	}

	if err := s.authorizeReview(evaluator, target); err != nil {
		return nil, err
	}

	// الدرجات اختيارية، بس إذا انطاها لازم تكون ١-٥. رقم برّا المدى
	// يخرب كل متوسط ينحسب منه بعدين.
	if !model.ValidReviewScore(req.CommitmentScore) || !model.ValidReviewScore(req.SpeedScore) || !model.ValidReviewScore(req.QualityScore) {
		return nil, errors.New("درجة التقييم لازم تكون من ١ إلى ٥")
	}

	review, err := s.repo.Create(req.EmployeeID, evaluatorID, req.Rating, req.Reason, req.BookingID,
		req.CommitmentScore, req.SpeedScore, req.QualityScore)
	if err != nil {
		return nil, err
	}
	// ═══ كل نوع وأثره ═══
	//
	// ⚠️ التدريب بس هو الي ينفّذ تلقائياً. المخالفة وخلل الالتزام
	// **يبلّغون الإدارة** ولا يغرّمون — الليدر يبلّغ والإدارة تقرر.
	// إعطاء الليدر سلطة غرامة مباشرة على زملائه يخلي أي خلاف شخصي
	// يتحوّل خصم من راتب، والنظام يصير سلاح مو أداة.
	switch req.Rating {
	case model.ReviewNeedsTraining:
		_ = s.employees.SetTrainee(req.EmployeeID, true)

	case model.ReviewMisconduct, model.ReviewCommitment:
		if s.notifications != nil {
			target, _ := s.employees.FindByID(req.EmployeeID)
			by, _ := s.employees.FindByID(evaluatorID)
			name, byName := req.EmployeeID, evaluatorID
			if target != nil {
				name = target.Name
			}
			if by != nil {
				byName = by.Name
			}
			msg := fmt.Sprintf("⚠️ بلاغ %s: %s — بلّغ عنه %s (السبب: %s). القرار إلك.",
				model.ReviewRatingLabels[req.Rating], name, byName, req.Reason)
			// يروح لإداري الكوادر وللمراقب: الاثنين مخوّلين بالإجراء،
			// وواحد بس يخلي البلاغ ينتظر لو كان بإجازة.
			_ = s.notifications.CreateForRole("HR_COORDINATOR", "review_flag", msg)
			_ = s.notifications.CreateForRole("MONITOR", "review_flag", msg)
		}
	}
	return review, nil
}

func (s *PerformanceReviewService) authorizeReview(evaluator, target *model.Employee) error {
	if evaluator.ID == target.ID {
		return errors.New("ما تكدر تقيّم نفسك")
	}
	// ⚠️ المالك والمراقب المدقق ينضافون: الشاشة أصلاً معروضة إلهم
	// بالقائمة، وبدون هذا تنفتح إلهم وكل أزرارها ترجّع رفضاً.
	if s.hasMonitorAccess(evaluator) {
		return nil
	}
	// ⚠️ أبو الكوادر يقيّم الليدرية والفنيين العاديين سوا — كان محصوراً
	// بالليدرية بس («تيم ليدرات الفرق»)، وصاحب النظام قرر توسيعها: هو
	// المسؤول عن الطاقم الميداني كله لا شريحة منه بس.
	if evaluator.Role == "HR_COORDINATOR" {
		if target.IsLeader || target.Role == "TECHNICIAN" {
			return nil
		}
		return errors.New("إداري الكوادر يقيّم الليدرية والفنيين فقط")
	}
	if evaluator.IsLeader {
		if target.Role != "TECHNICIAN" || target.IsLeader {
			return errors.New("التيم ليدر يقيّم فنيي فريقه العاديين فقط")
		}
		mates, err := s.bookings.CrewMatesForEmployee(evaluator.ID)
		if err != nil {
			return errors.New("تعذر التحقق من فريق الحجوزات")
		}
		for _, m := range mates {
			if m.ID == target.ID {
				return nil
			}
		}
		return errors.New("تقدر تقيّم بس الفنيين الي طلعوا وياك بحجوزاتك")
	}
	return errors.New("لا تملك صلاحية تقييم الأداء")
}

func (s *PerformanceReviewService) ListForEmployee(employeeID string) ([]model.PerformanceReview, error) {
	return s.repo.ListForEmployee(employeeID)
}

func (s *PerformanceReviewService) List() ([]model.PerformanceReview, error) {
	return s.repo.List()
}

// BookingsAwaitingReview حجوزات الليدر المنجزة وكادر كل وحدة.
//
// ⚠️ المالك/المدير/المراقب يشوفون كل الحجوزات المنجزة مو حجوزاتهم:
// هذولا مو مكلّفين بأي حجز، فالنطاق القديم (الي يطلب يكونون بالفريق)
// چان يرجّعلهم صفراً دائماً والشاشة تطلع فارغة بأربع أصفار. نفس
// التفريع الموجود بـRatableEmployees فوك.
//
// والليدر يبقى على نطاقه: حجوزاته هو وبس.
//
// from/to اختياريان — إذا الاثنان فاضيان نرجع للافتراضي: آخر ٣٠ يوم.
// تقييم شغلة صارت قبل أشهر ما يفيد، بس لمن يختار المدير فترة صراحةً
// نجيبها إله، لأن «تجيب الداتا الي عدنه» ما تنفع بنافذة مقفلة.
func (s *PerformanceReviewService) BookingsAwaitingReview(
	viewerID, from, to string,
) ([]model.BookingAwaitingReview, error) {
	viewer, err := s.employees.FindByID(viewerID)
	if err != nil || viewer == nil {
		return nil, errors.New("تعذر تحديد هوية المستخدم")
	}
	if from == "" && to == "" {
		from = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	return s.repo.BookingsAwaitingReview(viewerID, s.hasMonitorAccess(viewer), from, to)
}

// EvaluatorLeaderboard ترتيب الإداريين حسب نشاط المراجعة (كم حجز
// راجعوا) أسبوعياً وشهرياً — «تقييم بين الإداريين». ⚠️ نفس نمط
// `KpiService.RoleLeaderboard` بالضبط (فترة حالية + فترة سابقة
// للمقارنة عبر `applyDeltas` المشتركة بالحزمة).
func (s *PerformanceReviewService) EvaluatorLeaderboard() (*model.RoleKpiLeaderboard, error) {
	const day = "2006-01-02"
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7).Format(day)
	monthAgo := now.AddDate(0, -1, 0).Format(day)
	twoWeeksAgo := now.AddDate(0, 0, -14).Format(day)
	twoMonthsAgo := now.AddDate(0, -2, 0).Format(day)

	weekly, err := s.repo.ReviewerLeaderboard(weekAgo, "")
	if err != nil {
		return nil, err
	}
	monthly, err := s.repo.ReviewerLeaderboard(monthAgo, "")
	if err != nil {
		return nil, err
	}
	prevWeekly, err := s.repo.ReviewerLeaderboard(twoWeeksAgo, weekAgo)
	if err != nil {
		return nil, err
	}
	prevMonthly, err := s.repo.ReviewerLeaderboard(twoMonthsAgo, monthAgo)
	if err != nil {
		return nil, err
	}

	applyDeltas(weekly, prevWeekly)
	applyDeltas(monthly, prevMonthly)

	return &model.RoleKpiLeaderboard{Role: "EVALUATORS", Weekly: weekly, Monthly: monthly}, nil
}
