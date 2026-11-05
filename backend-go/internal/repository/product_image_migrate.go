package repository

import (
	"bytes"
	"context"
	"encoding/base64"
	"strings"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/storage"
)

// MigrateProductImages (٧ل، 11-05): صور المنتجات القديمة محفوظة base64 داخل القاعدة،
// فقائمة المنتجات تنزّل صور الكل ويا البيانات. هذا ينقلها لمخزن الملفات مرة وحدة:
// يحفظ الملف، **يقراه ويقارنه**، وبس بعدها يبدّل العمود برابط /api/files/…
// أي فشل = الصورة تبقى مثل ما هي (ماكو شي ينمسح بلا نسخة مؤكدة).
// عروض الأسعار القديمة ما تنلمس: صورتها لقطة وقت الإصدار.
func MigrateProductImages(db *sqlx.DB, store storage.Store) (moved int, err error) {
	rows := []struct {
		ID  string `db:"id"`
		Img string `db:"imageBase64"`
	}{}
	if err := db.Select(&rows, `SELECT id, "imageBase64" FROM "Product" WHERE "imageBase64" LIKE 'data:image/%;base64,%' LIMIT 500`); err != nil {
		return 0, err
	}
	ctx := context.Background()
	for _, r := range rows {
		head, b64, ok := strings.Cut(r.Img, ",")
		if !ok {
			continue
		}
		ct := strings.TrimSuffix(strings.TrimPrefix(head, "data:"), ";base64")
		data, err := base64.StdEncoding.DecodeString(b64)
		if err != nil || len(data) == 0 {
			continue
		}
		key := storage.NewKey("products", ct)
		if err := store.Put(ctx, key, data, ct); err != nil {
			continue
		}
		back, _, err := store.Get(ctx, key)
		if err != nil || !bytes.Equal(back, data) {
			continue
		}
		// الشرط على القيمة القديمة: لو أحد غيّر الصورة بنفس اللحظة ما ندوس عليها
		res, err := db.Exec(`UPDATE "Product" SET "imageBase64" = $2 WHERE id = $1 AND "imageBase64" = $3`, r.ID, "/api/files/"+key, r.Img)
		if err != nil {
			continue
		}
		if n, _ := res.RowsAffected(); n == 1 {
			moved++
		}
	}
	return moved, nil
}
