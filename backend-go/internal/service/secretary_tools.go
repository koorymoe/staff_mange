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
// النظام مباشرة. أداة التنفيذ الوحيدة propose_task: المهمة الكاملة (موظف واحد + موعد)
// تنرسل فوراً من المعالج (دازاي 10-23)، والناقصة يسأل عنها بالكلام.

// SecretaryDeps مصادر الأدوات.
type SecretaryDeps struct {
	DB     *sqlx.DB
	Client *anthropic.Client
	Model  string
	Repo   *repository.SecretaryRepository
	Matrix AgentDeps
	// NotifyOwner لاقتراحات الذكاء الي تحتاج برمجة (دازاي 10-28)
	NotifyOwner func(msg string)
}

// NewSecretaryAgent وكيل جديد لكل دور: الأدوات تكتب بطاقاتها ومقترحها على
// tc، فما يتشاركون حالة بين مستخدمين أو طلبين بنفس الوكت.
func NewSecretaryAgent(d SecretaryDeps, tc *SecretaryTurnCtx) *MatrixAgent {
	a := NewMatrixAgent(d.DB, d.Client, d.Model)
	a.ShowNames()
	a.RegisterReadTools(d.Matrix)
	registerSecretaryTools(a, d.Repo, tc)
	a.RegisterSuggestTool(d.DB, "IRIS", d.NotifyOwner)
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
	// 🙋 «منو آني؟» — ملف المستخدم الي يحاچي ايريس هسه (دازاي 10-23: «سألته تعرف منو اني؟ ما عرفني»).
	a.AddTool(agentTool{Name: "whoami",
		Desc: "ملف الشخص الي يحاچيك هسه: اسمه، دوره، قسمه، شهادته، تاريخ تعيينه. استعملها إذا سأل «منو آني؟» أو تحتاج تعرف عنه.",
		Run: func(map[string]any, *nameMask) (string, error) {
			var me struct {
				Name        string  `db:"name"`
				Role        string  `db:"role"`
				Department  *string `db:"department"`
				Certificate *string `db:"certificate"`
				HireDate    *string `db:"hireDate"`
			}
			if err := a.db.Get(&me, `SELECT name, role::text AS role, department, certificate, to_char("hireDate", 'YYYY-MM-DD') AS "hireDate"
				FROM "Employee" WHERE id = $1`, tc.UserID); err != nil {
				return "", err
			}
			role := map[string]string{"OWNER": "مالك الشركة", "ADMIN": "مدير النظام"}[me.Role]
			if role == "" {
				role = me.Role
			}
			out := fmt.Sprintf("الاسم: %s · الدور: %s", me.Name, role)
			if me.Department != nil && *me.Department != "" {
				out += " · القسم: " + *me.Department
			}
			if me.Certificate != nil && *me.Certificate != "" {
				out += " · الشهادة: " + *me.Certificate
			}
			if me.HireDate != nil && *me.HireDate != "" {
				out += " · تاريخ التعيين: " + *me.HireDate
			}
			if strings.TrimSpace(me.Name) == "المالك" {
				out += "\n(اسمه بالنظام «المالك» — إذا يريد تناديه باسمه الحقيقي، يغيّره من إدارة الموظفين ← ملفه.)"
			}
			return out, nil
		}})
	// 📔 مذكرات المستخدم وجدوله — قراءة بس، ملاحظاته هو (دازاي 10-23).
	a.AddTool(agentTool{Name: "my_notes",
		Desc:  "اقرا مذكرات المستخدم («مذكراتي») أو بنود جدوله، الأحدث أول. q اختياري للبحث (مثلاً اسم موظف). استعملها من يسأل «شنو سويت ويا فلان» أو «شنو كتبت».",
		Props: map[string]any{"q": map[string]any{"type": "string", "description": "كلمة بحث اختيارية"}},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			rows, err := repo.Notes(tc.UserID, "", strArg(in, "q"), 20)
			if err != nil {
				return "", err
			}
			if len(rows) == 0 {
				return "ماكو مذكرات مطابقة.", nil
			}
			var b strings.Builder
			for _, n := range rows {
				kind := "مذكرة"
				if n.Kind == "AGENDA" {
					kind = "بند جدول (" + secFmt(n.At) + ")"
				}
				fmt.Fprintf(&b, "- [%s · %s] %s\n", bagh(n.CreatedAt).Format("2006-01-02"), kind, n.Body)
			}
			return b.String(), nil
		}})
	// (دازاي 10-23) «بحثت عن مهندس اسمه مصطفى محمد جابلي حجوزات زبائن بنفس الاسم» — البحث
	// يفرز حسب النوع (kind)، والنتائج مقسّمة بعناوين، وإذا الاسم يطابق موظف وزبون سوه ايريس تسأل
	// قبل ما تخلط. ⚠️ أرقام الهواتف ما تروح للنموذج — بالبطاقات (للشاشة) بس.
	// (دازاي 10-25) «يخبط بين الموظفين والزبائن»: أداة وحدة ترجّع الكل سوه چانت تلخبط
	// النموذج. هسه أداة لكل جدول، واسمها يكول شنو ترجّع — ما يكدر يخلط.
	findKind := func(kind, q string) (string, error) {
		want := func(k string) bool { return kind == "all" || kind == k }
		f, err := repo.Find(q)
		if err != nil {
			return "", err
		}
		var b strings.Builder
		section := func(title string, n int) {
			if n > 0 {
				fmt.Fprintf(&b, "\n### %s (%d)\n", title, n)
			}
		}
		if want("employee") {
			section("موظفين", len(f.Employees))
			for _, x := range f.Employees {
				fmt.Fprintf(&b, "- موظف: %s · %s\n", x.Name, x.Role)
			}
		}
		if want("customer") {
			section("زبائن", len(f.Customers))
			for _, x := range f.Customers {
				fmt.Fprintf(&b, "- زبون #%d: %s · حجوزاته: %d\n", x.Code, x.Name, x.Bookings)
				tc.addCard(SecretaryCard{Kind: "customer", Title: x.Name,
					Lines: []string{"الهاتف: " + deref(x.Phone, "—"), fmt.Sprintf("حجوزاته: %d", x.Bookings)},
					Link:  strPtr("/customers")})
			}
		}
		if want("booking") {
			section("حجوزات (اسم الزبون مو اسم موظف)", len(f.Bookings))
			for _, x := range f.Bookings {
				fmt.Fprintf(&b, "- حجز %s — الزبون: %s · %s · الحالة: %s · الموعد: %s · الكادر: %s\n",
					x.Code, deref(x.Customer, "—"), deref(x.Service, "—"), x.Station, secFmt(x.ScheduledAt), deref(x.Crew, "بلا كادر"))
				tc.addCard(SecretaryCard{Kind: "booking", Title: "حجز " + x.Code + " — " + deref(x.Customer, ""),
					Status: x.Station, Owner: x.Crew, UpdatedAt: secISO(x.UpdatedAt),
					Lines: []string{"الخدمة: " + deref(x.Service, "—"), "الموعد: " + secFmt(x.ScheduledAt)},
					Link:  strPtr("/bookings?focus=" + x.Code)})
			}
		}
		if want("project") {
			section("مشاريع", len(f.Projects))
			for _, x := range f.Projects {
				fmt.Fprintf(&b, "- مشروع %s %s — المرحلة: %s · المسؤول: %s\n", deref(x.Code, ""), x.Name, x.Stage, deref(x.Owner, "—"))
				tc.addCard(SecretaryCard{Kind: "project", Title: strings.TrimSpace(deref(x.Code, "") + " " + x.Name),
					Status: x.Stage, Owner: x.Owner, UpdatedAt: secISO(x.UpdatedAt), Lines: []string{}, Link: strPtr("/projects")})
			}
		}
		if b.Len() == 0 {
			return "ما لگيت شي بهالاسم أو الكود (النوع: " + kind + ").", nil
		}
		if kind == "all" && len(f.Employees) > 0 && (len(f.Customers) > 0 || len(f.Bookings) > 0) {
			b.WriteString("\n⚠️ الاسم يطابق موظف وزبون/حجوزات سوه. إذا ما واضح من سؤاله منو يقصد، اسأله «تقصد الموظف لو الزبون؟» — ولا تخلط النتائج.")
		}
		return strings.TrimSpace(b.String()), nil
	}
	findTool := func(name, kind, desc string) {
		a.AddTool(agentTool{Name: name, Desc: desc,
			Props: map[string]any{"q": map[string]any{"type": "string", "description": "الاسم أو الكود"}},
			Req:   []string{"q"},
			Run:   func(in map[string]any, _ *nameMask) (string, error) { return findKind(kind, strArg(in, "q")) }})
	}
	findTool("find_employee", "employee", "دوّر على **موظف** بالشركة بالاسم (مهندس، فني، ليدر، محاسب، إداري…). ترجّع موظفين بس — مو زبائن.")
	findTool("find_customer", "customer", "دوّر على **زبون** بالاسم أو رقم الزبون. ترجّع زبائن بس — مو موظفين.")
	findTool("find_booking", "booking", "دوّر على **حجز** بكوده أو باسم الزبون صاحب الحجز. الاسم بالنتيجة اسم الزبون، والكادر مكتوب لحاله.")
	findTool("find_project", "project", "دوّر على **مشروع** بكوده أو اسمه.")
	findTool("find_any", "all", "إذا ما واضح هل يقصد موظف لو زبون لو حجز: يدوّر بالكل ويفصل النتائج. إذا الاسم طلع موظف وزبون سوه، اسأله منو يقصد قبل ما تجاوب.")

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
				return "ما لگيت حجز بهالكود — جرّب find_booking باسم الزبون.", nil
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
		Desc: "وجّه مهمة لموظف — تنرسل له فوراً بالنظام. لازم الموظف واضح وآخر موعد للإنجاز محدد (من كلامه: «باچر الصبح»، «الخميس ١٢»…). إذا ناقص شي، الأداة ترجّعلك شنو تسأل. مرة وحدة بالدور.",
		Props: map[string]any{
			"employee":    map[string]any{"type": "string", "description": "اسم الموظف كما گاله المستخدم"},
			"title":       map[string]any{"type": "string", "description": "عنوان قصير (لحد ٨ كلمات)"},
			"description": map[string]any{"type": "string", "description": "شنو المطلوب بالضبط"},
			"priority":    map[string]any{"type": "string", "enum": []string{"NORMAL", "URGENT"}},
			"dueAt":       map[string]any{"type": "string", "description": "آخر موعد للإنجاز YYYY-MM-DDTHH:MM بتوقيت بغداد — حوّل كلامه لوقت (الصبح=10:00، الظهر=13:00، آخر الدوام=16:00)"},
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
			ids := matchEmployeeBest(strArg(in, "employee"), list)
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
			// (دازاي 10-23) «اريده يشتغل كلشي عني» — بلا قوائم: إذا الاسم يطابق أكثر من واحد
			// أو الموعد ناقص، السكرتير يسأل بالكلام مثل سكرتير حقيقي، وما تنعمل مسودة.
			if len(ids) > 1 {
				opts := make([]string, 0, len(ids))
				for _, id := range ids {
					opts = append(opts, names[id])
				}
				return "ما انرسلت — الاسم يطابق: " + strings.Join(opts, "، ") + ". إذا واضح من الكلام منو يقصد عيد الأداة بالاسم الكامل، وإلا اسأله بجملة وحدة منو يقصد (اذكر الأسماء).", nil
			}
			if f.DueAt == "" {
				return "ما انرسلت — ناقص آخر موعد للإنجاز. اسأله بجملة وحدة: شوكت لازم يخلص؟", nil
			}
			cands := []SecretaryCandidate{{ID: ids[0], Name: names[ids[0]]}}
			f.EmployeeID, f.EmployeeName = ids[0], names[ids[0]]
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
			return "تمام — المهمة راح تنرسل لـ" + f.EmployeeName + " هسه. أكّدله بجملة قصيرة (منو، شنو، شوكت).", nil
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
