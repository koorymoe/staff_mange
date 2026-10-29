package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"staffmange-api/internal/model"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/service"
)

// ═══ 💬 دردشة ماتركس — المالك والمدير (قرار (ع) 10-10) ═══
// «نسولف ويا ويرد مثل أي ذكاء اصطناعي». ماتركس يتذكر المحادثة، ويفتح أدواته
// (قراءة بس) لما يحتاج أرقام الشركة.
type MatrixChatHandler struct {
	db    *sqlx.DB
	agent *service.MatrixAgent
	// chatModel موديل المحادثة المباشرة (أذكى من شغل الخلفية) — فارغ = موديل الوكيل
	chatModel string
	perms     func(employeeID string) []string
}

// SetPermissions مصدر صلاحيات الموظف — لفلترة أدوات ماتركس (دازاي 10-29).
func (h *MatrixChatHandler) SetPermissions(f func(employeeID string) []string) { h.perms = f }

// chatStaffDailyLimit حد رسائل الموظف (غير المدير/المالك) باليوم.
const chatStaffDailyLimit = 30

func isTopRole(r *http.Request) bool {
	role := middleware.RoleFromContext(r)
	return role == "OWNER" || role == "ADMIN"
}

// toolOpts فلتر الأدوات والأدوات الشخصية حسب منو يحاچي.
func (h *MatrixChatHandler) toolOpts(r *http.Request) (func(string) bool, []string) {
	me := middleware.EmployeeIDFromContext(r)
	var perms []string
	if h.perms != nil {
		perms = h.perms(me)
	}
	allow := service.MatrixToolFilter(middleware.RoleFromContext(r), perms)
	// (دازاي 10-29) قراءة بيانات مجاله — النسخة المحصورة تجي من extraTools
	if allow != nil && service.MatrixDataScope(perms) != nil {
		base := allow
		allow = func(n string) bool { return n == "query_data" || n == "describe_data" || base(n) }
	}
	return allow, perms
}

// SetChatModel موديل المحادثة المباشرة (AI_MODEL_CHAT).
func (h *MatrixChatHandler) SetChatModel(m string) { h.chatModel = m }

func NewMatrixChatHandler(db *sqlx.DB, agent *service.MatrixAgent) *MatrixChatHandler {
	return &MatrixChatHandler{db: db, agent: agent}
}

const chatDailyLimit = 200

const chatSystemPrompt = `أنت «ماتركس»، الذكاء الاصطناعي لشركة الأماني بالعراق (كاميرات، جي بي اس، طاقة شمسية، أنظمة أمنية، مشاريع). تسولف ويا المالك أو المدير.
- أنت مدير النظام: ادخل بالموضوع مباشرة بلا ترحيب ولا مقدمات — افحص واحكم وگول شنو المطلوب.
- الترتيب: ممنوع ** و# وعلامات ماركداون (الشاشة ما تعرضها). استعمل أسطر قصيرة تبدي بـ«- » وأرقام واضحة. وفلوس الموظفين من revenue_by_responsible بس: كل حجز إله مسؤول واحد، لا تكرر المبلغ على كل الكادر.
- جاوب على قد الرسالة: تحية = رد قصير بس، وسؤال محدد = جوابه بس، بلا إضافات ما انطلبت. التحليل والاقتراحات لما يسأل سؤال مفتوح.
- فكّر بالإنكليزي (التفكير يطلع للمالك)، والجواب بالعراقي.
- إذا لاحظت نقص حقيقي بأدواتك أو بالنظام (سؤال ما تگدر تجاوبه، شي يتكرر)، سجّله بـsuggest_improvement وكمّل.
- سولف طبيعي باللهجة العراقية مثل صاحب ذكي: ردود حيّة ومتنوعة، مو قالب. جاوب على أي سؤال، حتى العام.
- لما الكلام عن الشركة (فلوس، فواتير، موظفين، حجوزات، مشاريع، مخازن، دوام، تقييم) افتح أدواتك وشوف بعينك قبل ما تحكم، واربط بين المصادر.
- لا تخترع رقم ما شفته بأداة. إذا ما تگدر تتأكد گول بصراحة.
- الموظفين يجوك كرموز «موظف#n» — استعمل نفس الرمز بالضبط.
- لا تقترح غرامة أو خصم فلوس، ولا تخمّن أسباب شخصية. اقتراحاتك عملية.
- أدواتك قراءة بس: إذا طلبوا منك تغيّر شي، گلهم شنو يسوون هم ومن وين.
- إذا دزّولك صورة (سكرين من النظام): اقرا الي بيها، افتح system_guide حتى تفهم الشاشة، وإذا بيها كود حجز أو مشروع افتحه بـlookup_code، وبعدين اشرح «شنو هاي وليش» بالأرقام الي شفتها.`

