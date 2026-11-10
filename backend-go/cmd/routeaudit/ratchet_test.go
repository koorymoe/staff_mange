package main

import (
	"os"
	"regexp"
	"testing"
)

// قفل «ون بيس» (11-10): الأرقام تنزل بس، ما تصعد. أي مسار جديد ينكتب عبر
// internal/core (فعل + قاعدة)، مو mux.Handle بحارس الدخول وحده. إذا نقلت
// مسار للنواة نزّل الرقم هنا.
const (
	maxPublic        = 39 // كلها مقصودة: الزبائن، الدخول، روابط الاستقبال، الملفات الموقّعة
	maxWriteAuthOnly = 0  // (11-10) كلها انتقلت للنواة — ولا وحدة جديدة
	maxLegacy        = 39 // المسارات القديمة (mux.Handle) — الجديد ينكتب بالنواة، والرقم ينزل بس
)

func TestRouteRatchet(t *testing.T) {
	routes, err := Audit("../api/main.go")
	if err != nil {
		t.Fatal(err)
	}
	pub, write := 0, 0
	for _, r := range routes {
		switch r.Risk {
		case "عام (بلا دخول)":
			pub++
			if pub > maxPublic {
				t.Errorf("مسار عام جديد: %s %s (سطر %d) — لازم يكون مقصود، وإذا مقصود ارفع الحد ويا السبب", r.Method, r.Path, r.Line)
			}
		case "كتابة بالدخول بس":
			write++
			if write > maxWriteAuthOnly {
				t.Errorf("كتابة جديدة بحارس الدخول بس: %s %s (سطر %d) — اكتبها عبر internal/core بقاعدة", r.Method, r.Path, r.Line)
			}
		}
	}
	if len(routes) > maxLegacy {
		t.Errorf("مسارات قديمة %d > %d — المسار الجديد ينكتب عبر internal/core (فعل + قاعدة)", len(routes), maxLegacy)
	}
	if pub < maxPublic || write < maxWriteAuthOnly {
		t.Logf("👍 نزل الرقم (عام %d، كتابة %d) — نزّل الحد بهذا الملف", pub, write)
	}
}

// (11-10) كل النظام صار على النواة: ممنوع mux.Handle بمسار مو نص ثابت (حلقة/متغير)
// — هيچ مسارات تفلت من الجرد ومن القفل. سجّلها بـcore (SetApp.Add).
func TestNoDynamicMuxRoutes(t *testing.T) {
	b, err := os.ReadFile("../api/main.go")
	if err != nil {
		t.Fatal(err)
	}
	if m := regexp.MustCompile(`mux\.Handle(Func)?\([^"]`).FindAll(b, -1); len(m) > 0 {
		t.Fatalf("mux.Handle بمسار متغير (%d) — سجّله عبر internal/core", len(m))
	}
}
