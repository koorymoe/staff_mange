package service

import (
	"fmt"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ ملخص الإنجازات اليومي (ماتركس) ═══
//
// تقارير الإنجازات لا تنرسل فرديّاً لكل تقرير (سبام على ٤٥+ موظف) —
// تتجمّع بملخص يوم واحد يوصل لمدير النظام والمالك حصراً، نفس فلسفة
// «الفضفضة اليومية» بالضبط. التفصيل الكامل يبقى بشاشة الإنجازات
// (`requireAdmin`).
type AchievementDigestService struct {
	repo      *repository.AchievementRepository
	aiRepo    *repository.AiRepository
	notifRepo *repository.NotificationRepository
}

func NewAchievementDigestService(
	repo *repository.AchievementRepository,
	aiRepo *repository.AiRepository,
	notifRepo *repository.NotificationRepository,
) *AchievementDigestService {
	return &AchievementDigestService{repo: repo, aiRepo: aiRepo, notifRepo: notifRepo}
}

// RunIfDue تُستدعى دورياً (كل ٣٠ دقيقة) — نفس نافذة ونمط
// `DailyDebriefService.RunIfDue` بالضبط، بمفتاح علامة مختلف.
func (s *AchievementDigestService) RunIfDue() error {
	now := time.Now().In(debriefLoc)
	if now.Hour() < dailyDebriefHour {
		return nil
	}
	day := now.Format("2006-01-02")
	claimed, err := s.aiRepo.ClaimDailyMarker("ACHIEVEMENT_DIGEST_SENT", day)
	if err != nil || !claimed {
		return err
	}

	total, byRole, err := s.repo.CountsForDay(now)
	if err != nil {
		return err
	}
	if total == 0 {
		return nil
	}

	parts := make([]string, 0, len(byRole))
	for _, r := range byRole {
		parts = append(parts, fmt.Sprintf("%s: %d", r.Role, r.Count))
	}
	message := fmt.Sprintf(
		"📋 ماتركس — إنجازات اليوم: %d تقرير (%s). التفاصيل بشاشة الإنجازات.",
		total, joinArabicAnd(parts))
	if err := s.notifRepo.CreateForRole("OWNER", "AI_ACHIEVEMENT_DIGEST", message); err != nil {
		return err
	}
	return s.notifRepo.CreateForRole("ADMIN", "AI_ACHIEVEMENT_DIGEST", message)
}
