package service

import (
	"fmt"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ ربط الأخطاء بفجوة تدريب حقيقية (ماتركس) ═══
//
// تكرار نوع إشارة معيّن (توقف عمل) عند موظفين ما أخذوا (أو ما نجحوا
// بـ) برنامج التدريب المطابق → اقتراح دورة تدريبية مو غرامة. يحوّل
// «منو غلط» لـ«شنو ناقص بالمنظومة».
const (
	trainingGapMinStops     = 3
	trainingGapLookbackDays = 30
)

type TrainingGapService struct {
	aiRepo    *repository.AiRepository
	notifRepo *repository.NotificationRepository
}

func NewTrainingGapService(aiRepo *repository.AiRepository, notifRepo *repository.NotificationRepository) *TrainingGapService {
	return &TrainingGapService{aiRepo: aiRepo, notifRepo: notifRepo}
}

// RunWeeklyIfDue تُستدعى دورياً (كل ٦ ساعات) — يوم الاثنين، نفس حيلة
// الاستكشاف الأسبوعي.
func (s *TrainingGapService) RunWeeklyIfDue() error {
	now := time.Now().In(debriefLoc)
	if now.Weekday() != time.Monday {
		return nil
	}
	weekKey := now.Format("2006-01-02")
	claimed, err := s.aiRepo.ClaimDailyMarker("WEEKLY_TRAINING_GAP_SENT", weekKey)
	if err != nil || !claimed {
		return err
	}

	rows, err := s.aiRepo.TrainingGapCandidates(trainingGapMinStops, trainingGapLookbackDays)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	lines := make([]string, 0, len(rows))
	for _, r := range rows {
		lines = append(lines, fmt.Sprintf("%s بخدمة %s (%d توقف)", r.EmployeeName, r.ServiceName, r.StopCount))
	}
	msg := fmt.Sprintf(
		"🎓 ماتركس — فجوة تدريب محتملة: %s. ما عندهم برنامج تدريب ناجح يغطّي هذي الخدمة — يمكن الفجوة بالتدريب مو تقصير.",
		joinArabicAnd(lines))
	if err := s.notifRepo.CreateForRole("OWNER", "AI_TRAINING_GAP", msg); err != nil {
		return err
	}
	return s.notifRepo.CreateForRole("ADMIN", "AI_TRAINING_GAP", msg)
}
