package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type WorkReportService struct {
	repo *repository.WorkReportRepository
	// guard يقرّر منو يحق يسوي ورق الحجز — nil يعني بلا قاعدة (سلوك قديم).
	guard *PaperworkGuard
}

func NewWorkReportService(repo *repository.WorkReportRepository, guard *PaperworkGuard) *WorkReportService {
	return &WorkReportService{repo: repo, guard: guard}
}

func (s *WorkReportService) Create(employeeID, role string, req model.CreateWorkReportRequest) (*model.WorkReport, error) {
	if req.BookingID == "" {
		return nil, errors.New("bookingId مطلوب")
	}
	// ⚠️ بخدمات مؤشّرة، التقرير على **مسؤول الخدمة** مو على الفني.
	// وبقية الخدمات ما تنلمس — الحارس يمرّرها بلا أي فحص.
	if s.guard != nil {
		if err := s.guard.Check(req.BookingID, employeeID, role); err != nil {
			return nil, err
		}
	}
	if req.WorkStatus != "COMPLETED" && req.WorkStatus != "STOPPED" {
		return nil, errors.New("حالة العمل غير صحيحة")
	}
	if req.WorkStatus == "COMPLETED" && (req.Events == nil || *req.Events == "") {
		return nil, errors.New("يرجى كتابة تقرير الأحداث والمشاكل")
	}
	if req.WorkStatus == "STOPPED" && (req.StopReason == nil || *req.StopReason == "") {
		return nil, errors.New("يرجى كتابة سبب التوقف")
	}
	return s.repo.Create(employeeID, req)
}

// Review «تم التدقيق» من المراقب (ع 10-16).
func (s *WorkReportService) Review(id, byID, note string) error {
	return s.repo.Review(id, byID, note)
}

func (s *WorkReportService) List(employeeID string) ([]model.WorkReport, error) {
	return s.repo.List(employeeID)
}

func (s *WorkReportService) ListByReview(employeeID, review string) ([]model.WorkReport, error) {
	return s.repo.ListByReview(employeeID, review)
}

func (s *WorkReportService) ReviewCounts(employeeID string) (int, int, error) {
	return s.repo.ReviewCounts(employeeID)
}

// ByBookingCode تقارير العمل لحجز بكوده.
func (s *WorkReportService) ByBookingCode(code, employeeID string) (*repository.WorkReportBookingLookup, error) {
	return s.repo.ByBookingCode(code, employeeID)
}
