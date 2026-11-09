package safeguard

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// الطابور الثگيل: وحدة وحدة، والفشل ما يعلّگ الطابور.
func TestQueue_HeavyOneAtATime(t *testing.T) {
	var cur, peak int64
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			runQueued("heavy-test", time.Hour, func() {
				n := atomic.AddInt64(&cur, 1)
				for {
					p := atomic.LoadInt64(&peak)
					if n <= p || atomic.CompareAndSwapInt64(&peak, p, n) {
						break
					}
				}
				time.Sleep(20 * time.Millisecond)
				atomic.AddInt64(&cur, -1)
				if i == 0 {
					panic("boom")
				}
			})
		}(i)
	}
	wg.Wait()
	if peak != int64(cap(heavySlots)) {
		t.Fatalf("peak=%d want %d", peak, cap(heavySlots))
	}
	for _, s := range Jobs() {
		if s.Name == "heavy-test" && (s.Runs != 4 || s.Failures != 1 || s.Running) {
			t.Fatalf("stats %+v", s)
		}
	}
}

// الأولوية للموظفين: المهمة الثگيلة تنطر لحد ما تخف الطلبات.
func TestQueue_YieldsToUsers(t *testing.T) {
	for i := int64(0); i <= yieldInFlight; i++ {
		RequestStarted()
	}
	started := make(chan time.Time, 1)
	go runQueued("yield-test", time.Hour, func() { started <- time.Now() })
	time.Sleep(400 * time.Millisecond)
	select {
	case <-started:
		t.Fatal("اشتغلت والسيرفر مشغول")
	default:
	}
	released := time.Now()
	for i := int64(0); i <= yieldInFlight; i++ {
		RequestDone()
	}
	if at := <-started; at.Before(released) {
		t.Fatal("اشتغلت قبل ما تخف الطلبات")
	}
}
