package service

import (
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type PermissionService struct {
	permissions *repository.PermissionRepository
	employees   *repository.EmployeeRepository
	seeded      bool
}

func NewPermissionService(permissions *repository.PermissionRepository, employees *repository.EmployeeRepository) *PermissionService {
	return &PermissionService{permissions: permissions, employees: employees}
}

func (s *PermissionService) ensureSeeded() error {
	if s.seeded {
		return nil
	}
	if err := s.permissions.EnsureSeeded(); err != nil {
		return err
	}
	s.seeded = true
	return nil
}

func (s *PermissionService) ListAll() ([]model.Permission, error) {
	if err := s.ensureSeeded(); err != nil {
		return nil, err
	}
	return s.permissions.ListAll()
}

func (s *PermissionService) ListForEmployee(employeeID string) ([]model.Permission, error) {
	return s.permissions.ListForEmployee(employeeID)
}

// EmployeesWithPermission يرجّع الموظفين الي يوصلون لصلاحية معيّنة (بالصلاحية
// نفسها أو بدور يعطيها) — لتعبئة القوائم المنسدلة بالواجهة.
func (s *PermissionService) EmployeesWithPermission(permissionName string, alsoRoles []string) ([]model.EmployeeBrief, error) {
	if err := s.ensureSeeded(); err != nil {
		return nil, err
	}
	return s.permissions.ListEmployeesWithPermission(permissionName, alsoRoles)
}

func (s *PermissionService) SetForEmployee(employeeID string, permissionIDs []string) ([]model.Permission, error) {
	if err := s.permissions.ReplaceForEmployee(employeeID, permissionIDs); err != nil {
		return nil, err
	}
	return s.permissions.ListForEmployee(employeeID)
}

func (s *PermissionService) ApplyDefaults(employeeID string) ([]model.Permission, error) {
	employee, err := s.employees.FindByID(employeeID)
	if err != nil {
		return nil, err
	}

	defaults := model.RoleDefaultPermissions[employee.Role]
	if len(defaults) == 0 {
		return []model.Permission{}, nil
	}

	if err := s.ensureSeeded(); err != nil {
		return nil, err
	}

	all, err := s.permissions.ListAll()
	if err != nil {
		return nil, err
	}

	defaultSet := make(map[string]bool, len(defaults))
	for _, name := range defaults {
		defaultSet[name] = true
	}

	var defaultIDs []string
	for _, p := range all {
		if defaultSet[p.Name] {
			defaultIDs = append(defaultIDs, p.ID)
		}
	}

	if err := s.permissions.AddMissingForEmployee(employeeID, defaultIDs); err != nil {
		return nil, err
	}

	return s.permissions.ListForEmployee(employeeID)
}

func (s *PermissionService) RoleDefaults() map[string][]string {
	return model.RoleDefaultPermissions
}

// MissingRoleDefault موظف ناقصه شي من صلاحيات دوره الافتراضية.
type MissingRoleDefault struct {
	EmployeeID   string   `json:"employeeId"`
	EmployeeName string   `json:"employeeName"`
	Role         string   `json:"role"`
	Missing      []string `json:"missing"`
	MissingLabel []string `json:"missingLabels"`
	// ⚠️ صلاحيات مذكورة بخريطة الأدوار وماكو إلها صف بقاعدة البيانات
	// أصلاً — هاي مو «موظف ناقصه»، هاي «الصلاحية مو منزرعة بالنظام».
	// الفرق مهم: الأولى تنحل بمنح، والثانية لازم تنزرع أول.
	NotSeeded []string `json:"notSeeded"`
}

// AuditRoleDefaults يكشف منو من الموظفين ناقصه صلاحيات دوره.
//
// ⚠️⚠️ ليش هاي موجودة: `RoleDefaultPermissions` **خريطة اقتراح مو
// قاعدة منفَّذة**. إنشاء موظف جديد ما يمنحه ولا صلاحية منها،
// والتطبيق يصير بضغطة يدوية من المدير (`ApplyDefaults`). واكو زرع
// بأثر رجعي بس لأزواج محددة انتذكّرها أحد.
//
// فالنتيجة: موظف بدور MONITOR أو FINANCE ممكن يكون **بلا** صلاحية
// «المالية» — والقائمة تعرضله «التدقيق اليومي» وكل طلباته تنرفض.
// وهاي ما تظهر بأي مكان: مو خطأ بالكود، بيانات ناقصة بحساب واحد.
//
// ⚠️ والفحص **قراءة بس** — ما يمنح ولا صلاحية لحاله. منح صلاحية
// لعشرات الحسابات دفعة وحدة قرار مالك مو قرار فحص.
func (s *PermissionService) AuditRoleDefaults() ([]MissingRoleDefault, error) {
	employees, err := s.employees.List()
	if err != nil {
		return nil, err
	}
	all, err := s.permissions.ListAll()
	if err != nil {
		return nil, err
	}
	labelOf := make(map[string]string, len(all))
	for _, p := range all {
		labelOf[p.Name] = p.Label
	}

	out := []MissingRoleDefault{}
	for _, e := range employees {
		// الموقوفون ما ينحسبون: حسابه مقفل أصلاً فنقص صلاحياته مو خلل.
		if e.Status != "ACTIVE" {
			continue
		}
		defaults := model.RoleDefaultPermissions[e.Role]
		if len(defaults) == 0 {
			continue
		}
		have, err := s.permissions.ListForEmployee(e.ID)
		if err != nil {
			return nil, err
		}
		haveSet := make(map[string]bool, len(have))
		for _, p := range have {
			haveSet[p.Name] = true
		}
		var missing, labels, notSeeded []string
		for _, name := range defaults {
			if haveSet[name] {
				continue
			}
			missing = append(missing, name)
			lbl, seeded := labelOf[name]
			if !seeded || lbl == "" {
				lbl = name
				notSeeded = append(notSeeded, name)
			}
			labels = append(labels, lbl)
		}
		if len(missing) > 0 {
			out = append(out, MissingRoleDefault{
				EmployeeID: e.ID, EmployeeName: e.Name, Role: e.Role,
				Missing: missing, MissingLabel: labels, NotSeeded: notSeeded,
			})
		}
	}
	return out, nil
}
