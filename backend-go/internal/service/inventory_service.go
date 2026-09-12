package service

import (
	"errors"
	"log"
	"strings"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type InventoryService struct {
	repo *repository.InventoryRepository
	// ═══ إشعار القرار ═══
	// «إداري الكوادر يحتاج الإشعارات الي توصل إله حتى يشوف طلباته
	// مرفوضة لو مقبولة».
	//
	// ⚠️ الطلب كان **يروح بلا رجعة**: الموظف يطلب أداة، وإداري
	// الكميات يوافق أو يرفض، وما يوصل الطالب ولا خبر. يضل يفتح
	// «طلباتي» كل يوم يشوف تغيّرت لو لا — أو يتصل يسأل.
	// (الإجازات والطلبات الإدارية كانت الوحيدة الي تشعر بالقرار.)
	notifications *repository.NotificationRepository
	// طلب أداة مو متوفرة بالمخزن يتحول لطلب مشتريات يوصل للمحاسب — نحتاج
	// مستودع المشتريات حتى ننشئه. اختياري (SetProcurementRepository) حتى ما
	// ينكسر أي كود ينشئ الخدمة بدونه.
	procurement *repository.ProcurementRepository
}

func NewInventoryService(repo *repository.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

func (s *InventoryService) SetProcurementRepository(p *repository.ProcurementRepository) {
	s.procurement = p
}

func (s *InventoryService) SetNotificationRepository(n *repository.NotificationRepository) {
	s.notifications = n
}

// notifyRequester يوصل قرار الطلب لصاحبه.
// ⚠️ فشل الإشعار ما يلغي القرار — القرار صار فعلاً، وإرجاعه لأن
// الإشعار فشل أسوأ من إشعار ضايع.
func (s *InventoryService) notifyRequester(employeeID, notifType, message string) {
	if s.notifications == nil || employeeID == "" {
		return
	}
	if err := s.notifications.Create(employeeID, notifType, message); err != nil {
		log.Printf("[inventory] تعذر إرسال إشعار القرار للموظف %s: %v", employeeID, err)
	}
}

func (s *InventoryService) CreateInventoryCheck(employeeID string, req model.CreateInventoryCheckRequest) (*model.InventoryCheck, error) {
	return s.repo.CreateInventoryCheck(employeeID, req)
}

func (s *InventoryService) TodaysInventoryChecks() ([]model.InventoryCheck, error) {
	return s.repo.TodaysInventoryChecks()
}

// BookingCrewInventory حالة جرد كل واحد بكادر حجز معيّن.
//
// ⚠️ محصورة بمن هو **بنفس الحجز**: بدون هالشرط أي موظف يقرا نواقص
// عدة أي زميل بتبديل رقم بالرابط. والليدر يشوف فريقه هو، مو كل
// الشركة.
func (s *InventoryService) BookingCrewInventory(bookingID, viewerID string) ([]model.BookingCrewInventoryState, error) {
	rows, err := s.repo.BookingCrewInventory(bookingID)
	if err != nil {
		return nil, err
	}
	for _, c := range rows {
		if c.EmployeeID == viewerID {
			return rows, nil
		}
	}
	return nil, errors.New("هذا الحجز مو من حجوزاتك")
}

func (s *InventoryService) LastInventoryCheck(employeeID string) (*model.InventoryCheck, error) {
	return s.repo.LastInventoryCheck(employeeID)
}

func (s *InventoryService) ResolveInventoryCheck(id string, resolvedByID string) (*model.InventoryCheck, error) {
	return s.repo.ResolveInventoryCheck(id, resolvedByID)
}

func (s *InventoryService) ListPersonalTools(employeeID string) ([]model.PersonalTool, error) {
	return s.repo.ListPersonalTools(employeeID)
}

func (s *InventoryService) CreatePersonalTool(req model.CreatePersonalToolRequest, actorID *string) (*model.PersonalTool, error) {
	// الباركود ما عاد مطلوب — انشال من الفورم ويتولّد بالمستودع لما يجي فاضي
	if req.EmployeeID == "" || strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("لازم تختار الموظف وتكتب اسم الأداة")
	}
	req.Name = strings.TrimSpace(req.Name)
	return s.repo.CreatePersonalTool(req.EmployeeID, req.Name, req.Barcode, actorID)
}

func (s *InventoryService) UpdatePersonalTool(id string, req model.UpdatePersonalToolRequest, actorID *string) (*model.PersonalTool, error) {
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return nil, errors.New("اسم الأداة ما يصير فاضي")
		}
		req.Name = &trimmed
	}
	if req.Status != nil {
		if _, ok := model.PersonalToolStatusLabels[*req.Status]; !ok {
			return nil, errors.New("حالة الأداة غير صحيحة")
		}
	}
	return s.repo.UpdatePersonalTool(id, req, actorID)
}

