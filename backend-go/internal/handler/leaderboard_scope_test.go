package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
)

func TestLeaderboardScope(t *testing.T) {
	req := func(emp, role string) *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := context.WithValue(r.Context(), middleware.ContextEmployeeID, emp)
		return r.WithContext(context.WithValue(ctx, middleware.ContextRole, role))
	}
	rows := []model.KpiLeaderboardEntry{
		{EmployeeID: "me", EmployeeName: "أنا", Points: 10, AttendedDays: 20, CompletedBookings: 5},
		{EmployeeID: "x", EmployeeName: "زميل", Points: 12, AttendedDays: 22, CompletedBookings: 7},
	}
	out := limitKpiRows(req("me", "TECHNICIAN"), rows)
	if out[0].AttendedDays != 20 || out[0].Limited {
		t.Error("صفّي لازم يبقى كامل")
	}
	if out[1].AttendedDays != 0 || out[1].CompletedBookings != 0 || !out[1].Limited || out[1].Points != 12 || out[1].EmployeeName != "زميل" {
		t.Errorf("الزميل لازم الاسم والمجموع بس: %+v", out[1])
	}
	if full := limitKpiRows(req("boss", "ADMIN"), rows); full[1].AttendedDays != 22 || full[1].Limited {
		t.Error("المشرف يشوف التفاصيل")
	}
	sk := limitSmartKpi(req("me", "TECHNICIAN"), []model.SmartKpiResult{{EmployeeID: "x", EmployeeName: "زميل", TotalPoints: 9, Breakdown: model.SmartKpiBreakdown{}}})
	if !sk[0].Limited || sk[0].TotalPoints != 9 {
		t.Errorf("%+v", sk[0])
	}
}
