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
	expenses := []model.Expense{}
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
		// ⚠️ كود الحجز للعرض: الي يعتمد المصروف لازم يعرف **على أي
		// حجز** ينصرف — «مصروف ٥٠ ألف لفلان» بلا حجز يُعتمد بالثقة
		// مو بالمراجعة.
		if expenses[i].BookingID != nil && *expenses[i].BookingID != "" {
			var code string
			if err := r.db.Get(&code, `SELECT code FROM "Booking" WHERE id = $1`, *expenses[i].BookingID); err == nil {
				expenses[i].BookingCode = &code
			}
		}
	}
	return expenses, nil
}

// IsExpenseResponsible هل هذا الموظف مسؤول عن مصاريف هالحجز؟
//
// ⚠️⚠️ **الفحص بالخادم مو بقائمة الواجهة.** الواجهة تعرض للّيدر
// حجوزاته بس، لكن القائمة **مو حماية**: نداء مباشر بمعرّف حجز ثانٍ
// يتخطّاها ويحمّل حجز زميله مصروفاً.
//
// ⚠️ والشرط الي يقبله النظام اليوم للمسؤولية: المشرف على المشروع، أو
// المكلّف بالمصاريف، أو ليدر مكلّف بالحجز — نفس ما تحسبه الواجهة.
func (r *ExpenseRepository) IsExpenseResponsible(employeeID, bookingID string) (bool, error) {
	var n int
	err := r.db.Get(&n, `
		SELECT COUNT(*) FROM "Booking" b
		WHERE b.id = $2
		  AND ( b."expenseResponsibleId" = $1
		     OR b."projectSupervisorId"  = $1
		     OR EXISTS (SELECT 1 FROM "BookingAssignment" ba
		                JOIN "Employee" e ON e.id = ba."employeeId"
		                WHERE ba."bookingId" = b.id AND ba."employeeId" = $1 AND e."isLeader") )`,
		employeeID, bookingID)
	return n > 0, err
}

func (r *ExpenseRepository) Create(employeeID string, amount float64, description *string, bookingID *string) (*model.Expense, error) {
	var e model.Expense
	err := r.db.Get(&e, `
		INSERT INTO "Expense" (id, "employeeId", amount, description, "bookingId")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4)
		RETURNING *
	`, employeeID, amount, description, bookingID)
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
