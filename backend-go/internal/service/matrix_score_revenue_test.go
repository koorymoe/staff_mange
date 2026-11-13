package service

import (
	"testing"
	"time"

	"staffmange-api/internal/repository"
)

// (دازاي 11-13) «ماكو حد — المهم كل موظف يدخّل فلوس أكثر من راتبه»: الي ما يشتغل
// ما يطلع عالي، والي يغطي راتبه وأكثر ياخذ زيادة؛ وكل دور من مصدر شغله.
func TestApplyRevenue(t *testing.T) {
	f := func(v float64) *float64 { return &v }
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	now := to.Add(time.Hour) // شهر كامل
	rows := []StaffScore{
		{ID: "idle", Group: GroupTechs, Final: f(95)},   // ماكو شغل ولا مشاكل
		{ID: "half", Group: GroupTechs, Final: f(80)},   // غطّى نص راتبه
		{ID: "star", Group: GroupSales, Final: f(80)},   // سجّل حجوزات دخّلت ٣ أضعاف راتبه
		{ID: "acct", Group: GroupFinance, Final: f(80)}, // المحاسب ما يدخّل فلوس مباشرة
		{ID: "techSold", Group: GroupTechs, Final: f(80)},
	}
	rev := []repository.RevenueRow{
		{EmployeeID: "half", Kind: "WORK", Amount: 500_000, Jobs: 4},
		{EmployeeID: "star", Kind: "BOOKED", Amount: 3_000_000, Jobs: 20},
		{EmployeeID: "techSold", Kind: "BOOKED", Amount: 9_000_000, Jobs: 1}, // فني سجّل حجز: مو مصدره
	}
	sal := map[string]float64{"idle": 1_000_000, "half": 1_000_000, "star": 1_000_000, "acct": 1_000_000, "techSold": 1_000_000}
	ApplyRevenue(rows, rev, sal, from, to, now)
	by := map[string]StaffScore{}
	for _, r := range rows {
		by[r.ID] = r
	}
	if *by["idle"].Final != 47.5 {
		t.Fatalf("الي ما دخّل شي ينزل للنص: %v", *by["idle"].Final)
	}
	if *by["half"].Final != 60 || *by["half"].SalaryCoverage != 50 {
		t.Fatalf("نص الراتب = ٧٥٪ من نهائيه: %v", *by["half"].Final)
	}
	if *by["star"].Final != 90 || by["star"].RevenueJobs != 20 {
		t.Fatalf("٣ أضعاف = +١٠ (السقف): %v", *by["star"].Final)
	}
	if *by["acct"].Final != 80 || by["acct"].SalaryCoverage != nil {
		t.Fatal("المحاسب ما ينحسب عليه")
	}
	if by["techSold"].Revenue != 0 {
		t.Fatal("الفني ينحسب من شغله بس، مو من الحجوزات الي سجّلها")
	}
	// نص الشهر: الراتب المتوقع نصه
	rows2 := []StaffScore{{ID: "half", Group: GroupTechs, Final: f(80)}}
	ApplyRevenue(rows2, rev, sal, from, to, from.AddDate(0, 0, 15))
	if *rows2[0].SalaryCoverage != 100 {
		t.Fatalf("بنص الشهر نص الراتب يكفي: %v", *rows2[0].SalaryCoverage)
	}
}
