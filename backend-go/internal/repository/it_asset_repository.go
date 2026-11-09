package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// أنواع الأجهزة وحالاتها — قائمة ثابتة، والواجهة تعرض الأسماء العربية.
var (
	ItAssetKinds    = map[string]bool{"COMPUTER": true, "LAPTOP": true, "SERVER": true, "NETWORK": true, "PRINTER": true, "CAMERA": true, "PHONE": true, "OTHER": true}
	ItAssetStatuses = map[string]bool{"ACTIVE": true, "REPAIR": true, "SPARE": true, "RETIRED": true}
	ItAssetLogKinds = map[string]bool{"NOTE": true, "REPAIR": true, "MAINTENANCE": true, "MOVE": true, "STATUS": true}
)

type ItAsset struct {
	ID                   string     `db:"id" json:"id"`
	Name                 string     `db:"name" json:"name"`
	Kind                 string     `db:"kind" json:"kind"`
	Status               string     `db:"status" json:"status"`
	Brand                *string    `db:"brand" json:"brand"`
	Model                *string    `db:"model" json:"model"`
	SerialNumber         *string    `db:"serialNumber" json:"serialNumber"`
	IPAddress            *string    `db:"ipAddress" json:"ipAddress"`
	Location             *string    `db:"location" json:"location"`
	AssignedEmployeeID   *string    `db:"assignedEmployeeId" json:"assignedEmployeeId"`
	AssignedEmployeeName *string    `db:"assignedEmployeeName" json:"assignedEmployeeName"`
	PurchaseDate         *time.Time `db:"purchaseDate" json:"purchaseDate"`
	WarrantyUntil        *time.Time `db:"warrantyUntil" json:"warrantyUntil"`
	Notes                *string    `db:"notes" json:"notes"`
	CreatedAt            time.Time  `db:"createdAt" json:"createdAt"`
	UpdatedAt            time.Time  `db:"updatedAt" json:"updatedAt"`
	LastLogAt            *time.Time `db:"lastLogAt" json:"lastLogAt"`
}

// ItAssetInput الحقول الي تنكتب من الشاشة — التواريخ نص YYYY-MM-DD.
type ItAssetInput struct {
	Name               string  `json:"name"`
	Kind               string  `json:"kind"`
	Status             string  `json:"status"`
	Brand              *string `json:"brand"`
	Model              *string `json:"model"`
	SerialNumber       *string `json:"serialNumber"`
	IPAddress          *string `json:"ipAddress"`
	Location           *string `json:"location"`
	AssignedEmployeeID *string `json:"assignedEmployeeId"`
	PurchaseDate       *string `json:"purchaseDate"`
	WarrantyUntil      *string `json:"warrantyUntil"`
	Notes              *string `json:"notes"`
}

type ItAssetLog struct {
	ID           string    `db:"id" json:"id"`
	AssetID      string    `db:"assetId" json:"assetId"`
	Kind         string    `db:"kind" json:"kind"`
	Note         string    `db:"note" json:"note"`
	Cost         *float64  `db:"cost" json:"cost"`
	EmployeeID   *string   `db:"employeeId" json:"employeeId"`
	EmployeeName *string   `db:"employeeName" json:"employeeName"`
	AssetName    *string   `db:"assetName" json:"assetName,omitempty"`
	CreatedAt    time.Time `db:"createdAt" json:"createdAt"`
}

type ItStats struct {
	Total          int            `json:"total"`
	ByKind         map[string]int `json:"byKind"`
	ByStatus       map[string]int `json:"byStatus"`
	WarrantySoon   []ItAsset      `json:"warrantySoon"`
	InRepair       []ItAsset      `json:"inRepair"`
	RepairCost90d  float64        `json:"repairCost90d"`
	RepairCount90d int            `json:"repairCount90d"`
	Unassigned     int            `json:"unassigned"`
	RecentLogs     []ItAssetLog   `json:"recentLogs"`
}

type ItAssetRepository struct{ db *sqlx.DB }

func NewItAssetRepository(db *sqlx.DB) *ItAssetRepository { return &ItAssetRepository{db: db} }

const itAssetSelect = `
	SELECT a.id, a.name, a.kind, a.status, a.brand, a.model, a."serialNumber", a."ipAddress", a.location,
	       a."assignedEmployeeId", e.name AS "assignedEmployeeName", a."purchaseDate", a."warrantyUntil",
	       a.notes, a."createdAt", a."updatedAt",
	       (SELECT MAX(l."createdAt") FROM "ItAssetLog" l WHERE l."assetId" = a.id) AS "lastLogAt"
	FROM "ItAsset" a
	LEFT JOIN "Employee" e ON e.id = a."assignedEmployeeId"`

