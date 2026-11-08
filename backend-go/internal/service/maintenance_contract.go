package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ 🔄 عقود الصيانة الدورية (ع) 11-08 ═══
// زبون عنده عقد صيانة (كل ٣ أو ٦ أشهر…): قبل الموعد بأسبوع النظام يفتح حجز الزيارة
// لحاله ويوصل للإداري ينسّقه مثل أي حجز. دخل ثابت ما يعتمد على ذاكرة أحد.
// الإيقاف بدل الحذف (ممنوع حذف بلا موافقة المالك).

const ContractLeadDays = 7

type MaintenanceContract struct {
	ID              string     `db:"id" json:"id"`
	CustomerID      string     `db:"customerId" json:"customerId"`
	CustomerName    *string    `db:"customerName" json:"customerName"`
	CustomerCode    *string    `db:"customerCode" json:"customerCode"`
	ServiceID       string     `db:"serviceId" json:"serviceId"`
	ServiceName     *string    `db:"serviceName" json:"serviceName"`
	IntervalMonths  int        `db:"intervalMonths" json:"intervalMonths"`
	NextVisitDate   string     `db:"nextVisitDate" json:"nextVisitDate"`
	Address         *string    `db:"address" json:"address"`
	Notes           *string    `db:"notes" json:"notes"`
	Active          bool       `db:"active" json:"active"`
	LastBookingCode *string    `db:"lastBookingCode" json:"lastBookingCode"`
	LastGeneratedAt *time.Time `db:"lastGeneratedAt" json:"lastGeneratedAt"`
	LastError       *string    `db:"lastError" json:"lastError"`
	CreatedAt       time.Time  `db:"createdAt" json:"createdAt"`
}

type ContractInput struct {
	CustomerID     string  `json:"customerId"`
	ServiceID      string  `json:"serviceId"`
	IntervalMonths int     `json:"intervalMonths"`
	NextVisitDate  string  `json:"nextVisitDate"`
	Address        *string `json:"address"`
	Notes          *string `json:"notes"`
	Active         *bool   `json:"active"`
}

type MaintenanceContractService struct {
	db       *sqlx.DB
	bookings *BookingService
	notif    *repository.NotificationRepository
}

func NewMaintenanceContractService(db *sqlx.DB, b *BookingService, n *repository.NotificationRepository) *MaintenanceContractService {
	return &MaintenanceContractService{db: db, bookings: b, notif: n}
}

const contractSelect = `SELECT c.id, c."customerId", cu.name AS "customerName", cu."customerCode"::text AS "customerCode",
	c."serviceId", s.name AS "serviceName", c."intervalMonths", c."nextVisitDate"::text AS "nextVisitDate", c.address, c.notes,
	c.active, b.code AS "lastBookingCode", c."lastGeneratedAt", c."lastError", c."createdAt"
	FROM "MaintenanceContract" c
	LEFT JOIN "Customer" cu ON cu.id = c."customerId"
	LEFT JOIN "Service" s ON s.id = c."serviceId"
	LEFT JOIN "Booking" b ON b.id = c."lastBookingId"`

func (s *MaintenanceContractService) List() ([]MaintenanceContract, error) {
	rows := []MaintenanceContract{}
	err := s.db.Select(&rows, contractSelect+` ORDER BY c.active DESC, c."nextVisitDate"`)
	return rows, err
}

func validContract(in ContractInput, creating bool) error {
	if creating && (in.CustomerID == "" || in.ServiceID == "") {
		return errors.New("اختار الزبون والخدمة")
	}
	if (creating || in.IntervalMonths != 0) && (in.IntervalMonths < 1 || in.IntervalMonths > 24) {
		return errors.New("كل كم شهر؟ من ١ لحد ٢٤")
	}
	if creating || in.NextVisitDate != "" {
		if _, err := time.Parse("2006-01-02", in.NextVisitDate); err != nil {
			return errors.New("اختار موعد أول زيارة")
		}
	}
	return nil
}

func (s *MaintenanceContractService) Create(by string, in ContractInput) (string, error) {
	if err := validContract(in, true); err != nil {
		return "", err
	}
	var id string
	err := s.db.Get(&id, `INSERT INTO "MaintenanceContract" ("customerId", "serviceId", "intervalMonths", "nextVisitDate", address, notes, "createdById")
		VALUES ($1, $2, $3, $4::date, NULLIF(btrim($5), ''), NULLIF(btrim($6), ''), NULLIF($7, '')) RETURNING id`,
		in.CustomerID, in.ServiceID, in.IntervalMonths, in.NextVisitDate, strOr(in.Address), strOr(in.Notes), by)
	if err != nil && strings.Contains(err.Error(), "foreign key") {
		return "", errors.New("الزبون أو الخدمة مو موجودين")
	}
	return id, err
}

