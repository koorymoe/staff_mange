package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// EntityAvatarModel مجسّم ثلاثي الأبعاد مرفوع من داخل النظام.
type EntityAvatarModel struct {
	ID           string     `db:"id" json:"id"`
	Label        string     `db:"label" json:"label"`
	FileKey      string     `db:"fileKey" json:"fileKey"`
	FileType     string     `db:"fileType" json:"fileType"`
	SizeBytes    int64      `db:"sizeBytes" json:"sizeBytes"`
	UploadedByID *string    `db:"uploadedById" json:"uploadedById"`
	ArchivedAt   *time.Time `db:"archivedAt" json:"archivedAt"`
	// IsActive شخصية النظام الحالية — الودجة وورقة القصة تحمّلان هذا.
	//
	// ⚠️ **إضافة الخانة هنا إلزامية مو تجميلاً**: كل الاستعلامات بهذا
	// الملف `SELECT *`، و`sqlx` يفشل على عمود ما يلگى إله حقلاً —
	// فخانة ناقصة بالبنية تكسر القراءة كلها بلا علاقة بالميزة.
	IsActive  bool      `db:"isActive" json:"isActive"`
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

type EntityModelRepository struct{ db *sqlx.DB }

func NewEntityModelRepository(db *sqlx.DB) *EntityModelRepository {
	return &EntityModelRepository{db: db}
}

// List المجسّمات غير المؤرشفة، الأحدث أول.
func (r *EntityModelRepository) List() ([]EntityAvatarModel, error) {
	rows := []EntityAvatarModel{}
	err := r.db.Select(&rows, `
		SELECT * FROM "EntityAvatarModel"
		WHERE "archivedAt" IS NULL
		ORDER BY "createdAt" DESC`)
	return rows, err
}

// Get صفّ واحد بمعرّفه — نحتاجه قبل الأرشفة وقبل الخدمة.
func (r *EntityModelRepository) Get(id string) (*EntityAvatarModel, error) {
	var row EntityAvatarModel
	if err := r.db.Get(&row, `SELECT * FROM "EntityAvatarModel" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &row, nil
}

// Create يسجّل مجسّماً مرفوعاً.
//
// ⚠️ الملف نفسه يكون **مخزوناً** قبل هالنداء، والصف يحمل مفتاحه بس.
func (r *EntityModelRepository) Create(label, fileKey, fileType string, size int64, uploadedBy string) (*EntityAvatarModel, error) {
	label = strings.TrimSpace(label)
	if label == "" {
		return nil, fmt.Errorf("اسم المجسّم مطلوب")
	}
	if strings.TrimSpace(fileKey) == "" {
		return nil, fmt.Errorf("ملف المجسّم مطلوب")
	}
	var row EntityAvatarModel
	err := r.db.Get(&row, `
		INSERT INTO "EntityAvatarModel" (id, label, "fileKey", "fileType", "sizeBytes", "uploadedById")
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING *`,
		uuid.NewString(), label, fileKey, fileType, size, uploadedBy)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// GetActive شخصية النظام الحالية.
//
// ⚠️ **ماكو نشط ≠ خطأ**: هاي الحالة الطبيعية لنظام ما بدّل شخصيته بعد،
// فنرجّع `(nil, nil)` والواجهة تقع على المجسّم المدمج. لو رجّعناها خطأً
// چان كل موظف يشوف خطأً بالكونسول بلا سبب.
func (r *EntityModelRepository) GetActive() (*EntityAvatarModel, error) {
	var row EntityAvatarModel
	err := r.db.Get(&row, `
		SELECT * FROM "EntityAvatarModel"
		WHERE "isActive" AND "archivedAt" IS NULL`)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// Activate يخلي مجسّماً شخصيةَ النظام لكل الموظفين.
//
// ⚠️ **بمعاملة واحدة**: التصفير والتفعيل لازم يصيرون معاً. لو صفّرنا
// ثم فشل التفعيل، يبقى النظام **بلا شخصية نشطة** وكل موظف يرجع
// للمجسّم المدمج — تراجع صامت. والفهرس الفريد يمنع نشطين، فالتصفير
// **قبل** التفعيل مو اختياراً.
//
// وID فاضي = «رجّعني للمجسّم المدمج»: نصفّر بس.
func (r *EntityModelRepository) Activate(id string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`UPDATE "EntityAvatarModel" SET "isActive" = false WHERE "isActive"`); err != nil {
		return err
	}
	if strings.TrimSpace(id) != "" {
		res, err := tx.Exec(`
			UPDATE "EntityAvatarModel" SET "isActive" = true
			WHERE id = $1 AND "archivedAt" IS NULL`, id)
		if err != nil {
			return err
		}
		// 🔴 المؤرشف ما ينفعّل: ملفه ممكن يكون مقصوداً للحذف، وتفعيله
		// يعني شخصيةً تختفي بلا ما نعرف ليش.
		if n, _ := res.RowsAffected(); n == 0 {
			return fmt.Errorf("المجسّم غير موجود أو مؤرشف")
		}
	}
	return tx.Commit()
}

// Archive أرشفة ناعمة — ما نحذف ملفاً ممكن يكون معروضاً.
func (r *EntityModelRepository) Archive(id string) error {
	res, err := r.db.Exec(`
		UPDATE "EntityAvatarModel"
		-- ⚠️ والتصفير **مع** الأرشفة: بلاه يبقى صفٌّ مؤرشف نشطاً،
		-- والودجة تطلب ملفاً مقصوداً للإخفاء — كسرة صامتة بشاشة كل
		-- موظف.
		SET "archivedAt" = CURRENT_TIMESTAMP, "isActive" = false
		WHERE id = $1 AND "archivedAt" IS NULL`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("المجسّم غير موجود أو مؤرشف أصلاً")
	}
	return nil
}