func (s *InventoryService) DeletePersonalTool(id string, actorID *string) error {
	return s.repo.DeletePersonalTool(id, actorID)
}

// ═══ استثناء أداة من نواقص موظف بعينه ═══
// الحذف يستثنيها تلقائياً؛ هذولا للتراجع وللعرض.
func (s *InventoryService) ListPersonalToolExemptions() ([]model.PersonalToolExemption, error) {
	return s.repo.ListPersonalToolExemptions()
}

func (s *InventoryService) ExemptPersonalTool(employeeID, toolName string, note, actorID *string) error {
	return s.repo.ExemptPersonalTool(employeeID, toolName, note, actorID)
}

func (s *InventoryService) UnexemptPersonalTool(employeeID, toolName string) error {
	return s.repo.UnexemptPersonalTool(employeeID, toolName)
}

func (s *InventoryService) ListToolEvents(toolID, employeeID string) ([]model.PersonalToolEvent, error) {
	return s.repo.ListToolEvents(toolID, employeeID)
}

func (s *InventoryService) ListVehicleTools(vehicleID string) ([]model.VehicleTool, error) {
	return s.repo.ListVehicleTools(vehicleID)
}

func (s *InventoryService) CreateVehicleTool(req model.CreateVehicleToolRequest) (*model.VehicleTool, error) {
	// الباركود صار اختياري — الكمية أخذت محله لأن نفس الأداة ممكن تتكرر
	// بنفس السيارة، والباركود الفريد كان يمنع هذي الحالة.
	if strings.TrimSpace(req.Name) == "" || req.VehicleID == "" {
		return nil, errors.New("اسم الأداة والسيارة مطلوبين")
	}
	qty := 1
	if req.Quantity != nil {
		if *req.Quantity < 1 {
			return nil, errors.New("الكمية لازم تكون 1 أو أكثر")
		}
		qty = *req.Quantity
	}
	barcode := req.Barcode
	if barcode != nil && strings.TrimSpace(*barcode) == "" {
		barcode = nil
	}
	return s.repo.CreateVehicleTool(strings.TrimSpace(req.Name), barcode, qty, req.VehicleID)
}

func (s *InventoryService) UpdateVehicleTool(id string, req model.UpdateVehicleToolRequest) (*model.VehicleTool, error) {
	if req.Quantity != nil && *req.Quantity < 1 {
		return nil, errors.New("الكمية لازم تكون 1 أو أكثر")
	}
	return s.repo.UpdateVehicleTool(id, req.Name, req.Barcode, req.Status, req.VehicleID, req.Quantity)
}

func (s *InventoryService) DeleteVehicleTool(id string) error {
	return s.repo.DeleteVehicleTool(id)
}

func (s *InventoryService) ListOnDemandTools() ([]model.OnDemandTool, error) {
	return s.repo.ListOnDemandTools()
}

func (s *InventoryService) CreateOnDemandTool(req model.CreateOnDemandToolRequest) (*model.OnDemandTool, error) {
	if req.Name == "" || req.Barcode == "" {
		return nil, errors.New("name and barcode are required")
	}
	// أداة جديدة توها ننضافت مو مسحوبة من قبل أي موظف بعد — لازم كل الكمية
	// تكون متوفرة فوراً. الواجهة ما ترسل availableQuantity وقت الإضافة أصلاً
	// (تطلب بس الكمية الإجمالية)، فلو اعتمدنا القيمة المرسلة تضل 0 دائماً
	// ويصير الموظفين ما يكدرون يطلبون الأداة إطلاقاً حتى لو أضفناها بكمية.
	return s.repo.CreateOnDemandTool(req.Name, req.Barcode, req.TotalQuantity, req.TotalQuantity)
}

