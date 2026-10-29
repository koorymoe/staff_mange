package repository

import (
	"errors"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestPrinciplesAcceptance_Live المبادئ العشرين (دازاي 10-29): مزروعة، الموافقة على
// النص بالضبط (بصمة)، تعديل نقطة يطلب موافقة جديدة، وكل موافقة بسجل ما ينمسح.
func TestPrinciplesAcceptance_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	const emp = "__pp_emp"
	cleanup := func() {
		db.Exec(`DELETE FROM "PrivacyPolicyAcceptanceLog" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "PrivacyPolicyAcceptance" WHERE "employeeId" = $1`, emp)
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1, 'تجربة مبادئ', 'ACTIVE', 'TECHNICIAN')`, emp)
	var n int
	db.Get(&n, `SELECT count(*) FROM "PrivacyPolicyPoint" WHERE "principleNo" BETWEEN 1 AND 20 AND "isActive"`)
	if n != 20 {
		t.Fatalf("المبادئ المزروعة %d مو ٢٠", n)
	}
	r := NewPrivacyPolicyRepository(db)
	st, err := r.Status(emp)
	if err != nil || !st.NeedsAcceptance || st.Hash == "" {
		t.Fatalf("الجديد لازم يوافق: %v %+v", err, st)
	}
	if err := r.Accept(emp, "wrong"); !errors.Is(err, ErrPolicyChanged) {
		t.Fatalf("بصمة غلط لازم ترفض: %v", err)
	}
	if err := r.Accept(emp, st.Hash); err != nil {
		t.Fatal(err)
	}
	if st2, _ := r.Status(emp); st2.NeedsAcceptance {
		t.Fatal("بعد الموافقة بعده يطلب")
	}
	// تعديل نص مبدأ ← يطلب موافقة جديدة
	var id, content string
	db.QueryRow(`SELECT id, content FROM "PrivacyPolicyPoint" WHERE "principleNo" = 1`).Scan(&id, &content)
	db.MustExec(`UPDATE "PrivacyPolicyPoint" SET content = content || ' (تجربة)' WHERE id = $1`, id)
	defer db.Exec(`UPDATE "PrivacyPolicyPoint" SET content = $2 WHERE id = $1`, id, content)
	st3, _ := r.Status(emp)
	if !st3.NeedsAcceptance || !st3.Changed {
		t.Fatal("تعديل النص ما طلب موافقة جديدة")
	}
	if err := r.Accept(emp, st3.Hash); err != nil {
		t.Fatal(err)
	}
	h, err := r.History(emp)
	if err != nil || len(h) != 2 || !h[0].Current || h[1].Current {
		t.Fatalf("السجل لازم موافقتين والأحدث الحالية: %v %+v", err, h)
	}
}
