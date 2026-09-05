package database

import (
	"github.com/jmoiron/sqlx"
)

// dropAttendanceDailyUniqueConstraint يشيل قيد UNIQUE(employeeId, date) عن جدول
// "Attendance" — هذا القيد كان يمنع الموظف من تسجيل أكثر من جلسة حضور
// (دخول/خروج) بنفس اليوم. نلگي اسم القيد الحقيقي من information_schema بدل
// افتراض اسمه (idempotent — إذا القيد مو موجود ما تصير أي عملية).
func dropAttendanceDailyUniqueConstraint(db *sqlx.DB) error {
	_, err := db.Exec(`
		DO $$
		DECLARE
			constraint_name text;
			index_name text;
		BEGIN
			-- Case 1: a real UNIQUE table constraint on (employeeId, date).
			SELECT tc.constraint_name INTO constraint_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			WHERE tc.table_name = 'Attendance'
				AND tc.constraint_type = 'UNIQUE'
				AND tc.table_schema = 'public'
			GROUP BY tc.constraint_name
			HAVING array_agg(kcu.column_name::text ORDER BY kcu.column_name) = ARRAY['date', 'employeeId']::text[]
			LIMIT 1;

			IF constraint_name IS NOT NULL THEN
				EXECUTE format('ALTER TABLE "Attendance" DROP CONSTRAINT %I', constraint_name);
			END IF;

			-- Case 2: a bare UNIQUE index on (employeeId, date) not backed by a
			-- constraint. information_schema.table_constraints only reports real
			-- constraints, so on some schema histories (this one included) the
			-- old daily-uniqueness rule survives only as a plain unique index —
			-- Case 1 alone misses it, leaving check-in silently broken.
			SELECT i.relname INTO index_name
			FROM pg_index ix
			JOIN pg_class t ON t.oid = ix.indrelid
			JOIN pg_class i ON i.oid = ix.indexrelid
			JOIN pg_namespace n ON n.oid = t.relnamespace
			WHERE t.relname = 'Attendance'
				AND n.nspname = 'public'
				AND ix.indisunique
				AND NOT EXISTS (
					SELECT 1 FROM pg_constraint c WHERE c.conindid = ix.indexrelid
				)
				AND (
					SELECT array_agg(a.attname::text ORDER BY a.attname)
					FROM unnest(ix.indkey) WITH ORDINALITY AS k(attnum, ord)
					JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = k.attnum
				) = ARRAY['date', 'employeeId']::text[]
			LIMIT 1;

			IF index_name IS NOT NULL THEN
				EXECUTE format('DROP INDEX IF EXISTS %I', index_name);
			END IF;
		END $$;
	`)
	return err
}

// seedKpiCriteria يزرع نقاط الكي بي اي الثمانية الأصلية مرة وحدة (idempotent) —
// بعدها تصير قابلة للإضافة والحذف من واجهة إدارة النقاط (صلاحية منفصلة).
func seedKpiCriteria(db *sqlx.DB) error {
	criteria := []string{
		"العلاقة مع الزملاء",
		"تنفيذ المهام الموكلة إليه",
		"استطيع ولا استطيع",
		"الالتزام بالآليات وتوجيهات المسؤول",
		"تنظيف السيارة",
		"سرعة الاستجابة بالمهام",
		"ترتيب العدد",
		"شكوى الزبائن",
	}
	for _, label := range criteria {
		if _, err := db.Exec(`
			INSERT INTO "KpiCriterion" (id, label)
			VALUES (gen_random_uuid()::text, $1)
			ON CONFLICT (label) DO NOTHING
		`, label); err != nil {
			return err
		}
	}
	return nil
}
