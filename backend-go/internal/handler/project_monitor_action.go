package handler

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ إجراءات المراقب على المشاريع المتأخرة — دازاي 10-31 ═══
// «المشاريع المتأخرة تطلع بدون اتخاذ إجراء، بس طلب استفسار». جنب «اطلب تفسير»:
//   - ⏰ مهلة ومتابعة: مهمة للمشرف بموعد، والمشروع ينعلّم «المراقب متابع».
//   - 📌 تصعيد: مشكلة بمكتب الإدارة الهندسية (الهيكل الجديد) ومديرها ينبلغ، والمالك والأدمن هم.
//   - ⚖️ مخالفة: نقاط انضباط على المشرف بسبب مكتوب، وتنرجع بالتراجع.

type projectMonitorDeps struct {
	tasks      *repository.ExtraTaskRepository
	discipline *repository.DisciplineRepository
	notif      *repository.NotificationRepository
}

// SetActionDeps يُحقن من main.
func (h *ProjectQuoteWatchHandler) SetActionDeps(t *repository.ExtraTaskRepository, d *repository.DisciplineRepository, n *repository.NotificationRepository) {
	h.deps = &projectMonitorDeps{tasks: t, discipline: d, notif: n}
}

type projectMonitorAction struct {
	ID        string     `db:"id" json:"id"`
	ProjectID string     `db:"projectId" json:"projectId"`
	Action    string     `db:"action" json:"action"`
	Note      string     `db:"note" json:"note"`
	DueAt     *time.Time `db:"dueAt" json:"dueAt,omitempty"`
	Points    int        `db:"points" json:"points"`
	ByName    *string    `db:"byName" json:"byName"`
	CreatedAt time.Time  `db:"createdAt" json:"createdAt"`
}

// watchActions الإجراءات الفعّالة لكل مشروع (للعلامة «👁️ المراقب متابع»).
func (h *ProjectQuoteWatchHandler) watchActions() map[string][]projectMonitorAction {
	rows := []projectMonitorAction{}
	if err := h.db.Select(&rows, `SELECT a.id, a."projectId", a.action, a.note, a."dueAt", a.points, e.name AS "byName", a."createdAt"
		FROM "ProjectMonitorAction" a LEFT JOIN "Employee" e ON e.id = a."byId"
		WHERE a."revokedAt" IS NULL ORDER BY a."createdAt" DESC LIMIT 2000`); err != nil {
		log.Printf("project monitor actions: %v", err)
	}
	out := map[string][]projectMonitorAction{}
	for _, r := range rows {
		out[r.ProjectID] = append(out[r.ProjectID], r)
	}
	return out
}

// GET /api/projects/monitor-watch — علامة «المراقب متابع» بشاشة المشاريع (الي عدهم المشاريع).
func (h *ProjectQuoteWatchHandler) WatchList(w http.ResponseWriter, r *http.Request) {
	type item struct {
		Action string     `json:"action"`
		At     time.Time  `json:"at"`
		DueAt  *time.Time `json:"dueAt,omitempty"`
	}
	out := map[string]item{}
	for pid, as := range h.watchActions() {
		out[pid] = item{Action: as[0].Action, At: as[0].CreatedAt, DueAt: as[0].DueAt}
	}
	WriteJSON(w, http.StatusOK, out)
}

