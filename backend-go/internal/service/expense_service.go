package service

import (
	"errors"
	"strings"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type ExpenseService struct {
	repo *repository.ExpenseRepository
}

func NewExpenseService(repo *repository.ExpenseRepository) *ExpenseService {
	return &ExpenseService{repo: repo}
}

func (s *ExpenseService) List(employeeID string) ([]model.Expense, error) {
	return s.repo.List(employeeID)
}

func (s *ExpenseService) Create(req model.CreateExpenseRequest) (*model.Expense, error) {
	if req.EmployeeID == "" || req.Amount == nil {
		return nil, errors.New("employeeId and amount are required")
	}
	if *req.Amount < 0 {
		return nil, errors.New("مبلغ المصروف ما يصير يكون بالسالب")
	}

	// ═══ المصروف على حجز — والليدر لازم يكون مسؤولاً عنه ═══
	//
	// ⚠️⚠️ **القائمة بالواجهة مو حماية**: نداء مباشر بمعرّف حجز ثانٍ
	// يتخطّاها ويحمّل حجز زميله مصروفاً — ويطلع «نقص» بحجز ما صرف
	// عليه أحد.
	//
	// ⚠️ والحجز **اختياري** عمداً: خلّيناه إجبارياً چان انكسر تسجيل
	// المصاريف العامة (وقود السيارة، أدوات الورشة) الي ما تخص حجزاً.
	// الي ما إله حجز يبقى برّا حساب الحجوزات، وهذا الصحيح.
	var bookingID *string
	if req.BookingID != nil {
		if id := strings.TrimSpace(*req.BookingID); id != "" {
			ok, err := s.repo.IsExpenseResponsible(req.EmployeeID, id)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, errors.New("ما تگدر تسجّل مصروفاً على حجز مو مسؤول عن مصاريفه")
			}
			bookingID = &id
		}
	}
	return s.repo.Create(req.EmployeeID, *req.Amount, req.Description, bookingID)
}

func (s *ExpenseService) UpdateStatus(id string, req model.UpdateExpenseStatusRequest) (*model.Expense, error) {
	if req.Status != "APPROVED" && req.Status != "REJECTED" {
		return nil, errors.New("status must be APPROVED or REJECTED")
	}
	return s.repo.UpdateStatus(id, req.Status)
}
