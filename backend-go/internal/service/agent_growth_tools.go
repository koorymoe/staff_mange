package service

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/errlog"
	"staffmange-api/internal/model"
	"staffmange-api/internal/monitor"
)

// ═══ 🌱 الوكلاء يطورون نفسهم (دازاي 11-06) ═══
// ١) قياس كل نداء أداة (عدد، فشل، وقت) بلا كلفة ذكاء.
// ٢) report_gap: سؤال ما گدر يجاوبه — إذا نفس النوع تكرر ٣ مرات خلال ١٤ يوم ينخلق اقتراح برمجة لحاله.
// ٣) my_performance: الوكيل يشوف أرقامه.
// ٤) propose_lesson: الوكيل يقترح درس على نفسه، يدخل البرومبت بس بعد موافقة المالك.

const (
	gapRepeatThreshold = 3
	gapWindowDays      = 14
)

// RecordToolStat يسجّل نداء أداة. الخطأ ينظّف (بلا أسماء وأرقام) وينقص لـ٢٠٠ حرف.
func RecordToolStat(db *sqlx.DB, agent, tool string, d time.Duration, err error) {
	if db == nil || agent == "" {
		return
	}
	var fail int
	var last *string
	if err != nil {
		fail = 1
		s := monitor.Scrub(err.Error())
		if r := []rune(s); len(r) > 200 {
			s = string(r[:200])
		}
		last = &s
	}
	day := time.Now().In(debriefLoc).Format("2006-01-02")
	errlog.Exec(db.Exec(`INSERT INTO "AgentToolStat" (day, agent, tool, calls, failures, "totalMs", "lastError")
		VALUES ($1::date, $2, $3, 1, $4, $5, $6)
		ON CONFLICT (day, agent, tool) DO UPDATE SET calls = "AgentToolStat".calls + 1,
			failures = "AgentToolStat".failures + EXCLUDED.failures,
			"totalMs" = "AgentToolStat"."totalMs" + EXCLUDED."totalMs",
			"lastError" = COALESCE(EXCLUDED."lastError", "AgentToolStat"."lastError")`,
		day, agent, tool, fail, d.Milliseconds(), last))
}

type toolStatRow struct {
	Tool      string  `db:"tool" json:"tool"`
	Calls     int     `db:"calls" json:"calls"`
	Failures  int     `db:"failures" json:"failures"`
	AvgMs     int     `db:"avgMs" json:"avgMs"`
	LastError *string `db:"lastError" json:"lastError"`
}

// AgentToolStats أرقام أدوات الوكيل لآخر days يوم، الأكثر فشلاً فوق.
func AgentToolStats(db *sqlx.DB, agent string, days int) ([]toolStatRow, error) {
	rows := []toolStatRow{}
	err := db.Select(&rows, `SELECT tool, sum(calls)::int AS calls, sum(failures)::int AS failures,
			(sum("totalMs") / greatest(sum(calls), 1))::int AS "avgMs", max("lastError") AS "lastError"
		FROM "AgentToolStat" WHERE agent = $1 AND day > (now() AT TIME ZONE 'Asia/Baghdad')::date - $2::int
		GROUP BY tool ORDER BY sum(failures) DESC, sum(calls) DESC LIMIT 30`, agent, days)
	return rows, err
}

type gapRow struct {
	Kind    string `db:"kind" json:"kind"`
	Count   int    `db:"n" json:"count"`
	Missing string `db:"missing" json:"missing"`
}

// AgentGaps الفجوات المتكررة لآخر days يوم.
func AgentGaps(db *sqlx.DB, agent string, days int) ([]gapRow, error) {
	rows := []gapRow{}
	err := db.Select(&rows, `SELECT kind, count(*)::int AS n, (array_agg(missing ORDER BY "createdAt" DESC))[1] AS missing
		FROM "AgentGap" WHERE agent = $1 AND "createdAt" > now() - make_interval(days => $2)
		GROUP BY kind ORDER BY n DESC LIMIT 20`, agent, days)
	return rows, err
}

