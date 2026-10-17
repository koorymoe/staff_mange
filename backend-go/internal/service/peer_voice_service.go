package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"staffmange-api/internal/repository"
)

// ═══ صوت الموظفين — الفضفضة الأسبوعية + تعمّق ماتركس + المشاكل الوظيفية ═══
//
// قرارات (ع) 10-05:
//   • كل أسبوع ماتركس يسأل الموظف شلونه، وشي مضايقه، وشلون زملاؤه وياه.
//   • إذا قيّم زميله سيء، ماتركس **ما يرفع شي فوراً**: يسأله شنو صار، وليش،
//     وشلون يريد ينحل، ويقترح عليه حل — وبعدين التقرير يروح للمراقب والمدير.
//   • النص الأصلي يشوفه المالك والمدير. المراقب يشوف تقارير التعمّق.
//     **والزميل المقصود ما يعرف أبداً** — ماكو مسار يرجّعله شي.
//   • النزاعات: ماتركس يسجّل ويقترح، والمراقب والمدير يقررون، وماتركس يتابع.
//   • هايكو يشوف النص **بعد شيل الأسماء** (redactNames). ماكو عقوبة تلقائية.

type PeerVoiceService struct {
	repo   *repository.PeerVoiceRepository
	names  func() []string
	notify func(roles []string, perm, msg string)
	client *anthropic.Client
	model  string
}

func NewPeerVoiceService(repo *repository.PeerVoiceRepository, names func() []string, notify func([]string, string, string)) *PeerVoiceService {
	return &PeerVoiceService{repo: repo, names: names, notify: notify}
}

func (s *PeerVoiceService) EnableModel(apiKey, modelName string) {
	if apiKey == "" {
		return
	}
	if modelName == "" {
		modelName = "claude-haiku-4-5"
	}
	c := anthropic.NewClient(option.WithAPIKey(apiKey))
	s.client, s.model = &c, modelName
}

// PeerWeek بداية أسبوع الفضفضة (السبت، بتوقيت بغداد — أسبوع الدوام العراقي).
func PeerWeek(now time.Time) time.Time {
	t := now.In(debriefLoc)
	back := (int(t.Weekday()) - int(time.Saturday) + 7) % 7
	d := t.AddDate(0, 0, -back)
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
}

// peerOpen ماتركس يسأل من الخميس (وتبقى مفتوحة لحد آخر الأسبوع).
func peerOpen(now time.Time) bool {
	w := now.In(debriefLoc).Weekday()
	return w == time.Thursday || w == time.Friday
}

// ── كلمات الخطر: تهديد/ضرب/تحرش… ← «عاجل» وإشعار فوري ──
var dangerWords = []string{"تهديد", "يهددني", "هددني", "ضربني", "ضرب", "عركة", "تحرش", "يتحرش", "اعتداء", "سلاح", "يسب", "سبني", "شتم", "شتمني", "قتل", "طعن"}

func peerSeverity(texts ...string) string {
	all := strings.Join(texts, " ")
	for _, w := range dangerWords {
		if strings.Contains(all, w) {
			return "URGENT"
		}
	}
	return "ATTENTION"
}

func strp(s string) *string {
	t := strings.TrimSpace(s)
	if t == "" {
		return nil
	}
	return &t
}

