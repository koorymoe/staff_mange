package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/whatsapp"
)

// ═══ 📥 الاستقبال: الزبون هو الي يتواصل (دازاي 11-15) ═══
// «اريد الزبون من يتصل يردون عليه وكلاء الذكاء الاصطناعي، سواء رسالة او رسالة صوتية او مكالمة».
//
//   - رسالة من رقم ما عنده محادثة مفتوحة ← تنفتح محادثة استقبال والوكيل يرد (چانت تنهمل).
//   - فويس ← ينسحب من Meta ويتحوّل لنص **داخل سيرفرنا** (Whisper — قرار 10-05: الفويس ما يطلع)،
//     والوكيل يجاوب عليه مثل الرسالة.
//   - مكالمة واتساب فايتة ← الوكيل يراسل الزبون فوراً ويكمل وياه كتابة أو فويس.
//     الرد الصوتي الحي على المكالمة نفسها مرحلة ثانية (يحتاج مزوّد صوت لحظي — قرار دازاي).
//
// الوكيل يسوي الحجز، يقترح وقت من جدول الكوادر الحقيقي، ويثبّته. وتكليف الكادر بعدها شغل
// ماتركس الموجود (matrix_auto_crew ببوابة الدقة) — نفس السلسلة، بلا مسار ثاني.
//
// الحدود نفسها: بلا أسعار ولا فلوس ولا حذف، كل فعل بـAiAgentAction وينلغى، والحجز بعد «اي» صريحة.
// اسم الزبون ما يروح للنموذج إلا بعد قرار الخصوصية (PrivacyApproved).

const waReceptionMaxMsgs = 60

// WaTranscriber يحوّل فويس الزبون لنص.
type WaTranscriber func(ctx context.Context, audio []byte, contentType string) (string, error)

// waSlots الأوقات الفاضية من محرك اقتراحات ماتركس.
type waSlots interface {
	FreeSlots(bookingID string, max int) []SuggestedSlot
	SlotFree(bookingID string, at time.Time) bool
}

func (s *WaAgentService) SetCustomers(c *repository.CustomerRepository) { s.customers = c }
func (s *WaAgentService) SetSlots(sl *MatrixSuggestService) {
	if sl != nil {
		s.slots = sl
	}
}
func (s *WaAgentService) SetTranscriber(t WaTranscriber) { s.hear = t }

// VoiceOn الفويس ينسمع؟ (للحالة بشاشة المالك)
func (s *WaAgentService) VoiceOn() bool { return s.hear != nil }

const (
	waVoiceMark     = "🎤 "
	waVoiceFailText = "🎤 [فويس ما انسمع]"
	waVoiceFailBack = "عذراً، ما كدرت أسمع الفويس زين 🙏 تكدر تعيده بصوت أوضح أو تكتبلي؟"
)

// Inbound وارد حقيقي من Meta: رسالة أو فويس أو مكالمة.
func (s *WaAgentService) Inbound(ctx context.Context, m whatsapp.Inbound) {
	if m.Call != nil {
		s.inboundCall(ctx, m)
		return
	}
	conv := s.inboundConv(m.From)
	if conv == nil {
		return
	}
	id := m.MetaID
	if m.Audio != nil {
		text, ok := s.transcribe(ctx, m.Audio)
		if !ok {
			s.sayFixed(ctx, conv, waVoiceFailText, &id, waVoiceFailBack)
			return
		}
		m.Text = waVoiceMark + text
	}
	if _, _, err := s.Say(ctx, conv.ID, m.Text, &id); err != nil {
		log.Printf("[wa] inbound: %v", err)
	}
}