// recordGap يسجّل فجوة، ويصنع اقتراح برمجة إذا النوع تكرر.
func recordGap(db *sqlx.DB, agent, kind, missing string, notifyOwner func(string)) (string, error) {
	kind = strings.TrimSpace(kind)
	if r := []rune(kind); len(r) > 80 {
		kind = string(r[:80])
	}
	if r := []rune(missing); len(r) > 400 {
		missing = string(r[:400])
	}
	if kind == "" {
		return "", fmt.Errorf("نوع السؤال فارغ")
	}
	if _, err := db.Exec(`INSERT INTO "AgentGap" (agent, kind, missing) VALUES ($1, $2, $3)`, agent, kind, missing); err != nil {
		return "", err
	}
	var n int
	errlog.Err(db.Get(&n, `SELECT count(*) FROM "AgentGap" WHERE agent = $1 AND kind = $2 AND "createdAt" > now() - make_interval(days => $3)`,
		agent, kind, gapWindowDays))
	if n < gapRepeatThreshold {
		return "انسجلت الفجوة. كمّل جوابك بأحسن شي تگدره.", nil
	}
	title := "سؤال يتكرر وما عندي أداة: " + kind
	if r := []rune(title); len(r) > 120 {
		title = string(r[:120])
	}
	var dup, today int
	errlog.Err(db.Get(&dup, `SELECT count(*) FROM "AiSuggestion" WHERE agent = $1 AND lower(title) = lower($2) AND "createdAt" > now() - interval '30 days'`, agent, title))
	errlog.Err(db.Get(&today, `SELECT count(*) FROM "AiSuggestion" WHERE "createdAt" > now() - interval '24 hours'`))
	if dup > 0 || today >= aiSuggestDailyMax {
		return "انسجلت الفجوة (وهي مطلوبة قبل كاقتراح).", nil
	}
	details := fmt.Sprintf("انسأل هالنوع %d مرات خلال %d يوم وما گدرت أجاوب. الناقص (آخر مرة): %s", n, gapWindowDays, missing)
	if _, err := db.Exec(`INSERT INTO "AiSuggestion" (agent, title, details, "needsCode") VALUES ($1, $2, $3, true)`, agent, title, details); err != nil {
		return "", err
	}
	if notifyOwner != nil {
		notifyOwner("🧠 " + model.AgentName(agent) + " يطلب أداة: «" + kind + "» انسألت " + fmt.Sprint(n) + " مرات وما گدر يجاوب — بشاشة «اقتراحات الذكاء».")
	}
	return "انسجلت الفجوة، ولأنها تكررت صارت طلب تحديث للمالك.", nil
}

// RegisterGrowthTools أدوات التطوير الذاتي: report_gap وmy_performance وpropose_lesson.
func (a *MatrixAgent) RegisterGrowthTools(db *sqlx.DB, agent string, notifyOwner func(string)) {
	a.name = agent
	a.AddTool(agentTool{
		Name: "report_gap",
		Desc: "سجّل سؤال ما گدرت تجاوبه لأن ماكو أداة أو بيانات (مو لأن السؤال برّا صلاحيات السائل). kind = نوع السؤال بكلمات قليلة ثابتة حتى يتجمّع (مثل «رصيد مخزن مادة»، «رواتب الشهر»)، missing = شنو الناقص بالضبط. إذا نفس النوع تكرر ٣ مرات يصير طلب تحديث للمالك لحاله. استعملها وكمّل جوابك.",
		Props: map[string]any{
			"kind":    map[string]any{"type": "string", "description": "نوع السؤال بكلمات قليلة ثابتة"},
			"missing": map[string]any{"type": "string", "description": "الأداة أو البيانات الناقصة"},
		},
		Req: []string{"kind", "missing"},
		Run: func(in map[string]any, names *nameMask) (string, error) {
			kind, missing := strArg(in, "kind"), strArg(in, "missing")
			if names != nil && !a.plainNames {
				kind, missing = names.Hide(kind), names.Hide(missing) // ما تنحفظ أسماء
			}
			return recordGap(db, agent, kind, missing, notifyOwner)
		},
	})
	a.AddTool(agentTool{
		Name: "my_performance",
		Desc: "أرقامك آخر ٧ أيام: أدواتك الأكثر فشلاً وأبطأها، والأسئلة الي ما گدرت تجاوبها، وتقييم المستخدمين لأجوبتك (👍/👎). استعملها من تريد تعرف شنو لازم تحسّن أو تطلب.",
		Run: func(map[string]any, *nameMask) (string, error) {
			return performanceText(db, agent, 7), nil
		},
	})
	a.AddTool(agentTool{
		Name: "propose_lesson",
		Desc: "اقترح درس على نفسك من غلط لاحظته (مثل «لا تجاوب عن الرواتب بلا أداة»، «التاريخ دائماً بتوقيت بغداد»). الدرس ما يدخل تعليماتك إلا بعد ما المالك يوافق عليه. جملة وحدة واضحة وعامة، بلا أسماء.",
		Props: map[string]any{
			"text": map[string]any{"type": "string", "description": "الدرس بجملة وحدة"},
			"why":  map[string]any{"type": "string", "description": "شنو الي صار وخلّاك تتعلمه"},
		},
		Req: []string{"text"},
		Run: func(in map[string]any, names *nameMask) (string, error) {
			text, why := strArg(in, "text"), strArg(in, "why")
			if names != nil && !a.plainNames {
				text, why = names.Hide(text), names.Hide(why)
			}
			return proposeLesson(db, agent, text, why, notifyOwner)
		},
	})
}

