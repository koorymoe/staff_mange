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
	notifications    *repository.NotificationRepository
	inventory        *repository.InventoryRepository
}

func NewBookingService(repo *repository.BookingRepository, employees *repository.EmployeeRepository, customers *repository.CustomerRepository, qualityFollowUps *repository.QualityFollowUpRepository, notifications *repository.NotificationRepository, inventory *repository.InventoryRepository) *BookingService {
	return &BookingService{repo: repo, employees: employees, customers: customers, qualityFollowUps: qualityFollowUps, notifications: notifications, inventory: inventory}
}

func (s *BookingService) List(status, customerID, date string) ([]model.Booking, error) {
	return s.repo.List(status, customerID, date)
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

	// الخدمات: نقبل قائمة (خدمات متعددة بنفس الحجز) وننزل على الخدمة المفردة
	// لو ما انرسلت قائمة — حتى أي شاشة قديمة تضل تشتغل بدون تعديل.
	serviceIDs := req.ServiceIDs
	if len(serviceIDs) == 0 && req.ServiceID != nil && *req.ServiceID != "" {
		serviceIDs = []string{*req.ServiceID}
	}
	primaryService := req.ServiceID
	if len(serviceIDs) > 0 {
		primaryService = &serviceIDs[0]
	}

	b := &model.Booking{
		ID:                 uuid.NewString(),
		Code:               fmt.Sprintf("B%d", seq),
		SequenceNumber:     &seq,
		CustomerID:         req.CustomerID,
		ServiceID:          primaryService,
		Notes:              req.Notes,
		VehicleType:        req.VehicleType,
		Priority:           priority,
		TransferEmployeeID: req.TransferEmployeeID,
		Address:            req.Address,
		MapLatitude:        req.MapLatitude,
		MapLongitude:       req.MapLongitude,
		LocationUrl:        req.LocationUrl,
	}

	if err := s.repo.Create(b); err != nil {
		return nil, err
	}

	if len(serviceIDs) > 0 {
		if err := s.repo.SetServices(b.ID, serviceIDs); err != nil {
			return nil, err
		}
	}

	if req.Address != nil || req.MapLatitude != nil || req.MapLongitude != nil {
		_ = s.customers.UpdateLocation(req.CustomerID, req.Address, req.MapLatitude, req.MapLongitude)
	}

	// نوسم الزبون تلقائياً بالخدمة الي طلبها (جي بي اس، كاميرات...) حتى يظهر
	// بقائمة "زبائن هذي الخدمة" لاحقاً، بدون أي كود منفصل — نفس كود الزبون الموحّد.
	// نوسم الزبون بكل خدمة طلبها بهذا الحجز، مو بالخدمة الرئيسية بس
	for i := range serviceIDs {
		_ = s.customers.EnsureServiceTag(req.CustomerID, &serviceIDs[i])
	}

	return s.repo.FindByID(b.ID)
}

func (s *BookingService) Confirm(id string, req model.ConfirmBookingRequest) (*model.Booking, error) {
	if req.QuotedPrice != nil && *req.QuotedPrice < 0 {
		return nil, errors.New("المبلغ المقدّر ما يصير يكون بالسالب")
	}
	if err := s.repo.Confirm(id, req, req.ScheduledAt); err != nil {
		return nil, err
	}
	booking, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	// نذكّر إداري الكوادر بالحجز المثبّت حديثاً حتى يوجّه الكادر المناسب له —
	// نفس فكرة تذكير الكي بي اي، بس هذا يخص تنسيق العمل مو تقييم الأداء.
	if booking != nil && s.notifications != nil {
		customerName := ""
		if booking.Customer != nil {
			customerName = booking.Customer.Name
		}
		_ = s.notifications.CreateForRole("HR_COORDINATOR", "booking_confirmed",
			fmt.Sprintf("📌 حجز جديد مثبّت (%s) للزبون %s — يحتاج تحديد الكادر المناسب له", booking.Code, customerName))
	}
	return booking, nil
}

func (s *BookingService) UpdateDetails(id string, req model.UpdateBookingDetailsRequest, editorID string) (*model.Booking, error) {
	if req.QuotedPrice != nil && *req.QuotedPrice < 0 {
		return nil, errors.New("المبلغ المقدّر ما يصير يكون بالسالب")
	}
	// تعديل قائمة الخدمات (لو انرسلت) — الزبون ممكن يزيد منظومة أو يشيل وحدة
	if req.ServiceIDs != nil {
		if err := s.repo.SetServices(id, req.ServiceIDs); err != nil {
			return nil, err
		}
	}
	if err := s.repo.UpdateDetails(id, req); err != nil {
		return nil, err
	}
	_ = s.repo.TouchLastEdited(id, editorID)
	return s.repo.FindByID(id)
}

