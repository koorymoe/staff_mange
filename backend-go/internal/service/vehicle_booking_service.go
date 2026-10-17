package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type VehicleBookingService struct {
	repo *repository.VehicleBookingRepository
}

func NewVehicleBookingService(repo *repository.VehicleBookingRepository) *VehicleBookingService {
	return &VehicleBookingService{repo: repo}
}

func formatOverlapWindow(b model.VehicleBooking) string {
	return fmt.Sprintf("من %s إلى %s", b.StartAt.Format("2006-01-02 15:04"), b.EndAt.Format("2006-01-02 15:04"))
}

func (s *VehicleBookingService) CreateBooking(requestedByID string, req model.CreateVehicleBookingRequest) (*model.VehicleBooking, error) {
	if req.VehicleID == "" {
		return nil, errors.New("السيارة مطلوبة")
	}
	if req.Purpose == "" {
		return nil, errors.New("سبب الحجز مطلوب")
	}
	if req.StartAt == "" || req.EndAt == "" {
		return nil, errors.New("وقت بداية ونهاية الحجز مطلوبان")
	}

	overlaps, err := s.repo.FindOverlapping(req.VehicleID, req.StartAt, req.EndAt, nil)
	if err != nil {
		return nil, err
	}
	if len(overlaps) > 0 {
		return nil, fmt.Errorf("السيارة محجوزة مسبقاً بفترة متعارضة (%s)", formatOverlapWindow(overlaps[0]))
	}

	return s.repo.Create(requestedByID, req.VehicleID, req.Purpose, req.StartAt, req.EndAt)
}

func (s *VehicleBookingService) DecideBooking(bookingID, approverID string, req model.DecideVehicleBookingRequest) (*model.VehicleBooking, error) {
	booking, err := s.repo.Get(bookingID)
	if err != nil {
		return nil, errors.New("الحجز غير موجود")
	}
	if booking.Status != "PENDING" {
		return nil, errors.New("لا يمكن اتخاذ قرار بشأن حجز غير معلّق")
	}

	if req.Approve {
		overlaps, err := s.repo.FindOverlapping(booking.VehicleID, booking.StartAt.Format("2006-01-02T15:04:05"), booking.EndAt.Format("2006-01-02T15:04:05"), &booking.ID)
		if err != nil {
			return nil, err
		}
		for _, o := range overlaps {
			if o.Status == "APPROVED" {
				return nil, fmt.Errorf("لا يمكن الاعتماد — يوجد حجز معتمد آخر متعارض (%s)", formatOverlapWindow(o))
			}
		}
	} else {
		if req.RejectionReason == nil || *req.RejectionReason == "" {
			return nil, errors.New("سبب الرفض مطلوب")
		}
	}

	return s.repo.Decide(bookingID, approverID, req.Approve, req.RejectionReason)
}

func (s *VehicleBookingService) CancelBooking(bookingID, employeeID string, isAdmin bool) (*model.VehicleBooking, error) {
	booking, err := s.repo.Get(bookingID)
	if err != nil {
		return nil, errors.New("الحجز غير موجود")
	}
	if booking.RequestedByID != employeeID && !isAdmin {
		return nil, errors.New("لا تملك صلاحية إلغاء هذا الحجز")
	}
	if booking.Status != "PENDING" && booking.Status != "APPROVED" {
		return nil, errors.New("لا يمكن إلغاء هذا الحجز بحالته الحالية")
	}
	if !booking.StartAt.After(time.Now()) {
		return nil, errors.New("لا يمكن إلغاء حجز بدأ وقته أصلاً")
	}
	return s.repo.Cancel(bookingID)
}

func (s *VehicleBookingService) List(filters model.VehicleBookingFilters) ([]model.VehicleBooking, error) {
	return s.repo.List(filters)
}

// CheckApprovedBookingConflict يُستخدم عند بدء مهمة فعلية — يرجع رسالة تحذير لو
// يوجد حجز معتمد حالياً لموظف مختلف عن سائق المهمة (بدون منع صارم — تحذير فقط).
func (s *VehicleBookingService) CheckApprovedBookingConflict(vehicleID, driverID string) (string, error) {
	b, err := s.repo.FindApprovedCoveringNow(vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if b.RequestedByID == driverID {
		return "", nil
	}
	requesterName := "موظف آخر"
	if b.RequestedBy != nil {
		requesterName = b.RequestedBy.Name
	}
	return fmt.Sprintf("تنبيه: هذه السيارة محجوزة حالياً للموظف %s ضمن حجز معتمد", requesterName), nil
}
