package service

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"staffmange-api/internal/model"
)

// ═══ ماتركس ٩ و ١٠: صحة الزبون + هدر المواد ═══
//
// 🔴 نفس القواعد: لا غرامات ولا نقاط تلقائية، ولا أسماء أو أرقام هواتف
// بالأدلة — معرّفات وأرقام بس.

const (
	repeatComplaintWindowDays = 60
	materialMinSamples        = 5
	materialOverFactor        = 1.5
	materialUnderFactor       = 0.5
	materialLookbackDays      = 180
)

// classifyMaterialUsage يحدد اتجاه الشذوذ لسطر مادة: OVER لو الكمية
// أكثر من ١.٥× الوسيط، UNDER لو أقل من ٠.٥×، و"" غير هيچ أو لو
// العيّنات أقل من ٥.
func classifyMaterialUsage(qty, median float64, samples int) string {
	if samples < materialMinSamples || median <= 0 {
		return ""
	}
	if qty > materialOverFactor*median {
		return "OVER"
	}
	if qty < materialUnderFactor*median {
		return "UNDER"
	}
	return ""
}

// filterMaterialLines يرجّع الأسطر الي بنفس الاتجاه بس.
func filterMaterialLines(lines []model.MaterialUsageLine, direction string) []model.MaterialUsageLine {
	out := []model.MaterialUsageLine{}
	for _, l := range lines {
		if classifyMaterialUsage(l.Quantity, l.Median, l.Samples) == direction {
			out = append(out, l)
		}
	}
	return out
}

// ── مصادر البيانات (AiRepository يطبّقها) — تنفحص بـtype assertion على
// المسجّل حتى ما نحتاج أسلاك جديدة بـmain.go.

type complaintAiLookup interface {
	ComplaintLeaderID(complaintID string) (string, error)
	RepeatComplaintSnapshot(complaintID, leaderID string, windowDays int) (*model.RepeatComplaintEvidence, error)
}

type materialAiLookup interface {
	MaterialUsageBaseline(invoiceID string) (string, []model.MaterialUsageLine, error)
}

// ── الإطلاق ──

// recordRepeatComplaint يسجّل إشارة لو الزبون اشتكى ≥٢ خلال ٦٠ يوم على
// نفس الليدر. أي فشل ينطبع باللوق وبس — ما يوقف إنشاء الشكوى.
func recordRepeatComplaint(ai AiSignalRecorder, complaintID string) {
	if ai == nil {
		return
	}
	lk, ok := ai.(complaintAiLookup)
	if !ok {
		return
	}
	leaderID, err := lk.ComplaintLeaderID(complaintID)
	if err != nil {
		log.Printf("[ai] تعذر تحديد ليدر الشكوى %s: %v", complaintID, err)
		return
	}
	if leaderID == "" {
		return // ماكو ليدر ينعرف — صمت
	}
	ev, err := lk.RepeatComplaintSnapshot(complaintID, leaderID, repeatComplaintWindowDays)
	if err != nil {
		log.Printf("[ai] تعذر عد شكاوى الزبون %s: %v", complaintID, err)
		return
	}
	if ev.Count < 2 {
		return
	}
	if _, err := ai.RecordSignal(model.AiSignal{
		Kind:       model.AiSignalCustomerRepeatComplaint,
		EntityType: "COMPLAINT",
		EntityID:   complaintID,
		EmployeeID: &leaderID,
	}); err != nil {
		log.Printf("[ai] تعذر تسجيل إشارة شكاوى متكررة %s: %v", complaintID, err)
	}
}

// recordMaterialUsage يسجّل إشارة وحدة لكل اتجاه (زيادة/نقص) بالفاتورة.
// أي فشل ينطبع باللوق وبس — ما يوقف إنشاء الفاتورة.
func recordMaterialUsage(ai AiSignalRecorder, invoiceID, leaderID string) {
	if ai == nil {
		return
	}
	lk, ok := ai.(materialAiLookup)
	if !ok {
		return
	}
	_, lines, err := lk.MaterialUsageBaseline(invoiceID)
	if err != nil {
		log.Printf("[ai] تعذر حساب وسيط مواد الفاتورة %s: %v", invoiceID, err)
		return
	}
	for dir, kind := range map[string]string{"OVER": model.AiSignalMaterialOveruse, "UNDER": model.AiSignalMaterialUnderuse} {
		if len(filterMaterialLines(lines, dir)) == 0 {
			continue
		}
		emp := leaderID
		if _, err := ai.RecordSignal(model.AiSignal{
			Kind:       kind,
			EntityType: "LEADER_INVOICE",
			EntityID:   invoiceID,
			EmployeeID: &emp,
		}); err != nil {
			log.Printf("[ai] تعذر تسجيل إشارة %s للفاتورة %s: %v", kind, invoiceID, err)
		}
	}
}

// ── جامعو الأدلة ──

// CollectForRepeatComplaint — EntityID = معرّف الشكوى، EmployeeID = الليدر.
func (s *AiEvidenceService) CollectForRepeatComplaint(signal model.AiSignal) (*model.AiEvidence, error) {
	if signal.EmployeeID == nil {
		return nil, fmt.Errorf("إشارة شكاوى متكررة بلا ليدر")
	}
	ev, err := s.db.RepeatComplaintSnapshot(signal.EntityID, *signal.EmployeeID, repeatComplaintWindowDays)
	if err != nil {
		return nil, fmt.Errorf("الشكوى مو موجودة: %w", err)
	}
	gaps := []string{}
	if ev.Count < 2 {
		gaps = append(gaps, "وقت الجمع ما لكينا غير شكوى وحدة بالنافذة")
	}
	facts, _ := json.Marshal(ev)
	gapsJSON, _ := json.Marshal(gaps)
	return s.db.SaveEvidence(signal.ID, facts, gapsJSON)
}

