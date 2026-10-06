package service

import (
	"fmt"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type AttendanceService struct {
	repo *repository.AttendanceRepository
}

func NewAttendanceService(repo *repository.AttendanceRepository) *AttendanceService {
	return &AttendanceService{repo: repo}
}

func (s *AttendanceService) CheckIn(employeeID string) (*model.Attendance, error) {
	return s.repo.CheckIn(employeeID)
}

func (s *AttendanceService) CheckOut(employeeID string) (*model.Attendance, error) {
	return s.repo.CheckOut(employeeID)
}

func (s *AttendanceService) Today() ([]model.Attendance, error) {
	return s.repo.Today()
}

func (s *AttendanceService) TodaySummary() ([]model.EmployeeDailyAttendanceSummary, error) {
	return s.repo.TodaySummary()
}

func (s *AttendanceService) DaySummary(date string) ([]model.EmployeeDailyAttendanceSummary, error) {
	return s.repo.DaySummary(date)
}

func (s *AttendanceService) Mine(employeeID string) (*model.Attendance, error) {
	return s.repo.FindToday(employeeID)
}

// OpenSession ترجع جلسة الحضور المفتوحة حالياً للموظف (لو موجودة) — هذا هو
// المعنى الحقيقي لـ"مسجل دخول حالياً" بعد دعم الجلسات المتعددة باليوم.
func (s *AttendanceService) OpenSession(employeeID string) (*model.Attendance, error) {
	return s.repo.FindOpenSession(employeeID)
}

// TodaySessions ترجع كل جلسات اليوم للموظف مع مجموع الدقائق (المؤكدة + الجارية
// لو فيه جلسة مفتوحة حالياً).
func (s *AttendanceService) TodaySessions(employeeID string) ([]model.Attendance, int, bool, error) {
	sessions, err := s.repo.TodaySessions(employeeID)
	if err != nil {
		return nil, 0, false, err
	}
	total := 0
	open := false
	now := time.Now()
	for _, sess := range sessions {
		if sess.CheckOut != nil {
			total += int(sess.CheckOut.Sub(sess.CheckIn).Minutes())
		} else {
			total += int(now.Sub(sess.CheckIn).Minutes())
			open = true
		}
	}
	return sessions, total, open, nil
}

// month is expected in "YYYY-MM" form; defaults to the current month when empty.
func (s *AttendanceService) MonthlyReport(employeeID, month string) (*model.MonthlyAttendanceReport, error) {
	var start time.Time
	var err error
	if month == "" {
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	} else {
		start, err = time.Parse("2006-01", month)
		if err != nil {
			return nil, fmt.Errorf("صيغة الشهر غير صحيحة")
		}
	}
	end := start.AddDate(0, 1, 0)

	records, err := s.repo.ForEmployeeInRange(employeeID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}

	report := &model.MonthlyAttendanceReport{
		EmployeeID: employeeID,
		Month:      start.Format("2006-01"),
		Days:       groupByCalendarDate(records),
	}
	for _, day := range report.Days {
		report.TotalMinutes += day.TotalMinutes
	}
	report.DaysPresent = len(report.Days)
	return report, nil
}

// groupByCalendarDate يجمّع صفوف الحضور (ممكن أكثر من صف باليوم الواحد بعد
// دعم الجلسات المتعددة) بصف واحد لكل تاريخ تقويمي.
func groupByCalendarDate(records []model.Attendance) []model.DailyAttendance {
	order := []string{}
	byDate := map[string]*model.DailyAttendance{}
	now := time.Now()

	for _, rec := range records {
		key := rec.Date.Format("2006-01-02")
		day, ok := byDate[key]
		if !ok {
			day = &model.DailyAttendance{Date: key, FirstCheckIn: rec.CheckIn}
			byDate[key] = day
			order = append(order, key)
		}
		day.Sessions = append(day.Sessions, rec)
		if rec.CheckIn.Before(day.FirstCheckIn) {
			day.FirstCheckIn = rec.CheckIn
		}
		if rec.CheckOut != nil {
			if day.LastCheckOut == nil || rec.CheckOut.After(*day.LastCheckOut) {
				day.LastCheckOut = rec.CheckOut
			}
			day.TotalMinutes += int(rec.CheckOut.Sub(rec.CheckIn).Minutes())
		} else {
			day.StillOpen = true
			day.TotalMinutes += int(now.Sub(rec.CheckIn).Minutes())
		}
	}

	days := make([]model.DailyAttendance, 0, len(order))
	for _, key := range order {
		days = append(days, *byDate[key])
	}
	return days
}

func (s *AttendanceService) Correct(id string, req model.SetAttendanceCorrectionRequest) (*model.Attendance, error) {
	return s.repo.Correct(id, req.CheckIn, req.CheckOut)
}

// ═══ الحضور الإجباري — قرار (ع) 10-06 ═══
// «اول ما يفتح النظام الموظف اريده ينطلب منه تسجيل الحضور… ميعبر للنقطه
// الي بعدها». والي ما يسجّل انصراف ينسكّر تلقائياً بعد الشفت.

type AttendanceGate struct {
	Required   bool      `json:"required"`   // لازم يسجّل حضور قبل ما يكمّل
	Open       bool      `json:"open"`       // عنده جلسة مفتوحة
	AfterShift bool      `json:"afterShift"` // دوامه خلص — اسأله «آخر حجز؟»
	ShiftStart time.Time `json:"shiftStart"`
	ShiftEnd   time.Time `json:"shiftEnd"`
	EndLabel   string    `json:"endLabel"`
	Evening    bool      `json:"evening"`
}

// Gate حالة الموظف هسه. المالك ومدير النظام والي بإجازة معتمدة ما ينطلب منهم.
func (s *AttendanceService) Gate(employeeID string, now time.Time) (*AttendanceGate, error) {
	g, err := s.repo.Gate(employeeID)
	if err != nil {
		return nil, err
	}
	day := now.In(debriefLoc)
	from, to := ShiftWindow(g.Shift, g.ShiftStart, g.ShiftEnd, day)
	// المسائي بعد نص الليل: الشفت مال البارحة.
	if day.Before(from) {
		pf, pt := ShiftWindow(g.Shift, g.ShiftStart, g.ShiftEnd, day.AddDate(0, 0, -1))
		if day.Before(pt.Add(autoCheckoutGrace)) {
			from, to = pf, pt
		}
	}
	exempt := g.Role == "ADMIN" || g.Role == "OWNER"
	out := &AttendanceGate{Open: g.HasOpen, ShiftStart: from, ShiftEnd: to, EndLabel: clockLabel(to),
		Evening: g.Shift != nil && *g.Shift == "EVENING"}
	out.Required = !exempt && !g.OnLeave && !g.HasOpen && !g.HadToday
	out.AfterShift = !exempt && g.HasOpen && !now.Before(to)
	return out, nil
}

// AutoCheckout يسكّر الجلسات المفتوحة بعد نهاية الشفت بـ٣ ساعات.
func (s *AttendanceService) AutoCheckout(now time.Time, notify func(employeeID, msg string)) (int, error) {
	rows, err := s.repo.OpenWithShift()
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range rows {
		_, to := ShiftWindow(r.Shift, r.ShiftStart, r.ShiftEnd, r.CheckIn)
		// جلسة بدت قبل بداية الشفت بيوم (مسائي بعد نص الليل): نفس الشفت.
		at, ok := AutoCheckoutAt(r.CheckIn, to, r.LastActivity, now)
		if !ok {
			continue
		}
		reason := "ما سجّل انصراف — انسكّر بنهاية الشفت (" + clockLabel(to) + ")."
		if at.After(to) {
			reason = "ما سجّل انصراف — انسكّر بوقت آخر حجز خلّصه (" + clockLabel(at) + ")."
		}
		if err := s.repo.AutoClose(r.ID, r.EmployeeID, at, reason); err != nil {
			continue
		}
		n++
		if notify != nil {
			notify(r.EmployeeID, "🤖 ماتركس: ما سجّلت انصراف، فسجّلتلك انصراف تلقائي الساعة "+clockLabel(at)+". لا تنسى تسجّل انصرافك بنفسك من تخلص.")
		}
	}
	return n, nil
}
