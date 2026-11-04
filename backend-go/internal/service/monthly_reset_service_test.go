package service

import (
	"testing"
	"time"
)

func TestResetDue(t *testing.T) {
	at := func(d, h int) time.Time { return time.Date(2026, 10, d, h, 0, 0, 0, debriefLoc) }
	cases := []struct {
		t    time.Time
		want bool
	}{{at(26, 23), false}, {at(27, 22), false}, {at(27, 23), true}, {at(28, 1), true}, {at(3, 23), false}}
	for _, c := range cases {
		if got := resetDue(c.t); got != c.want {
			t.Errorf("%v: got %v want %v", c.t, got, c.want)
		}
	}
}
