package repository

import (
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type VehicleRepository struct {
	db *sqlx.DB
}

func NewVehicleRepository(db *sqlx.DB) *VehicleRepository {
	return &VehicleRepository{db: db}
}

func (r *VehicleRepository) loadEmployeeBrief(id *string) *model.EmployeeBrief {
	if id == nil {
		return nil
	}
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name FROM "Employee" WHERE id = $1`, *id); err != nil {
		return nil
	}
	return &brief
}

// ── Vehicle ──

func (r *VehicleRepository) List() ([]model.Vehicle, error) {
	vehicles := []model.Vehicle{}
	err := r.db.Select(&vehicles, `SELECT * FROM "Vehicle" ORDER BY "createdAt" DESC`)
	return vehicles, err
}

func (r *VehicleRepository) Create(req model.CreateVehicleRequest) (*model.Vehicle, error) {
	var v model.Vehicle
	err := r.db.Get(&v, `
		INSERT INTO "Vehicle" (id, name, "plateNumber", color, type)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4)
		RETURNING *
	`, req.Name, req.PlateNumber, req.Color, req.Type)
	return &v, err
}

func (r *VehicleRepository) Get(id string) (*model.Vehicle, error) {
	var v model.Vehicle
	err := r.db.Get(&v, `SELECT * FROM "Vehicle" WHERE id = $1`, id)
	return &v, err
}

func (r *VehicleRepository) Update(id string, req model.UpdateVehicleRequest) (*model.Vehicle, error) {
	var v model.Vehicle
	err := r.db.Get(&v, `
		UPDATE "Vehicle" SET
			name = COALESCE($2, name),
			"plateNumber" = COALESCE($3, "plateNumber"),
			color = COALESCE($4, color),
			type = COALESCE($5, type),
			model = COALESCE($6, model),
			year = COALESCE($7, year),
			"chassisNumber" = COALESCE($8, "chassisNumber"),
			"engineNumber" = COALESCE($9, "engineNumber"),
			"fuelType" = COALESCE($10, "fuelType"),
			"currentOdometer" = COALESCE($11, "currentOdometer"),
			condition = COALESCE($12, condition),
			"isActive" = COALESCE($13, "isActive")
		WHERE id = $1
		RETURNING *
	`, id, req.Name, req.PlateNumber, req.Color, req.Type, req.Model, req.Year, req.ChassisNumber,
		req.EngineNumber, req.FuelType, req.CurrentOdometer, req.Condition, req.IsActive)
	return &v, err
}

// Delete يحذف السيارة نهائياً. أغلب الجداول المرتبطة (وثائق، صور، سجلات،
// حوادث، حالة شهرية، تقييمات، قطع، أدوات) عندها ON DELETE CASCADE بالمخطط،
// فتنحذف تلقائياً. جداول المهام والحجوزات (VehicleMission/VehicleBooking)
// بالعمد ما عندها cascade — نخليها تفشل بخطأ FK واضح بدل ما نحذف سجل تاريخي
// مهم بصمت (سجل مهمة/حجز فعلي صار للسيارة).
func (r *VehicleRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "Vehicle" WHERE id = $1`, id)
	return err
}

// ── VehicleDocument ──

func (r *VehicleRepository) ListDocuments(vehicleID string) ([]model.VehicleDocument, error) {
	docs := []model.VehicleDocument{}
	err := r.db.Select(&docs, `SELECT * FROM "VehicleDocument" WHERE "vehicleId" = $1 ORDER BY "createdAt" DESC`, vehicleID)
	return docs, err
}

func (r *VehicleRepository) CreateDocument(vehicleID string, req model.CreateVehicleDocumentRequest) (*model.VehicleDocument, error) {
	var d model.VehicleDocument
	err := r.db.Get(&d, `
		INSERT INTO "VehicleDocument" (id, "vehicleId", "documentType", "documentNumber", "issueDate", "expiryDate", "fileUrl", notes)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4::timestamp, $5::timestamp, $6, $7)
		RETURNING *
	`, vehicleID, req.DocumentType, req.DocumentNumber, req.IssueDate, req.ExpiryDate, req.FileURL, req.Notes)
	return &d, err
}

