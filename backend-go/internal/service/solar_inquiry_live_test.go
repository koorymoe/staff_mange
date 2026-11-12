package service

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/repository"
)

// (ل) 11-12 دازاي: استفسار طاقة شمسية — المسجّل يسجّل، التقني ياخذه ويسجّل اتصاله، ابو الجودة
// يتصل ويتأكد. التقني كاتب «اتصلت» والزبون يگول لا = مفارقة + تنبيه. الغريب = 404.
func TestSolarInquiryFlow_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ids := []string{"si-in", "si-tech", "si-tech2", "si-q", "si-out"}
	clean := func() {
		db.Exec(`DELETE FROM "SolarInquiry" WHERE "createdById" = 'si-in'`)
		for _, id := range ids {
			db.Exec(`DELETE FROM "Notification" WHERE "employeeId" = $1`, id)
			db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = $1`, id)
			db.Exec(`DELETE FROM "Employee" WHERE id = $1`, id)
		}
	}
	clean()
	defer clean()
	for _, id := range ids {
		db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ($1, $1, 'SALES')`, id)
	}
	perm := func(emp, name string) {
		db.MustExec(`INSERT INTO "Permission"(id, name, label) VALUES (gen_random_uuid()::text, $1, $1) ON CONFLICT (name) DO NOTHING`, name)
		db.MustExec(`INSERT INTO "EmployeePermission"(id, "employeeId", "permissionId") SELECT gen_random_uuid()::text, $1, id FROM "Permission" WHERE name = $2`, emp, name)
	}
	perm("si-in", "solar_inquiry")
	perm("si-tech", "solar_inquiry_tech")
	perm("si-tech2", "solar_inquiry_tech")
	perm("si-q", "quality_control")
	s := NewSolarInquiryService(db, repository.NewNotificationRepository(db))

	if _, err := s.Create("si-out", "SALES", "زبون", "07701234567", "", ""); err != ErrInquiryHidden {
		t.Fatalf("الي ماعنده صلاحية = 404: %v", err)
	}
	if _, err := s.Create("si-in", "SALES", "زبون", "077", "", ""); err == nil {
		t.Fatal("رقم ناقص لازم ينرفض")
	}
	x, err := s.Create("si-in", "SALES", "أبو علي", "0770 123 4567", "كربلاء", "يسأل عن منظومة ٥ كيلو")
	if err != nil || x.Stage != "NEW" {
		t.Fatalf("التسجيل: %v %+v", err, x)
	}
	var n int
	db.Get(&n, `SELECT count(*) FROM "Notification" WHERE "employeeId" IN ('si-tech','si-tech2') AND type = 'SOLAR_INQUIRY'`)
	if n != 2 {
		t.Fatalf("التقنيين يوصلهم تنبيه: %d", n)
	}
	if _, err := s.Claim("si-tech", "SALES", x.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Claim("si-tech2", "SALES", x.ID); err == nil {
		t.Fatal("تقني ثاني ما ياخذه")
	}
	if _, err := s.TechContact("si-tech2", "SALES", x.ID, "JUST_ASKING", "حچيت وياه", nil, nil); err != ErrInquiryHidden {
		t.Fatalf("تقني ثاني ما يسجّل عليه: %v", err)
	}
	yes, amt := true, 4500000.0
	x, err = s.TechContact("si-tech", "SALES", x.ID, "WANTS_INSTALL", "يريد منظومة ٥ كيلو، عطيته سعر", &yes, &amt)
	if err != nil || x.Stage != "TECH_DONE" {
		t.Fatalf("اتصال التقني: %v %+v", err, x)
	}
	db.Get(&n, `SELECT count(*) FROM "Notification" WHERE "employeeId" = 'si-q' AND type = 'SOLAR_INQUIRY'`)
	if n != 1 {
		t.Fatalf("ابو الجودة يوصله تنبيه: %d", n)
	}
	if _, err := s.QualityCheck("si-tech", "SALES", x.ID, QualityInput{Reached: true}); err != ErrInquiryHidden {
		t.Fatal("التقني ما يسجّل جودة")
	}
	x, err = s.QualityCheck("si-q", "SALES", x.ID, QualityInput{Reached: true, TechCalled: "NO", Intent: "WANTS_INSTALL", GotQuote: "NO", QuoteFair: "UNSURE", Note: "يگول محد اتصل"})
	if err != nil || x.Stage != "CLOSED" || !x.Mismatch {
		t.Fatalf("الجودة والمفارقة: %v %+v", err, x)
	}
	rows, acc, err := s.List("si-in", "SALES")
	if err != nil || !acc.Intake || acc.Tech || len(rows) != 1 {
		t.Fatalf("المسجّل يشوف الي سجّلها بس: %v %+v %d", err, acc, len(rows))
	}
	if _, _, err := s.List("si-out", "SALES"); err != ErrInquiryHidden {
		t.Fatal("الغريب ما يشوف")
	}
}
