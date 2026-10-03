package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// DesignFormService يدير استمارات "وحدة التصميم" — كل استمارة مستقلة بأسئلتها
// ورابطها العام الخاص بيها، المدير يضيف/يعدّل/يرتّب الأسئلة يدوياً بدون أي
// محتوى ثابت مكتوب بالكود، والزبون يعبّي الاستمارة عبر الرابط العام بدون
// تسجيل دخول ولا اطلاع على أي جزء ثاني من النظام.
type DesignFormService struct {
	repo *repository.DesignFormRepository
}

func NewDesignFormService(repo *repository.DesignFormRepository) *DesignFormService {
	return &DesignFormService{repo: repo}
}

func (s *DesignFormService) ListForms() ([]model.DesignForm, error) {
	return s.repo.ListForms()
}

func (s *DesignFormService) CreateForm(req model.CreateDesignFormRequest) (*model.DesignForm, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("اسم الاستمارة مطلوب")
	}
	return s.repo.CreateForm(uuid.NewString(), req.Name, uuid.NewString())
}

func (s *DesignFormService) DeleteForm(id string) error {
	return s.repo.DeleteForm(id)
}

func (s *DesignFormService) GetFormByToken(token string) (*model.DesignForm, error) {
	return s.repo.GetFormByToken(token)
}

func (s *DesignFormService) List(formID string) ([]model.DesignFormQuestion, error) {
	return s.repo.List(formID)
}

func (s *DesignFormService) Create(req model.CreateDesignFormQuestionRequest) (*model.DesignFormQuestion, error) {
	if strings.TrimSpace(req.FormID) == "" {
		return nil, fmt.Errorf("الاستمارة مطلوبة")
	}
	if strings.TrimSpace(req.Label) == "" {
		return nil, fmt.Errorf("نص السؤال مطلوب")
	}
	if !model.IsValidDesignFormQuestionType(req.Type) {
		return nil, fmt.Errorf("نوع السؤال غير معروف")
	}
	order, err := s.repo.NextOrder(req.FormID)
	if err != nil {
		return nil, err
	}
	return s.repo.Create(uuid.NewString(), req.FormID, req.Label, req.Type, req.Options, req.Required, order)
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

// Submit يستقبل جواب زبون عبر الرابط العام — يتأكد الاستمارة موجودة وكل سؤال
// مطلوب (required) عنده جواب قبل الحفظ.
// حدود التقديم العام — تمنع إغراق قاعدة البيانات من نموذج مفتوح للكل.
const (
	maxAnswerCount  = 200
	maxAnswerLen    = 5000
	maxAnswersBytes = 256 * 1024
)

func (s *DesignFormService) Submit(token string, req model.SubmitDesignFormRequest) (*model.DesignFormSubmission, error) {
	form, err := s.repo.GetFormByToken(token)
	if err != nil {
		return nil, fmt.Errorf("الاستمارة غير موجودة")
	}
	questions, err := s.repo.List(form.ID)
	if err != nil {
		return nil, err
	}
	if req.Answers == nil {
		req.Answers = map[string]any{}
	}
	// حد منطقي على حجم التقديم — الاستمارة عامة (بلا تسجيل دخول)، فبدون هذا
	// الحد يقدر أي واحد عنده الرابط يخزن حمولات ضخمة ويكبّر قاعدة البيانات
	// والنسخ الاحتياطية بلا سقف.
	if len(req.Answers) > maxAnswerCount {
		return nil, fmt.Errorf("عدد الإجابات أكبر من المسموح")
	}
	if encoded, err := json.Marshal(req.Answers); err == nil && len(encoded) > maxAnswersBytes {
		return nil, fmt.Errorf("حجم الإجابات أكبر من المسموح")
	}
	for k, v := range req.Answers {
		if sv, ok := v.(string); ok && len(sv) > maxAnswerLen {
			req.Answers[k] = sv[:maxAnswerLen]
		}
	}
	for _, q := range questions {
		if !q.Required {
			continue
		}
		v, ok := req.Answers[q.ID]
		if !ok || v == nil || v == "" {
			return nil, fmt.Errorf("الرجاء تعبئة السؤال: %s", q.Label)
		}
	}
	answersJSON, err := json.Marshal(req.Answers)
	if err != nil {
		return nil, err
	}
	return s.repo.CreateSubmission(uuid.NewString(), form.ID, answersJSON)
}

func (s *DesignFormService) ListSubmissions(formID string) ([]model.DesignFormSubmission, error) {
	return s.repo.ListSubmissions(formID)
}
