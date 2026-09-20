package service

import (
	"errors"
	"strings"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type AchievementService struct {
	repo     *repository.AchievementRepository
	bookings *repository.BookingRepository
}

func NewAchievementService(repo *repository.AchievementRepository, bookings *repository.BookingRepository) *AchievementService {
	return &AchievementService{repo: repo, bookings: bookings}
}

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
	return s.repo.Create(employeeID, role, req)
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
