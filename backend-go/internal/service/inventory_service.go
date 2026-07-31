package service

import (
	"errors"
	"strings"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type InventoryService struct {
	repo *repository.InventoryRepository
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

func (s *InventoryService) CreateInventoryCheck(employeeID string, req model.CreateInventoryCheckRequest) (*model.InventoryCheck, error) {
	return s.repo.CreateInventoryCheck(employeeID, req)
}

func (s *InventoryService) TodaysInventoryChecks() ([]model.InventoryCheck, error) {
	return s.repo.TodaysInventoryChecks()
}

func (s *InventoryService) ResolveInventoryCheck(id string, resolvedByID string) (*model.InventoryCheck, error) {
	return s.repo.ResolveInventoryCheck(id, resolvedByID)
}

func (s *InventoryService) ListPersonalTools(employeeID string) ([]model.PersonalTool, error) {
	return s.repo.ListPersonalTools(employeeID)
}

func (s *InventoryService) CreatePersonalTool(req model.CreatePersonalToolRequest) (*model.PersonalTool, error) {
	if req.EmployeeID == "" || req.Name == "" || req.Barcode == "" {
		return nil, errors.New("employeeId, name and barcode are required")
	}
	return s.repo.CreatePersonalTool(req.EmployeeID, req.Name, req.Barcode)
}

func (s *InventoryService) UpdatePersonalTool(id string, req model.UpdatePersonalToolRequest) (*model.PersonalTool, error) {
	return s.repo.UpdatePersonalTool(id, req.Status, req.CheckedOut)
}

func (s *InventoryService) DeletePersonalTool(id string) error {
	return s.repo.DeletePersonalTool(id)
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
	return s.repo.CreateToolRequest(req.EmployeeID, req.ToolID, req.Reason, desc)
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
		return s.repo.ApproveToolRequest(id, req.ApprovedByID, nil, nil)
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
	return s.repo.ApproveToolRequest(id, req.ApprovedByID, &price, &pr.ID)
}

func (s *InventoryService) RejectToolRequest(id string) (*model.ToolRequest, error) {
	return s.repo.RejectToolRequest(id)
}

func (s *InventoryService) ReturnToolRequest(id string) (*model.ToolRequest, error) {
	return s.repo.ReturnToolRequest(id)
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