func sval(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// ═══ الموظف ═══

type PeerNoteIn struct {
	TargetID string `json:"targetId"`
	Score    int    `json:"score"`
	Kind     string `json:"kind"`
	Text     string `json:"text"`
}

type PeerCheckinIn struct {
	Mood       *int         `json:"mood"`
	Worry      string       `json:"worry"`
	Suggestion string       `json:"suggestion"`
	Nothing    bool         `json:"nothing"`
	Notes      []PeerNoteIn `json:"notes"`
}

type MyPeerState struct {
	Open       bool                        `json:"open"`
	Week       time.Time                   `json:"week"`
	Checkin    *repository.PeerCheckin     `json:"checkin"`
	Notes      []repository.PeerNote       `json:"notes"`
	Colleagues []repository.Colleague      `json:"colleagues"`
	Followups  []repository.WorkplaceIssue `json:"followups"`
}

func (s *PeerVoiceService) Mine(me string) (*MyPeerState, error) {
	now := time.Now()
	week := PeerWeek(now)
	st := &MyPeerState{Open: peerOpen(now), Week: week, Notes: []repository.PeerNote{}}
	if c, err := s.repo.Checkin(me, week); err == nil {
		st.Checkin = c
		st.Notes, _ = s.repo.NotesOfCheckin(c.ID)
		st.Open = true // بدا يكتب — تبقى مفتوحة يكمّل
	}
	st.Colleagues, _ = s.repo.Colleagues(me)
	st.Followups, _ = s.repo.FollowupsFor(me)
	if st.Followups == nil {
		st.Followups = []repository.WorkplaceIssue{}
	}
	return st, nil
}

// Submit يحفظ فضفضة الأسبوع. الملاحظات السيئة ترجع «تحتاج تعمّق».
func (s *PeerVoiceService) Submit(me string, in PeerCheckinIn) (*MyPeerState, error) {
	week := PeerWeek(time.Now())
	if in.Mood != nil && (*in.Mood < 1 || *in.Mood > 5) {
		return nil, errors.New("المزاج من ١ لـ٥")
	}
	urgent := peerSeverity(in.Worry, in.Suggestion) == "URGENT"
	id, err := s.repo.UpsertCheckin(me, week, in.Mood, strp(in.Worry), strp(in.Suggestion), in.Nothing, urgent)
	if err != nil {
		return nil, err
	}
	s.repo.DropDraftNotes(id)
	seen := map[string]bool{}
	for _, n := range in.Notes {
		if n.TargetID == "" || n.TargetID == me || seen[n.TargetID] || !s.repo.ActiveEmployee(n.TargetID) {
			continue
		}
		if n.Score < 1 || n.Score > 5 {
			return nil, errors.New("التقييم من ١ لـ٥ نجوم")
		}
		kind := n.Kind
		if kind != "GOOD" && kind != "PROBLEM" {
			kind = "NOTE"
		}
		seen[n.TargetID] = true
		needs := n.Score <= 2 || kind == "PROBLEM"
		sev := "NORMAL"
		if needs {
			sev = peerSeverity(n.Text)
		}
		if _, err := s.repo.AddNote(id, me, n.TargetID, n.Score, kind, strp(n.Text), needs, sev); err != nil {
			return nil, err
		}
		if sev == "URGENT" {
			urgent = true
		}
	}
	if urgent && s.notify != nil {
		go s.notify([]string{"ADMIN", "MONITOR"}, "monitoring",
			"🚨 ماتركس — بفضفضة هالأسبوع أكو كلام عن تهديد أو اعتداء أو إساءة. شوفه فوراً بـ«صوت الموظفين».")
	}
	return s.Mine(me)
}

// Followup الموظف يجاوب أسئلة ماتركس الثلاث، وماتركس يقترح حل.
func (s *PeerVoiceService) Followup(me, noteID, what, why, wish string) (*repository.PeerNote, error) {
	n, err := s.repo.Note(noteID)
	if err != nil || n.AuthorID != me {
		return nil, errors.New("هالملاحظة مو إلك")
	}
	if strings.TrimSpace(what) == "" {
		return nil, errors.New("گول شنو صار حتى أگدر أساعدك")
	}
	sev := peerSeverity(sval(n.Text), what, why, wish)
	sugg := s.suggest(sval(n.Text), what, why, wish)
	if err := s.repo.SaveFollowup(noteID, strp(what), strp(why), strp(wish), strp(sugg), sev); err != nil {
		return nil, err
	}
	if sev == "URGENT" && s.notify != nil {
		go s.notify([]string{"ADMIN", "MONITOR"}, "monitoring",
			"🚨 ماتركس — موظف حچى عن تهديد أو اعتداء ويا زميله. التفاصيل بـ«صوت الموظفين» الحين.")
	}
	return s.repo.Note(noteID)
}

// Accept الموظف يگول الاقتراح يناسبه لو لا — وبعدها التقرير يروح للمراقب والمدير.
func (s *PeerVoiceService) Accept(me, noteID string, accepted bool, note string) (*repository.PeerNote, error) {
	n, err := s.repo.Note(noteID)
	if err != nil || n.AuthorID != me {
		return nil, errors.New("هالملاحظة مو إلك")
	}
	if err := s.repo.Accept(noteID, accepted, strp(note)); err != nil {
		return nil, err
	}
	if s.notify != nil {
		ok := "وافق على اقتراح ماتركس"
		if !accepted {
			ok = "ما وافق على اقتراح ماتركس ويريد تدخّل"
		}
		go s.notify([]string{"ADMIN", "MONITOR"}, "monitoring",
			fmt.Sprintf("🤖 ماتركس — تقرير جديد من الفضفضة: %s عنده مشكلة ويا زميل، و%s. التفاصيل بـ«صوت الموظفين».", n.AuthorName, ok))
	}
	return s.repo.Note(noteID)
}

const peerSystemPrompt = `أنت «ماتركس» بنظام شركة الأماني (العراق). موظف حچالك عن مشكلة ويا زميله بالشغل.
الأسماء مشالة ومكانها [موظف]. اكتب اقتراح حل عملي وهادي بالعراقي البسيط، ٣ جمل كحد أقصى:
- لا تحكم منو الغلطان، ولا تقترح عقوبة أو خصم.
- اقترح خطوة يگدر يسويها هو، وخطوة يسويها المراقب أو المدير إذا احتاج.
- إذا أكو تهديد أو اعتداء، گول بوضوح إن المدير لازم يتدخل فوراً.
بلا مقدمات وبلا علامات تنصيص.`

// suggest اقتراح الحل: هايكو على نص منظّف، أو قواعد إذا المفتاح مطفي أو ماكو مفتاح.
func (s *PeerVoiceService) suggest(texts ...string) string {
	all := strings.Join(texts, "\n")
	if s.client != nil && AIFeatureOn("PEER") {
		clean := redactNames(all, s.names(), "")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
			Model: anthropic.Model(s.model), MaxTokens: 300,
			System:   []anthropic.TextBlockParam{{Text: peerSystemPrompt}},
			Messages: []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(clean))},
		})
		if err == nil {
			RecordAIUsage("PEER", s.model, resp.Usage)
			var b strings.Builder
			for _, block := range resp.Content {
				if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
					b.WriteString(tb.Text)
				}
			}
			if out := strings.TrimSpace(b.String()); out != "" {
				return out
			}
		}
	}
	return ruleSuggestion(all)
}

