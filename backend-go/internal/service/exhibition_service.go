package service

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ExhibitionService يدير "إدارة المعارض" (وحدة التقنيين) — أي تقني أو إداري
// يفتح معرضاً جديداً، لكن الترشيح (مين يروح فعلياً) حصراً بيد المدير. تقرير
// الزيارة يُولَّد بالذكاء الصناعي (نفس اتصال Gemini المستخدم بالمساعد الذكي).
type ExhibitionService struct {
	repo      *repository.ExhibitionRepository
	assistant *AssistantService
}

func NewExhibitionService(repo *repository.ExhibitionRepository, assistant *AssistantService) *ExhibitionService {
	return &ExhibitionService{repo: repo, assistant: assistant}
}

func (s *ExhibitionService) List() ([]model.Exhibition, error) {
	return s.repo.List()
}

func (s *ExhibitionService) Create(req model.CreateExhibitionRequest, createdByID string) (*model.Exhibition, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, fmt.Errorf("اسم المعرض مطلوب")
	}
	return s.repo.Create(uuid.NewString(), req.Title, req.Location, req.StartDate, req.EndDate, req.Companies, req.ProductsToShow, createdByID)
}

func (s *ExhibitionService) Nominate(id string, employeeIDs []string) (*model.Exhibition, error) {
	return s.repo.Nominate(id, employeeIDs)
}

func (s *ExhibitionService) AddPhotos(id string, photoUrls []string) (*model.Exhibition, error) {
	return s.repo.AddPhotos(id, photoUrls)
}

func (s *ExhibitionService) SetFindings(id, keyFindings string) (*model.Exhibition, error) {
	return s.repo.SetFindings(id, keyFindings)
}

func (s *ExhibitionService) Archive(id string) (*model.Exhibition, error) {
	return s.repo.Archive(id)
}

// GenerateVisitReport يبني تقرير زيارة نصي بالذكاء الصناعي من الشركات الحاضرة
// وأهم ما اكتُشف بالمعرض (يكتبه التقني بحقل keyFindings)، ويحفظه بالسجل.
func (s *ExhibitionService) GenerateVisitReport(id string) (*model.Exhibition, error) {
	ex, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("المعرض غير موجود")
	}

	findings := ""
	if ex.KeyFindings != nil {
		findings = *ex.KeyFindings
	}
	prompt := fmt.Sprintf(`اكتب تقرير زيارة معرض تجاري احترافي باللغة العربية بناءً على المعطيات التالية — منظّم بعناوين فرعية واضحة (نظرة عامة، الشركات الحاضرة، أهم الملاحظات والفرص):
اسم المعرض: %s
الموقع: %s
من %s إلى %s
الشركات الحاضرة: %s
أهم ما اكتُشف/لوحظ بالمعرض: %s
عدد صور كارتات الأعمال المرفقة: %d
اجعل التقرير مختصراً ومباشراً (لا يتجاوز 300 كلمة).`,
		ex.Title, ex.Location, ex.StartDate, ex.EndDate,
		strings.Join(ex.Companies, "، "), findings, len(ex.BusinessCardPhotos))

	report, err := s.assistant.GenerateReport(prompt)
	if err != nil {
		return nil, fmt.Errorf("تعذر توليد التقرير بالذكاء الصناعي حالياً")
	}
	return s.repo.SetVisitReport(id, report)
}