// POST /api/project-quote-watch/{id}/action {action, note, hours, points}
func (h *ProjectQuoteWatchHandler) Act(w http.ResponseWriter, r *http.Request) {
	if h.deps == nil {
		WriteError(w, http.StatusServiceUnavailable, "الإجراءات مو جاهزة")
		return
	}
	var b struct {
		Action string `json:"action"`
		Note   string `json:"note"`
		Hours  int    `json:"hours"`
		Points int    `json:"points"`
	}
	if err := DecodeJSON(r, &b); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	note := strings.TrimSpace(b.Note)
	var p struct {
		ID     string  `db:"id"`
		Code   string  `db:"code"`
		Name   string  `db:"name"`
		Stage  string  `db:"stage"`
		Sup    *string `db:"sup"`
		SupRol *string `db:"supRole"`
	}
	if err := h.db.Get(&p, `SELECT p.id, p.code, p.name, p.stage,
		COALESCE(p."delegatedToEmployeeId", p."responsibleEmployeeId") AS sup, e.role::text AS "supRole"
		FROM "Project" p LEFT JOIN "Employee" e ON e.id = COALESCE(p."delegatedToEmployeeId", p."responsibleEmployeeId")
		WHERE p.id = $1`, r.PathValue("id")); err != nil {
		WriteError(w, http.StatusNotFound, "المشروع مو موجود")
		return
	}
	me := middleware.EmployeeIDFromContext(r)
	what := fmt.Sprintf("مشروع %s (%s) — %s", p.Code, p.Name, p.Stage)
	extra := ""
	if note != "" {
		extra = " — ملاحظة المراقب: " + note
	}
	var dueAt *time.Time
	var taskID, issueID *string
	points := 0
	done := ""
	switch b.Action {
	case "FOLLOW_UP":
		if p.Sup == nil {
			WriteError(w, http.StatusBadRequest, "المشروع ما إله مشرف — عيّن مشرف أول")
			return
		}
		if b.Hours < 4 || b.Hours > 168 {
			b.Hours = 24
		}
		due := time.Now().Add(time.Duration(b.Hours) * time.Hour)
		dueAt = &due
		dueS := due.Format(time.RFC3339)
		desc := "متابعة من المراقب: " + what + extra
		t, err := h.deps.tasks.Create(model.CreateExtraTaskRequest{Title: "متابعة مشروع " + p.Code, Description: &desc,
			AssignedToID: *p.Sup, Priority: model.ExtraTaskUrgent, DueAt: &dueS}, me)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "تعذر إنشاء المهلة: "+err.Error())
			return
		}
		taskID = &t.ID
		_ = h.deps.notif.Create(*p.Sup, "extra_task", fmt.Sprintf("⏰ المراقب حطلك مهلة %d ساعة على %s%s", b.Hours, what, extra))
		done = fmt.Sprintf("انحطت مهلة %d ساعة للمشرف — والمشروع صار «👁️ المراقب متابع»", b.Hours)
	case "ESCALATE":
		if utf8.RuneCountInString(note) < 5 {
			WriteError(w, http.StatusBadRequest, "اكتب سبب التصعيد")
			return
		}
		var clause *int
		var c int
		if err := h.db.Get(&c, `SELECT "principleNo" FROM "PrivacyPolicyPoint" WHERE content LIKE 'م2/ف1%' AND "principleNo" IS NOT NULL LIMIT 1`); err == nil {
			clause = &c
		}
		var id string
		if err := h.db.Get(&id, `INSERT INTO "OrgIssueReport" ("divisionId", "unitId", title, problem, impact, scope, principle, severity, source, "authorId")
			VALUES ('eng', 'eng.projects', $1, $2, 'تأخير على الزبون وعلى عرض السعر', 'SPECIFIC', $3, 'HIGH', 'MONITOR', $4) RETURNING id`,
			"👁️ تصعيد من المراقب: مشروع "+p.Code+" متأخر", what+" — "+note, clause, me); err != nil {
			log.Printf("monitor escalate issue: %v", err)
			WriteError(w, http.StatusInternalServerError, "تعذر فتح المشكلة بمكتب الإدارة")
			return
		}
		issueID = &id
		heads := []string{}
		_ = h.db.Select(&heads, `SELECT DISTINCT "employeeId" FROM "OrgAssignment" WHERE "endedAt" IS NULL
			AND position IN ('DIRECTOR', 'DEPUTY') AND "divisionId" = 'eng' AND ("unitId" IS NULL OR "unitId" = 'eng.projects')`)
		for _, hd := range heads {
			_ = h.deps.notif.Create(hd, "org_issue", "📌 المراقب صعّد إلك: "+what+" — "+note+". اكتب خطة الحل بمكتب الإدارة.")
		}
		_ = h.deps.notif.CreateForRolesOrPermission([]string{"OWNER", "ADMIN"}, "", "monitor_escalation", "⬆️ تصعيد من المراقب — "+what+" — "+note)
		done = "انصعّد: انفتحت مشكلة بمكتب الإدارة الهندسية ومديرها انبلغ، والمالك والأدمن هم"
	case "VIOLATION":
		if p.Sup == nil {
			WriteError(w, http.StatusBadRequest, "المشروع ما إله مشرف")
			return
		}
		if p.SupRol != nil && (*p.SupRol == "ADMIN" || *p.SupRol == "OWNER") {
			WriteError(w, http.StatusBadRequest, "الإدارة العليا ما تنسجل عليها مخالفات")
			return
		}
		if utf8.RuneCountInString(note) < 5 {
			WriteError(w, http.StatusBadRequest, "اكتب سبب المخالفة")
			return
		}
		if b.Points < 1 || b.Points > 10 {
			WriteError(w, http.StatusBadRequest, "النقاط من ١ لـ١٠ (النقطة = ١٠ آلاف)")
			return
		}
		_, applied, err := h.deps.discipline.AdjustKind(*p.Sup, -b.Points, model.DisciplineMonitorFine, "مخالفة من المراقب: "+what+" — "+note, me)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "تعذر تسجيل المخالفة")
			return
		}
		points = -applied
		if points < 0 {
			points = -points
		}
		_ = h.deps.notif.Create(*p.Sup, "discipline", fmt.Sprintf("⚖️ انسجلت عليك مخالفة من المراقب (%d نقطة): %s — %s", points, what, note))
		done = fmt.Sprintf("انسجلت المخالفة: %d نقطة على المشرف", points)
	default:
		WriteError(w, http.StatusBadRequest, "إجراء غير معروف")
		return
	}
	if _, err := h.db.Exec(`INSERT INTO "ProjectMonitorAction" ("projectId", "employeeId", action, note, "dueAt", points, "taskId", "issueId", "byId")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`, p.ID, p.Sup, b.Action, note, dueAt, points, taskID, issueID, me); err != nil {
		log.Printf("project monitor action log: %v", err)
	}
	WriteJSON(w, http.StatusOK, map[string]string{"done": done})
}

