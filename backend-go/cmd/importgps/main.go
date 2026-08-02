// أداة استيراد لمرة واحدة: تنقل بيانات الجي بي اس القديمة (الي كانت بإكسل)
// لقاعدة البيانات — الشرائح والزبائن والأجهزة، مربوطين ببعض.
//
// الاستخدام:
//
//	python3 extract_gps_excel.py بيانات_GPS.xlsx الشرائح.xlsx > gps_data.json
//	go run ./cmd/importgps gps_data.json
//
// الاستيراد آمن للتكرار: أي شريحة برقمها موجودة أصلاً، أو زبون بنفس الاسم
// والهاتف، يتم تحديثه بدل ما ينضاف مرتين — فتقدر تشغّله أكثر من مرة بأمان.
//
// الربط: عمود «رقم ال GPS» بملف الزبائن = عمود SUBNO بملف الشرائح. هذا الي
// يخلينا نعرف أي شريحة بيد أي زبون بدل ما نخمّن.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	"staffmange-api/internal/config"
	"staffmange-api/internal/database"
)

type simRow struct {
	SimNumber string  `json:"simNumber"`
	ICCID     *string `json:"iccid"`
	Operator  string  `json:"operator"`
	Status    string  `json:"status"`
	RawStatus *string `json:"rawStatus"`
}

type customerRow struct {
	FullName        string   `json:"fullName"`
	Phone           *string  `json:"phone"`
	GpsNumber       *string  `json:"gpsNumber"`
	DeviceImei      *string  `json:"deviceImei"`
	SubscriptionEnd *string  `json:"subscriptionEnd"`
	InstalledAt     *string  `json:"installedAt"`
	RequestedAt     *string  `json:"requestedAt"`
	InstallerName   *string  `json:"installerName"`
	InstallCost     *float64 `json:"installCost"`
	VehicleType     *string  `json:"vehicleType"`
	InstallNote     *string  `json:"installNote"`
	Status          *string  `json:"status"`
	NotifyStage     *float64 `json:"notifyStage"`
	Notified1       bool     `json:"notified1"`
	Notified2       bool     `json:"notified2"`
	Notified40      bool     `json:"notified40"`
	SimBurned       bool     `json:"simBurned"`
}

type payload struct {
	Sims      []simRow      `json:"sims"`
	Customers []customerRow `json:"customers"`
}

func parseDate(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("الاستخدام: go run ./cmd/importgps gps_data.json")
	}
	_ = godotenv.Load()
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("تعذر الاتصال بقاعدة البيانات: %v", err)
	}
	defer db.Close()

	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("تعذر قراءة الملف: %v", err)
	}
	var p payload
	if err := json.Unmarshal(raw, &p); err != nil {
		log.Fatalf("ملف JSON غير صالح: %v", err)
	}

	// أي موظف صالح كصاحب الطلب — عمود employeeId إلزامي بجدول الطلبات،
	// وبيانات الإكسل ما بيها معرّف موظف، بس بيها اسم المنفّذ (نخزنه بالملاحظات).
	var ownerID string
	if err := db.Get(&ownerID, `SELECT id FROM "Employee" ORDER BY "createdAt" ASC LIMIT 1`); err != nil {
		log.Fatalf("ما اكو ولا موظف بقاعدة البيانات: %v", err)
	}

	simIDs := importSims(db, p.Sims)
	importCustomers(db, p.Customers, simIDs, ownerID)
}

// importSims يستورد الشرائح ويرجّع خريطة رقم الشريحة → معرّفها.
func importSims(db *sqlx.DB, sims []simRow) map[string]string {
	ids := map[string]string{}
	added, updated := 0, 0
	for _, s := range sims {
		var id string
		err := db.Get(&id, `SELECT id FROM "SimCard" WHERE "simNumber" = $1`, s.SimNumber)
		if err == nil {
			// موجودة أصلاً — نحدّث حالتها بس، ما نلمس ارتباطها بزبون
			if _, err := db.Exec(`UPDATE "SimCard" SET iccid = COALESCE($2, iccid), operator = $3::"SimOperator" WHERE id = $1`,
				id, s.ICCID, s.Operator); err != nil {
				log.Printf("تحذير: تعذر تحديث الشريحة %s: %v", s.SimNumber, err)
			}
			updated++
		} else {
			notes := ""
			if s.RawStatus != nil {
				notes = "الحالة بالإكسل: " + *s.RawStatus
			}
			if err := db.Get(&id, `
				INSERT INTO "SimCard" (id, "simNumber", iccid, operator, status, notes)
				VALUES (gen_random_uuid()::text, $1, $2, $3::"SimOperator", $4::"SimStatus", NULLIF($5,''))
				RETURNING id`, s.SimNumber, s.ICCID, s.Operator, s.Status, notes); err != nil {
				log.Printf("تحذير: تعذرت إضافة الشريحة %s: %v", s.SimNumber, err)
				continue
			}
			added++
		}
		ids[s.SimNumber] = id
	}
	log.Printf("الشرائح: %d جديدة، %d محدّثة، المجموع %d", added, updated, len(ids))
	return ids
}

