package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ سجل الانضباط الوظيفي — الملخّص لكل موظف + تحليل ماتركس (قواعد) ═══

type DisciplineRecordService struct {
	repo *repository.DisciplineRecordRepository
}

func NewDisciplineRecordService(repo *repository.DisciplineRecordRepository) *DisciplineRecordService {
	return &DisciplineRecordService{repo: repo}
}

type DisciplineEmployeeSummary struct {
	EmployeeID   string     `json:"employeeId"`
	EmployeeName string     `json:"employeeName"`
	Count        int        `json:"count"`      // كم مرة انخصم (مال أو نقاط أو تسوية)
	Amount       float64    `json:"amount"`     // مجموع المبالغ المخصومة
	Returned     float64    `json:"returned"`   // المنرجع منها
	Net          float64    `json:"net"`        // الصافي
	PointsLost   int        `json:"pointsLost"` // نقاط الانضباط المخصومة
	PointsBack   int        `json:"pointsBack"` // نقاط رجعت
	ThisMonth    int        `json:"thisMonth"`
	LastMonth    int        `json:"lastMonth"`
	TopReason    string     `json:"topReason"`
	Last         *time.Time `json:"last"`
	Insights     []string   `json:"insights"`
}

type DisciplineRecord struct {
	Employees []DisciplineEmployeeSummary  `json:"employees"`
	Entries   []repository.DisciplineEntry `json:"entries"`
	Totals    map[string]float64           `json:"totals"`
	Insights  []string                     `json:"insights"`
}

func isDeduction(e repository.DisciplineEntry) bool {
	switch e.Source {
	case "POINTS":
		return e.Points < 0
	default:
		return true
	}
}

func shortReason(r string) string {
	r = strings.TrimSpace(r)
	if i := strings.IndexAny(r, "—\n:("); i > 0 && i < 60 {
		r = strings.TrimSpace(r[:i])
	}
	if len([]rune(r)) > 60 {
		r = string([]rune(r)[:60]) + "…"
	}
	if r == "" {
		return "بلا سبب مكتوب"
	}
	return r
}

func (s *DisciplineRecordService) Build(employeeID string) (*DisciplineRecord, error) {
	rows, err := s.repo.All(employeeID)
	if err != nil {
		return nil, err
	}
	now := time.Now().In(debriefLoc)
	thisM := now.Format("2006-01")
	lastM := now.AddDate(0, -1, 0).Format("2006-01")
	per := map[string]*DisciplineEmployeeSummary{}
	reasons := map[string]map[string]int{}
	allReasons := map[string]int{}
	tot := map[string]float64{"amount": 0, "returned": 0, "count": 0}
	for _, e := range rows {
		p := per[e.EmployeeID]
		if p == nil {
			p = &DisciplineEmployeeSummary{EmployeeID: e.EmployeeID, EmployeeName: e.EmployeeName, Insights: []string{}}
			per[e.EmployeeID] = p
			reasons[e.EmployeeID] = map[string]int{}
		}
		if e.Source == "POINTS" {
			if e.Points < 0 {
				p.PointsLost += -e.Points
			} else {
				p.PointsBack += e.Points
			}
		}
		if !isDeduction(e) {
			continue
		}
		p.Count++
		tot["count"]++
		p.Amount += e.Amount
		tot["amount"] += e.Amount
		if e.Returned {
			p.Returned += e.Amount
			tot["returned"] += e.Amount
		}
		m := e.At.In(debriefLoc).Format("2006-01")
		if m == thisM {
			p.ThisMonth++
		} else if m == lastM {
			p.LastMonth++
		}
		if p.Last == nil || e.At.After(*p.Last) {
			t := e.At
			p.Last = &t
		}
		r := shortReason(e.Reason)
		reasons[e.EmployeeID][r]++
		allReasons[r]++
	}
	top := func(m map[string]int) (string, int) {
		best, n := "", 0
		for k, v := range m {
			if v > n || (v == n && k < best) {
				best, n = k, v
			}
		}
		return best, n
	}
	out := &DisciplineRecord{Employees: []DisciplineEmployeeSummary{}, Entries: rows, Totals: tot, Insights: []string{}}
	for id, p := range per {
		p.Net = p.Amount - p.Returned
		if r, n := top(reasons[id]); n > 0 {
			p.TopReason = r
			if n >= 2 {
				p.Insights = append(p.Insights, fmt.Sprintf("أكثر سبب يتكرر عليه: «%s» (%d مرات).", r, n))
			}
		}
		if p.ThisMonth > p.LastMonth && p.ThisMonth >= 2 {
			p.Insights = append(p.Insights, fmt.Sprintf("خصوماته زادت: هالشهر %d، والشهر الفات %d.", p.ThisMonth, p.LastMonth))
		} else if p.LastMonth >= 2 && p.ThisMonth == 0 {
			p.Insights = append(p.Insights, fmt.Sprintf("تحسّن: الشهر الفات انخصم %d مرات، وهالشهر ولا مرة.", p.LastMonth))
		}
		out.Employees = append(out.Employees, *p)
	}
	sort.Slice(out.Employees, func(i, j int) bool {
		if out.Employees[i].Count != out.Employees[j].Count {
			return out.Employees[i].Count > out.Employees[j].Count
		}
		return out.Employees[i].Net > out.Employees[j].Net
	})
	if tot["count"] == 0 {
		out.Insights = append(out.Insights, "ماكو خصومات مسجّلة بالنظام.")
	} else {
		out.Insights = append(out.Insights, fmt.Sprintf("من بداية النظام انخصم %.0f مرة، والمجموع %.0f د.ع، رجع منها %.0f د.ع.", tot["count"], tot["amount"], tot["returned"]))
		if r, n := top(allReasons); n >= 2 {
			out.Insights = append(out.Insights, fmt.Sprintf("أكثر سبب خصم بالشركة: «%s» (%d مرات).", r, n))
		}
	}
	return out, nil
}