type chatRow struct {
	ID        string    `db:"id" json:"id"`
	Title     string    `db:"title" json:"title"`
	UpdatedAt time.Time `db:"updatedAt" json:"updatedAt"`
}

type chatMsg struct {
	ID        string    `db:"id" json:"id"`
	Role      string    `db:"role" json:"role"`
	Text      string    `db:"text" json:"text"`
	Steps     *string   `db:"steps" json:"steps"`
	Image     *string   `db:"image" json:"image"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

func (h *MatrixChatHandler) own(r *http.Request, id string) bool {
	var ok bool
	_ = h.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "MatrixTalk" WHERE id = $1 AND "employeeId" = $2)`, id, middleware.EmployeeIDFromContext(r))
	return ok
}

// GET /api/matrix/chats
func (h *MatrixChatHandler) List(w http.ResponseWriter, r *http.Request) {
	rows := []chatRow{}
	if err := h.db.Select(&rows, `SELECT id, title, "updatedAt" FROM "MatrixTalk" WHERE "employeeId" = $1 ORDER BY "updatedAt" DESC LIMIT 50`,
		middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المحادثات")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"chats": rows, "enabled": h.agent.Enabled("CHAT")})
}

// POST /api/matrix/chats
func (h *MatrixChatHandler) Create(w http.ResponseWriter, r *http.Request) {
	var row chatRow
	if err := h.db.Get(&row, `INSERT INTO "MatrixTalk" (id, "employeeId") VALUES (gen_random_uuid()::text, $1) RETURNING id, title, "updatedAt"`,
		middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر إنشاء المحادثة")
		return
	}
	WriteJSON(w, http.StatusOK, row)
}

// GET /api/matrix/chats/{id}
func (h *MatrixChatHandler) Messages(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !h.own(r, id) {
		WriteError(w, http.StatusNotFound, "المحادثة غير موجودة")
		return
	}
	rows := []chatMsg{}
	_ = h.db.Select(&rows, `SELECT id, role, text, steps, image, "createdAt" FROM "MatrixTalkMessage" WHERE "chatId" = $1 ORDER BY "createdAt"`, id)
	WriteJSON(w, http.StatusOK, rows)
}

// DELETE /api/matrix/chats/{id}
func (h *MatrixChatHandler) Delete(w http.ResponseWriter, r *http.Request) {
	_, _ = h.db.Exec(`DELETE FROM "MatrixTalk" WHERE id = $1 AND "employeeId" = $2`, r.PathValue("id"), middleware.EmployeeIDFromContext(r))
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// POST /api/matrix/chats/{id}/messages {text}
// chatTurnIn رسالة جاهزة للوكيل بعد التحقق والذاكرة.
type chatTurnIn struct {
	id, text, rawImage, system, prompt string
	img                                *service.AgentImage
	turns                              []service.AgentTurn
}

// prepare يتحقق من الطلب ويجهّز التاريخ والذاكرة — يكتب الخطأ بنفسه ويرجّع nil.
func (h *MatrixChatHandler) prepare(w http.ResponseWriter, r *http.Request) *chatTurnIn {
	id, me := r.PathValue("id"), middleware.EmployeeIDFromContext(r)
	if !h.own(r, id) {
		WriteError(w, http.StatusNotFound, "المحادثة غير موجودة")
		return nil
	}
	var in struct {
		Text  string `json:"text"`
		Image string `json:"image"` // data:image/...;base64,...
	}
	if err := DecodeJSON(r, &in); err != nil || (strings.TrimSpace(in.Text) == "" && in.Image == "") {
		WriteError(w, http.StatusBadRequest, "اكتب رسالتك")
		return nil
	}
	var img *service.AgentImage
	if in.Image != "" {
		mt, data, ok := parseChatImage(in.Image)
		if !ok {
			WriteError(w, http.StatusBadRequest, "الصورة لازم PNG أو JPG أو WEBP وأصغر من 4 ميگا")
			return nil
		}
		img = &service.AgentImage{MediaType: mt, Data: data}
	}
	text := strings.TrimSpace(in.Text)
	if text == "" {
		text = "شنو هاي الشاشة؟ اشرحلي شنو الي بيها وليش."
	}
	if len([]rune(text)) > 2000 {
		WriteError(w, http.StatusBadRequest, "الرسالة طويلة — خليها أقصر")
		return nil
	}
	if !h.agent.Enabled("CHAT") {
		WriteError(w, http.StatusServiceUnavailable, "ماتركس يحتاج مفتاح هايكو (ANTHROPIC_API_KEY) حتى يسولف — أو مفتاح الدردشة مطفي")
		return nil
	}
	var today int
	_ = h.db.Get(&today, `SELECT count(*) FROM "MatrixTalkMessage" m JOIN "MatrixTalk" c ON c.id = m."chatId"
		WHERE c."employeeId" = $1 AND m.role = 'USER' AND m."createdAt" > now() - interval '24 hours'`, me)
	limit := chatDailyLimit
	if !isTopRole(r) {
		limit = chatStaffDailyLimit
	}
	if today >= limit {
		WriteError(w, http.StatusTooManyRequests, "وصلت حد الرسائل اليوم — باچر يرجع")
		return nil
	}
	// 🧠 الذاكرة (دازاي 10-23): ملخص المحادثة + آخر رسائل بس — الكلفة ثابتة مهما طالت
	var summary string
	_ = h.db.Get(&summary, `SELECT summary FROM "MatrixTalk" WHERE id = $1`, id)
	hist := []chatMsg{}
	_ = h.db.Select(&hist, `SELECT * FROM (SELECT id, role, text, steps, image, "createdAt" FROM "MatrixTalkMessage" WHERE "chatId" = $1
		ORDER BY "createdAt" DESC LIMIT $2) x ORDER BY "createdAt"`, id, service.MemoryKeep)
	turns := make([]service.AgentTurn, 0, len(hist))
	for _, m := range hist {
		t := m.Text
		if m.Image != nil {
			t = "[أرفق صورة] " + t
		}
		// رسالتين ورا بعض من نفس الطرف (مثل «تصحيح:») ينلزكن بوحدة
		if n := len(turns); n > 0 && turns[n-1].Role == m.Role {
			turns[n-1].Text += "\n" + t
			continue
		}
		turns = append(turns, service.AgentTurn{Role: m.Role, Text: t})
	}
	// التاريخ لازم يبدي برسالة مستخدم (النموذج يرفض غيرها)
	for len(turns) > 0 && turns[0].Role != "USER" {
		turns = turns[1:]
	}
	prompt := text
	if n := len(turns); n > 0 && turns[n-1].Role == "USER" {
		prompt = turns[n-1].Text + "\n" + text
		turns = turns[:n-1]
	}
	system := chatSystemPrompt + h.lessonsBlock()
	if !isTopRole(r) {
		var name, role string
		_ = h.db.QueryRow(`SELECT name, role::text FROM "Employee" WHERE id = $1`, me).Scan(&name, &role)
		if l, ok := staffRoleLabels[role]; ok {
			role = l
		}
		// (دازاي 10-29) الاسم ينخفى برمز (موظف#n) قبل ما يوصل النموذج، والخادم يرجّعه
		// بالجواب — فلازم يعرف إن الرمز هو اسمه ويستعمله، مو يگول «ما أعرف اسمك».
		// (دازاي 10-29) «لازم الذكاء يعرف المقامات والحدود مال كل موظف»
		_, perms := h.toolOpts(r)
		labels := map[string]string{}
		for _, p := range model.DefaultPermissions {
			labels[p.Name] = p.Label
		}
		var can []string
		for _, p := range perms {
			if l := labels[p]; l != "" && p != "agent_matrix" {
				can = append(can, l)
			}
		}
		scope := "ماعنده صلاحيات إضافية — شغله حسب دوره بس."
		if len(can) > 0 {
			scope = "صلاحياته بالنظام: " + strings.Join(can, "، ") + "."
		}
		if ds := service.MatrixDataScope(perms); ds != nil {
			scope += " وعنده قراءة لجداول مجاله بـdescribe_data ثم query_data: " + strings.Join(ds, "، ") + " — دوّر بيها قبل ما تگول ماكو بيانات."
		}
		system += "\n\n【حدوده ومقامه】 " + scope + " هو موظف تحت إدارة المدير والمالك: ما يوجّه مهام أو أوامر للمدير أو المالك ولا لموظفين برّا صلاحياته، وما يطّلع على شي برّا صلاحياته. إذا طلب شي برّا حدوده (مثل «وجّه مهمة للمدير»، أو بيانات قسم ثاني)، گوله بوضوح واحترام إنه **مو ضمن صلاحياتك/دورك** — لا تگول «ما عندي أداة» ولا تعامله كنقص بالنظام ولا تسجّله بـsuggest_improvement — واقترح الطريق الصحيح (يراجع مديره، أو يطلب من المسؤول). suggest_improvement بس لنقص حقيقي يفيد شغله ضمن صلاحياته."
		system += "\n\n【تحاچي موظف مو المدير】 هو " + name + "، دوره " + role + ". اسمه يوصلك كرمز (موظف#n) — استعمل الرمز نفسه كأنه اسمه وناديه بيه، والنظام يبدّله باسمه الحقيقي قبل ما يوصله. لا تگول أبداً «ما أعرف اسمك» ولا تذكر إن الأسماء رموز. أدواتك محدودة بصلاحياته هو بس — إذا سأل عن شي ما عندك أداته (فلوس، تقييم غيره، رواتب، بيانات موظفين ثانيين) گول بأدب «هذا مو ضمن صلاحياتك» ولا تخمّن. أرقامه هو من my_stats. لا تكشف تقييم أو حال أو فلوس موظف ثاني إلا إذا الأداة نفسها انفتحتلك."
	}
	if strings.TrimSpace(summary) != "" {
		system += "\n\nذاكرة هالمحادثة (ملخص الي صار قبل آخر الرسائل):\n" + summary
	}
	return &chatTurnIn{id: id, text: text, rawImage: in.Image, system: system, prompt: prompt, img: img, turns: turns}
}

func (h *MatrixChatHandler) Send(w http.ResponseWriter, r *http.Request) {
	// الوكيل يسمح لنفسه 60 ثانية والخادم يقطع بـ30 — نمد مهلة هالرد بس
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(90 * time.Second))
	in := h.prepare(w, r)
	if in == nil {
		return
	}
	allow, perms := h.toolOpts(r)
	run, err := h.agent.ChatStream("CHAT", in.system, in.turns, in.prompt, in.img, service.ChatOpts{Model: h.chatModel,
		AllowTool: allow, ExtraTools: h.agent.StaffTools(h.db, middleware.EmployeeIDFromContext(r), isTopRole(r), perms)})
	if err != nil || strings.TrimSpace(run.Answer) == "" {
		WriteError(w, http.StatusBadGateway, chatFailMsg(err))
		return
	}
	reply, err := h.save(in, run)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, reply)
}

