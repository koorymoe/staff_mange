package service

import (
	"errors"
	"fmt"

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
}

func NewPerformanceReviewService(
	repo *repository.PerformanceReviewRepository,
	employees *repository.EmployeeRepository,
	bookings *repository.BookingRepository,
	notifications *repository.NotificationRepository,
) *PerformanceReviewService {
	return &PerformanceReviewService{repo: repo, employees: employees, bookings: bookings, notifications: notifications}
}

// RatableEmployees يرجّع الموظفين الي المستخدم الحالي يقدر يقيّمهم حسب السلسلة
// الهرمية — الأدمن/إداري الكوادر يشوفون كل التيم ليدرات، أما التيم ليدر فيشوف
// بس فنيي حجوزاته الفعليين (زملاءه بالحجز، مو كل فنيي النظام).
func (s *PerformanceReviewService) RatableEmployees(evaluatorID string) ([]model.EmployeeBrief, error) {
	evaluator, err := s.employees.FindByID(evaluatorID)
	if err != nil || evaluator == nil {
		return nil, errors.New("تعذر تحديد هوية المقيّم")
	}

	if evaluator.Role == "ADMIN" || evaluator.Role == "HR_COORDINATOR" {
		all, err := s.employees.List()
		if err != nil {
			return nil, err
		}
		result := make([]model.EmployeeBrief, 0, len(all))
		for _, e := range all {
			if e.IsLeader {
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
	if evaluator.Role == "ADMIN" {
		return nil
	}
	if evaluator.Role == "HR_COORDINATOR" {
		if target.IsLeader {
			return nil
		}
		return errors.New("إداري الكوادر يقيّم تيم ليدرات الفرق فقط")
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
func (s *PerformanceReviewService) BookingsAwaitingReview(leaderID string) ([]model.BookingAwaitingReview, error) {
	return s.repo.BookingsAwaitingReview(leaderID)
}
