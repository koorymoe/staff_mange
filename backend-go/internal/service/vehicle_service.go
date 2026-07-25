package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type VehicleService struct {
	repo *repository.VehicleRepository
}

func NewVehicleService(repo *repository.VehicleRepository) *VehicleService {
	return &VehicleService{repo: repo}
}

func (s *VehicleService) List() ([]model.Vehicle, error) { return s.repo.List() }

func (s *VehicleService) Create(req model.CreateVehicleRequest) (*model.Vehicle, error) {
	if req.Name == "" || req.PlateNumber == "" {
		return nil, errors.New("اسم السيارة ورقم اللوحة مطلوبان")
	}
	return s.repo.Create(req)
}

func (s *VehicleService) Update(id string, req model.UpdateVehicleRequest) (*model.Vehicle, error) {
	if req.Name != nil && *req.Name == "" {
		return nil, errors.New("اسم السيارة لا يمكن أن يكون فارغاً")
	}
	if req.PlateNumber != nil && *req.PlateNumber == "" {
		return nil, errors.New("رقم اللوحة لا يمكن أن يكون فارغاً")
	}
	if req.CurrentOdometer != nil && *req.CurrentOdometer < 0 {
		return nil, errors.New("عداد الكيلومترات لا يمكن أن يكون بالسالب")
	}
	return s.repo.Update(id, req)
}

func (s *VehicleService) ListDocuments(vehicleID string) ([]model.VehicleDocument, error) {
	return s.repo.ListDocuments(vehicleID)
}

func (s *VehicleService) CreateDocument(vehicleID string, req model.CreateVehicleDocumentRequest) (*model.VehicleDocument, error) {
	if req.DocumentType == "" {
		return nil, errors.New("نوع الوثيقة مطلوب")
	}
	return s.repo.CreateDocument(vehicleID, req)
}

func (s *VehicleService) DeleteDocument(vehicleID, docID string) error {
	return s.repo.DeleteDocument(vehicleID, docID)
}

func (s *VehicleService) ListPhotos(vehicleID string) ([]model.VehiclePhoto, error) {
	return s.repo.ListPhotos(vehicleID)
}

func (s *VehicleService) CreatePhoto(vehicleID string, req model.CreateVehiclePhotoRequest) (*model.VehiclePhoto, error) {
	if req.URL == "" {
		return nil, errors.New("رابط/محتوى الصورة مطلوب")
	}
	return s.repo.CreatePhoto(vehicleID, req)
}

func (s *VehicleService) DeletePhoto(vehicleID, photoID string) error {
	return s.repo.DeletePhoto(vehicleID, photoID)
}

func (s *VehicleService) ListLogs(vehicleID string) ([]model.VehicleLog, error) {
	return s.repo.ListLogs(vehicleID)
}

func (s *VehicleService) CreateLog(vehicleID string, req model.CreateVehicleLogRequest, recordedByID string) (*model.VehicleLog, error) {
	if req.Type != "FUEL" && req.Type != "CLEANING" && req.Type != "OIL_CHANGE" {
		return nil, errors.New("نوع السجل غير صحيح")
	}
	return s.repo.CreateLog(vehicleID, req, recordedByID)
}

func (s *VehicleService) ListIncidents(vehicleID string) ([]model.VehicleIncident, error) {
	return s.repo.ListIncidents(vehicleID)
}

func (s *VehicleService) CreateIncident(vehicleID string, req model.CreateVehicleIncidentRequest, reportedByID string) (*model.VehicleIncident, error) {
	if req.Type != "FAULT" && req.Type != "DAMAGE" {
		return nil, errors.New("نوع الحادثة غير صحيح")
	}
	if req.Description == "" {
		return nil, errors.New("وصف العطل/الضرر مطلوب")
	}
	return s.repo.CreateIncident(vehicleID, req, reportedByID)
}

func (s *VehicleService) UpdateIncident(id string, req model.UpdateVehicleIncidentRequest) (*model.VehicleIncident, error) {
	return s.repo.UpdateIncident(id, req)
}

func (s *VehicleService) ListMonthlyStatus(vehicleID string) ([]model.VehicleMonthlyStatus, error) {
	return s.repo.ListMonthlyStatus(vehicleID)
}

func (s *VehicleService) SetMonthlyStatus(vehicleID string, req model.SetVehicleMonthlyStatusRequest, recordedByID string) (*model.VehicleMonthlyStatus, error) {
	if req.Month == "" {
		return nil, errors.New("الشهر مطلوب")
	}
	return s.repo.SetMonthlyStatus(vehicleID, req, recordedByID)
}

// قيمة نقطة الفني والحد الأعلى الشهري — نفس القيم الافتراضية بملف إكسل الشركة
// (200 د.ع للنقطة، 60000 د.ع حد أعلى شهري). للتعديل لاحقاً إذا تغيرت سياسة الشركة.
const (
	VehicleWashPointValue = 200.0
	VehicleWashMonthlyCap = 60000.0
)

func (s *VehicleService) CreateDailyRating(vehicleID string, req model.CreateVehicleDailyRatingRequest, recordedByID string) (*model.VehicleDailyRating, error) {
	if vehicleID == "" {
		return nil, errors.New("السيارة مطلوبة")
	}
	for _, tr := range req.TechnicianRatings {
		if tr.Score < 0 || tr.Score > 2 {
			return nil, errors.New("تقييم الفني لازم يكون بين 0 و 2")
		}
	}
	req.VehicleID = vehicleID
	return s.repo.CreateDailyRating(req, recordedByID)
}

func (s *VehicleService) ListDailyRatings(vehicleID, since string) ([]model.VehicleDailyRating, error) {
	return s.repo.ListDailyRatings(vehicleID, since)
}

func (s *VehicleService) VehicleScoreSummaries(since string) ([]model.VehicleScoreSummary, error) {
	return s.repo.VehicleScoreSummaries(since)
}

func (s *VehicleService) TechnicianWashSummaries(since string) ([]model.TechnicianWashSummary, error) {
	return s.repo.TechnicianWashSummaries(since, VehicleWashPointValue, VehicleWashMonthlyCap)
}
