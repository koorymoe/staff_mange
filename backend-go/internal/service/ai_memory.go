package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

// ═══ 🧠 ذاكرة المحادثة (دازاي 10-23: «لازم نسوي ذاكرة للذكاء لان اذا طالت المحادثة راح يصرف اكثر») ═══
// النموذج ياخذ ملخص المحادثة + آخر MemoryKeep رسائل بس، فكلفة السؤال ثابتة مهما طالت.
// كلما تجمّعت MemoryThreshold رسالة برّا الملخص، نداء رخيص واحد يدمجهن بالملخص.

const (
	MemoryKeep      = 8
	MemoryThreshold = 12
	memoryMaxTokens = 600
)

const memoryPrompt = `انت ذاكرة محادثة. عندك ملخص سابق (ممكن فاضي) ورسائل جديدة من المحادثة.
اكتب ملخص واحد محدّث بالعربي، نقاط قصيرة، لحد ١٥ نقطة:
- الحقائق والأرقام والأسماء والتواريخ الي انذكرت.
- القرارات والطلبات والمهام الي انرسلت أو اتفقوا عليها، وشنو باقي معلّق.
- تفضيلات المستخدم إذا بانت.
لا تكتب مقدمة. لا تخترع شي مو موجود. احذف الكلام الي ما إله قيمة.`

// Remember يدمج رسائل جديدة بالملخص السابق. feature لعدّاد الكلفة.
func Remember(ctx context.Context, llm agentLLM, model, feature, prev string, turns []AgentTurn) (string, error) {
	if llm == nil {
		return "", errors.New("بلا ذكاء")
	}
	var b strings.Builder
	if strings.TrimSpace(prev) != "" {
		b.WriteString("الملخص السابق:\n" + prev + "\n\n")
	}
	b.WriteString("الرسائل الجديدة:\n")
	for _, t := range turns {
		who := "المستخدم"
		if t.Role == "ASSISTANT" || t.Role == "assistant" {
			who = "المساعد"
		}
		text := t.Text
		if r := []rune(text); len(r) > 1500 {
			text = string(r[:1500]) + "…"
		}
		b.WriteString(who + ": " + text + "\n")
	}
	ctx, cancel := context.WithTimeout(ctx, 40*time.Second)
	defer cancel()
	resp, err := llm.New(ctx, anthropic.MessageNewParams{
		Model: anthropic.Model(model), MaxTokens: memoryMaxTokens,
		System:   []anthropic.TextBlockParam{{Text: memoryPrompt}},
		Messages: []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(b.String()))},
	})
	if err != nil {
		return "", err
	}
	RecordAIUsage(feature, model, resp.Usage)
	var out strings.Builder
	for _, blk := range resp.Content {
		if t, ok := blk.AsAny().(anthropic.TextBlock); ok {
			out.WriteString(t.Text)
		}
	}
	s := strings.TrimSpace(out.String())
	if s == "" {
		return "", errors.New("ملخص فاضي")
	}
	return s, nil
}
