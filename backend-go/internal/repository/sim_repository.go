package repository

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// ═══ مستودع مختبر المحاكاة ═══
//
// ⚠️ **قاعدة السلامة الأساسية بالمشروع كله موجودة هنا مو بالمعالج**:
// المتدرّب ما يشوف إلا محتوى `PUBLISHED` و`verified = true`. المحتوى
// المسحوب من الإنترنت وما انتأكد منه على جهاز حقيقي يبقى للمالك وحده.
//
// ليش بالمستودع مو بالمعالج؟ لأن المعالج ممكن ينضاف له مسار جديد بكرة
// وينسى الفحص. الاستعلام نفسه ما ينسى.
//
// تعريف أسلاك غلط لقفل بـ١٥ سلك يعني فني يحرق قفلاً حقيقياً، أو يتكهرب
// بـ٢٢٠ فولت بالطاقة الشمسية. المحاكي الغلط أخطر من ما اكو محاكي.
type SimRepository struct{ db *sqlx.DB }

func NewSimRepository(db *sqlx.DB) *SimRepository { return &SimRepository{db: db} }

// visibilityClause شرط الرؤية حسب منو يسأل.
// المالك يشوف كلشي (لأنه هو الي يراجع ويعتمد)، وغيره يشوف المنشور
// المحقّق بس.
func visibilityClause(alias string, isOwner bool) string {
	if isOwner {
		return ""
	}
	return ` AND ` + alias + `.status = 'PUBLISHED' AND ` + alias + `.verified = TRUE`
}

// ListCategories فئات المختبر مع عدد التمارين المتاحة للسائل.
func (r *SimRepository) ListCategories(isOwner bool) ([]model.SimCategory, error) {
	rows := []model.SimCategory{}
	q := `
		SELECT c.*, srv.name AS "serviceName",
		       (SELECT COUNT(*) FROM "SimExercise" e
		         WHERE e."categoryId" = c.id` + visibilityClause("e", isOwner) + `) AS "exerciseCount"
		FROM "SimCategory" c
		LEFT JOIN "Service" srv ON srv.id = c."serviceId"
		WHERE c.archived = FALSE
		ORDER BY c."sortOrder", c.name
		LIMIT 200`
	// ⚠️ الأعمدة المحسوبة (`serviceName`, `exerciseCount`) معلّمة `db:"-"`
	// بالنموذج حتى ما تنكتب بالإدخال، فنجيبها بمسح يدوي.
	type scanRow struct {
		model.SimCategory
		ServiceNameCol   *string `db:"serviceName"`
		ExerciseCountCol int     `db:"exerciseCount"`
	}
	scanned := []scanRow{}
	if err := r.db.Select(&scanned, q); err != nil {
		return nil, err
	}
	for _, s := range scanned {
		c := s.SimCategory
		c.ServiceName = s.ServiceNameCol
		c.ExerciseCount = s.ExerciseCountCol
		rows = append(rows, c)
	}
	return rows, nil
}

// ListExercises تمارين فئة.
func (r *SimRepository) ListExercises(categoryID string, isOwner bool, employeeID string) ([]model.SimExercise, error) {
	rows := []model.SimExercise{}
	q := `
		SELECT e.* FROM "SimExercise" e
		WHERE e."categoryId" = $1` + visibilityClause("e", isOwner) + `
		ORDER BY e."sortOrder", e.difficulty, e.title
		LIMIT 300`
	if err := r.db.Select(&rows, q, categoryID); err != nil {
		return nil, err
	}
	return r.attachMastery(rows, employeeID)
}

// attachMastery يعبّي أفضل نتيجة للموظف على كل تمرين بنداء واحد.
func (r *SimRepository) attachMastery(rows []model.SimExercise, employeeID string) ([]model.SimExercise, error) {
	if len(rows) == 0 || employeeID == "" {
		return rows, nil
	}
	ids := make([]string, 0, len(rows))
	for _, e := range rows {
		ids = append(ids, e.ID)
	}
	type mrow struct {
		ExerciseID string `db:"exerciseId"`
		BestScore  int    `db:"bestScore"`
		Passed     bool   `db:"passed"`
	}
	q, args, err := sqlx.In(`
		SELECT "exerciseId", "bestScore", passed FROM "SimMastery"
		WHERE "employeeId" = ? AND "exerciseId" IN (?)`, employeeID, ids)
	if err != nil {
		return rows, nil
	}
	ms := []mrow{}
	if err := r.db.Select(&ms, r.db.Rebind(q), args...); err != nil {
		return rows, nil
	}
	byID := map[string]mrow{}
	for _, m := range ms {
		byID[m.ExerciseID] = m
	}
	for i := range rows {
		if m, ok := byID[rows[i].ID]; ok {
			s := m.BestScore
			rows[i].BestScore = &s
			rows[i].Passed = m.Passed
		}
	}
	return rows, nil
}

