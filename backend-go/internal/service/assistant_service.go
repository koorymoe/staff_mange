package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
	apiKey      string
	dailyCap    int
	employees   *repository.EmployeeRepository
	kpi         *repository.KpiRepository
	perfReviews *repository.PerformanceReviewRepository
	bookings    *repository.BookingRepository
	missions    *repository.MissionRepository
	expenses    *repository.ExpenseRepository
	gps         *repository.GpsRepository
	qualityFUs  *repository.QualityFollowUpRepository
	complaints  *repository.ComplaintRepository

	mu        sync.Mutex
	usedToday int
	resetDate string
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
) *AssistantService {
	return &AssistantService{
		apiKey: apiKey, dailyCap: dailyCap, employees: employees, kpi: kpi, perfReviews: perfReviews,
		bookings: bookings, missions: missions, expenses: expenses, gps: gps, qualityFUs: qualityFUs, complaints: complaints,
		resetDate: time.Now().Format("2006-01-02"),
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

	systemContext := fmt.Sprintf(`أنت مساعد ذكي داخلي لنظام إدارة موظفين شركة الأماني. جاوب بس عن بيانات الموظف الحالي أدناه (بياناته الشخصية + بيانات شغله المرتبطة بدوره الوظيفي المذكورة أدناه)، بأسلوب ودود ومختصر بالعربي (لهجة عراقية بسيطة إذا مناسب). لا تختلق معلومات غير موجودة هنا، وإذا السؤال خارج نطاق بياناته الشخصية أو شغله اعتذر بأدب وقول تراجع الإدارة. إذا سألك يشرحلك شلون يسوي شي بالنظام (مثلاً "شلون أنسق حجز؟")، اشرحله خطوة بخطوة بأسلوب تعليمي بسيط.

بيانات الموظف السائل:
- الاسم: %s
- الدور الوظيفي: %s
- %s
- عدد المهارات المتقنة: %d من أصل %d
- مجموع نقاط الكي بي اي هذا الشهر: %d

بيانات شغله الحالية (حسب دوره الوظيفي):
%s

سؤال الموظف: %s`, employee.Name, employee.Role, salaryLine, skillsKnown, skillsTotal, kpiPointsThisMonth, domainBlock, message)

	return s.callGemini(systemContext)
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
		pending, _ := s.bookings.List("PENDING", "")
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

	prompt := fmt.Sprintf(`أنت مساعد ذكي داخلي لمدراء ومراقبي شركة الأماني. تكدر تجاوب عن أي موظف بالشركة (رواتب، مهارات، كي بي اي، تقييمات أداء) وتسوي تقارير مفصلة إذا طلب منك المدير/المراقب. جاوب بالعربي (لهجة عراقية بسيطة إذا مناسب)، بأسلوب محادثة طبيعي مو تقرير جامد إلا إذا طلب المستخدم "تقرير شامل" صراحة. لا تختلق معلومات غير موجودة بالبيانات المعطاة لك.

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
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (s *AssistantService) callGemini(prompt string) (string, error) {
	reqBody, err := json.Marshal(geminiRequest{Contents: []geminiContent{{Parts: []geminiPart{{Text: prompt}}}}})
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
