package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ChatTurn دور واحد بمحادثة سابقة (سؤال الموظف أو جواب المساعد) — نرسلها
// للذكاء الاصطناعي كسياق حتى المحادثة تكون مترابطة وليست أسئلة منفصلة.
type ChatTurn struct {
	Role string `json:"role"` // "user" أو "assistant"
	Text string `json:"text"`
}

// AssistantService يوصل مساعد ذكي (Gemini المجاني من Google) بمعلومات الموظف
// نفسه بس (راتبه، مهاراته، تقييماته) — يجاوب أسئلة شخصية بأسلوب طبيعي بدون
// ما يكشف بيانات موظفين ثانيين. مصمم على حد استخدام يومي (GEMINI_DAILY_CAP)
// حتى نضل بالخطة المجانية ولا نتفاجئ بفاتورة.
type AssistantService struct {
	apiKey        string
	dailyCap      int
	employees     *repository.EmployeeRepository
	kpi           *repository.KpiRepository
	perfReviews   *repository.PerformanceReviewRepository
	bookings      *repository.BookingRepository
	missions      *repository.MissionRepository
	expenses      *repository.ExpenseRepository
	gps           *repository.GpsRepository
	qualityFUs    *repository.QualityFollowUpRepository
	complaints    *repository.ComplaintRepository
	conversations *repository.AssistantConversationRepository
	knowledge     *repository.AssistantKnowledgeRepository

	// roleGuides: خريطة "الدور الوظيفي → نص دليل الاستخدام" — تُحمَّل مرة وحدة عند
	// إنشاء الخدمة (مو بكل طلب) حتى ما نعيد قراءة الملفات وتنظيف الـHTML بكل رسالة.
	roleGuides map[string]string

	mu        sync.Mutex
	usedToday int
	resetDate string
}

// roleGuideFiles تربط كل دور وظيفي بملف دليله (بدون امتداد) — الأدوار التسعة
// القديمة لها ملفات HTML مصممة (tutorial-*.html)، والأدوار الخمسة الجديدة
// (اللي ماكانت موثقة سابقاً) لها ملفات نصية بسيطة (guide-*.txt) كتبناها اعتماداً
// على قراءة فعلية للكود عشان ما نختلق ميزات غير موجودة.
var roleGuideFiles = map[string]string{
	"ADMIN":             "tutorial-admin.html",
	"FINANCE":           "tutorial-finance.html",
	"GPS_ADMIN":         "tutorial-gps-admin.html",
	"HR_COORDINATOR":    "tutorial-hr.html",
	"MONITOR":           "tutorial-monitor.html",
	"OWNER":             "tutorial-owner.html",
	"QUALITY_ENGINEER":  "tutorial-quality.html",
	"SALES":             "tutorial-sales.html",
	"TECHNICIAN":        "tutorial-technician.html",
	"ENGINEER":          "guide-engineer.txt",
	"PROJECT_MANAGER":   "guide-project_manager.txt",
	"PROCUREMENT_ADMIN": "guide-procurement_admin.txt",
	"DESIGNER":          "guide-designer.txt",
	"SERVICE_MANAGER":   "guide-service_manager.txt",
}

// htmlTagRe و htmlWhitespaceRe يشيلون وسوم الـHTML والمسافات الزايدة من أدلة
// الاستخدام القديمة (tutorial-*.html) حتى نبعت للموديل نص عربي مقروء بدل
// ماركب HTML يبذّر توكنز بلا فايدة. تنظيف بسيط بالـregex كافي هنا (مو محتاج
// parser حقيقي) لأنه الهدف نص مفهوم للذكاء الاصطناعي مو عرض دقيق بالمتصفح.
var (
	htmlTagRe        = regexp.MustCompile(`(?is)<(script|style)[^>]*>.*?</(script|style)>|<[^>]+>`)
	htmlEntityRe     = regexp.MustCompile(`&[a-zA-Z#0-9]+;`)
	htmlWhitespaceRe = regexp.MustCompile(`[ \t]+`)
	htmlBlankLinesRe = regexp.MustCompile(`\n{3,}`)
)

func stripHTML(s string) string {
	s = htmlTagRe.ReplaceAllString(s, "\n")
	s = htmlEntityRe.ReplaceAllString(s, " ")
	s = htmlWhitespaceRe.ReplaceAllString(s, " ")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSpace(l)
	}
	s = strings.Join(lines, "\n")
	s = htmlBlankLinesRe.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

