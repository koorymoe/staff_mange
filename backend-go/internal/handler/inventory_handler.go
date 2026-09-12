package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/service"
)

type InventoryHandler struct {
	service *service.InventoryService
	// permissions: منو يشوف عدة غيره (أبو الكميات ومتابع الجرد).
	permissions *repository.PermissionRepository
}

func NewInventoryHandler(s *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{service: s}
}

// SetPermissions يربط مستودع الصلاحيات بعد البناء.
func (h *InventoryHandler) SetPermissions(p *repository.PermissionRepository) { h.permissions = p }

// canSeeOthersTools منو يشوف عدة موظف غيره.
//
// ⚠️ الأدوار وحدها ما تكفي: **أبو الكميات** (`PROCUREMENT_ADMIN`) مو
// ضمن `canSeeOperational`، وهو صاحب شاشة الجرد الأصلية — فالاعتماد
// على الدور بس يكسر شغله.
func (h *InventoryHandler) canSeeOthersTools(r *http.Request) bool {
	if canSeeOperational(middleware.RoleFromContext(r)) ||
		middleware.RoleFromContext(r) == "PROCUREMENT_ADMIN" {
		return true
	}
	if h.permissions == nil {
		return false
	}
	rows, err := h.permissions.ListForEmployee(middleware.EmployeeIDFromContext(r))
	if err != nil {
		return false
	}
	for _, p := range rows {
		if p.Name == "inventory" || p.Name == "inventory_follow" {
			return true
		}
	}
	return false
}

// ── Inventory Checks ──────────────────────────────────────────────────────────

func (h *InventoryHandler) CreateInventoryCheck(w http.ResponseWriter, r *http.Request) {
	var req model.CreateInventoryCheckRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	check, err := h.service.CreateInventoryCheck(middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, check)
}