// inboundConv محادثة الرقم المفتوحة، أو محادثة استقبال جديدة إذا الاستقبال مشغّل.
func (s *WaAgentService) inboundConv(phone string) *repository.WaConversation {
	if conv, err := s.repo.OpenByPhone(phone); err == nil {
		return conv
	}
	if !s.Enabled(WaAgentReception) {
		log.Printf("[wa] رسالة من رقم بلا محادثة مفتوحة — الاستقبال مطفي، تنتظر موظف")
		return nil
	}
	var cust *string
	if s.customers != nil {
		if c, err := s.customers.FindByPhone(model.LocalPhone(phone)); err == nil && c != nil {
			cust = &c.ID
		}
	}
	conv, err := s.repo.CreateInbound(WaAgentBooking, cust, phone, false)
	if err != nil {
		log.Printf("[wa] فتح محادثة استقبال: %v", err)
		return nil
	}
	return conv
}

func (s *WaAgentService) transcribe(ctx context.Context, a *whatsapp.Audio) (string, bool) {
	if s.hear == nil {
		return "", false
	}
	data, mime, err := s.wa.Media(ctx, a.MediaID)
	if err != nil {
		log.Printf("[wa] سحب الفويس: %v", err)
		return "", false
	}
	if mime == "" {
		mime = a.MimeType
	}
	text, err := s.hear(ctx, data, mime)
	if text = strings.TrimSpace(text); err != nil || text == "" {
		if err != nil {
			log.Printf("[wa] تحويل الفويس: %v", err)
		}
		return "", false
	}
	return text, true
}

// sayFixed يحفظ وارد الزبون ويرد بنص ثابت بلا نموذج (فويس ما انسمع، مكالمة فايتة).
// يرجّع false إذا الوارد مكرر أو الرد ما انرسل.
func (s *WaAgentService) sayFixed(ctx context.Context, conv *repository.WaConversation, inText string, metaID *string, reply string) bool {
	if _, inserted, err := s.repo.AddMessage(conv.ID, "IN", "CUSTOMER", nil, inText, metaID); err != nil || !inserted {
		return false
	}
	if conv.Status != "BOT" || (!conv.Sandbox && !s.Enabled(convAgent(conv))) {
		return false // إنسان ماسكها أو الوكيل مطفي — الوارد انحفظ والموظف يشوفه
	}
	if !conv.Sandbox {
		if _, err := s.wa.SendText(ctx, conv.Phone, reply); err != nil {
			log.Printf("[wa] send: %v", err)
			return false
		}
	}
	_, _, _ = s.repo.AddMessage(conv.ID, "OUT", "AGENT", nil, reply, nil)
	return true
}

// inboundCall مكالمة واتساب من الزبون خلصت. الفايتة: الوكيل يراسله فوراً. وإذا الرسالة ما راحت
// (Meta ترفض خارج نافذة ٢٤ ساعة)، ينبَّه الموظفين يرجعون يتصلون.
func (s *WaAgentService) inboundCall(ctx context.Context, m whatsapp.Inbound) {
	conv := s.inboundConv(m.From)
	if conv == nil {
		return
	}
	id := m.MetaID
	if m.Call.Answered {
		mins := (m.Call.Duration + 59) / 60
		_, _, _ = s.repo.AddMessage(conv.ID, "IN", "CUSTOMER", nil, fmt.Sprintf("📞 مكالمة واتساب انردّ عليها (%d د)", mins), &id)
		return
	}
	reply := fmt.Sprintf("هلا بيك 🌹 شفنا اتصالك على شركة الأماني. وياك %s، مساعد ذكي — اكتبلي أو دزلي فويس بشنو تحتاج وأخدمك هسه.",
		s.name(WaAgentReception))
	if s.sayFixed(ctx, conv, "📞 مكالمة واتساب فايتة", &id, reply) {
		return
	}
	if conv.Status == "BOT" && s.notif != nil {
		_ = s.notif.CreateForRolesOrPermission([]string{"OWNER"}, "whatsapp_chats", "wa_handoff",
			"📞 زبون اتصل عالواتساب وما أحد رد، ورسالة الوكيل ما وصلته — ارجعوا اتصلوا بيه من «محادثات الواتساب».")
	}
}

