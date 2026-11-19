package repository

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// (دازاي 11-19) بياع غادر الشركة (موقوف) وما إله حجوزات بالشهر ما يطلع بالإحصائيات.
func TestSalesStatsHideInactiveLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	cleanup := func() { db.Exec(`DELETE FROM "Employee" WHERE id IN ('st_gone', 'st_here')`) }
	cleanup()
	t.Cleanup(func() { cleanup(); db.Close() })
	db.MustExec(`INSERT INTO "Employee" (id, name, username, password, role, status) VALUES
		('st_gone', 'غادر', 'st_gone', 'x', 'SALES', 'INACTIVE'), ('st_here', 'باقي', 'st_here', 'x', 'SALES', 'ACTIVE')`)
	now := time.Now()
	rows, err := NewStatsRepository(db).SalesStats(now.Truncate(24*time.Hour), now.AddDate(0, 0, -now.Day()+1))
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, r := range rows {
		seen[r.EmployeeID] = true
	}
	if seen["st_gone"] || !seen["st_here"] {
		t.Fatalf("seen = %v", seen)
	}
}
