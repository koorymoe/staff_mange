package service

import (
	"fmt"
	"strings"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// LeaderInvoiceService ينسّق حساب فاتورة الليدر كاملة: تكاليف التنفيذ (بالكتالوج)
// + بنود المواد (بالأرشيف أو يدوي) + الخصم + كود المحاسبة، ثم يحفظها. كما يحسب
// ويحفظ عمولات الليدر والفنيين المشاركين تلقائياً بنفس عملية الإنشاء.
type LeaderInvoiceService struct {
	invoices    *repository.LeaderInvoiceRepository
	catalog     *repository.SystemPriceCatalogRepository
	materials   *repository.MaterialRepository
	commissions *repository.EmployeeCommissionRepository
	bookings    *repository.BookingRepository
	employees   *repository.EmployeeRepository
	durations   *JobDurationEstimatorService
}

func NewLeaderInvoiceService(
	invoices *repository.LeaderInvoiceRepository,
	catalog *repository.SystemPriceCatalogRepository,
	materials *repository.MaterialRepository,
	commissions *repository.EmployeeCommissionRepository,
	bookings *repository.BookingRepository,
	employees *repository.EmployeeRepository,
	durations *JobDurationEstimatorService,
) *LeaderInvoiceService {
	return &LeaderInvoiceService{
		invoices:    invoices,
		catalog:     catalog,
		materials:   materials,
		commissions: commissions,
		bookings:    bookings,
		employees:   employees,
		durations:   durations,
	}
}

// Create يبني ويحفظ فاتورة ليدر جديدة من طلب الإنشاء، بحساب تكاليف التنفيذ
// ومجموع المواد صرفياً بالسيرفر (مو من الواجهة) — نفس منطق شيت "تكاليف المشروع".
func (s *LeaderInvoiceService) Create(employeeID string, req model.CreateLeaderInvoiceRequest) (*model.LeaderInvoice, error) {
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("يجب اختيار بند تنفيذ واحد على الأقل")
	}
	if len(req.Systems) == 0 {
		return nil, fmt.Errorf("يجب اختيار منظومة واحدة على الأقل")
	}
	if len(req.Systems) > 3 {
		return nil, fmt.Errorf("لا يمكن اختيار أكثر من 3 منظومات بالفاتورة الواحدة")
	}

	catalog, err := s.catalog.All()
	if err != nil {
		return nil, err
	}

	totalDeviceCount := 0
	for _, item := range req.Items {
		totalDeviceCount += item.Count
	}

	executionCost, err := CalculateExecutionCost(req.Items, catalog, totalDeviceCount)
	if err != nil {
		return nil, err
	}

	var materialLines []model.LeaderInvoiceMaterialItem
	var materialsTotal float64
	for _, line := range req.Materials {
		if line.Quantity <= 0 {
			return nil, fmt.Errorf("كمية المادة يجب أن تكون أكبر من صفر")
		}
		var materialID *string
		name := ""
		unitPrice := 0.0
		profit := 0.0
		if line.MaterialCode != nil && *line.MaterialCode != "" {
			mat, err := s.materials.GetByCode(*line.MaterialCode)
			if err != nil {
				return nil, err
			}
			if mat == nil {
				return nil, fmt.Errorf("لا توجد مادة بالكود %s", *line.MaterialCode)
			}
			materialID = &mat.ID
			name = mat.Name
			unitPrice = mat.SellPrice
			profit = mat.ProfitPerUnit
		} else if line.Name != nil && *line.Name != "" {
			name = *line.Name
			if line.UnitPrice != nil {
				unitPrice = *line.UnitPrice
			}
			if line.ProfitPerUnit != nil {
				profit = *line.ProfitPerUnit
			}
		} else {
			return nil, fmt.Errorf("كل بند مادة يحتاج كود أو اسم")
		}

		lineTotal := unitPrice * line.Quantity
		materialsTotal += lineTotal
		materialLines = append(materialLines, model.LeaderInvoiceMaterialItem{
			MaterialID:    materialID,
			Name:          name,
			Quantity:      line.Quantity,
			UnitPrice:     unitPrice,
			ProfitPerUnit: profit,
			LineTotal:     lineTotal,
		})
	}

	discount := req.DiscountValue
	if discount < 0 {
		discount = 0
	}
	netTotal := float64(executionCost) + materialsTotal - discount
	if netTotal < 0 {
		netTotal = 0
	}

	inv := &model.LeaderInvoice{
		BookingID:        req.BookingID,
		EmployeeID:       employeeID,
		CustomerName:     req.CustomerName,
		CustomerPhone:    req.CustomerPhone,
		CustomerAddress:  req.CustomerAddress,
		Systems:          req.Systems,
		Items:            req.Items,
		TotalDeviceCount: totalDeviceCount,
		ExecutionCost:    float64(executionCost),
		MaterialsTotal:   materialsTotal,
		DiscountValue:    discount,
		NetTotal:         netTotal,
		Status:           "SUBMITTED",
	}

	saved, err := s.invoices.Create(inv, materialLines)
	if err != nil {
		return nil, err
	}

	s.computeAndSaveCommissions(saved)
	s.recordDurationSample(saved)

	return saved, nil
}

