package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type BookingService struct {
	repo             *repository.BookingRepository
	employees        *repository.EmployeeRepository
	customers        *repository.CustomerRepository
	qualityFollowUps *repository.QualityFollowUpRepository
}

func NewBookingService(repo *repository.BookingRepository, employees *repository.EmployeeRepository, customers *repository.CustomerRepository, qualityFollowUps *repository.QualityFollowUpRepository) *BookingService {
	return &BookingService{repo: repo, employees: employees, customers: customers, qualityFollowUps: qualityFollowUps}
}

func (s *BookingService) List(status, customerID string) ([]model.Booking, error) {
	return s.repo.List(status, customerID)
}

func (s *BookingService) Create(req model.CreateBookingRequest) (*model.Booking, error) {
	if req.CustomerID == "" {
		return nil, errors.New("customerId is required")
	}

	seq, err := s.repo.NextSequenceNumber()
	if err != nil {
		return nil, err
	}

	priority := "NORMAL"
	if req.Priority != nil {
		priority = *req.Priority
	}

	b := &model.Booking{
		ID:                 uuid.NewString(),
		Code:               fmt.Sprintf("B%d", seq),
		SequenceNumber:     &seq,
		CustomerID:         req.CustomerID,
		ServiceID:          req.ServiceID,
		Notes:              req.Notes,
		VehicleType:        req.VehicleType,
		Priority:           priority,
		TransferEmployeeID: req.TransferEmployeeID,
		Address:            req.Address,
		MapLatitude:        req.MapLatitude,
		MapLongitude:       req.MapLongitude,
	}

	if err := s.repo.Create(b); err != nil {
		return nil, err
	}

	if req.Address != nil || req.MapLatitude != nil || req.MapLongitude != nil {
		_ = s.customers.UpdateLocation(req.CustomerID, req.Address, req.MapLatitude, req.MapLongitude)
	}

	// نوسم الزبون تلقائياً بالخدمة الي طلبها (جي بي اس، كاميرات...) حتى يظهر
	// بقائمة "زبائن هذي الخدمة" لاحقاً، بدون أي كود منفصل — نفس كود الزبون الموحّد.
	_ = s.customers.EnsureServiceTag(req.CustomerID, req.ServiceID)

	return s.repo.FindByID(b.ID)
}

func (s *BookingService) Confirm(id string, req model.ConfirmBookingRequest) (*model.Booking, error) {
	if req.QuotedPrice != nil && *req.QuotedPrice < 0 {
		return nil, errors.New("المبلغ المقدّر ما يصير يكون بالسالب")
	}
	if err := s.repo.Confirm(id, req, req.ScheduledAt); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *BookingService) UpdateDetails(id string, req model.UpdateBookingDetailsRequest) (*model.Booking, error) {
	if req.QuotedPrice != nil && *req.QuotedPrice < 0 {
		return nil, errors.New("المبلغ المقدّر ما يصير يكون بالسالب")
	}
	if err := s.repo.UpdateDetails(id, req); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *BookingService) ScheduleLog(id string) ([]model.ScheduleChangeLog, error) {
	return s.repo.ScheduleLog(id)
}

func (s *BookingService) SetSchedule(id, changedByID, scheduledAt string) (*model.Booking, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("Booking not found")
	}

	var oldTime *string
	if existing.ScheduledAt != nil {
		s := existing.ScheduledAt.Format("2006-01-02T15:04:05")
		oldTime = &s
	}

	if err := s.repo.SetSchedule(id, scheduledAt); err != nil {
		return nil, err
	}
	if changedByID != "" {
		if err := s.repo.CreateScheduleLog(id, changedByID, oldTime, scheduledAt); err != nil {
			return nil, err
		}
	}
	return s.repo.FindByID(id)
}

