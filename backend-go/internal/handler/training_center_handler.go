package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// ═══ مركز التدريب — قرار (ع) 10-13 ═══
//
// «شي إعجازي بالتدريب»: ماتركس يحدد منو لازم يتدرّب (من شغله الحقيقي)،
// ملاحظات السرعة والعمل لكل كادر، مسارات تخصصية، خطة شهرية ويا تفتيش،
// متابعة الجدد، منهج لكل دور، وبنك أسئلة (يدوي + ماتركس يقترح والمدرب
// يعتمد) واختبار شهري يسويه الموظف من حسابه. بلا خصم نقاط — تقرير وتدريب.
//
// ⚠️ تحديد المرشحين **قواعد بالأرقام** بالقاعدة، مو ذكاء — ولا اسم موظف
// يطلع لذكاء خارجي. الذكاء بس يكتب أسئلة من عناوين المنهج.

type TrainingCenterHandler struct {
	db    *sqlx.DB
	Agent *service.MatrixAgent
}

func NewTrainingCenterHandler(db *sqlx.DB) *TrainingCenterHandler { return &TrainingCenterHandler{db: db} }

func month() string { return time.Now().In(baghdadLoc()).Format("2006-01") }

func baghdadLoc() *time.Location {
	if l, err := time.LoadLocation("Asia/Baghdad"); err == nil {
		return l
	}
	return time.FixedZone("AST", 3*3600)
}

// ── ١) المرشحين للتدريب + ملاحظات السرعة والعمل ──

type TrainingInsight struct {
	ID          string   `db:"id" json:"id"`
	Name        string   `db:"name" json:"name"`
	Role        string   `db:"role" json:"role"`
	DaysIn      int      `db:"daysIn" json:"daysIn"`
	Jobs        int      `db:"jobs" json:"jobs"`
	AvgMinutes  *float64 `db:"avgMinutes" json:"avgMinutes"`
	TeamMinutes *float64 `db:"teamMinutes" json:"teamMinutes"`
	Complaints  int      `db:"complaints" json:"complaints"`
	Stops       int      `db:"stops" json:"stops"`
	Partials    int      `db:"partials" json:"partials"`
	Deductions  int      `db:"deductions" json:"deductions"`
	Skills      int      `db:"skills" json:"skills"`
	LastExamPct *int     `db:"lastExamPct" json:"lastExamPct"`
	Score       int      `json:"score"`
	Reasons     []string `json:"reasons"`
	Strengths   []string `json:"strengths"`
}

