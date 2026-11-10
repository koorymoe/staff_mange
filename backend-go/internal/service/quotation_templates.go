package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ 📋 نماذج عروض الأسعار (ل) 11-10 ═══
// دازاي: «اريدك تحفظ هذا عرض السعر… يقرا ويدرسه ويشوفه شلون مصمم ويحفظه… وراها اكله
// سويلي واحد مثله لفلان مشروع وغيّر المنتجات والضوابط» — ويدوي بنفس الوقت.
// النموذج = شكل العرض (layout/الغلاف/الملحق/الشروط/المدة) + فصوله وبنوده. اسم الزبون
// ورقمه ما ينحفظن بالنموذج أبداً.

type QuotationTemplateService struct {
	db     *sqlx.DB
	quotes *QuotationService
}

func NewQuotationTemplateService(db *sqlx.DB) *QuotationTemplateService {
	return &QuotationTemplateService{db: db, quotes: NewQuotationService(repository.NewQuotationRepository(db), repository.NewPermissionRepository(db))}
}

type templateItem struct {
	Category    string  `json:"category"`
	ProductName string  `json:"productName"`
	ModelName   string  `json:"modelName,omitempty"`
	Specs       string  `json:"specs,omitempty"`
	Unit        string  `json:"unit,omitempty"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	Image       string  `json:"image,omitempty"`
}

type templateData struct {
	Layout   string           `json:"layout"`
	Cover    *json.RawMessage `json:"cover,omitempty"`
	Annex    *json.RawMessage `json:"annex,omitempty"`
	Terms    []string         `json:"terms,omitempty"`
	Duration string           `json:"duration,omitempty"`
	Notes    string           `json:"notes,omitempty"`
	Discount float64          `json:"discountPercent,omitempty"`
	Items    []templateItem   `json:"items"`
}

type QuotationTemplate struct {
	ID         string    `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	StyleNotes *string   `db:"styleNotes" json:"styleNotes"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
	By         *string   `db:"by" json:"by"`
	Data       []byte    `db:"data" json:"-"`
	Summary    string    `db:"-" json:"summary"`
}

// CanUse يگدر يحفظ/يستعمل نماذج؟ (نفس صلاحيات عروض الأسعار)
func (s *QuotationTemplateService) CanUse(emp, role string) bool {
	if role == "" {
		_ = s.db.Get(&role, `SELECT role::text FROM "Employee" WHERE id = $1`, emp)
	}
	if model.IsAdminRole(role) {
		return true
	}
	var ok bool
	_ = s.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId"
		WHERE ep."employeeId" = $1 AND p.name IN ($2, $3, $4, 'quotation_system'))`, emp, QuotationPermCreate, QuotationPermEditOwn, QuotationPermManageAll)
	return ok
}

