package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// AiRepository نواة الذكاء الاصطناعي — إشارات وأدلة وأحكام ومؤشرات.
//
// ⚠️ اقرا رأس schema_ai_core.go: الأدلة حقائق نحسبها، والحكم تفسير.
type AiRepository struct {
	db *sqlx.DB
}

func NewAiRepository(db *sqlx.DB) *AiRepository { return &AiRepository{db: db} }

// ═══ الإشارات ═══

// RecordSignal يسجّل إشارة. الفهرس الفريد يمنع تكرار نفس الحدث.
//
// ⚠️ ما يرجّع خطأ يوقف العملية الأصلية: فشل تسجيل إشارة تحليل ما
// يصير يمنع الموظف من إيقاف شغله.
func (r *AiRepository) RecordSignal(in model.AiSignal) (*model.AiSignal, error) {
	if len(in.Payload) == 0 {
		in.Payload = []byte(`{}`)
	}
	if in.OccurredAt.IsZero() {
		in.OccurredAt = time.Now()
	}
	var row model.AiSignal
	err := r.db.Get(&row, `
		INSERT INTO "AiSignal" (id, kind, "entityType", "entityId", "employeeId", payload, "occurredAt")
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (kind, "entityType", "entityId", "occurredAt") DO NOTHING
		RETURNING *`,
		uuid.NewString(), in.Kind, in.EntityType, in.EntityID, in.EmployeeID, in.Payload, in.OccurredAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // انسجّلت من قبل — مو خطأ
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// PendingSignals الإشارات الي لسه ما انجمعت أدلتها.
func (r *AiRepository) PendingSignals(limit int) ([]model.AiSignal, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows := []model.AiSignal{}
	err := r.db.Select(&rows, `
		SELECT * FROM "AiSignal" WHERE status = 'PENDING'
		ORDER BY "occurredAt" ASC LIMIT $1`, limit)
	return rows, err
}

func (r *AiRepository) SetSignalStatus(id, status string) error {
	_, err := r.db.Exec(`UPDATE "AiSignal" SET status = $2 WHERE id = $1`, id, status)
	return err
}

// ListSignals للعرض — مع الأدلة والحكم مهدرجين.
func (r *AiRepository) ListSignals(kind string, limit int) ([]model.AiSignal, error) {
	if limit <= 0 || limit > 300 {
		limit = 100
	}
	rows := []model.AiSignal{}
	q := `SELECT * FROM "AiSignal"`
	args := []any{}
	if kind != "" {
		q += ` WHERE kind = $1`
		args = append(args, kind)
	}
	q += ` ORDER BY "occurredAt" DESC LIMIT $` + itoa(len(args)+1)
	args = append(args, limit)
	if err := r.db.Select(&rows, q, args...); err != nil {
		return nil, err
	}
	r.hydrateSignals(rows)
	return rows, nil
}

// hydrateSignals يلزق الأدلة والحكم والأسماء — بدفعة وحدة مو استعلام لكل صف.
func (r *AiRepository) hydrateSignals(rows []model.AiSignal) {
	if len(rows) == 0 {
		return
	}
	ids := make([]string, 0, len(rows))
	empIDs := map[string]bool{}
	for i := range rows {
		ids = append(ids, rows[i].ID)
		if rows[i].EmployeeID != nil {
			empIDs[*rows[i].EmployeeID] = true
		}
	}

	evByID := map[string]model.AiEvidence{}
	if q, args, err := sqlx.In(`SELECT * FROM "AiEvidence" WHERE "signalId" IN (?)`, ids); err == nil {
		list := []model.AiEvidence{}
		if err := r.db.Select(&list, r.db.Rebind(q), args...); err == nil {
			for _, e := range list {
				ee := e
				// نفك الـJSON هنا حتى الواجهة تلگاه كائناً مو نص
				_ = json.Unmarshal(ee.Facts, &ee.FactsMap)
				_ = json.Unmarshal(ee.Gaps, &ee.GapsList)
				evByID[e.SignalID] = ee
			}
		}
	}

	vByID := map[string]model.AiVerdict{}
	if q, args, err := sqlx.In(
		`SELECT DISTINCT ON ("signalId") * FROM "AiVerdict" WHERE "signalId" IN (?)
		 ORDER BY "signalId", "createdAt" DESC`, ids); err == nil {
		list := []model.AiVerdict{}
		if err := r.db.Select(&list, r.db.Rebind(q), args...); err == nil {
			for _, v := range list {
				vByID[v.SignalID] = v
				if v.BlameEmployeeID != nil {
					empIDs[*v.BlameEmployeeID] = true
				}
			}
		}
	}

	names := map[string]string{}
	if len(empIDs) > 0 {
		list := make([]string, 0, len(empIDs))
		for id := range empIDs {
			list = append(list, id)
		}
		if q, args, err := sqlx.In(`SELECT id, name FROM "Employee" WHERE id IN (?)`, list); err == nil {
			briefs := []model.EmployeeBrief{}
			if err := r.db.Select(&briefs, r.db.Rebind(q), args...); err == nil {
				for _, b := range briefs {
					names[b.ID] = b.Name
				}
			}
		}
	}

	for i := range rows {
		if e, ok := evByID[rows[i].ID]; ok {
			ee := e
			rows[i].Evidence = &ee
		}
		if v, ok := vByID[rows[i].ID]; ok {
			vv := v
			if vv.BlameEmployeeID != nil {
				if n, ok := names[*vv.BlameEmployeeID]; ok {
					vv.BlameEmployeeName = &n
				}
			}
			rows[i].Verdict = &vv
		}
		if rows[i].EmployeeID != nil {
			if n, ok := names[*rows[i].EmployeeID]; ok {
				rows[i].EmployeeName = &n
			}
		}
	}
}

// ═══ الأدلة ═══

func (r *AiRepository) SaveEvidence(signalID string, facts, gaps []byte) (*model.AiEvidence, error) {
	var row model.AiEvidence
	err := r.db.Get(&row, `
		INSERT INTO "AiEvidence" (id, "signalId", facts, gaps)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT ("signalId") DO UPDATE
			SET facts = EXCLUDED.facts, gaps = EXCLUDED.gaps, "collectedAt" = now()
		RETURNING *`, uuid.NewString(), signalID, facts, gaps)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(row.Facts, &row.FactsMap)
	_ = json.Unmarshal(row.Gaps, &row.GapsList)
	return &row, nil
}

// ═══ الحكم ═══

func (r *AiRepository) SaveVerdict(v model.AiVerdict) (*model.AiVerdict, error) {
	if v.Source == "" {
		v.Source = model.AiSourceRules
	}
	var row model.AiVerdict
	err := r.db.Get(&row, `
		INSERT INTO "AiVerdict"
			(id, "signalId", source, "modelName", headline, reasoning, confidence, severity, "blameEmployeeId", suggestion)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING *`,
		uuid.NewString(), v.SignalID, v.Source, v.ModelName, v.Headline, v.Reasoning,
		v.Confidence, v.Severity, v.BlameEmployeeID, v.Suggestion)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// RecentJudgedFeedback أمثلة حقيقية: حكم سابق لنفس صنف الإشارة + قرار
// المراقب الحقيقي عليه. هذي مادة التعلّم — بلا جدول جديد، لأن قرار
// المراقب مخزون أصلاً بصندوقه (`MonitorReview`، محطة `AI_VERDICT`).
func (r *AiRepository) RecentJudgedFeedback(kind string, limit int) ([]model.MonitorFeedbackExample, error) {
	if limit <= 0 || limit > 20 {
		limit = 5
	}
	rows := []model.MonitorFeedbackExample{}
	err := r.db.Select(&rows, `
		SELECT v.headline, v.reasoning, e.facts,
		       mr.status AS "monitorStatus", mr.note AS "monitorNote"
		FROM "AiVerdict" v
		JOIN "AiSignal" s ON s.id = v."signalId"
		JOIN "AiEvidence" e ON e."signalId" = s.id
		JOIN "MonitorReview" mr ON mr."entityType" = 'AI_VERDICT' AND mr."entityId" = v.id
		WHERE s.kind = $1 AND mr.status <> 'PENDING'
		ORDER BY mr."reviewedAt" DESC
		LIMIT $2`, kind, limit)
	return rows, err
}

// VerdictSeverityCounts أحكام ماتركس بفترة معيّنة مبوّبة بالخطورة —
// مادة الفضفضة اليومية.
type VerdictSeverityCounts struct {
	Critical int `db:"critical"`
	Warn     int `db:"warn"`
	Watch    int `db:"watch"`
	Info     int `db:"info"`
	Total    int `db:"total"`
}

func (r *AiRepository) VerdictSeverityCounts(from, to time.Time) (VerdictSeverityCounts, error) {
	var c VerdictSeverityCounts
	err := r.db.Get(&c, `
		SELECT
			COUNT(*) FILTER (WHERE severity = 'CRITICAL') AS critical,
			COUNT(*) FILTER (WHERE severity = 'WARN') AS warn,
			COUNT(*) FILTER (WHERE severity = 'WATCH') AS watch,
			COUNT(*) FILTER (WHERE severity = 'INFO') AS info,
			COUNT(*) AS total
		FROM "AiVerdict"
		WHERE "createdAt" >= $1 AND "createdAt" < $2`, from, to)
	return c, err
}

// ClaimDailyMarker «يحجز» مفتاحاً بيوم معيّن — أول Loop يوصل يفوز
// (`ON CONFLICT DO NOTHING`)، والباقي يرجعله false فما يكرر نفس
// الشغلة بنفس اليوم. نفس جدول `AiMetric` الموجود، بدون جدول جديد —
// scope="MARKER" يفصلها عن المؤشرات الحقيقية بوضوح.
func (r *AiRepository) ClaimDailyMarker(metricKey, day string) (bool, error) {
	anchor, err := time.Parse("2006-01-02", day)
	if err != nil {
		return false, err
	}
	res, err := r.db.Exec(`
		INSERT INTO "AiMetric" (id, "metricKey", scope, "scopeId", "periodStart", "periodEnd", value, "sampleCount", details)
		VALUES ($1, $2, 'MARKER', $3, $4, $4, 1, 0, '{}')
		ON CONFLICT ("metricKey", scope, COALESCE("scopeId",''), "periodStart", "periodEnd") DO NOTHING
	`, uuid.NewString(), metricKey, day, anchor)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// MonitorAgreementCounts شكد من أحكام ماتركس بفترة معيّنة وافق عليها
// المراقب (OK) من أصل الي بتّ فيها. مقياس الدقة — بلا جدول جديد.
func (r *AiRepository) MonitorAgreementCounts(from, to time.Time) (agreed, total int, err error) {
	var row struct {
		Agreed int `db:"agreed"`
		Total  int `db:"total"`
	}
	err = r.db.Get(&row, `
		SELECT COUNT(*) FILTER (WHERE status = 'OK') AS agreed, COUNT(*) AS total
		FROM "MonitorReview"
		WHERE "entityType" = 'AI_VERDICT' AND status <> 'PENDING'
		  AND "reviewedAt" >= $1 AND "reviewedAt" <= $2`, from, to)
	return row.Agreed, row.Total, err
}

// ═══ الأدلة الخام — استعلامات الحقائق ═══

// ProcurementCounts أرقام طلبات المواد لحجز.
type ProcurementCounts struct {
	Total      int `db:"total"`
	BeforeStop int `db:"beforeStop"`
}

// ProcurementSummary: طلب مادة لهذا الحجز؟ وكم منها قبل التوقف؟
//
// ⚠️ «قبل التوقف» هو المفتاح كله: طلب **قبل** ما يوقّف يعني كان
// منتبه وينتظر التوفير — والمسؤولية تنتقل لإداري الكميات. طلب
// **بعد** التوقف يعني اكتشفها بالموقع.
func (r *AiRepository) ProcurementSummary(bookingID string, stoppedAt time.Time) (ProcurementCounts, string, error) {
	var c ProcurementCounts
	err := r.db.Get(&c, `
		SELECT COUNT(*) AS total,
		       COUNT(*) FILTER (WHERE "createdAt" <= $2) AS "beforeStop"
		FROM "ProcurementRequest" WHERE "bookingId" = $1`, bookingID, stoppedAt)
	if err != nil {
		return c, "", err
	}
	var status sql.NullString
	_ = r.db.Get(&status, `
		SELECT status FROM "ProcurementRequest"
		WHERE "bookingId" = $1 ORDER BY "createdAt" DESC LIMIT 1`, bookingID)
	return c, status.String, nil
}

// CartSummary: كم مادة بسلة الزبون، وكم انضافت **بعد** ما بدأ الشغل.
//
// ⚠️ الي انضاف بعد البداية = الزبون طلب زيادة بالموقع. هذي تبرّئ
// الموظف من تهمة «نسى مادة» — وبلا هذا الرقم الاثنين يتشابهون.
func (r *AiRepository) CartSummary(bookingID string, startedAt *time.Time) (int, int, error) {
	var total int
	if err := r.db.Get(&total, `SELECT COUNT(*) FROM "CartItem" WHERE "bookingId" = $1`, bookingID); err != nil {
		return 0, 0, err
	}
	if startedAt == nil {
		return total, 0, nil
	}
	var after int
	if err := r.db.Get(&after, `
		SELECT COUNT(*) FROM "CartItem" WHERE "bookingId" = $1 AND "createdAt" > $2`,
		bookingID, *startedAt); err != nil {
		return total, 0, err
	}
	return total, after, nil
}

// StopCountForEmployee كم مرة وقّف هذا الموظف الشغل بآخر كذا يوم.
// مرة = ظرف، خمس مرات = نمط.
func (r *AiRepository) StopCountForEmployee(employeeID string, days int) (int, error) {
	return r.SignalCountForEmployee(model.AiSignalWorkStopped, employeeID, days)
}

// SignalCountForEmployee كم مرة تكررت نفس الإشارة لهذا الموظف بآخر
// كذا يوم — نفس فكرة «مرة ظرف، خمس مرات نمط» بس لأي صنف إشارة.
func (r *AiRepository) SignalCountForEmployee(kind, employeeID string, days int) (int, error) {
	var n int
	err := r.db.Get(&n, `
		SELECT COUNT(*) FROM "AiSignal"
		WHERE kind = $1 AND "employeeId" = $2
		  AND "occurredAt" > now() - ($3 || ' days')::interval`,
		kind, employeeID, days)
	return n, err
}

// SignalCountForEmployeeBetween كم مرة تكررت نفس الإشارة لهذا الموظف
// بين تاريخين — نفس `SignalCountForEmployee` بس بمدى صريح مو «آخر كذا
// يوم من الآن»، حتى نقارن أسبوعين ببعض (مثلاً «هذا الأسبوع» مقابل
// «الأسبوع الماضي») — أساس مديح ماتركس الأسبوعي.
func (r *AiRepository) SignalCountForEmployeeBetween(kind, employeeID string, from, to time.Time) (int, error) {
	var n int
	err := r.db.Get(&n, `
		SELECT COUNT(*) FROM "AiSignal"
		WHERE kind = $1 AND "employeeId" = $2
		  AND "occurredAt" >= $3 AND "occurredAt" < $4`,
		kind, employeeID, from, to)
	return n, err
}

// DaysSinceLastSignal شكد يوم مرّ من آخر إشارة من نفس الصنف لنفس
// الموظف **قبل** لحظة معيّنة (عادة وقت الإشارة الحالية — نقارن
// بالتاريخ الحقيقي للحدث مو بوقت معالجتها، لأن الكنسة ممكن تتأخر
// دقايق). `nil` يعني هذي أول مرة إطلاقاً — ماكو سجل قبلها، فما
// نقدر نحچي عن «فترة نظيفة» أصلاً.
//
// ⚠️ هذا نصف التصعيد بالاتجاهين: النصف الأول (التشديد) موجود من
// زمان بـ`StopsLast30Days`/`LateCountLast30Days`. هذا يضيف النصف
// الثاني — التساهل بعد فترة نظيفة، مو التشديد بس بالتكرار.
func (r *AiRepository) DaysSinceLastSignal(kind, employeeID string, before time.Time) (*int, error) {
	var maxOccurred sql.NullTime
	err := r.db.Get(&maxOccurred, `
		SELECT MAX("occurredAt") FROM "AiSignal"
		WHERE kind = $1 AND "employeeId" = $2 AND "occurredAt" < $3`,
		kind, employeeID, before)
	if err != nil {
		return nil, err
	}
	if !maxOccurred.Valid {
		return nil, nil
	}
	days := int(before.Sub(maxOccurred.Time).Hours() / 24)
	return &days, nil
}

// LatestVisitCrewSize عدد كادر آخر طلعة لحجز معيّن — أساس كشف
// «تناقض التقرير الذاتي» (SELF_REPORT_MISMATCH): الموظف ادّعى إنه
// سوى الشغل لحاله والكادر الحقيقي أكثر من واحد.
func (r *AiRepository) LatestVisitCrewSize(bookingID string) (int, error) {
	var n int
	err := r.db.Get(&n, `
		SELECT COUNT(*) FROM "BookingVisitCrew" vc
		WHERE vc."visitId" = (
			SELECT id FROM "BookingVisit" WHERE "bookingId" = $1 ORDER BY "visitNumber" DESC LIMIT 1
		)`, bookingID)
	return n, err
}

// TrainingGapCandidate موظف تكرر عنده توقف عمل بخدمة معيّنة ولسه ما
// أخذ (أو ما نجح بـ) تدريب يغطّي هذي الخدمة — مادة إشارة «فجوة
// تدريب» (يحوّل «منو غلط» لـ«شنو ناقص بالمنظومة»).
type TrainingGapCandidate struct {
	EmployeeID   string `db:"employeeId"`
	EmployeeName string `db:"employeeName"`
	ServiceID    string `db:"serviceId"`
	ServiceName  string `db:"serviceName"`
	StopCount    int    `db:"stopCount"`
}

// TrainingGapCandidates: موظفين عندهم `minStops` توقف عمل فأكثر بنفس
// الخدمة خلال آخر `days` يوم، وما عندهم برنامج تدريب ناجح يغطّي
// مهارة هذي الخدمة (`TrainingProgramSkill` → `Skill.serviceId`).
func (r *AiRepository) TrainingGapCandidates(minStops, days int) ([]TrainingGapCandidate, error) {
	rows := []TrainingGapCandidate{}
	err := r.db.Select(&rows, `
		SELECT s."employeeId" AS "employeeId", e.name AS "employeeName",
		       b."serviceId" AS "serviceId", sv.name AS "serviceName",
		       COUNT(*) AS "stopCount"
		FROM "AiSignal" s
		JOIN "Booking" b ON b.id = s."entityId" AND s."entityType" = 'BOOKING'
		JOIN "Employee" e ON e.id = s."employeeId"
		JOIN "Service" sv ON sv.id = b."serviceId"
		WHERE s.kind = $1 AND s."occurredAt" > now() - ($2 || ' days')::interval
		  AND s."employeeId" IS NOT NULL AND b."serviceId" IS NOT NULL
		GROUP BY s."employeeId", e.name, b."serviceId", sv.name
		HAVING COUNT(*) >= $3
		  AND NOT EXISTS (
		    SELECT 1 FROM "TrainingProgramParticipant" tpp
		    JOIN "TrainingProgramSkill" tps ON tps."programId" = tpp."programId"
		    JOIN "Skill" sk ON sk.id = tps."skillId"
		    WHERE tpp."employeeId" = s."employeeId" AND tpp.passed = true AND sk."serviceId" = b."serviceId"
		  )
		ORDER BY "stopCount" DESC`,
		model.AiSignalWorkStopped, days, minStops)
	return rows, err
}

// FuelAnomalySnapshot أدلة شذوذ تعبئة وقود — يعيد حساب نفس منطق
// `VehicleService.CheckFuelAnomaly` (متوسط آخر ٥ تعبئات قبل هذا
// السجل) وقت التحليل، بدل الاعتماد على نتيجة لحظة الإدخال.
func (r *AiRepository) FuelAnomalySnapshot(logID string) (*model.FuelAnomalyEvidence, error) {
	var row struct {
		VehicleID string    `db:"vehicleId"`
		Cost      float64   `db:"cost"`
		CreatedAt time.Time `db:"createdAt"`
	}
	if err := r.db.Get(&row, `SELECT "vehicleId", cost, "createdAt" FROM "VehicleLog" WHERE id = $1 AND cost IS NOT NULL`, logID); err != nil {
		return nil, err
	}
	ev := &model.FuelAnomalyEvidence{NewCost: row.Cost}
	_ = r.db.Get(&ev.VehiclePlate, `SELECT "plateNumber" FROM "Vehicle" WHERE id = $1`, row.VehicleID)
	var avg sql.NullFloat64
	_ = r.db.Get(&avg, `
		SELECT AVG(cost) FROM (
			SELECT cost FROM "VehicleLog"
			WHERE "vehicleId" = $1 AND type = 'FUEL' AND cost IS NOT NULL AND "createdAt" < $2
			ORDER BY "createdAt" DESC LIMIT 5
		) recent`, row.VehicleID, row.CreatedAt)
	if avg.Valid && avg.Float64 > 0 {
		ev.AverageCost = avg.Float64
		ev.PercentAbove = (row.Cost - avg.Float64) / avg.Float64 * 100
	}
	return ev, nil
}

// OverdueMaintenanceVehicle مركبة تجاوزت صيانتها المجدولة (تاريخاً أو
// عداد كيلومترات) وبعدها تُرسل لمهمة جارية أو حجز معتمد.
type OverdueMaintenanceVehicle struct {
	VehicleName string `db:"vehicleName"`
	PlateNumber string `db:"plateNumber"`
}

// OverdueMaintenanceVehiclesInUse: آخر سجل صيانة (`VehicleLog` نوع
// MAINTENANCE) لكل مركبة يحدّد موعدها الجاي (`nextDueAt`) أو عداد
// كيلومتراتها (`nextDueOdometer`) — لو تجاوزته المركبة **وبعدها**
// مستمرة بمهمة جارية أو حجز معتمد لسه ما خلص، هذا يستاهل تنبيهاً.
func (r *AiRepository) OverdueMaintenanceVehiclesInUse() ([]OverdueMaintenanceVehicle, error) {
	rows := []OverdueMaintenanceVehicle{}
	err := r.db.Select(&rows, `
		WITH latest_maintenance AS (
			SELECT DISTINCT ON ("vehicleId") "vehicleId", "nextDueAt", "nextDueOdometer"
			FROM "VehicleLog"
			WHERE type = 'MAINTENANCE'
			ORDER BY "vehicleId", "performedAt" DESC
		)
		SELECT v.name AS "vehicleName", v."plateNumber" AS "plateNumber"
		FROM latest_maintenance lm
		JOIN "Vehicle" v ON v.id = lm."vehicleId"
		WHERE (
			(lm."nextDueAt" IS NOT NULL AND lm."nextDueAt" < now())
			OR (lm."nextDueOdometer" IS NOT NULL AND v."currentOdometer" >= lm."nextDueOdometer")
		)
		AND (
			EXISTS (SELECT 1 FROM "VehicleMission" vm WHERE vm."vehicleId" = v.id AND vm.status = 'IN_PROGRESS')
			OR EXISTS (SELECT 1 FROM "VehicleBooking" vb WHERE vb."vehicleId" = v.id AND vb.status = 'APPROVED' AND vb."endAt" >= now())
		)`)
	return rows, err
}

// ═══ ساعات الدوام ═══

func (r *AiRepository) WorkWindow() (*model.AiWorkWindow, error) {
	var w model.AiWorkWindow
	err := r.db.Get(&w, `SELECT * FROM "AiWorkWindow" WHERE id = 'default'`)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *AiRepository) SetWorkWindow(startHour, endHour int) error {
	_, err := r.db.Exec(`
		UPDATE "AiWorkWindow" SET "startHour" = $1, "endHour" = $2, "updatedAt" = now()
		WHERE id = 'default'`, startHour, endHour)
	return err
}

// ═══ المؤشرات ═══

func (r *AiRepository) UpsertMetric(m model.AiMetric) error {
	if len(m.Details) == 0 {
		m.Details = []byte(`{}`)
	}
	_, err := r.db.Exec(`
		INSERT INTO "AiMetric" (id, "metricKey", scope, "scopeId", "periodStart", "periodEnd", value, "sampleCount", details)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT ("metricKey", scope, COALESCE("scopeId",''), "periodStart", "periodEnd")
		DO UPDATE SET value = EXCLUDED.value, "sampleCount" = EXCLUDED."sampleCount",
		              details = EXCLUDED.details, "computedAt" = now()`,
		uuid.NewString(), m.MetricKey, m.Scope, m.ScopeID, m.PeriodStart, m.PeriodEnd,
		m.Value, m.SampleCount, m.Details)
	return err
}

func (r *AiRepository) ListMetrics(from, to time.Time) ([]model.AiMetric, error) {
	rows := []model.AiMetric{}
	// ⚠️ scope='MARKER' علامات داخلية (مثلاً «الفضفضة اليومية انرسلت
	// اليوم») — مو مؤشراً حقيقياً، تُستثنى حتى ما تطلع كبطاقة بالواجهة.
	err := r.db.Select(&rows, `
		SELECT * FROM "AiMetric"
		WHERE "periodStart" >= $1 AND "periodEnd" <= $2 AND scope <> 'MARKER'
		ORDER BY "metricKey", scope`, from, to)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		if len(rows[i].Details) == 0 {
			continue
		}
		var m map[string]any
		if json.Unmarshal(rows[i].Details, &m) == nil {
			rows[i].DetailsMap = m
		}
	}
	return rows, nil
}

// itoa بسيط لبناء رقم المعامل بالاستعلام — بدون استيراد strconv لعملية وحدة.
func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