func proposeLesson(db *sqlx.DB, agent, text, why string, notifyOwner func(string)) (string, error) {
	if r := []rune(text); len(r) < 8 || len(r) > 300 {
		return "", fmt.Errorf("الدرس لازم جملة واضحة (٨–٣٠٠ حرف)")
	}
	var pending, dup int
	errlog.Err(db.Get(&pending, `SELECT count(*) FROM "MatrixLesson" WHERE agent = $1 AND status = 'PROPOSED'`, agent))
	if pending >= 10 {
		return "عندك ١٠ دروس تنتظر موافقة المالك — انتظر لحد ما يراجعها.", nil
	}
	errlog.Err(db.Get(&dup, `SELECT count(*) FROM "MatrixLesson" WHERE agent = $1 AND lower(text) = lower($2)`, agent, text))
	if dup > 0 {
		return "هذا الدرس موجود.", nil
	}
	if _, err := db.Exec(`INSERT INTO "MatrixLesson" (id, text, context, agent, status, active) VALUES (gen_random_uuid()::text, $1, $2, $3, 'PROPOSED', false)`,
		text, why, agent); err != nil {
		return "", err
	}
	if notifyOwner != nil {
		notifyOwner("📘 " + model.AgentName(agent) + " يقترح درس على نفسه: «" + text + "» — وافق أو ارفض بشاشة «اقتراحات الذكاء».")
	}
	return "انسجل الدرس المقترح، ويدخل تعليماتك بعد موافقة المالك.", nil
}

// performanceText ملخص أرقام الوكيل (للأداة وللمراجعة الذاتية).
func performanceText(db *sqlx.DB, agent string, days int) string {
	var b strings.Builder
	stats, _ := AgentToolStats(db, agent, days)
	fmt.Fprintf(&b, "أدواتك آخر %d أيام:\n", days)
	if len(stats) == 0 {
		b.WriteString("- ماكو نداءات مسجّلة.\n")
	}
	for i, s := range stats {
		if i >= 10 {
			break
		}
		fmt.Fprintf(&b, "- %s: %d نداء، %d فشل، معدل %dms", s.Tool, s.Calls, s.Failures, s.AvgMs)
		if s.Failures > 0 && s.LastError != nil {
			fmt.Fprintf(&b, " (آخر خطأ: %s)", *s.LastError)
		}
		b.WriteString("\n")
	}
	gaps, _ := AgentGaps(db, agent, days)
	if len(gaps) > 0 {
		b.WriteString("أسئلة ما گدرت تجاوبها:\n")
		for _, g := range gaps {
			fmt.Fprintf(&b, "- %s ×%d: %s\n", g.Kind, g.Count, g.Missing)
		}
	}
	var good, bad int
	errlog.Err(db.Get(&good, `SELECT count(*) FROM "AgentFeedback" WHERE agent = $1 AND good AND "createdAt" > now() - make_interval(days => $2)`, agent, days))
	errlog.Err(db.Get(&bad, `SELECT count(*) FROM "AgentFeedback" WHERE agent = $1 AND NOT good AND "createdAt" > now() - make_interval(days => $2)`, agent, days))
	fmt.Fprintf(&b, "تقييم المستخدمين: 👍 %d · 👎 %d\n", good, bad)
	var notes []string
	errlog.Err(db.Select(&notes, `SELECT note FROM "AgentFeedback" WHERE agent = $1 AND NOT good AND note <> '' AND "createdAt" > now() - make_interval(days => $2) ORDER BY "createdAt" DESC LIMIT 5`, agent, days))
	for _, n := range notes {
		fmt.Fprintf(&b, "- 👎: %s\n", n)
	}
	return b.String()
}

