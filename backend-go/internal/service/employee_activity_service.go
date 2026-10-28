package service

import (
	"fmt"
	"strings"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ «شنو سوّى اليوم» — تقرير ماتركس عن الموظف ═══
//
// (ع): «دخلت على المراقب، هو يگول اليوم وقّفت حسابات الي طلعوا وعدّلت
// الملفات الي بيها خطأ — أريد يطلعلي شنو سوّى». مصدرين:
//  1. أفعال مسجّلة أصلاً بالجداول (اعتماد فاتورة، قرار إجازة، إيقاف حساب...)
//     — تشتغل حتى للأيام قبل سجل النشاط.
//  2. سجل النشاط (0305): كل حفظ ناجح بأي شاشة، يتوصف بالعربي.
// الي يغطيه المصدر الأول ما يتكرر من الثاني.

type ActivityLine struct {
	At    string `json:"at"` // HH:MM بغداد
	Text  string `json:"text"`
	Count int    `json:"count"`
	Area  string `json:"area"`
}

var knownText = map[string]struct{ f, area string }{
	"INV_APPROVE":     {"اعتمد فاتورة %s", "الفواتير"},
	"INV_AUDIT":       {"دقّق فاتورة %s", "الفواتير"},
	"INV_MONITOR":     {"قرّر على فاتورة %s (مراقبة)", "الفواتير"},
	"INV_RETURN":      {"رجّع فاتورة %s للتصحيح", "الفواتير"},
	"INV_CREATE":      {"سجّل فاتورة %s", "الفواتير"},
	"MONITOR_REVIEW":  {"راجع بند بصندوق المراقب%s", "المراقبة"},
	"DELETE_DECIDE":   {"قرّر على طلب حذف الحجز %s", "الحجوزات"},
	"DELETE_REQUEST":  {"طلب حذف الحجز %s", "الحجوزات"},
	"LEAVE_DECIDE":    {"قرّر على إجازة %s", "الإجازات"},
	"SUSPEND":         {"أوقف حساب %s", "الموظفين"},
	"BOOKING_CREATE":  {"سجّل حجز %s", "الحجوزات"},
	"BOOKING_CONFIRM": {"ثبّت الحجز %s ويا الزبون", "الحجوزات"},
	"MATERIALS_READY": {"جهّز مواد الحجز %s", "الحجوزات"},
	"AUDIT_ISSUE":     {"رفع بلاغ تدقيق على الحجز %s", "التدقيق"},
	"QUALITY_CONTACT": {"اتصل بزبون متابعة جودة%s", "الجودة"},
}

// مسارات يغطيها المصدر الأول — ما نكررها من السجل.
var coveredPatterns = []string{
	"/api/leader-invoices/{id}/approve", "/api/leader-invoices/{id}/audit", "/api/leader-invoices/{id}/monitor-decide",
	"/api/leader-invoices/{id}/return", "POST /api/leader-invoices", "/api/monitor-reviews/{id}/decide",
	"/api/booking-delete-requests/{id}/decide", "/delete-request", "/api/leaves/{id}/decide",
	"POST /api/bookings", "/api/bookings/{id}/audit",
}

// مسارات ما تهم المدير (قراءة إشعار، دخول، سؤال ماتركس...).
var noisyPatterns = []string{"/api/notifications", "/api/auth/", "/api/ai/ask", "/api/privacy-policy", "/heartbeat", "/api/cart", "/api/matrix"}

var sectionAr = map[string]string{
	"bookings": "حجز", "employees": "موظف", "leader-invoices": "فاتورة", "inventory": "المخزن", "gps": "جي بي اس",
	"vehicles": "سيارة", "solar": "الطاقة الشمسية", "sim": "شريحة", "suppliers": "مورد", "funds": "الدوار",
	"exhibitions": "معرض", "complaints": "شكوى", "services": "خدمة", "projects": "مشروع", "extra-tasks": "مهمة إضافية",
	"vehicle-missions": "مهمة سيارة", "training-programs": "برنامج تدريب", "training": "تدريب", "service-studies": "دراسة خدمة",
	"quality-follow-ups": "متابعة جودة", "product-requests": "طلب منتج", "network-cost": "كلفة شبكة", "leaves": "إجازة",
	"it": "جهاز آيتي", "vehicle-incidents": "عطل سيارة", "vehicle-bookings": "حجز سيارة", "quotations": "عرض سعر",
	"products": "منتج", "procurement": "طلب مواد", "kpi": "نقاط الأداء", "design-forms": "نموذج تصميم", "design-form": "نموذج تصميم",
	"attendance": "الحضور", "announcements": "إعلان", "customers": "زبون", "expenses": "مصروف", "missions": "مهمة ميدانية",
	"letters": "كتاب", "staff-requests": "طلب كادر", "discipline": "انضباط", "permissions": "صلاحيات", "departments": "قسم",
	"achievements": "إنجاز يومي", "materials": "مادة", "work-reports": "تقرير عمل", "monitor-reviews": "مراجعة مراقب",
	"booking-survey-reports": "تقرير كشف", "survey-photos": "صور كشف", "team-inventory": "جرد الفريق", "stories": "قصة",
	"device-maintenance": "صيانة جهاز", "vip-customers": "زبون مميز", "kpi-criteria": "معايير النقاط",
	// (ع) 10-16: «المدير رجال بسيط يشوف شي بسيط» — كل قسم بالعربي، ماكو إنكليزي
	"duplicate-candidates": "زبون مكرر", "justifications": "تبرير", "staff-ratings": "تقييم موظف", "crew-ratings": "تقييم كادر",
	"audit-issues": "بلاغ تدقيق", "booking-delete-requests": "طلب حذف حجز", "attendance-icon-requests": "طلب أيقونة حضور",
	"car-wash": "غسل سيارة", "checklists": "قائمة فحص", "company-location": "موقع الشركة", "coordination-alerts": "تنبيه تنسيق",
	"credits": "الآجل", "decor-rates": "أسعار الديكور", "department-heads": "مسؤول قسم", "design-assets": "ملف تصميم",
	"discipline-record": "سجل انضباط", "free-work-reasons": "سبب شغل مجاني", "internal-bookings": "شغل داخلي",
	"media": "مشروع تصوير", "monitor": "المراقبة", "monitor-desk": "مكتب المراقب", "peer": "صوت الموظفين",
	"performance-reviews": "مراجعة أداء", "procurement-quotes": "عرض سعر مورد", "procurement-standards": "معيار شراء",
	"procurement-policy": "سياسة الشراء", "product-procurements": "شراء منتج", "progress-reports": "تقرير مشروع يومي",
	"project-payments": "دفعة مشروع", "project-work-types": "نوع شغل مشروع", "quality": "الجودة", "quotation-terms": "شروط عرض السعر",
	"remote-hours": "ساعات العمل من البيت", "service-managers": "مسؤول خدمة", "service-prices": "سعر خدمة", "skills": "مهارة",
	"staff-points": "نقاط موظف", "store-tools": "عدة المخزن", "supplier-relations": "علاقة مورد", "system-price-catalog": "جدول الكلفة",
	"tech-showcase": "عرض تقني", "tech-suppliers": "مورد تقني", "training-center": "مركز التدريب", "vehicle-parts": "قطعة سيارة",
	"work-schedule": "جدول الدوام", "workplace-issues": "مشكلة وظيفية", "late-invoices": "فاتورة متأخرة", "project-quote-watch": "متابعة عرض سعر",
	"files": "ملف", "finance": "الحسابات", "owner": "إعدادات المالك", "security": "الأمان", "system": "إعدادات النظام",
	"location-pings": "الموقع", "geo": "الموقع", "locate": "الموقع", "presence": "التواجد",
}

// جمل خاصة لأفعال شايعة — أوضح من «عدّل زبون مكرر».
var phraseAr = map[string]string{
	"duplicate-candidates/dismiss": "قرّر إنه زبون مو مكرر", "duplicate-candidates/merge": "دمج زبونين مكررين",
	"duplicate-candidates/request-delete": "طلب حذف زبون مكرر", "POST duplicate-candidates": "فحص زبائن مكررين",
	"POST justifications": "طلب تبرير من موظف", "justifications/answer": "جاوب على طلب تبرير", "justifications/decide": "حكم على تبرير",
	"POST staff-ratings": "قيّم موظف", "POST crew-ratings": "قيّم كادر الحجز", "monitor-reviews/decide": "دقّق بند بصندوق المراقب",
	"work-reports/review": "دقّق تقرير عمل", "progress/review": "دقّق تقارير مشروع", "quality-follow-ups/monitor-verdict": "قيّم شغل مهندس الجودة",
}

var verbAr = map[string]string{
	"suspend": "أوقف حساب", "reactivate": "رجّع تفعيل حساب", "approve": "وافق على", "reject": "رفض", "decide": "قرّر على",
	"audit": "دقّق", "assign": "كلّف كادر لـ", "status": "غيّر حالة", "profile": "عدّل ملف", "identity": "عدّل اسم/دور",
	"all": "عدّل بيانات", "schedule": "عدّل دوام", "skills": "عدّل مهارات", "permissions": "عدّل صلاحيات",
	"waiting": "حط بالانتظار", "resume": "رجّع للطابور", "complete": "خلّص", "start": "بدأ", "stage": "حدّث مرحلة",
	"cancel": "ألغى", "stock": "جرد", "review": "قيّم", "return": "رجّع", "revoke": "ألغى", "archive": "أرشف",
	"restore": "رجّع من الأرشيف", "confirm": "ثبّت", "postpone": "أجّل", "reschedule": "غيّر موعد", "check-in": "سجّل حضور",
	"check-out": "سجّل انصراف", "materials-ready": "جهّز مواد", "close": "سكّر", "resolve": "حلّ",
}

type EmployeeActivityService struct {
	repo *repository.EmployeeActivityRepository
}

func NewEmployeeActivityService(r *repository.EmployeeActivityRepository) *EmployeeActivityService {
	return &EmployeeActivityService{repo: r}
}

func (s *EmployeeActivityService) Record(employeeID, method, pattern, path string, status int) {
	if pattern == "" {
		pattern = method + " " + path
	}
	for _, n := range noisyPatterns {
		if strings.Contains(pattern, n) {
			return
		}
	}
	s.repo.Record(employeeID, method, pattern, path, status)
}

// entityName يسمّي الشي حسب القسم: الموظف باسمه، والحجز بكوده...
func (s *EmployeeActivityService) entityName(section, id string) string {
	switch section {
	case "employees":
		return s.repo.Lookup("Employee", "name", id)
	case "bookings":
		return s.repo.Lookup("Booking", "code", id)
	case "vehicles":
		return s.repo.Lookup("Vehicle", "name", id)
	case "customers":
		return s.repo.Lookup("Customer", "name", id)
	case "leader-invoices":
		return s.repo.Lookup("LeaderInvoice", `"customerName"`, id)
	}
	return ""
}

func (s *EmployeeActivityService) describe(r repository.ActivityRow) (text, area string) {
	pat := strings.TrimPrefix(r.Pattern, r.Method+" ")
	segs := strings.Split(strings.Trim(strings.TrimPrefix(pat, "/api/"), "/"), "/")
	real := strings.Split(strings.Trim(strings.TrimPrefix(r.Path, "/api/"), "/"), "/")
	section := segs[0]
	noun := sectionAr[section]
	if noun == "" {
		// ما نعرض كلمة إنكليزية للمدير أبداً
		noun = "شغلة بالنظام"
	}
	name := ""
	if len(segs) > 1 && segs[1] == "{id}" && len(real) > 1 {
		name = s.entityName(section, real[1])
	}
	last := segs[len(segs)-1]
	// جملة جاهزة إذا الفعل معروف (مثلاً «دمج زبونين مكررين»)
	for _, k := range []string{section + "/" + last, r.Method + " " + section, last + "/" + last} {
		if p, ok := phraseAr[k]; ok && (k != section+"/"+last || len(segs) > 1) {
			if strings.HasSuffix(pat, "/progress/review") {
				p = phraseAr["progress/review"]
			}
			if name != "" {
				p += " «" + name + "»"
			}
			return p, noun
		}
	}
	if strings.HasSuffix(pat, "/progress/review") {
		return phraseAr["progress/review"], "تقرير مشروع يومي"
	}
	verb := verbAr[last]
	if verb == "" {
		switch r.Method {
		case "POST":
			verb = "أضاف"
		case "DELETE":
			verb = "حذف"
		default:
			verb = "عدّل"
		}
	}
	text = verb + " " + noun
	if name != "" {
		text += " «" + name + "»"
	}
	return text, noun
}

func (s *EmployeeActivityService) Day(employeeID, day string) []ActivityLine {
	type item struct {
		at         time.Time
		text, area string
	}
	items := []item{}
	if known, err := s.repo.Known(employeeID, day); err == nil {
		for _, k := range known {
			t := knownText[k.Kind]
			ref := k.Ref
			if ref != "" && strings.Contains(t.f, "%s") && !strings.HasPrefix(t.f, "راجع") && !strings.HasPrefix(t.f, "اتصل") {
				ref = "«" + ref + "»"
			}
			items = append(items, item{k.At, strings.TrimSpace(fmt.Sprintf(t.f, ref)), t.area})
		}
	}
	if logged, err := s.repo.Logged(employeeID, day); err == nil {
	next:
		for _, r := range logged {
			for _, c := range coveredPatterns {
				if strings.Contains(r.Pattern, c) {
					continue next
				}
			}
			t, a := s.describe(r)
			items = append(items, item{r.At, t, a})
		}
	}
	// ترتيب زمني، ونفس الفعل المتتالي يتجمّع («عدّل ملف ×٣»).
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j].at.Before(items[j-1].at); j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
	out := []ActivityLine{}
	for _, it := range items {
		if n := len(out); n > 0 && out[n-1].Text == it.text {
			out[n-1].Count++
			continue
		}
		out = append(out, ActivityLine{At: it.at.In(debriefLoc).Format("15:04"), Text: it.text, Count: 1, Area: it.area})
	}
	return out
}