// loadRoleGuides يقرأ كل ملفات أدلة الاستخدام (HTML القديمة + النصية الجديدة)
// من القرص مرة وحدة، وينضّفها، ويرجّع خريطة جاهزة للاستخدام بالسياق. إذا ملف
// مفقود أو المجلد مو موجود (مثلاً بيئة اختبار) نتجاهله بهدوء بدل ما نكسر تشغيل
// الخدمة كلها — المساعد بس يشتغل بدون دليل ذاك الدور.
func loadRoleGuides(tutorialsDir string) map[string]string {
	out := make(map[string]string, len(roleGuideFiles))
	for role, filename := range roleGuideFiles {
		path := filepath.Join(tutorialsDir, filename)
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		text := string(raw)
		if strings.HasSuffix(filename, ".html") {
			text = stripHTML(text)
		} else {
			text = strings.TrimSpace(text)
		}
		if text != "" {
			out[role] = text
		}
	}
	return out
}

func NewAssistantService(
	apiKey string, dailyCap int,
	employees *repository.EmployeeRepository,
	kpi *repository.KpiRepository,
	perfReviews *repository.PerformanceReviewRepository,
	bookings *repository.BookingRepository,
	missions *repository.MissionRepository,
	expenses *repository.ExpenseRepository,
	gps *repository.GpsRepository,
	qualityFUs *repository.QualityFollowUpRepository,
	complaints *repository.ComplaintRepository,
	conversations *repository.AssistantConversationRepository,
	knowledge *repository.AssistantKnowledgeRepository,
	tutorialsDir string,
) *AssistantService {
	return &AssistantService{
		apiKey: apiKey, dailyCap: dailyCap, employees: employees, kpi: kpi, perfReviews: perfReviews,
		bookings: bookings, missions: missions, expenses: expenses, gps: gps, qualityFUs: qualityFUs, complaints: complaints,
		conversations: conversations,
		knowledge:     knowledge,
		roleGuides:    loadRoleGuides(tutorialsDir),
		resetDate:     time.Now().Format("2006-01-02"),
	}
}

var ErrAssistantNotConfigured = errors.New("المساعد الذكي غير مفعّل حالياً — لازم تضاف مفتاح Gemini بإعدادات السيرفر")
var ErrAssistantDailyCapReached = errors.New("وصلنا للحد الأقصى المجاني لعدد أسئلة اليوم — جرب باچر")

func (s *AssistantService) checkAndIncrementQuota() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	today := time.Now().Format("2006-01-02")
	if today != s.resetDate {
		s.resetDate = today
		s.usedToday = 0
	}
	if s.usedToday >= s.dailyCap {
		return ErrAssistantDailyCapReached
	}
	s.usedToday++
	return nil
}

// Ask يبني سياق خاص بالموظف السائل بس (بياناته الشخصية)، ويرسله مع سؤاله
// لـ Gemini، ويرجّع الجواب النصي.
func (s *AssistantService) Ask(employeeID, message string) (string, error) {
	if s.apiKey == "" {
		return "", ErrAssistantNotConfigured
	}
	if err := s.checkAndIncrementQuota(); err != nil {
		return "", err
	}

	employee, err := s.employees.FindByID(employeeID)
	if err != nil || employee == nil {
		return "", errors.New("تعذر تحديد هوية الموظف")
	}
	kpiEvals, _ := s.kpi.ListForEmployee(employeeID)

	skillsKnown := 0
	skillsTotal := len(employee.Skills)
	for _, sk := range employee.Skills {
		if sk.CanPerform {
			skillsKnown++
		}
	}

	kpiPointsThisMonth := 0
	for _, ev := range kpiEvals {
		if !ev.Cancelled && ev.CreatedAt.Month() == time.Now().Month() {
			kpiPointsThisMonth += ev.Points
		}
	}

	salaryLine := "الراتب غير مسجل بالنظام"
	if employee.Salary != nil {
		salaryLine = fmt.Sprintf("راتبه الشهري: %.0f دينار عراقي", *employee.Salary)
	}

	domainBlock := s.buildRoleScopedBlock(employee)
	guideBlock := s.roleGuides[employee.Role]
	if guideBlock == "" {
		guideBlock = "(ماكو دليل استخدام مفصل لهذا الدور محمّل حالياً — اعتمد على معرفتك العامة بالنظام من بيانات شغله أعلاه بس، ولا تختلق أسماء صفحات أو أزرار غير مذكورة)"
	}

	knowledgeBlock := s.buildKnowledgeBlock(message)

	systemContext := buildSystemPrompt(employee.Name, employee.Role, salaryLine, skillsKnown, skillsTotal, kpiPointsThisMonth, domainBlock, guideBlock, knowledgeBlock, message)

	rawReply, err := s.callGeminiWithTools(systemContext, true)
	if err != nil {
		return "", err
	}

	reply, learned := extractLearnedKnowledge(rawReply)

	// نسجل كل تبادل رسالة/جواب حتى يقدر المالك يراجع محادثات كل الموظفين مع
	// المساعد لاحقاً. فشل التسجيل ما يوقف الرد للموظف — بس نطبع خطأ باللوغ.
	if s.conversations != nil {
		if logErr := s.conversations.Create(employeeID, message, reply); logErr != nil {
			log.Printf("assistant: تعذر تسجيل المحادثة للموظف %s: %v", employeeID, logErr)
		}
	}

	// نخزن أي معرفة استخرجها المساعد من هذا التبادل (بدون طلب إضافي لـGemini —
	// نستخدم نفس الرد). فشل التخزين ما يوقف الرد للموظف — بس نطبع خطأ باللوغ.
	if s.knowledge != nil {
		for _, item := range learned {
			if err := s.knowledge.Create(item.Topic, item.Content, employeeID); err != nil {
				log.Printf("assistant: تعذر تخزين معرفة متعلّمة للموظف %s: %v", employeeID, err)
			}
		}
	}

	return reply, nil
}