// GET /api/training-center/insights — كل كادر بأرقامه + سبب تدريبه
func (h *TrainingCenterHandler) Insights(w http.ResponseWriter, r *http.Request) {
	rows := []TrainingInsight{}
	err := h.db.Select(&rows, `
		WITH jobs AS (
			SELECT ba."employeeId" AS id, b.id AS bid, b."serviceId",
			       EXTRACT(EPOCH FROM (b."completedAt" - COALESCE(b."startedAt", b."arrivedAt"))) / 60 AS mins
			FROM "BookingAssignment" ba JOIN "Booking" b ON b.id = ba."bookingId"
			WHERE b.status = 'COMPLETED' AND b."completedAt" > now() - interval '60 days'
			  AND COALESCE(b."startedAt", b."arrivedAt") IS NOT NULL AND b."completedAt" > COALESCE(b."startedAt", b."arrivedAt")
		), svc AS (SELECT "serviceId", AVG(mins) AS m FROM jobs WHERE mins BETWEEN 5 AND 1440 GROUP BY 1)
		SELECT e.id, e.name, e.role::text AS role,
		       GREATEST(0, (CURRENT_DATE - COALESCE(e."hireDate"::date, e."createdAt"::date)))::int AS "daysIn",
		       (SELECT COUNT(DISTINCT j.bid) FROM jobs j WHERE j.id = e.id)::int AS jobs,
		       (SELECT AVG(j.mins) FROM jobs j WHERE j.id = e.id AND j.mins BETWEEN 5 AND 1440)::float8 AS "avgMinutes",
		       (SELECT AVG(s.m) FROM jobs j JOIN svc s ON s."serviceId" = j."serviceId" WHERE j.id = e.id AND j.mins BETWEEN 5 AND 1440)::float8 AS "teamMinutes",
		       (SELECT COUNT(*) FROM "Complaint" c JOIN "BookingAssignment" ba ON ba."bookingId" = c."bookingId"
		         WHERE ba."employeeId" = e.id AND c."createdAt" > now() - interval '60 days')::int AS complaints,
		       (SELECT COUNT(*) FROM "Booking" b JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
		         WHERE ba."employeeId" = e.id AND b."workStoppedAt" > now() - interval '60 days')::int AS stops,
		       (SELECT COUNT(*) FROM "Booking" b JOIN "BookingAssignment" ba ON ba."bookingId" = b.id
		         WHERE ba."employeeId" = e.id AND b."lastPartialAt" > now() - interval '60 days')::int AS partials,
		       (SELECT COUNT(*) FROM "KpiEvaluation" k WHERE k."employeeId" = e.id AND NOT k.cancelled AND k.points < 0
		         AND k."createdAt" > now() - interval '60 days')::int AS deductions,
		       (SELECT COUNT(*) FROM "EmployeeSkill" es WHERE es."employeeId" = e.id AND es."canPerform")::int AS skills,
		       (SELECT (a.score * 100 / NULLIF(a.total, 0)) FROM "TrainingAttempt" a WHERE a."employeeId" = e.id AND a."finishedAt" IS NOT NULL
		         ORDER BY a."finishedAt" DESC LIMIT 1)::int AS "lastExamPct"
		FROM "Employee" e
		WHERE e.status = 'ACTIVE' AND e.role::text NOT IN ('ADMIN', 'OWNER')
		ORDER BY e.name`)
	if err != nil {
		log.Printf("training insights: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر حساب التدريب")
		return
	}
	field := map[string]bool{"TECHNICIAN": true, "TECHNICAL": true}
	for i := range rows {
		x := &rows[i]
		x.Reasons, x.Strengths = []string{}, []string{}
		add := func(pts int, why string) { x.Score += pts; x.Reasons = append(x.Reasons, why) }
		if x.Complaints > 0 {
			add(3*x.Complaints, fmt.Sprintf("%d شكوى على حجوزاته (٦٠ يوم)", x.Complaints))
		}
		if x.Stops > 0 {
			add(2*x.Stops, fmt.Sprintf("%d مرة توقف العمل", x.Stops))
		}
		if x.Partials > 1 {
			add(x.Partials, fmt.Sprintf("%d إنجاز جزئي (ما كمّل بالطلعة)", x.Partials))
		}
		if x.Deductions > 0 {
			add(2*x.Deductions, fmt.Sprintf("%d خصم كي بي اي", x.Deductions))
		}
		if x.AvgMinutes != nil && x.TeamMinutes != nil && *x.TeamMinutes > 0 && x.Jobs >= 3 {
			ratio := *x.AvgMinutes / *x.TeamMinutes
			if ratio >= 1.4 {
				add(3, fmt.Sprintf("أبطأ من زملائه بنفس الخدمات %d%% (معدله %d د مقابل %d د)", int((ratio-1)*100), int(*x.AvgMinutes), int(*x.TeamMinutes)))
			} else if ratio <= 0.8 {
				x.Strengths = append(x.Strengths, fmt.Sprintf("أسرع من زملائه %d%%", int((1-ratio)*100)))
			}
		}
		if field[x.Role] && x.Skills < 3 {
			add(2, fmt.Sprintf("مهاراته المؤكدة قليلة (%d)", x.Skills))
		}
		if x.LastExamPct != nil && *x.LastExamPct < 60 {
			add(3, fmt.Sprintf("رسب بآخر اختبار (%d%%)", *x.LastExamPct))
		} else if x.LastExamPct != nil && *x.LastExamPct >= 85 {
			x.Strengths = append(x.Strengths, fmt.Sprintf("ممتاز بآخر اختبار (%d%%)", *x.LastExamPct))
		}
		if x.DaysIn <= 90 {
			add(1, fmt.Sprintf("موظف جديد (يومه %d)", x.DaysIn))
		}
		if x.Jobs >= 10 && x.Complaints == 0 && x.Stops == 0 {
			x.Strengths = append(x.Strengths, fmt.Sprintf("%d حجز بلا شكوى ولا توقف", x.Jobs))
		}
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Score > rows[j].Score })
	WriteJSON(w, http.StatusOK, rows)
}

