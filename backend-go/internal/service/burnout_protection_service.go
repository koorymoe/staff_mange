package service

import (
	"fmt"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ حماية الموظف من الإرهاق (ماتركس) ═══
//
// طلب صاحب النظام: نمط — عمل أيام متواصلة كثيرة بلا راحة + ارتفاع
// إشاراته بنفس الفترة → ماتركس يقول "هذا تعب مو تقصير" ويقترح راحة
// مو عقوبة. يقلب دور ماتركس من مراقب إلى حامٍ.
//
// ⚠️ هذا مسار **مستقل** عن خط أنابيب AiSignal/AiVerdict/صندوق
// المراقب عمداً: ماكو حكم يحتاج "أوافق/عندي ملاحظة" هنا — هذا تنبيه
// حمائي للمالك ومدير النظام يقترح راحة، مو ملاحظة أداء تحتاج مراجعة.
const (
	burnoutConsecutiveDaysThreshold = 10
	burnoutLookbackDays             = 30
)

type BurnoutProtectionService struct {
	aiRepo     *repository.AiRepository
	employees  *repository.EmployeeRepository
	attendance *repository.AttendanceRepository
	notifRepo  *repository.NotificationRepository
}

func NewBurnoutProtectionService(
	aiRepo *repository.AiRepository,
	employees *repository.EmployeeRepository,
	attendance *repository.AttendanceRepository,
	notifRepo *repository.NotificationRepository,
) *BurnoutProtectionService {
	return &BurnoutProtectionService{aiRepo: aiRepo, employees: employees, attendance: attendance, notifRepo: notifRepo}
}

// RunWeeklyIfDue تُستدعى دورياً (كل ٦ ساعات) — يوم الاثنين، نفس حيلة
// الاستكشاف الأسبوعي.
func (s *BurnoutProtectionService) RunWeeklyIfDue() error {
	now := time.Now().In(debriefLoc)
	if now.Weekday() != time.Monday {
		return nil
	}
	weekKey := now.Format("2006-01-02")
	claimed, err := s.aiRepo.ClaimDailyMarker("WEEKLY_BURNOUT_SENT", weekKey)
	if err != nil || !claimed {
		return err
	}

	employees, err := s.employees.List()
	if err != nil {
		return err
	}
	weekAgo := now.Add(-7 * 24 * time.Hour)
	twoWeeksAgo := now.Add(-14 * 24 * time.Hour)

	var atRisk []string
	for _, e := range employees {
		streak, err := s.consecutiveWorkDays(e.ID, now)
		if err != nil || streak < burnoutConsecutiveDaysThreshold {
			continue
		}
		thisWeek := s.signalTotal(e.ID, weekAgo, now)
		lastWeek := s.signalTotal(e.ID, twoWeeksAgo, weekAgo)
		if thisWeek <= lastWeek {
			continue // ماكو ارتفاع بالإشارات — بس دوام متواصل بلا مؤشر تعب فعلي
		}
		atRisk = append(atRisk, fmt.Sprintf("%s (%d يوم متواصل، إشاراته صعدت من %d إلى %d)", e.Name, streak, lastWeek, thisWeek))
	}
	if len(atRisk) == 0 {
		return nil
	}
	msg := fmt.Sprintf(
		"🛡️ ماتركس — احتمال إرهاق: %s. يمكن هذا تعب مو تقصير — يستاهلون يوم راحة قبل ما يتصاعد.",
		joinArabicAnd(atRisk))
	if err := s.notifRepo.CreateForRole("OWNER", "AI_BURNOUT_RISK", msg); err != nil {
		return err
	}
	return s.notifRepo.CreateForRole("ADMIN", "AI_BURNOUT_RISK", msg)
}

func (s *BurnoutProtectionService) signalTotal(employeeID string, from, to time.Time) int {
	total := 0
	for _, kind := range praiseSignalKinds {
		if n, err := s.aiRepo.SignalCountForEmployeeBetween(kind, employeeID, from, to); err == nil {
			total += n
		}
	}
	return total
}

// consecutiveWorkDays شكد يوم متواصل اشتغل الموظف بلا انقطاع، رجوعاً
// من اليوم — يعتمد على سجل الحضور الحقيقي (`Attendance.date`)، بلا
// أي تخمين.
func (s *BurnoutProtectionService) consecutiveWorkDays(employeeID string, now time.Time) (int, error) {
	from := now.AddDate(0, 0, -burnoutLookbackDays).Format("2006-01-02")
	to := now.Format("2006-01-02")
	rows, err := s.attendance.ForEmployeeInRange(employeeID, from, to)
	if err != nil {
		return 0, err
	}
	worked := make(map[string]bool, len(rows))
	for _, r := range rows {
		worked[r.Date.Format("2006-01-02")] = true
	}
	streak := 0
	cursor := now
	for worked[cursor.Format("2006-01-02")] {
		streak++
		cursor = cursor.AddDate(0, 0, -1)
	}
	return streak, nil
}
