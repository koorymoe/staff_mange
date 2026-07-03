package repository

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type ProcurementRepository struct {
	db *sqlx.DB
}

func NewProcurementRepository(db *sqlx.DB) *ProcurementRepository {
	return &ProcurementRepository{db: db}
}

func (r *ProcurementRepository) loadEmployeeBrief(id string) *model.EmployeeBrief {
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name, role FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &brief
}

func (r *ProcurementRepository) loadBookingBrief(id string) *model.ProcurementBookingBrief {
	var b model.ProcurementBookingBrief
	var customerID string
	if err := r.db.Get(&customerID, `SELECT "customerId" FROM "Booking" WHERE id = $1`, id); err != nil {
		return nil
	}
	b.ID = id
	var customer model.CustomerBrief
	if err := r.db.Get(&customer, `SELECT id, name, phone FROM "Customer" WHERE id = $1`, customerID); err == nil {
		b.Customer = &customer
	}
	return &b
}

func (r *ProcurementRepository) hydrate(req *model.ProcurementRequest, withFulfilledBy bool) {
	items := []model.ProcurementItem{}
	if err := r.db.Select(&items, `SELECT * FROM "ProcurementItem" WHERE "requestId" = $1`, req.ID); err == nil {
		req.Items = items
	}
	req.RequestedBy = r.loadEmployeeBrief(req.RequestedByID)
	if withFulfilledBy && req.FulfilledByID != nil {
		req.FulfilledBy = r.loadEmployeeBrief(*req.FulfilledByID)
	}
	if req.BookingID != nil {
		req.Booking = r.loadBookingBrief(*req.BookingID)
	}
}

