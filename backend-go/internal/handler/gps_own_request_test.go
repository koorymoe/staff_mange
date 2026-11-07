package handler

import (
	"context"
	"net/http/httptest"
	"testing"

	"staffmange-api/internal/middleware"
)

// (11-07) موظف عادي ما يگدر يسجّل طلب جي بي اس باسم زميله؛ الأدمن ومدير الجي بي اس يگدرون.
func TestOwnGpsRequest(t *testing.T) {
	call := func(role string) (string, bool) {
		r := httptest.NewRequest("POST", "/", nil)
		ctx := context.WithValue(r.Context(), middleware.ContextEmployeeID, "me")
		ctx = context.WithValue(ctx, middleware.ContextRole, role)
		r = r.WithContext(ctx)
		other, adm := "colleague", "boss"
		emp, admin := &other, &adm
		ownGpsRequest(r, &emp, &admin)
		return *emp, admin != nil
	}
	if emp, hasAdmin := call("SALES"); emp != "me" || hasAdmin {
		t.Fatalf("موظف عادي: employeeId=%s adminId باقي=%v", emp, hasAdmin)
	}
	for _, role := range []string{"ADMIN", "OWNER", "GPS_ADMIN"} {
		if emp, _ := call(role); emp != "colleague" {
			t.Fatalf("%s ما يگدر يحدد غيره", role)
		}
	}
}
