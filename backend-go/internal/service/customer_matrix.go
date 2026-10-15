package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/jmoiron/sqlx"
)

// ═══ ماتركس تطبيق الزبائن — قرار (ع) 10-14 ═══
//
// ماتركس هو الي يرد على الزبون بالتطبيق: يجاوب عن خدماتنا، يستفسر،
// ويستلم طلب الحجز ويرحّله للحجوزات بنفسه (الحجز ينكتب «أنشأه: ماتركس»).
//
// ⚠️ حدود صارمة: أدواته تشوف بس خدمات الشركة والعروض الفعّالة وحجوزات
// **هذا الزبون**. ما يشوف زبون ثاني، ولا موظفين، ولا أسعار داخلية.

const customerMatrixFeature = "customer_app_chat"

// CustomerBookingDraft طلب الحجز الي يجمعه ماتركس من الزبون.
type CustomerBookingDraft struct {
	ServiceIDs    []string
	Address       string
	LocationURL   string
	PreferredTime string
	Notes         string
	DeviceCount   int
}

// CustomerMatrix وكيل المحادثة ويّا الزبون.
type CustomerMatrix struct {
	db    *sqlx.DB
	llm   agentLLM
	model string
	// CreateBooking ينفّذ الحجز (يربطه main بخدمة الحجوزات) — يرجّع كود الحجز.
	CreateBooking func(accountID string, d CustomerBookingDraft) (string, error)
}

func NewCustomerMatrix(db *sqlx.DB, client *anthropic.Client, model string) *CustomerMatrix {
	m := &CustomerMatrix{db: db, model: model}
	if client != nil {
		m.llm = anthropicLLM{c: client}
	}
	return m
}

func (m *CustomerMatrix) Enabled() bool { return m != nil && m.llm != nil }

// ErrCustomerMatrixOff ماكو مفتاح للنموذج.
var ErrCustomerMatrixOff = errors.New("ماتركس مو متوفر هسه")

func customerMatrixSystem(name string) string {
	return `أنت «ماتركس»، المساعد الذكي لشركة الأماني بالعراق، وتحچي ويّا زبون بتطبيق الشركة.
اسم الزبون: ` + name + `.
- احچي بلهجة عراقية مهذبة وودودة، وجمل قصيرة وواضحة. استخدم إيموجي قليلة.
- عرّف الزبون بخدماتنا (استخدم أداة list_services) والعروض الحالية (active_offers) وروّج إلها بلطف بدون مبالغة.
- إذا سأل عن أنواع الأجهزة أو الخيارات ضمن خدمة، استخدم catalog_products واذكر فقط المنتجات والموديلات والمواصفات الموجودة فعلاً بالكتالوج.
- إذا يريد حجز: افهم شنو الخدمة، واسأله عن العنوان (المحافظة والمنطقة وأقرب نقطة دالة)، والوقت المناسب، وأي تفاصيل (مثلاً عدد السيارات أو الأجهزة لخدمات الجي بي اس). اسأل سؤال أو سؤالين بالمرة مو أكثر.
- قبل ما تسجّل الحجز اعرض عليه ملخص قصير واطلب تأكيده. بعد ما يأكد استخدم create_booking، وبلّغه بكود الحجز وإن فريقنا راح يتواصل وياه لتثبيت الموعد.
- لا تعطي أسعار نهائية أبداً — الأسعار يحددها الفريق بعد الكشف أو التواصل. ولا تعطي وعود بمواعيد مؤكدة.
- رقم هاتف الزبون محفوظ بحسابه؛ لا تطلب رقمه أبداً ولا تطلب إعادة كتابته.
- إذا ما كدرت تحل طلبه، أو يحتاج تدخل بشري، لا تقل "دز رقمك". اقترح بلطف فقط: «تحب أحوّلك لأحد موظفينا؟» وإذا وافق، أكّد أن فريق الشركة سيتابع الطلب من حسابه.
- عن حجوزاته السابقة استخدم my_bookings.
- لا تذكر أسماء موظفين، ولا أي معلومة داخلية عن الشركة، ولا تجاوب على شي ما يخص خدمات الشركة (اعتذر بلطف).
- إذا الزبون زعلان أو عنده شكوى، اعتذر واهتم، واطلب التفاصيل وكلّه فريقنا راح يتابع.
التاريخ اليوم: ` + time.Now().Format("2006-01-02") + `.`
}

