package service

import (
	"encoding/json"
	"fmt"
	"log"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ العقل ═══
//
// ياخذ الأدلة ويطلّع حكماً: شنو صار، منو المسؤول، وشنو نسوي.
//
// ⚠️ اليوم يشتغل بمحرّك قواعد **حتمي** — بلا اشتراك بولا منصّة.
// صاحب العمل گال «حالياً فقط تهيكله لحد ما ننشترك بمنصّة»، وهذا
// الهيكل: `Judge` واجهة، والقواعد تنفيذ أول، والمنصّة تنفيذ ثاني
// ينضاف جنبه بلا ما ينتغيّر ولا سطر بالمنادين.
//
// ⚠️ وليش القواعد أصلاً مو ننتظر المنصّة؟ لأن نص الي طلبه صاحب العمل
// **حتمي**: «الدوام ينتهي ١٢ ليلاً» مقارنة ساعة، و«طلب مادة لو لا»
// استعلام. رميها للنموذج يعني ندفع فلوس حتى يخمّن شي نعرفه أكيد،
// ويغلط بيه أحياناً. المنصّة تجي للي القواعد ما تقدر عليه: قراءة
// نص الملاحظات، وربط أنماط متفرقة، وصياغة تقرير يقراه بني آدم.

// Judge منو يحكم على الأدلة. القواعد أو المنصّة — نفس التوقيع.
type Judge interface {
	Name() string
	Judge(signal model.AiSignal, ev model.WorkStopEvidence) (*model.AiVerdict, error)
}

type AiBrainService struct {
	repo     *repository.AiRepository
	evidence *AiEvidenceService
	// judge ينبدل بالمنصّة لما ننشترك — نقطة التوصيل الوحيدة.
	judge Judge
}

func NewAiBrainService(repo *repository.AiRepository, evidence *AiEvidenceService) *AiBrainService {
	return &AiBrainService{repo: repo, evidence: evidence, judge: RulesJudge{}}
}

// SetJudge يبدّل الحاكم — هنا تنوصل منصّة الذكاء الاصطناعي.
func (s *AiBrainService) SetJudge(j Judge) { s.judge = j }

// Process يمشي بالإشارات المعلّقة: يجمع الأدلة، بعدين يحكم.
//
// ⚠️ منفصل عن لحظة الحدث بالكامل: الموظف يضغط «توقف العمل» ويكمل
// شغله فوراً، والتحليل يصير بالخلفية. ربطهم چان يخلي بطء التحليل
// (أو فشله) يعطّل شغل ميداني.
func (s *AiBrainService) Process(limit int) (int, error) {
	signals, err := s.repo.PendingSignals(limit)
	if err != nil {
		return 0, err
	}
	done := 0
	for _, sig := range signals {
		if sig.Kind != model.AiSignalWorkStopped {
			// بقية الأنواع لسه ما إلها جامع أدلة — نأشرها متخطّاة
			// بدل ما تبقى معلّقة للأبد وتخفي الشغل الحقيقي.
			_ = s.repo.SetSignalStatus(sig.ID, "SKIPPED")
			continue
		}
		ev, err := s.evidence.CollectForWorkStop(sig)
		if err != nil {
			log.Printf("[ai] فشل جمع الأدلة لإشارة %s: %v", sig.ID, err)
			continue
		}
		_ = s.repo.SetSignalStatus(sig.ID, "COLLECTED")

		var facts model.WorkStopEvidence
		if err := json.Unmarshal(ev.Facts, &facts); err != nil {
			log.Printf("[ai] أدلة مو مقروءة لإشارة %s: %v", sig.ID, err)
			continue
		}
		verdict, err := s.judge.Judge(sig, facts)
		if err != nil || verdict == nil {
			continue
		}
		verdict.SignalID = sig.ID
		if _, err := s.repo.SaveVerdict(*verdict); err != nil {
			log.Printf("[ai] فشل حفظ الحكم لإشارة %s: %v", sig.ID, err)
			continue
		}
		_ = s.repo.SetSignalStatus(sig.ID, "ANALYZED")
		done++
	}
	return done, nil
}

// ═══ محرّك القواعد ═══
//
// كل قاعدة هنا تجاوب على سؤال طرحه صاحب العمل حرفياً.

type RulesJudge struct{}

func (RulesJudge) Name() string { return "rules-v1" }

func (RulesJudge) Judge(sig model.AiSignal, ev model.WorkStopEvidence) (*model.AiVerdict, error) {
	v := &model.AiVerdict{
		Source:     model.AiSourceRules,
		Severity:   model.AiSeverityInfo,
		Confidence: 60,
	}
	reason := ""
	suggestion := ""

	switch {
	// ═══ ١. الزبون طلب زيادة بالموقع ═══
	// «الزبون ما جان طلبها وطلبها» — الدليل: مادة انضافت للسلة **بعد**
	// ما بدأ الشغل. هذي تبرّئ الموظف، ولازم تنقال قبل أي لوم.
	case ev.CartItemsAfterStart > 0:
		v.Headline = "الزبون طلب زيادة بالموقع"
		reason = fmt.Sprintf(
			"انضافت %d مادة لسلة الزبون بعد ما بدأ الشغل — يعني الطلب توسّع بالموقع مو الموظف ناسي.",
			ev.CartItemsAfterStart)
		suggestion = "احسب الزيادة بالفاتورة، وراجع إذا الكشف الأولي كان ناقص."
		v.Confidence = 85

	// ═══ ٢. طلب المادة وما انوفّرت ═══
	// «لو إنت طلبتها وأبو الكميات ما وفّرها» — المسؤولية تنتقل.
	case ev.RequestedBeforeStop && ev.LastRequestStatus != "FULFILLED":
		v.Headline = "المادة انطلبت وما انوفّرت"
		reason = fmt.Sprintf(
			"الموظف طلب المادة قبل ما يوقّف (حالة آخر طلب: %s) — التوقف سببه التوفير مو الموظف.",
			ev.LastRequestStatus)
		suggestion = "راجع إداري الكميات: شكد ياخذ من الطلب للتوفير."
		v.Severity = model.AiSeverityWarn
		v.Confidence = 88

	// ═══ ٣. ما طلب ولا شي ═══
	// «لو إنت ناسيها» — ماكو طلب وماكو زيادة بالسلة، يبقى احتمال
	// النسيان هو الأقوى.
	// ⚠️ بس ما نجزم: نأشرها WATCH ونطلب مراجعة، ما نكتب «الموظف مقصّر».
	case ev.ProcurementRequests == 0 && ev.CartItemsAfterStart == 0:
		v.Headline = "توقف بلا طلب مادة ولا زيادة بالسلة"
		reason = "ماكو طلب مواد لهذا الحجز وماكو زيادة بالسلة بعد البداية — يحتمل نقص تحضير، ويحتمل سبب ثاني مو مسجّل."
		suggestion = "اسأل الموظف شنو نقص بالضبط، وشوف إذا الكشف الأولي كان ناقص."
		v.Severity = model.AiSeverityWatch
		v.Confidence = 55
		v.BlameEmployeeID = sig.EmployeeID

	// ═══ ٤. الدوام خالص فعلاً ═══
	// «الوقت لا يكفي» — نفحصها مقابل الساعة الحقيقية مو نقبلها.
	case ev.MinutesToShiftEnd <= 60:
		v.Headline = "توقف قرب نهاية الدوام"
		reason = fmt.Sprintf(
			"وقّف الساعة %d وباقي %d دقيقة على نهاية الدوام — التوقف منطقي.",
			ev.StoppedAtHour, ev.MinutesToShiftEnd)
		v.Confidence = 90

	default:
		v.Headline = "توقف عمل يحتاج مراجعة"
		reason = fmt.Sprintf("وقّف بعد %d دقيقة شغل، وباقي %d دقيقة على نهاية الدوام.",
			ev.WorkedMinutes, ev.MinutesToShiftEnd)
		v.Severity = model.AiSeverityWatch
	}

	// ═══ فوق كلشي: النمط ═══
	// «مرة» ظرف و«خمس مرات» نمط. هاي تتفوق على أي قاعدة فوق لأنها
	// تحچي عن الموظف مو عن الحادثة.
	if ev.StopsLast30Days >= 4 {
		v.Severity = model.AiSeverityCritical
		reason += fmt.Sprintf(" ⚠️ ونفس الموظف وقّف %d مرات بآخر ٣٠ يوم — هذا نمط مو حادثة.",
			ev.StopsLast30Days)
		v.BlameEmployeeID = sig.EmployeeID
	}

	// الفجوات تنزّل الثقة: حكم على أدلة ناقصة ما يستاهل نفس الوزن.
	v.Reasoning = &reason
	if suggestion != "" {
		v.Suggestion = &suggestion
	}
	return v, nil
}