func (r *VehicleRepository) DeleteDocument(vehicleID, docID string) error {
	_, err := r.db.Exec(`DELETE FROM "VehicleDocument" WHERE id = $1 AND "vehicleId" = $2`, docID, vehicleID)
	return err
}

// ── VehiclePhoto ──

func (r *VehicleRepository) ListPhotos(vehicleID string) ([]model.VehiclePhoto, error) {
	photos := []model.VehiclePhoto{}
	err := r.db.Select(&photos, `SELECT * FROM "VehiclePhoto" WHERE "vehicleId" = $1 ORDER BY "createdAt" DESC`, vehicleID)
	return photos, err
}

func (r *VehicleRepository) CreatePhoto(vehicleID string, req model.CreateVehiclePhotoRequest) (*model.VehiclePhoto, error) {
	var p model.VehiclePhoto
	err := r.db.Get(&p, `
		INSERT INTO "VehiclePhoto" (id, "vehicleId", url, caption)
		VALUES (gen_random_uuid()::text, $1, $2, $3)
		RETURNING *
	`, vehicleID, req.URL, req.Caption)
	return &p, err
}

func (r *VehicleRepository) DeletePhoto(vehicleID, photoID string) error {
	_, err := r.db.Exec(`DELETE FROM "VehiclePhoto" WHERE id = $1 AND "vehicleId" = $2`, photoID, vehicleID)
	return err
}

// ── VehicleLog (fuel / cleaning / oil change) ──

func (r *VehicleRepository) ListLogs(vehicleID string) ([]model.VehicleLog, error) {
	logs := []model.VehicleLog{}
	if err := r.db.Select(&logs, `SELECT * FROM "VehicleLog" WHERE "vehicleId" = $1 ORDER BY "performedAt" DESC`, vehicleID); err != nil {
		return nil, err
	}
	for i := range logs {
		logs[i].RecordedBy = r.loadEmployeeBrief(logs[i].RecordedByID)
	}
	return logs, nil
}

func (r *VehicleRepository) CreateLog(vehicleID string, req model.CreateVehicleLogRequest, recordedByID string) (*model.VehicleLog, error) {
	var l model.VehicleLog
	err := r.db.Get(&l, `
		INSERT INTO "VehicleLog" (id, "vehicleId", type, "performedAt", "nextDueAt", "nextDueOdometer", odometer, cost, notes, "recordedById")
		VALUES (gen_random_uuid()::text, $1, $2, COALESCE($3::timestamp, now()), $4::timestamp, $5, $6, $7, $8, $9)
		RETURNING *
	`, vehicleID, req.Type, req.PerformedAt, req.NextDueAt, req.NextDueOdometer, req.Odometer, req.Cost, req.Notes, recordedByID)
	if err != nil {
		return nil, err
	}
	l.RecordedBy = r.loadEmployeeBrief(l.RecordedByID)
	return &l, nil
}

// ── VehicleIncident (fault / damage) ──

func (r *VehicleRepository) ListIncidents(vehicleID string) ([]model.VehicleIncident, error) {
	incidents := []model.VehicleIncident{}
	if err := r.db.Select(&incidents, `SELECT * FROM "VehicleIncident" WHERE "vehicleId" = $1 ORDER BY "createdAt" DESC`, vehicleID); err != nil {
		return nil, err
	}
	for i := range incidents {
		incidents[i].ResponsibleEmployee = r.loadEmployeeBrief(incidents[i].ResponsibleEmployeeID)
		incidents[i].ReportedBy = r.loadEmployeeBrief(incidents[i].ReportedByID)
		incidents[i].Driver = r.loadEmployeeBrief(incidents[i].DriverID)
	}
	return incidents, nil
}

