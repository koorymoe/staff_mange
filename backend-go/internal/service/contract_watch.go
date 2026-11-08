package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/repository"
)

// ═══ 📅 متابعة العقود (ع) 11-08 ═══
// «ليش اكو موظفين ماكاتبين موعد انتهاء العقد؟» — محد چان يذكّر. هسه:
//   - أسبوعياً: إداري الكوادر يوصله عدد وأسماء الموظفين الي بلا «نهاية العقد» بالملف.
//   - قبل انتهاء أي عقد بشهر: يوصله ويوصل للمالك، مرة وحدة لكل تاريخ.
// المدير والمالك برّا (ما إلهم عقد بالنظام).

func StartContractWatch(db *sqlx.DB, notif *repository.NotificationRepository) {
	go func() {
		time.Sleep(6 * time.Minute)
		for {
			ContractWatchOnce(db, notif, time.Now())
			time.Sleep(6 * time.Hour)
		}
	}()
}

func ContractWatchOnce(db *sqlx.DB, notif *repository.NotificationRepository, now time.Time) (expiring, missing int) {
	if notif == nil {
		return
	}
	hr := func(msg string) {
		_ = notif.CreateForRolesOrPermission([]string{"HR_COORDINATOR"}, "staff_management", "CONTRACT_WATCH", msg)
	}
	// ١. تنتهي خلال ٣٠ يوم — مرة لكل موظف ولكل تاريخ انتهاء
	rows := []struct {
		Name string    `db:"name"`
		End  time.Time `db:"end"`
	}{}
	if err := db.Select(&rows, `INSERT INTO "ContractReminder" (kind, "employeeId", "forDate")
		SELECT 'EXPIRING', e.id, e."contractEndDate"::date FROM "Employee" e
		WHERE e.status = 'ACTIVE' AND e.role::text NOT IN ('ADMIN', 'OWNER')
		  AND e."contractEndDate" IS NOT NULL AND e."contractEndDate"::date BETWEEN baghdad_today() AND baghdad_today() + 30
		ON CONFLICT DO NOTHING
		RETURNING (SELECT name FROM "Employee" WHERE id = "employeeId") AS name, "forDate" AS end`); err != nil {
		log.Printf("[contracts-watch] %v", err)
	}
	for _, r := range rows {
		days := int(r.End.Sub(now).Hours()/24) + 1
		msg := fmt.Sprintf("📅 عقد %s ينتهي %s (بعد %d يوم) — جدّده أو بلّغ المالك.", r.Name, r.End.Format("2006-01-02"), days)
		hr(msg)
		_ = notif.CreateForRole("OWNER", "CONTRACT_WATCH", msg)
	}
	expiring = len(rows)
	// ٢. أسبوعياً (أول دورة بكل أسبوع): الي بلا نهاية عقد
	week := now.AddDate(0, 0, -int(now.Weekday())).Format("2006-01-02")
	var claimed bool
	_ = db.Get(&claimed, `INSERT INTO "ContractReminder" (kind, "employeeId", "forDate") VALUES ('MISSING', '', $1::date)
		ON CONFLICT DO NOTHING RETURNING true`, week)
	if !claimed {
		return
	}
	names := []string{}
	_ = db.Select(&names, `SELECT name FROM "Employee" WHERE status = 'ACTIVE' AND role::text NOT IN ('ADMIN', 'OWNER')
		AND "contractEndDate" IS NULL ORDER BY name`)
	if missing = len(names); missing > 0 {
		shown := names
		if len(shown) > 10 {
			shown = shown[:10]
		}
		more := ""
		if len(names) > 10 {
			more = fmt.Sprintf(" و%d غيرهم", len(names)-10)
		}
		hr(fmt.Sprintf("📅 %d موظف بلا «نهاية العقد» بملفه: %s%s — عبّيها من «إدارة الكوادر» ← تعديل كل البيانات.", missing, strings.Join(shown, "، "), more))
	}
	return
}
