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
	var n int
	err := r.db.Get(&n, `
		SELECT COUNT(*) FROM "AiSignal"
		WHERE kind = $1 AND "employeeId" = $2
		  AND "occurredAt" > now() - ($3 || ' days')::interval`,
		model.AiSignalWorkStopped, employeeID, days)
	return n, err
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
	err := r.db.Select(&rows, `
		SELECT * FROM "AiMetric"
		WHERE "periodStart" >= $1 AND "periodEnd" <= $2
		ORDER BY "metricKey", scope`, from, to)
	return rows, err
}

// itoa بسيط لبناء رقم المعامل بالاستعلام — بدون استيراد strconv لعملية وحدة.
func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
