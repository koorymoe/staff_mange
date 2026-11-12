// routeaudit — «ون بيس»: جرد المسارات القديمة من main.go (قراءة الكود، بلا تشغيل).
// لكل mux.Handle: الطريقة، المسار، والحراس بالسلسلة. ويأشّر الخطر:
//   - كتابة (POST/PUT/PATCH/DELETE) بحارس الدخول بس؛
//   - مسار بلا حارس دخول أصلاً (عام).
//
// go run ./cmd/routeaudit [-json] cmd/api/main.go
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Route struct {
	Method  string   `json:"method"`
	Path    string   `json:"path"`
	Guards  []string `json:"guards"`
	Line    int      `json:"line"`
	Risk    string   `json:"risk,omitempty"`
	Handler string   `json:"handler"`
}

func exprName(fs *token.FileSet, e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		return exprName(fs, v.X) + "." + v.Sel.Name
	case *ast.IndexExpr:
		return exprName(fs, v.X) + "[" + exprName(fs, v.Index) + "]"
	case *ast.BasicLit:
		return v.Value
	case *ast.CallExpr:
		args := []string{}
		for _, a := range v.Args {
			args = append(args, exprName(fs, a))
		}
		return exprName(fs, v.Fun) + "(" + strings.Join(args, ",") + ")"
	case *ast.FuncLit:
		return "func"
	case *ast.CompositeLit:
		return "{…}"
	}
	return fmt.Sprintf("%T", e)
}

// unwrap يفك الطبقات: middleware.Chain(h, mws...)، و X(mw)(h)، و http.HandlerFunc(f).
func unwrap(fs *token.FileSet, e ast.Expr, guards *[]string) string {
	call, ok := e.(*ast.CallExpr)
	if !ok {
		return exprName(fs, e)
	}
	fn := exprName(fs, call.Fun)
	switch {
	case fn == "middleware.Chain":
		for _, m := range call.Args[1:] {
			*guards = append(*guards, exprName(fs, m))
		}
		return unwrap(fs, call.Args[0], guards)
	case fn == "http.HandlerFunc" || fn == "http.Handler":
		return unwrap(fs, call.Args[0], guards)
	}
	// شكل mw(h): نداء بوسيط واحد — الدالة حارس
	if inner, ok := call.Fun.(*ast.CallExpr); ok || len(call.Args) == 1 {
		_ = inner
		*guards = append(*guards, fn)
		if len(call.Args) == 1 {
			return unwrap(fs, call.Args[0], guards)
		}
	}
	return fn
}

// Audit يقرا ملف المسارات ويرجع الجرد.
func Audit(path string) ([]Route, error) {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, path, nil, 0)
	if err != nil {
		return nil, err
	}
	routes := []Route{}
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) != 2 {
			return true
		}
		if s, ok := call.Fun.(*ast.SelectorExpr); !ok || s.Sel.Name != "Handle" || exprName(fs, s.X) != "mux" {
			return true
		}
		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok {
			return true
		}
		pat, _ := strconv.Unquote(lit.Value)
		method, p := "ANY", pat
		if i := strings.IndexByte(pat, ' '); i > 0 {
			method, p = pat[:i], pat[i+1:]
		}
		r := Route{Method: method, Path: p, Line: fs.Position(call.Pos()).Line}
		r.Handler = unwrap(fs, call.Args[1], &r.Guards)
		auth := false
		for _, g := range r.Guards {
			if g == "requireAuth" || strings.Contains(g, "RequireAuth") {
				auth = true
			}
		}
		write := method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE"
		switch {
		case !auth:
			r.Risk = "عام (بلا دخول)"
		case write && len(r.Guards) == 1:
			r.Risk = "كتابة بالدخول بس"
		}
		routes = append(routes, r)
		return true
	})
	sort.Slice(routes, func(i, j int) bool { return routes[i].Path < routes[j].Path })
	return routes, nil
}

func main() {
	asJSON := len(os.Args) > 1 && os.Args[1] == "-json"
	routes, err := Audit(os.Args[len(os.Args)-1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if asJSON {
		_ = json.NewEncoder(os.Stdout).Encode(routes)
		return
	}
	risky := map[string]int{}
	for _, r := range routes {
		if r.Risk != "" {
			risky[r.Risk]++
		}
	}
	fmt.Printf("المسارات: %d\n", len(routes))
	for k, v := range risky {
		fmt.Printf("  %s: %d\n", k, v)
	}
}
