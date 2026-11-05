package repository

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/storage"
)

// InlineFileColumn عمود بالقاعدة ممكن بيه ملف محفوظ data URL (base64).
// Array = العمود مصفوفة نصوص (كل عنصر ملف لحاله).
type InlineFileColumn struct {
	Table, Column, Folder string
	Array                 bool
}

// InlineFileColumns (٧ل): القائمة الوحيدة لأعمدة الملفات — تستعملها أداة
// cmd/migratefiles والترحيل التلقائي وقت تشغيل السيرفر.
// (WorkReport.tookPhotos منطقي مو صور — ما يخص.)
var InlineFileColumns = []InlineFileColumn{
	{"Product", "imageBase64", "products", false},
	{"ProductProcurement", "receiptImage", "receipts", false},
	{"RevolvingFundTxn", "receiptImage", "receipts", false},
	{"VehicleLog", "receiptPhotoBase64", "vehicles", false},
	{"VehicleDocument", "fileUrl", "vehicles", false},
	{"Project", "contractPdfBase64", "projects", false},
	{"Project", "signedContractPdfBase64", "projects", false},
	{"GpsDeviceRequest", "invoicePhotoUrl", "gps", false},
	{"Exhibition", "businessCardPhotos", "exhibitions", true},
	{"ProjectChecklist", "photoUrls", "reports", true},
}

// MigrateInlineFiles (٧ل، 11-05): الملفات القديمة المحفوظة base64 داخل القاعدة
// تنتقل لمخزن الملفات بالخلفية. لكل ملف: يحفظه، **يقراه ويقارن البايتات**،
// وبس بعدها يبدّل العمود برابط /api/files/… بشرط القيمة ما تغيّرت بنفس اللحظة.
// أي فشل = الملف يبقى مثل ما هو (ماكو شي ينمسح بلا نسخة مؤكدة).
// أعمدة المصفوفات تبقى للأداة اليدوية. ٥٠٠ صف لكل عمود بكل تشغيل.
func MigrateInlineFiles(db *sqlx.DB, store storage.Store) map[string]int {
	moved := map[string]int{}
	ctx := context.Background()
	for _, c := range InlineFileColumns {
		if c.Array {
			continue
		}
		rows := []struct {
			ID  string `db:"id"`
			Val string `db:"value"`
		}{}
		q := fmt.Sprintf(`SELECT id, %q AS value FROM %q WHERE %q LIKE 'data:%%;base64,%%' LIMIT 500`, c.Column, c.Table, c.Column)
		if err := db.Select(&rows, q); err != nil {
			continue // الجدول أو العمود مو موجود بهاي النسخة
		}
		for _, r := range rows {
			url, ok := storeVerified(ctx, store, c.Folder, r.Val)
			if !ok {
				continue
			}
			res, err := db.Exec(fmt.Sprintf(`UPDATE %q SET %q = $2 WHERE id = $1 AND %q = $3`, c.Table, c.Column, c.Column), r.ID, url, r.Val)
			if err != nil {
				continue
			}
			if n, _ := res.RowsAffected(); n == 1 {
				moved[c.Table+"."+c.Column]++
			}
		}
	}
	return moved
}

func storeVerified(ctx context.Context, store storage.Store, folder, dataURL string) (string, bool) {
	head, b64, ok := strings.Cut(dataURL, ",")
	if !ok {
		return "", false
	}
	ct := strings.TrimSuffix(strings.TrimPrefix(head, "data:"), ";base64")
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil || len(data) == 0 {
		return "", false
	}
	key := storage.NewKey(folder, ct)
	if err := store.Put(ctx, key, data, ct); err != nil {
		return "", false
	}
	back, _, err := store.Get(ctx, key)
	if err != nil || !bytes.Equal(back, data) {
		return "", false
	}
	return "/api/files/" + key, true
}
