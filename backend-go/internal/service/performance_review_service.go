package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type PerformanceReviewService struct {
	repo      *repository.PerformanceReviewRepository
	employees *repository.EmployeeRepository
}

func NewPerformanceReviewService(repo *repository.PerformanceReviewRepository, employees *repository.EmployeeRepository) *PerformanceReviewService {
	return &PerformanceReviewService{repo: repo, employees: employees}
}

// Create يسجل تقييم أداء بعد التحقق من السلسلة الهرمية:
//   - الأدمن يقيّم أي أحد
//   - إداري الكوادر (HR_COORDINATOR) يقيّم التيم ليدرات بس
//   - أي تيم ليدر يقيّم الفنيين العاديين (اللي مو ليدرات) بس
//
// تقييم سلبي يرجّع الموظف لوضع "متدرب" (نفس آلية قفل التدريب المستخدمة بـKPI).
func (s *PerformanceReviewService) Create(evaluatorID string, req model.CreatePerformanceReviewRequest) (*model.PerformanceReview, error) {
	if req.EmployeeID == "" || req.Reason == "" {
		return nil, errors.New("الموظف وسبب التقييم مطلوبين")
	}
	if req.Rating != "POSITIVE" && req.Rating != "NEGATIVE" {
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

	if err := authorizeReview(evaluator, target); err != nil {
		return nil, err
	}

	review, err := s.repo.Create(req.EmployeeID, evaluatorID, req.Rating, req.Reason)
	if err != nil {
		return nil, err
	}
	if req.Rating == "NEGATIVE" {
		_ = s.employees.SetTrainee(req.EmployeeID, true)
	}
	return review, nil
}

func authorizeReview(evaluator, target *model.Employee) error {
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
		if target.Role == "TECHNICIAN" && !target.IsLeader {
			return nil
		}
		return errors.New("التيم ليدر يقيّم فنيي فريقه العاديين فقط")
	}
	return errors.New("لا تملك صلاحية تقييم الأداء")
}

func (s *PerformanceReviewService) ListForEmployee(employeeID string) ([]model.PerformanceReview, error) {
	return s.repo.ListForEmployee(employeeID)
}

func (s *PerformanceReviewService) List() ([]model.PerformanceReview, error) {
	return s.repo.List()
}
