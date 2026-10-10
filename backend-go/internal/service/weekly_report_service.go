package service

import (
	"fmt"
	"sort"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ ماتركس — تقرير المالك الأسبوعي ═══
//
// الأسبوع = أسبوع ISO (الاثنين ٠٠:٠٠ بغداد → الاثنين الجاي). ينحسب
// لحظياً بلا جدول. الأسبوع الحالي (ما خلص) ينقارن بنفس المدة المنقضية
// من الأسبوع الي قبله — حتى ما نقارن ٦ أيام بـ٧.
//
// حركة الموظفين: نقاط = منجز − ٢×توقف − ٣×شكوى لكل نافذة؛ الفرق بين
// النافذتين. أعلى ٣ تحسّن (فرق > ٠) وأعلى ٣ تراجع (فرق < ٠). توجيه بس.
//
// الإشعار: كل أحد بعد ٨ الصبح بغداد، مرة بكل أسبوع ISO (ClaimDailyMarker
// بمفتاح اثنين الأسبوع)، للمالك ومدير النظام.

const (
	weeklyReportHour  = 8
	weeklyMoversTopN  = 3
	weeklyHistoryWeek = 12
)

type WeeklyMover struct {
	EmployeeID string `json:"employeeId"`
	Name       string `json:"name"`
	ThisScore  int    `json:"thisScore"`
	LastScore  int    `json:"lastScore"`
	Delta      int    `json:"delta"`
	Note       string `json:"note"`
}

type WeeklyReport struct {
	Week          string                   `json:"week"`
	From          time.Time                `json:"from"`
	To            time.Time                `json:"to"`
	Partial       bool                     `json:"partial"`
	This          repository.WeekTotals    `json:"this"`
	Last          repository.WeekTotals    `json:"last"`
	Improving     []WeeklyMover            `json:"improving"`
	Declining     []WeeklyMover            `json:"declining"`
	OpenDecisions repository.OpenDecisions `json:"openDecisions"`
	Weeks         []string                 `json:"weeks"`
	// Matrix دقة ماتركس لآخر ٣٠ يوم — فاضي لو سجل الأفعال مو مربوط.
	Matrix *model.MatrixAccuracy `json:"matrix,omitempty"`
}

type WeeklyReportService struct {
	aiRepo    *repository.AiRepository
	notifRepo *repository.NotificationRepository
	actions   *repository.AiActionRepository
}

// SetActions يربط سجل أفعال ماتركس حتى التقرير يذكر دقته.
func (s *WeeklyReportService) SetActions(a *repository.AiActionRepository) { s.actions = a }

func NewWeeklyReportService(aiRepo *repository.AiRepository, notifRepo *repository.NotificationRepository) *WeeklyReportService {
	return &WeeklyReportService{aiRepo: aiRepo, notifRepo: notifRepo}
}

// isoWeekLabel «2026-W39».
func isoWeekLabel(t time.Time) string {
	y, w := t.In(debriefLoc).ISOWeek()
	return fmt.Sprintf("%04d-W%02d", y, w)
}

// parseISOWeek يرجّع اثنين الأسبوع (٠٠:٠٠ بغداد).
func parseISOWeek(s string) (time.Time, error) {
	var y, w int
	if _, err := fmt.Sscanf(s, "%4d-W%2d", &y, &w); err != nil || w < 1 || w > 53 {
		return time.Time{}, fmt.Errorf("صيغة الأسبوع لازم YYYY-Www")
	}
	// ٤ كانون الثاني دائماً بأول أسبوع ISO.
	jan4 := time.Date(y, 1, 4, 0, 0, 0, 0, debriefLoc)
	mon := jan4.AddDate(0, 0, -((int(jan4.Weekday()) + 6) % 7))
	mon = mon.AddDate(0, 0, 7*(w-1))
	if yy, ww := mon.ISOWeek(); yy != y || ww != w {
		return time.Time{}, fmt.Errorf("أسبوع غير موجود")
	}
	return mon, nil
}

// weekWindows النافذتين: هذا الأسبوع (مقصوص على now) والي قبله بنفس الطول.
func weekWindows(mon, now time.Time) (curFrom, curTo, prevFrom, prevTo time.Time, partial bool) {
	curFrom, curTo = mon, mon.AddDate(0, 0, 7)
	if now.Before(curTo) {
		curTo, partial = now, true
	}
	prevFrom = mon.AddDate(0, 0, -7)
	prevTo = prevFrom.Add(curTo.Sub(curFrom))
	return
}

type moverTally struct {
	name      string
	cur, prev int
}

// weeklyMovers يحسب أعلى n تحسّن وتراجع من أحداث النافذتين.
func weeklyMovers(events []repository.EmployeeEvent, curFrom, curTo, prevFrom, prevTo time.Time, n int) (up, down []WeeklyMover) {
	by := map[string]*moverTally{}
	for _, e := range events {
		pts := 0
		switch e.Kind {
		case "DONE":
			pts = 1
		case "STOP":
			pts = -2
		case "COMPLAINT":
			pts = -3
		}
		t := by[e.EmployeeID]
		if t == nil {
			t = &moverTally{name: e.Name}
			by[e.EmployeeID] = t
		}
		switch {
		case !e.At.Before(curFrom) && e.At.Before(curTo):
			t.cur += pts
		case !e.At.Before(prevFrom) && e.At.Before(prevTo):
			t.prev += pts
		}
	}
	all := []WeeklyMover{}
	for id, t := range by {
		d := t.cur - t.prev
		if d == 0 {
			continue
		}
		all = append(all, WeeklyMover{EmployeeID: id, Name: t.name, ThisScore: t.cur, LastScore: t.prev, Delta: d,
			Note: fmt.Sprintf("%d ← %d", t.cur, t.prev)})
	}
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].Delta != all[j].Delta {
			return all[i].Delta > all[j].Delta
		}
		return all[i].Name < all[j].Name
	})
	up, down = []WeeklyMover{}, []WeeklyMover{}
	for _, m := range all {
		if m.Delta > 0 && len(up) < n {
			up = append(up, m)
		}
	}
	for i := len(all) - 1; i >= 0 && len(down) < n; i-- {
		if all[i].Delta < 0 {
			down = append(down, all[i])
		}
	}
	return up, down
}

