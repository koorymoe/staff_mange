package service

import (
	"database/sql"
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type VehicleMissionRatingService struct {
	repo        *repository.VehicleMissionRatingRepository
	missionRepo *repository.VehicleMissionRepository
}

func NewVehicleMissionRatingService(repo *repository.VehicleMissionRatingRepository, missionRepo *repository.VehicleMissionRepository) *VehicleMissionRatingService {
	return &VehicleMissionRatingService{repo: repo, missionRepo: missionRepo}
}

func validateRatingScore(v int, label string) error {
	if v < 1 || v > 5 {
		return errors.New(label + " يجب أن يكون بين 1 و 5")
	}
	return nil
}

func (s *VehicleMissionRatingService) CreateRating(missionID, ratedByID string, req model.CreateVehicleMissionRatingRequest) (*model.VehicleMissionRating, error) {
	mission, err := s.missionRepo.Get(missionID)
	if err != nil {
		return nil, errors.New("المهمة غير موجودة")
	}
	if mission.Status != "COMPLETED" {
		return nil, errors.New("لا يمكن تقييم مهمة لم تنتهِ بعد")
	}
	if err := validateRatingScore(req.Commitment, "الالتزام"); err != nil {
		return nil, err
	}
	if err := validateRatingScore(req.VehicleCare, "المحافظة على السيارة"); err != nil {
		return nil, err
	}
	if err := validateRatingScore(req.Driving, "القيادة"); err != nil {
		return nil, err
	}
	if err := validateRatingScore(req.Cleanliness, "النظافة"); err != nil {
		return nil, err
	}

	if _, err := s.repo.GetByMission(missionID); err == nil {
		return nil, errors.New("تم تقييم هذه المهمة مسبقاً")
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return s.repo.Create(missionID, ratedByID, req)
}

func (s *VehicleMissionRatingService) GetByMission(missionID string) (*model.VehicleMissionRating, error) {
	rating, err := s.repo.GetByMission(missionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return rating, err
}

func (s *VehicleMissionRatingService) GetDriverRatingSummary(employeeID string) (*model.DriverRatingSummary, error) {
	return s.repo.GetDriverRatingSummary(employeeID)
}