func clean(p *string) *string {
	if p == nil {
		return nil
	}
	v := strings.TrimSpace(*p)
	if v == "" {
		return nil
	}
	return &v
}

func (in *ItAssetInput) normalize() error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return errors.New("اسم الجهاز مطلوب")
	}
	if in.Kind == "" {
		in.Kind = "COMPUTER"
	}
	if !ItAssetKinds[in.Kind] {
		return errors.New("نوع الجهاز غير معروف")
	}
	if in.Status == "" {
		in.Status = "ACTIVE"
	}
	if !ItAssetStatuses[in.Status] {
		return errors.New("حالة الجهاز غير معروفة")
	}
	for _, p := range []**string{&in.Brand, &in.Model, &in.SerialNumber, &in.IPAddress, &in.Location, &in.AssignedEmployeeID, &in.PurchaseDate, &in.WarrantyUntil, &in.Notes} {
		*p = clean(*p)
	}
	for _, d := range []*string{in.PurchaseDate, in.WarrantyUntil} {
		if d != nil {
			if _, err := time.Parse("2006-01-02", *d); err != nil {
				return errors.New("التاريخ لازم يكون بصيغة YYYY-MM-DD")
			}
		}
	}
	return nil
}

func (r *ItAssetRepository) List() ([]ItAsset, error) {
	out := []ItAsset{}
	err := r.db.Select(&out, itAssetSelect+` ORDER BY a.status = 'RETIRED', a.kind, a.name`)
	return out, err
}

func (r *ItAssetRepository) Get(id string) (*ItAsset, error) {
	var a ItAsset
	if err := r.db.Get(&a, itAssetSelect+` WHERE a.id = $1`, id); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *ItAssetRepository) Create(in ItAssetInput, byID string) (*ItAsset, error) {
	if err := in.normalize(); err != nil {
		return nil, err
	}
	var id string
	err := r.db.Get(&id, `
		INSERT INTO "ItAsset" (id, name, kind, status, brand, model, "serialNumber", "ipAddress", location,
			"assignedEmployeeId", "purchaseDate", "warrantyUntil", notes, "createdById")
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10::date, $11::date, $12, $13)
		RETURNING id`,
		in.Name, in.Kind, in.Status, in.Brand, in.Model, in.SerialNumber, in.IPAddress, in.Location,
		in.AssignedEmployeeID, in.PurchaseDate, in.WarrantyUntil, in.Notes, byID)
	if err != nil {
		return nil, err
	}
	return r.Get(id)
}

// Update يبدّل الجهاز، ولو تغيّرت حالته أو صاحبه ينكتب بالسجل تلقائياً —
// حتى تاريخ الجهاز ما يعتمد على إن الموظف يتذكر يكتب ملاحظة.
func (r *ItAssetRepository) Update(id string, in ItAssetInput, byID string) (*ItAsset, error) {
	if err := in.normalize(); err != nil {
		return nil, err
	}
	old, err := r.Get(id)
	if err != nil {
		return nil, err
	}
	if _, err := r.db.Exec(`
		UPDATE "ItAsset" SET name=$2, kind=$3, status=$4, brand=$5, model=$6, "serialNumber"=$7, "ipAddress"=$8,
			location=$9, "assignedEmployeeId"=$10, "purchaseDate"=$11::date, "warrantyUntil"=$12::date, notes=$13,
			"updatedAt"=now()
		WHERE id=$1`,
		id, in.Name, in.Kind, in.Status, in.Brand, in.Model, in.SerialNumber, in.IPAddress, in.Location,
		in.AssignedEmployeeID, in.PurchaseDate, in.WarrantyUntil, in.Notes); err != nil {
		return nil, err
	}
	if old.Status != in.Status {
		_, _ = r.AddLog(id, "STATUS", "تغيّرت الحالة: "+old.Status+" ← "+in.Status, nil, byID)
	}
	if strOr(old.AssignedEmployeeID) != strOr(in.AssignedEmployeeID) || strOr(old.Location) != strOr(in.Location) {
		_, _ = r.AddLog(id, "MOVE", "انتقل: "+orDash(old.Location)+" ← "+orDash(in.Location), nil, byID)
	}
	return r.Get(id)
}