// Save يحفظ عرض كنموذج. العرض لازم يكون مالته أو عنده «إدارة كل العروض» أو مدير.
// يرجع ملخص «دراسة» العرض (الشكل والفصول والبنود والشروط) — ماتركس يقراه.
func (s *QuotationTemplateService) Save(emp, role, quotationRef, name, notes string) (string, string, error) {
	if !s.CanUse(emp, role) {
		return "", "", ErrDesignHidden
	}
	var q struct {
		ID       string           `db:"id"`
		Creator  string           `db:"createdByEmployeeId"`
		Layout   string           `db:"layout"`
		Cover    *json.RawMessage `db:"cover"`
		Annex    *json.RawMessage `db:"annex"`
		Terms    *json.RawMessage `db:"terms"`
		Duration *string          `db:"duration"`
		Notes    *string          `db:"notes"`
		Discount float64          `db:"discountPercent"`
	}
	if err := s.db.Get(&q, `SELECT id, "createdByEmployeeId", layout, cover, annex, terms, duration, notes, "discountPercent"
		FROM "Quotation" WHERE id = $1 OR "quotationNumber" = $1 LIMIT 1`, strings.TrimSpace(quotationRef)); err != nil {
		return "", "", errors.New("العرض مو موجود")
	}
	if q.Creator != emp && !model.IsAdminRole(role) {
		var all bool
		_ = s.db.Get(&all, `SELECT EXISTS (SELECT 1 FROM "EmployeePermission" ep JOIN "Permission" p ON p.id = ep."permissionId"
			WHERE ep."employeeId" = $1 AND p.name = $2)`, emp, QuotationPermManageAll)
		if !all {
			return "", "", errors.New("العرض مو موجود")
		}
	}
	items := []struct {
		Cat   *string `db:"category"`
		Name  string  `db:"productName"`
		Model *string `db:"modelName"`
		Specs *string `db:"specs"`
		Unit  *string `db:"unit"`
		Qty   int     `db:"quantity"`
		Price float64 `db:"unitPrice"`
		Img   *string `db:"imageBase64"`
	}{}
	if err := s.db.Select(&items, `SELECT category, "productName", "modelName", specs, unit, quantity, "unitPrice", "imageBase64"
		FROM "QuotationItem" WHERE "quotationId" = $1 ORDER BY "sortIndex"`, q.ID); err != nil {
		return "", "", err
	}
	str := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}
	d := templateData{Layout: q.Layout, Cover: q.Cover, Annex: q.Annex, Duration: str(q.Duration), Notes: str(q.Notes), Discount: q.Discount}
	if q.Terms != nil {
		_ = json.Unmarshal(*q.Terms, &d.Terms)
	}
	for _, it := range items {
		img := str(it.Img)
		if strings.HasPrefix(img, "data:") {
			img = "" // صور قديمة مضمّنة — ما تنحفظ بالنموذج
		}
		d.Items = append(d.Items, templateItem{Category: strings.TrimSpace(str(it.Cat)), ProductName: it.Name, ModelName: str(it.Model),
			Specs: str(it.Specs), Unit: str(it.Unit), Quantity: it.Qty, UnitPrice: it.Price, Image: img})
	}
	name = strings.TrimSpace(name)
	if len([]rune(name)) < 3 {
		return "", "", errors.New("اكتب اسم للنموذج (٣ حروف على الأقل)")
	}
	raw, _ := json.Marshal(d)
	var id string
	if err := s.db.Get(&id, `INSERT INTO "QuotationTemplate" (name, "sourceQuotationId", data, "styleNotes", "createdById")
		VALUES ($1, $2, $3, NULLIF(btrim($4), ''), NULLIF($5, '')) RETURNING id`, name, q.ID, raw, notes, emp); err != nil {
		return "", "", err
	}
	return id, summarize(d), nil
}

