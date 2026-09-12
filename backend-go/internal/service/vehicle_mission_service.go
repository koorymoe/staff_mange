package service

import (
	"database/sql"
	"errors"
	"log"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type VehicleMissionService struct {
	repo        *repository.VehicleMissionRepository
	vehicleRepo *repository.VehicleRepository
}

func NewVehicleMissionService(repo *repository.VehicleMissionRepository, vehicleRepo *repository.VehicleRepository) *VehicleMissionService {
	return &VehicleMissionService{repo: repo, vehicleRepo: vehicleRepo}
}

// StartMission يبدأ مهمة جديدة لسيارة، بعد التأكد إنها مو بمهمة فعالة أصلاً
// (منع الحجز المزدوج لنفس السيارة بنفس الوقت).
func (s *VehicleMissionService) StartMission(actorEmployeeID string, req model.StartVehicleMissionRequest) (*model.VehicleMission, error) {
	if req.VehicleID == "" {
		return nil, errors.New("السيارة مطلوبة")
	}
	if req.Purpose == "" {
		return nil, errors.New("سبب المهمة مطلوب")
	}
	if req.Destination == "" {
		return nil, errors.New("الوجهة مطلوبة")
	}
	if req.StartOdometer < 0 {
		return nil, errors.New("عداد الكيلومترات لا يمكن أن يكون بالسالب")
	}

	driverID := actorEmployeeID
	if req.DriverID != nil && *req.DriverID != "" {
		driverID = *req.DriverID
	}

	if _, err := s.repo.GetActiveMissionForVehicle(req.VehicleID); err == nil {
		return nil, errors.New("هذه السيارة بمهمة حالياً، لا يمكن بدء مهمة جديدة لها إلا بعد إنهاء المهمة الحالية")
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	vehicle, err := s.vehicleRepo.Get(req.VehicleID)
	if err != nil {
		return nil, errors.New("السيارة غير موجودة")
	}
	if req.StartOdometer < vehicle.CurrentOdometer {
		// لا نمنع — قد يكون خطأ إدخال بسيط — بس نسجله بالسجلات كتحذير.
		log.Printf("تحذير: عداد بداية المهمة (%d) أقل من عداد السيارة المسجل حالياً (%d) للسيارة %s", req.StartOdometer, vehicle.CurrentOdometer, req.VehicleID)
	}

	return s.repo.StartMission(driverID, req.VehicleID, req.Purpose, req.Destination, req.StartOdometer, req.PassengerIDs)
}

func (s *VehicleMissionService) EndMission(missionID string, req model.EndVehicleMissionRequest) (*model.VehicleMission, error) {
	mission, err := s.repo.Get(missionID)
	if err != nil {
		return nil, errors.New("المهمة غير موجودة")
	}
	if mission.Status == "COMPLETED" {
		return nil, errors.New("هذه المهمة منتهية أصلاً")
	}
	if req.EndOdometer < mission.StartOdometer {
		return nil, errors.New("عداد النهاية لا يمكن أن يكون أقل من عداد البداية")
	}
	return s.repo.EndMission(missionID, req.EndOdometer, req.Notes)
}

func (s *VehicleMissionService) List(filters model.VehicleMissionFilters) ([]model.VehicleMission, error) {
	return s.repo.List(filters)
}

func (s *VehicleMissionService) Get(id string) (*model.VehicleMission, error) {
	return s.repo.Get(id)
}

func (s *VehicleMissionService) GetActiveMissionForVehicle(vehicleID string) (*model.VehicleMission, error) {
	m, err := s.repo.GetActiveMissionForVehicle(vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return m, err
}
