package repository

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"staffmange-api/internal/model"
)

// BotReportRepository بوت تلغرام لبلاغات الموظفين (دازاي 11-05): الربط، والمسودّة، والبلاغات.
type BotReportRepository struct {
	db *sqlx.DB
}

func NewBotReportRepository(db *sqlx.DB) *BotReportRepository {
	return &BotReportRepository{db: db}
}

// LinkTTL عمر رمز الربط. قصير بقصد: لو أحد صوّر الرابط ما يفيده بعد ١٠ دقايق.
const LinkTTL = 10 * time.Minute

// CreateLinkToken رمز ربط مرة وحدة للموظف (٢٤ بايت عشوائي، يصلح بـ /start).
// الرموز القديمة غير المستعملة للموظف نفسه تنمسح — رمز وحد صالح بكل لحظة.
func (r *BotReportRepository) CreateLinkToken(employeeID string) (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	tok := base64.RawURLEncoding.EncodeToString(b)
	tx, err := r.db.Beginx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback() //nolint:errcheck
	if _, err := tx.Exec(`DELETE FROM "TelegramLinkToken" WHERE "employeeId" = $1 AND "usedAt" IS NULL`, employeeID); err != nil {
		return "", err
	}
	if _, err := tx.Exec(`INSERT INTO "TelegramLinkToken" (token, "employeeId", "expiresAt") VALUES ($1, $2, $3)`,
		tok, employeeID, time.Now().UTC().Add(LinkTTL)); err != nil {
		return "", err
	}
	return tok, tx.Commit()
}

var (
	ErrLinkTokenInvalid = errors.New("رمز الربط منتهي أو مستعمل")
	ErrAlreadyLinked    = errors.New("الحساب مربوط أصلاً")
	ErrChatTaken        = errors.New("هذا التلغرام مربوط بحساب ثاني")
)

// LinkWithToken يستهلك الرمز (مرة وحدة، ذرّياً) ويربط المحادثة بالموظف.
// (دازاي 11-05) **الربط مرة وحدة**: الحساب المربوط ما ينربط بتلغرام ثاني، والتلغرام
// المربوط ما ياخذ حساب ثاني — الفك بس من المالك. فلو انسرق رابط ربط موظف مربوط، ما يفيد.
func (r *BotReportRepository) LinkWithToken(token string, chatID int64, username string) (string, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback() //nolint:errcheck
	var empID string
	err = tx.Get(&empID, `UPDATE "TelegramLinkToken" SET "usedAt" = now()
		WHERE token = $1 AND "usedAt" IS NULL AND "expiresAt" > $2 RETURNING "employeeId"`, token, time.Now().UTC())
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrLinkTokenInvalid
	}
	if err != nil {
		return "", err
	}
	var taken struct {
		Emp  int `db:"emp"`
		Chat int `db:"chat"`
	}
	if err := tx.Get(&taken, `SELECT count(*) FILTER (WHERE "employeeId" = $1) AS emp, count(*) FILTER (WHERE "chatId" = $2) AS chat
		FROM "TelegramLink" WHERE "employeeId" = $1 OR "chatId" = $2`, empID, chatID); err != nil {
		return "", err
	}
	if taken.Emp > 0 {
		return "", ErrAlreadyLinked
	}
	if taken.Chat > 0 {
		return "", ErrChatTaken
	}
	var un *string
	if username != "" {
		un = &username
	}
	if _, err := tx.Exec(`INSERT INTO "TelegramLink" ("employeeId", "chatId", "tgUsername") VALUES ($1, $2, $3)`, empID, chatID, un); err != nil {
		return "", err
	}
	if _, err := tx.Exec(`DELETE FROM "BotReportDraft" WHERE "chatId" = $1`, chatID); err != nil {
		return "", err
	}
	return empID, tx.Commit()
}

// BotLinkedEmployee الموظف المربوط بمحادثة — من النظام نفسه، مو من كلامه.
type BotLinkedEmployee struct {
	EmployeeID string `db:"id"`
	Name       string `db:"name"`
	Role       string `db:"role"`
	Status     string `db:"status"`
}