// Report للأسبوع المطلوب ("" = الحالي).
func (s *WeeklyReportService) Report(week string) (*WeeklyReport, error) {
	now := time.Now().In(debriefLoc)
	if week == "" {
		week = isoWeekLabel(now)
	}
	mon, err := parseISOWeek(week)
	if err != nil {
		return nil, err
	}
	if mon.After(now) {
		return nil, fmt.Errorf("الأسبوع بعده ما بدا")
	}
	curFrom, curTo, prevFrom, prevTo, partial := weekWindows(mon, now)
	cur, err := s.aiRepo.WeekTotals(curFrom, curTo)
	if err != nil {
		return nil, err
	}
	prev, err := s.aiRepo.WeekTotals(prevFrom, prevTo)
	if err != nil {
		return nil, err
	}
	events, err := s.aiRepo.EmployeeEvents(nil, prevFrom, curTo)
	if err != nil {
		return nil, err
	}
	up, down := weeklyMovers(events, curFrom, curTo, prevFrom, prevTo, weeklyMoversTopN)
	open, err := s.aiRepo.OpenDecisions()
	if err != nil {
		return nil, err
	}
	weeks := []string{}
	thisMon, _ := parseISOWeek(isoWeekLabel(now))
	for i := 0; i < weeklyHistoryWeek; i++ {
		weeks = append(weeks, isoWeekLabel(thisMon.AddDate(0, 0, -7*i)))
	}
	rep := &WeeklyReport{Week: week, From: curFrom, To: curTo, Partial: partial, This: *cur, Last: *prev,
		Improving: up, Declining: down, OpenDecisions: *open, Weeks: weeks}
	if s.actions != nil {
		if acc, err := s.actions.Accuracy(); err == nil {
			rep.Matrix = acc
		}
	}
	return rep, nil
}

// RunSundayIfDue كل أحد بعد ٨ الصبح بغداد — مرة وحدة بالأسبوع.
func (s *WeeklyReportService) RunSundayIfDue() error {
	now := time.Now().In(debriefLoc)
	if now.Weekday() != time.Sunday || now.Hour() < weeklyReportHour {
		return nil
	}
	claimed, err := s.aiRepo.ClaimDailyMarker("WEEKLY_OWNER_REPORT_SENT", isoWeekMonday(now))
	if err != nil || !claimed {
		return err
	}
	r, err := s.Report("")
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("📊 ماتركس — تقرير الأسبوع %s: منجز %d (قبل %d)، محصّل %s د.ع (قبل %s)، شكاوى %d (قبل %d)، توقفات %d، ورق متأخر %d. قرارات تنتظرك: %d إجازة، %d حذف، %d حكم ماتركس. التفاصيل بشاشة «التقرير الأسبوعي».",
		r.Week, r.This.Completed, r.Last.Completed, fmtIQD(r.This.Revenue), fmtIQD(r.Last.Revenue),
		r.This.Complaints, r.Last.Complaints, r.This.WorkStops, r.This.LatePaperwork,
		r.OpenDecisions.PendingLeaves, r.OpenDecisions.DeleteRequests, r.OpenDecisions.UnreviewedVerdicts)
	if m := r.Matrix; m != nil && m.Actions > 0 {
		msg += fmt.Sprintf(" 🤖 ماتركس بآخر ٣٠ يوم: %d تذكير، انحل منها %d، وصعد %d، ورفضت %d.", m.Actions, m.Resolved, m.Escalated, m.Rejected)
	}
	return notifyOwnerAndAdmin(s.notifRepo, "AI_WEEKLY_REPORT", msg)
}
