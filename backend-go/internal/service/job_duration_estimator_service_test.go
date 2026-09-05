package service

import (
	"testing"

	"staffmange-api/internal/model"
)

// TestEstimateExpectedMinutes_NotEnoughData يتحقق من أن النظام "ما عنده رأي" فعلاً
// (nil) لو عدد العيّنات أقل من الحد الأدنى (5) — لا يجوز يخمّن رقم بأي حال.
func TestEstimateExpectedMinutes_NotEnoughData(t *testing.T) {
	samples := []model.JobDurationSample{
		{SystemName: "كاميرات انلوك", JobType: model.JobTypeInstall, ItemCount: 10, CrewSize: 2, DurationMinutes: 120},
		{SystemName: "كاميرات انلوك", JobType: model.JobTypeInstall, ItemCount: 8, CrewSize: 2, DurationMinutes: 100},
		{SystemName: "كاميرات انلوك", JobType: model.JobTypeInstall, ItemCount: 12, CrewSize: 3, DurationMinutes: 150},
		{SystemName: "كاميرات انلوك", JobType: model.JobTypeInstall, ItemCount: 6, CrewSize: 2, DurationMinutes: 90},
	}
	got, sampleCount, err := estimateExpectedMinutesFromSamples(samples, 10, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil estimate with only %d samples, got %v", len(samples), *got)
	}
	if sampleCount != 4 {
		t.Fatalf("expected sampleCount=4, got %d", sampleCount)
	}
}

// TestEstimateExpectedMinutes_HandComputedAverage يتحقق من صحة حساب المتوسط بمثال
// محسوب يدوياً بـ5 عيّنات بالضبط (الحد الأدنى):
// pace_i = duration_i * crewSize_i / itemCount_i
//
//	عيّنة 1: 100 * 2 / 10 = 20
//	عيّنة 2: 150 * 3 / 15 = 30
//	عيّنة 3: 80  * 2 / 8  = 20
//	عيّنة 4: 200 * 4 / 20 = 40
//	عيّنة 5: 90  * 2 / 9  = 20
//
// متوسط الوتيرة = (20+30+20+40+20)/5 = 130/5 = 26 دقيقة-عامل لكل عنصر.
// لوظيفة جديدة itemCount=12, crewSize=3: expected = 26*12/3 = 104 دقيقة بالضبط.
func TestEstimateExpectedMinutes_HandComputedAverage(t *testing.T) {
	samples := []model.JobDurationSample{
		{SystemName: "كاميرات IP", JobType: model.JobTypeInstall, ItemCount: 10, CrewSize: 2, DurationMinutes: 100},
		{SystemName: "كاميرات IP", JobType: model.JobTypeInstall, ItemCount: 15, CrewSize: 3, DurationMinutes: 150},
		{SystemName: "كاميرات IP", JobType: model.JobTypeInstall, ItemCount: 8, CrewSize: 2, DurationMinutes: 80},
		{SystemName: "كاميرات IP", JobType: model.JobTypeInstall, ItemCount: 20, CrewSize: 4, DurationMinutes: 200},
		{SystemName: "كاميرات IP", JobType: model.JobTypeInstall, ItemCount: 9, CrewSize: 2, DurationMinutes: 90},
	}
	got, sampleCount, err := estimateExpectedMinutesFromSamples(samples, 12, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected a computed estimate with %d samples, got nil", len(samples))
	}
	if sampleCount != 5 {
		t.Fatalf("expected sampleCount=5, got %d", sampleCount)
	}
	const want = 104.0
	if *got < want-0.0001 || *got > want+0.0001 {
		t.Fatalf("expected expectedMinutes=%.4f, got %.4f", want, *got)
	}
}

// TestEstimateExpectedMinutes_InvalidCurrentJobInputs يتحقق أنه لو الوظيفة الجديدة
// نفسها بلا itemCount/crewSize صالح، نرجع nil حتى لو عندنا عيّنات كافية — تفادياً
// لقسمة على صفر أو تقدير بلا معنى.
func TestEstimateExpectedMinutes_InvalidCurrentJobInputs(t *testing.T) {
	samples := []model.JobDurationSample{
		{SystemName: "كاميرات IP", JobType: model.JobTypeInstall, ItemCount: 10, CrewSize: 2, DurationMinutes: 100},
		{SystemName: "كاميرات IP", JobType: model.JobTypeInstall, ItemCount: 15, CrewSize: 3, DurationMinutes: 150},
		{SystemName: "كاميرات IP", JobType: model.JobTypeInstall, ItemCount: 8, CrewSize: 2, DurationMinutes: 80},
		{SystemName: "كاميرات IP", JobType: model.JobTypeInstall, ItemCount: 20, CrewSize: 4, DurationMinutes: 200},
		{SystemName: "كاميرات IP", JobType: model.JobTypeInstall, ItemCount: 9, CrewSize: 2, DurationMinutes: 90},
	}
	got, _, err := estimateExpectedMinutesFromSamples(samples, 0, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil estimate with itemCount=0, got %v", *got)
	}
}
