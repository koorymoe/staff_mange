package service

import (
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// StatsManagementService يبني "إدارة الإحصائيات" — يومية/أسبوعية/مشاريع (مو
// الشهرية، هذي موجودة أصلاً بـEmployeeMonthlyStatsService). OWNER/ADMIN فقط.
type StatsManagementService struct {
	employees       *repository.EmployeeRepository
	bookings        *repository.BookingRepository
	commissions     *repository.EmployeeCommissionRepository
	projects        *repository.ProjectRepository
	leaderInvoices  *repository.LeaderInvoiceRepository
	attendance      *repository.AttendanceRepository
	employeeMonthly *EmployeeMonthlyStatsService
}

func NewStatsManagementService(
	employees *repository.EmployeeRepository,
	bookings *repository.BookingRepository,
	commissions *repository.EmployeeCommissionRepository,
	projects *repository.ProjectRepository,
	leaderInvoices *repository.LeaderInvoiceRepository,
	attendance *repository.AttendanceRepository,
	employeeMonthly *EmployeeMonthlyStatsService,
) *StatsManagementService {
	return &StatsManagementService{
		employees: employees, bookings: bookings, commissions: commissions, projects: projects,
		leaderInvoices: leaderInvoices, attendance: attendance, employeeMonthly: employeeMonthly,
	}
}

// Daily يبني إحصائية يوم معيّن (date بصيغة "YYYY-MM-DD") — اليوم افتراضياً لو
// ما تحدد تاريخ، حتى تكدر ترجع لأي يوم سابق وتشوف نفس الإحصائية بالضبط.
func (s *StatsManagementService) Daily(date string) (*model.DailyStats, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	employees, err := s.employees.List()
	if err != nil {
		return nil, err
	}

	total, err := s.bookings.CountForDate(date)
	if err != nil {
		return nil, err
	}
	morning, evening, err := s.bookings.CountMorningEveningForDate(date)
	if err != nil {
		return nil, err
	}
	crewOut, err := s.bookings.CountDistinctCrewForDate(date)
	if err != nil {
		return nil, err
	}
	vehiclesOut, err := s.bookings.CountDistinctVehiclesForDate(date)
	if err != nil {
		return nil, err
	}
	salesAmount, err := s.leaderInvoices.SumNetTotalForDate(date)
	if err != nil {
		return nil, err
	}
	profitAmount, err := s.commissions.SumForDate(date)
	if err != nil {
		return nil, err
	}
	attendanceRows, err := s.attendance.DaySummary(date)
	if err != nil {
		return nil, err
	}
	checkedIn := map[string]bool{}
	for _, a := range attendanceRows {
		checkedIn[a.EmployeeID] = true
	}

	result := &model.DailyStats{
		Date: date, TotalBookings: total, MorningBookings: morning, EveningBookings: evening,
		CrewOutCount: crewOut, VehiclesOutCount: vehiclesOut, TotalEmployeesCount: len(employees),
		TotalSalesAmount: salesAmount, TotalProfitAmount: profitAmount,
		Employees: []model.DailyEmployeeStats{},
	}
	for _, e := range employees {
		assigned, completed, cerr := s.bookings.AssignedAndCompletedForEmployeeOnDate(e.ID, date)
		if cerr != nil || (assigned == 0 && !checkedIn[e.ID]) {
			continue
		}
		result.Employees = append(result.Employees, model.DailyEmployeeStats{
			EmployeeID: e.ID, EmployeeName: e.Name, Role: e.Role,
			BookingsAssigned: assigned, BookingsCompleted: completed, CheckedIn: checkedIn[e.ID],
		})
	}
	return result, nil
}

// Weekly يبني الإحصائية الأسبوعية بمدى تاريخ حر (from/to بصيغة "YYYY-MM-DD") —
// افتراضياً آخر 7 أيام لو ما تحدد مدى. الفقرة الأولى: حجم المبيعات صباحي/
// مسائي/مجموع. الفقرة الثانية: نفس أعمدة الإحصائية الشهرية محسوبة على مدى
// الأسبوع (بدون تعديل "عدد الخدمات" — هذا خاص بالشهرية فقط).
func (s *StatsManagementService) Weekly(from, to string) (*model.WeeklyStats, error) {
	if from == "" || to == "" {
		to = time.Now().Format("2006-01-02")
		from = time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	}

	morning, evening, err := s.leaderInvoices.SumNetTotalMorningEveningForRange(from, to)
	if err != nil {
		return nil, err
	}
	total, err := s.leaderInvoices.SumNetTotalForRange(from, to)
	if err != nil {
		return nil, err
	}
	employees, err := s.employeeMonthly.Range(from, to)
	if err != nil {
		return nil, err
	}

	return &model.WeeklyStats{
		From: from, To: to,
		MorningSalesAmount: morning, EveningSalesAmount: evening, TotalSalesAmount: total,
		Employees: employees,
	}, nil
}

func (s *StatsManagementService) ProjectStages() ([]model.ProjectStageStats, error) {
	projects, err := s.projects.List()
	if err != nil {
		return nil, err
	}
	counts := map[string]int{}
	order := []string{}
	for _, p := range projects {
		if _, ok := counts[p.Stage]; !ok {
			order = append(order, p.Stage)
		}
		counts[p.Stage]++
	}
	result := make([]model.ProjectStageStats, 0, len(order))
	for _, stage := range order {
		result = append(result, model.ProjectStageStats{Stage: stage, Count: counts[stage]})
	}
	return result, nil
}
