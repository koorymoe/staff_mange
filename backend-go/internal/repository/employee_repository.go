package repository

import (
	"sort"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

type EmployeeRepository struct {
	db *sqlx.DB
}

func NewEmployeeRepository(db *sqlx.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

// List يرجع الموظفين النشطين فقط (يستثني المؤرشفين والمحذوفين والموقوفين) —
// هذا افتراضي لكل واجهات النظام العادية.
func (r *EmployeeRepository) List() ([]model.Employee, error) {
	employees := []model.Employee{}
	// حساب المالك (OWNER) ما يطلع بأي قائمة موظفين عادية أبداً — حساب مخفي تماماً
	if err := r.db.Select(&employees, `SELECT * FROM "Employee" WHERE status NOT IN ('ARCHIVED', 'DELETED', 'SUSPENDED') AND role != 'OWNER' ORDER BY name ASC`); err != nil {
		return nil, err
	}
	if err := r.attachSkills(employees); err != nil {
		return nil, err
	}
	return employees, nil
}

// ListArchived يرجع المؤرشفين والمحذوفين والموقوفين تلقائياً (بسبب محاولات
// اختراق) — للأدمن/المالك حصراً، لمراجعة تاريخهم أو استرجاعهم عند الحاجة.
func (r *EmployeeRepository) ListArchived() ([]model.Employee, error) {
	employees := []model.Employee{}
	if err := r.db.Select(&employees, `SELECT * FROM "Employee" WHERE status IN ('ARCHIVED', 'DELETED', 'SUSPENDED') ORDER BY name ASC`); err != nil {
		return nil, err
	}
	if err := r.attachSkills(employees); err != nil {
		return nil, err
	}
	return employees, nil
}

func (r *EmployeeRepository) FindByID(id string) (*model.Employee, error) {
	var e model.Employee
	if err := r.db.Get(&e, `SELECT * FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	skills, err := r.SkillsForEmployee(e.ID)
	if err != nil {
		return nil, err
	}
	e.Skills = skills
	return &e, nil
}

// StatusByID استعلام خفيف يرجع حالة الموظف بس (بدون بقية بياناته) — يستخدمه
// RequireAuth بكل طلب يتحقق الحساب لسه فعّال (مو موقوف SUSPENDED) حتى لو
// التوكن نفسه لسه صالح.
func (r *EmployeeRepository) StatusByID(id string) (string, error) {
	var status string
	err := r.db.Get(&status, `SELECT status FROM "Employee" WHERE id = $1`, id)
	return status, err
}

// StatusAndRoleByID استعلام خفيف يرجع الحالة والدور الحاليين — يستخدمه RequireAuth
// بكل طلب حتى الدور المعتمد بالفحص يكون دايماً الدور الحقيقي بقاعدة البيانات
// الآن، مو الدور القديم المخزّن جوا التوكن وقت تسجيل الدخول (ثغرة كانت موجودة:
// تنزيل موظف من ADMIN لدور عادي ما يبطل صلاحياته إلا بعد انتهاء التوكن، لغاية
// ١٢ ساعة).
func (r *EmployeeRepository) StatusAndRoleByID(id string) (status string, role string, err error) {
	row := struct {
		Status string `db:"status"`
		Role   string `db:"role"`
	}{}
	err = r.db.Get(&row, `SELECT status, role FROM "Employee" WHERE id = $1`, id)
	return row.Status, row.Role, err
}

// IsLeaderFreshByID استعلام خفيف يرجع "isLeader" الحالية من قاعدة البيانات مباشرة —
// تُستخدم بگيت (middleware) صيانة الأجهزة العامة وجرد الفريق حتى القرار المعتمد
// دايماً يكون حسب قاعدة البيانات الآن، مو أي دور/علم قديم مخزّن جوا التوكن وقت
// تسجيل الدخول (نفس منطق StatusAndRoleByID تماماً بس لعلم isLeader).
func (r *EmployeeRepository) IsLeaderFreshByID(id string) (bool, error) {
	var isLeader bool
	err := r.db.Get(&isLeader, `SELECT "isLeader" FROM "Employee" WHERE id = $1`, id)
	return isLeader, err
}

// RecordAuthzViolation تسجل محاولة وصول مرفوضة (طلب عملية الموظف مو مخوّل لها
// فعلياً حسب دوره/صلاحياته الحاليين) — بس تسجيل وعدّاد، بدون أي إيقاف تلقائي
// للحساب. الإيقاف التلقائي (قبل هذا التعديل) كان خطر حقيقي: أي خطأ بالواجهة
// أو صلاحية ناقصة بالخطأ يقفل حساب موظف شرعي بمنتصف شغله بدون تدخل بشري —
// هسه بس ينبّه الإدارة (بعد grantRolePermission threshold) وتقرر هي.
func (r *EmployeeRepository) RecordAuthzViolation(id string) (violations int, err error) {
	err = r.db.Get(&violations, `
		UPDATE "Employee" SET "authzViolations" = "authzViolations" + 1
		WHERE id = $1
		RETURNING "authzViolations"
	`, id)
	return violations, err
}

// NameByID استعلام خفيف يرجع اسم الموظف بس — يستخدمه تنبيه محاولات الوصول
// المرفوضة المتكررة حتى الإدارة تعرف مين بالضبط تحتاج تراجعه.
func (r *EmployeeRepository) NameByID(id string) (string, error) {
	var name string
	err := r.db.Get(&name, `SELECT name FROM "Employee" WHERE id = $1`, id)
	return name, err
}

func (r *EmployeeRepository) FindByUsername(username string) (*model.Employee, error) {
	var e model.Employee
	if err := r.db.Get(&e, `SELECT * FROM "Employee" WHERE username = $1`, username); err != nil {
		return nil, err
	}
	skills, err := r.SkillsForEmployee(e.ID)
	if err != nil {
		return nil, err
	}
	e.Skills = skills
	return &e, nil
}

func (r *EmployeeRepository) SetTrainee(id string, isTrainee bool) error {
	_, err := r.db.Exec(`UPDATE "Employee" SET "isTrainee" = $2 WHERE id = $1`, id, isTrainee)
	return err
}

// SetPassword يحدّث كلمة مرور موظف معيّن (هاش جاهز مسبقاً) — تغيير ذاتي من
// إعدادات الموظف نفسه.
func (r *EmployeeRepository) SetPassword(id, hashedPassword string) error {
	_, err := r.db.Exec(`UPDATE "Employee" SET password = $2 WHERE id = $1`, id, hashedPassword)
	return err
}

// SetCommandPassword باسورد مركز القيادة — منفصل عن العادي.
func (r *EmployeeRepository) SetCommandPassword(id, hashedPassword string) error {
	_, err := r.db.Exec(`
		UPDATE "Employee"
		SET "commandPassword" = $2, "commandPasswordSetAt" = now()
		WHERE id = $1`, id, hashedPassword)
	return err
}

// SetAttendanceIcon يحدّث الأيقونة الشخصية لموظف معيّن — بعد موافقة الإداري
// على طلب تغيير الرمز.
func (r *EmployeeRepository) SetAttendanceIcon(id, icon string) error {
	_, err := r.db.Exec(`UPDATE "Employee" SET "attendanceIcon" = $2 WHERE id = $1`, id, icon)
	return err
}

func (r *EmployeeRepository) Create(e *model.Employee) error {
	_, err := r.db.NamedExec(`
		INSERT INTO "Employee" (id, name, certificate, position, phone, username, password, "jobTitle", salary, shift, "shiftStart", "shiftEnd", role, division)
		VALUES (:id, :name, :certificate, :position, :phone, :username, :password, :jobTitle, :salary, :shift, :shiftStart, :shiftEnd, :role, :division)
	`, e)
	return err
}

func (r *EmployeeRepository) Update(e *model.Employee) error {
	_, err := r.db.NamedExec(`
		UPDATE "Employee" SET
			name = :name,
			certificate = :certificate,
			position = :position,
			phone = :phone,
			status = :status,
			role = :role,
			"onDuty" = :onDuty,
			username = :username,
			password = COALESCE(NULLIF(:password, ''), password),
			"hasDrivingLicense" = :hasDrivingLicense,
			"hasSafetyCertificate" = :hasSafetyCertificate,
			"isLeader" = :isLeader,
			"isTrainee" = :isTrainee,
			salary = :salary,
			shift = :shift,
			"shiftStart" = :shiftStart,
			"shiftEnd" = :shiftEnd,
			"monthlyLeaves" = :monthlyLeaves,
			"jobTitle" = :jobTitle,
			"authzViolations" = :authzViolations,
			-- ملف الموارد البشرية المنقول من نظام الطاقة الشمسية
			department = :department,
			"hireDate" = :hireDate,
			"experienceYears" = :experienceYears,
			"lastReview" = :lastReview,
			"careerStatus" = :careerStatus,
			"jobLevel" = :jobLevel,
			"nextRole" = :nextRole,
			"trainingNeeds" = :trainingNeeds
		WHERE id = :id
	`, e)
	return err
}

// Supervisors يرجّع تيم ليدرز ومدراء المشاريع النشطين المؤهلين للإشراف على تكليف الفنيين
func (r *EmployeeRepository) Supervisors() ([]model.Employee, error) {
	employees := []model.Employee{}
	err := r.db.Select(&employees, `
		SELECT * FROM "Employee"
		WHERE status = 'ACTIVE' AND (role = 'PROJECT_MANAGER' OR "isLeader" = true)
		ORDER BY name ASC
	`)
	return employees, err
}

// MatchForService يرجّع الكوادر الي ينفع تنكلّف بخدمة معيّنة، مع علامة
// إذا يمتلكون مهارتها.
//
// جان بيها قيدين خربوا التنسيق:
//
//	١. `"onDuty" = true` — يعني الموظف ما يظهر إلا إذا مسجّل حضور بهاي
//	   اللحظة. والمنسّق يوزّع حجوزات باچر وبعده، فيفتح الحجز ويلكه
//	   «لا يوجد موظف متاح» لمجرد إن الفني ما سجّل دوام هسه.
//
//	٢. `role = 'TECHNICIAN'` — يشيل المهندسين وأي كادر ثاني يمتلك مهارة
//	   الخدمة فعلاً. والمهارة هي المقياس الصح مو المسمّى الوظيفي.
//
// هسه: الفنيين والمهندسين، وأي موظف عنده مهارة هاي الخدمة مهما جان
// دوره. والي عنده المهارة يطلع أول، وبعده الباقي.
func (r *EmployeeRepository) MatchForService(serviceID string) ([]model.Employee, error) {
	employees := []model.Employee{}
	if err := r.db.Select(&employees, `
		SELECT e.* FROM "Employee" e
		WHERE e.status = 'ACTIVE'
		  AND (
		    e.role IN ('TECHNICIAN', 'TECHNICAL', 'ENGINEER')
		    OR EXISTS (
		      SELECT 1 FROM "EmployeeSkill" es
		      JOIN "Skill" sk ON sk.id = es."skillId"
		      WHERE es."employeeId" = e.id AND es."canPerform" = true AND sk."serviceId" = $1
		    )
		  )
		ORDER BY e.name ASC
	`, serviceID); err != nil {
		return nil, err
	}
	if err := r.attachSkills(employees); err != nil {
		return nil, err
	}
	for i := range employees {
		hasSkill := false
		for _, s := range employees[i].Skills {
			if s.CanPerform && s.Skill != nil && s.Skill.ServiceID == serviceID {
				hasSkill = true
				break
			}
		}
		employees[i].HasRequiredSkill = &hasSkill
	}
	// الي عنده المهارة أول — المنسّق يشوف المناسب كدامه بدل ما يدوّر
	sort.SliceStable(employees, func(a, b int) bool {
		return employees[a].HasRequiredSkill != nil && *employees[a].HasRequiredSkill &&
			(employees[b].HasRequiredSkill == nil || !*employees[b].HasRequiredSkill)
	})
	return employees, nil
}

// attachSkills يعبّي مهارات كل الموظفين سوه باستعلامين، مهما جان عددهم.
//
// SkillsForEmployee تسوي استعلام للمهارات + استعلام لكل مهارة حتى تجيب
// تفاصيلها. نداؤها بحلقة على الموظفين معناها استعلام لكل موظف على الأقل،
// وقائمة الموظفين تنطلب بشاشات كثيرة (التنسيق، التعيين، الإحصاءات) —
// فالتأخير ينضرب بعدد الشاشات.
func (r *EmployeeRepository) attachSkills(employees []model.Employee) error {
	if len(employees) == 0 {
		return nil
	}
	ids := make([]string, 0, len(employees))
	for i := range employees {
		employees[i].Skills = []model.EmployeeSkillDetail{}
		ids = append(ids, employees[i].ID)
	}

	// المهارة وتفاصيلها بضربة وحدة (JOIN) بدل استعلام لكل مهارة
	type row struct {
		model.EmployeeSkillDetail `db:",inline"`
		SkillID2                  *string    `db:"s_id"`
		SkillName                 *string    `db:"s_name"`
		SkillServiceID            *string    `db:"s_serviceId"`
		SkillCreatedAt            *time.Time `db:"s_createdAt"`
	}
	rows := []row{}
	if err := r.db.Select(&rows, `
		SELECT es.*, s.id AS "s_id", s.name AS "s_name", s."serviceId" AS "s_serviceId", s."createdAt" AS "s_createdAt"
		FROM "EmployeeSkill" es
		LEFT JOIN "Skill" s ON s.id = es."skillId"
		WHERE es."employeeId" = ANY($1)`, pq.Array(ids)); err != nil {
		return err
	}

	byEmployee := map[string][]model.EmployeeSkillDetail{}
	for _, rw := range rows {
		d := rw.EmployeeSkillDetail
		if rw.SkillID2 != nil {
			d.Skill = &model.Skill{ID: *rw.SkillID2}
			if rw.SkillName != nil {
				d.Skill.Name = *rw.SkillName
			}
			if rw.SkillServiceID != nil {
				d.Skill.ServiceID = *rw.SkillServiceID
			}
			if rw.SkillCreatedAt != nil {
				d.Skill.CreatedAt = *rw.SkillCreatedAt
			}
		}
		byEmployee[d.EmployeeID] = append(byEmployee[d.EmployeeID], d)
	}
	for i := range employees {
		if list := byEmployee[employees[i].ID]; list != nil {
			employees[i].Skills = list
		}
	}
	return nil
}

// SkillsForEmployee يجلب كل مهارات موظف مع تفاصيل المهارة والخدمة المرتبطة بيها
func (r *EmployeeRepository) SkillsForEmployee(employeeID string) ([]model.EmployeeSkillDetail, error) {
	skills := []model.EmployeeSkillDetail{}
	if err := r.db.Select(&skills, `SELECT * FROM "EmployeeSkill" WHERE "employeeId" = $1`, employeeID); err != nil {
		return nil, err
	}
	for i := range skills {
		var skill model.Skill
		if err := r.db.Get(&skill, `SELECT * FROM "Skill" WHERE id = $1`, skills[i].SkillID); err == nil {
			skills[i].Skill = &skill
		}
	}
	return skills, nil
}

// CountDistinctServicesKnown يرجّع عدد الخدمات المميزة الي الموظف يعرف عليها
// مهارة واحدة على الأقل (canPerform = true) — يُستخدم بصفحة إحصائيات الموظفين
// الشهرية بدل تكرار نقاط الكي بي اي (موجودة أصلاً بصفحة التقديرات).
func (r *EmployeeRepository) CountDistinctServicesKnown(employeeID string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(DISTINCT sk."serviceId") FROM "EmployeeSkill" es
		JOIN "Skill" sk ON sk.id = es."skillId"
		WHERE es."employeeId" = $1 AND es."canPerform" = true
	`, employeeID)
	return count, err
}

// SkillDivisions يرجّع "division" الخدمة (Service) المالكة لكل مهارة من قائمة
// معرّفات المهارات المعطاة — يستخدمها EmployeeService.SetSkills حتى يتأكد إن
// كل مهارة يراد إسنادها تنتمي لنفس شعبة الموظف (ENGINEERING/DECOR) قبل الحفظ.
func (r *EmployeeRepository) SkillDivisions(skillIDs []string) (map[string]string, error) {
	result := make(map[string]string, len(skillIDs))
	if len(skillIDs) == 0 {
		return result, nil
	}
	rows := []struct {
		ID       string `db:"id"`
		Division string `db:"division"`
	}{}
	query, args, err := sqlx.In(`
		SELECT sk.id AS id, sv.division AS division
		FROM "Skill" sk JOIN "Service" sv ON sv.id = sk."serviceId"
		WHERE sk.id IN (?)
	`, skillIDs)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	if err := r.db.Select(&rows, query, args...); err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ID] = row.Division
	}
	return result, nil
}

