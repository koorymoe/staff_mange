package core

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
)

// Action شنو يسوي المسار — «app.verb» مثل "bookings.view" أو "worktypes.edit".
type Action string

// Rule قاعدة السماح لفعل واحد. الترتيب:
//  1. Public: بلا دخول (نادر — لازم يكون مقصود).
//  2. OwnerOnly: المالك الأصلي بس.
//  3. المالك ومدير النظام يعبرون (إلا OwnerOnly).
//  4. Anyone: أي موظف داخل.
//  5. الدور (أصلي/فعلي/منصب) أو الصلاحية.
//  6. Check: شرط إضافي على الشي نفسه (مثلاً «طرف بالحجز») — إذا Roles/Perms
//     فاضية، Check وحده يقرر.
type Rule struct {
	Label     string // بالعربي — يطلع بمصفوفة الصلاحيات
	Public    bool
	OwnerOnly bool
	Anyone    bool
	Roles     []string
	Perms     []string
	Check     func(a *Actor, r *http.Request) bool
}

// Policies سجل القواعد — مكان واحد لكل النظام.
type Policies struct {
	mu    sync.RWMutex
	rules map[Action]Rule
}

func NewPolicies() *Policies { return &Policies{rules: map[Action]Rule{}} }

// Register يسجّل قاعدة — التكرار خطأ برمجي (تطبيقين يعرّفون نفس الفعل).
func (p *Policies) Register(act Action, rule Rule) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, dup := p.rules[act]; dup {
		panic(fmt.Sprintf("core: الفعل %q متعرّف مرتين", act))
	}
	if rule.Label == "" {
		panic(fmt.Sprintf("core: الفعل %q بلا وصف عربي", act))
	}
	if !rule.Public && !rule.OwnerOnly && !rule.Anyone && len(rule.Roles) == 0 && len(rule.Perms) == 0 && rule.Check == nil {
		panic(fmt.Sprintf("core: الفعل %q بلا قاعدة — ما يفتح لأحد غير المدير، اكتبها صريحة", act))
	}
	p.rules[act] = rule
}

func (p *Policies) Rule(act Action) (Rule, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	r, ok := p.rules[act]
	return r, ok
}

// Allowed القرار. فعل مو مسجّل = لا.
func (p *Policies) Allowed(a *Actor, act Action, r *http.Request) bool {
	rule, ok := p.Rule(act)
	return ok && decide(rule, a, r, false)
}

// decide القرار الصافي. assumeCheck: الشرط الإضافي يُعتبر متحقق (للمحاكاة).
func decide(rule Rule, a *Actor, r *http.Request, assumeCheck bool) bool {
	if rule.Public {
		return true
	}
	if a == nil {
		return false
	}
	if rule.OwnerOnly {
		return a.Owner()
	}
	if a.Admin() {
		return true
	}
	onlyCheck := len(rule.Roles) == 0 && len(rule.Perms) == 0 && !rule.Anyone
	byRole := rule.Anyone || a.HasRole(rule.Roles...) || a.Can(rule.Perms...)
	if rule.Check == nil {
		return byRole
	}
	checked := assumeCheck || rule.Check(a, r)
	if onlyCheck {
		return checked
	}
	return byRole && checked
}

// Actions كل الأفعال مرتبة — للمصفوفة والوثائق.
func (p *Policies) Actions() []Action {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]Action, 0, len(p.rules))
	for a := range p.rules {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