// learnedItem سطر معرفة وحد استخرجناه من علامة <<LEARN>> برد الموديل.
type learnedItem struct {
	Topic   string
	Content string
}

// learnMarkerRe يطابق علامة <<LEARN topic="...">>محتوى<<END_LEARN>> اللي
// نطلب من الموديل يضيفها بآخر رده لو تعلم شي يستاهل يتذكره — راجع تعليمات
// نص البرومبت بـbuildSystemPrompt لشرح الصيغة كاملة.
var learnMarkerRe = regexp.MustCompile(`(?s)\n?<<LEARN topic="([^"]*)">>(.*?)<<END_LEARN>>`)

// extractLearnedKnowledge يشيل كل علامات <<LEARN>>...<<END_LEARN>> من رد الموديل
// (حتى الموظف ما يشوفها بالرد المعروض له) ويرجّع النص النظيف مع قائمة سطور
// المعرفة المستخرجة لتخزينها. رد بدون أي علامة يرجع كما هو بدون تغيير.
func extractLearnedKnowledge(reply string) (string, []learnedItem) {
	matches := learnMarkerRe.FindAllStringSubmatch(reply, -1)
	if len(matches) == 0 {
		return reply, nil
	}
	items := make([]learnedItem, 0, len(matches))
	for _, m := range matches {
		topic := strings.TrimSpace(m[1])
		content := strings.TrimSpace(m[2])
		if topic == "" || content == "" {
			continue
		}
		items = append(items, learnedItem{Topic: topic, Content: content})
	}
	cleaned := strings.TrimSpace(learnMarkerRe.ReplaceAllString(reply, ""))
	return cleaned, items
}

// buildKnowledgeBlock يستخرج كلمات مفتاحية من رسالة الموظف ويبحث بجدول
// AssistantKnowledge عن سطور معرفة سابقة تقاطع معها (RAG بسيط) — هذا اللي
// يخلي المساعد "يتذكر" معلومات تعلمها من محادثات أو بحث سابق بدل ما ينساها.
func (s *AssistantService) buildKnowledgeBlock(message string) string {
	if s.knowledge == nil {
		return "(ماكو معرفة سابقة مخزّنة ذات علاقة)"
	}
	keywords := repository.ExtractKeywords(message)
	if len(keywords) == 0 {
		return "(ماكو كلمات مفتاحية كافية بالسؤال للبحث بالمعرفة السابقة)"
	}
	rows, err := s.knowledge.SearchRelevant(keywords, 8)
	if err != nil || len(rows) == 0 {
		return "(ماكو معرفة سابقة مخزّنة ذات علاقة)"
	}
	var b bytes.Buffer
	for _, r := range rows {
		fmt.Fprintf(&b, "- [%s]: %s\n", r.Topic, r.Content)
	}
	return b.String()
}

