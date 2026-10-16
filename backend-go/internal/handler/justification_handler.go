package handler

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
)

// ═══ تبرير الموظف — قرار (ع) 10-13 ═══
// المراقب يطلب تبرير ← الموظف يشوفه أول ما يفتح النظام ويكتب عذره ←
// يرجع للمراقب بصندوقه (تبويب التبريرات) ← يحكم مقبول/مرفوض ← النتيجة توصل
// للموظف وتبقى محفوظة. بلا غرامة ولا نقاط — توثيق ومتابعة بس.
type JustificationHandler struct {
	db    *sqlx.DB
	notif *repository.NotificationRepository
}

func NewJustificationHandler(db *sqlx.DB, n *repository.NotificationRepository) *JustificationHandler {
	return &JustificationHandler{db: db, notif: n}
}

// Request ينادى من «اتخاذ إجراء ← طلب تبرير».
func (h *JustificationHandler) Request(reviewID, employeeID, monitorID, question string) error {
	_, err := h.db.Exec(`INSERT INTO "MonitorJustification" (id, "reviewId", "employeeId", "monitorId", question) VALUES ($1,NULLIF($2,''),$3,$4,$5)`,
		uuid.NewString(), reviewID, employeeID, monitorID, question)
	return err
}

type justificationRow struct {
	ID           string     `db:"id" json:"id"`
	EmployeeID   string     `db:"employeeId" json:"employeeId"`
	EmployeeName string     `db:"employeeName" json:"employeeName"`
	MonitorName  *string    `db:"monitorName" json:"monitorName"`
	Question     string     `db:"question" json:"question"`
	Answer       *string    `db:"answer" json:"answer"`
	AnsweredAt   *time.Time `db:"answeredAt" json:"answeredAt"`
	Verdict      *string    `db:"verdict" json:"verdict"`
	VerdictNote  *string    `db:"verdictNote" json:"verdictNote"`
	DecidedAt    *time.Time `db:"decidedAt" json:"decidedAt"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`
}

const justSelect = `SELECT j.id, j."employeeId", e.name AS "employeeName", m.name AS "monitorName", j.question, j.answer,
	j."answeredAt", j.verdict, j."verdictNote", j."decidedAt", j."createdAt"
	FROM "MonitorJustification" j JOIN "Employee" e ON e.id = j."employeeId" LEFT JOIN "Employee" m ON m.id = j."monitorId"`

// GET /api/justifications/mine — الي مطلوبة مني + آخر النتائج
func (h *JustificationHandler) Mine(w http.ResponseWriter, r *http.Request) {
	rows := []justificationRow{}
	_ = h.db.Select(&rows, justSelect+` WHERE j."employeeId" = $1 AND (j."answeredAt" IS NULL OR j."createdAt" > now() - interval '30 days')
		ORDER BY (j."answeredAt" IS NULL) DESC, j."createdAt" DESC LIMIT 30`, middleware.EmployeeIDFromContext(r))
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/justifications/{id}/answer {answer}
func (h *JustificationHandler) Answer(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Answer string `json:"answer"`
	}
	if err := DecodeJSON(r, &b); err != nil || len([]rune(strings.TrimSpace(b.Answer))) < 3 {
		WriteError(w, http.StatusBadRequest, "اكتب عذرك")
		return
	}
	var monitorID *string
	var question string
	err := h.db.QueryRowx(`UPDATE "MonitorJustification" SET answer = $3, "answeredAt" = now()
		WHERE id = $1 AND "employeeId" = $2 AND "answeredAt" IS NULL RETURNING "monitorId", question`,
		r.PathValue("id"), middleware.EmployeeIDFromContext(r), strings.TrimSpace(b.Answer)).Scan(&monitorID, &question)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "التبرير منطي قبل أو مو إلك")
		return
	}
	if monitorID != nil {
		_ = h.notif.Create(*monitorID, "monitor_justify_answer", "📝 وصل تبرير على: "+question+" — شوفه بصندوق المراقب ← التبريرات")
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/justifications — للمراقب: كل الطلبات وحالتها
func (h *JustificationHandler) List(w http.ResponseWriter, r *http.Request) {
	rows := []justificationRow{}
	_ = h.db.Select(&rows, justSelect+` WHERE j."createdAt" > now() - interval '60 days'
		ORDER BY (j."answeredAt" IS NOT NULL AND j.verdict IS NULL) DESC, j."createdAt" DESC LIMIT 200`)
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/justifications/{id}/decide {verdict: ACCEPTED|REJECTED, note}
func (h *JustificationHandler) Decide(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Verdict string `json:"verdict"`
		Note    string `json:"note"`
	}
	if err := DecodeJSON(r, &b); err != nil || (b.Verdict != "ACCEPTED" && b.Verdict != "REJECTED") {
		WriteError(w, http.StatusBadRequest, "اختار مقبول لو مرفوض")
		return
	}
	var empID, question string
	err := h.db.QueryRowx(`UPDATE "MonitorJustification" SET verdict = $2, "verdictNote" = $3, "decidedAt" = now()
		WHERE id = $1 AND "answeredAt" IS NOT NULL RETURNING "employeeId", question`,
		r.PathValue("id"), b.Verdict, strings.TrimSpace(b.Note)).Scan(&empID, &question)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "ما ينحكم قبل ما الموظف يبرّر")
		return
	}
	res := "✅ المراقب قبل تبريرك"
	if b.Verdict == "REJECTED" {
		res = "❌ المراقب ما قبل تبريرك"
	}
	if n := strings.TrimSpace(b.Note); n != "" {
		res += " — " + n
	}
	_ = h.notif.Create(empID, "monitor_justify_result", res+" (على: "+question+")")
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/justifications/direct {employeeId, question} — (ع 10-16) من خانة
// «المتأخر»: المراقب يطلب تفسير مباشرة (مشرف ما تواصل، محاسب تأخر بالتدقيق…)
// بلا صف بالصندوق. نفس المسار بعدها: الموظف يجاوب، والمراقب يحكم.
func (h *JustificationHandler) Direct(w http.ResponseWriter, r *http.Request) {
	var b struct {
		EmployeeID string `json:"employeeId"`
		Question   string `json:"question"`
	}
	if err := DecodeJSON(r, &b); err != nil || b.EmployeeID == "" || len([]rune(strings.TrimSpace(b.Question))) < 5 {
		WriteError(w, http.StatusBadRequest, "حدد الموظف واكتب شنو تريد يفسّر")
		return
	}
	me := middleware.EmployeeIDFromContext(r)
	if err := h.Request("", b.EmployeeID, me, strings.TrimSpace(b.Question)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر تسجيل الطلب")
		return
	}
	if err := h.notif.Create(b.EmployeeID, "monitor_justify", "📝 المراقب يطلب تفسير: "+strings.TrimSpace(b.Question)+". اكتب عذرك — يطلعلك أول ما تفتح النظام."); err != nil {
		log.Printf("justify notify: %v", err)
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
