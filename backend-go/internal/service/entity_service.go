package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
	"staffmange-api/internal/storage"
)

// ═══ الكيان — مراقب ومساعد شخصي لكل موظف ═══
//
// «كيان يهابه ويخافه الموظف… يرحّب بيه أول ما يفتح النظام، ويحذّره
// قبل ما تنزل عليه الغرامة، ويتغيّر وجهه لو ما التزم — كأنما بشر
// يراقب بشر».
//
// ⚠️⚠️ **الكيان ما يخترع رقماً أبداً.** كل جملة يقولها مبنية على
// نفس البيانات الي تبني عليها الغرامة الفعلية:
//   - مهلة الورق وقيمتها بالدينار: ثوابت `model/discipline.go`
//   - الحجوزات الناقص ورقها: `DisciplineRepository.PendingPaperworkForEmployee`
//     — نفس شروط مكانس الغرامة حرفياً
//   - المهام الإضافية: `ExtraTaskRepository`
//   - رصيد الانضباط وأحداثه: `DisciplineRepository`
// تحذير كاذب مرة وحدة يخلّي الموظف يتجاهل كل تحذير بعدها.

// EntityService يبني تقرير الكيان الحي ويولّد شخصيته.
type EntityService struct {
	characters  *repository.EmployeeCharacterRepository
	employees   *repository.EmployeeRepository
	discipline  *repository.DisciplineRepository
	extraTasks  *repository.ExtraTaskRepository
	bookings    *repository.BookingRepository
	kpi         *repository.KpiRepository
	permissions *repository.PermissionRepository
	assistant   *AssistantService
	store       storage.Store
	apiKey      string
	imageModel  string
}

func NewEntityService(
	characters *repository.EmployeeCharacterRepository,
	employees *repository.EmployeeRepository,
	discipline *repository.DisciplineRepository,
	extraTasks *repository.ExtraTaskRepository,
	bookings *repository.BookingRepository,
	kpi *repository.KpiRepository,
	permissions *repository.PermissionRepository,
	assistant *AssistantService,
	store storage.Store,
	apiKey, imageModel string,
) *EntityService {
	return &EntityService{
		characters: characters, employees: employees, discipline: discipline,
		extraTasks: extraTasks, bookings: bookings, kpi: kpi, permissions: permissions,
		assistant: assistant, store: store, apiKey: apiKey, imageModel: imageModel,
	}
}

// ─────────────────────────── التقرير الحي ───────────────────────────

// Briefing يبني كل الي يحتاجه الكيان حتى يتكلم مع صاحبه بصدق.
//
// ⚠️ الموظف يجيب **تقريره هو بس** — المعالج يمرّر هويته من التوكن،
// وماكو معامل employeeId من الطلب. تقرير موظف بيد زميله يعني كشف
// غراماته ورصيده لأي أحد.
func (s *EntityService) Briefing(employeeID string) (*model.EntityBriefing, error) {
	emp, err := s.employees.FindByID(employeeID)
	if err != nil || emp == nil {
		return nil, errors.New("تعذر تحديد هوية الموظف")
	}

	out := &model.EntityBriefing{
		Mood:           model.EntityMoodHappy,
		Points:         model.DisciplineStartingPoints,
		Lines:          []model.EntityLine{},
		CharacterState: "NONE",
	}
	out.Greeting = fmt.Sprintf("مرحبا أستاذ %s — آني المراقب عليك والمساعد بنفس الوقت.", firstName(emp.Name))

	// ① الورق المتأخر — التحذير قبل الغرامة، وبنفس أرقامها
	urgentPaperwork := s.appendPaperworkLines(employeeID, emp, out)

	// ② المهام الإضافية المعلّقة
	urgentTasks := s.appendExtraTaskLines(employeeID, out)

	// ③ حجوزات بانتظار تواصل الإداري — لمن هذا شغله فعلاً
	s.appendBookingLines(employeeID, emp, out)

	// ④ الانضباط: الرصيد والخصم الجديد
	recentPenalty, restoredPoint := s.appendDisciplineLines(employeeID, out)

	// المزاج يُشتق من نفس البيانات — مو عشوائي ولا مزيّن.
	//
	// ⚠️ الترتيب مقصود: التحذير يسبق الفرح. موظف عنده تأخير وبنفس
	// الوقت خلّص ورقة ثانية **ما يفرح** — الي يحتاج يشوفه هو
	// التأخير.
	switch {
	case recentPenalty || urgentPaperwork || urgentTasks:
		out.Mood = model.EntityMoodAngry
	case len(out.Lines) > 0:
		out.Mood = model.EntityMoodWatching
	case restoredPoint || s.recentPaperworkDone(employeeID):
		// نظيف **وصار شي إيجابي حقيقي** بآخر ٢٤ ساعة.
		out.Mood = model.EntityMoodPositive
	}

	s.attachCharacter(employeeID, out)
	return out, nil
}

