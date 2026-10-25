package service

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

// ═══ مهمة بالفويس — قرار (ع) 10-19 ═══
// «المدير يوجّه مهمة لأي موظف برسالة صوتية، هايكو يفهمها ويدزها على شكل مهمة».
//
// المسار: الصوت ينحوّل لنص بـWhisper **الداخلي** (ما يطلع)، أسماء الموظفين
// تنطابق بسيرفرنا وتنشال من النص، وهايكو يشوف النص بلا أسماء ويرتّبه
// (عنوان، تفاصيل، أولوية، موعد). الناتج **مسودة** — المدير يعدّلها ويختار
// الموظف ويضغط «وجّه» بنفسه. الصوت نفسه ما ينحفظ (قراره: النص بس).

// VoiceTaskEmployee موظف بقائمة المطابقة.
type VoiceTaskEmployee struct {
	ID   string
	Name string
}

// VoiceTaskDraft مسودة المهمة الي ترجع للمدير.
type VoiceTaskDraft struct {
	Transcript           string   `json:"transcript"`
	Title                string   `json:"title"`
	Description          string   `json:"description"`
	Priority             string   `json:"priority"` // NORMAL | URGENT
	DueAt                string   `json:"dueAt"`    // YYYY-MM-DDTHH:MM بتوقيت بغداد، أو فاضي
	SuggestedEmployeeIDs []string `json:"suggestedEmployeeIds"`
	Understood           bool     `json:"understood"` // false = هايكو ما اشتغل، المسودة من النص الخام
}

// matchEmployees يدوّر أسماء الموظفين بالنص: الاسم الكامل، أو أول كلمتين،
// أو الاسم الأول لو ٤ أحرف فأكثر — نفس قاعدة redactNames، وحروف الجر
// الملتصقة («لعلي»، «ومصطفى») مقبولة.
func matchEmployees(text string, emps []VoiceTaskEmployee) []string {
	ids := []string{}
	for _, e := range emps {
		w := strings.Fields(strings.TrimSpace(e.Name))
		if len(w) == 0 {
			continue
		}
		cands := []string{strings.Join(w, " ")}
		if len(w) >= 2 {
			cands = append(cands, w[0]+" "+w[1])
		}
		if len([]rune(w[0])) >= 4 {
			cands = append(cands, w[0])
		}
		for _, c := range cands {
			re, err := regexp.Compile(`(^|[^\p{L}])(?:و|ف|ب|ل|ك)?` + regexp.QuoteMeta(c) + `($|[^\p{L}])`)
			if err == nil && re.MatchString(text) {
				ids = append(ids, e.ID)
				break
			}
		}
	}
	return ids
}

const voiceTaskPrompt = `أنت «ماتركس» بنظام شركة الأماني (العراق). يوصلك نص فويس سجّله مدير يوجّه بيه مهمة لموظف.
النص محوّل آلياً من الصوت، فبيه أخطاء إملائية — افهم المقصود. الأسماء مشالة ومكانها [موظف] أو [زبون].
رجّع JSON بس، بلا أي كلام قبله أو بعده:
{"title":"عنوان قصير للمهمة (لحد ٨ كلمات)","description":"شنو المطلوب بالضبط بالعراقي البسيط","priority":"NORMAL","dueAt":""}
قواعد:
- priority: "URGENT" بس إذا المدير گال صراحة مستعجل/هسه/فوراً/ضروري اليوم، وإلا "NORMAL".
- dueAt: بصيغة YYYY-MM-DDTHH:MM بتوقيت بغداد، بس إذا المدير ذكر موعد (باچر، الساعة ٣، الخميس...). احسبه من «الوقت الحالي» المعطى. ما ذكر موعد = "".
- لا تضيف شي ما انگال، ولا تحكم على أحد، ولا تذكر غرامة أو عقوبة.`

// TaskDraft يحوّل فويس المدير لمسودة مهمة.
func (s *VoiceAnalysisService) TaskDraft(ctx context.Context, audio []byte, contentType string, emps []VoiceTaskEmployee) (*VoiceTaskDraft, error) {
	if !s.Enabled() {
		return nil, errors.New("تحويل الصوت لنص مطفي على السيرفر (WHISPER_URL)")
	}
	text, err := s.transcribe(ctx, audio, contentType)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("ما انسمع كلام بالفويس — سجّل مرة ثانية")
	}
	d := &VoiceTaskDraft{Transcript: text, Priority: "NORMAL", SuggestedEmployeeIDs: matchEmployees(text, emps)}

	names := make([]string, 0, len(emps))
	for _, e := range emps {
		names = append(names, e.Name)
	}
	if u, err := s.understandTask(ctx, redactNames(text, names, "")); err == nil {
		d.Title, d.Description, d.Priority, d.DueAt, d.Understood = u.Title, u.Description, u.Priority, u.DueAt, true
	}
	if !d.Understood {
		// بلا هايكو: المسودة من النص نفسه، والمدير يرتّبها.
		r := []rune(text)
		if len(r) > 60 {
			r = append(r[:60], '…')
		}
		d.Title, d.Description = string(r), text
	}
	// [موظف] بالمسودة: لو انطابق موظف وحد، نرجّع اسمه (بسيرفرنا، بعد هايكو).
	if len(d.SuggestedEmployeeIDs) == 1 {
		for _, e := range emps {
			if e.ID == d.SuggestedEmployeeIDs[0] {
				first := strings.Fields(e.Name)[0]
				d.Title = strings.ReplaceAll(d.Title, "[موظف]", first)
				d.Description = strings.ReplaceAll(d.Description, "[موظف]", first)
			}
		}
	}
	return d, nil
}