// recordDurationSample يسجّل عيّنة زمن تنفيذ حقيقية لهذا الحجز (لو مربوط بحجز فعلاً)
// حتى يتعلّم النظام وتيرة العمل تلقائياً — itemCount = مجموع أعداد بنود الفاتورة،
// crewSize = عدد الموظفين الفعليين المسندين (ليدر + فنيين)، durationMinutes = من
// startedAt (أو arrivedAt احتياطياً) إلى لحظة إنشاء الفاتورة (نعتبرها لحظة الإنجاز
// الفعلية لعدم توفر completedAt منفصل هنا). لو ما توفر لا startedAt ولا arrivedAt
// نتخطى التسجيل صراحة (لا نخترع مدة) مع سطر لوق واضح بالسبب.
func (s *LeaderInvoiceService) recordDurationSample(saved *model.LeaderInvoice) {
	if s.durations == nil || s.bookings == nil {
		return
	}
	if saved.BookingID == nil || *saved.BookingID == "" {
		return
	}
	booking, err := s.bookings.FindByID(*saved.BookingID)
	if err != nil || booking == nil {
		return
	}

	startTime := booking.StartedAt
	if startTime == nil {
		startTime = booking.ArrivedAt
	}
	if startTime == nil {
		fmt.Printf("job_duration_estimator: تخطي تسجيل عيّنة للحجز %s — لا يوجد startedAt ولا arrivedAt\n", booking.ID)
		return
	}

	crewSize := 1 // الليدر نفسه
	seen := map[string]bool{saved.EmployeeID: true}
	for _, a := range booking.Assignments {
		if a.Role != "TECH_1" && a.Role != "TECH_2" && a.Role != "TECH_3" {
			continue
		}
		if seen[a.EmployeeID] {
			continue
		}
		seen[a.EmployeeID] = true
		crewSize++
	}

	itemCount := 0
	for _, item := range saved.Items {
		if item.Count > 0 {
			itemCount += item.Count
		}
	}
	if itemCount <= 0 {
		itemCount = saved.TotalDeviceCount
	}

	durationMinutes := int(saved.CreatedAt.Sub(*startTime).Minutes())
	if durationMinutes <= 0 {
		durationMinutes = int(time.Since(*startTime).Minutes())
	}

	systemName := ""
	if len(saved.Systems) > 0 {
		systemName = saved.Systems[0]
	}
	if systemName == "" {
		return
	}

	bookingID := booking.ID
	_ = s.durations.RecordSample(model.JobDurationSample{
		SystemName:      systemName,
		JobType:         model.JobTypeInstall,
		ItemCount:       itemCount,
		CrewSize:        crewSize,
		DurationMinutes: durationMinutes,
		BookingID:       &bookingID,
	})

	employeeName := saved.EmployeeID
	if s.employees != nil {
		if emp, eerr := s.employees.FindByID(saved.EmployeeID); eerr == nil && emp != nil {
			employeeName = emp.Name
		}
	}
	_ = s.durations.CheckOverrunAndNotify(systemName, model.JobTypeInstall, itemCount, crewSize, durationMinutes, employeeName, "تركيب "+systemName, &bookingID)
}

