package service

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"staffmange-api/internal/model"
)

// حاكم بديل يسجّل إذا انستدعى — حتى نثبت الرجوع صار فعلاً.
type spyJudge struct {
	called bool
}

func (s *spyJudge) Name() string { return "spy" }
func (s *spyJudge) Judge(model.AiSignal, model.AiEvidence) (*model.AiVerdict, error) {
	s.called = true
	return &model.AiVerdict{Source: model.AiSourceRules, Headline: "حكم القواعد"}, nil
}

func workStopSignal() (model.AiSignal, model.AiEvidence) {
	empID := "emp-1"
	sig := model.AiSignal{
		ID:         "sig-1",
		Kind:       model.AiSignalWorkStopped,
		EntityType: "BOOKING",
		EntityID:   "bk-1",
		EmployeeID: &empID,
		OccurredAt: time.Date(2026, 9, 17, 14, 30, 0, 0, time.UTC),
	}
	facts, _ := json.Marshal(model.WorkStopEvidence{
		StopReason: "مادة ناقصة", StoppedAtHour: 14, MinutesToShiftEnd: 600,
		WorkedMinutes: 20, ProcurementRequests: 0,
	})
	gaps, _ := json.Marshal([]string{"ما قدرنا نقرا سلة الزبون"})
	return sig, model.AiEvidence{SignalID: "sig-1", Facts: facts, Gaps: gaps}
}

// ⚠️ أهم اختبار بالملف: الحد اليومي يخلص → يرجع للقواعد بهدوء.
// بدونه، خلوص الحد يعني إشارات بلا حكم ومراقب يشوف صندوقاً فاضي.
func TestModelJudge_QuotaExhausted_FallsBackToRules(t *testing.T) {
	spy := &spyJudge{}
	j := NewModelJudge("sk-fake", "claude-haiku-4-5", 0, spy) // سقف صفر = الحد خالص
	sig, ev := workStopSignal()

	v, err := j.Judge(sig, ev)
	if err != nil {
		t.Fatalf("لازم يرجع حكم القواعد بلا خطأ، طلع: %v", err)
	}
	if !spy.called {
		t.Fatal("ما رجع للقواعد — الحد خلص ومع ذلك حاول ينادي النموذج")
	}
	if v.Source != model.AiSourceRules {
		t.Fatalf("مصدر الحكم لازم RULES، طلع %q", v.Source)
	}
}

// النموذج ما يكدر يلصق التهمة بموظف ما موجود بالإشارة — المعرّف
// ينحط من الإشارة حصراً. هاي تمنع تهمة على بريء بسبب هلوسة.
func TestToVerdict_BlameComesFromSignalOnly(t *testing.T) {
	j := NewModelJudge("sk-fake", "claude-haiku-4-5", 10, RulesJudge{})
	sig, _ := workStopSignal()

	withBlame := j.toVerdict(sig, modelVerdictOut{
		Headline: "عنوان", Severity: model.AiSeverityWarn, Confidence: 80,
		BlameOnSignalEmployee: true,
	})
	if withBlame.BlameEmployeeID == nil || *withBlame.BlameEmployeeID != "emp-1" {
		t.Fatalf("لازم ياخذ معرّف الإشارة، طلع %v", withBlame.BlameEmployeeID)
	}

	noBlame := j.toVerdict(sig, modelVerdictOut{
		Headline: "عنوان", Severity: model.AiSeverityInfo, Confidence: 40,
		BlameOnSignalEmployee: false,
	})
	if noBlame.BlameEmployeeID != nil {
		t.Fatalf("«ماكو مسؤول واضح» جواب مشروع — ما لازم ينلصق أحد، طلع %v", *noBlame.BlameEmployeeID)
	}
}

// خطورة مخترعة تكسر ترتيب صندوق المراقب. والافتراضي WATCH مو INFO:
// رد مشوّه لازم ينشاف مو ينخفي.
func TestNormalizeSeverity_RejectsInvented(t *testing.T) {
	cases := map[string]string{
		model.AiSeverityInfo:     model.AiSeverityInfo,
		"warn":                   model.AiSeverityWarn,
		"  critical  ":           model.AiSeverityCritical,
		"URGENT":                 model.AiSeverityWatch, // مخترعة
		"":                       model.AiSeverityWatch,
		"DROP TABLE Employee":    model.AiSeverityWatch,
		model.AiSeverityCritical: model.AiSeverityCritical,
	}
	for in, want := range cases {
		if got := normalizeSeverity(in); got != want {
			t.Errorf("normalizeSeverity(%q) = %q، المتوقع %q", in, got, want)
		}
	}
}

