package service

import (
	"errors"
	"sort"
	"strconv"
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

func (s *VehicleService) CreateLog(vehicleID string, req model.CreateVehicleLogRequest, recordedByID string) (*model.VehicleLogCreateResult, error) {
	if req.Type != "FUEL" && req.Type != "CLEANING" && req.Type != "OIL_CHANGE" && req.Type != "MAINTENANCE" {
		return nil, errors.New("نوع السجل غير صحيح")
	}
	// نحسب الشذوذ من تاريخ التعبئات القديم *قبل* إدخال السجل الجديد، وإلا صار السجل
	// الجديد يدخل بحساب متوسطه هو نفسه ويشوّه النتيجة.
	var anomaly *model.FuelAnomalyResult
	if req.Type == "FUEL" && req.Cost != nil {
		anomaly, _ = s.CheckFuelAnomaly(vehicleID, *req.Cost)
	}

	log, err := s.repo.CreateLog(vehicleID, req, recordedByID)
	if err != nil {
		return nil, err
	}
	result := &model.VehicleLogCreateResult{VehicleLog: log, FuelAnomaly: anomaly}
	return result, nil
}

// ── شذوذ تكلفة الوقود ──

// FuelAnomalyThresholdPercent نسبة الزيادة عن متوسط آخر 5 تعبئات وقود للسيارة التي تُعتبر شذوذاً.
const FuelAnomalyThresholdPercent = 40.0

// CheckFuelAnomaly يقارن تكلفة تعبئة وقود جديدة بمتوسط آخر 5 تعبئات للسيارة نفسها؛
// يعتبرها شذوذاً لو تجاوزت المتوسط بأكثر من FuelAnomalyThresholdPercent%.
func (s *VehicleService) CheckFuelAnomaly(vehicleID string, newCost float64) (*model.FuelAnomalyResult, error) {
	costs, err := s.repo.LastFuelLogCosts(vehicleID, 5)
	if err != nil {
		return nil, err
	}
	if len(costs) == 0 {
		return &model.FuelAnomalyResult{IsAnomaly: false, AverageCost: 0, NewCost: newCost}, nil
	}
	var sum float64
	for _, c := range costs {
		sum += c
	}
	avg := sum / float64(len(costs))
	if avg <= 0 {
		return &model.FuelAnomalyResult{IsAnomaly: false, AverageCost: avg, NewCost: newCost}, nil
	}
	percentAbove := ((newCost - avg) / avg) * 100
	return &model.FuelAnomalyResult{
		IsAnomaly:       percentAbove > FuelAnomalyThresholdPercent,
		AverageCost:     avg,
		NewCost:         newCost,
		PercentAboveAvg: percentAbove,
	}, nil
}

func (s *VehicleService) ListIncidents(vehicleID string) ([]model.VehicleIncident, error) {
	return s.repo.ListIncidents(vehicleID)
}

func (s *VehicleService) CreateIncident(vehicleID string, req model.CreateVehicleIncidentRequest, reportedByID string) (*model.VehicleIncident, error) {
	if req.Type != "FAULT" && req.Type != "DAMAGE" && req.Type != "ACCIDENT" {
		return nil, errors.New("نوع الحادثة غير صحيح")
	}
	if req.Description == "" {
		return nil, errors.New("وصف العطل/الضرر/الحادث مطلوب")
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

	// شذوذ تكلفة آخر تعبئة وقود لكل سيارة نشطة (مقارنة بمتوسط آخر 5 تعبئات)
	for vehicleID, v := range vehiclesByID {
		costs, err := s.repo.LastFuelLogCosts(vehicleID, 5)
		if err != nil || len(costs) < 2 {
			continue
		}
		latest := costs[0]
		anomaly, err := s.checkFuelAnomalyAgainstHistory(costs[1:], latest)
		if err != nil || anomaly == nil || !anomaly.IsAnomaly {
			continue
		}
		alerts = append(alerts, model.VehicleAlert{
			VehicleID:   v.ID,
			VehicleName: v.Name,
			AlertType:   "FUEL_ANOMALY",
			Message:     "آخر تعبئة وقود أعلى من المعدل المعتاد بنسبة تقارب " + formatPercent(anomaly.PercentAboveAvg) + "%",
			Severity:    "warning",
		})
	}

	return alerts, nil
}

// checkFuelAnomalyAgainstHistory نسخة مساعدة تحسب الشذوذ من قائمة تكاليف تاريخية جاهزة
// (بدون استعلام قاعدة بيانات إضافي) — تُستخدم من VehicleAlerts لتفادي جلب البيانات مرتين.
func (s *VehicleService) checkFuelAnomalyAgainstHistory(historyCosts []float64, newCost float64) (*model.FuelAnomalyResult, error) {
	if len(historyCosts) == 0 {
		return &model.FuelAnomalyResult{IsAnomaly: false, NewCost: newCost}, nil
	}
	var sum float64
	for _, c := range historyCosts {
		sum += c
	}
	avg := sum / float64(len(historyCosts))
	if avg <= 0 {
		return &model.FuelAnomalyResult{IsAnomaly: false, AverageCost: avg, NewCost: newCost}, nil
	}
	percentAbove := ((newCost - avg) / avg) * 100
	return &model.FuelAnomalyResult{
		IsAnomaly:       percentAbove > FuelAnomalyThresholdPercent,
		AverageCost:     avg,
		NewCost:         newCost,
		PercentAboveAvg: percentAbove,
	}, nil
}

func formatPercent(p float64) string {
	if p < 0 {
		p = 0
	}
	return strconv.Itoa(int(p + 0.5))
}

// ── ملخص مصاريف السيارة (وقود/صيانة/قطع/حوادث/تنظيف) ──

// ExpenseSummary يجمع مصاريف سيارة خلال فترة شهرية (YYYY-MM) أو سنوية (YYYY).
// لو ما انمرر أي منهما يُستخدم الشهر الحالي افتراضياً.
func (s *VehicleService) ExpenseSummary(vehicleID, month, year string) (*model.VehicleExpenseSummary, error) {
	var from, to time.Time
	var period string
	if year != "" {
		y, err := strconv.Atoi(year)
		if err != nil {
			return nil, errors.New("سنة غير صحيحة")
		}
		from = time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC)
		to = time.Date(y+1, 1, 1, 0, 0, 0, 0, time.UTC)
		period = year
	} else {
		m := month
		if m == "" {
			m = time.Now().Format("2006-01")
		}
		t, err := time.Parse("2006-01", m)
		if err != nil {
			return nil, errors.New("شهر غير صحيح (الصيغة المطلوبة YYYY-MM)")
		}
		from = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		to = from.AddDate(0, 1, 0)
		period = m
	}

	fuelCost, err := s.repo.SumLogCostByType(vehicleID, "FUEL", from, to)
	if err != nil {
		return nil, err
	}
	oilCost, err := s.repo.SumLogCostByType(vehicleID, "OIL_CHANGE", from, to)
	if err != nil {
		return nil, err
	}
	maintCost, err := s.repo.SumLogCostByType(vehicleID, "MAINTENANCE", from, to)
	if err != nil {
		return nil, err
	}
	cleaningCost, err := s.repo.SumLogCostByType(vehicleID, "CLEANING", from, to)
	if err != nil {
		return nil, err
	}
	incidentCost, err := s.repo.SumIncidentCost(vehicleID, from, to)
	if err != nil {
		return nil, err
	}
	partsCost, err := s.repo.SumPartsCost(vehicleID, from, to)
	if err != nil {
		return nil, err
	}

	maintenanceTotal := oilCost + maintCost
	total := fuelCost + maintenanceTotal + partsCost + incidentCost + cleaningCost

	summary := &model.VehicleExpenseSummary{
		VehicleID:       vehicleID,
		Period:          period,
		FuelCost:        fuelCost,
		MaintenanceCost: maintenanceTotal,
		PartsCost:       partsCost,
		IncidentCost:    incidentCost,
		CleaningCost:    cleaningCost,
		TotalCost:       total,
	}

	// تقدير المسافة المقطوعة بالفترة: أولاً من فرق قراءات العداد بسجلات الفترة،
	// وإلا من مجموع مسافات المهمات المكتملة بنفس الفترة.
	minOdo, maxOdo, err := s.repo.OdometerRangeInPeriod(vehicleID, from, to)
	if err == nil && minOdo != nil && maxOdo != nil && *maxOdo > *minOdo {
		distance := *maxOdo - *minOdo
		summary.DistanceKm = &distance
	} else {
		if dist, err := s.repo.DistanceFromMissionsInPeriod(vehicleID, from, to); err == nil && dist != nil && *dist > 0 {
			summary.DistanceKm = dist
		}
	}
	if summary.DistanceKm != nil && *summary.DistanceKm > 0 && total > 0 {
		avg := total / float64(*summary.DistanceKm)
		summary.AvgCostPerKm = &avg
	}

	return summary, nil
}

// ── لوحة التحكم الشاملة للأسطول ──

// FleetDashboard يجمع كل مؤشرات لوحة التحكم الشاملة (GET /api/vehicles/dashboard)
// بنداء واحد، معتمداً على الاستعلامات والدوال الموجودة أصلاً (التنبيهات وملخص المصاريف)
// بدل تكرار منطقها.
func (s *VehicleService) FleetDashboard() (*model.FleetDashboardSummary, error) {
	now := time.Now()
	period := now.Format("2006-01")
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	vehicles, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	maintenanceIDs, err := s.repo.VehicleIDsWithOpenIncidentType("FAULT")
	if err != nil {
		return nil, err
	}
	inMaintenance := map[string]bool{}
	for _, id := range maintenanceIDs {
		inMaintenance[id] = true
	}

	onMissionIDs, err := s.repo.VehicleIDsWithInProgressMission()
	if err != nil {
		return nil, err
	}
	onMission := map[string]bool{}
	for _, id := range onMissionIDs {
		onMission[id] = true
	}

	alerts, err := s.VehicleAlerts()
	if err != nil {
		return nil, err
	}
	needsService := map[string]bool{}
	expiringDocs := map[string]bool{}
	for _, a := range alerts {
		switch a.AlertType {
		case "MAINTENANCE", "PART":
			needsService[a.VehicleID] = true
		case "DOCUMENT":
			expiringDocs[a.VehicleID] = true
		}
	}

	summary := &model.FleetDashboardSummary{
		Period:            period,
		TotalVehicles:     len(vehicles),
		Alerts:            alerts,
		NeedsServiceCount: len(needsService),
		ExpiringDocsCount: len(expiringDocs),
		VehicleExpenses:   []model.VehicleExpenseRow{},
		TopByUsage:        []model.VehicleUsageRankRow{},
		TopByCost:         []model.VehicleUsageRankRow{},
	}

	activeCount := 0
	maintenanceCount := 0
	onMissionCount := 0

	usageRows, err := s.repo.MissionUsageInPeriod(from, to)
	if err != nil {
		return nil, err
	}
	usageByVehicle := map[string]repository.VehicleUsageRow{}
	for _, u := range usageRows {
		usageByVehicle[u.VehicleID] = u
	}

	var fleetFuelCost, fleetTotalCost float64
	rankRows := make([]model.VehicleUsageRankRow, 0, len(vehicles))

	for _, v := range vehicles {
		if inMaintenance[v.ID] {
			maintenanceCount++
		} else if v.IsActive {
			activeCount++
		}
		if onMission[v.ID] {
			onMissionCount++
		}

		expSummary, err := s.ExpenseSummary(v.ID, period, "")
		if err != nil {
			continue
		}
		fleetFuelCost += expSummary.FuelCost
		fleetTotalCost += expSummary.TotalCost

		if expSummary.TotalCost > 0 {
			summary.VehicleExpenses = append(summary.VehicleExpenses, model.VehicleExpenseRow{
				VehicleID:   v.ID,
				VehicleName: v.Name,
				PlateNumber: v.PlateNumber,
				TotalCost:   expSummary.TotalCost,
				FuelCost:    expSummary.FuelCost,
			})
		}

		usage := usageByVehicle[v.ID]
		if usage.MissionCount > 0 || expSummary.TotalCost > 0 {
			rankRows = append(rankRows, model.VehicleUsageRankRow{
				VehicleID:    v.ID,
				VehicleName:  v.Name,
				PlateNumber:  v.PlateNumber,
				MissionCount: usage.MissionCount,
				DistanceKm:   usage.DistanceKm,
				TotalCost:    expSummary.TotalCost,
			})
		}
	}

	summary.ActiveVehiclesCount = activeCount
	summary.InMaintenanceCount = maintenanceCount
	summary.OnMissionCount = onMissionCount
	summary.FleetFuelCostThisMonth = fleetFuelCost
	summary.FleetTotalCostThisMonth = fleetTotalCost

	sort.Slice(summary.VehicleExpenses, func(i, j int) bool {
		return summary.VehicleExpenses[i].TotalCost > summary.VehicleExpenses[j].TotalCost
	})

	byUsage := make([]model.VehicleUsageRankRow, len(rankRows))
	copy(byUsage, rankRows)
	sort.Slice(byUsage, func(i, j int) bool {
		if byUsage[i].MissionCount != byUsage[j].MissionCount {
			return byUsage[i].MissionCount > byUsage[j].MissionCount
		}
		return byUsage[i].DistanceKm > byUsage[j].DistanceKm
	})
	if len(byUsage) > 5 {
		byUsage = byUsage[:5]
	}
	summary.TopByUsage = byUsage

	byCost := make([]model.VehicleUsageRankRow, len(rankRows))
	copy(byCost, rankRows)
	sort.Slice(byCost, func(i, j int) bool {
		return byCost[i].TotalCost > byCost[j].TotalCost
	})
	if len(byCost) > 5 {
		byCost = byCost[:5]
	}
	summary.TopByCost = byCost

	return summary, nil
}
