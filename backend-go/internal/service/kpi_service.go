package service

import (
	"errors"
	"log"
	"strconv"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// announcementPublisher الجزء الي نحتاجه من مستودع الإعلانات — واجهة
// صغيرة بدل ما نربط الخدمة بالمستودع كامل.
type announcementPublisher interface {
	Create(body, byID string, expiresInDays int) (*model.Announcement, error)
}

// storyEmitter محرّك القصص — واجهة صغيرة بنفس نمط `announcementPublisher`
// فوق: نربط الخدمة بالي تحتاجه بس، مو بالمحرّك كامل.
type storyEmitter interface {
	Emit(req model.EmitStoryRequest)
}

type KpiService struct {
	repo          *repository.KpiRepository
	employees     *repository.EmployeeRepository
	notifications *repository.NotificationRepository
	announcements announcementPublisher
	stories       storyEmitter
}

// SetStories يركّب محرّك القصص بعد الإنشاء.
//
// ⚠️ **بـsetter مو بالمُنشئ بقصد**: القصة **إضافة على** الخصم، مو
// شرط له. خدمة الكي بي اي تشتغل كاملة بلا محرّك قصص — ولو ربطناه
// بالمُنشئ نخلي إجراءً مالياً/إدارياً يعتمد على ميزة عرض.
func (s *KpiService) SetStories(e storyEmitter) { s.stories = e }

func NewKpiService(repo *repository.KpiRepository, employees *repository.EmployeeRepository, notifications *repository.NotificationRepository, announcements announcementPublisher) *KpiService {
	return &KpiService{repo: repo, employees: employees, notifications: notifications, announcements: announcements}
}

func (s *KpiService) List() ([]model.KpiEvaluation, error) {
	return s.repo.List()
}

func (s *KpiService) ListForEmployee(employeeID string) ([]model.KpiEvaluation, error) {
	return s.repo.ListForEmployee(employeeID)
}

// Create ينشئ تقييماً جديداً. إذا كان التقييم سلبياً، يرجّع الموظف تلقائياً لوضع
// "متدرب" (isTrainee) حتى يكمل التدريب المطلوب قبل ما يرجع لباقي شاشات النظام.
// كذلك يرسل إشعارات: للموظف نفسه عند خصم نقطة، ولكل زملائه بنفس الدور إذا تغيّر
// المتصدر بالترتيب الشهري بعد هذا التقييم — إعلام بس، بدون أي تغيير تلقائي بالراتب.
func (s *KpiService) Create(req model.CreateKpiEvaluationRequest) (*model.KpiEvaluation, error) {
	if req.EmployeeID == "" || req.EvaluatorID == "" || req.Points == nil {
		return nil, errors.New("employeeId, evaluatorId, and points are required")
	}

	employee, empErr := s.employees.FindByID(req.EmployeeID)

	// ⚠️ الخصم يتقرّر بالمقدار مو بالإشارة. الواجهة تدزّ النقاط
	// بالموجب (١..٨) والشرط القديم چان `*req.Points < 0`، فما چان
	// ينطبق ولا مرة: الموظف ما چان يوصله إشعار بالخصم، و«انشرها
	// بلوحة الإعلانات» چان صندوقاً ميتاً — يتأشّر وماكو شي ينتشر.
	// نقرأ المقدار حتى يشتغل بالحالتين، والمخزون ما يتغيّر: نمرّر
	// req.Points مثل ما إجت حتى البيانات القديمة تبقى مقروءة بنفس
	// الطريقة.
	magnitude := *req.Points
	if magnitude < 0 {
		magnitude = -magnitude
	}
	deductionAmount := float64(magnitude) * 10000
	eval, err := s.repo.Create(req.EmployeeID, req.EvaluatorID, *req.Points, req.Reason, deductionAmount)
	if err != nil {
		return nil, err
	}
	if magnitude > 0 {
		_ = s.employees.SetTrainee(req.EmployeeID, true)
		reason := req.Reason
		if reason == "" {
			reason = "بدون سبب مذكور"
		}
		if s.notifications != nil {
			_ = s.notifications.Create(req.EmployeeID, "kpi_deduction",
				"⚠️ تم خصم "+strconv.Itoa(magnitude)+" نقطة من رصيدك (السبب: "+reason+")")
		}
		// نشر المخالفة بلوحة الإعلانات — بطلب المدير وقت التسجيل، ولمدة
		// ثلاثة أيام بس. فشل النشر ما يلغي التقييم، التقييم انسجل أصلاً.
		if req.Announce && s.announcements != nil && empErr == nil && employee != nil {
			body := "⚠️ مخالفة: " + employee.Name + " — خصم " +
				strconv.Itoa(magnitude) + " نقطة (" + reason + ")"
			if _, err := s.announcements.Create(body, req.EvaluatorID, model.AnnouncementPenaltyDays); err != nil {
				log.Printf("نشر مخالفة %s بلوحة الإعلانات: %v", req.EmployeeID, err)
			}
		}
	}

	// ⚠️ **القصة بعد ما ينجح التقييم ويترسّخ** — «الحركة لا تسبق
	// نجاح العملية». وفشلها ما يلغي الخصم: `Emit` ما ترجّع خطأ،
	// تسجّله بالسجل.
	//
	// ⚠️ **واسم المُقيِّم ينمرّر**: قرار (ع) الصريح إن الموظف يعرف
	// منو خصمه — تغيير سياسة، اليوم الإشعار يوصل بلا اسم.
	if s.stories != nil && magnitude > 0 && eval != nil {
		reason := req.Reason
		if reason == "" {
			reason = "بدون سبب مذكور"
		}
		senderName := ""
		if ev, err := s.employees.FindByID(req.EvaluatorID); err == nil && ev != nil {
			senderName = ev.Name
		}
		recipientName := ""
		if employee != nil {
			recipientName = employee.Name
		}
		evaluatorID := req.EvaluatorID
		s.stories.Emit(model.EmitStoryRequest{
			EventID:     eval.ID,
			EventKind:   model.StoryEventPointDeducted,
			SenderID:    &evaluatorID,
			SenderName:  senderName,
			RecipientID:   req.EmployeeID,
			RecipientName: recipientName,
			Payload: map[string]any{
				"title":  "انخصمت منك " + strconv.Itoa(magnitude) + " نقطة",
				"reason": reason,
				"points": magnitude,
				"dinar":  deductionAmount,
				"link":   "/kpi",
			},
		})
	}

	// ⚠️⚠️ إشعار «تصدّر التصنيف» انشال — چان يهنّي أكثر واحد انخصم منه.
	// RoleLeaderboard يرتّب `ORDER BY SUM(points) DESC` (kpi_repository.go:255)،
	// والخصومات مخزونة بالموجب، فأكثر موظف انعاقب يطلع board[0] —
	// والنظام چان يدزّ لكل زملائه بنفس الدور «🏆 فلان تصدر التصنيف
	// الشهري بالمركز الأول!». يعني الي انعاقب ينهنّى علناً كدامهم.
	//
	// وما ينفع نعكس الترتيب: نفس العمود يحمل الخصم والمكافأة
	// (CompleteTraining يسجّل +1 مكافأةً) بلا أي شي يفرّقهن، فالبيانات
	// نفسها ما تكدر تجاوب منو المتصدّر. إعلان متصدّر غلط أسوأ من
	// عدم إعلان أحد — فينشال لحد ما ينفصل الخصم عن المكافأة بعمود
	// أو نوع صريح، وبعدها يرجع صح.

	return eval, nil
}

// CompleteTraining يخرج الموظف من وضع "متدرب" ويسجل تقييماً إيجابياً يعكس
// اجتيازه للتدريب، بدل ما يظل التقييم السلبي وحده بسجله.
func (s *KpiService) CompleteTraining(employeeID, evaluatorID string) (*model.KpiEvaluation, error) {
	if err := s.employees.SetTrainee(employeeID, false); err != nil {
		return nil, err
	}
	points := 1
	deductionAmount := float64(points) * 10000
	return s.repo.Create(employeeID, evaluatorID, points, "إكمال التدريب بنجاح", deductionAmount)
}

func (s *KpiService) Delete(id string) error {
	return s.repo.Delete(id)
}

// Cancel "يرجّع" نقطة كي بي اي تم تسجيلها بالغلط أو بعد ما الموظف عدّل سلوكه —
// يحتفظ بالسجل بدل الحذف حتى يضل تاريخه واضح للمراقب.
func (s *KpiService) Cancel(id, cancelledByEmployeeID string) (*model.KpiEvaluation, error) {
	return s.repo.Cancel(id, cancelledByEmployeeID)
}

// RoleLeaderboard يرجع ترتيب موظفي دور معيّن أسبوعياً وشهرياً معاً
//
// ⚠️⚠️ ADMIN/OWNER مستثنون — نفس مبدأ الاستثناء الموجود أصلاً بـ
// PermissionLeaderboard (عندهم كل الصلاحيات بحكم موقعهم، فيطلعون
// بكل تصنيف ويزاحمون الي يشتغل الشغل فعلاً). بدون هذا الرفض، حساب
// إداري عليا بلا مسار عمل حقيقي يقارَن بحسابات ADMIN/OWNER الثانية
// بنقاط KPI/حجوزات — مقارنة بلا معنى لعمل إداري. الحارس هنا دفاع من
// العمق: حتى نداء مباشر بلا فتح الواجهة يُرفض.
func (s *KpiService) RoleLeaderboard(role string) (*model.RoleKpiLeaderboard, error) {
	if role == "ADMIN" || role == "OWNER" {
		return nil, errors.New("لا تصنيف شخصي للإداريين العليا")
	}
	const day = "2006-01-02"
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7).Format(day)
	monthAgo := now.AddDate(0, -1, 0).Format(day)
	// الفترة السابقة: الأسبوع/الشهر الي قبل الحالي بالضبط — للمقارنة
	twoWeeksAgo := now.AddDate(0, 0, -14).Format(day)
	twoMonthsAgo := now.AddDate(0, -2, 0).Format(day)

	weekly, err := s.repo.RoleLeaderboard(role, weekAgo, "")
	if err != nil {
		return nil, err
	}
	monthly, err := s.repo.RoleLeaderboard(role, monthAgo, "")
	if err != nil {
		return nil, err
	}
	prevWeekly, err := s.repo.RoleLeaderboard(role, twoWeeksAgo, weekAgo)
	if err != nil {
		return nil, err
	}
	prevMonthly, err := s.repo.RoleLeaderboard(role, twoMonthsAgo, monthAgo)
	if err != nil {
		return nil, err
	}

	applyDeltas(weekly, prevWeekly)
	applyDeltas(monthly, prevMonthly)

	return &model.RoleKpiLeaderboard{Role: role, Weekly: weekly, Monthly: monthly}, nil
}

