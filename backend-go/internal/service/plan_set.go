package service

import (
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ 📐 سيت المخططات (ك) 11-10 ═══
// بطلب دازاي: «هذن ينضافن بوحدة التصاميم… تختار الي تريده وتسوي ابلود… ومن تسوي
// مجموعة من هاي التصاميم تسوي تنزيل ملف PDF بالهوية البصرية مال الشركة بي كل
// التصاميم بالإضافة الى عرض السعر». شاشة جديدة لحالها، ما تلمس «صيت التصاميم».
// الأقسام والأنواع نفس ورقة «Package Contents» مال الشركة (١٠ أقسام).
// الصلاحيات نفس صيت التصاميم: يدير = DesignSetService.CanManage، يشوف ويرفع = CanSeeAll.

type PlanType struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type PlanGroup struct {
	No    int        `json:"no"`
	Name  string     `json:"name"`
	En    string     `json:"en"`
	Types []PlanType `json:"types"`
}

// PlanCatalog الأقسام العشرة بأنواعها. الأرقام مصحّحة: بالورقة القسم ٥ و٦ و١٠ چانت
// أنواعها مرقّمة بأرقام قسم ثاني (4.x / 8.x)، فخليناها تبع قسمها.
var PlanCatalog = []PlanGroup{
	{1, "المخططات المعمارية", "Architectural Drawings", []PlanType{
		{"1.1", "مخطط الموقع العام"}, {"1.2", "مخطط الأبعاد"}, {"1.3", "مخطط الأثاث"},
		{"1.4", "الواجهات الخارجية"}, {"1.5", "الواجهات الداخلية"}, {"1.6", "المقاطع"},
		{"1.7", "مخطط الأسقف"}, {"1.8", "مخطط سطح المبنى"}, {"1.9", "مخطط تفصيلي خارجي"},
		{"1.10", "مخطط تفصيلي داخلي"},
	}},
	{2, "المخططات الإنشائية", "Structural Drawings", []PlanType{
		{"2.1", "مخطط الأساس والحفريات"}, {"2.2", "مخطط الأعمدة"}, {"2.3", "مخطط الأسقف"},
		{"2.4", "مخطط الأسطح والدرج"}, {"2.5", "جدول تسليح الإنشائي"}, {"2.6", "مخططات إنشائية تفصيلية"},
	}},
	{3, "المخططات الكهربائية", "Electrical Drawings", []PlanType{
		{"3.1", "مخطط الإنارة"}, {"3.2", "مخطط القوى والمآخذ الكهربائية"},
		{"3.3", "مخطط تغذية التكييف والأجهزة الثقيلة"}, {"3.4", "مخطط التغذية والتوصيلات الكهربائية"},
		{"3.5", "جدول الأحمال ولوحات التوزيع"}, {"3.6", "مخطط الخط الواحد"},
		{"3.7", "مخطط التأريض ومانعة الصواعق"}, {"3.8", "التفاصيل التنفيذية للكهرباء"},
	}},
	{4, "المخططات الميكانيكية والتكييف والصحي", "Mechanical, HVAC & Plumbing Drawings", []PlanType{
		{"4.1", "مخطط التكييف والتهوية"}, {"4.2", "مخطط التغذية بالمياه"},
		{"4.3", "مخطط تصريف مياه الأمطار والصرف الصحي"}, {"4.4", "مخطط مكافح الحريق"},
	}},
	{5, "مخطط الكاميرات والشبكة والصوت", "CCTV, Network & Sound System", []PlanType{
		{"5.1", "مخططات الكاميرات وتغطيتها"}, {"5.2", "مخطط الخريطة الحرارية لشبكة Wi-Fi"},
		{"5.3", "مخطط ربط الشبكات السلكية"}, {"5.4", "مخطط توزيع الهاتف الأرضي"},
		{"5.5", "مخطط مسارات وحوامل الكابلات الرئيسية والفرعية"}, {"5.6", "مخطط توزيع أجهزة القدرة الاحتياطية"},
		{"5.7", "مخطط أنظمة الصوت والإذاعة الداخلية"}, {"5.8", "مخطط الشاشات والعرض"},
		{"5.9", "مخطط التفاصيل التنفيذية"},
	}},
	{6, "جدول الأبواب والشبابيك والتشطيبات", "Door, Window & Finishes Schedules", []PlanType{
		{"6.1", "جدول الأبواب"}, {"6.2", "جدول الشبابيك والفتحات"}, {"6.3", "جدول التشطيبات الداخلية"},
	}},
	{7, "جدول الكميات", "BOQ (Bill of Quantities)", []PlanType{{"7.1", "جدول الكميات"}}},
	{8, "تفاصيل البناء", "Construction Details", []PlanType{{"8.1", "تفاصيل البناء"}}},
	{9, "المواصفات الفنية", "Technical Specifications", []PlanType{{"9.1", "المواصفات الفنية"}}},
	{10, "حزمة العرض", "3D Renderings & Presentation Boards", []PlanType{
		{"10.1", "المفهوم الفلسفي"}, {"10.2", "اللقطات الرئيسية"}, {"10.3", "اللقطات الثانوية"},
		{"10.4", "اللقطات المنظورية"},
	}},
}

func validPlanType(code string) bool {
	for _, g := range PlanCatalog {
		for _, t := range g.Types {
			if t.Code == code {
				return true
			}
		}
	}
	return false
}

func planTypeCount() int {
	n := 0
	for _, g := range PlanCatalog {
		n += len(g.Types)
	}
	return n
}

type PlanSet struct {
	ID            string     `db:"id" json:"id"`
	Title         string     `db:"title" json:"title"`
	ProjectID     *string    `db:"projectId" json:"projectId"`
	ProjectCode   *string    `db:"projectCode" json:"projectCode"`
	ProjectName   *string    `db:"projectName" json:"projectName"`
	Notes         *string    `db:"notes" json:"notes"`
	CreatedByName *string    `db:"createdByName" json:"createdByName"`
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`
	ArchivedAt    *time.Time `db:"archivedAt" json:"archivedAt"`
	Files         int        `db:"files" json:"files"`
	Cover         *string    `db:"cover" json:"cover"`
}

type PlanFile struct {
	ID             string    `db:"id" json:"id"`
	SetID          string    `db:"setId" json:"setId"`
	TypeCode       string    `db:"typeCode" json:"typeCode"`
	Title          string    `db:"title" json:"title"`
	FileKey        string    `db:"fileKey" json:"fileKey"`
	FileType       *string   `db:"fileType" json:"fileType"`
	UploadedByName *string   `db:"uploadedByName" json:"uploadedByName"`
	CreatedAt      time.Time `db:"createdAt" json:"createdAt"`
}

type PlanSetDetail struct {
	PlanSet
	Items     []PlanFile `json:"items"`
	CanManage bool       `json:"canManage"`
}

type PlanStats struct {
	Types    int `json:"types"`
	Projects int `db:"projects" json:"projects"`
	Sets     int `db:"sets" json:"sets"`
	Files    int `db:"files" json:"files"`
}

type PlanSetService struct {
	db     *sqlx.DB
	design *DesignSetService
}

func NewPlanSetService(db *sqlx.DB, design *DesignSetService) *PlanSetService {
	return &PlanSetService{db: db, design: design}
}

func (s *PlanSetService) CanManage(emp, role string, perms []string) bool {
	return s.design.CanManage(emp, role, perms)
}

func (s *PlanSetService) CanSee(emp, role string, perms []string) bool {
	return s.design.CanSeeAll(emp, role, perms)
}

func (s *PlanSetService) Projects() []ProjectOption { return s.design.Projects() }

// imageTypes الأنواع الي تنعرض صورة (الغلاف وحزمة الـPDF).
const planImageTypes = `('image/jpeg', 'image/png', 'image/webp')`

const planSetCols = `SELECT d.id, d.title, d."projectId", p.code AS "projectCode", p.name AS "projectName",
	d.notes, e.name AS "createdByName", d."createdAt", d."archivedAt",
	(SELECT count(*) FROM "PlanFile" f WHERE f."setId" = d.id AND f."removedAt" IS NULL)::int AS files,
	(SELECT f."fileKey" FROM "PlanFile" f WHERE f."setId" = d.id AND f."removedAt" IS NULL
	   AND f."fileType" IN ` + planImageTypes + ` ORDER BY f."createdAt" LIMIT 1) AS cover
	FROM "PlanSet" d LEFT JOIN "Project" p ON p.id = d."projectId" LEFT JOIN "Employee" e ON e.id = d."createdById"`

func (s *PlanSetService) Stats() (PlanStats, error) {
	st := PlanStats{Types: planTypeCount()}
	err := s.db.Get(&st, `SELECT
		(SELECT count(DISTINCT "projectId") FROM "PlanSet" WHERE "archivedAt" IS NULL AND "projectId" IS NOT NULL)::int AS projects,
		(SELECT count(*) FROM "PlanSet" WHERE "archivedAt" IS NULL)::int AS sets,
		(SELECT count(*) FROM "PlanFile" f JOIN "PlanSet" d ON d.id = f."setId"
		  WHERE f."removedAt" IS NULL AND d."archivedAt" IS NULL)::int AS files`)
	st.Types = planTypeCount()
	return st, err
}

func (s *PlanSetService) List(emp, role string, perms []string) ([]PlanSet, error) {
	rows := []PlanSet{}
	if !s.CanSee(emp, role, perms) {
		return rows, ErrDesignHidden
	}
	err := s.db.Select(&rows, planSetCols+` ORDER BY (d."archivedAt" IS NOT NULL), d."createdAt" DESC LIMIT 300`)
	return rows, err
}

type PlanSetInput struct {
	Title     string `json:"title"`
	ProjectID string `json:"projectId"`
	Notes     string `json:"notes"`
}

func (s *PlanSetService) Create(emp, role string, perms []string, in PlanSetInput) (string, error) {
	if !s.CanManage(emp, role, perms) {
		return "", ErrDesignHidden
	}
	in.Title = strings.TrimSpace(in.Title)
	if n := len([]rune(in.Title)); n < 3 || n > 200 {
		return "", errors.New("اكتب اسم السيت (مثلاً: سيت مشروع الجامعة)")
	}
	var id string
	err := s.db.Get(&id, `INSERT INTO "PlanSet" (title, "projectId", notes, "createdById")
		VALUES ($1, NULLIF($2, ''), NULLIF(btrim($3), ''), $4) RETURNING id`,
		in.Title, strings.TrimSpace(in.ProjectID), in.Notes, emp)
	if err != nil && strings.Contains(err.Error(), "foreign key") {
		return "", errors.New("المشروع مو موجود")
	}
	return id, err
}

func (s *PlanSetService) Detail(emp, role string, perms []string, id string) (*PlanSetDetail, error) {
	if !s.CanSee(emp, role, perms) {
		return nil, ErrDesignHidden
	}
	d := &PlanSetDetail{Items: []PlanFile{}, CanManage: s.CanManage(emp, role, perms)}
	if err := s.db.Get(&d.PlanSet, planSetCols+` WHERE d.id = $1`, id); err != nil {
		return nil, ErrDesignHidden
	}
	err := s.db.Select(&d.Items, `SELECT f.id, f."setId", f."typeCode", f.title, f."fileKey", f."fileType",
		e.name AS "uploadedByName", f."createdAt"
		FROM "PlanFile" f LEFT JOIN "Employee" e ON e.id = f."uploadedById"
		WHERE f."setId" = $1 AND f."removedAt" IS NULL ORDER BY f."createdAt"`, id)
	return d, err
}

func (s *PlanSetService) Archive(emp, role string, perms []string, id string, archived bool) error {
	if !s.CanManage(emp, role, perms) {
		return ErrDesignHidden
	}
	q := `UPDATE "PlanSet" SET "archivedAt" = NULL WHERE id = $1`
	if archived {
		q = `UPDATE "PlanSet" SET "archivedAt" = now() WHERE id = $1`
	}
	_, err := s.db.Exec(q, id)
	return err
}

type PlanFileInput struct {
	TypeCode string `json:"typeCode"`
	Title    string `json:"title"`
	FileKey  string `json:"fileKey"`
	FileType string `json:"fileType"`
}

func (s *PlanSetService) AddFile(emp, role string, perms []string, setID string, in PlanFileInput) (string, error) {
	if !s.CanSee(emp, role, perms) {
		return "", ErrDesignHidden
	}
	var archived *time.Time
	if err := s.db.Get(&archived, `SELECT "archivedAt" FROM "PlanSet" WHERE id = $1`, setID); err != nil {
		return "", ErrDesignHidden
	}
	if archived != nil {
		return "", errors.New("السيت مؤرشف — رجّعه أول")
	}
	if !validPlanType(in.TypeCode) {
		return "", errors.New("اختار نوع المخطط")
	}
	in.FileKey, in.Title = strings.TrimSpace(in.FileKey), strings.TrimSpace(in.Title)
	// نفس مجلد صيت التصاميم: يقبل PDF وDWG والصور لحد ٢٤ ميگا.
	if !strings.HasPrefix(in.FileKey, "/api/files/designs/") {
		return "", errors.New("الملف لازم ينرفع لمجلد التصاميم")
	}
	if in.Title == "" {
		in.Title = in.TypeCode
	}
	if len([]rune(in.Title)) > 200 {
		return "", errors.New("الاسم طويل هواية")
	}
	var id string
	err := s.db.Get(&id, `INSERT INTO "PlanFile" ("setId", "typeCode", title, "fileKey", "fileType", "uploadedById")
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6) RETURNING id`,
		setID, in.TypeCode, in.Title, in.FileKey, strings.TrimSpace(in.FileType), emp)
	return id, err
}

// RemoveFile يشيل الملف من السيت (يبقى بالقاعدة بـremovedAt — ماكو حذف فعلي).
// يشيله الي رفعه أو الي يدير.
func (s *PlanSetService) RemoveFile(emp, role string, perms []string, fileID string) error {
	var up *string
	if err := s.db.Get(&up, `SELECT "uploadedById" FROM "PlanFile" WHERE id = $1 AND "removedAt" IS NULL`, fileID); err != nil {
		return ErrDesignHidden
	}
	if !s.CanManage(emp, role, perms) && (up == nil || *up != emp || !s.CanSee(emp, role, perms)) {
		return ErrDesignHidden
	}
	_, err := s.db.Exec(`UPDATE "PlanFile" SET "removedAt" = now(), "removedById" = $2 WHERE id = $1`, fileID, emp)
	return err
}

// ═══ واجهة «سيت المخططات» الجديدة — (ك) 11-10 بطلب دازاي «اريد هذا بالضبط» ═══
// كروت المشاريع، وآخر النشاطات، ونسبة النمو. كل الأرقام محسوبة من الجداول — ماكو رقم وهمي.

// PlanProjectCard مشروع بيه سيتات: الغلاف أول صورة بآخر سيت، والحالة «مكتمل» إذا كل سيتاته مؤرشفة.
type PlanProjectCard struct {
	ProjectID string    `db:"projectId" json:"projectId"`
	Code      string    `db:"code" json:"code"`
	Name      string    `db:"name" json:"name"`
	Location  *string   `db:"location" json:"location"`
	Sets      int       `db:"sets" json:"sets"`
	OpenSets  int       `db:"openSets" json:"openSets"`
	LastAt    time.Time `db:"lastAt" json:"lastAt"`
	Cover     *string   `db:"cover" json:"cover"`
	LatestSet string    `db:"latestSet" json:"latestSet"`
}

// PlanActivity آخر شي صار: سيت انفتح أو مخطط انرفع.
type PlanActivity struct {
	Kind  string    `db:"kind" json:"kind"` // set | file
	Title string    `db:"title" json:"title"`
	Sub   *string   `db:"sub" json:"sub"`
	SetID string    `db:"setId" json:"setId"`
	At    time.Time `db:"at" json:"at"`
}

// PlanGrowth نسبة الي انضاف آخر ٣٠ يوم مقابل الي چان قبلها — nil إذا ماكو أساس للمقارنة.
type PlanGrowth struct {
	Projects *float64 `json:"projects"`
	Sets     *float64 `json:"sets"`
	Files    *float64 `json:"files"`
	Pdfs     *float64 `json:"pdfs"`
}

type PlanOverview struct {
	Pdfs     int               `json:"pdfs"`
	Growth   PlanGrowth        `json:"growth"`
	Projects []PlanProjectCard `json:"projectCards"`
	Activity []PlanActivity    `json:"activity"`
}

func growth(newer, older int) *float64 {
	if older <= 0 {
		return nil
	}
	v := float64(newer) * 100 / float64(older)
	return &v
}

func (s *PlanSetService) Overview() (PlanOverview, error) {
	out := PlanOverview{Projects: []PlanProjectCard{}, Activity: []PlanActivity{}}
	var g struct {
		PdfAll  int `db:"pdfAll"`
		PdfNew  int `db:"pdfNew"`
		SetAll  int `db:"setAll"`
		SetNew  int `db:"setNew"`
		FileAll int `db:"fileAll"`
		FileNew int `db:"fileNew"`
		PrjAll  int `db:"prjAll"`
		PrjNew  int `db:"prjNew"`
	}
	const live = `f."removedAt" IS NULL AND d."archivedAt" IS NULL`
	const recent = `> now() - interval '30 days'`
	err := s.db.Get(&g, `SELECT
		(SELECT count(*) FROM "PlanFile" f JOIN "PlanSet" d ON d.id = f."setId" WHERE `+live+` AND f."fileType" = 'application/pdf')::int AS "pdfAll",
		(SELECT count(*) FROM "PlanFile" f JOIN "PlanSet" d ON d.id = f."setId" WHERE `+live+` AND f."fileType" = 'application/pdf' AND f."createdAt" `+recent+`)::int AS "pdfNew",
		(SELECT count(*) FROM "PlanSet" d WHERE d."archivedAt" IS NULL)::int AS "setAll",
		(SELECT count(*) FROM "PlanSet" d WHERE d."archivedAt" IS NULL AND d."createdAt" `+recent+`)::int AS "setNew",
		(SELECT count(*) FROM "PlanFile" f JOIN "PlanSet" d ON d.id = f."setId" WHERE `+live+`)::int AS "fileAll",
		(SELECT count(*) FROM "PlanFile" f JOIN "PlanSet" d ON d.id = f."setId" WHERE `+live+` AND f."createdAt" `+recent+`)::int AS "fileNew",
		(SELECT count(DISTINCT "projectId") FROM "PlanSet" WHERE "archivedAt" IS NULL AND "projectId" IS NOT NULL)::int AS "prjAll",
		(SELECT count(*) FROM (SELECT "projectId" FROM "PlanSet" WHERE "archivedAt" IS NULL AND "projectId" IS NOT NULL
		   GROUP BY "projectId" HAVING min("createdAt") `+recent+`) x)::int AS "prjNew"`)
	if err != nil {
		return out, err
	}
	out.Pdfs = g.PdfAll
	out.Growth = PlanGrowth{
		Projects: growth(g.PrjNew, g.PrjAll-g.PrjNew), Sets: growth(g.SetNew, g.SetAll-g.SetNew),
		Files: growth(g.FileNew, g.FileAll-g.FileNew), Pdfs: growth(g.PdfNew, g.PdfAll-g.PdfNew),
	}
	if err := s.db.Select(&out.Projects, `SELECT p.id AS "projectId", p.code, p.name, p.location,
			count(d.id)::int AS sets, count(d.id) FILTER (WHERE d."archivedAt" IS NULL)::int AS "openSets",
			max(d."createdAt") AS "lastAt",
			(SELECT f."fileKey" FROM "PlanFile" f JOIN "PlanSet" d2 ON d2.id = f."setId"
			  WHERE d2."projectId" = p.id AND f."removedAt" IS NULL AND f."fileType" IN `+planImageTypes+`
			  ORDER BY f."createdAt" DESC LIMIT 1) AS cover,
			(SELECT d3.id FROM "PlanSet" d3 WHERE d3."projectId" = p.id ORDER BY (d3."archivedAt" IS NOT NULL), d3."createdAt" DESC LIMIT 1) AS "latestSet"
		FROM "PlanSet" d JOIN "Project" p ON p.id = d."projectId"
		GROUP BY p.id ORDER BY max(d."createdAt") DESC LIMIT 12`); err != nil {
		return out, err
	}
	err = s.db.Select(&out.Activity, `SELECT * FROM (
		SELECT 'set' AS kind, d.title, p.name AS sub, d.id AS "setId", d."createdAt" AS at
		  FROM "PlanSet" d LEFT JOIN "Project" p ON p.id = d."projectId"
		UNION ALL
		SELECT 'file', f.title, d.title, d.id, f."createdAt"
		  FROM "PlanFile" f JOIN "PlanSet" d ON d.id = f."setId" WHERE f."removedAt" IS NULL
		) a ORDER BY at DESC LIMIT 6`)
	return out, err
}