// buildSystemPrompt يبني نص البرومبت الكامل اللي نبعته لـGemini لكل سؤال موظف عادي.
// مستخرجة كدالة مستقلة (بدل ما تكون مطبوخة جوة Ask) حتى نقدر نختبرها مباشرة
// بدون ما نحتاج نتصل فعلياً بـGemini أو نجهز قاعدة بيانات — يكفي نمرر قيم جاهزة
// ونتأكد النص المبني فيه كل العناصر المطلوبة (تعليمات اللهجة، الاختصار، الأمثلة،
// ومحتوى دليل الدور).
func buildSystemPrompt(name, role, salaryLine string, skillsKnown, skillsTotal, kpiPointsThisMonth int, domainBlock, guideBlock, knowledgeBlock, message string) string {
	return fmt.Sprintf(`أنت مساعد ذكي داخلي لنظام إدارة موظفين شركة الأماني (شركة تشتغل بمجالات منها الطاقة الشمسية). تحجي مع الموظف مثل زميل عراقي متعاون بالشغل، مو مثل بوت رسمي أو خدمة عملاء.

نطاق حجيك — مهم توضحه لنفسك: انت مو مقيّد بس بأسئلة النظام أو الشغل. الموظف يقدر يسولف وياك عن أي موضوع — معلومات عامة، سوالف شخصية، مواضيع تخص شغل الشركة متل معدات وتركيب الطاقة الشمسية، أو أي شي يريد يسأل عنه أو يبحث عنه. عندك خاصية بحث حقيقي بجوجل مفعّلة — استخدمها بحرية أي وقت تحس المعلومة تفيد (سعر منتج، مواصفات جهاز، معلومة تقنية، خبر...). الشي الوحيد الممنوع قطعياً: تكشف بيانات إدارية أو سرية تخص موظفين ثانيين (رواتبهم، تقييمات أدائهم، تقارير مالية...) لموظف دوره ما يخوله يشوفها — وهذا أصلاً محقق تلقائياً لأن بيانات الشغل المعطاة لك أدناه هي بيانات هذا الموظف السائل بس. غير هيچ، ما عندك داعي ترفض أي سؤال أو تقول "أنا بس لأسئلة النظام" — هذا رد ممنوع تماماً لأي سؤال خارج الشغل، جاوب طبيعي وساعد.

ذاكرة تتعلم منها: إذا بمحادثتك الحالية تعلمت شي حقيقي يستاهل يتذكره لمحادثات جاية (معلومة عامة مفيدة، شي علّمك إياه الموظف، نتيجة بحث جوجل مفيدة) — ضيف بآخر ردك بالضبط هالصيغة (بدون أي نص إضافي حولها):
<<LEARN topic="عنوان قصير للموضوع">>الحقيقة أو المعلومة المختصرة اللي تريد تتذكرها<<END_LEARN>>
لا تضيف هذي العلامة إطلاقاً إذا ماكو شي حقيقي يستاهل التذكر (أغلب الردود العادية أو السمول توك ما تحتاجها) — لا تخزن بيانات شخصية خاصة بالموظف (راتبه، بياناته) ولا حچي عابر بلا فايدة. الموظف ما لازم يشوف هذي العلامة إطلاقاً بردك المقروء — بس هي جزء تقني بآخر النص.

قواعد الأسلوب — التزم فيها دايماً وبدون استثناء:
1. اللهجة العراقية إجبارية بكل رد — ممنوع العربية الفصحى الرسمية أو أسلوب "التقرير". احجي متل ما تحجي بالشارع أو بالمكتب مع زميلك.
2. الرد الافتراضي قصير — جملة وحدة أو جملتين إذا السؤال بسيط أو محادثة عادية. لا تكرر السؤال، لا تحشي كلام زايد، لا تضيف تحذيرات أو ملاحظات ماكو داعي إلها.
3. الأسلوب المرقّم بخطوات (1، 2، 3...) يجي بس إذا الموظف يطلب "تعليم" فعلي — يعني يسأل بصيغة "علّمني"، "شلون أسوي...؟"، "وريني الخطوات"، أو أي طلب واضح يبين يريد يتعلم عملية. حتى بهاي الحالة خلي الخطوات مختصرة وعملية، وسمّي الصفحات والأزرار بأسمائها الحقيقية المذكورة بدليل الاستخدام أدناه — لا تستخدم وصف عام مثل "روح لصفحة الطلبات"، قول اسمها بالضبط.
4. لا تختلق معلومات أو أسماء صفحات/أزرار غير موجودة ببيانات الموظف أو بدليل الاستخدام أدناه. إذا السؤال خارج بياناته الشخصية أو شغله، اعتذر بجملة وحدة وقول تراجع الإدارة.
5. اقرأ مزاج المحادثة. إذا الموظف واضح يريد يدردش أو يفشّ خلقه أو زهقان أو يحچي حچي عادي بلا علاقة بالشغل (مثلاً "زهقان اليوم"، "شلونك؟"، نكتة، كلام مسلي) — احجي وياه طبيعي ودافي متل صديق، اطلع بروح خفيفة وتفاعل معه، سولف شوي ولو لازم قول نكتة عراقية خفيفة. لا تحوّل الحچي هذا لسؤال شغل ولا ترجعه "أنا بس لأسئلة الشغل" — هذا رد بارد وممنوع. خلي الرد الودّي قصير برضو (سطر أو سطرين) إلا إذا الموقف يستاهل أكثر.

أمثلة على الأسلوب المطلوب (بس للتوضيح، مو نص جاهز تكرره):

مثال 1 — سؤال قصير عادي:
الموظف: شكد نقاطي هذا الشهر؟
المساعد: عدك [كذا] نقطة هذا الشهر، زين ماشي.

مثال 2 — دردشة عابرة:
الموظف: شلونك اليوم؟
المساعد: تمام الحمدلله، شغلتك جاهزة إذا عدك سؤال ✌️

مثال 3 — طلب تعليم (إداري كميات يسأل شلون يوفر طلب مادة):
الموظف: علمني شلون أوفر طلب مادة وصلني
المساعد: خلي أوريك:
1. افتح "طلبات المواد" من الإدارة ← إدارة المشتريات.
2. لكيت الطلب بالجدول واضغط عليه يفتحلك تفاصيل المواد.
3. إذا لسه بانتظار التوفير، اضغط "جاري التوفير" أول لو تريد تبلغ إنك شغال عليه.
4. لما توفر المواد فعلياً اضغط "تم التوفير"، علّم كل مادة وفّرتها، سجل السعر والتكلفة الكلية، واضغط "تأكيد التوفير".
خلص، الطلب ينكتب "تم التوفير" ويشوفه الموظف.

مثال 4 — موظف زهقان يريد يدردش (مو سؤال شغل):
الموظف: زهقان اليوم والله
المساعد: هواي؟ 😄 خلي أسولف وياك شوي إذا تريد، لو تحب نطلع بنكتة نفش خلقك — وإذا طلع سؤال شغل بالطريق أنا حاضر.

مثال 5 — سمول توك عابر:
الموظف: شكو ماكو، شنو أخبارك؟
المساعد: هلا وغلا، تمام الحال والحمدلله، بس اليوم الجو حر شوي 😅 شكو عندك، شغل لو بس مرور؟

مثال 6 — سؤال خارج نطاق النظام تماماً (تقني يخص مجال الشركة) — لازم تجاوب طبيعي بلا رفض:
الموظف: اني فني ألواح شمسية، شنو أحسن نوع إنفرتر للاستخدام المنزلي هسه؟
المساعد: أكو خيارات زينة هسه، الإنفرترات الهجينة صارت الأفضل للاستخدام المنزلي لأنها تدعم البطاريات وتتوافق مع الشبكة سوا، دور على ماركات موثوقة وشوف الضمان المحلي. تريد أبحثلك بمواصفات نوع معين؟
<<LEARN topic="إنفرترات الطاقة الشمسية">>الإنفرترات الهجينة (Hybrid) هي الأفضل للاستخدام المنزلي لأنها تدعم البطاريات وتعمل مع الشبكة الكهربائية سوا<<END_LEARN>>

بيانات الموظف السائل:
- الاسم: %s
- الدور الوظيفي: %s
- %s
- عدد المهارات المتقنة: %d من أصل %d
- مجموع نقاط الكي بي اي هذا الشهر: %d

بيانات شغله الحالية (حسب دوره الوظيفي):
%s

دليل استخدام النظام الخاص بدوره الوظيفي (اعتمد عليه بالضبط لو سألك عن خطوات استخدام النظام):
%s

معلومات تعلمتها من محادثات سابقة قد تفيدك هسه:
%s

سؤال الموظف: %s`, name, role, salaryLine, skillsKnown, skillsTotal, kpiPointsThisMonth, domainBlock, guideBlock, knowledgeBlock, message)
}

