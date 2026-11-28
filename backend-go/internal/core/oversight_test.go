package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"staffmange-api/internal/middleware"
)

type fixedPerms []string

func (f fixedPerms) Names(string) ([]string, error) { return f, nil }

// (دازاي 11-28) صلاحية «رقابة وتدقيق — مشاهدة»: تفتح الحزمة للقراءة، والتعديل يبقى مسدود.
func TestOversightViewReadOnly(t *testing.T) {
	req := func(m string) *http.Request {
		r := httptest.NewRequest(m, "/x", nil)
		ctx := context.WithValue(r.Context(), middleware.ContextEmployeeID, "e1")
		ctx = context.WithValue(ctx, middleware.ContextRole, "TECHNICIAN")
		return r.WithContext(ctx)
	}
	rule := Rule{Label: "x", Roles: []string{"MONITOR"}, Perms: []string{"monitoring"}}
	get := resolveActor(req(http.MethodGet), fixedPerms{"oversight_view"}, nil, nil)
	if !decide("a.x", rule, get, nil, false) || !get.Can("finance_audit", "audit_issues") {
		t.Fatal("القراءة لازم تنفتح بالصلاحية")
	}
	post := resolveActor(req(http.MethodPost), fixedPerms{"oversight_view"}, nil, nil)
	if decide("a.x", rule, post, nil, false) {
		t.Fatal("التعديل لازم يبقى مسدود")
	}
	none := resolveActor(req(http.MethodGet), fixedPerms{}, nil, nil)
	if decide("a.x", rule, none, nil, false) {
		t.Fatal("بلا صلاحية ما ينفتح")
	}
}