// POST /api/project-monitor-actions/{id}/revoke — تراجع؛ المخالفة ترجع نقاطها.
func (h *ProjectQuoteWatchHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	if h.deps == nil {
		WriteError(w, http.StatusServiceUnavailable, "الإجراءات مو جاهزة")
		return
	}
	me := middleware.EmployeeIDFromContext(r)
	var a struct {
		Action string  `db:"action"`
		Emp    *string `db:"employeeId"`
		Points int     `db:"points"`
		Code   string  `db:"code"`
	}
	if err := h.db.Get(&a, `UPDATE "ProjectMonitorAction" pa SET "revokedAt" = now(), "revokedById" = $2
		FROM "Project" p WHERE pa.id = $1 AND pa."revokedAt" IS NULL AND p.id = pa."projectId"
		RETURNING pa.action, pa."employeeId", pa.points, p.code`, r.PathValue("id"), me); err != nil {
		WriteError(w, http.StatusNotFound, "الإجراء مو موجود أو متراجع عنه")
		return
	}
	if a.Action == "VIOLATION" && a.Emp != nil && a.Points > 0 {
		if _, _, err := h.deps.discipline.AdjustKind(*a.Emp, a.Points, model.DisciplineMonitorFineUndo, "تراجع المراقب عن مخالفة مشروع "+a.Code, me); err != nil {
			log.Printf("monitor fine undo: %v", err)
		}
		_ = h.deps.notif.Create(*a.Emp, "discipline", fmt.Sprintf("↩️ المراقب تراجع عن مخالفة مشروع %s — رجعتلك %d نقطة", a.Code, a.Points))
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
