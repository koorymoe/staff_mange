package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type ExpenseRepository struct {
	db *sqlx.DB
}

func NewExpenseRepository(db *sqlx.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (r *ExpenseRepository) List(employeeID string) ([]model.Expense, error) {
	var expenses []model.Expense
	var err error
	if employeeID != "" {
		err = r.db.Select(&expenses, `SELECT * FROM "Expense" WHERE "employeeId" = $1 ORDER BY "createdAt" DESC`, employeeID)
	} else {
		err = r.db.Select(&expenses, `SELECT * FROM "Expense" ORDER BY "createdAt" DESC`)
	}
	if err != nil {
		return nil, err
	}

	for i := range expenses {
		var brief model.EmployeeBrief
		if err := r.db.Get(&brief, `SELECT id, name, position FROM "Employee" WHERE id = $1`, expenses[i].EmployeeID); err == nil {
			expenses[i].Employee = &brief
		}
	}
	return expenses, nil
}

func (r *ExpenseRepository) Create(employeeID string, amount float64, description *string) (*model.Expense, error) {
	var e model.Expense
	err := r.db.Get(&e, `
		INSERT INTO "Expense" (id, "employeeId", amount, description)
		VALUES (gen_random_uuid()::text, $1, $2, $3)
		RETURNING *
	`, employeeID, amount, description)
	if err != nil {
		return nil, err
	}
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name, position FROM "Employee" WHERE id = $1`, e.EmployeeID); err == nil {
		e.Employee = &brief
	}
	return &e, nil
}

func (r *ExpenseRepository) UpdateStatus(id, status string) (*model.Expense, error) {
	var e model.Expense
	err := r.db.Get(&e, `UPDATE "Expense" SET status = $2 WHERE id = $1 RETURNING *`, id, status)
	if err != nil {
		return nil, err
	}
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name, position FROM "Employee" WHERE id = $1`, e.EmployeeID); err == nil {
		e.Employee = &brief
	}
	return &e, nil
}
