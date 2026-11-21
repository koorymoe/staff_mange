package service

import (
	"testing"
	"time"

	"github.com/lib/pq"

	"staffmange-api/internal/repository"
)

// (دازاي 11-21) جي بي اس: «ما يكدر يسوي فاتورة ولا تقرير… المفروض يترحلن لمسؤول الخدمة».
func TestGpsPaperOnServiceManager(t *testing.T) {
	now := time.Now()
	old := now.AddDate(-1, 0, 0)
	s := &MatrixChainService{names: func(string) string { return "" }, thAt: now, thGlobal: map[string]int{},
		thSvc: map[string]map[string]int{}, contactSince: &old, contactAt: now}
	for _, d := range chainStations {
		s.thGlobal[d.Key] = 60
	}
	str := func(v string) *string { return &v }
	done := now.Add(-48 * time.Hour)
	f := repository.ChainFacts{ID: "g1", Code: "G1", Status: "COMPLETED", BookingType: "REGULAR", CreatedAt: done.Add(-time.Hour),
		CompletedAt: &done, Solo: true, LeaderID: str("tech"), LeaderName: str("الفني"), CrewIDs: pq.StringArray{"tech"},
		ServiceManagerIDs: pq.StringArray{"mgr"}}
	paper := func(ch *BookingChain) ChainStation {
		for _, st := range ch.Stations {
			if st.Key == "PAPER" {
				return st
			}
		}
		t.Fatal("ماكو محطة ورق")
		return ChainStation{}
	}
	st := paper(s.chainOf(&f, now))
	if len(st.Owners) != 1 || st.Owners[0].ID != "mgr" {
		t.Fatalf("الورق لازم على مسؤول الخدمة: %+v", st.Owners)
	}
	f.ServiceManagerIDs = nil
	if st := paper(s.chainOf(&f, now)); st.Status != ChainNA {
		t.Fatalf("بلا مسؤول: ما ينحسب على الفني — %s %+v", st.Status, st.Owners)
	}
	f.Solo = false
	if st := paper(s.chainOf(&f, now)); len(st.Owners) == 0 || st.Owners[0].ID != "tech" {
		t.Fatalf("الخدمة العادية: الورق على الليدر — %+v", st.Owners)
	}
}
