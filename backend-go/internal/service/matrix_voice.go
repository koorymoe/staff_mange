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

// voiceMaxGrowth أقصى نسبة طول الصياغة للأصل — أطول من هيچ يعني ضاف كلام.
const voiceMaxGrowth = 1.6

// voiceTagRe رموز الموظفين المخفية.
var voiceTagRe = regexp.MustCompile(`موظف#\d+`)

func voiceSet(xs []string) map[string]bool {
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}
	return m
}

// voiceSafe الصياغة الجديدة تحافظ على الحقائق وما بيها كلام ممنوع؟
// (دازاي 11-15) «إعادة الصياغة تضيف حقائق»: مو بس لازم كل رقم/كود بالأصل يبقى —
// ولا رقم ولا كود حجز/مشروع ولا رمز «موظف#n» جديد يطلع، والطول ما يتجاوز ١.٦ ضعف.
func voiceSafe(orig, out string) bool {
	if strings.TrimSpace(out) == "" || float64(len([]rune(out))) > voiceMaxGrowth*float64(len([]rune(orig))) {
		return false
	}
	// المقارنة بالرموز الكاملة مو بالـsubstring: «3» ما ينحسب موجود لأن «B1234» بيه 3.
	origFacts, outFacts := voiceSet(voiceFacts(orig)), voiceSet(voiceFacts(out))
	for f := range origFacts {
		if !outFacts[f] {
			return false
		}
	}
	for f := range outFacts {
		if !origFacts[f] {
			return false
		}
	}
	origTags := voiceSet(voiceTagRe.FindAllString(orig, -1))
	for _, t := range voiceTagRe.FindAllString(out, -1) {
		if !origTags[t] {
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
		"شروط: غيّر الصياغة بس — نفس المعنى والطلب بالضبط. كل رقم واسم وكود وتاريخ ورمز «موظف#n» يبقى حرفياً. " +
		"ممنوع تضيف أي شي مو موجود بالتنبيه: لا رقم ولا اسم ولا كود ولا موعد ولا سبب ولا نتيجة. لا غرامة ولا تهديد. خليه بنفس الطول تقريباً. " +
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
