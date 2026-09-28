package service

import (
	"fmt"
	"strings"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ صحة الزبون (ماتركس ٩) — زبائن مخلصين انقطعوا ═══
//
// زبون عنده ≥٣ حجوزات مكتملة وما حجز من ١٢٠ يوم — حالة مستمرة مو
// حدثاً، فمسارها تنبيه أسبوعي للمالك ومدير النظام مو إشارة AiSignal
// (نفس سبب تنبيه صيانة المركبات). أكواد زبائن بس، بلا أسماء أو هواتف.
type LapsedCustomerAlertService struct {
	aiRepo    *repository.AiRepository
	notifRepo *repository.NotificationRepository
}

const (
	lapsedMinCompleted = 3
	lapsedIdleDays     = 120
	lapsedTopCodes     = 10
)

func NewLapsedCustomerAlertService(aiRepo *repository.AiRepository, notifRepo *repository.NotificationRepository) *LapsedCustomerAlertService {
	return &LapsedCustomerAlertService{aiRepo: aiRepo, notifRepo: notifRepo}
}

// isoWeekMonday يرجّع تاريخ اثنين الأسبوع (ISO) — مفتاح «مرة بالأسبوع».
func isoWeekMonday(t time.Time) string {
	offset := (int(t.Weekday()) + 6) % 7 // الاثنين = 0
	return t.AddDate(0, 0, -offset).Format("2006-01-02")
}

// RunWeeklyIfDue تُستدعى دورياً — أول مرة بكل أسبوع ISO تحجز المفتاح
// وترسل، والباقي يطلع بصمت.
func (s *LapsedCustomerAlertService) RunWeeklyIfDue() error {
	now := time.Now().In(debriefLoc)
	claimed, err := s.aiRepo.ClaimDailyMarker("WEEKLY_LAPSED_CUSTOMERS_SENT", isoWeekMonday(now))
	if err != nil || !claimed {
		return err
	}
	rows, err := s.aiRepo.LapsedLoyalCustomers(lapsedMinCompleted, lapsedIdleDays)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	codes := []string{}
	for i, r := range rows {
		if i >= lapsedTopCodes {
			break
		}
		codes = append(codes, fmt.Sprintf("%s (%d)", model.Customer{CustomerCode: r.CustomerCode}.FormatCode(), r.CompletedCount))
	}
	msg := fmt.Sprintf(
		"🤝 ماتركس — %d زبون عندهم %d حجوزات مكتملة أو أكثر وما حجزوا من %d يوم. أبرزهم: %s. يستاهلون اتصال متابعة.",
		len(rows), lapsedMinCompleted, lapsedIdleDays, strings.Join(codes, "، "))
	if err := s.notifRepo.CreateForRole("OWNER", "AI_LAPSED_CUSTOMERS", msg); err != nil {
		return err
	}
	return s.notifRepo.CreateForRole("ADMIN", "AI_LAPSED_CUSTOMERS", msg)
}
