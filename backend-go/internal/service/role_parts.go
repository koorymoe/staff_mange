package service

import (
	"fmt"
	"math"

	"staffmange-api/internal/repository"
)

// ═══ التقييم الجديد لبقية الأدوار (دازاي 11-19 «ابدي بيهم كلهم بنفس السياق») ═══
// نفس فكرة الفني والليدر: بنود بأوزان (من ٨٠) + رأي ماتركس ٢٠. كل بند يجي من:
//   Rules  — نقاط ماتركس لهالقواعد (مكسوب ÷ أقصى)
//   Stages — معدّل نجوم هالمقيّمين
//   Rel    — جزء من الاعتمادية (حضور، مهام، زملاء)
//   Fact   — رقم خاص (تحويل البيّاع)
// البند الي ماله بيانات وزنه يتوزع. الأوزان بـdocs/التقييم-الجديد.md القسم ٣.

type partDef struct {
	Key, Label string
	Weight     float64
	Rules      []string
	Stages     []string
	Rel        string
	Fact       string
}

var monitorStages = []string{"MONITOR_PERIODIC", "MONITOR_DAILY"}

func common(att, tasks, peers float64) []partDef {
	out := []partDef{{Key: "ATTENDANCE", Label: "الحضور", Weight: att, Rel: "ATTENDANCE"}, {Key: "TASK_DUE", Label: "المهام الإضافية", Weight: tasks, Rel: "TASK_DUE"}}
	if peers > 0 {
		out = append(out, partDef{Key: "PEERS", Label: "الزملاء (المشاكل الوظيفية)", Weight: peers, Rel: "PEERS"})
	}
	return out
}

