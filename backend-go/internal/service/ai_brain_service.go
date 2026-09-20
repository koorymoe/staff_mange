package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/safeguard"
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
//
// ⚠️ التوقيع ياخذ `AiEvidence` الخام مو `WorkStopEvidence` المفكوكة،
// لسببين:
//
//	١. **أصناف الإشارات خمسة مو واحد.** تثبيت نوع أدلة واحد بالتوقيع
//	   يعني كل صنف جديد يكسر الواجهة وكل تنفيذ يتعدّل معاها.
//
//	٢. **الفجوات (`Gaps`) لازم توصل للحاكم.** مكتوب برأس
//	   schema_ai_core.go: «الفراغ المعلن أأمن من الفراغ الصامت —
//	   النموذج لازم يعرف شنو ما شافه بدل ما يفترض إنه ماكو». وهي
//	   چانت تنضيع قبل ما توصل لأن البرين يفك الحقائق بس.
//
// وكل تنفيذ يفك الحقائق حسب `signal.Kind`.
type Judge interface {
	Name() string
	Judge(signal model.AiSignal, ev model.AiEvidence) (*model.AiVerdict, error)
}

type AiBrainService struct {
	repo     *repository.AiRepository
	evidence *AiEvidenceService
	// judge ينبدل بالمنصّة لما ننشترك — نقطة التوصيل الوحيدة.
	judge Judge
	// monitor: صندوق المراقب. اختياري — نفس سبب بقية الخدمات: فشل
	// صف مراقبة ما يصير يوقف التحليل نفسه.
	monitor MonitorFeed
}

func NewAiBrainService(repo *repository.AiRepository, evidence *AiEvidenceService) *AiBrainService {
	return &AiBrainService{repo: repo, evidence: evidence, judge: RulesJudge{}}
}

// SetJudge يبدّل الحاكم — هنا تنوصل منصّة الذكاء الاصطناعي.
func (s *AiBrainService) SetJudge(j Judge) { s.judge = j }

// SetMonitorFeed يربط صندوق المراقب بعد البناء — كل حكم يوصله.
func (s *AiBrainService) SetMonitorFeed(m MonitorFeed) { s.monitor = m }

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
		ev, err := s.evidence.CollectFor(sig)
		if err != nil {
			log.Printf("[ai] فشل جمع الأدلة لإشارة %s (%s): %v", sig.ID, sig.Kind, err)
			continue
		}
		_ = s.repo.SetSignalStatus(sig.ID, "COLLECTED")

		verdict, err := s.judge.Judge(sig, *ev)
		if err != nil || verdict == nil {
			log.Printf("[ai] ماكو حكم لإشارة %s: %v", sig.ID, err)
			continue
		}
		verdict.SignalID = sig.ID
		saved, err := s.repo.SaveVerdict(*verdict)
		if err != nil {
			log.Printf("[ai] فشل حفظ الحكم لإشارة %s: %v", sig.ID, err)
			continue
		}
		_ = s.repo.SetSignalStatus(sig.ID, "ANALYZED")
		s.stageForMonitor(sig, saved)
		done++
	}
	return done, nil
}

// stageForMonitor يوصّل الحكم لصندوق المراقب — نفس الصندوق الي
// تستعمله بقية الأقسام، بمحطة جديدة `AI_VERDICT`.
//
// ⚠️ entityId معرّف **الحكم** مو الحجز: فهرس الصندوق فريد على
// (stage, entityType, entityId)، ولو استعملنا معرّف الحجز، حكم ثاني
// لنفس الحجز ينبلع بصمت (نفس حل تعديلات الفاتورة الموجود أصلاً).
func (s *AiBrainService) stageForMonitor(sig model.AiSignal, v *model.AiVerdict) {
	if s.monitor == nil || v == nil {
		return
	}
	summary := ""
	if v.Reasoning != nil {
		summary = *v.Reasoning
	}
	if v.Suggestion != nil && *v.Suggestion != "" {
		summary += " ← الإجراء المقترح: " + *v.Suggestion
	}
	ownerRole := "TECHNICIAN"
	if sig.Kind == model.AiSignalInvoiceAdjusted {
		ownerRole = "FINANCE"
	}
	urgent := v.Severity == model.AiSeverityCritical || v.Severity == model.AiSeverityWarn
	s.monitor.StageUrgent(model.MonitorStageAiVerdict, "AI_VERDICT", v.ID,
		v.Headline, summary, ownerRole, v.BlameEmployeeID, urgent)
}

