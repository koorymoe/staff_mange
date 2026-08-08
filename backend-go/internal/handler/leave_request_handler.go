package handler

import (
	"log"
	"net/http"
	"time"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type LeaveRequestHandler struct {
	repo        *repository.LeaveRequestRepository
	permissions *repository.PermissionRepository
	notify      *repository.NotificationRepository
}

func NewLeaveRequestHandler(r *repository.LeaveRequestRepository, p *repository.PermissionRepository, n *repository.NotificationRepository) *LeaveRequestHandler {
	return &LeaveRequestHandler{repo: r, permissions: p, notify: n}
}

// routesFor الشفتات الي يقدر هذا الشخص يبت بطلباتها.
//
// المالك ومدير النظام يغطون كل الشفتات وكل الموظفين. غيرهم حسب
// صلاحياتهم — إداري الكوادر ياخذ صلاحية شفته هو بس، فما يقدر ينطي
// إجازة لموظف بشفت ثاني. ولو انتقلت المسؤولية من شخص لشخص تنتقل
// بالصلاحية مو بتعديل الكود.
func (h *LeaveRequestHandler) routesFor(r *http.Request) []string {
	role := middleware.RoleFromContext(r)
	if role == "OWNER" || role == "ADMIN" {
		return model.AllLeaveRoutes()
	}
	empID := middleware.EmployeeIDFromContext(r)
	routes := []string{}
	for route, perm := range model.LeaveRoutePermission {
		if ok, err := h.permissions.HasPermission(empID, perm); err == nil && ok {
			routes = append(routes, route)
		}
	}
	return routes
}

// POST /api/leaves — الموظف يقدّم طلب إجازة

// blockSelfDecision يمنع الشخص يبت بطلب إجازته هو.
//
// المسار ينتحدد من **شفت مقدّم الطلب**، ومسؤول الشفت الصباحي نفسه
// صباحي — يعني طلب إجازته يوصل لصندوقه هو، ويوافق عليه بضغطة زر بلا
// ما يشوفه أحد. هاي مو ثغرة تقنية، هاي ثغرة رقابية: أي واحد عنده
// صلاحية الموافقة يكدر ياخذ إجازات بلا حسيب.
//
// المالك مستثنى — هو أعلى سلطة وماكو فوقه أحد يوافقه، ولو منعناه
// يبقى طلبه معلق للأبد.
func (h *LeaveRequestHandler) blockSelfDecision(w http.ResponseWriter, r *http.Request, leaveOwnerID string) bool {
	if middleware.RoleFromContext(r) == "OWNER" {
		return false
	}
	if leaveOwnerID != "" && leaveOwnerID == middleware.EmployeeIDFromContext(r) {
		WriteError(w, http.StatusForbidden, "ما تكدر تبت بطلب إجازتك إنت — لازم يوافق عليه المالك أو مخوّل ثاني")
		return true
	}
	return false
}

func (h *LeaveRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateLeaveRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "تاريخ بداية الإجازة غير صحيح")
		return
	}
	end := start
	if req.EndDate != "" {
		if end, err = time.Parse("2006-01-02", req.EndDate); err != nil {
			WriteError(w, http.StatusBadRequest, "تاريخ نهاية الإجازة غير صحيح")
			return
		}
	}
	if end.Before(start) {
		WriteError(w, http.StatusBadRequest, "تاريخ النهاية قبل البداية")
		return
	}
	// الإجازة تُطلب قبل يومين على الأقل — حتى يلحكون يرتبون الشفت
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	minStart := today.AddDate(0, 0, model.LeaveMinNoticeDays)
	if start.Before(minStart) {
		WriteError(w, http.StatusBadRequest, "لازم تقدّم طلب الإجازة قبل يومين على الأقل من تاريخها")
		return
	}

	leave, err := h.repo.Create(middleware.EmployeeIDFromContext(r), start, end, req.Reason)
	if err != nil {
		log.Printf("create leave: %v", err)
		WriteError(w, http.StatusBadRequest, "تعذر تقديم طلب الإجازة")
		return
	}

	// ننبّه المخوّل بهذا المسار حتى ما يضل الطلب معلّق بلا ما أحد يدري
	if h.notify != nil {
		msg := "🏖️ طلب إجازة جديد من " + leave.EmployeeName + " بتاريخ " + leave.StartDate.Format("2006-01-02")
		if perm, ok := model.LeaveRoutePermission[leave.Route]; ok {
			_ = h.notify.CreateForPermission(perm, "leave_request", msg)
		}
		_ = h.notify.CreateForRole("OWNER", "leave_request", msg)
	}
	WriteJSON(w, http.StatusOK, leave)
}

