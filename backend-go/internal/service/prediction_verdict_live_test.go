package service

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/repository"
)

// (ل) 11-12 دازاي: «خليها تنحسب من دقة التوقعات» — بنهاية الدوام كل توقع «ما راح يخلّص»
// ينحكم: بقى عليه شغل = صح، خلّص = غلط. والثقة = الصح ÷ المحكوم بآخر ٣٠ يوم.
func TestPredictionVerdict_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	clean := func() {
		db.Exec(`DELETE FROM "MatrixProposal" WHERE signature LIKE 'PRED|pv-%'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = 'pv-mon'`)
	}
	clean()
	defer clean()
	db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ('pv-mon', 'pv-mon', 'MONITOR')`)
	day := time.Now().In(debriefLoc).Format("2006-01-02")
	add := func(id, sig, ev string) {
		db.MustExec(`INSERT INTO "MatrixProposal"(id, kind, title, evidence, payload, signature, status)
			VALUES ($1, 'PREDICTION', $1, $2::jsonb, '{"employeeId":"pv-mon"}'::jsonb, $3, 'EXPIRED')`, id, ev, sig)
	}
	add("pv-1", "PRED|pv-mon|NOPE|"+day, `{"key":"NOPE","share":1}`)       // شغلة ما باقي بيها شي = غلط
	add("pv-2", "PRED|pv-old|X|"+day, `{"done":1}`)                        // توقع قديم بلا مفتاح = ما ينحكم
	add("pv-3", "PRED|pv-mon|NOPE|2000-01-01", `{"key":"NOPE","share":1}`) // يوم ثاني = مو هسه
	props := repository.NewMatrixProposalRepository(db)
	watch := NewMatrixAutopilotService(repository.NewAiActionRepository(db), repository.NewAiRepository(db),
		repository.NewNotificationRepository(db), repository.NewSystemSwitchRepository(db), nil, nil)
	s := NewMatrixLearningService(props, nil, repository.NewAiActionRepository(db), repository.NewAiRepository(db), repository.NewNotificationRepository(db), watch)
	base, _ := s.PredictionAccuracy() // بيانات قديمة بالقاعدة — نقارن بالفرق
	c, w := s.VerifyPredictions(day)
	if c != 0 || w != 1 {
		t.Fatalf("خلّص الشغلة = التوقع غلط (0 صح، 1 غلط): %d %d", c, w)
	}
	var v *string
	db.Get(&v, `SELECT verdict FROM "MatrixProposal" WHERE id = 'pv-2'`)
	if v != nil {
		t.Fatal("التوقع بلا مفتاح ما ينحكم")
	}
	db.Get(&v, `SELECT verdict FROM "MatrixProposal" WHERE id = 'pv-3'`)
	if v != nil {
		t.Fatal("توقع يوم ثاني ما ينحكم هسه")
	}
	if c2, w2 := s.VerifyPredictions(day); c2+w2 != 0 {
		t.Fatal("ما ينحكم مرتين")
	}
	props.SetPredictionVerdict("pv-2", true, 3)
	acc, err := s.PredictionAccuracy()
	if err != nil || acc["correct"].(int)-base["correct"].(int) != 1 || acc["judged"].(int)-base["judged"].(int) != 2 || acc["pct"] == nil {
		t.Fatalf("الدقة = الصح ÷ المحكوم (+1 صح من +2): %v %+v %+v", err, base, acc)
	}
	if predShare(true, 3) != 3 || predShare(false, 3) != 1 || predShare(true, 1) != 1 {
		t.Fatal("الحصة")
	}
}