func (r *VehicleRepository) CreateIncident(vehicleID string, req model.CreateVehicleIncidentRequest, reportedByID string) (*model.VehicleIncident, error) {
	var inc model.VehicleIncident
	err := r.db.Get(&inc, `
		INSERT INTO "VehicleIncident" (
			id, "vehicleId", type, description, "responsibleEmployeeId", cost, "reportedById",
			location, "driverId", "peoplePresent", "policeReportNumber", "repairCost"
		)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING *
	`, vehicleID, req.Type, req.Description, req.ResponsibleEmployeeID, req.Cost, reportedByID,
		req.Location, req.DriverID, req.PeoplePresent, req.PoliceReportNumber, req.RepairCost)
	if err != nil {
		return nil, err
	}
	inc.ResponsibleEmployee = r.loadEmployeeBrief(inc.ResponsibleEmployeeID)
	inc.ReportedBy = r.loadEmployeeBrief(inc.ReportedByID)
	inc.Driver = r.loadEmployeeBrief(inc.DriverID)
	return &inc, nil
}

func (r *VehicleRepository) UpdateIncident(id string, req model.UpdateVehicleIncidentRequest) (*model.VehicleIncident, error) {
	var resolvedAtExpr string
	if req.Status != nil && *req.Status == "RESOLVED" {
		resolvedAtExpr = "now()"
	} else {
		resolvedAtExpr = `"resolvedAt"`
	}
	var inc model.VehicleIncident
	err := r.db.Get(&inc, `
		UPDATE "VehicleIncident" SET
			status = COALESCE($2, status),
			cost = COALESCE($3, cost),
			"resolvedAt" = `+resolvedAtExpr+`
		WHERE id = $1
		RETURNING *
	`, id, req.Status, req.Cost)
	if err != nil {
		return nil, err
	}
	inc.ResponsibleEmployee = r.loadEmployeeBrief(inc.ResponsibleEmployeeID)
	inc.ReportedBy = r.loadEmployeeBrief(inc.ReportedByID)
	inc.Driver = r.loadEmployeeBrief(inc.DriverID)
	return &inc, nil
}

// ── VehicleMonthlyStatus ──

func (r *VehicleRepository) ListMonthlyStatus(vehicleID string) ([]model.VehicleMonthlyStatus, error) {
	statuses := []model.VehicleMonthlyStatus{}
	err := r.db.Select(&statuses, `SELECT * FROM "VehicleMonthlyStatus" WHERE "vehicleId" = $1 ORDER BY month DESC`, vehicleID)
	return statuses, err
}

func (r *VehicleRepository) SetMonthlyStatus(vehicleID string, req model.SetVehicleMonthlyStatusRequest, recordedByID string) (*model.VehicleMonthlyStatus, error) {
	var s model.VehicleMonthlyStatus
	err := r.db.Get(&s, `
		INSERT INTO "VehicleMonthlyStatus" (id, "vehicleId", month, "hasIssue", "issueDescription", resolved, notes, "recordedById")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT ("vehicleId", month) DO UPDATE SET
			"hasIssue" = EXCLUDED."hasIssue",
			"issueDescription" = EXCLUDED."issueDescription",
			resolved = EXCLUDED.resolved,
			notes = EXCLUDED.notes,
			"recordedById" = EXCLUDED."recordedById"
		RETURNING *
	`, vehicleID, req.Month, req.HasIssue, req.IssueDescription, req.Resolved, req.Notes, recordedByID)
	return &s, err
}

// ── VehicleDailyRating (تقييم يومي للسيارة + جودة غسيل الفنيين) ──

