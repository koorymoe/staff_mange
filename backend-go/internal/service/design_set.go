package service

import (
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/errlog"
	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ 🏗️ صيت التصاميم (ع) 11-08 ═══
// «كل مشروع إله مجموعة تصاميم — مثلاً قطعة أرض من البناء والهيكل لآخر شي الطاقة
// الشمسية… كلشي بمكان واحد». الصيت = تصاميم مشروع واحد، مقسّمة على تخصصات؛ لكل
// تخصص: مصمم وموعد (يصير مهمة)، أدوات خارجية بزر، ملفات/روابط، ومراجعة واعتماد.
// المواقع الخارجية (تي بي لينك، داهوا…) ما تنطي API — فالمصمم يصمم هناك ويرجّع
// الشغل هنا (لصق/سحب/رابط مشاركة). الخطة ٢ (محاكاة جوّه النظام) مؤجلة: docs/design-studio-plan.md

// DesignDisciplines التخصصات بالترتيب — كود ← اسم.
type DesignDiscipline struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

var DesignDisciplines = []DesignDiscipline{
	{"ARCH", "معماري", "🏛️"},
	{"STRUCT", "إنشائي", "🧱"},
	{"TOPVIEW", "توب فيو (مسقط)", "🗺️"},
	{"VIEW3D", "ثري دي", "🧊"},
	{"INTERIOR", "ديكور داخلي", "🛋️"},
	{"CERAMIC", "سيراميك وأرضيات", "🟫"},
	{"WINDOWS", "شبابيك وألمنيوم", "🪟"},
	{"ELECTRICAL", "كهرباء", "⚡"},
	{"CAMERAS", "كاميرات مراقبة", "📹"},
	{"NETWORK", "شبكات وواي فاي", "📶"},
	{"AUDIO", "صوتيات", "🔊"},
	{"FIRE", "إنذار حريق", "🔥"},
	{"SMART", "أنظمة ذكية وأقفال", "🔐"},
	{"SOLAR", "طاقة شمسية", "☀️"},
	{"CONTRACTING", "مقاولات وجدول كميات", "🏗️"},
}

func validDiscipline(c string) bool {
	for _, d := range DesignDisciplines {
		if d.Code == c {
			return true
		}
	}
	return false
}

type DesignTool struct {
	ID         string `db:"id" json:"id"`
	Discipline string `db:"discipline" json:"discipline"`
	Name       string `db:"name" json:"name"`
	URL        string `db:"url" json:"url"`
	HowTo      string `db:"howTo" json:"howTo"`
	Sort       int    `db:"sort" json:"sort"`
	Active     bool   `db:"active" json:"active"`
}

type DesignSet struct {
	ID          string     `db:"id" json:"id"`
	Title       string     `db:"title" json:"title"`
	ProjectID   *string    `db:"projectId" json:"projectId"`
	ProjectCode *string    `db:"projectCode" json:"projectCode"`
	CustomerID  *string    `db:"customerId" json:"customerId"`
	Customer    *string    `db:"customerName" json:"customerName"`
	Notes       *string    `db:"notes" json:"notes"`
	CreatedBy   *string    `db:"createdByName" json:"createdByName"`
	CreatedAt   time.Time  `db:"createdAt" json:"createdAt"`
	ArchivedAt  *time.Time `db:"archivedAt" json:"archivedAt"`
	Sections    int        `db:"sections" json:"sections"`
	Approved    int        `db:"approved" json:"approved"`
	InReview    int        `db:"inReview" json:"inReview"`
	Late        int        `db:"late" json:"late"`
}

type DesignSection struct {
	ID           string       `db:"id" json:"id"`
	SetID        string       `db:"setId" json:"setId"`
	Discipline   string       `db:"discipline" json:"discipline"`
	Status       string       `db:"status" json:"status"` // TODO | IN_PROGRESS | REVIEW | APPROVED | RETURNED
	AssigneeID   *string      `db:"assigneeId" json:"assigneeId"`
	AssigneeName *string      `db:"assigneeName" json:"assigneeName"`
	DueDate      *string      `db:"dueDate" json:"dueDate"`
	ReviewNote   *string      `db:"reviewNote" json:"reviewNote"`
	ReviewedBy   *string      `db:"reviewedByName" json:"reviewedByName"`
	ReviewedAt   *time.Time   `db:"reviewedAt" json:"reviewedAt"`
	TaskID       *string      `db:"taskId" json:"taskId"`
	Files        []DesignFile `db:"-" json:"files"`
}

type DesignFile struct {
	ID         string    `db:"id" json:"id"`
	SectionID  string    `db:"sectionId" json:"sectionId"`
	Title      string    `db:"title" json:"title"`
	FileKey    *string   `db:"fileKey" json:"fileKey"`
	FileType   *string   `db:"fileType" json:"fileType"`
	URL        *string   `db:"url" json:"url"`
	ToolName   *string   `db:"toolName" json:"toolName"`
	UploadedBy *string   `db:"uploadedByName" json:"uploadedByName"`
	CreatedAt  time.Time `db:"createdAt" json:"createdAt"`
}

type DesignSetDetail struct {
	DesignSet
	Sections  []DesignSection `json:"sections"`
	CanManage bool            `json:"canManage"`
	CanReview bool            `json:"canReview"`
}

type DesignSetService struct {
	db    *sqlx.DB
	tasks *repository.ExtraTaskRepository
	notif *repository.NotificationRepository
}

func NewDesignSetService(db *sqlx.DB, tasks *repository.ExtraTaskRepository, notif *repository.NotificationRepository) *DesignSetService {
	return &DesignSetService{db: db, tasks: tasks, notif: notif}
}

var ErrDesignHidden = errors.New("مو موجود")

// ── الصلاحيات ──
// يدير (ينشئ صيت، يوزّع، يعدّل الأدوات): المالك/مدير النظام، مدير/نائب الهندسة،
// مدير/نائب وحدة التصميم، ومن عنده إدارة المشاريع.
// يراجع ويعتمد: المالك/مدير النظام، ومدير/نائب الهندسة أو وحدة التصميم.
// يشوف: الي يدير + أعضاء وحدة التصميم + المصمم المكلّف بتخصص (يشوف صيته بس).
func (s *DesignSetService) inDesignUnit(emp string, positions ...string) bool {
	var n int
	errlog.Err(s.db.Get(&n, `SELECT count(*) FROM "OrgAssignment" WHERE "employeeId" = $1 AND "endedAt" IS NULL
		AND (("divisionId" = 'eng' AND "unitId" IS NULL AND position = ANY($2)) OR ("unitId" = 'eng.design' AND position = ANY($3)))`,
		emp, pq.Array([]string{"DIRECTOR", "DEPUTY"}), pq.Array(positions)))
	return n > 0
}

func (s *DesignSetService) CanReview(emp, role string) bool {
	return model.IsAdminRole(role) || s.inDesignUnit(emp, "DIRECTOR", "DEPUTY")
}

func (s *DesignSetService) CanManage(emp, role string, perms []string) bool {
	if s.CanReview(emp, role) {
		return true
	}
	for _, p := range perms {
		if p == "project_management" {
			return true
		}
	}
	return false
}

func (s *DesignSetService) CanSeeAll(emp, role string, perms []string) bool {
	return s.CanManage(emp, role, perms) || s.inDesignUnit(emp, "DIRECTOR", "DEPUTY", "MEMBER")
}

func (s *DesignSetService) canSeeSet(emp, role string, perms []string, setID string) bool {
	if s.CanSeeAll(emp, role, perms) {
		return true
	}
	var n int
	errlog.Err(s.db.Get(&n, `SELECT count(*) FROM "DesignSection" WHERE "setId" = $1 AND "assigneeId" = $2`, setID, emp))
	return n > 0
}

// ── الأدوات ──

func (s *DesignSetService) Tools(all bool) ([]DesignTool, error) {
	rows := []DesignTool{}
	q := `SELECT id, discipline, name, url, "howTo", sort, active FROM "DesignTool"`
	if !all {
		q += ` WHERE active`
	}
	err := s.db.Select(&rows, q+` ORDER BY discipline, sort, name`)
	return rows, err
}

type DesignToolInput struct {
	Discipline string `json:"discipline"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	HowTo      string `json:"howTo"`
	Active     *bool  `json:"active"`
}

func (s *DesignSetService) SaveTool(emp, role string, perms []string, id string, in DesignToolInput) (string, error) {
	if !s.CanManage(emp, role, perms) {
		return "", ErrDesignHidden
	}
	in.Name, in.URL = strings.TrimSpace(in.Name), strings.TrimSpace(in.URL)
	if !validDiscipline(in.Discipline) || in.Name == "" || !strings.HasPrefix(in.URL, "https://") {
		return "", errors.New("اختار التخصص واكتب اسم الأداة ورابطها (يبدي بـ https://)")
	}
	active := true
	if in.Active != nil {
		active = *in.Active
	}
	if id == "" {
		err := s.db.Get(&id, `INSERT INTO "DesignTool" (discipline, name, url, "howTo", active, sort)
			VALUES ($1, $2, $3, $4, $5, COALESCE((SELECT max(sort) + 1 FROM "DesignTool" WHERE discipline = $1), 0)) RETURNING id`,
			in.Discipline, in.Name, in.URL, strings.TrimSpace(in.HowTo), active)
		return id, err
	}
	_, err := s.db.Exec(`UPDATE "DesignTool" SET discipline = $2, name = $3, url = $4, "howTo" = $5, active = $6 WHERE id = $1`,
		id, in.Discipline, in.Name, in.URL, strings.TrimSpace(in.HowTo), active)
	return id, err
}

// ── الصيتات ──

const designSetCols = `SELECT d.id, d.title, d."projectId", p.code AS "projectCode", d."customerId", c.name AS "customerName",
	d.notes, e.name AS "createdByName", d."createdAt", d."archivedAt",
	(SELECT count(*) FROM "DesignSection" x WHERE x."setId" = d.id)::int AS sections,
	(SELECT count(*) FROM "DesignSection" x WHERE x."setId" = d.id AND x.status = 'APPROVED')::int AS approved,
	(SELECT count(*) FROM "DesignSection" x WHERE x."setId" = d.id AND x.status = 'REVIEW')::int AS "inReview",
	(SELECT count(*) FROM "DesignSection" x WHERE x."setId" = d.id AND x.status <> 'APPROVED' AND x."dueDate" < baghdad_today())::int AS late
	FROM "DesignSet" d
	LEFT JOIN "Project" p ON p.id = d."projectId"
	LEFT JOIN "Customer" c ON c.id = d."customerId"
	LEFT JOIN "Employee" e ON e.id = d."createdById"`

func (s *DesignSetService) List(emp, role string, perms []string) ([]DesignSet, error) {
	rows := []DesignSet{}
	if s.CanSeeAll(emp, role, perms) {
		err := s.db.Select(&rows, designSetCols+` ORDER BY (d."archivedAt" IS NOT NULL), d."createdAt" DESC LIMIT 300`)
		return rows, err
	}
	err := s.db.Select(&rows, designSetCols+` WHERE EXISTS (SELECT 1 FROM "DesignSection" x WHERE x."setId" = d.id AND x."assigneeId" = $1)
		ORDER BY (d."archivedAt" IS NOT NULL), d."createdAt" DESC LIMIT 300`, emp)
	return rows, err
}

type DesignSetInput struct {
	Title       string   `json:"title"`
	ProjectID   string   `json:"projectId"`
	CustomerID  string   `json:"customerId"`
	Notes       string   `json:"notes"`
	Disciplines []string `json:"disciplines"`
}

func (s *DesignSetService) Create(emp, role string, perms []string, in DesignSetInput) (string, error) {
	if !s.CanManage(emp, role, perms) {
		return "", ErrDesignHidden
	}
	in.Title = strings.TrimSpace(in.Title)
	if len([]rune(in.Title)) < 3 {
		return "", errors.New("اكتب اسم الصيت (مثلاً: فيلا أبو علي — الجادرية)")
	}
	ds := []string{}
	for _, d := range in.Disciplines {
		if validDiscipline(d) {
			ds = append(ds, d)
		}
	}
	if len(ds) == 0 {
		return "", errors.New("اختار تخصص واحد على الأقل")
	}
	tx, err := s.db.Beginx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback() //nolint:errcheck
	var id string
	if err := tx.Get(&id, `INSERT INTO "DesignSet" (title, "projectId", "customerId", notes, "createdById")
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF(btrim($4), ''), $5) RETURNING id`,
		in.Title, in.ProjectID, in.CustomerID, in.Notes, emp); err != nil {
		if strings.Contains(err.Error(), "foreign key") {
			return "", errors.New("المشروع أو الزبون مو موجود")
		}
		return "", err
	}
	for _, d := range ds {
		if _, err := tx.Exec(`INSERT INTO "DesignSection" ("setId", discipline) VALUES ($1, $2)`, id, d); err != nil {
			return "", err
		}
	}
	return id, tx.Commit()
}

// AddSection يضيف تخصص لصيت موجود.
func (s *DesignSetService) AddSection(emp, role string, perms []string, setID, discipline string) error {
	if !s.CanManage(emp, role, perms) {
		return ErrDesignHidden
	}
	if !validDiscipline(discipline) {
		return errors.New("تخصص غير معروف")
	}
	_, err := s.db.Exec(`INSERT INTO "DesignSection" ("setId", discipline) VALUES ($1, $2) ON CONFLICT ("setId", discipline) DO NOTHING`, setID, discipline)
	return err
}

func (s *DesignSetService) Archive(emp, role string, perms []string, setID string, archived bool) error {
	if !s.CanManage(emp, role, perms) {
		return ErrDesignHidden
	}
	_, err := s.db.Exec(`UPDATE "DesignSet" SET "archivedAt" = CASE WHEN $2 THEN now() ELSE NULL END WHERE id = $1`, setID, archived)
	return err
}

func (s *DesignSetService) Detail(emp, role string, perms []string, setID string) (*DesignSetDetail, error) {
	if !s.canSeeSet(emp, role, perms, setID) {
		return nil, ErrDesignHidden
	}
	d := &DesignSetDetail{}
	if err := s.db.Get(&d.DesignSet, designSetCols+` WHERE d.id = $1`, setID); err != nil {
		return nil, ErrDesignHidden
	}
	d.CanManage, d.CanReview = s.CanManage(emp, role, perms), s.CanReview(emp, role)
	d.Sections = []DesignSection{}
	if err := s.db.Select(&d.Sections, `SELECT x.id, x."setId", x.discipline, x.status, x."assigneeId", a.name AS "assigneeName",
		x."dueDate"::text AS "dueDate", x."reviewNote", r.name AS "reviewedByName", x."reviewedAt", x."taskId"
		FROM "DesignSection" x LEFT JOIN "Employee" a ON a.id = x."assigneeId" LEFT JOIN "Employee" r ON r.id = x."reviewedById"
		WHERE x."setId" = $1`, setID); err != nil {
		return nil, err
	}
	files := []DesignFile{}
	if err := s.db.Select(&files, `SELECT f.id, f."sectionId", f.title, f."fileKey", f."fileType", f.url, f."toolName", e.name AS "uploadedByName", f."createdAt"
		FROM "DesignFile" f JOIN "DesignSection" x ON x.id = f."sectionId" LEFT JOIN "Employee" e ON e.id = f."uploadedById"
		WHERE x."setId" = $1 AND f."removedAt" IS NULL ORDER BY f."createdAt" DESC`, setID); err != nil {
		return nil, err
	}
	by := map[string][]DesignFile{}
	for _, f := range files {
		by[f.SectionID] = append(by[f.SectionID], f)
	}
	order := map[string]int{}
	for i, dd := range DesignDisciplines {
		order[dd.Code] = i
	}
	for i := range d.Sections {
		d.Sections[i].Files = by[d.Sections[i].ID]
		if d.Sections[i].Files == nil {
			d.Sections[i].Files = []DesignFile{}
		}
	}
	sortSections(d.Sections, order)
	return d, nil
}

func sortSections(xs []DesignSection, order map[string]int) {
	for i := 1; i < len(xs); i++ {
		for j := i; j > 0 && order[xs[j].Discipline] < order[xs[j-1].Discipline]; j-- {
			xs[j], xs[j-1] = xs[j-1], xs[j]
		}
	}
}

type designSectionScope struct {
	SetID      string  `db:"setId"`
	Discipline string  `db:"discipline"`
	Status     string  `db:"status"`
	Assignee   *string `db:"assigneeId"`
	TaskID     *string `db:"taskId"`
	Title      string  `db:"title"`
}

func (s *DesignSetService) section(id string) (*designSectionScope, error) {
	var sc designSectionScope
	if err := s.db.Get(&sc, `SELECT x."setId", x.discipline, x.status, x."assigneeId", x."taskId", d.title
		FROM "DesignSection" x JOIN "DesignSet" d ON d.id = x."setId" WHERE x.id = $1`, id); err != nil {
		return nil, ErrDesignHidden
	}
	return &sc, nil
}

func disciplineName(code string) string {
	for _, d := range DesignDisciplines {
		if d.Code == code {
			return d.Name
		}
	}
	return code
}

// Assign يحدد المصمم والموعد — ويصير مهمة بمهامه (تنحسب بتقييمه).
func (s *DesignSetService) Assign(emp, role string, perms []string, sectionID, assigneeID, due string) error {
	if !s.CanManage(emp, role, perms) {
		return ErrDesignHidden
	}
	sc, err := s.section(sectionID)
	if err != nil {
		return err
	}
	dueT, err := time.ParseInLocation("2006-01-02", due, baghdad)
	if err != nil {
		return errors.New("حدد موعد التسليم")
	}
	var ok bool
	errlog.Err(s.db.Get(&ok, `SELECT EXISTS (SELECT 1 FROM "Employee" WHERE id = $1 AND status = 'ACTIVE')`, assigneeID))
	if !ok {
		return errors.New("اختار المصمم")
	}
	var taskID *string
	// تبديل المصمم أو الموعد: المهمة القديمة تنلغى ويا سبب، والجديدة تنفتح
	if sc.TaskID != nil {
		_, _ = s.db.Exec(`UPDATE "ExtraTask" SET status = 'CANCELLED', "updatedAt" = now() WHERE id = $1 AND status IN ('NEW', 'IN_PROGRESS')`, *sc.TaskID)
	}
	if s.tasks != nil {
		desc := "تصميم «" + disciplineName(sc.Discipline) + "» لصيت «" + sc.Title + "». صمّم بالأداة المناسبة (أزرارها بالصيت)، وارفع الشغل على نفس التخصص، واضغط «للمراجعة». تفتحه من «🏗️ صيت التصاميم» أو الرابط: /design-sets/" + sc.SetID
		if id, err := s.tasks.CreateFor("🎨 تصميم "+disciplineName(sc.Discipline)+" — "+sc.Title, desc, assigneeID, emp, dueT.Add(23*time.Hour)); err == nil {
			taskID = &id
		}
	}
	_, err = s.db.Exec(`UPDATE "DesignSection" SET "assigneeId" = $2, "dueDate" = $3::date, "taskId" = $4,
		status = CASE WHEN status = 'TODO' THEN 'IN_PROGRESS' ELSE status END WHERE id = $1`, sectionID, assigneeID, due, taskID)
	return err
}

type DesignFileInput struct {
	Title    string `json:"title"`
	FileKey  string `json:"fileKey"`
	FileType string `json:"fileType"`
	URL      string `json:"url"`
	ToolName string `json:"toolName"`
}

func (s *DesignSetService) canWork(emp, role string, perms []string, sc *designSectionScope) bool {
	return s.CanManage(emp, role, perms) || (sc.Assignee != nil && *sc.Assignee == emp) || s.inDesignUnit(emp, "DIRECTOR", "DEPUTY", "MEMBER")
}

// AddFile ملف مرفوع (مفتاح من /api/files?folder=designs) أو رابط مشاركة من أداة خارجية.
func (s *DesignSetService) AddFile(emp, role string, perms []string, sectionID string, in DesignFileInput) (string, error) {
	sc, err := s.section(sectionID)
	if err != nil {
		return "", err
	}
	if !s.canWork(emp, role, perms, sc) {
		return "", ErrDesignHidden
	}
	if sc.Status == "APPROVED" {
		return "", errors.New("التخصص معتمد — رجّعه للتعديل أول")
	}
	in.Title, in.URL, in.FileKey = strings.TrimSpace(in.Title), strings.TrimSpace(in.URL), strings.TrimSpace(in.FileKey)
	if (in.FileKey == "") == (in.URL == "") {
		return "", errors.New("ارفع ملف أو الصق رابط")
	}
	if in.URL != "" && !strings.HasPrefix(in.URL, "https://") {
		return "", errors.New("الرابط لازم يبدي بـ https://")
	}
	if in.FileKey != "" && !strings.HasPrefix(in.FileKey, "/api/files/designs/") {
		return "", errors.New("الملف لازم ينرفع لمجلد التصاميم")
	}
	if in.Title == "" {
		in.Title = disciplineName(sc.Discipline)
	}
	if len([]rune(in.Title)) > 200 || len(in.URL) > 2000 {
		return "", errors.New("الاسم أو الرابط طويل هواية")
	}
	var id string
	err = s.db.Get(&id, `INSERT INTO "DesignFile" ("sectionId", title, "fileKey", "fileType", url, "toolName", "uploadedById")
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), $7) RETURNING id`,
		sectionID, in.Title, in.FileKey, in.FileType, in.URL, strings.TrimSpace(in.ToolName), emp)
	if err == nil {
		_, _ = s.db.Exec(`UPDATE "DesignSection" SET status = 'IN_PROGRESS' WHERE id = $1 AND status IN ('TODO', 'RETURNED')`, sectionID)
	}
	return id, err
}

// RemoveFile يخفي الملف (ما ينحذف من التخزين — يبقى بالسجل).
func (s *DesignSetService) RemoveFile(emp, role string, perms []string, fileID string) error {
	var sectionID string
	if err := s.db.Get(&sectionID, `SELECT "sectionId" FROM "DesignFile" WHERE id = $1`, fileID); err != nil {
		return ErrDesignHidden
	}
	sc, err := s.section(sectionID)
	if err != nil {
		return err
	}
	if !s.canWork(emp, role, perms, sc) {
		return ErrDesignHidden
	}
	if sc.Status == "APPROVED" {
		return errors.New("التخصص معتمد — رجّعه للتعديل أول")
	}
	_, err = s.db.Exec(`UPDATE "DesignFile" SET "removedAt" = now(), "removedById" = $2 WHERE id = $1`, fileID, emp)
	return err
}

// Submit المصمم يرفع التخصص للمراجعة.
func (s *DesignSetService) Submit(emp, role string, perms []string, sectionID string) error {
	sc, err := s.section(sectionID)
	if err != nil {
		return err
	}
	if !s.canWork(emp, role, perms, sc) {
		return ErrDesignHidden
	}
	var files int
	errlog.Err(s.db.Get(&files, `SELECT count(*) FROM "DesignFile" WHERE "sectionId" = $1 AND "removedAt" IS NULL`, sectionID))
	if files == 0 {
		return errors.New("ارفع التصميم أول (ملف أو رابط)")
	}
	if sc.Status == "APPROVED" || sc.Status == "REVIEW" {
		return errors.New("التخصص مرفوع أو معتمد أصلاً")
	}
	if _, err := s.db.Exec(`UPDATE "DesignSection" SET status = 'REVIEW', "submittedAt" = now() WHERE id = $1`, sectionID); err != nil {
		return err
	}
	if s.notif != nil {
		msg := "🎨 تصميم «" + disciplineName(sc.Discipline) + "» لصيت «" + sc.Title + "» ينتظر مراجعتك."
		_ = s.notif.CreateForRole("ADMIN", "DESIGN_REVIEW", msg)
		_ = s.notif.CreateForRole("OWNER", "DESIGN_REVIEW", msg)
		reviewers := []string{}
		errlog.Err(s.db.Select(&reviewers, `SELECT DISTINCT "employeeId" FROM "OrgAssignment" WHERE "endedAt" IS NULL AND position IN ('DIRECTOR', 'DEPUTY')
			AND (("divisionId" = 'eng' AND "unitId" IS NULL) OR "unitId" = 'eng.design') AND "employeeId" <> $1`, emp))
		for _, r := range reviewers {
			_ = s.notif.Create(r, "DESIGN_REVIEW", msg)
		}
	}
	return nil
}

// Review يعتمد أو يرجّع (الإرجاع بملاحظة). الاعتماد يخلّص مهمة المصمم؛ ويكدر يرجّعه بعد الاعتماد.
func (s *DesignSetService) Review(emp, role, sectionID string, approve bool, note string) error {
	if !s.CanReview(emp, role) {
		return ErrDesignHidden
	}
	sc, err := s.section(sectionID)
	if err != nil {
		return err
	}
	note = strings.TrimSpace(note)
	status := "APPROVED"
	if !approve {
		status = "RETURNED"
		if len([]rune(note)) < 5 {
			return errors.New("اكتب شنو لازم يتعدل حتى ترجّعه")
		}
	} else if sc.Status != "REVIEW" {
		return errors.New("التخصص مو مرفوع للمراجعة")
	}
	if _, err := s.db.Exec(`UPDATE "DesignSection" SET status = $2, "reviewNote" = NULLIF($3, ''), "reviewedById" = $4, "reviewedAt" = now() WHERE id = $1`,
		sectionID, status, note, emp); err != nil {
		return err
	}
	if approve && sc.TaskID != nil {
		_, _ = s.db.Exec(`UPDATE "ExtraTask" SET status = 'DONE', "doneAt" = now(), "doneNote" = 'انعتمد التصميم بصيت التصاميم', "updatedAt" = now()
			WHERE id = $1 AND status IN ('NEW', 'IN_PROGRESS')`, *sc.TaskID)
	}
	if s.notif != nil && sc.Assignee != nil {
		msg := "✅ انعتمد تصميم «" + disciplineName(sc.Discipline) + "» لصيت «" + sc.Title + "»."
		if !approve {
			msg = "↩️ تصميم «" + disciplineName(sc.Discipline) + "» لصيت «" + sc.Title + "» رجع للتعديل: " + note
		}
		_ = s.notif.Create(*sc.Assignee, "DESIGN_REVIEW", msg)
	}
	return nil
}

// ProjectOption مشروع يتربط بيه الصيت (للي يدير بس).
type ProjectOption struct {
	ID   string `db:"id" json:"id"`
	Code string `db:"code" json:"code"`
	Name string `db:"name" json:"name"`
}

func (s *DesignSetService) Projects() []ProjectOption {
	rows := []ProjectOption{}
	errlog.Err(s.db.Select(&rows, `SELECT id, code, name FROM "Project" ORDER BY "createdAt" DESC LIMIT 400`))
	return rows
}
