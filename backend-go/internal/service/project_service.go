package service

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/lib/pq"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

var projectStages = []string{
	"1. اتصال بالزبون",
	"2. مرحلة الكشف",
	"3. عرض السعر",
	"4. العقد",
	"5. البدء بالتنفيذ",
	"✅ مكتمل",
	"❌ مرفوض",
}

// ProjectExecutionStage المرحلة الي تفتح الحجز عند إداري الحجوزات:
// خلصت إجراءات المشروع (اتصال، كشف، عرض سعر، عقد) وصار جاهز للتنفيذ،
// و**نفس كادر الشد** هو الي راح ينفّذ — فيرجع للتنسيق العادي.
const ProjectExecutionStage = "5. البدء بالتنفيذ"

// stageUnlocksBooking المكتمل يفتح بعد، لو المشروع قفز المرحلة.
func stageUnlocksBooking(stage string) bool {
	return stage == ProjectExecutionStage || stage == "✅ مكتمل"
}

// BookingUnlocker يخلي خدمة المشاريع تفتح الحجز بدون ما تعتمد على
// خدمة الحجوزات كاملة — نفس أسلوب بقية الحقن بالمشروع.
type BookingUnlocker interface {
	MarkProjectExecution(bookingID string) error
}

type ProjectService struct {
	repo *repository.ProjectRepository
	// bookings يُحقن من main — يفتح الحجز لما المشروع يوصل التنفيذ
	bookings BookingUnlocker
	// vipRepo يُحقن من main — أي مشروع ينضاف يرحّل صاحبه للشخصيات المهمة
	vipRepo      *repository.VipCustomerRepository
	customerRepo *repository.CustomerRepository
}

func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

// SetVipRepositories يربط ترحيل الشخصيات المهمة (يُنادى من main).
func (s *ProjectService) SetVipRepositories(vip *repository.VipCustomerRepository, customers *repository.CustomerRepository) {
	s.vipRepo = vip
	s.customerRepo = customers
}

// markProjectOwnerAsVip يرحّل صاحب المشروع للشخصيات المهمة.
//
// الناس الي يطلبون مشاريع دائماً ناس مهمين — فما ننتظر أحد يعلّمهم يدوياً.
// نلاقي الزبون من الحجز المرتبط، وإذا ماكو حجز نلاقيه (أو ننشئه) بالهاتف.
// فشل الترحيل ما يفشّل إنشاء المشروع — المشروع أهم.
func (s *ProjectService) markProjectOwnerAsVip(p *model.Project, createdBy *string) {
	if s.vipRepo == nil || s.customerRepo == nil || createdBy == nil || *createdBy == "" {
		return
	}
	customerID := ""
	if p.BookingID != nil && *p.BookingID != "" {
		if id, err := s.repo.CustomerIDForBooking(*p.BookingID); err == nil {
			customerID = id
		}
	}
	if customerID == "" && p.Phone != nil && *p.Phone != "" {
		if c, err := s.customerRepo.FindOrCreateByPhone(*p.Phone, p.Name); err == nil && c != nil {
			customerID = c.ID
		} else if err != nil {
			log.Printf("ترحيل VIP للمشروع %s: تعذر إيجاد/إنشاء الزبون: %v", p.Code, err)
		}
	}
	if customerID == "" {
		return
	}
	// المنصب: الشخص المرجعي بالمشروع هو أقرب شي عندنا لمنصب الزبون
	if err := s.vipRepo.MarkFromProject(p.ID, customerID, p.Name, p.RefPerson, *createdBy); err != nil {
		log.Printf("ترحيل VIP للمشروع %s: %v", p.Code, err)
	}
}

func computeProjectStats(projects []model.Project) model.ProjectStats {
	var stats model.ProjectStats
	for _, p := range projects {
		s := p.Stage
		switch {
		case strings.Contains(s, "اتصال"):
			stats.Contact++
		case strings.Contains(s, "كشف"):
			stats.Survey++
		case strings.Contains(s, "سعر"):
			stats.Price++
		case strings.Contains(s, "عقد"):
			stats.Contract++
		case strings.Contains(s, "تنفيذ"):
			stats.Execute++
		case strings.Contains(s, "مكتمل"):
			stats.Done++
		case strings.Contains(s, "مرفوض"):
			stats.Rejected++
		}
	}
	return stats
}