// StartBackgroundLoop يخلّي النظام **يفكر لحاله**.
//
// ⚠️ قبل هذا، `Process` چانت تنشتغل بس لما أحد يضرب `POST /api/ai/process`
// بالإيد — يعني الإشارات تتكدس بالطابور وماكو أحد يشوفها. الهيكل چان
// جاهز والمحرّك مطفي.
//
// الفترة كل ١٥ دقيقة: التحليل ما يستعجل (الموظف كمّل شغله من زمان)، بس
// المراقب لازم يشوف الخلل بنفس الدوام مو باچر. والدفعة ٢٠ إشارة حتى
// طابور متراكم ينمشي على دورات بدل ما يضرب الحد اليومي بدورة وحدة.
//
// ⚠️ ولازم تمر بـ`safeguard.Loop` مو goroutine عارية: أي panic هنا
// (رد مشوّه، أدلة ناقصة) چان يقتل السيرفر كله.
func (s *AiBrainService) StartBackgroundLoop() {
	safeguard.Loop("كنسة تحليل الذكاء", 4*time.Minute, 15*time.Minute, func() {
		n, err := s.Process(20)
		if err != nil {
			log.Printf("[ai] كنسة التحليل فشلت: %v", err)
			return
		}
		if n > 0 {
			log.Printf("[ai] انحللت %d إشارة", n)
		}
	})
}

// ═══ محرّك القواعد ═══
//
// كل قاعدة هنا تجاوب على سؤال طرحه صاحب العمل حرفياً.

// escalationCleanStreakDays حد «الفترة النظيفة» — رجوع الموظف بعد
// هذا العدد من الأيام بلا نفس المخالفة يستاهل نبرة متساهلة بالسبب،
// مو تصعيداً. نصف التصعيد بالاتجاهين الثاني (التساهل)؛ التشديد
// (النمط المتكرر) موجود من زمان بعتبات كل قاعدة أدناه.
const escalationCleanStreakDays = 14

// newEmployeeGraceDays عدالة الموظف الجديد — منحنى تعلّم. موظف بأول
// ٣ أسابيع معدل أخطائه الطبيعي أعلى من موظف خبير؛ نمط يستاهل
// CRITICAL لموظف خبير يبقى WARN بس لموظف جديد (يستاهل مراجعة، مو
// إنذاراً أحمر فوري).
const newEmployeeGraceDays = 21

// capSeverityForNewEmployee يخفّف خطورة CRITICAL لـWARN لو الموظف
// بعده بفترة التجربة — نفس السبب يبقى بالنص، بس النبرة تختلف.
func capSeverityForNewEmployee(v *model.AiVerdict, reason *string, tenureDays *int) {
	if v.Severity != model.AiSeverityCritical || tenureDays == nil || *tenureDays >= newEmployeeGraceDays {
		return
	}
	v.Severity = model.AiSeverityWarn
	*reason += fmt.Sprintf(" ⚖️ بس الموظف جديد (%d يوم بالشركة) — النمط يستاهل مراجعة، مو إنذاراً حاداً لحد ما يكمّل فترة التجربة.", *tenureDays)
}

type RulesJudge struct{}

func (RulesJudge) Name() string { return "rules-v1" }

