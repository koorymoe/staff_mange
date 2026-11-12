package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ ☀️ استفسارات الطاقة الشمسية (دازاي 11-12) ═══
// «زبون يتصل يستفسر… يسجلون رقمه بالاستفسار، ويترحل للتقنيين حتى يتواصلون وياه ويشوفون
//  شنو يريد. ونريد نتأكد إن التقنيين جاي يتواصلون… يترحل لابو الجودة يتأكد: هل فعلاً
//  التقني اتصل؟ الزبون چان بس يستفسر؟ قدّمله عروض أسعار مناسبة ومنافسة للسوق؟ وإذا التقني
//  كاتب «ما رد»، ابو الجودة يتصل بالزبون ويسأله: اتصل بيك التقني؟»
//
// المراحل:
//   NEW          انسجل وينتظر تقني (بعد ٢٤ ساعة بلا تواصل = متأخر، ويطلع لابو الجودة)
//   TECH_DONE    التقني سجّل نتيجة اتصاله (حتى «ما رد») ← لابو الجودة
//   CLOSED       ابو الجودة اتصل بالزبون وسجّل الأجوبة
// المفارقة: التقني كاتب اتصل، والزبون يگول ما اتصل بيه ← تنبيه للمدير.
//
// الصلاحيات (ينطيها المالك لمن يريد): solar_inquiry = يسجّل؛ solar_inquiry_tech = التقني؛
// quality_control = ابو الجودة. المدير والمالك يشوفون كلشي. الي ماله دخل = 404.

var ErrInquiryHidden = errors.New("مو موجود")

const inquiryTechSLA = 24 * time.Hour

type SolarInquiry struct {
	ID            string     `db:"id" json:"id"`
	Number        int        `db:"number" json:"number"`
	CustomerName  string     `db:"customerName" json:"customerName"`
	Phone         string     `db:"phone" json:"phone"`
	Area          *string    `db:"area" json:"area"`
	Note          *string    `db:"note" json:"note"`
	CreatedBy     *string    `db:"createdBy" json:"createdBy"`
	CreatedAt     time.Time  `db:"createdAt" json:"createdAt"`
	TechID        *string    `db:"techId" json:"techId"`
	TechName      *string    `db:"techName" json:"techName"`
	TechOutcome   *string    `db:"techOutcome" json:"techOutcome"`
	TechNote      *string    `db:"techNote" json:"techNote"`
	TechContactAt *time.Time `db:"techContactAt" json:"techContactAt"`
	QuoteGiven    *bool      `db:"quoteGiven" json:"quoteGiven"`
	QuoteAmount   *float64   `db:"quoteAmount" json:"quoteAmount"`
	QualityBy     *string    `db:"qualityBy" json:"qualityBy"`
	QualityAt     *time.Time `db:"qualityAt" json:"qualityAt"`
	QReached      *bool      `db:"qCustomerReached" json:"qCustomerReached"`
	QTechCalled   *string    `db:"qTechCalled" json:"qTechCalled"`
	QIntent       *string    `db:"qIntent" json:"qIntent"`
	QGotQuote     *string    `db:"qGotQuote" json:"qGotQuote"`
	QQuoteFair    *string    `db:"qQuoteFair" json:"qQuoteFair"`
	QNote         *string    `db:"qNote" json:"qNote"`

	Stage    string `db:"-" json:"stage"`    // NEW · TECH_DONE · CLOSED
	Overdue  bool   `db:"-" json:"overdue"`  // ما تواصل تقني خلال ٢٤ ساعة
	Mismatch bool   `db:"-" json:"mismatch"` // التقني گال اتصل، والزبون گال لا
}

// InquiryAccess شنو يگدر يسوي هذا الشخص.
type InquiryAccess struct {
	Intake  bool `json:"intake"`
	Tech    bool `json:"tech"`
	Quality bool `json:"quality"`
	Admin   bool `json:"admin"`
}

func (a InquiryAccess) any() bool { return a.Intake || a.Tech || a.Quality || a.Admin }

type SolarInquiryService struct {
	db    *sqlx.DB
	notif *repository.NotificationRepository
}

func NewSolarInquiryService(db *sqlx.DB, notif *repository.NotificationRepository) *SolarInquiryService {
	return &SolarInquiryService{db: db, notif: notif}
}

func (s *SolarInquiryService) Access(emp, role string) InquiryAccess {
	a := InquiryAccess{Admin: model.IsAdminRole(role)}
	_ = s.db.QueryRow(`SELECT `+model.HasPermSQL(`$1`, `= 'solar_inquiry'`)+`, `+model.HasPermSQL(`$1`, `= 'solar_inquiry_tech'`)+`, `+
		model.HasPermSQL(`$1`, `= 'quality_control'`)+` OR EXISTS (SELECT 1 FROM "Employee" WHERE id = $1 AND role::text = 'QUALITY_ENGINEER')`, emp).
		Scan(&a.Intake, &a.Tech, &a.Quality)
	return a
}