func (s *InventoryService) UpdateOnDemandTool(id string, req model.UpdateOnDemandToolRequest) (*model.OnDemandTool, error) {
	return s.repo.UpdateOnDemandTool(id, req.Name, req.Barcode, req.Status, req.TotalQuantity, req.AvailableQuantity)
}

func (s *InventoryService) ListToolRequests(employeeID string) ([]model.ToolRequest, error) {
	return s.repo.ListToolRequests(employeeID)
}

func (s *InventoryService) CreateToolRequest(req model.CreateToolRequestRequest) (*model.ToolRequest, error) {
	if req.EmployeeID == "" || req.ToolID == "" {
		return nil, errors.New("employeeId and toolId are required")
	}
	if req.Reason == "" {
		return nil, errors.New("لازم تختار سبب الطلب")
	}
	if !model.IsValidToolRequestReason(req.Reason) {
		return nil, errors.New("سبب الطلب غير صحيح")
	}
	desc := req.Description
	if desc != nil && strings.TrimSpace(*desc) == "" {
		desc = nil
	}
	// "سبب آخر" بدون شرح ما ينفع الي راح يوافق — يوصله طلب بلا معلومة.
	if req.Reason == model.ToolRequestReasonOther && desc == nil {
		return nil, errors.New("لازم تكتب شرح للسبب لما تختار «سبب آخر»")
	}
	return s.repo.CreateToolRequest(req.EmployeeID, req.ToolID, req.Reason, req.RequestKind, desc)
}

func (s *InventoryService) DeleteToolRequest(id string) error {
	return s.repo.DeleteToolRequest(id)
}

func (s *InventoryService) ApproveToolRequest(id string, req model.ApproveToolRequestRequest) (*model.ToolRequest, error) {
	if req.ApprovedByID == "" {
		return nil, errors.New("approvedById is required")
	}
	existing, err := s.repo.GetToolRequest(id)
	if err != nil {
		return nil, errors.New("الطلب غير موجود")
	}
	if existing.Status != "PENDING" {
		return nil, errors.New("هذا الطلب متعامل معه من قبل")
	}

	// الأداة موجودة بالشركة؟ إذا موجودة نوافق مباشرة. إذا ما موجودة، إداري
	// الكميات مضطر يشتريها، فلازم يدخل سعرها، وننشئ طلب مشتريات يوصل للمحاسب
	// حتى تنغلق الدورة المالية بدل ما يضيع سعر الشراء بلا أثر محاسبي.
	tool, err := s.repo.GetOnDemandTool(existing.ToolID)
	if err != nil {
		return nil, errors.New("الأداة غير موجودة")
	}
	if tool.AvailableQuantity > 0 {
		// انعطت من الرف — ننقّصها من مخزن إداري الكميات
		out, err := s.repo.ApproveToolRequest(id, req.ApprovedByID, nil, nil, true)
		if err != nil {
			return nil, err
		}
		s.notifyRequester(out.EmployeeID, "tool_request_decision",
			"✅ انوافق على طلبك للأداة «"+tool.Name+"» — متوفرة بالمخزن، راجع إداري الكميات تستلمها")
		return out, nil
	}

	if req.PurchasePrice == nil || *req.PurchasePrice <= 0 {
		return nil, errors.New("الأداة مو متوفرة بالمخزن — لازم تدخل سعر الشراء حتى تنحول للمحاسب")
	}
	if s.procurement == nil {
		return nil, errors.New("تعذر إنشاء طلب المشتريات")
	}

	price := *req.PurchasePrice
	notes := "طلب شراء أداة «" + tool.Name + "» متولّد تلقائياً من موافقة على طلب أداة غير متوفرة بالمخزن."
	if existing.Reason != nil {
		notes += " السبب: " + model.ToolRequestReasonLabels[*existing.Reason] + "."
	}
	if existing.Description != nil {
		notes += " الشرح: " + *existing.Description
	}
	pr, err := s.procurement.Create(model.CreateProcurementRequestRequest{
		RequestedByID: existing.EmployeeID,
		RequestType:   model.RequestTypePersonalSupply,
		Notes:         &notes,
		Items: []model.CreateProcurementItemRequest{{
			ProductName: tool.Name,
			Quantity:    1,
			UnitPrice:   &price,
			TotalPrice:  &price,
		}},
	})
	if err != nil {
		return nil, errors.New("تعذر إنشاء طلب المشتريات للمحاسب")
	}
	// انشترت خصيصاً — ما كانت بالمخزن أصلاً فماكو شي ينتنقص
	out, err := s.repo.ApproveToolRequest(id, req.ApprovedByID, &price, &pr.ID, false)
	if err != nil {
		return nil, err
	}
	s.notifyRequester(out.EmployeeID, "tool_request_decision",
		"✅ انوافق على طلبك للأداة «"+tool.Name+"» — مو متوفرة بالمخزن، فانفتح طلب شراء إلها ("+pr.Code+")")
	return out, nil
}

