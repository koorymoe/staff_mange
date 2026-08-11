package service

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/timeutil"
)

type BookingService struct {
	repo             *repository.BookingRepository
	employees        *repository.EmployeeRepository
	customers        *repository.CustomerRepository
	qualityFollowUps *repository.QualityFollowUpRepository
	notifications    *repository.NotificationRepository
	inventory        *repository.InventoryRepository
	// discipline: فحص عدالة التوزيع وقت التعيين. اختياري — لو ما انربط
	// النظام يشتغل عادي بلا غرامات.
	discipline AssignmentBalanceChecker
	// solar: سعر المنظومة لحجز الطاقة الشمسية. اختياري.
	solar SolarPricer
	// monitor: صندوق المراقب. اختياري — بدونه النظام يشتغل عادي بس
	// المراقب ما يوصله شي.
	monitor MonitorFeed
	// ai: مسجّل إشارات التحليل. اختياري — بدونه النظام يشتغل عادي
	// بس ماكو تحليل يتراكم.
	ai AiSignalRecorder
}

// AiSignalRecorder يفصل خدمة الحجوزات عن نواة الذكاء الاصطناعي حتى
// ما يصير اعتماد متبادل بين الاثنين.
type AiSignalRecorder interface {
	RecordSignal(model.AiSignal) (*model.AiSignal, error)
}

// SetAiRecorder يربط مسجّل الإشارات بعد البناء.
func (s *BookingService) SetAiRecorder(a AiSignalRecorder) { s.ai = a }

// SetMonitorFeed يربط صندوق المراقب بعد بناء الخدمتين.
func (s *BookingService) SetMonitorFeed(m MonitorFeed) { s.monitor = m }

// AssignmentBalanceChecker يفصل خدمة الحجوزات عن خدمة الانضباط حتى ما
// يصير اعتماد دائري بين الاثنين.
type AssignmentBalanceChecker interface {
	CheckAssignmentBalance(adminID, assignedLeaderID, bookingID, bookingCode string, activeByLeader map[string]int, leaderNames map[string]string)
}

// SolarPricer يعطي سعر المنظومة بدون ما تعتمد خدمة الحجوزات على
// مستودع الطاقة الشمسية كاملاً.
type SolarPricer interface {
	SystemTotalPrice(systemID string) (float64, error)
}

// SetSolarPricer يربط تسعير المنظومات.
func (s *BookingService) SetSolarPricer(p SolarPricer) { s.solar = p }

// SetDisciplineChecker يربط فحص عدالة التوزيع بعد بناء الخدمتين.
func (s *BookingService) SetDisciplineChecker(c AssignmentBalanceChecker) {
	s.discipline = c
}

func NewBookingService(repo *repository.BookingRepository, employees *repository.EmployeeRepository, customers *repository.CustomerRepository, qualityFollowUps *repository.QualityFollowUpRepository, notifications *repository.NotificationRepository, inventory *repository.InventoryRepository) *BookingService {
	return &BookingService{repo: repo, employees: employees, customers: customers, qualityFollowUps: qualityFollowUps, notifications: notifications, inventory: inventory}
}

func (s *BookingService) List(status, customerID, date string, limit int) ([]model.Booking, error) {
	return s.repo.List(status, customerID, date, limit)
}

