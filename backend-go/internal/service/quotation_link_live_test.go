package service

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/repository"
)

// بلاغ #33 (مؤمل 11-12): عرض انسوّى قبل المشروع يرتبط بيه بعدين. نفس الزبون يطلع أول،
// عروض غيري (edit_own) ما تطلع، الغريب 404، والعرض/المشروع المربوط ما ينربط مرة ثانية.
func TestQuotationLinkToProject_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	emps := []string{"ql-own", "ql-other", "ql-out"}
	clean := func() {
		db.Exec(`DELETE FROM "Quotation" WHERE id LIKE 'ql-%'`)
		db.Exec(`DELETE FROM "Project" WHERE id LIKE 'ql-%'`)
		for _, id := range emps {
			db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = $1`, id)
			db.Exec(`DELETE FROM "Employee" WHERE id = $1`, id)
		}
	}
	clean()
	defer clean()
	for _, id := range emps {
		db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ($1, $1, 'SALES')`, id)
	}
	for _, id := range []string{"ql-own", "ql-other"} {
		db.MustExec(`INSERT INTO "Permission"(id, name, label) VALUES (gen_random_uuid()::text, 'quotation_edit_own', 'x') ON CONFLICT (name) DO NOTHING`)
		db.MustExec(`INSERT INTO "EmployeePermission"(id, "employeeId", "permissionId") SELECT gen_random_uuid()::text, $1, id FROM "Permission" WHERE name = 'quotation_edit_own'`, id)
	}
	db.MustExec(`INSERT INTO "Project"(id, code, name, phone) VALUES ('ql-p1','QL-1','احمد محمد','0781 117 7720'), ('ql-p2','QL-2','زبون ثاني',NULL)`)
	q := func(id, num, name, phone, by string) {
		db.MustExec(`INSERT INTO "Quotation"(id, "quotationNumber", "customerName", "customerPhone", "createdByEmployeeId", "grandTotal", "netTotal")
			VALUES ($1,$2,$3,NULLIF($4,''),$5,100,100)`, id, num, name, phone, by)
	}
	q("ql-a", "QL-Q-1", "ابو احمد", "+964 781 117 7720", "ql-own") // نفس الرقم
	q("ql-b", "QL-Q-2", "شخص ثاني", "", "ql-own")                    // بالبحث بس
	q("ql-c", "QL-Q-3", "احمد محمد", "", "ql-other")                 // عرض غيري
	s := NewQuotationService(repository.NewQuotationRepository(db), repository.NewPermissionRepository(db))

	if _, err := s.LinkCandidates("ql-p1", "", "ql-out", "SALES"); err != ErrQuotationHidden {
		t.Fatalf("الي ماعنده صلاحية عروض = 404: %v", err)
	}
	rows, err := s.LinkCandidates("ql-p1", "", "ql-own", "SALES")
	if err != nil || len(rows) != 1 || rows[0].ID != "ql-a" || !rows[0].Match {
		t.Fatalf("نفس الزبون بالرقم، وعروض غيري ما تطلع: %v %+v", err, rows)
	}
	rows, _ = s.LinkCandidates("ql-p1", "QL-Q-2", "ql-own", "SALES")
	if len(rows) != 2 || rows[0].ID != "ql-a" {
		t.Fatalf("البحث برقم العرض يضيف، والمطابق يبقى أول: %+v", rows)
	}
	rows, _ = s.LinkCandidates("ql-p1", "", "boss", "OWNER")
	if len(rows) != 2 {
		t.Fatalf("المالك يشوف المطابق بالرقم والاسم: %+v", rows)
	}
	if _, err := s.LinkToProject("ql-p1", "ql-c", "ql-own", "SALES"); err != ErrQuotationHidden {
		t.Fatalf("عرض غيري ما أربطه: %v", err)
	}
	if _, err := s.LinkToProject("ql-nope", "ql-a", "ql-own", "SALES"); err != ErrQuotationHidden {
		t.Fatalf("مشروع مو موجود = 404: %v", err)
	}
	x, err := s.LinkToProject("ql-p1", "ql-a", "ql-own", "SALES")
	if err != nil || x.ProjectID == nil || *x.ProjectID != "ql-p1" {
		t.Fatalf("الربط: %v %+v", err, x)
	}
	if got, _ := s.ByProject("ql-p1", "ql-own", "SALES"); got == nil || got.ID != "ql-a" {
		t.Fatal("المشروع صار يشوف عرضه")
	}
	if _, err := s.LinkToProject("ql-p1", "ql-b", "ql-own", "SALES"); err != ErrLinkTaken {
		t.Fatalf("المشروع عنده عرض: %v", err)
	}
	if _, err := s.LinkToProject("ql-p2", "ql-a", "ql-own", "SALES"); err != ErrLinkTaken {
		t.Fatalf("العرض مربوط بمشروع ثاني: %v", err)
	}
}
