package database

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
)

// ═══ 🔎 أدوار قراءة الذكاء (دازاي 10-23: «اريده يوصل لكل معلومة بالنظام») ═══
// ايريس وماتركس يقرون النظام كله بـquery_data — بس من خلال أدوار Postgres بصلاحية
// SELECT على أعمدة محددة. الحماية بالقاعدة نفسها، مو بفلترة نص بالكود:
//   - ممنوع على الاثنين: الباسوردات والتوكنات والأسرار وأرقام الهواتف ووثائق الهوية،
//     ومذكرات المالك ومحادثات ايريس وسجل وضع الفحص.
//   - ماتركس فوكاها: أسماء الزبائن وعناوينهم ومواقعهم ورسائل الواتساب (قاعدة الخصوصية).
// الأدوار NOLOGIN، والخادم يلبسها بـSET LOCAL ROLE داخل معاملة READ ONLY بس.
// تنعاد كل إقلاع حتى الجداول والأعمدة الجديدة تدخل تلقائياً.

const (
	AIRoleIris   = "ai_iris"
	AIRoleMatrix = "ai_matrix"
)

var (
	aiDenyColumn       = regexp.MustCompile(`(?i)(password|codehash|token|secret|otp|apikey|phone|mobile|whatsapp|idcard|residencecard|nationalid)`)
	aiDenyMatrixColumn = regexp.MustCompile(`(?i)(address|customername|guestname|locationurl|latitude|longitude|maplat|maplng|maplong|email)`)
	aiDenyTables       = map[string]bool{"AiFile": true, "AiReminder": true, "AiWatch": true, "SecretaryNote": true, "SecretaryMessage": true, "SecretaryAction": true, "SecretaryPref": true,
		"StealthLog": true, "schema_migrations": true}
	aiDenyMatrixTables = map[string]bool{"WaMessage": true, "WaConversation": true, "MeetingRequest": true, "AiConversation": true}
)

// AIReadersReady الأدوار جاهزة؟ (إذا مستخدم القاعدة ما يكدر يسوي أدوار، أدوات query_data تنطفي)
var AIReadersReady bool

// EnsureAIReaders ينشئ الأدوار ويعيد صلاحياتها. ما يوكّف الإقلاع إذا فشل.
func EnsureAIReaders(db *sqlx.DB) {
	if err := ensureAIReaders(db); err != nil {
		log.Printf("🔎 أدوار قراءة الذكاء ما انعملت (query_data مطفي): %v", err)
		AIReadersReady = false
		return
	}
	AIReadersReady = true
	log.Printf("🔎 أدوار قراءة الذكاء جاهزة (ai_iris · ai_matrix)")
}

func ensureAIReaders(db *sqlx.DB) error {
	for _, role := range []string{AIRoleIris, AIRoleMatrix} {
		if _, err := db.Exec(fmt.Sprintf(`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '%[1]s') THEN CREATE ROLE %[1]s NOLOGIN; END IF;
		END $$`, role)); err != nil {
			return err
		}
		if _, err := db.Exec(fmt.Sprintf(`GRANT %s TO CURRENT_USER`, role)); err != nil {
			return err
		}
		if _, err := db.Exec(fmt.Sprintf(`GRANT USAGE ON SCHEMA public TO %[1]s; REVOKE ALL ON ALL TABLES IN SCHEMA public FROM %[1]s`, role)); err != nil {
			return err
		}
	}
	var cols []struct {
		Table  string `db:"table_name"`
		Column string `db:"column_name"`
	}
	if err := db.Select(&cols, `SELECT c.table_name, c.column_name FROM information_schema.columns c
		JOIN information_schema.tables t ON t.table_schema = c.table_schema AND t.table_name = c.table_name
		WHERE c.table_schema = 'public' AND t.table_type = 'BASE TABLE' ORDER BY c.table_name, c.ordinal_position`); err != nil {
		return err
	}
	byTable := map[string][]string{}
	order := []string{}
	for _, c := range cols {
		if _, ok := byTable[c.Table]; !ok {
			order = append(order, c.Table)
		}
		byTable[c.Table] = append(byTable[c.Table], c.Column)
	}
	for _, table := range order {
		if aiDenyTables[table] {
			continue
		}
		var iris, matrix []string
		for _, col := range byTable[table] {
			if aiDenyColumn.MatchString(col) {
				continue
			}
			q := `"` + strings.ReplaceAll(col, `"`, `""`) + `"`
			iris = append(iris, q)
			if aiDenyMatrixColumn.MatchString(col) || (table == "Customer" && col == "name") {
				continue
			}
			matrix = append(matrix, q)
		}
		tq := `"` + strings.ReplaceAll(table, `"`, `""`) + `"`
		if len(iris) > 0 {
			if _, err := db.Exec(fmt.Sprintf(`GRANT SELECT (%s) ON %s TO %s`, strings.Join(iris, ", "), tq, AIRoleIris)); err != nil {
				return fmt.Errorf("%s: %w", table, err)
			}
		}
		if len(matrix) > 0 && !aiDenyMatrixTables[table] {
			if _, err := db.Exec(fmt.Sprintf(`GRANT SELECT (%s) ON %s TO %s`, strings.Join(matrix, ", "), tq, AIRoleMatrix)); err != nil {
				return fmt.Errorf("%s: %w", table, err)
			}
		}
	}
	return nil
}
