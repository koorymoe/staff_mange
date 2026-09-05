package service

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

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
	// monitor: صندوق المراقب. اختياري.
	monitor MonitorFeed
	// notifications: إشعارات المراقب والمحاسب. اختيارية — نفس سبب
	// `monitor`: فشل إشعار ما يصير يمنع عملية مالية.
	notifications *repository.NotificationRepository
}

// SetNotifications يربط الإشعارات بعد البناء — نفس نمط صندوق المراقب.
func (s *LeaderInvoiceService) SetNotifications(n *repository.NotificationRepository) {
	s.notifications = n
}

// SetMonitorFeed يربط صندوق المراقب بعد بناء الخدمات.
func (s *LeaderInvoiceService) SetMonitorFeed(m MonitorFeed) { s.monitor = m }

// monitorInvoiceSummary سطر واحد يوضّح المبالغ بلا ما يفتح الفاتورة.
func monitorInvoiceSummary(inv *model.LeaderInvoice) (string, string) {
	name := ""
	if inv.CustomerName != nil {
		name = *inv.CustomerName
	}
	title := fmt.Sprintf("فاتورة %s — %s", inv.AccountingCode, name)
	summary := fmt.Sprintf("التنفيذ: %.0f • المواد: %.0f • الخصم: %.0f • الصافي: %.0f",
		inv.ExecutionCost, inv.MaterialsTotal, inv.DiscountValue, inv.NetTotal)
	if inv.IsFree {
		summary += " • شغل مجاني"
	}
	return title, summary
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

	// ── الشغل المجاني ──
	//
	// ⚠️ ما نصفّر الكلفة والمواد — نصفّر **الصافي** بس، بخصم يساوي
	// المبلغ كله. ليش؟ لأن الشغل المجاني كلّفنا فعلاً: وقت كادر ومواد
	// من المخزن. لو صفّرنا كل شي، الضمان يبين مجاني على الشركة وهو مو
	// مجاني — وما نقدر نجاوب «شكد كلّفنا الضمان هالسنة؟».
	//
	// فالفاتورة تبقى تحمل الكلفة الحقيقية، والزبون يدفع صفر.
	if req.IsFree {
		if req.FreeReasonID == nil || *req.FreeReasonID == "" {
			return nil, fmt.Errorf("لازم تحدد سبب المجانية")
		}
		reason, rerr := s.invoices.FreeReason(*req.FreeReasonID)
		if rerr != nil || reason == nil {
			return nil, fmt.Errorf("سبب المجانية مو معروف")
		}
		if reason.NeedsNote && utf8.RuneCountInString(strings.TrimSpace(derefString(req.FreeReasonNote))) < 5 {
			return nil, fmt.Errorf("سبب «%s» يحتاج توضيح مكتوب", reason.Label)
		}
		discount = float64(executionCost) + materialsTotal
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
		IsFree:           req.IsFree,
		FreeReasonID:     req.FreeReasonID,
		FreeReasonNote:   req.FreeReasonNote,
	}

	saved, err := s.invoices.Create(inv, materialLines)
	if err != nil {
		return nil, err
	}

	// ═══ الفاتورة تورّث حكم المحاسب على حجزها ═══
	//
	// ⚠️⚠️ **الترتيب الزمني بالميدان معكوس**: المحاسب يدقّق الحجز يوم
	// إنجازه، والليدر يرفع فاتورته بعدها بيوم أو أسبوع. فحكم «خطأ
	// بالسعر» چان ينكتب على الحجز **قبل** ما توجد الفاتورة — والفاتورة
	// الي تجي بعده تبدي نظيفة، فتروح لطابور التدقيق والمحاسب يعيد نفس
	// الحكم مرة ثانية. وهذا الي بلّغ بيه صاحب النظام.
	//
	// ⚠️ والتوريث **ما يدهس حكماً موجوداً** (الشرط بالـSQL): لو المحاسب
	// أشّر على الفاتورة نفسها بعدين، رأيه الأحدث يبقى.
	if saved.BookingID != nil && *saved.BookingID != "" {
		s.invoices.InheritBookingVerdict(saved.ID, *saved.BookingID)
		if fresh, err := s.invoices.GetByID(saved.ID); err == nil && fresh != nil {
			saved = fresh
		}
	}

	s.computeAndSaveCommissions(saved)
	s.recordDurationSample(saved)

	// الفاتورة توصل المراقب **قبل** ما يدققها المحاسب — يعني يشوف
	// الأرقام الأصلية قبل أي تعديل، وهذا كل الفايدة.
	if s.monitor != nil {
		title, summary := monitorInvoiceSummary(saved)
		s.monitor.InvoiceStage(model.MonitorStageInvoiceBeforeAudit, saved.ID, title, summary, "TECHNICIAN", &saved.EmployeeID, false)
	}

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
	// ⚠️ الـslice الفارغة بـGo تنكتب `null` بالـJSON مو `[]`، والواجهة
	// تسوي عليها `.reduce(...)` فتنكسر الشاشة كلها بـ«صار خطأ غير
	// متوقع». وهاي تصير بحالة عادية جداً: الليدر يضيف بند وما يختار
	// عنصر التركيب بعد. نضمن قائمة فارغة حقيقية بدل null.
	if breakdown == nil {
		breakdown = []model.ExecutionCostBreakdownLine{}
	}
	if minimums == nil {
		minimums = []model.ExecutionCostSystemMinimum{}
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

	// ═══ التدقيق قبل الاعتماد — إجباري ═══
	//
	// ⚠️⚠️ **الفحص بالخادم مو بإخفاء الزر.** الواجهة تخفي زر الاعتماد
	// عن الفواتير الي ما انتدققت، بس الإخفاء **مو منع**: نداء مباشر
	// على المسار يتخطّاه. وهاي فاتورة مالية — الي ينحرس بالواجهة وحدها
	// محروس بالنية مو بالكود.
	//
	// ⚠️ والأحكام الثلاثة **كلها تمرّ**: مطابق وغير مطابق وخطأ بالسعر.
	// النظام يطلب **قراراً**، ما يطلب **موافقة** — حجب الاعتماد عن
	// «غير مطابق» يعني نظاماً ياخذ قراراً مالياً بدل صاحبه. المحاسب
	// يشوف الحكم قدّامه ويقرر.
	current, err := s.invoices.GetByID(id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, fmt.Errorf("الفاتورة غير موجودة")
	}
	if current.AuditVerdict == nil || strings.TrimSpace(*current.AuditVerdict) == "" {
		// ⚠️ الرسالة تقول **شنو يسوي** مو «ممنوع»: المحاسب الي يقرا
		// «ما عندك صلاحية» يتصل بالإدارة، والي يقرا «دقّق أول» يعرف
		// الخطوة الجاية بلا ما يسأل أحداً.
		return nil, fmt.Errorf("دقّق الفاتورة أول — اختر مطابق أو غير مطابق أو خطأ بالسعر، وبعدها تنتقل لطابور الاعتماد")
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
	// وبعد التدقيق كمان: المراقب يقارن الأرقام قبل وبعد، فأي تعديل
	// من المحاسب ينبيّن.
	if s.monitor != nil {
		title, summary := monitorInvoiceSummary(inv)
		s.monitor.InvoiceStage(model.MonitorStageInvoiceAfterAudit, inv.ID, title,
			summary+" • رقم الفاتورة المحاسبية: "+externalNumber, "FINANCE", &approverEmployeeID, false)
	}
	return inv, nil
}

// ═══ المحاسب يرسلها للمراقب ═══
//
// «يطلعله خيار وي الاعتماد: إرساله للمراقب حتى المراقب يراجعها
// ويدققها».
//
// ⚠️ الطريق الثالث للشك: بلاه المحاسب محصور بين «اعتمدها» و«اتركها
// معلّقة». والمعلّقة بلا سبب مكتوب تبقى معلّقة شهوراً، وما أحد يعرف
// إنها تنتظر رأياً — لأن الانتظار ما انسجّل بمكان.
func (s *LeaderInvoiceService) RequestMonitorReview(id, byEmployeeID, note string) (*model.LeaderInvoice, error) {
	inv, err := s.invoices.RequestMonitorReview(id, byEmployeeID, strings.TrimSpace(note))
	if err != nil {
		return nil, err
	}
	// ⚠️ **إشعار للمراقب مو صف بصندوقه**: صندوق المراقب قراره
	// «سليم/ملاحظة» على **حدث**، وهنا القرار على **فاتورة** ويرجّعها
	// للمحاسب. سطحان للقرار على نفس الشي يعني مراقباً يبتّ بمكان
	// والفاتورة تبقى واگفة بالمكان الثاني.
	if s.notifications != nil {
		msg := "🧾 المحاسب يطلب مراجعتك لفاتورة " + inv.AccountingCode
		if n := strings.TrimSpace(note); n != "" {
			msg += " — " + n
		}
		_ = s.notifications.CreateForRole("MONITOR", "invoice_monitor", msg)
		_ = s.notifications.CreateForRole("ADMIN", "invoice_monitor", msg)
		// ⚠️⚠️ **المالك دوره `OWNER` مو `ADMIN`** بقاعدة البيانات،
		// و`CreateForRole("ADMIN")` تطابق العمود حرفياً — يعني المالك
		// **ما چان يوصله ولا إشعار** عن فاتورة راحت للمراقب. وهو
		// الوحيد الي يگدر يرجّعها للمحاسب، فما يعرف بشي يحتاج قراره.
		_ = s.notifications.CreateForRole("OWNER", "invoice_monitor", msg)
	}
	if s.monitor != nil {
		title, summary := monitorInvoiceSummary(inv)
		s.monitor.InvoiceStage(model.MonitorStageInvoiceBeforeAudit, inv.ID, title,
			summary+" • ⏳ المحاسب طلب مراجعة المراقب", "MONITOR", &byEmployeeID, false)
	}
	return inv, nil
}

// ═══ المراقب يبتّ ═══
func (s *LeaderInvoiceService) DecideMonitorReview(id, verdict, note, byEmployeeID string) (*model.LeaderInvoice, error) {
	verdict = strings.ToUpper(strings.TrimSpace(verdict))
	if verdict != "OK" && verdict != "FLAGGED" {
		return nil, fmt.Errorf("الحكم لازم يكون: سليمة أو عندي ملاحظة")
	}
	note = strings.TrimSpace(note)
	// ⚠️ الملاحظة إجبارية بالتأشير — نفس قاعدة صندوق المراقب. «عندي
	// ملاحظة» بلا نص يوگف الفاتورة بلا ما يعرف المحاسب شنو يصلّح.
	if verdict == "FLAGGED" && utf8.RuneCountInString(note) < 5 {
		return nil, fmt.Errorf("اكتب ملاحظتك — المحاسب لازم يعرف شنو يصلّح")
	}
	inv, err := s.invoices.DecideMonitorReview(id, verdict, note, byEmployeeID)
	if err != nil {
		return nil, err
	}
	if s.notifications != nil {
		lbl := "✅ سليمة"
		if verdict == "FLAGGED" {
			lbl = "⚠️ عليها ملاحظة: " + note
		}
		note := "👁️ المراقب راجع فاتورة " + inv.AccountingCode + " — " + lbl
		_ = s.notifications.CreateForRole("FINANCE", "invoice_monitor", note)
		// ⚠️ والمالك يتابع الدورة كاملة: راحت للمراقب ورجعت بحكمه.
		// إشعار بالذهاب بلا إشعار بالرجوع يخلّيه يظن إنها لسه واگفة.
		_ = s.notifications.CreateForRole("OWNER", "invoice_monitor", note)
	}
	return inv, nil
}

// ═══ المالك يرجّعها للمحاسب ═══
//
// «الفواتير الي راحن قبل هذا التعديل — صلاحية فقط للمالك أرجعهن
// للمحاسب حتى يرتبهن من جديد».
//
// ⚠️ **للمالك وحده ولا حتى الإداري**: هاي تشيل اعتماداً صار وترجّع
// فاتورة لأول الطابور. صلاحية بهالوزن بيد أكثر من واحد تعني إن
// المالك ما يعرف منو غيّر شنو.
func (s *LeaderInvoiceService) ReturnToAccountant(id, byEmployeeID, reason string) (*model.LeaderInvoice, error) {
	reason = strings.TrimSpace(reason)
	// ⚠️ السبب إجباري: الفاتورة راح ترجع للمحاسب وهو ما يعرف ليش —
	// فيعيد نفس الشغل بنفس الطريقة، وترجع له مرة ثانية.
	if utf8.RuneCountInString(reason) < 5 {
		return nil, fmt.Errorf("اكتب سبب الإرجاع — المحاسب لازم يعرف شنو يرتّب")
	}
	inv, err := s.invoices.ReturnToAccountant(id, byEmployeeID, reason)
	if err != nil {
		return nil, err
	}
	if s.notifications != nil {
		back := "↩️ المالك رجّع فاتورة " + inv.AccountingCode + " للترتيب من جديد — السبب: " + reason
		_ = s.notifications.CreateForRole("FINANCE", "invoice_returned", back)
		// ⚠️ والمراقب هم: لو چانت عنده وانسحبت منه، لازم يعرف ليش
		// اختفت من طابوره — وإلا يدوّر عليها ويظن النظام بلعها.
		_ = s.notifications.CreateForRole("MONITOR", "invoice_returned", back)
	}
	if s.monitor != nil {
		title, summary := monitorInvoiceSummary(inv)
		s.monitor.InvoiceStage(model.MonitorStageInvoiceAdjusted, inv.ID+":ret", title,
			summary+" • ↩️ المالك رجّعها للمحاسب — السبب: "+reason, "FINANCE", &byEmployeeID, false)
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
func (s *LeaderInvoiceService) AdjustAmounts(id string, req model.AdjustLeaderInvoiceRequest, byEmployeeID string) (*model.LeaderInvoice, error) {
	if strings.TrimSpace(req.Reason) == "" {
		return nil, fmt.Errorf("سبب التعديل مطلوب")
	}
	if req.ExecutionCost < 0 || req.MaterialsTotal < 0 || req.DiscountValue < 0 {
		return nil, fmt.Errorf("المبالغ ما تصير بالسالب")
	}
	inv, err := s.invoices.AdjustAmounts(id, req.ExecutionCost, req.MaterialsTotal, req.DiscountValue,
		strings.TrimSpace(req.Reason), byEmployeeID)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, fmt.Errorf("الفاتورة غير موجودة")
	}
	// تعديل مبلغ فاتورة هو اللحظة الي تحتاج عين ثانية — وقبل، ما كانت
	// تدزّ ولا شي للمراقب (بعكس الإصدار والاعتماد).
	if s.monitor != nil {
		title, summary := monitorInvoiceSummary(inv)
		// ⚠️ مفتاح الصف = معرّف **التعديل** مو الفاتورة.
		// صندوق المراقب عنده فهرس فريد (نوع+معرّف+محطة) يمنع التكرار،
		// فلو استعملنا معرّف الفاتورة چان التعديل الثاني والثالث
		// انبلعوا بصمت — والمراقب يشوف الأول بس. وكل تعديل على مبلغ
		// يستاهل عين لحاله.
		entityID := inv.ID
		if rows, err := s.invoices.Adjustments(inv.ID); err == nil && len(rows) > 0 {
			entityID = rows[0].ID
		}
		s.monitor.Stage(model.MonitorStageInvoiceAdjusted, "INVOICE_ADJUSTMENT", entityID, title,
			summary+" • سبب التعديل: "+strings.TrimSpace(req.Reason), "FINANCE", &byEmployeeID)
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

// ListByStage فواتير مرحلة معيّنة — فارغة تعني الكل.
func (s *LeaderInvoiceService) ListByStage(employeeID, stage string) ([]model.LeaderInvoice, error) {
	return s.invoices.ListByStage(employeeID, stage)
}

func (s *LeaderInvoiceService) List(employeeID string) ([]model.LeaderInvoice, error) {
	return s.invoices.List(employeeID)
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// FreeReasons أسباب الشغل المجاني الفعّالة.
func (s *LeaderInvoiceService) FreeReasons() ([]model.FreeWorkReason, error) {
	return s.invoices.FreeReasons()
}


// Adjustments سجل تعديلات فاتورة — «شنو كان وشنو صار ومنو غيّره».
func (s *LeaderInvoiceService) Adjustments(invoiceID string) ([]model.LeaderInvoiceAdjustment, error) {
	return s.invoices.Adjustments(invoiceID)
}

// ═══ التدقيق ═══

// SetAuditVerdict حكم المحاسب: مطابق / غير مطابق / خطأ بالسعر.
//
// «من ينطي هنا شي يكتب ملاحظة، والمدقق يكدر يقرأه والمدير يكدر
// يقرأه» — فالحكم يدزّ صف لصندوق المراقب.
func (s *LeaderInvoiceService) SetAuditVerdict(id string, req model.AuditVerdictRequest, byEmployeeID string) (*model.LeaderInvoice, error) {
	if !model.ValidAuditVerdict(req.Verdict) {
		return nil, fmt.Errorf("الحكم لازم يكون: مطابق أو غير مطابق أو خطأ بالسعر")
	}
	note := strings.TrimSpace(req.Note)
	// ⚠️ «غير مطابق» أو «خطأ بالسعر» بلا شرح ما تفيد أحد: المراقب
	// والمدير الاثنين يقرونها، ولازم يعرفون **شنو** الاختلاف.
	if req.Verdict != model.AuditVerdictMatched && utf8.RuneCountInString(note) < 5 {
		return nil, fmt.Errorf("اكتب ملاحظة توضّح الاختلاف — المراقب والمدير راح يقرونها")
	}
	inv, err := s.invoices.SetAuditVerdict(id, req.Verdict, note, byEmployeeID, req.Amount)
	if err != nil {
		return nil, err
	}
	// ⚠️⚠️ «غير مطابق» تعني إن بالفاتورة مشكلة حقيقية، مو تصليح رقم.
	// فترتفع للمراقب **عاجلة**: صف عاجل بصندوقه + إشعار فوري.
	// و«خطأ بالسعر» تمرّ عادية — قراره الصريح.
	//
	// ⚠️ والثلاثة الأحكام كلهن يبقن يروحن لـ«بانتظار الاعتماد» مثل
	// ما هن: التصعيد **تنبيه إضافي**، مو تحويل مسار الفاتورة.
	urgent := req.Verdict == model.AuditVerdictMismatch
	if s.monitor != nil && inv != nil {
		title, summary := monitorInvoiceSummary(inv)
		detail := " • حكم التدقيق: " + model.AuditVerdictLabel(req.Verdict)
		if urgent {
			detail = " • 🔴 عاجل" + detail
		}
		if note != "" {
			detail += " — " + note
		}
		s.monitor.InvoiceStage(model.MonitorStageInvoiceBeforeAudit, inv.ID, title,
			summary+detail, "FINANCE", &byEmployeeID, urgent)
	}
	if urgent && s.notifications != nil && inv != nil {
		msg := "🔴 المحاسب أشّر «غير مطابق» على فاتورة " + inv.AccountingCode
		if note != "" {
			msg += " — " + note
		}
		// ⚠️ الثلاثة سوا: المالك دوره `OWNER` مو `ADMIN`، و
		// `CreateForRole` تطابق العمود حرفياً — بدون سطره ما يوصله شي.
		_ = s.notifications.CreateForRole("MONITOR", "invoice_mismatch", msg)
		_ = s.notifications.CreateForRole("ADMIN", "invoice_mismatch", msg)
		_ = s.notifications.CreateForRole("OWNER", "invoice_mismatch", msg)
	}
	return inv, nil
}

// ═══ فاتورة خدمة بسعر يدوي (جي بي اس / داش كام) ═══
//
// «هذا ما يرادله تيم وليدر… والي يسوّي الفاتورة هو مسؤول الخدمة
// نفسها، واني أخلي السعر بكيفي».
//
// ⚠️ **مسار منفصل عن فاتورة الليدر بقصد**: فاتورة الليدر تلزم
// منظومات وبنود تنفيذ وسعرها ينحسب بالسيرفر من الكتالوگ — وهذا
// مقصود ويبقى. لو فكّينا الإلزام عنها حتى تخدم هالحالة، صار كل
// أحد يگدر يكتب سعراً بالإيد وانفرغ حساب الكلفة من معناه.
//
// ⚠️ **والحارس بالمسار**: `invoice_gps` أو `invoice_dashcam` حسب
// النوع — مسؤول الجي بي اس ما يفوتر داش كام.
func (s *LeaderInvoiceService) CreateServiceInvoice(employeeID string, req model.CreateServiceInvoiceRequest) (*model.LeaderInvoice, error) {
	if _, ok := model.ServiceInvoicePermission[req.Kind]; !ok {
		return nil, fmt.Errorf("نوع الفاتورة لازم يكون جي بي اس أو داش كام")
	}
	// ⚠️ السعر حر بس مو سالباً ولا صفراً: فاتورة بصفر تعني «مجاني»
	// وهذا مسار ثاني إله سببه المكتوب — خلطهما يخفي كلفة الضمان.
	if req.Price <= 0 {
		return nil, fmt.Errorf("اكتب سعر الفاتورة")
	}
	kindLabel := model.ServiceInvoiceKindLabel[req.Kind]
	inv := &model.LeaderInvoice{
		BookingID:       req.BookingID,
		EmployeeID:      employeeID,
		CustomerName:    req.CustomerName,
		CustomerPhone:   req.CustomerPhone,
		CustomerAddress: req.CustomerAddress,
		// ⚠️ المنظومة تنكتب باسم الخدمة حتى الفاتورة تبيّن **شنو هي**
		// بكل الشاشات الموجودة بلا ما نضيف عموداً ونعدّل كل شاشة.
		Systems:          []string{kindLabel},
		Items:            []model.ExecutionCostItem{},
		TotalDeviceCount: 0,
		// السعر كله «تنفيذ»: ماكو مواد ولا خصم بهذي الخدمات.
		ExecutionCost:  req.Price,
		MaterialsTotal: 0,
		DiscountValue:  0,
		NetTotal:       req.Price,
		Status:         "SUBMITTED",
	}
	saved, err := s.invoices.Create(inv, nil)
	if err != nil {
		return nil, err
	}
	s.computeAndSaveCommissions(saved)
	// نفس مسار المراقب مثل أي فاتورة — السعر اليدوي **بالذات** يستاهل
	// عينه، لأنه ماكو جدول كلفة يراجعه.
	if s.monitor != nil {
		title, summary := monitorInvoiceSummary(saved)
		detail := " • " + kindLabel + " — السعر بالإيد من مسؤول الخدمة"
		if req.Note != nil && strings.TrimSpace(*req.Note) != "" {
			detail += " — " + strings.TrimSpace(*req.Note)
		}
		s.monitor.InvoiceStage(model.MonitorStageInvoiceBeforeAudit, saved.ID, title,
			summary+detail, "SERVICE_MANAGER", &saved.EmployeeID, false)
	}
	return saved, nil
}

// RevokeApproval يسحب اعتماد فاتورة انعتمدت بالغلط.
func (s *LeaderInvoiceService) RevokeApproval(id, reason, byEmployeeID string) (*model.LeaderInvoice, error) {
	reason = strings.TrimSpace(reason)
	// ⚠️ السبب إجباري: سحب اعتماد بلا سبب يخلي السجل يقول «انسحب» وبس،
	// والمراقب ما يعرف إذا كان غلط بالرقم لو بالمبلغ لو تلاعب.
	if utf8.RuneCountInString(reason) < 5 {
		return nil, fmt.Errorf("اكتب سبب سحب الاعتماد")
	}
	inv, err := s.invoices.RevokeApproval(id, reason, byEmployeeID)
	if err != nil {
		return nil, err
	}
	// السحب لحظة تستاهل عين المراقب — نفس وزن الاعتماد.
	if s.monitor != nil && inv != nil {
		title, summary := monitorInvoiceSummary(inv)
		s.monitor.InvoiceStage(model.MonitorStageInvoiceAfterAudit, inv.ID, title,
			summary+" • ⚠️ انسحب الاعتماد — السبب: "+reason, "FINANCE", &byEmployeeID, false)
	}
	return inv, nil
}

// ListApprovedWithoutNumber الفواتير المعتمدة بلا رقم — الفجوة الي كلّفت.
func (s *LeaderInvoiceService) ListApprovedWithoutNumber() ([]model.LeaderInvoice, error) {
	return s.invoices.ListApprovedWithoutNumber()
}