// ruleIssueSuggestion اقتراح للمراقب والمدير (مو للموظف).
func ruleIssueSuggestion(t string) string {
	switch {
	case peerSeverity(t) == "URGENT":
		return "🚨 يحتاج تدخّل فوري: المدير يسمع كل طرف لحاله اليوم، ويفصل بينهم بالشغل لحد ما ينحسم الموضوع، ويسجّل القرار هنا."
	case strings.Contains(t, "عدّة") || strings.Contains(t, "عدة") || strings.Contains(t, "أداة"):
		return "حدّدوا عدّة كل واحد باسمه بالجرد، وأي استعارة تنسجّل. وجلسة قصيرة بين الطرفين بحضور المراقب."
	case strings.Contains(t, "يتأخر") || strings.Contains(t, "تأخير"):
		return "المراقب يحدد مسؤولية كل واحد بالحجوزات المشتركة بالاسم، ويتابع الالتزام بالوقت أسبوعين."
	default:
		return "اسمعوا إفادة الطرفين كل واحد لحاله، وبعدها جلسة مشتركة قصيرة يتفقون بيها على طريقة شغل، وماتركس يتابع وياهم بعد أسبوع."
	}
}

func ruleSuggestion(t string) string {
	switch {
	case peerSeverity(t) == "URGENT":
		return "هالموضوع خطير وما ينسكت عليه. المدير لازم يتدخل فوراً ويسمع الطرفين، وإنت لا ترد على الإساءة بإساءة."
	case strings.Contains(t, "يتأخر") || strings.Contains(t, "تأخير") || strings.Contains(t, "ما يجي") || strings.Contains(t, "يغيب"):
		return "اتفق وياه على وقت واضح للشغل المشترك. وإذا تكرر، المراقب يحدد مسؤولية كل واحد بالحجز حتى ما يتحمّل واحد شغل الثاني."
	case strings.Contains(t, "ما يساعد") || strings.Contains(t, "ما يشتغل") || strings.Contains(t, "كسلان") || strings.Contains(t, "يتركني"):
		return "اطلب من الليدر أو المراقب يقسّم الشغل بينكم بالاسم قبل الطلعة، حتى كل واحد يعرف شغله ويتحاسب عليه."
	case strings.Contains(t, "يصيح") || strings.Contains(t, "احترام") || strings.Contains(t, "يستهزئ") || strings.Contains(t, "اسلوب"):
		return "حاول تحچي وياه بهدوء وبعيد عن الناس. وإذا ما تحسّن، المراقب يجمعكم بجلسة قصيرة يوضّح بيها طريقة التعامل المطلوبة."
	default:
		return "جلسة قصيرة بينكم بحضور المراقب حتى كل واحد يگول اللي بقلبه وتتفقون على طريقة شغل. ماتركس راح يسألك بعد أسبوع شلون صارت الأمور."
	}
}