// GetExercise تمرين واحد ومعه أجهزة مشهده كاملة.
func (r *SimRepository) GetExercise(id string, isOwner bool) (*model.SimExercise, error) {
	var e model.SimExercise
	q := `SELECT * FROM "SimExercise" WHERE id = $1` + visibilityClause("SimExercise", isOwner)
	if err := r.db.Get(&e, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("التمرين مو موجود")
		}
		return nil, err
	}
	devs, err := r.devicesForScene(e.Scene)
	if err != nil {
		return nil, err
	}
	e.Devices = devs
	return &e, nil
}

// devicesForScene يجيب كل جهاز مذكور بمشهد التمرين.
//
// ⚠️ الأجهزة تنجاب كاملة مع التمرين بنداء واحد: المحرّك بالواجهة يحتاج
// أطراف كل جهاز قبل ما يرسم أول سلك، ونداء لكل جهاز يعني شاشة تتقطّع
// قدّام المتدرّب.
func (r *SimRepository) devicesForScene(scene json.RawMessage) ([]model.SimDevice, error) {
	var parsed struct {
		Devices []struct {
			DeviceID string `json:"deviceId"`
		} `json:"devices"`
	}
	if err := json.Unmarshal(scene, &parsed); err != nil || len(parsed.Devices) == 0 {
		return []model.SimDevice{}, nil
	}
	ids := make([]string, 0, len(parsed.Devices))
	seen := map[string]bool{}
	for _, d := range parsed.Devices {
		if d.DeviceID != "" && !seen[d.DeviceID] {
			seen[d.DeviceID] = true
			ids = append(ids, d.DeviceID)
		}
	}
	if len(ids) == 0 {
		return []model.SimDevice{}, nil
	}
	q, args, err := sqlx.In(`SELECT * FROM "SimDevice" WHERE id IN (?)`, ids)
	if err != nil {
		return nil, err
	}
	rows := []model.SimDevice{}
	if err := r.db.Select(&rows, r.db.Rebind(q), args...); err != nil {
		return nil, err
	}
	return rows, nil
}

// ListLessons دروس فئة.
func (r *SimRepository) ListLessons(categoryID string, isOwner bool) ([]model.SimLesson, error) {
	rows := []model.SimLesson{}
	extra := ""
	if !isOwner {
		extra = ` AND status = 'PUBLISHED'`
	}
	q := `SELECT * FROM "SimLesson" WHERE "categoryId" = $1` + extra + ` ORDER BY "sortOrder" LIMIT 200`
	if err := r.db.Select(&rows, q, categoryID); err != nil {
		return nil, err
	}
	return rows, nil
}

// ═══ المحاولات ═══

