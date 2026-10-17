package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"staffmange-api/internal/repository"
	"staffmange-api/internal/storage"
)

// ═══ ماتركس يسمع الفويس — طلب (ع) 10-05 ═══
//
// ١. الصوت يتحول لنص **داخل سيرفرنا** (Whisper بحاوية «whisper» على الشبكة
//    الداخلية). الصوت ما يطلع من السيرفر أبداً.
// ٢. من النص تنشال أسماء الموظفين واسم زبون الحجز المربوط.
// ٣. النص النظيف بس يروح لهايكو: ملخّص، الشغل المعدود، طريقة الكلام،
//    والعبارات العراقية (تتجمع بقاموس يتعلّم منه ماتركس اللهجة).
//
// ⚠️ ماكو نقاط ولا غرامات من هذا — قراءة ومساعدة للمدير بس.

const voiceDailyCap = 150

type VoiceAnalysisService struct {
	voices     *repository.AchievementVoiceRepository
	store      storage.Store
	whisperURL string
	http       *http.Client
	client     *anthropic.Client
	model      string

	mu      sync.Mutex
	running map[string]bool
	day     string
	used    int

	dialectMu   sync.Mutex
	dialectAt   time.Time
	dialectText string
}

func NewVoiceAnalysisService(voices *repository.AchievementVoiceRepository, store storage.Store, whisperURL string) *VoiceAnalysisService {
	return &VoiceAnalysisService{voices: voices, store: store, whisperURL: strings.TrimRight(whisperURL, "/"),
		http: &http.Client{Timeout: 5 * time.Minute}, running: map[string]bool{}}
}

func (s *VoiceAnalysisService) EnableModel(apiKey, modelName string) {
	if apiKey == "" {
		return
	}
	if modelName == "" {
		modelName = "claude-haiku-4-5"
	}
	c := anthropic.NewClient(option.WithAPIKey(apiKey))
	s.client, s.model = &c, modelName
}

// Enabled التحويل لنص شغّال؟ (بلا Whisper الفويس يبقى محفوظ بلا تحليل).
func (s *VoiceAnalysisService) Enabled() bool { return s.whisperURL != "" }

// Start عامل الخلفية: كل دقيقتين ياخذ المعلّق.
func (s *VoiceAnalysisService) Start() {
	if !s.Enabled() {
		log.Printf("[voice] WHISPER_URL فاضي — الفويس ينحفظ بلا تحليل")
		return
	}
	go func() {
		for {
			s.RunPending()
			time.Sleep(2 * time.Minute)
		}
	}()
}

func (s *VoiceAnalysisService) RunPending() {
	rows, err := s.voices.Pending(5)
	if err != nil {
		return
	}
	for _, v := range rows {
		s.Analyze(v.AchievementID)
	}
}

// Kick بعد الرفع مباشرة — بالخلفية.
func (s *VoiceAnalysisService) Kick(id string) {
	if !s.Enabled() {
		return
	}
	s.voices.ResetAnalysis(id)
	go s.Analyze(id)
}

func (s *VoiceAnalysisService) Analyze(id string) {
	s.mu.Lock()
	if s.running[id] {
		s.mu.Unlock()
		return
	}
	s.running[id] = true
	s.mu.Unlock()
	defer func() { s.mu.Lock(); delete(s.running, id); s.mu.Unlock() }()

	if err := s.analyze(id); err != nil {
		log.Printf("[voice] %s: %v", id, err)
		s.voices.MarkAttempt(id, err.Error())
	}
}

