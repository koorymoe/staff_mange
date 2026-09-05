package handler

import (
	"errors"
	"net/http"
	"strings"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type WorkReportHandler struct {
	service *service.WorkReportService
	// permissions: منو يراجع تقارير غيره. اختياري — بدونه المراجعة
	// تنحصر بالأدوار الإشرافية.
	permissions *repository.PermissionRepository
}

func NewWorkReportHandler(s *service.WorkReportService) *WorkReportHandler {
	return &WorkReportHandler{service: s}
}

// SetPermissions يربط مستودع الصلاحيات بعد البناء.
func (h *WorkReportHandler) SetPermissions(p *repository.PermissionRepository) { h.permissions = p }

// canReviewReports منو يشوف تقارير غيره.
//
// ⚠️⚠️ `selfOrSupervisor` وحدها **ما تكفي هنا**: شاشة «مراجعة تقارير
// العمل» حارسها `monitoring`/`quality_control`، ومهندس الجودة **مو**
// ضمن الأدوار الإشرافية بـ`canSeeOperational` — فالاعتماد عليها
// يفرّغ شاشته وهو شغّال اليوم.
func (h *WorkReportHandler) canReviewReports(r *http.Request) bool {
	if canSeeOperational(middleware.RoleFromContext(r)) {
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
		if p.Name == "monitoring" || p.Name == "quality_control" || p.Name == "auditing" {
			return true
		}
	}
	return false
}

func (h *WorkReportHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateWorkReportRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	report, err := h.service.Create(middleware.EmployeeIDFromContext(r), middleware.RoleFromContext(r), req)
	if err != nil {
		// ⚠️ الرفض بسبب الصلاحية يطلع ٤٠٣ مو ٤٠٠: الواجهة تفرّق
		// بينهما، و«بيانات الطلب غير صحيحة» على منع صلاحية تخلّي
		// الفني يعيد كتابة تقريره عشر مرات بلا فايدة.
		if errors.Is(err, service.ErrPaperworkNotYours) || strings.Contains(err.Error(), "مسؤول الخدمة") {
			WriteError(w, http.StatusForbidden, err.Error())
			return
		}
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, report)
}

// GET /api/work-reports?employeeId= — بدون employeeId: كل التقارير (للمراقب/الجودة)
func (h *WorkReportHandler) List(w http.ResponseWriter, r *http.Request) {
	// ⚠️ چان `employeeId` فاضي يرجّع **كل تقارير الشركة** لأي موظف
	// مسجّل. هسه: المراجِع (مراقبة/جودة) يشوف الكل، وغيره يشوف
	// تقاريره هو مهما كتب بالرابط — نفس نمط `ExpenseHandler`.
	employeeID := r.URL.Query().Get("employeeId")
	self := middleware.EmployeeIDFromContext(r)
	if employeeID != self && !h.canReviewReports(r) {
		employeeID = self
	}
	reports, err := h.service.List(employeeID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب تقارير العمل")
		return
	}
	WriteJSON(w, http.StatusOK, reports)
}