// recentPaperworkDone هل خلّص ورق حجز بآخر ٢٤ ساعة؟
//
// ⚠️ الإشارة الثانية للفرح. الفشل يرجّع `false` — يعني **ما يفرح**
// عند الشك. الاحتفال الكاذب أسوأ من فرح ضايع.
//
// ⚠️ وتنسأل **بس** لمن يكون الموظف نظيفاً أصلاً (آخر فرع بالمزاج)،
// فما تنضاف نداة قاعدة بيانات لكل بريفينغ.
func (s *EntityService) recentPaperworkDone(employeeID string) bool {
	if s.bookings == nil {
		return false
	}
	done, err := s.bookings.RecentPaperworkDone(employeeID, time.Now().Add(-24*time.Hour))
	if err != nil {
		log.Printf("[entity] تعذر فحص الورق المنجَز للموظف %s: %v", employeeID, err)
		return false
	}
	return done
}

// appendPaperworkLines يحذّر من ورق الحجوزات الناقص — قبل الغرامة
// وبعدها. يرجّع صحيح لو أكو حجز **فاتت مهلته فعلاً**.
func (s *EntityService) appendPaperworkLines(employeeID string, emp *model.Employee, out *model.EntityBriefing) bool {
	rows, err := s.discipline.PendingPaperworkForEmployee(employeeID)
	if err != nil {
		log.Printf("[entity] تعذر جلب ورق الموظف %s: %v", employeeID, err)
		return false
	}
	overdue := false
	for i := range rows {
		r := rows[i]
		// المهلة تختلف: الليدر ٢٤ ساعة لأنه الي طلع وسوّى الشغل،
		// والإداري ٤٨ لأنه المسؤول عن متابعة كادره بعدها.
		limit := model.DisciplinePaperworkHours
		who := "كإداري كلّف الكادر"
		if r.AsLeader {
			limit = model.DisciplineLeaderPaperworkHours
			who = "كليدر طلع بالحجز"
		}
		missing := "الفاتورة والتقرير"
		switch {
		case r.HasInvoice && !r.HasReport:
			missing = "التقرير"
		case !r.HasInvoice && r.HasReport:
			missing = "الفاتورة"
		}
		passed := time.Since(r.CompletedAt).Hours()
		remaining := float64(limit) - passed
		if remaining <= 0 {
			overdue = true
			out.DinarAtRisk += model.DisciplineDinarPerPoint
			out.Lines = append(out.Lines, model.EntityLine{
				Kind: model.EntityLinePaperwork,
				Text: fmt.Sprintf("الحجز %s فاتت مهلته (%d ساعة %s) وباقي عليه %s — الغرامة %s د.ع.",
					r.BookingCode, limit, who, missing, formatDinar(model.DisciplineDinarPerPoint)),
				Link:   "/bookings",
				Urgent: true,
			})
			continue
		}
		out.Lines = append(out.Lines, model.EntityLine{
			Kind: model.EntityLinePaperwork,
			Text: fmt.Sprintf("لا تنسَ %s للحجز %s — باقي %s قبل غرامة %s د.ع.",
				missing, r.BookingCode, humanHours(remaining), formatDinar(model.DisciplineDinarPerPoint)),
			Link: "/bookings",
		})
	}
	_ = emp
	return overdue
}

// appendExtraTaskLines المهام الموجّهة الي ما فتحها أو تأخرت.
func (s *EntityService) appendExtraTaskLines(employeeID string, out *model.EntityBriefing) bool {
	rows, err := s.extraTasks.ListForEmployee(employeeID, false)
	if err != nil {
		log.Printf("[entity] تعذر جلب مهام الموظف %s: %v", employeeID, err)
		return false
	}
	overdue := false
	unseen := 0
	for i := range rows {
		t := rows[i]
		if t.Status != model.ExtraTaskNew && t.Status != model.ExtraTaskInProgress {
			continue
		}
		if t.DueAt != nil && time.Now().After(*t.DueAt) {
			overdue = true
			out.Lines = append(out.Lines, model.EntityLine{
				Kind:   model.EntityLineExtraTask,
				Text:   fmt.Sprintf("مهمة «%s» فات موعدها وما خلصتها.", t.Title),
				Link:   "/my-extra-tasks",
				Urgent: true,
			})
			continue
		}
		if t.Status == model.ExtraTaskNew && t.SeenAt == nil {
			unseen++
		}
	}
	if unseen > 0 {
		out.Lines = append(out.Lines, model.EntityLine{
			Kind: model.EntityLineExtraTask,
			Text: fmt.Sprintf("عندك %d مهمة موجّهة لسه ما فتحتها.", unseen),
			Link: "/my-extra-tasks",
		})
	}
	return overdue
}