// ── ٢) المسارات التخصصية: كل خدمة بموادها ومهاراتها ومنو يتقنها ──

type trainingTrack struct {
	ID        string `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	Division  string `db:"division" json:"division"`
	Materials int    `db:"materials" json:"materials"`
	Skills    int    `db:"skills" json:"skills"`
	Masters   int    `db:"masters" json:"masters"`
	Assigned  int    `db:"assigned" json:"assigned"`
	Programs  int    `db:"programs" json:"programs"`
}

func (h *TrainingCenterHandler) Tracks(w http.ResponseWriter, r *http.Request) {
	rows := []trainingTrack{}
	if err := h.db.Select(&rows, `
		SELECT s.id, s.name, COALESCE(s.division::text, 'ENGINEERING') AS division,
		       (SELECT COUNT(*) FROM "TrainingMaterial" m WHERE m."serviceId" = s.id)::int AS materials,
		       (SELECT COUNT(*) FROM "Skill" k WHERE k."serviceId" = s.id)::int AS skills,
		       (SELECT COUNT(DISTINCT es."employeeId") FROM "EmployeeSkill" es JOIN "Skill" k ON k.id = es."skillId"
		         JOIN "Employee" e ON e.id = es."employeeId" AND e.status = 'ACTIVE'
		         WHERE k."serviceId" = s.id AND es."canPerform")::int AS masters,
		       (SELECT COUNT(*) FROM "EmployeeTrainingAssignment" a WHERE a."serviceId" = s.id)::int AS assigned,
		       (SELECT COUNT(DISTINCT ps."programId") FROM "TrainingProgramSkill" ps JOIN "Skill" k ON k.id = ps."skillId" WHERE k."serviceId" = s.id)::int AS programs
		FROM "Service" s ORDER BY s.name`); err != nil {
		log.Printf("training tracks: %v", err)
	}
	WriteJSON(w, http.StatusOK, rows)
}

// ── ٣) المنهج لكل دور ──

type curriculumItem struct {
	ID      string  `db:"id" json:"id"`
	Role    string  `db:"role" json:"role"`
	Title   string  `db:"title" json:"title"`
	Details *string `db:"details" json:"details"`
	Ord     int     `db:"ord" json:"ord"`
}

func (h *TrainingCenterHandler) Curriculum(w http.ResponseWriter, r *http.Request) {
	rows := []curriculumItem{}
	_ = h.db.Select(&rows, `SELECT id, role, title, details, ord FROM "TrainingCurriculum" ORDER BY role, ord, "createdAt"`)
	WriteJSON(w, http.StatusOK, rows)
}

func (h *TrainingCenterHandler) AddCurriculum(w http.ResponseWriter, r *http.Request) {
	var b curriculumItem
	if err := DecodeJSON(r, &b); err != nil || strings.TrimSpace(b.Title) == "" || b.Role == "" {
		WriteError(w, http.StatusBadRequest, "اكتب الدور والعنوان")
		return
	}
	_, err := h.db.Exec(`INSERT INTO "TrainingCurriculum" (id, role, title, details, ord) VALUES ($1,$2,$3,$4,
		(SELECT COALESCE(MAX(ord), 0) + 1 FROM "TrainingCurriculum" WHERE role = $2))`, uuid.NewString(), b.Role, strings.TrimSpace(b.Title), b.Details)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الحفظ")
		return
	}
	h.Curriculum(w, r)
}

func (h *TrainingCenterHandler) DeleteCurriculum(w http.ResponseWriter, r *http.Request) {
	_, _ = h.db.Exec(`DELETE FROM "TrainingCurriculum" WHERE id = $1`, r.PathValue("id"))
	h.Curriculum(w, r)
}

// ── ٤) بنك الأسئلة ──

type trainingQuestion struct {
	ID       string         `db:"id" json:"id"`
	Role     string         `db:"role" json:"role"`
	Text     string         `db:"text" json:"text"`
	Options  pq.StringArray `db:"options" json:"options"`
	Correct  int            `db:"correct" json:"correct"`
	Approved bool           `db:"approved" json:"approved"`
	Source   string         `db:"source" json:"source"`
}

func (h *TrainingCenterHandler) Questions(w http.ResponseWriter, r *http.Request) {
	rows := []trainingQuestion{}
	_ = h.db.Select(&rows, `SELECT id, role, text, options, correct, approved, source FROM "TrainingQuestion"
		WHERE ($1 = '' OR role = $1) ORDER BY approved, "createdAt" DESC`, r.URL.Query().Get("role"))
	WriteJSON(w, http.StatusOK, rows)
}

func (h *TrainingCenterHandler) AddQuestion(w http.ResponseWriter, r *http.Request) {
	var b trainingQuestion
	if err := DecodeJSON(r, &b); err != nil || strings.TrimSpace(b.Text) == "" || len(b.Options) < 2 || b.Correct < 0 || b.Correct >= len(b.Options) {
		WriteError(w, http.StatusBadRequest, "اكتب السؤال وخيارين على الأقل وحدد الصح")
		return
	}
	_, err := h.db.Exec(`INSERT INTO "TrainingQuestion" (id, role, text, options, correct, approved, source) VALUES ($1,$2,$3,$4,$5,true,'MANUAL')`,
		uuid.NewString(), b.Role, strings.TrimSpace(b.Text), pq.Array([]string(b.Options)), b.Correct)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *TrainingCenterHandler) ApproveQuestion(w http.ResponseWriter, r *http.Request) {
	_, _ = h.db.Exec(`UPDATE "TrainingQuestion" SET approved = true WHERE id = $1`, r.PathValue("id"))
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *TrainingCenterHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	_, _ = h.db.Exec(`DELETE FROM "TrainingQuestion" WHERE id = $1`, r.PathValue("id"))
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

var jsonArrayRe = regexp.MustCompile(`(?s)\[.*\]`)

// POST /api/training-center/questions/generate {role, roleLabel, count}
// ماتركس يكتب أسئلة من عناوين منهج الدور — بلا أسماء — وتبقى «تنتظر اعتماد».
func (h *TrainingCenterHandler) GenerateQuestions(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Role      string `json:"role"`
		RoleLabel string `json:"roleLabel"`
		Count     int    `json:"count"`
	}
	if err := DecodeJSON(r, &b); err != nil || b.Role == "" {
		WriteError(w, http.StatusBadRequest, "اختار الدور")
		return
	}
	if b.Count <= 0 || b.Count > 15 {
		b.Count = 8
	}
	if h.Agent == nil || !h.Agent.Enabled("ASK") {
		WriteError(w, http.StatusServiceUnavailable, "ذكاء ماتركس مطفي — اكتب الأسئلة بإيدك")
		return
	}
	topics := []string{}
	_ = h.db.Select(&topics, `SELECT title || COALESCE(' — ' || details, '') FROM "TrainingCurriculum" WHERE role = $1 ORDER BY ord`, b.Role)
	if len(topics) == 0 {
		WriteError(w, http.StatusBadRequest, "ضيف مفردات المنهج لهذا الدور أول — ماتركس يكتب الأسئلة منها")
		return
	}
	task := fmt.Sprintf(`اكتب %d أسئلة اختيار من متعدد باللهجة العراقية البسيطة لاختبار شهري لموظفين دورهم «%s» بشركة تركيب وصيانة (كاميرات، GPS، أقفال، شبكات، طاقة شمسية).
المفردات: %s
كل سؤال عملي من الشغل اليومي، ٤ خيارات، واحد صح. لا تستعمل أي أداة. رجّع JSON بس بهالشكل بلا أي نص ثاني:
[{"text":"...","options":["...","...","...","..."],"correct":0}]`, b.Count, b.RoleLabel, strings.Join(topics, " | "))
	run, err := h.Agent.Run("ASK", "انت مدرب تقني بشركة الأماني تكتب أسئلة اختبارات واضحة وقصيرة.", task)
	if err != nil {
		WriteError(w, http.StatusBadGateway, "ماتركس ما رد — جرّب بعد شوية")
		return
	}
	var qs []struct {
		Text    string   `json:"text"`
		Options []string `json:"options"`
		Correct int      `json:"correct"`
	}
	if err := json.Unmarshal([]byte(jsonArrayRe.FindString(run.Answer)), &qs); err != nil {
		WriteError(w, http.StatusBadGateway, "جواب ماتركس ما انقرا — جرّب مرة ثانية")
		return
	}
	n := 0
	for _, q := range qs {
		if strings.TrimSpace(q.Text) == "" || len(q.Options) < 2 || q.Correct < 0 || q.Correct >= len(q.Options) {
			continue
		}
		if _, err := h.db.Exec(`INSERT INTO "TrainingQuestion" (id, role, text, options, correct, approved, source) VALUES ($1,$2,$3,$4,$5,false,'AI')`,
			uuid.NewString(), b.Role, q.Text, pq.Array(q.Options), q.Correct); err == nil {
			n++
		}
	}
	WriteJSON(w, http.StatusOK, map[string]int{"added": n})
}

// ── ٥) الاختبار الشهري ──

type trainingExam struct {
	ID          string         `db:"id" json:"id"`
	Role        string         `db:"role" json:"role"`
	Month       string         `db:"month" json:"month"`
	Title       string         `db:"title" json:"title"`
	QuestionIDs pq.StringArray `db:"questionIds" json:"questionIds"`
	Minutes     int            `db:"minutes" json:"minutes"`
	PassPct     int            `db:"passPct" json:"passPct"`
}

// POST /api/training-center/exams {role, title, count, minutes} — من الأسئلة المعتمدة
func (h *TrainingCenterHandler) CreateExam(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Role    string `json:"role"`
		Title   string `json:"title"`
		Count   int    `json:"count"`
		Minutes int    `json:"minutes"`
	}
	if err := DecodeJSON(r, &b); err != nil || b.Role == "" {
		WriteError(w, http.StatusBadRequest, "اختار الدور")
		return
	}
	if b.Count <= 0 {
		b.Count = 10
	}
	if b.Minutes <= 0 {
		b.Minutes = 15
	}
	ids := []string{}
	_ = h.db.Select(&ids, `SELECT id FROM "TrainingQuestion" WHERE role = $1 AND approved ORDER BY random() LIMIT $2`, b.Role, b.Count)
	if len(ids) < 3 {
		WriteError(w, http.StatusBadRequest, "لازم ٣ أسئلة معتمدة على الأقل لهذا الدور")
		return
	}
	if strings.TrimSpace(b.Title) == "" {
		b.Title = "اختبار شهر " + month()
	}
	_, err := h.db.Exec(`INSERT INTO "TrainingExam" (id, role, month, title, "questionIds", minutes, "createdById") VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (role, month) DO UPDATE SET title = EXCLUDED.title, "questionIds" = EXCLUDED."questionIds", minutes = EXCLUDED.minutes`,
		uuid.NewString(), b.Role, month(), b.Title, pq.Array(ids), b.Minutes, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر إنشاء الاختبار")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]int{"questions": len(ids)})
}

type examReportRow struct {
	ExamID   string     `db:"examId" json:"examId"`
	Role     string     `db:"role" json:"role"`
	Month    string     `db:"month" json:"month"`
	Title    string     `db:"title" json:"title"`
	PassPct  int        `db:"passPct" json:"passPct"`
	EmpID    string     `db:"employeeId" json:"employeeId"`
	EmpName  string     `db:"employeeName" json:"employeeName"`
	Score    *int       `db:"score" json:"score"`
	Total    *int       `db:"total" json:"total"`
	Finished *time.Time `db:"finishedAt" json:"finishedAt"`
}

// GET /api/training-center/exams/report?month= — كل اختبار وكل موظف بدوره (سوّاه لو لا)
func (h *TrainingCenterHandler) ExamReport(w http.ResponseWriter, r *http.Request) {
	m := r.URL.Query().Get("month")
	if m == "" {
		m = month()
	}
	rows := []examReportRow{}
	if err := h.db.Select(&rows, `
		SELECT x.id AS "examId", x.role, x.month, x.title, x."passPct", e.id AS "employeeId", e.name AS "employeeName",
		       a.score, a.total, a."finishedAt"
		FROM "TrainingExam" x
		JOIN "Employee" e ON e.role::text = x.role AND e.status = 'ACTIVE'
		LEFT JOIN "TrainingAttempt" a ON a."examId" = x.id AND a."employeeId" = e.id
		WHERE x.month = $1 ORDER BY x.role, a.score DESC NULLS LAST, e.name`, m); err != nil {
		log.Printf("exam report: %v", err)
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/training-center/my-exam — اختبار شهري لدوري (بلا الأجوبة الصحيحة)
func (h *TrainingCenterHandler) MyExam(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	var x trainingExam
	err := h.db.Get(&x, `SELECT x.id, x.role, x.month, x.title, x."questionIds", x.minutes, x."passPct" FROM "TrainingExam" x
		JOIN "Employee" e ON e.role::text = x.role WHERE e.id = $1 AND x.month = $2`, me, month())
	if err != nil {
		WriteJSON(w, http.StatusOK, map[string]any{"exam": nil})
		return
	}
	var att struct {
		Score    *int       `db:"score"`
		Total    *int       `db:"total"`
		Finished *time.Time `db:"finishedAt"`
		Started  time.Time  `db:"startedAt"`
	}
	hasAtt := h.db.Get(&att, `SELECT score, total, "finishedAt", "startedAt" FROM "TrainingAttempt" WHERE "examId" = $1 AND "employeeId" = $2`, x.ID, me) == nil
	type q struct {
		ID      string         `db:"id" json:"id"`
		Text    string         `db:"text" json:"text"`
		Options pq.StringArray `db:"options" json:"options"`
	}
	qs := []q{}
	if !hasAtt || att.Finished == nil {
		_ = h.db.Select(&qs, `SELECT id, text, options FROM "TrainingQuestion" WHERE id = ANY($1) ORDER BY array_position($1, id)`, x.QuestionIDs)
	}
	out := map[string]any{"exam": map[string]any{"id": x.ID, "title": x.Title, "minutes": x.Minutes, "passPct": x.PassPct, "count": len(x.QuestionIDs)}, "questions": qs}
	if hasAtt {
		out["attempt"] = map[string]any{"score": att.Score, "total": att.Total, "finishedAt": att.Finished, "startedAt": att.Started}
	}
	WriteJSON(w, http.StatusOK, out)
}

// POST /api/training-center/my-exam/start — يبدي الوكت
func (h *TrainingCenterHandler) StartExam(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	_, _ = h.db.Exec(`INSERT INTO "TrainingAttempt" (id, "examId", "employeeId")
		SELECT $1, x.id, e.id FROM "TrainingExam" x JOIN "Employee" e ON e.role::text = x.role WHERE e.id = $2 AND x.month = $3
		ON CONFLICT DO NOTHING`, uuid.NewString(), me, month())
	h.MyExam(w, r)
}

// POST /api/training-center/my-exam/submit {answers: {questionId: optionIndex}}
func (h *TrainingCenterHandler) SubmitExam(w http.ResponseWriter, r *http.Request) {
	me := middleware.EmployeeIDFromContext(r)
	var b struct {
		Answers map[string]int `json:"answers"`
	}
	if err := DecodeJSON(r, &b); err != nil {
		WriteError(w, http.StatusBadRequest, "الأجوبة مو مفهومة")
		return
	}
	var x trainingExam
	if err := h.db.Get(&x, `SELECT x.id, x."questionIds", x.minutes FROM "TrainingExam" x JOIN "Employee" e ON e.role::text = x.role
		WHERE e.id = $1 AND x.month = $2`, me, month()); err != nil {
		WriteError(w, http.StatusNotFound, "ماكو اختبار")
		return
	}
	type key struct {
		ID      string `db:"id"`
		Correct int    `db:"correct"`
	}
	keys := []key{}
	_ = h.db.Select(&keys, `SELECT id, correct FROM "TrainingQuestion" WHERE id = ANY($1) ORDER BY array_position($1, id)`, x.QuestionIDs)
	score := 0
	ans := make([]int64, len(keys))
	for i, k := range keys {
		a, ok := b.Answers[k.ID]
		if !ok {
			a = -1
		}
		ans[i] = int64(a)
		if a == k.Correct {
			score++
		}
	}
	res, err := h.db.Exec(`UPDATE "TrainingAttempt" SET answers = $3, score = $4, total = $5, "finishedAt" = now()
		WHERE "examId" = $1 AND "employeeId" = $2 AND "finishedAt" IS NULL`, x.ID, me, pq.Array(ans), score, len(keys))
	if n, _ := res.RowsAffected(); err != nil || n == 0 {
		WriteError(w, http.StatusBadRequest, "الاختبار منسلّم قبل")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]int{"score": score, "total": len(keys)})
}

// ── ٦) الخطة الشهرية والتفتيش ──

type trainingPlan struct {
	ID               string         `db:"id" json:"id"`
	Month            string         `db:"month" json:"month"`
	Title            string         `db:"title" json:"title"`
	Role             *string        `db:"role" json:"role"`
	EmployeeIDs      pq.StringArray `db:"employeeIds" json:"employeeIds"`
	EmployeeNames    pq.StringArray `db:"employeeNames" json:"employeeNames"`
	PlannedDate      *time.Time     `db:"plannedDate" json:"plannedDate"`
	TrainerID        *string        `db:"trainerId" json:"trainerId"`
	TrainerName      *string        `db:"trainerName" json:"trainerName"`
	InspectorName    *string        `db:"inspectorName" json:"inspectorName"`
	InspectionResult *string        `db:"inspectionResult" json:"inspectionResult"`
	InspectionNotes  *string        `db:"inspectionNotes" json:"inspectionNotes"`
	InspectedAt      *time.Time     `db:"inspectedAt" json:"inspectedAt"`
}

func (h *TrainingCenterHandler) Plans(w http.ResponseWriter, r *http.Request) {
	m := r.URL.Query().Get("month")
	if m == "" {
		m = month()
	}
	rows := []trainingPlan{}
	if err := h.db.Select(&rows, `SELECT p.id, p.month, p.title, p.role, p."employeeIds",
		ARRAY(SELECT e.name FROM "Employee" e WHERE e.id = ANY(p."employeeIds") ORDER BY e.name) AS "employeeNames",
		p."plannedDate", p."trainerId", t.name AS "trainerName", i.name AS "inspectorName",
		p."inspectionResult", p."inspectionNotes", p."inspectedAt"
		FROM "TrainingPlan" p LEFT JOIN "Employee" t ON t.id = p."trainerId" LEFT JOIN "Employee" i ON i.id = p."inspectorId"
		WHERE p.month = $1 ORDER BY p."plannedDate" NULLS LAST, p."createdAt"`, m); err != nil {
		log.Printf("training plans: %v", err)
	}
	WriteJSON(w, http.StatusOK, rows)
}

func (h *TrainingCenterHandler) AddPlan(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Month       string   `json:"month"`
		Title       string   `json:"title"`
		Role        string   `json:"role"`
		EmployeeIDs []string `json:"employeeIds"`
		PlannedDate string   `json:"plannedDate"`
		TrainerID   string   `json:"trainerId"`
	}
	if err := DecodeJSON(r, &b); err != nil || strings.TrimSpace(b.Title) == "" {
		WriteError(w, http.StatusBadRequest, "اكتب عنوان التدريب")
		return
	}
	if b.Month == "" {
		b.Month = month()
	}
	if b.EmployeeIDs == nil {
		b.EmployeeIDs = []string{}
	}
	_, err := h.db.Exec(`INSERT INTO "TrainingPlan" (id, month, title, role, "employeeIds", "plannedDate", "trainerId")
		VALUES ($1,$2,$3,NULLIF($4,''),$5,NULLIF($6,'')::date,NULLIF($7,''))`,
		uuid.NewString(), b.Month, strings.TrimSpace(b.Title), b.Role, pq.Array(b.EmployeeIDs), b.PlannedDate, b.TrainerID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/training-center/plans/{id}/inspect {result: DONE|PARTIAL|NOT_DONE, notes}
func (h *TrainingCenterHandler) InspectPlan(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Result string `json:"result"`
		Notes  string `json:"notes"`
	}
	if err := DecodeJSON(r, &b); err != nil || (b.Result != "DONE" && b.Result != "PARTIAL" && b.Result != "NOT_DONE") {
		WriteError(w, http.StatusBadRequest, "اختار نتيجة التفتيش")
		return
	}
	_, _ = h.db.Exec(`UPDATE "TrainingPlan" SET "inspectionResult" = $2, "inspectionNotes" = $3, "inspectorId" = $4, "inspectedAt" = now() WHERE id = $1`,
		r.PathValue("id"), b.Result, strings.TrimSpace(b.Notes), middleware.EmployeeIDFromContext(r))
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *TrainingCenterHandler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	_, _ = h.db.Exec(`DELETE FROM "TrainingPlan" WHERE id = $1`, r.PathValue("id"))
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ── ٧) الموظفين الجدد: مشرف لكل واحد ──

type mentorRow struct {
	EmployeeID string  `db:"employeeId" json:"employeeId"`
	MentorID   *string `db:"mentorId" json:"mentorId"`
	MentorName *string `db:"mentorName" json:"mentorName"`
	Notes      *string `db:"notes" json:"notes"`
}

func (h *TrainingCenterHandler) Mentors(w http.ResponseWriter, r *http.Request) {
	rows := []mentorRow{}
	_ = h.db.Select(&rows, `SELECT m."employeeId", m."mentorId", e.name AS "mentorName", m.notes FROM "TrainingMentor" m
		LEFT JOIN "Employee" e ON e.id = m."mentorId"`)
	WriteJSON(w, http.StatusOK, rows)
}

func (h *TrainingCenterHandler) SetMentor(w http.ResponseWriter, r *http.Request) {
	var b struct {
		MentorID string `json:"mentorId"`
		Notes    string `json:"notes"`
	}
	if err := DecodeJSON(r, &b); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	_, err := h.db.Exec(`INSERT INTO "TrainingMentor" ("employeeId", "mentorId", notes) VALUES ($1, NULLIF($2,''), NULLIF($3,''))
		ON CONFLICT ("employeeId") DO UPDATE SET "mentorId" = EXCLUDED."mentorId", notes = EXCLUDED.notes, "updatedAt" = now()`,
		r.PathValue("id"), b.MentorID, strings.TrimSpace(b.Notes))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
