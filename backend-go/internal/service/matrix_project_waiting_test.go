package service

import (
	"testing"
	"time"

	"staffmange-api/internal/repository"
)

// بلاغ #41 (11-14): مشروع واكف بانتظار الزبون ما يتحاسب عليه المسؤول؛ وعرض السعر/العقد حدّهن ٣٠ يوم.
func TestProjectWaitingCustomer(t *testing.T) {
	now := time.Now()
	own := "e1"
	since := now.AddDate(0, 0, -31)
	s := &MatrixProjectService{}
	base := repository.ProjectFacts{ID: "p", Name: "x", Stage: "3. عرض السعر", ResponsibleID: &own, CreatedAt: since, UpdatedAt: since, StageSince: &since}

	v := s.judge(base, map[string]int{"3. عرض السعر": 15}, now)
	if v.StageLimit != CustomerStageMinDays {
		t.Fatalf("حد عرض السعر = %d", v.StageLimit)
	}
	if v.Status == ChainNA || v.Status == ChainOK {
		t.Fatalf("بلا انتظار لازم يطلع متأخر: %s", v.Status)
	}

	w := now.AddDate(0, 0, -10)
	note := "ينتظر يوقّع العقد"
	f := base
	f.WaitingSince, f.WaitingNote = &w, &note
	if v := s.judge(f, nil, now); v.Status != ChainNA {
		t.Fatalf("بانتظار الزبون لازم ما ينحسب: %s %s", v.Status, v.Verdict)
	}
	if _, ok := StationPoints(ChainNA); ok {
		t.Fatal("ChainNA لازم ما ينطي نقاط")
	}

	// الانتظار قبل ما تتغيّر المرحلة ينشال وحده
	old := since.AddDate(0, 0, -5)
	f.WaitingSince = &old
	if v := s.judge(f, nil, now); v.Status == ChainNA {
		t.Fatal("انتظار من مرحلة قديمة ما يغطي المرحلة الجديدة")
	}

	// المرحلة الي مو بإيد الزبون حدّها مثل ما هو
	g := base
	g.Stage = "6. البدء بالتنفيذ"
	if v := s.judge(g, map[string]int{"6. البدء بالتنفيذ": 10}, now); v.StageLimit != 10 {
		t.Fatalf("حد التنفيذ تغيّر: %d", v.StageLimit)
	}
}
