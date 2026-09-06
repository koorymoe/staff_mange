package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
)

// سجلّ موظف واحد (دوام، تقييم أداء، KPI، مهام، صلاحيات) بيانات شخصية:
// الموظف يشوف مالته، والمشرف يشوف مال الي تحت إيده. بدون هذا الفحص أي
// موظف مسجّل دخول يقدر يبدّل رقم الموظف بالرابط ويقرأ سجل أي زميل —
// الحماية بالواجهة ما تكفي لأن الرابط ينندى مباشرة.
//
// ليش هنا مو بالميدل وير؟ لأن رقم الموظف يجي من مسار مختلف بكل راوت
// ({id} أو {employeeId})، والميدل وير ما يعرف أي واحد منهم يخص موظف.

// selfOrSupervisor هل الطالب هو صاحب السجل، أو مشرف مخوّل يشوف غيره.
func selfOrSupervisor(r *http.Request, targetEmployeeID string) bool {
	if targetEmployeeID != "" && targetEmployeeID == middleware.EmployeeIDFromContext(r) {
		return true
	}
	return canSeeOperational(middleware.RoleFromContext(r))
}

// requireSelfOrSupervisor يكتب 403 ويرجّع false إذا الطالب مو مخوّل.
// الاستخدام:
//
//	if !requireSelfOrSupervisor(w, r, id) { return }
func requireSelfOrSupervisor(w http.ResponseWriter, r *http.Request, targetEmployeeID string) bool {
	if selfOrSupervisor(r, targetEmployeeID) {
		return true
	}
	WriteError(w, http.StatusForbidden, "ما تكدر تشوف سجل موظف غيرك")
	return false
}
