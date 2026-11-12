package models

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Plan الفرق بين الموديلات والقاعدة.
//   - Safe: تحديثات إضافة بس (جدول، عمود، قيد، فهرس، نوع، قيمة نوع) — تنكتب تلقائياً.
//   - NeedsApproval: أي حذف أو تغيير على شي موجود — ما يتولّد أبداً بلا موافقة دازاي.
type Plan struct {
	Safe          []string
	NeedsApproval []string
}

func (p Plan) Empty() bool { return len(p.Safe) == 0 && len(p.NeedsApproval) == 0 }

func qi(s string) string { return `"` + strings.ReplaceAll(s, `"`, `""`) + `"` }
func ql(s string) string { return `'` + strings.ReplaceAll(s, `'`, `''`) + `'` }

var seqDefault = regexp.MustCompile(`^nextval\('("?[^']+?"?)'::regclass\)$`)

func colDef(c Column) string {
	s := qi(c.Name) + " " + c.Type
	if c.Default != "" {
		s += " DEFAULT " + c.Default
	}
	if c.NotNull {
		s += " NOT NULL"
	}
	return s
}

// Diff want = الموديلات، have = القاعدة الحية (Introspect).
func Diff(want []Table, wantEnums []Enum, have []Table, haveEnums []Enum) Plan {
	return DiffAll(want, wantEnums, nil, have, haveEnums, nil)
}

