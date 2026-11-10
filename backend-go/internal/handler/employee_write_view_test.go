package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
)

// (11-10) رد تعديل الموظف: المفوَّض (غير أدمن) ما يشوف راتب وتلفون ويوزر الزميل؛ الأدمن يشوف الكل.
func TestWriteEmployeeForHidesPrivate(t *testing.T) {
	sal, phone, user := 900000.0, "07701234567", "colleague_u"
	e := &model.Employee{ID: "other", Name: "زميل", Salary: &sal, Phone: &phone, Username: &user}
	call := func(role string) map[string]any {
		r := httptest.NewRequest("PUT", "/", nil)
		ctx := context.WithValue(r.Context(), middleware.ContextEmployeeID, "me")
		ctx = context.WithValue(ctx, middleware.ContextRole, role)
		w := httptest.NewRecorder()
		writeEmployeeFor(w, r.WithContext(ctx), e)
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return out
	}
	got := call("TECHNICIAN")
	for _, k := range []string{"salary", "phone", "username"} {
		if _, ok := got[k]; ok {
			t.Errorf("%s طلع لغير الأدمن", k)
		}
	}
	if got["name"] != "زميل" {
		t.Errorf("الاسم لازم يطلع: %v", got)
	}
	if _, ok := call("ADMIN")["salary"]; !ok {
		t.Error("الأدمن لازم يشوف الراتب")
	}
}
