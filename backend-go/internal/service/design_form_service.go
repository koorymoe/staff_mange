package service

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// DesignFormService يدير بنّاء أسئلة "وحدة التصميم" — المدير يضيف/يعدّل/يرتّب
// الأسئلة يدوياً حسب الاستمارة الي يريدها، بدون أي محتوى ثابت مكتوب بالكود.
type DesignFormService struct {
	repo *repository.DesignFormRepository
}

func NewDesignFormService(repo *repository.DesignFormRepository) *DesignFormService {
	return &DesignFormService{repo: repo}
}

func (s *DesignFormService) List() ([]model.DesignFormQuestion, error) {
	return s.repo.List()
}

func (s *DesignFormService) Create(req model.CreateDesignFormQuestionRequest) (*model.DesignFormQuestion, error) {
	if strings.TrimSpace(req.Label) == "" {
		return nil, fmt.Errorf("نص السؤال مطلوب")
	}
	if !model.IsValidDesignFormQuestionType(req.Type) {
		return nil, fmt.Errorf("نوع السؤال غير معروف")
	}
	order, err := s.repo.NextOrder()
	if err != nil {
		return nil, err
	}
	return s.repo.Create(uuid.NewString(), req.Label, req.Type, req.Options, req.Required, order)
}

func (s *DesignFormService) Update(id string, req model.UpdateDesignFormQuestionRequest) (*model.DesignFormQuestion, error) {
	if req.Type != nil && !model.IsValidDesignFormQuestionType(*req.Type) {
		return nil, fmt.Errorf("نوع السؤال غير معروف")
	}
	return s.repo.Update(id, req.Label, req.Type, req.Options, req.Required)
}

func (s *DesignFormService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *DesignFormService) Reorder(questionIDs []string) error {
	return s.repo.Reorder(questionIDs)
}
