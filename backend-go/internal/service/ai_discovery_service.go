package service

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ الاستكشاف الأسبوعي (ماتركس) ═══
//
// «ما أريد أشرحله بالمضبوط، أريده يكتشف الي أنا ما مكتشفه» — بس
// بمحاور **موجودة أصلاً** بالجدول (يوم، منظومة، وردية، موظف)، مو
// نموذجاً يبني استعلامات جديدة (خطر حقن وبيانات حساسة). كل خلية
// بالشبكة تُحسب حتمياً بالكود ضد **كل** البيانات، وتعبر بوابتين قبل
// ما توصل لأي نموذج لغوي:
//
//	١. حد أدنى للعيّنة — رقم من عيّنتين ما يستاهل قرار
//	٢. انحراف عن المعدل العام — نفس مفهوم `FuelAnomalyThresholdPercent`
//	   الموجود بـ`vehicle_service.go`، مو اختباراً إحصائياً رسمياً
const (
	DiscoveryMinSampleSize         = 5
	DiscoveryDeviationThresholdPct = 30.0
)

type AiDiscoveryService struct {
	repo      *repository.AiDiscoveryRepository
	aiRepo    *repository.AiRepository
	employees *repository.EmployeeRepository
	notifRepo *repository.NotificationRepository
	narrator  DiscoveryNarrator
}

func NewAiDiscoveryService(
	repo *repository.AiDiscoveryRepository,
	aiRepo *repository.AiRepository,
	employees *repository.EmployeeRepository,
	notifRepo *repository.NotificationRepository,
) *AiDiscoveryService {
	return &AiDiscoveryService{
		repo: repo, aiRepo: aiRepo, employees: employees, notifRepo: notifRepo,
		narrator: DeterministicDiscoveryNarrator{},
	}
}

// SetNarrator يربط نموذجاً خارجياً اختيارياً — بلا مفتاح تبقى الجملة
// الحتمية شغّالة (نفس فصل `ModelJudge`/`RulesJudge`).
func (s *AiDiscoveryService) SetNarrator(n DiscoveryNarrator) {
	if n != nil {
		s.narrator = n
	}
}

// RunWeeklyIfDue تُستدعى دورياً (كل ٦ ساعات). ما تسوي شي إلا يوم
// الاثنين بتوقيت بغداد، وبعلامة أسبوع لسه ما انحجزت — نفس حيلة
// `DailyDebriefService` بالضبط، بس المفتاح تاريخ يوم الاثنين نفسه.
func (s *AiDiscoveryService) RunWeeklyIfDue() error {
	now := time.Now().In(debriefLoc)
	if now.Weekday() != time.Monday {
		return nil
	}
	weekKey := now.Format("2006-01-02")
	claimed, err := s.aiRepo.ClaimDailyMarker("WEEKLY_DISCOVERY_SENT", weekKey)
	if err != nil || !claimed {
		return err
	}

	flagged, err := s.RunGrid()
	if err != nil {
		return err
	}
	if len(flagged) == 0 {
		return nil
	}
	msg := fmt.Sprintf(
		"🔎 الاستكشاف الأسبوعي: %d نمط شاذ عن المعدل يستاهل مراجعة — التفاصيل بشاشة مؤشرات الذكاء الاصطناعي.",
		len(flagged),
	)
	if err := s.notifRepo.CreateForRole("OWNER", "AI_WEEKLY_DISCOVERY", msg); err != nil {
		return err
	}
	return s.notifRepo.CreateForRole("ADMIN", "AI_WEEKLY_DISCOVERY", msg)
}

type discoveryFetcher struct {
	key        string
	byAxis     func(axis string) ([]repository.AxisCell, error)
	byEmployee func() ([]repository.AxisCell, error)
}