var vehicleRatingWeights = []struct {
	get    func(*model.VehicleDailyRating) *int
	weight float64
}{
	{func(r *model.VehicleDailyRating) *int { return r.Wash }, 0.08},
	{func(r *model.VehicleDailyRating) *int { return r.ExteriorClean }, 0.07},
	{func(r *model.VehicleDailyRating) *int { return r.ExteriorCondition }, 0.10},
	{func(r *model.VehicleDailyRating) *int { return r.TireCondition }, 0.10},
	{func(r *model.VehicleDailyRating) *int { return r.GlassClean }, 0.05},
	{func(r *model.VehicleDailyRating) *int { return r.LightsCondition }, 0.05},
	{func(r *model.VehicleDailyRating) *int { return r.TechnicalFaults }, 0.25},
	{func(r *model.VehicleDailyRating) *int { return r.InteriorClean }, 0.10},
	{func(r *model.VehicleDailyRating) *int { return r.SeatsCondition }, 0.07},
	{func(r *model.VehicleDailyRating) *int { return r.InteriorDirt }, 0.08},
	{func(r *model.VehicleDailyRating) *int { return r.Smell }, 0.05},
}

// computeWeightedScore يحسب النتيجة الموزونة % لتقييم سيارة — نفس معادلة ملف
// الإكسل: الخلايا الفارغة ما تُحتسب لا بالبسط ولا بالمقام.
func computeWeightedScore(r *model.VehicleDailyRating) *float64 {
	var num, den float64
	any := false
	for _, item := range vehicleRatingWeights {
		if v := item.get(r); v != nil {
			num += float64(*v) * item.weight
			den += 4 * item.weight
			any = true
		}
	}
	if !any || den == 0 {
		return nil
	}
	score := num / den
	return &score
}

func (r *VehicleRepository) CreateDailyRating(req model.CreateVehicleDailyRatingRequest, recordedByID string) (*model.VehicleDailyRating, error) {
	ratedDate := "CURRENT_DATE"
	var args []any
	args = append(args, req.VehicleID, req.Wash, req.ExteriorClean, req.ExteriorCondition, req.TireCondition,
		req.GlassClean, req.LightsCondition, req.TechnicalFaults, req.FaultDescription, req.InteriorClean,
		req.SeatsCondition, req.InteriorDirt, req.Smell, req.Notes, recordedByID)
	dateClause := ratedDate
	if req.RatedDate != nil && *req.RatedDate != "" {
		dateClause = "$16"
		args = append(args, *req.RatedDate)
	}

	var rating model.VehicleDailyRating
	err := r.db.Get(&rating, `
		INSERT INTO "VehicleDailyRating" (
			id, "vehicleId", "ratedDate", wash, "exteriorClean", "exteriorCondition", "tireCondition",
			"glassClean", "lightsCondition", "technicalFaults", "faultDescription", "interiorClean",
			"seatsCondition", "interiorDirt", smell, notes, "recordedById"
		)
		VALUES (gen_random_uuid()::text, $1, `+dateClause+`, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING *
	`, args...)
	if err != nil {
		return nil, err
	}

	for _, tr := range req.TechnicianRatings {
		var wr model.VehicleWashRating
		if err := r.db.Get(&wr, `
			INSERT INTO "VehicleWashRating" (id, "dailyRatingId", "employeeId", score)
			VALUES (gen_random_uuid()::text, $1, $2, $3)
			RETURNING *
		`, rating.ID, tr.EmployeeID, tr.Score); err == nil {
			wr.Employee = r.loadEmployeeBrief(&wr.EmployeeID)
			rating.WashRatings = append(rating.WashRatings, wr)
		}
	}

	rating.RecordedBy = r.loadEmployeeBrief(rating.RecordedByID)
	rating.WeightedScore = computeWeightedScore(&rating)
	return &rating, nil
}