// Assign يعيّن فني لمهمة الحجز، يتحقق من المهارة والدوام، ويحدد المسؤول عن المصاريف تلقائياً
func (s *BookingService) Assign(id string, req model.AssignBookingRequest) (*model.Booking, error) {
	if req.EmployeeID == "" || req.Role == "" {
		return nil, errors.New("employeeId and role are required")
	}

	booking, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, errors.New("Booking not found")
	}

	skillWarning := false
	if booking.ServiceID != nil {
		hasSkill, err := s.repo.EmployeeHasSkillForService(req.EmployeeID, *booking.ServiceID)
		if err != nil {
			return nil, err
		}
		skillWarning = !hasSkill
	}

	employee, err := s.employees.FindByID(req.EmployeeID)
	if err != nil || employee == nil {
		return nil, errors.New("الموظف غير موجود")
	}
	if !employee.OnDuty || employee.Status != "ACTIVE" {
		return nil, errors.New("هذا الموظف غير متاح حالياً (خارج الدوام)")
	}

	if err := s.repo.UpsertAssignment(id, req.EmployeeID, req.Role); err != nil {
		return nil, err
	}

	if req.AssignedVehicle != nil {
		if err := s.repo.SetAssignedVehicle(id, *req.AssignedVehicle); err != nil {
			return nil, err
		}
	}

	if skillWarning {
		warning := fmt.Sprintf("⚠️ تنبيه: الموظف %s لا يمتلك المهارة اللازمة لهذه الخدمة", employee.Name)
		existingNotes := ""
		if booking.AdminNotes != nil {
			existingNotes = *booking.AdminNotes
		}
		if existingNotes == "" || !strings.Contains(existingNotes, warning) {
			combined := warning
			if existingNotes != "" {
				combined = existingNotes + "\n" + warning
			}
			if err := s.repo.SetAdminNotes(id, combined); err != nil {
				return nil, err
			}
		}
	}

	// تعيين المسؤول عن المصاريف تلقائياً: الليدر أولاً، وإلا الفني الوحيد المعيّن
	if booking.ExpenseResponsibleID == nil {
		if employee.IsLeader {
			if err := s.repo.SetExpenseResponsible(id, req.EmployeeID); err != nil {
				return nil, err
			}
		} else {
			assignments, err := s.repo.ListAssignments(id)
			if err == nil {
				hasLeader := false
				for _, a := range assignments {
					emp, _ := s.employees.FindByID(a.EmployeeID)
					if emp != nil && emp.IsLeader {
						hasLeader = true
						break
					}
				}
				if !hasLeader && len(assignments) == 1 {
					if err := s.repo.SetExpenseResponsible(id, assignments[0].EmployeeID); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	return s.repo.FindByID(id)
}

func (s *BookingService) SetSupervisor(id string, employeeID *string) (*model.Booking, error) {
	if employeeID != nil {
		employee, err := s.employees.FindByID(*employeeID)
		if err != nil || employee == nil || (employee.Role != "PROJECT_MANAGER" && !employee.IsLeader) {
			return nil, errors.New("يجب أن يكون تيم ليدر أو مدير مشاريع")
		}
	}
	if err := s.repo.SetSupervisor(id, employeeID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *BookingService) Start(id string) (*model.Booking, error) {
	if err := s.repo.StartWithResponseTime(id); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

// SetMaterialsReady يسمح فقط لتيم ليدر مسند لهذا الحجز (أو أدمن/مراقب) بتأكيد تجهيز
// المواد — لحظة الضغط تصير بداية عدّاد استجابة الفنيين.
func (s *BookingService) SetMaterialsReady(id, employeeID string) (*model.Booking, error) {
	employee, err := s.employees.FindByID(employeeID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, errors.New("الموظف غير موجود")
	}
	if !employee.IsLeader && employee.Role != "ADMIN" && employee.Role != "MONITOR" {
		return nil, errors.New("هذا الإجراء يقتصر على تيم ليدر الفريق أو الإدارة")
	}
	if err := s.repo.SetMaterialsReady(id, employeeID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *BookingService) Complete(id string, req model.CompleteBookingRequest) (*model.Booking, error) {
	if req.AmountCollected != nil && *req.AmountCollected < 0 {
		return nil, errors.New("المبلغ المحصّل ما يصير يكون بالسالب")
	}
	if req.AdvancePaid != nil && *req.AdvancePaid < 0 {
		return nil, errors.New("الدفعة المقدمة ما يصير تكون بالسالب")
	}
	if err := s.repo.Complete(id, req); err != nil {
		return nil, err
	}
	booking, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	// يترحل الحجز تلقائياً لمهندس الجودة يتواصل مع الزبون ويتأكد ما اكو مشاكل —
	// فشل هذا الترحيل ما يوقف إكمال الحجز نفسه (ثانوي).
	if booking != nil {
		_ = s.qualityFollowUps.CreateForBooking(booking.ID, booking.CustomerID)
	}
	return booking, nil
}

func (s *BookingService) Verify(id string) (*model.Booking, error) {
	if err := s.repo.Verify(id); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}
