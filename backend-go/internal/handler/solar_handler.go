package handler

import (
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// SolarHandler نظام الطاقة الشمسية المنقول من Google Sheets.
type SolarHandler struct {
	repo *repository.SolarRepository
}

func NewSolarHandler(repo *repository.SolarRepository) *SolarHandler {
	return &SolarHandler{repo: repo}
}

// GET /api/solar/stats
func (h *SolarHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.Stats()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب أرقام الطاقة الشمسية")
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}

// GET /api/solar/low-stock
func (h *SolarHandler) LowStock(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.LowStock()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب تنبيهات المخزن")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// ═══ المكوّنات ═══

// GET /api/solar/components?category=PANEL
func (h *SolarHandler) ListComponents(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListComponents(r.URL.Query().Get("category"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب مواد المخزن")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/solar/components
func (h *SolarHandler) CreateComponent(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeComponent(w, r)
	if !ok {
		return
	}
	c, err := h.repo.CreateComponent(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, componentError(err))
		return
	}
	WriteJSON(w, http.StatusCreated, c)
}

// PUT /api/solar/components/{id}
func (h *SolarHandler) UpdateComponent(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeComponent(w, r)
	if !ok {
		return
	}
	c, err := h.repo.UpdateComponent(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, componentError(err))
		return
	}
	WriteJSON(w, http.StatusOK, c)
}

// DELETE /api/solar/components/{id}
func (h *SolarHandler) DeleteComponent(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.DeleteComponent(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func decodeComponent(w http.ResponseWriter, r *http.Request) (model.SaveSolarComponentRequest, bool) {
	var req model.SaveSolarComponentRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات المادة غير صحيحة")
		return req, false
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, "اسم المادة مطلوب")
		return req, false
	}
	switch req.Category {
	case model.SolarPanel, model.SolarInverter, model.SolarBattery, model.SolarBoard, model.SolarIron:
	default:
		WriteError(w, http.StatusBadRequest, "تصنيف المادة مو صحيح")
		return req, false
	}
	// الكمية والسعر ما يصيرون بالسالب — لا بالإدخال ولا بالتعديل
	if req.Quantity < 0 || req.Price < 0 || req.MinStock < 0 {
		WriteError(w, http.StatusBadRequest, "الكمية والسعر والحد الأدنى ما يصيرون بالسالب")
		return req, false
	}
	return req, true
}

// componentError يترجم تصادم الفهرس الفريد لرسالة يفهمها المستخدم بدل
// نص بوستكرس الإنكليزي.
func componentError(err error) string {
	if strings.Contains(err.Error(), "SolarComponent_name_category_key") {
		return "هذي المادة موجودة بنفس التصنيف"
	}
	return err.Error()
}

// ═══ المنظومات ═══

// GET /api/solar/systems?brand=Deye
func (h *SolarHandler) ListSystems(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListSystems(r.URL.Query().Get("brand"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المنظومات")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/solar/systems/{id}
func (h *SolarHandler) GetSystem(w http.ResponseWriter, r *http.Request) {
	s, err := h.repo.FindSystem(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusNotFound, "المنظومة مو موجودة")
		return
	}
	WriteJSON(w, http.StatusOK, s)
}

// POST /api/solar/systems
func (h *SolarHandler) CreateSystem(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeSystem(w, r)
	if !ok {
		return
	}
	s, err := h.repo.SaveSystem("", req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, s)
}

// PUT /api/solar/systems/{id}
func (h *SolarHandler) UpdateSystem(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeSystem(w, r)
	if !ok {
		return
	}
	s, err := h.repo.SaveSystem(r.PathValue("id"), req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, s)
}

// DELETE /api/solar/systems/{id}
func (h *SolarHandler) DeleteSystem(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.DeleteSystem(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func decodeSystem(w http.ResponseWriter, r *http.Request) (model.SaveSolarSystemRequest, bool) {
	var req model.SaveSolarSystemRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات المنظومة غير صحيحة")
		return req, false
	}
	req.Brand = strings.TrimSpace(req.Brand)
	req.Model = strings.TrimSpace(req.Model)
	req.Capacity = strings.TrimSpace(req.Capacity)
	if req.Brand == "" || req.Model == "" || req.Capacity == "" {
		WriteError(w, http.StatusBadRequest, "الماركة والموديل والسعة مطلوبات")
		return req, false
	}
	return req, true
}

// ═══ التجهيز والمتابعة ═══

// POST /api/solar/systems/{id}/process
//
// تجهيز منظومة لزبون — يخصم من المخزن. لو المخزن ما يكفي يرد ٤٠٩ مع
// شرح شنو الناقص بالضبط، بدل ما يجهّز نص منظومة بهدوء.
func (h *SolarHandler) ProcessSystem(w http.ResponseWriter, r *http.Request) {
	var req model.ProcessSolarSystemRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات التجهيز غير صحيحة")
		return
	}
	req.CustomerName = strings.TrimSpace(req.CustomerName)
	if req.CustomerID == nil || strings.TrimSpace(*req.CustomerID) == "" {
		if req.CustomerName == "" || strings.TrimSpace(req.CustomerPhone) == "" {
			WriteError(w, http.StatusBadRequest, "اسم الزبون ورقم هاتفه مطلوبين")
			return
		}
	}
	if strings.TrimSpace(req.InstallDate) == "" {
		WriteError(w, http.StatusBadRequest, "تاريخ التركيب مطلوب")
		return
	}
	inst, err := h.repo.ProcessSystem(r.PathValue("id"), req, middleware.EmployeeIDFromContext(r))
	if err != nil {
		// نقص المخزن مو خطأ بالطلب — حالة تعارض، والواجهة تعرضها كتحذير
		if strings.HasPrefix(err.Error(), "ما نكدر نجهّز") {
			WriteError(w, http.StatusConflict, err.Error())
			return
		}
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, inst)
}

// GET /api/solar/installations?status=PENDING
func (h *SolarHandler) ListInstallations(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListInstallations(r.URL.Query().Get("status"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب تركيبات الزبائن")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/solar/installations/{id}/contacted
func (h *SolarHandler) MarkContacted(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Notes string `json:"notes"`
	}
	_ = DecodeJSON(r, &body)
	inst, err := h.repo.MarkContacted(r.PathValue("id"), middleware.EmployeeIDFromContext(r), strings.TrimSpace(body.Notes))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر تسجيل الاتصال")
		return
	}
	WriteJSON(w, http.StatusOK, inst)
}
