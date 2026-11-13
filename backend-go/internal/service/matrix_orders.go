package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ 🎯 أوامر ماتركس (دازاي 11-13) ═══
// «مهام ماتركس اكدر اوجه لماتركس مهمه — اكله روح سويلي عرض سعر، اكله روح اخصم
// هذا الموظف فلان نقطه». المالك/المدير يكتب الأمر بكلامه، ماتركس يفهمه ويقترح
// إجراءات بالضبط (منو، شكد، ليش) — وما يتنفّذ شي إلا من يضغط «نفّذ».
// الأسماء ما تطلع للذكاء الاصطناعي: الموظفين برموز موظف#n، والزبائن برموز زبون#n.

const (
	OrderProposed  = "PROPOSED"
	OrderRunning   = "RUNNING"
	OrderDone      = "DONE"
	OrderCancelled = "CANCELLED"
	OrderFailed    = "FAILED"
)

// OrderAction إجراء واحد مقترح. الحقول حسب النوع.
type OrderAction struct {
	Kind         string      `json:"kind"` // points | task | notify | watch | quotation
	EmployeeID   string      `json:"employeeId,omitempty"`
	EmployeeName string      `json:"employeeName,omitempty"`
	Points       int         `json:"points,omitempty"`
	Reason       string      `json:"reason,omitempty"`
	Title        string      `json:"title,omitempty"`
	Message      string      `json:"message,omitempty"`
	Days         int         `json:"days,omitempty"`
	CustomerName string      `json:"customerName,omitempty"`
	ProjectName  string      `json:"projectName,omitempty"`
	Template     string      `json:"template,omitempty"`
	Items        []OrderItem `json:"items,omitempty"`
	// بعد التنفيذ
	Done   bool   `json:"done,omitempty"`
	Result string `json:"result,omitempty"`
	Link   string `json:"link,omitempty"`
	Error  string `json:"error,omitempty"`
	Skip   bool   `json:"skip,omitempty"`
}

