package repository

import (
	"database/sql"
	"encoding/json"

	"github.com/lib/pq"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// LeaderInvoiceRepository يخزّن فواتير الليدر وبنود المواد التابعة لها.
type LeaderInvoiceRepository struct {
	db *sqlx.DB
}

func NewLeaderInvoiceRepository(db *sqlx.DB) *LeaderInvoiceRepository {
	return &LeaderInvoiceRepository{db: db}
}

// Create يحفظ فاتورة ليدر جديدة مع كل بنود موادها بمعاملة واحدة (كلها تنجح أو
// كلها تفشل معاً).
func (r *LeaderInvoiceRepository) Create(inv *model.LeaderInvoice, materials []model.LeaderInvoiceMaterialItem) (*model.LeaderInvoice, error) {
	systemsJSON, err := json.Marshal(inv.Systems)
	if err != nil {
		return nil, err
	}
	itemsJSON, err := json.Marshal(inv.Items)
	if err != nil {
		return nil, err
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// الكود المحاسبي النهائي يعتمد على معرف الفاتورة (gen_random_uuid) وتاريخ
	// الإنشاء، وكلاهما لا يُعرف إلا بعد الإدراج — لذا نستخدم كوداً مؤقتاً فريداً
	// بالإدراج الأول ثم نحدّثه بالكود الحقيقي بنفس المعاملة.
	var saved model.LeaderInvoice
	err = tx.Get(&saved, `
		INSERT INTO "LeaderInvoice" (
			id, "bookingId", "employeeId", "customerName", "customerPhone", "customerAddress",
			systems, items, "totalDeviceCount", "executionCost", "materialsTotal",
			"discountValue", "netTotal", "accountingCode", status
		) VALUES (
			gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, gen_random_uuid()::text, $13
		) RETURNING *
	`,
		inv.BookingID, inv.EmployeeID, inv.CustomerName, inv.CustomerPhone, inv.CustomerAddress,
		string(systemsJSON), string(itemsJSON), inv.TotalDeviceCount, inv.ExecutionCost, inv.MaterialsTotal,
		inv.DiscountValue, inv.NetTotal, inv.Status,
	)
	if err != nil {
		return nil, err
	}

	for i := range materials {
		materials[i].LeaderInvoiceID = saved.ID
		var savedItem model.LeaderInvoiceMaterialItem
		err = tx.Get(&savedItem, `
			INSERT INTO "LeaderInvoiceMaterialItem" (
				id, "leaderInvoiceId", "materialId", name, quantity, "unitPrice", "profitPerUnit", "lineTotal"
			) VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7)
			RETURNING *
		`, materials[i].LeaderInvoiceID, materials[i].MaterialID, materials[i].Name,
			materials[i].Quantity, materials[i].UnitPrice, materials[i].ProfitPerUnit, materials[i].LineTotal)
		if err != nil {
			return nil, err
		}
		saved.Materials = append(saved.Materials, savedItem)
	}

	finalCode := model.GenerateAccountingCode(saved.ID, saved.CreatedAt)
	if _, err := tx.Exec(`UPDATE "LeaderInvoice" SET "accountingCode" = $2 WHERE id = $1`, saved.ID, finalCode); err != nil {
		return nil, err
	}
	saved.AccountingCode = finalCode

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	saved.Systems = inv.Systems
	saved.Items = inv.Items
	return &saved, nil
}

// CountForEmployeeMonth يرجّع عدد فواتير الليدر (المبيعات) لموظف معيّن خلال
// شهر معيّن (monthPrefix بصيغة "YYYY-MM").
func (r *LeaderInvoiceRepository) CountForEmployeeMonth(employeeID, monthPrefix string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(*) FROM "LeaderInvoice"
		WHERE "employeeId" = $1 AND to_char("createdAt", 'YYYY-MM') = $2
	`, employeeID, monthPrefix)
	return count, err
}

// CountForEmployeeRange نفس CountForEmployeeMonth لكن لمدى تاريخ حر (from/to
// بصيغة "YYYY-MM-DD").
func (r *LeaderInvoiceRepository) CountForEmployeeRange(employeeID, from, to string) (int, error) {
	var count int
	err := r.db.Get(&count, `
		SELECT COUNT(*) FROM "LeaderInvoice"
		WHERE "employeeId" = $1 AND "createdAt"::date BETWEEN $2::date AND $3::date
	`, employeeID, from, to)
	return count, err
}

// SumNetTotalForRange يرجّع مجموع "netTotal" لكل فواتير الليدر خلال مدى تاريخ
// حر — إجمالي حجم المبيعات بالإحصائية الأسبوعية.
func (r *LeaderInvoiceRepository) SumNetTotalForRange(from, to string) (float64, error) {
	var total sql.NullFloat64
	err := r.db.Get(&total, `
		SELECT COALESCE(SUM("netTotal"), 0) FROM "LeaderInvoice" WHERE "createdAt"::date BETWEEN $1::date AND $2::date
	`, from, to)
	if err != nil {
		return 0, err
	}
	return total.Float64, nil
}

// SumNetTotalMorningEveningForRange يرجّع مجموع "netTotal" مقسوماً صباحي
// (قبل الساعة 12 ظهراً حسب وقت إنشاء الفاتورة) ومسائي خلال مدى تاريخ حر.
func (r *LeaderInvoiceRepository) SumNetTotalMorningEveningForRange(from, to string) (morning float64, evening float64, err error) {
	var m, e sql.NullFloat64
	if err = r.db.Get(&m, `
		SELECT COALESCE(SUM("netTotal"), 0) FROM "LeaderInvoice"
		WHERE "createdAt"::date BETWEEN $1::date AND $2::date AND EXTRACT(HOUR FROM "createdAt") < 12
	`, from, to); err != nil {
		return 0, 0, err
	}
	if err = r.db.Get(&e, `
		SELECT COALESCE(SUM("netTotal"), 0) FROM "LeaderInvoice"
		WHERE "createdAt"::date BETWEEN $1::date AND $2::date AND EXTRACT(HOUR FROM "createdAt") >= 12
	`, from, to); err != nil {
		return 0, 0, err
	}
	return m.Float64, e.Float64, nil
}

// SumNetTotalForDate يرجّع مجموع "netTotal" لكل فواتير الليدر المنشأة بتاريخ
// معيّن — يُستخدم كـ"إجمالي المبيعات" اليومي.
func (r *LeaderInvoiceRepository) SumNetTotalForDate(date string) (float64, error) {
	var total sql.NullFloat64
	err := r.db.Get(&total, `
		SELECT COALESCE(SUM("netTotal"), 0) FROM "LeaderInvoice" WHERE "createdAt"::date = $1::date
	`, date)
	return total.Float64, err
}

// Approve يعتمد فاتورة SUBMITTED فقط (لا يسمح باعتماد فاتورة معتمدة أصلاً
// مرة ثانية) — يرجّع nil لو الفاتورة غير موجودة أو معتمدة أصلاً.
func (r *LeaderInvoiceRepository) Approve(id, approverEmployeeID, externalNumber string) (*model.LeaderInvoice, error) {
	var inv model.LeaderInvoice
	err := r.db.Get(&inv, `
		UPDATE "LeaderInvoice"
		SET status = 'APPROVED',
		    "approvedByEmployeeId" = $2,
		    "approvedAt" = CURRENT_TIMESTAMP,
		    "externalInvoiceNumber" = $3,
		    "externalInvoiceAt" = CURRENT_TIMESTAMP
		WHERE id = $1 AND status != 'APPROVED'
		RETURNING *
	`, id, approverEmployeeID, externalNumber)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := r.hydrate(&inv); err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *LeaderInvoiceRepository) hydrate(inv *model.LeaderInvoice) error {
	if inv.SystemsJSON != "" {
		_ = json.Unmarshal([]byte(inv.SystemsJSON), &inv.Systems)
	}
	if inv.ItemsJSON != "" {
		_ = json.Unmarshal([]byte(inv.ItemsJSON), &inv.Items)
	}
	materials := []model.LeaderInvoiceMaterialItem{}
	if err := r.db.Select(&materials, `SELECT * FROM "LeaderInvoiceMaterialItem" WHERE "leaderInvoiceId" = $1 ORDER BY "createdAt"`, inv.ID); err != nil {
		return err
	}
	inv.Materials = materials

	// التفاصيل الي يحتاجها المحاسب: منو رفعها ومنو اعتمدها وأي حجز
	// تخص. بدونها الجدول يعرض «—» بمكان الزبون والليدر.
	var emp struct {
		Name  string  `db:"name"`
		Role  string  `db:"role"`
		Phone *string `db:"phone"`
	}
	if err := r.db.Get(&emp, `SELECT name, role::text, phone FROM "Employee" WHERE id = $1`, inv.EmployeeID); err == nil {
		inv.EmployeeName = emp.Name
		inv.EmployeeRole = emp.Role
		inv.EmployeePhone = emp.Phone
	}
	if inv.ApprovedByEmployeeID != nil {
		var name string
		if err := r.db.Get(&name, `SELECT name FROM "Employee" WHERE id = $1`, *inv.ApprovedByEmployeeID); err == nil {
			inv.ApprovedByName = &name
		}
	}
	if inv.BookingID != nil {
		var b model.Booking
		if err := r.db.Get(&b, `SELECT * FROM "Booking" WHERE id = $1`, *inv.BookingID); err == nil {
			code := b.Code
			inv.BookingCode = &code
			// نعبّي الزبون والخدمة والكادر حتى المحاسب يشوف كل شي بمكان واحد
			var c model.Customer
			if err := r.db.Get(&c, `SELECT * FROM "Customer" WHERE id = $1`, b.CustomerID); err == nil {
				b.Customer = &c
			}
			if b.ServiceID != nil {
				var sv model.Service
				if err := r.db.Get(&sv, `SELECT * FROM "Service" WHERE id = $1`, *b.ServiceID); err == nil {
					b.Service = &sv
				}
			}
			assignments := []model.BookingAssignment{}
			if err := r.db.Select(&assignments, `
				SELECT a.* FROM "BookingAssignment" a WHERE a."bookingId" = $1`, b.ID); err == nil {
				for i := range assignments {
					var e model.Employee
					if err := r.db.Get(&e, `SELECT * FROM "Employee" WHERE id = $1`, assignments[i].EmployeeID); err == nil {
						assignments[i].Employee = e
					}
				}
				b.Assignments = assignments
			}
			inv.Booking = &b
		}
	}
	return nil
}

// GetByID يرجّع فاتورة واحدة مع بنودها.
func (r *LeaderInvoiceRepository) GetByID(id string) (*model.LeaderInvoice, error) {
	var inv model.LeaderInvoice
	err := r.db.Get(&inv, `SELECT * FROM "LeaderInvoice" WHERE id = $1`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := r.hydrate(&inv); err != nil {
		return nil, err
	}
	return &inv, nil
}

// List يرجّع الفواتير، مصفّاة حسب الموظف لو تم تمريره.
func (r *LeaderInvoiceRepository) List(employeeID string) ([]model.LeaderInvoice, error) {
	invoices := []model.LeaderInvoice{}
	var err error
	if employeeID == "" {
		err = r.db.Select(&invoices, `SELECT * FROM "LeaderInvoice" ORDER BY "createdAt" DESC`)
	} else {
		err = r.db.Select(&invoices, `SELECT * FROM "LeaderInvoice" WHERE "employeeId" = $1 ORDER BY "createdAt" DESC`, employeeID)
	}
	if err != nil {
		return nil, err
	}
	if err := r.hydrateAll(invoices); err != nil {
		return nil, err
	}
	return invoices, nil
}

// hydrateAll يعبّي كل الفواتير سوه بعدد ثابت من الاستعلامات، بدل ما ينادي
// hydrate لكل فاتورة على حدة.
//
// hydrate الواحدة تسوي ٨ استعلامات (مواد، ليدر، معتمد، حجز، زبون، خدمة،
// تعيينات، وموظف لكل تعيين). فمعناها القائمة جانت تسوي ٨×عدد الفواتير:
// ٥٠٠ فاتورة = ٤٠٠٠ رحلة لقاعدة البيانات بفتحة وحدة للصفحة. وكل ما
// ينضاف شهر شغل تصير الصفحة أبطأ لين توقف تحمّل.
//
// هنا نجمع المفاتيح كلها أول، وننزّل كل جدول بطلب واحد (WHERE id = ANY)،
// وندمج بالذاكرة. النتيجة نفسها بالضبط — بس بعدد استعلامات ثابت مهما
// كبرت القائمة.
func (r *LeaderInvoiceRepository) hydrateAll(invoices []model.LeaderInvoice) error {
	if len(invoices) == 0 {
		return nil
	}

	invoiceIDs := make([]string, 0, len(invoices))
	employeeIDs := map[string]bool{}
	bookingIDs := []string{}
	for i := range invoices {
		inv := &invoices[i]
		if inv.SystemsJSON != "" {
			_ = json.Unmarshal([]byte(inv.SystemsJSON), &inv.Systems)
		}
		if inv.ItemsJSON != "" {
			_ = json.Unmarshal([]byte(inv.ItemsJSON), &inv.Items)
		}
		inv.Materials = []model.LeaderInvoiceMaterialItem{}
		invoiceIDs = append(invoiceIDs, inv.ID)
		employeeIDs[inv.EmployeeID] = true
		if inv.ApprovedByEmployeeID != nil {
			employeeIDs[*inv.ApprovedByEmployeeID] = true
		}
		if inv.BookingID != nil {
			bookingIDs = append(bookingIDs, *inv.BookingID)
		}
	}

	// ① المواد — كلها بطلب واحد ثم نوزّعها على فواتيرها
	materials := []model.LeaderInvoiceMaterialItem{}
	if err := r.db.Select(&materials, `
		SELECT * FROM "LeaderInvoiceMaterialItem"
		WHERE "leaderInvoiceId" = ANY($1) ORDER BY "createdAt"`, pq.Array(invoiceIDs)); err != nil {
		return err
	}
	byInvoice := map[string][]model.LeaderInvoiceMaterialItem{}
	for _, m := range materials {
		byInvoice[m.LeaderInvoiceID] = append(byInvoice[m.LeaderInvoiceID], m)
	}

	// ② الحجوزات وما يتعلق بيها (زبون، خدمة، كادر منفّذ)
	bookings := map[string]*model.Booking{}
	assignmentsByBooking := map[string][]model.BookingAssignment{}
	if len(bookingIDs) > 0 {
		rows := []model.Booking{}
		if err := r.db.Select(&rows, `SELECT * FROM "Booking" WHERE id = ANY($1)`, pq.Array(bookingIDs)); err != nil {
			return err
		}
		customerIDs := []string{}
		serviceIDs := []string{}
		foundIDs := make([]string, 0, len(rows))
		for i := range rows {
			b := &rows[i]
			bookings[b.ID] = b
			foundIDs = append(foundIDs, b.ID)
			customerIDs = append(customerIDs, b.CustomerID)
			if b.ServiceID != nil {
				serviceIDs = append(serviceIDs, *b.ServiceID)
			}
		}

		customers := []model.Customer{}
		if err := r.db.Select(&customers, `SELECT * FROM "Customer" WHERE id = ANY($1)`, pq.Array(customerIDs)); err != nil {
			return err
		}
		customerByID := map[string]*model.Customer{}
		for i := range customers {
			customerByID[customers[i].ID] = &customers[i]
		}

		serviceByID := map[string]*model.Service{}
		if len(serviceIDs) > 0 {
			services := []model.Service{}
			if err := r.db.Select(&services, `SELECT * FROM "Service" WHERE id = ANY($1)`, pq.Array(serviceIDs)); err != nil {
				return err
			}
			for i := range services {
				serviceByID[services[i].ID] = &services[i]
			}
		}

		assignments := []model.BookingAssignment{}
		if err := r.db.Select(&assignments, `
			SELECT * FROM "BookingAssignment" WHERE "bookingId" = ANY($1)`, pq.Array(foundIDs)); err != nil {
			return err
		}
		for _, a := range assignments {
			employeeIDs[a.EmployeeID] = true
			assignmentsByBooking[a.BookingID] = append(assignmentsByBooking[a.BookingID], a)
		}

		for id, b := range bookings {
			b.Customer = customerByID[b.CustomerID]
			if b.ServiceID != nil {
				b.Service = serviceByID[*b.ServiceID]
			}
			b.Assignments = assignmentsByBooking[id]
		}
	}

	// ③ الموظفون — الليدرات والمعتمدون والكادر المنفّذ، كلهم بطلب واحد
	ids := make([]string, 0, len(employeeIDs))
	for id := range employeeIDs {
		ids = append(ids, id)
	}
	employees := []model.Employee{}
	if len(ids) > 0 {
		if err := r.db.Select(&employees, `SELECT * FROM "Employee" WHERE id = ANY($1)`, pq.Array(ids)); err != nil {
			return err
		}
	}
	employeeByID := map[string]*model.Employee{}
	for i := range employees {
		employeeByID[employees[i].ID] = &employees[i]
	}

	// الكادر المنفّذ ينلزق بتعييناته بعد ما صار عدنا كل الموظفين
	for _, list := range assignmentsByBooking {
		for i := range list {
			if e := employeeByID[list[i].EmployeeID]; e != nil {
				list[i].Employee = *e
			}
		}
	}

	// ④ الدمج النهائي
	for i := range invoices {
		inv := &invoices[i]
		if m := byInvoice[inv.ID]; m != nil {
			inv.Materials = m
		}
		if e := employeeByID[inv.EmployeeID]; e != nil {
			inv.EmployeeName = e.Name
			inv.EmployeeRole = e.Role
			inv.EmployeePhone = e.Phone
		}
		if inv.ApprovedByEmployeeID != nil {
			if e := employeeByID[*inv.ApprovedByEmployeeID]; e != nil {
				name := e.Name
				inv.ApprovedByName = &name
			}
		}
		if inv.BookingID != nil {
			if b := bookings[*inv.BookingID]; b != nil {
				code := b.Code
				inv.BookingCode = &code
				inv.Booking = b
			}
		}
	}
	return nil
}

// FindByExternalNumber يدوّر فاتورة برقم المحاسب الخارجي — هذا سبب
// أرشفة الرقم أصلاً: يلكاها بيه لمن يحتاجها.
func (r *LeaderInvoiceRepository) FindByExternalNumber(number string) (*model.LeaderInvoice, error) {
	var inv model.LeaderInvoice
	err := r.db.Get(&inv, `
		SELECT * FROM "LeaderInvoice"
		WHERE lower(btrim("externalInvoiceNumber")) = lower(btrim($1))
		LIMIT 1
	`, number)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := r.hydrate(&inv); err != nil {
		return nil, err
	}
	return &inv, nil
}

// SetExternalNumber يربط رقم الفاتورة المحاسبية بفاتورة معتمدة أصلاً —
// للفواتير الي انعتمدت قبل ما يصير الرقم إجبارياً.
func (r *LeaderInvoiceRepository) SetExternalNumber(id, number string) (*model.LeaderInvoice, error) {
	var inv model.LeaderInvoice
	err := r.db.Get(&inv, `
		UPDATE "LeaderInvoice"
		SET "externalInvoiceNumber" = $2, "externalInvoiceAt" = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING *
	`, id, number)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := r.hydrate(&inv); err != nil {
		return nil, err
	}
	return &inv, nil
}

// AdjustAmounts يعدّل مبالغ الفاتورة — للمحاسب حصراً.
// تقدير الإداري يطلع غلط أحياناً والفاتورة الي بيد الليدر هي الصح،
// فالمحاسب لازم يكدر يطابق. نخزن سبب التعديل حتى يبقى أثر.
func (r *LeaderInvoiceRepository) AdjustAmounts(id string, executionCost, materialsTotal, discountValue float64, reason string) (*model.LeaderInvoice, error) {
	net := executionCost + materialsTotal - discountValue
	if net < 0 {
		net = 0
	}
	var inv model.LeaderInvoice
	err := r.db.Get(&inv, `
		UPDATE "LeaderInvoice"
		SET "executionCost" = $2, "materialsTotal" = $3, "discountValue" = $4, "netTotal" = $5,
		    "adjustedReason" = $6, "adjustedAt" = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING *
	`, id, executionCost, materialsTotal, discountValue, net, reason)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := r.hydrate(&inv); err != nil {
		return nil, err
	}
	return &inv, nil
}
