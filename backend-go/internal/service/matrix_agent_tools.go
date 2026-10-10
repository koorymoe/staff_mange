package service

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ عيون ماتركس — أدوات القراءة (المرحلة الأولى) ═══
// كل أداة تلف استعلاماً موجوداً بالنظام. ما تكتب شي. الإخفاء مركزي بالحلقة،
// وأسماء الزبائن ما تدخل المخرجات أصلاً.

// AgentDeps مصادر الأدوات — كلها موجودة أصلاً بالنظام.
type AgentDeps struct {
	Command   *MatrixCommandService
	Projects  func() (*ProjectsReport, error)
	Scores    func(month string) (*StaffScoreBoard, error)
	ScoreRepo *repository.MatrixScoreRepository
}

var monthRe2 = regexp.MustCompile(`^\d{4}-\d{2}$`)

func bagh(t time.Time) time.Time { return t.In(debriefLoc) }

func money(v float64) string {
	return fmt.Sprintf("%s د.ع", strings.TrimSuffix(fmt.Sprintf("%.0f", v), ".0"))
}

// RegisterReadTools يركّب كل عيون القراءة على الوكيل.
func (a *MatrixAgent) RegisterReadTools(d AgentDeps) {
	monthProp := map[string]any{"month": map[string]any{"type": "string", "description": "الشهر بصيغة YYYY-MM (فاضي = هالشهر)"}}

	a.AddTool(agentTool{Name: "money_month", Desc: "أرقام شهر كامل: الحجوزات (انفتحت/انجزت/انلغت)، المشاريع الجديدة، الأعمال الداخلية، الزبائن الجدد، الشكاوى، فلوس الحجوزات، فلوس المشاريع، المصاريف. استعملها للفلوس والمقارنة بين الأشهر.",
		Props: monthProp,
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			m := strArg(in, "month")
			if !monthRe2.MatchString(m) {
				m = bagh(time.Now()).Format("2006-01")
			}
			st, err := repository.MonthStatsFor(a.db, m)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("شهر %s: حجوزات انفتحت %d، انجزت %d، انلغت %d، أعمال داخلية %d، مشاريع جديدة %d، زبائن جدد %d، شكاوى %d. فلوس الحجوزات %s، فلوس المشاريع %s، المصاريف %s.",
				m, st.BookingsCreated, st.BookingsCompleted, st.BookingsCancelled, st.InternalWorks, st.ProjectsCreated, st.NewCustomers, st.Complaints,
				money(st.BookingsRevenue), money(st.ProjectsRevenue), money(st.Expenses)), nil
		}})

	a.AddTool(agentTool{Name: "invoices_status", Desc: "حالة فواتير الليدر هالشهر: كم مقدّمة بلا تدقيق، مدقّقة تنتظر اعتماد، معتمدة، مجانية، أحكام غير مطابق/خطأ بالسعر، حجوزات منجزة بلا فاتورة، والفرق بين المفوتر والمستلم.",
		Run: func(map[string]any, *nameMask) (string, error) {
			var r struct {
				Submitted, Audited, Approved, Free, Mismatch, PriceErr, NoInvoice int
				Billed, Collected                                                 float64
			}
			err := a.db.QueryRow(`
				SELECT
				 count(*) FILTER (WHERE li.status <> 'APPROVED' AND li."auditVerdict" IS NULL),
				 count(*) FILTER (WHERE li.status <> 'APPROVED' AND li."auditVerdict" IS NOT NULL),
				 count(*) FILTER (WHERE li.status = 'APPROVED'),
				 count(*) FILTER (WHERE li."isFree"),
				 count(*) FILTER (WHERE li."auditVerdict" = 'MISMATCH'),
				 count(*) FILTER (WHERE li."auditVerdict" = 'PRICE_ERROR'),
				 (SELECT count(*) FROM "Booking" b WHERE b.status = 'COMPLETED' AND b."archivedAt" IS NULL
				    AND b."bookingType" IS DISTINCT FROM 'INTERNAL'
				    AND to_char(baghdad_date(b."completedAt"), 'YYYY-MM') = to_char(baghdad_today(), 'YYYY-MM')
				    AND NOT EXISTS (SELECT 1 FROM "LeaderInvoice" x WHERE x."bookingId" = b.id AND x."revokedAt" IS NULL)),
				 COALESCE(sum(li."netTotal") FILTER (WHERE NOT li."isFree"), 0),
				 COALESCE((SELECT sum(COALESCE(b."amountCollected",0) + COALESCE(b."advancePaid",0)) FROM "Booking" b
				    WHERE b.id IN (SELECT "bookingId" FROM "LeaderInvoice" y WHERE y."revokedAt" IS NULL
				      AND to_char(baghdad_date(y."createdAt"), 'YYYY-MM') = to_char(baghdad_today(), 'YYYY-MM'))), 0)
				FROM "LeaderInvoice" li
				WHERE li."revokedAt" IS NULL AND to_char(baghdad_date(li."createdAt"), 'YYYY-MM') = to_char(baghdad_today(), 'YYYY-MM')`).
				Scan(&r.Submitted, &r.Audited, &r.Approved, &r.Free, &r.Mismatch, &r.PriceErr, &r.NoInvoice, &r.Billed, &r.Collected)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("فواتير هالشهر: %d مقدّمة بلا تدقيق، %d مدقّقة تنتظر اعتماد، %d معتمدة، %d مجانية. أحكام: %d غير مطابق، %d خطأ بالسعر. حجوزات منجزة هالشهر بلا فاتورة: %d. المفوتر (غير المجاني) %s مقابل المستلم بحجوزاتها %s.",
				r.Submitted, r.Audited, r.Approved, r.Free, r.Mismatch, r.PriceErr, r.NoInvoice, money(r.Billed), money(r.Collected)), nil
		}})

	a.AddTool(agentTool{Name: "late_bookings", Desc: "التأخير بالحجوزات آخر ٣ أيام مقابل الـ٣ الي قبلها، والحجوزات المتأخرة نفسها (الكود والموعد والحالة).",
		Run: func(map[string]any, *nameMask) (string, error) {
			if d.Command == nil {
				return "", fmt.Errorf("المصدر مو مربوط")
			}
			lf, err := d.Command.LateFocus()
			if err != nil {
				return "", err
			}
			var b strings.Builder
			fmt.Fprintf(&b, "آخر ٣ أيام: %d حجز، متأخر %d (%d٪)، مفتوح وموعده فات %d، بلا كادر %d، جزئي %d. الـ٣ أيام الي قبلها: %d٪.\n",
				lf.Recent.Total, lf.Recent.Late, lf.RecentPct, lf.Recent.OpenNow, lf.Recent.Unstaffed, lf.Recent.Partial, lf.PrevPct)
			for i, it := range lf.Items {
				if i >= 15 {
					break
				}
				fmt.Fprintf(&b, "- %+v\n", it)
			}
			return b.String(), nil
		}})

	a.AddTool(agentTool{Name: "team_performance", Desc: "أداء كل مجموعات الموظفين (٣٠ يوم): لكل موظف حجوزاته/شغل دوره، المنجز، المفتوح، الحضور والتأخير اليوم.",
		Run: func(map[string]any, *nameMask) (string, error) {
			if d.Command == nil {
				return "", fmt.Errorf("المصدر مو مربوط")
			}
			c := d.Command.buildContext()
			return strings.Join(c.lines, "\n"), nil
		}})

	a.AddTool(agentTool{Name: "employee_profile", Desc: "ملف موظف واحد: دوره، حضوره آخر ١٤ يوم (أيام وتأخير)، حجوزاته آخر ٣٠ يوم (مكلّف، منجز، مفتوح، متأخر)، آخر ٥ حجوزات، وتقييم ماتركس هالشهر. استعمل رمز الموظف (موظف#n) أو اسمه.",
		Props: map[string]any{"who": map[string]any{"type": "string", "description": "رمز الموظف مثل موظف#3"}}, Req: []string{"who"},
		Run: func(in map[string]any, mask *nameMask) (string, error) {
			id, name := mask.Resolve(strArg(in, "who"))
			if id == "" {
				return "", fmt.Errorf("ما لگيت الموظف")
			}
			var role string
			_ = a.db.Get(&role, `SELECT role::text FROM "Employee" WHERE id = $1`, id)
			var att struct{ Days, LateDays int }
			_ = a.db.QueryRow(`SELECT count(DISTINCT date), count(DISTINCT date) FILTER (WHERE (("checkIn" AT TIME ZONE 'UTC') + interval '3 hours')::time > time '09:15')
				FROM "Attendance" WHERE "employeeId" = $1 AND "checkIn" > now() - interval '14 days'`, id).Scan(&att.Days, &att.LateDays)
			var bk struct{ Total, Done, Open, Late int }
			_ = a.db.QueryRow(`SELECT count(DISTINCT b.id),
				count(DISTINCT b.id) FILTER (WHERE b.status::text IN ('COMPLETED','PARTIAL')),
				count(DISTINCT b.id) FILTER (WHERE b.status::text IN ('CONFIRMED','IN_PROGRESS')),
				count(DISTINCT b.id) FILTER (WHERE b."completedAt" > b."scheduledAt" + interval '1 day' OR (b.status::text IN ('CONFIRMED','IN_PROGRESS') AND b."scheduledAt" < now() - interval '1 day'))
				FROM "Booking" b LEFT JOIN "BookingAssignment" a ON a."bookingId" = b.id
				WHERE (a."employeeId" = $1 OR b."projectSupervisorId" = $1) AND b."archivedAt" IS NULL AND b."scheduledAt" > now() - interval '30 days'`, id).
				Scan(&bk.Total, &bk.Done, &bk.Open, &bk.Late)
			recent := []struct {
				Code   string     `db:"code"`
				Status string     `db:"status"`
				At     *time.Time `db:"scheduledAt"`
				Done   *time.Time `db:"completedAt"`
			}{}
			_ = a.db.Select(&recent, `SELECT DISTINCT b.code, b.status::text AS status, b."scheduledAt", b."completedAt"
				FROM "Booking" b LEFT JOIN "BookingAssignment" a ON a."bookingId" = b.id
				WHERE (a."employeeId" = $1 OR b."projectSupervisorId" = $1) AND b."archivedAt" IS NULL
				ORDER BY b."scheduledAt" DESC NULLS LAST LIMIT 5`, id)
			var b strings.Builder
			fmt.Fprintf(&b, "%s (%s): حضر %d يوم بآخر ١٤ يوم، منها %d يوم متأخر. حجوزات ٣٠ يوم: %d (منجز %d، مفتوح %d، متأخر %d).\nآخر الحجوزات:\n",
				name, role, att.Days, att.LateDays, bk.Total, bk.Done, bk.Open, bk.Late)
			for _, r := range recent {
				at, done := "—", "—"
				if r.At != nil {
					at = bagh(*r.At).Format("2006-01-02 15:04")
				}
				if r.Done != nil {
					done = bagh(*r.Done).Format("2006-01-02 15:04")
				}
				fmt.Fprintf(&b, "- %s: %s، موعده %s، خلص %s\n", r.Code, r.Status, at, done)
			}
			if d.Scores != nil {
				if board, err := d.Scores(""); err == nil {
					for _, s := range board.Staff {
						if s.ID == id && s.Final != nil {
							fmt.Fprintf(&b, "تقييم ماتركس هالشهر: %.0f٪", *s.Final)
							if s.Reliability != nil {
								fmt.Fprintf(&b, "، الاعتمادية %.0f٪", *s.Reliability)
							}
							for i, l := range s.TopLosses {
								if i >= 3 {
									break
								}
								fmt.Fprintf(&b, "\n  نزّله: %s (%d مرة)", l.Title, l.Count)
							}
						}
					}
				}
			}
			return b.String(), nil
		}})

	a.AddTool(agentTool{Name: "attendance_today", Desc: "دوام اليوم: كم موظف سجّل حضور، منو ما حضر، ومنو تأخر (بلا المدير والمالك).",
		Run: func(map[string]any, *nameMask) (string, error) {
			rows := []struct {
				Name string     `db:"name"`
				In   *time.Time `db:"checkIn"`
			}{}
			err := a.db.Select(&rows, `SELECT e.name, (SELECT min(x."checkIn") FROM "Attendance" x WHERE x."employeeId" = e.id AND x.date = baghdad_today()) AS "checkIn"
				FROM "Employee" e WHERE e.status = 'ACTIVE' AND e.role::text NOT IN ('ADMIN','OWNER') ORDER BY e.name`)
			if err != nil {
				return "", err
			}
			var absent, late []string
			in := 0
			for _, r := range rows {
				if r.In == nil {
					absent = append(absent, r.Name)
					continue
				}
				in++
				if t := bagh(*r.In); t.Hour()*60+t.Minute() > 9*60+15 {
					late = append(late, fmt.Sprintf("%s (%s)", r.Name, t.Format("15:04")))
				}
			}
			return fmt.Sprintf("حضر %d من %d. ما حضر بعد: %s. وصل بعد ٩:١٥: %s.", in, len(rows), strings.Join(absent, "، "), strings.Join(late, "، ")), nil
		}})

	a.AddTool(agentTool{Name: "project_delays", Desc: "المشاريع المفتوحة: المسؤول، المرحلة، أيامها مقابل حدها، اشتغل عليها أحد اليوم، وسبب التأخير المكتوب.",
		Run: func(map[string]any, *nameMask) (string, error) {
			if d.Projects == nil {
				return "", fmt.Errorf("المصدر مو مربوط")
			}
			rep, err := d.Projects()
			if err != nil {
				return "", err
			}
			var b strings.Builder
			for _, p := range rep.Projects {
				if p.Status == ChainNA {
					continue
				}
				reason := "—"
				if p.DelayReason != nil {
					reason = *p.DelayReason
				}
				fmt.Fprintf(&b, "- %s: المسؤول %s، «%s» %d/%d يوم، متجاوز=%v، اشتغل اليوم=%v، السبب: %s\n",
					p.Code, p.OwnerName, p.Stage, p.DaysInStage, p.StageLimit, p.OverLimit, p.WorkedToday, reason)
			}
			if b.Len() == 0 {
				return "ماكو مشاريع مفتوحة.", nil
			}
			return b.String(), nil
		}})

	a.AddTool(agentTool{Name: "procurement_watch", Desc: "شغل المخازن (إداري الكميات): طلبات المواد والأدوات المعلّقة وأعمارها، نواقص الجرد المفتوحة، السيارات الي ما انقيّمت اليوم، والصيانة الي فات موعدها.",
		Run: func(map[string]any, *nameMask) (string, error) {
			if d.ScoreRepo == nil {
				return "", fmt.Errorf("المصدر مو مربوط")
			}
			var b strings.Builder
			list := func(title string, rows []repository.ProcDecision) {
				fmt.Fprintf(&b, "%s: %d معلّق", title, len(rows))
				for i, r := range rows {
					if i >= 5 {
						break
					}
					fmt.Fprintf(&b, "؛ %s من %.0f ساعة", r.Label, time.Since(r.CreatedAt).Hours())
				}
				b.WriteString("\n")
			}
			if m, err := d.ScoreRepo.MaterialDecisions(true); err == nil {
				list("طلبات المواد", m)
			}
			if t, err := d.ScoreRepo.ToolDecisions(true); err == nil {
				list("طلبات الأدوات", t)
			}
			if s, err := d.ScoreRepo.Shortages(true); err == nil {
				list("نواقص الجرد", s)
			}
			if u, err := d.ScoreRepo.VehiclesUnratedToday(); err == nil {
				fmt.Fprintf(&b, "سيارات ما انقيّمت اليوم: %d\n", len(u))
			}
			if f, err := d.ScoreRepo.FleetOverdue(); err == nil {
				fmt.Fprintf(&b, "صيانة/وثائق فات موعدها: %s\n", strings.Join(f, "؛ "))
			}
			return b.String(), nil
		}})

	a.AddTool(agentTool{Name: "staff_scores", Desc: "تقييم ماتركس للموظفين هالشهر (أو شهر محدد): لكل موظف مجموعته والنهائي والاعتمادية وأكثر شي نزّله.",
		Props: monthProp,
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			if d.Scores == nil {
				return "", fmt.Errorf("المصدر مو مربوط")
			}
			board, err := d.Scores(strArg(in, "month"))
			if err != nil {
				return "", err
			}
			var b strings.Builder
			for _, s := range board.Staff {
				final := "—"
				if s.Final != nil {
					final = fmt.Sprintf("%.0f٪", *s.Final)
				}
				fmt.Fprintf(&b, "- %s (%s): %s", s.Name, s.GroupLabel, final)
				if len(s.TopLosses) > 0 {
					fmt.Fprintf(&b, "، نزّله: %s", s.TopLosses[0].Title)
				}
				b.WriteString("\n")
			}
			return b.String(), nil
		}})
}