type voiceTaskUnderstanding struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	DueAt       string `json:"dueAt"`
}

func (s *VoiceAnalysisService) understandTask(ctx context.Context, text string) (*voiceTaskUnderstanding, error) {
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

	if r := []rune(text); len(r) > 3000 {
		text = string(r[:3000])
	}
	now := time.Now().In(debriefLoc)
	msg := "الوقت الحالي: " + now.Format("2006-01-02T15:04") + " (" + now.Weekday().String() + ")\nنص الفويس:\n" + text
	resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(s.model),
		MaxTokens: 400,
		System:    []anthropic.TextBlockParam{{Text: voiceTaskPrompt}},
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(msg))},
	})
	if err != nil {
		return nil, err
	}
	RecordAIUsage("VOICE", s.model, resp.Usage)
	var b strings.Builder
	for _, block := range resp.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			b.WriteString(tb.Text)
		}
	}
	return parseVoiceTask(b.String())
}

// parseVoiceTask يفك رد هايكو وينظّفه — رد النموذج مدخل خارجي.
func parseVoiceTask(raw string) (*voiceTaskUnderstanding, error) {
	if i, j := strings.Index(raw, "{"), strings.LastIndex(raw, "}"); i >= 0 && j > i {
		raw = raw[i : j+1]
	}
	var u voiceTaskUnderstanding
	if err := json.Unmarshal([]byte(raw), &u); err != nil {
		return nil, errors.New("رد هايكو مو JSON")
	}
	u.Title = strings.TrimSpace(u.Title)
	if u.Title == "" {
		return nil, errors.New("هايكو ما طلّع عنوان")
	}
	if r := []rune(u.Title); len(r) > 120 {
		u.Title = string(r[:120])
	}
	if r := []rune(u.Description); len(r) > 2000 {
		u.Description = string(r[:2000])
	}
	if u.Priority != "URGENT" {
		u.Priority = "NORMAL"
	}
	if _, err := time.Parse("2006-01-02T15:04", u.DueAt); err != nil {
		u.DueAt = ""
	}
	return &u, nil
}

// arNorm يوحّد الحروف الي تنكتب بأكثر من شكل (أ/إ/آ ← ا، ة ← ه، ى ← ي) — حتى
// «فاطمة» و«فاطمه» تنطابق.
var arNorm = strings.NewReplacer("أ", "ا", "إ", "ا", "آ", "ا", "ة", "ه", "ى", "ي", "ـ", "")

// matchEmployeeBest للحقل الي ياخذ موظف واحد (ايريس — دازاي 10-25: «لكيت أكثر
// من اسم» حتى لو گال الاسم كامل). كل موظف ينحسب بشكد كلمة من اسمه انذكرت
// بالترتيب؛ الأعلى يفوز. «علي حسين كاظم» ما يتساوى ويا «علي حسين جبار» إذا
// انگال الاسم الثلاثي. يرجّع أكثر من واحد بس إذا تعادلوا فعلاً.
func matchEmployeeBest(text string, emps []VoiceTaskEmployee) []string {
	t := arNorm.Replace(strings.TrimSpace(text))
	best, ids := 0, []string{}
	for _, e := range emps {
		w := strings.Fields(arNorm.Replace(strings.TrimSpace(e.Name)))
		score := 0
		for n := len(w); n >= 1; n-- {
			c := strings.Join(w[:n], " ")
			if n == 1 && len([]rune(c)) < 3 {
				break
			}
			re, err := regexp.Compile(`(^|[^\p{L}])(?:و|ف|ب|ل|ك)?` + regexp.QuoteMeta(c) + `($|[^\p{L}])`)
			if err == nil && re.MatchString(t) {
				score = n
				// الاسم الكامل بالضبط يغلب أي تطابق جزئي ويا غيره
				if n == len(w) {
					score += 10
				}
				break
			}
		}
		switch {
		case score == 0:
		case score > best:
			best, ids = score, []string{e.ID}
		case score == best:
			ids = append(ids, e.ID)
		}
	}
	return ids
}