const inquirySelect = `SELECT q.id, q.number, q."customerName", q.phone, q.area, q.note, c.name AS "createdBy", q."createdAt",
	q."techId", t.name AS "techName", q."techOutcome", q."techNote", q."techContactAt", q."quoteGiven", q."quoteAmount"::float8 AS "quoteAmount",
	g.name AS "qualityBy", q."qualityAt", q."qCustomerReached", q."qTechCalled", q."qIntent", q."qGotQuote", q."qQuoteFair", q."qNote"
	FROM "SolarInquiry" q LEFT JOIN "Employee" c ON c.id = q."createdById"
	LEFT JOIN "Employee" t ON t.id = q."techId" LEFT JOIN "Employee" g ON g.id = q."qualityById"`

func decorate(x *SolarInquiry, now time.Time) {
	switch {
	case x.QualityAt != nil:
		x.Stage = "CLOSED"
	case x.TechOutcome != nil:
		x.Stage = "TECH_DONE"
	default:
		x.Stage = "NEW"
		x.Overdue = now.Sub(x.CreatedAt) > inquiryTechSLA
	}
	x.Mismatch = x.TechOutcome != nil && *x.TechOutcome != "NO_ANSWER" && x.QTechCalled != nil && *x.QTechCalled == "NO"
}

// List الاستفسارات حسب الصلاحية: المسجّل يشوف الي سجّلها؛ التقني والجودة والمدير الكل (آخر ٩٠ يوم).
func (s *SolarInquiryService) List(emp, role string) ([]SolarInquiry, InquiryAccess, error) {
	a := s.Access(emp, role)
	out := []SolarInquiry{}
	if !a.any() {
		return out, a, ErrInquiryHidden
	}
	where := `q."createdAt" > now() - interval '90 days'`
	args := []any{}
	if !a.Tech && !a.Quality && !a.Admin {
		where += ` AND q."createdById" = $1`
		args = append(args, emp)
	}
	if err := s.db.Select(&out, inquirySelect+` WHERE `+where+` ORDER BY q."createdAt" DESC LIMIT 500`, args...); err != nil {
		return nil, a, err
	}
	now := time.Now()
	for i := range out {
		decorate(&out[i], now)
	}
	return out, a, nil
}

func (s *SolarInquiryService) get(id string) (*SolarInquiry, error) {
	var x SolarInquiry
	if err := s.db.Get(&x, inquirySelect+` WHERE q.id = $1`, id); err != nil {
		return nil, ErrInquiryHidden
	}
	decorate(&x, time.Now())
	return &x, nil
}

// Create تسجيل استفسار (صلاحية solar_inquiry أو المدير) ← ينبّه التقنيين.
func (s *SolarInquiryService) Create(emp, role, name, phone, area, note string) (*SolarInquiry, error) {
	a := s.Access(emp, role)
	if !a.Intake && !a.Admin {
		return nil, ErrInquiryHidden
	}
	name, phone = strings.TrimSpace(name), strings.TrimSpace(phone)
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)
	if name == "" || len(digits) < 10 {
		return nil, errors.New("اكتب اسم الزبون ورقم هاتف صحيح")
	}
	var id string
	if err := s.db.Get(&id, `INSERT INTO "SolarInquiry" ("customerName", phone, area, note, "createdById")
		VALUES ($1, $2, NULLIF($3,''), NULLIF($4,''), $5) RETURNING id`, name, phone, strings.TrimSpace(area), strings.TrimSpace(note), emp); err != nil {
		return nil, err
	}
	x, err := s.get(id)
	if err == nil && s.notif != nil {
		_ = s.notif.CreateForPermission("solar_inquiry_tech", "SOLAR_INQUIRY",
			fmt.Sprintf("☀️ استفسار طاقة شمسية جديد #%d — تواصل ويا الزبون وسجّل النتيجة (لازم خلال ٢٤ ساعة)", x.Number))
	}
	return x, err
}

// Claim التقني ياخذ الاستفسار (حتى ما يتصلون اثنين بنفس الزبون).
func (s *SolarInquiryService) Claim(emp, role, id string) (*SolarInquiry, error) {
	a := s.Access(emp, role)
	if !a.Tech && !a.Admin {
		return nil, ErrInquiryHidden
	}
	res, err := s.db.Exec(`UPDATE "SolarInquiry" SET "techId" = $2, "techClaimedAt" = now()
		WHERE id = $1 AND ("techId" IS NULL OR "techId" = $2) AND "techOutcome" IS NULL`, id, emp)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("الاستفسار ماخذه تقني ثاني أو انسجلت نتيجته")
	}
	return s.get(id)
}

