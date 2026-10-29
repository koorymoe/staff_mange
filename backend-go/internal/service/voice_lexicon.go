package service

import (
	_ "embed"
	"errors"
	"log"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ 👂 «أذن» الأماني — قاموس يتعلم (دازاي 10-29) ═══
// «اريد يفهم العراقي… يتعلم من المصطلحات الي تمر عليه… كتبت غلط يدري شنو اقصد».
// ثلاث طبقات:
//   ١. ورقة المصطلحات (docs/iraqi-voice-glossary.md — نسخة مدمجة هنا) + كلمات النظام
//      (الخدمات، المنتجات، المواد، الأقسام، السيارات، أسماء الموظفين) تتحدث لحالها.
//   ٢. القاموس يروح للأذن (Gemini) ويا كل بصمة، فتكتب كلماتنا صح.
//   ٣. ذاكرة التصحيح (VoiceFix): «سمعها X وهي Y» — تنطبق على كل نص يطلع من الأذن.
//      تنحفظ من تصحيح المالك لكمال، ومن ورقة المصطلحات (بس لكلمات الشغل).
// نفس الأذن تخدم كمال هسه، والواتساب وتطبيق الزبائن بعدين.

//go:embed voice_glossary_seed.md
var voiceGlossarySeed string

// VoiceLexicon القاموس — نسخة بالذاكرة تتحدث كل ١٠ دقايق.
type VoiceLexicon struct {
	db      *sqlx.DB
	mu      sync.RWMutex
	prompt  string
	fixes   []voiceFix
	loaded  time.Time
	maxTerm int
}

type voiceFix struct {
	re    *regexp.Regexp
	right string
}

func NewVoiceLexicon(db *sqlx.DB) *VoiceLexicon {
	return &VoiceLexicon{db: db, maxTerm: 400}
}

var glossLineRe = regexp.MustCompile(`^- (.+?)(?: ← (.+?))?(?: ← (.+))?$`)

var seedVoiceFixes = map[string]string{
	"الدش كاب": "الداش كام", "داشكام": "داش كام", "دشكام": "داش كام", "دش كام": "داش كام",
	"جيبي اس": "جي بي اس", "جي بي إس": "جي بي اس",
	"ماتريكس": "ماتركس", "منضومة": "منظومة", "انفيرتر": "انفرتر", "انفرتير": "انفرتر",
	"راوتير": "راوتر", "سوتش": "سويچ", "انضار": "إنذار", "عرض صعر": "عرض سعر",
	// (دازاي 10-29) من أول تجربة صوت حقيقية
	"شكبارة": "شخبارك", "شكبارك": "شخبارك", "شخبارة": "شخبارك", "زيان": "زين",
}

// SeedGlossary يدخّل ورقة المصطلحات (مرة، بلا تكرار) كتلميح للأذن.
func (l *VoiceLexicon) SeedGlossary() {
	section := ""
	n := 0
	l.db.Exec(`DELETE FROM "VoiceFix" WHERE source = 'SEED'`)
	for _, line := range strings.Split(voiceGlossarySeed, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## ") {
			section = strings.TrimPrefix(line, "## ")
			continue
		}
		m := glossLineRe.FindStringSubmatch(line)
		if m == nil || strings.TrimSpace(m[1]) == "" {
			continue
		}
		terms := splitTerms(m[1])
		meaning := strings.TrimSpace(m[3])
		if meaning == "" {
			meaning = strings.TrimSpace(m[2])
		}
		for _, t := range terms {
			if _, err := l.db.Exec(`INSERT INTO "VoiceTerm" (term, meaning, category, source) VALUES ($1, $2, $3, 'SEED')
				ON CONFLICT (term) DO NOTHING`, t, meaning, section); err == nil {
				n++
			}
		}
	}
	// تصحيحات مضمونة بس — عمود «الأشكال» بالورقة مو كله غلط إملائي (بيه أصناف وأمثلة)،
	// فالباقي تلميح للأذن، والتصحيح الحقيقي يجي من المالك بـlearn_term.
	for wrong, right := range seedVoiceFixes {
		l.AddFix(wrong, right, "SEED")
	}
	_ = n
}

