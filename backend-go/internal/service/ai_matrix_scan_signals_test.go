package service

import (
	"database/sql"
	"strings"
	"testing"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// كلمات لوم ممنوعة بأحكام هالإشارات.
var blameWords = []string{"مقصّر", "مخالفة", "غرامة", "خصم", "عقوبة للموظف", "كذب", "مهمل"}

func assertNoBlame(t *testing.T, v *model.AiVerdict) {
	t.Helper()
	if v.BlameEmployeeID != nil {
		t.Fatalf("ما لازم ينسب لوم: %v", *v.BlameEmployeeID)
	}
	txt := v.Headline + *v.Reasoning + *v.Suggestion
	for _, w := range blameWords {
		if strings.Contains(txt, w) {
			t.Fatalf("كلمة لوم %q بالنص: %s", w, txt)
		}
	}
}

func TestCustomerRiskFactors(t *testing.T) {
	r := repository.CustomerRiskRow{MaxPostpone: 1, PostponeCode: "B1"}
	if n := len(customerRiskFactors(r)); n != 0 {
		t.Fatalf("تأجيل وحد مو عامل: %d", n)
	}
	r = repository.CustomerRiskRow{MaxPostpone: 2, PostponeCode: "B1", OverdueDays: 3, OverdueCode: "B2",
		OpenComplaints: 1, LowRatings: 1, MinRating: 2, LeaderID: sql.NullString{}}
	if n := len(customerRiskFactors(r)); n != 4 {
		t.Fatalf("متوقع ٤ عوامل: %d", n)
	}
	r = repository.CustomerRiskRow{LowRatings: 1, MinRating: 3}
	if n := len(customerRiskFactors(r)); n != 0 {
		t.Fatalf("تقييم ٣ مو واطي: %d", n)
	}
}

func TestJudgeCustomerAtRisk(t *testing.T) {
	f := model.CustomerRiskFactor{Kind: riskFactorPostponed, Value: 2, Ref: "B1"}
	g := model.CustomerRiskFactor{Kind: riskFactorOpenComplaint, Value: 1}
	h := model.CustomerRiskFactor{Kind: riskFactorLowRating, Value: 1}
	for _, c := range []struct {
		fs  []model.CustomerRiskFactor
		sev string
	}{{[]model.CustomerRiskFactor{f, g}, model.AiSeverityWarn}, {[]model.CustomerRiskFactor{f, g, h}, model.AiSeverityCritical}} {
		v, _ := RulesJudge{}.judgeCustomerAtRisk(model.AiSignal{}, model.CustomerAtRiskEvidence{Factors: c.fs})
		if v.Severity != c.sev {
			t.Fatalf("%d عوامل: %s", len(c.fs), v.Severity)
		}
		assertNoBlame(t, v)
		if !strings.Contains(*v.Suggestion, "مهندس الجودة") {
			t.Fatal("الاقتراح لازم يوجّه لمهندس الجودة")
		}
	}
}

func TestClassifyPriceOutlier(t *testing.T) {
	cases := []struct {
		net, med float64
		n        int
		free     bool
		want     string
	}{
		{250, 100, 8, false, "HIGH"},
		{200, 100, 8, false, ""},
		{39, 100, 8, false, "LOW"},
		{40, 100, 8, false, ""},
		{250, 100, 7, false, ""},
		{250, 100, 8, true, ""},
		{0, 100, 8, false, ""},
	}
	for _, c := range cases {
		if got := classifyPriceOutlier(c.net, c.med, c.n, c.free); got != c.want {
			t.Fatalf("%+v: %q", c, got)
		}
	}
}

func TestJudgePriceOutlier(t *testing.T) {
	v, _ := RulesJudge{}.judgePriceOutlier(model.AiSignal{}, model.PriceOutlierEvidence{NetTotal: 300, Median: 100, Samples: 9, Ratio: 3, Direction: "HIGH"})
	if v.Severity != model.AiSeverityWatch {
		t.Fatalf("متوقع WATCH: %s", v.Severity)
	}
	assertNoBlame(t, v)
}

func TestJudgeLatePaperwork(t *testing.T) {
	items := []model.LatePaperworkItem{{BookingCode: "B1", MissingInvoice: true}, {BookingCode: "B2", MissingReport: true}, {BookingCode: "B3", MissingReport: true}}
	for n, want := range map[int]string{1: model.AiSeverityWatch, 2: model.AiSeverityWatch, 3: model.AiSeverityWarn} {
		v, _ := RulesJudge{}.judgeLatePaperwork(model.AiSignal{}, model.LatePaperworkEvidence{LateCount: n, Bookings: items[:n]})
		if v.Severity != want {
			t.Fatalf("%d حجز: %s", n, v.Severity)
		}
		assertNoBlame(t, v)
		if !strings.Contains(*v.Reasoning, "B1") {
			t.Fatal("لازم يذكر أكواد الحجوزات")
		}
	}
}

func TestGroupLatePaperwork(t *testing.T) {
	rows := []repository.LatePaperworkRow{
		{LeaderID: "L1", LatePaperworkItem: model.LatePaperworkItem{BookingCode: "B1"}},
		{LeaderID: "L1", LatePaperworkItem: model.LatePaperworkItem{BookingCode: "B2"}},
		{LeaderID: "L2", LatePaperworkItem: model.LatePaperworkItem{BookingCode: "B3"}},
		{LeaderID: "", LatePaperworkItem: model.LatePaperworkItem{BookingCode: "B4"}},
	}
	g := groupLatePaperwork(rows)
	if len(g) != 2 || len(g["L1"]) != 2 || len(g["L2"]) != 1 {
		t.Fatalf("تجميع غلط: %+v", g)
	}
}

func TestJudgeAttendanceWorkGap(t *testing.T) {
	v, _ := RulesJudge{}.judgeAttendanceWorkGap(model.AiSignal{}, model.AttendanceWorkGapEvidence{Day: "2026-09-27", Kind: gapPresentNoWork, AttendedMinutes: 400})
	if v.Severity != model.AiSeverityInfo {
		t.Fatalf("متوقع INFO: %s", v.Severity)
	}
	w, _ := RulesJudge{}.judgeAttendanceWorkGap(model.AiSignal{}, model.AttendanceWorkGapEvidence{Day: "2026-09-27", Kind: gapWorkNoAttendance, BookingCodes: []string{"B1"}})
	if w.Severity != model.AiSeverityWatch {
		t.Fatalf("متوقع WATCH: %s", w.Severity)
	}
	for _, x := range []*model.AiVerdict{v, w} {
		assertNoBlame(t, x)
		if !strings.Contains(*x.Reasoning, "نقص تسجيل") || !strings.Contains(*x.Reasoning, "مو اتهام") {
			t.Fatal("لازم يوضّح إنه ممكن نقص تسجيل مو اتهام")
		}
	}
}

func TestBaghdadMidnightAndWeek(t *testing.T) {
	m := baghdadMidnight("2026-09-28")
	if m.IsZero() || m.UTC().Hour() != 21 {
		t.Fatalf("منتصف ليل بغداد = ٢١ UTC: %v", m.UTC())
	}
	if isoWeekMonday(baghdadMidnight("2026-10-04")) != "2026-09-28" {
		t.Fatal("الأحد يرجع لاثنين نفس الأسبوع")
	}
}
