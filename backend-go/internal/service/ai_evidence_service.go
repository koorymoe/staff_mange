package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ جامع الأدلة ═══
//
// هذا الي وصفه صاحب العمل حرفياً:
//
//   «مو يجي الموظف يسوي توقف العمل… النظام يضل يفكر بالموضوع: يراجع
//    الساعة — إحنا دوامنا ينتهي ١٢ ليلاً ويبدي ٩ صباحاً. أو مادة
//    ناقصة — ليش ناقصة؟ الزبون ما جان طلبها وطلبها، لو إنت ناسيها،
//    لو إنت طلبتها وأبو الكميات ما وفّرها؟ يروح يشوف سلة الزبون…
//    نوب يروح لأبو الكميات… ترجع تشوف شوكت انضافت المادة».
//
// ⚠️ هذا الملف **ما يحكم على أحد**. شغله ينتزع الحقائق بس:
//   طلب مادة؟ متى؟ انوفّرت؟ السلة زادت بعد ما بدأ؟ الدوام خالص؟
// الحكم بملف ثاني، والسبب بالتفصيل برأس schema_ai_core.go.
//
// وليش نجمعها بأنفسنا بدل ما نرمي كلشي للنموذج؟ لأن النموذج بلا
// حقائق يخمّن بثقة — والتخمين هنا يتحول لظلم موظف. الحقائق تنعرض
// للمالك حتى لو النموذج غلط بتفسيرها.

type AiEvidenceService struct {
	db           *repository.AiRepository
	bookings     *repository.BookingRepository
	invoices     *repository.LeaderInvoiceRepository
	progress     *repository.BookingProgressRepository
	achievements *repository.AchievementRepository
	employees    *repository.EmployeeRepository
}

func NewAiEvidenceService(
	db *repository.AiRepository,
	bookings *repository.BookingRepository,
	invoices *repository.LeaderInvoiceRepository,
	progress *repository.BookingProgressRepository,
	achievements *repository.AchievementRepository,
	employees *repository.EmployeeRepository,
) *AiEvidenceService {
	return &AiEvidenceService{db: db, bookings: bookings, invoices: invoices, progress: progress, achievements: achievements, employees: employees}
}

// tenureDaysFor عدالة الموظف الجديد — شكد يوم مرّ من تاريخ تعيينه.
// nil يعني ماكو تاريخ تعيين مسجّل، مو صفر (بلا تخمين).
func (s *AiEvidenceService) tenureDaysFor(employeeID string) *int {
	emp, err := s.employees.FindByID(employeeID)
	if err != nil || emp == nil || emp.HireDate == nil {
		return nil
	}
	days := int(time.Since(*emp.HireDate).Hours() / 24)
	if days < 0 {
		days = 0
	}
	return &days
}

