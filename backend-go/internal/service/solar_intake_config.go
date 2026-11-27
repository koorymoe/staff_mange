package service

import (
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"

	"staffmange-api/internal/model"
)

// ═══ ☀️ إعدادات استفسار الطاقة الشمسية (دازاي 11-27) ═══
// «منين اخلي الاسعار مال اقتصادي…؟» — كلشي يدوي بالإعدادات:
//  • استهلاك كل جهاز (كيلو واط أقل/أكثر) — منه ينحسب حجم المنظومة.
//  • أسعار المنظومات حسب الحجم: لكل حجم سعر الاقتصادي والوسط والغالي.
// إذا جدول الأسعار فاضي، يرجع للحسبة القديمة (من المنظومات المسجّلة).

type SolarAppliance struct {
	Name  string  `json:"name"`
	MinKw float64 `json:"minKw"`
	MaxKw float64 `json:"maxKw"`
}

type SolarBand struct {
	Kw      float64 `json:"kw"`
	Economy float64 `json:"economy"`
	Mid     float64 `json:"mid"`
	High    float64 `json:"high"`
}

type SolarIntakeConfig struct {
	Appliances []SolarAppliance `json:"appliances"`
	Bands      []SolarBand      `json:"bands"`
}

var ErrSolarConfig = errors.New("تأكد: اسم الجهاز ما مكرر، الأرقام موجبة، والأقل ما يعبر الأكثر")

func (s *SolarInquiryService) Config(emp, role string) (*SolarIntakeConfig, error) {
	if !s.Access(emp, role).any() {
		return nil, ErrInquiryHidden
	}
	return s.loadConfig(), nil
}

func (s *SolarInquiryService) loadConfig() *SolarIntakeConfig {
	c := &SolarIntakeConfig{Appliances: []SolarAppliance{}, Bands: []SolarBand{}}
	var a, b []byte
	if err := s.db.QueryRow(`SELECT appliances, bands FROM "SolarIntakeConfig" WHERE id = 1`).Scan(&a, &b); err == nil {
		_ = json.Unmarshal(a, &c.Appliances)
		_ = json.Unmarshal(b, &c.Bands)
	}
	return c
}

// SaveConfig المدير والمالك بس.
func (s *SolarInquiryService) SaveConfig(emp, role string, c SolarIntakeConfig) (*SolarIntakeConfig, error) {
	if !model.IsAdminRole(role) {
		return nil, ErrInquiryHidden
	}
	seen := map[string]bool{}
	apps := []SolarAppliance{}
	for _, x := range c.Appliances {
		x.Name = strings.TrimSpace(x.Name)
		if x.Name == "" {
			continue
		}
		if seen[x.Name] || x.MinKw < 0 || x.MaxKw <= 0 || x.MinKw > x.MaxKw {
			return nil, ErrSolarConfig
		}
		seen[x.Name] = true
		apps = append(apps, x)
	}
	bands := []SolarBand{}
	kws := map[float64]bool{}
	for _, b := range c.Bands {
		if b.Kw <= 0 && b.Economy == 0 && b.Mid == 0 && b.High == 0 {
			continue
		}
		if b.Kw <= 0 || kws[b.Kw] || b.Economy < 0 || b.Mid < 0 || b.High < 0 {
			return nil, errors.New("كل حجم لازم يكون رقم موجب وما مكرر، والأسعار ما بيها سالب")
		}
		kws[b.Kw] = true
		bands = append(bands, b)
	}
	sort.Slice(bands, func(i, j int) bool { return bands[i].Kw < bands[j].Kw })
	aj, _ := json.Marshal(apps)
	bj, _ := json.Marshal(bands)
	if _, err := s.db.Exec(`INSERT INTO "SolarIntakeConfig" (id, appliances, bands, "updatedById", "updatedAt") VALUES (1, $1, $2, NULLIF($3,''), now())
		ON CONFLICT (id) DO UPDATE SET appliances = $1, bands = $2, "updatedById" = NULLIF($3,''), "updatedAt" = now()`, string(aj), string(bj), emp); err != nil {
		return nil, err
	}
	return s.loadConfig(), nil
}

// bandTiers أسعار الدرجات من الجدول اليدوي للأحجام بين minKw و maxKw
// (وإذا ماكو حجم بالمدى، أقرب حجم أكبر منه، وإلا الأكبر). nil = الجدول فاضي.
func bandTiers(bands []SolarBand, minKw, maxKw float64) []PriceTier {
	if len(bands) == 0 {
		return nil
	}
	pick := []SolarBand{}
	for _, b := range bands {
		if b.Kw >= minKw && b.Kw <= maxKw {
			pick = append(pick, b)
		}
	}
	if len(pick) == 0 {
		best := bands[len(bands)-1]
		for _, b := range bands {
			if b.Kw >= maxKw {
				best = b
				break
			}
		}
		pick = []SolarBand{best}
	}
	out := []PriceTier{}
	for _, t := range []struct {
		name string
		get  func(SolarBand) float64
	}{{"ECONOMY", func(b SolarBand) float64 { return b.Economy }}, {"MID", func(b SolarBand) float64 { return b.Mid }}, {"HIGH", func(b SolarBand) float64 { return b.High }}} {
		lo, hi, n := math.Inf(1), 0.0, 0
		for _, b := range pick {
			if v := t.get(b); v > 0 {
				lo, hi, n = math.Min(lo, v), math.Max(hi, v), n+1
			}
		}
		if n > 0 {
			out = append(out, PriceTier{Tier: t.name, Min: lo, Max: hi, Count: n})
		}
	}
	return out
}