func (r *VehicleRepository) ListDailyRatings(vehicleID string, since string) ([]model.VehicleDailyRating, error) {
	ratings := []model.VehicleDailyRating{}
	err := r.db.Select(&ratings, `
		SELECT * FROM "VehicleDailyRating"
		WHERE "vehicleId" = $1 AND "ratedDate" >= $2::date
		ORDER BY "ratedDate" DESC
	`, vehicleID, since)
	if err != nil {
		return nil, err
	}
	for i := range ratings {
		ratings[i].RecordedBy = r.loadEmployeeBrief(ratings[i].RecordedByID)
		ratings[i].WeightedScore = computeWeightedScore(&ratings[i])
		washRatings := []model.VehicleWashRating{}
		if err := r.db.Select(&washRatings, `SELECT * FROM "VehicleWashRating" WHERE "dailyRatingId" = $1`, ratings[i].ID); err == nil {
			for j := range washRatings {
				washRatings[j].Employee = r.loadEmployeeBrief(&washRatings[j].EmployeeID)
			}
			ratings[i].WashRatings = washRatings
		}
	}
	return ratings, nil
}

// VehicleScoreSummaries متوسط النتيجة الموزونة لكل سيارة خلال فترة — للوحة تذكير المراقب.
func (r *VehicleRepository) VehicleScoreSummaries(since string) ([]model.VehicleScoreSummary, error) {
	ratings := []model.VehicleDailyRating{}
	if err := r.db.Select(&ratings, `SELECT * FROM "VehicleDailyRating" WHERE "ratedDate" >= $1::date`, since); err != nil {
		return nil, err
	}
	type acc struct {
		sum   float64
		count int
	}
	byVehicle := map[string]*acc{}
	for i := range ratings {
		score := computeWeightedScore(&ratings[i])
		if score == nil {
			continue
		}
		a, ok := byVehicle[ratings[i].VehicleID]
		if !ok {
			a = &acc{}
			byVehicle[ratings[i].VehicleID] = a
		}
		a.sum += *score
		a.count++
	}

	vehicles := []model.Vehicle{}
	if err := r.db.Select(&vehicles, `SELECT * FROM "Vehicle"`); err != nil {
		return nil, err
	}
	summaries := []model.VehicleScoreSummary{}
	for _, v := range vehicles {
		a, ok := byVehicle[v.ID]
		if !ok || a.count == 0 {
			continue
		}
		summaries = append(summaries, model.VehicleScoreSummary{
			VehicleID:    v.ID,
			VehicleName:  v.Name,
			RatingsCount: a.count,
			AverageScore: a.sum / float64(a.count),
		})
	}
	return summaries, nil
}

// TechnicianWashSummaries مجموع نقاط الغسيل لكل فني خلال فترة، والراتب المقترح
// (نقطة الفني × قيمة النقطة، بحد أعلى شهري) — تذكير للمراقب بدون أي تعديل تلقائي.
func (r *VehicleRepository) TechnicianWashSummaries(since string, pointValue, monthlyCap float64) ([]model.TechnicianWashSummary, error) {
	type row struct {
		EmployeeID     string `db:"employeeId"`
		EmployeeName   string `db:"employeeName"`
		TotalPoints    int    `db:"totalPoints"`
		VehiclesWashed int    `db:"vehiclesWashed"`
	}
	rows := []row{}
	err := r.db.Select(&rows, `
		SELECT e.id AS "employeeId", e.name AS "employeeName",
			COALESCE(SUM(w.score), 0) AS "totalPoints",
			COUNT(DISTINCT w."dailyRatingId") AS "vehiclesWashed"
		FROM "VehicleWashRating" w
		JOIN "Employee" e ON e.id = w."employeeId"
		JOIN "VehicleDailyRating" d ON d.id = w."dailyRatingId"
		WHERE d."ratedDate" >= $1::date
		GROUP BY e.id, e.name
		ORDER BY "totalPoints" DESC
	`, since)
	if err != nil {
		return nil, err
	}
	summaries := make([]model.TechnicianWashSummary, len(rows))
	for i, rr := range rows {
		wage := float64(rr.TotalPoints) * pointValue
		if wage > monthlyCap {
			wage = monthlyCap
		}
		summaries[i] = model.TechnicianWashSummary{
			EmployeeID:     rr.EmployeeID,
			EmployeeName:   rr.EmployeeName,
			VehiclesWashed: rr.VehiclesWashed,
			TotalPoints:    rr.TotalPoints,
			SuggestedWage:  wage,
			MonthlyCap:     monthlyCap,
		}
	}
	return summaries, nil
}