// TechContact التقني يسجّل نتيجة اتصاله ← يروح لابو الجودة.
func (s *SolarInquiryService) TechContact(emp, role, id, outcome, note string, quoteGiven *bool, quoteAmount *float64) (*SolarInquiry, error) {
	a := s.Access(emp, role)
	if !a.Tech && !a.Admin {
		return nil, ErrInquiryHidden
	}
	switch outcome {
	case "JUST_ASKING", "WANTS_INSTALL", "NO_ANSWER", "NOT_INTERESTED":
	default:
		return nil, errors.New("اختار نتيجة الاتصال")
	}
	if outcome != "NO_ANSWER" && len([]rune(strings.TrimSpace(note))) < 5 {
		return nil, errors.New("اكتب شنو حچيتوا (جملة وحدة على الأقل)")
	}
	res, err := s.db.Exec(`UPDATE "SolarInquiry" SET "techId" = COALESCE("techId", $2), "techOutcome" = $3, "techNote" = NULLIF($4,''),
			"techContactAt" = now(), "quoteGiven" = $5, "quoteAmount" = $6
		WHERE id = $1 AND ("techId" IS NULL OR "techId" = $2 OR $7) AND "qualityAt" IS NULL`,
		id, emp, outcome, strings.TrimSpace(note), quoteGiven, quoteAmount, a.Admin)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrInquiryHidden
	}
	x, err := s.get(id)
	if err == nil && s.notif != nil {
		msg := fmt.Sprintf("⭐ استفسار #%d: التقني سجّل اتصاله — اتصل بالزبون وتأكد", x.Number)
		if outcome == "NO_ANSWER" {
			msg = fmt.Sprintf("⭐ استفسار #%d: التقني كاتب «ما رد» — اتصل بالزبون واسأله: اتصل بيك التقني؟", x.Number)
		}
		_ = s.notif.CreateForPermission("quality_control", "SOLAR_INQUIRY", msg)
	}
	return x, err
}

// QualityInput أجوبة ابو الجودة من اتصاله بالزبون.
type QualityInput struct {
	Reached    bool   `json:"customerReached"`
	TechCalled string `json:"techCalled"` // YES · NO · UNSURE
	Intent     string `json:"intent"`     // JUST_ASKING · WANTS_INSTALL · NOT_INTERESTED
	GotQuote   string `json:"gotQuote"`   // YES · NO · UNSURE
	QuoteFair  string `json:"quoteFair"`  // YES · NO · UNSURE
	Note       string `json:"note"`
}

func oneOf(v string, ok ...string) bool {
	for _, o := range ok {
		if v == o {
			return true
		}
	}
	return false
}

// QualityCheck ابو الجودة يسجّل اتصاله. يگدر يسويه حتى قبل التقني (إذا التقني متأخر).
func (s *SolarInquiryService) QualityCheck(emp, role, id string, in QualityInput) (*SolarInquiry, error) {
	a := s.Access(emp, role)
	if !a.Quality && !a.Admin {
		return nil, ErrInquiryHidden
	}
	if in.Reached {
		if !oneOf(in.TechCalled, "YES", "NO", "UNSURE") || !oneOf(in.Intent, "JUST_ASKING", "WANTS_INSTALL", "NOT_INTERESTED") ||
			!oneOf(in.GotQuote, "YES", "NO", "UNSURE") || !oneOf(in.QuoteFair, "YES", "NO", "UNSURE") {
			return nil, errors.New("جاوب على كل الأسئلة")
		}
	} else {
		in.TechCalled, in.Intent, in.GotQuote, in.QuoteFair = "", "", "", ""
	}
	res, err := s.db.Exec(`UPDATE "SolarInquiry" SET "qualityById" = $2, "qualityAt" = now(), "qCustomerReached" = $3,
			"qTechCalled" = NULLIF($4,''), "qIntent" = NULLIF($5,''), "qGotQuote" = NULLIF($6,''), "qQuoteFair" = NULLIF($7,''), "qNote" = NULLIF($8,'')
		WHERE id = $1 AND "qualityAt" IS NULL`, id, emp, in.Reached, in.TechCalled, in.Intent, in.GotQuote, in.QuoteFair, strings.TrimSpace(in.Note))
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, errors.New("الاستفسار مسكّر من قبل")
	}
	x, err := s.get(id)
	if err == nil && s.notif != nil && x.Mismatch {
		_ = s.notif.CreateForRole("ADMIN", "SOLAR_INQUIRY", fmt.Sprintf("🚩 استفسار #%d: التقني %s كاتب «اتصلت» والزبون يگول ما اتصل بيه أحد",
			x.Number, deref(x.TechName, "—")))
	}
	return x, err
}
