package service

import (
	"fmt"
	"log"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ التصفير الشهري — يوم ٢٧ الساعة ١١ بالليل (بغداد) ═══
//
// قرار (ع): «كل يوم ٢٧ بالشهر ترجع كل النقاط للموظفين بس يبقى سجل…
// وساعات العمل يوم ٢٧ الساعة ١١ ليلاً».
//  • التقييمات: كل خصم فعّال يتسجّل «رجع بالتصفير الشهري» — ما ينحذف.
//  • ساعات البيت: الفترة تنسكّر، والعدّادات تبدي من صفر، والسجلات تبقى بالإكسل.
// مرة وحدة بالشهر (علامة بالقاعدة)، ولو السيرفر چان طافي وقتها يلحق أول ما يشتغل.

const (
	ResetDayOfMonth = 27
	resetHour       = 23
)

type MonthlyResetService struct {
	kpi    *repository.KpiRepository
	remote *repository.RemoteHoursRepository
	aiRepo *repository.AiRepository
	notif  *repository.NotificationRepository
}

func NewMonthlyResetService(kpi *repository.KpiRepository, remote *repository.RemoteHoursRepository,
	aiRepo *repository.AiRepository, notif *repository.NotificationRepository) *MonthlyResetService {
	return &MonthlyResetService{kpi: kpi, remote: remote, aiRepo: aiRepo, notif: notif}
}

// due هل وصلنا موعد تصفير هالشهر؟
func resetDue(now time.Time) bool {
	return now.Day() > ResetDayOfMonth || (now.Day() == ResetDayOfMonth && now.Hour() >= resetHour)
}

func (s *MonthlyResetService) RunIfDue() error {
	now := time.Now().In(debriefLoc)
	if !resetDue(now) {
		return nil
	}
	ok, err := s.aiRepo.ClaimDailyMarker("MONTHLY_RESET", now.Format("2006-01")+"-01")
	if err != nil || !ok {
		return err
	}
	evals, pts, kerr := s.kpi.ResetAllActive()
	if kerr != nil {
		log.Printf("monthly reset kpi: %v", kerr)
	}
	hours, herr := s.remote.CloseAll("")
	if herr != nil {
		log.Printf("monthly reset remote hours: %v", herr)
	}
	msg := fmt.Sprintf("🔄 التصفير الشهري (%s): رجعت %d نقطة من %d تقييم للموظفين، وتصفّرت ساعات البيت (%d تسجيل). السجلات كلها باقية.",
		now.Format("2006-01-02 15:04"), pts, evals, hours)
	return notifyOwnerAndAdmin(s.notif, "monthly_reset", msg)
}
