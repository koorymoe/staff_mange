package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/xuri/excelize/v2"

	"staffmange-api/internal/middleware"
)

// ═══ سجل الآجل — قرار (ع) 10-14 ═══
//
// حجز أو مشروع نكمله وما ناخذ فلوسه مباشرة: الفلوس تتأجل، دفعات أو مبلغ
// كامل. الحجز يدخل هنا من التدقيق اليومي بزر «⏳ الآجل»، والمشروع يضيفه
// المحاسب يدوياً. كل دفعة تنسجّل، والدين يتسكّر لمن يكتمل، وأكو إكسل شهري.
//
// 💰 الدفعة على حجز تنضاف لمستلمه (amountCollected) — حتى أرقام الأرباح
// تبقى حقيقية: الفلوس تنحسب يوم توصل، مو يوم خلص الشغل.

type CreditHandler struct{ db *sqlx.DB }

func NewCreditHandler(db *sqlx.DB) *CreditHandler { return &CreditHandler{db: db} }

type creditRow struct {
	ID           string     `db:"id" json:"id"`
	Kind         string     `db:"kind" json:"kind"`
	BookingID    *string    `db:"bookingId" json:"bookingId"`
	BookingCode  *string    `db:"bookingCode" json:"bookingCode"`
	ProjectID    *string    `db:"projectId" json:"projectId"`
	ProjectCode  *string    `db:"projectCode" json:"projectCode"`
	Title        string     `db:"title" json:"title"`
	CustomerName *string    `db:"customerName" json:"customerName"`
	Phone        *string    `db:"phone" json:"phone"`
	Amount       float64    `db:"amount" json:"amount"`
	Paid         float64    `db:"paid" json:"paid"`
	Note         *string    `db:"note" json:"note"`
	DueDate      *time.Time `db:"dueDate" json:"dueDate"`
	CreatedBy    *string    `db:"createdByName" json:"createdByName"`
	CreatedAt    time.Time  `db:"createdAt" json:"createdAt"`
	SettledAt    *time.Time `db:"settledAt" json:"settledAt"`
	LastPaidAt   *time.Time `db:"lastPaidAt" json:"lastPaidAt"`
}

const creditSelect = `SELECT cs.id, cs.kind, cs."bookingId", b.code AS "bookingCode", cs."projectId", p.code AS "projectCode",
		cs.title, cs."customerName", cs.phone, cs.amount::float8 AS amount,
		COALESCE((SELECT sum(amount) FROM "CreditPayment" x WHERE x."creditId" = cs.id), 0)::float8 AS paid,
		cs.note, cs."dueDate", e.name AS "createdByName", cs."createdAt", cs."settledAt",
		(SELECT max("paidAt") FROM "CreditPayment" x WHERE x."creditId" = cs.id) AS "lastPaidAt"
	FROM "CreditSale" cs
	LEFT JOIN "Booking" b ON b.id = cs."bookingId"
	LEFT JOIN "Project" p ON p.id = cs."projectId"
	LEFT JOIN "Employee" e ON e.id = cs."createdById"`

// GET /api/credits?status=open|settled|all&q=
func (h *CreditHandler) List(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	where := `WHERE cs."settledAt" IS NULL`
	switch status {
	case "settled":
		where = `WHERE cs."settledAt" IS NOT NULL`
	case "all":
		where = `WHERE true`
	}
	rows := []creditRow{}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	err := h.db.Select(&rows, creditSelect+" "+where+` AND ($1 = '' OR cs.title ILIKE '%'||$1||'%' OR cs."customerName" ILIKE '%'||$1||'%'
		OR cs.phone LIKE '%'||$1||'%' OR b.code ILIKE '%'||$1||'%' OR p.code ILIKE '%'||$1||'%')
		ORDER BY cs."settledAt" NULLS FIRST, cs."createdAt" DESC LIMIT 500`, q)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الآجل")
		return
	}
	WriteJSON(w, http.StatusOK, rows)
}

