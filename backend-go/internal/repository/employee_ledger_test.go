package repository

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestLedgerSQL كل استعلامات «شنو سوّى الموظف» لازم تشتغل على القاعدة الحقيقية —
// قسم مكسور ينخفي بصمت ويطلع الموظف «ما سوّى شي» وهو سوّى.
func TestLedgerSQL(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Skip("القاعدة ما تشتغل: ", err)
	}
	defer db.Close()
	to := time.Now().UTC()
	from := to.AddDate(0, -1, 0)
	for _, s := range ledgerSections {
		items := []LedgerItem{}
		if err := db.Select(&items, s.sql+` ORDER BY at DESC LIMIT 1`, "x", from, to); err != nil {
			t.Errorf("%s: %v", s.key, err)
		}
	}
}
