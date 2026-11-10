package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
)

// (11-10) المسار القديم للصلاحيات: الأدمن ما يمنح ولا يسحب صلاحية مقفولة على المالك؛ المالك يگدر؛
// والصلاحيات العادية تمشي للأدمن مثل قبل.
func TestLegacyPermLock(t *testing.T) {
	all := []model.Permission{{ID: "1", Name: "finance_audit"}, {ID: "2", Name: "bookings"}}
	req := func(role string) *httptestReq {
		r := httptest.NewRequest("PUT", "/", nil)
		return &httptestReq{r.WithContext(context.WithValue(r.Context(), middleware.ContextRole, role))}
	}
	admin, owner := req("ADMIN"), req("OWNER")
	if legacyPermLock(admin.r, all, nil, []string{"1"}) == "" {
		t.Error("الأدمن منح صلاحية المالك")
	}
	if legacyPermLock(admin.r, all, []model.Permission{{ID: "1", Name: "finance_audit"}}, nil) == "" {
		t.Error("الأدمن سحب صلاحية المالك")
	}
	if why := legacyPermLock(admin.r, all, []model.Permission{{ID: "1", Name: "finance_audit"}}, []string{"1", "2"}); why != "" {
		t.Errorf("إضافة صلاحية عادية مع إبقاء المقفولة انرفضت: %s", why)
	}
	if why := legacyPermLock(owner.r, all, nil, []string{"1"}); why != "" {
		t.Errorf("المالك انرفض: %s", why)
	}
}

type httptestReq struct{ r *http.Request }
