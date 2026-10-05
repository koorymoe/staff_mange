package service

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ عدّاد كلفة الذكاء الاصطناعي ═══
// كل نداء لهايكو يسجّل توكناته هنا (RecordAIUsage بعد كل Messages.New).
// السعر بالدولار لكل مليون توكن — من جدول أسعار Anthropic (أيلول ٢٠٢٦).

var aiUsageRepo *repository.AiUsageRepository

// ═══ مفاتيح الكلفة ═══
var (
	aiSwitchRepo *repository.SystemSwitchRepository
	aiSwitchMu   sync.Mutex
	aiSwitchAt   time.Time
	aiSwitchMap  map[string]bool
)

func SetAISwitches(r *repository.SystemSwitchRepository) { aiSwitchRepo = r }

// ErrAIFeatureOff الميزة مطفية من مفاتيح الكلفة — كل مكان يرجع لقواعده.
var ErrAIFeatureOff = errors.New("ميزة الذكاء مطفية لتوفير الكلفة")

// AIFeatureOn هل الميزة مسموح تستعمل هايكو؟ (كاش دقيقة)
func AIFeatureOn(feature string) bool {
	key, ok := model.AIFeatureSwitch[feature]
	if !ok || aiSwitchRepo == nil {
		return true
	}
	aiSwitchMu.Lock()
	defer aiSwitchMu.Unlock()
	if aiSwitchMap == nil || time.Since(aiSwitchAt) > time.Minute {
		if m, err := aiSwitchRepo.All(); err == nil {
			aiSwitchMap, aiSwitchAt = m, time.Now()
		}
	}
	if aiSwitchMap == nil {
		return true
	}
	return aiSwitchMap[key]
}

func SetAIUsageRepo(r *repository.AiUsageRepository) { aiUsageRepo = r }

func RecordAIUsage(feature, model string, u anthropic.Usage) {
	if aiUsageRepo == nil {
		return
	}
	go aiUsageRepo.Add(feature, model, u.InputTokens, u.OutputTokens, u.CacheReadInputTokens, u.CacheCreationInputTokens)
}

type modelPrice struct{ In, Out float64 }

// الكاش: القراءة ١٠٪ من سعر الداخل، والكتابة ١٢٥٪.
func priceOf(model string) modelPrice {
	m := strings.ToLower(model)
	switch {
	case strings.Contains(m, "haiku"):
		return modelPrice{1, 5}
	case strings.Contains(m, "sonnet-4"):
		return modelPrice{3, 15}
	case strings.Contains(m, "sonnet"):
		return modelPrice{2, 10}
	case strings.Contains(m, "opus-5-5"):
		return modelPrice{4, 20}
	case strings.Contains(m, "opus"):
		return modelPrice{5, 25}
	}
	return modelPrice{1, 5}
}

func usageCost(r repository.AiUsageRow) float64 {
	p := priceOf(r.Model)
	return (float64(r.InputTokens)*p.In + float64(r.OutputTokens)*p.Out +
		float64(r.CacheRead)*p.In*0.1 + float64(r.CacheWrite)*p.In*1.25) / 1e6
}

var AIFeatureLabels = map[string]string{
	"GUIDE":           "توجيهات ماتركس للموظفين",
	"JUDGE":           "أحكام ماتركس",
	"DISCOVERY":       "ماتركس اكتشف",
	"VOICE":           "تحليل الفويس",
	"ASK":             "اسأل ماتركس",
	"LEARNING":        "توقعات ماتركس",
	"EMPLOYEE_REPORT": "تقارير الموظفين",
}

type AIUsageFeature struct {
	Feature   string  `json:"feature"`
	Label     string  `json:"label"`
	Calls     int     `json:"calls"`
	Tokens    int64   `json:"tokens"`
	CostToday float64 `json:"costToday"`
	Cost30    float64 `json:"cost30"`
}

type AIUsageReport struct {
	Today    float64           `json:"today"`
	Month    float64           `json:"month"` // الشهر الحالي لحد اليوم
	Last30   float64           `json:"last30"`
	Forecast float64           `json:"forecast"` // توقّع الشهر كامل بمعدل آخر ٧ أيام
	Features []AIUsageFeature  `json:"features"`
	Switches map[string]string `json:"switches"`
	Days     []map[string]any  `json:"days"`
}

func AIUsageSummary() (*AIUsageReport, error) {
	rep := &AIUsageReport{Features: []AIUsageFeature{}, Days: []map[string]any{}, Switches: model.AIFeatureSwitch}
	if aiUsageRepo == nil {
		return rep, nil
	}
	rows, err := aiUsageRepo.Since(31)
	if err != nil {
		return nil, err
	}
	now := time.Now().In(debriefLoc)
	today := now.Format("2006-01-02")
	monthPrefix := now.Format("2006-01")
	weekAgo := now.AddDate(0, 0, -7).Format("2006-01-02")
	feat := map[string]*AIUsageFeature{}
	byDay := map[string]float64{}
	week := 0.0
	for _, r := range rows {
		c := usageCost(r)
		d := r.Day.Format("2006-01-02")
		byDay[d] += c
		rep.Last30 += c
		if d == today {
			rep.Today += c
		}
		if strings.HasPrefix(d, monthPrefix) {
			rep.Month += c
		}
		if d > weekAgo {
			week += c
		}
		f := feat[r.Feature]
		if f == nil {
			f = &AIUsageFeature{Feature: r.Feature, Label: AIFeatureLabels[r.Feature]}
			feat[r.Feature] = f
		}
		f.Calls += r.Calls
		f.Tokens += r.InputTokens + r.OutputTokens + r.CacheRead + r.CacheWrite
		f.Cost30 += c
		if d == today {
			f.CostToday += c
		}
	}
	rep.Forecast = week / 7 * 30
	// كل الميزات تطلع حتى الي ما صرفت، حتى مفتاحها يبين.
	for k := range model.AIFeatureSwitch {
		if feat[k] == nil {
			feat[k] = &AIUsageFeature{Feature: k, Label: AIFeatureLabels[k]}
		}
	}
	for _, f := range feat {
		rep.Features = append(rep.Features, *f)
	}
	for i := 0; i < 30; i++ {
		d := now.AddDate(0, 0, -i).Format("2006-01-02")
		rep.Days = append(rep.Days, map[string]any{"day": d, "cost": byDay[d]})
	}
	return rep, nil
}
