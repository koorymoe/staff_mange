package service

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type EmployeeService struct {
	repo          *repository.EmployeeRepository
	inventoryRepo *repository.InventoryRepository
}

func NewEmployeeService(repo *repository.EmployeeRepository) *EmployeeService {
	return &EmployeeService{repo: repo}
}

// SetInventoryRepository يربط مستودع المخزون بعد الإنشاء (تجنباً لتغيير كل نقاط
// استدعاء NewEmployeeService الحالية) — يُستخدم فقط لتطبيق العدة القياسية
// (PersonalToolTemplateItem) تلقائياً على أي موظف جديد وقت الإنشاء.
func (s *EmployeeService) SetInventoryRepository(inventoryRepo *repository.InventoryRepository) {
	s.inventoryRepo = inventoryRepo
}

func (s *EmployeeService) List() ([]model.Employee, error) {
	return s.repo.List()
}

func (s *EmployeeService) ListArchived() ([]model.Employee, error) {
	return s.repo.ListArchived()
}

func (s *EmployeeService) Get(id string) (*model.Employee, error) {
	return s.repo.FindByID(id)
}

func (s *EmployeeService) Create(req model.CreateEmployeeRequest) (*model.Employee, error) {
	if req.Name == "" {
		return nil, errors.New("الاسم مطلوب")
	}

	role := "TECHNICIAN"
	if req.Role != nil && *req.Role != "" {
		role = *req.Role
	}
	if role == "OWNER" {
		return nil, errors.New("دور المالك محجوز لحساب واحد بس، ما ينمنح من الواجهة")
	}
	shift := "MORNING"
	if req.Shift != nil && *req.Shift != "" {
		shift = *req.Shift
	}

	// division: "ENGINEERING" (افتراضي، الكادر الحالي) أو "DECOR" (الكادر الجديد) —
	// أول سؤال بفورم إضافة موظف بالواجهة. أي قيمة ثانية غير الاثنتين مرفوضة صراحة
	// حتى ما ينزرع صف بقيمة غلط بالغلط.
	division := model.DivisionEngineering
	if req.Division != nil && *req.Division != "" {
		division = *req.Division
	}
	if division != model.DivisionEngineering && division != model.DivisionDecor {
		return nil, errors.New("الشعبة يجب أن تكون شعبة هندسية أو ديكور")
	}
	// موظفو شعبة الديكور يُمنحون دائماً دور TECHNICIAN العادي (نفس دور الفنيين
	// الحاليين) بغض النظر عن أي دور آخر يُرسل بالطلب — هذا يعيد استخدام 100%
	// من منطق الصلاحيات/التوجيه الحالي بدل بناء نموذج صلاحيات موازٍ، وصلاحيات
	// TECHNICIAN الافتراضية أصلاً محدودة جداً (RoleDefaultPermissions["TECHNICIAN"]
	// = "expenses" بس)، فيتحقق شرط "صفر صلاحيات تقريباً" مباشرة بدون كود إضافي.
	if division == model.DivisionDecor {
		role = "TECHNICIAN"
	}

	employee := &model.Employee{
		ID:          uuid.NewString(),
		Name:        req.Name,
		Certificate: req.Certificate,
		Position:    req.Position,
		Phone:       req.Phone,
		Username:    req.Username,
		JobTitle:    req.JobTitle,
		Salary:      req.Salary,
		Shift:       &shift,
		ShiftStart:  req.ShiftStart,
		ShiftEnd:    req.ShiftEnd,
		Role:        role,
		Division:    division,
		Skills:      []model.EmployeeSkillDetail{},
	}

	if req.Password != nil && *req.Password != "" {
		hashed, err := HashPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		employee.Password = &hashed
	}

	if err := s.repo.Create(employee); err != nil {
		return nil, err
	}
	// موظف جديد ياخذ العدة القياسية (PersonalToolTemplateItem) كاملة تلقائياً —
	// بدون أي خطوة يدوية من الإداري بعدها. فشل هالخطوة ما يوقف إنشاء الموظف
	// نفسه (تسجيل الموظف أهم)، بس نسجّل الخطأ بالسجلات.
	if s.inventoryRepo != nil {
		if err := s.inventoryRepo.ApplyPersonalToolTemplateToEmployee(employee.ID); err != nil {
			log.Printf("تحذير: تعذر تطبيق العدة القياسية على الموظف الجديد %s: %v", employee.ID, err)
		}
	}
	return employee, nil
}

