package service

import "testing"

func TestNormTrainingWindow(t *testing.T) {
	s := func(v string) *string { return &v }
	if f, to, err := normTrainingWindow(s("10:00"), s("11:30"), nil, nil); err != nil || *f != "10:00" || *to != "11:30" {
		t.Fatalf("valid rejected: %v", err)
	}
	if f, to, err := normTrainingWindow(s(""), s(""), s("10:00"), s("11:00")); err != nil || f != nil || to != nil {
		t.Fatal("clear failed")
	}
	if _, _, err := normTrainingWindow(s("12:00"), nil, s("10:00"), s("11:00")); err == nil {
		t.Fatal("start after end accepted")
	}
	for _, bad := range []string{"9:00", "25:00", "10:00x"} {
		if _, _, err := normTrainingWindow(s(bad), s("23:00"), nil, nil); err == nil {
			t.Fatalf("%q accepted", bad)
		}
	}
	if _, _, err := normTrainingWindow(s("10:00"), nil, nil, nil); err == nil {
		t.Fatal("half window accepted")
	}
}
