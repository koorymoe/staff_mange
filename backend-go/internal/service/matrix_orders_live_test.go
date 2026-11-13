package service

import (
	"os"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/repository"
)

// أوامر ماتركس (دازاي 11-13): الاقتراح ما ينفّذ شي، «نفّذ» ينفّذ مرة وحدة بس،
// المالك ما ينخصم منه، والمشطوب ما يتنفّذ، والزبون يوصل للموديل برمز.
func TestMatrixOrdersLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("بلا DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	clean := func() {
		db.Exec(`DELETE FROM "MatrixOrder" WHERE "actorId" = 'mo-boss'`)
		db.Exec(`DELETE FROM "KpiEvaluation" WHERE "employeeId" IN ('mo-emp', 'mo-boss2')`)
		db.Exec(`DELETE FROM "Notification" WHERE "employeeId" IN ('mo-emp', 'mo-boss2', 'mo-boss')`)
		db.Exec(`DELETE FROM "ExtraTask" WHERE "assignedToId" = 'mo-emp' OR "createdById" = 'mo-boss'`)
		db.Exec(`DELETE FROM "QuotationItem" WHERE "quotationId" IN (SELECT id FROM "Quotation" WHERE "createdByEmployeeId" = 'mo-boss')`)
		db.Exec(`DELETE FROM "Quotation" WHERE "createdByEmployeeId" = 'mo-boss'`)
		db.Exec(`DELETE FROM "Customer" WHERE id = 'mo-cu'`)
		db.Exec(`DELETE FROM "Employee" WHERE id IN ('mo-emp', 'mo-boss', 'mo-boss2')`)
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name, role) VALUES ('mo-boss', 'mo boss', 'OWNER'), ('mo-boss2', 'mo boss2', 'ADMIN'), ('mo-emp', 'mo emp', 'TECHNICIAN')`)
	db.MustExec(`INSERT INTO "Customer"(id, name, phone) VALUES ('mo-cu', 'شركة الموصلي للتجارة', '07700009911')`)

	notif := repository.NewNotificationRepository(db)
	kpi := NewKpiService(repository.NewKpiRepository(db), repository.NewEmployeeRepository(db), notif, repository.NewAnnouncementRepository(db))
	s := NewMatrixOrderService(db, nil, OrderDeps{Kpi: kpi, Tasks: repository.NewExtraTaskRepository(db), Notif: notif,
		Quotations: NewQuotationService(repository.NewQuotationRepository(db), repository.NewPermissionRepository(db))})

	masked, cm := s.maskCustomers("سويلي عرض سعر لشركة الموصلي للتجارة")
	if strings.Contains(masked, "الموصلي") || !strings.Contains(masked, "زبون#1") || cm.show("زبون#1") != "شركة الموصلي للتجارة" {
		t.Fatalf("اسم الزبون لازم يتخفى: %q", masked)
	}

	s.propose = func(actor, text string) (string, []OrderAction, error) {
		return "فهمت", []OrderAction{
			{Kind: "points", EmployeeID: "mo-emp", EmployeeName: "mo emp", Points: 20, Reason: "تأخر عالحجز"},
			{Kind: "points", EmployeeID: "mo-boss2", EmployeeName: "mo boss2", Points: 1, Reason: "x"}, // المدير ما ينخصم
			{Kind: "task", EmployeeID: "mo-emp", EmployeeName: "mo emp", Title: "رتّب المخزن"},
			{Kind: "notify", EmployeeID: "mo-emp", EmployeeName: "mo emp", Message: "راجع الإدارة"},
			{Kind: "quotation", CustomerName: "شركة الموصلي للتجارة", Items: []OrderItem{{ProductName: "كاميرا", Quantity: 4, UnitPrice: 50000}}},
		}, nil
	}
	o, err := s.Propose("mo-boss", "اخصم نقطة …")
	if err != nil {
		t.Fatal(err)
	}
	if len(o.Actions) != 4 || o.Actions[0].Points != 8 || o.Status != OrderProposed {
		t.Fatalf("الاقتراح: %+v", o)
	}
	var n int
	db.Get(&n, `SELECT count(*) FROM "KpiEvaluation" WHERE "employeeId" = 'mo-emp'`)
	if n != 0 {
		t.Fatal("الاقتراح ما ينفّذ شي")
	}
	done, err := s.Run(o.ID, "mo-boss", "OWNER", []int{3}) // نشطب عرض السعر
	if err != nil {
		t.Fatal(err)
	}
	for i, a := range done.Actions[:3] {
		if !a.Done {
			t.Fatalf("الإجراء %d ما تنفّذ: %+v", i, a)
		}
	}
	if !done.Actions[3].Skip || done.Status != OrderDone {
		t.Fatalf("المشطوب: %+v", done)
	}
	var pts int
	db.Get(&pts, `SELECT points FROM "KpiEvaluation" WHERE "employeeId" = 'mo-emp'`)
	if pts != -8 {
		t.Fatalf("الخصم = %d", pts)
	}
	db.Get(&n, `SELECT count(*) FROM "Quotation" WHERE "createdByEmployeeId" = 'mo-boss'`)
	if n != 0 {
		t.Fatal("المشطوب انفّذ")
	}
	if _, err := s.Run(o.ID, "mo-boss", "OWNER", nil); err == nil {
		t.Fatal("ينفّذ مرة وحدة بس")
	}
	// أمر ثاني: عرض سعر، وغير صاحبه ما يشوفه
	o2, _ := s.Propose("mo-boss", "عرض")
	if _, err := s.Run(o2.ID, "mo-boss2", "ADMIN", nil); err == nil {
		t.Fatal("غير صاحب الأمر ما ينفّذه")
	}
	d2, err := s.Run(o2.ID, "mo-boss", "OWNER", []int{0, 1, 2})
	if err != nil || !d2.Actions[3].Done || !strings.Contains(d2.Actions[3].Result, "عرض سعر") {
		t.Fatalf("عرض السعر: %v %+v", err, d2.Actions[3])
	}
	if err := s.Cancel(o2.ID, "mo-boss"); err == nil {
		t.Fatal("المنفّذ ما ينلغي")
	}
}