// RunGrid تحسب الشبكة كاملة (٥ مقاييس × ٤ محاور)، تحفظ الخلايا الي
// عبرت البوابتين بـ`AiMetric`، وترجّعهن للعدّ بالإشعار.
func (s *AiDiscoveryService) RunGrid() ([]DiscoveryCell, error) {
	fetchers := []discoveryFetcher{
		{
			key:        model.AiDiscoveryLateStartRate,
			byAxis:     func(axis string) ([]repository.AxisCell, error) { return s.repo.SignalRateByBookingAxis(model.AiSignalLateStart, axis) },
			byEmployee: func() ([]repository.AxisCell, error) { return s.repo.SignalRateByEmployee(model.AiSignalLateStart) },
		},
		{
			key:        model.AiDiscoveryWorkStopRate,
			byAxis:     func(axis string) ([]repository.AxisCell, error) { return s.repo.SignalRateByBookingAxis(model.AiSignalWorkStopped, axis) },
			byEmployee: func() ([]repository.AxisCell, error) { return s.repo.SignalRateByEmployee(model.AiSignalWorkStopped) },
		},
		{key: model.AiDiscoveryPartialRate, byAxis: s.repo.PartialRateByBookingAxis, byEmployee: s.repo.PartialRateByEmployee},
		{key: model.AiDiscoveryInvoiceAdjustRate, byAxis: s.repo.InvoiceAdjustRateByBookingAxis, byEmployee: s.repo.InvoiceAdjustRateByEmployee},
		{key: model.AiDiscoveryDisciplineRate, byAxis: s.repo.DisciplineRateByBookingAxis, byEmployee: s.repo.DisciplineRateByEmployee},
	}

	// ⚠️ periodEnd = periodStart عمداً (لقطة لحظية بيوم الاثنين، مو
	// فترة مستقبلية) — لو حطينا +٧ أيام، `ListMetrics` (شرطها
	// periodEnd <= to) ما ترجّع الصف إلا بعد أسبوع كامل من حسابه،
	// والمالك ما يشوف الاستكشاف إلا متأخراً أسبوعاً عن وقته الحقيقي.
	now := time.Now().In(debriefLoc)
	periodStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, debriefLoc).UTC()
	periodEnd := periodStart

	var flagged []DiscoveryCell
	for _, f := range fetchers {
		for _, axis := range []string{model.AiAxisDayOfWeek, model.AiAxisSystemType, model.AiAxisShift} {
			rows, err := f.byAxis(axis)
			if err != nil {
				log.Printf("[ai-discovery] %s/%s: %v", f.key, axis, err)
				continue
			}
			flagged = append(flagged, s.evaluateAndSave(f.key, axis, rows, periodStart, periodEnd)...)
		}
		empRows, err := f.byEmployee()
		if err != nil {
			log.Printf("[ai-discovery] %s/EMPLOYEE: %v", f.key, err)
			continue
		}
		flagged = append(flagged, s.evaluateAndSave(f.key, model.AiAxisEmployee, empRows, periodStart, periodEnd)...)
	}
	return flagged, nil
}

// evaluateAndSave يطبّق البوابتين على خلايا محور وحد، ويحفظ الي يعبرهن.
func (s *AiDiscoveryService) evaluateAndSave(metricKey, axis string, rows []repository.AxisCell, periodStart, periodEnd time.Time) []DiscoveryCell {
	totalNum, totalDen := 0, 0
	for _, r := range rows {
		totalNum += r.Numerator
		totalDen += r.Denominator
	}
	// ماكو عيّنات كافية أصلاً بهذا المحور — ماكو معدل نقارن بيه.
	if totalDen == 0 {
		return nil
	}
	overallAvg := float64(totalNum) * 100 / float64(totalDen)
	// معدل عام صفر يعني الحدث ما صار إطلاقاً — أي ظهور نادر يطلّع
	// انحرافاً لا نهائياً بلا معنى حقيقي، فنتخطى هذا المحور كلياً.
	if overallAvg <= 0 {
		return nil
	}

	var out []DiscoveryCell
	for _, r := range rows {
		if r.Denominator < DiscoveryMinSampleSize { // بوابة ١
			continue
		}
		rate := float64(r.Numerator) * 100 / float64(r.Denominator)
		deviation := (rate - overallAvg) / overallAvg * 100
		if deviation < 0 {
			deviation = -deviation
		}
		if deviation < DiscoveryDeviationThresholdPct { // بوابة ٢
			continue
		}

		label := s.axisValueLabel(axis, r.ScopeID)
		cell := DiscoveryCell{
			MetricKey: metricKey, Axis: axis, ScopeID: r.ScopeID, ValueLabel: label,
			Rate: rate, OverallAvg: overallAvg, Sample: r.Denominator,
		}
		narration := s.narrator.Narrate(cell)

		details, _ := json.Marshal(map[string]any{
			"axisLabel":  model.AiAxisLabel(axis),
			"valueLabel": label,
			"overallAvg": overallAvg,
			"narration":  narration,
		})
		scopeID := r.ScopeID
		if err := s.aiRepo.UpsertMetric(model.AiMetric{
			MetricKey: metricKey, Scope: axis, ScopeID: &scopeID,
			PeriodStart: periodStart, PeriodEnd: periodEnd,
			Value: rate, SampleCount: r.Denominator, Details: details,
		}); err != nil {
			log.Printf("[ai-discovery] حفظ %s/%s/%s: %v", metricKey, axis, r.ScopeID, err)
			continue
		}
		out = append(out, cell)
	}
	return out
}

func (s *AiDiscoveryService) axisValueLabel(axis, scopeID string) string {
	switch axis {
	case model.AiAxisDayOfWeek:
		return model.AiDayOfWeekLabel(scopeID)
	case model.AiAxisShift:
		return model.AiShiftLabel(scopeID)
	case model.AiAxisEmployee:
		if name, err := s.employees.NameByID(scopeID); err == nil && name != "" {
			return name
		}
		return "موظف محذوف"
	default:
		return scopeID
	}
}