// ── VehicleIncidentAttachment ──

func (r *VehicleRepository) ListIncidentAttachments(incidentID string) ([]model.VehicleIncidentAttachment, error) {
	attachments := []model.VehicleIncidentAttachment{}
	err := r.db.Select(&attachments, `SELECT * FROM "VehicleIncidentAttachment" WHERE "incidentId" = $1 ORDER BY "createdAt" DESC`, incidentID)
	return attachments, err
}

func (r *VehicleRepository) CreateIncidentAttachment(incidentID string, req model.CreateVehicleIncidentAttachmentRequest) (*model.VehicleIncidentAttachment, error) {
	mediaType := req.MediaType
	if mediaType == "" {
		mediaType = "IMAGE"
	}
	var a model.VehicleIncidentAttachment
	err := r.db.Get(&a, `
		INSERT INTO "VehicleIncidentAttachment" (id, "incidentId", url, "mediaType")
		VALUES (gen_random_uuid()::text, $1, $2, $3)
		RETURNING *
	`, incidentID, req.URL, mediaType)
	return &a, err
}

func (r *VehicleRepository) DeleteIncidentAttachment(incidentID, attachmentID string) error {
	_, err := r.db.Exec(`DELETE FROM "VehicleIncidentAttachment" WHERE id = $1 AND "incidentId" = $2`, attachmentID, incidentID)
	return err
}

// ── VehiclePart (إطارات وبطاريات) ──

func (r *VehicleRepository) ListParts(vehicleID string) ([]model.VehiclePart, error) {
	parts := []model.VehiclePart{}
	err := r.db.Select(&parts, `SELECT * FROM "VehiclePart" WHERE "vehicleId" = $1 ORDER BY "installedAt" DESC`, vehicleID)
	return parts, err
}

func (r *VehicleRepository) GetPart(id string) (*model.VehiclePart, error) {
	var p model.VehiclePart
	err := r.db.Get(&p, `SELECT * FROM "VehiclePart" WHERE id = $1`, id)
	return &p, err
}

func (r *VehicleRepository) CreatePart(vehicleID string, req model.CreateVehiclePartRequest) (*model.VehiclePart, error) {
	var p model.VehiclePart
	err := r.db.Get(&p, `
		INSERT INTO "VehiclePart" (id, "vehicleId", "partType", "installedAt", "installedOdometer", "expectedLifespanKm", "expectedLifespanMonths", notes, cost)
		VALUES (gen_random_uuid()::text, $1, $2, COALESCE($3::timestamp, now()), $4, $5, $6, $7, $8)
		RETURNING *
	`, vehicleID, req.PartType, req.InstalledAt, req.InstalledOdometer, req.ExpectedLifespanKm, req.ExpectedLifespanMonths, req.Notes, req.Cost)
	return &p, err
}

func (r *VehicleRepository) MarkPartReplaced(id string) (*model.VehiclePart, error) {
	var p model.VehiclePart
	err := r.db.Get(&p, `UPDATE "VehiclePart" SET "replacedAt" = now() WHERE id = $1 RETURNING *`, id)
	return &p, err
}

// ── Alerts (استحقاق صيانة/قطع/وثائق قريبة) ──

func (r *VehicleRepository) VehiclesByID() (map[string]model.Vehicle, error) {
	vehicles := []model.Vehicle{}
	if err := r.db.Select(&vehicles, `SELECT * FROM "Vehicle" WHERE "isActive" = true`); err != nil {
		return nil, err
	}
	m := make(map[string]model.Vehicle, len(vehicles))
	for _, v := range vehicles {
		m[v.ID] = v
	}
	return m, nil
}