// StartAttempt يفتح محاولة جديدة، أو يرجّع المفتوحة إذا كان واقفاً بالنص.
//
// ⚠️ ما نفتح محاولة جديدة لو عنده وحدة شغّالة: الفني يوقّف بنص تمرين
// توصيل طويل ويرجع بعدين، ولو فتحنا وحدة جديدة يضيع شغله.
func (r *SimRepository) StartAttempt(exerciseID, employeeID string, isOwner bool) (*model.SimAttempt, error) {
	ex, err := r.GetExercise(exerciseID, isOwner)
	if err != nil {
		return nil, err
	}
	var open model.SimAttempt
	err = r.db.Get(&open, `
		SELECT * FROM "SimAttempt"
		WHERE "exerciseId" = $1 AND "employeeId" = $2 AND status = 'IN_PROGRESS'
		ORDER BY "startedAt" DESC LIMIT 1`, exerciseID, employeeID)
	if err == nil {
		return &open, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if ex.MaxAttempts != nil {
		var used int
		if err := r.db.Get(&used, `
			SELECT COUNT(*) FROM "SimAttempt"
			WHERE "exerciseId" = $1 AND "employeeId" = $2 AND status <> 'IN_PROGRESS'`,
			exerciseID, employeeID); err == nil && used >= *ex.MaxAttempts {
			return nil, errors.New("خلصت محاولاتك على هذا التمرين")
		}
	}

	var steps []json.RawMessage
	_ = json.Unmarshal(ex.Steps, &steps)

	var a model.SimAttempt
	err = r.db.Get(&a, `
		INSERT INTO "SimAttempt" (id, "exerciseId", "exerciseVersion", "employeeId", "stepsTotal", state)
		VALUES ($1, $2, $3, $4, $5, '{}'::jsonb) RETURNING *`,
		uuid.NewString(), exerciseID, ex.Version, employeeID, len(steps))
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// SaveProgress يحفظ حالة المشهد وأحداث الدفعة.
func (r *SimRepository) SaveProgress(attemptID, employeeID string, req model.SaveAttemptProgressRequest) error {
	res, err := r.db.Exec(`
		UPDATE "SimAttempt"
		SET state = COALESCE($3, state), "stepsPassed" = $4, "hintsUsed" = $5,
		    "wrongCount" = $6, "updatedAt" = NOW()
		WHERE id = $1 AND "employeeId" = $2 AND status = 'IN_PROGRESS'`,
		attemptID, employeeID, nullableJSON(req.State), req.StepsPassed, req.HintsUsed, req.WrongCount)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("المحاولة مو إلك أو منتهية")
	}
	return r.insertEvents(attemptID, req.Events)
}

func (r *SimRepository) insertEvents(attemptID string, events []model.SaveAttemptEventItem) error {
	if len(events) == 0 {
		return nil
	}
	// ⚠️ سقف على الدفعة: الأحداث تجي من المتصفح، وبلا سقف أي خلل
	// بالواجهة (أو عبث متعمّد) يملأ الجدول.
	if len(events) > 500 {
		events = events[:500]
	}
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for _, e := range events {
		payload := e.Payload
		if len(payload) == 0 {
			payload = json.RawMessage(`{}`)
		}
		if _, err := tx.Exec(`
			INSERT INTO "SimAttemptEvent" (id, "attemptId", "stepIndex", kind, payload, "atMs")
			VALUES ($1, $2, $3, $4, $5, $6)`,
			uuid.NewString(), attemptID, e.StepIndex, e.Kind, []byte(payload), e.AtMs); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// FinishAttempt ينهي المحاولة ويحسب الدرجة **بالسيرفر**.
//
// ⚠️ الدرجة ما تجي من الواجهة. الواجهة ترسل شنو صار (خطوات نجحت،
// تلميحات، أغلاط) والسيرفر يحسب من أوزان الخطوات المخزونة عنده. لو
// قبلنا الدرجة كما هي، أي واحد يفتح أدوات المطوّر ويرسل ١٠٠.
func (r *SimRepository) FinishAttempt(attemptID, employeeID string, req model.FinishAttemptRequest) (*model.SimAttempt, error) {
	var a model.SimAttempt
	if err := r.db.Get(&a, `
		SELECT * FROM "SimAttempt" WHERE id = $1 AND "employeeId" = $2`, attemptID, employeeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("المحاولة مو إلك")
		}
		return nil, err
	}
	if a.Status != model.SimAttemptInProgress {
		return nil, errors.New("المحاولة منتهية من قبل")
	}

	var ex model.SimExercise
	if err := r.db.Get(&ex, `SELECT * FROM "SimExercise" WHERE id = $1`, a.ExerciseID); err != nil {
		return nil, err
	}
	if err := r.insertEvents(attemptID, req.Events); err != nil {
		return nil, err
	}

	// ⚠️ التحقّق الحقيقي: نعدّ الخطوات الي عندها حدث `PASS` مسجّل فعلاً،
	// وناخذ الأقل بينها وبين الي تدّعيه الواجهة. بدون هذا، أي واحد يفتح
	// أدوات المطوّر ويرسل «نجحت كل الخطوات» ياخذ ١٠٠ بلا ما يلمس سلكاً.
	// هسه لازم يزوّر الأحداث بعد — وهاي عتبة أعلى بكثير، والسجل يبقى
	// متّسقاً مع الدرجة فينفع للمراجعة.
	var passedEvents int
	if err := r.db.Get(&passedEvents, `
		SELECT COUNT(DISTINCT "stepIndex") FROM "SimAttemptEvent"
		WHERE "attemptId" = $1 AND kind = 'PASS' AND "stepIndex" IS NOT NULL`,
		attemptID); err != nil {
		return nil, err
	}
	if req.StepsPassed > passedEvents {
		req.StepsPassed = passedEvents
	}

	// ⚠️ الدرجة تنحسب **بعد** التصحيح مو قبله — وإلا التحقّق ما ينفع شي.
	score := computeScore(ex, req)
	status := model.SimAttemptFailed
	if score >= ex.PassScore {
		status = model.SimAttemptPassed
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if err := tx.Get(&a, `
		UPDATE "SimAttempt"
		SET status = $3, score = $4, "stepsPassed" = $5, "hintsUsed" = $6, "wrongCount" = $7,
		    "durationSec" = $8, state = COALESCE($9, state), "finishedAt" = NOW(), "updatedAt" = NOW()
		WHERE id = $1 AND "employeeId" = $2 RETURNING *`,
		attemptID, employeeID, status, score, req.StepsPassed, req.HintsUsed, req.WrongCount,
		req.DurationSec, nullableJSON(req.State)); err != nil {
		return nil, err
	}

	// إتقان: أفضل نتيجة تبقى، والعدّاد يزيد.
	if _, err := tx.Exec(`
		INSERT INTO "SimMastery" (id, "employeeId", "exerciseId", "bestScore", attempts, passed, "firstPassAt", "lastAt")
		VALUES ($1, $2, $3, $4, 1, $5, CASE WHEN $5 THEN NOW() ELSE NULL END, NOW())
		ON CONFLICT ("employeeId", "exerciseId") DO UPDATE SET
			"bestScore"   = GREATEST("SimMastery"."bestScore", EXCLUDED."bestScore"),
			attempts      = "SimMastery".attempts + 1,
			passed        = "SimMastery".passed OR EXCLUDED.passed,
			"firstPassAt" = COALESCE("SimMastery"."firstPassAt", EXCLUDED."firstPassAt"),
			"lastAt"      = NOW()`,
		uuid.NewString(), employeeID, a.ExerciseID, score, status == model.SimAttemptPassed); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &a, nil
}

// computeScore يحسب الدرجة من أوزان الخطوات المخزونة بالسيرفر.
//
// المعادلة بسيطة عمداً حتى تنشرح للموظف: مجموع أوزان الخطوات الي نجحها،
// ناقص عقوبة كل تلميح وكل غلط، ومطبّعة على ١٠٠.
func computeScore(ex model.SimExercise, req model.FinishAttemptRequest) int {
	type step struct {
		Weight       int `json:"weight"`
		HintPenalty  int `json:"hintPenalty"`
		WrongPenalty int `json:"wrongPenalty"`
	}
	var steps []step
	if err := json.Unmarshal(ex.Steps, &steps); err != nil || len(steps) == 0 {
		return 0
	}
	total, hintPen, wrongPen := 0, 0, 0
	for _, s := range steps {
		w := s.Weight
		if w <= 0 {
			w = 10
		}
		total += w
		if s.HintPenalty > hintPen {
			hintPen = s.HintPenalty
		}
		if s.WrongPenalty > wrongPen {
			wrongPen = s.WrongPenalty
		}
	}
	if hintPen == 0 {
		hintPen = 5
	}
	if wrongPen == 0 {
		wrongPen = 5
	}

	// ⚠️ عدد الخطوات الناجحة ينحدّ بعدد الخطوات الفعلي: الواجهة ممكن
	// ترسل رقماً أكبر (خلل أو عبث)، والسيرفر ما يصدّقه.
	passed := req.StepsPassed
	if passed > len(steps) {
		passed = len(steps)
	}
	if passed < 0 {
		passed = 0
	}
	earned := 0
	for i := 0; i < passed; i++ {
		w := steps[i].Weight
		if w <= 0 {
			w = 10
		}
		earned += w
	}
	raw := earned - (req.HintsUsed * hintPen) - (req.WrongCount * wrongPen)
	if raw < 0 {
		raw = 0
	}
	if total == 0 {
		return 0
	}
	return raw * 100 / total
}

// MyAttempts آخر محاولات الموظف.
func (r *SimRepository) MyAttempts(employeeID string, limit int) ([]model.SimAttempt, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows := []model.SimAttempt{}
	if err := r.db.Select(&rows, `
		SELECT * FROM "SimAttempt" WHERE "employeeId" = $1
		ORDER BY "startedAt" DESC LIMIT $2`, employeeID, limit); err != nil {
		return nil, err
	}
	return rows, nil
}

func nullableJSON(v json.RawMessage) any {
	if len(v) == 0 {
		return nil
	}
	return []byte(v)
}
