package service

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// (ل) 11-10: «١٠٠٠ صورة للطاقة الشمسية — كل عرض يستعمل وحدة».
func TestWorkTypeImagePickRotatesLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	sfx := uuid.NewString()[:6]
	wt := uuid.NewString()
	db.MustExec(`INSERT INTO "ProjectWorkType" (id, name) VALUES ($1, $2)`, wt, "طاقة شمسية "+sfx)
	t.Cleanup(func() { db.Exec(`DELETE FROM "ProjectWorkType" WHERE id = $1`, wt) }) //nolint:errcheck
	s := NewWorkTypeImageService(db)
	if n, err := s.Add("", wt, []string{"/api/files/q/s1.png", "/api/files/q/s2.png", "/api/files/q/s3.png", "/api/files/q/s1.png"}); err != nil || n != 3 {
		t.Fatalf("الإضافة: %d %v (المكرر ما ينضاف)", n, err)
	}
	if _, err := s.Add("", wt, []string{"https://evil/x.png"}); err == nil {
		t.Fatal("صورة من برا النظام انقبلت")
	}
	quotes := []string{}
	for i := 0; i < 3; i++ {
		q := uuid.NewString()
		db.MustExec(`INSERT INTO "Quotation" (id, "quotationNumber", "customerName", "createdByEmployeeId") VALUES ($1, $2, 'زبون', (SELECT id FROM "Employee" LIMIT 1))`, q, "QT-"+sfx+string(rune('a'+i)))
		quotes = append(quotes, q)
	}
	t.Cleanup(func() {
		for _, q := range quotes {
			db.Exec(`DELETE FROM "Quotation" WHERE id = $1`, q) //nolint:errcheck
		}
	})
	chapter := "منظومة طاقة شمسية " + sfx // الفصل بيه اسم النوع
	seen := map[string]bool{}
	for _, q := range quotes {
		m, err := s.Pick(q, []string{chapter, "شي ماله نوع"}, false)
		if err != nil || m[chapter] == "" {
			t.Fatalf("ما انختارت صورة: %v %v", m, err)
		}
		if _, ok := m["شي ماله نوع"]; ok {
			t.Fatal("فصل بلا نوع عمل انطاه صورة")
		}
		seen[m[chapter]] = true
	}
	if len(seen) != 3 {
		t.Fatalf("٣ عروض لازم ياخذن ٣ صور مختلفة، طلع %d", len(seen))
	}
	first, _ := s.Pick(quotes[0], []string{chapter}, false)
	again, _ := s.Pick(quotes[0], []string{chapter}, false)
	if first[chapter] != again[chapter] {
		t.Fatal("نفس العرض لازم يبقى على نفس الصورة")
	}
	other, _ := s.Pick(quotes[0], []string{chapter}, true)
	if other[chapter] == first[chapter] {
		t.Fatal("«صورة غيرها» رجّعت نفس الصورة")
	}
}
