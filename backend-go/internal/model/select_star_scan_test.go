package model

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"staffmange-api/internal/models"
)

// TestSelectStarExhaustive يمسح كل كود internal/ (بلا DB) ويلقط كل استعلام
// SELECT * / SELECT alias.* / RETURNING * ويحدد الجدول، وبعدين يتأكد إن الـstruct
// الي اسمه نفس اسم الجدول (بأي باكج تحت internal/) إله حقل db لكل عمود بـinternal/models.
// جدول بلا struct بنفس الاسم لازم ينضاف لـstarStructAlias أو starIgnore مع سبب.
var (
	reStar      = regexp.MustCompile(`(?is)SELECT\s+(?:DISTINCT\s+)?\*\s+FROM\s+"([A-Za-z_]+)"`)
	reRet       = regexp.MustCompile(`(?is)(?:INSERT\s+INTO|UPDATE|DELETE\s+FROM)\s+"([A-Za-z_]+)"[\s\S]*?RETURNING\s+\*`)
	reAliasStar = regexp.MustCompile(`(?is)SELECT\s+(?:DISTINCT\s+)?([a-z_]+)\.\*`)
	reFromAlias = regexp.MustCompile(`(?is)(?:FROM|JOIN)\s+"([A-Za-z_]+)"\s+(?:AS\s+)?([a-z_]+)`)
)

// جدول → اسم struct لما يختلفون.
var starStructAlias = map[string]string{
	"AiUsage":              "AiUsageRow",
	"IntegrityRun":         "Run", // integrity.Run
	"ProjectDelegationLog": "ProjectDelegationLogEntry",
	"CustomerAccount":      "customerAccount", // handler/customer_app_handler.go
}

// جداول تنقرا بـ* بس مو لـstruct كامل (مثلاً EXISTS / COUNT داخلي).
var starIgnore = map[string]string{
	"ProcurementQuote": "LATERAL (SELECT * ...) b بـpurchase_control_handler — ينقرا منه أعمدة محددة بس",
}

type astStruct struct {
	tags   map[string]bool
	embeds []string
}

