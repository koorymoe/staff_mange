package handler

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"staffmange-api/internal/model"
)

// الحقول الي ما يجوز تطلع أبداً لموظف عادي — لا بقيمتها ولا حتى باسمها.
//
// الاسم لحاله تسريب: `"salary": null` تقول للي يفتح F12 «أكو راتب مخزون».
var forbiddenForPlainEmployee = []string{
	"password", "passwordHash", "salary", "username",
	"lockedAt", "lockedReason", "lockedDetail",
	"failedLoginStreak", "authzViolations", "sessionsInvalidatedAt",
}

func sampleEmployee() *model.Employee {
	salary := 750000.0
	user := "target_user"
	reason := "FAILED_LOGINS"
	detail := "تعطيل مؤقت"
	phone := "07701234567"
	now := time.Now()
	pw := "$2a$10$hashed"
	return &model.Employee{
		ID: "emp-1", Name: "موظف الهدف", Role: "TECHNICIAN", Status: "ACTIVE",
		Salary: &salary, Username: &user, Password: &pw, Phone: &phone,
		LockedAt: &now, LockedReason: &reason, LockedDetail: &detail,
		FailedLoginStreak: 3, AuthzViolations: 5,
	}
}

func marshalView(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("تعذر ترميز الاستجابة: %v", err)
	}
	return string(b)
}

// TestPlainEmployeeSeesNoSensitiveFields الاختبار الأساسي: فني عادي يفتح
// قائمة الموظفين ما يلكه ولا حقل حساس — ولا حتى اسمه.
func TestPlainEmployeeSeesNoSensitiveFields(t *testing.T) {
	out := marshalView(t, ViewEmployee(sampleEmployee(), "TECHNICIAN", "someone-else"))
	for _, field := range forbiddenForPlainEmployee {
		if strings.Contains(out, `"`+field+`"`) {
			t.Errorf("الحقل %q طلع بالاستجابة لموظف عادي — لازم ينشال بالكامل.\nالاستجابة: %s", field, out)
		}
	}
	if !strings.Contains(out, `"name"`) || !strings.Contains(out, `"role"`) {
		t.Errorf("الاستجابة ناقصة حقول أساسية تحتاجها الواجهة: %s", out)
	}
}

// TestEmployeeSeesOwnSalary الموظف يشوف راتبه هو — بس ما يشوف راتب غيره.
func TestEmployeeSeesOwnSalary(t *testing.T) {
	e := sampleEmployee()
	own := marshalView(t, ViewEmployee(e, "TECHNICIAN", e.ID))
	if !strings.Contains(own, `"salary"`) {
		t.Errorf("الموظف ما شاف راتبه هو: %s", own)
	}
	// حتى بملفه هو، بيانات الحظر إدارية وما تخصه
	if strings.Contains(own, `"lockedReason"`) {
		t.Errorf("بيانات الحظر طلعت للموظف بملفه الشخصي: %s", own)
	}
}

// TestAdminSeesAdminFields إدارة النظام تشوف اسم المستخدم وحالة الحظر.
func TestAdminSeesAdminFields(t *testing.T) {
	out := marshalView(t, ViewEmployee(sampleEmployee(), "ADMIN", "admin-1"))
	for _, field := range []string{"username", "lockedAt", "lockedReason", "salary"} {
		if !strings.Contains(out, `"`+field+`"`) {
			t.Errorf("مدير النظام ما شاف الحقل %q: %s", field, out)
		}
	}
}

// TestPasswordNeverLeaks كلمة السر ما تطلع لأي دور كان — ولا للمالك.
func TestPasswordNeverLeaks(t *testing.T) {
	for _, role := range []string{"TECHNICIAN", "HR_COORDINATOR", "FINANCE", "ADMIN", "OWNER"} {
		out := marshalView(t, ViewEmployee(sampleEmployee(), role, "x"))
		if strings.Contains(out, "password") || strings.Contains(out, "$2a$") {
			t.Errorf("كلمة السر طلعت للدور %s: %s", role, out)
		}
	}
}

// TestHRSeesOperationalNotFinancial منسّق الكوادر يشوف الدوام والهاتف،
// بس الراتب يبقى محجوب عنه إلا إذا إله صلاحية الرواتب.
func TestHRSeesOperationalNotFinancial(t *testing.T) {
	out := marshalView(t, ViewEmployee(sampleEmployee(), "HR_COORDINATOR", "hr-1"))
	if !strings.Contains(out, `"phone"`) {
		t.Errorf("منسّق الكوادر ما شاف الهاتف: %s", out)
	}
	if strings.Contains(out, `"lockedReason"`) {
		t.Errorf("منسّق الكوادر شاف بيانات الحظر الإدارية: %s", out)
	}
}

// TestSkillsAlwaysArray الموظف الي ماعنده ولا مهارة لازم يرجع skills: []
// مو بدون الحقل أصلاً.
//
// هذا مو تجميل: حذف الحقل خلّى صفحة الموظفين تنادي .filter على undefined
// وتطيح كلها بوجه المالك. تقليل البيانات يخص القيم الحساسة، مو شكل
// القوائم الي الواجهة تمر عليها دائماً.
func TestSkillsAlwaysArray(t *testing.T) {
	e := sampleEmployee()
	e.Skills = nil
	for _, role := range []string{"TECHNICIAN", "HR_COORDINATOR", "ADMIN", "OWNER"} {
		out := marshalView(t, ViewEmployee(e, role, "x"))
		if !strings.Contains(out, `"skills":[]`) {
			t.Errorf("الدور %s ما استلم skills كمصفوفة فاضية: %s", role, out)
		}
	}
}
