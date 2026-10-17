package service

import (
	"errors"

	"github.com/google/uuid"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type TechShowcaseService struct {
	repo *repository.TechShowcaseRepository
}

func NewTechShowcaseService(repo *repository.TechShowcaseRepository) *TechShowcaseService {
	return &TechShowcaseService{repo: repo}
}

func (s *TechShowcaseService) List() ([]model.TechShowcaseItem, error) {
	return s.repo.List()
}

func (s *TechShowcaseService) Create(req model.CreateTechShowcaseItemRequest, employeeID string) (*model.TechShowcaseItem, error) {
	if req.Title == "" {
		return nil, errors.New("عنوان العمل مطلوب")
	}
	return s.repo.Create(uuid.NewString(), employeeID, req.Title, req.Description)
}

func (s *TechShowcaseService) AddMedia(id string, req model.AddTechShowcaseMediaRequest) (*model.TechShowcaseItem, error) {
	if len(req.MediaUrls) == 0 {
		return nil, errors.New("يرجى إرفاق صورة أو فيديو واحد على الأقل")
	}
	return s.repo.AddMedia(id, req.MediaUrls)
}