// ── سياق الدور ──

// turnContext الي يشوفه العقل: سياق الحجز إذا المحادثة مربوطة، وحجوزات الزبون المفتوحة بالاستقبال.
func (s *WaAgentService) turnContext(conv *repository.WaConversation, c *repository.WaBookingContext) string {
	if !conv.Inbound {
		return s.contextText(c)
	}
	var b strings.Builder
	if c != nil {
		b.WriteString("الحجز المحدد بهالمحادثة — " + s.contextText(c))
	} else {
		b.WriteString("المحادثة بعدها بلا حجز محدد.")
	}
	switch {
	case conv.CustomerID == nil:
		b.WriteString("\nالزبون: رقم جديد، ما مسجّل عدنا.")
	default:
		rows, err := s.repo.OpenBookings(*conv.CustomerID)
		if err != nil || len(rows) == 0 {
			b.WriteString("\nالزبون: مسجّل عدنا، وماكو إله حجز مفتوح.")
			break
		}
		b.WriteString("\nالزبون: مسجّل عدنا. حجوزاته المفتوحة:")
		for _, r := range rows {
			b.WriteString("\n- " + waBookingLine(r))
		}
	}
	if conv.CustomerID != nil {
		if _, code := s.repo.PendingFollowUp(*conv.CustomerID); code != "" {
			b.WriteString("\nمتابعة جودة معلّقة: خلّصنا إله الحجز " + code + " قبل أيام وبعدنا ما أخذنا رأيه بالشغل.")
		}
	}
	return b.String()
}

func waBookingLine(r repository.WaOpenBooking) string {
	line := r.Code + " · " + waStatusAr(r.Status)
	if r.Services != "" {
		line += " · " + r.Services
	}
	if r.ScheduledAt != nil {
		line += " · الموعد " + secFmt(r.ScheduledAt)
	} else {
		line += " · بلا موعد بعد"
	}
	return line
}

func waStatusAr(st string) string {
	switch st {
	case "PENDING":
		return "بانتظار التثبيت"
	case "CONFIRMED":
		return "مثبّت"
	case "IN_PROGRESS":
		return "قيد التنفيذ"
	case "WAITING":
		return "متوقف مؤقتاً"
	case "PARTIAL":
		return "منجز جزئياً"
	}
	return st
}

// ── أدوات الاستقبال ──

// sandboxBooking بالتجربة الحجز ما ينخلق فعلاً — إذا الوكيل «سوّاه» قبل، نكمل على حجز وهمي.
func (t *waTurnCtx) sandboxBooking() *repository.WaBookingContext {
	rows, err := t.svc.repo.Actions("", t.conv.ID)
	if err != nil {
		return nil
	}
	for _, a := range rows {
		if a.Kind == "CREATE_BOOKING" && a.UndoneAt == nil {
			return &repository.WaBookingContext{Code: "تجربة", Status: "PENDING", Type: "REGULAR"}
		}
	}
	return nil
}