var genericParts = map[string][]partDef{
	// (دازاي 11-21) المبيعات: «يسجلون حجز ويسجلون شكوى، وكم حجز سجلوه تم واشتغلنا للزبون،
	// والحجز الي تحول لمشروع ينحسب الهم».
	GroupSales: append([]partDef{
		{Key: "CREATE", Label: "تسجيل الحجوزات كاملة وبوقتها", Weight: 30, Rules: []string{"CREATE"}},
		{Key: "CONVERSION", Label: "حجوزاته الي تمت (أو صارت مشروع) مو انلغت", Weight: 25, Fact: "CONVERSION"},
		{Key: "COMPLAINT_LOG", Label: "تسجيل الشكاوى كاملة (مربوطة بالحجز وشرح واضح)", Weight: 10, Rules: []string{"COMPLAINT_LOG"}},
		{Key: "SUPERVISOR", Label: "تقييم المراقب", Weight: 8, Stages: monitorStages},
	}, common(5, 1, 1)...),
	// (11-21) + تقييمه اليومي للموظفين بوقته (واجب التقييم)
	GroupMonitors: append([]partDef{
		{Key: "VERDICT", Label: "حسم بنود التدقيق بسرعة", Weight: 25, Rules: []string{"MONITOR"}},
		{Key: "AUDIT", Label: "تدقيق غير الحجوزات (التغطية)", Weight: 20, Rules: []string{"MONITOR_AUDIT"}},
		{Key: "RATING_DUTY", Label: "يقيّم الموظفين يومياً بوقته", Weight: 10, Rules: []string{"RATING_DUTY"}},
		{Key: "SUPERVISOR", Label: "تقييم المدير/المالك", Weight: 15, Stages: []string{"ADMIN_PERIODIC"}},
	}, common(5, 3, 2)...),
	GroupFinance: append([]partDef{
		{Key: "ACCOUNT", Label: "تدقيق واعتماد الفواتير خلال يوم", Weight: 45, Rules: []string{"ACCOUNT"}},
		{Key: "SUPERVISOR", Label: "تقييم المراقب", Weight: 15, Stages: monitorStages},
	}, common(10, 5, 5)...),
	// (دازاي 11-21) الجودة: يتصل بكل زبون خلص حجزه بنفس اليوم، والشكوى يتصل عليها بسرعة ويحلها بأسرع وقت.
	GroupQuality: append([]partDef{
		{Key: "QUALITY", Label: "الاتصال بالزبون خلال يوم من الإنجاز", Weight: 30, Rules: []string{"QUALITY"}},
		{Key: "COMPLAINT_CONTACT", Label: "الاتصال بصاحب الشكوى بسرعة", Weight: 15, Rules: []string{"COMPLAINT_CONTACT"}},
		{Key: "COMPLAINT_RESOLVE", Label: "حل الشكوى بأسرع وقت", Weight: 15, Rules: []string{"COMPLAINT_RESOLVE"}},
		{Key: "SUPERVISOR", Label: "تقييم المراقب", Weight: 10, Stages: monitorStages},
	}, common(7, 2, 1)...),
	GroupProcurement: append([]partDef{
		{Key: "REQUESTS", Label: "الرد على طلبات المواد والأدوات", Weight: 25, Rules: []string{"PROC_RESPONSE", "TOOL_RESPONSE"}},
		{Key: "SHORTAGES", Label: "حل نواقص الجرد", Weight: 20, Rules: []string{"SHORTAGE_RESOLVE"}},
		{Key: "VEHICLES", Label: "متابعة السيارات اليومية", Weight: 20, Rules: []string{"VEHICLE_DAILY"}},
		{Key: "SUPERVISOR", Label: "تقييم المراقب", Weight: 10, Stages: monitorStages},
	}, common(3, 2, 0)...),
	GroupTechnical: append([]partDef{
		{Key: "TECH_STEPS", Label: "محطات التقني: الاتصال، القرار، الزيارة، الكشف", Weight: 25, Rules: []string{"TECH_CONTACT", "TECH_DECIDE", "TECH_VISIT", "TECH_DIAG"}},
		{Key: "TECH_FIX", Label: "المعالجة بوقتها", Weight: 15, Rules: []string{"TECH_FIX"}},
		{Key: "COORD", Label: "التنسيق (تواصل، تثبيت، كادر، ترحيل)", Weight: 10, Rules: []string{"CONTACT", "CONFIRM", "CREW", "HANDOFF"}},
		{Key: "SUPERVISOR", Label: "تقييم المراقب", Weight: 15, Stages: monitorStages},
	}, common(10, 3, 2)...),
	GroupDesign: append([]partDef{
		{Key: "DESIGN", Label: "تسليم أقسام التصاميم بموعدها", Weight: 40, Rules: []string{"DESIGN_SECTION"}},
		{Key: "SUPERVISOR", Label: "تقييم المراقب", Weight: 20, Stages: monitorStages},
	}, common(10, 5, 5)...),
	GroupProjects: append([]partDef{
		{Key: "PROJECTS", Label: "حركة المشروع كل أسبوع", Weight: 40, Rules: []string{"PROJECT_WEEK"}},
		{Key: "SUPERVISOR", Label: "تقييم المراقب", Weight: 20, Stages: monitorStages},
	}, common(10, 5, 5)...),
	// الدعم التقني وبقية الأقسام (GPS، الإعلام، الغسل، تطوير الخدمات، العلاقات العامة): ماكو شغل
	// خاص بالنظام ينقاس — تقريره اليومي وصدقه ومهامه.
	GroupIT: append([]partDef{
		{Key: "REPORT", Label: "تقرير الإنجاز اليومي وصدقه", Weight: 30, Rules: []string{"DAILY_REPORT", "REPORT_MATCH"}},
		{Key: "SUPERVISOR", Label: "تقييم المراقب", Weight: 15, Stages: monitorStages},
	}, common(10, 20, 5)...),
	GroupOthers: append([]partDef{
		{Key: "REPORT", Label: "تقرير الإنجاز اليومي وصدقه", Weight: 30, Rules: []string{"DAILY_REPORT", "REPORT_MATCH"}},
		{Key: "SUPERVISOR", Label: "تقييم المراقب", Weight: 15, Stages: monitorStages},
	}, common(10, 20, 5)...),
}

// projectsPart (دازاي 11-21) «اي موظف عنده مشاريع يتقيّم عليهن… ماكو تأخير»: بند لكل الأدوار،
// يبين بس للي عنده مشروع (البند الي ماله بيانات وزنه يتوزع).
var projectsPart = partDef{Key: "PROJECTS", Label: "المشاريع الموجّهة إله (بلا تأخير)", Weight: 15, Rules: []string{"PROJECT_WEEK"}}

