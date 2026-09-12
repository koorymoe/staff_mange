package service

import (
	"encoding/json"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ حاسبة المؤشرات ═══
//
// «أريده يحلل ويقيس ويؤشّر… تصير عندي إحصائية بالإحصائيات اسمها
// إحصائيات ومؤشرات الذكاء الاصطناعي».
//
// ⚠️ المؤشرات تنحسب **من الأدلة المتراكمة** مو من عدّادات خام. الفرق:
// «١٢ توقف هذا الشهر» رقم خام ما يگول شي، أما «٤٠٪ من التوقفات سببها
// مادة ما انطلبت» فهذا يوجّه قرار (راجع تحضير الكادر) — والثاني ما
// ينحسب إلا لأننا جمّعنا الأدلة لكل توقف.
//
// ⚠️ وما تنحسب من فراغ: المؤشر بلا عيّنات كافية يضلّل. نحفظ
// sampleCount مع كل رقم، والواجهة تعرضه — «٥٠٪» من عيّنتين مو مثل
// «٥٠٪» من مئتين.
type AiMetricsService struct {
	repo *repository.AiRepository
}

func NewAiMetricsService(repo *repository.AiRepository) *AiMetricsService {
	return &AiMetricsService{repo: repo}
}

// Recompute يعيد حساب مؤشرات الفترة.
//
// ⚠️ إعادة الحساب تحدّث نفس الصف (فهرس فريد على المفتاح+النطاق+الفترة)
// مو تضيف صفاً ثانياً — وإلا كل ضغطة تضاعف الأرقام.
func (s *AiMetricsService) Recompute(from, to time.Time) (int, error) {
	signals, err := s.repo.ListSignals(model.AiSignalWorkStopped, 300)
	if err != nil {
		return 0, err
	}

	// نمشي بالإشارات الي **إلها أدلة** بس: الإشارة بلا أدلة ما نعرف
	// عنها شي، وحسابها ضمن المجموع يطلّع نسب كذب.
	type stopFacts = model.WorkStopEvidence
	facts := []stopFacts{}
	for i := range signals {
		sig := signals[i]
		if sig.OccurredAt.Before(from) || sig.OccurredAt.After(to) {
			continue
		}
		if sig.Evidence == nil {
			continue
		}
		var f stopFacts
		if err := json.Unmarshal(sig.Evidence.Facts, &f); err != nil {
			continue
		}
		facts = append(facts, f)
	}

	n := len(facts)
	save := func(key string, value float64, details map[string]any) error {
		raw, _ := json.Marshal(details)
		return s.repo.UpsertMetric(model.AiMetric{
			MetricKey: key, Scope: "COMPANY", PeriodStart: from, PeriodEnd: to,
			Value: value, SampleCount: n, Details: raw,
		})
	}

	// ⚠️ بلا عيّنات ما نحفظ أصفاراً: صفر بالشاشة يعني «قسنا وطلع صفر»،
	// والحقيقة إننا ما قسنا أصلاً. الفرق مهم لأن الأول يطمّن والثاني
	// يعني «انتظر بيانات».
	if n == 0 {
		return 0, nil
	}

	pct := func(count int) float64 { return float64(count) * 100 / float64(n) }

	// ١. توقف بسبب مادة ما انطلبت: ماكو طلب وماكو زيادة بالسلة.
	miss := 0
	// ٢. الزبون طلب زيادة بالموقع: السلة زادت بعد ما بدأ الشغل.
	creep := 0
	// ٣. التوفير تأخر: طلب قبل التوقف وما انوفّر.
	procDelay := 0
	totalMinutes := 0
	for _, f := range facts {
		if f.ProcurementRequests == 0 && f.CartItemsAfterStart == 0 {
			miss++
		}
		if f.CartItemsAfterStart > 0 {
			creep++
		}
		if f.RequestedBeforeStop && f.LastRequestStatus != "FULFILLED" {
			procDelay++
		}
		totalMinutes += f.WorkedMinutes
	}

	saved := 0
	for _, m := range []struct {
		key   string
		value float64
		det   map[string]any
	}{
		{model.AiMetricMaterialMissRate, pct(miss), map[string]any{"count": miss, "of": n}},
		{model.AiMetricScopeCreepRate, pct(creep), map[string]any{"count": creep, "of": n}},
		{model.AiMetricProcurementDelay, pct(procDelay), map[string]any{"count": procDelay, "of": n}},
		{model.AiMetricStopMinutesAvg, float64(totalMinutes) / float64(n), map[string]any{"totalMinutes": totalMinutes}},
	} {
		if err := save(m.key, m.value, m.det); err == nil {
			saved++
		}
	}
	return saved, nil
}
