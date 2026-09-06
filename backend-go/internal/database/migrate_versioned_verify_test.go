package database

// اختبار حي (Postgres حقيقي عبر DATABASE_URL) لآلية الترحيل الإصداري الجديدة
// (جدول "SchemaMigration" و runVersionedMigrations بـ migrate.go) اللي حلّت محل
// النظام القديم "أعد تنفيذ كل شي كل إقلاع". يتحقق من اثنين:
//  1. تشغيل Migrate مرتين متتاليتين على نفس القاعدة ينجح بالمرتين (الثانية
//     تتخطى كل الترحيلات المسجّلة أصلاً بدون أي خطأ).
//  2. تشغيل Migrate على schema فارغة تماماً (schema جديد داخل نفس القاعدة، بدون
//     أي جدول موجود مسبقاً) ينجح من الصفر وينشئ جدول "SchemaMigration" ويطبّق
//     كل الترحيلات.

import (
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestMigrateIsIdempotent_Live يتأكد إن تشغيل Migrate مرتين على نفس القاعدة الحية
// ينجح بالمرتين وإن التشغيلة الثانية ما تعيد تطبيق أي ترحيل (skip-through فقط).
func TestMigrateIsIdempotent_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار الترحيل الحي")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	defer db.Close()

	if err := Migrate(db, "testowner", "testpassword123"); err != nil {
		t.Fatalf("first Migrate run failed: %v", err)
	}

	var countAfterFirst int
	if err := db.Get(&countAfterFirst, `SELECT COUNT(*) FROM "SchemaMigration"`); err != nil {
		t.Fatalf("could not count SchemaMigration rows after first run: %v", err)
	}
	if countAfterFirst == 0 {
		t.Fatalf("expected at least one applied migration to be recorded after first run")
	}

	if err := Migrate(db, "testowner", "testpassword123"); err != nil {
		t.Fatalf("second Migrate run (should be a clean no-op skip-through) failed: %v", err)
	}

	var countAfterSecond int
	if err := db.Get(&countAfterSecond, `SELECT COUNT(*) FROM "SchemaMigration"`); err != nil {
		t.Fatalf("could not count SchemaMigration rows after second run: %v", err)
	}
	if countAfterSecond != countAfterFirst {
		t.Fatalf("expected identical SchemaMigration row count after second run (no new migrations applied), got %d before / %d after", countAfterFirst, countAfterSecond)
	}
	t.Logf("SchemaMigration rows: %d (stable across both runs)", countAfterFirst)
}

// TestMigrateOnFreshSchema_Live يتأكد إن Migrate ينجح من الصفر على قاعدة بيانات
// فارغة تماماً (بدون أي جدول أو نوع "ENUM" موجود مسبقاً) — يحاكي أول إقلاع للسيرفر
// على قاعدة جديدة كلياً. نستخدم CREATE DATABASE حقيقية (مو مجرد schema داخل نفس
// القاعدة) لأن أنواع Postgres المخصصة (زي "EmployeeStatus") تعيش بكاتالوج pg_type
// المشترك على مستوى القاعدة كلها بغض النظر عن الـschema — لو استخدمنا schema فرعي
// بنفس القاعدة الحية اللي عندها هذي الأنواع أصلاً بـpublic، فحص "IF NOT EXISTS"
// (اللي يتحقق من pg_type.typname بدون تأهيل schema) كان راح يعتبرها موجودة أصلاً
// ويتخطى إنشاءها بالـschema الجديد، فتفشل باقي الترحيلات اللي تعتمد عليها.
func TestMigrateOnFreshSchema_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار الترحيل على قاعدة فارغة")
	}

	adminDB, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect error: %v", err)
	}
	defer adminDB.Close()

	dbName := fmt.Sprintf("test_fresh_%d", time.Now().UnixNano())
	if _, err := adminDB.Exec(fmt.Sprintf(`CREATE DATABASE %q`, dbName)); err != nil {
		t.Skipf("could not create a throwaway database (likely insufficient DB privileges in this environment), skipping: %v", err)
	}
	t.Cleanup(func() {
		_, _ = adminDB.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS %q WITH (FORCE)`, dbName))
	})

	freshDSN, err := replaceDatabaseName(dsn, dbName)
	if err != nil {
		t.Fatalf("could not build DSN for fresh database: %v", err)
	}
	freshDB, err := sqlx.Connect("postgres", freshDSN)
	if err != nil {
		t.Fatalf("connect to fresh database error: %v", err)
	}
	defer freshDB.Close()

	if err := Migrate(freshDB, "testowner", "testpassword123"); err != nil {
		t.Fatalf("Migrate on fresh empty database failed: %v", err)
	}

	var tableCount int
	if err := freshDB.Get(&tableCount, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'`); err != nil {
		t.Fatalf("could not count tables in fresh database: %v", err)
	}
	if tableCount == 0 {
		t.Fatalf("expected Migrate to create tables in the fresh database, found none")
	}
	t.Logf("fresh database %s has %d tables after Migrate", dbName, tableCount)
}

// replaceDatabaseName يستبدل اسم القاعدة بمسار/DSN اتصال Postgres (شكل URL أو
// شكل key=value) باسم قاعدة جديد — يدعم شكل postgresql://user:pass@host:port/dbname?params
func replaceDatabaseName(dsn, newDBName string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}
	u.Path = "/" + newDBName
	return u.String(), nil
}