// GET /api/credits/{id}/payments
func (h *CreditHandler) Payments(w http.ResponseWriter, r *http.Request) {
	rows := []struct {
		ID     string    `db:"id" json:"id"`
		Amount float64   `db:"amount" json:"amount"`
		PaidAt time.Time `db:"paidAt" json:"paidAt"`
		Note   *string   `db:"note" json:"note"`
		By     *string   `db:"byName" json:"byName"`
	}{}
	_ = h.db.Select(&rows, `SELECT x.id, x.amount::float8 AS amount, x."paidAt", x.note, e.name AS "byName"
		FROM "CreditPayment" x LEFT JOIN "Employee" e ON e.id = x."byId" WHERE x."creditId" = $1 ORDER BY x."paidAt", x."createdAt"`, r.PathValue("id"))
	WriteJSON(w, http.StatusOK, rows)
}

// POST /api/credits — آجل يدوي (مشروع، أو حجز ما مر بالتدقيق)
func (h *CreditHandler) Create(w http.ResponseWriter, r *http.Request) {
	var b struct {
		ProjectCode  string  `json:"projectCode"`
		BookingCode  string  `json:"bookingCode"`
		Title        string  `json:"title"`
		CustomerName string  `json:"customerName"`
		Phone        string  `json:"phone"`
		Amount       float64 `json:"amount"`
		Note         string  `json:"note"`
		DueDate      string  `json:"dueDate"`
	}
	if err := DecodeJSON(r, &b); err != nil || b.Amount <= 0 {
		WriteError(w, http.StatusBadRequest, "اكتب المبلغ المستحق")
		return
	}
	kind, title := "OTHER", strings.TrimSpace(b.Title)
	var projectID, bookingID *string
	if c := strings.TrimSpace(b.ProjectCode); c != "" {
		var p struct {
			ID    string  `db:"id"`
			Name  string  `db:"name"`
			Phone *string `db:"phone"`
		}
		if err := h.db.Get(&p, `SELECT id, name, phone FROM "Project" WHERE code ILIKE $1`, c); err != nil {
			WriteError(w, http.StatusBadRequest, "ما لگيت مشروع بالكود "+c)
			return
		}
		kind, projectID = "PROJECT", &p.ID
		if title == "" {
			title = "مشروع " + strings.ToUpper(c) + " — " + p.Name
		}
		if b.Phone == "" && p.Phone != nil {
			b.Phone = *p.Phone
		}
	} else if c := strings.TrimPrefix(strings.TrimSpace(b.BookingCode), "#"); c != "" {
		var bk struct {
			ID    string  `db:"id"`
			Code  string  `db:"code"`
			Name  *string `db:"name"`
			Phone *string `db:"phone"`
		}
		if err := h.db.Get(&bk, `SELECT b.id, b.code, c.name, c.phone FROM "Booking" b LEFT JOIN "Customer" c ON c.id = b."customerId" WHERE b.code ILIKE $1`, c); err != nil {
			WriteError(w, http.StatusBadRequest, "ما لگيت حجز بالكود "+c)
			return
		}
		kind, bookingID = "BOOKING", &bk.ID
		if title == "" {
			title = "حجز " + bk.Code
		}
		if b.CustomerName == "" && bk.Name != nil {
			b.CustomerName = *bk.Name
		}
		if b.Phone == "" && bk.Phone != nil {
			b.Phone = *bk.Phone
		}
	}
	if title == "" {
		WriteError(w, http.StatusBadRequest, "اكتب كود المشروع أو الحجز، أو عنوان للآجل")
		return
	}
	var due *string
	if d := strings.TrimSpace(b.DueDate); d != "" {
		due = &d
	}
	_, err := h.db.Exec(`INSERT INTO "CreditSale" (id, kind, "bookingId", "projectId", title, "customerName", phone, amount, note, "dueDate", "createdById")
		VALUES ($1,$2,$3,$4,$5,NULLIF($6,''),NULLIF($7,''),$8,NULLIF($9,''),$10::date,$11)`,
		uuid.NewString(), kind, bookingID, projectID, title, strings.TrimSpace(b.CustomerName), strings.TrimSpace(b.Phone), b.Amount, strings.TrimSpace(b.Note), due,
		middleware.EmployeeIDFromContext(r))
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			WriteError(w, http.StatusConflict, "هذا الحجز مسجّل بالآجل أصلاً")
			return
		}
		WriteError(w, http.StatusBadRequest, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]bool{"ok": true})
}