// Judge يفك الحقائق حسب صنف الإشارة، ويحوّلها لدالة القاعدة المناسبة.
// صنف ما يعرفه يرجع خطأ بدل ما يخترع حكماً.
func (r RulesJudge) Judge(sig model.AiSignal, ev model.AiEvidence) (*model.AiVerdict, error) {
	switch sig.Kind {
	case model.AiSignalWorkStopped:
		var facts model.WorkStopEvidence
		if err := json.Unmarshal(ev.Facts, &facts); err != nil {
			return nil, fmt.Errorf("أدلة مو مقروءة: %w", err)
		}
		return r.judgeWorkStop(sig, facts)
	case model.AiSignalLateStart:
		var facts model.LateStartEvidence
		if err := json.Unmarshal(ev.Facts, &facts); err != nil {
			return nil, fmt.Errorf("أدلة مو مقروءة: %w", err)
		}
		return r.judgeLateStart(sig, facts)
	case model.AiSignalRepeatPostpone:
		var facts model.RepeatPostponeEvidence
		if err := json.Unmarshal(ev.Facts, &facts); err != nil {
			return nil, fmt.Errorf("أدلة مو مقروءة: %w", err)
		}
		return r.judgeRepeatPostpone(sig, facts)
	case model.AiSignalInvoiceAdjusted:
		var facts model.InvoiceAdjustedEvidence
		if err := json.Unmarshal(ev.Facts, &facts); err != nil {
			return nil, fmt.Errorf("أدلة مو مقروءة: %w", err)
		}
		return r.judgeInvoiceAdjusted(sig, facts)
	case model.AiSignalRepeatPartial:
		var facts model.RepeatPartialEvidence
		if err := json.Unmarshal(ev.Facts, &facts); err != nil {
			return nil, fmt.Errorf("أدلة مو مقروءة: %w", err)
		}
		return r.judgeRepeatPartial(sig, facts)
	case model.AiSignalSelfReportMismatch:
		var facts model.SelfReportMismatchEvidence
		if err := json.Unmarshal(ev.Facts, &facts); err != nil {
			return nil, fmt.Errorf("أدلة مو مقروءة: %w", err)
		}
		return r.judgeSelfReportMismatch(sig, facts)
	}
	return nil, fmt.Errorf("محرّك القواعد ما يعرف صنف الإشارة %q", sig.Kind)
}

func (RulesJudge) judgeWorkStop(sig model.AiSignal, ev model.WorkStopEvidence) (*model.AiVerdict, error) {
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
	} else if ev.DaysSinceLastStop != nil && *ev.DaysSinceLastStop >= escalationCleanStreakDays {
		// ═══ التصعيد بالاتجاهين — النصف الثاني: التساهل ═══
		// صفر توقفات بآخر ٣٠ يوم وحدها ما تفرّق بين «أول مرة إطلاقاً»
		// و«رجع بعد فترة نظيفة طويلة» — والثانية تستاهل تنقال بصراحة،
		// مو بس السكوت عنها.
		reason += fmt.Sprintf(" ✅ وآخر توقف لنفس الموظف كان من %d يوم — أول ملاحظة بعد فترة نظيفة، مو نمط متكرر.",
			*ev.DaysSinceLastStop)
	}

	// عدالة الموظف الجديد — قبل التخزين، بعد كل قرارات الخطورة فوق.
	capSeverityForNewEmployee(v, &reason, ev.TenureDays)

	// الفجوات تنزّل الثقة: حكم على أدلة ناقصة ما يستاهل نفس الوزن.
	v.Reasoning = &reason
	if suggestion != "" {
		v.Suggestion = &suggestion
	}
	return v, nil
}

