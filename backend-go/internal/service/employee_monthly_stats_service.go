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

		result = append(result, stats)
	}

	return result, nil
}