// ═══ المدير والمراقب ═══

type PeerVoiceReport struct {
	Full     bool                    `json:"full"` // المدير/المالك: النصوص كاملة
	Mood     []repository.MoodWeek   `json:"mood"`
	Scores   []repository.PeerScore  `json:"scores"`
	Tensions []repository.Tension    `json:"tensions"`
	Notes    []repository.PeerNote   `json:"notes"`
	Checkins []repository.CheckinRow `json:"checkins"`
	Insights []string                `json:"insights"`
}

func (s *PeerVoiceService) Report(full bool) (*PeerVoiceReport, error) {
	rep := &PeerVoiceReport{Full: full, Insights: []string{}}
	rep.Mood, _ = s.repo.MoodByWeek(8)
	rep.Scores, _ = s.repo.Scores()
	rep.Tensions, _ = s.repo.Tensions()
	var err error
	rep.Notes, err = s.repo.Notes(60, !full)
	if err != nil {
		return nil, err
	}
	if full {
		rep.Checkins, _ = s.repo.Checkins(60)
	} else {
		rep.Checkins = []repository.CheckinRow{}
	}
	for _, x := range []*[]repository.MoodWeek{&rep.Mood} {
		if *x == nil {
			*x = []repository.MoodWeek{}
		}
	}
	if rep.Scores == nil {
		rep.Scores = []repository.PeerScore{}
	}
	if rep.Tensions == nil {
		rep.Tensions = []repository.Tension{}
	}
	// تحليل ماتركس (قواعد).
	if n := len(rep.Mood); n > 0 && rep.Mood[n-1].Avg != nil {
		line := fmt.Sprintf("مزاج الفريق هالأسبوع %.1f من 5 (%d موظف جاوب)", *rep.Mood[n-1].Avg, rep.Mood[n-1].Count)
		if n > 1 && rep.Mood[n-2].Avg != nil {
			d := *rep.Mood[n-1].Avg - *rep.Mood[n-2].Avg
			if math.Abs(d) >= 0.3 {
				dir := "أحسن"
				if d < 0 {
					dir = "أسوأ"
				}
				line += fmt.Sprintf("، %s من الأسبوع الفات (%.1f)", dir, *rep.Mood[n-2].Avg)
			}
		}
		rep.Insights = append(rep.Insights, line+".")
	} else {
		rep.Insights = append(rep.Insights, "بعد ماكو فضفضات هالفترة.")
	}
	if len(rep.Tensions) > 0 {
		rep.Insights = append(rep.Insights, fmt.Sprintf("⚠️ أكو %d توتر متبادل بين موظفين (كل واحد مقيّم الثاني سيء). يستاهل يتفتح إلها مشكلة وظيفية.", len(rep.Tensions)))
	}
	low := []string{}
	for _, sc := range rep.Scores {
		if sc.Count >= 3 && sc.Avg <= 2.5 {
			low = append(low, sc.Name)
		}
	}
	if len(low) > 0 {
		rep.Insights = append(rep.Insights, "زملاؤهم مو راضين عن: "+strings.Join(low, "، ")+".")
	}
	urgent := 0
	for _, n := range rep.Notes {
		if n.Severity == "URGENT" {
			urgent++
		}
	}
	if urgent > 0 {
		rep.Insights = append(rep.Insights, fmt.Sprintf("🚨 %d ملاحظة بيها كلام عن تهديد أو اعتداء — تحتاج تدخّل.", urgent))
	}
	return rep, nil
}

