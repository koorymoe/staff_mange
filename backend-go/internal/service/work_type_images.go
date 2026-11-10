package service

import (
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ 🖼️ مكتبة صور أنواع العمل (ل) 11-10 ═══
// دازاي: «اكتب طاقة شمسية تطلعلي صور طاقة شمسية… نولّد ١٠٠٠ صورة والنظام بكل عرض
// يستعمل وحدة». الكادر يرفع صور لكل نوع عمل (ProjectWorkType)، وفصل العرض (category)
// ينطابق ويا نوع العمل بالاسم، والنظام ياخذ الصورة الأقل استعمالاً (والأقدم استعمالاً
// عند التساوي) — فكل عرض يطلع بصورة غير. الاختيار ينحفظ للعرض (QuotationImagePick).

type WorkTypeImageService struct{ db *sqlx.DB }

func NewWorkTypeImageService(db *sqlx.DB) *WorkTypeImageService { return &WorkTypeImageService{db: db} }

type WorkTypeImage struct {
	ID         string     `db:"id" json:"id"`
	WorkTypeID string     `db:"workTypeId" json:"workTypeId"`
	WorkType   string     `db:"workType" json:"workType"`
	FileKey    string     `db:"fileKey" json:"fileKey"`
	UseCount   int        `db:"useCount" json:"useCount"`
	LastUsedAt *time.Time `db:"lastUsedAt" json:"lastUsedAt"`
	CreatedAt  time.Time  `db:"createdAt" json:"createdAt"`
}

// WorkTypeImageStat كم صورة لكل نوع — للشاشة.
type WorkTypeImageStat struct {
	WorkTypeID string `db:"id" json:"workTypeId"`
	Name       string `db:"name" json:"name"`
	Images     int    `db:"images" json:"images"`
}

func (s *WorkTypeImageService) Stats() ([]WorkTypeImageStat, error) {
	out := []WorkTypeImageStat{}
	err := s.db.Select(&out, `SELECT w.id, w.name, count(i.id)::int AS images FROM "ProjectWorkType" w
		LEFT JOIN "WorkTypeImage" i ON i."workTypeId" = w.id AND i."removedAt" IS NULL
		GROUP BY w.id ORDER BY w.name`)
	return out, err
}

func (s *WorkTypeImageService) List(workTypeID string) ([]WorkTypeImage, error) {
	out := []WorkTypeImage{}
	err := s.db.Select(&out, `SELECT i.id, i."workTypeId", w.name AS "workType", i."fileKey", i."useCount", i."lastUsedAt", i."createdAt"
		FROM "WorkTypeImage" i JOIN "ProjectWorkType" w ON w.id = i."workTypeId"
		WHERE i."workTypeId" = $1 AND i."removedAt" IS NULL ORDER BY i."createdAt" DESC LIMIT 500`, workTypeID)
	return out, err
}

// Add يضيف صور (مسارات مرفوعة) لنوع عمل. نفس الصورة مرتين ما تنضاف.
func (s *WorkTypeImageService) Add(emp, workTypeID string, keys []string) (int, error) {
	n := 0
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if !strings.HasPrefix(k, "/api/files/") {
			return n, errors.New("الصورة لازم تنرفع للنظام أول")
		}
		res, err := s.db.Exec(`INSERT INTO "WorkTypeImage" ("workTypeId", "fileKey", "createdById")
			SELECT $1, $2, NULLIF($3, '') WHERE EXISTS (SELECT 1 FROM "ProjectWorkType" WHERE id = $1)
			  AND NOT EXISTS (SELECT 1 FROM "WorkTypeImage" WHERE "workTypeId" = $1 AND "fileKey" = $2 AND "removedAt" IS NULL)`,
			workTypeID, k, emp)
		if err != nil {
			return n, err
		}
		m, _ := res.RowsAffected()
		n += int(m)
	}
	return n, nil
}

// Remove يشيل الصورة من المكتبة (تبقى بالقاعدة — ماكو حذف فعلي).
func (s *WorkTypeImageService) Remove(id string) error {
	_, err := s.db.Exec(`UPDATE "WorkTypeImage" SET "removedAt" = now() WHERE id = $1 AND "removedAt" IS NULL`, id)
	return err
}

// MatchWorkType نوع العمل المطابق لاسم فصل: نفس الاسم، أو الاسم جوّا الفصل، أو العكس.
func (s *WorkTypeImageService) MatchWorkType(chapter string) (string, bool) {
	c := strings.TrimSpace(chapter)
	if c == "" {
		return "", false
	}
	var id string
	err := s.db.Get(&id, `SELECT id FROM "ProjectWorkType"
		WHERE lower(name) = lower($1) OR position(lower(name) IN lower($1)) > 0 OR position(lower($1) IN lower(name)) > 0
		ORDER BY (lower(name) = lower($1)) DESC, length(name) DESC LIMIT 1`, c)
	return id, err == nil
}

// Pick صورة لكل فصل بالعرض. الي انختار قبل يرجع نفسه (العرض ثابت)، إلا إذا again=true
// (زر «صورة غيرها») — ساعتها ياخذ الأقل استعمالاً غير الحالية.
func (s *WorkTypeImageService) Pick(quotationID string, chapters []string, again bool) (map[string]string, error) {
	out := map[string]string{}
	var exists bool
	if err := s.db.Get(&exists, `SELECT EXISTS (SELECT 1 FROM "Quotation" WHERE id = $1)`, quotationID); err != nil || !exists {
		return nil, errors.New("العرض مو موجود")
	}
	for _, ch := range chapters {
		ch = strings.TrimSpace(ch)
		if ch == "" {
			continue
		}
		if !again {
			var key string
			if s.db.Get(&key, `SELECT i."fileKey" FROM "QuotationImagePick" p JOIN "WorkTypeImage" i ON i.id = p."imageId"
				WHERE p."quotationId" = $1 AND p.chapter = $2 AND i."removedAt" IS NULL`, quotationID, ch) == nil {
				out[ch] = key
				continue
			}
		}
		wt, ok := s.MatchWorkType(ch)
		if !ok {
			continue
		}
		var img struct {
			ID  string `db:"id"`
			Key string `db:"fileKey"`
		}
		err := s.db.Get(&img, `SELECT i.id, i."fileKey" FROM "WorkTypeImage" i
			WHERE i."workTypeId" = $1 AND i."removedAt" IS NULL
			  AND i.id IS DISTINCT FROM (SELECT "imageId" FROM "QuotationImagePick" WHERE "quotationId" = $2 AND chapter = $3)
			ORDER BY i."useCount", i."lastUsedAt" NULLS FIRST, random() LIMIT 1`, wt, quotationID, ch)
		if err != nil {
			continue // ماكو صور لهالنوع بعد
		}
		tx, err := s.db.Beginx()
		if err != nil {
			return nil, err
		}
		tx.MustExec(`INSERT INTO "QuotationImagePick" ("quotationId", chapter, "imageId") VALUES ($1, $2, $3)
			ON CONFLICT ("quotationId", chapter) DO UPDATE SET "imageId" = EXCLUDED."imageId", "pickedAt" = now()`, quotationID, ch, img.ID)
		tx.MustExec(`UPDATE "WorkTypeImage" SET "useCount" = "useCount" + 1, "lastUsedAt" = now() WHERE id = $1`, img.ID)
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		out[ch] = img.Key
	}
	return out, nil
}
