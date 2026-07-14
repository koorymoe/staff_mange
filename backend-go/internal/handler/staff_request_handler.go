package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type StaffRequestHandler struct {
	repo *repository.StaffRequestRepository
}

func NewStaffRequestHandler(repo *repository.StaffRequestRepository) *StaffRequestHandler {
	return &StaffRequestHandler{repo: repo}
}

// Create مدير المشاريع (أو من عنده صلاحية إدارة المشاريع) يقدم طلب كادر
func (h *StaffRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateStaffRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	if len(req.EmployeeIDs) == 0 {
		WriteError(w, http.StatusBadRequest, "اختر موظف واحد على الأقل")
		return
	}
	if req.NeededAt.IsZero() {
		WriteError(w, http.StatusBadRequest, "حدد وقت الحاجة للكادر")
		return
	}
	if req.DurationHours <= 0 {
		req.DurationHours = 8
	}
	sr, err := h.repo.Create(middleware.EmployeeIDFromContext(r), req)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر إنشاء طلب الكادر")
		return
	}
	WriteJSON(w, http.StatusCreated, sr)
}

// List إدارة الكوادر والأدمن يشوفون كل الطلبات؛ الباقي يشوف طلباته فقط (?mine=1 يجبرها)
func (h *StaffRequestHandler) List(w http.ResponseWriter, r *http.Request) {
	role := middleware.RoleFromContext(r)
	requesterID := middleware.EmployeeIDFromContext(r)
	if (role == "ADMIN" || role == "HR_COORDINATOR") && r.URL.Query().Get("mine") != "1" {
		requesterID = ""
	}
	list, err := h.repo.List(requesterID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب طلبات الكادر")
		return
	}
	WriteJSON(w, http.StatusOK, list)
}

// UpdateStatus إدارة الكوادر تلبي/ترفض الطلب
func (h *StaffRequestHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateStaffRequestStatus
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	valid := map[string]bool{"APPROVED": true, "REJECTED": true, "FULFILLED": true, "PENDING": true}
	if !valid[req.Status] {
		WriteError(w, http.StatusBadRequest, "حالة غير معروفة")
		return
	}
	sr, err := h.repo.UpdateStatus(r.PathValue("id"), req.Status, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر تحديث حالة الطلب")
		return
	}
	WriteJSON(w, http.StatusOK, sr)
}