// ═══ المشاكل الوظيفية ═══

type IssueDetail struct {
	repository.WorkplaceIssue
	Entries []repository.IssueEntry `json:"entries"`
	Context repository.IssueContext `json:"context"`
}

func (s *PeerVoiceService) Issues() ([]repository.WorkplaceIssue, error) { return s.repo.Issues() }

func (s *PeerVoiceService) Issue(id string) (*IssueDetail, error) {
	w, err := s.repo.Issue(id)
	if err != nil {
		return nil, errors.New("المشكلة مو موجودة")
	}
	d := &IssueDetail{WorkplaceIssue: *w, Context: s.repo.Context(w.PartyAID, w.PartyBID)}
	d.Entries, _ = s.repo.Entries(id)
	if d.Entries == nil {
		d.Entries = []repository.IssueEntry{}
	}
	return d, nil
}

func (s *PeerVoiceService) OpenIssue(by, title, a string, b *string, desc, severity string, source *string) (string, error) {
	if strings.TrimSpace(title) == "" || strings.TrimSpace(desc) == "" || a == "" {
		return "", errors.New("العنوان والوصف والطرف الأول مطلوبين")
	}
	if b != nil && *b == a {
		return "", errors.New("الطرفين نفس الشخص")
	}
	if severity != "NORMAL" && severity != "URGENT" {
		severity = "ATTENTION"
	}
	if peerSeverity(desc) == "URGENT" {
		severity = "URGENT"
	}
	id, err := s.repo.CreateIssue(strings.TrimSpace(title), a, b, by, strings.TrimSpace(desc), severity, source)
	if err != nil {
		return "", err
	}
	_ = s.Analyze(id)
	return id, nil
}

func (s *PeerVoiceService) AddEntry(issueID, by, kind, text string) error {
	if strings.TrimSpace(text) == "" {
		return errors.New("اكتب الإفادة")
	}
	if kind != "STATEMENT" {
		kind = "NOTE"
	}
	if err := s.repo.AddEntry(issueID, &by, kind, strings.TrimSpace(text)); err != nil {
		return err
	}
	// كل إفادة جديدة تحدّث تحليل ماتركس.
	return s.Analyze(issueID)
}

const issueSystemPrompt = `أنت «ماتركس» بنظام شركة الأماني (العراق). عندك مشكلة وظيفية بين موظفين (الطرف أ والطرف ب)، والأسماء مشالة.
اكتب للمراقب والمدير بالعراقي البسيط وبدقة:
١. الملخّص (جملتين).
٢. نقاط الخلاف (نقاط قصيرة).
٣. اقتراحات حل عملية (٢-٣)، بلا عقوبة تلقائية وبلا تحيّز لطرف.
إذا أكو تهديد أو اعتداء، ابدي بـ«🚨 يحتاج تدخّل فوري». بلا مقدمات.`

