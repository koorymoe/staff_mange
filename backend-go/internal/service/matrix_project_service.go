package service

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ ماتركس ٢٠٥٠ — عين على المشاريع ═══
// طلب (ع) 10-05: «عدنه تقنيين ومصممين ومشرفين مشاريع، أي احد يتوجهله مشروع…
// بغض النظر عن الدور ايضاً كون عليه رقابه. نريد نشوف شنو ترتيب العمل بالمشروع».
// لكل مشروع: مراحله وشكد بقى بكل وحدة (مقابل وسيط الشركة)، منو لمسه وبأي
// ترتيب، ووين واقف. ولكل شخص: مشاريعه والعالق منها. قراءة بس — ماكو نقاط.

type MatrixProjectService struct {
	repo *repository.MatrixProjectRepository
}

func NewMatrixProjectService(repo *repository.MatrixProjectRepository) *MatrixProjectService {
	return &MatrixProjectService{repo: repo}
}

func projectFinal(stage string) bool {
	return strings.Contains(stage, "مكتمل") || strings.Contains(stage, "مرفوض")
}

type ProjectVerdict struct {
	repository.ProjectFacts
	OwnerID     string   `json:"ownerId"`
	OwnerName   string   `json:"ownerName"`
	DaysInStage int      `json:"daysInStage"`
	StageLimit  int      `json:"stageLimit"`
	IdleDays    int      `json:"idleDays"`
	Status      string   `json:"status"` // OK | LATE | MISSED | ISSUE | NA
	Problems    []string `json:"problems"`
	Verdict     string   `json:"verdict"`
}

type ProjectPersonStat struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Owned    int      `json:"owned"`
	Open     int      `json:"open"`
	Stuck    int      `json:"stuck"`
	Done     int      `json:"done"`
	AvgIdle  float64  `json:"avgIdle"`
	Insights []string `json:"insights"`
}

type ProjectsReport struct {
	Projects []ProjectVerdict    `json:"projects"`
	People   []ProjectPersonStat `json:"people"`
	StageMed map[string]float64  `json:"stageMedianDays"`
	Insights []string            `json:"insights"`
}

func (s *MatrixProjectService) limits() (map[string]int, map[string]float64) {
	rows, _ := s.repo.StageDurations()
	by := map[string][]float64{}
	for _, r := range rows {
		by[r.Stage] = append(by[r.Stage], r.Days)
	}
	lim, med := map[string]int{}, map[string]float64{}
	for st, v := range by {
		sort.Float64s(v)
		m := v[len(v)/2]
		med[st] = math.Round(m*10) / 10
		l := int(math.Ceil(m * 1.5))
		if l < 3 {
			l = 3
		}
		lim[st] = l
	}
	return lim, med
}

