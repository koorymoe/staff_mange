package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// ExtraTaskRepository المهام الإضافية — شغل موجّه مو مربوط بحجز.
type ExtraTaskRepository struct {
	db *sqlx.DB
}

func NewExtraTaskRepository(db *sqlx.DB) *ExtraTaskRepository {
	return &ExtraTaskRepository{db: db}
}

func (r *ExtraTaskRepository) Create(in model.CreateExtraTaskRequest, byEmployeeID string) (*model.ExtraTask, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, errors.New("اكتب عنوان المهمة")
	}
	if strings.TrimSpace(in.AssignedToID) == "" {
		return nil, errors.New("اختار الموظف الي تريد توجّهه")
	}
	priority := model.ExtraTaskNormal
	if in.Priority == model.ExtraTaskUrgent {
		priority = model.ExtraTaskUrgent
	}
	var due *time.Time
	if in.DueAt != nil && strings.TrimSpace(*in.DueAt) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(*in.DueAt))
		if err != nil {
			return nil, errors.New("صيغة الموعد مو صحيحة")
		}
		due = &t
	}

	var row model.ExtraTask
	err := r.db.Get(&row, `
		INSERT INTO "ExtraTask" (id, title, description, "assignedToId", "assignedById", priority, "dueAt")
		VALUES ($1,$2,NULLIF($3,''),$4,$5,$6,$7)
		RETURNING *`,
		uuid.NewString(), title, derefStrLocal(in.Description), in.AssignedToID, byEmployeeID, priority, due)
	if err != nil {
		return nil, err
	}
	r.hydrate([]*model.ExtraTask{&row})
	return &row, nil
}

// ListForEmployee مهام موظف — الي انتوجّهت له هو.
//
// المنجزة والملغاة تنجي بعد: الموظف لازم يشوف شغله الي خلّصه، مو
// يختفي أول ما يضغط «تم» فيحس إن شغله ضاع.
func (r *ExtraTaskRepository) ListForEmployee(employeeID string, includeDone bool) ([]model.ExtraTask, error) {
	q := `SELECT * FROM "ExtraTask" WHERE "assignedToId" = $1`
	if !includeDone {
		q += ` AND status IN ('NEW','IN_PROGRESS')`
	}
	// المستعجل أول، بعدين الأقدم موعداً — الي بلا موعد بالآخر.
	q += ` ORDER BY (status IN ('NEW','IN_PROGRESS')) DESC,
	                (priority = 'URGENT') DESC,
	                "dueAt" ASC NULLS LAST, "createdAt" DESC LIMIT 300`
	rows := []model.ExtraTask{}
	if err := r.db.Select(&rows, q, employeeID); err != nil {
		return nil, err
	}
	r.hydrateSlice(rows)
	return rows, nil
}

// ListAll للمدير — كل المهام مع فلاتر.
func (r *ExtraTaskRepository) ListAll(status, assigneeID string) ([]model.ExtraTask, error) {
	where := []string{}
	args := []any{}
	if status != "" {
		args = append(args, status)
		where = append(where, fmt.Sprintf(`status = $%d`, len(args)))
	}
	if assigneeID != "" {
		args = append(args, assigneeID)
		where = append(where, fmt.Sprintf(`"assignedToId" = $%d`, len(args)))
	}
	q := `SELECT * FROM "ExtraTask"`
	if len(where) > 0 {
		q += ` WHERE ` + strings.Join(where, " AND ")
	}
	q += ` ORDER BY (status IN ('NEW','IN_PROGRESS')) DESC,
	                (priority = 'URGENT') DESC,
	                "dueAt" ASC NULLS LAST, "createdAt" DESC LIMIT 400`
	rows := []model.ExtraTask{}
	if err := r.db.Select(&rows, q, args...); err != nil {
		return nil, err
	}
	r.hydrateSlice(rows)
	return rows, nil
}