// POST /api/credits/{id}/payments {amount, paidAt, note} — دفعة (أو المبلغ كامل)
func (h *CreditHandler) AddPayment(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Amount float64 `json:"amount"`
		PaidAt string  `json:"paidAt"`
		Note   string  `json:"note"`
	}
	if err := DecodeJSON(r, &b); err != nil || b.Amount <= 0 {
		WriteError(w, http.StatusBadRequest, "اكتب مبلغ الدفعة")
		return
	}
	id := r.PathValue("id")
	tx, err := h.db.Beginx()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "خطأ")
		return
	}
	defer func() { _ = tx.Rollback() }()
	var c struct {
		Amount    float64    `db:"amount"`
		Paid      float64    `db:"paid"`
		BookingID *string    `db:"bookingId"`
		Settled   *time.Time `db:"settledAt"`
	}
	if err := tx.Get(&c, `SELECT amount::float8 AS amount, "bookingId", "settledAt",
			COALESCE((SELECT sum(amount) FROM "CreditPayment" WHERE "creditId" = $1), 0)::float8 AS paid
		FROM "CreditSale" WHERE id = $1 FOR UPDATE`, id); err != nil {
		WriteError(w, http.StatusNotFound, "الآجل غير موجود")
		return
	}
	remaining := c.Amount - c.Paid
	if c.Settled != nil || remaining <= 0 {
		WriteError(w, http.StatusBadRequest, "هذا الآجل مسدّد بالكامل")
		return
	}
	if b.Amount > remaining+0.5 {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("الدفعة أكبر من الباقي (%s)", formatIQD(remaining)))
		return
	}
	paidAt := strings.TrimSpace(b.PaidAt)
	if paidAt == "" {
		paidAt = time.Now().In(baghdadLoc()).Format("2006-01-02")
	}
	if _, err := tx.Exec(`INSERT INTO "CreditPayment" (id, "creditId", amount, "paidAt", note, "byId") VALUES ($1,$2,$3,$4::date,NULLIF($5,''),$6)`,
		uuid.NewString(), id, b.Amount, paidAt, strings.TrimSpace(b.Note), middleware.EmployeeIDFromContext(r)); err != nil {
		WriteError(w, http.StatusBadRequest, "تعذر تسجيل الدفعة — تأكد من التاريخ")
		return
	}
	// الفلوس وصلت فعلاً → تنضاف لمستلم الحجز
	if c.BookingID != nil {
		_, _ = tx.Exec(`UPDATE "Booking" SET "amountCollected" = COALESCE("amountCollected", 0) + $2 WHERE id = $1`, *c.BookingID, b.Amount)
	}
	if b.Amount >= remaining-0.5 {
		_, _ = tx.Exec(`UPDATE "CreditSale" SET "settledAt" = now() WHERE id = $1`, id)
	}
	if err := tx.Commit(); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/credits/export?month=YYYY-MM — إكسل الآجل الشهري:
