package service

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ٦هـ — نفس الحجز: طلب واحد بالمرة؛ حجز ثاني ما ينتظر.
func TestInvoiceBookingLockSerializes(t *testing.T) {
	var inside, maxInside int32
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unlock := invoiceBookingLock("bk-1")
			n := atomic.AddInt32(&inside, 1)
			for {
				m := atomic.LoadInt32(&maxInside)
				if n <= m || atomic.CompareAndSwapInt32(&maxInside, m, n) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			atomic.AddInt32(&inside, -1)
			unlock()
		}()
	}
	wg.Wait()
	if maxInside != 1 {
		t.Fatalf("دخل %d طلب سوه على نفس الحجز", maxInside)
	}
	done := make(chan struct{})
	unlock := invoiceBookingLock("bk-1")
	go func() { u := invoiceBookingLock("bk-2"); u(); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("حجز ثاني انتظر قفل حجز ثاني")
	}
	unlock()
}