func (s *ProjectService) List() (*model.ProjectListResponse, error) {
	projects, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	return &model.ProjectListResponse{
		Projects: projects,
		Stats:    computeProjectStats(projects),
		Stages:   projectStages,
	}, nil
}

// ListDelegatedTo مشاريع مُسلَّمة لموظف معيّن — نفس شكل القائمة العامة (مع
// إحصائياتها والمراحل) حتى الواجهة تعرضها بنفس طريقة إدارة المشاريع بالضبط.
func (s *ProjectService) ListDelegatedTo(employeeID string) (*model.ProjectListResponse, error) {
	projects, err := s.repo.ListDelegatedTo(employeeID)
	if err != nil {
		return nil, err
	}
	return &model.ProjectListResponse{
		Projects: projects,
		Stats:    computeProjectStats(projects),
		Stages:   projectStages,
	}, nil
}

// Statistics إحصائيات المشاريع الكاملة: نظرة عامة + سطر لكل مشروع بقيمته +
// سطر لكل موظف يبيّن دوره الفعلي (أضاف / طلع كشف / كان مسؤول / استلم مشروع).
func (s *ProjectService) Statistics() (*model.ProjectStatisticsResponse, error) {
	projects, err := s.repo.ProjectValueRows()
	if err != nil {
		return nil, err
	}
	employees, err := s.repo.ProjectEmployeeStats()
	if err != nil {
		return nil, err
	}

	ov := model.ProjectStatisticsOverview{TotalProjects: len(projects)}
	for _, p := range projects {
		val := 0.0
		if p.PriceValue != nil {
			val = *p.PriceValue
			ov.TotalValue += val
			ov.PricedProjects++
		}
		if p.DelegatedToName != nil {
			ov.DelegatedCount++
		}
		if p.HasSurvey {
			ov.SurveysFilled++
		}
		switch {
		case strings.Contains(p.Stage, "مكتمل"):
			ov.CompletedCount++
			ov.CompletedValue += val
		case strings.Contains(p.Stage, "مرفوض"):
			ov.RejectedCount++
		default:
			ov.ActiveCount++
			ov.InProgressValue += val
		}
	}
	if ov.PricedProjects > 0 {
		ov.AverageValue = ov.TotalValue / float64(ov.PricedProjects)
	}
	// نعيد استخدام نفس دالة توزيع المراحل المستخدمة بصفحة المشاريع حتى
	// الرقمين ما يختلفون بين الصفحتين
	stageSource := make([]model.Project, 0, len(projects))
	for _, p := range projects {
		stageSource = append(stageSource, model.Project{Stage: p.Stage})
	}
	ov.StageBreakdown = computeProjectStats(stageSource)

	return &model.ProjectStatisticsResponse{
		Overview:  ov,
		Projects:  projects,
		Employees: employees,
	}, nil
}

// Delegate يسلّم المشروع لموظف، أو يسحب التسليم لو employeeID فاضي.
func (s *ProjectService) Delegate(projectID, employeeID string, byEmployeeID *string, note string) (*model.Project, error) {
	var target *string
	if employeeID != "" {
		target = &employeeID
	}
	return s.repo.Delegate(projectID, target, byEmployeeID, note)
}

// HasAnyDelegation صحيح لو الموظف موجّه له أي مشروع.
func (s *ProjectService) HasAnyDelegation(employeeID string) (bool, error) {
	return s.repo.HasAnyDelegation(employeeID)
}

// IsDelegatedTo يفحص ملكية التوجيه — يستعملها الراوت قبل ما يسمح بالتعديل.
func (s *ProjectService) IsDelegatedTo(projectID, employeeID string) (bool, error) {
	return s.repo.IsDelegatedTo(projectID, employeeID)
}