// summarize «الدراسة»: شكل العرض بكلمات — بلا أسماء زبائن.
func summarize(d templateData) string {
	chapters := map[string]int{}
	order := []string{}
	for _, it := range d.Items {
		c := it.Category
		if c == "" {
			c = "التجهيزات"
		}
		if chapters[c] == 0 {
			order = append(order, c)
		}
		chapters[c]++
	}
	parts := []string{}
	for _, c := range order {
		parts = append(parts, fmt.Sprintf("%s (%d بند)", c, chapters[c]))
	}
	layout := "كلاسيكي"
	if d.Layout == "PROJECT" {
		layout = "عرض مشاريع (غلاف + فصول + ملحق فني)"
	}
	return fmt.Sprintf("الشكل: %s · الفصول: %s · الشروط: %d · المدة: %s · الخصم: %.0f%%",
		layout, strings.Join(parts, "، "), len(d.Terms), orDash(d.Duration), d.Discount)
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func (s *QuotationTemplateService) List() ([]QuotationTemplate, error) {
	out := []QuotationTemplate{}
	err := s.db.Select(&out, `SELECT t.id, t.name, t."styleNotes", t."createdAt", e.name AS by, t.data
		FROM "QuotationTemplate" t LEFT JOIN "Employee" e ON e.id = t."createdById"
		WHERE t."archivedAt" IS NULL ORDER BY t."createdAt" DESC LIMIT 200`)
	for i := range out {
		var d templateData
		_ = json.Unmarshal(out[i].Data, &d)
		out[i].Summary = summarize(d)
	}
	return out, err
}

// TemplateItemChange تغيير على بنود النموذج وقت الاستعمال.
type TemplateItemChange struct {
	Category    string   `json:"category"`
	ProductName string   `json:"productName"`
	ModelName   string   `json:"modelName"`
	Specs       string   `json:"specs"`
	Unit        string   `json:"unit"`
	Quantity    *int     `json:"quantity"`
	UnitPrice   *float64 `json:"unitPrice"`
	Remove      bool     `json:"remove"`
}

// TemplateUse عرض جديد من نموذج: الزبون والمشروع، وتغييرات اختيارية على البنود والشروط.
type TemplateUse struct {
	CustomerName  string               `json:"customerName"`
	CustomerPhone string               `json:"customerPhone"`
	ProjectName   string               `json:"projectName"`
	ProjectID     string               `json:"projectId"`
	ReplaceItems  bool                 `json:"replaceItems"` // true = البنود الجديدة بدل بنود النموذج
	Items         []TemplateItemChange `json:"items"`        // false = تعديل/إضافة على بنود النموذج (بالاسم)
	Terms         []string             `json:"terms"`        // فاضي = شروط النموذج
	Duration      string               `json:"duration"`
}

func (s *QuotationTemplateService) Use(emp, role, templateRef string, in TemplateUse) (*model.Quotation, error) {
	if !s.CanUse(emp, role) {
		return nil, ErrDesignHidden
	}
	var raw []byte
	if err := s.db.Get(&raw, `SELECT data FROM "QuotationTemplate" WHERE (id = $1 OR name = $1) AND "archivedAt" IS NULL
		ORDER BY "createdAt" DESC LIMIT 1`, strings.TrimSpace(templateRef)); err != nil {
		return nil, errors.New("النموذج مو موجود")
	}
	var d templateData
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.CustomerName) == "" {
		return nil, errors.New("اكتب اسم الزبون")
	}
	items := d.Items
	if in.ReplaceItems {
		items = nil
	}
	for _, ch := range in.Items {
		idx := -1
		for i, it := range items {
			if strings.EqualFold(strings.TrimSpace(it.ProductName), strings.TrimSpace(ch.ProductName)) {
				idx = i
				break
			}
		}
		if ch.Remove {
			if idx >= 0 {
				items = append(items[:idx], items[idx+1:]...)
			}
			continue
		}
		if idx < 0 {
			items = append(items, templateItem{ProductName: ch.ProductName, Quantity: 1})
			idx = len(items) - 1
		}
		it := &items[idx]
		set := func(dst *string, v string) {
			if strings.TrimSpace(v) != "" {
				*dst = strings.TrimSpace(v)
			}
		}
		set(&it.Category, ch.Category)
		set(&it.ModelName, ch.ModelName)
		set(&it.Specs, ch.Specs)
		set(&it.Unit, ch.Unit)
		if ch.Quantity != nil {
			it.Quantity = *ch.Quantity
		}
		if ch.UnitPrice != nil {
			it.UnitPrice = *ch.UnitPrice
		}
	}
	if len(items) == 0 {
		return nil, errors.New("العرض بلا بنود")
	}
	p := func(v string) *string {
		v = strings.TrimSpace(v)
		if v == "" {
			return nil
		}
		return &v
	}
	req := model.CreateQuotationRequest{CustomerName: strings.TrimSpace(in.CustomerName), CustomerPhone: p(in.CustomerPhone),
		ProjectName: p(in.ProjectName), ProjectID: p(in.ProjectID), Layout: p(d.Layout), Cover: d.Cover, Annex: d.Annex,
		Notes: p(d.Notes), Duration: p(d.Duration), Terms: d.Terms, DiscountPercent: &d.Discount}
	if len(in.Terms) > 0 {
		req.Terms = in.Terms
	}
	if in.Duration != "" {
		req.Duration = p(in.Duration)
	}
	for _, it := range items {
		req.Items = append(req.Items, model.QuotationItemInput{ProductName: it.ProductName, Unit: p(it.Unit), Quantity: it.Quantity,
			UnitPrice: it.UnitPrice, ModelName: p(it.ModelName), Specs: p(it.Specs), Category: p(it.Category), ImageBase64: p(it.Image)})
	}
	return s.quotes.Create(req, emp, role)
}

// templateNames أسماء النماذج (لماتركس).
func (s *QuotationTemplateService) templateNames() string {
	rows, _ := s.List()
	names := []string{}
	for _, r := range rows {
		names = append(names, "«"+r.Name+"» — "+r.Summary)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return "ماكو نماذج محفوظة بعد."
	}
	return strings.Join(names, "\n")
}

