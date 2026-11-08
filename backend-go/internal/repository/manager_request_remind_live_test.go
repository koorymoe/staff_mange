package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// (ع) 11-08: الطلب المفتوح +٤٨ ساعة يذكّر مرة وحدة (لحد ٤٨ ساعة ثانية)، والجديد والمحسوم لا.
func TestManagerRequestRemind_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	clean := func() {
		db.Exec(`DELETE FROM "ManagerRequest" WHERE "requesterId" = 'mr-emp'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = 'mr-emp'`)
	}
	clean()
	defer clean()
	db.MustExec(`INSERT INTO "Employee"(id, name) VALUES ('mr-emp','طالب تجربة')`)
	ins := func(title, status, ago string) {
		db.MustExec(`INSERT INTO "ManagerRequest"(id, "requesterId", title, status, "createdAt") VALUES (gen_random_uuid()::text, 'mr-emp', $1, $2, now() - $3::interval)`, title, status, ago)
	}
	ins("قديم مفتوح", "OPEN", "3 days")
	ins("جديد", "OPEN", "5 hours")
	ins("قديم محسوم", "DONE", "5 days")
	r := NewManagerRequestRepository(db, nil)
	if n, err := r.RemindStale(); err != nil || n < 1 {
		t.Fatalf("لازم يذكّر بالقديم المفتوح: %d %v", n, err)
	}
	var got []string
	db.Select(&got, `SELECT title FROM "ManagerRequest" WHERE "requesterId" = 'mr-emp' AND "remindedAt" IS NOT NULL`) //nolint:errcheck
	if len(got) != 1 || got[0] != "قديم مفتوح" {
		t.Fatalf("التذكير غلط: %v", got)
	}
	if n, _ := r.RemindStale(); n != 0 {
		var mine int
		db.Get(&mine, `SELECT count(*) FROM "ManagerRequest" WHERE "requesterId" = 'mr-emp' AND "remindedAt" > now() - interval '1 minute'`) //nolint:errcheck
		if mine != 1 {
			t.Fatal("ما لازم يذكّر مرتين بنفس اليوم")
		}
	}
}
