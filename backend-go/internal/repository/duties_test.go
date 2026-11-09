package repository

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestDutiesSQL كل واجب لكل منصب لازم يشتغل على القاعدة — واجب مكسور
// يطلع «ما سوّى» على موظف سوّى.
func TestDutiesSQL(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Skip("القاعدة ما تشتغل: ", err)
	}
	defer db.Close()
	r := NewDutiesRepository(db)
	for pos, defs := range dutiesByPosition {
		for _, d := range defs {
			var n int
			if err := db.Get(&n, r.fill(d.done), "x"); err != nil {
				t.Errorf("%s/%s done: %v", pos, d.key, err)
			}
			if d.pending != "" && d.pending != "once" {
				q := r.fill(d.pending)
				var err error
				if strings.Contains(q, "$1") {
					err = db.Get(&n, q, "x")
				} else {
					err = db.Get(&n, q)
				}
				if err != nil {
					t.Errorf("%s/%s pending: %v", pos, d.key, err)
				}
			}
		}
	}
}
