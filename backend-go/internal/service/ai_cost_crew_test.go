package service

import (
	"math"
	"testing"

	"staffmange-api/internal/repository"
)

// usageCost: كلفة الذكاء بالدولار — الكاش قراءة ١٠٪ وكتابة ١٢٥٪ من سعر الإدخال.
func TestUsageCost(t *testing.T) {
	r := repository.AiUsageRow{Model: "claude-sonnet-4-6", InputTokens: 1_000_000, OutputTokens: 1_000_000, CacheRead: 1_000_000, CacheWrite: 1_000_000}
	want := 3 + 15 + 0.3 + 3.75
	if got := usageCost(r); math.Abs(got-want) > 1e-9 {
		t.Fatalf("%v، المتوقع %v", got, want)
	}
	if usageCost(repository.AiUsageRow{Model: "x"}) != 0 {
		t.Error("بلا توكنات = صفر")
	}
	// الأحدث قبل العام: opus-5-5 مو سعر opus القديم
	if priceOf("claude-opus-5-5") == priceOf("claude-opus-4-1") {
		t.Error("opus-5-5 لازم إله سعر خاص")
	}
}

// scoreCrew: المهارة ترفع، والمشاكل والحمل بنفس اليوم تنزّل، والخبرة لها سقف.
func TestScoreCrewOrdering(t *testing.T) {
	base := repository.CrewCandidateRow{ServiceHasSkills: true, HasSkill: true, DoneCount: 10}
	skilled := scoreCrew(base).Score
	noSkill := base
	noSkill.HasSkill = false
	if !(scoreCrew(noSkill).Score < skilled) {
		t.Error("بلا مهارة لازم أقل")
	}
	busy := base
	busy.DayLoad = 2
	if !(scoreCrew(busy).Score < skilled) {
		t.Error("المشغول بنفس اليوم لازم أقل")
	}
	trouble := base
	trouble.ProblemCount = 5
	s := scoreCrew(trouble)
	if !(s.Score < skilled) || s.ProblemRatePct == nil || *s.ProblemRatePct != 50 {
		t.Errorf("نسبة المشاكل: %+v", s)
	}
	huge := base
	huge.DoneCount = 100000
	capped := base
	capped.DoneCount = crewDoneCap
	if scoreCrew(huge).Score != scoreCrew(capped).Score {
		t.Error("الخبرة لازم إلها سقف")
	}
}

func TestCurveScore(t *testing.T) {
	if got := curveScore(CurveWeek{Completed: 10, Stops: 1, Complaints: 1}); got != 5 {
		t.Fatalf("%d، المتوقع 5 (10 - 2×1 - 3×1)", got)
	}
}
