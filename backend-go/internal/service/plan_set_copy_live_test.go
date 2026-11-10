package service

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
)

// (ل) 11-10: النسخ من نموذج جاهز لسيت مشروع — قسم كامل أو مخطط، بلا تكرار، والنموذج ما يتغيّر.
func TestPlanSetCopyLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	defer db.Close()
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	tpl, dst, emp := uuid.NewString(), uuid.NewString(), uuid.NewString()
	db.MustExec(`INSERT INTO "Employee" (id, name, username, role, status) VALUES ($1, 'ناسخ', $2, 'OWNER', 'ACTIVE')`, emp, "cp_"+emp[:6])
	t.Cleanup(func() { db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp) }) //nolint:errcheck
	db.MustExec(`INSERT INTO "PlanSet" (id, title, "isTemplate") VALUES ($1, 'نموذج تجربة', true), ($2, 'سيت مشروع', false)`, tpl, dst)
	t.Cleanup(func() { db.Exec(`DELETE FROM "PlanSet" WHERE id IN ($1, $2)`, tpl, dst) }) //nolint:errcheck
	var cam1, solar string
	db.Get(&cam1, `INSERT INTO "PlanFile" ("setId","typeCode",title,"fileKey") VALUES ($1,'5.1','كاميرات','/api/files/designs/c1') RETURNING id`, tpl) //nolint:errcheck
	db.MustExec(`INSERT INTO "PlanFile" ("setId","typeCode",title,"fileKey") VALUES ($1,'5.4','هاتف','/api/files/designs/c2')`, tpl)
	db.Get(&solar, `INSERT INTO "PlanFile" ("setId","typeCode",title,"fileKey") VALUES ($1,'3.1','إنارة','/api/files/designs/e1') RETURNING id`, tpl) //nolint:errcheck
	db.MustExec(`INSERT INTO "PlanFile" ("setId","typeCode",title,"fileKey") VALUES ($1,'10.1','مفهوم','/api/files/designs/p1')`, tpl)
	s := NewPlanSetService(db, NewDesignSetService(db, nil, nil))
	n, err := s.Copy(emp, "OWNER", nil, dst, PlanCopyInput{FromSetID: tpl, Sections: []int{5}, FileIDs: []string{solar}})
	if err != nil || n != 3 {
		t.Fatalf("النسخ: %d %v (المتوقع ٣: قسم ٥ كامل + مخطط ٣.١)", n, err)
	}
	if n, _ = s.Copy(emp, "OWNER", nil, dst, PlanCopyInput{FromSetID: tpl, Sections: []int{5}}); n != 0 {
		t.Fatalf("نسخ مرتين كرّر %d", n)
	}
	var in10 int
	db.Get(&in10, `SELECT count(*) FROM "PlanFile" WHERE "setId" = $1 AND "typeCode" LIKE '1%'`, dst) //nolint:errcheck
	if in10 != 0 {
		t.Fatal("قسم ١٠ انسخ وهو مو مطلوب (5. لازم ما يطابق 10.)")
	}
	var tplCount int
	db.Get(&tplCount, `SELECT count(*) FROM "PlanFile" WHERE "setId" = $1`, tpl) //nolint:errcheck
	if tplCount != 4 {
		t.Fatal("النموذج تغيّر")
	}
	if _, err := s.Copy(emp, "TECHNICIAN", nil, dst, PlanCopyInput{FromSetID: tpl, Sections: []int{5}}); err != ErrDesignHidden {
		t.Fatal("الفني نسخ")
	}
}
