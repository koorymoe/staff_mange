package repository

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// (ل) 11-12 دازاي: زبائن ينسجلون بنفس اللحظة — كلهم ينحفظون بأرقام مختلفة (چان «أكبر رقم + ١»
// فاثنين ياخذون نفس الرقم وواحد يطلعله «ما انحفظ»). والشاشة الثانية (من التسلسل) ما تصطدم.
func TestCustomerCodeConcurrent_Live(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	defer db.Exec(`DELETE FROM "Customer" WHERE phone LIKE '0799000%'`) //nolint:errcheck
	db.Exec(`DELETE FROM "Customer" WHERE phone LIKE '0799000%'`)       //nolint:errcheck
	r := NewCustomerRepository(db)
	const n = 20
	var wg sync.WaitGroup
	codes := make([]int, n)
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c, err := r.FindOrCreateByPhone(fmt.Sprintf("0799000%04d", i), "زبون متزامن")
			errs[i] = err
			if c != nil {
				codes[i] = c.CustomerCode
			}
		}(i)
	}
	wg.Wait()
	seen := map[int]bool{}
	for i := 0; i < n; i++ {
		if errs[i] != nil {
			t.Fatalf("زبون %d ما انحفظ: %v", i, errs[i])
		}
		if seen[codes[i]] {
			t.Fatalf("رقم مكرر %d", codes[i])
		}
		seen[codes[i]] = true
	}
	// الشاشة الثانية (افتراضي العمود) بعدها ما تصطدم
	var code int
	if err := db.Get(&code, `INSERT INTO "Customer" (id, name, phone) VALUES (gen_random_uuid()::text, 'ثاني', '07990009999') RETURNING "customerCode"`); err != nil || seen[code] {
		t.Fatalf("الإضافة من التسلسل: %v %d", err, code)
	}
}