// DelegationLog سجل التسليم لمشروع معيّن أو للكل.
func (s *ProjectService) DelegationLog(projectID string) ([]model.ProjectDelegationLogEntry, error) {
	return s.repo.DelegationLog(projectID)
}

func (s *ProjectService) Create(req model.CreateProjectRequest, createdBy *string) (*model.Project, error) {
	if req.Name == "" {
		return nil, errors.New("اسم المؤسسة مطلوب")
	}
	priority := "عادي"
	if req.Priority != nil {
		priority = *req.Priority
	}

	// الكود يُبنى من أكبر رقم موجود مو من عدد الصفوف، ومع ذلك يبقى احتمال تضارب
	// لو موظفين ضافوا مشروع بنفس اللحظة — فنعيد المحاولة عدة مرات على خطأ
	// التكرار (23505) بدل ما نطلّع الخطأ بوجه المستخدم.
	next, err := s.repo.NextCodeNumber()
	if err != nil {
		return nil, err
	}
	var lastErr error
	for attempt := 0; attempt < 10; attempt++ {
		code := fmt.Sprintf("PRJ-%04d", next+attempt)
		p, err := s.repo.Create(code, req.Name, req.Rep, req.Phone, req.Location, req.MapLatitude, req.MapLongitude, req.WorkType, req.RefPerson, priority, req.DeliveryDate, req.BookingID,
			emptyToNil(req.ResponsibleEmployeeID), emptyToNil(req.SurveyorEmployeeID), req.LocationUrl, createdBy)
		if err == nil {
			s.markProjectOwnerAsVip(p, createdBy)
			return p, nil
		}
		lastErr = err
		if !isDuplicateKeyErr(err) {
			return nil, err
		}
	}
	return nil, lastErr
}

// isDuplicateKeyErr يميّز خطأ تكرار المفتاح الفريد بـPostgres (SQLSTATE 23505).
func isDuplicateKeyErr(err error) bool {
	if err == nil {
		return false
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return strings.Contains(err.Error(), "23505")
}

// emptyToNil يحوّل "" لـnil حتى لا نحاول نخزن سترنغ فاضي بعمود مفتاح أجنبي
// (الواجهة ترسل "" لما المستخدم ما يختار موظف من القائمة المنسدلة).
func emptyToNil(v *string) *string {
	if v == nil || *v == "" {
		return nil
	}
	return v
}

func (s *ProjectService) Get(id string) (*model.Project, error) {
	return s.repo.GetByID(id)
}

func (s *ProjectService) Update(id string, req model.UpdateProjectRequest) (*model.Project, error) {
	req.ResponsibleEmployeeID = emptyToNil(req.ResponsibleEmployeeID)
	req.SurveyorEmployeeID = emptyToNil(req.SurveyorEmployeeID)
	p, err := s.repo.Update(id, req)
	if err != nil {
		return nil, err
	}
	// وصول المشروع للتنفيذ يفتح حجزه عند إداري الحجوزات. الفشل هنا ما
	// يرجّع خطأ — تحديث المشروع نجح فعلاً، وما يصير نرجّع فشل ونخلي
	// مدير المشاريع يعيد المحاولة ويغيّر المرحلة مرتين.
	if s.bookings != nil && p != nil && p.BookingID != nil && req.Stage != nil && stageUnlocksBooking(*req.Stage) {
		if err := s.bookings.MarkProjectExecution(*p.BookingID); err != nil {
			log.Printf("[project] تعذر فتح حجز المشروع %s: %v", id, err)
		}
	}
	return p, nil
}

// SetBookingUnlocker يربط فتح الحجز (يُنادى من main) — بعد البناء حتى
// ما يصير اعتماد دائري بين خدمة المشاريع وخدمة الحجوزات.
func (s *ProjectService) SetBookingUnlocker(b BookingUnlocker) { s.bookings = b }

func (s *ProjectService) Delete(id string) error {
	return s.repo.Delete(id)
}