func (s *MatrixProjectService) judge(f repository.ProjectFacts, lim map[string]int, now time.Time) ProjectVerdict {
	v := ProjectVerdict{ProjectFacts: f, Problems: []string{}}
	switch {
	case f.DelegatedToID != nil:
		v.OwnerID, v.OwnerName = *f.DelegatedToID, str(f.DelegatedTo)
	case f.ResponsibleID != nil:
		v.OwnerID, v.OwnerName = *f.ResponsibleID, str(f.Responsible)
	case f.SurveyorID != nil:
		v.OwnerID, v.OwnerName = *f.SurveyorID, str(f.Surveyor)
	case f.CreatedByID != nil:
		v.OwnerID, v.OwnerName = *f.CreatedByID, str(f.CreatedBy)
	}
	since := f.UpdatedAt
	if f.StageSince != nil {
		since = *f.StageSince
	}
	v.DaysInStage = int(now.Sub(since).Hours() / 24)
	last := f.UpdatedAt
	if f.LastActivityAt != nil && f.LastActivityAt.After(last) {
		last = *f.LastActivityAt
	}
	v.IdleDays = int(now.Sub(last).Hours() / 24)
	v.StageLimit = lim[f.Stage]
	if v.StageLimit == 0 {
		v.StageLimit = 7 // ماكو عينات بعد: أسبوع
	}
	if projectFinal(f.Stage) {
		v.Status = ChainNA
		v.Verdict = "المشروع " + strings.TrimSpace(strings.TrimLeft(f.Stage, "✅❌ ")) + "."
		return v
	}
	if v.OwnerID == "" {
		v.Problems = append(v.Problems, "بلا مسؤول — محد متوجهله.")
	}
	if v.DaysInStage > v.StageLimit {
		v.Problems = append(v.Problems, fmt.Sprintf("صارله %d يوم بمرحلة «%s» والمعتاد لحد %d.", v.DaysInStage, f.Stage, v.StageLimit))
	}
	if v.IdleDays >= 7 {
		v.Problems = append(v.Problems, fmt.Sprintf("ماكو أي حركة عليه من %d يوم.", v.IdleDays))
	}
	if f.DeliveryDate != nil && *f.DeliveryDate != "" {
		if d, err := time.ParseInLocation("2006-01-02", (*f.DeliveryDate)[:minInt(10, len(*f.DeliveryDate))], debriefLoc); err == nil && now.After(d.Add(24*time.Hour)) {
			v.Problems = append(v.Problems, fmt.Sprintf("فات موعد التسليم (%s) بـ%d يوم.", d.Format("2006-01-02"), int(now.Sub(d).Hours()/24)))
		}
	}
	if strings.HasPrefix(f.Stage, "5") && !f.HasSigned {
		v.Problems = append(v.Problems, "بالتنفيذ بلا عقد موقّع مرفوع.")
	} else if strings.HasPrefix(f.Stage, "4") && !f.HasContract && v.DaysInStage >= 2 {
		v.Problems = append(v.Problems, "بمرحلة العقد وبعد ما انرفع عقد.")
	}
	switch {
	case v.IdleDays >= 14 || (v.DaysInStage > v.StageLimit*2):
		v.Status = ChainMissed
	case len(v.Problems) > 0 && (v.DaysInStage > v.StageLimit || v.IdleDays >= 7):
		v.Status = ChainLate
	case len(v.Problems) > 0:
		v.Status = ChainIssue
	default:
		v.Status = ChainOK
	}
	if len(v.Problems) == 0 {
		v.Verdict = fmt.Sprintf("ماشي: %d يوم بالمرحلة من أصل %d، وآخر حركة قبل %d يوم.", v.DaysInStage, v.StageLimit, v.IdleDays)
	} else {
		v.Verdict = strings.Join(v.Problems, " ")
	}
	return v
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *MatrixProjectService) Report() (*ProjectsReport, error) {
	facts, err := s.repo.All()
	if err != nil {
		return nil, err
	}
	lim, med := s.limits()
	now := time.Now()
	rep := &ProjectsReport{Projects: []ProjectVerdict{}, People: []ProjectPersonStat{}, StageMed: med, Insights: []string{}}
	people := map[string]*ProjectPersonStat{}
	idle := map[string][]int{}
	stuck, open := 0, 0
	for _, f := range facts {
		v := s.judge(f, lim, now)
		rep.Projects = append(rep.Projects, v)
		if v.OwnerID != "" {
			p := people[v.OwnerID]
			if p == nil {
				p = &ProjectPersonStat{ID: v.OwnerID, Name: v.OwnerName, Insights: []string{}}
				people[v.OwnerID] = p
			}
			p.Owned++
			if projectFinal(f.Stage) {
				if strings.Contains(f.Stage, "مكتمل") {
					p.Done++
				}
			} else {
				p.Open++
				idle[v.OwnerID] = append(idle[v.OwnerID], v.IdleDays)
				if v.Status == ChainLate || v.Status == ChainMissed {
					p.Stuck++
				}
			}
		}
		if !projectFinal(f.Stage) {
			open++
			if v.Status == ChainLate || v.Status == ChainMissed {
				stuck++
			}
		}
	}
	// الأسوأ أول: الي ما انسوّت، بعدين المتأخر، بعدين النواقص.
	rank := map[string]int{ChainMissed: 0, ChainLate: 1, ChainIssue: 2, ChainOK: 3, ChainNA: 4}
	sort.SliceStable(rep.Projects, func(i, j int) bool {
		return rank[rep.Projects[i].Status] < rank[rep.Projects[j].Status]
	})
	for id, p := range people {
		if xs := idle[id]; len(xs) > 0 {
			sum := 0
			for _, x := range xs {
				sum += x
			}
			p.AvgIdle = math.Round(float64(sum)*10/float64(len(xs))) / 10
		}
		if p.Open > 0 && p.Stuck*2 >= p.Open && p.Stuck >= 2 {
			p.Insights = append(p.Insights, fmt.Sprintf("نص مشاريعه المفتوحة أو أكثر عالقة (%d من %d).", p.Stuck, p.Open))
		}
		if p.AvgIdle >= 7 {
			p.Insights = append(p.Insights, fmt.Sprintf("معدل سكون مشاريعه %.1f يوم بلا حركة.", p.AvgIdle))
		}
		if p.Done > 0 && p.Owned >= 3 {
			p.Insights = append(p.Insights, fmt.Sprintf("خلّص %d من %d.", p.Done, p.Owned))
		}
		rep.People = append(rep.People, *p)
	}
	sort.Slice(rep.People, func(i, j int) bool {
		if rep.People[i].Stuck != rep.People[j].Stuck {
			return rep.People[i].Stuck > rep.People[j].Stuck
		}
		return rep.People[i].Open > rep.People[j].Open
	})
	if open == 0 {
		rep.Insights = append(rep.Insights, "ماكو مشاريع مفتوحة هسه.")
	} else {
		rep.Insights = append(rep.Insights, fmt.Sprintf("%d مشروع مفتوح — %d منها عالق أو واقف.", open, stuck))
	}
	slow, sv := "", 0.0
	for st, m := range med {
		if !projectFinal(st) && m > sv {
			slow, sv = st, m
		}
	}
	if slow != "" {
		rep.Insights = append(rep.Insights, fmt.Sprintf("أبطأ مرحلة بالشركة «%s» — المشروع يبقى بيها بالوسط %.1f يوم.", slow, sv))
	}
	return rep, nil
}

