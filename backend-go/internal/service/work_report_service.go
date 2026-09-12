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

func (s *WorkReportService) List(employeeID string) ([]model.WorkReport, error) {
	return s.repo.List(employeeID)
}
