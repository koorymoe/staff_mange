package service

import "context"

// ═══ «السكرتير» — الأنواع المشتركة (العقد: docs/secretary.md) ═══
// ناروتو: التخزين والمسارات والأدوات. لوفي: العقل (SecretaryBrain.Reply).
// ساسكي: الواجهة — تقرا نفس الأشكال بالـJSON. أي تغيير هنا يتغيّر بالعقد أول.

// SecretaryCard بطاقة جواب — مو كتلة كلام.
type SecretaryCard struct {
	Kind      string   `json:"kind"` // booking | project | customer | employee | attendance | task | summary | info
	Title     string   `json:"title"`
	Status    string   `json:"status,omitempty"`
	Owner     *string  `json:"owner"`
	UpdatedAt *string  `json:"updatedAt"`
	Lines     []string `json:"lines"`
	Link      *string  `json:"link"`
}

// SecretaryTaskFields حقول مسودة المهمة — المستخدم يعدّلها قبل «إرسال».
type SecretaryTaskFields struct {
	EmployeeID   string `json:"employeeId"`
	EmployeeName string `json:"employeeName"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Priority     string `json:"priority"` // NORMAL | URGENT
	DueAt        string `json:"dueAt"`    // YYYY-MM-DDTHH:MM بتوقيت بغداد، أو ""
}

// SecretaryCandidate موظف مطابق للاسم — إذا أكثر من واحد، المستخدم يختار.
type SecretaryCandidate struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SecretaryProposal مقترح تنفيذ — ما ينفّذ إلا بـ«إرسال».
type SecretaryProposal struct {
	ActionID   string               `json:"actionId"`
	Kind       string               `json:"kind"` // task
	Summary    string               `json:"summary"`
	Fields     SecretaryTaskFields  `json:"fields"`
	Candidates []SecretaryCandidate `json:"candidates"`
}

// SecretaryHistory رسالة سابقة يقراها العقل حتى «وهذا الحجز؟» تنفهم.
type SecretaryHistory struct {
	Role string // user | assistant
	Text string
}

// SecretaryTurn دور واحد: ناروتو يجهّزه ويسلّمه للعقل.
type SecretaryTurn struct {
	UserID   string
	UserName string
	// UserTitle لقبه من ملفه («دكتور»، «أستاذ»، «مهندس»، أو «سيدنا») — دازاي 10-23.
	UserTitle string
	// UserRole «المالك» أو «المدير» — حتى يعرف منو يحاچي.
	UserRole string
	Text     string
	History  []SecretaryHistory
	Prefs    []string
}

// SecretaryResult جواب العقل.
type SecretaryResult struct {
	Reply    string
	Cards    []SecretaryCard
	Proposal *SecretaryProposal
}

// SecretaryReplier واجهة العقل — لوفي يكتب SecretaryBrain بيها. لحد ما يوصل،
// المسار يرد برسالة واضحة بدل ما يطيح.
type SecretaryReplier interface {
	Reply(ctx context.Context, t SecretaryTurn) (*SecretaryResult, error)
}

// SecretaryTurnCtx يجمع شنو طلّعت الأدوات بهالدور: بطاقات ومقترح. الأدوات
// ترجّع نص للنموذج، وبنفس الوكت تحط البطاقات هنا حتى الواجهة تستلمها
// منظّمة — الأرقام من النظام مباشرة، مو من صياغة النموذج.
type SecretaryTurnCtx struct {
	UserID   string
	Request  string // نص الطلب — ينحفظ ويا المقترح بالسجل
	Cards    []SecretaryCard
	Proposal *SecretaryProposal
}

func (c *SecretaryTurnCtx) addCard(card SecretaryCard) {
	if card.Lines == nil {
		card.Lines = []string{}
	}
	if len(c.Cards) < 12 {
		c.Cards = append(c.Cards, card)
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
