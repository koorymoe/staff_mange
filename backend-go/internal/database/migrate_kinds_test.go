package database

import (
	"regexp"
	"strconv"
	"testing"
)

// ٧ف (11-05): الترحيلات صنفين بس — **مخطط** (CREATE/ALTER/DROP) أو **بيانات**
// (UPDATE/INSERT/DELETE: تصليح أو بذر). الترحيل المخلوط يصعّب المراجعة والاسترجاع:
// لو تصليح البيانات غلط، ما نگدر نعيده بلا ما نعيد المخطط.
// القديمة (لحد 0441) انطبّقت وأرقامها مسجّلة بالقاعدة فما تنلمس — الفهرس بـ
// docs/adr/0009-migrations-kinds.md. من هسه: كل ترحيل جديد صنف واحد، والبيانات
// اسمها يبدي بـ data_ (مثال: 0450_data_fix_x).
const migrationKindsBaseline = 441

var (
	migDDL = regexp.MustCompile(`(?i)\b(CREATE|ALTER|DROP)\s+(TABLE|INDEX|UNIQUE|TYPE|EXTENSION|FUNCTION|TRIGGER|OR|CONSTRAINT|VIEW|SEQUENCE)`)
	migDML = regexp.MustCompile(`(?i)\b(UPDATE\s+"|DELETE\s+FROM|INSERT\s+INTO)`)
	migNum = regexp.MustCompile(`^(\d{4})[a-z]?_(.*)$`)
)

// MigrationKind يصنّف ترحيل: schema أو data أو mixed.
func MigrationKind(sql string) string {
	a, b := migDDL.MatchString(sql), migDML.MatchString(sql)
	switch {
	case a && b:
		return "mixed"
	case b:
		return "data"
	}
	return "schema"
}

func TestNewMigrationsAreSingleKind(t *testing.T) {
	for _, m := range versionedMigrations() {
		g := migNum.FindStringSubmatch(m.Version)
		if g == nil {
			t.Errorf("رقم ترحيل غلط: %s", m.Version)
			continue
		}
		n, _ := strconv.Atoi(g[1])
		if n <= migrationKindsBaseline {
			continue
		}
		switch k := MigrationKind(m.SQL); {
		case k == "mixed":
			t.Errorf("%s: يخلط المخطط ويا تصليح البيانات — قسّمه لترحيلين (المخطط، وبعده data_…)", m.Version)
		case k == "data" && !regexp.MustCompile(`^data_`).MatchString(g[2]):
			t.Errorf("%s: ترحيل بيانات — اسمه لازم يبدي بـ data_", m.Version)
		}
	}
}

func TestMigrationKindClassifier(t *testing.T) {
	cases := map[string]string{
		`ALTER TABLE "X" ADD COLUMN IF NOT EXISTS y TEXT`:                        "schema",
		`UPDATE "X" SET y = 1 WHERE y IS NULL`:                                   "data",
		`CREATE TABLE IF NOT EXISTS "Y" (id TEXT); INSERT INTO "Y" VALUES ('a')`: "mixed",
	}
	for sql, want := range cases {
		if got := MigrationKind(sql); got != want {
			t.Errorf("%q = %s، المتوقع %s", sql, got, want)
		}
	}
}