// LessonsBlock دروس الوكيل المعتمدة لبرومبته (آخر ٥٠). الدروس المقترحة ما تدخل قبل الموافقة.
func LessonsBlock(db *sqlx.DB, agent string) string {
	if db == nil {
		return ""
	}
	rows := []string{}
	errlog.Err(db.Select(&rows, `SELECT text FROM "MatrixLesson" WHERE active AND status = 'ACTIVE' AND agent = $1 ORDER BY "createdAt" DESC LIMIT 50`, agent))
	if len(rows) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n\nدروس معتمدة من المالك — التزم بيها وقدّمها على أي استنتاج ثاني:\n")
	for _, t := range rows {
		b.WriteString("- " + t + "\n")
	}
	return b.String()
}

// ═══ 🪞 المراجعة الذاتية الأسبوعية (دازاي 11-06) ═══
// كل جمعة بعد ١١:٤٥ الليل بغداد: الوكيل يقرا أرقام أسبوعه (أدوات فاشلة، فجوات، 👎، نتائج تأكد)
// ويكتب تقرير قصير ينحفظ بشاشة الاقتراحات، ويگدر يقترح دروس أو تحديثات بأدواته.
// تشتغل بس إذا أكو بيانات، وتحت مفتاح «التعلّم».

func selfReviewDue(now time.Time, lastWeek string) bool {
	now = now.In(debriefLoc)
	if now.Weekday() != time.Friday || now.Hour()*60+now.Minute() < 23*60+45 {
		return false
	}
	return now.Format("2006-01-02") != lastWeek
}

type SelfReviewer struct {
	db       *sqlx.DB
	agent    *MatrixAgent
	name     string
	mu       sync.Mutex
	lastWeek string
}

func NewSelfReviewer(db *sqlx.DB, agent *MatrixAgent, name string) *SelfReviewer {
	return &SelfReviewer{db: db, agent: agent, name: name}
}

func (s *SelfReviewer) RunIfDue() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if !selfReviewDue(now, s.lastWeek) || !s.agent.Enabled("LEARNING") {
		return nil
	}
	s.lastWeek = now.In(debriefLoc).Format("2006-01-02")
	var activity int
	errlog.Err(s.db.Get(&activity, `SELECT
		(SELECT coalesce(sum(calls), 0) FROM "AgentToolStat" WHERE agent = $1 AND day > current_date - 7) +
		(SELECT count(*) FROM "AgentGap" WHERE agent = $1 AND "createdAt" > now() - interval '7 days') +
		(SELECT count(*) FROM "AgentFeedback" WHERE agent = $1 AND "createdAt" > now() - interval '7 days')`, s.name))
	if activity == 0 {
		return nil
	}
	system := "إنت " + model.AgentName(s.name) + "، وكيل نظام شركة الأماني. هذي مراجعتك الذاتية الأسبوعية بلا محادثة: تطلّع على أرقامك وتقرر شنو تحسّن. أدوات قراءة واقتراح بس، ولا تسوي أي إجراء ثاني."
	task := "١) استدعي my_performance. ٢) حلّل: شنو أكثر شي فشل، وشنو الأسئلة الي ما گدرت تجاوبها، وشنو اشتكى منه المستخدمين (👎). " +
		"٣) لكل مشكلة حقيقية متكررة: إذا تحتاج أداة أو برمجة سجّلها بـ suggest_improvement (needs_code=true)، وإذا غلط بطريقة جوابك اقترح درس بـ propose_lesson. بالكثير ٣ بالمجموع، وبس الي يستاهل. " +
		"٤) بالأخير اكتب تقرير قصير (٥ أسطر بالكثير): شنو تحسّن، شنو بعده ضعيف، وشنو طلبت."
	run, err := s.agent.Run("LEARNING", system, task)
	if err != nil || run == nil || strings.TrimSpace(run.Answer) == "" {
		return err
	}
	title := "📋 مراجعة ذاتية أسبوعية " + now.In(debriefLoc).Format("2006-01-02")
	errlog.Exec(s.db.Exec(`INSERT INTO "AiSuggestion" (agent, title, details, "needsCode") VALUES ($1, $2, $3, false)`,
		s.name, title, strings.TrimSpace(run.Answer)))
	return nil
}
