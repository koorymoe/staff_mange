package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"staffmange-api/internal/model"
)

// ═══ أداء المجموعة — جدول غني لعين المدير + أكثر ٣ مشاكل ═══
//
// لكل موظف بالمجموعة: حضوره اليوم وتأخيره، حجوزاته بآخر ٣٠ يوم (منجز/جزئي/مفتوح)،
// معامل سرعته (الفعلي ÷ المتوقع) مقابل الشهر الي قبله، ونسبة الجزئي.
// ترتيب حسب الأسوأ. ماكو غرامات ولا نقاط — أرقام وروابط بس.

type GroupMember struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	CheckIn   string   `json:"checkIn,omitempty"`
	Late      int      `json:"late"` // دقائق
	Absent    bool     `json:"absent"`
	Jobs      int      `json:"jobs"`
	Completed int      `json:"completed"`
	Partial   int      `json:"partial"`
	Open      int      `json:"open"`
	Speed     *float64 `json:"speed"`     // الفعلي ÷ المتوقع (١ = طبيعي)
	PrevSpeed *float64 `json:"prevSpeed"` // نفسه للشهر الي قبله
	Score     int      `json:"score"`     // للترتيب بس — أعلى = يحتاج انتباه أكثر
}

type GroupProblem struct {
	Text  string   `json:"text"`
	Link  string   `json:"link,omitempty"`
	Fixes []string `json:"fixes"`
}

type GroupPerformance struct {
	Group    string         `json:"group"`
	Members  []GroupMember  `json:"members"`
	Problems []GroupProblem `json:"problems"`
}

func speedOf(timed int, act, exp float64) *float64 {
	if timed < 2 || exp <= 0 {
		return nil
	}
	v := float64(int(act/exp*100)) / 100
	return &v
}

func (s *MatrixEmployeeReportService) GroupPerformance(group string) (*GroupPerformance, error) {
	subs, err := s.watch.actions.ActiveSubjects()
	if err != nil {
		return nil, err
	}
	now := time.Now().In(debriefLoc)
	day := now.Format("2006-01-02")
	dayT, _ := time.ParseInLocation("2006-01-02", day, debriefLoc)
	out := &GroupPerformance{Group: group, Members: []GroupMember{}, Problems: []GroupProblem{}}
	var late, absent, slow, partial []string
	for _, sub := range subs {
		if WatchGroup(sub) != group {
			continue
		}
		m := GroupMember{ID: sub.ID, Name: sub.Name}
		if e, err := s.repo.Employee(sub.ID); err == nil {
			start, _, _ := shiftWindow(e)
			if att, _ := s.repo.Attendance(sub.ID, day); len(att) > 0 {
				m.CheckIn = hm(att[0].CheckIn)
				if l := mins(att[0].CheckIn.Sub(atClock(dayT, start))); l > lateGraceMin {
					m.Late = l
					late = append(late, sub.Name)
				}
			} else if now.After(atClock(dayT, start).Add(time.Hour)) && now.Weekday() != time.Friday {
				m.Absent = true
				absent = append(absent, sub.Name)
			}
		}
		if p, err := s.repo.Performance(sub.ID, 0); err == nil {
			m.Jobs, m.Completed, m.Partial = p.Total, p.Completed, p.Partial
			m.Open = p.Total - p.Completed - p.Partial
			m.Speed = speedOf(p.Timed, p.AvgActual, p.AvgExpected)
			if m.Speed != nil && *m.Speed >= 1.4 {
				slow = append(slow, sub.Name)
			}
			if p.Total >= 4 && p.Partial*4 >= p.Total {
				partial = append(partial, sub.Name)
			}
		}
		if pp, err := s.repo.Performance(sub.ID, 30); err == nil {
			m.PrevSpeed = speedOf(pp.Timed, pp.AvgActual, pp.AvgExpected)
		}
		m.Score = m.Late/10 + m.Open*2 + m.Partial*3
		if m.Absent {
			m.Score += 10
		}
		if m.Speed != nil && *m.Speed > 1 {
			m.Score += int((*m.Speed - 1) * 20)
		}
		out.Members = append(out.Members, m)
	}
	sort.SliceStable(out.Members, func(i, j int) bool { return out.Members[i].Score > out.Members[j].Score })

	add := func(names []string, text string, link string, fixes ...string) {
		if len(names) > 0 {
			out.Problems = append(out.Problems, GroupProblem{Text: fmt.Sprintf(text, len(names), strings.Join(firstN(names, 3), "، ")), Link: link, Fixes: fixes})
		}
	}
	add(absent, "%d ما سجّلوا حضور لهسه: %v", "/work-schedule", "اتصل بيهم تتأكد.", "تأكد إن جدولهم مسجّل صح.")
	add(slow, "%d يطوّلون بالحجز أكثر من المتوقع بـ٤٠٪ أو أكثر: %v", "/bookings", "وزّع الحجوزات الكبيرة على كادر أكبر.", "تدريب على الخدمات الي يطوّلون بيها.")
	add(partial, "%d ربع حجوزاتهم أو أكثر جزئية: %v", "/bookings", "تأكد من تجهيز المواد والعدّة قبل الطلعة.")
	add(late, "%d تأخّروا اليوم: %v", "/work-schedule", "ذكّرهم قبل الدوام.", "راجع إذا وقت البداية يناسب شغلهم.")
	if len(out.Problems) > 3 {
		out.Problems = out.Problems[:3]
	}
	return out, nil
}

