package service

import (
	"fmt"
	"log"
	"time"

	"staffmange-api/internal/repository"
	"staffmange-api/internal/timeutil"
)

// ═══ التصفير الشهري — يوم ٢٧ الساعة ١١ بالليل (بغداد) ═══
//
// قرار (ع): «كل يوم ٢٧ بالشهر ترجع كل النقاط للموظفين بس يبقى سجل…
// وساعات العمل يوم ٢٧ الساعة ١١ ليلاً».
//  • التقييمات: كل خصم فعّال يتسجّل «رجع بالتصفير الشهري» — ما ينحذف.
//  • ساعات البيت: الفترة تنسكّر، والعدّادات تبدي من صفر، والسجلات تبقى بالإكسل.
// مرة وحدة بالشهر (علامة بالقاعدة)، ولو السيرفر چان طافي وقتها يلحق أول ما يشتغل.

// (دازاي 11-27) التصفير بآخر يوم بفترة الشركة (٢٥) الساعة ١١ بالليل — چان يوم ٢٧ والفترة ٢٦←٢٥.
const (
	ResetDayOfMonth = 25
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
	// يصفّر بس الي قبل بداية الفترة الجديدة (٢٦) — خصم انسجّل يوم ٢٦ لو بعده يبقى للفترة الجديدة
	_, cutoff, _ := timeutil.PeriodStart(now.Format("2006-01"))
	evals, pts, kerr := s.kpi.ResetBefore(cutoff)
	if kerr != nil {
		log.Printf("monthly reset kpi: %v", kerr)
	}
	// (دازاي 11-26) ساعات البيت صار إلها موعد رفع خاص من إعدادات الأدمن — RunRemoteCloseIfDue
	msg := fmt.Sprintf("🔄 التصفير الشهري (%s): رجعت %d نقطة من %d تقييم للموظفين. السجلات كلها باقية.",
		now.Format("2006-01-02 15:04"), pts, evals)
	return notifyOwnerAndAdmin(s.notif, "monthly_reset", msg)
}

// RunRemoteCloseIfDue (دازاي 11-26) يرفع ساعات البيت بموعد الأدمن: الفترة الي خلصت تنسكّر
// لحد آخر يوم بيها بس — الي اشتغل بعدها (بالفترة الجديدة) يبقى مفتوح.
func (s *MonthlyResetService) RunRemoteCloseIfDue() error {
	now := time.Now().In(debriefLoc)
	c := s.remote.Config()
	cur := repository.PeriodOf(now, c.PeriodStartDay)
	m, _ := time.Parse("2006-01", cur)
	prev := m.AddDate(0, -1, 0).Format("2006-01")
	if now.Before(repository.CloseMoment(prev, c, debriefLoc)) {
		return nil
	}
	ok, err := s.aiRepo.ClaimDailyMarker("REMOTE_CLOSE", prev+"-01")
	if err != nil || !ok {
		return err
	}
	from, to := repository.PeriodRange(prev, c.PeriodStartDay)
	n, err := s.remote.CloseUpTo(to, "")
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("🏠 انرفعت ساعات البيت لفترة %s (من %s إلى %s): %d تسجيل. الفترة الجديدة بدت من %s.",
		prev, from.Format("2006-01-02"), to.Format("2006-01-02"), n, to.AddDate(0, 0, 1).Format("2006-01-02"))
	return notifyOwnerAndAdmin(s.notif, "monthly_reset", msg)
}
