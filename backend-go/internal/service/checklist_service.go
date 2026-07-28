package service

import (
	"errors"

	"github.com/google/uuid"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type ChecklistService struct {
	repo *repository.ChecklistRepository
}

func NewChecklistService(repo *repository.ChecklistRepository) *ChecklistService {
	return &ChecklistService{repo: repo}
}

func (s *ChecklistService) List() ([]model.ProjectChecklist, error) {
	return s.repo.List()
}

func (s *ChecklistService) Create(req model.CreateChecklistRequest, createdByID string) (*model.ProjectChecklist, error) {
	if req.Title == "" {
		return nil, errors.New("عنوان الكشف مطلوب")
	}
	return s.repo.Create(uuid.NewString(), req.ProjectID, req.Title, createdByID)
}

func (s *ChecklistService) AddPhotos(id string, req model.AddChecklistPhotosRequest) (*model.ProjectChecklist, error) {
	if len(req.PhotoUrls) == 0 {
		return nil, errors.New("يرجى إرفاق صورة واحدة على الأقل")
	}
	return s.repo.AddPhotos(id, req.PhotoUrls)
}
