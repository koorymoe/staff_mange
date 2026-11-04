package timeutil

import (
	"testing"
	"time"
)

func TestTodayIsBaghdad(t *testing.T) {
	want := time.Now().In(companyLoc).Format("2006-01-02")
	if Today() != want || ThisMonth() != want[:7] {
		t.Fatalf("Today=%s ThisMonth=%s want %s", Today(), ThisMonth(), want)
	}
}