type OrderItem struct {
	ProductName string  `json:"productName"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	Unit        string  `json:"unit,omitempty"`
}

type MatrixOrder struct {
	ID        string          `db:"id" json:"id"`
	Text      string          `db:"text" json:"text"`
	Status    string          `db:"status" json:"status"`
	RawAction json.RawMessage `db:"actions" json:"-"`
	Actions   []OrderAction   `db:"-" json:"actions"`
	Reply     *string         `db:"reply" json:"reply"`
	CreatedAt time.Time       `db:"createdAt" json:"createdAt"`
	DoneAt    *time.Time      `db:"doneAt" json:"doneAt"`
}

// OrderDeps الخدمات الي تنفّذ — كل وحدة بنفس مسارها العادي (إشعارات، سجل، قواعد).
type OrderDeps struct {
	Kpi        *KpiService
	Tasks      *repository.ExtraTaskRepository
	Notif      *repository.NotificationRepository
	Missions   *MatrixMissionService
	Quotations *QuotationService
	Templates  *QuotationTemplateService
}

type MatrixOrderService struct {
	db    *sqlx.DB
	agent *MatrixAgent
	deps  OrderDeps
	// propose بديل الموديل بالاختبار
	propose func(actor, text string) (string, []OrderAction, error)
}

func NewMatrixOrderService(db *sqlx.DB, agent *MatrixAgent, deps OrderDeps) *MatrixOrderService {
	s := &MatrixOrderService{db: db, agent: agent, deps: deps}
	s.propose = s.proposeAI
	return s
}

var ErrOrderNotFound = errors.New("الأمر مو موجود")

const orderSelect = `SELECT id, text, status, actions, reply, "createdAt", "doneAt" FROM "MatrixOrder"`

func (o *MatrixOrder) decode() {
	o.Actions = []OrderAction{}
	_ = json.Unmarshal(o.RawAction, &o.Actions)
}

func (s *MatrixOrderService) List(actor string) ([]MatrixOrder, error) {
	rows := []MatrixOrder{}
	if err := s.db.Select(&rows, orderSelect+` WHERE "actorId" = $1 ORDER BY "createdAt" DESC LIMIT 30`, actor); err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].decode()
	}
	return rows, nil
}

func (s *MatrixOrderService) get(id, actor string) (*MatrixOrder, error) {
	var o MatrixOrder
	if err := s.db.Get(&o, orderSelect+` WHERE id = $1 AND "actorId" = $2`, id, actor); err != nil {
		return nil, ErrOrderNotFound
	}
	o.decode()
	return &o, nil
}

// Propose يفهم الأمر ويخزن الإجراءات المقترحة (ما ينفّذ شي).
func (s *MatrixOrderService) Propose(actor, text string) (*MatrixOrder, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("اكتب الأمر")
	}
	if len([]rune(text)) > 1500 {
		return nil, errors.New("الأمر طويل — اختصره")
	}
	reply, acts, err := s.propose(actor, text)
	if err != nil {
		return nil, err
	}
	acts = s.validate(acts)
	raw, _ := json.Marshal(acts)
	status := OrderProposed
	if len(acts) == 0 {
		status = OrderCancelled
		if strings.TrimSpace(reply) == "" {
			reply = "ما فهمت شنو أسوي — اكتبه بشكل أوضح (منو، شنو، شكد)."
		}
	}
	var id string
	if err := s.db.Get(&id, `INSERT INTO "MatrixOrder"(id, "actorId", text, status, actions, reply)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, NULLIF($5, '')) RETURNING id`, actor, text, status, raw, reply); err != nil {
		return nil, err
	}
	return s.get(id, actor)
}

// validate يشيل الناقص ويحدّ القيم — الموديل يقترح، والقواعد هنا.
func (s *MatrixOrderService) validate(in []OrderAction) []OrderAction {
	out := []OrderAction{}
	for _, a := range in {
		needsEmp := a.Kind != "quotation"
		if needsEmp {
			if a.EmployeeID == "" || !s.scorableActive(a.EmployeeID) {
				continue
			}
		}
		switch a.Kind {
		case "points":
			if a.Points < 0 {
				a.Points = -a.Points
			}
			if a.Points < 1 {
				a.Points = 1
			}
			if a.Points > 8 {
				a.Points = 8
			}
			if strings.TrimSpace(a.Reason) == "" {
				continue
			}
		case "task":
			if strings.TrimSpace(a.Title) == "" {
				continue
			}
			if a.Days < 1 || a.Days > 30 {
				a.Days = 3
			}
		case "notify":
			if strings.TrimSpace(a.Message) == "" {
				continue
			}
		case "watch":
			if a.Days < 1 || a.Days > 30 {
				a.Days = 7
			}
		case "quotation":
			if strings.TrimSpace(a.CustomerName) == "" {
				continue
			}
			items := a.Items[:0]
			for _, it := range a.Items {
				if strings.TrimSpace(it.ProductName) == "" {
					continue
				}
				if it.Quantity < 1 {
					it.Quantity = 1
				}
				if it.UnitPrice < 0 {
					it.UnitPrice = 0
				}
				items = append(items, it)
			}
			a.Items = items
		default:
			continue
		}
		out = append(out, a)
		if len(out) >= 10 {
			break
		}
	}
	return out
}

// scorableActive: الأوامر على الموظفين بس — مو على المالك والمدير (ما ينقيّمون).
func (s *MatrixOrderService) scorableActive(id string) bool {
	var ok bool
	_ = s.db.Get(&ok, `SELECT EXISTS(SELECT 1 FROM "Employee" WHERE id = $1 AND status = 'ACTIVE' AND role NOT IN ('ADMIN', 'OWNER'))`, id)
	return ok
}

// Cancel يلغي أمر ما انفّذ.
func (s *MatrixOrderService) Cancel(id, actor string) error {
	res, err := s.db.Exec(`UPDATE "MatrixOrder" SET status = $3, "doneAt" = now() WHERE id = $1 AND "actorId" = $2 AND status = $4`,
		id, actor, OrderCancelled, OrderProposed)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrOrderNotFound
	}
	return nil
}

// Run ينفّذ الإجراءات (عدا المشطوبة بـskip) مرة وحدة بس — القفل بالحالة.
func (s *MatrixOrderService) Run(id, actor, role string, skip []int) (*MatrixOrder, error) {
	var raw json.RawMessage
	if err := s.db.Get(&raw, `UPDATE "MatrixOrder" SET status = $3 WHERE id = $1 AND "actorId" = $2 AND status = $4 RETURNING actions`,
		id, actor, OrderRunning, OrderProposed); err != nil {
		return nil, ErrOrderNotFound
	}
	acts := []OrderAction{}
	_ = json.Unmarshal(raw, &acts)
	skipped := map[int]bool{}
	for _, i := range skip {
		skipped[i] = true
	}
	ok := 0
	for i := range acts {
		if skipped[i] {
			acts[i].Skip = true
			continue
		}
		res, link, err := s.exec(actor, role, acts[i])
		if err != nil {
			acts[i].Error = err.Error()
			continue
		}
		acts[i].Done, acts[i].Result, acts[i].Link = true, res, link
		ok++
	}
	status := OrderDone
	if ok == 0 && len(skipped) < len(acts) {
		status = OrderFailed
	}
	out, _ := json.Marshal(acts)
	if _, err := s.db.Exec(`UPDATE "MatrixOrder" SET status = $2, actions = $3, "doneAt" = now() WHERE id = $1`, id, status, out); err != nil {
		return nil, err
	}
	return s.get(id, actor)
}

func (s *MatrixOrderService) exec(actor, role string, a OrderAction) (string, string, error) {
	if a.Kind != "quotation" && !s.scorableActive(a.EmployeeID) {
		return "", "", errors.New("الموظف مو فعّال")
	}
	switch a.Kind {
	case "points":
		if s.deps.Kpi == nil {
			return "", "", errors.New("الخصم مو متاح")
		}
		p := a.Points
		if _, err := s.deps.Kpi.Create(model.CreateKpiEvaluationRequest{EmployeeID: a.EmployeeID, EvaluatorID: actor, Points: &p, Reason: "🎯 أمر ماتركس: " + a.Reason}); err != nil {
			return "", "", err
		}
		return fmt.Sprintf("انخصم %d نقطة من %s", p, a.EmployeeName), "/kpi", nil
	case "task":
		due := time.Now().AddDate(0, 0, a.Days)
		if _, err := s.deps.Tasks.CreateFor(a.Title, a.Message, a.EmployeeID, actor, due); err != nil {
			return "", "", err
		}
		return fmt.Sprintf("انطت مهمة «%s» لـ%s (موعدها %s)", a.Title, a.EmployeeName, due.Format("2006-01-02")), "/extra-tasks", nil
	case "notify":
		if err := s.deps.Notif.Create(a.EmployeeID, "MATRIX_ORDER", "🎯 من الإدارة: "+a.Message); err != nil {
			return "", "", err
		}
		return "وصل التنبيه لـ" + a.EmployeeName, "", nil
	case "watch":
		if s.deps.Missions == nil {
			return "", "", errors.New("المراقبة مو متاحة")
		}
		if _, err := s.deps.Missions.Create(actor, a.EmployeeID, a.Reason, a.Days); err != nil {
			return "", "", err
		}
		return fmt.Sprintf("ماتركس يراقب %s %d يوم", a.EmployeeName, a.Days), "", nil
	case "quotation":
		return s.execQuotation(actor, role, a)
	}
	return "", "", errors.New("نوع مو معروف")
}

func (s *MatrixOrderService) execQuotation(actor, role string, a OrderAction) (string, string, error) {
	if a.Template != "" && s.deps.Templates != nil {
		ch := make([]TemplateItemChange, 0, len(a.Items))
		for _, it := range a.Items {
			q := it.Quantity
			ch = append(ch, TemplateItemChange{ProductName: it.ProductName, Unit: it.Unit, Quantity: &q})
		}
		q, err := s.deps.Templates.Use(actor, role, a.Template, TemplateUse{CustomerName: a.CustomerName, ProjectName: a.ProjectName, Items: ch})
		if err != nil {
			return "", "", err
		}
		return "انسوى عرض سعر " + q.QuotationNumber + " (مسودة) من نموذج «" + a.Template + "»", "/quotations/" + q.ID + "/edit", nil
	}
	req := model.CreateQuotationRequest{CustomerName: a.CustomerName}
	if a.ProjectName != "" {
		p := a.ProjectName
		req.ProjectName = &p
	}
	if a.Reason != "" {
		n := a.Reason
		req.Notes = &n
	}
	for _, it := range a.Items {
		in := model.QuotationItemInput{ProductName: it.ProductName, Quantity: it.Quantity, UnitPrice: it.UnitPrice}
		if it.Unit != "" {
			u := it.Unit
			in.Unit = &u
		}
		req.Items = append(req.Items, in)
	}
	q, err := s.deps.Quotations.Create(req, actor, role)
	if err != nil {
		return "", "", err
	}
	return "انسوى عرض سعر " + q.QuotationNumber + " (مسودة) — كمّل الأسعار والتفاصيل من الشاشة", "/quotations/" + q.ID + "/edit", nil
}

// ── الفهم بالذكاء الاصطناعي ──

const orderSystem = `أنت ماتركس، مساعد إدارة شركة. المدير يعطيك أمر بلهجة عراقية.
مهمتك: تحوّله لإجراءات بأدوات order_* — كل أداة تسجّل اقتراح بس، ما تنفّذ. المدير يشوفها ويضغط «نفّذ».
- الموظفين يجون برموز مثل موظف#3 — مرّر الرمز نفسه للأداة. والزبائن برموز زبون#n — مرّر الرمز نفسه.
- إذا تحتاج تعرف منو الموظف (مثلاً «فني الحجز الفلاني» أو «أكثر واحد متأخر») استعمل أدوات القراءة.
- «اخصم نقطة» = order_points بالنقاط والسبب. «انطيه شغلة/مهمة» = order_task. «بلّغه/كله» = order_notify.
  «راقبه» = order_watch. «سويلي عرض سعر» = order_quotation (استعمل quotation_template_list إذا ذكر نموذج).