// PermissionLeaderboard ترتيب حسب الشغل — كل صاحب صلاحية معيّنة
// ينقارن بأصحابها، مهما كان اسم دوره.
func (s *KpiService) PermissionLeaderboard(permission string) (*model.RoleKpiLeaderboard, error) {
	const day = "2006-01-02"
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7).Format(day)
	monthAgo := now.AddDate(0, -1, 0).Format(day)
	twoWeeksAgo := now.AddDate(0, 0, -14).Format(day)
	twoMonthsAgo := now.AddDate(0, -2, 0).Format(day)

	weekly, err := s.repo.PermissionLeaderboard(permission, weekAgo, "")
	if err != nil {
		return nil, err
	}
	monthly, err := s.repo.PermissionLeaderboard(permission, monthAgo, "")
	if err != nil {
		return nil, err
	}
	prevWeekly, err := s.repo.PermissionLeaderboard(permission, twoWeeksAgo, weekAgo)
	if err != nil {
		return nil, err
	}
	prevMonthly, err := s.repo.PermissionLeaderboard(permission, twoMonthsAgo, monthAgo)
	if err != nil {
		return nil, err
	}

	applyDeltas(weekly, prevWeekly)
	applyDeltas(monthly, prevMonthly)

	return &model.RoleKpiLeaderboard{Role: permission, Weekly: weekly, Monthly: monthly}, nil
}

