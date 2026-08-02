package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// RevolvingFundRepository يدير الدوار: الصرف للموظفين والتسويات والتدقيق.
//
// قاعدة أساسية: التسوية ما تأثر على ولا رصيد إلا بعد ما المحاسب يوافق.
// كل الحسابات تعتمد status = 'APPROVED' — التسوية المعلّقة ما تنقص دَين
// الموظف ولا ترجّع فلوس للدوار.
type RevolvingFundRepository struct {
	db *sqlx.DB
}

func NewRevolvingFundRepository(db *sqlx.DB) *RevolvingFundRepository {
	return &RevolvingFundRepository{db: db}
}

func (r *RevolvingFundRepository) ListFunds() ([]model.RevolvingFund, error) {
	funds := []model.RevolvingFund{}
	if err := r.db.Select(&funds, `SELECT * FROM "RevolvingFund" ORDER BY name`); err != nil {
		return nil, err
	}
	// المبلغ الي بيد الموظفين وما انتسوّى — لكل دوار
	for i := range funds {
		_ = r.db.Get(&funds[i].OutstandingTotal, `
			SELECT COALESCE(SUM(d.amount), 0) - COALESCE((
				SELECT SUM(s."spentAmount" + s."returnedAmount") FROM "RevolvingFundTxn" s
				WHERE s."fundId" = $1 AND s.kind = 'SETTLEMENT' AND s.status = 'APPROVED'
			), 0)
			FROM "RevolvingFundTxn" d
			WHERE d."fundId" = $1 AND d.kind = 'DISBURSE'`, funds[i].ID)
	}
	return funds, nil
}

func (r *RevolvingFundRepository) UpdateFund(id string, req model.UpsertFundRequest) (*model.RevolvingFund, error) {
	var f model.RevolvingFund
	err := r.db.Get(&f, `
		UPDATE "RevolvingFund" SET
			name = COALESCE($2, name),
			balance = COALESCE($3, balance),
			"isActive" = COALESCE($4, "isActive"),
			"updatedAt" = now()
		WHERE id = $1 RETURNING *`, id, req.Name, req.Balance, req.IsActive)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

const fundTxnSelect = `SELECT t.*, f.name AS "fundName", e.name AS "employeeName", rv.name AS "reviewedByName"
	FROM "RevolvingFundTxn" t
	JOIN "RevolvingFund" f ON f.id = t."fundId"
	LEFT JOIN "Employee" e ON e.id = t."employeeId"
	LEFT JOIN "Employee" rv ON rv.id = t."reviewedById"`

func decorateTxns(rows []model.RevolvingFundTxn) []model.RevolvingFundTxn {
	for i := range rows {
		rows[i].KindLabel = model.FundTxnKindLabels[rows[i].Kind]
		rows[i].StatusLabel = model.FundTxnStatusLabels[rows[i].Status]
	}
	return rows
}

// Disburse المحاسب يسلّم موظف مبلغ من الدوار.
//
// نخصم من رصيد الدوار بنفس المعاملة، وشرط `balance >= amount` بالتحديث
// نفسه يمنع الصرف لو ما اكو رصيد كافي — حتى لو صرفين انطلبو بنفس اللحظة.
func (r *RevolvingFundRepository) Disburse(req model.DisburseRequest, byID *string) (*model.RevolvingFundTxn, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("المبلغ لازم يكون أكبر من صفر")
	}
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var newBalance float64
	if err := tx.Get(&newBalance, `
		UPDATE "RevolvingFund" SET balance = balance - $2, "updatedAt" = now()
		WHERE id = $1 AND "isActive" = true AND balance >= $2
		RETURNING balance`, req.FundID, req.Amount); err != nil {
		return nil, fmt.Errorf("رصيد الدوار ما يكفي، أو الدوار موقوف")
	}

	var t model.RevolvingFundTxn
	if err := tx.Get(&t, `
		INSERT INTO "RevolvingFundTxn"
			(id, "fundId", "employeeId", kind, amount, "bookingId", notes, status, "createdById")
		VALUES (gen_random_uuid()::text, $1, $2, 'DISBURSE', $3, $4, NULLIF($5,''), 'APPROVED', $6)
		RETURNING *`, req.FundID, req.EmployeeID, req.Amount, req.BookingID, deref(req.Notes), byID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	t.KindLabel = model.FundTxnKindLabels[t.Kind]
	t.StatusLabel = model.FundTxnStatusLabels[t.Status]
	return &t, nil
}

// Topup تغذية الدوار نفسه.
func (r *RevolvingFundRepository) Topup(fundID string, amount float64, notes *string, byID *string) error {
	if amount <= 0 {
		return fmt.Errorf("المبلغ لازم يكون أكبر من صفر")
	}
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE "RevolvingFund" SET balance = balance + $2, "updatedAt" = now() WHERE id = $1`, fundID, amount); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT INTO "RevolvingFundTxn" (id, "fundId", kind, amount, notes, status, "createdById")
		VALUES (gen_random_uuid()::text, $1, 'TOPUP', $2, NULLIF($3,''), 'APPROVED', $4)`,
		fundID, amount, deref(notes), byID); err != nil {
		return err
	}
	return tx.Commit()
}

