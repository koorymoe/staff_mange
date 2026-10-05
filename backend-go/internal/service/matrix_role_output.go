package service

import "staffmange-api/internal/repository"

// ═══ شغل كل دور — ماتركس يقيس كل موظف بشغل دوره هو ═══
//
// (ع): «ماتركس يحسب لكل موظف عدد الحجوزات الي طلعلها، وهذا غلط — عندي
// محاسب ومراقب ومدير. خصصنا لكل موظف عين حتى نوزّع الشغل ويكون الحساب صحيح».
//
// الحجوزات والطلعات والسرعة والجزئي للميدانيين (الليدرية والفنيين) بس.
// الباقين ينقاسون بجداول دورهم: فواتير، أحكام صندوق، حجوزات مسجّلة…
// كلها أرقام محسوبة من الجداول، بلا تخمين ولا عقوبة.

// IsFieldGroup المجموعات الي شغلها حجوزات وطلعات.
func IsFieldGroup(group string) bool { return group == "LEADERS" || group == "TECHS" }

type RoleMetric struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

type RoleOutput struct {
	Title   string       `json:"title"`   // «📈 شغل المحاسبة»
	Metrics []RoleMetric `json:"metrics"` // منجز بآخر N يوم
	Done    int          `json:"done"`    // مجموع الشغل الأساسي للدور
	Pending int          `json:"pending"` // بالانتظار الآن (طابور الدور)
	Lines   []ReportLine `json:"lines"`
}

var roleTitle = map[string]string{
	"FINANCE": "📈 شغل المحاسبة", "MONITORS": "📈 شغل المراقبة", "COORDINATORS": "📈 شغل التنسيق",
	"QUALITY": "📈 شغل الجودة", "DESIGN": "📈 شغل التصميم", "IT": "📈 شغل الدعم الفني",
	"ADMINS": "📈 الإدارة", "STAFF": "📈 شغل التقنيين", "MEDIA": "📈 شغل الإعلام",
	"SALES": "📈 شغل المبيعات", "PROJECTS": "📈 شغل المشاريع", "GPS": "📈 شغل الجي بي اس",
}

// roleOutput شغل الموظف المكتبي بآخر days يوم + طابور دوره الحالي.
func (s *MatrixAutopilotService) roleOutput(subj repository.WatchSubject, days int) RoleOutput {
	g := WatchGroup(subj)
	c := s.actions.RoleCounts(subj.ID, days)
	out := RoleOutput{Title: roleTitle[g], Metrics: []RoleMetric{}, Lines: []ReportLine{}}
	if out.Title == "" {
		out.Title = "📈 شغله بالنظام"
	}
	add := func(label string, v int) {
		out.Metrics = append(out.Metrics, RoleMetric{Label: label, Value: v})
		out.Done += v
	}
	switch g {
	case "FINANCE":
		add("فواتير اعتمدها", c.InvoicesApproved)
		add("فواتير دقّقها", c.InvoicesAudited)
	case "MONITORS":
		add("بنود حكم عليها", c.Reviews)
		add("فواتير دقّقها", c.InvoicesAudited)
	case "COORDINATORS":
		add("حجوزات سجّلها", c.BookingsCreated)
	case "QUALITY":
		add("متابعات جودة سوّاها", c.QualityDone)
	case "DESIGN":
		add("تصاميم رفعها", c.DesignUploads)
	case "SALES":
		add("حجوزات سجّلها", c.BookingsCreated)
	case "PROJECTS":
		add("مشاريع سجّلها أو مسؤول عنها", c.Projects)
	case "GPS":
		add("طلبات جي بي اس سجّلها", c.GpsRequests)
		add("اتصالات تجديد", c.GpsCalls)
	}
	// كل عملية حفظ بالنظام — المقياس الوحيد للأدوار بلا جدول شغل خاص.
	out.Metrics = append(out.Metrics, RoleMetric{Label: "عمليات بالنظام", Value: c.Actions})
	if len(out.Metrics) == 1 {
		out.Done = c.Actions
	}

	for _, w := range s.workload(subj) {
		out.Pending += w.Left
		if w.Left > 0 {
			out.Lines = append(out.Lines, rlLink(Fmt("بالانتظار هسه: %d — %s.", w.Left, w.Label), "INFO", w.Route))
		}
	}
	if len(out.Metrics) > 1 {
		line := Fmt("آخر %d يوم:", days)
		for i, m := range out.Metrics {
			if i > 0 {
				line += "،"
			}
			line += Fmt(" %s %d", m.Label, m.Value)
		}
		out.Lines = append([]ReportLine{rl(line+".", "OK")}, out.Lines...)
	} else {
		out.Lines = append([]ReportLine{rl(Fmt("آخر %d يوم: %d عملية بالنظام.", days, c.Actions), "INFO")}, out.Lines...)
	}
	if g == "MONITORS" && c.Reviews >= 10 {
		out.Lines = append(out.Lines, rl(Fmt("من أحكامه: %d٪ «عليه ملاحظة».", c.ReviewsFlagged*100/c.Reviews), "INFO"))
	}
	if c.Actions == 0 && g != "ADMINS" {
		out.Lines = append(out.Lines, rl(Fmt("ما سجّل أي عملية بالنظام بآخر %d يوم — تأكد يستخدم حسابه هو.", days), "WARN"))
	}
	return out
}

// rlLink سطر ويا رابط «ودّيني».
func rlLink(text, tone, link string) ReportLine {
	l := rl(text, tone)
	l.Link = link
	return l
}
