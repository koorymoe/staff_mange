package service

import (
	"encoding/json"
	"testing"

	"staffmange-api/internal/model"
)

// ═══ محرّك القواعد — الأصناف الأربعة الجديدة ═══
//
// اختبار سطحي بس مهم: يثبت إن كل صنف يتفكّ ويرجّع حكماً حقيقياً
// (مو خطأ «صنف ما نعرفه») — الغلطة السهلة بعد إضافة صنف جديد هي
// نسيانه بسويتش RulesJudge.Judge ويرجع يظهر كـ«صنف مو معروف».

func TestRulesJudge_KnowsAllFourNewKinds(t *testing.T) {
	empID := "emp-1"
	cases := []struct {
		kind  string
		facts any
	}{
		{model.AiSignalLateStart, model.LateStartEvidence{MinutesLate: 90, ThresholdMinutes: 60}},
		{model.AiSignalRepeatPostpone, model.RepeatPostponeEvidence{PostponeCount: 2}},
		{model.AiSignalInvoiceAdjusted, model.InvoiceAdjustedEvidence{OldNetTotal: 100000, NewNetTotal: 90000, DifferenceAmount: -10000}},
		{model.AiSignalRepeatPartial, model.RepeatPartialEvidence{PartialCount: 2, LastPercentDone: 60}},
	}
	for _, c := range cases {
		facts, _ := json.Marshal(c.facts)
		sig := model.AiSignal{ID: "sig-1", Kind: c.kind, EmployeeID: &empID}
		ev := model.AiEvidence{SignalID: "sig-1", Facts: facts, Gaps: []byte(`[]`)}

		v, err := RulesJudge{}.Judge(sig, ev)
		if err != nil {
			t.Fatalf("صنف %q لازم يعرفه محرّك القواعد، طلع خطأ: %v", c.kind, err)
		}
		if v == nil || v.Headline == "" {
			t.Fatalf("صنف %q رجّع حكماً فارغاً", c.kind)
		}
		if v.Source != model.AiSourceRules {
			t.Errorf("صنف %q: المصدر لازم RULES، طلع %q", c.kind, v.Source)
		}
	}
}

// تعديل فاتورة بالنزول (الليدر بالغ بالتقدير) يلصق التهمة به —
// هذا بالضبط سؤال صاحب العمل: «ليش هلكد ناقص عن الفاتورة».
func TestJudgeInvoiceAdjusted_BlamesOnlyOnDecrease(t *testing.T) {
	empID := "emp-leader"

	// نزول: يلصق التهمة
	down := model.InvoiceAdjustedEvidence{OldNetTotal: 200000, NewNetTotal: 150000, DifferenceAmount: -50000}
	sig := model.AiSignal{ID: "s1", Kind: model.AiSignalInvoiceAdjusted, EmployeeID: &empID}
	v, err := RulesJudge{}.judgeInvoiceAdjusted(sig, down)
	if err != nil {
		t.Fatalf("خطأ غير متوقع: %v", err)
	}
	if v.BlameEmployeeID == nil || *v.BlameEmployeeID != empID {
		t.Error("النزول لازم يلصق التهمة بالليدر — هذا سؤال صاحب العمل نفسه")
	}
	if v.Severity != model.AiSeverityWarn {
		t.Errorf("النزول لازم يطلع WARN على الأقل، طلع %q", v.Severity)
	}

	// زيادة: ما يلصق تهمة تلقائية
	up := model.InvoiceAdjustedEvidence{OldNetTotal: 150000, NewNetTotal: 200000, DifferenceAmount: 50000}
	v2, err := RulesJudge{}.judgeInvoiceAdjusted(sig, up)
	if err != nil {
		t.Fatalf("خطأ غير متوقع: %v", err)
	}
	if v2.BlameEmployeeID != nil {
		t.Error("الزيادة ما تستاهل تهمة تلقائية — الليدر نقّص من جيبه")
	}
}

// نمط متكرر (تأخر أو تعديل فواتير) يرفع الخطورة لـCRITICAL — نفس
// منطق «مرة ظرف، خمس مرات نمط» الموجود لتوقف العمل.
func TestJudgeLateStart_PatternEscalatesToCritical(t *testing.T) {
	empID := "emp-1"
	sig := model.AiSignal{ID: "s1", Kind: model.AiSignalLateStart, EmployeeID: &empID}

	once := model.LateStartEvidence{MinutesLate: 90, ThresholdMinutes: 60, LateCountLast30Days: 1}
	v, _ := RulesJudge{}.judgeLateStart(sig, once)
	if v.Severity == model.AiSeverityCritical {
		t.Error("مرة وحدة ما تستاهل CRITICAL")
	}

	pattern := model.LateStartEvidence{MinutesLate: 90, ThresholdMinutes: 60, LateCountLast30Days: 5}
	v2, _ := RulesJudge{}.judgeLateStart(sig, pattern)
	if v2.Severity != model.AiSeverityCritical {
		t.Errorf("خمس مرات بآخر ٣٠ يوم لازم CRITICAL، طلع %q", v2.Severity)
	}
}

// جامع الأدلة يوزّع على الصنف الصحيح، ويرفض صنف ما يعرفه بدل ما
// يرجّع أدلة فاضية بصمت.
func TestAiEvidenceService_CollectFor_UnknownKindErrors(t *testing.T) {
	s := &AiEvidenceService{}
	_, err := s.CollectFor(model.AiSignal{Kind: "SOME_FUTURE_SIGNAL_KIND"})
	if err == nil {
		t.Fatal("صنف ما نعرفه لازم يرجع خطأ")
	}
}
