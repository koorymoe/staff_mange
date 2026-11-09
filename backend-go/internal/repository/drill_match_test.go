package repository

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// ═══ حارس «الرقم = القائمة» — (ع) 10-16 ═══
// المدير لگى بطاقات تگول رقم وتفتح شي ثاني. هذا الملف يمسك المشكلة
// يوم تنكتب، مو يوم يشتكي أحد.

// TestBoardCountsUseDrillDefs بنيوي بلا قاعدة: كل رقم بلوحة اليوم ونبض
// الإدارة ينحسب من DrillCountSQL — مو من شرط مكتوب باليد يفترق عن القائمة.
func TestBoardCountsUseDrillDefs(t *testing.T) {
	src, err := os.ReadFile("today_repository.go")
	if err != nil {
		t.Fatal(err)
	}
	body := string(src)
	for _, alias := range []string{"bookingsToday", "inField", "completedToday", "newToday",
		"needsContact", "needsCrew", "needsPaper", "needsFinish", "weekTotal"} {
		re := regexp.MustCompile(`DrillCountSQL\("[a-z_]+"\)\+` + "`" + ` AS "` + alias + `"`)
		if !re.MatchString(body) {
			t.Errorf("لوحة اليوم: الرقم %q ما ينحسب من DrillCountSQL — رقمه يفترق عن قائمته", alias)
		}
	}
	dash, err := os.ReadFile("../handler/dashboard_handler.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"today_scheduled", "open_missions", "needs_coord", "overdue"} {
		if !strings.Contains(string(dash), `DrillCountSQL("`+key+`")`) {
			t.Errorf("نبض الإدارة: %q ما ينحسب من DrillCountSQL", key)
		}
	}
	for _, d := range DrillDefs() {
		if d.Title == "" || d.Explain == "" || d.Where == "" || d.AgeFrom == "" {
			t.Errorf("تعريف %q ناقص (عنوان/شرح/شرط)", d.Key)
		}
	}
}

// TestDrillCountMatchesList حي: لكل رقم، عدّ البطاقة = عدد قائمتها.
func TestDrillCountMatchesList(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Skip("القاعدة ما تشتغل: ", err)
	}
	defer db.Close()
	for _, d := range DrillDefs() {
		var card int
		if err := db.Get(&card, `SELECT `+DrillCountSQL(d.Key)+` FROM "Booking" b WHERE `+drillBase()); err != nil {
			t.Fatalf("%s: عدّ البطاقة: %v", d.Key, err)
		}
		res, err := Drill(db, d.Key, 100000)
		if err != nil {
			t.Fatalf("%s: القائمة: %v", d.Key, err)
		}
		if res.Count != card || len(res.Rows) != card {
			t.Errorf("%s: البطاقة %d بس القائمة %d (صفوف %d)", d.Key, card, res.Count, len(res.Rows))
		}
	}
}