// ByChat الموظف المربوط بهاي المحادثة (nil = مو مربوط).
func (r *BotReportRepository) ByChat(chatID int64) (*BotLinkedEmployee, error) {
	var e BotLinkedEmployee
	err := r.db.Get(&e, `SELECT e.id, e.name, e.role::text AS role, e.status::text AS status
		FROM "TelegramLink" l JOIN "Employee" e ON e.id = l."employeeId" WHERE l."chatId" = $1`, chatID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &e, err
}

func (r *BotReportRepository) LinkOf(employeeID string) (*model.TelegramLink, error) {
	var l model.TelegramLink
	err := r.db.Get(&l, `SELECT * FROM "TelegramLink" WHERE "employeeId" = $1`, employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &l, err
}

func (r *BotReportRepository) Unlink(employeeID string) error {
	_, err := r.db.Exec(`DELETE FROM "TelegramLink" WHERE "employeeId" = $1`, employeeID)
	return err
}

// ═══ المسودّة: النص والصور يجون برسائل منفصلة، فيتجمعن لحد ما يكمل البلاغ ═══

// Draft المسودّة إذا عمرها أقل من ٣٠ دقيقة (الأقدم تنتسى). ⚠️ المقارنة بساعة
// القاعدة نفسها: العمود timestamp بلا منطقة وانكتب بـ now() القاعدة (قرار ٠٠٠٣).
func (r *BotReportRepository) Draft(chatID int64) (*model.BotReportDraft, error) {
	var d model.BotReportDraft
	err := r.db.Get(&d, `SELECT * FROM "BotReportDraft" WHERE "chatId" = $1 AND "updatedAt" > now() - interval '30 minutes'`, chatID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &d, err
}

func (r *BotReportRepository) SaveDraft(d *model.BotReportDraft) error {
	if d.Photos == nil {
		d.Photos = pq.StringArray{} // nil ينكتب NULL والعمود NOT NULL
	}
	_, err := r.db.Exec(`INSERT INTO "BotReportDraft" ("chatId", "employeeId", body, photos, "updatedAt")
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT ("chatId") DO UPDATE SET "employeeId" = $2, body = $3, photos = $4, "updatedAt" = now()`,
		d.ChatID, d.EmployeeID, d.Body, d.Photos)
	return err
}

func (r *BotReportRepository) ClearDraft(chatID int64) error {
	_, err := r.db.Exec(`DELETE FROM "BotReportDraft" WHERE "chatId" = $1`, chatID)
	return err
}

// ═══ البلاغات ═══

// Create يحفظ البلاغ ويرجع رقمه للموظف.
func (r *BotReportRepository) Create(rep *model.BotReport) error {
	rep.ID = uuid.NewString()
	if rep.Source == "" {
		rep.Source = "telegram"
	}
	return r.db.Get(rep, `INSERT INTO "BotReport" (id, "employeeId", "chatId", body, photos, "mediaGroupId", source, page)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, number, "employeeId", "chatId", body, photos, "mediaGroupId", status, "adminNote", "createdAt", "updatedAt", source, page, "photosPurgedAt", '' AS "employeeName", '' AS "employeeRole"`,
		rep.ID, rep.EmployeeID, rep.ChatID, rep.Body, pq.StringArray(rep.Photos), rep.MediaGroupID, rep.Source, rep.Page)
}

// Employee اسم الموظف ودوره (للبلاغ الي ينكتب من النظام).
func (r *BotReportRepository) Employee(id string) (*BotLinkedEmployee, error) {
	var e BotLinkedEmployee
	err := r.db.Get(&e, `SELECT id, name, role::text AS role, status::text AS status FROM "Employee" WHERE id = $1`, id)
	return &e, err
}

// BotLinkRow صف بشاشة المالك: منو مربوط بأي تلغرام.
type BotLinkRow struct {
	EmployeeID string    `db:"employeeId" json:"employeeId"`
	Name       string    `db:"name" json:"name"`
	Role       string    `db:"role" json:"role"`
	Status     string    `db:"status" json:"status"`
	TgUsername *string   `db:"tgUsername" json:"tgUsername"`
	LinkedAt   time.Time `db:"linkedAt" json:"linkedAt"`
}

func (r *BotReportRepository) Links() ([]BotLinkRow, error) {
	out := []BotLinkRow{}
	err := r.db.Select(&out, `SELECT l."employeeId", e.name, e.role::text AS role, e.status::text AS status, l."tgUsername", l."linkedAt"
		FROM "TelegramLink" l JOIN "Employee" e ON e.id = l."employeeId" ORDER BY l."linkedAt" DESC`)
	return out, err
}

// AttachToGroup ألبوم صور: التلغرام يرسل كل صورة برسالة، وكلها بنفس media_group_id.
// البلاغ ينخلق ويا أول صورة، والباقي تنضاف له (خلال ٣ دقايق). يرجع رقم البلاغ (٠ = ماكو).
func (r *BotReportRepository) AttachToGroup(chatID int64, group, photo string) (int, error) {
	var n int
	err := r.db.Get(&n, `UPDATE "BotReport" SET photos = array_append(photos, $3), "updatedAt" = now()
		WHERE "chatId" = $1 AND "mediaGroupId" = $2 AND "createdAt" > now() - interval '3 minutes'
		  AND cardinality(photos) < 10
		RETURNING number`, chatID, group, photo)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return n, err
}

// CountLastHour بلاغات الموظف بآخر ساعة — حد الإزعاج (بساعة القاعدة، نفس عمود createdAt).
func (r *BotReportRepository) CountLastHour(employeeID string) (int, error) {
	var n int
	err := r.db.Get(&n, `SELECT count(*) FROM "BotReport" WHERE "employeeId" = $1 AND "createdAt" > now() - interval '1 hour'`, employeeID)
	return n, err
}

const botReportCols = `b.id, b.number, b."employeeId", b."chatId", b.body, b.photos, b."mediaGroupId", b.status, b."adminNote",
	b."createdAt", b."updatedAt", b.source, b.page, b."photosPurgedAt", e.name AS "employeeName", e.role::text AS "employeeRole"`

// List البلاغات (الأحدث أول). status فاضي = الكل. حد ٣٠٠ (٧م).
func (r *BotReportRepository) List(status string) ([]model.BotReport, error) {
	out := []model.BotReport{}
	err := r.db.Select(&out, `SELECT `+botReportCols+` FROM "BotReport" b JOIN "Employee" e ON e.id = b."employeeId"
		WHERE ($1 = '' OR b.status = $1) ORDER BY b."createdAt" DESC LIMIT 300`, status)
	return out, err
}

// Mine آخر بلاغات الموظف (لأمر /my بالبوت).
func (r *BotReportRepository) Mine(employeeID string, limit int) ([]model.BotReport, error) {
	out := []model.BotReport{}
	err := r.db.Select(&out, `SELECT `+botReportCols+` FROM "BotReport" b JOIN "Employee" e ON e.id = b."employeeId"
		WHERE b."employeeId" = $1 ORDER BY b."createdAt" DESC LIMIT $2`, employeeID, limit)
	return out, err
}

func (r *BotReportRepository) Get(id string) (*model.BotReport, error) {
	var b model.BotReport
	err := r.db.Get(&b, `SELECT `+botReportCols+` FROM "BotReport" b JOIN "Employee" e ON e.id = b."employeeId" WHERE b.id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &b, err
}

// SetStatus يغيّر الحالة والملاحظة (الملاحظة nil = تبقى مثل ما هي).
func (r *BotReportRepository) SetStatus(id, status string, note *string) error {
	_, err := r.db.Exec(`UPDATE "BotReport" SET status = $2, "adminNote" = COALESCE($3, "adminNote"), "updatedAt" = now() WHERE id = $1`, id, status, note)
	return err
}

// OwnerChats محادثات المالك المربوطة — لإشعار «بلاغ جديد» الفوري (رسائل البوت مجانية).
func (r *BotReportRepository) OwnerChats() ([]int64, error) {
	out := []int64{}
	err := r.db.Select(&out, `SELECT l."chatId" FROM "TelegramLink" l JOIN "Employee" e ON e.id = l."employeeId"
		WHERE e.role::text = 'OWNER' AND e.status::text = 'ACTIVE'`)
	return out, err
}

// ═══ تنظيف الصور الأسبوعي (دازاي 11-05: «كل اسبوع يتصفن وينحذفن من السيرفر») ═══

// PhotosOlderThanWeek البلاغات الي صورها عمرها أكثر من ٧ أيام وبعدها ما انمسحت.
func (r *BotReportRepository) PhotosOlderThanWeek() ([]model.BotReport, error) {
	out := []model.BotReport{}
	err := r.db.Select(&out, `SELECT id, number, photos FROM "BotReport"
		WHERE cardinality(photos) > 0 AND "createdAt" < now() - interval '7 days' ORDER BY "createdAt" LIMIT 1000`)
	return out, err
}

// MarkPhotosPurged يفرّغ الصور ويأشّر وقت المسح (النص يبقى).
func (r *BotReportRepository) MarkPhotosPurged(id string) error {
	_, err := r.db.Exec(`UPDATE "BotReport" SET photos = '{}', "photosPurgedAt" = now() WHERE id = $1`, id)
	return err
}

// StaleDrafts المسودّات الأقدم من يوم (صور انرفعت وما كمل البلاغ) — صورها تنمسح ويا المسودّة.
func (r *BotReportRepository) StaleDrafts() ([]model.BotReportDraft, error) {
	out := []model.BotReportDraft{}
	err := r.db.Select(&out, `SELECT * FROM "BotReportDraft" WHERE "updatedAt" < now() - interval '1 day' LIMIT 1000`)
	return out, err
}
