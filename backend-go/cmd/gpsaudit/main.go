// gpsaudit يطابق بيانات الجي بي اس بين ملف الإكسل المستخرج والنظام.
//
// الاستخدام:
//
//	go run ./cmd/gpsaudit                      # يقارن الملف الافتراضي بالنظام
//	go run ./cmd/gpsaudit /path/to/gps_data.json
//
// ليش أداة مستقلة مو صفحة بالنظام؟ لأن هذا تدقيق يتسوّى مرة أو مرتين
// وقت الاستيراد، وما يستاهل شاشة دائمة — ولأنه يقرأ ملف الإكسل الي
// مو موجود بالسيرفر أصلاً.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/joho/godotenv"

	"staffmange-api/internal/config"
	"staffmange-api/internal/database"
)

type simRow struct {
	SimNumber string `json:"simNumber"`
	Status    string `json:"status"`
	RawStatus string `json:"rawStatus"`
}

type custRow struct {
	FullName  string `json:"fullName"`
	Phone     string `json:"phone"`
	GpsNumber string `json:"gpsNumber"`
}

type excelFile struct {
	Sims      []simRow  `json:"sims"`
	Customers []custRow `json:"customers"`
}

func norm(s string) string { return strings.Join(strings.Fields(strings.TrimSpace(s)), " ") }

func main() {
	_ = godotenv.Load()
	path := "cmd/importgps/gps_data.json"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("تعذر قراءة %s: %v", path, err)
	}
	var f excelFile
	if err := json.Unmarshal(raw, &f); err != nil {
		log.Fatalf("ملف غير صالح: %v", err)
	}

	cfg := config.Load()
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("تعذر الاتصال بقاعدة البيانات: %v", err)
	}
	defer db.Close()

	fmt.Println("════════ تدقيق الجي بي اس: الإكسل مقابل النظام ════════")

	// ── الشرائح ──
	excelSims := map[string]string{}
	for _, s := range f.Sims {
		excelSims[norm(s.SimNumber)] = s.Status
	}
	var dbSimCount int
	_ = db.Get(&dbSimCount, `SELECT COUNT(*) FROM "SimCard"`)
	dbSims := map[string]string{}
	rows := []struct {
		SimNumber string `db:"simNumber"`
		Status    string `db:"status"`
	}{}
	_ = db.Select(&rows, `SELECT "simNumber", status::text AS status FROM "SimCard"`)
	for _, r := range rows {
		dbSims[norm(r.SimNumber)] = r.Status
	}

	fmt.Printf("\n【الشرائح】\n")
	fmt.Printf("  بالإكسل: %d    بالنظام: %d    %s\n", len(excelSims), dbSimCount, verdict(len(excelSims), dbSimCount))
	fmt.Printf("  حالات الإكسل : %s\n", countBy(f.Sims))
	fmt.Printf("  حالات النظام : %s\n", countStatuses(rows))

	missing := []string{}
	for n := range excelSims {
		if _, ok := dbSims[n]; !ok {
			missing = append(missing, n)
		}
	}
	extra := []string{}
	for n := range dbSims {
		if _, ok := excelSims[n]; !ok {
			extra = append(extra, n)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	fmt.Printf("  بالإكسل وما انصعدت للنظام: %d %s\n", len(missing), sample(missing))
	fmt.Printf("  بالنظام ومو بالإكسل (مضافة يدوي): %d %s\n", len(extra), sample(extra))

	diff := 0
	for n, st := range excelSims {
		if dbSt, ok := dbSims[n]; ok && dbSt != st {
			diff++
		}
	}
	fmt.Printf("  حالتها تغيّرت بعد الاستيراد: %d  (طبيعي — الربط بزبون يقلبها IN_USE)\n", diff)

	// ── الزبائن ──
	people := map[string]bool{}      // اسم+هاتف
	triples := map[string]bool{}     // اسم+هاتف+رقم GPS
	dupGps := map[string][]string{}  // رقم GPS → أسماء
	for _, c := range f.Customers {
		nm, ph, gp := norm(c.FullName), norm(c.Phone), norm(c.GpsNumber)
		people[nm+"|"+ph] = true
		triples[nm+"|"+ph+"|"+gp] = true
		if gp != "" {
			dupGps[gp] = appendUniq(dupGps[gp], nm)
		}
	}
	var dbCust, dbReq int
	_ = db.Get(&dbCust, `SELECT COUNT(*) FROM "GpsCustomer"`)
	_ = db.Get(&dbReq, `SELECT COUNT(*) FROM "GpsDeviceRequest"`)

	fmt.Printf("\n【الزبائن】\n")
	fmt.Printf("  سطور الإكسل: %d\n", len(f.Customers))
	fmt.Printf("  أشخاص فريدين (اسم+هاتف): %d    بالنظام: %d    %s\n", len(people), dbCust, verdict(len(people), dbCust))
	fmt.Printf("  اشتراكات فريدة (اسم+هاتف+رقم GPS): %d    طلبات بالنظام: %d    %s\n", len(triples), dbReq, verdict(len(triples), dbReq))

	// ── التكرار ──
	conflicts := 0
	for _, names := range dupGps {
		if len(names) > 1 {
			conflicts++
		}
	}
	fmt.Printf("\n【التكرار】\n")
	fmt.Printf("  أرقام GPS عندها أكثر من اسم: %d  ← نقل جهاز أو غلط إملائي، راجعها\n", conflicts)
	fmt.Printf("  سطور مكررة تماماً بالإكسل: %d  (انطوت بالاستيراد)\n", len(f.Customers)-len(triples))

	// الأسماء المتشابهة جداً بنفس الهاتف = نفس الشخص بإملاء مختلف
	byPhone := map[string][]string{}
	for _, c := range f.Customers {
		ph := norm(c.Phone)
		if ph != "" {
			byPhone[ph] = appendUniq(byPhone[ph], norm(c.FullName))
		}
	}
	sameLine := 0
	for _, names := range byPhone {
		if len(names) > 1 {
			sameLine++
		}
	}
	fmt.Printf("  أرقام هاتف عندها أكثر من اسم: %d  ← يمكن نفس الشخص بإملاء مختلف\n", sameLine)

	fmt.Println("\n════════ خلص التدقيق ════════")
}

func appendUniq(list []string, v string) []string {
	for _, x := range list {
		if x == v {
			return list
		}
	}
	return append(list, v)
}

func verdict(a, b int) string {
	if a == b {
		return "✔ مطابق"
	}
	return fmt.Sprintf("✘ فرق %d", a-b)
}

func sample(list []string) string {
	if len(list) == 0 {
		return ""
	}
	n := len(list)
	if n > 5 {
		n = 5
	}
	return "→ " + strings.Join(list[:n], "، ")
}

func countBy(sims []simRow) string {
	m := map[string]int{}
	for _, s := range sims {
		m[s.Status]++
	}
	return fmtMap(m)
}

func countStatuses(rows []struct {
	SimNumber string `db:"simNumber"`
	Status    string `db:"status"`
}) string {
	m := map[string]int{}
	for _, r := range rows {
		m[r.Status]++
	}
	return fmtMap(m)
}

func fmtMap(m map[string]int) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", k, m[k]))
	}
	return strings.Join(parts, "  ")
}