// POST /api/matrix-chats/{id}/stream — نفس Send بس بث حي (SSE): تفكير بالإنكليزي،
// والأدوات الي يفتش بيها، والجواب كلمة كلمة، وبالآخر done ويا الرسالة المحفوظة.
func (h *MatrixChatHandler) Stream(w http.ResponseWriter, r *http.Request) {
	rc := http.NewResponseController(w)
	_ = rc.SetWriteDeadline(time.Now().Add(150 * time.Second))
	in := h.prepare(w, r)
	if in == nil {
		return
	}
	allow, perms := h.toolOpts(r)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	var mu sync.Mutex
	send := func(v any) {
		b, _ := json.Marshal(v)
		mu.Lock()
		defer mu.Unlock()
		_, _ = w.Write([]byte("data: "))
		_, _ = w.Write(b)
		_, _ = w.Write([]byte("\n\n"))
		_ = rc.Flush()
	}
	run, err := h.agent.ChatStream("CHAT", in.system, in.turns, in.prompt, in.img, service.ChatOpts{
		Ctx: r.Context(), Model: h.chatModel, Think: true, Emit: func(e service.AgentEvent) { send(e) },
		AllowTool: allow, ExtraTools: h.agent.StaffTools(h.db, middleware.EmployeeIDFromContext(r), isTopRole(r), perms),
	})
	if r.Context().Err() != nil {
		return // المستخدم وكّف — ما نحفظ نص ناقص
	}
	if err != nil || strings.TrimSpace(run.Answer) == "" {
		send(map[string]string{"type": "error", "text": chatFailMsg(err)})
		return
	}
	reply, err := h.save(in, run)
	if err != nil {
		send(map[string]string{"type": "error", "text": "تعذر الحفظ"})
		return
	}
	send(map[string]any{"type": "done", "message": reply})
}

