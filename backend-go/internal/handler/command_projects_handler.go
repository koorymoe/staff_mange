package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
)

// ═══ 🏗️ مركز القيادة — المشاريع والمقاولات (دازاي 11-01) ═══
// معزول عن نظام الشركة: جدول «CmdProject» لحاله، والمسارات ورا RequireCommandRealm.
// الإدخال سطر أوامر بالواجهة — يتفكك محلياً، والسيرفر يتحقق ويحفظ ويسجّل الأمر.

type CommandProjectsHandler struct{ db *sqlx.DB }

func NewCommandProjectsHandler(db *sqlx.DB) *CommandProjectsHandler {
	return &CommandProjectsHandler{db: db}
}

type cmdProject struct {
	ID             string     `db:"id" json:"id"`
	Name           string     `db:"name" json:"name"`
	Location       string     `db:"location" json:"location"`
	Owner          string     `db:"owner" json:"owner"`
	ContractIQD    float64    `db:"contractIQD" json:"contractIQD"`
	DurationMonths *int       `db:"durationMonths" json:"durationMonths"`
	StartDate      *time.Time `db:"startDate" json:"startDate"`
	Stage          string     `db:"stage" json:"stage"`
	Progress       int        `db:"progress" json:"progress"`
	CreatedByID    *string    `db:"createdById" json:"-"`
	CreatedAt      time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time  `db:"updatedAt" json:"updatedAt"`
}

var cmdStages = map[string]bool{"TENDER": true, "SIGNED": true, "EXECUTION": true, "HANDOVER": true, "CLOSED": true}

type cmdProjectIn struct {
	Command        string   `json:"command"`
	Name           *string  `json:"name"`
	Location       *string  `json:"location"`
	Owner          *string  `json:"owner"`
	ContractIQD    *float64 `json:"contractIQD"`
	DurationMonths *int     `json:"durationMonths"`
	StartDate      *string  `json:"startDate"`
	Stage          *string  `json:"stage"`
	Progress       *int     `json:"progress"`
}

func (in *cmdProjectIn) validate() string {
	if in.Name != nil && len([]rune(strings.TrimSpace(*in.Name))) < 2 {
		return "اسم المشروع قصير"
	}
	if in.ContractIQD != nil && *in.ContractIQD < 0 {
		return "قيمة العقد ما تكون سالبة"
	}
	if in.DurationMonths != nil && (*in.DurationMonths < 1 || *in.DurationMonths > 240) {
		return "المدة بين شهر و٢٤٠ شهر"
	}
	if in.StartDate != nil && *in.StartDate != "" {
		if _, err := time.Parse("2006-01-02", *in.StartDate); err != nil {
			return "تاريخ البدء لازم YYYY-MM-DD"
		}
	}
	if in.Stage != nil && !cmdStages[*in.Stage] {
		return "مرحلة مو معروفة"
	}
	if in.Progress != nil && (*in.Progress < 0 || *in.Progress > 100) {
		return "الإنجاز بين ٠ و١٠٠"
	}
	return ""
}

func nullDate(s *string) any {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}

func (h *CommandProjectsHandler) log(projectID *string, emp, cmd, result string) {
	if strings.TrimSpace(cmd) == "" {
		return
	}
	_, _ = h.db.Exec(`INSERT INTO "CmdProjectLog" ("projectId", "employeeId", command, result) VALUES ($1, $2, $3, $4)`,
		projectID, emp, cmd, result)
}

// GET /api/command/projects
func (h *CommandProjectsHandler) List(w http.ResponseWriter, r *http.Request) {
	rows := []cmdProject{}
	if err := h.db.Select(&rows, `SELECT * FROM "CmdProject" ORDER BY "updatedAt" DESC`); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر التحميل")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/command/projects
func (h *CommandProjectsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in cmdProjectIn
	if err := DecodeJSON(r, &in); err != nil || in.Name == nil {
		WriteError(w, http.StatusBadRequest, "اسم المشروع مطلوب")
		return
	}
	if msg := in.validate(); msg != "" {
		WriteError(w, http.StatusBadRequest, msg)
		return
	}
	str := func(p *string) string {
		if p == nil {
			return ""
		}
		return strings.TrimSpace(*p)
	}
	stage, progress, contract := "TENDER", 0, 0.0
	if in.Stage != nil {
		stage = *in.Stage
	}
	if in.Progress != nil {
		progress = *in.Progress
	}
	if in.ContractIQD != nil {
		contract = *in.ContractIQD
	}
	me := middleware.EmployeeIDFromContext(r)
	var p cmdProject
	err := h.db.Get(&p, `INSERT INTO "CmdProject" (name, location, owner, "contractIQD", "durationMonths", "startDate", stage, progress, "createdById")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *`,
		str(in.Name), str(in.Location), str(in.Owner), contract, in.DurationMonths, nullDate(in.StartDate), stage, progress, me)
	if err != nil {
		if strings.Contains(err.Error(), "CmdProject_name_uq") {
			WriteError(w, http.StatusConflict, "أكو مشروع بنفس الاسم")
			return
		}
		WriteError(w, http.StatusBadRequest, "تعذر الحفظ")
		return
	}
	h.log(&p.ID, me, in.Command, "انضاف مشروع "+p.Name)
	WriteJSON(w, http.StatusCreated, p)
}

// PATCH /api/command/projects/{id}
func (h *CommandProjectsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var in cmdProjectIn
	if err := DecodeJSON(r, &in); err != nil {
		WriteError(w, http.StatusBadRequest, "طلب غير صالح")
		return
	}
	if msg := in.validate(); msg != "" {
		WriteError(w, http.StatusBadRequest, msg)
		return
	}
	id, me := r.PathValue("id"), middleware.EmployeeIDFromContext(r)
	var p cmdProject
	err := h.db.Get(&p, `UPDATE "CmdProject" SET
			name = COALESCE(NULLIF(trim($2), ''), name), location = COALESCE($3, location), owner = COALESCE($4, owner),
			"contractIQD" = COALESCE($5, "contractIQD"), "durationMonths" = COALESCE($6, "durationMonths"),
			"startDate" = COALESCE($7::date, "startDate"), stage = COALESCE($8, stage), progress = COALESCE($9, progress),
			"updatedAt" = now()
		WHERE id = $1 RETURNING *`,
		id, in.Name, in.Location, in.Owner, in.ContractIQD, in.DurationMonths, nullDate(in.StartDate), in.Stage, in.Progress)
	if err != nil {
		WriteError(w, http.StatusNotFound, "المشروع مو موجود")
		return
	}
	h.log(&p.ID, me, in.Command, "انحدّث "+p.Name)
	WriteJSON(w, http.StatusOK, p)
}