// computeAndSaveCommissions يحسب ويحفظ عمولة الليدر (تنفيذ + مبيعات) وعمولة كل
// فني آخر مربوط بنفس الحجز (TECH_1/TECH_2/TECH_3 غير الليدر نفسه) — تلقائياً
// بدون أي تشغيل يدوي منفصل. أخطاء الحفظ هنا لا توقف إنشاء الفاتورة نفسها (تم
// حفظها بنجاح فعلاً) لكنها تُسجَّل ضمنياً عبر تجاهل الخطأ بحرص — العمولة عملية
// تابعة، لا يصح أن تفشل بها الفاتورة الأساسية.
func (s *LeaderInvoiceService) computeAndSaveCommissions(saved *model.LeaderInvoice) {
	if s.commissions == nil {
		return
	}

	executionCommission, salesCommission, _ := model.CalculateLeaderCommission(saved.ExecutionCost, saved.Materials)
	_, _ = s.commissions.Create(saved.EmployeeID, saved.ID, model.CommissionRoleLeader, executionCommission, salesCommission)

	if saved.BookingID == nil || *saved.BookingID == "" || s.bookings == nil {
		return
	}

	assignments, err := s.bookings.ListAssignments(*saved.BookingID)
	if err != nil {
		return
	}
	technicianCommission := model.CalculateTechnicianCommission(saved.ExecutionCost)
	seen := map[string]bool{saved.EmployeeID: true}
	for _, a := range assignments {
		if a.Role != "TECH_1" && a.Role != "TECH_2" && a.Role != "TECH_3" {
			continue
		}
		if seen[a.EmployeeID] {
			continue // الليدر نفسه مو فني إضافي، وما نكرر نفس الموظف مرتين
		}
		seen[a.EmployeeID] = true
		_, _ = s.commissions.Create(a.EmployeeID, saved.ID, model.CommissionRoleTechnician, technicianCommission, 0)
	}
}

// Estimate يحسب تكلفة تنفيذ تقريبية بدون حفظ أي شي — يستخدمه الليدر لما زبون
// يستفسر عن سعر ("حساب كلفة")، بنفس محرك CalculateExecutionCost بالضبط.
func (s *LeaderInvoiceService) Estimate(items []model.ExecutionCostItem) (*model.EstimateExecutionCostResponse, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("أضف بند تنفيذ واحد على الأقل")
	}
	catalog, err := s.catalog.All()
	if err != nil {
		return nil, err
	}
	totalDeviceCount := 0
	for _, item := range items {
		totalDeviceCount += item.Count
	}
	executionCost, breakdown, minimums, err := CalculateExecutionCostDetailed(items, catalog, totalDeviceCount)
	if err != nil {
		return nil, err
	}
	return &model.EstimateExecutionCostResponse{
		ExecutionCost:    executionCost,
		TotalDeviceCount: totalDeviceCount,
		Breakdown:        breakdown,
		SystemMinimums:   minimums,
	}, nil
}

