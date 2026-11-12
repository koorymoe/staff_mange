// makemigration — 🏴‍☠️ ون بيس: يقارن الموديلات (internal/models) ويا القاعدة، ويكتب
// تحديث الإضافة لحاله بـinternal/database/schema_auto.go (مثل makemigrations بجانغو).
//
//	DATABASE_URL=… go run ./cmd/makemigration -name add_x   ← يكتب التحديث
//	DATABASE_URL=… go run ./cmd/makemigration -check        ← يطبع الفرق بس
//
// الحذف أو تغيير شي موجود ما يتولّد أبداً: يطلع بقائمة «يحتاج موافقة» والأداة توقف.
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/database"
	"staffmange-api/internal/models"
)

const autoFile = "internal/database/schema_auto.go"
const marker = "// ‹makemigration›"

func main() {
	name := flag.String("name", "", "وصف قصير بالإنگليزي (حروف صغيرة و_)")
	check := flag.Bool("check", false, "اطبع الفرق بس")
	flag.Parse()
	db := sqlx.MustConnect("postgres", os.Getenv("DATABASE_URL"))
	if err := database.Migrate(db, "", ""); err != nil {
		fail(err.Error())
	}
	have, haveE, err := models.Introspect(db)
	haveO, _ := models.IntrospectObjects(db)
	if err != nil {
		fail(err.Error())
	}
	p := models.DiffAll(models.Tables(), models.Enums(), models.Objects(), have, haveE, haveO)
	if p.Empty() {
		fmt.Println("✅ الموديلات والقاعدة متطابقين — ماكو تحديث")
		return
	}
	for _, s := range p.Safe {
		fmt.Println("＋", strings.ReplaceAll(s, "\n", " "))
	}
	for _, s := range p.NeedsApproval {
		fmt.Println("⚠️ يحتاج موافقة:", s)
	}
	if *check {
		os.Exit(1)
	}
	if len(p.NeedsApproval) > 0 {
		fail("بيه تغييرات تحتاج موافقة دازاي — ما انكتب شي. صلّح الموديل، أو اكتب تحديث يدوي بموافقة.")
	}
	if !regexp.MustCompile(`^[a-z][a-z0-9_]{2,40}$`).MatchString(*name) {
		fail("لازم -name (مثلاً add_booking_note)")
	}
	next := 0
	for _, m := range database.AllVersions() {
		if g := regexp.MustCompile(`^(\d{4})_`).FindStringSubmatch(m); g != nil {
			if n, _ := strconv.Atoi(g[1]); n > next {
				next = n
			}
		}
	}
	version := fmt.Sprintf("%04d_%s", next+1, *name)
	b, err := os.ReadFile(autoFile)
	if err != nil {
		fail(err.Error())
	}
	sql := "\t\t\t" + strings.Join(p.Safe, ";\n\t\t\t") + ";"
	entry := fmt.Sprintf("{\n\t\t\tVersion: %q,\n\t\t\tSQL: `\n%s\n\t\t`,\n\t\t},\n\t\t", version, sql)
	out := strings.Replace(string(b), marker, entry+marker, 1)
	if out == string(b) {
		fail("ما لگيت العلامة بـ" + autoFile)
	}
	if err := os.WriteFile(autoFile, []byte(out), 0o644); err != nil {
		fail(err.Error())
	}
	fmt.Printf("✍️ انكتب %s (%d عبارة) — شغّل السيرفر/الاختبار حتى ينطبق\n", version, len(p.Safe))
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, "❌", msg)
	os.Exit(1)
}