// judgeLateStart «تأخر بالخروج للزبون» — نفس حد DEPART بالخط الزمني.
func (RulesJudge) judgeLateStart(sig model.AiSignal, ev model.LateStartEvidence) (*model.AiVerdict, error) {
	v := &model.AiVerdict{
		Source:          model.AiSourceRules,
		Headline:        "تأخر بالخروج للزبون",
		Severity:        model.AiSeverityWatch,
		Confidence:      75,
		BlameEmployeeID: sig.EmployeeID,
	}
	reason := fmt.Sprintf("طلع للحجز متأخر %d دقيقة عن الموعد المجدول (الحد المسموح %d دقيقة).",
		ev.MinutesLate, ev.ThresholdMinutes)
	suggestion := "اسأل الفني ليش تأخر — أزمة طريق، حجز سابق طوّل، أو تأخير بلا سبب."

	// «مرة» ظرف و«أربع مرات» نمط — نفس منطق توقف العمل بالأعلى.
	if ev.LateCountLast30Days >= 4 {
		v.Severity = model.AiSeverityCritical
		v.Confidence = 90
		reason += fmt.Sprintf(" ⚠️ ونفس الموظف تأخر %d مرات بآخر ٣٠ يوم — هذا نمط مو حادثة.",
			ev.LateCountLast30Days)
	} else if ev.DaysSinceLastLate != nil && *ev.DaysSinceLastLate >= escalationCleanStreakDays {
		reason += fmt.Sprintf(" ✅ وآخر تأخر لنفس الموظف كان من %d يوم — أول ملاحظة بعد فترة نظيفة، مو نمط متكرر.",
			*ev.DaysSinceLastLate)
	}

	capSeverityForNewEmployee(v, &reason, ev.TenureDays)

	v.Reasoning = &reason
	v.Suggestion = &suggestion
	return v, nil
}

// judgeRepeatPostpone «تأجيل متكرر لنفس الحجز».
//
// ⚠️ ما نلصق التهمة بموظف: التأجيل غالباً قرار الزبون أو تنسيق داخلي
// متعثر، والدليل المتوفر ما يفرّق بينهم. «ما نعرف» أشرف من تخمين.
func (RulesJudge) judgeRepeatPostpone(sig model.AiSignal, ev model.RepeatPostponeEvidence) (*model.AiVerdict, error) {
	v := &model.AiVerdict{
		Source:     model.AiSourceRules,
		Headline:   "تأجيل متكرر لنفس الحجز",
		Severity:   model.AiSeverityWatch,
		Confidence: 60,
	}
	reason := fmt.Sprintf("هذا الحجز انأجّل %d مرة.", ev.PostponeCount)
	if ev.LastReason != "" {
		reason += " آخر سبب مسجّل: " + ev.LastReason + "."
	}
	if ev.AwaitingReschedule {
		reason += " والحجز حالياً بلا موعد بديل — الزبون نفسه ما محدّد وقته."
	}
	suggestion := "راجع مع الزبون: هل التأجيل بطلبه، أو تنسيق داخلي متعثر؟"

	if ev.PostponeCount >= 4 {
		v.Severity = model.AiSeverityWarn
		v.Confidence = 75
		reason += " ⚠️ أربع تأجيلات وأكثر لحجز وحد يستاهل مراجعة إدارية مباشرة."
	}

	v.Reasoning = &reason
	v.Suggestion = &suggestion
	return v, nil
}

// judgeInvoiceAdjusted «تعديل على فاتورة» — هذا بالضبط سؤال صاحب
// العمل: «ليش هلكد ناقص عن الفاتورة».
//
// ⚠️ النزول هو الأخطر: يعني الفاتورة الأصلية كانت زايدة عن الحقيقة —
// فلوس دخلت وما انحسبت. الزيادة (نزول مبلغ الليدر من جيبه) أهون.
func (RulesJudge) judgeInvoiceAdjusted(sig model.AiSignal, ev model.InvoiceAdjustedEvidence) (*model.AiVerdict, error) {
	v := &model.AiVerdict{
		Source:     model.AiSourceRules,
		Headline:   "تعديل على فاتورة",
		Severity:   model.AiSeverityInfo,
		Confidence: 65,
	}
	direction := "زادت"
	if ev.DifferenceAmount < 0 {
		direction = "نزلت"
	}
	reason := fmt.Sprintf("المحاسب عدّل الفاتورة: %s من %.0f إلى %.0f (فرق %.0f، %.1f٪).",
		direction, ev.OldNetTotal, ev.NewNetTotal, math.Abs(ev.DifferenceAmount), math.Abs(ev.DifferencePct))
	if ev.Reason != "" {
		reason += " السبب المسجّل: " + ev.Reason
	}
	suggestion := "راجع فاتورة الليدر الأصلية مقابل سبب التعديل المسجّل."

	if ev.DifferenceAmount < 0 {
		v.Severity = model.AiSeverityWarn
		v.Confidence = 75
		v.BlameEmployeeID = sig.EmployeeID
	}
	if ev.AdjustCountLast30Days >= 3 {
		v.Severity = model.AiSeverityCritical
		v.Confidence = 88
		v.BlameEmployeeID = sig.EmployeeID
		reason += fmt.Sprintf(" ⚠️ وفواتير نفس الليدر انعدّلت %d مرات بآخر ٣٠ يوم — نمط مو غلطة وحدة.",
			ev.AdjustCountLast30Days)
	}

	v.Reasoning = &reason
	v.Suggestion = &suggestion
	return v, nil
}

