package service

import (
	"testing"
	"time"

	"staffmange-api/internal/repository"
)

func TestStageLogPoints(t *testing.T) {
	at := func(m int) *time.Time {
		x := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC).Add(time.Duration(m) * time.Minute)
		return &x
	}
	good := repository.StageLogFact{DepartedAt: at(0), ArrivedAt: at(30), StartedAt: at(35), CompletedAt: at(120), AutoArrivedAt: at(25), AutoCompletedAt: at(125)}
	if p, _ := stageLogPoints(good); p != 2 {
		t.Fatalf("good=%d", p)
	}
	late := good
	late.ArrivedAt = at(70)
	if p, why := stageLogPoints(late); p != 1 || why == "" {
		t.Fatalf("late=%d", p)
	}
	batch := good
	batch.Batch = 3
	if p, _ := stageLogPoints(batch); p != 1 {
		t.Fatalf("batch=%d", p)
	}
	forgot := good
	forgot.AutoArrive = true
	if p, _ := stageLogPoints(forgot); p != 0 {
		t.Fatalf("forgot=%d", p)
	}
}