func strOr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func orDash(p *string) string {
	if p == nil || *p == "" {
		return "—"
	}
	return *p
}

func (r *ItAssetRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM "ItAsset" WHERE id = $1`, id)
	return err
}

func (r *ItAssetRepository) AddLog(assetID, kind, note string, cost *float64, byID string) (*ItAssetLog, error) {
	note = strings.TrimSpace(note)
	if note == "" {
		return nil, errors.New("اكتب شنو انسوّى")
	}
	if kind == "" {
		kind = "NOTE"
	}
	if !ItAssetLogKinds[kind] {
		return nil, errors.New("نوع السجل غير معروف")
	}
	var l ItAssetLog
	err := r.db.Get(&l, `
		WITH ins AS (
			INSERT INTO "ItAssetLog" (id, "assetId", kind, note, cost, "employeeId")
			VALUES (gen_random_uuid()::text, $1, $2, $3, $4, NULLIF($5, ''))
			RETURNING *
		)
		SELECT ins.*, e.name AS "employeeName" FROM ins LEFT JOIN "Employee" e ON e.id = ins."employeeId"`,
		assetID, kind, note, cost, byID)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *ItAssetRepository) Logs(assetID string) ([]ItAssetLog, error) {
	out := []ItAssetLog{}
	err := r.db.Select(&out, `
		SELECT l.*, e.name AS "employeeName" FROM "ItAssetLog" l
		LEFT JOIN "Employee" e ON e.id = l."employeeId"
		WHERE l."assetId" = $1 ORDER BY l."createdAt" DESC`, assetID)
	return out, err
}

func (r *ItAssetRepository) Stats() (*ItStats, error) {
	s := &ItStats{ByKind: map[string]int{}, ByStatus: map[string]int{}}
	type kv struct {
		K string `db:"k"`
		N int    `db:"n"`
	}
	var rows []kv
	if err := r.db.Select(&rows, `SELECT kind AS k, COUNT(*) AS n FROM "ItAsset" GROUP BY kind`); err != nil {
		return nil, err
	}
	for _, x := range rows {
		s.ByKind[x.K] = x.N
		s.Total += x.N
	}
	rows = nil
	if err := r.db.Select(&rows, `SELECT status AS k, COUNT(*) AS n FROM "ItAsset" GROUP BY status`); err != nil {
		return nil, err
	}
	for _, x := range rows {
		s.ByStatus[x.K] = x.N
	}
	s.WarrantySoon = []ItAsset{}
	if err := r.db.Select(&s.WarrantySoon, itAssetSelect+`
		WHERE a.status <> 'RETIRED' AND a."warrantyUntil" IS NOT NULL
		  AND a."warrantyUntil" <= CURRENT_DATE + 60
		ORDER BY a."warrantyUntil" LIMIT 20`); err != nil {
		return nil, err
	}
	s.InRepair = []ItAsset{}
	if err := r.db.Select(&s.InRepair, itAssetSelect+` WHERE a.status = 'REPAIR' ORDER BY a."updatedAt" LIMIT 20`); err != nil {
		return nil, err
	}
	if err := r.db.Get(&s.Unassigned, `
		SELECT COUNT(*) FROM "ItAsset" WHERE status = 'ACTIVE' AND "assignedEmployeeId" IS NULL AND location IS NULL`); err != nil {
		return nil, err
	}
	var agg struct {
		N int     `db:"n"`
		C float64 `db:"c"`
	}
	if err := r.db.Get(&agg, `
		SELECT COUNT(*) AS n, COALESCE(SUM(cost), 0) AS c FROM "ItAssetLog"
		WHERE kind IN ('REPAIR', 'MAINTENANCE') AND "createdAt" >= now() - interval '90 days'`); err != nil {
		return nil, err
	}
	s.RepairCount90d, s.RepairCost90d = agg.N, agg.C
	s.RecentLogs = []ItAssetLog{}
	if err := r.db.Select(&s.RecentLogs, `
		SELECT l.*, e.name AS "employeeName", a.name AS "assetName"
		FROM "ItAssetLog" l
		JOIN "ItAsset" a ON a.id = l."assetId"
		LEFT JOIN "Employee" e ON e.id = l."employeeId"
		ORDER BY l."createdAt" DESC LIMIT 12`); err != nil {
		return nil, err
	}
	return s, nil
}
