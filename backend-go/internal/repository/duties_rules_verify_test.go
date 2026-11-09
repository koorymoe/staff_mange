package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestDutiesRules_Live قرارات (ع) 10-22 على الواجبات: المالك ما ينقيّم، الطابور
// المشترك يحاسب على المتأخر ٢٤ ساعة بس، الفاتورة المكررة ما تنحسب، المهمة
// تنحسب من يوصل موعدها، الإجازة يوم عطلة، واللقطة وحدة باليوم.
func TestDutiesRules_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود بالبيئة — تخطي اختبار القاعدة الحية")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	r := NewDutiesRepository(db)
	const own, fin, ldr, cust, bk = "__du_own", "__du_fin", "__du_ldr", "__du_cust", "__du_bk"
	cleanup := func() {
		db.Exec(`DELETE FROM "DutySnapshot" WHERE "employeeId" IN ($1, $2, $3)`, own, fin, ldr)
		db.Exec(`DELETE FROM "ExtraTask" WHERE "assignedToId" = $1`, fin)
		db.Exec(`DELETE FROM "LeaveRequest" WHERE "employeeId" = $1`, fin)
		db.Exec(`DELETE FROM "LeaderInvoice" WHERE "bookingId" = $1`, bk)
		db.Exec(`DELETE FROM "Booking" WHERE id = $1`, bk)
		db.Exec(`DELETE FROM "Customer" WHERE id = $1`, cust)
		db.Exec(`DELETE FROM "Employee" WHERE id IN ($1, $2, $3)`, own, fin, ldr)
	}
	cleanup()
	defer cleanup()
	db.MustExec(`INSERT INTO "Employee"(id, name, status, role) VALUES ($1, 'مالك', 'ACTIVE', 'OWNER'), ($2, 'محاسب', 'ACTIVE', 'FINANCE'), ($3, 'ليدر', 'ACTIVE', 'TECHNICIAN')`, own, fin, ldr)

	if rep := r.Today(own); rep.Position != "ADMIN" || len(rep.Duties) != 0 {
		t.Fatalf("المالك ما ينقيّم: %+v", rep)
	}
	friday := r.dayOff(fin) == "FRIDAY"

	audit := func() Duty {
		for _, d := range r.Today(fin).Duties {
			if d.Key == "audit" {
				return d
			}
		}
		t.Fatal("ماكو واجب تدقيق")
		return Duty{}
	}
	base := audit()
	if base.Late == nil {
		t.Fatal("واجب التدقيق لازم يحسب المتأخر")
	}
	l0 := *base.Late

	// فاتورة وصلت هسه: بالطابور بس مو متأخرة
	db.MustExec(`INSERT INTO "Customer"(id, "customerCode", name, phone, "createdAt") VALUES ($1, 987656, 'ز', '07700000002', now())`, cust)
	db.MustExec(`INSERT INTO "Booking"(id, code, "customerId", status, "createdAt", "updatedAt") VALUES ($1, 'DU-1', $2, 'COMPLETED', now(), now())`, bk, cust)
	db.MustExec(`INSERT INTO "LeaderInvoice"(id, "bookingId", "employeeId", systems, items, "netTotal", "accountingCode", status)
		VALUES ('__du_inv1', $1, $2, '[]', '[]', 1000, '__du_c1', 'SUBMITTED')`, bk, ldr)
	if d := audit(); *d.Pending != *base.Pending+1 || *d.Late != l0 {
		t.Fatalf("الجديدة تنحسب بالطابور بس مو متأخرة: %+v", d)
	}
	db.MustExec(`UPDATE "LeaderInvoice" SET "createdAt" = "createdAt" - interval '25 hours' WHERE id = '__du_inv1'`)
	if d := audit(); *d.Late != l0+1 {
		t.Fatalf("بعد ٢٤ ساعة تصير متأخرة: %+v", d)
	}

	// توأم معتمد بنفس اليوم: المكررة تطلع من الطابور وما تنعتمد
	db.MustExec(`INSERT INTO "LeaderInvoice"(id, "bookingId", "employeeId", systems, items, "netTotal", "accountingCode", status, "externalInvoiceNumber", "createdAt")
		SELECT '__du_inv2', $1, $2, '[]', '[]', 1000, '__du_c2', 'APPROVED', '__du_n2', "createdAt" FROM "LeaderInvoice" WHERE id = '__du_inv1'`, bk, ldr)
	if d := audit(); *d.Late != l0 {
		t.Fatalf("المكررة لازم تطلع من الطابور: %+v", d)
	}
	if !NewLeaderInvoiceRepository(db).HasApprovedTwin("__du_inv1") {
		t.Fatal("HasApprovedTwin لازم يلگى التوأم")
	}

	// مهمة موعدها الأسبوع الجاي: معلومة مو ناقصة
	db.MustExec(`INSERT INTO "ExtraTask"(id, title, "assignedToId", priority, status, "dueAt", "createdAt") VALUES ('__du_t1', 'بعدين', $1, 'NORMAL', 'NEW', now() + interval '7 days', now())`, fin)
	for _, d := range r.Today(fin).Duties {
		if d.Key == "task:__du_t1" && d.Status != "INFO" {
			t.Fatalf("مهمة موعدها بعدين ما تنحسب: %+v", d)
		}
	}

	// إجازة معتمدة اليوم: يوم عطلة
	db.MustExec(`INSERT INTO "LeaveRequest"(id, "employeeId", "startDate", "endDate", status, route, reason, "createdAt")
		VALUES ('__du_lv', $1, baghdad_today(), baghdad_today(), 'APPROVED', 'ADMIN', 'اختبار', now())`, fin)
	if rep := r.Today(fin); rep.Verdict != "OFF" || rep.Total != 0 {
		t.Fatalf("الإجازة يوم عطلة: %+v", rep)
	} else if !friday && rep.Off != "LEAVE" {
		t.Fatalf("Off = %q", rep.Off)
	}

	// اللقطة: وحدة باليوم
	if _, err := r.SnapshotToday(); err != nil {
		t.Fatalf("SnapshotToday: %v", err)
	}
	if _, err := r.SnapshotToday(); err != nil {
		t.Fatalf("SnapshotToday 2: %v", err)
	}
	var n int
	db.Get(&n, `SELECT count(*) FROM "DutySnapshot" WHERE "employeeId" IN ($1, $2, $3) AND day = baghdad_today()`, own, fin, ldr)
	if n != 2 {
		t.Fatalf("اللقطة لازم وحدة لكل موظف (بلا المالك): %d", n)
	}
	if h, err := r.History(fin, 30); err != nil || len(h) != 1 || h[0].Verdict != "OFF" {
		t.Fatalf("History: %v %+v", err, h)
	}
}
