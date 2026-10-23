package service

import (
	"testing"
	"time"

	"staffmange-api/internal/repository"
)

// قرار (ع) 10-07: الخطوة المتروكة (تواصل/تثبيت) تنحسب على منو ثبّت الحجز
// فعلياً — ضغط تثبيت، وإلا حط الموعد، وإلا سجّل الحجز.
func TestChainChargesEffectiveConfirmer(t *testing.T) {
	now := time.Now()
	old := now.AddDate(-1, 0, 0)
	s := &MatrixChainService{names: func(string) string { return "" }, thAt: now, thGlobal: map[string]int{},
		thSvc: map[string]map[string]int{}, contactSince: &old, contactAt: now}
	for _, d := range chainStations {
		s.thGlobal[d.Key] = 60
	}
	str := func(v string) *string { return &v }
	created := now.Add(-48 * time.Hour)
	done := now.Add(-24 * time.Hour)
	base := func() repository.ChainFacts {
		return repository.ChainFacts{ID: "b1", Code: "B1", Status: "COMPLETED", BookingType: "REGULAR",
			CreatedAt: created, CreatedByID: str("creator"), CreatedByName: str("المنشئ"), CompletedAt: &done}
	}
	owner := func(ch *BookingChain, key string) (string, string, string) {
		for _, st := range ch.Stations {
			if st.Key == key {
				id := ""
				if len(st.Owners) > 0 {
					id = st.Owners[0].ID
				}
				return id, st.Status, st.Verdict
			}
		}
		return "", "", ""
	}

	f := base()
	f.EffConfirmerID, f.EffConfirmerName, f.EffConfirmerVia = str("sched"), str("حاط الموعد"), "SCHEDULE"
	ch := s.chainOf(&f, now)
	for _, k := range []string{"CONTACT", "CONFIRM"} {
		if id, st, v := owner(ch, k); id != "sched" || st != ChainMissed {
			t.Fatalf("%s: owner=%q status=%q verdict=%q", k, id, st, v)
		}
	}

	f = base()
	f.EffConfirmerID, f.EffConfirmerName, f.EffConfirmerVia = str("creator"), str("المنشئ"), "CREATE"
	ch = s.chainOf(&f, now)
	if id, _, v := owner(ch, "CONFIRM"); id != "creator" || v == "" {
		t.Fatalf("fallback to creator: %q %q", id, v)
	}
}