// appendBookingLines حجوزات مثبّتة بانتظار تواصل الإداري.
//
// ⚠️ ما ننسبها لموظف بعينه: الحجز `PENDING` قبل التثبيت ماكو إله
// «إداري مسؤول» محدد بالنظام (يطلع «غير محدد» بشاشة التنسيق نفسها).
// فالكيان يقولها كطابور شغل مشترك لمن هذا شغله — نسبتها لشخص تعني
// اتهامه بشي مو مسجّل عليه.
func (s *EntityService) appendBookingLines(employeeID string, emp *model.Employee, out *model.EntityBriefing) {
	if !s.doesCoordination(employeeID, emp) {
		return
	}
	rows, err := s.bookings.List("PENDING", "", "", 0)
	if err != nil {
		log.Printf("[entity] تعذر جلب الحجوزات المعلّقة: %v", err)
		return
	}
	waiting := 0
	for i := range rows {
		if rows[i].ConfirmationContactedAt == nil {
			waiting++
		}
	}
	if waiting > 0 {
		out.Lines = append(out.Lines, model.EntityLine{
			Kind: model.EntityLineBooking,
			Text: fmt.Sprintf("أكو %d حجز بانتظار تواصل الإداري وتثبيته.", waiting),
			Link: "/coordinator",
		})
	}
}

// doesCoordination هل تنسيق الحجوزات شغل هذا الموظف فعلاً؟
func (s *EntityService) doesCoordination(employeeID string, emp *model.Employee) bool {
	if emp.Role == "ADMIN" || emp.Role == "OWNER" || emp.Role == "HR_COORDINATOR" {
		return true
	}
	perms, err := s.permissions.ListForEmployee(employeeID)
	if err != nil {
		return false
	}
	for _, p := range perms {
		if p.Name == "coordinator" || p.Name == "crew_management" {
			return true
		}
	}
	return false
}

// appendDisciplineLines الرصيد والخصم الجديد. يرجّع صحيح لو انخصم
// منه شي بآخر ٢٤ ساعة — وهذا الي يقلب وجه الكيان لغاضب.
// يرجّع: (انخصم منه بآخر ٢٤ ساعة، رجعتله نقطة بآخر ٢٤ ساعة).
func (s *EntityService) appendDisciplineLines(employeeID string, out *model.EntityBriefing) (bool, bool) {
	all, err := s.discipline.List()
	if err == nil {
		for i := range all {
			if all[i].EmployeeID == employeeID {
				out.Points = all[i].Points
				break
			}
		}
	}
	events, err := s.discipline.Events(employeeID, 20)
	if err != nil {
		return false, false
	}
	recent := false
	restored := false
	for i := range events {
		e := events[i]
		if time.Since(e.CreatedAt) > 24*time.Hour {
			continue
		}
		// ⚠️ نقطة رجعتله = حدث إيجابي حقيقي مؤرّخ. نلتقطه بنفس
		// الجولة — بلا استعلام ثاني.
		if e.Delta > 0 {
			restored = true
			continue
		}
		if e.Delta == 0 {
			continue
		}
		recent = true
		out.Lines = append(out.Lines, model.EntityLine{
			Kind: model.EntityLineDiscipline,
			Text: fmt.Sprintf("انخصمت منك %d نقطة: %s — رصيدك هسه %d من %d.",
				-e.Delta, e.Reason, out.Points, model.DisciplineStartingPoints),
			Link:   "/discipline",
			Urgent: true,
		})
		break // خصم واحد يكفي — تكرارها يخلّي الفقاعة جدار
	}
	if out.Points < model.DisciplineStartingPoints {
		out.DinarAtRisk += 0 // الخصم صار فعلاً، مو «معرّض» — يبقى بالرصيد
	}
	return recent, restored
}

