package service

import (
	"testing"
	"time"
)

// (دازاي 11-21) الجودة: الشكوى تنتصل بسرعة وتنحل بأسرع وقت · المشاريع لكل الأدوار.
func TestComplaintPoints(t *testing.T) {
	d := func(h float64) *time.Duration { x := time.Duration(h * float64(time.Hour)); return &x }
	cases := []struct {
		took *time.Duration
		open float64
		pts  int
		ok   bool
	}{
		{d(1), 5, 2, true},   // اتصل خلال ساعة
		{d(5), 6, 1, true},   // خلال يوم
		{d(30), 40, 0, true}, // بعد يوم
		{nil, 3, 0, false},   // بعده مفتوح وما عدّى الحد — ما تنحسب
		{nil, 30, 0, true},   // مفتوح وعدّى يوم = صفر
	}
	for i, c := range cases {
		p, ok := ComplaintPoints(c.took, time.Duration(c.open*float64(time.Hour)), ComplaintContactFull, ComplaintContactHalf)
		if p != c.pts || ok != c.ok {
			t.Errorf("%d: %d %v", i, p, ok)
		}
	}
}

func TestQualityWeightsAndProjectsEverywhere(t *testing.T) {
	sum := 0.0
	for _, d := range genericParts[GroupQuality] {
		if d.Key != "PROJECTS" {
			sum += d.Weight
		}
	}
	if sum != 80 {
		t.Fatalf("أوزان الجودة (بلا المشاريع) لازم ٨٠: %v", sum)
	}
	for g, w := range roleWeightsOf {
		found := false
		for _, x := range w {
			found = found || x.Key == "PROJECTS"
		}
		if !found {
			t.Errorf("%s: ماكو بند المشاريع", g)
		}
	}
	if !RuleFitsGroup("PROJECT_WEEK", GroupQuality) || !RuleFitsGroup("PROJECT_WEEK", GroupTechs) {
		t.Fatal("المشروع لازم يتحسب بند دور لكل الفئات مو شغل إضافي")
	}
}

// (11-21) رأي ماتركس ما يتساهل ويه الغياب والمهام الفايتة حتى لو وزنهن صغير بالدور.
func TestGenericOpinionFloors(t *testing.T) {
	defs := genericParts[GroupMonitors]
	v, _ := genericOpinion(defs, map[string]float64{"ATTENDANCE": 43, "TASK_DUE": 0})
	if v > 65 {
		t.Fatalf("مراقب غايب ومفوّت مهامه أخذ رأي %v", v)
	}
	v, _ = genericOpinion(defs, map[string]float64{"ATTENDANCE": 100, "TASK_DUE": 100})
	if v != 100 {
		t.Fatalf("الكامل لازم ١٠٠: %v", v)
	}
}

// (11-21) المبيعات: تسجيل الشكوى + أوزانهم ٨٠.
func TestSalesComplaintLogAndWeights(t *testing.T) {
	if p, _ := ComplaintLogPoints(true, 6); p != 2 {
		t.Fatal(p)
	}
	if p, _ := ComplaintLogPoints(false, 6); p != 1 {
		t.Fatal(p)
	}
	if p, _ := ComplaintLogPoints(false, 2); p != 0 {
		t.Fatal(p)
	}
	sum := 0.0
	for _, d := range genericParts[GroupSales] {
		if d.Key != "PROJECTS" {
			sum += d.Weight
		}
	}
	if sum != 80 {
		t.Fatalf("أوزان المبيعات %v", sum)
	}
	if RuleFitsGroup("COMPLAINT_LOG", GroupQuality) {
		t.Fatal("تسجيل الشكوى للمبيعات بس")
	}
}