// ورقة «الآجل» (كل دين مفتوح أو انفتح/انسدّ بالشهر، بالمدفوع والباقي)
// وورقة «الدفعات» (كل دفعة وصلت بالشهر).
func (h *CreditHandler) Export(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	start, err := time.Parse("2006-01", month)
	if err != nil {
		start, _ = time.Parse("2006-01", time.Now().In(baghdadLoc()).Format("2006-01"))
		month = start.Format("2006-01")
	}
	end := start.AddDate(0, 1, 0)
	rows := []creditRow{}
	if err := h.db.Select(&rows, creditSelect+` WHERE cs."createdAt" < $2 AND (cs."settledAt" IS NULL OR cs."settledAt" >= $1)
		ORDER BY cs."createdAt"`, start, end); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب الآجل")
		return
	}
	pays := []struct {
		Title  string    `db:"title"`
		Cust   *string   `db:"customerName"`
		Amount float64   `db:"amount"`
		PaidAt time.Time `db:"paidAt"`
		Note   *string   `db:"note"`
		By     *string   `db:"byName"`
	}{}
	_ = h.db.Select(&pays, `SELECT cs.title, cs."customerName", x.amount::float8 AS amount, x."paidAt", x.note, e.name AS "byName"
		FROM "CreditPayment" x JOIN "CreditSale" cs ON cs.id = x."creditId" LEFT JOIN "Employee" e ON e.id = x."byId"
		WHERE x."paidAt" >= $1 AND x."paidAt" < $2 ORDER BY x."paidAt"`, start, end)

	f := excelize.NewFile()
	head, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"0F2040"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"}})
	bold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	rtl := true
	writeSheet := func(name string, headers []string, data [][]any, widths []float64) {
		_ = f.SetSheetView(name, 0, &excelize.ViewOptions{RightToLeft: &rtl})
		for i, hd := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			_ = f.SetCellValue(name, cell, hd)
			_ = f.SetCellStyle(name, cell, cell, head)
			col, _ := excelize.ColumnNumberToName(i + 1)
			_ = f.SetColWidth(name, col, col, widths[i])
		}
		for ri, row := range data {
			for ci, v := range row {
				cell, _ := excelize.CoordinatesToCellName(ci+1, ri+2)
				_ = f.SetCellValue(name, cell, v)
			}
		}
	}
	str := func(p *string) string {
		if p == nil || *p == "" {
			return "—"
		}
		return *p
	}
	d := func(t *time.Time) string {
		if t == nil {
			return "—"
		}
		return t.In(baghdadLoc()).Format("2006-01-02")
	}
	s1 := "الآجل"
	f.SetSheetName("Sheet1", s1)
	var tAmt, tPaid float64
	data := [][]any{}
	for _, c := range rows {
		state := "مفتوح"
		if c.SettledAt != nil {
			state = "مسدّد"
		}
		data = append(data, []any{c.Title, str(c.CustomerName), str(c.Phone), c.Amount, c.Paid, c.Amount - c.Paid, state,
			d(&c.CreatedAt), d(c.LastPaidAt), d(c.DueDate), str(c.CreatedBy), str(c.Note)})
		tAmt += c.Amount
		tPaid += c.Paid
	}
	data = append(data, []any{fmt.Sprintf("المجموع — %d", len(rows)), "", "", tAmt, tPaid, tAmt - tPaid})
	writeSheet(s1, []string{"الحجز/المشروع", "الزبون", "الهاتف", "المستحق", "المدفوع", "الباقي", "الحالة", "تاريخ الآجل", "آخر دفعة", "موعد السداد", "سجّله", "ملاحظة"},
		data, []float64{34, 22, 16, 14, 14, 14, 10, 13, 13, 13, 18, 30})
	_ = f.SetCellStyle(s1, fmt.Sprintf("A%d", len(data)+1), fmt.Sprintf("F%d", len(data)+1), bold)

	s2 := "الدفعات"
	_, _ = f.NewSheet(s2)
	var tPay float64
	pdata := [][]any{}
	for _, p := range pays {
		pdata = append(pdata, []any{p.PaidAt.Format("2006-01-02"), p.Title, str(p.Cust), p.Amount, str(p.By), str(p.Note)})
		tPay += p.Amount
	}
	pdata = append(pdata, []any{"المجموع", fmt.Sprintf("%d دفعة", len(pays)), "", tPay})
	writeSheet(s2, []string{"التاريخ", "الحجز/المشروع", "الزبون", "المبلغ", "استلمها", "ملاحظة"}, pdata, []float64{13, 34, 22, 14, 18, 30})
	_ = f.SetCellStyle(s2, fmt.Sprintf("A%d", len(pdata)+1), fmt.Sprintf("D%d", len(pdata)+1), bold)

	filename := fmt.Sprintf("الآجل-%s.xlsx", month)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, url.QueryEscape(filename)))
	if err := f.Write(w); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر إنشاء ملف الإكسل")
	}
}