// IsAssignedTo هل الموظف طرف بهذا الحجز (مكلّف/مشرف/مسؤول مصاريف/رحّله).
func (s *BookingService) IsAssignedTo(bookingID, employeeID string) (bool, error) {
	return s.repo.IsAssignedTo(bookingID, employeeID)
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
func (s *BookingService) Assign(id string, req model.AssignBookingRequest, editorID string) (*model.Booking, error) {
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

	_ = s.repo.TouchLastEdited(id, editorID)
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

// StartWithToolsCheck تسجّل (لو انطلبت) لقطة الأدوات الشخصية الناقصة عند الموظف
// بلحظة استلامه الحجز، وبعدين تكمل نفس منطق Start العادي (تلف الدالة، ما تكررها).
// لو الموظف ماله أدوات شخصية مسجّلة أصلاً بالنظام، ما نسجل أي لقطة (نتجنب صف فاضي
// بلا فايدة) ونكمل الاستلام عادي — القرار: تخطي الفحص بالكامل بدل ما نجبر الموظف
// يواجه واجهة فاضية أو معطلة.
func (s *BookingService) StartWithToolsCheck(id, employeeID string, missingToolIDs []string) (*model.Booking, error) {
	if employeeID != "" && s.inventory != nil {
		allTools, err := s.inventory.ListPersonalTools(employeeID)
		if err == nil && len(allTools) > 0 && len(missingToolIDs) > 0 {
			missingTools, err := s.inventory.ListPersonalToolsByIDs(missingToolIDs)
			if err == nil && len(missingTools) > 0 {
				names := make([]string, 0, len(missingTools))
				for _, t := range missingTools {
					names = append(names, t.Name)
				}
				joined := strings.Join(names, "، ")
				_, _ = s.inventory.CreateBookingToolCheck(id, employeeID, &joined)
			}
		}
	}
	return s.Start(id)
}

// ListToolChecks ترجّع لقطات الأدوات الناقصة المسجّلة عند استلام حجز معيّن.
func (s *BookingService) ListToolChecks(bookingID string) ([]model.BookingToolCheck, error) {
	if s.inventory == nil {
		return []model.BookingToolCheck{}, nil
	}
	return s.inventory.ListBookingToolChecks(bookingID)
}

// MarkConfirmationContacted تسجّل لحظة "تم" ضغطها الإداري بعد ما تواصل فعلاً مع
// الزبون قبل تثبيت الحجز — خطوة سابقة ومنفصلة عن Confirm نفسها، حتى يقدر المراقب
// (صلاحية crew_management) يدقق هل صار التواصل قبل التثبيت الفعلي أو لا.
// ReturnToCrew يرجّع حجز محوّل لإدارة المشاريع رجعة لكادر الشد، وينبّه
// إداري الكوادر إنه رجع له حتى ما يضيع بلا متابعة.
func (s *BookingService) ReturnToCrew(id string, note *string) (*model.Booking, error) {
	if err := s.repo.ReturnToCrew(id, note); err != nil {
		return nil, err
	}
	booking, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if booking != nil && s.notifications != nil {
		customerName := ""
		if booking.Customer != nil {
			customerName = booking.Customer.Name
		}
		reason := ""
		if note != nil && *note != "" {
			reason = " — السبب: " + *note
		}
		_ = s.notifications.CreateForRole("HR_COORDINATOR", "booking_returned_to_crew",
			fmt.Sprintf("↩️ الحجز (%s) للزبون %s رجع من إدارة المشاريع لكادر الشد%s",
				booking.Code, customerName, reason))
	}
	return booking, nil
}

func (s *BookingService) MarkConfirmationContacted(id, employeeID string) (*model.Booking, error) {
	if err := s.repo.MarkConfirmationContacted(id, employeeID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

// MarkArrived يسجّل وصول الفنيين لموقع الزبون — خطوة سابقة لبدء العمل، تُستخدم كبديل
// عن startedAt عند حساب عيّنات مدة العمل التاريخية لو لم يُضغط "بدأ العمل" بشكل منفصل.
func (s *BookingService) MarkArrived(id string) (*model.Booking, error) {
	if err := s.repo.MarkArrived(id); err != nil {
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