// SetSkills يستبدل مهارات الموظف بالكامل (حذف القديم وإدخال الجديد بنفس المعاملة)
func (r *EmployeeRepository) SetSkills(employeeID string, skills []model.EmployeeSkillInput) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM "EmployeeSkill" WHERE "employeeId" = $1`, employeeID); err != nil {
		return err
	}
	for _, s := range skills {
		if _, err := tx.Exec(`
			INSERT INTO "EmployeeSkill" (id, "employeeId", "skillId", "canPerform")
			VALUES (gen_random_uuid()::text, $1, $2, $3)
		`, employeeID, s.SkillID, s.CanPerform); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// LinkHistoricalRecords يربط سجلات تاريخية (حجوزات/شكاوى مستوردة من نظام قديم بس
// بالاسم النصي، بدون حساب فعلي وقتها) بحساب موظف حالي بنفس الاسم — يستخدمها الأدمن
// يدوياً لما موظف قديم يرجع ويصير له حساب جديد بنفس اسمه بالضبط.
func (r *EmployeeRepository) LinkHistoricalRecords(employeeID, employeeName string) (bookingsLinked int, complaintsLinked int, err error) {
	res, err := r.db.Exec(`
		UPDATE "Booking" SET "confirmedByEmployeeId" = $1
		WHERE "confirmedByName" = $2 AND "confirmedByEmployeeId" IS NULL
	`, employeeID, employeeName)
	if err != nil {
		return 0, 0, err
	}
	if n, e := res.RowsAffected(); e == nil {
		bookingsLinked = int(n)
	}

	res, err = r.db.Exec(`
		UPDATE "Complaint" SET "relatedEmployeeId" = $1
		WHERE "relatedEmployeeName" = $2 AND "relatedEmployeeId" IS NULL
	`, employeeID, employeeName)
	if err != nil {
		return bookingsLinked, 0, err
	}
	if n, e := res.RowsAffected(); e == nil {
		complaintsLinked = int(n)
	}

	return bookingsLinked, complaintsLinked, nil
}
