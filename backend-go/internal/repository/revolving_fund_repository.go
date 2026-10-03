package repository

import (
	"fmt"
	"strings"
	"unicode/utf8"

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
		// الي الدوار ناطره من المحاسب: مصروف مدقّق وما انخرّج بعد
		_ = r.db.Get(&funds[i].AwaitingDischargeTotal, `
			SELECT COALESCE(SUM("spentAmount"), 0) FROM "RevolvingFundTxn"
			WHERE "fundId" = $1 AND kind = 'SETTLEMENT' AND status = 'APPROVED'
			  AND "spentAmount" > 0 AND "dischargedAt" IS NULL`, funds[i].ID)
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

const fundTxnSelect = `SELECT t.*, f.name AS "fundName", e.name AS "employeeName",
		rv.name AS "reviewedByName", dc.name AS "dischargedByName"
	FROM "RevolvingFundTxn" t
	JOIN "RevolvingFund" f ON f.id = t."fundId"
	LEFT JOIN "Employee" e ON e.id = t."employeeId"
	LEFT JOIN "Employee" rv ON rv.id = t."reviewedById"
	LEFT JOIN "Employee" dc ON dc.id = t."dischargedById"`

func decorateTxns(rows []model.RevolvingFundTxn) []model.RevolvingFundTxn {
	for i := range rows {
		rows[i].KindLabel = model.FundTxnKindLabels[rows[i].Kind]
		rows[i].StatusLabel = model.FundTxnStatusLabels[rows[i].Status]
		rows[i].AwaitingDischarge = rows[i].Kind == model.FundTxnSettlement &&
			rows[i].Status == model.FundTxnApproved &&
			rows[i].SpentAmount > 0 && rows[i].DischargedAt == nil
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
			(id, "fundId", "employeeId", kind, amount, "bookingId", "requestId", notes, status, "createdById")
		VALUES (gen_random_uuid()::text, $1, $2, 'DISBURSE', $3, $4, $5, NULLIF($6,''), 'APPROVED', $7)
		RETURNING *`, req.FundID, req.EmployeeID, req.Amount, req.BookingID, req.RequestID, deref(req.Notes), byID); err != nil {
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
	// ═══ التسوية لدوار واحد بعينه ═══
	//
	// ⚠️ الواجهة چانت **تخمّن** الدوار من آخر عملية تسليم، فتسوية
	// تغطي دوارين تنختم بالدوار الي دفع آخر مرة، والفلوس ترجع كلها
	// إله. هسه الدوار إجباري، والتحقق على رصيد **ذاك الدوار** لا
	// على المجموع.
	if strings.TrimSpace(req.FundID) == "" {
		return nil, fmt.Errorf("لازم تحدد من أي دوار هاي التسوية")
	}
	outstanding, err := r.EmployeeOutstandingInFund(employeeID, req.FundID)
	if err != nil {
		return nil, err
	}
	// ما نخلي الموظف يسوّي أكثر من الي بيده **من هذا الدوار** — وإلا
	// يرجع لدوار فلوس ما طلعت منه، ويطلع رصيد الدوار الثاني بالسالب
	if total > outstanding+0.001 {
		return nil, fmt.Errorf("المبلغ (%.0f) أكبر من الي بيدك من هذا الدوار (%.0f)", total, outstanding)
	}
	if req.SpentAmount > 0 && (req.ReceiptImage == nil || *req.ReceiptImage == "") {
		return nil, fmt.Errorf("لازم ترفع صورة الوصل للمبلغ المصروف")
	}
	// ═══ بيان الصرف إجباري ═══
	//
	// «الموظف الي أحوّله فلوس من الدوار، لما يسوي حسابه يطلعله
	// ريكوست فيلد».
	//
	// الوصل يثبت إن الفلوس انصرفت، بس **ما يگول على شنو**. بدون بيان
	// المحاسب يشوف «صرف ٢٠٠ ألف» ولازم يتصل يسأل، وبعد أسبوع الموظف
	// ما يتذكر. والحجز المربوط يجاوب لحاله، فأي واحد منهم يكفي.
	//
	// ⚠️ شرط تعبئة بس — ماكو مبلغ ولا رصيد ينتغيّر.
	if req.SpentAmount > 0 {
		hasBooking := req.BookingID != nil && strings.TrimSpace(*req.BookingID) != ""
		note := ""
		if req.Notes != nil {
			note = strings.TrimSpace(*req.Notes)
		}
		// ⚠️ العدّ بالحروف مو بالبايتات: «شريت» عربية ٥ حروف بس ١٠
		// بايتات، وlen() چان يقبل كلمة وحدة كأنها بيان كامل.
		if !hasBooking && utf8.RuneCountInString(note) < 5 {
			return nil, fmt.Errorf("اكتب على شنو انصرفت الفلوس، أو اربطها بحجز")
		}
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

// Discharge المحاسب يأشر إن المادة انخرجت على النظام المحاسبي ويحدد على
// أي حساب انحسبت — وهنا بس يرجع المبلغ المصروف للدوار.
//
// المبلغ الي يرجع هو "spentAmount" وحده، مو كل السلفة: المرتجع رجع من
// قبل وقت الموافقة على التسوية. يعني موظف أخذ 50 وصرف 20 ورجّع 30 —
// رجعت الـ30 بالتدقيق، والـ20 ترجع هنا. وموظف أخذ 100 وصرفهن كلهن —
// ما يرجع ولا دينار قبل التخريج، وبالتخريج ترجع الـ100 كاملة.
//
// شرط "dischargedAt" IS NULL بالتحديث نفسه يمنع تخريج نفس التسوية
// مرتين — وإلا انضاف المبلغ للدوار مرتين من ضغطتين متزامنتين.
func (r *RevolvingFundRepository) Discharge(id, account string, note *string, byID *string) (*model.RevolvingFundTxn, error) {
	if account == "" {
		return nil, fmt.Errorf("لازم تحدد على أي حساب انخرجت المادة")
	}
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var t model.RevolvingFundTxn
	if err := tx.Get(&t, `
		UPDATE "RevolvingFundTxn"
		SET "dischargedAt" = now(), "dischargedById" = $2,
			"dischargeAccount" = $3, "dischargeNote" = NULLIF($4,'')
		WHERE id = $1 AND kind = 'SETTLEMENT' AND status = 'APPROVED'
		  AND "spentAmount" > 0 AND "dischargedAt" IS NULL
		RETURNING *`, id, byID, account, deref(note)); err != nil {
		return nil, fmt.Errorf("التسوية مو جاهزة للتخريج — لازم تكون مدققة، بيها مبلغ مصروف، وما انخرجت من قبل")
	}

	if _, err := tx.Exec(`UPDATE "RevolvingFund" SET balance = balance + $2, "updatedAt" = now() WHERE id = $1`,
		t.FundID, t.SpentAmount); err != nil {
		return nil, err
	}
	// حركة مستقلة حتى يبين بالكشف من وين رجع المبلغ وعلى أي حساب
	if _, err := tx.Exec(`
		INSERT INTO "RevolvingFundTxn"
			(id, "fundId", "employeeId", kind, amount, "dischargeAccount", notes, status, "createdById", "dischargedById", "dischargedAt")
		VALUES (gen_random_uuid()::text, $1, $2, 'DISCHARGE', $3, $4, NULLIF($5,''), 'APPROVED', $6, $6, now())`,
		t.FundID, t.EmployeeID, t.SpentAmount, account, deref(note), byID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	t.KindLabel = model.FundTxnKindLabels[t.Kind]
	t.StatusLabel = model.FundTxnStatusLabels[t.Status]
	return &t, nil
}

// DischargeAccounts الحسابات الي انستخدمت من قبل — تنعرض كقائمة منسدلة
// بجنب خانة الكتابة الحرة، حتى ما يكتب المحاسب نفس الحساب بصيغتين.
func (r *RevolvingFundRepository) DischargeAccounts() ([]string, error) {
	rows := []string{}
	err := r.db.Select(&rows, `
		SELECT DISTINCT "dischargeAccount" FROM "RevolvingFundTxn"
		WHERE "dischargeAccount" IS NOT NULL AND "dischargeAccount" <> ''
		ORDER BY 1`)
	return rows, err
}

// employeeFundLinesSQL الي بيد الموظف مفصَّلاً **لكل دوار**.
//
// ⚠️ هذا أساس الإصلاح كله: قبل، الرصيد جان يتحسب `WHERE "employeeId" = $1`
// بلا `fundId` بالمرة — فالنظام يعرف «عليه ٣٠ ألف» وما يعرف ١٠ من وين
// و٢٠ من وين. وبلا هالبيانة، ما اكو طريقة نرجّع كل مبلغ لدواره.
//
// نفس معادلة EmployeeOutstanding بالضبط، بس مجمّعة على الدوار:
// المستلم − (المصروف + المرتجع) للتسويات **المقبولة** فقط.
const employeeFundLinesSQL = `
	SELECT f.id AS "fundId", f.name AS "fundName",
		COALESCE(SUM(CASE WHEN t.kind = 'DISBURSE' THEN t.amount ELSE 0 END), 0) AS "totalTaken",
		COALESCE(SUM(CASE WHEN t.kind = 'DISBURSE' THEN t.amount
			WHEN t.kind = 'SETTLEMENT' AND t.status = 'APPROVED'
			THEN -(t."spentAmount" + t."returnedAmount") ELSE 0 END), 0) AS outstanding,
		COALESCE(SUM(CASE WHEN t.kind = 'SETTLEMENT' AND t.status = 'PENDING' THEN 1 ELSE 0 END), 0) AS "pendingSettlements"
	FROM "RevolvingFundTxn" t
	JOIN "RevolvingFund" f ON f.id = t."fundId"
	WHERE t."employeeId" = $1
	GROUP BY f.id, f.name
	HAVING COALESCE(SUM(CASE WHEN t.kind = 'DISBURSE' THEN t.amount ELSE 0 END), 0) > 0
	ORDER BY outstanding DESC, f.name`

// EmployeeFundLines الي بيد الموظف من كل دوار على حِدة.
func (r *RevolvingFundRepository) EmployeeFundLines(employeeID string) ([]model.EmployeeFundLine, error) {
	rows := []model.EmployeeFundLine{}
	err := r.db.Select(&rows, employeeFundLinesSQL, employeeID)
	return rows, err
}

// EmployeeOutstandingInFund شكد بيد الموظف من **دوار واحد** بعينه.
//
// ⚠️ هاي الي تتحقق منها التسوية — لا المجموع. بلاها، موظف أخذ ١٠ من
// دوار الطاقة يقدر يرفع تسوية ٣٠ ألف عليه (لأن مجموعه ٣٠)، فيرجع
// لدوار الطاقة ٢٠ ألف ما طلعت منه أصلاً.
func (r *RevolvingFundRepository) EmployeeOutstandingInFund(employeeID, fundID string) (float64, error) {
	var v float64
	err := r.db.Get(&v, `
		SELECT COALESCE(SUM(CASE WHEN kind = 'DISBURSE' THEN amount ELSE 0 END), 0)
		     - COALESCE(SUM(CASE WHEN kind = 'SETTLEMENT' AND status = 'APPROVED'
		                         THEN "spentAmount" + "returnedAmount" ELSE 0 END), 0)
		FROM "RevolvingFundTxn" WHERE "employeeId" = $1 AND "fundId" = $2`, employeeID, fundID)
	return v, err
}

// EmployeeOutstanding شكد بيد هذا الموظف وما انتسوّى — **مجموع كل الدوارات**.
//
// ⚠️ للعرض بس (بطاقة «مطلوب منك»). أي تحقق قبل حركة فلوس لازم يستعمل
// EmployeeOutstandingInFund، وإلا الفلوس ترجع لدوار غلط.
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
	// التفصيل لكل دوار — بلاه الموظف ما يعرف من وين يسوّي، والواجهة
	// تضطر تخمّن الدوار (وهاي چانت العلّة).
	if lines, err := r.EmployeeFundLines(employeeID); err == nil {
		b.Funds = lines
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
	if err != nil {
		return nil, err
	}
	// نفس التفصيل للمحاسب — يشوف الموظف عليه شكد **من كل دوار**، مو
	// رقماً واحداً مجموعاً ما يدل على وين ترجع الفلوس.
	for i := range rows {
		if lines, e := r.EmployeeFundLines(rows[i].EmployeeID); e == nil {
			rows[i].Funds = lines
		}
	}
	return rows, nil
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
