package service

import (
	"fmt"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// EmployeeMonthlyStatsService يجمع كل مؤشرات الموظف الشهرية بنداء واحد (صفحة
// إحصائيات الموظفين — OWNER/ADMIN فقط): نقاط الكي بي اي (يعيد استخدام آلية
// KpiEvaluation الموجودة أصلاً)، سرعة العمل (placeholder لغاية ما تنتهي ميزة
// تقدير مدة التنفيذ)، نظافة السيارة (يعيد استخدام "تقييم السائقين بعد المهمة"
// الموجود أصلاً)، الشكاوى، المبيعات، الحجوزات المكتملة، ومجموع العمولات.
type EmployeeMonthlyStatsService struct {
	employees            *repository.EmployeeRepository
	kpi                  *repository.KpiRepository
	complaints           *repository.ComplaintRepository
	leaderInvoices       *repository.LeaderInvoiceRepository
	bookings             *repository.BookingRepository
	vehicleMissionRating *repository.VehicleMissionRatingRepository
	commissions          *repository.EmployeeCommissionRepository
}

func NewEmployeeMonthlyStatsService(
	employees *repository.EmployeeRepository,
	kpi *repository.KpiRepository,
	complaints *repository.ComplaintRepository,
	leaderInvoices *repository.LeaderInvoiceRepository,
	bookings *repository.BookingRepository,
	vehicleMissionRating *repository.VehicleMissionRatingRepository,
	commissions *repository.EmployeeCommissionRepository,
) *EmployeeMonthlyStatsService {
	return &EmployeeMonthlyStatsService{
		employees:            employees,
		kpi:                  kpi,
		complaints:           complaints,
		leaderInvoices:       leaderInvoices,
		bookings:             bookings,
		vehicleMissionRating: vehicleMissionRating,
		commissions:          commissions,
	}
}

// Monthly يبني إحصائيات كل موظف نشط خلال شهر معيّن (month بصيغة "YYYY-MM").
func (s *EmployeeMonthlyStatsService) Monthly(month string) ([]model.EmployeeMonthlyStats, error) {
	if _, err := time.Parse("2006-01", month); err != nil {
		return nil, fmt.Errorf("صيغة الشهر يجب أن تكون YYYY-MM")
	}

	employees, err := s.employees.List()
	if err != nil {
		return nil, err
	}

	result := make([]model.EmployeeMonthlyStats, 0, len(employees))
	for _, e := range employees {
		stats := model.EmployeeMonthlyStats{
			EmployeeID:   e.ID,
			EmployeeName: e.Name,
			Role:         e.Role,
			Month:        month,
			// WorkSpeedScore يبقى nil عمداً — TODO يُملأ بعد اكتمال ميزة تقدير
			// مدة تنفيذ العمل (job-duration-estimation)، مبنية بفريق موازٍ.
			WorkSpeedScore: nil,
		}

		if points, kerr := s.kpi.SumPointsForEmployeeMonth(e.ID, month); kerr == nil {
			stats.KpiPoints = points
			stats.KpiPointsValue = float64(points) * 10000
		}

		if avg, count, verr := s.vehicleMissionRating.GetCleanlinessAvgForDriverMonth(e.ID, month); verr == nil {
			stats.VehicleCleanlinessScore = avg
			stats.VehicleRatingsCount = count
		}

		if count, cerr := s.complaints.CountForEmployeeMonth(e.ID, month); cerr == nil {
			stats.ComplaintsCount = count
		}

		if count, serr := s.leaderInvoices.CountForEmployeeMonth(e.ID, month); serr == nil {
			stats.SalesCount = count
		}

		if count, berr := s.bookings.CountCompletedForEmployeeMonth(e.ID, month); berr == nil {
			stats.CompletedBookingsCount = count
		}

		if total, comErr := s.commissions.SumForEmployeeMonth(e.ID, month); comErr == nil {
			stats.TotalCommission = total
		}

		if count, err := s.bookings.CountAssignedForEmployeeMonth(e.ID, month); err == nil {
			stats.TotalBookingsCount = count
		}
		if count, err := s.bookings.CountMaintenanceForEmployeeMonth(e.ID, month); err == nil {
			stats.MaintenanceBookingsCount = count
		}
		if count, err := s.bookings.CountFreeMaintenanceForEmployeeMonth(e.ID, month); err == nil {
			stats.FreeMaintenanceCount = count
		}

		if count, err := s.employees.CountDistinctServicesKnown(e.ID); err == nil {
			stats.ServicesKnownCount = count
		}

		// القائمة تنرسل دائماً — فاضية لو ماكو، حتى الواجهة ما تنكسر
		stats.InHouseWorkTypes = []string{}
		if count, types, err := s.bookings.CountInHouseForEmployeeMonth(e.ID, month); err == nil {
			stats.InHouseWorksCount = count
			if len(types) > 0 {
				stats.InHouseWorkTypes = types
			}
		}

		result = append(result, stats)
	}

	return result, nil
}

// Range يبني نفس صفوف Monthly (نقاط الكي بي اي، الشكاوى، المبيعات، الحجوزات،
// العمولة...) لكن لمدى تاريخ حر (from/to بصيغة "YYYY-MM-DD") بدل شهر كامل —
// تُستخدم بالإحصائية الأسبوعية بفلتر من/إلى. لا تحسب ServicesKnownCount عمداً
// (التعديل الخاص بإزالة تكرار نقاط الكي بي اي يخص الشهرية فقط، مو الأسبوعية).
func (s *EmployeeMonthlyStatsService) Range(from, to string) ([]model.EmployeeMonthlyStats, error) {
	if _, err := time.Parse("2006-01-02", from); err != nil {
		return nil, fmt.Errorf("صيغة تاريخ البداية يجب أن تكون YYYY-MM-DD")
	}
	if _, err := time.Parse("2006-01-02", to); err != nil {
		return nil, fmt.Errorf("صيغة تاريخ النهاية يجب أن تكون YYYY-MM-DD")
	}

	employees, err := s.employees.List()
	if err != nil {
		return nil, err
	}

	result := make([]model.EmployeeMonthlyStats, 0, len(employees))
	for _, e := range employees {
		stats := model.EmployeeMonthlyStats{
			EmployeeID: e.ID, EmployeeName: e.Name, Role: e.Role,
			From: &from, To: &to,
		}

		if points, kerr := s.kpi.SumPointsForEmployeeRange(e.ID, from, to); kerr == nil {
			stats.KpiPoints = points
			stats.KpiPointsValue = float64(points) * 10000
		}
		if avg, count, verr := s.vehicleMissionRating.GetCleanlinessAvgForDriverRange(e.ID, from, to); verr == nil {
			stats.VehicleCleanlinessScore = avg
			stats.VehicleRatingsCount = count
		}
		if count, cerr := s.complaints.CountForEmployeeRange(e.ID, from, to); cerr == nil {
			stats.ComplaintsCount = count
		}
		if count, serr := s.leaderInvoices.CountForEmployeeRange(e.ID, from, to); serr == nil {
			stats.SalesCount = count
		}
		if count, berr := s.bookings.CountCompletedForEmployeeRange(e.ID, from, to); berr == nil {
			stats.CompletedBookingsCount = count
		}
		if total, comErr := s.commissions.SumForEmployeeRange(e.ID, from, to); comErr == nil {
			stats.TotalCommission = total
		}
		if count, err := s.bookings.CountAssignedForEmployeeRange(e.ID, from, to); err == nil {
			stats.TotalBookingsCount = count
		}
		if count, err := s.bookings.CountMaintenanceForEmployeeRange(e.ID, from, to); err == nil {
			stats.MaintenanceBookingsCount = count
		}
		if count, err := s.bookings.CountFreeMaintenanceForEmployeeRange(e.ID, from, to); err == nil {
			stats.FreeMaintenanceCount = count
		}

		result = append(result, stats)
	}

	return result, nil
}

// Curve يبني منحنى أداء موظف واحد — نقاط الكي بي اي والعمولات شهرياً لآخر
// monthsCount شهر (6 افتراضياً) — يُستخدم بمنحنى الأداء المتحرك.
// Curve يبني منحنى أداء موظف واحد — نقاط الكي بي اي والعمولات لكل شهر تقويمي
// حقيقي (بدل نافذة متحركة ثابتة) — month (بصيغة "YYYY-MM") يحدد الشهر
// الأخير بالمنحنى، يبقى المستخدم يقدر يتصفح لأي شهر سابق يريده.
func (s *EmployeeMonthlyStatsService) Curve(employeeID string, monthsCount int, month string) (*model.EmployeePerformanceCurve, error) {
	if monthsCount <= 0 {
		monthsCount = 6
	}
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	if _, err := time.Parse("2006-01", month); err != nil {
		return nil, fmt.Errorf("صيغة الشهر يجب أن تكون YYYY-MM")
	}
	endMonth := month + "-01"

	employee, err := s.employees.FindByID(employeeID)
	if err != nil || employee == nil {
		return nil, fmt.Errorf("الموظف غير موجود")
	}
	points, err := s.kpi.MonthlyPointsSeriesForEmployee(employeeID, monthsCount, endMonth)
	if err != nil {
		return nil, err
	}
	commission, err := s.commissions.MonthlyCommissionSeriesForEmployee(employeeID, monthsCount, endMonth)
	if err != nil {
		return nil, err
	}
	return &model.EmployeePerformanceCurve{
		EmployeeID: employee.ID, EmployeeName: employee.Name,
		Points: points, Commission: commission,
	}, nil
}