// GET /api/leaves/mine — طلبات الموظف الحالي وأرشيفه
func (h *LeaveRequestHandler) Mine(w http.ResponseWriter, r *http.Request) {
	rows, err := h.repo.ListForEmployee(middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب طلباتك")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/leaves/inbox?status= — الطلبات الي يحق للمستخدم يبت بيها
func (h *LeaveRequestHandler) Inbox(w http.ResponseWriter, r *http.Request) {
	routes := h.routesFor(r)
	if len(routes) == 0 {
		WriteError(w, http.StatusForbidden, "ما عندك صلاحية الموافقة على الإجازات")
		return
	}
	rows, err := h.repo.ListForApprover(routes, r.URL.Query().Get("status"))
	if err != nil {
		log.Printf("leave inbox: %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر جلب طلبات الإجازة")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// PUT /api/leaves/{id}/decide — موافقة أو رفض
// PUT /api/leaves/{id}/preliminary — موافقة أولية + ملاحظة.
//
// الموظف ياخذ إشعار إنها أولية ومو نهائية، والطلب يبقى بصندوق المدير
// لحد ما يبت بيه — ما ينشال، وإلا ينتسى والموظف ينظلم.
func (h *LeaveRequestHandler) Preliminary(w http.ResponseWriter, r *http.Request) {
	var req model.PreliminaryLeaveRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	id := r.PathValue("id")
	route, err := h.repo.RouteOf(id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "الطلب غير موجود")
		return
	}
	allowed := false
	for _, rt := range h.routesFor(r) {
		if rt == route {
			allowed = true
			break
		}
	}
	if !allowed {
		WriteError(w, http.StatusForbidden, "ما عندك صلاحية البت بطلبات "+model.LeaveRouteLabels[route])
		return
	}

	if owner, err := h.repo.OwnerOf(id); err == nil && h.blockSelfDecision(w, r, owner) {
		return
	}

	leave, err := h.repo.Preliminary(id, req.Note, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.notify != nil {
		msg := "🔎 اجت موافقة أولية على طلب إجازتك بتاريخ " + leave.StartDate.Format("2006-01-02") +
			" — لسّه بحالة المراجعة، القرار النهائي راح يوصلك"
		if leave.PreliminaryNote != nil && *leave.PreliminaryNote != "" {
			msg += " — ملاحظة: " + *leave.PreliminaryNote
		}
		_ = h.notify.Create(leave.EmployeeID, "leave_preliminary", msg)
	}
	WriteJSON(w, http.StatusOK, leave)
}

func (h *LeaveRequestHandler) Decide(w http.ResponseWriter, r *http.Request) {
	var req model.DecideLeaveRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "بيانات الطلب غير صحيحة")
		return
	}
	id := r.PathValue("id")
	route, err := h.repo.RouteOf(id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "الطلب غير موجود")
		return
	}
	// ما يكفي إنه مخوّل بالإجازات — لازم يكون مخوّل *بهذا المسار* بالذات،
	// وإلا مسؤول الكوادر المسائية يوافق على إجازات الفنيين.
	allowed := false
	for _, rt := range h.routesFor(r) {
		if rt == route {
			allowed = true
			break
		}
	}
	if !allowed {
		WriteError(w, http.StatusForbidden, "ما عندك صلاحية البت بطلبات "+model.LeaveRouteLabels[route])
		return
	}

	if owner, err := h.repo.OwnerOf(id); err == nil && h.blockSelfDecision(w, r, owner) {
		return
	}

	leave, err := h.repo.Decide(id, req.Approve, req.Note, middleware.EmployeeIDFromContext(r))
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// يرجع تقرير للموظف — هذا الي طلبته: يعرف إن طلبه انقبل أو انرفض
	if h.notify != nil {
		verdict := "❌ انرفض طلب إجازتك بتاريخ " + leave.StartDate.Format("2006-01-02")
		if req.Approve {
			verdict = "✅ انقبل طلب إجازتك بتاريخ " + leave.StartDate.Format("2006-01-02")
		}
		if leave.DecisionNote != nil && *leave.DecisionNote != "" {
			verdict += " — " + *leave.DecisionNote
		}
		_ = h.notify.Create(leave.EmployeeID, "leave_decision", verdict)
	}
	WriteJSON(w, http.StatusOK, leave)
}

// DELETE /api/leaves/{id} — الموظف يسحب طلبه قبل البت بيه
func (h *LeaveRequestHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Cancel(r.PathValue("id"), middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/leaves/pending-count — شارة العدد بالقائمة، وهل هذا الشخص مخوّل أصلاً.
//
// canApprove يرجع من هنا حتى الواجهة ما تضطر تجرّب /inbox وتاكل 403 لكل
// موظف عادي — الرد 200 للكل، والفرق بالمحتوى مو بالخطأ.
func (h *LeaveRequestHandler) PendingCount(w http.ResponseWriter, r *http.Request) {
	routes := h.routesFor(r)
	n, err := h.repo.PendingCountFor(routes)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر العد")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"count": n, "canApprove": len(routes) > 0})
}