func importCustomers(db *sqlx.DB, rows []customerRow, simIDs map[string]string, ownerID string) {
	added, linked, devices := 0, 0, 0
	for _, c := range rows {
		phone := ""
		if c.Phone != nil {
			phone = *c.Phone
		}

		// نطابق بالاسم+الهاتف حتى إعادة التشغيل ما تكرّر الزبون
		var custID string
		err := db.Get(&custID, `SELECT id FROM "GpsCustomer" WHERE "fullName" = $1 AND COALESCE(phone,'') = $2`, c.FullName, phone)
		if err != nil {
			if err := db.Get(&custID, `
				INSERT INTO "GpsCustomer" (id, "fullName", phone)
				VALUES (gen_random_uuid()::text, $1, $2) RETURNING id`, c.FullName, phone); err != nil {
				log.Printf("تحذير: تعذرت إضافة الزبون %s: %v", c.FullName, err)
				continue
			}
			added++
		}

		// الشريحة: نربطها بالزبون عن طريق رقم الجي بي اس
		var simID *string
		if c.GpsNumber != nil {
			if id, ok := simIDs[*c.GpsNumber]; ok {
				simID = &id
				// الشريحة المحروقة تبقى محروقة — ما نرجّعها IN_USE
				if _, err := db.Exec(`
					UPDATE "SimCard" SET "customerId" = $2, "assignedAt" = COALESCE("assignedAt", now()),
						status = CASE WHEN status = 'BURNED' THEN status ELSE 'IN_USE'::"SimStatus" END
					WHERE id = $1`, id, custID); err != nil {
					log.Printf("تحذير: تعذر ربط الشريحة %s: %v", *c.GpsNumber, err)
				} else {
					linked++
				}
			}
		}

		// الجهاز/الاشتراك
		notes := []string{}
		if c.InstallerName != nil {
			notes = append(notes, "منفّذ الشد: "+*c.InstallerName)
		}
		// التكلفة تنخزن بصيغة ثابتة حتى حساب تكاليف الشد يقدر يستخرجها
		if c.InstallCost != nil && *c.InstallCost > 0 {
			notes = append(notes, fmt.Sprintf("تكلفة الشد: %.0f", *c.InstallCost))
		}
		if c.VehicleType != nil {
			notes = append(notes, "نوع المركبة: "+*c.VehicleType)
		}
		if c.InstallNote != nil {
			notes = append(notes, *c.InstallNote)
		}
		if c.Status != nil {
			notes = append(notes, "الحالة بالإكسل: "+*c.Status)
		}
		noteText := ""
		for i, n := range notes {
			if i > 0 {
				noteText += " · "
			}
			noteText += n
		}

		var existing string
		dupErr := db.Get(&existing, `SELECT id FROM "GpsDeviceRequest" WHERE "customerId" = $1 AND COALESCE("gpsNumber",'') = COALESCE($2,'')`,
			custID, c.GpsNumber)
		if dupErr == nil {
			continue // مستورد من قبل
		}

		if _, err := db.Exec(`
			INSERT INTO "GpsDeviceRequest"
				(id, "customerId", "employeeId", "purchaseType", "subscriptionType",
				 "subscriptionStart", "subscriptionEnd", "subscriptionStatus", status,
				 "simCardId", "gpsNumber", notes, "isChecked", "isActivated", "isDelivered",
				 "activationDate", "createdAt")
			VALUES (gen_random_uuid()::text, $1, $2, 'DEVICE_SIM', 'YEARLY',
				$3, $4::timestamptz,
				-- حالة الاشتراك تنحسب من التاريخ نفسه، مو قيمة ثابتة.
				-- نصرّح بالنوع بكل استعمال لأن بوستكرس ما يقدر يستنتجه لما
				-- نفس المعامل ينستعمل كعمود وكتاريخ بنفس الجملة.
				CASE WHEN $4::timestamptz IS NULL OR $4::timestamptz::date >= CURRENT_DATE
					THEN 'ACTIVE'::"GpsSubscriptionStatus"
					ELSE 'EXPIRED'::"GpsSubscriptionStatus" END,
				'DELIVERED',
				$5, $6, NULLIF($7,''), true, true, true,
				$3, COALESCE($8, now()))`,
			custID, ownerID, parseDate(c.InstalledAt), parseDate(c.SubscriptionEnd),
			simID, c.GpsNumber, noteText, parseDate(c.RequestedAt)); err != nil {
			log.Printf("تحذير: تعذرت إضافة جهاز الزبون %s: %v", c.FullName, err)
			continue
		}
		devices++
	}
	log.Printf("الزبائن: %d جديد | الأجهزة: %d | شرائح انربطت بزبائن: %d", added, devices, linked)
}