func (s *EmployeeService) Update(id string, req model.UpdateEmployeeRequest) (*model.Employee, error) {
	employee, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		employee.Name = *req.Name
	}
	if req.Certificate != nil {
		employee.Certificate = req.Certificate
	}
	if req.Position != nil {
		employee.Position = req.Position
	}
	if req.Phone != nil {
		employee.Phone = req.Phone
	}
	if req.Status != nil {
		employee.Status = *req.Status
		// استرجاع حساب موقوف (SUSPENDED) يصفّر عداد محاولات الاختراق —
		// نعطيه بداية جديدة بعد ما الأدمن راجع الموضوع ووافق يرجّعه
		if *req.Status == "ACTIVE" {
			employee.AuthzViolations = 0
		}
	}
	if req.Role != nil {
		if *req.Role == "OWNER" && employee.Role != "OWNER" {
			return nil, errors.New("دور المالك محجوز لحساب واحد بس، ما ينمنح من الواجهة")
		}
		if *req.Role == "ENGINEER" && employee.Role != "ENGINEER" {
			if err := requireEngineeringSkills(employee); err != nil {
				return nil, err
			}
		}
		employee.Role = *req.Role
	}
	if req.OnDuty != nil {
		employee.OnDuty = *req.OnDuty
	}
	if req.Username != nil {
		employee.Username = req.Username
	}
	if req.HasDrivingLicense != nil {
		employee.HasDrivingLicense = *req.HasDrivingLicense
	}
	if req.HasSafetyCertificate != nil {
		employee.HasSafetyCertificate = *req.HasSafetyCertificate
	}
	if req.IsLeader != nil {
		employee.IsLeader = *req.IsLeader
	}
	if req.IsTrainee != nil {
		employee.IsTrainee = *req.IsTrainee
	}
	if req.Salary != nil {
		employee.Salary = req.Salary
	}
	if req.Shift != nil {
		employee.Shift = req.Shift
	}
	if req.ShiftStart != nil {
		employee.ShiftStart = req.ShiftStart
	}
	if req.ShiftEnd != nil {
		employee.ShiftEnd = req.ShiftEnd
	}
	if req.MonthlyLeaves != nil {
		employee.MonthlyLeaves = *req.MonthlyLeaves
	}
	if req.JobTitle != nil {
		employee.JobTitle = req.JobTitle
	}

	// ═══ ملف الموارد البشرية ═══
	if req.Department != nil {
		employee.Department = req.Department
	}
	if req.HireDate != nil {
		if t, err := time.Parse("2006-01-02", *req.HireDate); err == nil {
			employee.HireDate = &t
		} else if *req.HireDate == "" {
			employee.HireDate = nil
		}
	}
	if req.ExperienceYears != nil {
		employee.ExperienceYears = req.ExperienceYears
	}
	if req.LastReview != nil {
		employee.LastReview = req.LastReview
	}
	if req.JobLevel != nil && *req.JobLevel >= 1 && *req.JobLevel <= 10 {
		employee.JobLevel = *req.JobLevel
	}
	if req.NextRole != nil {
		employee.NextRole = req.NextRole
	}
	if req.TrainingNeeds != nil {
		employee.TrainingNeeds = req.TrainingNeeds
	}
	// الحالة الوظيفية تنحسب بالسيرفر من الخبرة والمستوى والتقييم — قاعدة
	// وحدة للكل. الإداري يقدر يفرض «تحت المراقبة» يدوياً، وهاي الوحيدة
	// الي ما تنحسب لأنها قرار مو نتيجة.
	if req.CareerStatus != nil && *req.CareerStatus == model.CareerWatched {
		employee.CareerStatus = model.CareerWatched
	} else {
		exp := 0.0
		if employee.ExperienceYears != nil {
			exp = *employee.ExperienceYears
		}
		review := ""
		if employee.LastReview != nil {
			review = *employee.LastReview
		}
		employee.CareerStatus = model.EvaluateCareerStatus(exp, employee.JobLevel, review)
	}

	if req.Password != nil && *req.Password != "" {
		hashed, err := HashPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		employee.Password = &hashed
	} else {
		empty := ""
		employee.Password = &empty
	}

	if err := s.repo.Update(employee); err != nil {
		return nil, err
	}
	// تغيير الدور أو صفة الليدر يغيّر الاستحقاق للعدة القياسية: فني ينتقل
	// للمبيعات لازم تنشال عدته، وموظف يصير ليدر لازم ياخذها. فشل المزامنة
	// ما يبطّل تعديل الموظف — نسجّله بس.
	if s.inventoryRepo != nil && (req.Role != nil || req.IsLeader != nil) {
		if err := s.inventoryRepo.SyncPersonalToolKitForEmployee(id); err != nil {
			log.Printf("تحذير: تعذرت مزامنة العدة القياسية للموظف %s: %v", id, err)
		}
	}
	return s.repo.FindByID(id)
}