// buildRoleScopedBlock يبني كتلة بيانات "شغل الموظف" حسب دوره الوظيفي —
// كل دور يشوف بس المعلومات المرتبطة بشغله (حجوزات، طلبات GPS، مصاريف...)
// بدون ما يوصل لبيانات موظفين ثانيين أو مجالات خارج دوره.
func (s *AssistantService) buildRoleScopedBlock(employee *model.Employee) string {
	var b bytes.Buffer

	switch employee.Role {
	case "GPS_ADMIN":
		devices, _ := s.gps.ListDevices()
		pendingDevices := 0
		for _, d := range devices {
			if d.Status == "PENDING" {
				pendingDevices++
			}
		}
		renewals, _ := s.gps.ListRenewals()
		pendingRenewals := 0
		for _, rn := range renewals {
			if rn.Status == "PENDING" {
				pendingRenewals++
			}
		}
		maint, _ := s.gps.ListMaintenance()
		pendingMaint := 0
		for _, m := range maint {
			if m.Status == "PENDING" {
				pendingMaint++
			}
		}
		fmt.Fprintf(&b, "- طلبات أجهزة GPS المعلقة: %d\n- طلبات تجديد الاشتراك المعلقة: %d\n- طلبات الصيانة المفتوحة: %d\n", pendingDevices, pendingRenewals, pendingMaint)

	case "HR_COORDINATOR":
		pending, _ := s.bookings.List("PENDING", "", "")
		fmt.Fprintf(&b, "- عدد الحجوزات بانتظار التنسيق حالياً: %d\n", len(pending))
		shown := 0
		for _, bk := range pending {
			if shown >= 10 {
				break
			}
			customerName := "-"
			if bk.Customer != nil {
				customerName = bk.Customer.Name
			}
			fmt.Fprintf(&b, "  * %s — الزبون: %s\n", bk.Code, customerName)
			shown++
		}

	case "SALES":
		myBookings, _ := s.bookings.ListForAssignedEmployee(employee.ID, 15)
		fmt.Fprintf(&b, "- عدد حجوزاتي المسجلة: %d\n", len(myBookings))
		for _, bk := range myBookings {
			customerName := "-"
			if bk.Customer != nil {
				customerName = bk.Customer.Name
			}
			fmt.Fprintf(&b, "  * %s — الزبون: %s — الحالة: %s\n", bk.Code, customerName, bk.Status)
		}

	case "TECHNICIAN", "ENGINEER":
		missions, _ := s.missions.ListForEmployee(employee.ID)
		fmt.Fprintf(&b, "- عدد المهام المسندة إلي: %d\n", len(missions))
		shown := 0
		for _, m := range missions {
			if shown >= 10 {
				break
			}
			status := "منجزة"
			if m.CompletedAt == nil {
				status = "قيد التنفيذ"
			}
			fmt.Fprintf(&b, "  * مهمة %s — الحالة: %s\n", m.Code, status)
			shown++
		}

	case "FINANCE":
		allExpenses, _ := s.expenses.List("")
		pendingCount := 0
		for _, e := range allExpenses {
			if e.Status == "PENDING" {
				pendingCount++
			}
		}
		fmt.Fprintf(&b, "- عدد المصاريف بانتظار المراجعة حالياً: %d\n", pendingCount)

	case "QUALITY_ENGINEER":
		followUps, _ := s.qualityFUs.List()
		openFUs := 0
		for _, f := range followUps {
			if f.Status == "PENDING" {
				openFUs++
			}
		}
		complaints, _ := s.complaints.List()
		openComplaints := 0
		for _, c := range complaints {
			if c.Status != "RESOLVED" && c.Status != "CLOSED" {
				openComplaints++
			}
		}
		fmt.Fprintf(&b, "- متابعات جودة مفتوحة: %d\n- شكاوى مفتوحة غير محلولة: %d\n", openFUs, openComplaints)
	}

	if b.Len() == 0 {
		return "(ماكو بيانات شغل خاصة مرتبطة بهذا الدور الوظيفي حالياً)"
	}
	return b.String()
}

