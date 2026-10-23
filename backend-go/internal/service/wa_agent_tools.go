package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ أدوات وكيلي الواتساب (ناروتو) — كل أداة تكتب AiAgentAction، وبالتجربة ما تنفّذ ═══

type waTurnCtx struct {
	svc     *WaAgentService
	conv    *repository.WaConversation
	booking *repository.WaBookingContext
	actions []repository.AiAgentAction
}

func (t *waTurnCtx) record(kind, summary, reason, targetType string, targetID *string, payload any) {
	tt := targetType
	cid := t.conv.ID
	a, err := t.svc.repo.AddAction(repository.AiAgentAction{Agent: t.conv.Agent, ConversationID: &cid, Kind: kind,
		Summary: summary, Reason: reason, TargetType: &tt, TargetID: targetID, Sandbox: t.conv.Sandbox}, payload)
	if err == nil {
		if d, err := t.svc.decorate(a, nil); err == nil {
			t.actions = append(t.actions, *d)
		}
	}
}

func intArg(in map[string]any, k string) *int {
	switch v := in[k].(type) {
	case float64:
		n := int(v)
		if n >= 1 && n <= 5 {
			return &n
		}
	case int:
		if v >= 1 && v <= 5 {
			return &v
		}
	}
	return nil
}

func (t *waTurnCtx) tools() []agentTool {
	b := t.booking
	sandboxNote := func(s string) string {
		if t.conv.Sandbox {
			return s + " (وضع تجربة — ما انحفظ بالنظام)"
		}
		return s
	}
	common := []agentTool{
		{Name: "get_context", Desc: "تفاصيل الحجز الحالي: الكود والنوع والحالة والخدمة والموعد وليدر الكادر.",
			Run: func(map[string]any, *nameMask) (string, error) { return t.svc.contextText(b), nil }},
		{Name: "handoff", Desc: "حوّل المحادثة لموظف إنسان ووكّف. استعملها إذا الزبون طلب إنسان، أو عصّب، أو الطلب برّا شغلك، أو ما متأكد.",
			Props: map[string]any{"reason": map[string]any{"type": "string", "description": "ليش تحوّلها"}}, Req: []string{"reason"},
			Run: func(in map[string]any, _ *nameMask) (string, error) {
				reason := strArg(in, "reason")
				t.svc.handoff(t.conv, reason)
				t.record("HANDOFF", "حوّل المحادثة لموظف: "+reason, reason, "CONVERSATION", &t.conv.ID, nil)
				return "انحوّلت لموظف. ودّع الزبون بجملة قصيرة وكله موظف راح يكمل وياه.", nil
			}},
	}
	if t.conv.Agent == WaAgentQuality {
		return append(common,
			agentTool{Name: "record_contact",
				Desc: "سجّل «تم التواصل» بمتابعة الجودة: outcome = OK إذا الزبون راضي بلا مشكلة، ISSUE إذا عنده مشكلة. والتقييمين من ١ لـ٥ إذا انطاهن. سجّلها مرة وحدة بآخر المحادثة، وقبل create_booking إذا اكو مشكلة.",
				Props: map[string]any{
					"outcome":         map[string]any{"type": "string", "enum": []string{"OK", "ISSUE"}},
					"customer_rating": map[string]any{"type": "integer", "description": "تقييم الزبون للشغل ١–٥"},
					"crew_rating":     map[string]any{"type": "integer", "description": "تقييم الزبون للكادر ١–٥"},
					"notes":           map[string]any{"type": "string", "description": "شنو كال الزبون بجملة أو جملتين"},
				}, Req: []string{"outcome", "notes"},
				Run: func(in map[string]any, _ *nameMask) (string, error) {
					if t.conv.FollowUpID == nil {
						return "", errors.New("ماكو متابعة جودة لهذا الحجز")
					}
					ok := strArg(in, "outcome") != "ISSUE"
					cr, kr, notes := intArg(in, "customer_rating"), intArg(in, "crew_rating"), strArg(in, "notes")
					if !t.conv.Sandbox {
						if err := t.svc.repo.RecordContactByAgent(*t.conv.FollowUpID, t.conv.Agent, ok, cr, kr, notes); err != nil {
							return "", err
						}
					}
					sum := "تم التواصل — الزبون راضي"
					if !ok {
						sum = "تم التواصل — الزبون عنده مشكلة"
					}
					t.record("RECORD_CONTACT", sum+": "+notes, notes, "FOLLOWUP", t.conv.FollowUpID,
						map[string]any{"ok": ok, "customerRating": cr, "crewRating": kr})
					return sandboxNote("انسجّل التواصل والتقييم."), nil
				}},
			agentTool{Name: "create_booking",
				Desc: "سوّي حجز جديد لنفس الزبون إذا عنده مشكلة — بعد ما يوافق الزبون صراحة. kind: MAINTENANCE صيانة لعطل واضح، COMPLAINT شكوى على الشغل أو الكادر، INSPECTION كشف إذا المشكلة مو واضحة وتحتاج فني يشوفها.",
				Props: map[string]any{
					"kind":    map[string]any{"type": "string", "enum": []string{"MAINTENANCE", "COMPLAINT", "INSPECTION"}},
					"details": map[string]any{"type": "string", "description": "المشكلة بكلام الزبون، وشنو يحتاج"},
				}, Req: []string{"kind", "details"},
				Run: func(in map[string]any, _ *nameMask) (string, error) {
					kind, details := strArg(in, "kind"), strArg(in, "details")
					label := map[string]string{"MAINTENANCE": "صيانة", "COMPLAINT": "شكوى", "INSPECTION": "كشف للمشكلة"}[kind]
					if label == "" {
						return "", errors.New("نوع الحجز غلط")
					}
					note := fmt.Sprintf("🤖 %s (%s) من متابعة حجز %s: %s", t.svc.name(t.conv.Agent), label, b.Code, details)
					var newID *string
					code := "تجربة"
					if !t.conv.Sandbox {
						bt := "MAINTENANCE"
						if kind == "COMPLAINT" {
							bt = "COMPLAINT"
						}
						nb, err := t.svc.bookings.Create(model.CreateBookingRequest{CustomerID: b.CustomerID, BookingType: &bt, Notes: &note})
						if err != nil {
							return "", err
						}
						newID, code = &nb.ID, nb.Code
						if t.conv.FollowUpID != nil {
							_, _ = t.svc.followups.LinkToBooking(*t.conv.FollowUpID, nb.ID)
						}
					}
					t.record("CREATE_BOOKING", fmt.Sprintf("سوّى حجز %s %s — %s", label, code, details), details, "BOOKING", newID,
						map[string]any{"kind": kind, "fromBooking": b.Code})
					return sandboxNote("انسوى حجز " + label + " " + code + ". كول للزبون إن الحجز انسجّل وراح يتواصلون وياه يحددون الموعد."), nil
				}},
			agentTool{Name: "suggest_to_matrix",
				Desc:  "قدّم اقتراح لماتركس (تحسين، نمط مشاكل، ملاحظة على الكادر أو الخدمة). مو للزبون — للإدارة.",
				Props: map[string]any{"title": map[string]any{"type": "string"}, "rationale": map[string]any{"type": "string"}}, Req: []string{"title", "rationale"},
				Run: func(in map[string]any, _ *nameMask) (string, error) {
					title, why := strArg(in, "title"), strArg(in, "rationale")
					if !t.conv.Sandbox {
						if _, err := t.svc.proposals.Create(model.MatrixProposal{Kind: model.ProposalAgentNote, Source: "AI_" + t.conv.Agent,
							Title: "💬 " + t.svc.name(t.conv.Agent) + ": " + title, Rationale: why,
							Evidence:  toJSON(map[string]any{"booking": b.Code, "conversationId": t.conv.ID}),
							Signature: "AGENT|" + t.conv.ID + "|" + title}); err != nil {
							return "", err
						}
					}
					t.record("SUGGEST_MATRIX", "اقتراح لماتركس: "+title, why, "PROPOSAL", nil, nil)
					return sandboxNote("وصل الاقتراح لماتركس."), nil
				}},
		)
	}
	return append(common,
		agentTool{Name: "add_booking_details",
			Desc:  "رحّل تفاصيل أخذتها من الزبون ويا الحجز (كادر يريده، نوعية، عنوان أوضح، وقت يناسبه…).",
			Props: map[string]any{"details": map[string]any{"type": "string"}}, Req: []string{"details"},
			Run: func(in map[string]any, _ *nameMask) (string, error) {
				d := strArg(in, "details")
				if !t.conv.Sandbox {
					if err := t.svc.repo.AppendBookingNote(b.BookingID, "🤖 "+t.svc.name(t.conv.Agent)+": "+d); err != nil {
						return "", err
					}
				}
				t.record("ADD_DETAILS", "رحّل تفاصيل ويا الحجز: "+d, d, "BOOKING", &b.BookingID, nil)
				return sandboxNote("انضافت التفاصيل للحجز."), nil
			}},
		agentTool{Name: "confirm_booking",
			Desc:  "ثبّت الحجز بموعد وافق عليه الزبون صراحة. at بتوقيت بغداد YYYY-MM-DDTHH:MM، ولازم بالمستقبل وبساعات الدوام (٩ص–٦م).",
			Props: map[string]any{"at": map[string]any{"type": "string"}, "notes": map[string]any{"type": "string"}}, Req: []string{"at"},
			Run: func(in map[string]any, _ *nameMask) (string, error) {
				at, err := time.ParseInLocation("2006-01-02T15:04", strArg(in, "at"), debriefLoc)
				if err != nil {
					return "", errors.New("صيغة الموعد غلط — YYYY-MM-DDTHH:MM")
				}
				if at.Before(time.Now()) {
					return "", errors.New("الموعد لازم بالمستقبل")
				}
				if h := at.Hour(); h < 9 || h >= 18 {
					return "", errors.New("الموعد برّا ساعات الدوام (٩ص–٦م)")
				}
				if !t.conv.Sandbox {
					if err := t.svc.repo.MarkContactedByAgent(b.BookingID); err != nil {
						return "", err
					}
					local := at.Format("2006-01-02T15:04")
					name := t.svc.name(t.conv.Agent)
					notes := "🤖 ثبّته " + name
					if n := strArg(in, "notes"); n != "" {
						notes += ": " + n
					}
					if _, err := t.svc.bookings.Confirm(b.BookingID, model.ConfirmBookingRequest{ConfirmedByName: &name, AdminNotes: &notes, ScheduledAt: &local}); err != nil {
						return "", err
					}
				}
				t.record("CONFIRM_BOOKING", "ثبّت الحجز "+b.Code+" على "+at.Format("2006-01-02 15:04"), strArg(in, "notes"), "BOOKING", &b.BookingID,
					map[string]any{"at": at.Format(time.RFC3339)})
				return sandboxNote("انثبّت الحجز. كول للزبون الموعد بالضبط."), nil
			}},
	)
}

// ── العقل الافتراضي فوق Claude (لوفي يطوّر السياسة بـwa_agent_prompts.go) ──

type claudeWaBrain struct {
	db      *sqlx.DB
	client  *anthropic.Client
	model   string
	feature string
}

// NewClaudeWaBrain عقل Claude لوكيل. client nil = بلا ذكاء (الخدمة تستعمل العقل المؤقت).
func NewClaudeWaBrain(db *sqlx.DB, client *anthropic.Client, model, agent string) WaBrain {
	if client == nil {
		return nil
	}
	return &claudeWaBrain{db: db, client: client, model: model, feature: agent + "_AGENT"}
}

func (c *claudeWaBrain) Reply(_ context.Context, t WaTurn) (string, error) {
	a := NewMatrixAgent(c.db, c.client, c.model)
	for _, tool := range t.Tools {
		a.AddTool(tool)
	}
	run, err := a.Chat(c.feature, WaSystemPrompt(t), t.History, t.Message)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(run.Answer), nil
}
