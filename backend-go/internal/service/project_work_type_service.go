package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type ProjectWorkTypeService struct {
	repo *repository.ProjectWorkTypeRepository
}

func NewProjectWorkTypeService(repo *repository.ProjectWorkTypeRepository) *ProjectWorkTypeService {
	return &ProjectWorkTypeService{repo: repo}
}

func (s *ProjectWorkTypeService) List() ([]model.ProjectWorkType, error) {
	return s.repo.List()
}

func (s *ProjectWorkTypeService) Create(req model.CreateProjectWorkTypeRequest) (*model.ProjectWorkType, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("اسم نوع العمل مطلوب")
	}
	return s.repo.Create(uuid.NewString(), name)
}

func (s *ProjectWorkTypeService) Delete(id string) error {
	return s.repo.Delete(id)
}

// ListProjectCandidates يرجّع مرشحي المشروع مرتّبين بالتسلسل المطلوب:
// مهندسين ← تقنيين ← ليدريه ← فنيين ← إداريين ← مصممين.
func (s *ProjectWorkTypeService) ListProjectCandidates() ([]model.ProjectCandidate, error) {
	candidates, err := s.repo.ListProjectCandidates(model.EngineeringSkillNames)
	if err != nil {
		return nil, err
	}
	rank := make(map[string]int, len(model.ProjectCandidateGroupOrder))
	for i, g := range model.ProjectCandidateGroupOrder {
		rank[g] = i
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return rank[candidates[i].Group] < rank[candidates[j].Group]
	})
	return candidates, nil
}
