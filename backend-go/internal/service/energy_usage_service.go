package service

import (
	"encoding/json"
	"fmt"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ مؤشر الطاقة المستخدمة الشهري (ماتركس) ═══
//
// طلب صاحب النظام حرفياً: «هذا الموظف استخدم كذا من طاقته هذا الشهر —
// دوام ساعات إضافية هواي، طلع حجوزات هواي، حل مشاكل هواي». مؤشر
// تجميعي من بيانات موجودة أصلاً — صفر جدول جديد، تجميع بس.
//
// ⚠️ العناصر الثلاثة الحقيقية المتوفرة اليوم (بلا تخمين بشي ماكو):
//   - ساعات الدوام الفعلية (AttendanceService.MonthlyReport)
//   - عدد الحجوزات المكلَّف بيها (EmployeeMonthlyStatsService)
//   - حجوزات الصيانة/المشاكل المحلولة (نفس المصدر)
//
// نحسب لكل عنصر نسبة الموظف ÷ متوسط زملائه (نفس مبدأ `WorkSpeedScore`
// الموجود أصلاً: فوق ١ = أعلى من المعدل)، ومتوسط النسب الثلاث هو
// المؤشر. ونفس بوابتي الاستكشاف الأسبوعي بالضبط: عيّنة حقيقية (عنده
// شغل هذا الشهر أصلاً) وانحراف يتجاوز نسبة معلنة عن المعدل العام —
// قبل ما يوصل أي اسم للمالك. هذا تقرير «مين اشتغل فوق المعتاد»، مو
// تصنيف شامل لكل الموظفين كل شهر.
const energyDeviationThresholdPct = 40.0

type EnergyUsageService struct {
	aiRepo     *repository.AiRepository
	stats      *EmployeeMonthlyStatsService
	attendance *AttendanceService
	notifRepo  *repository.NotificationRepository
}

func NewEnergyUsageService(
	aiRepo *repository.AiRepository,
	stats *EmployeeMonthlyStatsService,
	attendance *AttendanceService,
	notifRepo *repository.NotificationRepository,
) *EnergyUsageService {
	return &EnergyUsageService{aiRepo: aiRepo, stats: stats, attendance: attendance, notifRepo: notifRepo}
}

// RunMonthlyIfDue تُستدعى دورياً (كل ٦ ساعات) — ما تسوي شي إلا أول
// ٣ أيام من الشهر (حتى الشهر السابق يكون اكتمل فعلياً)، وبعلامة شهر
// لسه ما انحسبت (نفس حيلة الاستكشاف الأسبوعي، بس بمفتاح شهري).
func (s *EnergyUsageService) RunMonthlyIfDue() error {
	now := time.Now().In(debriefLoc)
	if now.Day() > 3 {
		return nil
	}
	prevMonth := now.AddDate(0, 0, -now.Day()).Format("2006-01")
	claimed, err := s.aiRepo.ClaimDailyMarker("MONTHLY_ENERGY_SENT", prevMonth)
	if err != nil || !claimed {
		return err
	}

	rows, err := s.stats.Monthly(prevMonth)
	if err != nil {
		return err
	}

	minutesByEmp := make(map[string]int, len(rows))
	var totalBookings, totalMaintenance, totalMinutes float64
	n := 0
	for _, row := range rows {
		minutes := 0
		if report, aerr := s.attendance.MonthlyReport(row.EmployeeID, prevMonth); aerr == nil && report != nil {
			minutes = report.TotalMinutes
		}
		minutesByEmp[row.EmployeeID] = minutes
		if row.TotalBookingsCount == 0 && minutes == 0 {
			continue // ماكو شغل مسجّل هذا الشهر أصلاً — يُستثنى من المقارنة
		}
		totalBookings += float64(row.TotalBookingsCount)
		totalMaintenance += float64(row.MaintenanceBookingsCount)
		totalMinutes += float64(minutes)
		n++
	}
	if n == 0 {
		return nil
	}
	avgBookings := totalBookings / float64(n)
	avgMaintenance := totalMaintenance / float64(n)
	avgMinutes := totalMinutes / float64(n)
	periodStart := monthStart(prevMonth)

	var standouts []string
	for _, row := range rows {
		minutes := minutesByEmp[row.EmployeeID]
		if row.TotalBookingsCount == 0 && minutes == 0 {
			continue
		}
		score := (ratioComponent(float64(row.TotalBookingsCount), avgBookings) +
			ratioComponent(float64(row.MaintenanceBookingsCount), avgMaintenance) +
			ratioComponent(float64(minutes), avgMinutes)) / 3

		details, _ := json.Marshal(map[string]any{
			"bookings":         row.TotalBookingsCount,
			"maintenanceCount": row.MaintenanceBookingsCount,
			"workedMinutes":    minutes,
			"avgBookings":      avgBookings,
			"avgMaintenance":   avgMaintenance,
			"avgMinutes":       avgMinutes,
		})
		empID := row.EmployeeID
		if uerr := s.aiRepo.UpsertMetric(model.AiMetric{
			MetricKey: model.AiMetricEnergyUsed, Scope: model.AiAxisEmployee, ScopeID: &empID,
			PeriodStart: periodStart, PeriodEnd: periodStart,
			Value: score, SampleCount: row.TotalBookingsCount, Details: details,
		}); uerr != nil {
			continue
		}

		deviation := (score - 1.0) * 100
		if deviation >= energyDeviationThresholdPct {
			standouts = append(standouts, fmt.Sprintf("%s (%.0f٪ فوق معدل زملائه)", row.EmployeeName, deviation))
		}
	}
	if len(standouts) == 0 {
		return nil
	}
	msg := fmt.Sprintf(
		"⚡ ماتركس — طاقة الشهر: %s استخدموا طاقة أعلى بكثير من زملائهم هذا الشهر (حجوزات + ساعات دوام + صيانات). التفاصيل بمؤشرات الذكاء الاصطناعي.",
		joinArabicAnd(standouts))
	if err := s.notifRepo.CreateForRole("OWNER", "AI_MONTHLY_ENERGY", msg); err != nil {
		return err
	}
	return s.notifRepo.CreateForRole("ADMIN", "AI_MONTHLY_ENERGY", msg)
}

func ratioComponent(value, avg float64) float64 {
	if avg <= 0 {
		return 1
	}
	return value / avg
}

func monthStart(month string) time.Time {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return time.Now()
	}
	return t
}