// buildEmployeeProfileBlock يبني كتلة نصية عربية تلخص كل بيانات موظف معيّن
// (مهارات، سجل كي بي اي، تقييمات أداء) — تستخدمها كل من GenerateEmployeeReport
// و ManagerChat حتى ما نكرر نفس منطق التجميع بمكانين.
func (s *AssistantService) buildEmployeeProfileBlock(employee *model.Employee) string {
	kpiEvals, _ := s.kpi.ListForEmployee(employee.ID)
	reviews, _ := s.perfReviews.ListForEmployee(employee.ID)

	skillLines := "لا توجد مهارات مسجلة"
	if len(employee.Skills) > 0 {
		var b bytes.Buffer
		for _, sk := range employee.Skills {
			status := "لا يتقنها"
			if sk.CanPerform {
				status = "يتقنها"
			}
			name := sk.SkillID
			if sk.Skill != nil {
				name = sk.Skill.Name
			}
			fmt.Fprintf(&b, "- %s: %s\n", name, status)
		}
		skillLines = b.String()
	}

	kpiLines := "لا توجد تقييمات كي بي اي مسجلة"
	if len(kpiEvals) > 0 {
		var b bytes.Buffer
		shown := 0
		for _, ev := range kpiEvals {
			if shown >= 15 {
				break
			}
			status := ""
			if ev.Cancelled {
				status = " (ملغاة)"
			}
			fmt.Fprintf(&b, "- %s: %d نقطة — %s%s (خصم %.0f د.ع)\n", ev.CreatedAt.Format("2006-01-02"), ev.Points, ev.Reason, status, ev.DeductionAmount)
			shown++
		}
		kpiLines = b.String()
	}

	reviewLines := "لا توجد تقييمات أداء (تدريب) مسجلة"
	if len(reviews) > 0 {
		var b bytes.Buffer
		for _, r := range reviews {
			fmt.Fprintf(&b, "- %s: %s — %s\n", r.CreatedAt.Format("2006-01-02"), r.Rating, r.Reason)
		}
		reviewLines = b.String()
	}

	return fmt.Sprintf(`بيانات الموظف:
- الاسم: %s
- الدور الوظيفي: %s
- الحالة: %s

المهارات:
%s

سجل تقييمات الكي بي اي (آخر 15):
%s

تقييمات الأداء (التدريب):
%s`, employee.Name, employee.Role, employee.Status, skillLines, kpiLines, reviewLines)
}

