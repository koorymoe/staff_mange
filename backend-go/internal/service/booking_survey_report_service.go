package service

import (
	"errors"
	"strings"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type BookingSurveyReportService struct {
	repo *repository.BookingSurveyReportRepository
}

func NewBookingSurveyReportService(repo *repository.BookingSurveyReportRepository) *BookingSurveyReportService {
	return &BookingSurveyReportService{repo: repo}
}

// Create يسجّل نتائج زيارة معاينة («كشف») — بلا حارس ورق محاسبي: هذا
// مو فاتورة ولا تقرير عمل، أي كادر مكلّف بالحجز يقدر يسلّمه.
func (s *BookingSurveyReportService) Create(employeeID string, req model.CreateBookingSurveyReportRequest) (*model.BookingSurveyReport, error) {
	if req.BookingID == "" {
		return nil, errors.New("bookingId مطلوب")
	}
	if strings.TrimSpace(req.CustomerWants) == "" {
		return nil, errors.New("اكتب شنو يريد الزبون")
	}
	return s.repo.Create(employeeID, req)
}

func (s *BookingSurveyReportService) List(bookingID string) ([]model.BookingSurveyReport, error) {
	if bookingID == "" {
		return nil, errors.New("bookingId مطلوب")
	}
	return s.repo.List(bookingID)
}