func (r *ExtraTaskRepository) FindByID(id string) (*model.ExtraTask, error) {
	var row model.ExtraTask
	if err := r.db.Get(&row, `SELECT * FROM "ExtraTask" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	r.hydrate([]*model.ExtraTask{&row})
	return &row, nil
}

// MarkSeen يأشّر إن الموظف فتحها — «ما وصلني» ما عاد إلها محل.
func (r *ExtraTaskRepository) MarkSeen(id, employeeID string) error {
	_, err := r.db.Exec(`
		UPDATE "ExtraTask" SET "seenAt" = COALESCE("seenAt", now()), "updatedAt" = now()
		WHERE id = $1 AND "assignedToId" = $2`, id, employeeID)
	return err
}

// Start الموظف يبدي المهمة.
//
// ⚠️ الشرط على assignedToId بكل عملية: بدونه أي موظف يكدر يبدي أو
// ينهي مهمة موظف ثاني بمعرّفها. الحارس بالمسار يفحص إنه موظف، مو
// إنها **مهمته هو**.
func (r *ExtraTaskRepository) Start(id, employeeID string) error {
	res, err := r.db.Exec(`
		UPDATE "ExtraTask"
		SET status = 'IN_PROGRESS', "startedAt" = COALESCE("startedAt", now()),
		    "seenAt" = COALESCE("seenAt", now()), "updatedAt" = now()
		WHERE id = $1 AND "assignedToId" = $2 AND status = 'NEW'`, id, employeeID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("المهمة مو إلك أو بديتها من قبل")
	}
	return nil
}

// Complete الموظف ينهي المهمة بوصف شنو سوّى.
func (r *ExtraTaskRepository) Complete(id, employeeID, doneNote string) error {
	res, err := r.db.Exec(`
		UPDATE "ExtraTask"
		SET status = 'DONE', "doneAt" = now(), "doneNote" = $3,
		    "startedAt" = COALESCE("startedAt", now()),
		    "seenAt" = COALESCE("seenAt", now()), "updatedAt" = now()
		WHERE id = $1 AND "assignedToId" = $2 AND status IN ('NEW','IN_PROGRESS')`,
		id, employeeID, doneNote)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("المهمة مو إلك أو انخلصت من قبل")
	}
	return nil
}

// Cancel المدير يلغي المهمة بسبب.
func (r *ExtraTaskRepository) Cancel(id, reason string) error {
	res, err := r.db.Exec(`
		UPDATE "ExtraTask"
		SET status = 'CANCELLED', "cancelledAt" = now(), "cancelReason" = $2, "updatedAt" = now()
		WHERE id = $1 AND status IN ('NEW','IN_PROGRESS')`, id, reason)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("المهمة منجزة أو ملغاة من قبل")
	}
	return nil
}

// OpenCountForEmployee عدد المهام المفتوحة — للشارة بالقائمة.
func (r *ExtraTaskRepository) OpenCountForEmployee(employeeID string) (int, error) {
	var n int
	err := r.db.Get(&n, `
		SELECT COUNT(*) FROM "ExtraTask"
		WHERE "assignedToId" = $1 AND status IN ('NEW','IN_PROGRESS')`, employeeID)
	return n, err
}

// ═══ الهدرجة ═══

func (r *ExtraTaskRepository) hydrateSlice(rows []model.ExtraTask) {
	ptrs := make([]*model.ExtraTask, len(rows))
	for i := range rows {
		ptrs[i] = &rows[i]
	}
	r.hydrate(ptrs)
}

func (r *ExtraTaskRepository) hydrate(rows []*model.ExtraTask) {
	if len(rows) == 0 {
		return
	}
	ids := map[string]bool{}
	for _, t := range rows {
		ids[t.AssignedToID] = true
		if t.AssignedByID != nil {
			ids[*t.AssignedByID] = true
		}
	}
	list := make([]string, 0, len(ids))
	for id := range ids {
		list = append(list, id)
	}
	names := map[string]string{}
	if q, args, err := sqlx.In(`SELECT id, name FROM "Employee" WHERE id IN (?)`, list); err == nil {
		briefs := []model.EmployeeBrief{}
		if err := r.db.Select(&briefs, r.db.Rebind(q), args...); err == nil {
			for _, b := range briefs {
				names[b.ID] = b.Name
			}
		}
	}
	now := time.Now()
	for _, t := range rows {
		if n, ok := names[t.AssignedToID]; ok {
			nn := n
			t.AssignedToName = &nn
		}
		if t.AssignedByID != nil {
			if n, ok := names[*t.AssignedByID]; ok {
				nn := n
				t.AssignedByName = &nn
			}
		}
		// ⚠️ «متأخرة» تنحسب بس للمفتوحة: المنجزة بعد موعدها انخلصت
		// خلاص، وتأشيرها متأخرة للأبد يخلي القائمة كلها حمرة.
		t.Overdue = t.DueAt != nil && t.DueAt.Before(now) &&
			(t.Status == model.ExtraTaskNew || t.Status == model.ExtraTaskInProgress)
	}
}

// derefStrLocal يفك المؤشر لنص فاضي — NULLIF بالاستعلام يحوّله NULL.
func derefStrLocal(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}
