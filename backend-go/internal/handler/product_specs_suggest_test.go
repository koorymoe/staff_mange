package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProductSpecsSuggest(t *testing.T) {
	var lines []string
	lines = append(lines, "تفضل المواصفات:", "", "- **الدقة**: 4MP", "* العدسة: 2.8mm")
	for i := 0; i < 30; i++ {
		lines = append(lines, "بند: قيمة")
	}
	h := &ProductSpecsSuggest{Ask: func(_, task string) (string, error) {
		if !strings.Contains(task, "DS-2CD") {
			t.Errorf("الموديل ما وصل: %q", task)
		}
		return strings.Join(lines, "\n"), nil
	}}
	w := httptest.NewRecorder()
	h.Suggest(w, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"كاميرا","modelName":"DS-2CD"}`)))
	var out struct{ Specs string }
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	got := strings.Split(out.Specs, "\n")
	if len(got) != 20 || got[0] != "الدقة: 4MP" || got[1] != "العدسة: 2.8mm" {
		t.Fatalf("%d %q", w.Code, got[:3])
	}
	// ما يعرف = فاضي
	h.Ask = func(_, _ string) (string, error) { return "ماعرف", nil }
	w = httptest.NewRecorder()
	h.Suggest(w, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"شي"}`)))
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	if out.Specs != "" {
		t.Fatalf("لازم فاضي: %q", out.Specs)
	}
}