// attachCharacter يعلّق صور الشخصية إذا انولدت.
func (s *EntityService) attachCharacter(employeeID string, out *model.EntityBriefing) {
	ch, err := s.characters.FindByEmployee(employeeID)
	if err != nil || ch == nil {
		return
	}
	out.CharacterState = ch.Status
	if ch.Persona != nil {
		out.Persona = *ch.Persona
	}
	if ch.Status != model.CharacterReady {
		return
	}
	out.CalmURL = fileURL(ch.CalmKey)
	out.HappyURL = fileURL(ch.HappyKey)
	out.AngryURL = fileURL(ch.AngryKey)
}

func fileURL(key *string) string {
	if key == nil || *key == "" {
		return ""
	}
	return "/api/files/" + *key
}

// ─────────────────────────── توليد الشخصية ───────────────────────────

// Character يرجّع شخصية موظف (أو nil لو ما انولدت).
func (s *EntityService) Character(employeeID string) (*model.EmployeeCharacter, error) {
	return s.characters.FindByEmployee(employeeID)
}

// GenerateCharacter يولّد شخصية الكيان لموظف: وصف مشتق من بياناته
// الحقيقية، وثلاث صور بملامح مختلفة.
//
// ⚠️ **بالطلب لا للجميع دفعة وحدة**: سقف Gemini اليومي محدود، وتوليد
// عشرات الموظفين × ثلاث صور يحرقه بنداء واحد. المدير يولّد موظفاً
// موظفاً.
func (s *EntityService) GenerateCharacter(employeeID, byEmployeeID string) (*model.EmployeeCharacter, error) {
	emp, err := s.employees.FindByID(employeeID)
	if err != nil || emp == nil {
		return nil, errors.New("الموظف غير موجود")
	}
	if s.apiKey == "" {
		return nil, errors.New("مفتاح الذكاء الاصطناعي غير مضبوط — ما نقدر نولّد الشخصية")
	}
	if err := s.characters.MarkPending(employeeID, byEmployeeID); err != nil {
		return nil, err
	}

	persona := s.derivePersona(emp)
	prompt := s.characterPrompt(emp, persona)

	keys := map[string]string{}
	for _, mood := range []string{"calm", "happy", "angry"} {
		data, err := s.generateImage(prompt + "\n\n" + moodClause(mood))
		if err != nil {
			_ = s.characters.MarkFailed(employeeID, err.Error())
			return nil, err
		}
		key := storage.NewKey("characters", "image/png")
		if err := s.store.Put(context.Background(), key, data, "image/png"); err != nil {
			_ = s.characters.MarkFailed(employeeID, "تعذر حفظ صورة الشخصية")
			return nil, errors.New("تعذر حفظ صورة الشخصية")
		}
		keys[mood] = key
	}

	if err := s.characters.SaveReady(employeeID, persona, prompt, keys["calm"], keys["happy"], keys["angry"]); err != nil {
		return nil, err
	}
	return s.characters.FindByEmployee(employeeID)
}

// derivePersona يشتق وصف الشخصية من بيانات الموظف الحقيقية — دوره
// ورصيد انضباطه ومهاراته. مو وصف مزاجي: هو الي يحدد شكل الكيان
// وطريقة كلامه، فلازم يطلع من واقع الموظف مو من خيال الموديل.
func (s *EntityService) derivePersona(emp *model.Employee) string {
	points := model.DisciplineStartingPoints
	if all, err := s.discipline.List(); err == nil {
		for i := range all {
			if all[i].EmployeeID == emp.ID {
				points = all[i].Points
				break
			}
		}
	}
	skills := 0
	if list, err := s.employees.SkillsForEmployee(emp.ID); err == nil {
		for i := range list {
			if list[i].CanPerform {
				skills++
			}
		}
	}
	facts := fmt.Sprintf("الاسم: %s · الدور: %s · رصيد الانضباط: %d من %d · عدد المهارات: %d · المنصب: %s",
		emp.Name, roleLabel(emp.Role), points, model.DisciplineStartingPoints, skills, safeStr(emp.Position))

	if s.assistant == nil {
		return facts
	}
	out, err := s.assistant.GenerateReport(
		"اكتب سطر واحد بالعربية (١٥ كلمة كحد أقصى) يوصف شخصية هذا الموظف بالعمل، " +
			"مبني على هذي الحقائق بس وبلا مبالغة ولا مجاملة:\n" + facts +
			"\nالمطلوب: سطر واحد بس، بلا مقدمات ولا علامات تنسيق.")
	if err != nil || strings.TrimSpace(out) == "" {
		// فشل التوليد ما يوقف الشخصية — الحقائق نفسها وصف كافٍ
		return facts
	}
	return strings.TrimSpace(out)
}