type dueMaintenanceLogRow struct {
	VehicleID       string     `db:"vehicleId"`
	Type            string     `db:"type"`
	NextDueAt       *time.Time `db:"nextDueAt"`
	NextDueOdometer *int       `db:"nextDueOdometer"`
}

// DueSoonLogs يجيب أحدث سجل لكل (سيارة، نوع) له استحقاق تاريخ أو عداد قادم لم يُتجاوز بعد.
func (r *VehicleRepository) DueSoonLogs() ([]dueMaintenanceLogRow, error) {
	rows := []dueMaintenanceLogRow{}
	err := r.db.Select(&rows, `
		SELECT DISTINCT ON ("vehicleId", type) "vehicleId", type, "nextDueAt", "nextDueOdometer"
		FROM "VehicleLog"
		WHERE "nextDueAt" IS NOT NULL OR "nextDueOdometer" IS NOT NULL
		ORDER BY "vehicleId", type, "performedAt" DESC
	`)
	return rows, err
}

func (r *VehicleRepository) AllActivePartsInUse() ([]model.VehiclePart, error) {
	parts := []model.VehiclePart{}
	err := r.db.Select(&parts, `SELECT * FROM "VehiclePart" WHERE "replacedAt" IS NULL`)
	return parts, err
}

func (r *VehicleRepository) DocumentsExpiringWithin(days int) ([]model.VehicleDocument, error) {
	docs := []model.VehicleDocument{}
	err := r.db.Select(&docs, `
		SELECT * FROM "VehicleDocument"
		WHERE "expiryDate" IS NOT NULL AND "expiryDate" <= now() + ($1 || ' days')::interval
		ORDER BY "expiryDate" ASC
	`, days)
	return docs, err
}

// ── مصاريف السيارة (المرحلة 3) ──

// SumLogCostByType مجموع تكلفة سجلات السيارة من نوع معيّن ضمن فترة [from, to).
func (r *VehicleRepository) SumLogCostByType(vehicleID, logType string, from, to time.Time) (float64, error) {
	var total float64
	err := r.db.Get(&total, `
		SELECT COALESCE(SUM(cost), 0) FROM "VehicleLog"
		WHERE "vehicleId" = $1 AND type = $2 AND "performedAt" >= $3 AND "performedAt" < $4
	`, vehicleID, logType, from, to)
	return total, err
}

// SumIncidentCost مجموع تكلفة الحوادث/الأعطال/الأضرار (cost + repairCost) ضمن فترة.
func (r *VehicleRepository) SumIncidentCost(vehicleID string, from, to time.Time) (float64, error) {
	var total float64
	err := r.db.Get(&total, `
		SELECT COALESCE(SUM(COALESCE(cost, 0) + COALESCE("repairCost", 0)), 0) FROM "VehicleIncident"
		WHERE "vehicleId" = $1 AND "createdAt" >= $2 AND "createdAt" < $3
	`, vehicleID, from, to)
	return total, err
}

// SumPartsCost مجموع تكلفة قطع (إطارات/بطاريات) رُكّبت ضمن فترة.
func (r *VehicleRepository) SumPartsCost(vehicleID string, from, to time.Time) (float64, error) {
	var total float64
	err := r.db.Get(&total, `
		SELECT COALESCE(SUM(cost), 0) FROM "VehiclePart"
		WHERE "vehicleId" = $1 AND "installedAt" >= $2 AND "installedAt" < $3
	`, vehicleID, from, to)
	return total, err
}

