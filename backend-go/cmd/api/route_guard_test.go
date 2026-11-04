package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// ═══ 🛡️ اختبار الحارس (تعديلات بند ١، المرحلة ٣ — 11-04) ═══
// يمر على كل مسار بـ main.go: لازم يكون محمي بدخول الموظف (requireAuth) أو
// دخول الزبون (RequireCustomer). أي مسار جديد بلا حماية **يفشّل البناء** — إلا
// إذا انضاف هنا لقائمة العامّة ويا سببه، يعني قرار مقصود ومكتوب.
var publicRoutes = map[string]string{
	"GET /api/health":                              "فحص الخادم للمراقبة",
	"POST /api/auth/login":                         "الدخول نفسه (ويا تحديد المحاولات)",
	"GET /api/files/":                              "<img> ما يرسل ترويسة — الوسم الموقّع يتحقق بالمعالج",
	"GET /api/public/design-forms/{token}":         "استمارة تصميم للزبون برابط موقّع (ويا تحديد)",
	"POST /api/public/design-forms/{token}/submit": "إرسال الاستمارة برابط موقّع (ويا تحديد)",
	"POST /api/c/signup":                           "تطبيق الزبائن: تسجيل (ويا تحديد)",
	"POST /api/c/verify":                           "تطبيق الزبائن: تأكيد (ويا تحديد)",
	"POST /api/c/resend":                           "تطبيق الزبائن: إعادة إرسال (ويا تحديد)",
	"POST /api/c/login":                            "تطبيق الزبائن: دخول (ويا تحديد)",
	"POST /api/c/forgot":                           "تطبيق الزبائن: نسيت الرمز (ويا تحديد)",
	"POST /api/c/reset":                            "تطبيق الزبائن: تغيير الرمز (ويا تحديد)",
	"GET /api/c/img/{id}":                          "صور الخدمات للتطبيق (ويا تحديد)",
	"GET /api/c/img/s/{id}":                        "صور الخدمات للتطبيق (ويا تحديد)",
	"GET /api/wa/webhook":                          "ميتا تتحقق من الويب هوك (رمز التحقق بالمعالج)",
	"POST /api/wa/webhook":                         "رسائل واتساب من ميتا (التوقيع بالمعالج)",
}

func TestEveryRouteIsGuarded(t *testing.T) {
	src, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	var open []string
	seen := map[string]bool{}
	total, authOnly := 0, 0
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || (sel.Sel.Name != "Handle" && sel.Sel.Name != "HandleFunc") || len(call.Args) < 2 {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); !ok || id.Name != "mux" {
			return true
		}
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok {
			// دالة مساعدة تسجّل مسارات (مثل tcRoute) — لازم الحماية بجسمها نفسه
			b := string(src[fset.Position(call.Args[1].Pos()).Offset:fset.Position(call.End()).Offset])
			if !strings.Contains(b, "requireAuth") && !strings.Contains(b, "RequireCustomer") {
				t.Errorf("%s: مسار من دالة مساعدة بلا حماية", fset.Position(call.Pos()))
			}
			return true
		}
		pattern, _ := strconv.Unquote(lit.Value)
		total++
		seen[pattern] = true
		body := string(src[fset.Position(call.Args[1].Pos()).Offset:fset.Position(call.End()).Offset])
		switch {
		case strings.Contains(body, "requireAuth"):
			if strings.Count(body, "require")+strings.Count(body, "Require")+strings.Count(body, "guards.")+strings.Count(body, "authz.") <= 1 {
				authOnly++ // دخول بس — الصلاحية جوّا المعالج (المرحلة ٤ تنقلها لـ authz.Can)
			}
		case strings.Contains(body, "RequireCustomer"), strings.HasPrefix(strings.TrimSpace(body), "rc("),
			strings.Contains(body, "(rc("):
		default:
			if _, ok := publicRoutes[pattern]; !ok {
				open = append(open, fmt.Sprintf("%s  (%s)", pattern, fset.Position(call.Pos())))
			}
		}
		return true
	})
	sort.Strings(open)
	if len(open) > 0 {
		t.Errorf("❌ %d مسار بلا حماية — ضيف requireAuth، أو إذا مقصود عام ضيفه لـ publicRoutes ويا السبب:\n%s", len(open), strings.Join(open, "\n"))
	}
	for p := range publicRoutes {
		if !seen[p] {
			t.Errorf("publicRoutes بيها %q وهو مو موجود بـ main.go — شيله", p)
		}
	}
	t.Logf("المسارات: %d · دخول بس (الصلاحية بالمعالج): %d · عامة مقصودة: %d", total, authOnly, len(publicRoutes))
}
