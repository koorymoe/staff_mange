package service

import (
	"fmt"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ صيانة المركبات مقابل الاستخدام الفعلي (ماتركس) ═══
//
// «مركبة صيانتها متأخرة ولسه تُرسل لحجوزات» — حالة مستمرة مو حدثاً
// لمرة وحدة (نفس سبب حماية الإرهاق ومؤشر الطاقة)، فمسارها تنبيه دوري
// مباشر للمالك ومدير النظام، مو إشارة AiSignal تحتاج مراجعة.
type VehicleMaintenanceAlertService struct {
	aiRepo    *repository.AiRepository
	notifRepo *repository.NotificationRepository
}

func NewVehicleMaintenanceAlertService(aiRepo *repository.AiRepository, notifRepo *repository.NotificationRepository) *VehicleMaintenanceAlertService {
	return &VehicleMaintenanceAlertService{aiRepo: aiRepo, notifRepo: notifRepo}
}

// RunWeeklyIfDue تُستدعى دورياً (كل ٦ ساعات) — يوم الاثنين، نفس حيلة
// الاستكشاف الأسبوعي.
func (s *VehicleMaintenanceAlertService) RunWeeklyIfDue() error {
	now := time.Now().In(debriefLoc)
	if now.Weekday() != time.Monday {
		return nil
	}
	weekKey := now.Format("2006-01-02")
	claimed, err := s.aiRepo.ClaimDailyMarker("WEEKLY_VEHICLE_MAINTENANCE_SENT", weekKey)
	if err != nil || !claimed {
		return err
	}

	rows, err := s.aiRepo.OverdueMaintenanceVehiclesInUse()
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	lines := make([]string, 0, len(rows))
	for _, r := range rows {
		lines = append(lines, fmt.Sprintf("%s (%s)", r.VehicleName, r.PlateNumber))
	}
	msg := fmt.Sprintf(
		"🔧 ماتركس — صيانة متأخرة وبعدها بالخدمة: %s. صيانتها المجدولة فاتت وهي لسه تُرسل بمهام/حجوزات.",
		joinArabicAnd(lines))
	if err := s.notifRepo.CreateForRole("OWNER", "AI_VEHICLE_MAINTENANCE_OVERDUE", msg); err != nil {
		return err
	}
	return s.notifRepo.CreateForRole("ADMIN", "AI_VEHICLE_MAINTENANCE_OVERDUE", msg)
}
