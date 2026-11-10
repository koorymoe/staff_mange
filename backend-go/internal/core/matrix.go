package core

// MatrixRow سطر بمصفوفة الصلاحيات: فعل وقاعدته ومساراته — للتدقيق (دازاي:
// «هاي الأشياء مهمة جداً، إذا نريد ندقق بيهن»).
type MatrixRow struct {
	Action    Action   `json:"action"`
	Label     string   `json:"label"`
	Public    bool     `json:"public,omitempty"`
	OwnerOnly bool     `json:"ownerOnly,omitempty"`
	Strict    bool     `json:"strict,omitempty"`
	Anyone    bool     `json:"anyone,omitempty"`
	Roles     []string `json:"roles,omitempty"`
	Perms     []string `json:"perms,omitempty"`
	HasCheck  bool     `json:"hasCheck,omitempty"`
	Routes    []string `json:"routes"`
	Unused    bool     `json:"unused,omitempty"` // فعل بلا مسار — يا ينحذف يا ناقص مسار
}

// Matrix كل الأفعال ويا مساراتها.
func (rt *Router) Matrix() []MatrixRow {
	byAct := map[Action][]string{}
	for _, ri := range rt.routes {
		byAct[ri.Action] = append(byAct[ri.Action], ri.Pattern)
	}
	out := []MatrixRow{}
	for _, act := range rt.policies.Actions() {
		rule, _ := rt.policies.Rule(act)
		out = append(out, MatrixRow{Action: act, Label: rule.Label, Public: rule.Public, OwnerOnly: rule.OwnerOnly, Strict: rule.Strict,
			Anyone: rule.Anyone, Roles: rule.Roles, Perms: rule.Perms, HasCheck: rule.Check != nil,
			Routes: byAct[act], Unused: len(byAct[act]) == 0})
	}
	return out
}

// Simulate «هذا الشخص يگدر يسوي هذا الفعل؟» بلا طلب حقيقي — للاختبارات ولشاشة
// التدقيق. الشرط الإضافي (Check) يُعتبر متحقق: المحاكاة تفحص الدور والصلاحية.
func (p *Policies) Simulate(a *Actor, act Action) bool {
	rule, ok := p.Rule(act)
	return ok && decide(rule, a, nil, true)
}