// CollectForWorkStop يجمع أدلة توقف العمل — المسار الي وصفه صاحب العمل.
//
// ⚠️ كل خطوة ممكن تفشل بلا ما توقف الباقي: نسجّل الفجوة بـgaps ونكمل.
// أدلة ناقصة معلنة أحسن من ولا أدلة — والنموذج لازم يعرف شنو ما شافه
// بدل ما يفترض إنه ماكو.
func (s *AiEvidenceService) CollectForWorkStop(signal model.AiSignal) (*model.AiEvidence, error) {
	ev := model.WorkStopEvidence{}
	gaps := []string{}

	booking, err := s.bookings.FindByID(signal.EntityID)
	if err != nil || booking == nil {
		return nil, fmt.Errorf("الحجز مو موجود")
	}

	if booking.WorkStopReason != nil {
		ev.StopReason = *booking.WorkStopReason
	} else {
		gaps = append(gaps, "ما كتب سبب التوقف")
	}

	// ═══ الساعة ═══
	// «الوقت لا يكفي» تنفحص مقابل نهاية الدوام الحقيقية، مو تنقبل
	// مثل ما هي.
	win, err := s.db.WorkWindow()
	if err != nil {
		gaps = append(gaps, "ما قدرنا نقرا ساعات الدوام")
		win = &model.AiWorkWindow{StartHour: 9, EndHour: 24}
	}
	stoppedAt := signal.OccurredAt
	if booking.WorkStoppedAt != nil {
		stoppedAt = *booking.WorkStoppedAt
	}
	// بغداد = UTC+3. الحساب بالتوقيت المحلي وإلا «الساعة ١١ ليلاً»
	// تنقرا ٨ مساءً وتنقلب النتيجة.
	baghdad := stoppedAt.UTC().Add(3 * time.Hour)
	ev.StoppedAtHour = baghdad.Hour()
	ev.MinutesToShiftEnd = (win.EndHour * 60) - (baghdad.Hour()*60 + baghdad.Minute())
	if ev.MinutesToShiftEnd < 0 {
		ev.MinutesToShiftEnd = 0
	}

	if booking.StartedAt != nil {
		ev.WorkedMinutes = int(stoppedAt.Sub(*booking.StartedAt).Minutes())
		if ev.WorkedMinutes < 0 {
			ev.WorkedMinutes = 0
		}
	} else {
		gaps = append(gaps, "ماكو وقت بداية شغل مسجّل")
	}

	// ═══ خيط المواد: هل طلب من إداري الكميات؟ ═══
	reqs, status, err := s.db.ProcurementSummary(booking.ID, stoppedAt)
	if err != nil {
		gaps = append(gaps, "ما قدرنا نقرا طلبات المواد")
	} else {
		ev.ProcurementRequests = reqs.Total
		ev.RequestedBeforeStop = reqs.BeforeStop > 0
		ev.LastRequestStatus = status
	}

	// ═══ خيط السلة: الزبون طلب زيادة بالموقع؟ ═══
	total, afterStart, err := s.db.CartSummary(booking.ID, booking.StartedAt)
	if err != nil {
		gaps = append(gaps, "ما قدرنا نقرا سلة الزبون")
	} else {
		ev.CartItemsTotal = total
		ev.CartItemsAfterStart = afterStart
	}

	// ═══ سجل الموظف: مرة ولا نمط؟ ═══
	if signal.EmployeeID != nil {
		n, err := s.db.StopCountForEmployee(*signal.EmployeeID, 30)
		if err != nil {
			gaps = append(gaps, "ما قدرنا نقرا سجل توقفات الموظف")
		} else {
			ev.StopsLast30Days = n
		}
		// ⚠️ نصف التصعيد الثاني (التساهل) — بأفضل جهد: فشلها ما يوقف
		// التحليل، بس بلا ملاحظة «فترة نظيفة» بالحكم.
		if days, err := s.db.DaysSinceLastSignal(model.AiSignalWorkStopped, *signal.EmployeeID, signal.OccurredAt); err == nil {
			ev.DaysSinceLastStop = days
		}
		ev.TenureDays = s.tenureDaysFor(*signal.EmployeeID)
	} else {
		gaps = append(gaps, "الإشارة بلا موظف")
	}

	facts, _ := json.Marshal(ev)
	gapsJSON, _ := json.Marshal(gaps)
	return s.db.SaveEvidence(signal.ID, facts, gapsJSON)
}

// selfReportSoloPhrases الجمل الصريحة الي تعتبر ادّعاء عمل لحاله —
// نص حرفي بس، بلا أي تفسير لأسلوب الكتابة.
var selfReportSoloPhrases = []string{"وحدي", "لحالي", "بروحي", "بمفردي"}

// detectSoloClaim يدوّر أول جملة صريحة بالنص. فاضي يعني ماكو ادّعاء —
// التقرير ما يذكر شي عن العدد، وهذا مختلف عن الادّعاء الصريح.
func detectSoloClaim(text string) string {
	for _, phrase := range selfReportSoloPhrases {
		if strings.Contains(text, phrase) {
			return phrase
		}
	}
	return ""
}

// CollectForSelfReportMismatch أدلة تناقض التقرير الذاتي — الموظف
// ادّعى عملاً لحاله بتقرير إنجاز مربوط بحجز، وكادر آخر طلعة أكثر من
// واحد. EntityID هنا معرّف **الإنجاز** (`Achievement.ID`) مو الحجز.
func (s *AiEvidenceService) CollectForSelfReportMismatch(signal model.AiSignal) (*model.AiEvidence, error) {
	ev := model.SelfReportMismatchEvidence{}
	gaps := []string{}

	a, err := s.achievements.FindByID(signal.EntityID)
	if err != nil || a == nil || a.BookingID == nil {
		return nil, fmt.Errorf("الإنجاز أو ربطه بالحجز مو موجود")
	}
	ev.ReportText = a.ReportText
	ev.ClaimPhrase = detectSoloClaim(a.ReportText)
	if ev.ClaimPhrase == "" {
		gaps = append(gaps, "ماكو جملة ادّعاء صريحة بالتقرير وقت الجمع")
	}

	crewSize, err := s.db.LatestVisitCrewSize(*a.BookingID)
	if err != nil {
		gaps = append(gaps, "ما قدرنا نقرا كادر آخر طلعة لهذا الحجز")
	} else {
		ev.ActualCrewSize = crewSize
	}
	if a.BookingCode != nil {
		ev.BookingCode = *a.BookingCode
	}

	facts, _ := json.Marshal(ev)
	gapsJSON, _ := json.Marshal(gaps)
	return s.db.SaveEvidence(signal.ID, facts, gapsJSON)
}