- لا تخترع سبب أو رقم ما قاله المدير؛ إذا السبب مو واضح اكتب كلامه نفسه.
- بالنهاية جاوب بسطر أو سطرين: شنو فهمت، وإذا اكو شي ناقص أو مو واضح كوله.`

var orderBlockedTools = map[string]bool{"raise_manager_request": true, "quotation_from_template": true, "quotation_template_save": true}

var custTagRe = regexp.MustCompile(`زبون#(\d+)`)

// custMask يخفي أسماء الزبائن المعروفة من نص الأمر.
type custMask struct{ names []string }

func (s *MatrixOrderService) maskCustomers(text string) (string, *custMask) {
	m := &custMask{}
	rows := []string{}
	_ = s.db.Select(&rows, `SELECT DISTINCT name FROM "Customer" WHERE length(name) >= 4 AND position(name in $1) > 0 LIMIT 20`, text)
	sort.Slice(rows, func(i, j int) bool { return len(rows[i]) > len(rows[j]) })
	for _, n := range rows {
		if !strings.Contains(text, n) {
			continue
		}
		m.names = append(m.names, n)
		text = strings.ReplaceAll(text, n, fmt.Sprintf("زبون#%d", len(m.names)))
	}
	return text, m
}

func (m *custMask) show(s string) string {
	return custTagRe.ReplaceAllStringFunc(s, func(t string) string {
		var n int
		fmt.Sscanf(custTagRe.FindStringSubmatch(t)[1], "%d", &n)
		if n >= 1 && n <= len(m.names) {
			return m.names[n-1]
		}
		return t
	})
}

