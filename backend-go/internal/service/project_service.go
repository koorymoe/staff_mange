package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

var projectStages = []string{
	"1. اتصال بالزبون",
	"2. مرحلة الكشف",
	"3. عرض السعر",
	"4. العقد",
	"5. البدء بالتنفيذ",
	"✅ مكتمل",
	"❌ مرفوض",
}

type ProjectService struct {
	repo *repository.ProjectRepository
}

func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func computeProjectStats(projects []model.Project) model.ProjectStats {
	var stats model.ProjectStats
	for _, p := range projects {
		s := p.Stage
		switch {
		case strings.Contains(s, "اتصال"):
			stats.Contact++
		case strings.Contains(s, "كشف"):
			stats.Survey++
		case strings.Contains(s, "سعر"):
			stats.Price++
		case strings.Contains(s, "عقد"):
			stats.Contract++
		case strings.Contains(s, "تنفيذ"):
			stats.Execute++
		case strings.Contains(s, "مكتمل"):
			stats.Done++
		case strings.Contains(s, "مرفوض"):
			stats.Rejected++
		}
	}
	return stats
}

func (s *ProjectService) List() (*model.ProjectListResponse, error) {
	projects, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	return &model.ProjectListResponse{
		Projects: projects,
		Stats:    computeProjectStats(projects),
		Stages:   projectStages,
	}, nil
}

// ListDelegatedTo مشاريع مُسلَّمة لموظف معيّن — نفس شكل القائمة العامة (مع
// إحصائياتها والمراحل) حتى الواجهة تعرضها بنفس طريقة إدارة المشاريع بالضبط.
func (s *ProjectService) ListDelegatedTo(employeeID string) (*model.ProjectListResponse, error) {
	projects, err := s.repo.ListDelegatedTo(employeeID)
	if err != nil {
		return nil, err
	}
	return &model.ProjectListResponse{
		Projects: projects,
		Stats:    computeProjectStats(projects),
		Stages:   projectStages,
	}, nil
}

// Statistics إحصائيات المشاريع الكاملة: نظرة عامة + سطر لكل مشروع بقيمته +
// سطر لكل موظف يبيّن دوره الفعلي (أضاف / طلع كشف / كان مسؤول / استلم مشروع).
func (s *ProjectService) Statistics() (*model.ProjectStatisticsResponse, error) {
	projects, err := s.repo.ProjectValueRows()
	if err != nil {
		return nil, err
	}
	employees, err := s.repo.ProjectEmployeeStats()
	if err != nil {
		return nil, err
	}

	ov := model.ProjectStatisticsOverview{TotalProjects: len(projects)}
	for _, p := range projects {
		val := 0.0
		if p.PriceValue != nil {
			val = *p.PriceValue
			ov.TotalValue += val
			ov.PricedProjects++
		}
		if p.DelegatedToName != nil {
			ov.DelegatedCount++
		}
		if p.HasSurvey {
			ov.SurveysFilled++
		}
		switch {
		case strings.Contains(p.Stage, "مكتمل"):
			ov.CompletedCount++
			ov.CompletedValue += val
		case strings.Contains(p.Stage, "مرفوض"):
			ov.RejectedCount++
		default:
			ov.ActiveCount++
			ov.InProgressValue += val
		}
	}
	if ov.PricedProjects > 0 {
		ov.AverageValue = ov.TotalValue / float64(ov.PricedProjects)
	}
	// نعيد استخدام نفس دالة توزيع المراحل المستخدمة بصفحة المشاريع حتى
	// الرقمين ما يختلفون بين الصفحتين
	stageSource := make([]model.Project, 0, len(projects))
	for _, p := range projects {
		stageSource = append(stageSource, model.Project{Stage: p.Stage})
	}
	ov.StageBreakdown = computeProjectStats(stageSource)

	return &model.ProjectStatisticsResponse{
		Overview:  ov,
		Projects:  projects,
		Employees: employees,
	}, nil
}

// Delegate يسلّم المشروع لموظف، أو يسحب التسليم لو employeeID فاضي.
func (s *ProjectService) Delegate(projectID, employeeID string, byEmployeeID *string, note string) (*model.Project, error) {
	var target *string
	if employeeID != "" {
		target = &employeeID
	}
	return s.repo.Delegate(projectID, target, byEmployeeID, note)
}

// HasAnyDelegation صحيح لو الموظف موجّه له أي مشروع.
func (s *ProjectService) HasAnyDelegation(employeeID string) (bool, error) {
	return s.repo.HasAnyDelegation(employeeID)
}

// IsDelegatedTo يفحص ملكية التوجيه — يستعملها الراوت قبل ما يسمح بالتعديل.
func (s *ProjectService) IsDelegatedTo(projectID, employeeID string) (bool, error) {
	return s.repo.IsDelegatedTo(projectID, employeeID)
}

// DelegationLog سجل التسليم لمشروع معيّن أو للكل.
func (s *ProjectService) DelegationLog(projectID string) ([]model.ProjectDelegationLogEntry, error) {
	return s.repo.DelegationLog(projectID)
}

func (s *ProjectService) Create(req model.CreateProjectRequest, createdBy *string) (*model.Project, error) {
	if req.Name == "" {
		return nil, errors.New("اسم المؤسسة مطلوب")
	}
	priority := "عادي"
	if req.Priority != nil {
		priority = *req.Priority
	}

	// الكود يُبنى من أكبر رقم موجود مو من عدد الصفوف، ومع ذلك يبقى احتمال تضارب
	// لو موظفين ضافوا مشروع بنفس اللحظة — فنعيد المحاولة عدة مرات على خطأ
	// التكرار (23505) بدل ما نطلّع الخطأ بوجه المستخدم.
	next, err := s.repo.NextCodeNumber()
	if err != nil {
		return nil, err
	}
	var lastErr error
	for attempt := 0; attempt < 10; attempt++ {
		code := fmt.Sprintf("PRJ-%04d", next+attempt)
		p, err := s.repo.Create(code, req.Name, req.Rep, req.Phone, req.Location, req.MapLatitude, req.MapLongitude, req.WorkType, req.RefPerson, priority, req.DeliveryDate, req.BookingID,
			emptyToNil(req.ResponsibleEmployeeID), emptyToNil(req.SurveyorEmployeeID), req.LocationUrl, createdBy)
		if err == nil {
			return p, nil
		}
		lastErr = err
		if !isDuplicateKeyErr(err) {
			return nil, err
		}
	}
	return nil, lastErr
}

// isDuplicateKeyErr يميّز خطأ تكرار المفتاح الفريد بـPostgres (SQLSTATE 23505).
func isDuplicateKeyErr(err error) bool {
	if err == nil {
		return false
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return strings.Contains(err.Error(), "23505")
}

// emptyToNil يحوّل "" لـnil حتى لا نحاول نخزن سترنغ فاضي بعمود مفتاح أجنبي
// (الواجهة ترسل "" لما المستخدم ما يختار موظف من القائمة المنسدلة).
func emptyToNil(v *string) *string {
	if v == nil || *v == "" {
		return nil
	}
	return v
}

func (s *ProjectService) Get(id string) (*model.Project, error) {
	return s.repo.GetByID(id)
}

func (s *ProjectService) Update(id string, req model.UpdateProjectRequest) (*model.Project, error) {
	req.ResponsibleEmployeeID = emptyToNil(req.ResponsibleEmployeeID)
	req.SurveyorEmployeeID = emptyToNil(req.SurveyorEmployeeID)
	return s.repo.Update(id, req)
}

func (s *ProjectService) Delete(id string) error {
	return s.repo.Delete(id)
}