func splitTerms(s string) []string {
	out := []string{}
	for _, p := range regexp.MustCompile(`[/،,]`).Split(s, -1) {
		p = strings.TrimSpace(regexp.MustCompile(`\(.*?\)`).ReplaceAllString(p, ""))
		if p != "" && len([]rune(p)) <= 40 {
			out = append(out, p)
		}
	}
	return out
}

// HarvestSystem يسحب كلمات النظام (أسماء الخدمات والمنتجات… والموظفين) للقاموس.
func (l *VoiceLexicon) HarvestSystem() int {
	sources := map[string]string{
		"خدمة":   `SELECT DISTINCT name FROM "Service"`,
		"منتج":   `SELECT DISTINCT name FROM "Product" UNION SELECT DISTINCT "modelName" FROM "Product" WHERE "modelName" IS NOT NULL`,
		"مادة":   `SELECT DISTINCT name FROM "Material"`,
		"ديكور":  `SELECT DISTINCT name FROM "DecorRate"`,
		"قسم":    `SELECT DISTINCT name FROM "Department"`,
		"سيارة":  `SELECT DISTINCT name FROM "Vehicle"`,
		"موظف":   `SELECT DISTINCT name FROM "Employee" WHERE status = 'ACTIVE'`,
		"منظومة": `SELECT DISTINCT "systemName" FROM "SystemPriceCatalog"`,
	}
	n := 0
	for cat, q := range sources {
		var names []string
		if err := l.db.Select(&names, q); err != nil {
			continue
		}
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name == "" || len([]rune(name)) > 60 {
				continue
			}
			res, err := l.db.Exec(`INSERT INTO "VoiceTerm" (term, meaning, category, source) VALUES ($1, $2, $3, 'SYSTEM')
				ON CONFLICT (term) DO NOTHING`, name, cat, "النظام — "+cat)
			if err == nil {
				if k, _ := res.RowsAffected(); k > 0 {
					n++
				}
			}
		}
	}
	l.invalidate()
	return n
}

// AddFix «سمعها wrong وهي right».
func (l *VoiceLexicon) AddFix(wrong, right, source string) {
	wrong, right = strings.TrimSpace(wrong), strings.TrimSpace(right)
	if wrong == "" || right == "" || wrong == right || len([]rune(wrong)) < 2 {
		return
	}
	_, _ = l.db.Exec(`INSERT INTO "VoiceFix" (wrong, "right", source) VALUES ($1, $2, $3)
		ON CONFLICT (wrong) DO UPDATE SET "right" = EXCLUDED."right", source = EXCLUDED.source`, wrong, right, source)
	l.invalidate()
}

// Learn كلمة جديدة يتعلمها (من المالك أو من محادثة) — ويا شكلها الغلط إذا معروف.
func (l *VoiceLexicon) Learn(term, meaning, heardAs, source string) error {
	term = strings.TrimSpace(term)
	if term == "" {
		return nil
	}
	if _, err := l.db.Exec(`INSERT INTO "VoiceTerm" (term, meaning, category, source) VALUES ($1, $2, 'تعلّم', $3)
		ON CONFLICT (term) DO UPDATE SET meaning = COALESCE(NULLIF(EXCLUDED.meaning, ''), "VoiceTerm".meaning), "updatedAt" = now()`,
		term, strings.TrimSpace(meaning), source); err != nil {
		return err
	}
	if heardAs != "" {
		l.AddFix(heardAs, term, source)
	}
	l.invalidate()
	return nil
}

func (l *VoiceLexicon) Invalidate() { l.invalidate() }

func (l *VoiceLexicon) invalidate() {
	l.mu.Lock()
	l.loaded = time.Time{}
	l.mu.Unlock()
}