// DiffAll مثل Diff ويا الأشياء غير الجداول (wantObj/haveObj = nil يعني تجاهلها).
func DiffAll(want []Table, wantEnums []Enum, wantObj []Object, have []Table, haveEnums []Enum, haveObj []Object) Plan {
	var p Plan
	objMissing := map[string][]string{}
	if wantObj != nil || haveObj != nil {
		ho := map[string]string{}
		for _, o := range haveObj {
			ho[o.Kind+":"+o.Key] = o.Def
		}
		wo := map[string]bool{}
		for _, o := range wantObj {
			k := o.Kind + ":" + o.Key
			wo[k] = true
			d, ok := ho[k]
			switch {
			case !ok:
				objMissing[o.Kind] = append(objMissing[o.Kind], o.Def)
			case strings.TrimSpace(d) != strings.TrimSpace(o.Def):
				p.NeedsApproval = append(p.NeedsApproval, fmt.Sprintf("%s %s تغيّر بالقاعدة عن الموديل", o.Kind, o.Key))
			}
		}
		for _, o := range haveObj {
			if !wo[o.Kind+":"+o.Key] {
				p.NeedsApproval = append(p.NeedsApproval, fmt.Sprintf("%s %s بالقاعدة ومو بالموديلات", o.Kind, o.Key))
			}
		}
	}
	p.Safe = append(p.Safe, objMissing["ext"]...)
	he := map[string]Enum{}
	for _, e := range haveEnums {
		he[e.Name] = e
	}
	we := map[string]bool{}
	for _, e := range wantEnums {
		we[e.Name] = true
		h, ok := he[e.Name]
		if !ok {
			vals := make([]string, len(e.Values))
			for i, v := range e.Values {
				vals[i] = ql(v)
			}
			p.Safe = append(p.Safe, fmt.Sprintf(`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = %s) THEN CREATE TYPE %s AS ENUM (%s); END IF; END $$`, ql(e.Name), qi(e.Name), strings.Join(vals, ", ")))
			continue
		}
		hv := map[string]bool{}
		for _, v := range h.Values {
			hv[v] = true
		}
		wv := map[string]bool{}
		for _, v := range e.Values {
			wv[v] = true
			if !hv[v] {
				p.Safe = append(p.Safe, fmt.Sprintf(`ALTER TYPE %s ADD VALUE IF NOT EXISTS %s`, qi(e.Name), ql(v)))
			}
		}
		for _, v := range h.Values {
			if !wv[v] {
				p.NeedsApproval = append(p.NeedsApproval, fmt.Sprintf("النوع %s: القيمة %s بالقاعدة ومو بالموديل", e.Name, v))
			}
		}
	}
	for _, e := range haveEnums {
		if !we[e.Name] {
			p.NeedsApproval = append(p.NeedsApproval, fmt.Sprintf("النوع %s بالقاعدة ومو بالموديلات", e.Name))
		}
	}

	p.Safe = append(p.Safe, objMissing["seq"]...)
	ht := map[string]Table{}
	for _, t := range have {
		ht[t.Name] = t
	}
	wt := map[string]bool{}
	var post []string // القيود والفهارس بعد إنشاء كل الجداول (المفاتيح الأجنبية تأشّر على جداول ثانية)
	for _, t := range want {
		wt[t.Name] = true
		h, ok := ht[t.Name]
		if !ok {
			defs := []string{}
			for _, c := range t.Columns {
				if m := seqDefault.FindStringSubmatch(c.Default); m != nil {
					p.Safe = append(p.Safe, "CREATE SEQUENCE IF NOT EXISTS "+m[1])
					post = append(post, fmt.Sprintf("ALTER SEQUENCE %s OWNED BY %s.%s", m[1], qi(t.Name), qi(c.Name)))
				}
				defs = append(defs, colDef(c))
			}
			p.Safe = append(p.Safe, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n\t%s\n)", qi(t.Name), strings.Join(defs, ",\n\t")))
			for _, k := range t.Constraints {
				post = append(post, fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s", qi(t.Name), qi(k.Name), k.Def))
			}
			for _, x := range t.Indexes {
				post = append(post, ifNotExists(x))
			}
			continue
		}
		hc := map[string]Column{}
		for _, c := range h.Columns {
			hc[c.Name] = c
		}
		wc := map[string]bool{}
		for _, c := range t.Columns {
			wc[c.Name] = true
			x, ok := hc[c.Name]
			if !ok {
				if m := seqDefault.FindStringSubmatch(c.Default); m != nil {
					p.Safe = append(p.Safe, "CREATE SEQUENCE IF NOT EXISTS "+m[1])
					post = append(post, fmt.Sprintf("ALTER SEQUENCE %s OWNED BY %s.%s", m[1], qi(t.Name), qi(c.Name)))
				}
				p.Safe = append(p.Safe, fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s", qi(t.Name), colDef(c)))
				if c.NotNull && c.Default == "" {
					p.NeedsApproval = append(p.NeedsApproval, fmt.Sprintf("%s.%s: NOT NULL بلا قيمة افتراضية — يفشل إذا الجدول بيه صفوف", t.Name, c.Name))
				}
				continue
			}
			if x.Type != c.Type || x.NotNull != c.NotNull || x.Default != c.Default {
				p.NeedsApproval = append(p.NeedsApproval, fmt.Sprintf("%s.%s تغيّر: القاعدة (%s، notnull=%v، default=%q) ← الموديل (%s، notnull=%v، default=%q)",
					t.Name, c.Name, x.Type, x.NotNull, x.Default, c.Type, c.NotNull, c.Default))
			}
		}
		for _, c := range h.Columns {
			if !wc[c.Name] {
				p.NeedsApproval = append(p.NeedsApproval, fmt.Sprintf("%s.%s بالقاعدة ومو بالموديل (حذفه يحتاج موافقة)", t.Name, c.Name))
			}
		}
		hk := map[string]string{}
		for _, k := range h.Constraints {
			hk[k.Name] = k.Def
		}
		wk := map[string]bool{}
		for _, k := range t.Constraints {
			wk[k.Name] = true
			d, ok := hk[k.Name]
			switch {
			case !ok:
				post = append(post, fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s", qi(t.Name), qi(k.Name), k.Def))
			case d != k.Def:
				p.NeedsApproval = append(p.NeedsApproval, fmt.Sprintf("%s: القيد %s تغيّر (%s ← %s)", t.Name, k.Name, d, k.Def))
			}
		}
		for _, k := range h.Constraints {
			if !wk[k.Name] {
				p.NeedsApproval = append(p.NeedsApproval, fmt.Sprintf("%s: القيد %s بالقاعدة ومو بالموديل", t.Name, k.Name))
			}
		}
		hx := map[string]bool{}
		for _, x := range h.Indexes {
			hx[x] = true
		}
		wx := map[string]bool{}
		for _, x := range t.Indexes {
			wx[x] = true
			if !hx[x] {
				post = append(post, ifNotExists(x))
			}
		}
		for _, x := range h.Indexes {
			if !wx[x] {
				p.NeedsApproval = append(p.NeedsApproval, fmt.Sprintf("%s: فهرس بالقاعدة ومو بالموديل: %s", t.Name, x))
			}
		}
	}
	for _, t := range have {
		if !wt[t.Name] {
			p.NeedsApproval = append(p.NeedsApproval, fmt.Sprintf("الجدول %s بالقاعدة ومو بالموديلات (حذفه يحتاج موافقة)", t.Name))
		}
	}
	// الدوال والـviews بعد الجداول (بعض الدوال تقرا جداول) وقبل الفهارس (فهارس تستعمل دوال)
	if len(objMissing["fn"]) > 0 {
		// الدوال تعتمد على بعض — فحص الجسم يتأجل لحد ما تنخلق كلها
		p.Safe = append(p.Safe, "SET check_function_bodies = off")
		p.Safe = append(p.Safe, objMissing["fn"]...)
		p.Safe = append(p.Safe, "RESET check_function_bodies")
	}
	p.Safe = append(p.Safe, objMissing["view"]...)
	// الترتيب: المفاتيح الأساسية والفريدة ← الفهارس ← الشروط ← الأجنبية (تأشّر على مفاتيح لازم تكون موجودة)
	rank := func(s string) int {
		switch {
		case strings.Contains(s, " PRIMARY KEY") || strings.Contains(s, " UNIQUE "):
			return 0
		case strings.HasPrefix(s, "CREATE "):
			return 1
		case strings.Contains(s, " FOREIGN KEY"):
			return 3
		}
		return 2
	}
	sort.SliceStable(post, func(i, j int) bool { return rank(post[i]) < rank(post[j]) })
	p.Safe = append(p.Safe, post...)
	p.Safe = append(p.Safe, objMissing["trg"]...)
	sort.Strings(p.NeedsApproval)
	return p
}

var createIdx = regexp.MustCompile(`^CREATE (UNIQUE )?INDEX (IF NOT EXISTS )?`)

func ifNotExists(def string) string {
	return createIdx.ReplaceAllStringFunc(def, func(m string) string {
		if strings.Contains(m, "UNIQUE") {
			return "CREATE UNIQUE INDEX IF NOT EXISTS "
		}
		return "CREATE INDEX IF NOT EXISTS "
	})
}
