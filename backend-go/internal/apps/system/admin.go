package system

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ لوحة الأدمن التلقائية (ون بيس 11-10) ═══
// مثل أدمن جانغو: كل جدول بالقاعدة إله شاشة تصفّح وبحث — بلا كود لكل جدول.
// **قراءة بس** (المالك وحده): التعديل يبقى بشاشات النظام الي بيها قواعدها.
// الأسرار (كلمات السر، التوكنات، الرموز) تطلع «•••» دايماً.

var secretCol = regexp.MustCompile(`(?i)(password|passwd|hash|secret|token|otp|apikey|api_key|privatekey|pin$|code_hash|sealed)`)

type tableInfo struct {
	Name    string `db:"name" json:"name"`
	Rows    int64  `db:"rows" json:"rows"`
	Columns int    `db:"columns" json:"columns"`
}

type column struct {
	Name string `db:"name" json:"name"`
	Type string `db:"type" json:"type"`
}

func (a *App) tables(db *sqlx.DB) ([]tableInfo, error) {
	out := []tableInfo{}
	err := db.Select(&out, `SELECT c.relname AS name, GREATEST(c.reltuples, 0)::bigint AS rows,
		(SELECT count(*) FROM information_schema.columns ic WHERE ic.table_schema = 'public' AND ic.table_name = c.relname)::int AS columns
		FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public' AND c.relkind = 'r' AND c.relname NOT LIKE 'schema_%'
		ORDER BY c.relname`)
	return out, err
}

func (a *App) adminTables(w http.ResponseWriter, _ *http.Request) {
	t, err := a.tables(a.deps.DB)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "تعذر جلب الجداول")
		return
	}
	writeJSON(w, t)
}

// adminRows صفحة من جدول. الاسم ينفحص مقابل قائمة الجداول (ما ينحط نص المستخدم بالاستعلام).
func (a *App) adminRows(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("table")
	var ok bool
	if err := a.deps.DB.Get(&ok, `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1 AND table_type = 'BASE TABLE')`, name); err != nil || !ok {
		writeErr(w, http.StatusNotFound, "مو موجود")
		return
	}
	cols := []column{}
	if err := a.deps.DB.Select(&cols, `SELECT column_name AS name, data_type AS type FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1 ORDER BY ordinal_position`, name); err != nil {
		writeErr(w, http.StatusInternalServerError, "تعذر")
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	const size = 50
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	where, args := "", []any{}
	if q != "" {
		parts := []string{}
		for _, c := range cols {
			if secretCol.MatchString(c.Name) {
				continue // البحث ما يمر على الأسرار — حتى ما ينكشف وجود قيمة
			}
			parts = append(parts, fmt.Sprintf(`%s::text ILIKE $1`, quoteIdent(c.Name)))
		}
		if len(parts) > 0 {
			where = " WHERE " + strings.Join(parts, " OR ")
			args = append(args, "%"+q+"%")
		}
	}
	order := ""
	for _, c := range cols {
		if c.Name == "createdAt" {
			order = ` ORDER BY "createdAt" DESC NULLS LAST`
		}
	}
	var total int64
	_ = a.deps.DB.Get(&total, `SELECT count(*) FROM `+quoteIdent(name)+where, args...)
	rows, err := a.deps.DB.Queryx(fmt.Sprintf(`SELECT * FROM %s%s%s LIMIT %d OFFSET %d`, quoteIdent(name), where, order, size, (page-1)*size), args...)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "تعذر قراءة الجدول")
		return
	}
	defer rows.Close()
	data := []map[string]any{}
	for rows.Next() {
		m := map[string]any{}
		if err := rows.MapScan(m); err != nil {
			continue
		}
		for k, v := range m {
			switch {
			case v == nil:
			case secretCol.MatchString(k):
				m[k] = "•••"
			default:
				switch x := v.(type) {
				case []byte:
					s := string(x)
					if len(s) > 300 {
						s = s[:300] + "…"
					}
					m[k] = s
				case time.Time:
					m[k] = x.Format(time.RFC3339)
				}
			}
		}
		data = append(data, m)
	}
	writeJSON(w, map[string]any{"table": name, "columns": cols, "rows": data, "total": total, "page": page, "size": size})
}

func quoteIdent(s string) string { return `"` + strings.ReplaceAll(s, `"`, `""`) + `"` }

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "error": msg})
}
