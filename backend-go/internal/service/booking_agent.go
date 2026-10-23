package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// ═══ 📅 «ابو الحجوزات» — تثبيت الحجز الجديد ويا الزبون عالواتساب (دازاي 10-23) ═══
// يجي حجز جديد بلا تواصل: يرحّب، يأكد نوع الشغل والعنوان، ياخذ التفضيلات (كادر،
// نوعية) ويرحّلها ويا الحجز، يقترح وقت من الفاضي، ويثبّت بعد موافقة الزبون.

// BookingAgent عقل ابو الحجوزات.
type BookingAgent struct{ *waAgent }

// NewBookingAgent toolsFor = أدوات ناروتو (get_booking، add_booking_details،
// confirm_booking، handoff).
func NewBookingAgent(client *anthropic.Client, model string, toolsFor func(*WaTurnCtx) []agentTool, dailyCap int) *BookingAgent {
	a := newWaAgent("BOOKING_AGENT", client, model, toolsFor, dailyCap)
	a.system = bookingSystem
	return &BookingAgent{a}
}

// BookingOpening أول رسالة (قالب Meta). service = نوع الشغل من الحجز (ممكن فاضي).
func BookingOpening(agentName, service string) string {
	s := "مرحبا بيك 🌹 وياك " + waAgentName(agentName) + "، المساعد الذكي لحجوزات شركة الأماني. وصلنا طلبك"
	if service = strings.TrimSpace(service); service != "" {
		s += " لـ" + service
	}
	return s + "، وحاب أكمّل وياك التفاصيل حتى نثبّته. هذا الطلب صحيح؟"
}

const bookingPolicy = `انت «%s» — المساعد الذكي لحجوزات شركة الأماني (طاقة شمسية، كهرباء، شبكات، كاميرات).
شغلك: تكمّل تفاصيل حجز جديد ويا الزبون وتثبّته. الرسالة الأولى (الترحيب + «هذا الطلب صحيح؟») انرسلت.
كمّل بهالترتيب، وكل خطوة سؤال واحد:
١) يأكد نوع الشغل (من معلومات الحجز). إذا يريد يغيّره، سجّل التغيير بـadd_booking_details.
٢) يأكد العنوان أو المنطقة إذا ناقص أو مو واضح.
٣) اكو تفضيل؟ (كادر معيّن، نوعية أو ماركة معيّنة، ملاحظة للفنيين) — سجّله بـadd_booking_details. إذا ماكو، كمّل.
٤) اقترح وقت من الأوقات الفاضية بمعلومات الحجز (لا تخترع وقت). إذا ماكو أوقات بالمعلومات، حوّل لإنسان.
٥) اكتب ملخص بسطر (الشغل، المنطقة، اليوم والساعة) واسأل «نثبّته؟». بعد «اي» صريحة: confirm_booking بالوقت.
بعد التثبيت: اشكره بسطر وكله الفريق راح يتواصل وياه قبل الموعد.
إذا جاوب على شي قبل وقته، لا تعيد السؤال.
لا تذكر سعر أبداً — إذا سأل عن السعر، كله الفريق يبلّغه بعد الكشف، أو حوّل لإنسان إذا أصرّ.
` + waCommonRules

func bookingSystem(t WaAgentTurn) string {
	return fmt.Sprintf(bookingPolicy, waAgentName(t.AgentName)) + waContext(t)
}

// Reply رد ابو الحجوزات — دائماً يرجّع نتيجة (الفشل = تحويل لإنسان).
func (b *BookingAgent) Reply(ctx context.Context, t WaAgentTurn) *WaAgentResult {
	return b.waAgent.Reply(ctx, t)
}
