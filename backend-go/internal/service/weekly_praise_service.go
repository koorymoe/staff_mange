package service

import (
	"fmt"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ ماتركس يمدح مو بس يعاتب ═══
//
// فجوة حقيقية بكل إشارات الذكاء الخمسة: كلها سلبية — صفر مسار
// إيجابي. هذا إشعار أسبوعي شخصي لكل موظف (له هو، مو للمالك) يعرض
// تحسّنه: "صفر تأخير هذا الأسبوع" أو "تحسّنت من ٣ توقفات لصفر".
//
// ⚠️ بلا خطر بالتصميم: ما تعاقب أحد ولا تسكت عنه — لو ماكو مادة
// إيجابية حقيقية هذا الأسبوع (لا تحسّن ولا فترة نظيفة تستاهل تنقال)،
// الموظف بس ما ياخذ إشعار. صمت، مو لوم.
var praiseSignalKinds = []string{model.AiSignalWorkStopped, model.AiSignalLateStart}

type WeeklyPraiseService struct {
	aiRepo    *repository.AiRepository
	employees *repository.EmployeeRepository
	notifRepo *repository.NotificationRepository
}

func NewWeeklyPraiseService(
	aiRepo *repository.AiRepository,
	employees *repository.EmployeeRepository,
	notifRepo *repository.NotificationRepository,
) *WeeklyPraiseService {
	return &WeeklyPraiseService{aiRepo: aiRepo, employees: employees, notifRepo: notifRepo}
}

// RunWeeklyIfDue تُستدعى دورياً (كل ٦ ساعات) — ما تسوي شي إلا يوم
// الاثنين بتوقيت بغداد، وبعلامة أسبوع لسه ما انرسلت (نفس حيلة
// `AiDiscoveryService.RunWeeklyIfDue` بالضبط).
func (s *WeeklyPraiseService) RunWeeklyIfDue() error {
	now := time.Now().In(debriefLoc)
	if now.Weekday() != time.Monday {
		return nil
	}
	weekKey := now.Format("2006-01-02")
	claimed, err := s.aiRepo.ClaimDailyMarker("WEEKLY_PRAISE_SENT", weekKey)
	if err != nil || !claimed {
		return err
	}

	employees, err := s.employees.List()
	if err != nil {
		return err
	}
	weekAgo := now.Add(-7 * 24 * time.Hour)
	twoWeeksAgo := now.Add(-14 * 24 * time.Hour)

	for _, e := range employees {
		lines := s.praiseLinesFor(e.ID, now, weekAgo, twoWeeksAgo)
		if len(lines) == 0 {
			continue
		}
		message := "🌟 ماتركس: " + joinArabicAnd(lines)
		_ = s.notifRepo.Create(e.ID, "AI_WEEKLY_PRAISE", message)
	}
	return nil
}

// praiseLinesFor يبني جمل المديح لموظف واحد — سطر لكل صنف إشارة عنده
// مادة إيجابية حقيقية هذا الأسبوع، وبلا شي لو ماكو.
func (s *WeeklyPraiseService) praiseLinesFor(employeeID string, now, weekAgo, twoWeeksAgo time.Time) []string {
	lines := []string{}
	for _, kind := range praiseSignalKinds {
		thisWeek, err := s.aiRepo.SignalCountForEmployeeBetween(kind, employeeID, weekAgo, now)
		if err != nil || thisWeek > 0 {
			continue // هذا الأسبوع فيه إشارة فعلية — ماكو مديح على هذا النوع
		}
		label := model.AiSignalLabel(kind)
		lastWeek, err := s.aiRepo.SignalCountForEmployeeBetween(kind, employeeID, twoWeeksAgo, weekAgo)
		if err == nil && lastWeek > 0 {
			lines = append(lines, fmt.Sprintf("تحسّنت من %d إلى صفر بـ«%s» هذا الأسبوع 👏", lastWeek, label))
			continue
		}
		days, err := s.aiRepo.DaysSinceLastSignal(kind, employeeID, now)
		if err == nil && days != nil && *days >= escalationCleanStreakDays {
			lines = append(lines, fmt.Sprintf("صفر «%s» من %d يوم — مستمر بالانضباط 👏", label, *days))
		}
	}
	return lines
}

func joinArabicAnd(lines []string) string {
	out := ""
	for i, l := range lines {
		if i > 0 {
			out += " و"
		}
		out += l
	}
	return out
}