// CollectForFuelAnomaly أدلة شذوذ تعبئة وقود — EntityID هنا معرّف
// سجل السيارة (`VehicleLog.ID`) مو المركبة.
func (s *AiEvidenceService) CollectForFuelAnomaly(signal model.AiSignal) (*model.AiEvidence, error) {
	ev, err := s.db.FuelAnomalySnapshot(signal.EntityID)
	if err != nil {
		return nil, fmt.Errorf("سجل تعبئة الوقود مو موجود: %w", err)
	}
	gaps := []string{}
	if ev.AverageCost == 0 {
		gaps = append(gaps, "ماكو تاريخ تعبئات كافٍ لهذي المركبة وقت الجمع")
	}
	facts, _ := json.Marshal(ev)
	gapsJSON, _ := json.Marshal(gaps)
	return s.db.SaveEvidence(signal.ID, facts, gapsJSON)
}

// CollectFor يوزّع على الجامع الصحيح حسب صنف الإشارة. نقطة دخول واحدة
// حتى البرين ما يحتاج يعرف تفاصيل كل صنف.
func (s *AiEvidenceService) CollectFor(signal model.AiSignal) (*model.AiEvidence, error) {
	switch signal.Kind {
	case model.AiSignalWorkStopped:
		return s.CollectForWorkStop(signal)
	case model.AiSignalLateStart:
		return s.CollectForLateStart(signal)
	case model.AiSignalRepeatPostpone:
		return s.CollectForRepeatPostpone(signal)
	case model.AiSignalInvoiceAdjusted:
		return s.CollectForInvoiceAdjusted(signal)
	case model.AiSignalRepeatPartial:
		return s.CollectForRepeatPartial(signal)
	case model.AiSignalSelfReportMismatch:
		return s.CollectForSelfReportMismatch(signal)
	case model.AiSignalFuelAnomaly:
		return s.CollectForFuelAnomaly(signal)
	}
	return nil, fmt.Errorf("ماكو جامع أدلة لصنف %q", signal.Kind)
}

// CollectForLateStart أدلة تأخر الخروج للحجز.
//
// ⚠️ نفس حساب «تأخر الخروج» بجدول الخط الزمني (DEPART) — هنا يوصل
// للتحليل بدل ما يبقى رقماً بشاشة وحدها.
func (s *AiEvidenceService) CollectForLateStart(signal model.AiSignal) (*model.AiEvidence, error) {
	ev := model.LateStartEvidence{ThresholdMinutes: model.DelayDepartMinutes}
	gaps := []string{}

	booking, err := s.bookings.FindByID(signal.EntityID)
	if err != nil || booking == nil {
		return nil, fmt.Errorf("الحجز مو موجود")
	}

	if booking.ScheduledAt != nil && booking.StartedAt != nil {
		late := int(booking.StartedAt.Sub(*booking.ScheduledAt).Minutes())
		if late < 0 {
			// طلع قبل موعده — مو تأخير، نصفّرها بدل رقم سالب يربك القارئ.
			late = 0
		}
		ev.MinutesLate = late
	} else {
		gaps = append(gaps, "ماكو موعد مجدول أو وقت بداية شغل مسجّل")
	}

	if signal.EmployeeID != nil {
		n, err := s.db.SignalCountForEmployee(model.AiSignalLateStart, *signal.EmployeeID, 30)
		if err != nil {
			gaps = append(gaps, "ما قدرنا نقرا سجل تأخر الموظف")
		} else {
			ev.LateCountLast30Days = n
		}
		if days, err := s.db.DaysSinceLastSignal(model.AiSignalLateStart, *signal.EmployeeID, signal.OccurredAt); err == nil {
			ev.DaysSinceLastLate = days
		}
		ev.TenureDays = s.tenureDaysFor(*signal.EmployeeID)
	} else {
		gaps = append(gaps, "الإشارة بلا موظف")
	}

	facts, _ := json.Marshal(ev)
	gapsJSON, _ := json.Marshal(gaps)
	return s.db.SaveEvidence(signal.ID, facts, gapsJSON)
}

