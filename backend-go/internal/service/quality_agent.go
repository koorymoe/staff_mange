package service

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

// ═══ ⭐ «ابو الجودة» — متابعة الزبون بعد الشغل عالواتساب (دازاي 10-23) ═══
// يشتغل بس على QualityFollowUp بحالة PENDING لحجوزات انجزت آخر ٧ أيام (ناروتو يختار).
// الترتيب: خلص زين؟ ← تقييم الشغل ← تقييم الكادر ← اكو مشكلة؟ ← (حجز إذا لازم) ← تسجيل.

// QualityAgent عقل ابو الجودة.
type QualityAgent struct{ *waAgent }

// NewQualityAgent toolsFor = أدوات ناروتو (get_followup، record_contact، create_booking،
// suggest_to_matrix، handoff). client = nil ← كل رسالة تتحول لإنسان.
func NewQualityAgent(client *anthropic.Client, model string, toolsFor func(*WaTurnCtx) []agentTool, dailyCap int) *QualityAgent {
	a := newWaAgent("QUALITY_AGENT", client, model, toolsFor, dailyCap)
	a.system = qualitySystem
	return &QualityAgent{a}
}

// QualityOpening أول رسالة (قالب Meta الموافق عليه) — نص ثابت مو من النموذج.
// ⚠️ سؤال مفتوح لدازاي: «مهندس الجودة» لو «المساعد الذكي لقسم الجودة»؟ لحد القرار
// نعرّف إنه مساعد ذكي (قاعدة الخطة: أول رسالة تعرّف إنه مساعد ذكي).
func QualityOpening(agentName string) string {
	return fmt.Sprintf("مرحبا زبوننا الكريم 🌹 وياك %s، المساعد الذكي لقسم الجودة بشركة الأماني. حبينا نطمّن عليك: الشغل الي سويناه إلك خلص زين؟", waAgentName(agentName))
}

const qualityPolicy = `انت «%s» — المساعد الذكي لقسم الجودة بشركة الأماني (طاقة شمسية، كهرباء، شبكات، كاميرات).
شغلك: تطمّن على زبون خلصنا شغله قريباً، وتاخذ تقييمه، وإذا عنده مشكلة ترتب حلها.
الرسالة الأولى (التعريف + «الشغل خلص زين؟») انرسلت. كمّل بهالترتيب، وكل خطوة سؤال واحد:
١) هل الشغل خلص زين؟ (انسألت)
٢) تقييمك للشغل من ٥؟
٣) تقييمك للكادر (الفنيين) من ٥؟
٤) اكو أي مشكلة أو ملاحظة؟
إذا الزبون جاوب على شي قبل وقته، لا تعيد السؤال — كمّل من وين وصل.
إذا اكو مشكلة:
- افهمها بسؤال أو سؤالين منطقيين (شنو بالضبط؟ من شوكت؟ يشتغل لو عاطل؟).
- قرر نوع الحجز: MAINTENANCE (عطل يحتاج تصليح) · COMPLAINT (زعل من الشغل أو التعامل) · INSPECTION (ما واضح، يحتاج كشف).
- اكتب للزبون ملخص بسطر («راح نسويلك حجز صيانة للعاكسة الي تطفي — نثبّت؟») واستنى «اي». بعدها create_booking.
بالنهاية (بعد ما ياخذ التقييمين ويعرف إذا اكو مشكلة): record_contact بالتقييمين ونوع البلاغ والملاحظات، واشكره بسطر.
إذا لكيت شي يستاهل المدير يعرفه (نمط، شكوى على كادر، فكرة تحسين): suggest_to_matrix.
` + waCommonRules

func qualitySystem(t WaAgentTurn) string {
	return fmt.Sprintf(qualityPolicy, waAgentName(t.AgentName)) + waContext(t)
}

// Reply رد ابو الجودة — دائماً يرجّع نتيجة (الفشل = تحويل لإنسان).
func (q *QualityAgent) Reply(ctx context.Context, t WaAgentTurn) *WaAgentResult {
	return q.waAgent.Reply(ctx, t)
}