func TestSelectStarExhaustive(t *testing.T) {
	root := filepath.Join("..")
	structs := map[string][]*astStruct{}
	uses := map[string][]string{}
	destOf := map[string]string{} // table\x00struct → موقع، لما الوجهة مستنتجة وتختلف عن اسم الجدول
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(p, ".go") {
			return err
		}
		fs := token.NewFileSet()
		f, perr := parser.ParseFile(fs, p, nil, 0)
		if perr != nil {
			return nil
		}
		isTest := strings.HasSuffix(p, "_test.go")
		dest := map[token.Pos]string{}
		if !isTest {
			inferDest(f, dest)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeSpec:
				st, ok := x.Type.(*ast.StructType)
				if !ok || isTest {
					return true
				}
				s := &astStruct{tags: map[string]bool{}}
				for _, fld := range st.Fields.List {
					tag := ""
					if fld.Tag != nil {
						raw, _ := strconv.Unquote(fld.Tag.Value)
						tag = strings.Split(reflectTag(raw, "db"), ",")[0]
					}
					if len(fld.Names) == 0 && tag == "" {
						s.embeds = append(s.embeds, typeName(fld.Type))
						continue
					}
					if tag != "" && tag != "-" {
						s.tags[tag] = true
					}
				}
				structs[x.Name.Name] = append(structs[x.Name.Name], s)
			case *ast.BasicLit:
				if x.Kind != token.STRING || isTest {
					return true
				}
				q, _ := strconv.Unquote(x.Value)
				at := p + ":" + strconv.Itoa(fs.Position(x.Pos()).Line)
				add := func(tb string, idx int) {
					uses[tb] = append(uses[tb], at)
					// الوجهة المستنتجة تخص الـSELECT الخارجي بس (مو EXISTS/LATERAL داخلي).
					if d := dest[x.Pos()]; d != "" && d != tb && depthAt(q, idx) == 0 {
						destOf[tb+"\x00"+d] = at
					}
				}
				for _, m := range reStar.FindAllStringSubmatchIndex(q, -1) {
					add(q[m[2]:m[3]], m[0])
				}
				for _, m := range reRet.FindAllStringSubmatchIndex(q, -1) {
					add(q[m[2]:m[3]], m[0])
				}
				for _, m := range reAliasStar.FindAllStringSubmatchIndex(q, -1) {
					for _, a := range reFromAlias.FindAllStringSubmatch(q, -1) {
						if a[2] == q[m[2]:m[3]] {
							add(a[1], m[0])
						}
					}
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var collect func(name string, out map[string]bool, depth int) bool
	collect = func(name string, out map[string]bool, depth int) bool {
		ss := structs[name]
		if len(ss) == 0 || depth > 5 {
			return false
		}
		for _, s := range ss { // لو أكثر من struct بنفس الاسم ناخذ الاتحاد
			for k := range s.tags {
				out[k] = true
			}
			for _, e := range s.embeds {
				collect(e, out, depth+1)
			}
		}
		return true
	}
	// structs flagged يدوياً (أعمدة كانت ناقصة) — تنفحص حتى لو ماكو * حالياً، حتى ما ترجع تنقص.
	for _, tb := range []string{"AchievementVoice", "MatrixMissionLog", "SecretaryNote", "ManagerRequest", "WaMessage", "VehicleLog"} {
		if _, ok := uses[tb]; !ok {
			uses[tb] = []string{"flagged"}
		}
	}
	tables := make([]string, 0, len(uses))
	for tb := range uses {
		tables = append(tables, tb)
	}
	sort.Strings(tables)
	covered := 0
	for _, tb := range tables {
		if _, skip := starIgnore[tb]; skip {
			continue
		}
		tbl, ok := models.Get(tb)
		if !ok {
			t.Errorf("%s: يُقرأ بـ* (%s) بس مو بسجل internal/models", tb, uses[tb][0])
			continue
		}
		sn := tb
		if a, ok := starStructAlias[tb]; ok {
			sn = a
		}
		tags := map[string]bool{}
		if !collect(sn, tags, 0) {
			if hasDest(destOf, tb) {
				continue // الوجهة مستنتجة وتنفحص تحت
			}
			t.Errorf("%s: يُقرأ بـ* (%s) بس ماكو struct اسمه %s — أضفه لـstarStructAlias", tb, uses[tb][0], sn)
			continue
		}
		covered++
		for _, c := range tbl.Columns {
			if !tags[c.Name] {
				t.Errorf("%s.%s: عمود بلا حقل db بـstruct %s — SELECT * راح يفشل (%s)", tb, c.Name, sn, uses[tb][0])
			}
		}
	}
	// وجهات مستنتجة من الكود (Get/Select(&x, query)) — نفحصها هم.
	keys := make([]string, 0, len(destOf))
	for k := range destOf {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		tb, sn, _ := strings.Cut(k, "\x00")
		if _, skip := starIgnore[tb]; skip {
			continue
		}
		tbl, ok := models.Get(tb)
		if !ok {
			continue
		}
		tags := map[string]bool{}
		if !collect(sn, tags, 0) {
			continue // نوع مو struct محلي (map وغيره)
		}
		for _, c := range tbl.Columns {
			if !tags[c.Name] {
				t.Errorf("%s.%s: عمود بلا حقل db بـstruct %s — SELECT * راح يفشل (%s)", tb, c.Name, sn, destOf[k])
			}
		}
	}
	t.Logf("covered %d tables: %s", covered, strings.Join(tables, ", "))
}

func typeName(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		return typeName(x.X)
	case *ast.SelectorExpr:
		return x.Sel.Name
	}
	return ""
}

func reflectTag(raw, key string) string {
	for _, part := range strings.Fields(raw) {
		if strings.HasPrefix(part, key+":") {
			v, _ := strconv.Unquote(strings.TrimPrefix(part, key+":"))
			return v
		}
	}
	return ""
}

// inferDest يربط كل نص استعلام بنوع الوجهة لما يكون الاستدعاء بشكل
// db.Get/Select(&x, q, ...) أو QueryRowx(q).StructScan(&x) ونوع x معرّف بنفس الدالة.
func inferDest(f *ast.File, dest map[token.Pos]string) {
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}
		vars := varTypes{}
		strs := map[string][]strAssign{}
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.ValueSpec:
				for i, nm := range x.Names {
					if x.Type != nil {
						vars.set(nm.Name, nm.Pos(), elemName(x.Type))
					}
					if i < len(x.Values) {
						litsInto(x.Values[i], nm, true, vars, strs)
					}
				}
			case *ast.AssignStmt:
				for i, l := range x.Lhs {
					if id, ok := l.(*ast.Ident); ok && i < len(x.Rhs) {
						litsInto(x.Rhs[i], id, x.Tok == token.DEFINE, vars, strs)
					}
				}
			}
			return true
		})
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			c, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := c.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			var target string
			var qargs []ast.Expr
			switch sel.Sel.Name {
			case "Get", "Select", "GetContext", "SelectContext":
				for i, a := range c.Args {
					if u, ok := a.(*ast.UnaryExpr); ok && u.Op == token.AND {
						if id, ok := u.X.(*ast.Ident); ok {
							target = vars.at(id.Name, c.Pos())
							qargs = c.Args[i+1:]
						}
						break
					}
				}
			case "StructScan":
				if len(c.Args) == 1 {
					if u, ok := c.Args[0].(*ast.UnaryExpr); ok {
						if id, ok := u.X.(*ast.Ident); ok {
							target = vars.at(id.Name, c.Pos())
						}
					}
					if inner, ok := sel.X.(*ast.CallExpr); ok {
						qargs = inner.Args
					}
				}
			}
			if target == "" || len(qargs) == 0 {
				return true
			}
			ast.Inspect(qargs[0], func(m ast.Node) bool {
				switch y := m.(type) {
				case *ast.BasicLit:
					dest[y.Pos()] = target
				case *ast.Ident:
					for _, p := range strsAt(strs[y.Name], c.Pos()) {
						dest[p] = target
					}
				}
				return true
			})
			return true
		})
	}
}

