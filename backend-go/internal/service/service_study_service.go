package service

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ServiceStudyService يدير "إدارة الخدمات" (وحدة التقنيين) — خدمة جديدة
// مقترحة تحتاج دراسة قبل ما تصير خدمة رسمية: المدير يفتحها أو التقني، والمدير
// حصراً يوكّل التقنيين المسؤولين عنها، وكل موكَّل يرفع تقارير/دراسات.
type ServiceStudyService struct {
	repo *repository.ServiceStudyRepository
}

func NewServiceStudyService(repo *repository.ServiceStudyRepository) *ServiceStudyService {
	return &ServiceStudyService{repo: repo}
}

func (s *ServiceStudyService) List() ([]model.ServiceStudy, error) {
	return s.repo.List()
}

func (s *ServiceStudyService) Create(name, createdByID string) (*model.ServiceStudy, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("اسم الخدمة مطلوب")
	}
	return s.repo.Create(uuid.NewString(), name, createdByID)
}

func (s *ServiceStudyService) Assign(id string, employeeIDs []string) (*model.ServiceStudy, error) {
	if err := s.repo.SetAssignments(id, employeeIDs); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

// AddReport يسجّل تقرير/دراسة يرفعها تقني — يرفض العملية لو الموظف مو موكّل
// أصلاً بهذه الخدمة (المدير حصراً يحدد التوكيل عبر Assign).
func (s *ServiceStudyService) AddReport(serviceStudyID, employeeID, content string) (*model.ServiceStudyReport, error) {
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("محتوى التقرير مطلوب")
	}
	assigned, err := s.repo.IsAssigned(serviceStudyID, employeeID)
	if err != nil {
		return nil, err
	}
	if !assigned {
		return nil, fmt.Errorf("أنت غير موكَّل بدراسة هذه الخدمة")
	}
	return s.repo.AddReport(uuid.NewString(), serviceStudyID, employeeID, content)
}

func (s *ServiceStudyService) Archive(id string) (*model.ServiceStudy, error) {
	return s.repo.Archive(id)
}
