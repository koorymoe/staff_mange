package service

import (
	"errors"
	"fmt"
	"strings"

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

func (s *ProjectService) Create(req model.CreateProjectRequest) (*model.Project, error) {
	if req.Name == "" {
		return nil, errors.New("اسم المؤسسة مطلوب")
	}
	count, err := s.repo.CountAll()
	if err != nil {
		return nil, err
	}
	code := fmt.Sprintf("PRJ-%04d", count+1)
	priority := "عادي"
	if req.Priority != nil {
		priority = *req.Priority
	}
	return s.repo.Create(code, req.Name, req.Rep, req.Phone, req.Location, req.MapLatitude, req.MapLongitude, req.WorkType, req.RefPerson, priority, req.DeliveryDate, req.BookingID,
		emptyToNil(req.ResponsibleEmployeeID), emptyToNil(req.SurveyorEmployeeID))
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
