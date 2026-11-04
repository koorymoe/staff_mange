package service

import (
	"context"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

// ═══ صوت ماتركس — كل تنبيه بأسلوب جديد (قرار (ع) 10-10) ═══
// «كلامه وطريقة تنبيهه لازم يكون كأنما ذكاء اصطناعي — بين فترة وفترة يتغيّر
// الأسلوب، ما يضل يحاور بنفس الطريقة». المعنى والأرقام ثابتة بالكود، وهايكو
// يغيّر الصياغة بس. السياج: كل رقم وكود بالأصل لازم يبقى، ولا كلمة غرامة أو
// تهديد؛ وأي خلل يرجّع الأصل.

var voiceTones = []string{
	"مباشر ومختصر، جملة أو جملتين",
	"ودّي وقريب، مثل زميل يذكّر زميله",
	"تشجيعي يركّز على إنه يگدر يخلّصها بسرعة",
	"يبدي بسؤال خفيف يحفّز، وبعدين الطلب",
	"تذكير هادئ ومحترم",
	"رسمي خفيف ومرتّب",
	"عملي: شنو المطلوب بالضبط وشنو الخطوة الجاية",
}

var voiceBanned = []string{"غرامة", "خصم", "عقوبة", "فصل", "طرد", "راح تنحاسب", "تهديد"}

// الأرقام (عربية أو هندية) وأكواد الحجوزات والمشاريع — لازم تبقى كلها.
var voiceFactRe = regexp.MustCompile(`[A-Za-z]+-?[A-Za-z0-9]*\d+[A-Za-z0-9-]*|\d+(?:[.,:/-]\d+)*|[٠-٩]+`)

func voiceFacts(s string) []string { return voiceFactRe.FindAllString(s, -1) }

// voiceSafe الصياغة الجديدة تحافظ على الحقائق وما بيها كلام ممنوع؟
func voiceSafe(orig, out string) bool {
	if strings.TrimSpace(out) == "" || len([]rune(out)) > 2*len([]rune(orig))+80 {
		return false
	}
	for _, f := range voiceFacts(orig) {
		if !strings.Contains(out, f) {
			return false
		}
	}
	for _, b := range voiceBanned {
		if strings.Contains(out, b) && !strings.Contains(orig, b) {
			return false
		}
	}
	return true
}

// Rewrite يعيد صياغة تنبيه ماتركس لموظف — وأي فشل يرجّع الأصل.
func (a *MatrixAgent) Rewrite(recent []string, message string) string {
	if !a.Enabled("TONE") || strings.TrimSpace(message) == "" {
		return message
	}
	mask := a.newMask()
	hidden := mask.Hide(message)
	prev := make([]string, 0, len(recent))
	for _, r := range recent {
		prev = append(prev, "- "+mask.Hide(r))
	}
	tone := voiceTones[rand.Intn(len(voiceTones))]
	prompt := "أعد صياغة تنبيه «ماتركس» هذا لموظف بالشركة، باللهجة العراقية، بأسلوب: " + tone + ".\n" +
		"شروط: نفس المعنى والطلب بالضبط. كل رقم وكود وتاريخ ورمز «موظف#n» يبقى حرفياً. لا تضيف معلومة جديدة، ولا غرامة ولا تهديد. " +
		"خليه طبيعي مثل إنسان مو قالب، وابدي بغير الطريقة الي بدت بيها الرسائل السابقة. رجّع نص الرسالة بس.\n\n" +
		"الرسائل السابقة لنفس الموظف (لا تكررها):\n" + strings.Join(prev, "\n") + "\n\nالتنبيه:\n" + hidden
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	resp, err := a.llm.New(ctx, anthropic.MessageNewParams{
		Model: anthropic.Model(a.model), MaxTokens: 400, Temperature: anthropic.Float(1),
		Messages: []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(prompt))},
	})
	if err != nil {
		return message
	}
	RecordAIUsage("TONE", a.model, resp.Usage)
	var b strings.Builder
	for _, block := range resp.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			b.WriteString(tb.Text)
		}
	}
	out := strings.TrimSpace(b.String())
	if !voiceSafe(hidden, out) {
		return message
	}
	out = mask.Show(out)
	if !strings.HasPrefix(out, "🤖") {
		out = "🤖 " + out
	}
	return out
}
