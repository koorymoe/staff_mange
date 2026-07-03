package repository

import (
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

type GpsRepository struct {
	db *sqlx.DB
}

func NewGpsRepository(db *sqlx.DB) *GpsRepository {
	return &GpsRepository{db: db}
}

func (r *GpsRepository) loadEmployeeBrief(id string) *model.EmployeeBrief {
	var brief model.EmployeeBrief
	if err := r.db.Get(&brief, `SELECT id, name FROM "Employee" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &brief
}

func (r *GpsRepository) loadCustomer(id string) *model.GpsCustomer {
	var c model.GpsCustomer
	if err := r.db.Get(&c, `SELECT * FROM "GpsCustomer" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &c
}

func (r *GpsRepository) loadSimCard(id string) *model.SimCard {
	var s model.SimCard
	if err := r.db.Get(&s, `SELECT * FROM "SimCard" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &s
}

func (r *GpsRepository) loadDeviceRequestBare(id string) *model.GpsDeviceRequest {
	var d model.GpsDeviceRequest
	if err := r.db.Get(&d, `SELECT * FROM "GpsDeviceRequest" WHERE id = $1`, id); err != nil {
		return nil
	}
	return &d
}

// ── GPS Customers ────────────────────────────────────────────────────────────

func (r *GpsRepository) ListCustomers() ([]model.GpsCustomer, error) {
	customers := []model.GpsCustomer{}
	err := r.db.Select(&customers, `SELECT * FROM "GpsCustomer" ORDER BY "fullName" ASC`)
	return customers, err
}

func (r *GpsRepository) CreateCustomer(req model.UpsertGpsCustomerRequest) (*model.GpsCustomer, error) {
	var c model.GpsCustomer
	err := r.db.Get(&c, `
		INSERT INTO "GpsCustomer" (id, "fullName", "fatherName", "grandfatherName", phone, address, governorate, "idCardFrontUrl", "idCardBackUrl", "residenceCardFrontUrl", "residenceCardBackUrl")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING *
	`, req.FullName, req.FatherName, req.GrandfatherName, req.Phone, req.Address, req.Governorate, req.IDCardFrontURL, req.IDCardBackURL, req.ResidenceCardFrontURL, req.ResidenceCardBackURL)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *GpsRepository) UpdateCustomer(id string, req model.UpsertGpsCustomerRequest) (*model.GpsCustomer, error) {
	var c model.GpsCustomer
	err := r.db.Get(&c, `
		UPDATE "GpsCustomer" SET
			"fullName" = COALESCE($2, "fullName"),
			"fatherName" = COALESCE($3, "fatherName"),
			"grandfatherName" = COALESCE($4, "grandfatherName"),
			phone = COALESCE($5, phone),
			address = COALESCE($6, address),
			governorate = COALESCE($7, governorate),
			"idCardFrontUrl" = COALESCE($8, "idCardFrontUrl"),
			"idCardBackUrl" = COALESCE($9, "idCardBackUrl"),
			"residenceCardFrontUrl" = COALESCE($10, "residenceCardFrontUrl"),
			"residenceCardBackUrl" = COALESCE($11, "residenceCardBackUrl")
		WHERE id = $1
		RETURNING *
	`, id, req.FullName, req.FatherName, req.GrandfatherName, req.Phone, req.Address, req.Governorate, req.IDCardFrontURL, req.IDCardBackURL, req.ResidenceCardFrontURL, req.ResidenceCardBackURL)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ── SIM Cards ────────────────────────────────────────────────────────────────

func (r *GpsRepository) ListSims() ([]model.SimCard, error) {
	sims := []model.SimCard{}
	err := r.db.Select(&sims, `SELECT * FROM "SimCard" ORDER BY "createdAt" DESC`)
	return sims, err
}

func (r *GpsRepository) CreateSim(req model.UpsertSimCardRequest) (*model.SimCard, error) {
	status := "AVAILABLE"
	if req.Status != nil {
		status = *req.Status
	}
	var s model.SimCard
	err := r.db.Get(&s, `
		INSERT INTO "SimCard" (id, "simNumber", iccid, operator, status, "customerId", notes)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6)
		RETURNING *
	`, req.SimNumber, req.ICCID, req.Operator, status, req.CustomerID, req.Notes)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *GpsRepository) UpdateSim(id string, req model.UpsertSimCardRequest) (*model.SimCard, error) {
	var s model.SimCard
	err := r.db.Get(&s, `
		UPDATE "SimCard" SET
			"simNumber" = COALESCE($2, "simNumber"),
			iccid = COALESCE($3, iccid),
			operator = COALESCE($4, operator),
			status = COALESCE($5, status),
			"customerId" = COALESCE($6, "customerId"),
			notes = COALESCE($7, notes)
		WHERE id = $1
		RETURNING *
	`, id, req.SimNumber, req.ICCID, req.Operator, req.Status, req.CustomerID, req.Notes)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// ── Device Requests ──────────────────────────────────────────────────────────

func (r *GpsRepository) hydrateDevice(d *model.GpsDeviceRequest) {
	d.Customer = r.loadCustomer(d.CustomerID)
	d.Employee = r.loadEmployeeBrief(d.EmployeeID)
	if d.SimCardID != nil {
		d.SimCard = r.loadSimCard(*d.SimCardID)
	}
}

func (r *GpsRepository) ListDevices() ([]model.GpsDeviceRequest, error) {
	devices := []model.GpsDeviceRequest{}
	if err := r.db.Select(&devices, `SELECT * FROM "GpsDeviceRequest" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	for i := range devices {
		r.hydrateDevice(&devices[i])
	}
	return devices, nil
}

func (r *GpsRepository) CreateDevice(req model.UpsertGpsDeviceRequest) (*model.GpsDeviceRequest, error) {
	var d model.GpsDeviceRequest
	err := r.db.Get(&d, `
		INSERT INTO "GpsDeviceRequest" (id, "customerId", "employeeId", "adminId", "purchaseType", "subscriptionType", "subscriptionStart", "subscriptionEnd", "simCardId", notes, "invoicePhotoUrl", "gpsNumber", "residenceCardNumber")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING *
	`, req.CustomerID, req.EmployeeID, req.AdminID, req.PurchaseType, req.SubscriptionType, req.SubscriptionStart, req.SubscriptionEnd, req.SimCardID, req.Notes, req.InvoicePhotoURL, req.GpsNumber, req.ResidenceCardNumber)
	if err != nil {
		return nil, err
	}
	r.hydrateDevice(&d)
	return &d, nil
}

func (r *GpsRepository) UpdateDevice(id string, req model.UpsertGpsDeviceRequest) (*model.GpsDeviceRequest, error) {
	var d model.GpsDeviceRequest
	err := r.db.Get(&d, `
		UPDATE "GpsDeviceRequest" SET
			"customerId" = COALESCE($2, "customerId"),
			"employeeId" = COALESCE($3, "employeeId"),
			"adminId" = COALESCE($4, "adminId"),
			"purchaseType" = COALESCE($5, "purchaseType"),
			"subscriptionType" = COALESCE($6, "subscriptionType"),
			"subscriptionStart" = COALESCE($7, "subscriptionStart"),
			"subscriptionEnd" = COALESCE($8, "subscriptionEnd"),
			"subscriptionStatus" = COALESCE($9, "subscriptionStatus"),
			status = COALESCE($10, status),
			"simCardId" = COALESCE($11, "simCardId"),
			notes = COALESCE($12, notes),
			"isChecked" = COALESCE($13, "isChecked"),
			"isActivated" = COALESCE($14, "isActivated"),
			"isDelivered" = COALESCE($15, "isDelivered"),
			"invoicePhotoUrl" = COALESCE($16, "invoicePhotoUrl"),
			"gpsNumber" = COALESCE($17, "gpsNumber"),
			"residenceCardNumber" = COALESCE($18, "residenceCardNumber"),
			"activationDate" = COALESCE($19, "activationDate"),
			"deliveredAt" = COALESCE($20, "deliveredAt")
		WHERE id = $1
		RETURNING *
	`, id, req.CustomerID, req.EmployeeID, req.AdminID, req.PurchaseType, req.SubscriptionType,
		req.SubscriptionStart, req.SubscriptionEnd, req.SubscriptionStatus, req.Status, req.SimCardID, req.Notes,
		req.IsChecked, req.IsActivated, req.IsDelivered, req.InvoicePhotoURL, req.GpsNumber, req.ResidenceCardNumber,
		req.ActivationDate, req.DeliveredAt)
	if err != nil {
		return nil, err
	}
	r.hydrateDevice(&d)
	return &d, nil
}

// ── Renewals ─────────────────────────────────────────────────────────────────

func (r *GpsRepository) hydrateRenewal(rn *model.GpsRenewalRequest) {
	rn.Customer = r.loadCustomer(rn.CustomerID)
	device := r.loadDeviceRequestBare(rn.DeviceRequestID)
	if device != nil {
		r.hydrateDevice(device)
	}
	rn.DeviceRequest = device
}

func (r *GpsRepository) ListRenewals() ([]model.GpsRenewalRequest, error) {
	renewals := []model.GpsRenewalRequest{}
	if err := r.db.Select(&renewals, `SELECT * FROM "GpsRenewalRequest" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	for i := range renewals {
		r.hydrateRenewal(&renewals[i])
	}
	return renewals, nil
}

func (r *GpsRepository) CreateRenewal(req model.UpsertGpsRenewalRequest) (*model.GpsRenewalRequest, error) {
	var rn model.GpsRenewalRequest
	err := r.db.Get(&rn, `
		INSERT INTO "GpsRenewalRequest" (id, "customerId", "deviceRequestId", "employeeId", "adminId", "subscriptionType", "newEndDate")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6)
		RETURNING *
	`, req.CustomerID, req.DeviceRequestID, req.EmployeeID, req.AdminID, req.SubscriptionType, req.NewEndDate)
	if err != nil {
		return nil, err
	}
	r.hydrateRenewal(&rn)
	return &rn, nil
}

func (r *GpsRepository) UpdateRenewal(id string, req model.UpsertGpsRenewalRequest) (*model.GpsRenewalRequest, error) {
	var rn model.GpsRenewalRequest
	err := r.db.Get(&rn, `
		UPDATE "GpsRenewalRequest" SET
			"customerId" = COALESCE($2, "customerId"),
			"deviceRequestId" = COALESCE($3, "deviceRequestId"),
			"employeeId" = COALESCE($4, "employeeId"),
			"adminId" = COALESCE($5, "adminId"),
			"subscriptionType" = COALESCE($6, "subscriptionType"),
			"newEndDate" = COALESCE($7, "newEndDate"),
			status = COALESCE($8, status)
		WHERE id = $1
		RETURNING *
	`, id, req.CustomerID, req.DeviceRequestID, req.EmployeeID, req.AdminID, req.SubscriptionType, req.NewEndDate, req.Status)
	if err != nil {
		return nil, err
	}
	r.hydrateRenewal(&rn)
	return &rn, nil
}

// ── Maintenance ──────────────────────────────────────────────────────────────

func (r *GpsRepository) hydrateMaintenance(m *model.GpsMaintenanceRequest) {
	m.Customer = r.loadCustomer(m.CustomerID)
	m.Employee = r.loadEmployeeBrief(m.EmployeeID)
}

func (r *GpsRepository) ListMaintenance() ([]model.GpsMaintenanceRequest, error) {
	records := []model.GpsMaintenanceRequest{}
	if err := r.db.Select(&records, `SELECT * FROM "GpsMaintenanceRequest" ORDER BY "createdAt" DESC`); err != nil {
		return nil, err
	}
	for i := range records {
		r.hydrateMaintenance(&records[i])
	}
	return records, nil
}

func (r *GpsRepository) CreateMaintenance(req model.UpsertGpsMaintenanceRequest) (*model.GpsMaintenanceRequest, error) {
	var m model.GpsMaintenanceRequest
	err := r.db.Get(&m, `
		INSERT INTO "GpsMaintenanceRequest" (id, "customerId", "employeeId", "adminId", "problemDescription", "adminNotes")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5)
		RETURNING *
	`, req.CustomerID, req.EmployeeID, req.AdminID, req.ProblemDescription, req.AdminNotes)
	if err != nil {
		return nil, err
	}
	r.hydrateMaintenance(&m)
	return &m, nil
}

func (r *GpsRepository) UpdateMaintenance(id string, req model.UpsertGpsMaintenanceRequest) (*model.GpsMaintenanceRequest, error) {
	var m model.GpsMaintenanceRequest
	err := r.db.Get(&m, `
		UPDATE "GpsMaintenanceRequest" SET
			"customerId" = COALESCE($2, "customerId"),
			"employeeId" = COALESCE($3, "employeeId"),
			"adminId" = COALESCE($4, "adminId"),
			"problemDescription" = COALESCE($5, "problemDescription"),
			status = COALESCE($6, status),
			"adminNotes" = COALESCE($7, "adminNotes"),
			"resolvedAt" = COALESCE($8, "resolvedAt")
		WHERE id = $1
		RETURNING *
	`, id, req.CustomerID, req.EmployeeID, req.AdminID, req.ProblemDescription, req.Status, req.AdminNotes, req.ResolvedAt)
	if err != nil {
		return nil, err
	}
	r.hydrateMaintenance(&m)
	return &m, nil
}

// ── Settings (prices) ────────────────────────────────────────────────────────

func (r *GpsRepository) ListSettings() ([]model.GpsSubscriptionPrice, error) {
	prices := []model.GpsSubscriptionPrice{}
	err := r.db.Select(&prices, `SELECT * FROM "GpsSubscriptionPrice"`)
	return prices, err
}

func (r *GpsRepository) UpsertSetting(id, subscriptionType string, price float64) (*model.GpsSubscriptionPrice, error) {
	var p model.GpsSubscriptionPrice
	err := r.db.Get(&p, `
		INSERT INTO "GpsSubscriptionPrice" (id, "subscriptionType", price)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET "subscriptionType" = $2, price = $3
		RETURNING *
	`, id, subscriptionType, price)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ── Stats ────────────────────────────────────────────────────────────────────

func (r *GpsRepository) Stats() (*model.GpsStats, error) {
	byStatus := []model.GpsStatusCount{}
	if err := r.db.Select(&byStatus, `SELECT status, COUNT(id) as count FROM "GpsDeviceRequest" GROUP BY status`); err != nil {
		return nil, err
	}
	var totalDevices, totalCustomers, totalSims, availableSims int
	if err := r.db.Get(&totalDevices, `SELECT COUNT(*) FROM "GpsDeviceRequest"`); err != nil {
		return nil, err
	}
	if err := r.db.Get(&totalCustomers, `SELECT COUNT(*) FROM "GpsCustomer"`); err != nil {
		return nil, err
	}
	if err := r.db.Get(&totalSims, `SELECT COUNT(*) FROM "SimCard"`); err != nil {
		return nil, err
	}
	if err := r.db.Get(&availableSims, `SELECT COUNT(*) FROM "SimCard" WHERE status = 'AVAILABLE'`); err != nil {
		return nil, err
	}

	return &model.GpsStats{
		DevicesByStatus: byStatus,
		TotalDevices:    totalDevices,
		TotalCustomers:  totalCustomers,
		TotalSims:       totalSims,
		AvailableSims:   availableSims,
		InUseSims:       totalSims - availableSims,
	}, nil
}