func (s *EntityService) characterPrompt(emp *model.Employee, persona string) string {
	return fmt.Sprintf(
		"Create a friendly 3D cartoon mascot character (Pixar-like stylized render) of an Iraqi company employee. "+
			"The character wears a bright blue work vest over a black long-sleeve shirt and black trousers, "+
			"standing full body, centered, on a dark navy blue gradient background. "+
			"Role: %s. Character notes: %s. "+
			"Square image, clean, no text, no logos, no watermark.",
		roleLabel(emp.Role), persona)
}

func moodClause(mood string) string {
	switch mood {
	case "happy":
		return "Expression: big warm smile, cheerful, giving a thumbs up."
	case "angry":
		return "Expression: stern and displeased, arms crossed, frowning, disciplinary look."
	default:
		return "Expression: calm and alert, watching attentively, neutral face, hands at sides."
	}
}

// ─────────────────────── نداء توليد الصور ───────────────────────

type geminiImageResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				InlineData *struct {
					MimeType string `json:"mimeType"`
					Data     string `json:"data"`
				} `json:"inlineData,omitempty"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// generateImage ينادي موديل الصور ويرجّع بايتات PNG.
//
// ⚠️ اسم الموديل من متغيّر بيئة (`GEMINI_IMAGE_MODEL`) مو مكتوب
// بالكود: موديلات الصور تتبدّل أسماؤها، وتثبيت الاسم يعني ميزة
// تنكسر بلا ما نلمس سطر.
func (s *EntityService) generateImage(prompt string) ([]byte, error) {
	body, _ := json.Marshal(map[string]any{
		"contents": []map[string]any{{
			"parts": []map[string]any{{"text": prompt}},
		}},
	})
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		s.imageModel, s.apiKey)
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, errors.New("تعذر الاتصال بمولّد الصور")
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var gr geminiImageResponse
	if err := json.Unmarshal(raw, &gr); err != nil {
		return nil, errors.New("رد غير متوقع من مولّد الصور")
	}
	if gr.Error != nil {
		return nil, fmt.Errorf("خطأ من مولّد الصور: %s", gr.Error.Message)
	}
	for _, c := range gr.Candidates {
		for _, p := range c.Content.Parts {
			if p.InlineData != nil && p.InlineData.Data != "" {
				data, err := base64.StdEncoding.DecodeString(p.InlineData.Data)
				if err != nil {
					return nil, errors.New("صورة غير صالحة من المولّد")
				}
				return data, nil
			}
		}
	}
	return nil, errors.New("مولّد الصور ما رجّع صورة")
}

// ─────────────────────────── مساعدات ───────────────────────────

func firstName(full string) string {
	parts := strings.Fields(strings.TrimSpace(full))
	if len(parts) == 0 {
		return full
	}
	return parts[0]
}

func safeStr(p *string) string {
	if p == nil {
		return "—"
	}
	return *p
}

// humanHours «٣ ساعات» / «٤٥ دقيقة» — الموظف ما يفهم «2.75 ساعة».
func humanHours(h float64) string {
	if h < 1 {
		m := int(h * 60)
		if m < 1 {
			m = 1
		}
		return fmt.Sprintf("%d دقيقة", m)
	}
	return fmt.Sprintf("%d ساعة", int(h))
}

// roleLabels تسمية الدور بالعربي — بيانات مو شرط بالكود، والإضافة
// تصير بسطر واحد. (`formatDinar` موجودة أصلاً بـ`discipline_service.go`
// ونستعملها هي بدل نسخة ثانية.)
var roleLabels = map[string]string{
	"OWNER": "مالك النظام", "ADMIN": "مدير النظام", "SALES": "موظف مبيعات",
	"HR_COORDINATOR": "إداري الكوادر", "TECHNICIAN": "فني", "TECHNICAL": "تقني",
	"PROJECT_MANAGER": "مدير مشاريع", "MONITOR": "مراقب", "FINANCE": "محاسب",
	"GPS_ADMIN": "مسؤول GPS", "QUALITY_ENGINEER": "مهندس جودة", "ENGINEER": "مهندس",
	"PROCUREMENT_ADMIN": "إداري الكميات", "DESIGNER": "مصمم", "SERVICE_MANAGER": "مسؤول خدمة",
}

func roleLabel(role string) string {
	if v, ok := roleLabels[role]; ok {
		return v
	}
	return role
}