// SubmitSettlement الموظف يرفع تسوية: شكد صرف، شكد رجّع، وصورة الوصل.
//
// تنحفظ PENDING — ما تأثر على ولا رصيد لحد ما المحاسب يوافق.
func (r *RevolvingFundRepository) SubmitSettlement(employeeID string, req model.SettlementRequest) (*model.RevolvingFundTxn, error) {
	total := req.SpentAmount + req.ReturnedAmount
	if total <= 0 {
		return nil, fmt.Errorf("لازم تحدد المبلغ المصروف أو المرتجع")
	}
	outstanding, err := r.EmployeeOutstanding(employeeID)
	if err != nil {
		return nil, err
	}
	// ما نخلي الموظف يسوّي أكثر من الي بيده — وإلا يطلع رصيده بالسالب
	if total > outstanding+0.001 {
		return nil, fmt.Errorf("المبلغ (%.0f) أكبر من الي بيدك (%.0f)", total, outstanding)
	}
	if req.SpentAmount > 0 && (req.ReceiptImage == nil || *req.ReceiptImage == "") {
		return nil, fmt.Errorf("لازم ترفع صورة الوصل للمبلغ المصروف")
	}

	var t model.RevolvingFundTxn
	err = r.db.Get(&t, `
		INSERT INTO "RevolvingFundTxn"
			(id, "fundId", "employeeId", kind, "spentAmount", "returnedAmount",
			 "bookingId", "receiptImage", notes, status, "createdById")
		VALUES (gen_random_uuid()::text, $1, $2, 'SETTLEMENT', $3, $4, $5, $6, NULLIF($7,''), 'PENDING', $2)
		RETURNING *`, req.FundID, employeeID, req.SpentAmount, req.ReturnedAmount,
		req.BookingID, req.ReceiptImage, deref(req.Notes))
	if err != nil {
		return nil, err
	}
	t.KindLabel = model.FundTxnKindLabels[t.Kind]
	t.StatusLabel = model.FundTxnStatusLabels[t.Status]
	return &t, nil
}

