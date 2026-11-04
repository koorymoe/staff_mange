package service

import (
	"testing"

	"staffmange-api/internal/model"
)

func TestSuggestAchievementReply(t *testing.T) {
	code := "B12"
	sec := 30
	long := "اليوم نصبت جهازين جي بي اس بحجز الزبون وفحصت الإشارة وسلمت الفاتورة"
	cases := []struct {
		a    model.Achievement
		want string
	}{
		{model.Achievement{ReportText: "تم"}, model.AchievementReviewNeeds},
		{model.Achievement{ReportText: "", VoiceSeconds: &sec}, model.AchievementReviewNeeds},
		{model.Achievement{ReportText: long, BookingCode: &code}, model.AchievementReviewGood},
		{model.Achievement{ReportText: long}, model.AchievementReviewGood},
	}
	for i, c := range cases {
		if got, reply := SuggestAchievementReply(c.a); got != c.want || reply == "" {
			t.Errorf("case %d: got %q %q", i, got, reply)
		}
	}
}
