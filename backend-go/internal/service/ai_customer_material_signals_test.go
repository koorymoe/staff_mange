package service

import (
	"encoding/json"
	"testing"
	"time"

	"staffmange-api/internal/model"
)

func judgeKind(t *testing.T, kind string, facts any) *model.AiVerdict {
	t.Helper()
	emp := "emp-1"
	b, _ := json.Marshal(facts)
	v, err := RulesJudge{}.Judge(model.AiSignal{ID: "s1", Kind: kind, EmployeeID: &emp},
		model.AiEvidence{SignalID: "s1", Facts: b, Gaps: []byte(`[]`)})
	if err != nil {
		t.Fatalf("صنف %q: %v", kind, err)
	}
	if v.Source != model.AiSourceRules || v.Headline == "" {
		t.Fatalf("صنف %q: حكم ناقص", kind)
	}
	if v.BlameEmployeeID != nil {
		t.Errorf("صنف %q ما لازم يلصق تهمة تلقائية", kind)
	}
	return v
}

func TestJudgeRepeatComplaint_Severity(t *testing.T) {
	two := model.RepeatComplaintEvidence{Count: 2, WindowDays: 60, Complaints: []model.RepeatComplaintItem{{ComplaintID: "c1", BookingCode: "B-1"}, {ComplaintID: "c2"}}}
	if v := judgeKind(t, model.AiSignalCustomerRepeatComplaint, two); v.Severity != model.AiSeverityWarn {
		t.Errorf("شكويين = WARN، طلع %q", v.Severity)
	}
	three := two
	three.Count = 3
	if v := judgeKind(t, model.AiSignalCustomerRepeatComplaint, three); v.Severity != model.AiSeverityCritical {
		t.Errorf("ثلاث = CRITICAL، طلع %q", v.Severity)
	}
}

func TestJudgeMaterialUsage_Severity(t *testing.T) {
	for _, kind := range []string{model.AiSignalMaterialOveruse, model.AiSignalMaterialUnderuse} {
		ev := model.MaterialUsageEvidence{Lines: []model.MaterialUsageLine{{MaterialID: "m1", Quantity: 20, Median: 10, Samples: 6}}, SameKindLast30Days: 1}
		if v := judgeKind(t, kind, ev); v.Severity != model.AiSeverityWatch {
			t.Errorf("%s مرة وحدة = WATCH، طلع %q", kind, v.Severity)
		}
		ev.SameKindLast30Days = 3
		if v := judgeKind(t, kind, ev); v.Severity != model.AiSeverityWarn {
			t.Errorf("%s ثلاث مرات = WARN، طلع %q", kind, v.Severity)
		}
	}
}

func TestClassifyMaterialUsage(t *testing.T) {
	cases := []struct {
		qty, median float64
		samples     int
		want        string
	}{
		{16, 10, 5, "OVER"},
		{15, 10, 5, ""}, // بالضبط ١.٥× مو أكثر
		{4.9, 10, 5, "UNDER"},
		{5, 10, 5, ""},   // بالضبط ٠.٥× مو أقل
		{100, 10, 4, ""}, // عيّنات ناقصة
		{3, 0, 9, ""},
	}
	for _, c := range cases {
		if got := classifyMaterialUsage(c.qty, c.median, c.samples); got != c.want {
			t.Errorf("qty=%v median=%v samples=%d: want %q got %q", c.qty, c.median, c.samples, c.want, got)
		}
	}
	lines := []model.MaterialUsageLine{{MaterialID: "a", Quantity: 30, Median: 10, Samples: 5}, {MaterialID: "b", Quantity: 1, Median: 10, Samples: 5}}
	if o := filterMaterialLines(lines, "OVER"); len(o) != 1 || o[0].MaterialID != "a" {
		t.Errorf("فلتر OVER غلط: %+v", o)
	}
}

func TestIsoWeekMonday(t *testing.T) {
	sun := time.Date(2026, 9, 27, 10, 0, 0, 0, time.UTC) // أحد
	if got := isoWeekMonday(sun); got != "2026-09-21" {
		t.Errorf("got %s", got)
	}
	mon := time.Date(2026, 9, 28, 1, 0, 0, 0, time.UTC)
	if got := isoWeekMonday(mon); got != "2026-09-28" {
		t.Errorf("got %s", got)
	}
}
