package service

import (
	"fmt"
	"strings"

	"staffmange-api/internal/model"
)

// ═══ 📝 مسودة عرض سعر من كمال (دازاي 10-28: «ليش ميكدر يسوي عرض سعر؟») ═══
// كمال يكتب العرض كامل (الزبون، البنود، الأسعار من جدول الأسعار)، وينحفظ بحالة
// NEW — يعني ما انرسل للزبون. المالك/المسؤول يراجعه بشاشة عروض الأسعار ويرسله.
// الأسعار تنكتب بالعرض مثل ما يحددها كمال، فالمراجعة البشرية هي الضمان.

func registerQuoteTools(a *MatrixAgent, quotes *QuotationService, tc *SecretaryTurnCtx) {
	if quotes == nil || tc == nil {
		return
	}
	a.AddTool(agentTool{Name: "price_lookup",
		Desc:  "دوّر على سعر منتج أو خدمة بجدول أسعار الشركة (المنتجات والأسعار القياسية). استعملها قبل draft_quotation حتى تحط أسعار حقيقية — لا تخمّن سعر.",
		Props: map[string]any{"query": map[string]any{"type": "string", "description": "اسم المنتج أو جزء منه (مثل: كاميرا 4 ميگا، بطارية، جي بي اس)"}},
		Req:   []string{"query"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			q := strings.TrimSpace(strArg(in, "query"))
			if q == "" {
				return "", fmt.Errorf("اكتب اسم المنتج")
			}
			rows := []struct {
				Name  string   `db:"name"`
				Unit  *string  `db:"unit"`
				Price *float64 `db:"price"`
				Src   string   `db:"src"`
			}{}
			err := a.db.Select(&rows, `
				SELECT name, unit, "defaultPrice" AS price, 'منتج' AS src FROM "Product" WHERE name ILIKE '%' || $1 || '%' OR "modelName" ILIKE '%' || $1 || '%'
				UNION ALL
				SELECT product AS name, NULL AS unit, price, 'سعر قياسي' AS src FROM "StandardPrice" WHERE product ILIKE '%' || $1 || '%'
				LIMIT 15`, q)
			if err != nil {
				return "", err
			}
			if len(rows) == 0 {
				return "ما لكيت سعر لـ«" + q + "» بجدول الأسعار. اسأل المالك عن السعر، لا تخمّن.", nil
			}
			var b strings.Builder
			for _, r := range rows {
				p := "بلا سعر"
				if r.Price != nil && *r.Price > 0 {
					p = money(*r.Price)
				}
				u := ""
				if r.Unit != nil && *r.Unit != "" {
					u = " / " + *r.Unit
				}
				fmt.Fprintf(&b, "- %s: %s%s (%s)\n", r.Name, p, u, r.Src)
			}
			return b.String(), nil
		}})

	a.AddTool(agentTool{Name: "draft_quotation",
		Desc: "سوّي مسودة عرض سعر (ما تنرسل للزبون — تنحفظ جديدة والمالك يراجعها ويرسلها). لازم اسم الزبون وبند واحد على الأقل بسعر حقيقي من price_lookup أو من المالك. إذا ناقص سعر أو كمية، اسأله أول.",
		Props: map[string]any{
			"customer_name":  map[string]any{"type": "string"},
			"customer_phone": map[string]any{"type": "string"},
			"project_name":   map[string]any{"type": "string", "description": "اسم المشروع أو وصف قصير"},
			"notes":          map[string]any{"type": "string"},
			"items": map[string]any{"type": "array", "items": map[string]any{"type": "object",
				"properties": map[string]any{
					"name":       map[string]any{"type": "string"},
					"unit":       map[string]any{"type": "string"},
					"quantity":   map[string]any{"type": "integer"},
					"unit_price": map[string]any{"type": "number"},
				}, "required": []string{"name", "quantity", "unit_price"}}},
		},
		Req: []string{"customer_name", "items"},
		Run: func(in map[string]any, names *nameMask) (string, error) {
			req := model.CreateQuotationRequest{CustomerName: strings.TrimSpace(strArg(in, "customer_name")), CreatedByEmployeeID: tc.UserID}
			if req.CustomerName == "" {
				return "", fmt.Errorf("اسم الزبون ناقص")
			}
			opt := func(k string) *string {
				if v := strings.TrimSpace(strArg(in, k)); v != "" {
					return &v
				}
				return nil
			}
			req.CustomerPhone, req.ProjectName = opt("customer_phone"), opt("project_name")
			note := "📝 مسودة من " + model.AgentName("IRIS") + " — راجعها قبل الإرسال."
			if n := opt("notes"); n != nil {
				note += "\n" + *n
			}
			req.Notes = &note
			raw, _ := in["items"].([]any)
			var total float64
			for _, it := range raw {
				m, _ := it.(map[string]any)
				name := strings.TrimSpace(fmt.Sprint(m["name"]))
				qty, _ := m["quantity"].(float64)
				price, _ := m["unit_price"].(float64)
				if name == "" || qty <= 0 || price < 0 {
					return "", fmt.Errorf("بند ناقص (الاسم أو الكمية أو السعر): %v", m)
				}
				item := model.QuotationItemInput{ProductName: name, Quantity: int(qty), UnitPrice: price}
				if u, ok := m["unit"].(string); ok && strings.TrimSpace(u) != "" {
					u = strings.TrimSpace(u)
					item.Unit = &u
				}
				req.Items = append(req.Items, item)
				total += qty * price
			}
			if len(req.Items) == 0 {
				return "", fmt.Errorf("ماكو بنود بالعرض")
			}
			q, err := quotes.Create(req)
			if err != nil {
				return "", err
			}
			link := "/quotations/" + q.ID + "/edit"
			tc.Cards = append(tc.Cards, SecretaryCard{Kind: "info", Title: "📝 مسودة عرض سعر " + q.QuotationNumber,
				Lines: []string{"الزبون: " + req.CustomerName, fmt.Sprintf("%d بند — المجموع %s", len(req.Items), money(total)), "ما انرسل للزبون — راجعه وأرسله من عروض الأسعار"},
				Link:  &link})
			return fmt.Sprintf("انحفظت مسودة العرض %s للزبون %s: %d بند، المجموع %s. ما انرسلت — كله يراجعها من البطاقة.", q.QuotationNumber, req.CustomerName, len(req.Items), money(total)), nil
		}})
}
