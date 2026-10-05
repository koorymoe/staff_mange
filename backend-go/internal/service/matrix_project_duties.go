package service

import (
	"fmt"
	"sync"
	"time"
)

// ═══ عين المشروع فوق عين الدور — طلب (ع) 10-05 ═══
// «اذا توجهلي مشروع… العين الي تراقبني تكوم تراقبني للضوابط المعروفه
// وللضوابط مال المشاريع». نفس أحكام «عين المشاريع» (judge) — ماكو ضوابط
// جديدة، بس تنحسب على عين صاحب المشروع ويا شغل دوره.

type projectDutiesCache struct {
	mu  sync.Mutex
	at  time.Time
	val map[string][]ProjectVerdict
}

// SetProjectDuties يربط مصدر المشاريع (MatrixProjectService.Report).
func (s *MatrixAutopilotService) SetProjectDuties(f func() (*ProjectsReport, error)) { s.projectsSrc = f }

// projectDuties المشاريع المفتوحة لكل مسؤول — مخزّنة دقيقتين حتى عين الهيدر ما تثقل.
func (s *MatrixAutopilotService) projectDuties() map[string][]ProjectVerdict {
	if s.projectsSrc == nil {
		return nil
	}
	c := &s.projCache
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.val != nil && time.Since(c.at) < 2*time.Minute {
		return c.val
	}
	rep, err := s.projectsSrc()
	if err != nil || rep == nil {
		return c.val
	}
	c.val, c.at = groupProjectDuties(rep.Projects), time.Now()
	return c.val
}

func groupProjectDuties(ps []ProjectVerdict) map[string][]ProjectVerdict {
	out := map[string][]ProjectVerdict{}
	for _, p := range ps {
		if p.OwnerID == "" || p.Status == ChainNA {
			continue
		}
		out[p.OwnerID] = append(out[p.OwnerID], p)
	}
	return out
}

// projectWatch بنود العين و«مشاريعي» من مشاريع الموظف.
func projectWatch(ps []ProjectVerdict) ([]WatchItem, *WorkloadItem) {
	if len(ps) == 0 {
		return nil, nil
	}
	items := []WatchItem{}
	w := &WorkloadItem{Key: "PROJECTS_MINE", Label: "مشاريعي", Verb: "ماشية", Route: "/projects"}
	for _, p := range ps {
		if p.Status == ChainOK {
			w.Done++
			continue
		}
		w.Left++
		since := p.UpdatedAt
		if p.StageSince != nil {
			since = *p.StageSince
		}
		items = append(items, WatchItem{Kind: "PROJECT", Label: "ضوابط المشروع",
			Summary:   fmt.Sprintf("📐 مشروع «%s»: %s", p.Name, p.Verdict),
			Escalated: p.Status == ChainMissed, Since: since})
	}
	return items, w
}
