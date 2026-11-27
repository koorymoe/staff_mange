package service

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

func TestSolarBandTiers(t *testing.T) {
	bands := []SolarBand{{Kw: 3, Economy: 3e6, Mid: 4e6, High: 5e6}, {Kw: 5, Economy: 5e6, Mid: 6e6, High: 8e6}, {Kw: 10, Economy: 9e6, Mid: 11e6, High: 14e6}}
	if bandTiers(nil, 1, 2) != nil {
		t.Fatal("فاضي لازم nil")
	}
	tr := bandTiers(bands, 3, 5)
	if len(tr) != 3 || tr[0].Min != 3e6 || tr[0].Max != 5e6 || tr[2].Max != 8e6 {
		t.Fatalf("3-5: %+v", tr)
	}
	if tr := bandTiers(bands, 6, 7); tr[1].Min != 11e6 { // أقرب حجم أكبر = 10
		t.Fatalf("6-7: %+v", tr)
	}
	if tr := bandTiers(bands, 20, 25); tr[0].Min != 9e6 { // أكبر من الكل = الأكبر
		t.Fatalf("20: %+v", tr)
	}
}

func TestSolarConfigLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	defer db.Close()
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	defer db.Exec(`DELETE FROM "SolarIntakeConfig"`)
	s := NewSolarInquiryService(db, nil)
	if _, err := s.SaveConfig("", "TECHNICIAN", SolarIntakeConfig{}); err != ErrInquiryHidden {
		t.Fatalf("الفني ما يعدّل: %v", err)
	}
	if _, err := s.SaveConfig("", "ADMIN", SolarIntakeConfig{Appliances: []SolarAppliance{{Name: "ثلاجة", MinKw: 2, MaxKw: 1}}}); err == nil {
		t.Fatal("الأقل أكبر من الأكثر لازم يترفض")
	}
	c, err := s.SaveConfig("", "ADMIN", SolarIntakeConfig{
		Appliances: []SolarAppliance{{Name: " ثلاجة ", MinKw: 0.3, MaxKw: 0.6}},
		Bands:      []SolarBand{{Kw: 5, Economy: 5e6, Mid: 6e6, High: 8e6}, {Kw: 3, Economy: 3e6}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.Appliances[0].Name != "ثلاجة" || c.Bands[0].Kw != 3 {
		t.Fatalf("%+v", c)
	}
	tr, err := s.PriceRanges("", "ADMIN", 4, 5)
	if err != nil || len(tr) != 3 || tr[0].Min != 5e6 {
		t.Fatalf("الأسعار من الجدول: %+v %v", tr, err)
	}
}