// ListAssignedTo يرجّع حجوزات الموظف المعيّن عليها فقط. حد أعلى ٢٠٠
// حجز — الفني ما يحتاج أرشيفه كامل بلوحة المهام، ويمنع طلب واحد ثقيل
// لو موظف قديم عليه آلاف المهام.
func (s *BookingService) ListAssignedTo(employeeID string) ([]model.Booking, error) {
	if employeeID == "" {
		return []model.Booking{}, nil
	}
	return s.repo.ListForAssignedEmployee(employeeID, 200)
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

	// حجز داخل الشركة: نثبت workLocation من هنا حتى إحصائية «الأعمال
	// داخل الشركة» تلقفه بلا ما ننتظر أحد يأشرها وقت الإنجاز.
	b.BookingType = "REGULAR"
	b.WorkLocation = model.WorkOnSite
	if req.BookingType != nil && *req.BookingType != "" {
		b.BookingType = *req.BookingType
	}
	// ═══ حجز طاقة شمسية ═══
	// المنظومة تجي من الكتالوك، فالسعر المقدّر ما ينكتب بالإيد — ينحسب
	// من مكوّناتها بأسعار المخزن اليوم. المبيعات ما يعرف أسعار المكوّنات
	// أصلاً، ولو خلّيناه يقدّر بالإيد يطلع رقم ما إله علاقة بالكلفة.
	if b.BookingType == "SOLAR" {
		b.SolarSystemID = req.SolarSystemID
		b.SolarMonthlyKwh = req.SolarMonthlyKwh
		if req.SolarSystemID != nil && *req.SolarSystemID != "" && s.solar != nil {
			if price, err := s.solar.SystemTotalPrice(*req.SolarSystemID); err == nil && price > 0 {
				b.QuotedPrice = &price
			}
		}
	}

	if b.BookingType == "INTERNAL" {
		b.WorkLocation = model.WorkInHouse
		b.InternalEmployeeName = req.InternalEmployeeName
		b.InternalEmployeePhone = req.InternalEmployeePhone
		b.InternalDepartment = req.InternalDepartment
		b.InternalApproved = req.InternalApproved
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
	// نفس قصة التوقيت: الموعد المكتوب بالإيد بغداد → عالمي قبل التخزين
	if req.ScheduledAt != nil {
		normalized := timeutil.NormalizeCompanyLocal(*req.ScheduledAt)
		req.ScheduledAt = &normalized
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
	if s.monitor != nil {
		s.monitor.BookingStage(model.MonitorStageBookingAfterConfirm, booking, "HR_COORDINATOR", nil)
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

// IsCartItemOfAssignedBooking هل عنصر السلة يتبع حجز الموظف طرف بيه.
func (s *BookingService) IsCartItemOfAssignedBooking(cartItemID, employeeID string) (bool, error) {
	return s.repo.IsCartItemOfAssignedBooking(cartItemID, employeeID)
}

// IsAssignedTo هل الموظف طرف بهذا الحجز (مكلّف/مشرف/مسؤول مصاريف/رحّله).
func (s *BookingService) IsAssignedTo(bookingID, employeeID string) (bool, error) {
	return s.repo.IsAssignedTo(bookingID, employeeID)
}

func (s *BookingService) ScheduleLog(id string) ([]model.ScheduleChangeLog, error) {
	return s.repo.ScheduleLog(id)
}

func (s *BookingService) SetSchedule(id, changedByID, scheduledAt string) (*model.Booking, error) {
	// الموظف يكتب الوقت بتوقيت بغداد، والنظام يخزن بالتوقيت العالمي.
	// بدونها الموعد يتقدم ثلاث ساعات كل ما ينحفظ.
	scheduledAt = timeutil.NormalizeCompanyLocal(scheduledAt)

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
	updated, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	// المراقب يشوف الحجز وموعده **قبل** التثبيت — هذا الوقت الي ينفع
	// بيه الاعتراض، بعد التثبيت يصير تصليح مو منع.
	if s.monitor != nil {
		s.monitor.BookingStage(model.MonitorStageBookingBeforeConfirm, updated, "HR_COORDINATOR", ptrOrNil(changedByID))
	}
	return updated, nil
}

// ptrOrNil يحوّل النص الفارغ لـnil حتى ما ننحفظ معرّف موظف فاضي.
func ptrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Assign يعيّن فني لمهمة الحجز، يتحقق من المهارة والدوام، ويحدد المسؤول عن المصاريف تلقائياً
// ensureNotProjectLocked يمنع تنسيق حجز لسه عند إدارة المشاريع.
//
// ⚠️ الإقفال لازم يكون هنا مو بالواجهة بس: زر مخفي ينلتف عليه بنداء
// API مباشر، والحجز يتنسّق وكادر الشد يتكلّف قبل ما المشروع يخلّص
// إجراءاته أصلاً.
func (s *BookingService) ensureNotProjectLocked(id string) error {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if b == nil {
		return errors.New("Booking not found")
	}
	if b.TransferToProjects && b.ProjectExecutionAt == nil {
		return errors.New("هذا الحجز عند إدارة المشاريع — ما تكدر تنسّقه لحد ما يوصل مرحلة التنفيذ")
	}
	return nil
}

func (s *BookingService) Assign(id string, req model.AssignBookingRequest, editorID string) (*model.Booking, error) {
	if err := s.ensureNotProjectLocked(id); err != nil {
		return nil, err
	}
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

	if err := s.repo.UpsertAssignment(id, req.EmployeeID, req.Role, editorID); err != nil {
		return nil, err
	}

	// عدالة التوزيع: لو الإداري كلّف ليدر عنده شغل وبنفس الوقت أكو ليدر
	// فاضي تماماً، النظام يغرّمه نقطة. الفحص يصير بعد التعيين مباشرة —
	// وقتها بس نعرف منو انكلّف فعلاً.
	if s.discipline != nil && employee.IsLeader {
		if counts, names, err := s.repo.ActiveCountByLeader(); err == nil {
			// نشيل هذا الحجز من العدّ حتى ما نحسبه على المكلَّف تواً
			if counts[req.EmployeeID] > 0 {
				counts[req.EmployeeID]--
			}
			s.discipline.CheckAssignmentBalance(editorID, req.EmployeeID, id, booking.Code, counts, names)
		}
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
	if err := s.ensureNotProjectLocked(id); err != nil {
		return nil, err
	}
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

// MarkProjectExecution يفتح حجز المشاريع للتنسيق — تناديها خدمة
// المشاريع لما المشروع يوصل مرحلة التنفيذ.
func (s *BookingService) MarkProjectExecution(bookingID string) error {
	if err := s.repo.MarkProjectExecution(bookingID); err != nil {
		return err
	}
	b, err := s.repo.FindByID(bookingID)
	if err != nil || b == nil || s.notifications == nil {
		return nil
	}
	customerName := ""
	if b.Customer != nil {
		customerName = b.Customer.Name
	}
	// بدون الإشعار الحجز ينفتح بالسكوت والمنسّق ما يدري — يعني نفس
	// المشكلة القديمة بس بشكل ثاني.
	_ = s.notifications.CreateForRole("HR_COORDINATOR", "project_booking_ready",
		fmt.Sprintf("🏗️ حجز المشاريع %s (%s) وصل مرحلة التنفيذ — جاهز للتنسيق بكادر الشد", b.Code, customerName))
	return nil
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
	// الحجز رجع لكادر الشد — ما يبقى عليه أثر مشروع انلغى
	_ = s.repo.ClearProjectExecution(id)
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

// StopWork يوقف العمل بسبب مكتوب. السبب إجباري — «توقف» بلا سبب ما
// تنفع لا للمتابعة ولا للتقرير.
func (s *BookingService) StopWork(id, reason, employeeID string) (*model.Booking, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, errors.New("سبب توقف العمل مطلوب")
	}
	if err := s.repo.StopWork(id, strings.TrimSpace(reason), employeeID); err != nil {
		return nil, err
	}
	// إشارة للتحليل: «ليش وقّف الشغل؟».
	//
	// ⚠️ تنسجّل وبس — التحليل يصير بالخلفية. ربط التحليل بهاي اللحظة
	// چان يخلي بطؤه (أو فشله) يعطّل موظف واقف بموقع الزبون.
	// ⚠️ وفشلها ما يرجّع خطأ: توقف العمل انحفظ فعلاً، وإفشال الطلب
	// بسبب إشارة تحليل يخلي الموظف يعيد المحاولة بلا فايدة.
	if s.ai != nil {
		if _, err := s.ai.RecordSignal(model.AiSignal{
			Kind:       model.AiSignalWorkStopped,
			EntityType: "BOOKING",
			EntityID:   id,
			EmployeeID: &employeeID,
		}); err != nil {
			log.Printf("[ai] تعذر تسجيل إشارة توقف العمل للحجز %s: %v", id, err)
		}
	}
	return s.repo.FindByID(id)
}

// ResumeWork يرجّع الحجز شغّال بعد ما ينحل سبب التوقف.
func (s *BookingService) ResumeWork(id string) (*model.Booking, error) {
	if err := s.repo.ResumeWork(id); err != nil {
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
	// بعد الإنجاز: المراقب يشوف شنو انعمل فعلاً قبل ما تصير فاتورة.
	if s.monitor != nil {
		s.monitor.BookingStage(model.MonitorStageBookingAfterComplete, booking, "TECHNICIAN", booking.ProjectSupervisorID)
	}
	return booking, nil
}

func (s *BookingService) Verify(id string) (*model.Booking, error) {
	if err := s.repo.Verify(id); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

// Unverify يفتح الحجز للتدقيق من جديد — لمدير النظام حصراً.
func (s *BookingService) Unverify(id string) (*model.Booking, error) {
	if err := s.repo.Unverify(id); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

// ═══ الأرشيف والتأجيل والانتظار ═══

// ListArchived الحجوزات المحذوفة — موجودة بالأرشيف مو ممحية.
func (s *BookingService) ListArchived(limit int) ([]model.Booking, error) {
	return s.repo.ListArchived(limit)
}

// Archive يحذف الحجز من الشاشات العاملة ويحفظه بالأرشيف.
func (s *BookingService) Archive(id, byEmployeeID, reason string) (*model.Booking, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, errors.New("سبب الحذف مطلوب")
	}
	if err := s.repo.Archive(id, byEmployeeID, reason); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

// Restore يرجّع حجز من الأرشيف للعمل.
func (s *BookingService) Restore(id string) (*model.Booking, error) {
	if err := s.repo.Restore(id); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

// Postpone يأجّل موعد الحجز. الموعد يجي بتوقيت بغداد وينخزن عالمي —
// بدون التحويل الموعد يتقدم ثلاث ساعات كل ما ينحفظ.
func (s *BookingService) Postpone(id, newTime, reason, byEmployeeID string) (*model.Booking, error) {
	if err := s.ensureNotProjectLocked(id); err != nil {
		return nil, err
	}
	// الموعد اختياري: أكثر التأجيلات تصير والزبون ما محدّد متى يناسبه،
	// والإداري كان يضطر يحط تاريخ من راسه حتى يمرّر الشاشة — فيطلع
	// موعد كذب بالجدول والكادر يتحضّر لحجز ماكو.
	//
	// ⚠️ السبب يضل إجباري: «تأجل» بلا سبب ما تنفع لا للإحصاء ولا
	// للكادر الي راح يسأل ليش.
	newTime = strings.TrimSpace(newTime)
	if strings.TrimSpace(reason) == "" {
		return nil, errors.New("سبب التأجيل مطلوب")
	}
	if newTime != "" {
		newTime = timeutil.NormalizeCompanyLocal(newTime)
	}
	if err := s.repo.Postpone(id, newTime, reason, byEmployeeID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

// ListPostponed الحجوزات المؤجلة بلا موعد — طابور قرارات الإداري.
func (s *BookingService) ListPostponed() ([]model.Booking, error) {
	return s.repo.ListPostponed()
}

// MarkWaiting يحط الحجز بحالة «في الانتظار» — اتصلنا بالزبون وما رد.
func (s *BookingService) MarkWaiting(id, note, byEmployeeID string) (*model.Booking, error) {
	if err := s.ensureNotProjectLocked(id); err != nil {
		return nil, err
	}
	if err := s.repo.MarkWaiting(id, note, byEmployeeID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

// ResumeFromWaiting يرجّع الحجز من الانتظار — الزبون رد.
func (s *BookingService) ResumeFromWaiting(id string) (*model.Booking, error) {
	if err := s.repo.ResumeFromWaiting(id); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

// ChangeType تغيير نوع الحجز — للمالك ومدير النظام (مقيّد بالراوتر).
// يرجّع الحجز بعد التغيير حتى الواجهة تحدّث الصف بلا إعادة تحميل.
func (s *BookingService) ChangeType(id, newType, byEmployeeID string) (*model.Booking, error) {
	if err := s.repo.ChangeType(id, newType, byEmployeeID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

// ═══ تتبّع المراحل ═══

// RecordCreator يسجّل منو أدخل الحجز — ينندى بعد الإنشاء مباشرة.
//
// ⚠️ منفصل عن Create مو معاه: Create عنده منادين آخرين (استيراد،
// مهام مجدولة) ما إلهم موظف، وتغيير توقيعه چان يفرض عليهم يمرّرون
// معرّفاً وهمياً — يعني ننسب الحجز لواحد ما أدخله.
//
// ⚠️ الفشل ما يوقف إنشاء الحجز: «منو أدخله» معلومة زينة، بس ضياعها
// ما يسوّي حجز الزبون ينرفض.
func (s *BookingService) RecordCreator(bookingID, employeeID string) {
	if bookingID == "" || employeeID == "" {
		return
	}
	_ = s.repo.SetCreatedBy(bookingID, employeeID)
}

// SetCrewNotes ملاحظة الإداري للكادر المنفّذ — يقراها الفريق بشاشته.
func (s *BookingService) SetCrewNotes(id, note, byEmployeeID string) (*model.Booking, error) {
	if err := s.repo.SetCrewNotes(id, note, byEmployeeID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

// SetProjectNotes ملاحظة الإداري لمدير المشاريع.
func (s *BookingService) SetProjectNotes(id, note, byEmployeeID string) (*model.Booking, error) {
	if err := s.repo.SetProjectNotes(id, note, byEmployeeID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

// Cancel إلغاء الحجز — السبب إجباري.
//
// ⚠️ «ملغى» بلا سبب ما تفيد أحد: المراقب ما يعرف ليش، والإحصائية ما
// تفرّق بين زبون بدّل رأيه وبين غلط منّا خسّرنا شغلاً.
func (s *BookingService) Cancel(id, reason, byEmployeeID string) (*model.Booking, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, errors.New("سبب الإلغاء إجباري")
	}
	if err := s.ensureNotProjectLocked(id); err != nil {
		return nil, err
	}
	if err := s.repo.Cancel(id, strings.TrimSpace(reason), byEmployeeID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

// ListByStageBucket و StageBucketCounts — تمرير مباشر للمستودع.
func (s *BookingService) ListByStageBucket(bucket string, limit int) ([]model.Booking, error) {
	return s.repo.ListByStageBucket(bucket, limit)
}

func (s *BookingService) StageBucketCounts() (map[string]int, error) {
	return s.repo.StageBucketCounts()
}

// Get حجز واحد بمعرّفه — للمنادين الي يحتاجون النسخة المهدرجة
// (بالأسماء وسلّة المرحلة) بعد عملية كتابة.
func (s *BookingService) Get(id string) (*model.Booking, error) {
	return s.repo.FindByID(id)
}