func (l *VoiceLexicon) load() {
	l.mu.RLock()
	fresh := time.Since(l.loaded) < 10*time.Minute
	l.mu.RUnlock()
	if fresh {
		return
	}
	terms := []struct {
		Term string `db:"term"`
	}{}
	_ = l.db.Select(&terms, `SELECT term FROM "VoiceTerm" ORDER BY (source IN ('OWNER', 'LEARNED')) DESC, (category = 'موظف') DESC, (source = 'SEED') DESC, uses DESC, "createdAt" LIMIT $1`, l.maxTerm)
	words := make([]string, 0, len(terms))
	for _, t := range terms {
		words = append(words, t.Term)
	}
	fixRows := []struct {
		Wrong string `db:"wrong"`
		Right string `db:"right"`
	}{}
	_ = l.db.Select(&fixRows, `SELECT wrong, "right" FROM "VoiceFix"`)
	// الأطول أول حتى «الدش كاب» تنطبق قبل «دش»
	sort.Slice(fixRows, func(i, j int) bool { return len([]rune(fixRows[i].Wrong)) > len([]rune(fixRows[j].Wrong)) })
	fixes := make([]voiceFix, 0, len(fixRows))
	for _, f := range fixRows {
		re, err := regexp.Compile(`(^|[^\p{L}])` + regexp.QuoteMeta(f.Wrong) + `($|[^\p{L}])`)
		if err == nil {
			fixes = append(fixes, voiceFix{re: re, right: f.Right})
		}
	}
	l.mu.Lock()
	l.prompt = strings.Join(words, "، ")
	l.fixes = fixes
	l.loaded = time.Now()
	l.mu.Unlock()
}

// PromptHint قائمة الكلمات للأذن.
func (l *VoiceLexicon) PromptHint() string {
	if l == nil {
		return ""
	}
	l.load()
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.prompt == "" {
		return ""
	}
	return "\nمرجع إملاء بس (كلمات شغل شركة الأماني وأسماء): إذا انگالت وحدة منها فعلاً اكتبها بهالشكل. لا تبدّل كلام عادي بكلمة من القائمة، والسلام والكلام اليومي اكتبه مثل ما انگال:\n" + l.prompt
}

// Apply يطبّق ذاكرة التصحيح على نص طالع من الأذن.
func (l *VoiceLexicon) Apply(text string) string {
	if l == nil || text == "" {
		return text
	}
	l.load()
	l.mu.RLock()
	defer l.mu.RUnlock()
	for _, f := range l.fixes {
		text = f.re.ReplaceAllString(text, "${1}"+f.right+"${2}")
	}
	return text
}

// StartVoiceLexicon يزرع الورقة ويسحب كلمات النظام (وبعدين كل ٦ ساعات).
func StartVoiceLexicon(l *VoiceLexicon) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("voice lexicon: %v", r)
			}
		}()
		l.SeedGlossary()
		n := l.HarvestSystem()
		log.Printf("👂 قاموس الأذن جاهز (+%d كلمة من النظام)", n)
	}()
}

// DefaultVoiceLexicon الأذن المشتركة (كمال هسه، والواتساب وتطبيق الزبائن بعدين). nil = مطفية.
var DefaultVoiceLexicon *VoiceLexicon

func registerVoiceTools(a *MatrixAgent, tc *SecretaryTurnCtx) {
	if tc == nil || DefaultVoiceLexicon == nil {
		return
	}
	a.AddTool(agentTool{Name: "learn_term",
		Desc: "علّم أذنك كلمة/مصطلح: من يصحّحلك كلمة انكتبت غلط بالصوت، أو يشرحلك كلمة عراقية أو شغل ما تعرفها. term = الكلمة الصح، heard_as = الشكل الغلط الي انكتب (إذا اكو)، meaning = معناها.",
		Props: map[string]any{
			"term":     map[string]any{"type": "string"},
			"meaning":  map[string]any{"type": "string"},
			"heard_as": map[string]any{"type": "string"},
		},
		Req: []string{"term"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			term := strings.TrimSpace(strArg(in, "term"))
			if l := len([]rune(term)); l < 2 || l > 60 {
				return "", errors.New("الكلمة لازم بين ٢ و٦٠ حرف")
			}
			if err := DefaultVoiceLexicon.Learn(term, strArg(in, "meaning"), strArg(in, "heard_as"), "OWNER"); err != nil {
				return "", err
			}
			return "تعلّمتها. گوله بكلمتين إنك حفظتها.", nil
		}})
}

// TopTerms أهم الكلمات (المالك أول، بعدين أسماء الموظفين وكلمات الشغل).
func (l *VoiceLexicon) TopTerms(n int) []string {
	if l == nil {
		return nil
	}
	l.load()
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.prompt == "" {
		return nil
	}
	all := strings.Split(l.prompt, "، ")
	if len(all) > n {
		all = all[:n]
	}
	return all
}

func urlEscape(s string) string { return url.QueryEscape(s) }
