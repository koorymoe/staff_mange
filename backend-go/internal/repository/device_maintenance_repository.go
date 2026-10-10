package repository

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// timeNowPtr يرجع مؤشر للوقت الحالي (UTC) — مستخدم لتعيين receivedAt/deliveredAt
// أول ما يتحقق الفعل (استلام/تسليم الجهاز)، مرة وحدة بس (idempotent بالخدمة).
func timeNowPtr() *time.Time {
	now := time.Now().UTC()
	return &now
}

type DeviceMaintenanceRepository struct {
	db *sqlx.DB
}

func NewDeviceMaintenanceRepository(db *sqlx.DB) *DeviceMaintenanceRepository {
	return &DeviceMaintenanceRepository{db: db}
}

func (r *DeviceMaintenanceRepository) loadEmployeeBrief(id string) *model.EmployeeBrief {
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name, position FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &brief
}

func (r *DeviceMaintenanceRepository) loadCustomer(id string) *model.CustomerResponse {
	var c model.Customer
	if err := r.db.Get(&c, `SELECT * FROM "Customer" WHERE id = $1`, id); err != nil {
		return nil
	}
	resp := c.ToResponse()
	return &resp
}

func (r *DeviceMaintenanceRepository) hydrate(t *model.DeviceMaintenanceTicket) {
	t.Customer = r.loadCustomer(t.CustomerID)
	t.Employee = r.loadEmployeeBrief(t.EmployeeID)
	t.Status = t.DeriveStatus()
}

// nextInvoiceNumber يولّد رقم فاتورة صيانة فريد بالاعتماد على تسلسل قاعدة البيانات
// (DeviceMaintenanceTicket_invoiceSeq)، بصيغة MT-00001 مطابقة لنمط أكواد الزبائن CUST-00001.
func (r *DeviceMaintenanceRepository) nextInvoiceNumber() (string, error) {
	var n int64
	if err := r.db.Get(&n, `SELECT nextval('"DeviceMaintenanceTicket_invoiceSeq"')`); err != nil {
		return "", err
	}
	return fmt.Sprintf("MT-%05d", n), nil
}

func (r *DeviceMaintenanceRepository) Create(employeeID, customerID string, req model.CreateDeviceMaintenanceTicketRequest) (*model.DeviceMaintenanceTicket, error) {
	invoiceNumber, err := r.nextInvoiceNumber()
	if err != nil {
		return nil, err
	}
	var t model.DeviceMaintenanceTicket
	err = r.db.Get(&t, `
		INSERT INTO "DeviceMaintenanceTicket"
			(id, "appointmentDate", "customerId", "deviceTypeName", problem, "deviceSerial", "invoiceNumber", "employeeId")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7)
		RETURNING *
	`, req.AppointmentDate, customerID, req.DeviceTypeName, req.Problem, req.DeviceSerial, invoiceNumber, employeeID)
	if err != nil {
		return nil, err
	}
	r.hydrate(&t)
	return &t, nil
}

func (r *DeviceMaintenanceRepository) List() ([]model.DeviceMaintenanceTicket, error) {
	tickets := []model.DeviceMaintenanceTicket{}
	if err := r.db.Select(&tickets, `SELECT * FROM "DeviceMaintenanceTicket" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	for i := range tickets {
		r.hydrate(&tickets[i])
	}
	return tickets, nil
}

func (r *DeviceMaintenanceRepository) GetByID(id string) (*model.DeviceMaintenanceTicket, error) {
	var t model.DeviceMaintenanceTicket
	if err := r.db.Get(&t, `SELECT * FROM "DeviceMaintenanceTicket" WHERE id = $1`, id); err != nil {
		return nil, err
	}
	r.hydrate(&t)
	return &t, nil
}

func (r *DeviceMaintenanceRepository) Update(id string, req model.UpdateDeviceMaintenanceTicketRequest) (*model.DeviceMaintenanceTicket, error) {
	current, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	appointmentDate := current.AppointmentDate
	if req.AppointmentDate != nil {
		appointmentDate = req.AppointmentDate
	}
	deviceTypeName := current.DeviceTypeName
	if req.DeviceTypeName != nil {
		deviceTypeName = *req.DeviceTypeName
	}
	problem := current.Problem
	if req.Problem != nil {
		problem = *req.Problem
	}
	deviceSerial := current.DeviceSerial
	if req.DeviceSerial != nil {
		deviceSerial = req.DeviceSerial
	}
	receivedAt := current.ReceivedAt
	if req.MarkReceived && receivedAt == nil {
		receivedAt = timeNowPtr()
	}
	deliveredAt := current.DeliveredAt
	if req.MarkDelivered && deliveredAt == nil {
		deliveredAt = timeNowPtr()
	}

	var t model.DeviceMaintenanceTicket
	err = r.db.Get(&t, `
		UPDATE "DeviceMaintenanceTicket"
		SET "appointmentDate" = $2, "deviceTypeName" = $3, problem = $4, "deviceSerial" = $5,
			"receivedAt" = $6, "deliveredAt" = $7
		WHERE id = $1
		RETURNING *
	`, id, appointmentDate, deviceTypeName, problem, deviceSerial, receivedAt, deliveredAt)
	if err != nil {
		return nil, err
	}
	r.hydrate(&t)
	return &t, nil
}
