package service

import (
	"testing"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

func iptr(n int) *int         { return &n }
func fptr(f float64) *float64 { return &f }

func TestCompareInvoiceWork(t *testing.T) {
	cases := []struct {
		name  string
		f     repository.InvoiceWorkFacts
		kinds []string
	}{
		{"٨ مقابل ٤ = فرق", repository.InvoiceWorkFacts{InvoiceDeviceCount: 8, BookedDeviceCount: iptr(4)}, []string{invoiceFindingDeviceCount}},
		{"٥ مقابل ٤ = طبيعي", repository.InvoiceWorkFacts{InvoiceDeviceCount: 5, BookedDeviceCount: iptr(4)}, nil},
		{"٢ مقابل ١ = فرق واحد بس", repository.InvoiceWorkFacts{InvoiceDeviceCount: 2, BookedDeviceCount: iptr(1)}, nil},
		{"أقل بالفاتورة", repository.InvoiceWorkFacts{InvoiceDeviceCount: 2, BookedDeviceCount: iptr(6)}, []string{invoiceFindingDeviceCount}},
		{"الحجز بلا عدد", repository.InvoiceWorkFacts{InvoiceDeviceCount: 20}, nil},
		{"سعر ضعف المقدّر", repository.InvoiceWorkFacts{InvoiceNetTotal: 300000, QuotedPrice: fptr(150000)}, []string{invoiceFindingQuotedPrice}},
		{"سعر قريب", repository.InvoiceWorkFacts{InvoiceNetTotal: 180000, QuotedPrice: fptr(150000)}, nil},
		{"مجانية ما تنقارن بالسعر", repository.InvoiceWorkFacts{InvoiceNetTotal: 10, QuotedPrice: fptr(150000), IsFree: true}, nil},
		{"الفرقين", repository.InvoiceWorkFacts{InvoiceDeviceCount: 8, BookedDeviceCount: iptr(4), InvoiceNetTotal: 50000, QuotedPrice: fptr(150000)},
			[]string{invoiceFindingDeviceCount, invoiceFindingQuotedPrice}},
	}
	for _, c := range cases {
		got := compareInvoiceWork(c.f)
		if len(got) != len(c.kinds) {
			t.Errorf("%s: توقعنا %d فرق، طلع %d", c.name, len(c.kinds), len(got))
			continue
		}
		for i, k := range c.kinds {
			if got[i].Kind != k {
				t.Errorf("%s: فرق %d = %s، طلع %s", c.name, i, k, got[i].Kind)
			}
		}
	}
}

func TestJudgeInvoiceWorkMismatch(t *testing.T) {
	price := model.InvoiceWorkMismatchEvidence{Findings: []model.InvoiceWorkFinding{{Kind: invoiceFindingQuotedPrice, Invoiced: 300000, Booked: 150000}}}
	if v := judgeKind(t, model.AiSignalInvoiceWorkMismatch, price); v.Severity != model.AiSeverityWatch {
		t.Errorf("فرق سعر لحاله = WATCH، طلع %q", v.Severity)
	}
	dev := model.InvoiceWorkMismatchEvidence{BookingCode: "B-1", Findings: []model.InvoiceWorkFinding{{Kind: invoiceFindingDeviceCount, Invoiced: 8, Booked: 4}}}
	v := judgeKind(t, model.AiSignalInvoiceWorkMismatch, dev)
	if v.Severity != model.AiSeverityWarn {
		t.Errorf("فرق أجهزة = WARN، طلع %q", v.Severity)
	}
	if v.Suggestion == nil || *v.Suggestion == "" {
		t.Error("لازم اقتراح مراجعة")
	}
}

func TestDelayRiskRules(t *testing.T) {
	stats := map[string]repository.DurationStat{
		"s1|":   {ServiceID: "s1", MedianMinutes: 240, Samples: 6},
		"s1|L1": {ServiceID: "s1", LeaderID: "L1", MedianMinutes: 300, Samples: 5},
		"s1|L2": {ServiceID: "s1", LeaderID: "L2", MedianMinutes: 100, Samples: 2},
		"s2|":   {ServiceID: "s2", MedianMinutes: 240, Samples: 4},
	}
	if m, _, b, ok := pickExpected(stats, "s1", "L1"); !ok || b != "LEADER" || m != 300 {
		t.Errorf("الليدر بـ٥ عيّنات أول: %v %v %v", m, b, ok)
	}
	if m, _, b, ok := pickExpected(stats, "s1", "L2"); !ok || b != "SERVICE" || m != 240 {
		t.Errorf("الليدر بعيّنتين ← الخدمة: %v %v %v", m, b, ok)
	}
	if _, _, _, ok := pickExpected(stats, "s2", ""); ok {
		t.Error("٤ عيّنات = ولا كلمة")
	}
	// موعد ٨ مساءً بغداد، نهاية الدوام ٢٤ ← ٤ ساعات متاحة.
	start := time.Date(2026, 9, 28, 20, 0, 0, 0, debriefLoc)
	avail, by := availableWindow(start, 24, nil)
	if avail != 240 || by != "SHIFT_END" {
		t.Errorf("متاح %v %s", avail, by)
	}
	next := start.Add(2 * time.Hour)
	avail, by = availableWindow(start, 24, &next)
	if avail != 120 || by != "NEXT_BOOKING" {
		t.Errorf("الحجز الجاي يقصّر: %v %s", avail, by)
	}
	if isDelayRisk(260, 240) {
		t.Error("فرق ٢٠ دقيقة مو خطر")
	}
	if !isDelayRisk(300, 240) {
		t.Error("فرق ساعة = خطر")
	}
}

func TestBuildOpportunities(t *testing.T) {
	now := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)
	gps := "GPS"
	services := []model.Service{
		{ID: "cam", Name: "كاميرات IP"}, {ID: "net", Name: "شبكات"}, {ID: "gps", Name: "GPS", ServiceKind: &gps},
		{ID: "paint", Name: "صباغة"},
	}
	hist := []repository.CustomerFamilyHistory{
		{CustomerID: "c1", CustomerCode: 7, ServiceID: "cam", ServiceName: "كاميرات IP", LastCompleted: now.AddDate(0, -12, 0)},
		{CustomerID: "c2", CustomerCode: 8, ServiceID: "paint", ServiceName: "صباغة", LastCompleted: now.AddDate(0, -12, 0)},
		{CustomerID: "c3", CustomerCode: 9, ServiceID: "cam", ServiceName: "كاميرات IP", LastCompleted: now.AddDate(0, -12, 0), LaterBookingOf: true},
	}
	booked := map[string]map[string]bool{"c1": {"cam": true, "net": true}, "c3": {"cam": true, "net": true, "gps": true}}
	ops := buildOpportunities(hist, booked, services, now)
	count := map[string]int{}
	for _, o := range ops {
		count[o.CustomerID+"|"+o.Reason]++
		if o.CustomerID == "c1" && o.Reason == OpportunityCrossSell && o.SuggestedServiceID == "net" {
			t.Error("c1 حاجز شبكات أصلاً — ما تنقترح")
		}
		if o.CustomerCode == "" {
			t.Error("كود الزبون ناقص")
		}
	}
	if count["c1|"+OpportunityMaintenanceDue] != 1 {
		t.Error("c1 صيانة مستحقة (١٢ شهر)")
	}
	if count["c2|"+OpportunityMaintenanceDue] != 0 {
		t.Error("الصباغة ما إلها صيانة سنوية")
	}
	if count["c3|"+OpportunityMaintenanceDue] != 0 {
		t.Error("c3 حاجز بعدها — لا صيانة")
	}
	if count["c1|"+OpportunityCrossSell] != 2 { // GPS + إنذار (بلا خدمة إنذار بالجدول)
		t.Errorf("c1 بيع متقاطع = 2، طلع %d", count["c1|"+OpportunityCrossSell])
	}
	if !isMaintenanceDue(now.AddDate(0, -11, 0), now) || isMaintenanceDue(now.AddDate(0, -10, 0), now) || isMaintenanceDue(now.AddDate(0, -14, 0), now) {
		t.Error("نافذة ١١–١٣ شهر")
	}
}
