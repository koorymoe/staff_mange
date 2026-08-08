package service

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ نظام الغرامات التلقائي ═══
//
// المبدأ: النظام هو الي يغرّم، مو المدير. الغرامة تنزل لحالها بلا ما
// يتدخل أحد، وتنعلن لكل الموظفين فوراً بلوحة الإعلانات — حتى ما تصير
// محاباة ولا نسيان ولا «تعال بكرة نحچي».
//
// منو ينغرم؟ الإداري الي كلّف الكادر — مو الليدر. سبب هذا: الليدر
// مسؤول عن شغله، بس الإداري مسؤول عن **متابعة** كادره. لو الليدر أنجز
// وما سوّى فاتورة وتقرير ومرت ١٦ ساعة، معناها الإداري ما تابعه.
type DisciplineService struct {
	repo          *repository.DisciplineRepository
	announcements *repository.AnnouncementRepository
	notifications *repository.NotificationRepository
	employees     *repository.EmployeeRepository
}

func NewDisciplineService(
	repo *repository.DisciplineRepository,
	announcements *repository.AnnouncementRepository,
	notifications *repository.NotificationRepository,
	employees *repository.EmployeeRepository,
) *DisciplineService {
	return &DisciplineService{repo: repo, announcements: announcements, notifications: notifications, employees: employees}
}

func (s *DisciplineService) List() ([]model.DisciplinePoints, error) {
	return s.repo.List()
}

func (s *DisciplineService) Events(employeeID string, limit int) ([]model.DisciplineEvent, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return s.repo.Events(employeeID, limit)
}

// announce ينشر الغرامة بلوحة الإعلانات لكل الموظفين. الإعلان جزء من
// العقوبة — الشفافية هي الي تخلي الناس تنتبه.
// Adjust تعديل يدوي على رصيد موظف من المالك أو مدير النظام.
//
// نشترط سبب مكتوب: تعديل بلا سبب ما ينفهم بعد شهر، لا من المالك ولا
// من الموظف الي انطلب منه توضيح. والموظف ينوصله إشعار بالتعديل
// وسببه — التعديل بالخفية يكسر ثقة الموظف بالنظام كله.
func (s *DisciplineService) Adjust(employeeID string, delta int, reason, byEmployeeID string) (*model.DisciplinePoints, error) {
	if employeeID == "" {
		return nil, errors.New("لازم تحدد الموظف")
	}
	if delta == 0 {
		return nil, errors.New("لازم تحدد كم نقطة تريد تزيد أو تنقص")
	}
	if utf8.RuneCountInString(strings.TrimSpace(reason)) < 3 {
		return nil, errors.New("لازم تكتب سبب التعديل")
	}

	remaining, applied, err := s.repo.Adjust(employeeID, delta, strings.TrimSpace(reason), byEmployeeID)
	if err != nil {
		return nil, err
	}
	if applied == 0 {
		return nil, errors.New("الرصيد وصل الحد — ما تغيّر شي (الرصيد بين ٠ و١٠٠)")
	}

	if s.notifications != nil {
		verb := "انزادت"
		amount := applied
		if applied < 0 {
			verb = "انخصمت"
			amount = -applied
		}
		_ = s.notifications.Create(employeeID, "discipline_manual",
			"📝 "+verb+" "+strconv.Itoa(amount)+" نقطة من رصيد الانضباط مالتك بتعديل إداري (السبب: "+
				strings.TrimSpace(reason)+"). رصيدك الحالي "+strconv.Itoa(remaining)+" من ١٠٠")
	}

	return &model.DisciplinePoints{
		EmployeeID:    employeeID,
		Points:        remaining,
		DeductedDinar: (model.DisciplineStartingPoints - remaining) * model.DisciplineDinarPerPoint,
	}, nil
}

func (s *DisciplineService) announce(body string) {
	if s.announcements == nil {
		return
	}
	// لوحة الإعلانات تشترط كاتب موجود بجدول الموظفين، والنظام ما عنده
	// حساب — فنكتبها باسم المالك. تنتهي بعد أسبوع حتى ما تتكدّس اللوحة.
	author, err := s.repo.SystemAuthorID()
	if err != nil || author == "" {
		log.Printf("[discipline] ماكو حساب يصلح لنشر الإعلان: %v", err)
		return
	}
	if _, err := s.announcements.Create(body, author, 7); err != nil {
		log.Printf("[discipline] تعذر نشر الإعلان: %v", err)
	}
}

// penalize يخصم نقطة ويعلنها ويشعّر صاحبها.
func (s *DisciplineService) penalize(employeeID, employeeName, kind, reason string, bookingID *string) {
	applied, left, err := s.repo.Penalize(employeeID, kind, reason, bookingID, 1)
	if err != nil {
		log.Printf("[discipline] تعذر تسجيل الغرامة: %v", err)
		return
	}
	if !applied {
		return // انسجّلت قبل — ما نعيدها
	}
	body := fmt.Sprintf(
		"⚠️ غرامة انضباط: خُصمت نقطة وحدة (%s د.ع) من %s — %s. الرصيد المتبقي: %d من %d نقطة.",
		formatDinar(model.DisciplineDinarPerPoint), employeeName, reason, left, model.DisciplineStartingPoints,
	)
	s.announce(body)
	if s.notifications != nil {
		_ = s.notifications.Create(employeeID, "discipline_penalty", body)
	}
	log.Printf("[discipline] غرامة: %s — %s (بقي %d)", employeeName, reason, left)
}

