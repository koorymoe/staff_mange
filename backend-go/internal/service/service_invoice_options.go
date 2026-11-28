package service

import (
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

// ═══ 🧾 خيارات الفاتورة لكل خدمة — تجريبي (ل) 11-24 دازاي ═══
// «نضيف فاتورة بالإعدادات: نوع الخدمة، ونضيف خانات — تسليك كيبل بسعر المتر، غسل لوح بسعر اللوح…
//  وآخر خيار ثابت: شنو أقل سعر تريده يظهر بالفاتورة» — الحد الأدنى على الشغل بس، والمواد فوگه.
// «تنضاف بس ما تنطلق، تبقى غير معتمدة — القديمة هي المعتمدة لحد ما نسوي تيست ونقرر نطلقها».
// يعني: هنا إعدادات + حاسبة تجربة بس؛ الفواتير الحقيقية ما تمسها.

type ServiceInvoiceOption struct {
	ID        string  `db:"id" json:"id"`
	Label     string  `db:"label" json:"label"`
	Unit      string  `db:"unit" json:"unit"`
	UnitPrice float64 `db:"unitPrice" json:"unitPrice"`
}

type ServiceInvoiceTemplate struct {
	ServiceID   string  `db:"serviceId" json:"serviceId"`
	ServiceName string  `db:"serviceName" json:"serviceName"`
	MinWork     float64 `db:"minWork" json:"minWork"`
	TiersRaw    []byte  `db:"minTiers" json:"-"`
	// MinTiers درجات أقل سعر (تصاعدي): الشغل يطلع بأصغر درجة توصل للمجموع، وإذا المجموع فات كل الدرجات ينحسب مثل ما هو.
	MinTiers []float64              `db:"-" json:"minTiers"`
	Options  []ServiceInvoiceOption `db:"-" json:"options"`
}

type ServiceInvoiceOptions struct{ db *sqlx.DB }

func NewServiceInvoiceOptions(db *sqlx.DB) *ServiceInvoiceOptions {
	return &ServiceInvoiceOptions{db: db}
}

// List كل الخدمات ووياها خياراتها وحدّها الأدنى (الخدمة البلا خيارات تطلع فاضية).
func (s *ServiceInvoiceOptions) List() ([]ServiceInvoiceTemplate, error) {
	ts := []ServiceInvoiceTemplate{}
	if err := s.db.Select(&ts, `SELECT sv.id AS "serviceId", sv.name AS "serviceName", COALESCE(r."minWork", 0)::float8 AS "minWork", COALESCE(r."minTiers", '[]'::jsonb) AS "minTiers"
		FROM "Service" sv LEFT JOIN "ServiceInvoiceRule" r ON r."serviceId" = sv.id ORDER BY sv.name`); err != nil {
		return nil, err
	}
	type row struct {
		ServiceInvoiceOption
		ServiceID string `db:"serviceId"`
	}
	rows := []row{}
	if err := s.db.Select(&rows, `SELECT id, "serviceId", label, unit, "unitPrice"::float8 AS "unitPrice"
		FROM "ServiceInvoiceOption" ORDER BY "serviceId", "sortOrder", "createdAt"`); err != nil {
		return nil, err
	}
	by := map[string][]ServiceInvoiceOption{}
	for _, r := range rows {
		by[r.ServiceID] = append(by[r.ServiceID], r.ServiceInvoiceOption)
	}
	for i := range ts {
		ts[i].MinTiers = []float64{}
		_ = json.Unmarshal(ts[i].TiersRaw, &ts[i].MinTiers)
		if len(ts[i].MinTiers) == 0 && ts[i].MinWork > 0 {
			ts[i].MinTiers = []float64{ts[i].MinWork}
		}
		ts[i].Options = by[ts[i].ServiceID]
		if ts[i].Options == nil {
			ts[i].Options = []ServiceInvoiceOption{}
		}
	}
	return ts, nil
}

// Save يبدّل خيارات الخدمة كلها بالي انرسل (بالترتيب) ويحفظ الحد الأدنى.
func (s *ServiceInvoiceOptions) Save(serviceID, by string, minWork float64, tiers []float64, opts []ServiceInvoiceOption) error {
	if minWork < 0 || math.IsNaN(minWork) {
		return errors.New("أقل سعر لازم صفر أو أكثر")
	}
	if tiers == nil && minWork > 0 {
		tiers = []float64{minWork}
	}
	clean := []float64{}
	for _, t := range tiers {
		if t < 0 || math.IsNaN(t) {
			return errors.New("درجات أقل سعر لازم صفر أو أكثر")
		}
		if t > 0 {
			clean = append(clean, t)
		}
	}
	if len(clean) > 20 {
		return errors.New("أكثر شي ٢٠ درجة")
	}
	sort.Float64s(clean)
	minWork = 0
	if len(clean) > 0 {
		minWork = clean[0]
	}
	tiersJSON, _ := json.Marshal(clean)
	if len(opts) > 40 {
		return errors.New("أكثر شي ٤٠ خيار للخدمة")
	}
	for i := range opts {
		opts[i].Label, opts[i].Unit = strings.TrimSpace(opts[i].Label), strings.TrimSpace(opts[i].Unit)
		if opts[i].Label == "" || len([]rune(opts[i].Label)) > 80 {
			return errors.New("كل خيار لازم إله اسم (لحد ٨٠ حرف)")
		}
		if opts[i].Unit == "" {
			opts[i].Unit = "قطعة"
		}
		if opts[i].UnitPrice < 0 || math.IsNaN(opts[i].UnitPrice) {
			return errors.New("السعر ما يصير سالب")
		}
	}
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var ok bool
	if err := tx.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "Service" WHERE id = $1)`, serviceID); err != nil || !ok {
		return ErrDesignHidden
	}
	if _, err := tx.Exec(`DELETE FROM "ServiceInvoiceOption" WHERE "serviceId" = $1`, serviceID); err != nil {
		return err
	}
	for i, o := range opts {
		if _, err := tx.Exec(`INSERT INTO "ServiceInvoiceOption" ("serviceId", label, unit, "unitPrice", "sortOrder") VALUES ($1, $2, $3, $4, $5)`,
			serviceID, o.Label, o.Unit, o.UnitPrice, i); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`INSERT INTO "ServiceInvoiceRule" ("serviceId", "minWork", "minTiers", "updatedById", "updatedAt") VALUES ($1, $2, $4::jsonb, NULLIF($3, ''), now())
		ON CONFLICT ("serviceId") DO UPDATE SET "minWork" = $2, "minTiers" = $4::jsonb, "updatedById" = NULLIF($3, ''), "updatedAt" = now()`, serviceID, minWork, by, string(tiersJSON)); err != nil {
		return err
	}
	return tx.Commit()
}

type QuoteLine struct {
	Label string  `json:"label"`
	Unit  string  `json:"unit"`
	Qty   float64 `json:"qty"`
	Price float64 `json:"price"`
	Total float64 `json:"total"`
}

type ServiceQuote struct {
	Lines     []QuoteLine `json:"lines"`
	Sum       float64     `json:"sum"`       // مجموع الخيارات
	MinWork   float64     `json:"minWork"`   // الحد الأدنى للشغل
	Work      float64     `json:"work"`      // الشغل = الأكبر بيناتهم
	Materials float64     `json:"materials"` // المواد (فوگ الحد)
	Total     float64     `json:"total"`
	MinUsed   bool        `json:"minUsed"`
}

// QuoteFor الحساب: الشغل = max(Σ كمية×سعر، الحد الأدنى)، والمواد تنضاف فوگه.
func QuoteFor(t ServiceInvoiceTemplate, qty map[string]float64, materials float64) ServiceQuote {
	q := ServiceQuote{Lines: []QuoteLine{}, MinWork: t.MinWork, Materials: math.Max(0, materials)}
	for _, o := range t.Options {
		n := qty[o.ID]
		if n <= 0 {
			continue
		}
		l := QuoteLine{Label: o.Label, Unit: o.Unit, Qty: n, Price: o.UnitPrice, Total: math.Round(n * o.UnitPrice)}
		q.Lines = append(q.Lines, l)
		q.Sum += l.Total
	}
	q.Work = q.Sum
	tiers := append([]float64(nil), t.MinTiers...)
	if len(tiers) == 0 && t.MinWork > 0 {
		tiers = []float64{t.MinWork}
	}
	sort.Float64s(tiers)
	for _, m := range tiers { // أصغر درجة توصل للمجموع
		if m >= q.Sum {
			if m > q.Sum {
				q.Work, q.MinUsed, q.MinWork = m, true, m
			}
			break
		}
	}
	q.Total = q.Work + q.Materials
	return q
}

// Quote حاسبة التجربة لخدمة.
func (s *ServiceInvoiceOptions) Quote(serviceID string, qty map[string]float64, materials float64) (ServiceQuote, error) {
	ts, err := s.List()
	if err != nil {
		return ServiceQuote{}, err
	}
	for _, t := range ts {
		if t.ServiceID == serviceID {
			return QuoteFor(t, qty, materials), nil
		}
	}
	return ServiceQuote{}, ErrDesignHidden
}
