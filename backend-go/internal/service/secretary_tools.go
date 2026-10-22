package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/repository"
)

// ═══ «السكرتير» — الأدوات (ناروتو · docs/secretary.md) ═══
// فوق عيون ماتركس الموجودة، بالأسماء الحقيقية (استثناء (ع) 10-22 للسكرتير
// وحده). كل أداة: نص للنموذج + بطاقات منظّمة بـSecretaryTurnCtx من أرقام
// النظام مباشرة. أداة التنفيذ الوحيدة propose_task: تحضّر مسودة بس، والإرسال
// يصير بالمسار بعد «إرسال» من المستخدم — النموذج ما يكدر يرسل شي بنفسه.

// SecretaryDeps مصادر الأدوات.
type SecretaryDeps struct {
	DB     *sqlx.DB
	Client *anthropic.Client
	Model  string
	Repo   *repository.SecretaryRepository
	Matrix AgentDeps
}

// NewSecretaryAgent وكيل جديد لكل دور: الأدوات تكتب بطاقاتها ومقترحها على
// tc، فما يتشاركون حالة بين مستخدمين أو طلبين بنفس الوكت.
func NewSecretaryAgent(d SecretaryDeps, tc *SecretaryTurnCtx) *MatrixAgent {
	a := NewMatrixAgent(d.DB, d.Client, d.Model)
	a.ShowNames()
	a.RegisterReadTools(d.Matrix)
	registerSecretaryTools(a, d.Repo, tc)
	return a
}

func secFmt(t *time.Time) string {
	if t == nil {
		return "—"
	}
	return bagh(*t).Format("2006-01-02 15:04")
}

func secISO(t time.Time) *string {
	s := t.UTC().Format(time.RFC3339)
	return &s
}