func (s *MaintenanceContractService) Update(id string, in ContractInput) error {
	if err := validContract(in, false); err != nil {
		return err
	}
	res, err := s.db.Exec(`UPDATE "MaintenanceContract" SET
		"intervalMonths" = COALESCE(NULLIF($2, 0), "intervalMonths"),
		"nextVisitDate" = COALESCE(NULLIF($3, '')::date, "nextVisitDate"),
		address = CASE WHEN $4::text IS NULL THEN address ELSE NULLIF(btrim($4), '') END,
		notes = CASE WHEN $5::text IS NULL THEN notes ELSE NULLIF(btrim($5), '') END,
		active = COALESCE($6, active)
		WHERE id = $1`, id, in.IntervalMonths, in.NextVisitDate, in.Address, in.Notes, in.Active)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("العقد مو موجود")
	}
	return nil
}

func strOr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// Start يفحص كل ٦ ساعات.
func (s *MaintenanceContractService) Start() {
	go func() {
		time.Sleep(4 * time.Minute)
		for {
			if n := s.GenerateDue(); n > 0 {
				log.Printf("[contracts] انفتح %d حجز صيانة دورية", n)
			}
			time.Sleep(6 * time.Hour)
		}
	}()
}

// GenerateDue يفتح حجز لكل عقد فعّال موعده خلال أسبوع، ويقدّم الموعد للزيارة الجاية.
func (s *MaintenanceContractService) GenerateDue() int {
	type due struct {
		ID         string  `db:"id"`
		CustomerID string  `db:"customerId"`
		ServiceID  string  `db:"serviceId"`
		Next       string  `db:"next"`
		Address    *string `db:"address"`
		Notes      *string `db:"notes"`
		Customer   *string `db:"customer"`
	}
	rows := []due{}
	if err := s.db.Select(&rows, `SELECT c.id, c."customerId", c."serviceId", c."nextVisitDate"::text AS next, c.address, c.notes, cu.name AS customer
		FROM "MaintenanceContract" c JOIN "Customer" cu ON cu.id = c."customerId"
		WHERE c.active AND c."nextVisitDate" <= baghdad_today() + $1::int`, ContractLeadDays); err != nil {
		log.Printf("[contracts] %v", err)
		return 0
	}
	n := 0
	for _, c := range rows {
		// نقدّم الموعد أول (بنفس الشرط) — حتى لو خادمين اشتغلوا سوه، حجز واحد بس
		res, err := s.db.Exec(`UPDATE "MaintenanceContract" SET "nextVisitDate" = "nextVisitDate" + make_interval(months => "intervalMonths"),
			"lastGeneratedAt" = now(), "lastError" = NULL WHERE id = $1 AND "nextVisitDate" = $2::date`, c.ID, c.Next)
		if err != nil {
			continue
		}
		if k, _ := res.RowsAffected(); k == 0 {
			continue
		}
		note := fmt.Sprintf("🔄 زيارة صيانة دورية حسب العقد — موعدها %s. تواصل ويا الزبون وثبّت الوقت.", c.Next)
		if c.Notes != nil && *c.Notes != "" {
			note += " ملاحظات العقد: " + *c.Notes
		}
		svc := c.ServiceID
		b, err := s.bookings.Create(model.CreateBookingRequest{CustomerID: c.CustomerID, ServiceID: &svc, Notes: &note, Address: c.Address})
		if err != nil {
			// ما انفتح (مثلاً الخدمة تحتاج عدد أجهزة) — نسجّل السبب وننبّه الإداري يفتحه بإيده
			msg := err.Error()
			_, _ = s.db.Exec(`UPDATE "MaintenanceContract" SET "lastError" = $2 WHERE id = $1`, c.ID, msg)
			if s.notif != nil {
				_ = s.notif.CreateForPermission("coordinator", "MAINTENANCE_CONTRACT",
					fmt.Sprintf("🔄 زيارة صيانة دورية مستحقة %s للزبون %s — ما گدرت أفتح الحجز لحالي (%s). افتحه من «حجز جديد».", c.Next, str(c.Customer), msg))
			}
			continue
		}
		_, _ = s.db.Exec(`UPDATE "MaintenanceContract" SET "lastBookingId" = $2 WHERE id = $1`, c.ID, b.ID)
		if s.notif != nil {
			_ = s.notif.CreateForPermission("coordinator", "MAINTENANCE_CONTRACT",
				fmt.Sprintf("🔄 انفتح حجز صيانة دورية %s للزبون %s — موعده %s. تواصل وثبّته.", b.Code, str(c.Customer), c.Next))
		}
		n++
	}
	return n
}