func (t *waTurnCtx) receptionTools() []agentTool {
	if t.booking == nil && t.conv.Sandbox {
		t.booking = t.sandboxBooking()
	}
	sandboxNote := func(s string) string {
		if t.conv.Sandbox {
			return s + " (وضع تجربة — ما انحفظ بالنظام)"
		}
		return s
	}
	tools := []agentTool{
		{Name: "handoff", Desc: "حوّل المحادثة لموظف إنسان ووكّف. استعملها إذا الزبون طلب إنسان، أو عصّب، أو يسأل عن سعر أو فاتورة أو دفع، أو الطلب برّا شغلك، أو ما متأكد.",
			Props: map[string]any{"reason": map[string]any{"type": "string", "description": "ليش تحوّلها"}}, Req: []string{"reason"},
			Run: func(in map[string]any, _ *nameMask) (string, error) {
				reason := strArg(in, "reason")
				t.svc.handoff(t.conv, reason)
				t.record("HANDOFF", "حوّل المحادثة لموظف: "+reason, reason, "CONVERSATION", &t.conv.ID, nil)
				return "انحوّلت لموظف. ودّع الزبون بجملة قصيرة وكله موظف راح يكمل وياه.", nil
			}},
		{Name: "services", Desc: "قائمة خدمات الشركة بأسمائها بالنظام. استعملها حتى تطابق طلب الزبون ويا خدمة قبل create_booking.",
			Run: func(map[string]any, *nameMask) (string, error) {
				rows, err := t.svc.repo.Services()
				if err != nil {
					return "", err
				}
				names := make([]string, 0, len(rows))
				for _, r := range rows {
					names = append(names, r.Name)
				}
				return "الخدمات: " + strings.Join(names, " · "), nil
			}},
		{Name: "my_bookings", Desc: "حجوزات هذا الزبون المفتوحة (الكود، الحالة، الخدمة، الموعد). استعملها إذا يسأل عن حجزه أو موعده.",
			Run: func(map[string]any, *nameMask) (string, error) {
				if t.conv.CustomerID == nil {
					return "رقم جديد — ماكو إله حجوزات عدنا.", nil
				}
				rows, err := t.svc.repo.OpenBookings(*t.conv.CustomerID)
				if err != nil {
					return "", err
				}
				if len(rows) == 0 {
					return "ماكو إله حجز مفتوح.", nil
				}
				lines := make([]string, 0, len(rows))
				for _, r := range rows {
					lines = append(lines, waBookingLine(r))
				}
				return strings.Join(lines, "\n"), nil
			}},
		{Name: "select_booking", Desc: "اختار واحد من حجوزات الزبون المفتوحة (بكوده) حتى تكمل عليه: تضيف تفاصيل، تقترح وقت، أو تثبّت.",
			Props: map[string]any{"code": map[string]any{"type": "string", "description": "كود الحجز مثل B123"}}, Req: []string{"code"},
			Run: func(in map[string]any, _ *nameMask) (string, error) {
				if t.conv.CustomerID == nil {
					return "", errors.New("رقم جديد — ماكو إله حجوزات")
				}
				code := strings.ToUpper(strings.TrimSpace(strArg(in, "code")))
				rows, err := t.svc.repo.OpenBookings(*t.conv.CustomerID)
				if err != nil {
					return "", err
				}
				for _, r := range rows {
					if strings.ToUpper(r.Code) != code {
						continue
					}
					bc, err := t.svc.repo.BookingContext(r.ID)
					if err != nil {
						return "", err
					}
					if err := t.svc.repo.Bind(t.conv.ID, nil, &r.ID); err != nil {
						return "", err
					}
					t.conv.BookingID, t.booking = &r.ID, bc
					return "انحدد الحجز: " + t.svc.contextText(bc), nil
				}
				return "", errors.New("هالكود مو من حجوزات هذا الزبون المفتوحة")
			}},
		{Name: "create_booking",
			Desc: "سوّي حجز جديد لهذا الزبون — بعد ما يوافق صراحة على ملخص كتبته إله. kind: NEW طلب شغل جديد (لازم service باسمه من أداة services)، MAINTENANCE صيانة لشغل سابق، COMPLAINT شكوى. للرقم الجديد customerName إجباري (اسأله اسمه).",
			Props: map[string]any{
				"kind":         map[string]any{"type": "string", "enum": []string{"NEW", "MAINTENANCE", "COMPLAINT"}},
				"service":      map[string]any{"type": "string", "description": "اسم الخدمة مثل ما بأداة services"},
				"details":      map[string]any{"type": "string", "description": "شنو يريد الزبون بكلامه"},
				"address":      map[string]any{"type": "string", "description": "المنطقة أو العنوان"},
				"customerName": map[string]any{"type": "string", "description": "اسم الزبون (للرقم الجديد)"},
				"deviceCount":  map[string]any{"type": "integer", "description": "عدد الأجهزة — بس للخدمات الي تطلبه (جي بي اس)"},
				"vehicleType":  map[string]any{"type": "string", "description": "نوع المركبة — بس للخدمات الي تطلبه"},
			}, Req: []string{"kind", "details"},
			Run: func(in map[string]any, _ *nameMask) (string, error) {
				return t.receptionCreate(in, sandboxNote)
			}},
	}
	// الجودة من أي تواصل (دازاي 11-16): الزبون الي عنده متابعة معلّقة وراسلنا أو اتصل — ناخذ رأيه هنا،
	// بلا ما ننتظر وكيل الجودة يراسله. نفس اسم الأداة حتى تنطبق حراسة التقييم (١–٥ من الزبون نفسه).
	if t.conv.CustomerID != nil && !t.conv.Sandbox {
		if fu, code := t.svc.repo.PendingFollowUp(*t.conv.CustomerID); fu != "" {
			tools = append(tools, agentTool{Name: "record_contact",
				Desc: "سجّل رأي الزبون بالشغل الأخير (الحجز " + code + ") بمتابعة الجودة: outcome = OK راضي، ISSUE عنده مشكلة. والتقييمين من ١ لـ٥ مثل ما گالهن الزبون. مرة وحدة.",
				Props: map[string]any{
					"outcome":        map[string]any{"type": "string", "enum": []string{"OK", "ISSUE"}},
					"customerRating": map[string]any{"type": "integer", "description": "تقييم الزبون للشغل ١–٥"},
					"crewRating":     map[string]any{"type": "integer", "description": "تقييم الزبون للكادر ١–٥"},
					"notes":          map[string]any{"type": "string", "description": "شنو گال بجملة أو جملتين"},
				}, Req: []string{"outcome", "notes"},
				Run: func(in map[string]any, _ *nameMask) (string, error) {
					ok := strArg(in, "outcome") != "ISSUE"
					cr, kr, notes := intArg(in, "customerRating"), intArg(in, "crewRating"), strArg(in, "notes")
					if err := t.svc.repo.RecordContactByAgent(fu, t.conv.Agent, ok, cr, kr, notes); err != nil {
						return "", err
					}
					sum := "أخذ رأي الزبون بحجز " + code + " — راضي"
					if !ok {
						sum = "أخذ رأي الزبون بحجز " + code + " — عنده مشكلة"
					}
					t.record("RECORD_CONTACT", sum+": "+notes, notes, "FOLLOWUP", &fu, map[string]any{"ok": ok, "customerRating": cr, "crewRating": kr})
					return "انسجّل رأيه. إذا عنده مشكلة، اعرض عليه حجز صيانة أو شكوى (create_booking) بعد موافقته.", nil
				}})
		}
	}
	return append(tools, t.bookingTools()...)
}