func (s *VoiceAnalysisService) analyze(id string) error {
	v, err := s.voices.Get(id)
	if err != nil {
		return errors.New("الفويس مو موجود")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	data, _, err := s.store.Get(ctx, v.FileKey)
	if err != nil {
		return errors.New("تعذر قراءة الفويس من التخزين")
	}
	text, err := s.transcribe(ctx, data, v.ContentType)
	if err != nil {
		return err
	}
	if strings.TrimSpace(text) == "" {
		return s.voices.SaveAnalysis(id, "", "الفويس ما بيه كلام مفهوم.", "", nil)
	}
	clean := redactNames(text, s.voices.Names(), s.voices.CustomerNameFor(id))
	res, err := s.understand(ctx, clean)
	if err != nil {
		// النص ينحفظ حتى لو هايكو ما رد — المدير يقراه.
		log.Printf("[voice] هايكو: %v", err)
		return s.voices.SaveAnalysis(id, clean, "", "", nil)
	}
	for _, p := range res.Phrases {
		ph, mn := strings.TrimSpace(p.Phrase), strings.TrimSpace(p.Meaning)
		if ph != "" && mn != "" && len([]rune(ph)) <= 60 && len([]rune(mn)) <= 120 {
			s.voices.LearnPhrase(ph, mn)
		}
	}
	return s.voices.SaveAnalysis(id, clean, res.Summary, res.Style, res.Tasks)
}

// transcribe يرسل الصوت لـWhisper **الداخلي** (حاوية على نفس السيرفر).
func (s *VoiceAnalysisService) transcribe(ctx context.Context, audio []byte, contentType string) (string, error) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("file", "voice"+voiceExtFor(contentType))
	_, _ = fw.Write(audio)
	_ = mw.WriteField("language", "ar")
	_ = mw.WriteField("response_format", "json")
	_ = mw.WriteField("temperature", "0")
	_ = mw.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.whisperURL+"/inference", &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := s.http.Do(req)
	if err != nil {
		return "", errors.New("خدمة تحويل الصوت ما ترد")
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("تحويل الصوت فشل (%d)", resp.StatusCode)
	}
	var out struct {
		Text  string `json:"text"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", errors.New("رد غير متوقع من تحويل الصوت")
	}
	if out.Error != "" {
		return "", errors.New("تحويل الصوت: " + out.Error)
	}
	return strings.TrimSpace(out.Text), nil
}

func voiceExtFor(ct string) string {
	switch ct {
	case "audio/webm":
		return ".webm"
	case "audio/ogg":
		return ".ogg"
	case "audio/mp4":
		return ".m4a"
	}
	return ".wav"
}

// redactNames يشيل أسماء الموظفين واسم الزبون من النص قبل ما يطلع لهايكو.
// كل اسم ينفحص كامل، وبكلمتين الأولى (الاسم الثنائي)، والاسم الأول لحاله
// إذا طوله ٤ أحرف أو أكثر (حتى ما نشيل كلمات عادية قصيرة).
func redactNames(text string, employees []string, customer string) string {
	type pat struct {
		s, tag string
	}
	pats := []pat{}
	add := func(full, tag string) {
		w := strings.Fields(strings.TrimSpace(full))
		if len(w) == 0 {
			return
		}
		pats = append(pats, pat{strings.Join(w, " "), tag})
		if len(w) >= 2 {
			pats = append(pats, pat{w[0] + " " + w[1], tag})
		}
		if len([]rune(w[0])) >= 4 {
			pats = append(pats, pat{w[0], tag})
		}
	}
	if customer != "" {
		add(customer, "[زبون]")
	}
	for _, n := range employees {
		add(n, "[موظف]")
	}
	// الأطول أول حتى «علي حسن» ما ينقص لـ«[موظف] حسن».
	sort.SliceStable(pats, func(i, j int) bool { return len([]rune(pats[i].s)) > len([]rune(pats[j].s)) })
	out := text
	for _, p := range pats {
		// حروف الجر والعطف تلتصق بالاسم بالعربي («ومصطفى»، «لعلي») — نخليها ونشيل الاسم.
		re, err := regexp.Compile(`(^|[^\p{L}])((?:و|ف|ب|ل|ك)?)` + regexp.QuoteMeta(p.s) + `($|[^\p{L}])`)
		if err != nil {
			continue
		}
		// مرتين: التطابقات المتلاصقة («علي، علي») تاكل الفاصل بينها بأول مرة.
		for i := 0; i < 2; i++ {
			out = re.ReplaceAllString(out, "${1}${2}"+p.tag+"${3}")
		}
	}
	return out
}

type voiceUnderstanding struct {
	Summary string                 `json:"summary"`
	Tasks   []repository.VoiceTask `json:"tasks"`
	Style   string                 `json:"style"`
	Phrases []struct {
		Phrase  string `json:"phrase"`
		Meaning string `json:"meaning"`
	} `json:"phrases"`
}

const voiceSystemPrompt = `أنت «ماتركس» بنظام شركة الأماني (العراق). يوصلك نص فويس سجّله موظف عن شغله اليوم.
النص محوّل آلياً من الصوت، فبيه أخطاء إملائية — افهم المقصود. الأسماء مشالة ومكانها [موظف] أو [زبون].
رجّع JSON بس، بلا أي كلام قبله أو بعده، بهالشكل:
{"summary":"جملتين بالعراقي البسيط عن شنو سوّى","tasks":[{"what":"حجز منجز","count":3}],"style":"وصف قصير لطريقة كلامه: مرتب/مشتت، واضح/غامض، مستعجل/هادي","phrases":[{"phrase":"عبارة عراقية انگالت","meaning":"معناها بالفصحى"}]}
قواعد:
- tasks: بس الشغل الي انذكر صراحة بعدد. لا تخمّن عدد ما انگال.
- phrases: لحد ٥ عبارات عامية عراقية حقيقية من النص (مو كلمات فصحى عادية)، بلا أسماء.
- لا تحكم على الموظف ولا تقترح عقوبة أو نقاط.`

func (s *VoiceAnalysisService) understand(ctx context.Context, text string) (*voiceUnderstanding, error) {
	if !AIFeatureOn("VOICE") {
		return nil, ErrAIFeatureOff
	}
	if s.client == nil {
		return nil, errors.New("ماكو مفتاح هايكو")
	}
	s.mu.Lock()
	today := time.Now().Format("2006-01-02")
	if s.day != today {
		s.day, s.used = today, 0
	}
	if s.used >= voiceDailyCap {
		s.mu.Unlock()
		return nil, errors.New("وصل الحد اليومي لتحليل الفويس")
	}
	s.used++
	s.mu.Unlock()

	if r := []rune(text); len(r) > 6000 {
		text = string(r[:6000])
	}
	resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(s.model),
		MaxTokens: 700,
		System:    []anthropic.TextBlockParam{{Text: voiceSystemPrompt}},
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock("نص الفويس:\n" + text))},
	})
	if err == nil {
		RecordAIUsage("VOICE", s.model, resp.Usage)
	}
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	for _, block := range resp.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			b.WriteString(tb.Text)
		}
	}
	raw := b.String()
	if i, j := strings.Index(raw, "{"), strings.LastIndex(raw, "}"); i >= 0 && j > i {
		raw = raw[i : j+1]
	}
	var out voiceUnderstanding
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, errors.New("رد هايكو مو JSON")
	}
	return &out, nil
}

// DialectHint أكثر العبارات العراقية الي سمعها ماتركس — تنضاف لتوجيهاته حتى
// يحچي أقرب للهجة الموظفين. تتحدّث كل ساعة.
func (s *VoiceAnalysisService) DialectHint() string {
	s.dialectMu.Lock()
	defer s.dialectMu.Unlock()
	if time.Since(s.dialectAt) < time.Hour {
		return s.dialectText
	}
	s.dialectAt = time.Now()
	rows, err := s.voices.Phrases(30)
	if err != nil || len(rows) == 0 {
		s.dialectText = ""
		return ""
	}
	var b strings.Builder
	b.WriteString("عبارات عراقية يستعملها موظفينا (من فويساتهم) — استعملها إذا تناسب الموقف:\n")
	for _, p := range rows {
		fmt.Fprintf(&b, "- %s = %s\n", p.Phrase, p.Meaning)
	}
	s.dialectText = b.String()
	return s.dialectText
}
