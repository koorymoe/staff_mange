package service

import "testing"

func TestFinePointsRange(t *testing.T) {
	cases := []struct {
		late      bool
		emp, cust int
		want      int
	}{
		{false, 0, 0, 3}, // أبسط شكوى = ٣٠ ألف
		{true, 0, 0, 5},  // ما انحلت خلال ٤٨ ساعة
		{false, 1, 0, 5}, // شكوى ثانية عليه
		{false, 5, 0, 7}, // التكرار سقفه +٤
		{true, 5, 3, 10}, // كلشي = السقف ١٠٠ ألف
	}
	for _, c := range cases {
		got, why := FinePoints(c.late, c.emp, c.cust)
		if got != c.want || len(why) == 0 {
			t.Errorf("FinePoints(%v,%d,%d)=%d want %d", c.late, c.emp, c.cust, got, c.want)
		}
		if got < orgFineMinPoints || got > orgFineMaxPoints {
			t.Errorf("خارج المدى: %d", got)
		}
	}
}

func TestOrgScoreWeights(t *testing.T) {
	f := func(v float64) *float64 { return &v }
	if s := OrgScore(OrgEvalComponents{Staff: f(40)}); s == nil || *s != 40 {
		t.Fatalf("العضو = تقييمه: %v", s)
	}
	if s := OrgScore(OrgEvalComponents{Staff: nil}); s != nil {
		t.Fatal("بلا بيانات لازم nil (ما نحكم)")
	}
	// المدير: ٦٠٪ تقييم + ٢٥٪ خطة + ١٥٪ ذكاء
	if s := OrgScore(OrgEvalComponents{Director: true, Staff: f(50), Plan: f(100), AIUsage: f(0)}); s == nil || *s != 55 {
		t.Fatalf("وزن المدير: %v", *s)
	}
	// بلا تقييم ماتركس ما نحكم — حتى المدير (ما ننحكم على الخطة والذكاء لحالهم)
	if s := OrgScore(OrgEvalComponents{Director: true, Plan: f(40), AIUsage: f(100)}); s != nil {
		t.Fatalf("المدير بلا تقييم لازم nil: %v", *s)
	}
	if s := OrgScore(OrgEvalComponents{Director: true, Staff: f(80)}); s != nil {
		t.Fatal("المدير بلا خطة/ذكاء محسوبين لازم nil")
	}
}