// RunPaperworkSweep يمر على الحجوزات المنجزة الي تأخر ورقها ويغرّم
// الإداري الي كلّف. يشتغل دورياً بالخلفية.
func (s *DisciplineService) RunPaperworkSweep() {
	rows, err := s.repo.OverduePaperwork(model.DisciplinePaperworkHours)
	if err != nil {
		log.Printf("[discipline] تعذر فحص الحجوزات المتأخرة: %v", err)
		return
	}
	for i := range rows {
		r := rows[i]
		missing := "الفاتورة والتقرير"
		switch {
		case r.HasInvoice && !r.HasReport:
			missing = "التقرير"
		case !r.HasInvoice && r.HasReport:
			missing = "الفاتورة"
		}
		bid := r.BookingID
		s.penalize(r.AdminID, r.AdminName, model.DisciplineLatePaperwork,
			fmt.Sprintf("مرّت %d ساعة على إنجاز الحجز %s بدون %s، وما تابع الكادر المكلّف",
				model.DisciplinePaperworkHours, r.BookingCode, missing),
			&bid)
	}
}

// RunRestoreSweep يرجّع نقطة وحدة لكل موظف اشتغل المدة المطلوبة بلا أي
// غرامة. هذا هو الجواب الوحيد على «أريد تخفيض بالنقاط»: ما اكو واسطة —
// اشتغل نظيف والنقطة ترجع لحالها.
func (s *DisciplineService) RunRestoreSweep() {
	ids, err := s.repo.EligibleForRestore(model.DisciplineCleanDaysToRestore)
	if err != nil {
		log.Printf("[discipline] تعذر فحص المستحقين لرجوع نقطة: %v", err)
		return
	}
	for _, id := range ids {
		emp, err := s.employees.FindByID(id)
		if err != nil || emp == nil {
			continue
		}
		reason := fmt.Sprintf("اشتغل %d أيام بلا أي غرامة", model.DisciplineCleanDaysToRestore)
		if err := s.repo.RestoreOne(id, reason); err != nil {
			log.Printf("[discipline] تعذر إرجاع نقطة: %v", err)
			continue
		}
		s.announce(fmt.Sprintf("✅ رجعت نقطة وحدة إلى %s — %s.", emp.Name, reason))
		log.Printf("[discipline] رجوع نقطة: %s", emp.Name)
	}
}

// CheckAssignmentBalance يغرّم الإداري لو كلّف ليدر عنده حجوزات شغّالة
// وبنفس الوقت أكو ليدر ثاني فاضي تماماً. التوزيع الغلط يتحاسب عليه.
//
// ما نغرّم إلا بالحالة الواضحة: المكلَّف عنده شغل *وأكو* واحد ماعنده
// ولا حجز. لو الاثنين مشغولين أو الاثنين فاضيين ما اكو خطأ.
func (s *DisciplineService) CheckAssignmentBalance(adminID, assignedLeaderID, bookingID, bookingCode string, activeByLeader map[string]int, leaderNames map[string]string) {
	if adminID == "" || adminID == assignedLeaderID {
		return
	}
	chosen := activeByLeader[assignedLeaderID]
	if chosen == 0 {
		return // كلّف واحد فاضي — صح
	}
	freeName := ""
	for id, count := range activeByLeader {
		if id != assignedLeaderID && count == 0 {
			freeName = leaderNames[id]
			break
		}
	}
	if freeName == "" {
		return // ما اكو ليدر فاضي — ما إله خيار ثاني
	}
	admin, err := s.employees.FindByID(adminID)
	if err != nil || admin == nil {
		return
	}
	bid := bookingID
	s.penalize(adminID, admin.Name, model.DisciplineUnbalancedAssign,
		fmt.Sprintf("كلّف %s بالحجز %s وعنده %d حجز شغّال، بينما %s فاضي تماماً",
			leaderNames[assignedLeaderID], bookingCode, chosen, freeName),
		&bid)
}

// StartBackgroundSweeps يشغّل الفحص الدوري. كل ساعة يكفي: المهلة ١٦
// ساعة، فما اكو داعي نفحص كل دقيقة ونحمّل قاعدة البيانات بلا فايدة.
func (s *DisciplineService) StartBackgroundSweeps() {
	go func() {
		// تأخير أولي حتى ما نزاحم إقلاع السيرفر
		time.Sleep(2 * time.Minute)
		for {
			s.RunPaperworkSweep()
			s.RunRestoreSweep()
			time.Sleep(time.Hour)
		}
	}()
}

// formatDinar يكتب المبلغ بفواصل الآلاف — ١٠٠٠٠ → 10,000
func formatDinar(n int) string {
	s := fmt.Sprintf("%d", n)
	out := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out += ","
		}
		out += string(c)
	}
	return out
}
