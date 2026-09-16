package handler

import (
	"log"
	"net/http"
	"staffmange-api/internal/middleware"

	"staffmange-api/internal/model"
	"staffmange-api/internal/service"
)

type ProjectHandler struct {
	service *service.ProjectService
	// canSeeAll يقرر هل هذا المستخدم يشوف كل المشاريع لو لا — يُحقن من main
	// حتى الهاندلر ما يحتاج يعرف تفاصيل مستودع الصلاحيات.
	canSeeAll func(r *http.Request) bool
}

func NewProjectHandler(s *service.ProjectService, canSeeAll func(r *http.Request) bool) *ProjectHandler {
	return &ProjectHandler{service: s, canSeeAll: canSeeAll}
}

// GET /api/v1/projects
//
// أمان: هذا الراوت مفتوح لأي موظف مسجّل دخول لأنه شاشات كثيرة تحتاج أسماء
// المشاريع. بس ما يصير أي موظف يشوف *كل* المشاريع — الي ما عنده صلاحية
// إدارة المشاريع يشوف بس المشاريع الموجّهة له. قبل هذا الإصلاح كان أي فني
// يقدر يفتح /api/projects ويشوف كل مشاريع الشركة.
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	if h.canSeeAll != nil && !h.canSeeAll(r) {
		result, err := h.service.ListDelegatedTo(middleware.EmployeeIDFromContext(r))
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "تعذر جلب المشاريع")
			return
		}
		WriteJSON(w, http.StatusOK, result)
		return
	}
	result, err := h.service.List()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المشاريع")
		return
	}
	WriteJSON(w, http.StatusOK, result)
}

// GET /api/projects/delegated-to-me — مشاريع مُسلَّمة للموظف الحالي. ما تحتاج
// صلاحية إدارة مشاريع: التسليم نفسه هو الصلاحية، وعلى هذي المشاريع بس.
func (h *ProjectHandler) ListDelegatedToMe(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ListDelegatedTo(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب المشاريع المُسلَّمة لك")
		return
	}
	WriteJSON(w, http.StatusOK, result)
}

// PUT /api/projects/{id}/delegate — تسليم المشروع لموظف أو سحبه منه.
func (h *ProjectHandler) Delegate(w http.ResponseWriter, r *http.Request) {
	var req model.DelegateProjectRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	by := middleware.EmployeeIDFromContext(r)
	p, err := h.service.Delegate(r.PathValue("id"), req.EmployeeID, &by, req.Note)
	if err != nil {
		// ما نرجّع نص خطأ قاعدة البيانات للمستخدم (تسريب تفاصيل داخلية) — نسجّله بس
		log.Printf("delegate project %s failed: %v", r.PathValue("id"), err)
		WriteError(w, http.StatusBadRequest, "تعذر تسليم المشروع")
		return
	}
	WriteJSON(w, http.StatusOK, p)
}

// GET /api/projects/statistics — إحصائيات المشاريع والموظفين داخلها.
func (h *ProjectHandler) Statistics(w http.ResponseWriter, r *http.Request) {
	res, err := h.service.Statistics()
	if err != nil {
		log.Printf("project statistics failed: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب إحصائيات المشاريع")
		return
	}
	WriteJSON(w, http.StatusOK, res)
}

// GET /api/projects/{id}/delegation-log — سجل تسليم مشروع معيّن.
func (h *ProjectHandler) DelegationLog(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.DelegationLog(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب سجل التسليم")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/projects/{id} — المشروع كامل بما بيه ملفات العقد (تُطلب عند الحاجة فقط)
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// نفس قاعدة القائمة: بدون صلاحية إدارة المشاريع، ما يفتح إلا مشروعه الموجّه
	if h.canSeeAll != nil && !h.canSeeAll(r) {
		ok, err := h.service.IsDelegatedTo(id, middleware.EmployeeIDFromContext(r))
		if err != nil || !ok {
			WriteError(w, http.StatusForbidden, "لا تملك صلاحية الوصول لهذا المشروع")
			return
		}
	}
	project, err := h.service.Get(id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "المشروع غير موجود")
		return
	}
	WriteJSON(w, http.StatusOK, project)
}

// POST /api/v1/projects
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProjectRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	// نسجّل منو أضاف المشروع (أو رحّل الحجز) حتى يظهر ببطاقته
	var createdBy *string
	if id := middleware.EmployeeIDFromContext(r); id != "" {
		createdBy = &id
	}
	project, err := h.service.Create(req, createdBy)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, project)
}

// PUT /api/v1/projects/{id}
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateProjectRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	project, err := h.service.Update(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, project)
}

// DELETE /api/projects/{id}/contract?which=plain|signed|both — حذف ملف العقد
// المرفوع. الراوت محمي بـrequireAdmin، يعني مدير النظام بس يقدر يحذف.
func (h *ProjectHandler) DeleteContract(w http.ResponseWriter, r *http.Request) {
	which := r.URL.Query().Get("which")
	req := model.UpdateProjectRequest{}
	switch which {
	case "plain":
		req.ClearContract = true
	case "signed":
		req.ClearSignedContract = true
	case "", "both":
		req.ClearContract = true
		req.ClearSignedContract = true
	default:
		WriteError(w, http.StatusBadRequest, "قيمة غير صحيحة")
		return
	}
	project, err := h.service.Update(r.PathValue("id"), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, project)
}

// DELETE /api/v1/projects/{id}
func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.PathValue("id")); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
