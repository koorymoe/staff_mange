package handler

import (
	"testing"
	"time"
)

// (11-07) الساعة ١:٣٠ فجر ١٠ تشرين الثاني ببغداد = ٢٢:٣٠ يوم ٩ بـUTC.
// چان «اليوم» ينحسب ٩، فإجازة تبدي ١١ (إشعار يوم واحد) تنقبل.
func TestLeaveMinStartUsesBaghdadDate(t *testing.T) {
	now := time.Date(2026, 11, 9, 22, 30, 0, 0, time.UTC)
	if got := leaveMinStart(now).Format("2006-01-02"); got != "2026-11-12" {
		t.Fatalf("أقرب بداية %s، المفروض 2026-11-12", got)
	}
}
