package repository

import "github.com/jmoiron/sqlx"

// ═══ أرقام شهر واحد (مقارنة الأشهر + أداة ماتركس) — قرار (ع) 10-10 ═══
type MonthStats struct {
	Month             string  `db:"month" json:"month"`
	BookingsCreated   int     `db:"bookingsCreated" json:"bookingsCreated"`
	BookingsCompleted int     `db:"bookingsCompleted" json:"bookingsCompleted"`
	BookingsCancelled int     `db:"bookingsCancelled" json:"bookingsCancelled"`
	InternalWorks     int     `db:"internalWorks" json:"internalWorks"`
	ProjectsCreated   int     `db:"projectsCreated" json:"projectsCreated"`
	NewCustomers      int     `db:"newCustomers" json:"newCustomers"`
	Complaints        int     `db:"complaints" json:"complaints"`
	BookingsRevenue   float64 `db:"bookingsRevenue" json:"bookingsRevenue"`
	ProjectsRevenue   float64 `db:"projectsRevenue" json:"projectsRevenue"`
	Expenses          float64 `db:"expenses" json:"expenses"`
}

// MonthStatsFor أرقام شهر (YYYY-MM) بنفس تعريفها بباقي النظام.
func MonthStatsFor(db *sqlx.DB, m string) (*MonthStats, error) {
	var st MonthStats
	err := db.Get(&st, `
			SELECT $1::text AS month,
			  (SELECT COUNT(*) FROM "Booking" b WHERE `+BookingCountableSQL("b")+`
			     AND company_period(baghdad_date(b."createdAt")) = $1) AS "bookingsCreated",
			  (SELECT COUNT(*) FROM "Booking" b WHERE `+BookingCountableSQL("b")+`
			     AND b.status::text IN ('COMPLETED','PARTIAL') AND b."bookingType" IS DISTINCT FROM 'INTERNAL'
			     AND company_period(baghdad_date(b."completedAt")) = $1) AS "bookingsCompleted",
			  (SELECT COUNT(*) FROM "Booking" b WHERE `+BookingCountableSQL("b")+`
			     AND b.status::text = 'CANCELLED' AND company_period(baghdad_date(b."createdAt")) = $1) AS "bookingsCancelled",
			  (SELECT COUNT(*) FROM "Booking" b WHERE `+BookingCountableSQL("b")+`
			     AND b."bookingType" = 'INTERNAL' AND b.status::text = 'COMPLETED'
			     AND company_period(baghdad_date(b."completedAt")) = $1) AS "internalWorks",
			  (SELECT COUNT(*) FROM "Project" p WHERE company_period(baghdad_date(p."createdAt")) = $1) AS "projectsCreated",
			  (SELECT COUNT(*) FROM "Customer" c WHERE company_period(baghdad_date(c."createdAt")) = $1) AS "newCustomers",
			  (SELECT COUNT(*) FROM "Complaint" q WHERE company_period(baghdad_date(q."createdAt")) = $1) AS "complaints",
			  (SELECT COALESCE(SUM(`+RevenueAmountSQL("b")+`), 0) FROM "Booking" b WHERE `+BookingCountableSQL("b")+`
			     AND `+RevenueBookingSQL("b")+` AND company_period(baghdad_date(b."completedAt")) = $1) AS "bookingsRevenue",
			  (SELECT COALESCE(SUM(pp.amount), 0) FROM "ProjectPayment" pp WHERE pp."cancelledAt" IS NULL
			     AND company_period(baghdad_date(pp."paidAt")) = $1) AS "projectsRevenue",
			  (SELECT COALESCE(SUM(e.amount), 0) FROM "Expense" e WHERE e.status = 'APPROVED'
			     AND company_period(baghdad_date(e."createdAt")) = $1) AS "expenses"`, m)
	if err != nil {
		return nil, err
	}
	return &st, nil
}