// Approve يعتمد فاتورة ليدر — محصور بمدير/محاسب (requireFinance بالراوت)، الليدر
// نفسه ما يقدر يعتمد فاتورته حتى لو كملها، لازم طرف ثاني يراجعها.
func (s *LeaderInvoiceService) Approve(id, approverEmployeeID, externalNumber string) (*model.LeaderInvoice, error) {
	// رقم الفاتورة المحاسبية إجباري: المحاسب يصدّر فواتيره بنظام ثاني،
	// وبدون الرقم ينقطع الخيط بين النظامين وما يبقى شي يربط فاتورتنا
	// المعتمدة بفاتورته الصادرة.
	externalNumber = strings.TrimSpace(externalNumber)
	if externalNumber == "" {
		return nil, fmt.Errorf("رقم الفاتورة المحاسبية مطلوب قبل الاعتماد")
	}
	inv, err := s.invoices.Approve(id, approverEmployeeID, externalNumber)
	if err != nil {
		// الرقم فريد — لو انستعمل بفاتورة ثانية نوضّح السبب بدل رسالة
		// قاعدة بيانات ما يفهمها أحد
		if strings.Contains(err.Error(), "leader_invoice_external_number_unique") {
			return nil, fmt.Errorf("رقم الفاتورة %s مستعمل بفاتورة ثانية — تأكد من الرقم", externalNumber)
		}
		return nil, err
	}
	if inv == nil {
		return nil, fmt.Errorf("الفاتورة غير موجودة أو معتمدة أصلاً")
	}
	return inv, nil
}

// SetExternalNumber يربط رقم فاتورة محاسبية بفاتورة معتمدة أصلاً —
// للفواتير الي انعتمدت قبل ما يصير الرقم إجبارياً.
func (s *LeaderInvoiceService) SetExternalNumber(id, number string) (*model.LeaderInvoice, error) {
	number = strings.TrimSpace(number)
	if number == "" {
		return nil, fmt.Errorf("اكتب رقم الفاتورة المحاسبية")
	}
	inv, err := s.invoices.SetExternalNumber(id, number)
	if err != nil {
		if strings.Contains(err.Error(), "leader_invoice_external_number_unique") {
			return nil, fmt.Errorf("رقم الفاتورة %s مستعمل بفاتورة ثانية — تأكد من الرقم", number)
		}
		return nil, err
	}
	if inv == nil {
		return nil, fmt.Errorf("الفاتورة غير موجودة")
	}
	return inv, nil
}

// AdjustAmounts يعدّل مبالغ الفاتورة — للمحاسب. السبب إجباري: التعديل
// على مبلغ ما يصير يمر بلا تفسير مكتوب.
func (s *LeaderInvoiceService) AdjustAmounts(id string, req model.AdjustLeaderInvoiceRequest) (*model.LeaderInvoice, error) {
	if strings.TrimSpace(req.Reason) == "" {
		return nil, fmt.Errorf("سبب التعديل مطلوب")
	}
	if req.ExecutionCost < 0 || req.MaterialsTotal < 0 || req.DiscountValue < 0 {
		return nil, fmt.Errorf("المبالغ ما تصير بالسالب")
	}
	inv, err := s.invoices.AdjustAmounts(id, req.ExecutionCost, req.MaterialsTotal, req.DiscountValue, strings.TrimSpace(req.Reason))
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, fmt.Errorf("الفاتورة غير موجودة")
	}
	return inv, nil
}

// FindByExternalNumber يدوّر فاتورة برقم المحاسب.
func (s *LeaderInvoiceService) FindByExternalNumber(number string) (*model.LeaderInvoice, error) {
	number = strings.TrimSpace(number)
	if number == "" {
		return nil, fmt.Errorf("اكتب رقم الفاتورة")
	}
	return s.invoices.FindByExternalNumber(number)
}

func (s *LeaderInvoiceService) approveOld(id, approverEmployeeID string) (*model.LeaderInvoice, error) {
	inv, err := s.invoices.Approve(id, approverEmployeeID, "")
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, fmt.Errorf("الفاتورة غير موجودة أو معتمدة أصلاً")
	}
	return inv, nil
}

func (s *LeaderInvoiceService) Get(id string) (*model.LeaderInvoice, error) {
	inv, err := s.invoices.GetByID(id)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, fmt.Errorf("الفاتورة غير موجودة")
	}
	return inv, nil
}

func (s *LeaderInvoiceService) List(employeeID string) ([]model.LeaderInvoice, error) {
	return s.invoices.List(employeeID)
}