// ReviewSettlement المحاسب يدقّق: يوافق أو يرفض.
//
// الموافقة هي الي تصفّر دَين الموظف وترجّع المبلغ المرتجع للدوار. الرفض
// يخلي المبلغ برقبة الموظف مثل ما هو.
func (r *RevolvingFundRepository) ReviewSettlement(id string, req model.ReviewSettlementRequest, byID *string) (*model.RevolvingFundTxn, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	status := model.FundTxnRejected
	if req.Approve {
		status = model.FundTxnApproved
	}

	var t model.RevolvingFundTxn
	// شرط status='PENDING' يمنع تدقيق نفس التسوية مرتين (ورجوع الفلوس مرتين)
	if err := tx.Get(&t, `
		UPDATE "RevolvingFundTxn"
		SET status = $2, "reviewedById" = $3, "reviewedAt" = now(), "reviewNote" = NULLIF($4,'')
		WHERE id = $1 AND kind = 'SETTLEMENT' AND status = 'PENDING'
		RETURNING *`, id, status, byID, deref(req.ReviewNote)); err != nil {
		return nil, fmt.Errorf("التسوية مو موجودة أو انتدققت من قبل")
	}

	// المبلغ المرتجع يرجع للدوار عند الموافقة بس
	if req.Approve && t.ReturnedAmount > 0 {
		if _, err := tx.Exec(`UPDATE "RevolvingFund" SET balance = balance + $2, "updatedAt" = now() WHERE id = $1`,
			t.FundID, t.ReturnedAmount); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	t.KindLabel = model.FundTxnKindLabels[t.Kind]
	t.StatusLabel = model.FundTxnStatusLabels[t.Status]
	return &t, nil
}

// EmployeeOutstanding شكد بيد هذا الموظف وما انتسوّى.
func (r *RevolvingFundRepository) EmployeeOutstanding(employeeID string) (float64, error) {
	var v float64
	err := r.db.Get(&v, `
		SELECT COALESCE(SUM(CASE WHEN kind = 'DISBURSE' THEN amount ELSE 0 END), 0)
		     - COALESCE(SUM(CASE WHEN kind = 'SETTLEMENT' AND status = 'APPROVED'
		                         THEN "spentAmount" + "returnedAmount" ELSE 0 END), 0)
		FROM "RevolvingFundTxn" WHERE "employeeId" = $1`, employeeID)
	return v, err
}

// EmployeeBalance ملخص رصيد موظف — يظهر بلوحته الرئيسية.
func (r *RevolvingFundRepository) EmployeeBalance(employeeID string) (*model.EmployeeFundBalance, error) {
	var b model.EmployeeFundBalance
	err := r.db.Get(&b, `
		SELECT e.id AS "employeeId", e.name AS "employeeName", e."jobTitle",
			COALESCE(SUM(CASE WHEN t.kind = 'DISBURSE' THEN t.amount ELSE 0 END), 0) AS "totalTaken",
			COALESCE(SUM(CASE WHEN t.kind = 'SETTLEMENT' AND t.status = 'APPROVED'
				THEN t."spentAmount" + t."returnedAmount" ELSE 0 END), 0) AS "totalSettled",
			COALESCE(SUM(CASE WHEN t.kind = 'DISBURSE' THEN t.amount
				WHEN t.kind = 'SETTLEMENT' AND t.status = 'APPROVED'
				THEN -(t."spentAmount" + t."returnedAmount") ELSE 0 END), 0) AS outstanding,
			COALESCE(SUM(CASE WHEN t.kind = 'SETTLEMENT' AND t.status = 'PENDING' THEN 1 ELSE 0 END), 0) AS "pendingSettlements"
		FROM "Employee" e
		LEFT JOIN "RevolvingFundTxn" t ON t."employeeId" = e.id
		WHERE e.id = $1
		GROUP BY e.id, e.name, e."jobTitle"`, employeeID)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// EmployeeBalances كل الموظفين الي ساحبين من الدوار — شاشة المحاسب.
func (r *RevolvingFundRepository) EmployeeBalances() ([]model.EmployeeFundBalance, error) {
	rows := []model.EmployeeFundBalance{}
	err := r.db.Select(&rows, `
		SELECT e.id AS "employeeId", e.name AS "employeeName", e."jobTitle",
			COALESCE(SUM(CASE WHEN t.kind = 'DISBURSE' THEN t.amount ELSE 0 END), 0) AS "totalTaken",
			COALESCE(SUM(CASE WHEN t.kind = 'SETTLEMENT' AND t.status = 'APPROVED'
				THEN t."spentAmount" + t."returnedAmount" ELSE 0 END), 0) AS "totalSettled",
			COALESCE(SUM(CASE WHEN t.kind = 'DISBURSE' THEN t.amount
				WHEN t.kind = 'SETTLEMENT' AND t.status = 'APPROVED'
				THEN -(t."spentAmount" + t."returnedAmount") ELSE 0 END), 0) AS outstanding,
			COALESCE(SUM(CASE WHEN t.kind = 'SETTLEMENT' AND t.status = 'PENDING' THEN 1 ELSE 0 END), 0) AS "pendingSettlements"
		FROM "Employee" e
		JOIN "RevolvingFundTxn" t ON t."employeeId" = e.id
		GROUP BY e.id, e.name, e."jobTitle"
		HAVING COALESCE(SUM(CASE WHEN t.kind = 'DISBURSE' THEN t.amount ELSE 0 END), 0) > 0
		ORDER BY outstanding DESC, e.name`)
	return rows, err
}

// ListTxns حركات الدوار. employeeID فاضي = الكل (للمحاسب).
func (r *RevolvingFundRepository) ListTxns(employeeID, status string) ([]model.RevolvingFundTxn, error) {
	rows := []model.RevolvingFundTxn{}
	q := fundTxnSelect + ` WHERE 1=1`
	args := []any{}
	if employeeID != "" {
		args = append(args, employeeID)
		q += fmt.Sprintf(` AND t."employeeId" = $%d`, len(args))
	}
	if status != "" {
		args = append(args, status)
		q += fmt.Sprintf(` AND t.status = $%d`, len(args))
	}
	q += ` ORDER BY t."createdAt" DESC LIMIT 500`
	if err := r.db.Select(&rows, q, args...); err != nil {
		return nil, err
	}
	return decorateTxns(rows), nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
