// أداة ترحيل: تنقل الصور والوثائق المخزّنة base64 داخل قاعدة البيانات
// إلى التخزين الخارجي (R2 أو القرص المحلي)، وتترك بمكانها الرابط بس.
//
// الاستخدام:
//
//	go run ./cmd/migratefiles            # تجربة بلا تعديل (dry-run)
//	go run ./cmd/migratefiles -apply     # التنفيذ الفعلي
//	go run ./cmd/migratefiles -apply -table Product
//
// آمنة للتكرار: أي قيمة مو data URL (يعني انترحّلت من قبل، أو رابط
// خارجي) تنتخطى. وأي صف يفشل ما يوقف الباقي — ينطبع ويكمل.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/lib/pq"

	"staffmange-api/internal/config"
	"staffmange-api/internal/database"
	"staffmange-api/internal/storage"
)

// target عمود واحد ننقله. folder يحدد مجلد التخزين المنطقي.
type target struct {
	table  string
	column string
	folder string
	// array = العمود مصفوفة نصوص (كل عنصر data URL لحاله)
	array bool
}

var targets = []target{
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
	// ملاحظة: WorkReport.tookPhotos عمود منطقي (هل انصورت الشغلة أو لا)
	// مو مصفوفة صور — اسمه مضلّل. ما يخص الترحيل.
}

func main() {
	apply := flag.Bool("apply", false, "نفّذ التعديل فعلاً (بدونها تجربة بس)")
	only := flag.String("table", "", "جدول واحد بس (اختياري)")
	flag.Parse()

	_ = godotenv.Load()
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("تعذر الاتصال بقاعدة البيانات: %v", err)
	}
	defer db.Close()

	store := buildStore(cfg)
	log.Printf("التخزين: %s", store.Kind())
	if !*apply {
		log.Printf("⚠️  وضع التجربة — ماكو أي تعديل. أضف -apply للتنفيذ.")
	}

	ctx := context.Background()
	var totalRows, totalFiles, totalBytes, failed int

	for _, t := range targets {
		if *only != "" && !strings.EqualFold(*only, t.table) {
			continue
		}
		if !columnExists(db, t.table, t.column) {
			continue
		}
		rows, files, bytes, fail := migrateTarget(ctx, db, store, t, *apply)
		if rows > 0 || fail > 0 {
			log.Printf("%-20s %-24s صفوف:%-5d ملفات:%-5d حجم:%-8s فشل:%d",
				t.table, t.column, rows, files, humanBytes(bytes), fail)
		}
		totalRows += rows
		totalFiles += files
		totalBytes += bytes
		failed += fail
	}

	log.Printf("─────────────────────────────────────────────")
	log.Printf("الإجمالي: %d صف · %d ملف · %s · %d فشل", totalRows, totalFiles, humanBytes(totalBytes), failed)
	if !*apply {
		log.Printf("هذي كانت تجربة. شغّلها مرة ثانية مع -apply حتى تنكتب.")
	}
}

func migrateTarget(ctx context.Context, db *sqlx.DB, store storage.Store, t target, apply bool) (rows, files, size, failed int) {
	if t.array {
		return migrateArray(ctx, db, store, t, apply)
	}
	return migrateScalar(ctx, db, store, t, apply)
}

// migrateScalar عمود نصي واحد.
func migrateScalar(ctx context.Context, db *sqlx.DB, store storage.Store, t target, apply bool) (rows, files, size, failed int) {
	type row struct {
		ID    string `db:"id"`
		Value string `db:"value"`
	}
	var list []row
	q := `SELECT id, "` + t.column + `" AS value FROM "` + t.table + `"
		WHERE "` + t.column + `" LIKE 'data:%'`
	if err := db.Select(&list, q); err != nil {
		log.Printf("  ✗ %s.%s: %v", t.table, t.column, err)
		return 0, 0, 0, 1
	}
	for _, r := range list {
		url, n, err := storeOne(ctx, store, t.folder, r.Value, apply)
		if err != nil {
			log.Printf("  ✗ %s %s: %v", t.table, r.ID, err)
			failed++
			continue
		}
		if apply {
			if _, err := db.Exec(`UPDATE "`+t.table+`" SET "`+t.column+`" = $2 WHERE id = $1`, r.ID, url); err != nil {
				log.Printf("  ✗ تحديث %s %s: %v", t.table, r.ID, err)
				failed++
				continue
			}
		}
		rows++
		files++
		size += n
	}
	return rows, files, size, failed
}