type ProjectChain struct {
	Verdict     ProjectVerdict                    `json:"verdict"`
	Stages      []repository.ProjectStageRow      `json:"stages"`
	Touchers    []repository.ProjectTouchRow      `json:"touchers"`
	Delegations []repository.ProjectDelegationRow `json:"delegations"`
	Order       []string                          `json:"order"` // ترتيب العمل بجمل
}

func (s *MatrixProjectService) Chain(id string) (*ProjectChain, error) {
	f, err := s.repo.One(id)
	if err != nil {
		return nil, fmt.Errorf("المشروع مو موجود")
	}
	lim, _ := s.limits()
	out := &ProjectChain{Verdict: s.judge(*f, lim, time.Now()), Order: []string{}}
	out.Stages, _ = s.repo.Stages(id)
	out.Touchers, _ = s.repo.Touchers(id)
	out.Delegations, _ = s.repo.Delegations(id)
	// ترتيب العمل: كل مرحلة شكد بقت ومنو نقلها.
	for i, st := range out.Stages {
		end := time.Now()
		if i+1 < len(out.Stages) {
			end = out.Stages[i+1].At
		}
		days := end.Sub(st.At).Hours() / 24
		by := ""
		if st.By != nil {
			by = " — نقله " + *st.By
		}
		tail := "بقى بيها"
		if i+1 == len(out.Stages) {
			tail = "صارله بيها"
		}
		out.Order = append(out.Order, fmt.Sprintf("%s: دخلها %s%s، %s %.1f يوم.", st.ToStage, st.At.In(debriefLoc).Format("2006-01-02"), by, tail, days))
	}
	return out, nil
}
