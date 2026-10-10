package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
)

// ═══ مركز القيادة: الأهداف والخطة (قرار (ع) 10-10) ═══
// المسارات كلها تحت /api/command/ ومحروسة بـRequireCommandRealm — توكن نظام
// الشركة ما يوصلها، وتوكن القيادة ما يوصل غيرها.
type CommandGoalsHandler struct{ db *sqlx.DB }

func NewCommandGoalsHandler(db *sqlx.DB) *CommandGoalsHandler { return &CommandGoalsHandler{db: db} }

type commandStep struct {
	ID      string     `db:"id" json:"id"`
	GoalID  string     `db:"goalId" json:"goalId"`
	Title   string     `db:"title" json:"title"`
	DueDate *string    `db:"dueDate" json:"dueDate"`
	DoneAt  *time.Time `db:"doneAt" json:"doneAt"`
}

type commandGoal struct {
	ID          string        `db:"id" json:"id"`
	Title       string        `db:"title" json:"title"`
	Description *string       `db:"description" json:"description"`
	Category    string        `db:"category" json:"category"`
	OwnerLabel  *string       `db:"ownerLabel" json:"ownerLabel"`
	TargetDate  *string       `db:"targetDate" json:"targetDate"`
	Status      string        `db:"status" json:"status"`
	CreatedAt   time.Time     `db:"createdAt" json:"createdAt"`
	Steps       []commandStep `db:"-" json:"steps"`
}

var goalStatuses = map[string]bool{"ACTIVE": true, "DONE": true, "PAUSED": true, "DROPPED": true}

func datePtr(s *string) *string {
	s = trimPtr(s)
	if s == nil || !washDateRe.MatchString(*s) {
		return nil
	}
	return s
}

// GET /api/command/goals
func (h *CommandGoalsHandler) List(w http.ResponseWriter, r *http.Request) {
	goals := []commandGoal{}
	if err := h.db.Select(&goals, `SELECT id, title, description, category, "ownerLabel",
		to_char("targetDate", 'YYYY-MM-DD') AS "targetDate", status, "createdAt"
		FROM "CommandGoal" ORDER BY (status = 'ACTIVE') DESC, "targetDate" NULLS LAST, "createdAt" DESC`); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الأهداف")
		return
	}
	steps := []commandStep{}
	if err := h.db.Select(&steps, `SELECT id, "goalId", title, to_char("dueDate", 'YYYY-MM-DD') AS "dueDate", "doneAt"
		FROM "CommandGoalStep" ORDER BY "dueDate" NULLS LAST, "createdAt"`); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المراحل")
		return
	}
	by := map[string][]commandStep{}
	for _, s := range steps {
		by[s.GoalID] = append(by[s.GoalID], s)
	}
	for i := range goals {
		goals[i].Steps = by[goals[i].ID]
		if goals[i].Steps == nil {
			goals[i].Steps = []commandStep{}
		}
	}
	WriteJSON(w, http.StatusOK, goals)
}

type goalInput struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Category    string  `json:"category"`
	OwnerLabel  *string `json:"ownerLabel"`
	TargetDate  *string `json:"targetDate"`
	Status      string  `json:"status"`
}

// POST /api/command/goals
func (h *CommandGoalsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in goalInput
	if err := DecodeJSON(r, &in); err != nil || strings.TrimSpace(in.Title) == "" {
		WriteError(w, http.StatusBadRequest, "اكتب الهدف")
		return
	}
	cat := strings.TrimSpace(in.Category)
	if cat == "" {
		cat = "عام"
	}
	if _, err := h.db.Exec(`INSERT INTO "CommandGoal" (id, title, description, category, "ownerLabel", "targetDate", "createdById")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5::date, $6)`,
		strings.TrimSpace(in.Title), trimPtr(in.Description), cat, trimPtr(in.OwnerLabel), datePtr(in.TargetDate),
		middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	h.List(w, r)
}

// PUT /api/command/goals/{id}
func (h *CommandGoalsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var in goalInput
	if err := DecodeJSON(r, &in); err != nil || strings.TrimSpace(in.Title) == "" {
		WriteError(w, http.StatusBadRequest, "اكتب الهدف")
		return
	}
	if !goalStatuses[in.Status] {
		in.Status = "ACTIVE"
	}
	cat := strings.TrimSpace(in.Category)
	if cat == "" {
		cat = "عام"
	}
	if _, err := h.db.Exec(`UPDATE "CommandGoal" SET title=$2, description=$3, category=$4, "ownerLabel"=$5,
		"targetDate"=$6::date, status=$7 WHERE id=$1`, r.PathValue("id"), strings.TrimSpace(in.Title),
		trimPtr(in.Description), cat, trimPtr(in.OwnerLabel), datePtr(in.TargetDate), in.Status); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	h.List(w, r)
}

// DELETE /api/command/goals/{id}
func (h *CommandGoalsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if _, err := h.db.Exec(`DELETE FROM "CommandGoal" WHERE id=$1`, r.PathValue("id")); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحذف")
		return
	}
	h.List(w, r)
}

// POST /api/command/goals/{id}/steps
func (h *CommandGoalsHandler) AddStep(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title   string  `json:"title"`
		DueDate *string `json:"dueDate"`
	}
	if err := DecodeJSON(r, &in); err != nil || strings.TrimSpace(in.Title) == "" {
		WriteError(w, http.StatusBadRequest, "اكتب المرحلة")
		return
	}
	if _, err := h.db.Exec(`INSERT INTO "CommandGoalStep" (id, "goalId", title, "dueDate")
		SELECT gen_random_uuid()::text, id, $2, $3::date FROM "CommandGoal" WHERE id=$1`,
		r.PathValue("id"), strings.TrimSpace(in.Title), datePtr(in.DueDate)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	h.List(w, r)
}

// PUT /api/command/steps/{id}/toggle — تمّت / ما تمّت
func (h *CommandGoalsHandler) ToggleStep(w http.ResponseWriter, r *http.Request) {
	if _, err := h.db.Exec(`UPDATE "CommandGoalStep" SET "doneAt" = CASE WHEN "doneAt" IS NULL THEN now() END WHERE id=$1`,
		r.PathValue("id")); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	h.List(w, r)
}

// DELETE /api/command/steps/{id}
func (h *CommandGoalsHandler) DeleteStep(w http.ResponseWriter, r *http.Request) {
	if _, err := h.db.Exec(`DELETE FROM "CommandGoalStep" WHERE id=$1`, r.PathValue("id")); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحذف")
		return
	}
	h.List(w, r)
}
