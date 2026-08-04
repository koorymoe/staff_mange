package repository

import (
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type StatsRepository struct {
	db *sqlx.DB
}

func NewStatsRepository(db *sqlx.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) count(query string, args ...any) (int, error) {
	var n int
	err := r.db.Get(&n, query, args...)
	return n, err
}

func (r *StatsRepository) sum(query string, args ...any) (float64, error) {
	var n float64
	err := r.db.Get(&n, query, args...)
	return n, err
}

func (r *StatsRepository) Totals() (*model.StatsTotals, error) {
	totalCustomers, err := r.count(`SELECT COUNT(*) FROM "Customer"`)
	if err != nil {
		return nil, err
	}
	totalBookings, err := r.count(`SELECT COUNT(*) FROM "Booking"`)
	if err != nil {
		return nil, err
	}
	pending, err := r.count(`SELECT COUNT(*) FROM "Booking" WHERE status = 'PENDING'`)
	if err != nil {
		return nil, err
	}
	confirmed, err := r.count(`SELECT COUNT(*) FROM "Booking" WHERE status = 'CONFIRMED'`)
	if err != nil {
		return nil, err
	}
	completed, err := r.count(`SELECT COUNT(*) FROM "Booking" WHERE status = 'COMPLETED'`)
	if err != nil {
		return nil, err
	}
	cancelled, err := r.count(`SELECT COUNT(*) FROM "Booking" WHERE status = 'CANCELLED'`)
	if err != nil {
		return nil, err
	}
	urgentPending, err := r.count(`SELECT COUNT(*) FROM "Booking" WHERE status = 'PENDING' AND priority = 'URGENT'`)
	if err != nil {
		return nil, err
	}
	revenue, err := r.sum(`SELECT COALESCE(SUM("amountCollected"), 0) FROM "Booking" WHERE "amountCollected" IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	unverifiedRevenue, err := r.sum(`SELECT COALESCE(SUM("amountCollected"), 0) FROM "Booking" WHERE "amountCollected" IS NOT NULL AND "amountVerified" = false`)
	if err != nil {
		return nil, err
	}

	return &model.StatsTotals{
		TotalCustomers:    totalCustomers,
		TotalBookings:     totalBookings,
		PendingBookings:   pending,
		ConfirmedBookings: confirmed,
		CompletedBookings: completed,
		CancelledBookings: cancelled,
		UrgentPending:     urgentPending,
		TotalRevenue:      revenue,
		UnverifiedRevenue: unverifiedRevenue,
	}, nil
}

func (r *StatsRepository) SalesStats(startOfToday, startOfMonth time.Time) ([]model.SalesStat, error) {
	type row struct {
		EmployeeID string `db:"id"`
		Name       string `db:"name"`
	}
	var employees []row
	if err := r.db.Select(&employees, `SELECT id, name FROM "Employee" WHERE role IN ('SALES', 'ADMIN')`); err != nil {
		return nil, err
	}

	stats := make([]model.SalesStat, 0, len(employees))
	for _, e := range employees {
		total, err := r.count(`SELECT COUNT(*) FROM "Booking" WHERE "transferEmployeeId" = $1`, e.EmployeeID)
		if err != nil {
			return nil, err
		}
		confirmed, err := r.count(`SELECT COUNT(*) FROM "Booking" WHERE "transferEmployeeId" = $1 AND status NOT IN ('PENDING', 'CANCELLED')`, e.EmployeeID)
		if err != nil {
			return nil, err
		}
		today, err := r.count(`SELECT COUNT(*) FROM "Booking" WHERE "transferEmployeeId" = $1 AND "createdAt" >= $2`, e.EmployeeID, startOfToday)
		if err != nil {
			return nil, err
		}
		thisMonth, err := r.count(`SELECT COUNT(*) FROM "Booking" WHERE "transferEmployeeId" = $1 AND "createdAt" >= $2`, e.EmployeeID, startOfMonth)
		if err != nil {
			return nil, err
		}
		stats = append(stats, model.SalesStat{
			EmployeeID:       e.EmployeeID,
			Name:             e.Name,
			TotalTransferred: total,
			Confirmed:        confirmed,
			Today:            today,
			ThisMonth:        thisMonth,
		})
	}
	return stats, nil
}

func (r *StatsRepository) CoordinatorStats(startOfToday, startOfMonth time.Time) ([]model.CoordinatorStat, error) {
	type row struct {
		EmployeeID string `db:"id"`
		Name       string `db:"name"`
	}
	var employees []row
	if err := r.db.Select(&employees, `SELECT id, name FROM "Employee" WHERE role IN ('HR_COORDINATOR', 'ADMIN')`); err != nil {
		return nil, err
	}

	stats := make([]model.CoordinatorStat, 0, len(employees))
	for _, e := range employees {
		total, err := r.count(`SELECT COUNT(*) FROM "Booking" WHERE "confirmedByEmployeeId" = $1`, e.EmployeeID)
		if err != nil {
			return nil, err
		}
		today, err := r.count(`SELECT COUNT(*) FROM "Booking" WHERE "confirmedByEmployeeId" = $1 AND "createdAt" >= $2`, e.EmployeeID, startOfToday)
		if err != nil {
			return nil, err
		}
		thisMonth, err := r.count(`SELECT COUNT(*) FROM "Booking" WHERE "confirmedByEmployeeId" = $1 AND "createdAt" >= $2`, e.EmployeeID, startOfMonth)
		if err != nil {
			return nil, err
		}
		stats = append(stats, model.CoordinatorStat{
			EmployeeID:     e.EmployeeID,
			Name:           e.Name,
			TotalConfirmed: total,
			Today:          today,
			ThisMonth:      thisMonth,
		})
	}
	return stats, nil
}

func (r *StatsRepository) TechnicianStats() ([]model.TechnicianStat, error) {
	type row struct {
		EmployeeID string `db:"id"`
		Name       string `db:"name"`
		OnDuty     bool   `db:"onDuty"`
	}
	var employees []row
	if err := r.db.Select(&employees, `SELECT id, name, "onDuty" FROM "Employee" WHERE role = 'TECHNICIAN'`); err != nil {
		return nil, err
	}

	stats := make([]model.TechnicianStat, 0, len(employees))
	for _, e := range employees {
		totalAssigned, err := r.count(`SELECT COUNT(*) FROM "BookingAssignment" WHERE "employeeId" = $1`, e.EmployeeID)
		if err != nil {
			return nil, err
		}
		completed, err := r.count(`
			SELECT COUNT(*) FROM "BookingAssignment" a
			JOIN "Booking" b ON b.id = a."bookingId"
			WHERE a."employeeId" = $1 AND b.status = 'COMPLETED'
		`, e.EmployeeID)
		if err != nil {
			return nil, err
		}
		revenue, err := r.sum(`
			SELECT COALESCE(SUM(b."amountCollected"), 0) FROM "BookingAssignment" a
			JOIN "Booking" b ON b.id = a."bookingId"
			WHERE a."employeeId" = $1 AND b.status = 'COMPLETED'
		`, e.EmployeeID)
		if err != nil {
			return nil, err
		}
		stats = append(stats, model.TechnicianStat{
			EmployeeID:     e.EmployeeID,
			Name:           e.Name,
			OnDuty:         e.OnDuty,
			TotalAssigned:  totalAssigned,
			Completed:      completed,
			RevenueHandled: revenue,
		})
	}
	return stats, nil
}

func (r *StatsRepository) ServiceBreakdown() ([]model.ServiceBreakdownEntry, error) {
	type row struct {
		ServiceID *string    `db:"serviceId"`
		Name      *string    `db:"name"`
		Count     int        `db:"count"`
		FirstAt   *time.Time `db:"firstAt"`
		LastAt    *time.Time `db:"lastAt"`
		Revenue   float64    `db:"revenue"`
		Cost      float64    `db:"cost"`
	}
	var rows []row
	// المبالغ والكلفة تنحسب بحجز واحد لكل صف قبل التجميع — لو ضمّينا
	// CartItem مباشرة، حجز بثلاث مواد يضاعف مبلغه ثلاث مرات.
	err := r.db.Select(&rows, `
		WITH per_booking AS (
			SELECT b.id, b."serviceId", b."createdAt",
			       COALESCE(b."amountCollected", 0) + COALESCE(b."advancePaid", 0) AS collected,
			       COALESCE((
			           SELECT SUM(ci.quantity * COALESCE(p."wholesalePrice", 0))
			           FROM "CartItem" ci
			           LEFT JOIN "Product" p ON p.name = ci."productName"
			           WHERE ci."bookingId" = b.id
			       ), 0) AS cost
			FROM "Booking" b
		)
		SELECT pb."serviceId", s.name,
		       COUNT(*) AS count,
		       MIN(pb."createdAt") AS "firstAt",
		       MAX(pb."createdAt") AS "lastAt",
		       COALESCE(SUM(pb.collected), 0) AS revenue,
		       COALESCE(SUM(pb.cost), 0) AS cost
		FROM per_booking pb
		LEFT JOIN "Service" s ON s.id = pb."serviceId"
		GROUP BY pb."serviceId", s.name
	`)
	if err != nil {
		return nil, err
	}
	entries := make([]model.ServiceBreakdownEntry, 0, len(rows))
	for _, rw := range rows {
		name := "بدون خدمة"
		if rw.Name != nil {
			name = *rw.Name
		}
		entries = append(entries, model.ServiceBreakdownEntry{
			ServiceID: rw.ServiceID, Name: name, Count: rw.Count,
			FirstAt: rw.FirstAt, LastAt: rw.LastAt,
			Revenue: rw.Revenue, Profit: rw.Revenue - rw.Cost,
		})
	}
	return entries, nil
}

func (r *StatsRepository) RoleCounts() ([]model.RoleCount, error) {
	counts := []model.RoleCount{}
	err := r.db.Select(&counts, `SELECT role, COUNT(*) as count FROM "Employee" GROUP BY role`)
	return counts, err
}

func (r *StatsRepository) RecentBookings(limit int) ([]model.RecentBookingEntry, error) {
	type row struct {
		ID           string    `db:"id"`
		Code         string    `db:"code"`
		Status       string    `db:"status"`
		Priority     string    `db:"priority"`
		CustomerName string    `db:"customerName"`
		ServiceName  *string   `db:"serviceName"`
		CreatedAt    time.Time `db:"createdAt"`
	}
	var rows []row
	err := r.db.Select(&rows, `
		SELECT b.id, b.code, b.status, b.priority, c.name as "customerName", s.name as "serviceName", b."createdAt"
		FROM "Booking" b
		JOIN "Customer" c ON c.id = b."customerId"
		LEFT JOIN "Service" s ON s.id = b."serviceId"
		ORDER BY b."createdAt" DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	entries := make([]model.RecentBookingEntry, 0, len(rows))
	for _, rw := range rows {
		entries = append(entries, model.RecentBookingEntry{
			ID:           rw.ID,
			Code:         rw.Code,
			Status:       rw.Status,
			Priority:     rw.Priority,
			CustomerName: rw.CustomerName,
			ServiceName:  rw.ServiceName,
			CreatedAt:    rw.CreatedAt,
		})
	}
	return entries, nil
}