type varDecl struct {
	pos token.Pos
	typ string
}

// varTypes: نوع كل متغير حسب أقرب تعريف قبل الاستدعاء (rows يتعاد تعريفه بأنواع مختلفة بنفس الدالة).
type varTypes map[string][]varDecl

func (v varTypes) set(name string, pos token.Pos, typ string) {
	v[name] = append(v[name], varDecl{pos, typ})
}

func (v varTypes) at(name string, pos token.Pos) string {
	best := varDecl{}
	for _, d := range v[name] {
		if d.pos < pos && d.pos > best.pos {
			best = d
		}
	}
	return best.typ
}

type strAssign struct {
	pos    token.Pos
	define bool
	lits   []token.Pos
}

// strsAt النصوص الي وصلت للمتغير من آخر := قبل الاستدعاء.
func strsAt(as []strAssign, pos token.Pos) []token.Pos {
	var out []token.Pos
	for _, a := range as {
		if a.pos >= pos {
			continue
		}
		if a.define {
			out = nil
		}
		out = append(out, a.lits...)
	}
	return out
}

func litsInto(e ast.Expr, id *ast.Ident, define bool, vars varTypes, strs map[string][]strAssign) {
	name := id.Name
	switch v := e.(type) {
	case *ast.CompositeLit:
		vars.set(name, id.Pos(), elemName(v.Type))
		return
	case *ast.CallExpr:
		if fn, ok := v.Fun.(*ast.Ident); ok && fn.Name == "make" && len(v.Args) > 0 {
			vars.set(name, id.Pos(), elemName(v.Args[0]))
			return
		}
	}
	a := strAssign{pos: id.Pos(), define: define}
	ast.Inspect(e, func(m ast.Node) bool {
		if bl, ok := m.(*ast.BasicLit); ok && bl.Kind == token.STRING {
			a.lits = append(a.lits, bl.Pos())
		}
		return true
	})
	strs[name] = append(strs[name], a)
}

func elemName(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.ArrayType:
		return elemName(x.Elt)
	case *ast.StarExpr:
		return elemName(x.X)
	case *ast.Ident, *ast.SelectorExpr:
		return typeName(x)
	}
	return ""
}

func depthAt(q string, idx int) int {
	d := 0
	for _, c := range q[:idx] {
		switch c {
		case '(':
			d++
		case ')':
			d--
		}
	}
	return d
}

func hasDest(destOf map[string]string, tb string) bool {
	for k := range destOf {
		if strings.HasPrefix(k, tb+"\x00") {
			return true
		}
	}
	return false
}