// judgeRepeatPartial «إنجاز جزئي متكرر بنفس الحجز».
//
// ⚠️ ماكو لوم تلقائي: تكرار الإنجاز الجزئي غالباً معناه الكشف الأولي
// قدّر حجم الشغل غلط، مو إن الكادر متكاسل.
func (RulesJudge) judgeRepeatPartial(sig model.AiSignal, ev model.RepeatPartialEvidence) (*model.AiVerdict, error) {
	v := &model.AiVerdict{
		Source:     model.AiSourceRules,
		Headline:   "إنجاز جزئي متكرر",
		Severity:   model.AiSeverityWatch,
		Confidence: 55,
	}
	reason := fmt.Sprintf("هذا الحجز انسجّل عليه إنجاز جزئي %d مرة. آخر نسبة إنجاز مسجّلة: %d٪.",
		ev.PartialCount, ev.LastPercentDone)
	if ev.LastBlockers != "" {
		reason += " المعوّق المسجّل آخر مرة: " + ev.LastBlockers
	}
	suggestion := "راجع مع الكادر: هل نفس المعوّق يتكرر، أو حجم الشغل فعلاً أكبر من يوم وحد؟"

	if ev.PartialCount >= 3 {
		v.Severity = model.AiSeverityWarn
		v.Confidence = 70
		reason += " ⚠️ ثلاث مرات وأكثر يستاهل مراجعة الكشف الأولي — يمكن حجم الشغل انقدّر غلط من البداية."
	}

	v.Reasoning = &reason
	v.Suggestion = &suggestion
	return v, nil
}

// judgeSelfReportMismatch «تقرير إنجاز يناقض كادر الحجز الحقيقي».
//
// 🔴 القرار المتفق عليه صراحة (بدل كشف مبالغة من أسلوب الكتابة):
// إشارة بلا لوم تلقائي — نفس فلسفة `judgeRepeatPartial`. تناقض رقمي
// (ادّعى "وحدي" والكادر أكثر من واحد) يستاهل مراجعة بشرية، مو حكماً
// جاهزاً على نية الموظف — يمكن الموظف يقصد "سويت الجزء الفني لحالي"
// وزميله سوى شي ثاني بنفس الطلعة.
func (RulesJudge) judgeSelfReportMismatch(sig model.AiSignal, ev model.SelfReportMismatchEvidence) (*model.AiVerdict, error) {
	v := &model.AiVerdict{
		Source:     model.AiSourceRules,
		Headline:   "تقرير إنجاز يناقض كادر الحجز",
		Severity:   model.AiSeverityWatch,
		Confidence: 60,
	}
	reason := fmt.Sprintf(
		"التقرير يذكر «%s» بينما كادر آخر طلعة لهذا الحجز %d أشخاص — تناقض رقمي يستاهل توضيح، مو تهمة.",
		ev.ClaimPhrase, ev.ActualCrewSize)
	if ev.BookingCode != "" {
		reason += " الحجز: " + ev.BookingCode
	}
	suggestion := "اسأل الموظف يوضّح: شنو بالضبط سواه لحاله وشنو سواه الكادر الثاني بنفس الطلعة."
	v.Reasoning = &reason
	v.Suggestion = &suggestion
	return v, nil
}
