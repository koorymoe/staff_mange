package integrity

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// كل فحص لازم: مفتاح فريد، خطورة معروفة، وما يكتب (READ ONLY يرفض الكتابة أصلاً).
func TestChecks_Static(t *testing.T) {
	seen := map[string]bool{}
	sev := map[string]bool{"critical": true, "high": true, "medium": true, "low": true}
	for _, c := range Checks {
		if seen[c.Key] {
			t.Errorf("مفتاح مكرر: %s", c.Key)
		}
		seen[c.Key] = true
		if !sev[c.Severity] || c.Title == "" || c.Count == "" || c.Sample == "" {
			t.Errorf("فحص ناقص: %s", c.Key)
		}
		for _, w := range []string{"INSERT ", "UPDATE ", "DELETE ", "DROP ", "ALTER ", "TRUNCATE "} {
			if strings.Contains(strings.ToUpper(c.Count+" "+c.Sample), w) {
				t.Errorf("%s بيه %s", c.Key, w)
			}
		}
	}
}

// يحتاج قاعدة حية: TEST_DATABASE_URL — كل فحص وأمثلته تشتغل بلا خطأ.
func TestChecks_Live(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	for _, r := range RunAll(db) {
		if r.Error != "" {
			t.Errorf("%s: %s", r.Key, r.Error)
			continue
		}
		t.Logf("%-50s %d", r.Key, r.Count)
		if _, _, err := Sample(db, r.Key); err != nil {
			t.Errorf("%s sample: %v", r.Key, err)
		}
	}
}
