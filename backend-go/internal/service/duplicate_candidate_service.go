package service

import (
	"errors"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type DuplicateCandidateService struct {
	repo *repository.DuplicateCandidateRepository
}

func NewDuplicateCandidateService(repo *repository.DuplicateCandidateRepository) *DuplicateCandidateService {
	return &DuplicateCandidateService{repo: repo}
}

// RunScan يفحص الحجوزات والزبائن — تُستدعى دورياً من الحلقة الخلفية.
// صفوف جديدة بس تنضاف (ON CONFLICT DO NOTHING بالمستودع)، فتكرار
// التشغيل آمن ولا يعيد زوجاً سبق فحصه أو انأشّر "مو تكرار".
func (s *DuplicateCandidateService) RunScan() (bookings int64, customers int64, err error) {
	bookings, err = s.repo.ScanBookings()
	if err != nil {
		return 0, 0, err
	}
	customers, err = s.repo.ScanCustomers()
	if err != nil {
		return bookings, 0, err
	}
	return bookings, customers, nil
}

func (s *DuplicateCandidateService) List(kind, status string) ([]model.DuplicateCandidate, error) {
	return s.repo.List(kind, status)
}

func (s *DuplicateCandidateService) Dismiss(id, byEmployeeID string) error {
	if id == "" {
		return errors.New("id مطلوب")
	}
	return s.repo.Dismiss(id, byEmployeeID)
}
