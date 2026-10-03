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

var ErrSurveyNotParty = errors.New("تقرير الكشف يسلّمه الكادر المكلّف بالحجز بس")

// Create يسجّل نتائج زيارة معاينة («كشف») — بلا حارس ورق محاسبي: هذا
// مو فاتورة ولا تقرير عمل، أي كادر مكلّف بالحجز يقدر يسلّمه.
// ⚠️ چان أي موظف يسجّل كشف على أي حجز — هسه المكلّف بس، أو الإدارة.
func (s *BookingSurveyReportService) Create(employeeID, role string, req model.CreateBookingSurveyReportRequest) (*model.BookingSurveyReport, error) {
	if req.BookingID == "" {
		return nil, errors.New("bookingId مطلوب")
	}
	switch role {
	case "ADMIN", "OWNER", "HR_COORDINATOR", "MONITOR":
	default:
		if !s.repo.IsBookingParty(req.BookingID, employeeID) {
			return nil, ErrSurveyNotParty
		}
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
