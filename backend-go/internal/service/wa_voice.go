package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ 📞 المكالمة الحية: الوكيل يحچي ويا الزبون بصوته وبالعراقي (دازاي 11-15) ═══
// «ابني الصوت الحي وخلي يحچي عراقي».
//
// القسمة:
//   - **منصة صوت** (ElevenLabs Agents أو أي منصة تدعم «Custom LLM») ترد على مكالمة الواتساب،
//     تسمع الزبون وتحوّل كلامه لنص، وتنطق الرد بصوت عراقي. هي الأذن واللسان بس.
//   - **العقل عدنا**: المنصة تنادي /api/voice/v1/chat/completions بكل جملة يگولها الزبون،
//     ونفس عقل الاستقبال يجاوب — بنفس الأدوات (حجز، أوقات فاضية، تثبيت) ونفس الحراسات
//     («اي» صريحة قبل الحجز، تحويل لإنسان، كل فعل بـAiAgentAction). عقل واحد للكتابة والصوت.
//
// المكالمة تنكتب بنفس محادثة الزبون (📞 قبل كل جملة) — الموظف يشوف الرسائل والمكالمات بخط واحد.
// مكالمة بلا رقم (تجربة من لوحة المنصة) = وضع تجربة: ولا شي ينكتب على النظام.
//
// 🔴 صوت الزبون يروح لمنصة الصوت (خارج السيرفر) — استثناء من قرار 10-05، فالمفتاح مطفي
// افتراضياً والمالك يشغّله بنفسه. النص الي يروح للنموذج نفس الكتابة: بلا اسم الزبون.

var ErrVoiceOff = errors.New("المكالمة الحية مو مربوطة بعد — حط VOICE_LLM_SECRET واربط منصة الصوت")

const (
	waCallMark = "📞 "
	// يطلعن بصوت — بلا إيموجي.
	voiceGreeting   = "هلا بيك، وياك مساعد شركة الأماني، مساعد ذكي. شلون أكدر أخدمك؟"
	voiceOffReply   = "هلا بيك. هسه ما نكدر نرد على المكالمة، دزلنا رسالة عالواتساب ونخدمك بأقرب وقت. مع السلامة."
	voiceHumanReply = "تدلل، زميلنا بالشركة راح يتصل بيك بأقرب وقت. مع السلامة."
	voiceErrorReply = "عذراً، صار عندي خلل. زميلنا بالشركة راح يتصل بيك بأقرب وقت. مع السلامة."
	voiceAgainReply = "عذراً، ما سمعتك زين. تكدر تعيد؟"
)

// waVoiceRules تنضاف لسياسة أي وكيل لمّا الدور مكالمة حية.
const waVoiceRules = `

قواعد المكالمة الصوتية (هذا الدور مكالمة تلفون — ردّك ينگال بصوت للزبون):
- كلام الزبون (الي يبدي بـ📞) انكتب نص تلقائياً من صوته، وممكن بيه أغلاط سمع. إذا شي مو واضح اطلب منه يعيده.
- جاوب بجملة أو جملتين قصار، بعراقي محچي طبيعي مثل موظف استقبال عراقي عالتلفون. بلا فصحى وبلا كلام رسمي.
- ممنوع إيموجي أو رموز أو قوائم أو أقواس — بس كلام ينگال.
- الأرقام والأوقات والتواريخ اكتبها بالحروف مثل ما تنگال («باچر الساعة عشرة الصبح»، «يوم الأحد»)، مو أرقام.
- لا تقرا كود الحجز إلا إذا الزبون طلبه.
- اقترح وقتين بالأكثر بالمرة.
- لا تعرّف بنفسك من جديد — الترحيب انگال أول المكالمة.`

// VoiceReply نتيجة دور بالمكالمة: شنو ينگال، وهل تنسد المكالمة بعده.
type VoiceReply struct {
	Text string
	End  bool
}