// BookingCrewInventory حالة جرد كادر حجز واحد — شاشة «جرد أدوات فريقي».
func (h *InventoryHandler) BookingCrewInventory(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.BookingCrewInventory(r.PathValue("id"), middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusForbidden, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

func (h *InventoryHandler) TodaysInventoryChecks(w http.ResponseWriter, r *http.Request) {
	checks, err := h.service.TodaysInventoryChecks()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب نتائج الجرد")
		return
	}
	WriteJSON(w, http.StatusOK, checks)
}

// GET /api/inventory/checks/mine — آخر جرد للفني نفسه (بلا أي بيانات
// عن بقية الفنيين). الفني يسوّي جرده ويشوف حالته هو؛ متابعة جرد
// الآخرين شغل الليدر والمراقب.
func (h *InventoryHandler) MyLastInventoryCheck(w http.ResponseWriter, r *http.Request) {
	check, err := h.service.LastInventoryCheck(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب آخر جرد")
		return
	}
	WriteJSON(w, http.StatusOK, check)
}

func (h *InventoryHandler) ResolveInventoryCheck(w http.ResponseWriter, r *http.Request) {
	check, err := h.service.ResolveInventoryCheck(r.PathValue("id"), middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, check)
}

// ── Personal Tools ──────────────────────────────────────────────────────────

func (h *InventoryHandler) ListPersonalTools(w http.ResponseWriter, r *http.Request) {
	// ⚠️ چان أي موظف يبدّل الرقم بالرابط ويقرا عدة أي زميل — أو
	// يتركه فاضي فيجيب عدة الشركة كلها. هسه: صاحب الجرد يشوف الكل،
	// وغيره **عدته هو** مهما كتب.
	employeeID := r.URL.Query().Get("employeeId")
	self := middleware.EmployeeIDFromContext(r)
	if employeeID != self && !h.canSeeOthersTools(r) {
		employeeID = self
	}
	tools, err := h.service.ListPersonalTools(employeeID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الأدوات الشخصية")
		return
	}
	WriteJSON(w, http.StatusOK, tools)
}

func (h *InventoryHandler) CreatePersonalTool(w http.ResponseWriter, r *http.Request) {
	var req model.CreatePersonalToolRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	tool, err := h.service.CreatePersonalTool(req, actorID(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, tool)
}

func (h *InventoryHandler) UpdatePersonalTool(w http.ResponseWriter, r *http.Request) {
	var req model.UpdatePersonalToolRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	tool, err := h.service.UpdatePersonalTool(r.PathValue("id"), req, actorID(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, tool)
}

// ═══ استثناءات العدة القياسية لموظف بعينه ═══

// GET /api/inventory/tool-exemptions
func (h *InventoryHandler) ListToolExemptions(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.ListPersonalToolExemptions()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الاستثناءات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/inventory/tool-exemptions — يشيل أداة من نواقص موظف
func (h *InventoryHandler) CreateToolExemption(w http.ResponseWriter, r *http.Request) {
	var req model.CreatePersonalToolExemptionRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات غير صحيحة")
		return
	}
	if err := h.service.ExemptPersonalTool(req.EmployeeID, req.ToolName, req.Note, actorID(r)); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]bool{"ok": true})
}

// DELETE /api/inventory/tool-exemptions — يرجّع الأداة لنواقصه
func (h *InventoryHandler) DeleteToolExemption(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employeeId")
	toolName := r.URL.Query().Get("toolName")
	if err := h.service.UnexemptPersonalTool(employeeID, toolName); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر إرجاع الأداة")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *InventoryHandler) DeletePersonalTool(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeletePersonalTool(r.PathValue("id"), actorID(r)); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// actorID منو سوّى الحركة — يُسجّل بسجل حركة الأداة. فاضي = مجهول.
func actorID(r *http.Request) *string {
	if id := middleware.EmployeeIDFromContext(r); id != "" {
		return &id
	}
	return nil
}

// ToolEvents سجل حركة الأدوات: لأداة وحدة (?toolId=) أو لموظف (?employeeId=)
// أو الكل. هنا يبين متى انفقدت كل أداة ومنو سجّل الفقدان.
func (h *InventoryHandler) ToolEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.service.ListToolEvents(r.URL.Query().Get("toolId"), r.URL.Query().Get("employeeId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب سجل حركة الأدوات")
		return
	}
	WriteJSON(w, http.StatusOK, events)
}

// ── Vehicle Tools ───────────────────────────────────────────────────────────

func (h *InventoryHandler) ListVehicleTools(w http.ResponseWriter, r *http.Request) {
	tools, err := h.service.ListVehicleTools(r.URL.Query().Get("vehicleId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب أدوات المركبة")
		return
	}
	WriteJSON(w, http.StatusOK, tools)
}

func (h *InventoryHandler) CreateVehicleTool(w http.ResponseWriter, r *http.Request) {
	var req model.CreateVehicleToolRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	tool, err := h.service.CreateVehicleTool(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, tool)
}

func (h *InventoryHandler) UpdateVehicleTool(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateVehicleToolRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	tool, err := h.service.UpdateVehicleTool(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, tool)
}

func (h *InventoryHandler) DeleteVehicleTool(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteVehicleTool(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// ── On-demand Tools ─────────────────────────────────────────────────────────

func (h *InventoryHandler) ListOnDemandTools(w http.ResponseWriter, r *http.Request) {
	tools, err := h.service.ListOnDemandTools()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الأدوات المشتركة")
		return
	}
	WriteJSON(w, http.StatusOK, tools)
}

func (h *InventoryHandler) CreateOnDemandTool(w http.ResponseWriter, r *http.Request) {
	var req model.CreateOnDemandToolRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	tool, err := h.service.CreateOnDemandTool(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, tool)
}

func (h *InventoryHandler) UpdateOnDemandTool(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateOnDemandToolRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	tool, err := h.service.UpdateOnDemandTool(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, tool)
}

// ── Tool Requests ───────────────────────────────────────────────────────────

// GET /api/inventory/requests
//
// الموظف العادي يشوف طلباته هو بس. كانت ترجع طلبات كل الكادر لأي موظف
// مسجّل دخول — يعني الفني يشوف منو طلب شنو وليش من F12.
func (h *InventoryHandler) ListToolRequests(w http.ResponseWriter, r *http.Request) {
	scope := r.URL.Query().Get("employeeId")
	switch middleware.RoleFromContext(r) {
	case "OWNER", "ADMIN", "PROCUREMENT_ADMIN", "HR_COORDINATOR":
		// هذول شغلهم يشوفون الطلبات كلها ويوافقون عليها
	default:
		scope = middleware.EmployeeIDFromContext(r)
	}
	requests, err := h.service.ListToolRequests(scope)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب طلبات الأدوات")
		return
	}
	WriteJSON(w, http.StatusOK, requests)
}

func (h *InventoryHandler) CreateToolRequest(w http.ResponseWriter, r *http.Request) {
	var req model.CreateToolRequestRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	request, err := h.service.CreateToolRequest(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, request)
}

func (h *InventoryHandler) DeleteToolRequest(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteToolRequest(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر حذف الطلب")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *InventoryHandler) ApproveToolRequest(w http.ResponseWriter, r *http.Request) {
	var req model.ApproveToolRequestRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	request, err := h.service.ApproveToolRequest(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, request)
}

func (h *InventoryHandler) RejectToolRequest(w http.ResponseWriter, r *http.Request) {
	request, err := h.service.RejectToolRequest(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, request)
}

func (h *InventoryHandler) ReturnToolRequest(w http.ResponseWriter, r *http.Request) {
	request, err := h.service.ReturnToolRequest(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, request)
}

// ── Personal Tool Template (العدة القياسية) ─────────────────────────────────

func (h *InventoryHandler) ListPersonalToolTemplateItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListPersonalToolTemplateItems()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب العدة القياسية")
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (h *InventoryHandler) CreatePersonalToolTemplateItem(w http.ResponseWriter, r *http.Request) {
	var req model.CreatePersonalToolTemplateItemRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	item, err := h.service.CreatePersonalToolTemplateItem(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, item)
}

func (h *InventoryHandler) DeletePersonalToolTemplateItem(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeletePersonalToolTemplateItem(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// ── Vehicle Tool Checks ──────────────────────────────────────────────────────

func (h *InventoryHandler) ListAllBookingToolChecks(w http.ResponseWriter, r *http.Request) {
	checks, err := h.service.ListAllBookingToolChecks()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب فحوصات أدوات الحجوزات")
		return
	}
	WriteJSON(w, http.StatusOK, checks)
}

func (h *InventoryHandler) ListVehicleToolChecks(w http.ResponseWriter, r *http.Request) {
	checks, err := h.service.ListVehicleToolChecks()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب فحوصات أدوات المركبات")
		return
	}
	WriteJSON(w, http.StatusOK, checks)
}

// ── إضافة الكميات للمخزون ────────────────────────────────────────────────────

// POST /api/inventory/stock-intake — إداري الكميات يضيف كمية لأداة
func (h *InventoryHandler) AddStock(w http.ResponseWriter, r *http.Request) {
	var req model.CreateStockIntakeRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	var by *string
	if id := middleware.EmployeeIDFromContext(r); id != "" {
		by = &id
	}
	in, err := h.service.AddStock(req, by)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, in)
}

// GET /api/inventory/stock-intake?toolId= — سجل إضافات الكميات
func (h *InventoryHandler) ListStockIntakes(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.ListStockIntakes(r.URL.Query().Get("toolId"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب سجل الإضافات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}
