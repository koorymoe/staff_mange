package safeguard

import (
	"sync/atomic"
	"testing"
	"time"
)

// الاختبار نفسه هو الدليل: لو الحماية ما اشتغلت، الـpanic يقتل عملية
// الاختبار كلها ويطلع فشل — مو مجرد تأكيد على قيمة.
func TestRunTrapsPanic(t *testing.T) {
	ok := Run("مهمة تنهار", func() { panic("انفجار مقصود") })
	if ok {
		t.Fatal("لازم يرجع false لمن تنهار المهمة")
	}
}

func TestRunReportsSuccess(t *testing.T) {
	called := false
	if ok := Run("مهمة سليمة", func() { called = true }); !ok {
		t.Fatal("لازم يرجع true لمن تخلص بسلام")
	}
	if !called {
		t.Fatal("ما انستدعت الدالة أصلاً")
	}
}

// الأهم: الانهيار بدورة ما يوگف الحلقة. قبل التصليح، أول panic كان
// يقتل السيرفر — يعني ما اكو دورة ثانية إطلاقاً.
func TestLoopSurvivesPanic(t *testing.T) {
	var runs int64
	Loop("حلقة تنهار كل مرة", time.Millisecond, 5*time.Millisecond, func() {
		atomic.AddInt64(&runs, 1)
		panic("انهيار متكرر")
	})

	deadline := time.After(500 * time.Millisecond)
	for {
		select {
		case <-deadline:
			t.Fatalf("الحلقة وگفت بعد %d دورة — الحماية ما اشتغلت", atomic.LoadInt64(&runs))
		default:
			if atomic.LoadInt64(&runs) >= 3 {
				return // اشتغلت ثلاث دورات رغم الانهيار بكل وحدة
			}
			time.Sleep(2 * time.Millisecond)
		}
	}
}
