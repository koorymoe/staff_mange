package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type WorkReportService struct {
	repo *repository.WorkReportRepository
}

func NewWorkReportService(repo *repository.WorkReportRepository) *WorkReportService {
	return &WorkReportService{repo: repo}
}

func (s *WorkReportService) Create(employeeID string, req model.CreateWorkReportRequest) (*model.WorkReport, error) {
	if req.BookingID == "" {
		return nil, errors.New("bookingId مطلوب")
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