// migrateArray عمود مصفوفة نصوص — كل عنصر ملف لحاله.
func migrateArray(ctx context.Context, db *sqlx.DB, store storage.Store, t target, apply bool) (rows, files, size, failed int) {
	type row struct {
		ID     string         `db:"id"`
		Values pq.StringArray `db:"value"`
	}
	var list []row
	q := `SELECT id, "` + t.column + `" AS value FROM "` + t.table + `"
		WHERE EXISTS (SELECT 1 FROM unnest("` + t.column + `") x WHERE x LIKE 'data:%')`
	if err := db.Select(&list, q); err != nil {
		log.Printf("  ✗ %s.%s: %v", t.table, t.column, err)
		return 0, 0, 0, 1
	}
	for _, r := range list {
		out := make(pq.StringArray, 0, len(r.Values))
		changed := false
		for _, v := range r.Values {
			if !strings.HasPrefix(v, "data:") {
				out = append(out, v)
				continue
			}
			url, n, err := storeOne(ctx, store, t.folder, v, apply)
			if err != nil {
				log.Printf("  ✗ %s %s: %v", t.table, r.ID, err)
				failed++
				out = append(out, v) // نبقي الأصل بدل ما نضيّعه
				continue
			}
			out = append(out, url)
			files++
			size += n
			changed = true
		}
		if !changed {
			continue
		}
		if apply {
			if _, err := db.Exec(`UPDATE "`+t.table+`" SET "`+t.column+`" = $2 WHERE id = $1`, r.ID, out); err != nil {
				log.Printf("  ✗ تحديث %s %s: %v", t.table, r.ID, err)
				failed++
				continue
			}
		}
		rows++
	}
	return rows, files, size, failed
}

// storeOne يفك data URL ويخزنه، ويرجّع الرابط الجديد وحجم الملف.
func storeOne(ctx context.Context, store storage.Store, folder, dataURL string, apply bool) (string, int, error) {
	data, contentType, err := storage.DecodeDataURL(dataURL)
	if err != nil {
		return "", 0, err
	}
	// النوع من المحتوى نفسه إذا انعرف — الترويسة بالـdata URL أحياناً غلط
	if sniffed := storage.SniffContentType(data); sniffed != "" {
		contentType = sniffed
	}
	key := storage.NewKey(folder, contentType)
	if apply {
		if err := store.Put(ctx, key, data, contentType); err != nil {
			return "", 0, err
		}
	}
	return "/api/files/" + key, len(data), nil
}

func columnExists(db *sqlx.DB, table, column string) bool {
	var n int
	err := db.Get(&n, `
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2`, table, column)
	return err == nil && n > 0
}

func buildStore(cfg *config.Config) storage.Store {
	r2 := storage.R2Config{
		Bucket:    cfg.R2Bucket,
		AccessKey: cfg.R2AccessKey,
		SecretKey: cfg.R2SecretKey,
		Endpoint:  cfg.R2Endpoint,
	}
	if r2.Configured() {
		if s, err := storage.NewR2Store(r2); err == nil {
			return s
		}
	}
	s, err := storage.NewLocalStore(cfg.UploadsDir)
	if err != nil {
		log.Fatalf("تعذر تهيئة التخزين المحلي: %v", err)
	}
	return s
}

func humanBytes(n int) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	}
	return fmt.Sprintf("%d B", n)
}