func chatFailMsg(err error) string {
	if errors.Is(err, service.ErrAgentOff) {
		return "ماتركس مطفي هسه"
	}
	if errors.Is(err, service.ErrAIBudget) {
		return "ماتركس وصل سقف الصرف اليومي — يرجع باچر أو زيد السقف"
	}
	return "ماتركس ما گدر يرد هسه — جرّب بعد شوية"
}

// save يحفظ رسالة المستخدم والرد ويحدّث العنوان والذاكرة.
func (h *MatrixChatHandler) save(in *chatTurnIn, run *service.AgentRun) (chatMsg, error) {
	id, text := in.id, in.text
	steps := strings.Join(run.Steps, ",")
	var reply chatMsg
	tx, err := h.db.Beginx()
	if err != nil {
		return reply, err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT INTO "MatrixTalkMessage" (id, "chatId", role, text, image) VALUES (gen_random_uuid()::text, $1, 'USER', $2, NULLIF($3, ''))`, id, text, in.rawImage); err != nil {
		return reply, err
	}
	err = tx.Get(&reply, `INSERT INTO "MatrixTalkMessage" (id, "chatId", role, text, steps, "createdAt")
		VALUES (gen_random_uuid()::text, $1, 'ASSISTANT', $2, NULLIF($3, ''), now() + interval '1 millisecond')
		RETURNING id, role, text, steps, "createdAt"`, id, run.Answer, steps)
	if err != nil {
		return reply, err
	}
	title := text
	if r := []rune(title); len(r) > 40 {
		title = string(r[:40]) + "…"
	}
	_, _ = tx.Exec(`UPDATE "MatrixTalk" SET "updatedAt" = now(), title = CASE WHEN title = 'محادثة جديدة' THEN $2 ELSE title END WHERE id = $1`, id, title)
	if err := tx.Commit(); err != nil {
		return reply, err
	}
	go h.updateMemory(id)
	return reply, nil
}

// parseChatImage يفك data-URL ويتأكد من النوع والحجم.
func parseChatImage(u string) (string, string, bool) {
	head, data, ok := strings.Cut(u, ",")
	if !ok || !strings.HasPrefix(head, "data:") || !strings.HasSuffix(head, ";base64") {
		return "", "", false
	}
	mt := strings.TrimSuffix(strings.TrimPrefix(head, "data:"), ";base64")
	switch mt {
	case "image/png", "image/jpeg", "image/webp":
	default:
		return "", "", false
	}
	if len(data) == 0 || len(data) > 4*1024*1024*4/3 {
		return "", "", false
	}
	return mt, data, true
}

// ── ماتركس يتعلّم: تصحيح المالك يصير درس دائم (قرار (ع) 10-11) ──

func (h *MatrixChatHandler) lessonsBlock() string {
	rows := []string{}
	_ = h.db.Select(&rows, `SELECT text FROM "MatrixLesson" WHERE active ORDER BY "createdAt" DESC LIMIT 50`)
	if len(rows) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n\nدروس من المالك (صحّحك بيها قبل) — التزم بيها وقدّمها على أي استنتاج ثاني:\n")
	for _, t := range rows {
		b.WriteString("- " + t + "\n")
	}
	return b.String()
}

type lessonRow struct {
	ID        string    `db:"id" json:"id"`
	Text      string    `db:"text" json:"text"`
	Context   *string   `db:"context" json:"context"`
	By        *string   `db:"by" json:"by"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

// POST /api/matrix/chats/{id}/messages/{msgId}/correct {correction}
func (h *MatrixChatHandler) Correct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !h.own(r, id) {
		WriteError(w, http.StatusNotFound, "المحادثة غير موجودة")
		return
	}
	var in struct {
		Correction string `json:"correction"`
	}
	if err := DecodeJSON(r, &in); err != nil || len([]rune(strings.TrimSpace(in.Correction))) < 5 {
		WriteError(w, http.StatusBadRequest, "اكتب شنو الصح")
		return
	}
	corr := strings.TrimSpace(in.Correction)
	if r := []rune(corr); len(r) > 600 {
		corr = string(r[:600])
	}
	// السياق: سؤال المستخدم الي قبل الرد الغلط
	var ctx string
	_ = h.db.Get(&ctx, `SELECT u.text FROM "MatrixTalkMessage" a JOIN "MatrixTalkMessage" u ON u."chatId" = a."chatId"
		AND u.role = 'USER' AND u."createdAt" < a."createdAt"
		WHERE a.id = $1 AND a."chatId" = $2 AND a.role = 'ASSISTANT' ORDER BY u."createdAt" DESC LIMIT 1`, r.PathValue("msgId"), id)
	if ctx == "" {
		WriteError(w, http.StatusNotFound, "الرد غير موجود")
		return
	}
	if r := []rune(ctx); len(r) > 200 {
		ctx = string(r[:200]) + "…"
	}
	if _, err := h.db.Exec(`INSERT INTO "MatrixLesson" (id, text, context, "createdById") VALUES (gen_random_uuid()::text, $1, $2, $3)`,
		corr, ctx, middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حفظ الدرس")
		return
	}
	_, _ = h.db.Exec(`INSERT INTO "MatrixTalkMessage" (id, "chatId", role, text) VALUES (gen_random_uuid()::text, $1, 'USER', $2)`,
		id, "تصحيح: "+corr)
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/matrix/lessons
func (h *MatrixChatHandler) Lessons(w http.ResponseWriter, r *http.Request) {
	rows := []lessonRow{}
	_ = h.db.Select(&rows, `SELECT l.id, l.text, l.context, e.name AS by, l."createdAt" FROM "MatrixLesson" l
		LEFT JOIN "Employee" e ON e.id = l."createdById" WHERE l.active ORDER BY l."createdAt" DESC`)
	WriteJSON(w, http.StatusOK, rows)
}

// DELETE /api/matrix/lessons/{id} — يطفّيه، ما يحذفه
func (h *MatrixChatHandler) DisableLesson(w http.ResponseWriter, r *http.Request) {
	_, _ = h.db.Exec(`UPDATE "MatrixLesson" SET active = false WHERE id = $1`, r.PathValue("id"))
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// updateMemory يدمج الرسائل القديمة بملخص محادثة ماتركس إذا تجمّعن (بالخلفية).
func (h *MatrixChatHandler) updateMemory(chatID string) {
	var c struct {
		Summary string     `db:"summary"`
		UpTo    *time.Time `db:"summarizedUpTo"`
	}
	if err := h.db.Get(&c, `SELECT summary, "summarizedUpTo" FROM "MatrixTalk" WHERE id = $1`, chatID); err != nil {
		return
	}
	msgs := []chatMsg{}
	if err := h.db.Select(&msgs, `SELECT id, role, text, steps, image, "createdAt" FROM "MatrixTalkMessage"
		WHERE "chatId" = $1 AND ($2::timestamptz IS NULL OR "createdAt" > $2) ORDER BY "createdAt"`, chatID, c.UpTo); err != nil {
		return
	}
	if len(msgs)-service.MemoryKeep < service.MemoryThreshold {
		return
	}
	msgs = msgs[:len(msgs)-service.MemoryKeep]
	turns := make([]service.AgentTurn, 0, len(msgs))
	for _, m := range msgs {
		turns = append(turns, service.AgentTurn{Role: m.Role, Text: m.Text})
	}
	sum, err := h.agent.Remember("CHAT", c.Summary, turns)
	if err != nil {
		log.Printf("[matrix-memory] %v", err)
		return
	}
	_, _ = h.db.Exec(`UPDATE "MatrixTalk" SET summary = $2, "summarizedUpTo" = $3 WHERE id = $1`, chatID, sum, msgs[len(msgs)-1].CreatedAt)
}

// staffRoleLabels أسماء الأدوار بالعربي (نفس frontend/src/roleLabels.ts) — للبرومبت.
var staffRoleLabels = map[string]string{
	"SALES": "موظف مبيعات", "HR_COORDINATOR": "إداري الكوادر", "TECHNICIAN": "فني", "TECHNICAL": "تقني",
	"PROJECT_MANAGER": "مدير مشاريع", "MONITOR": "مدقق / مراقب", "FINANCE": "محاسب", "GPS_ADMIN": "مسؤول GPS",
	"QUALITY_ENGINEER": "مهندس جودة", "ENGINEER": "مهندس", "PROCUREMENT_ADMIN": "إداري الكميات", "DESIGNER": "مصمم",
	"SERVICE_MANAGER": "مسؤول خدمة", "IT_SUPPORT": "الدعم التقني", "MEDIA": "الإعلام", "CAR_WASHER": "عامل الغسل",
	"SERVICE_DEVELOPER": "مطوّر خدمات", "PUBLIC_RELATIONS": "العلاقات العامة",
}