func registerSecretaryTools(a *MatrixAgent, repo *repository.SecretaryRepository, tc *SecretaryTurnCtx) {
	a.AddTool(agentTool{Name: "find",
		Desc:  "دوّر بالنظام بالاسم أو الكود أو رقم الهاتف: حجوزات، مشاريع، زبائن، موظفين (٥ لكل نوع). أول أداة لأي سؤال عن شخص أو حجز أو مشروع.",
		Props: map[string]any{"q": map[string]any{"type": "string", "description": "الاسم أو الكود أو الهاتف"}},
		Req:   []string{"q"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			f, err := repo.Find(strArg(in, "q"))
			if err != nil {
				return "", err
			}
			var b strings.Builder
			for _, x := range f.Bookings {
				fmt.Fprintf(&b, "حجز %s — %s · %s · الحالة: %s · الموعد: %s · الكادر: %s\n",
					x.Code, deref(x.Customer, "—"), deref(x.Service, "—"), x.Station, secFmt(x.ScheduledAt), deref(x.Crew, "بلا كادر"))
				tc.addCard(SecretaryCard{Kind: "booking", Title: "حجز " + x.Code + " — " + deref(x.Customer, ""),
					Status: x.Station, Owner: x.Crew, UpdatedAt: secISO(x.UpdatedAt),
					Lines: []string{"الخدمة: " + deref(x.Service, "—"), "الموعد: " + secFmt(x.ScheduledAt)},
					Link:  strPtr("/bookings?focus=" + x.Code)})
			}
			for _, x := range f.Projects {
				fmt.Fprintf(&b, "مشروع %s %s — المرحلة: %s · المسؤول: %s\n", deref(x.Code, ""), x.Name, x.Stage, deref(x.Owner, "—"))
				tc.addCard(SecretaryCard{Kind: "project", Title: strings.TrimSpace(deref(x.Code, "") + " " + x.Name),
					Status: x.Stage, Owner: x.Owner, UpdatedAt: secISO(x.UpdatedAt), Lines: []string{}, Link: strPtr("/projects")})
			}
			for _, x := range f.Customers {
				fmt.Fprintf(&b, "زبون #%d %s · %s · حجوزاته: %d\n", x.Code, x.Name, deref(x.Phone, "—"), x.Bookings)
				tc.addCard(SecretaryCard{Kind: "customer", Title: x.Name,
					Lines: []string{"الهاتف: " + deref(x.Phone, "—"), fmt.Sprintf("حجوزاته: %d", x.Bookings)},
					Link:  strPtr("/customers")})
			}
			for _, x := range f.Employees {
				fmt.Fprintf(&b, "موظف %s · %s · %s\n", x.Name, x.Role, deref(x.Phone, "—"))
			}
			if b.Len() == 0 {
				return "ما لگيت شي بهالاسم أو الكود.", nil
			}
			return b.String(), nil
		}})

	a.AddTool(agentTool{Name: "booking_timeline",
		Desc:  "مراحل حجز واحد بالترتيب (انسجّل، انتصل، انثبّت، استلم، انطلق، وصل، بدأ، انجز) ومنو سوّاها ووكتها. استعملها لـ«وين وصل الحجز؟».",
		Props: map[string]any{"code": map[string]any{"type": "string", "description": "كود الحجز"}},
		Req:   []string{"code"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			bk, steps, err := repo.BookingTimeline(strArg(in, "code"))
			if err != nil {
				return "", err
			}
			if bk == nil {
				return "ما لگيت حجز بهالكود — جرّب find.", nil
			}
			var b strings.Builder
			lines := []string{}
			fmt.Fprintf(&b, "حجز %s — %s · الحالة: %s · الموعد: %s\n", bk.Code, deref(bk.Customer, "—"), bk.Station, secFmt(bk.ScheduledAt))
			for _, s := range steps {
				at := s.At
				line := fmt.Sprintf("%s: %s", s.Label, secFmt(&at))
				if s.By != nil && *s.By != "" {
					line += " (" + *s.By + ")"
				}
				lines = append(lines, line)
				b.WriteString("- " + line + "\n")
			}
			tc.addCard(SecretaryCard{Kind: "booking", Title: "حجز " + bk.Code + " — " + deref(bk.Customer, ""),
				Status: bk.Station, Owner: bk.Crew, UpdatedAt: secISO(bk.UpdatedAt), Lines: lines,
				Link: strPtr("/bookings?focus=" + bk.Code)})
			return b.String(), nil
		}})

	a.AddTool(agentTool{Name: "attendance_now",
		Desc: "دوام اليوم لحظياً: منو داخل هسه، منو طلع، منو ما سجّل حضور، ووكت دخول كل واحد (بلا المدير والمالك).",
		Run: func(_ map[string]any, _ *nameMask) (string, error) {
			rows, err := repo.AttendanceToday()
			if err != nil {
				return "", err
			}
			var in, out, absent []string
			for _, r := range rows {
				switch {
				case r.CheckIn == nil:
					absent = append(absent, r.Name)
				case r.CheckOut != nil:
					out = append(out, fmt.Sprintf("%s (%s ← %s)", r.Name, bagh(*r.CheckIn).Format("15:04"), bagh(*r.CheckOut).Format("15:04")))
				default:
					in = append(in, fmt.Sprintf("%s (من %s)", r.Name, bagh(*r.CheckIn).Format("15:04")))
				}
			}
			tc.addCard(SecretaryCard{Kind: "attendance", Title: "دوام اليوم",
				Status: fmt.Sprintf("داخل %d · طلع %d · ما حضر %d", len(in), len(out), len(absent)),
				Lines:  []string{"داخل هسه: " + joinOr(in), "ما حضر: " + joinOr(absent)}, Link: strPtr("/attendance")})
			return fmt.Sprintf("داخل هسه (%d): %s\nطلع (%d): %s\nما سجّل حضور (%d): %s",
				len(in), joinOr(in), len(out), joinOr(out), len(absent), joinOr(absent)), nil
		}})

	a.AddTool(agentTool{Name: "pending_decisions",
		Desc: "الأشياء الي تنتظر قرار المدير: طلبات حذف حجوزات، وإجازات مسارها الإدارة.",
		Run: func(_ map[string]any, _ *nameMask) (string, error) {
			p, err := repo.Pending()
			if err != nil {
				return "", err
			}
			if len(p) == 0 {
				return "ماكو شي ينتظر قرارك.", nil
			}
			var b strings.Builder
			for _, x := range p {
				fmt.Fprintf(&b, "- %s · %s\n", x.Question, x.Why)
			}
			return b.String(), nil
		}})

	a.AddTool(agentTool{Name: "today_facts",
		Desc: "أرقام اليوم الخام: حجوزات اليوم، المنجزة، بالميدان، المتأخرة، تنتظر تنسيق، ناقصها ورق، الشكاوى المفتوحة، الحاضرين، والمهام المفتوحة.",
		Run: func(_ map[string]any, _ *nameMask) (string, error) {
			t, err := repo.Today()
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("حجوزات اليوم %d · خلصت %d · بالميدان %d · متأخرة +٢٤ ساعة %d · تنتظر تنسيق %d · ناقصها ورق %d · شكاوى مفتوحة %d · حاضرين هسه %d من %d · مهام إضافية مفتوحة %d",
				t.BookingsToday, t.CompletedToday, t.InField, t.Overdue, t.NeedsCoord, t.NeedsPaper, t.OpenComplaints, t.PresentNow, t.ActiveStaff, t.OpenTasks), nil
		}})

	a.AddTool(agentTool{Name: "propose_task",
		Desc: "حضّر مسودة مهمة لموظف. ما ترسلها — المستخدم يشوفها ويضغط «إرسال». استعملها مرة وحدة بالدور، لما يطلب يوجّه شغلة لموظف.",
		Props: map[string]any{
			"employee":    map[string]any{"type": "string", "description": "اسم الموظف كما گاله المستخدم"},
			"title":       map[string]any{"type": "string", "description": "عنوان قصير (لحد ٨ كلمات)"},
			"description": map[string]any{"type": "string", "description": "شنو المطلوب بالضبط"},
			"priority":    map[string]any{"type": "string", "enum": []string{"NORMAL", "URGENT"}},
			"dueAt":       map[string]any{"type": "string", "description": "YYYY-MM-DDTHH:MM بتوقيت بغداد، أو فاضي"},
		},
		Req: []string{"employee", "title"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			if tc.Proposal != nil {
				return "اكو مسودة بهالدور أصلاً — وحدة بس بكل مرة.", nil
			}
			emps, err := repo.ActiveEmployees()
			if err != nil {
				return "", err
			}
			list := make([]VoiceTaskEmployee, 0, len(emps))
			names := map[string]string{}
			for _, e := range emps {
				list = append(list, VoiceTaskEmployee{ID: e.ID, Name: e.Name})
				names[e.ID] = e.Name
			}
			ids := matchEmployees(strArg(in, "employee"), list)
			if len(ids) == 0 {
				return "ما لگيت موظف بهالاسم — اسأله منو يقصد.", nil
			}
			f := SecretaryTaskFields{Title: strArg(in, "title"), Description: strArg(in, "description"),
				Priority: "NORMAL", DueAt: ""}
			if strArg(in, "priority") == "URGENT" {
				f.Priority = "URGENT"
			}
			if r := []rune(f.Title); len(r) > 120 {
				f.Title = string(r[:120])
			}
			if d := strArg(in, "dueAt"); d != "" {
				if _, err := time.ParseInLocation("2006-01-02T15:04", d, debriefLoc); err == nil {
					f.DueAt = d
				}
			}
			cands := []SecretaryCandidate{}
			for _, id := range ids {
				cands = append(cands, SecretaryCandidate{ID: id, Name: names[id]})
			}
			if len(ids) == 1 {
				f.EmployeeID, f.EmployeeName = ids[0], names[ids[0]]
			}
			who := f.EmployeeName
			if who == "" {
				who = "(اختار الموظف)"
			}
			urgent := ""
			if f.Priority == "URGENT" {
				urgent = "عاجلة "
			}
			summary := fmt.Sprintf("فهمت منك: مهمة %sلـ%s — %s", urgent, who, f.Title)
			if f.DueAt != "" {
				summary += " · موعدها " + strings.Replace(f.DueAt, "T", " ", 1)
			}
			id, err := repo.CreateAction(tc.UserID, "task", tc.Request, summary, f)
			if err != nil {
				return "", err
			}
			tc.Proposal = &SecretaryProposal{ActionID: id, Kind: "task", Summary: summary, Fields: f, Candidates: cands}
			if len(ids) > 1 {
				return fmt.Sprintf("حضّرت المسودة، بس الاسم يطابق %d موظفين — گله يختار قبل «إرسال».", len(ids)), nil
			}
			return "حضّرت المسودة: " + summary + ". ما انرسلت — هو يشوفها ويضغط «إرسال».", nil
		}})
}

func deref(s *string, def string) string {
	if s == nil || strings.TrimSpace(*s) == "" {
		return def
	}
	return *s
}

func joinOr(xs []string) string {
	if len(xs) == 0 {
		return "—"
	}
	return strings.Join(xs, "، ")
}
