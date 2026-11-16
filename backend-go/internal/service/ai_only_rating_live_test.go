package service

import (
	"errors"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/repository"
)

// (دازاي 11-16) «ما أتقيّم من قبل البشر — بس ماتركس»: التقييم البشري ينرفض،
// ما ينطلب من أحد، والموظف يطلع من قائمة التقييم الدوري.
func TestAIOnlyRating_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	db.MustExec(`ALTER TABLE "Employee" ADD COLUMN IF NOT EXISTS "aiOnlyRating" boolean DEFAULT false NOT NULL`)

	const emp, rater = "ai-only-emp", "ai-only-rater"
	cleanup := func() {
		db.Exec(`DELETE FROM "StaffRating" WHERE "rateeId" = $1`, emp)
		db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2)`, emp, rater)
	}
	cleanup()
	t.Cleanup(cleanup)
	db.MustExec(`INSERT INTO "Employee"(id, name, "aiOnlyRating") VALUES ($1, 'مبرمج اختبار', true), ($2, 'مراقب اختبار', false)`, emp, rater)

	repo := repository.NewMatrixScoreRepository(db)
	err = repo.SaveRating(repository.StaffRatingInput{RaterID: rater, RateeID: emp, Stage: "MONITOR_DAILY", Period: "2026-11-16", Score: 3})
	if !errors.Is(err, repository.ErrAIOnlyRatee) {
		t.Fatalf("التقييم البشري لازم ينرفض، طلع: %v", err)
	}

	if periodicAllowed("MONITOR_PERIODIC", repository.Scorable{ID: emp, AIOnly: true, OffSystem: true}) {
		t.Fatal("ما لازم يطلع بالتقييم الدوري")
	}

	d := NewRatingDuty(db)
	got := d.dropAIOnly([]DutyItem{{Stage: "MONITOR_DAILY", RateeID: emp}, {Stage: "MONITOR_DAILY", RateeID: "someone-else"}})
	if len(got) != 1 || got[0].RateeID != "someone-else" {
		t.Fatalf("ما لازم ينطلب تقييمه: %+v", got)
	}
}
