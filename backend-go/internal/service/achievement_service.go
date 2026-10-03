package service

import (
	"errors"
	"log"
	"strings"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// BookingCrewSizer يفصل الإنجازات عن نواة الذكاء — نفس فلسفة
// `AiSignalRecorder`. `AiRepository` يحقّقها أصلاً (`LatestVisitCrewSize`).
type BookingCrewSizer interface {
	LatestVisitCrewSize(bookingID string) (int, error)
}

type AchievementService struct {
	repo     *repository.AchievementRepository
	bookings *repository.BookingRepository
	// ai/crew: اختياريان. يوصلون SELF_REPORT_MISMATCH لمّا التقرير
	// يدّعي عملاً لحاله وكادر آخر طلعة الحقيقي أكثر من واحد.
	ai   AiSignalRecorder
	crew BookingCrewSizer
}

func NewAchievementService(repo *repository.AchievementRepository, bookings *repository.BookingRepository) *AchievementService {
	return &AchievementService{repo: repo, bookings: bookings}
}

// SetAiRecorder يربط مسجّل إشارات ماتركس بعد البناء.
func (s *AchievementService) SetAiRecorder(a AiSignalRecorder) { s.ai = a }

// SetCrewSizer يربط قارئ كادر الحجز — بدونه ماكو فحص تناقض (صمت آمن).
func (s *AchievementService) SetCrewSizer(c BookingCrewSizer) { s.crew = c }

func (s *AchievementService) Create(employeeID, role string, req model.CreateAchievementRequest) (*model.Achievement, error) {
	req.ReportText = strings.TrimSpace(req.ReportText)
	if req.ReportText == "" {
		return nil, errors.New("يرجى كتابة تقرير الإنجاز")
	}
	if req.BookingID != nil && *req.BookingID != "" {
		b, err := s.bookings.FindByID(*req.BookingID)
		if err != nil || b == nil {
			return nil, errors.New("الحجز المربوط غير موجود")
		}
	} else {
		req.BookingID = nil
	}
	a, err := s.repo.Create(employeeID, role, req)
	if err != nil {
		return nil, err
	}
	s.checkSelfReportMismatch(a, employeeID)
	return a, nil
}

// checkSelfReportMismatch يفحص ادّعاء «وحدي/لحالي» مقابل كادر آخر
// طلعة الحقيقي — نفس الجملة الصريحة الي يعيد جامع الأدلة اكتشافها.
// ⚠️ فحص إضافي هنا مو الوحيد: لو فشل (ماكو `ai` أو خطأ قراءة)، ماكو
// إشارة، وهذا آمن — صمت أحسن من إشارة مبنية على تخمين.
func (s *AchievementService) checkSelfReportMismatch(a *model.Achievement, employeeID string) {
	if s.ai == nil || s.crew == nil || a.BookingID == nil {
		return
	}
	if detectSoloClaim(a.ReportText) == "" {
		return
	}
	crewSize, err := s.crew.LatestVisitCrewSize(*a.BookingID)
	if err != nil || crewSize <= 1 {
		return // ماكو تناقض حقيقي — الكادر فعلاً واحد أو ماكو بيانات كافية
	}
	if _, err := s.ai.RecordSignal(model.AiSignal{
		Kind:       model.AiSignalSelfReportMismatch,
		EntityType: "ACHIEVEMENT",
		EntityID:   a.ID,
		EmployeeID: &employeeID,
	}); err != nil {
		log.Printf("[ai] تعذر تسجيل إشارة تناقض التقرير الذاتي للإنجاز %s: %v", a.ID, err)
	}
}

func (s *AchievementService) List(employeeID, day string, limit int) ([]model.Achievement, error) {
	return s.repo.List(employeeID, day, limit)
}

func (s *AchievementService) Review(id, reviewerID string, req model.ReviewAchievementRequest) (*model.Achievement, error) {
	if req.Status != model.AchievementReviewGood && req.Status != model.AchievementReviewNeeds {
		return nil, errors.New("حالة المراجعة غير صحيحة")
	}
	return s.repo.Review(id, reviewerID, req.Status, req.Note)
}
