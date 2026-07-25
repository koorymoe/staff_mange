package service

import (
	"errors"
	"time"

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

func (s *VehicleService) Get(id string) (*model.Vehicle, error) { return s.repo.Get(id) }

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
	if req.Type != "FUEL" && req.Type != "CLEANING" && req.Type != "OIL_CHANGE" && req.Type != "MAINTENANCE" {
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

// ── VehicleIncidentAttachment ──

func (s *VehicleService) ListIncidentAttachments(incidentID string) ([]model.VehicleIncidentAttachment, error) {
	return s.repo.ListIncidentAttachments(incidentID)
}

func (s *VehicleService) CreateIncidentAttachment(incidentID string, req model.CreateVehicleIncidentAttachmentRequest) (*model.VehicleIncidentAttachment, error) {
	if req.URL == "" {
		return nil, errors.New("رابط/محتوى المرفق مطلوب")
	}
	if req.MediaType != "" && req.MediaType != "IMAGE" && req.MediaType != "VIDEO" {
		return nil, errors.New("نوع المرفق غير صحيح")
	}
	return s.repo.CreateIncidentAttachment(incidentID, req)
}

func (s *VehicleService) DeleteIncidentAttachment(incidentID, attachmentID string) error {
	return s.repo.DeleteIncidentAttachment(incidentID, attachmentID)
}

// ── VehiclePart (إطارات وبطاريات) ──

// معايير الاستحقاق القريب: خلال 2000 كم من الحد الأقصى بالمسافة، أو خلال 30
// يوم من الحد الأقصى بالزمن — أيهما تحقق يعتبر "قريب الاستحقاق".
const (
	PartDueSoonKmThreshold   = 2000
	PartDueSoonDaysThreshold = 30
)

// IsPartDueSoon يحسب فيما إذا كانت قطعة (إطار/بطارية) قريبة من الاستحقاق حسب
// المسافة المقطوعة منذ التركيب أو الزمن المنقضي، أيهما ينطبق.
func (s *VehicleService) IsPartDueSoon(part model.VehiclePart, currentOdometer int) bool {
	if part.ReplacedAt != nil {
		return false
	}
	if part.ExpectedLifespanKm != nil {
		remainingKm := part.InstalledOdometer + *part.ExpectedLifespanKm - currentOdometer
		if remainingKm <= PartDueSoonKmThreshold {
			return true
		}
	}
	if part.ExpectedLifespanMonths != nil {
		dueDate := part.InstalledAt.AddDate(0, *part.ExpectedLifespanMonths, 0)
		if time.Until(dueDate) <= PartDueSoonDaysThreshold*24*time.Hour {
			return true
		}
	}
	return false
}

func (s *VehicleService) ListParts(vehicleID string, currentOdometer int) ([]model.VehiclePart, error) {
	parts, err := s.repo.ListParts(vehicleID)
	if err != nil {
		return nil, err
	}
	for i := range parts {
		parts[i].DueSoon = s.IsPartDueSoon(parts[i], currentOdometer)
	}
	return parts, nil
}

func (s *VehicleService) CreatePart(vehicleID string, req model.CreateVehiclePartRequest) (*model.VehiclePart, error) {
	if req.PartType != "TIRE" && req.PartType != "BATTERY" {
		return nil, errors.New("نوع القطعة غير صحيح")
	}
	if req.InstalledOdometer < 0 {
		return nil, errors.New("عداد التركيب لا يمكن أن يكون بالسالب")
	}
	return s.repo.CreatePart(vehicleID, req)
}

func (s *VehicleService) MarkPartReplaced(id string) (*model.VehiclePart, error) {
	return s.repo.MarkPartReplaced(id)
}

// ── تنبيهات الصيانة والوثائق ──

func (s *VehicleService) VehicleAlerts() ([]model.VehicleAlert, error) {
	alerts := []model.VehicleAlert{}

	vehiclesByID, err := s.repo.VehiclesByID()
	if err != nil {
		return nil, err
	}

	// صيانة مستحقة قريباً (تاريخ أو عداد)
	logRows, err := s.repo.DueSoonLogs()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	for _, row := range logRows {
		v, ok := vehiclesByID[row.VehicleID]
		if !ok {
			continue
		}
		dueByDate := row.NextDueAt != nil && row.NextDueAt.Sub(now) <= 30*24*time.Hour
		dueByOdometer := row.NextDueOdometer != nil && (*row.NextDueOdometer-v.CurrentOdometer) <= 2000
		if dueByDate || dueByOdometer {
			severity := "warning"
			msg := "صيانة (" + row.Type + ") مستحقة قريباً"
			if dueByDate && row.NextDueAt.Before(now) {
				severity = "danger"
				msg = "صيانة (" + row.Type + ") متأخرة عن موعدها"
			}
			if dueByOdometer && *row.NextDueOdometer <= v.CurrentOdometer {
				severity = "danger"
				msg = "صيانة (" + row.Type + ") تجاوزت عداد الاستحقاق"
			}
			alerts = append(alerts, model.VehicleAlert{
				VehicleID:   v.ID,
				VehicleName: v.Name,
				AlertType:   "MAINTENANCE",
				Message:     msg,
				Severity:    severity,
			})
		}
	}

	// قطع (إطارات/بطاريات) قريبة الاستحقاق
	parts, err := s.repo.AllActivePartsInUse()
	if err != nil {
		return nil, err
	}
	for _, p := range parts {
		v, ok := vehiclesByID[p.VehicleID]
		if !ok {
			continue
		}
		if s.IsPartDueSoon(p, v.CurrentOdometer) {
			partLabel := "إطار"
			if p.PartType == "BATTERY" {
				partLabel = "بطارية"
			}
			alerts = append(alerts, model.VehicleAlert{
				VehicleID:   v.ID,
				VehicleName: v.Name,
				AlertType:   "PART",
				Message:     partLabel + " قريبة الاستحقاق للاستبدال",
				Severity:    "warning",
			})
		}
	}

	// وثائق قريبة الانتهاء (30 يوم)
	docs, err := s.repo.DocumentsExpiringWithin(30)
	if err != nil {
		return nil, err
	}
	for _, d := range docs {
		v, ok := vehiclesByID[d.VehicleID]
		if !ok {
			continue
		}
		severity := "warning"
		if d.ExpiryDate != nil && d.ExpiryDate.Before(now) {
			severity = "danger"
		}
		alerts = append(alerts, model.VehicleAlert{
			VehicleID:   v.ID,
			VehicleName: v.Name,
			AlertType:   "DOCUMENT",
			Message:     "وثيقة (" + d.DocumentType + ") قريبة/منتهية الصلاحية",
			Severity:    severity,
		})
	}

	return alerts, nil
}
