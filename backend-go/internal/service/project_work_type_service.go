package service

import (
	"fmt"
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