func (t *waTurnCtx) receptionCreate(in map[string]any, sandboxNote func(string) string) (string, error) {
	kind, details := strArg(in, "kind"), strings.TrimSpace(strArg(in, "details"))
	label := map[string]string{"NEW": "طلب جديد", "MAINTENANCE": "صيانة", "COMPLAINT": "شكوى"}[kind]
	if label == "" {
		return "", errors.New("نوع الحجز غلط")
	}
	if details == "" {
		return "", errors.New("اكتب شنو يريد الزبون بـdetails")
	}
	if t.booking != nil {
		return "", errors.New("هالمحادثة بيها حجز " + t.booking.Code + " أصلاً — كمّل عليه، أو حوّل لموظف إذا يريد حجز ثاني")
	}
	var serviceID *string
	if name := strings.TrimSpace(strArg(in, "service")); name != "" {
		rows, err := t.svc.repo.Services()
		if err != nil {
			return "", err
		}
		id := matchService(rows, name)
		if id == "" {
			return "", errors.New("الخدمة «" + name + "» مو موجودة — جيب الأسماء بأداة services واختار الأقرب")
		}
		serviceID = &id
	} else if kind == "NEW" {
		return "", errors.New("الطلب الجديد يحتاج service — جيب الأسماء بأداة services")
	}
	name := t.svc.name(WaAgentReception)
	note := fmt.Sprintf("🤖 %s (%s من واتساب الزبون): %s", name, label, details)
	req := model.CreateBookingRequest{ServiceID: serviceID, Notes: &note}
	if a := strings.TrimSpace(strArg(in, "address")); a != "" {
		req.Address = &a
	}
	if v := strings.TrimSpace(strArg(in, "vehicleType")); v != "" {
		req.VehicleType = &v
	}
	if n, ok := in["deviceCount"].(float64); ok && n >= 1 {
		c := int(n)
		req.DeviceCount = &c
	}
	if kind != "NEW" {
		bt := kind
		req.BookingType = &bt
	}
	code := "تجربة"
	var newID *string
	if t.conv.Sandbox {
		t.booking = &repository.WaBookingContext{Code: code, Status: "PENDING", Type: "REGULAR"}
	} else {
		custID, err := t.receptionCustomer(strings.TrimSpace(strArg(in, "customerName")))
		if err != nil {
			return "", err
		}
		req.CustomerID = custID
		if dup := t.svc.bookings.RecentDuplicate(req); dup != "" {
			return "", errors.New("اكو حجز مشابه انسوى تواً لنفس الزبون — لا تكرره، شوفه بـmy_bookings وكمّل عليه بـselect_booking")
		}
		nb, err := t.svc.bookings.Create(req)
		if err != nil {
			return "", err
		}
		bc, err := t.svc.repo.BookingContext(nb.ID)
		if err != nil {
			return "", err
		}
		if err := t.svc.repo.Bind(t.conv.ID, &custID, &nb.ID); err != nil {
			return "", err
		}
		newID, code = &nb.ID, nb.Code
		t.conv.CustomerID, t.conv.BookingID, t.booking = &custID, &nb.ID, bc
	}
	t.record("CREATE_BOOKING", fmt.Sprintf("سوّى حجز %s %s — %s", label, code, details), details, "BOOKING", newID,
		map[string]any{"kind": kind, "channel": "RECEPTION"})
	return sandboxNote("انسوى الحجز " + code + ". هسه اقترح على الزبون وقت بـfree_slots، وبعد موافقته ثبّته بـconfirm_booking."), nil
}

