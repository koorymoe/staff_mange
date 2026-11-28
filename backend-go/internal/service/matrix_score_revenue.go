package service

import (
	"math"
	"strconv"
	"strings"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ «ماكو حد — المهم كل موظف يدخّل فلوس أكثر من راتبه» — دازاي 11-13 ═══
// كل دور ياخذ الفلوس من مصدر شغله هو:
//   المبيعات والمنسّقين ← فواتير الحجوزات الي سجّلوها
//   الفنيين والليدرية والديكور والتقنيين ← حصتهم من فواتير الحجوزات الي اشتغلوها
//   إدارة المشاريع والتصميم ← دفعات المشاريع الي هم مسؤوليها
// تغطية الراتب = الي دخّله ÷ راتب الفترة. الي ما غطّى راتبه ينزل نهائيه بنسبة
// النقص (ما يشتغل = ما يطلع عالي حتى لو ماكو مشاكل)، والي زاد عليه ياخذ زيادة.
// الأدوار الي ما تدخّل فلوس مباشرة (محاسبة، مراقبة، آيتي…) ما تنحسب عليها.

var revenueKindOf = map[string]string{
	// (دازاي 11-13) الإداريين/المنسّقين ينسّقون بس — مو مصدر فلوس، فما ينحسب عليهم
	// إيراد ولا تغطية راتب؛ ينقيّمون بشغل التنسيق نفسه.
	GroupSales: "BOOKED",
	GroupTechs: "WORK", GroupLeaders: "WORK", GroupDecor: "WORK", GroupTechnical: "WORK",
	GroupProjects: "PROJECT", GroupDesign: "PROJECT",
}

const (
	// RevenueFloor: الي ما دخّل ولا دينار ينزل نهائيه للنص.
	RevenueFloor = 0.5
	// RevenueGraceDays (دازاي 11-17): بأول هالأيام من الفترة تغطية الراتب ما تنقّص.
	RevenueGraceDays = 7
	// RevenueBonusMax: أعلى زيادة لمن يغطي راتبه وأكثر (نقطة لكل ١٠٪ فوك الراتب).
	RevenueBonusMax = 10.0
)

// revenueTarget (دازاي 11-18) الهدف كنسبة من الراتب — الفني لازم يدخّل ١٣٠٪، البقية ١٠٠٪.
var revenueTarget = map[string]float64{GroupTechs: 130, GroupLeaders: 130}

func targetOf(group string) float64 {
	if t, ok := revenueTarget[group]; ok {
		return t
	}
	return 100
}

// salaryFor راتب الفترة: الشهر الجاري بنسبة الأيام الي مرّت بس.
// elapsedDays الأيام الي مرّت من الفترة لحد هسه (أو لنهايتها).
func elapsedDays(from, to, now time.Time) float64 {
	end := to
	if now.Before(end) {
		end = now
	}
	return end.Sub(from).Hours() / 24
}

func salaryFor(salary float64, from, to, now time.Time) float64 {
	end := to
	if now.Before(end) {
		end = now
	}
	days := end.Sub(from).Hours() / 24
	if days <= 0 {
		return 0
	}
	return salary * math.Min(days/30, 1)
}

// ApplyRevenue يضيف لكل صف الفلوس الي دخّلها وتغطية راتبه، ويعدّل النهائي.
func ApplyRevenue(rows []StaffScore, rev []repository.RevenueRow, salaries map[string]float64, from, to, now time.Time) {
	by := map[string]map[string]repository.RevenueRow{}
	for _, r := range rev {
		if by[r.EmployeeID] == nil {
			by[r.EmployeeID] = map[string]repository.RevenueRow{}
		}
		x := by[r.EmployeeID][r.Kind]
		x.Amount += r.Amount
		x.Jobs += r.Jobs
		by[r.EmployeeID][r.Kind] = x
	}
	for i := range rows {
		s := &rows[i]
		kind, ok := revenueKindOf[s.Group]
		if !ok {
			continue
		}
		got := by[s.ID][kind]
		// (دازاي 11-28) تجديد الـGPS فلوس دخّلها هو — تنضاف لأي دور يدخّل فلوس
		if kind != "GPS" {
			g := by[s.ID]["GPS"]
			got.Amount += g.Amount
			got.Jobs += g.Jobs
		}
		s.Revenue, s.RevenueJobs, s.RevenueKind = math.Round(got.Amount), got.Jobs, kind
		expected := salaryFor(salaries[s.ID], from, to, now)
		if expected <= 0 {
			continue
		}
		c := got.Amount / expected * 100
		s.SalaryCoverage = &c
		if s.Final == nil {
			continue
		}
		// (دازاي 11-17) أول أسبوع من الفترة (أول الشهر، أو أول أيام دور جديد) بعده ما لحق يشتغل
		// ويدخّل — التغطية تنعرض بس ما تنقّص الرقم. الزيادة (إذا غطّى أكثر) تبقى.
		// النسبة من الهدف (١٠٠٪ = وصل هدفه): الفني هدفه ١٣٠٪ من راتبه
		tc := c * 100 / targetOf(s.Group)
		s.RevenueTarget = targetOf(s.Group)
		if elapsedDays(from, to, now) < RevenueGraceDays && tc < 100 {
			s.RevenueGrace = true
			continue
		}
		f := *s.Final
		if tc < 100 {
			adj := f * (RevenueFloor + (1-RevenueFloor)*tc/100)
			s.RevenueAdj = adj - f
			f = adj
		} else {
			bonus := math.Min((tc-100)/10, RevenueBonusMax)
			s.RevenueAdj = math.Min(100, f+bonus) - f
			f = math.Min(100, f+bonus)
		}
		s.Final = &f
	}
}

// FmtIQD مبلغ بفواصل الآلاف: 1250000 → 1,250,000
func FmtIQD(v float64) string {
	s := strconv.FormatInt(int64(math.Round(v)), 10)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	if neg {
		return "-" + s
	}
	return s
}
