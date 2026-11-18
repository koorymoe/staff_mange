package service

import "testing"

// بلاغ #50 (11-18، خيار B): الحجز ينفتح للمنسّق من التجهيز.
func TestStageUnlocksBookingFromPrepare(t *testing.T) {
	for _, s := range []string{ProjectPrepareStage, ProjectExecutionStage, ProjectHandoverStage, "✅ مكتمل"} {
		if !stageUnlocksBooking(s) {
			t.Errorf("%s should unlock", s)
		}
	}
	for _, s := range []string{"4. العقد", ProjectMediaStage, ProjectRepricingStage, "3. عرض السعر"} {
		if stageUnlocksBooking(s) {
			t.Errorf("%s should stay locked", s)
		}
	}
}
