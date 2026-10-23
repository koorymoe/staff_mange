package service

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"staffmange-api/internal/repository"
)

// ═══ 💵 الحد اليومي بالدولار لكل وكيل (دازاي 10-23: «نتأكد ما يصرف فوك الحد») ═══
// كل نداء لهايكو يمر من NewAIClient: قبل ما يطلع الطلب نشوف شكد صرف الوكيل
// اليوم (بغداد). إذا وصل حده، الطلب ما يطلع أصلاً والميزة ترجع لقواعدها.
// الصرف يتحسب من جدول AiUsage (يبقى بعد إعادة التشغيل) + كل نداء جديد بالذاكرة.
// ⚠️ والضمان الأخير سقف الـWorkspace بالكونسول — هذا يمنع نوصله أصلاً.

const (
	AIAgentMatrix = "MATRIX"
	AIAgentIris   = "IRIS"
)

// ErrAIBudget الوكيل خلّص حده اليومي.
var ErrAIBudget = errors.New("الوكيل وصل حده اليومي بالدولار — يرجع باچر")

var (
	aiBudgetMu    sync.Mutex
	aiBudgetUSD   = map[string]float64{}
	aiBudgetSpent = map[string]float64{}
	aiBudgetDay   string
)

// SetAIBudget الحد اليومي بالدولار لوكيل (٠ = بلا حد بالكود).
func SetAIBudget(agent string, usd float64) {
	aiBudgetMu.Lock()
	aiBudgetUSD[agent] = usd
	aiBudgetMu.Unlock()
}

func aiAgentOf(feature string) string {
	switch feature {
	case "SECRETARY":
		return AIAgentIris
	case "QUALITY_AGENT":
		return WaAgentQuality
	case "BOOKING_AGENT":
		return WaAgentBooking
	}
	return AIAgentMatrix
}

// aiBudgetRollLocked أول نداء بكل يوم يحمّل صرف اليوم من القاعدة.
func aiBudgetRollLocked() {
	day := time.Now().In(debriefLoc).Format("2006-01-02")
	if day == aiBudgetDay || aiUsageRepo == nil {
		return
	}
	rows, err := aiUsageRepo.Since(1)
	if err != nil {
		return
	}
	// الي انحسب بالذاكرة قبل التحميل ينحفظ أصلاً بالقاعدة (Add)، فنبدي من القاعدة.
	aiBudgetDay = day
	aiBudgetSpent = map[string]float64{}
	for _, r := range rows {
		aiBudgetSpent[aiAgentOf(r.Feature)] += usageCost(r)
	}
}

func aiBudgetAdd(feature, model string, u anthropic.Usage) {
	aiBudgetMu.Lock()
	defer aiBudgetMu.Unlock()
	aiBudgetRollLocked()
	aiBudgetSpent[aiAgentOf(feature)] += usageCost(repository.AiUsageRow{Model: model,
		InputTokens: u.InputTokens, OutputTokens: u.OutputTokens,
		CacheRead: u.CacheReadInputTokens, CacheWrite: u.CacheCreationInputTokens})
}

// AIBudgetState صرف اليوم والحد لوكيل.
func AIBudgetState(agent string) (spent, limit float64) {
	aiBudgetMu.Lock()
	defer aiBudgetMu.Unlock()
	aiBudgetRollLocked()
	return aiBudgetSpent[agent], aiBudgetUSD[agent]
}

// NewAIClient عميل Claude لوكيل، والحد اليومي محروس قبل كل طلب.
func NewAIClient(agent, apiKey string) anthropic.Client {
	return anthropic.NewClient(option.WithAPIKey(apiKey),
		option.WithMiddleware(func(req *http.Request, next option.MiddlewareNext) (*http.Response, error) {
			if spent, limit := AIBudgetState(agent); limit > 0 && spent >= limit {
				return nil, ErrAIBudget
			}
			return next(req)
		}))
}
