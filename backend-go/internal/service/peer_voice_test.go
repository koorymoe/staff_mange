package service

import (
	"strings"
	"testing"
	"time"
)

func TestPeerSeverity(t *testing.T) {
	if peerSeverity("ضربني بالمخزن") != "URGENT" || peerSeverity("هددني إذا ما أسكت") != "URGENT" {
		t.Fatal("كلام الخطر لازم يطلع عاجل")
	}
	if peerSeverity("يتأخر كل يوم") != "ATTENTION" {
		t.Fatal("التأخير مو عاجل")
	}
	if !strings.Contains(ruleSuggestion("هو كل يوم يتأخر علينا"), "وقت واضح") {
		t.Fatal("اقتراح التأخير")
	}
	if !strings.Contains(ruleSuggestion("ضربني"), "المدير") {
		t.Fatal("الخطر لازم يوصل للمدير")
	}
}

// أسبوع الفضفضة يبدي السبت (بغداد).
func TestPeerWeek(t *testing.T) {
	thu := time.Date(2026, 10, 8, 12, 0, 0, 0, debriefLoc) // خميس
	w := PeerWeek(thu)
	if w.Weekday() != time.Saturday || w.Day() != 3 {
		t.Fatalf("الأسبوع لازم يبدي السبت ٣، طلع %v", w)
	}
	sat := time.Date(2026, 10, 10, 9, 0, 0, 0, debriefLoc)
	if PeerWeek(sat).Day() != 10 {
		t.Fatalf("السبت يبدي أسبوع جديد، طلع %v", PeerWeek(sat))
	}
}
