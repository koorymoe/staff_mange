package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type TeamInventoryCheckService struct {
	repo *repository.TeamInventoryCheckRepository
}

func NewTeamInventoryCheckService(repo *repository.TeamInventoryCheckRepository) *TeamInventoryCheckService {
	return &TeamInventoryCheckService{repo: repo}
}

func (s *TeamInventoryCheckService) ListTools() ([]model.TeamInventoryToolCatalog, error) {
	return s.repo.ListTools()
}

func (s *TeamInventoryCheckService) CreateTool(name string) (*model.TeamInventoryToolCatalog, error) {
	if name == "" {
		return nil, errors.New("اسم الأداة مطلوب")
	}
	return s.repo.CreateTool(name)
}

func (s *TeamInventoryCheckService) Create(leaderID string, req model.CreateTeamInventoryCheckRequest) (*model.TeamInventoryCheck, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("يجب تعبئة حالة الأدوات قبل حفظ الجرد")
	}
	for _, item := range req.Items {
		if item.ToolName == "" {
			return nil, errors.New("اسم الأداة مطلوب لكل عنصر")
		}
		if item.PersonRole != model.PersonRoleLeader && item.PersonRole != model.PersonRoleEmployee1 && item.PersonRole != model.PersonRoleEmployee2 {
			return nil, errors.New("دور غير صحيح لأحد عناصر الجرد")
		}
		if !item.Present {
			if item.Reason == nil || !model.ValidShortageReasons[*item.Reason] {
				return nil, errors.New("يجب اختيار سبب النقص لكل أداة غير متوفرة")
			}
		}
	}
	return s.repo.Create(leaderID, req)
}

func (s *TeamInventoryCheckService) List() ([]model.TeamInventoryCheck, error) {
	return s.repo.List()
}
