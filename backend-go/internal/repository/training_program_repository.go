package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

// TrainingProgramRepository برامج التدريب — منقولة من نظام الطاقة
// الشمسية، بس تشتغل على موظفي النظام ومهاراته الموجودة.
//
// المشاركين والمهارات جداول ربط مو نص JSON بخلية (متل ما كان بالنظام
// القديم) — يعني سؤال «هذا الموظف شنو تدرّب عليه؟» ينسأل باستعلام بدل
// ما نقرا كل الصفوف ونفكّهن بالكود.
type TrainingProgramRepository struct {
	db *sqlx.DB
}

func NewTrainingProgramRepository(db *sqlx.DB) *TrainingProgramRepository {
	return &TrainingProgramRepository{db: db}
}

func (r *TrainingProgramRepository) List(status string) ([]model.TrainingProgram, error) {
	rows := []model.TrainingProgram{}
	if err := r.db.Select(&rows, `
		SELECT * FROM "TrainingProgram"
		WHERE ($1 = '' OR status = $1)
		ORDER BY COALESCE("startDate", baghdad_date("createdAt")) DESC`, status); err != nil {
		return nil, err
	}
	return rows, r.hydrate(rows)
}

func (r *TrainingProgramRepository) Find(id string) (*model.TrainingProgram, error) {
	var p model.TrainingProgram
	if err := r.db.Get(&p, `SELECT * FROM "TrainingProgram" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	one := []model.TrainingProgram{p}
	if err := r.hydrate(one); err != nil {
		return nil, err
	}
	return &one[0], nil
}

// hydrate يجيب المشاركين والمهارات وأسماء المدربين لكل البرامج سوه —
// ثلاث استعلامات ثابتة مهما كان عدد البرامج.
func (r *TrainingProgramRepository) hydrate(rows []model.TrainingProgram) error {
	if len(rows) == 0 {
		return nil
	}
	ids := make([]string, len(rows))
	instructorIDs := []string{}
	for i, p := range rows {
		ids[i] = p.ID
		if p.InstructorID != nil && *p.InstructorID != "" {
			instructorIDs = append(instructorIDs, *p.InstructorID)
		}
	}

	type partRow struct {
		ProgramID string  `db:"programId"`
		ID        string  `db:"id"`
		Name      string  `db:"name"`
		Dept      *string `db:"department"`
		JobTitle  *string `db:"jobTitle"`
		Passed    *bool   `db:"passed"`
	}
	parts := []partRow{}
	if err := r.db.Select(&parts, `
		SELECT tp."programId", e.id, e.name, e.department, e."jobTitle", tp.passed
		FROM "TrainingProgramParticipant" tp
		JOIN "Employee" e ON e.id = tp."employeeId"
		WHERE tp."programId" = ANY($1)
		ORDER BY e.name`, pq.Array(ids)); err != nil {
		return err
	}
	byProgramParts := map[string][]model.TrainingParticipant{}
	for _, p := range parts {
		byProgramParts[p.ProgramID] = append(byProgramParts[p.ProgramID], model.TrainingParticipant{
			EmployeeID: p.ID, Name: p.Name, Department: p.Dept, JobTitle: p.JobTitle, Passed: p.Passed,
		})
	}

	type skillRow struct {
		ProgramID string `db:"programId"`
		ID        string `db:"id"`
		Name      string `db:"name"`
		Category  string `db:"category"`
	}
	skills := []skillRow{}
	if err := r.db.Select(&skills, `
		SELECT ts."programId", s.id, s.name, s.category
		FROM "TrainingProgramSkill" ts
		JOIN "Skill" s ON s.id = ts."skillId"
		WHERE ts."programId" = ANY($1)
		ORDER BY s.name`, pq.Array(ids)); err != nil {
		return err
	}
	byProgramSkills := map[string][]model.TrainingSkill{}
	for _, s := range skills {
		byProgramSkills[s.ProgramID] = append(byProgramSkills[s.ProgramID], model.TrainingSkill{
			SkillID: s.ID, Name: s.Name, Category: s.Category,
		})
	}

	names := map[string]string{}
	if len(instructorIDs) > 0 {
		type nameRow struct {
			ID   string `db:"id"`
			Name string `db:"name"`
		}
		nrows := []nameRow{}
		if err := r.db.Select(&nrows, `SELECT id, name FROM "Employee" WHERE id = ANY($1)`, pq.Array(instructorIDs)); err != nil {
			return err
		}
		for _, n := range nrows {
			names[n.ID] = n.Name
		}
	}

	for i := range rows {
		rows[i].Participants = byProgramParts[rows[i].ID]
		if rows[i].Participants == nil {
			rows[i].Participants = []model.TrainingParticipant{}
		}
		rows[i].Skills = byProgramSkills[rows[i].ID]
		if rows[i].Skills == nil {
			rows[i].Skills = []model.TrainingSkill{}
		}
		if rows[i].InstructorID != nil {
			if n, ok := names[*rows[i].InstructorID]; ok {
				nn := n
				rows[i].InstructorName = &nn
			}
		}
	}
	return nil
}

// Save ينشئ أو يعدّل برنامج مع مشاركينه ومهاراته — كلها بمعاملة وحدة،
// حتى ما يصير برنامج انحفظ ومشاركينه لا.
func (r *TrainingProgramRepository) Save(id string, req model.SaveTrainingProgramRequest, byEmployeeID string) (*model.TrainingProgram, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// تاريخ النهاية ينحسب من البداية والمدة — ما ننطيه للمستخدم يكتبه
	// بالإيد لأنه يطلع مخالف للمدة ويصير عدنا رقمين ما يتفقون.
	endExpr := `CASE WHEN $5::date IS NULL THEN NULL ELSE $5::date + ($4::int - 1) END`

	if id == "" {
		id = uuid.NewString()
		if _, err := tx.Exec(`
			INSERT INTO "TrainingProgram" (
				id, name, level, "durationDays", "startDate", "endDate",
				"targetDepartment", "instructorId", objectives, content,
				"passRate", cost, status, progress, "createdById")
			VALUES ($1,$2,$3,$4,$5::date, `+endExpr+`,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
			id, req.Name, req.Level, req.DurationDays, nullIfEmpty(req.StartDate),
			nullIfEmpty(req.TargetDepartment), nullIfEmpty(req.InstructorID),
			nullIfEmpty(req.Objectives), nullIfEmpty(req.Content),
			req.PassRate, req.Cost, req.Status, req.Progress, nullIfEmpty(byEmployeeID)); err != nil {
			return nil, err
		}
	} else {
		if _, err := tx.Exec(`
			UPDATE "TrainingProgram" SET
				name = $2, level = $3, "durationDays" = $4,
				"startDate" = $5::date, "endDate" = `+endExpr+`,
				"targetDepartment" = $6, "instructorId" = $7,
				objectives = $8, content = $9, "passRate" = $10,
				cost = $11, status = $12, progress = $13, "updatedAt" = now()
			WHERE id = $1`,
			id, req.Name, req.Level, req.DurationDays, nullIfEmpty(req.StartDate),
			nullIfEmpty(req.TargetDepartment), nullIfEmpty(req.InstructorID),
			nullIfEmpty(req.Objectives), nullIfEmpty(req.Content),
			req.PassRate, req.Cost, req.Status, req.Progress); err != nil {
			return nil, err
		}
	}

	// نمسح ونكتب من جديد: أبسط وأأمن من مقارنة الفروقات، والأعداد صغيرة
	if _, err := tx.Exec(`DELETE FROM "TrainingProgramParticipant" WHERE "programId" = $1`, id); err != nil {
		return nil, err
	}
	for _, empID := range req.ParticipantIDs {
		if empID == "" {
			continue
		}
		if _, err := tx.Exec(`
			INSERT INTO "TrainingProgramParticipant" ("programId", "employeeId")
			VALUES ($1, $2) ON CONFLICT DO NOTHING`, id, empID); err != nil {
			return nil, err
		}
	}

	if _, err := tx.Exec(`DELETE FROM "TrainingProgramSkill" WHERE "programId" = $1`, id); err != nil {
		return nil, err
	}
	for _, skillID := range req.SkillIDs {
		if skillID == "" {
			continue
		}
		if _, err := tx.Exec(`
			INSERT INTO "TrainingProgramSkill" ("programId", "skillId")
			VALUES ($1, $2) ON CONFLICT DO NOTHING`, id, skillID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.Find(id)
}

// Complete يخلّص البرنامج ويمنح مهاراته لكل المشاركين.
//
// هاي الشغلة الي جانت ناقصة بالنظام القديم: «إصدار الشهادات» جان يغيّر
// حالة البرنامج بس. يعني الموظف يتدرّب على تركيب الألواح، والنظام يضل
// ما يعرف إنه يعرفها — فما يطلع بفلترة الكوادر لشغل الطاقة الشمسية.
//
// هسه إكمال البرنامج يأشّر المهارة فعلاً بملف كل مشارك.
func (r *TrainingProgramRepository) Complete(id string) (*model.TrainingProgram, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(`
		UPDATE "TrainingProgram"
		SET status = 'مكتمل', progress = 100, "updatedAt" = now()
		WHERE id = $1 AND status <> 'مكتمل'`, id)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("البرنامج مكتمل من قبل أو مو موجود")
	}

	if _, err := tx.Exec(`
		INSERT INTO "EmployeeSkill" (id, "employeeId", "skillId", "canPerform")
		SELECT gen_random_uuid()::text, p."employeeId", s."skillId", true
		FROM "TrainingProgramParticipant" p
		JOIN "TrainingProgramSkill" s ON s."programId" = p."programId"
		WHERE p."programId" = $1
		ON CONFLICT ("employeeId", "skillId") DO UPDATE SET "canPerform" = true`, id); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(`
		UPDATE "TrainingProgramParticipant" SET passed = true WHERE "programId" = $1`, id); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.Find(id)
}

func (r *TrainingProgramRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "TrainingProgram" WHERE id = $1`, id)
	return err
}
