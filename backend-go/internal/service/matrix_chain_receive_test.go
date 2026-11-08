package service

import (
	"testing"
	"time"

	"staffmange-api/internal/repository"
)

// (ك) 11-08: بلاغ #21 — كادر انحدد قبل الموعد بأيام ما يطلّع «تجهيز المواد» متأخر.
func TestReceiveStart_EarlyCrewCountsFromAppointmentDay(t *testing.T) {
	loc := BaghdadLoc()
	at := func(d, h, m int) *time.Time { x := time.Date(2026, 11, d, h, m, 0, 0, loc).UTC(); return &x }

	// نفس حجز محمد فاضل: كادر 05/11 19:06، موعد 08/11 08:00، مواد جاهزة 08:47.
	f := repository.ChainFacts{FirstAssignAt: at(5, 19, 6), ScheduledAt: at(8, 8, 0), MaterialsAt: at(8, 8, 47)}
	if got := minsBetween(receiveStart(&f, 9), f.MaterialsAt); got == nil || *got != 47 {
		t.Fatalf("الموعد قبل بداية الدوام: العدّ لازم يبدي من الموعد (47 د)، طلع %v", got)
	}
	if got := minsBetween(receiveStart(&f, 7), f.MaterialsAt); got == nil || *got != 107 {
		t.Fatalf("دوام ٧ الصبح: لازم 107 د، طلع %v", got)
	}

	// الكادر انحدد بنفس يوم الموعد بعد بداية الدوام — الحساب ما يتغير.
	same := repository.ChainFacts{FirstAssignAt: at(8, 10, 0), ScheduledAt: at(8, 14, 0), MaterialsAt: at(8, 11, 0)}
	if got := receiveStart(&same, 9); !got.Equal(*same.FirstAssignAt) {
		t.Fatalf("نفس اليوم: البداية لازم تبقى وقت تحديد الكادر، طلع %v", got)
	}

	// بلا موعد — يبقى على وقت تحديد الكادر.
	noSch := repository.ChainFacts{FirstAssignAt: at(5, 19, 6)}
	if got := receiveStart(&noSch, 9); !got.Equal(*noSch.FirstAssignAt) {
		t.Fatalf("بلا موعد: لازم تبقى وقت تحديد الكادر، طلع %v", got)
	}
}
