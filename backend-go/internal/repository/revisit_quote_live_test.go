package repository

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// (ع) 11-07: الرجعة = حجز ثاني لنفس الزبون ونفس الخدمة خلال ٣٠ يوم من الإنجاز.
func TestRevisitReport_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	clean := func() {
		db.Exec(`DELETE FROM "Booking" WHERE code LIKE 'RV-%'`)
		db.Exec(`DELETE FROM "Customer" WHERE id IN ('rv-c1','rv-c2')`)
		db.Exec(`DELETE FROM "Service" WHERE id IN ('rv-s1','rv-s2')`)
	}
	clean()
	defer clean()
	db.MustExec(`INSERT INTO "Customer" (id, name, phone) VALUES ('rv-c1','زبون رجعة','07700000991'), ('rv-c2','زبون ثاني','07700000992')`)
	db.MustExec(`INSERT INTO "Service" (id, name) VALUES ('rv-s1','خدمة رجعة ١'), ('rv-s2','خدمة رجعة ٢')`)
	ins := func(code, cust, svc, status string, doneDaysAgo, createdDaysAgo int) {
		db.MustExec(`INSERT INTO "Booking" (id, code, "customerId", "serviceId", status, "completedAt", "createdAt")
			VALUES ($1, $1, $2, $3, $4::"BookingStatus",
			        CASE WHEN $5 >= 0 THEN now() - make_interval(days => $5) END, now() - make_interval(days => $6))`,
			code, cust, svc, status, doneDaysAgo, createdDaysAgo)
	}
	// c1/s1: انجز قبل ٤٠ يوم، ورجع بعد ١٠ أيام ← رجعة
	ins("RV-1", "rv-c1", "rv-s1", "COMPLETED", 40, 45)
	ins("RV-2", "rv-c1", "rv-s1", "PENDING", -1, 30)
	// c1/s2: خدمة ثانية لنفس الزبون ← مو رجعة
	ins("RV-3", "rv-c1", "rv-s2", "COMPLETED", 20, 25)
	// c2/s1: رجع بعد ٥٠ يوم ← برّا النافذة
	ins("RV-4", "rv-c2", "rv-s1", "COMPLETED", 80, 85)
	ins("RV-5", "rv-c2", "rv-s1", "PENDING", -1, 30)
	// ملغي ما ينحسب رجعة
	ins("RV-6", "rv-c2", "rv-s2", "COMPLETED", 15, 16)
	ins("RV-7", "rv-c2", "rv-s2", "CANCELLED", -1, 10)

	rep, err := NewRevisitRepository(db).Report(90)
	if err != nil {
		t.Fatalf("report: %v", err)
	}
	got := map[string]string{}
	for _, it := range rep.Items {
		if len(it.FirstCode) > 3 && it.FirstCode[:3] == "RV-" {
			got[it.FirstCode] = it.SecondCode
		}
	}
	if len(got) != 1 || got["RV-1"] != "RV-2" {
		t.Fatalf("الرجعات غلط: %v (نريد RV-1→RV-2 بس)", got)
	}
}

// صلاحية العرض: يومين قبل ينبّه مرة وحدة، والمقبول ما ينبّه.
func TestClaimExpiringQuotations_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()
	clean := func() {
		db.Exec(`DELETE FROM "Quotation" WHERE "quotationNumber" LIKE 'QX-%'`)
		db.Exec(`DELETE FROM "Employee" WHERE id = 'qx-emp'`)
	}
	clean()
	defer clean()
	db.MustExec(`INSERT INTO "Employee"(id, name) VALUES ('qx-emp','منشئ عروض')`)
	q := func(num, status string, days int) {
		db.MustExec(`INSERT INTO "Quotation" (id, "quotationNumber", "customerName", "createdByEmployeeId", status, "validUntil")
			VALUES (gen_random_uuid()::text, $1, 'زبون', 'qx-emp', $2, baghdad_today() + $3::int)`, num, status, days)
	}
	q("QX-1", "SENT", 1)     // ينبّه
	q("QX-2", "NEW", 10)     // بعيد
	q("QX-3", "ACCEPTED", 1) // انحسم
	q("QX-4", "NEW", -3)     // انتهى قريباً ← ينبّه
	q("QX-5", "SENT", -30)   // قديم هواية ← لا
	r := NewQuotationRepository(db)
	rows, err := r.ClaimExpiringQuotations()
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	seen := map[string]bool{}
	for _, x := range rows {
		seen[x.QuotationNumber] = true
	}
	if !seen["QX-1"] || !seen["QX-4"] || seen["QX-2"] || seen["QX-3"] || seen["QX-5"] {
		t.Fatalf("التنبيه غلط: %v", seen)
	}
	again, _ := r.ClaimExpiringQuotations()
	for _, x := range again {
		if x.QuotationNumber == "QX-1" || x.QuotationNumber == "QX-4" {
			t.Fatalf("انبّه مرتين على %s", x.QuotationNumber)
		}
	}
}
