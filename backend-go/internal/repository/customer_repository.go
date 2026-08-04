package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type CustomerRepository struct {
	db *sqlx.DB
}

func NewCustomerRepository(db *sqlx.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) List() ([]model.Customer, error) {
	customers := []model.Customer{}
	err := r.db.Select(&customers, `SELECT * FROM "Customer" ORDER BY "customerCode" ASC`)
	return customers, err
}

func (r *CustomerRepository) FindByPhone(phone string) (*model.Customer, error) {
	var c model.Customer
	err := r.db.Get(&c, `SELECT * FROM "Customer" WHERE phone = $1`, phone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// FindByCode يبحث عن زبون حسب كوده الرقمي (customerCode) — يستخدم بميزة صيانة
// الأجهزة العامة (كود الزبون => تعبئة الاسم/الهاتف/الموقع تلقائياً مثل الشيت).
func (r *CustomerRepository) FindByCode(code int) (*model.Customer, error) {
	var c model.Customer
	err := r.db.Get(&c, `SELECT * FROM "Customer" WHERE "customerCode" = $1`, code)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CustomerRepository) Create(name, phone string, location *string, lat, lng *float64) (*model.Customer, error) {
	var c model.Customer
	err := r.db.Get(&c, `
		INSERT INTO "Customer" (id, name, phone, location, "mapLatitude", "mapLongitude")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5)
		RETURNING *
	`, name, phone, location, lat, lng)
	return &c, err
}

func (r *CustomerRepository) FindByID(id string) (*model.Customer, error) {
	var c model.Customer
	err := r.db.Get(&c, `SELECT * FROM "Customer" WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// CountBookings يرجع عدد الحجوزات السابقة لزبون معيّن — يستخدم لتمييز "زبون جديد" عن
// "زبون قديم" وقت تسجيل حجز جديد له.
func (r *CustomerRepository) CountBookings(customerID string) (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM "Booking" WHERE "customerId" = $1`, customerID)
	return count, err
}

// UpdateLocation يحدّث موقع الزبون المحفوظ كلما ثبّت موظف المبيعات موقعاً جديداً بحجز،
// حتى تنترحل آخر نقطة معروفة تلقائياً بالمرة الجاية.
func (r *CustomerRepository) UpdateLocation(id string, location *string, lat, lng *float64) error {
	_, err := r.db.Exec(`
		UPDATE "Customer" SET
			location = COALESCE($2, location),
			"mapLatitude" = COALESCE($3, "mapLatitude"),
			"mapLongitude" = COALESCE($4, "mapLongitude")
		WHERE id = $1
	`, id, location, lat, lng)
	return err
}

// Update يعدّل بيانات الزبون — لتصحيح السجلات الخاطئة (اسم مو رباعي،
// رقم غلط، عنوان قديم). الحقول الاختيارية الفاضية تبقي القديم.
func (r *CustomerRepository) Update(id string, req model.UpdateCustomerRequest) (*model.Customer, error) {
	var c model.Customer
	err := r.db.Get(&c, `
		UPDATE "Customer" SET
			name = $2,
			phone = $3,
			location = COALESCE($4, location),
			"mapLatitude" = COALESCE($5, "mapLatitude"),
			"mapLongitude" = COALESCE($6, "mapLongitude"),
			"locationUrl" = COALESCE($7, "locationUrl"),
			position = COALESCE($8, position)
		WHERE id = $1
		RETURNING *
	`, id, req.Name, req.Phone, req.Location, req.MapLatitude, req.MapLongitude, req.LocationURL, req.Position)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &c, err
}

// ServiceTagsByCustomer يرجع خارطة customerId -> أسماء الخدمات الي طلبها، لتلوين قائمة الزبائن الكلية
func (r *CustomerRepository) ServiceTagsByCustomer() (map[string][]string, error) {
	rows := []struct {
		CustomerID string `db:"customerId"`
		Service    string `db:"service"`
	}{}
	if err := r.db.Select(&rows, `SELECT "customerId", service FROM "CustomerServiceTag"`); err != nil {
		return nil, err
	}
	out := map[string][]string{}
	for _, row := range rows {
		out[row.CustomerID] = append(out[row.CustomerID], row.Service)
	}
	return out, nil
}

// EnsureServiceTag يضيف وسم خدمة للزبون (جي بي اس، كاميرات...) إذا ما كان موجود أصلاً —
// يستدعى تلقائياً كل ما يسجل حجز بخدمة معينة لزبون، حتى نعرف بعدين شنو الخدمات الي طلبها كلها.
func (r *CustomerRepository) EnsureServiceTag(customerID string, serviceID *string) error {
	if serviceID == nil {
		return nil
	}
	_, err := r.db.Exec(`
		INSERT INTO "CustomerServiceTag" (id, "customerId", service, "createdAt")
		SELECT gen_random_uuid()::text, $1, s.name, now()
		FROM "Service" s WHERE s.id = $2
		ON CONFLICT ("customerId", service) DO NOTHING
	`, customerID, *serviceID)
	return err
}

// ListGpsCustomers يرجع كل زبون موسوم بـ"GPS" مع معلوماته الخاصة (رقم الجهاز، تاريخ انتهاء الاشتراك)
func (r *CustomerRepository) ListGpsCustomers() ([]model.CustomerGpsResponse, error) {
	rows := []struct {
		model.Customer
		GpsNumber       *string    `db:"gpsNumber"`
		DeviceID        *string    `db:"deviceId"`
		SubscriptionEnd *time.Time `db:"subscriptionEnd"`
	}{}
	err := r.db.Select(&rows, `
		SELECT c.*, gi."gpsNumber", gi."deviceId", gi."subscriptionEnd"
		FROM "Customer" c
		JOIN "CustomerServiceTag" t ON t."customerId" = c.id AND t.service = 'GPS'
		LEFT JOIN "CustomerGpsInfo" gi ON gi."customerId" = c.id
		ORDER BY c."customerCode" ASC
	`)
	if err != nil {
		return nil, err
	}
	out := make([]model.CustomerGpsResponse, len(rows))
	for i, row := range rows {
		out[i] = model.CustomerGpsResponse{
			CustomerResponse: row.Customer.ToResponse(),
			GpsNumber:        row.GpsNumber,
			DeviceID:         row.DeviceID,
			SubscriptionEnd:  row.SubscriptionEnd,
		}
	}
	return out, nil
}

// FindOrCreateByPhone يلاقي الزبون برقم هاتفه أو ينشئه إذا مو موجود.
// نستعمله لما ينضاف مشروع بدون حجز مرتبط — لازم يكون اكو سجل زبون حتى
// نقدر نرحّله للشخصيات المهمة.
func (r *CustomerRepository) FindOrCreateByPhone(phone, name string) (*model.Customer, error) {
	var c model.Customer
	if err := r.db.Get(&c, `SELECT * FROM "Customer" WHERE phone = $1 LIMIT 1`, phone); err == nil {
		return &c, nil
	}
	if name == "" {
		name = phone
	}
	// customerCode عمود رقمي — ناخذ أكبر رقم موجود ونزيد واحد
	err := r.db.Get(&c, `
		INSERT INTO "Customer" (id, "customerCode", name, phone)
		VALUES (gen_random_uuid()::text,
			COALESCE((SELECT MAX("customerCode") FROM "Customer"), 0) + 1,
			$1, $2)
		RETURNING *`, name, phone)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
