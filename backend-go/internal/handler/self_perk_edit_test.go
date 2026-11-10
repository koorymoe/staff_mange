package handler

import (
	"context"
	"net/http/httptest"
	"testing"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
)

// (11-10) غير الأدمن ما يرفع راتبه أو يسوي نفسه ليدر؛ تعديل زميل أو الأدمن يمشي.
func TestSelfPerkEdit(t *testing.T) {
	sal, yes := 5000000.0, true
	call := func(role, target string, req model.UpdateEmployeeRequest) string {
		r := httptest.NewRequest("PUT", "/", nil)
		ctx := context.WithValue(r.Context(), middleware.ContextEmployeeID, "me")
		ctx = context.WithValue(ctx, middleware.ContextRole, role)
		return selfPerkEdit(r.WithContext(ctx), target, &req)
	}
	if call("HR_COORDINATOR", "me", model.UpdateEmployeeRequest{Salary: &sal}) == "" {
		t.Error("رفع راتبه لحاله")
	}
	if call("MONITOR", "me", model.UpdateEmployeeRequest{IsLeader: &yes}) == "" {
		t.Error("سوّى نفسه ليدر")
	}
	if call("HR_COORDINATOR", "other", model.UpdateEmployeeRequest{Salary: &sal}) != "" {
		t.Error("تعديل راتب زميل انرفض (هذا قرار ثاني، مو هنا)")
	}
	phone := "0770"
	if call("HR_COORDINATOR", "me", model.UpdateEmployeeRequest{Phone: &phone}) != "" {
		t.Error("تعديل تلفونه انرفض")
	}
	if call("ADMIN", "me", model.UpdateEmployeeRequest{Salary: &sal}) != "" {
		t.Error("الأدمن انرفض")
	}
}