// OdometerRangeInPeriod يجيب أقل وأكبر قراءة عداد مسجّلة (من سجلات السيارة) ضمن فترة،
// لتقدير المسافة المقطوعة بالفترة إن كانت متوفرة.
func (r *VehicleRepository) OdometerRangeInPeriod(vehicleID string, from, to time.Time) (min, max *int, err error) {
	var row struct {
		Min *int `db:"min"`
		Max *int `db:"max"`
	}
	err = r.db.Get(&row, `
		SELECT MIN(odometer) AS min, MAX(odometer) AS max FROM "VehicleLog"
		WHERE "vehicleId" = $1 AND "performedAt" >= $2 AND "performedAt" < $3 AND odometer IS NOT NULL
	`, vehicleID, from, to)
	if err != nil {
		return nil, nil, err
	}
	return row.Min, row.Max, nil
}

// DistanceFromMissionsInPeriod مجموع مسافات مهمات السيارة المكتملة ضمن فترة (بديل احتياطي
// لتقدير المسافة إن ما توفرت قراءات عداد بسجلات الوقود بنفس الفترة).
func (r *VehicleRepository) DistanceFromMissionsInPeriod(vehicleID string, from, to time.Time) (*int, error) {
	var total *int
	err := r.db.Get(&total, `
		SELECT SUM("distanceKm") FROM "VehicleMission"
		WHERE "vehicleId" = $1 AND "startedAt" >= $2 AND "startedAt" < $3 AND "distanceKm" IS NOT NULL
	`, vehicleID, from, to)
	return total, err
}

// LastFuelLogCosts آخر N تكلفة لسجلات وقود السيارة (الأحدث أولاً) — تُستخدم لحساب متوسط
// شذوذ تكلفة تعبئة الوقود.
func (r *VehicleRepository) LastFuelLogCosts(vehicleID string, limit int) ([]float64, error) {
	costs := []float64{}
	err := r.db.Select(&costs, `
		SELECT cost FROM "VehicleLog"
		WHERE "vehicleId" = $1 AND type = 'FUEL' AND cost IS NOT NULL
		ORDER BY "performedAt" DESC, "createdAt" DESC
		LIMIT $2
	`, vehicleID, limit)
	return costs, err
}

// VehicleIDsWithOpenIncidentType معرفات السيارات التي لديها حادثة/عطل بحالة OPEN من نوع
// معيّن — تُستخدم لتقدير "السيارات بالصيانة حالياً" (عطل مفتوح لم يُحل بعد).
func (r *VehicleRepository) VehicleIDsWithOpenIncidentType(incidentType string) ([]string, error) {
	ids := []string{}
	err := r.db.Select(&ids, `
		SELECT DISTINCT "vehicleId" FROM "VehicleIncident"
		WHERE status = 'OPEN' AND type = $1
	`, incidentType)
	return ids, err
}

// VehicleIDsWithInProgressMission معرفات السيارات التي لديها مهمة قيد التنفيذ حالياً.
func (r *VehicleRepository) VehicleIDsWithInProgressMission() ([]string, error) {
	ids := []string{}
	err := r.db.Select(&ids, `
		SELECT DISTINCT "vehicleId" FROM "VehicleMission" WHERE status = 'IN_PROGRESS'
	`)
	return ids, err
}

// MissionCountInPeriod عدد المهام المبدوءة لكل سيارة ضمن فترة (لترتيب الأكثر استخداماً).
type VehicleUsageRow struct {
	VehicleID    string `db:"vehicleId"`
	MissionCount int    `db:"missionCount"`
	DistanceKm   int    `db:"distanceKm"`
}

func (r *VehicleRepository) MissionUsageInPeriod(from, to time.Time) ([]VehicleUsageRow, error) {
	rows := []VehicleUsageRow{}
	err := r.db.Select(&rows, `
		SELECT "vehicleId",
		       COUNT(*) AS "missionCount",
		       COALESCE(SUM(COALESCE("distanceKm", 0)), 0) AS "distanceKm"
		FROM "VehicleMission"
		WHERE "startedAt" >= $1 AND "startedAt" < $2
		GROUP BY "vehicleId"
	`, from, to)
	return rows, err
}