// ثقة ١٥٠ تعدي عتبة AiConfidenceTrusted وتخلي حكماً مشكوكاً يظهر كحقيقة.
func TestToVerdict_ClampsConfidence(t *testing.T) {
	j := NewModelJudge("sk-fake", "claude-haiku-4-5", 10, RulesJudge{})
	sig, _ := workStopSignal()
	for _, tc := range []struct{ in, want int }{{150, 100}, {-20, 0}, {75, 75}} {
		got := j.toVerdict(sig, modelVerdictOut{Headline: "x", Confidence: tc.in}).Confidence
		if got != tc.want {
			t.Errorf("ثقة %d صارت %d، المتوقع %d", tc.in, got, tc.want)
		}
	}
}

// القص بالبايت يكسر الحرف العربي بالنص ويطلّع رموزاً مشوّهة.
func TestClampText_DoesNotBreakArabic(t *testing.T) {
	long := strings.Repeat("مرحبا", 100) // ٥٠٠ حرف عربي
	got := clampText(long, 10)
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("لازم ينقص وينضاف «…»، طلع %q", got)
	}
	if strings.ContainsRune(got, '�') {
		t.Fatal("انكسر حرف عربي بالقص")
	}
	if n := len([]rune(strings.TrimSuffix(got, "…"))); n > 10 {
		t.Fatalf("القص بالأحرف مو بالبايت: طلع %d حرف", n)
	}
	if clampText("  قصير  ", 50) != "قصير" {
		t.Fatal("النص القصير لازم يرجع بلا فراغات ولا قص")
	}
}

// ⚠️⚠️ **أهم اختبار خصوصية بالمشروع**: ولا اسم زبون ولا هاتف ولا اسم
// موظف يطلع للمزوّد الخارجي. الأدلة أرقام وحالات — تنفهم بلا أسماء.
func TestBuildEvidencePrompt_LeaksNoIdentity(t *testing.T) {
	empName := "يوسف احمد"
	sig, ev := workStopSignal()
	sig.EmployeeName = &empName

	prompt := buildEvidencePrompt(sig, ev)

	for _, secret := range []string{"يوسف احمد", "07701101274", "emp-1"} {
		if strings.Contains(prompt, secret) {
			t.Errorf("تسرّبت هوية للمزوّد الخارجي: %q موجودة بالنص", secret)
		}
	}
	// وبنفس الوقت لازم الفجوات توصل — «الفراغ المعلن أأمن من الصامت».
	if !strings.Contains(prompt, "ما قدرنا نقرا سلة الزبون") {
		t.Error("الفجوات ما وصلت للنموذج")
	}
	if !strings.Contains(prompt, "توقف عمل") {
		t.Error("صنف الإشارة ما وصل")
	}
}

// محرّك القواعد ما يخترع حكماً لصنف ما يعرفه — يرجع خطأ فينسجّل.
func TestRulesJudge_UnknownKindReturnsError(t *testing.T) {
	sig, ev := workStopSignal()
	sig.Kind = model.AiSignalInvoiceAdjusted

	v, err := RulesJudge{}.Judge(sig, ev)
	if err == nil {
		t.Fatal("لازم يرجع خطأ لصنف ما يعرفه بدل ما يخترع حكماً")
	}
	if v != nil {
		t.Fatalf("ما لازم يرجع حكماً، طلع %+v", v)
	}
}

// الأدلة المشوّهة ما تسقّط النظام — ترجع خطأ ينسجّل.
func TestRulesJudge_MalformedFacts(t *testing.T) {
	sig, _ := workStopSignal()
	ev := model.AiEvidence{SignalID: "sig-1", Facts: []byte("{ مو JSON")}

	if _, err := (RulesJudge{}).Judge(sig, ev); err == nil {
		t.Fatal("أدلة مشوّهة لازم ترجع خطأ")
	}
}