// receptionCustomer زبون المحادثة: المسجّل، أو ينفتحله سجل جديد برقمه واسمه.
func (t *waTurnCtx) receptionCustomer(name string) (string, error) {
	if t.conv.CustomerID != nil {
		return *t.conv.CustomerID, nil
	}
	if t.svc.customers == nil {
		return "", errors.New("سجل الزبائن مو متوفر — حوّل لموظف")
	}
	phone := model.LocalPhone(t.conv.Phone)
	if c, err := t.svc.customers.FindByPhone(phone); err == nil && c != nil {
		return c.ID, nil
	}
	if name == "" {
		return "", errors.New("رقم جديد — اسأل الزبون عن اسمه وعيد النداء بـcustomerName")
	}
	c, err := t.svc.customers.FindOrCreateByPhone(phone, name)
	if err != nil {
		return "", err
	}
	return c.ID, nil
}

// matchService يطابق اسم الخدمة: حرفي أولاً، وبعدين احتواء بطرف واحد إذا المطابق وحيد.
func matchService(rows []repository.WaService, name string) string {
	norm := func(s string) string { return strings.ToLower(strings.Join(strings.Fields(s), " ")) }
	want := norm(name)
	found := ""
	for _, r := range rows {
		n := norm(r.Name)
		if n == want {
			return r.ID
		}
		if strings.Contains(n, want) || strings.Contains(want, n) {
			if found != "" {
				return "" // أكثر من خدمة — غامض
			}
			found = r.ID
		}
	}
	return found
}