func (m *CustomerMatrix) tools(accountID string) []agentTool {
	return []agentTool{
		{Name: "list_services", Desc: "قائمة خدمات شركة الأماني (المعرّف، الاسم، الصنف).", Run: func(map[string]any, *nameMask) (string, error) {
			rows := []struct {
				ID       string  `db:"id" json:"id"`
				Name     string  `db:"name" json:"name"`
				Category *string `db:"category" json:"category,omitempty"`
			}{}
			if err := m.db.Select(&rows, `SELECT id, name, category FROM "Service" WHERE category IS DISTINCT FROM 'إدارية' ORDER BY category NULLS LAST, name`); err != nil {
				return "", err
			}
			b, _ := json.Marshal(rows)
			return string(b), nil
		}},
		{Name: "catalog_products", Desc: "أسماء وموديلات ومواصفات منتجات كتالوج شركة الأماني؛ استخدمه لاقتراح الخيارات المناسبة للزبون. لا يضم أسعاراً.", Run: func(map[string]any, *nameMask) (string, error) {
			rows := []struct {
				Name string `db:"name" json:"name"`
				ModelName *string `db:"modelName" json:"modelName,omitempty"`
				Specs *string `db:"specs" json:"specs,omitempty"`
				ServiceName *string `db:"serviceName" json:"serviceName,omitempty"`
			}{}
			if err := m.db.Select(&rows, `SELECT p.name, p."modelName", p.specs, s.name AS "serviceName" FROM "Product" p LEFT JOIN "Service" s ON s.id = p."serviceId" ORDER BY s.name NULLS LAST, p.name LIMIT 300`); err != nil {
				return "", err
			}
			b, _ := json.Marshal(rows)
			return string(b), nil
		}},
		{Name: "active_offers", Desc: "العروض والخدمات المميزة المنشورة حالياً بالتطبيق.", Run: func(map[string]any, *nameMask) (string, error) {
			rows := []struct {
				Kind  string  `db:"kind" json:"kind"`
				Title string  `db:"title" json:"title"`
				Body  *string `db:"body" json:"body,omitempty"`
			}{}
			if err := m.db.Select(&rows, `SELECT kind, title, body FROM "PromoItem" WHERE active ORDER BY ord, "createdAt" DESC LIMIT 20`); err != nil {
				return "", err
			}
			b, _ := json.Marshal(rows)
			return string(b), nil
		}},
		{Name: "my_bookings", Desc: "حجوزات هذا الزبون (الحالية والسابقة) بحالتها.", Run: func(map[string]any, *nameMask) (string, error) {
			rows, err := CustomerBookings(m.db, accountID, 15)
			if err != nil {
				return "", err
			}
			b, _ := json.Marshal(rows)
			return string(b), nil
		}},
		{Name: "create_booking", Desc: "يسجّل طلب حجز جديد لهذا الزبون بعد ما يأكد الملخص. يرجّع كود الحجز.",
			Props: map[string]any{
				"service_ids":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "معرّفات الخدمات من list_services"},
				"address":        map[string]any{"type": "string", "description": "المحافظة والمنطقة وأقرب نقطة دالة"},
				"location_url":   map[string]any{"type": "string", "description": "رابط الموقع إذا انطاه الزبون"},
				"preferred_time": map[string]any{"type": "string", "description": "الوقت المناسب للزبون"},
				"notes":          map[string]any{"type": "string", "description": "تفاصيل الطلب"},
				"device_count":   map[string]any{"type": "integer", "description": "عدد السيارات/الأجهزة لخدمات الجي بي اس"},
			},
			Req: []string{"service_ids", "address"},
			Run: func(in map[string]any, _ *nameMask) (string, error) {
				if m.CreateBooking == nil {
					return "", errors.New("التسجيل غير متاح")
				}
				d := CustomerBookingDraft{
					Address: cmStr(in["address"]), LocationURL: cmStr(in["location_url"]),
					PreferredTime: cmStr(in["preferred_time"]), Notes: cmStr(in["notes"]),
				}
				if arr, ok := in["service_ids"].([]any); ok {
					for _, x := range arr {
						if s := cmStr(x); s != "" {
							d.ServiceIDs = append(d.ServiceIDs, s)
						}
					}
				}
				if n, ok := in["device_count"].(float64); ok {
					d.DeviceCount = int(n)
				}
				code, err := m.CreateBooking(accountID, d)
				if err != nil {
					return "", err
				}
				return `{"ok":true,"bookingCode":"` + code + `"}`, nil
			}},
	}
}