func (s *InventoryService) RejectToolRequest(id string) (*model.ToolRequest, error) {
	out, err := s.repo.RejectToolRequest(id)
	if err != nil {
		return nil, err
	}
	s.notifyRequester(out.EmployeeID, "tool_request_decision", "❌ انرفض طلب الأداة مالتك — راجع إداري الكميات لو تحتاج توضيح")
	return out, nil
}

// ReturnToolRequest الأداة رجعت لإداري الكميات — ترجع للمخزن بعد.
func (s *InventoryService) ReturnToolRequest(id string) (*model.ToolRequest, error) {
	out, err := s.repo.ReturnToolRequest(id)
	if err != nil {
		return nil, err
	}
	// الإرجاع صار فعلاً؛ فشل تحديث الرقم ما يلغيه — بس ينسجل
	if err := s.repo.ReturnToolStock(out.ToolID); err != nil {
		log.Printf("إرجاع أداة %s: تعذر تحديث كمية المخزن: %v", out.ToolID, err)
	}
	return out, nil
}

// ── Personal Tool Template (العدة القياسية) ─────────────────────────────────

func (s *InventoryService) ListPersonalToolTemplateItems() ([]model.PersonalToolTemplateItem, error) {
	return s.repo.ListPersonalToolTemplateItems()
}

func (s *InventoryService) CreatePersonalToolTemplateItem(req model.CreatePersonalToolTemplateItemRequest) (*model.PersonalToolTemplateItem, error) {
	if req.Name == "" {
		return nil, errors.New("اسم الأداة مطلوب")
	}
	return s.repo.CreatePersonalToolTemplateItem(req.Name)
}

func (s *InventoryService) DeletePersonalToolTemplateItem(id string) error {
	return s.repo.DeletePersonalToolTemplateItem(id)
}

// ── Vehicle Tool Check ───────────────────────────────────────────────────────

func (s *InventoryService) ListVehicleToolChecks() ([]model.VehicleToolCheck, error) {
	return s.repo.ListVehicleToolChecks()
}

func (s *InventoryService) ListAllBookingToolChecks() ([]model.BookingToolCheck, error) {
	return s.repo.ListAllBookingToolChecks()
}

func (s *InventoryService) CreateVehicleToolCheck(vehicleID, missionID, employeeID string, req model.CreateVehicleToolCheckRequest) (*model.VehicleToolCheck, error) {
	var missing *string
	if len(req.MissingToolNames) > 0 {
		joined := strings.Join(req.MissingToolNames, "، ")
		missing = &joined
	}
	return s.repo.CreateVehicleToolCheck(vehicleID, missionID, employeeID, missing)
}

// ── إضافة الكميات للمخزون ────────────────────────────────────────────────────

func (s *InventoryService) AddStock(req model.CreateStockIntakeRequest, byID *string) (*model.StockIntake, error) {
	if req.ToolID == "" {
		return nil, errors.New("لازم تحدد الأداة")
	}
	return s.repo.AddStock(req, byID)
}

func (s *InventoryService) ListStockIntakes(toolID string) ([]model.StockIntake, error) {
	return s.repo.ListStockIntakes(toolID)
}
