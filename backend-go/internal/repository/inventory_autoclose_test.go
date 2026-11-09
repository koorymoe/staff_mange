package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// (ع) 11-04: النقص القديم ينسكّر لحاله بس إذا آخر جرد أحدث أثبت وجود أدواته.
func TestAutoCloseShortages_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	emps := []string{"ac-emp-a", "ac-emp-b", "ac-emp-c"}
	clean := func() {
		for _, e := range emps {
			db.Exec(`DELETE FROM "InventoryCheck" WHERE "employeeId" = $1`, e)
			db.Exec(`DELETE FROM "Employee" WHERE id = $1`, e)
		}
	}
	clean()
	defer clean()
	for _, e := range emps {
		if _, err := db.Exec(`INSERT INTO "Employee"(id, name) VALUES ($1, $1)`, e); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	ins := func(id, emp string, complete bool, missing *string, daysAgo int) {
		if _, err := db.Exec(`INSERT INTO "InventoryCheck"(id, "employeeId", complete, "missingItems", "checkedAt")
			VALUES ($1, $2, $3, $4, now() - make_interval(days => $5))`, id, emp, complete, missing, daysAgo); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}
	p := func(s string) *string { return &s }
	// A: نقص قديم وبعده جرد مكتمل ← ينسكّر
	ins("ac-a1", "ac-emp-a", false, p("دريل، سلم"), 60)
	ins("ac-a2", "ac-emp-a", true, nil, 1)
	// B: نقص قديم، وآخر جرد ناقص بس بأداة ثانية ← ينسكّر؛ وآخر نقص (الأحدث) يبقى
	ins("ac-b1", "ac-emp-b", false, p("دريل"), 60)
	ins("ac-b2", "ac-emp-b", false, p("سلم"), 1)
	// C: آخر جرد بعده ناقص نفس الأداة ← يبقى مفتوح
	ins("ac-c1", "ac-emp-c", false, p("دريل، سلم"), 60)
	ins("ac-c2", "ac-emp-c", false, p("سلم"), 1)

	for _, e := range emps {
		if _, err := AutoCloseShortages(db, e); err != nil {
			t.Fatalf("autoclose: %v", err)
		}
	}
	want := map[string]bool{"ac-a1": true, "ac-b1": true, "ac-b2": false, "ac-c1": false, "ac-c2": false}
	for id, w := range want {
		var got bool
		var by *string
		if err := db.QueryRow(`SELECT resolved, "resolvedById" FROM "InventoryCheck" WHERE id = $1`, id).Scan(&got, &by); err != nil {
			t.Fatalf("read %s: %v", id, err)
		}
		if got != w {
			t.Errorf("%s: resolved=%v نريد %v", id, got, w)
		}
		if by != nil {
			t.Errorf("%s: الإغلاق التلقائي لازم يخلي resolvedById فاضي", id)
		}
	}
}
