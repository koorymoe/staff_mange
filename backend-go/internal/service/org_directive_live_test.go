package service

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
)

// TestDirectives_Live النقطة ٧: التوجيه حسب المقام، يوصل نطاقه بس، الإقرار بنطاقه بس،
// عدّاد المقرين، الإيقاف، يدخل موجز الذكاء، وكمال يصدره.
func TestDirectives_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	const own, head, mem, out = "__dir_own", "__dir_head", "__dir_mem", "__dir_out"
	emps := []string{own, head, mem, out}
	cleanup := func() {
		db.Exec(`DELETE FROM "OrgDirective" WHERE "issuedById" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "OrgAssignment" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "OrgPermissionGrant" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "EmployeePermission" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "PermissionAuditLog" WHERE "employeeId" = ANY($1)`, emps)
		db.Exec(`DELETE FROM "Employee" WHERE id = ANY($1)`, emps)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1,'مالك','ACTIVE','OWNER'),($2,'رئيس وحدة','ACTIVE','TECHNICIAN'),($3,'عضو','ACTIVE','TECHNICIAN'),($4,'برّا','ACTIVE','TECHNICIAN')`, own, head, mem, out)
	s := NewOrgService(db)
	_ = s.Seed()
	for _, a := range []OrgAssignInput{{EmployeeID: head, UnitID: "eng.design", Position: "DIRECTOR"}, {EmployeeID: mem, UnitID: "eng.design", Position: "MEMBER"}} {
		if _, err := s.Assign("", a); err != nil {
			t.Skipf("المكان مشغول محلياً: %v", err)
		}
	}
	if _, _, err := s.IssueDirective(head, "TECHNICIAN", DirectiveInput{Scope: "COMPANY", Title: "xxx", Body: "يييييييييييي"}); !errors.Is(err, ErrOrgHidden) {
		t.Fatal("رئيس وحدة أصدر توجيه للشركة")
	}
	c := 10
	uid, to, err := s.IssueDirective(head, "TECHNICIAN", DirectiveInput{Scope: "UNIT", UnitID: "eng.design", Title: "التصاميم بالمعايير", Body: "كل تصميم يتراجع بالمعايير الهندسية گبل التسليم", Clause: &c})
	if err != nil {
		t.Fatal(err)
	}
	if len(to) != 1 || to[0] != mem {
		t.Fatalf("توجيه الوحدة لازم يوصل عضوها بس: %v", to)
	}
	cid, toAll, err := s.IssueDirective(own, "OWNER", DirectiveInput{Scope: "COMPANY", Title: "الدوام ٨ الصبح", Body: "كل الكادر يبصم گبل ٨ الصبح بالضبط", Clause: intp(3)})
	if err != nil || len(toAll) < 3 {
		t.Fatalf("توجيه الشركة: %v %d", err, len(toAll))
	}
	mine, _ := s.MyDirectives(out)
	for _, d := range mine {
		if d.ID == uid {
			t.Fatal("الي برّا الوحدة شاف توجيهها")
		}
	}
	if err := s.AckDirective(out, uid); !errors.Is(err, ErrOrgHidden) {
		t.Fatal("أقر على توجيه مو بنطاقه")
	}
	if err := s.AckDirective(mem, uid); err != nil {
		t.Fatal(err)
	}
	mine, _ = s.MyDirectives(mem)
	var gotU, gotC bool
	for _, d := range mine {
		if d.ID == uid && d.AckAt != nil && d.ClauseCode != nil && *d.ClauseCode == "م2/ف4" {
			gotU = true
		}
		if d.ID == cid && d.AckAt == nil {
			gotC = true
		}
	}
	if !gotU || !gotC {
		t.Fatalf("توجيهاتي: %+v", mine)
	}
	man, _ := s.ManagedDirectives(head, "TECHNICIAN")
	found := false
	for _, d := range man {
		if d.ID == uid {
			found = true
			if d.Acks != 2 || d.Audience != 2 { // الرئيس (المُصدِر) + العضو
				t.Fatalf("العداد: %d/%d", d.Acks, d.Audience)
			}
		}
		if d.ID == cid {
			t.Fatal("رئيس الوحدة يتابع توجيه الشركة")
		}
	}
	if !found {
		t.Fatal("رئيس الوحدة ما يشوف توجيهه")
	}
	if !strings.Contains(OrgBrief(db, mem), "الدوام ٨ الصبح") {
		t.Fatal("التوجيه ما دخل موجز الذكاء")
	}
	if err := s.ArchiveDirective(head, "TECHNICIAN", uid); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(OrgBrief(db, mem), "التصاميم بالمعايير") {
		t.Fatal("الموقوف بعده بالموجز")
	}
	// كمال يصدر
	SetDirectiveDeps(s, nil)
	a := NewMatrixAgent(db, nil, "")
	registerDirectiveTool(a, &SecretaryTurnCtx{UserID: own})
	for _, tl := range a.tools {
		if tl.Name == "issue_directive" {
			res, err := tl.Run(map[string]any{"scope": "DIVISION", "division": "eng", "title": "تسليم بالوقت", "body": "كل المشاريع تتسلم بموعدها بلا تأخير"}, a.newMask())
			if err != nil || !strings.Contains(res, "انصدر") {
				t.Fatalf("كمال: %v %s", err, res)
			}
		}
	}
}
