package service

import (
	"fmt"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ الفضفضة اليومية (ماتركس) ═══
//
// طلب صاحب النظام: يجي كل يوم يفضفضله بملخص شنو صار — بلا ما يفتح
// شاشة ولا يدوّر. القرار الفعلي (سليم/ملاحظة) يبقى بصندوق المراقب
// الموجود — هذا تذكير وملخص بس، ما يبني آلية قرار موازية.

var debriefLoc = func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Baghdad"); err == nil {
		return loc
	}
	return time.FixedZone("+03", 3*60*60)
}()

// ٨ مساءً بغداد — بعد نهاية دوام العمل الاعتيادي.
const dailyDebriefHour = 20

type DailyDebriefService struct {
	aiRepo    *repository.AiRepository
	notifRepo *repository.NotificationRepository
}

func NewDailyDebriefService(aiRepo *repository.AiRepository, notifRepo *repository.NotificationRepository) *DailyDebriefService {
	return &DailyDebriefService{aiRepo: aiRepo, notifRepo: notifRepo}
}

// RunIfDue تُستدعى دورياً من الحلقة الخلفية. ما تسوي شي إلا لمن الوقت
// يدخل نافذة الساعة المحددة **ولسه ما انرسلت اليوم** — العلامة
// (`ClaimDailyMarker`) تمنع إرسالها مرتين لو الحلقة دارت أكثر من مرة
// بنفس النافذة.
func (s *DailyDebriefService) RunIfDue() error {
	now := time.Now().In(debriefLoc)
	if now.Hour() < dailyDebriefHour {
		return nil
	}
	day := now.Format("2006-01-02")
	claimed, err := s.aiRepo.ClaimDailyMarker("DAILY_DEBRIEF_SENT", day)
	if err != nil || !claimed {
		return err
	}

	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, debriefLoc)
	dayEnd := dayStart.Add(24 * time.Hour)
	counts, err := s.aiRepo.VerdictSeverityCounts(dayStart.UTC(), dayEnd.UTC())
	if err != nil {
		return err
	}
	// ماكو أحكام اليوم — ماكو داعي نزعج المالك بملخص فاضي.
	if counts.Total == 0 {
		return nil
	}

	message := fmt.Sprintf(
		"🧠 فضفضة ماتركس اليومية: %d حالة اليوم — %d حرجة، %d تنبيه، %d راقب، %d معلومة. التفاصيل بصندوق المراقب.",
		counts.Total, counts.Critical, counts.Warn, counts.Watch, counts.Info,
	)
	if err := s.notifRepo.CreateForRole("OWNER", "AI_DAILY_DEBRIEF", message); err != nil {
		return err
	}
	return s.notifRepo.CreateForRole("ADMIN", "AI_DAILY_DEBRIEF", message)
}