// ── متابعة مستمرة: تراجع الأداء (يومياً بعد ٥ العصر) ──
// يقارن معامل السرعة لآخر ٣٠ يوم بالي قبلها؛ لو نزل ٣٠٪ أو أكثر يحط اقتراح
// بصندوق القرارات — الموافقة ترسل للموظف رسالة لطيفة، بلا عقوبة.
const declineHour = 17

func (s *MatrixLearningService) SetReportService(r *MatrixEmployeeReportService) { s.reports = r }

func (s *MatrixLearningService) detectDeclines() int {
	if s.reports == nil {
		return 0
	}
	subs, err := s.actions.ActiveSubjects()
	if err != nil {
		return 0
	}
	month := time.Now().In(debriefLoc).Format("2006-01")
	n := 0
	for _, sub := range subs {
		cur, err1 := s.reports.repo.Performance(sub.ID, 0)
		prev, err2 := s.reports.repo.Performance(sub.ID, 30)
		if err1 != nil || err2 != nil || cur.Timed < 3 || prev.Timed < 3 || cur.AvgExpected <= 0 || prev.AvgExpected <= 0 {
			continue
		}
		c, p := cur.AvgActual/cur.AvgExpected, prev.AvgActual/prev.AvgExpected
		drop := (c - p) / p * 100
		if drop < 30 {
			continue
		}
		ok, _ := s.props.Create(model.MatrixProposal{
			Kind: model.ProposalPrediction, Source: "RULES",
			Title:     fmt.Sprintf("%s: صار يطوّل بالحجوزات %d٪ أكثر من الشهر الي قبله.", sub.Name, int(drop)),
			Rationale: fmt.Sprintf("معدّل الفعلي ÷ المتوقع صار %.2f بعد ما چان %.2f (%d حجز موقوت مقابل %d). شوف تقريره: /matrix/employee/%s", c, p, cur.Timed, prev.Timed, sub.ID),
			Evidence:  toJSON(map[string]any{"ratio": c, "prevRatio": p, "timed": cur.Timed, "prevTimed": prev.Timed}),
			Payload: toJSON(map[string]any{"employeeId": sub.ID, "route": "/matrix/employee/" + sub.ID,
				"message": "🤖 ماتركس — لاحظت حجوزاتك هالشهر تاخذ وكت أكثر من قبل. إذا أكو عائق (مواد، طريق، فريق) گوله لمسؤولك حتى ينحل."}),
			Signature: fmt.Sprintf("DECLINE|%s|%s", sub.ID, month),
		})
		if ok {
			n++
		}
	}
	return n
}