// ═══ أدوات ماتركس ═══
// «اريدك تحفظ هذا عرض السعر… سويلي واحد مثله لفلان مشروع وغيّر المنتجات والضوابط».
// ⚠️ ولا أداة ترجّع اسم زبون أو رقمه للذكاء الخارجي — رقم العرض بس.

func quotationTemplateTools(db *sqlx.DB, userID, role string) []agentTool {
	s := NewQuotationTemplateService(db)
	if role == "" {
		_ = db.Get(&role, `SELECT role::text FROM "Employee" WHERE id = $1`, userID)
	}
	if !s.CanUse(userID, role) {
		return nil
	}
	str := func(in map[string]any, k string) string { v, _ := in[k].(string); return strings.TrimSpace(v) }
	return []agentTool{{
		Name: "quotation_template_save",
		Desc: "يحفظ عرض سعر موجود كنموذج (شكله وغلافه وفصوله وبنوده وشروطه — بلا بيانات الزبون) ويرجّع دراسة شكله. استعملها لما يگلك «احفظ هذا العرض/ادرسه».",
		Props: map[string]any{
			"quotation_number": map[string]any{"type": "string", "description": "رقم العرض (مثلاً Q-2026-0042)"},
			"name":             map[string]any{"type": "string", "description": "اسم قصير للنموذج"},
			"style_notes":      map[string]any{"type": "string", "description": "ملاحظاتك عن تصميمه (اختياري)"},
		},
		Req: []string{"quotation_number", "name"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			_, sum, err := s.Save(userID, role, str(in, "quotation_number"), str(in, "name"), str(in, "style_notes"))
			if err != nil {
				return "", err
			}
			return "انحفظ النموذج «" + str(in, "name") + "». دراسته: " + sum, nil
		},
	}, {
		Name: "quotation_template_list",
		Desc: "قائمة نماذج عروض الأسعار المحفوظة ويا دراسة كل واحد.",
		Run:  func(map[string]any, *nameMask) (string, error) { return s.templateNames(), nil },
	}, {
		Name: "quotation_from_template",
		Desc: "يسوي عرض سعر جديد (مسودة) من نموذج محفوظ لمشروع/زبون ثاني، ويا تغييرات البنود والشروط. أكّد ويا المستخدم الاسم والتغييرات قبل. يرجّع رقم العرض الجديد — والمستخدم يكمّل التعديل يدوي من الشاشة.",
		Props: map[string]any{
			"template":      map[string]any{"type": "string", "description": "اسم النموذج"},
			"customer_name": map[string]any{"type": "string"},
			"project_name":  map[string]any{"type": "string"},
			"replace_items": map[string]any{"type": "boolean", "description": "true = البنود الجديدة بدل بنود النموذج"},
			"items": map[string]any{"type": "array", "description": "بنود تنضاف أو تتعدّل (بالاسم) أو تنشال (remove)", "items": map[string]any{"type": "object", "properties": map[string]any{
				"category": map[string]any{"type": "string"}, "productName": map[string]any{"type": "string"}, "modelName": map[string]any{"type": "string"},
				"specs": map[string]any{"type": "string"}, "unit": map[string]any{"type": "string"}, "quantity": map[string]any{"type": "integer"},
				"unitPrice": map[string]any{"type": "number"}, "remove": map[string]any{"type": "boolean"}}}},
			"terms":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "الشروط الجديدة (فاضي = شروط النموذج)"},
			"duration": map[string]any{"type": "string"},
		},
		Req: []string{"template", "customer_name"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			use := TemplateUse{CustomerName: str(in, "customer_name"), ProjectName: str(in, "project_name"), Duration: str(in, "duration")}
			use.ReplaceItems, _ = in["replace_items"].(bool)
			if raw, err := json.Marshal(in["items"]); err == nil {
				_ = json.Unmarshal(raw, &use.Items)
			}
			if raw, err := json.Marshal(in["terms"]); err == nil {
				_ = json.Unmarshal(raw, &use.Terms)
			}
			q, err := s.Use(userID, role, str(in, "template"), use)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("انسوى العرض %s (%d بند، الصافي %.0f). يفتحه من «عروض الأسعار» ويعدّله يدوي إذا يريد.", q.QuotationNumber, len(use.Items), q.NetTotal), nil
		},
	}}
}
