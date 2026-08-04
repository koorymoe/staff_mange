package handler

import (
	"time"

	"staffmange-api/internal/model"
)

// تقليل البيانات: الموظف يستلم أقل قدر يحتاجه لشغله، والحقول الي ما تخصه
// تنشال من الـJSON بالكامل — مو تنرجع باسمها وقيمتها null.
//
// ليش الفرق مهم؟ الحقل الفاضي يفضح إن الشي موجود: `"salary": null` تقول
// للي يفتح F12 «أكو راتب مخزون، بس ما ننطيك إياه». والاسم نفسه خارطة
// للمهاجم. الحذف الكامل ما يترك أي أثر.
//
// employeeView خريطة حرة بدل ستركت ثابت، لأن مجموعة الحقول تختلف حسب دور
// الطالب — وستركت واحد بـomitempty ما يميّز بين "فاضي" و"ممنوع".
type employeeView map[string]any

// حقول يشوفها الكل (زملاء الشغل لازم يعرفون منو الموظف وشنو دوره)
func basePublicFields(e *model.Employee) employeeView {
	v := employeeView{
		"id":               e.ID,
		"name":             e.Name,
		"role":             e.Role,
		"status":           e.Status,
		"onDuty":           e.OnDuty,
		"division":         e.Division,
		"isLeader":         e.IsLeader,
		"isTrainee":        e.IsTrainee,
		"leaderSkillLevel": e.LeaderSkillLevel,
		"createdAt":        e.CreatedAt,
	}
	putIfSet(v, "jobTitle", e.JobTitle)
	putIfSet(v, "position", e.Position)
	putIfSet(v, "attendanceIcon", e.AttendanceIcon)
	// ⚠️ القوائم تنرسل دائماً — فاضية لو ماكو، مو محذوفة.
	//
	// تقليل البيانات معناه ما ننشر قيم حساسة، مو إننا نخفي شكل الحقل.
	// حذف skills لموظف بلا مهارات خلّى الواجهة تنادي .filter على
	// undefined وصفحة الموظفين كلها تطيح. المصفوفة الفاضية ما تكشف شي.
	if e.Skills == nil {
		v["skills"] = []model.EmployeeSkillDetail{}
	} else {
		v["skills"] = e.Skills
	}
	return v
}

// حقول التشغيل: الدوام والشهادات — لازمة للتنسيق وتوزيع المهام
func addOperationalFields(v employeeView, e *model.Employee) {
	v["hasDrivingLicense"] = e.HasDrivingLicense
	v["hasSafetyCertificate"] = e.HasSafetyCertificate
	v["monthlyLeaves"] = e.MonthlyLeaves
	putIfSet(v, "shift", e.Shift)
	putIfSet(v, "shiftStart", e.ShiftStart)
	putIfSet(v, "shiftEnd", e.ShiftEnd)
	putIfSet(v, "phone", e.Phone)
	putIfSet(v, "certificate", e.Certificate)
}

// حقول إدارية حساسة: اسم المستخدم وحالة الحظر — إدارة النظام حصراً
func addAdminFields(v employeeView, e *model.Employee) {
	putIfSet(v, "username", e.Username)
	putIfSet(v, "lockedAt", e.LockedAt)
	putIfSet(v, "lockedReason", e.LockedReason)
	putIfSet(v, "lockedDetail", e.LockedDetail)
}

// putIfSet يحط الحقل بس إذا إله قيمة — الفاضي ينشال بدل ما يطلع null.
func putIfSet(v employeeView, key string, val any) {
	switch x := val.(type) {
	case *string:
		if x != nil && *x != "" {
			v[key] = *x
		}
	case *float64:
		if x != nil {
			v[key] = *x
		}
	case *time.Time:
		if x != nil {
			v[key] = *x
		}
	}
}

// ViewEmployee يبني نسخة الموظف المسموحة لهذا الطالب بالتحديد.
//
//   - كل واحد يشوف بياناته هو كاملة (بدون كلمة السر طبعاً)
//   - HR/المراقب/التنسيق: بيانات التشغيل (دوام، هاتف، شهادات)
//   - من عنده صلاحية الرواتب: الراتب
//   - إدارة النظام: اسم المستخدم وحالة الحظر
//   - غير هذول: الاسم والدور والحالة بس
func ViewEmployee(e *model.Employee, viewerRole, viewerID string) employeeView {
	v := basePublicFields(e)
	self := e.ID == viewerID

	if self || canSeeOperational(viewerRole) {
		addOperationalFields(v, e)
	}
	if self || canSeeSalaries(viewerRole) {
		putIfSet(v, "salary", e.Salary)
	}
	if isSystemAdmin(viewerRole) {
		addAdminFields(v, e)
	}
	return v
}

func ViewEmployees(list []model.Employee, viewerRole, viewerID string) []employeeView {
	out := make([]employeeView, 0, len(list))
	for i := range list {
		out = append(out, ViewEmployee(&list[i], viewerRole, viewerID))
	}
	return out
}

// canSeeOperational منو يحتاج الدوام والهاتف والشهادات لشغله.
func canSeeOperational(role string) bool {
	switch role {
	case "OWNER", "ADMIN", "HR_COORDINATOR", "MONITOR", "PROJECT_MANAGER", "SERVICE_MANAGER":
		return true
	default:
		return false
	}
}

func isSystemAdmin(role string) bool {
	return role == "OWNER" || role == "ADMIN"
}