// Analyze ماتركس يكتب الملخّص ونقاط الخلاف والاقتراحات من الإفادات والسياق.
func (s *PeerVoiceService) Analyze(id string) error {
	d, err := s.Issue(id)
	if err != nil {
		return err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "العنوان: %s\nالوصف: %s\n", d.Title, d.Description)
	partyOf := func(aid *string) string {
		if aid == nil {
			return "المراقب/المدير"
		}
		if *aid == d.PartyAID {
			return "الطرف أ"
		}
		if d.PartyBID != nil && *aid == *d.PartyBID {
			return "الطرف ب"
		}
		return "المراقب/المدير"
	}
	for _, e := range d.Entries {
		if e.Kind == "MATRIX" {
			continue
		}
		fmt.Fprintf(&b, "- (%s) %s: %s\n", partyOf(e.AuthorID), e.Kind, e.Text)
	}
	c := d.Context
	fmt.Fprintf(&b, "سياق: اشتغلوا سوة بـ%d حجز آخر ٣ أشهر. شكاوى زبائن على الطرف أ %d وعلى الطرف ب %d.", c.SharedBookings, c.AComplaints, c.BComplaints)
	text := b.String()
	names := s.names()
	names = append(names, d.PartyAName)
	if d.PartyBName != nil {
		names = append(names, *d.PartyBName)
	}
	clean := redactNames(text, names, "")
	summary := ""
	if s.client != nil && AIFeatureOn("PEER") {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
			Model: anthropic.Model(s.model), MaxTokens: 600,
			System:   []anthropic.TextBlockParam{{Text: issueSystemPrompt}},
			Messages: []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(clean))},
		})
		if err == nil {
			RecordAIUsage("PEER", s.model, resp.Usage)
			for _, block := range resp.Content {
				if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
					summary += tb.Text
				}
			}
		}
	}
	if strings.TrimSpace(summary) == "" {
		statements := 0
		for _, e := range d.Entries {
			if e.Kind == "STATEMENT" {
				statements++
			}
		}
		summary = fmt.Sprintf("ملخّص: %s.\nالإفادات المسجّلة: %d. اشتغلوا سوة بـ%d حجز آخر ٣ أشهر.\nالاقتراح: %s",
			d.Title, statements, c.SharedBookings, ruleIssueSuggestion(text))
		if statements < 2 && d.PartyBID != nil {
			summary += "\nقبل القرار: اسمع إفادة الطرفين كل واحد لحاله وسجّلها هنا."
		}
	}
	s.repo.SetSummary(id, strings.TrimSpace(summary))
	return nil
}

func (s *PeerVoiceService) Decide(id, by, decision, status string, followUp *time.Time) error {
	if strings.TrimSpace(decision) == "" {
		return errors.New("اكتب القرار")
	}
	if status != "RESOLVED" {
		status = "IN_PROGRESS"
	}
	if err := s.repo.Decide(id, strings.TrimSpace(decision), by, status, followUp); err != nil {
		return err
	}
	return s.repo.AddEntry(id, &by, "DECISION", strings.TrimSpace(decision))
}

// PartyFollowup الطرف يجاوب «شلون صارت الأمور ويا زميلك؟» بعد موعد المتابعة.
func (s *PeerVoiceService) PartyFollowup(id, me string, better bool, text string) error {
	if !s.repo.IsParty(id, me) {
		return errors.New("مو إلك هالمتابعة")
	}
	state := "الأمور تحسّنت"
	if !better {
		state = "الأمور ما تحسّنت"
	}
	msg := state
	if t := strings.TrimSpace(text); t != "" {
		msg += ": " + t
	}
	if err := s.repo.AddEntry(id, &me, "FOLLOWUP", msg); err != nil {
		return err
	}
	if !better && s.notify != nil {
		go s.notify([]string{"ADMIN", "MONITOR"}, "monitoring",
			"🤖 ماتركس — بمتابعة مشكلة وظيفية، أحد الطرفين گال الأمور ما تحسّنت. شوف «المشاكل الوظيفية».")
	}
	return nil
}