func (r *ProcurementRepository) List() ([]model.ProcurementRequest, error) {
	requests := []model.ProcurementRequest{}
	if err := r.db.Select(&requests, `SELECT * FROM "ProcurementRequest" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	for i := range requests {
		r.hydrate(&requests[i], true)
	}
	return requests, nil
}

var prCodeRe = regexp.MustCompile(`PR-(\d+)`)

func (r *ProcurementRepository) nextCode() (string, error) {
	var lastCode string
	err := r.db.Get(&lastCode, `SELECT code FROM "ProcurementRequest" ORDER BY code DESC LIMIT 1`)
	nextNum := 1
	if err == nil {
		if m := prCodeRe.FindStringSubmatch(lastCode); m != nil {
			if n, convErr := strconv.Atoi(m[1]); convErr == nil {
				nextNum = n + 1
			}
		}
	}
	return fmt.Sprintf("PR-%04d", nextNum), nil
}

func (r *ProcurementRepository) Create(req model.CreateProcurementRequestRequest) (*model.ProcurementRequest, error) {
	code, err := r.nextCode()
	if err != nil {
		return nil, err
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var request model.ProcurementRequest
	err = tx.Get(&request, `
		INSERT INTO "ProcurementRequest" (id, code, "requestedById", "bookingId", notes)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4)
		RETURNING *
	`, code, req.RequestedByID, req.BookingID, req.Notes)
	if err != nil {
		return nil, err
	}

	for _, item := range req.Items {
		if _, err := tx.Exec(`
			INSERT INTO "ProcurementItem" (id, "requestId", "productName", quantity, "unitPrice", "totalPrice")
			VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5)
		`, request.ID, item.ProductName, item.Quantity, item.UnitPrice, item.TotalPrice); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	r.hydrate(&request, false)
	return &request, nil
}

func (r *ProcurementRepository) UpdateStatus(id, status string) (*model.ProcurementRequest, error) {
	var request model.ProcurementRequest
	err := r.db.Get(&request, `UPDATE "ProcurementRequest" SET status = $2 WHERE id = $1 RETURNING *`, id, status)
	if err != nil {
		return nil, err
	}
	r.hydrate(&request, true)
	return &request, nil
}

func (r *ProcurementRepository) UpdateItem(id string, unitPrice, totalPrice *float64, fulfilled bool) error {
	_, err := r.db.Exec(`
		UPDATE "ProcurementItem" SET
			"unitPrice" = COALESCE($2, "unitPrice"),
			"totalPrice" = COALESCE($3, "totalPrice"),
			fulfilled = $4
		WHERE id = $1
	`, id, unitPrice, totalPrice, fulfilled)
	return err
}

func (r *ProcurementRepository) Fulfill(id string, req model.FulfillProcurementRequestRequest) (*model.ProcurementRequest, error) {
	for _, item := range req.Items {
		fulfilled := true
		if item.Fulfilled != nil {
			fulfilled = *item.Fulfilled
		}
		if err := r.UpdateItem(item.ID, item.UnitPrice, item.TotalPrice, fulfilled); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	var request model.ProcurementRequest
	err := r.db.Get(&request, `
		UPDATE "ProcurementRequest" SET
			status = 'FULFILLED',
			"fulfilledById" = $2,
			"totalCost" = $3,
			"fulfillmentNotes" = $4,
			"fulfilledAt" = $5
		WHERE id = $1
		RETURNING *
	`, id, req.FulfilledByID, req.TotalCost, req.FulfillmentNotes, now)
	if err != nil {
		return nil, err
	}
	r.hydrate(&request, true)
	return &request, nil
}

func (r *ProcurementRepository) Stats() (*model.ProcurementStats, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	sixMonthsAgo := time.Date(now.Year(), now.Month()-5, 1, 0, 0, 0, 0, time.UTC)

	var totalSpent sql.NullFloat64
	if err := r.db.Get(&totalSpent, `SELECT SUM("totalCost") FROM "ProcurementRequest" WHERE status = 'FULFILLED'`); err != nil {
		return nil, err
	}
	var totalItems int
	if err := r.db.Get(&totalItems, `SELECT COUNT(*) FROM "ProcurementItem"`); err != nil {
		return nil, err
	}
	var pendingCount int
	if err := r.db.Get(&pendingCount, `SELECT COUNT(*) FROM "ProcurementRequest" WHERE status = 'PENDING'`); err != nil {
		return nil, err
	}
	var monthlySpent sql.NullFloat64
	if err := r.db.Get(&monthlySpent, `SELECT SUM("totalCost") FROM "ProcurementRequest" WHERE status = 'FULFILLED' AND "fulfilledAt" >= $1`, startOfMonth); err != nil {
		return nil, err
	}
	var fulfilledCount int
	if err := r.db.Get(&fulfilledCount, `SELECT COUNT(*) FROM "ProcurementRequest" WHERE status = 'FULFILLED'`); err != nil {
		return nil, err
	}

	type monthlyRow struct {
		TotalCost   *float64  `db:"totalCost"`
		FulfilledAt time.Time `db:"fulfilledAt"`
	}
	var monthlyRows []monthlyRow
	if err := r.db.Select(&monthlyRows, `
		SELECT "totalCost", "fulfilledAt" FROM "ProcurementRequest"
		WHERE status = 'FULFILLED' AND "fulfilledAt" >= $1
	`, sixMonthsAgo); err != nil {
		return nil, err
	}

	byMonth := map[string]float64{}
	for _, row := range monthlyRows {
		key := fmt.Sprintf("%d-%02d", row.FulfilledAt.Year(), int(row.FulfilledAt.Month()))
		cost := 0.0
		if row.TotalCost != nil {
			cost = *row.TotalCost
		}
		byMonth[key] += cost
	}

	stats := &model.ProcurementStats{
		TotalItems:     totalItems,
		PendingCount:   pendingCount,
		FulfilledCount: fulfilledCount,
		ByMonth:        byMonth,
	}
	if totalSpent.Valid {
		stats.TotalSpent = totalSpent.Float64
	}
	if monthlySpent.Valid {
		stats.MonthlySpent = monthlySpent.Float64
	}
	return stats, nil
}