// applyDeltas يحسب فرق النقاط وفرق الترتيب عن الفترة السابقة.
//
// ⚠️ فرق الترتيب معكوس بقصد: المركز ٣ → ١ معناه **تقدّم**، رغم إن
// الرقم نزل. لو رجّعناه كما هو (-2) الواجهة تعرض سهم نازل أحمر على
// موظف تحسّن — وهذا يظلمه ويخليه يفقد الثقة بالشاشة كلها.
//
// ⚠️ الي ما كان موجود بالفترة السابقة (موظف جديد) ياخذ delta = 0 مو
// رقم ضخم: موظف داوم أسبوع واحد ما ينعرض إنه «قفز ٩ مراكز».
func applyDeltas(current, previous []model.KpiLeaderboardEntry) {
	prevPoints := make(map[string]int, len(previous))
	prevRank := make(map[string]int, len(previous))
	for i, p := range previous {
		prevPoints[p.EmployeeID] = p.Points
		prevRank[p.EmployeeID] = i + 1
	}
	for i := range current {
		id := current[i].EmployeeID
		if old, ok := prevPoints[id]; ok {
			current[i].PointsDelta = current[i].Points - old
		}
		if oldRank, ok := prevRank[id]; ok {
			current[i].RankDelta = oldRank - (i + 1) // موجب = تقدّم
		}
	}
}
