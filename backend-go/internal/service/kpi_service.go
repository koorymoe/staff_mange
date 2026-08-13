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

type KpiService struct {
	repo          *repository.KpiRepository
	employees     *repository.EmployeeRepository
	notifications *repository.NotificationRepository
	announcements announcementPublisher
}

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

	var monthAgo string
	var oldTopID string
	if empErr == nil && employee != nil {
		monthAgo = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
		if board, err := s.repo.RoleLeaderboard(employee.Role, monthAgo, ""); err == nil && len(board) > 0 {
			oldTopID = board[0].EmployeeID
		}
	}

	deductionAmount := float64(*req.Points) * 10000
	eval, err := s.repo.Create(req.EmployeeID, req.EvaluatorID, *req.Points, req.Reason, deductionAmount)
	if err != nil {
		return nil, err
	}
	if *req.Points < 0 {
		_ = s.employees.SetTrainee(req.EmployeeID, true)
		reason := req.Reason
		if reason == "" {
			reason = "بدون سبب مذكور"
		}
		if s.notifications != nil {
			_ = s.notifications.Create(req.EmployeeID, "kpi_deduction",
				"⚠️ تم خصم "+strconv.Itoa(-*req.Points)+" نقطة من رصيدك (السبب: "+reason+")")
		}
		// نشر المخالفة بلوحة الإعلانات — بطلب المدير وقت التسجيل، ولمدة
		// ثلاثة أيام بس. فشل النشر ما يلغي التقييم، التقييم انسجل أصلاً.
		if req.Announce && s.announcements != nil && empErr == nil && employee != nil {
			body := "⚠️ مخالفة: " + employee.Name + " — خصم " +
				strconv.Itoa(-*req.Points) + " نقطة (" + reason + ")"
			if _, err := s.announcements.Create(body, req.EvaluatorID, model.AnnouncementPenaltyDays); err != nil {
				log.Printf("نشر مخالفة %s بلوحة الإعلانات: %v", req.EmployeeID, err)
			}
		}
	}

	if empErr == nil && employee != nil && s.notifications != nil {
		if board, err := s.repo.RoleLeaderboard(employee.Role, monthAgo, ""); err == nil && len(board) > 0 {
			newTop := board[0]
			if newTop.EmployeeID != oldTopID {
				_ = s.notifications.CreateForRole(employee.Role, "kpi_leaderboard",
					"🏆 "+newTop.EmployeeName+" تصدر التصنيف الشهري بالمركز الأول!")
			}
		}
	}

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
func (s *KpiService) RoleLeaderboard(role string) (*model.RoleKpiLeaderboard, error) {
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