func init() {
	for g, defs := range genericParts {
		has := false
		for _, d := range defs {
			has = has || d.Key == projectsPart.Key
		}
		if !has {
			genericParts[g] = append(defs, projectsPart)
			defs = genericParts[g]
		}
		w := make([]roleWeight, 0, len(defs))
		for _, d := range defs {
			w = append(w, roleWeight{d.Key, d.Label, d.Weight})
		}
		roleWeightsOf[g] = w
	}
	// (11-19) البيّاع هدفه ٣ أضعاف راتبه، والتقني ١٣٠٪ مثل الفني
	revenueTarget[GroupSales] = 300
	revenueTarget[GroupTechnical] = 130
}

// ConversionPct البيّاع: حجوزاته الي انجزت ÷ (الي انجزت + الي انلغت). المعلّقة ما تنحسب بعد.
func ConversionPct(completed, cancelled int) (float64, bool) {
	if completed+cancelled == 0 {
		return 0, false
	}
	return float64(completed) * 100 / float64(completed+cancelled), true
}

// genericPct يحسب بنود دور من التعريف.
func genericPct(defs []partDef, s *StaffScore, stars func(string, ...string) (float64, string, bool), cf map[string]repository.CoordFact,
	set func(k string, v float64, d string)) {
	rel := map[string]ReliabilityPart{}
	for _, p := range s.RelParts {
		rel[p.Key] = p
	}
	for _, d := range defs {
		switch {
		case d.Rel != "":
			if p, ok := rel[d.Rel]; ok {
				set(d.Key, p.Pct, p.Detail)
			}
		case len(d.Stages) > 0:
			if v, det, ok := stars(s.ID, d.Stages...); ok {
				set(d.Key, v, det)
			}
		case d.Fact == "CONVERSION":
			if f, ok := cf[s.ID]; ok {
				if v, ok := ConversionPct(f.Completed+f.Projects, f.Cancelled); ok {
					d := fmt.Sprintf("%d تم، %d انلغى", f.Completed, f.Cancelled)
					if f.Projects > 0 {
						d = fmt.Sprintf("%d تم، %d صار مشروع، %d انلغى", f.Completed, f.Projects, f.Cancelled)
					}
					set("CONVERSION", v, d)
				}
			}
		case len(d.Rules) > 0:
			var sum [2]int
			for _, r := range d.Rules {
				v := s.ruleTotals[r]
				sum[0], sum[1] = sum[0]+v[0], sum[1]+v[1]
			}
			if v, ok := ratio(sum); ok {
				set(d.Key, v, fmt.Sprintf("%d من %d نقطة", sum[0], sum[1]))
			}
		}
	}
}

// genericOpinion رأي ماتركس لبقية الأدوار: كل بند شغل تحت ٧٠٪ (والحضور تحت ٩٠٪) ينقّص بقدر وزنه.
func genericOpinion(defs []partDef, pct map[string]float64) (float64, []string) {
	v := 100.0
	notes := []string{}
	for _, d := range defs {
		x, ok := pct[d.Key]
		if !ok {
			continue
		}
		limit := 70.0
		if d.Key == "ATTENDANCE" {
			limit = 90
		}
		if x < limit {
			cut := math.Min(40, d.Weight*1.25)
			// (دازاي 11-21، المحاكاة) الغياب والمهام الفايتة ينقّصن حد أدنى ثابت مهما چان وزنهن
			// بالدور — قبل چان المراقب الغايب يومين ياخذ رأي ٨٨٪ لأن وزن الحضور عنده ٥ بس.
			switch {
			case d.Key == "ATTENDANCE" && x < 80:
				cut = math.Max(cut, 20)
			case d.Key == "TASK_DUE":
				cut = math.Max(cut, 15)
			}
			v -= cut
			notes = append(notes, "يحتاج يحسّن: "+d.Label)
		}
	}
	return math.Max(0, v), notes
}