func cmStr(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

// Chat دورة محادثة وحدة. يرجّع الرد، وكود الحجز إذا انسجّل حجز بهالدورة.
func (m *CustomerMatrix) Chat(accountID, name string, history []AgentTurn, message string) (string, string, error) {
	if !m.Enabled() {
		return "", "", ErrCustomerMatrixOff
	}
	tools := m.tools(accountID)
	byName := map[string]agentTool{}
	params := make([]anthropic.ToolUnionParam, 0, len(tools))
	for _, t := range tools {
		byName[t.Name] = t
		props := t.Props
		if props == nil {
			props = map[string]any{}
		}
		params = append(params, anthropic.ToolUnionParam{OfTool: &anthropic.ToolParam{
			Name: t.Name, Description: anthropic.String(t.Desc),
			InputSchema: anthropic.ToolInputSchemaParam{Properties: props, Required: t.Req},
		}})
	}
	msgs := []anthropic.MessageParam{}
	for _, t := range history {
		if strings.TrimSpace(t.Text) == "" {
			continue
		}
		if t.Role == "ASSISTANT" {
			msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(t.Text)))
		} else {
			msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(t.Text)))
		}
	}
	msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(message)))
	ctx, cancel := context.WithTimeout(context.Background(), agentTimeout)
	defer cancel()
	booked := ""
	for step := 0; step < 8; step++ {
		resp, err := m.llm.New(ctx, anthropic.MessageNewParams{
			Model: anthropic.Model(m.model), MaxTokens: 900,
			System:   []anthropic.TextBlockParam{{Text: customerMatrixSystem(name)}},
			Messages: msgs, Tools: params,
		})
		if err != nil {
			return "", booked, err
		}
		RecordAIUsage(customerMatrixFeature, m.model, resp.Usage)
		msgs = append(msgs, resp.ToParam())
		var results []anthropic.ContentBlockParamUnion
		var text strings.Builder
		for _, block := range resp.Content {
			switch b := block.AsAny().(type) {
			case anthropic.TextBlock:
				text.WriteString(b.Text)
			case anthropic.ToolUseBlock:
				t, ok := byName[b.Name]
				out, isErr := "أداة غير معروفة", true
				if ok {
					var in map[string]any
					_ = json.Unmarshal(b.Input, &in)
					res, err := t.Run(in, nil)
					if err != nil {
						out, isErr = "خطأ: "+err.Error(), true
					} else {
						out, isErr = res, false
						if b.Name == "create_booking" {
							var r struct {
								BookingCode string `json:"bookingCode"`
							}
							_ = json.Unmarshal([]byte(res), &r)
							booked = r.BookingCode
						}
					}
				}
				results = append(results, anthropic.NewToolResultBlock(b.ID, out, isErr))
			}
		}
		if len(results) == 0 || resp.StopReason != anthropic.StopReasonToolUse {
			return strings.TrimSpace(text.String()), booked, nil
		}
		msgs = append(msgs, anthropic.NewUserMessage(results...))
	}
	return "", booked, fmt.Errorf("ماتركس ما لحك يجاوب")
}

// CustomerBookingRow حجز كما يشوفه الزبون — بلا أسماء موظفين.
type CustomerBookingRow struct {
	ID          string     `db:"id" json:"id"`
	Code        string     `db:"code" json:"code"`
	Status      string     `db:"status" json:"status"`
	StatusLabel string     `db:"-" json:"statusLabel"`
	Services    string     `db:"services" json:"services"`
	ServiceID   *string    `db:"serviceId" json:"serviceId"`
	Address     *string    `db:"address" json:"address,omitempty"`
	ScheduledAt *time.Time `db:"scheduledAt" json:"scheduledAt,omitempty"`
	CompletedAt *time.Time `db:"completedAt" json:"completedAt,omitempty"`
	CreatedAt   time.Time  `db:"createdAt" json:"createdAt"`
	FromApp     bool       `db:"fromApp" json:"fromApp"`
}

