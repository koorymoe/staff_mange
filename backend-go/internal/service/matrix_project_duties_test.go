package service

import (
	"testing"

	"staffmange-api/internal/repository"
)

func TestWatchGroupTechs(t *testing.T) {
	if g := WatchGroup(repository.WatchSubject{Role: "SERVICE_MANAGER"}); g != "STAFF" {
		t.Fatalf("service manager → %s", g)
	}
	if g := WatchGroup(repository.WatchSubject{Role: "MEDIA"}); g != "MEDIA" {
		t.Fatalf("media → %s", g)
	}
}

func TestProjectWatch(t *testing.T) {
	ps := groupProjectDuties([]ProjectVerdict{
		{OwnerID: "a", Status: ChainOK},
		{OwnerID: "a", Status: ChainLate, Verdict: "متأخر"},
		{OwnerID: "a", Status: ChainMissed, Verdict: "واگف"},
		{OwnerID: "a", Status: ChainNA},
		{OwnerID: "", Status: ChainLate},
	})
	items, w := projectWatch(ps["a"])
	if len(items) != 2 || w.Done != 1 || w.Left != 2 {
		t.Fatalf("items=%d done=%d left=%d", len(items), w.Done, w.Left)
	}
	if items[0].Escalated || !items[1].Escalated {
		t.Fatal("MISSED لازم يتصعّد")
	}
	if i, w := projectWatch(nil); i != nil || w != nil {
		t.Fatal("بلا مشاريع لازم فارغ")
	}
}