// CollectForRepeatPostpone أدلة تأجيل نفس الحجز أكثر من مرة.
//
// ⚠️ ماكو عدّاد «موظف» هنا عمداً: التأجيل غالباً قرار الزبون أو إداري
// الحجوزات مو الكادر الميداني — الحكم يصير من السبب المكتوب مو من لصق
// تهمة بمسجّل الحركة.
func (s *AiEvidenceService) CollectForRepeatPostpone(signal model.AiSignal) (*model.AiEvidence, error) {
	ev := model.RepeatPostponeEvidence{}
	gaps := []string{}

	booking, err := s.bookings.FindByID(signal.EntityID)
	if err != nil || booking == nil {
		return nil, fmt.Errorf("الحجز مو موجود")
	}
	ev.PostponeCount = booking.PostponeCount
	ev.AwaitingReschedule = booking.AwaitingReschedule
	if booking.PostponeReason != nil {
		ev.LastReason = *booking.PostponeReason
	} else {
		gaps = append(gaps, "ماكو سبب تأجيل مسجّل لآخر مرة")
	}

	facts, _ := json.Marshal(ev)
	gapsJSON, _ := json.Marshal(gaps)
	return s.db.SaveEvidence(signal.ID, facts, gapsJSON)
}

// CollectForInvoiceAdjusted أدلة تعديل مبالغ فاتورة بعد ما انسجّلت.
//
// ⚠️ EntityID هنا معرّف **الفاتورة** مو الحجز — الإشارة تنسجّل وقت
// التعديل، والمعرّف بالإشارة هو الليدر صاحب الفاتورة (شوف تعليق
// InvoiceAdjustedEvidence بالموديل).
func (s *AiEvidenceService) CollectForInvoiceAdjusted(signal model.AiSignal) (*model.AiEvidence, error) {
	ev := model.InvoiceAdjustedEvidence{}
	gaps := []string{}

	adjustments, err := s.invoices.Adjustments(signal.EntityID)
	if err != nil || len(adjustments) == 0 {
		return nil, fmt.Errorf("سجل تعديل الفاتورة مو موجود")
	}
	last := adjustments[0] // الأحدث أول (ORDER BY createdAt DESC)
	ev.OldNetTotal = last.OldNetTotal
	ev.NewNetTotal = last.NewNetTotal
	ev.DifferenceAmount = last.NewNetTotal - last.OldNetTotal
	if last.OldNetTotal != 0 {
		ev.DifferencePct = (ev.DifferenceAmount / last.OldNetTotal) * 100
	} else {
		gaps = append(gaps, "المبلغ الأصلي كان صفر — النسبة ما تنحسب")
	}
	ev.Reason = last.Reason

	if signal.EmployeeID != nil {
		n, err := s.db.SignalCountForEmployee(model.AiSignalInvoiceAdjusted, *signal.EmployeeID, 30)
		if err != nil {
			gaps = append(gaps, "ما قدرنا نقرا سجل تعديلات فواتير هذا الليدر")
		} else {
			ev.AdjustCountLast30Days = n
		}
	} else {
		gaps = append(gaps, "الإشارة بلا موظف")
	}

	facts, _ := json.Marshal(ev)
	gapsJSON, _ := json.Marshal(gaps)
	return s.db.SaveEvidence(signal.ID, facts, gapsJSON)
}

// CollectForRepeatPartial أدلة إنجاز جزئي متكرر بنفس الحجز.
func (s *AiEvidenceService) CollectForRepeatPartial(signal model.AiSignal) (*model.AiEvidence, error) {
	ev := model.RepeatPartialEvidence{}
	gaps := []string{}

	booking, err := s.bookings.FindByID(signal.EntityID)
	if err != nil || booking == nil {
		return nil, fmt.Errorf("الحجز مو موجود")
	}
	ev.PartialCount = booking.PartialCount

	reports, err := s.progress.Reports(booking.ID)
	if err != nil {
		gaps = append(gaps, "ما قدرنا نقرا تقارير الإنجاز الجزئي")
	} else if len(reports) == 0 {
		gaps = append(gaps, "ماكو تقارير إنجاز مسجّلة رغم عدّاد الإنجاز الجزئي")
	} else {
		last := reports[len(reports)-1] // ASC حسب رقم اليوم — آخر عنصر هو الأحدث
		ev.LastPercentDone = last.PercentDone
		ev.LastRemaining = last.RemainingWork
		if last.Blockers != nil {
			ev.LastBlockers = *last.Blockers
		}
	}

	facts, _ := json.Marshal(ev)
	gapsJSON, _ := json.Marshal(gaps)
	return s.db.SaveEvidence(signal.ID, facts, gapsJSON)
}