// VoiceTurn جملة گالها الزبون بالمكالمة. callID معرّف المكالمة عند المنصة، phone رقم المتصل
// (فاضي = تجربة)، n تسلسل الجملة، greeting الترحيب الي انگال أول المكالمة (ينحفظ مرة).
func (s *WaAgentService) VoiceTurn(ctx context.Context, callID, phone string, n int, text, greeting string) VoiceReply {
	text = strings.TrimSpace(text)
	if text == "" {
		return VoiceReply{Text: voiceAgainReply}
	}
	digits := waPhone(phone)
	sandbox := len(digits) < 8
	if !sandbox && !s.Enabled(WaAgentVoice) {
		return VoiceReply{Text: voiceOffReply, End: true}
	}
	conv, fresh, err := s.voiceConv(digits, callID, sandbox)
	if err != nil {
		log.Printf("[voice] محادثة المكالمة: %v", err)
		return VoiceReply{Text: voiceErrorReply, End: true}
	}
	if fresh && strings.TrimSpace(greeting) != "" {
		_, _, _ = s.repo.AddMessage(conv.ID, "OUT", "AGENT", nil, speakable(greeting), nil)
	}
	text = DefaultVoiceLexicon.Apply(text) // قاموس الأذن: كلمات الشغل الي تنسمع غلط
	var metaID *string
	if callID != "" {
		sum := sha256.Sum256([]byte(text))
		id := fmt.Sprintf("voice:%s:%d:%s", callID, n, hex.EncodeToString(sum[:4]))
		metaID = &id
	}
	if conv.Status != "BOT" {
		// موظف ماسك المحادثة — الجملة تنحفظ إله، والمكالمة تنسد بأدب.
		_, _, _ = s.repo.AddMessage(conv.ID, "IN", "CUSTOMER", nil, waCallMark+text, metaID)
		return VoiceReply{Text: voiceHumanReply, End: true}
	}
	out, _, err := s.say(ctx, conv.ID, waCallMark+text, metaID, true)
	if err != nil {
		log.Printf("[voice] %v", err)
		return VoiceReply{Text: voiceErrorReply, End: true}
	}
	reply := ""
	for _, m := range out {
		if m.Direction == "OUT" {
			reply = m.Text
		}
	}
	if reply == "" {
		// المنصة عادت نفس الجملة (انقطاع أو إعادة محاولة) — نرجّع آخر رد بلا ما نعيد التنفيذ.
		if reply = s.repo.LastOut(conv.ID); reply == "" {
			reply = voiceAgainReply
		}
	}
	after, err := s.repo.Conversation(conv.ID)
	return VoiceReply{Text: reply, End: err == nil && after.Status != "BOT"}
}

// voiceConv محادثة المكالمة: محادثة الزبون المفتوحة إذا موجودة (المكالمة تكمل على نفس الخط)،
// وإلا محادثة استقبال جديدة. التجربة: محادثة تجربة لكل مكالمة.
func (s *WaAgentService) voiceConv(digits, callID string, sandbox bool) (*repository.WaConversation, bool, error) {
	if sandbox {
		key := "voice-test:" + callID
		if c, err := s.repo.SandboxByPhone(key); err == nil {
			return c, false, nil
		}
		c, err := s.repo.CreateInbound(WaAgentBooking, nil, key, true)
		return c, true, err
	}
	if c, err := s.repo.OpenByPhone(digits); err == nil {
		return c, false, nil
	}
	var cust *string
	if s.customers != nil {
		if c, err := s.customers.FindByPhone(model.LocalPhone(digits)); err == nil && c != nil {
			cust = &c.ID
		}
	}
	c, err := s.repo.CreateInbound(WaAgentBooking, cust, digits, false)
	return c, true, err
}

// VoiceGreeting الترحيب الافتراضي إذا المنصة طلبت أول جملة منّا.
func (s *WaAgentService) VoiceGreeting() string {
	n := s.name(WaAgentVoice)
	if n == "" || n == "مساعد الحجوزات" {
		return voiceGreeting
	}
	return "هلا بيك، وياك " + n + " من شركة الأماني، مساعد ذكي. شلون أكدر أخدمك؟"
}

// speakable ينظّف النص حتى ينگال: يشيل الإيموجي ورموز التنسيق — محرك النطق يقراها حرفياً أو يتلعثم بيها.
func speakable(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case unicode.Is(unicode.So, r), unicode.Is(unicode.Sk, r), unicode.Is(unicode.Cf, r), unicode.Is(unicode.Mn, r) && r >= 0xFE00 && r <= 0xFE0F:
			continue
		case r == '*' || r == '#' || r == '_' || r == '`' || r == '•' || r == '«' || r == '»':
			continue
		case r == '\n' || r == '\t':
			b.WriteRune(' ')
		default:
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