func (s *MatrixOrderService) proposeAI(actor, text string) (string, []OrderAction, error) {
	if s.agent == nil {
		return "", nil, ErrAgentOff
	}
	masked, cm := s.maskCustomers(text)
	acts := []OrderAction{}
	str := func(in map[string]any, k string) string { v, _ := in[k].(string); return strings.TrimSpace(v) }
	num := func(in map[string]any, k string) int { v, _ := in[k].(float64); return int(v) }
	emp := func(in map[string]any, names *nameMask) (string, string, error) {
		id, name := names.Resolve(str(in, "employee"))
		if id == "" {
			return "", "", errors.New("ما لكيت هالموظف — استعمل رمز موظف#n من أدوات القراءة")
		}
		return id, name, nil
	}
	add := func(a OrderAction, names *nameMask) string {
		acts = append(acts, a)
		return fmt.Sprintf("انسجل اقتراح رقم %d (%s) — ينتظر موافقة المدير.", len(acts), a.Kind)
	}
	empProp := map[string]any{"type": "string", "description": "رمز الموظف مثل موظف#3"}
	tools := []agentTool{{
		Name:  "order_points",
		Desc:  "يقترح خصم نقاط من موظف (١–٨). ما ينفّذ إلا بموافقة المدير.",
		Props: map[string]any{"employee": empProp, "points": map[string]any{"type": "integer"}, "reason": map[string]any{"type": "string"}},
		Req:   []string{"employee", "points", "reason"},
		Run: func(in map[string]any, names *nameMask) (string, error) {
			id, n, err := emp(in, names)
			if err != nil {
				return "", err
			}
			return add(OrderAction{Kind: "points", EmployeeID: id, EmployeeName: n, Points: num(in, "points"), Reason: cm.show(names.Show(str(in, "reason")))}, names), nil
		},
	}, {
		Name:  "order_task",
		Desc:  "يقترح مهمة لموظف بعنوان وتفاصيل وعدد أيام للموعد.",
		Props: map[string]any{"employee": empProp, "title": map[string]any{"type": "string"}, "details": map[string]any{"type": "string"}, "days": map[string]any{"type": "integer"}},
		Req:   []string{"employee", "title"},
		Run: func(in map[string]any, names *nameMask) (string, error) {
			id, n, err := emp(in, names)
			if err != nil {
				return "", err
			}
			return add(OrderAction{Kind: "task", EmployeeID: id, EmployeeName: n, Title: cm.show(names.Show(str(in, "title"))), Message: cm.show(names.Show(str(in, "details"))), Days: num(in, "days")}, names), nil
		},
	}, {
		Name:  "order_notify",
		Desc:  "يقترح تنبيه/رسالة من الإدارة لموظف.",
		Props: map[string]any{"employee": empProp, "message": map[string]any{"type": "string"}},
		Req:   []string{"employee", "message"},
		Run: func(in map[string]any, names *nameMask) (string, error) {
			id, n, err := emp(in, names)
			if err != nil {
				return "", err
			}
			return add(OrderAction{Kind: "notify", EmployeeID: id, EmployeeName: n, Message: cm.show(names.Show(str(in, "message")))}, names), nil
		},
	}, {
		Name:  "order_watch",
		Desc:  "يقترح إن ماتركس يراقب موظف عدد أيام ويكتب يومية وخلاصة.",
		Props: map[string]any{"employee": empProp, "note": map[string]any{"type": "string"}, "days": map[string]any{"type": "integer"}},
		Req:   []string{"employee"},
		Run: func(in map[string]any, names *nameMask) (string, error) {
			id, n, err := emp(in, names)
			if err != nil {
				return "", err
			}
			return add(OrderAction{Kind: "watch", EmployeeID: id, EmployeeName: n, Reason: cm.show(names.Show(str(in, "note"))), Days: num(in, "days")}, names), nil
		},
	}, {
		Name: "order_quotation",
		Desc: "يقترح عرض سعر (مسودة) لزبون، من نموذج محفوظ إذا انذكر، ويا البنود الي قالها المدير.",
		Props: map[string]any{
			"customer": map[string]any{"type": "string", "description": "رمز الزبون زبون#n أو الاسم مثل ما كتبه"},
			"project":  map[string]any{"type": "string"},
			"template": map[string]any{"type": "string", "description": "اسم النموذج (اختياري)"},
			"notes":    map[string]any{"type": "string"},
			"items": map[string]any{"type": "array", "items": map[string]any{"type": "object", "properties": map[string]any{
				"productName": map[string]any{"type": "string"}, "quantity": map[string]any{"type": "integer"},
				"unitPrice": map[string]any{"type": "number"}, "unit": map[string]any{"type": "string"}}}},
		},
		Req: []string{"customer"},
		Run: func(in map[string]any, names *nameMask) (string, error) {
			a := OrderAction{Kind: "quotation", CustomerName: cm.show(str(in, "customer")), ProjectName: cm.show(names.Show(str(in, "project"))),
				Template: str(in, "template"), Reason: cm.show(names.Show(str(in, "notes")))}
			if arr, ok := in["items"].([]any); ok {
				for _, x := range arr {
					m, _ := x.(map[string]any)
					if m == nil {
						continue
					}
					p, _ := m["unitPrice"].(float64)
					a.Items = append(a.Items, OrderItem{ProductName: str(m, "productName"), Quantity: num(m, "quantity"), UnitPrice: p, Unit: str(m, "unit")})
				}
			}
			return add(a, names), nil
		},
	}}
	for _, t := range quotationTemplateTools(s.db, actor, "") {
		if t.Name == "quotation_template_list" {
			tools = append(tools, t)
		}
	}
	// أدوات تكتب ما تشتغل هنا — الكتابة بس من «نفّذ».
	run, err := s.agent.ChatStream("CHAT", orderSystem, nil, masked, nil, ChatOpts{ExtraTools: tools, Actor: actor,
		AllowTool: func(n string) bool { return !orderBlockedTools[n] }})
	if err != nil {
		return "", nil, err
	}
	return cm.show(run.Answer), acts, nil
}