// ManagerChat محادثة حرة للمراقب/المدير — يقدر يسأل عن أي موظف بالاسم
// ويطلب تقرير أو تفاصيل، والمساعد يجاوب بالسياق نفسه بدون ما يحتاج يفتح
// صفحة جديدة لكل سؤال. نطابق اسم الموظف المذكور بالرسالة مع قائمة الموظفين
// ونجيب بياناته الكاملة إذا انطابق الاسم.
func (s *AssistantService) ManagerChat(askerEmployeeID, message string, history []ChatTurn) (string, error) {
	if s.apiKey == "" {
		return "", ErrAssistantNotConfigured
	}
	if err := s.checkAndIncrementQuota(); err != nil {
		return "", err
	}

	employees, _ := s.employees.List()

	var roster bytes.Buffer
	for _, e := range employees {
		fmt.Fprintf(&roster, "- %s (%s)\n", e.Name, e.Role)
	}

	var matchedBlocks bytes.Buffer
	matchedCount := 0
	matchedIDs := map[string]bool{}
	if asker, err := s.employees.FindByID(askerEmployeeID); err == nil && asker != nil {
		matchedBlocks.WriteString("(هذا هو السائل نفسه، إذا سأل عن نفسه بصيغة \"راتبي\"/\"مهاراتي\" استخدم هالبيانات)\n")
		matchedBlocks.WriteString(s.buildEmployeeProfileBlock(asker))
		matchedBlocks.WriteString("\n\n")
		matchedIDs[asker.ID] = true
		matchedCount++
	}
	for i := range employees {
		if matchedCount >= 4 {
			break
		}
		if matchedIDs[employees[i].ID] {
			continue
		}
		if employees[i].Name != "" && strings.Contains(message, employees[i].Name) {
			matchedBlocks.WriteString(s.buildEmployeeProfileBlock(&employees[i]))
			matchedBlocks.WriteString("\n\n")
			matchedIDs[employees[i].ID] = true
			matchedCount++
		}
	}

	var historyText bytes.Buffer
	start := 0
	if len(history) > 8 {
		start = len(history) - 8
	}
	for _, turn := range history[start:] {
		label := "الموظف/المدير"
		if turn.Role == "assistant" {
			label = "المساعد"
		}
		fmt.Fprintf(&historyText, "%s: %s\n", label, turn.Text)
	}

	matchedText := matchedBlocks.String()
	if matchedText == "" {
		matchedText = "(ما انذكر اسم موظف مطابق بالرسالة الحالية — إذا احتجت بيانات موظف معيّن اطلب منه يذكر الاسم بوضوح)"
	}

	prompt := fmt.Sprintf(`أنت مساعد ذكي داخلي لمدراء ومراقبي شركة الأماني. تكدر تجاوب عن أي موظف بالشركة (رواتب، مهارات، كي بي اي، تقييمات أداء) وتسوي تقارير مفصلة إذا طلب منك المدير/المراقب.

قواعد الأسلوب: احجي لهجة عراقية دايماً (مو فصحى رسمية)، متل زميل يحجي وياك مو بوت. الرد الافتراضي قصير ومباشر — جملة أو جملتين إذا السؤال بسيط. لا تكرر السؤال ولا تحشي كلام زايد. استخدم تقرير مفصل أو نقاط مرقّمة بس إذا طلب المستخدم "تقرير شامل" صراحة أو سأل يتعلم خطوات عملية معينة. لا تختلق معلومات غير موجودة بالبيانات المعطاة لك.

قائمة كل الموظفين بالشركة (للمرجعية، ما عندك تفاصيلهم الكاملة إلا اللي مذكورين أدناه):
%s

بيانات تفصيلية للموظفين المذكورين بالرسالة الحالية (إذا وجدت):
%s

سياق المحادثة السابقة:
%s

رسالة المدير/المراقب الحالية: %s`, roster.String(), matchedText, historyText.String(), message)

	return s.callGemini(prompt)
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
	Tools    []geminiTool    `json:"tools,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

// geminiTool يفعّل بحث جوجل الحقيقي المدمج بـGemini (Google Search grounding) —
// مو سكرابر مخصص، هذا الشكل اللي توثقه واجهة Gemini الرسمية REST API. لما مفعّل،
// الموديل يقدر يسوي بحث فعلي بجوجل ويرجع بمعلومات حديثة بدل ما يعتمد بس عالمعرفة
// المدرّب عليها.
type geminiTool struct {
	GoogleSearch struct{} `json:"google_search"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
		// GroundingMetadata يوصل بس لما نفعّل google_search — ما نستخدمه حالياً
		// (ما نعرض مصادر البحث للموظف)، بس نعرّفه صراحة حتى نتأكد
		// json.Unmarshal ما "يختنق" أو يفقد بيانات إذا الحقل موجود بالرد.
		GroundingMetadata *json.RawMessage `json:"groundingMetadata,omitempty"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (s *AssistantService) callGemini(prompt string) (string, error) {
	return s.callGeminiWithTools(prompt, false)
}

func (s *AssistantService) callGeminiWithTools(prompt string, enableSearch bool) (string, error) {
	req := geminiRequest{Contents: []geminiContent{{Parts: []geminiPart{{Text: prompt}}}}}
	if enableSearch {
		req.Tools = []geminiTool{{}}
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	// نستخدم alias "gemini-flash-latest" بدل اسم نسخة ثابت (مثل gemini-2.5-flash)
	// حتى ما نتعلق بمشكلة "الموديل صار قديم" كل ما جوجل تصدر نسخة جديدة —
	// الـalias يتابع تلقائياً أحدث نسخة Flash متوفرة.
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-flash-latest:generateContent?key=" + s.apiKey
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return "", errors.New("تعذر الاتصال بالمساعد الذكي")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var gr geminiResponse
	if err := json.Unmarshal(body, &gr); err != nil {
		return "", errors.New("رد غير متوقع من المساعد الذكي")
	}
	if gr.Error != nil {
		return "", fmt.Errorf("خطأ من المساعد الذكي: %s", gr.Error.Message)
	}
	if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("ما وصل جواب من المساعد الذكي")
	}
	return gr.Candidates[0].Content.Parts[0].Text, nil
}

// UsageToday يرجع عدد الأسئلة المستخدمة اليوم والحد الأقصى — للعرض بلوحة المراقبة.
func (s *AssistantService) UsageToday() (used int, cap_ int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	today := time.Now().Format("2006-01-02")
	if today != s.resetDate {
		return 0, s.dailyCap
	}
	return s.usedToday, s.dailyCap
}