var customerStatusAr = map[string]string{
	"PENDING": "استلمنا طلبك — بانتظار التثبيت", "CONFIRMED": "تثبّت الموعد", "IN_PROGRESS": "فريقنا يشتغل عليه",
	"WAITING": "بالانتظار", "PARTIAL": "انجز جزء منه", "COMPLETED": "منجز ✓", "CANCELLED": "ملغي",
}

// CustomerBookings حجوزات الزبون صاحب الحساب: بزبونه المربوط أو المسجّلة من حسابه.
func CustomerBookings(db *sqlx.DB, accountID string, limit int) ([]CustomerBookingRow, error) {
	rows := []CustomerBookingRow{}
	err := db.Select(&rows, `
		SELECT b.id, b.code, b.status, b.address, b."serviceId", b."scheduledAt", b."completedAt", b."createdAt",
			(b.source = 'matrix_app') IS TRUE AS "fromApp",
			COALESCE((SELECT string_agg(s.name, '، ') FROM "BookingService" bs JOIN "Service" s ON s.id = bs."serviceId" WHERE bs."bookingId" = b.id),
				(SELECT s.name FROM "Service" s WHERE s.id = b."serviceId"), '') AS services
		FROM "Booking" b
		WHERE (b."customerAccountId" = $1
		   OR b."customerId" = (SELECT "customerId" FROM "CustomerAccount" WHERE id = $1))
		  AND b."archivedAt" IS NULL -- المحذوف ما يطلع للزبون
		ORDER BY b."createdAt" DESC LIMIT $2`, accountID, limit)
	for i := range rows {
		rows[i].StatusLabel = customerStatusAr[rows[i].Status]
		if rows[i].StatusLabel == "" {
			rows[i].StatusLabel = rows[i].Status
		}
	}
	return rows, err
}

// SuggestPromo ماتركس يكتب مسودة عرض ترويجي من خدمات الشركة (المالك يوافق قبل النشر).
func (m *CustomerMatrix) SuggestPromo() (string, string, error) {
	if !m.Enabled() {
		return "", "", ErrCustomerMatrixOff
	}
	var names []string
	_ = m.db.Select(&names, `SELECT name FROM "Service" WHERE category IS DISTINCT FROM 'إدارية' ORDER BY random() LIMIT 12`)
	var existing []string
	_ = m.db.Select(&existing, `SELECT title FROM "PromoItem" ORDER BY "createdAt" DESC LIMIT 15`)
	ctx, cancel := context.WithTimeout(context.Background(), agentTimeout)
	defer cancel()
	resp, err := m.llm.New(ctx, anthropic.MessageNewParams{
		Model: anthropic.Model(m.model), MaxTokens: 400,
		System: []anthropic.TextBlockParam{{Text: "أنت كاتب إعلانات لشركة الأماني بالعراق. اكتب عرض ترويجي قصير وجذاب لتطبيق الزبائن بلهجة عراقية مهذبة. بلا أسعار وبلا وعود ما نكدر نلتزم بيها. رجّع JSON فقط: {\"title\": \"عنوان قصير\", \"body\": \"سطرين أو ثلاثة\"}"}},
		Messages: []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(
			"خدماتنا: " + strings.Join(names, "، ") + "\nعناوين عروض سابقة (لا تكررها): " + strings.Join(existing, " | ")))},
	})
	if err != nil {
		return "", "", err
	}
	RecordAIUsage(customerMatrixFeature, m.model, resp.Usage)
	var text strings.Builder
	for _, b := range resp.Content {
		if t, ok := b.AsAny().(anthropic.TextBlock); ok {
			text.WriteString(t.Text)
		}
	}
	raw := text.String()
	if i, j := strings.Index(raw, "{"), strings.LastIndex(raw, "}"); i >= 0 && j > i {
		raw = raw[i : j+1]
	}
	var out struct{ Title, Body string }
	if err := json.Unmarshal([]byte(raw), &out); err != nil || out.Title == "" {
		return "", "", fmt.Errorf("رد غير مفهوم")
	}
	return out.Title, out.Body, nil
}