func (s *EmployeeService) Supervisors() ([]model.Employee, error) {
	return s.repo.Supervisors()
}

// Match كوادر ينفع تنكلّف بخدمة. serviceID فاضي مسموح: الحجز الي ما
// إله خدمة محددة (مثلاً حجز صيانة) لازم يطلعله كادر بعد — قبل، كان
// يرجع خطأ فتختفي خانات الفنيين كلها من الحجز.
func (s *EmployeeService) Match(serviceID string) ([]model.Employee, error) {
	return s.repo.MatchForService(serviceID)
}

// validateSkillsDivision يتأكد إن كل مهارة يراد تفعيلها فعلياً (canPerform=true)
// للموظف تنتمي لنفس شعبته (ENGINEERING/DECOR) — يمنع مثلاً تفعيل مهارة "حدادة"
// (ديكور) لموظف شعبة هندسية أو العكس. الواجهة ترسل دائماً القائمة الكاملة لكل
// خدمات النظام (كل الشُّعب) بكل طلب (استبدال كامل)، بس canPerform=false لأي
// خدمة مو من شعبة الموظف — فالتحقق يقتصر على المهارات المفعّلة فعلاً حتى ما
// يرفض حفظ فورم عادي يحوي مهارات كل الشعب بقيمة false.
func (s *EmployeeService) validateSkillsDivision(employeeDivision string, skills []model.EmployeeSkillInput) error {
	var activeIDs []string
	for _, sk := range skills {
		if sk.CanPerform {
			activeIDs = append(activeIDs, sk.SkillID)
		}
	}
	if len(activeIDs) == 0 {
		return nil
	}
	divisions, err := s.repo.SkillDivisions(activeIDs)
	if err != nil {
		return err
	}
	for _, id := range activeIDs {
		div, ok := divisions[id]
		if !ok {
			return errors.New("مهارة غير موجودة: " + id)
		}
		if div != employeeDivision {
			return errors.New("لا يمكن إسناد مهارة من شعبة مختلفة عن شعبة الموظف")
		}
	}
	return nil
}

func (s *EmployeeService) SetSkills(employeeID string, req model.SetEmployeeSkillsRequest) (*model.Employee, error) {
	employee, err := s.repo.FindByID(employeeID)
	if err != nil {
		return nil, err
	}
	if err := s.validateSkillsDivision(employee.Division, req.Skills); err != nil {
		return nil, err
	}
	if err := s.repo.SetSkills(employeeID, req.Skills); err != nil {
		return nil, err
	}
	employee, err = s.repo.FindByID(employeeID)
	if err != nil {
		return nil, err
	}
	skills, err := s.repo.SkillsForEmployee(employeeID)
	if err != nil {
		return nil, err
	}
	employee.Skills = skills
	return employee, nil
}

// requireEngineeringSkills يتأكد إن الموظف عنده المهارات الأربع الأساسية (تصميم/تخطيط/
// تنفيذ/إشراف) فعّالة قبل ما نسمح نرفعه لدور "مهندس".
func requireEngineeringSkills(employee *model.Employee) error {
	has := make(map[string]bool, len(model.EngineeringSkillNames))
	for _, s := range employee.Skills {
		if s.CanPerform && s.Skill != nil {
			has[s.Skill.Name] = true
		}
	}
	var missing []string
	for _, name := range model.EngineeringSkillNames {
		if !has[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return errors.New("الموظف ما عنده مهارات الهندسة المطلوبة بعد: " + strings.Join(missing, "، "))
	}
	return nil
}

// LinkHistoricalRecords يربط سجلات تاريخية (حجوزات/شكاوى مستوردة بالاسم بس) بحساب
// الموظف هذا — يستخدمها الأدمن لما يسوي حساب لموظف قديم رجع للشركة.
func (s *EmployeeService) LinkHistoricalRecords(id string) (bookingsLinked int, complaintsLinked int, err error) {
	emp, err := s.repo.FindByID(id)
	if err != nil {
		return 0, 0, err
	}
	return s.repo.LinkHistoricalRecords(id, emp.Name)
}