// CollectForMaterialUsage — EntityID = معرّف فاتورة الليدر.
func (s *AiEvidenceService) CollectForMaterialUsage(signal model.AiSignal) (*model.AiEvidence, error) {
	dir := "OVER"
	if signal.Kind == model.AiSignalMaterialUnderuse {
		dir = "UNDER"
	}
	serviceID, lines, err := s.db.MaterialUsageBaseline(signal.EntityID)
	if err != nil {
		return nil, fmt.Errorf("فاتورة الليدر مو مقروءة: %w", err)
	}
	ev := model.MaterialUsageEvidence{
		Direction: dir,
		InvoiceID: signal.EntityID,
		ServiceID: serviceID,
		Lines:     filterMaterialLines(lines, dir),
	}
	gaps := []string{}
	if len(ev.Lines) == 0 {
		gaps = append(gaps, "وقت الجمع ما بقى سطر مادة يتجاوز العتبة")
	}
	if signal.EmployeeID != nil {
		if n, err := s.db.SignalCountForEmployee(signal.Kind, *signal.EmployeeID, 30); err == nil {
			ev.SameKindLast30Days = n
		} else {
			gaps = append(gaps, "ما قدرنا نقرا سجل إشارات المواد للموظف")
		}
	}
	facts, _ := json.Marshal(ev)
	gapsJSON, _ := json.Marshal(gaps)
	return s.db.SaveEvidence(signal.ID, facts, gapsJSON)
}

// ── القواعد ──

// judgeRepeatComplaint — WARN من شكويين، CRITICAL من ثلاث. بلا لوم أبعد
// من الأرقام: «يستاهل مراجعة».
func (RulesJudge) judgeRepeatComplaint(sig model.AiSignal, ev model.RepeatComplaintEvidence) (*model.AiVerdict, error) {
	v := &model.AiVerdict{
		Source:     model.AiSourceRules,
		Headline:   "شكاوى متكررة من نفس الزبون",
		Severity:   model.AiSeverityWatch,
		Confidence: 55,
	}
	if ev.Count >= 2 {
		v.Severity = model.AiSeverityWarn
		v.Confidence = 65
	}
	if ev.Count >= 3 {
		v.Severity = model.AiSeverityCritical
		v.Confidence = 75
	}
	codes := []string{}
	for _, c := range ev.Complaints {
		if c.BookingCode != "" {
			codes = append(codes, c.BookingCode)
		}
	}
	reason := fmt.Sprintf("نفس الزبون اشتكى %d مرات خلال %d يوم على شغل نفس الليدر — يستاهل مراجعة.", ev.Count, ev.WindowDays)
	if ev.CustomerCode != "" {
		reason += " الزبون: " + ev.CustomerCode + "."
	}
	if len(codes) > 0 {
		reason += " الحجوزات: " + strings.Join(codes, "، ") + "."
	}
	suggestion := "راجع الشكاوى سوية: هل نفس المشكلة تتكرر؟ ممكن السبب بالمواد أو التصميم مو بالكادر."
	v.Reasoning = &reason
	v.Suggestion = &suggestion
	return v, nil
}

// judgeMaterialUsage — WATCH لمرة وحدة، WARN لو ≥٣ إشارات نفس الصنف
// خلال ٣٠ يوم. بلا لوم: شغل أكبر/أصغر سبب مشروع.
func (RulesJudge) judgeMaterialUsage(sig model.AiSignal, ev model.MaterialUsageEvidence) (*model.AiVerdict, error) {
	over := sig.Kind == model.AiSignalMaterialOveruse
	v := &model.AiVerdict{
		Source:     model.AiSourceRules,
		Severity:   model.AiSeverityWatch,
		Confidence: 50,
	}
	word := "أعلى"
	factor := "١.٥×"
	legit := "ممكن الشغل أكبر من المعتاد (أجهزة أكثر أو مسافات أطول) — سبب مشروع."
	v.Headline = "استهلاك مواد أعلى من المعتاد"
	if !over {
		word = "أقل"
		factor = "٠.٥×"
		legit = "ممكن الشغل أصغر من المعتاد أو الزبون وفّر المواد — سبب مشروع."
		v.Headline = "استهلاك مواد أقل من المعتاد"
	}
	parts := []string{}
	for _, l := range ev.Lines {
		parts = append(parts, fmt.Sprintf("مادة %s: %.2f مقابل وسيط %.2f (%d عيّنة)", l.MaterialID, l.Quantity, l.Median, l.Samples))
	}
	reason := fmt.Sprintf("%d سطر مادة %s من %s وسيط فواتير نفس الخدمة: %s. %s",
		len(ev.Lines), word, factor, strings.Join(parts, "؛ "), legit)
	if ev.SameKindLast30Days >= 3 {
		v.Severity = model.AiSeverityWarn
		v.Confidence = 65
		reason += fmt.Sprintf(" تكرر %d مرات خلال ٣٠ يوم — يستاهل مراجعة.", ev.SameKindLast30Days)
	}
	suggestion := "قارن الكميات بحجم الشغل الفعلي بالحجز قبل أي استنتاج."
	v.Reasoning = &reason
	v.Suggestion = &suggestion
	return v, nil
}
