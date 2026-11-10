package system

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"staffmange-api/internal/core"
	"staffmange-api/internal/middleware"
	"staffmange-api/internal/models"
)

// ═══ تعديل وإضافة من لوحة الأدمن (ون بيس 11-10) ═══
// المالك، أو مدير نظام نطاه المالك الفعل (👑). القواعد:
//   - ماكو حذف من اللوحة أبداً (قاعدة دازاي: لا حذف بلا اتفاق).
//   - السجلات والتدقيق (Log/Audit/Trail/Event/Ledger…) ما تتعدّل من أي أحد — حتى المالك.
//   - الفلوس والرواتب والصلاحيات: المالك بس (حتى لو نطى التعديل لمدير).
//   - الأسرار ما تنعرض ولا تنعدّل.
//   - كل تغيير بـAdminEditLog: منو، متى، قبل وبعد (بلا أسرار).

var (
	readOnlyTable = regexp.MustCompile(`(Log|Audit|Trail|Event|Ledger|History|Snapshot|SchemaMigration|OwnerGrant|EmployeePermission)$|^(AuditTrail|StealthLog|LoginAudit|SecurityEvent|AdminEditLog)$`)
	ownerOnlyData = regexp.MustCompile(`(?i)(fund|invoice|payment|expense|credit|salary|commission|ledger|permission|price|cost)`)
	pkCols        = regexp.MustCompile(`^PRIMARY KEY \((.*)\)$`)
)

// editAccess هل يگدر يعدّل/يضيف بهالجدول؟ (السبب إذا لا)
func editAccess(a *core.Actor, table string) (bool, string) {
	if readOnlyTable.MatchString(table) {
		return false, "هذا سجل تدقيق — ما يتعدّل من اللوحة أبداً"
	}
	if ownerOnlyData.MatchString(table) && !a.Owner() {
		return false, "جدول فلوس/صلاحيات — تعديله للمالك بس"
	}
	return true, ""
}

func primaryKey(t models.Table) []string {
	for _, c := range t.Constraints {
		if m := pkCols.FindStringSubmatch(c.Def); m != nil {
			out := []string{}
			for _, p := range strings.Split(m[1], ",") {
				out = append(out, strings.Trim(strings.TrimSpace(p), `"`))
			}
			return out
		}
	}
	return nil
}

type editIn struct {
	Key    map[string]any `json:"key"`    // قيم المفتاح الأساسي (للتعديل)
	Values map[string]any `json:"values"` // الأعمدة الي تتغيّر/تنضاف
}

func (a *App) adminWrite(w http.ResponseWriter, r *http.Request, create bool) {
	actor := core.ActorOf(r)
	name := r.PathValue("table")
	t, ok := models.Get(name)
	if !ok {
		writeErr(w, http.StatusNotFound, "مو موجود")
		return
	}
	if ok, why := editAccess(actor, name); !ok {
		writeErr(w, http.StatusForbidden, why)
		return
	}
	var in editIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || len(in.Values) == 0 {
		writeErr(w, http.StatusBadRequest, "ماكو تغيير")
		return
	}
	cols := map[string]models.Column{}
	for _, c := range t.Columns {
		cols[c.Name] = c
	}
	pk := primaryKey(t)
	if len(pk) == 0 {
		writeErr(w, http.StatusBadRequest, "الجدول بلا مفتاح أساسي — ما ينعدّل من اللوحة")
		return
	}
	isPK := map[string]bool{}
	for _, k := range pk {
		isPK[k] = true
	}
	names, sets, args := []string{}, []string{}, []any{}
	for k, v := range in.Values {
		c, ok := cols[k]
		if !ok {
			writeErr(w, http.StatusBadRequest, "عمود مو موجود: "+k)
			return
		}
		if secretCol.MatchString(k) {
			writeErr(w, http.StatusForbidden, "عمود سري — ما ينعدّل من اللوحة: "+k)
			return
		}
		if !create && isPK[k] {
			writeErr(w, http.StatusBadRequest, "المفتاح الأساسي ما يتغيّر")
			return
		}
		if (k == "salary" || ownerOnlyData.MatchString(k)) && !actor.Owner() {
			writeErr(w, http.StatusForbidden, "عمود فلوس — تعديله للمالك بس: "+k)
			return
		}
		args = append(args, sqlValue(v))
		names = append(names, quoteIdent(k))
		sets = append(sets, fmt.Sprintf("%s = $%d::%s", quoteIdent(k), len(args), c.Type))
	}
	tx, err := a.deps.DB.Beginx()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "تعذر")
		return
	}
	defer tx.Rollback() //nolint:errcheck
	var before, after []byte
	key := map[string]any{}
	if create {
		ph := make([]string, len(names))
		for i := range names {
			ph[i] = fmt.Sprintf("$%d::%s", i+1, cols[strings.Trim(names[i], `"`)].Type)
		}
		q := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) RETURNING to_jsonb(%s.*)`, quoteIdent(name), strings.Join(names, ", "), strings.Join(ph, ", "), quoteIdent(name))
		if err := tx.Get(&after, q, args...); err != nil {
			writeErr(w, http.StatusBadRequest, "ما انضاف: "+pgMsg(err))
			return
		}
		var row map[string]any
		_ = json.Unmarshal(after, &row)
		for _, k := range pk {
			key[k] = row[k]
		}
	} else {
		where, selWhere, keyArgs := []string{}, []string{}, []any{}
		for i, k := range pk {
			v, ok := in.Key[k]
			if !ok {
				writeErr(w, http.StatusBadRequest, "ناقص المفتاح: "+k)
				return
			}
			args = append(args, sqlValue(v))
			keyArgs = append(keyArgs, sqlValue(v))
			where = append(where, fmt.Sprintf("%s = $%d::%s", quoteIdent(k), len(args), cols[k].Type))
			selWhere = append(selWhere, fmt.Sprintf("t.%s = $%d::%s", quoteIdent(k), i+1, cols[k].Type))
			key[k] = v
		}
		wsql := strings.Join(where, " AND ")
		if err := tx.Get(&before, fmt.Sprintf(`SELECT to_jsonb(t.*) FROM %s t WHERE %s FOR UPDATE`, quoteIdent(name), strings.Join(selWhere, " AND ")), keyArgs...); err != nil {
			writeErr(w, http.StatusNotFound, "مو موجود")
			return
		}
		q := fmt.Sprintf(`UPDATE %s SET %s WHERE %s RETURNING to_jsonb(%s.*)`, quoteIdent(name), strings.Join(sets, ", "), wsql, quoteIdent(name))
		if err := tx.Get(&after, q, args...); err != nil {
			writeErr(w, http.StatusBadRequest, "ما انحفظ: "+pgMsg(err))
			return
		}
	}
	kj, _ := json.Marshal(key)
	action := "update"
	if create {
		action = "create"
	}
	if _, err := tx.Exec(`INSERT INTO "AdminEditLog" ("tableName", action, "rowKey", before, after, "byId") VALUES ($1, $2, $3, $4, $5, $6)`,
		name, action, kj, nullJSON(redact(before)), nullJSON(redact(after)), middleware.EmployeeIDFromContext(r)); err != nil {
		writeErr(w, http.StatusInternalServerError, "تعذر تسجيل التغيير — ما انحفظ شي")
		return
	}
	if err := tx.Commit(); err != nil {
		writeErr(w, http.StatusInternalServerError, "تعذر الحفظ")
		return
	}
	var row map[string]any
	_ = json.Unmarshal(redact(after), &row)
	writeJSON(w, map[string]any{"ok": true, "row": row})
}

func (a *App) adminCreate(w http.ResponseWriter, r *http.Request) { a.adminWrite(w, r, true) }
func (a *App) adminUpdate(w http.ResponseWriter, r *http.Request) { a.adminWrite(w, r, false) }

// adminEditLog آخر التغييرات من اللوحة.
func (a *App) adminEditLog(w http.ResponseWriter, _ *http.Request) {
	rows := []struct {
		Table  string          `db:"tableName" json:"table"`
		Action string          `db:"action" json:"action"`
		Key    json.RawMessage `db:"rowKey" json:"key"`
		Before json.RawMessage `db:"before" json:"before"`
		After  json.RawMessage `db:"after" json:"after"`
		By     string          `db:"by" json:"by"`
		At     string          `db:"at" json:"at"`
	}{}
	_ = a.deps.DB.Select(&rows, `SELECT l."tableName", l.action, l."rowKey", COALESCE(l.before, 'null') AS before, COALESCE(l.after, 'null') AS after,
		COALESCE(e.name, l."byId") AS by, to_char(l."createdAt" AT TIME ZONE 'Asia/Baghdad', 'YYYY-MM-DD HH24:MI') AS at
		FROM "AdminEditLog" l LEFT JOIN "Employee" e ON e.id = l."byId" ORDER BY l."createdAt" DESC LIMIT 100`)
	writeJSON(w, rows)
}

// sqlValue القيمة كنص (والقاعدة تحوّلها لنوع العمود) — null يبقى NULL.
func sqlValue(v any) any {
	switch x := v.(type) {
	case nil:
		return nil
	case string:
		return x
	case map[string]any, []any:
		b, _ := json.Marshal(x)
		return string(b)
	default:
		return fmt.Sprint(x)
	}
}

func redact(b []byte) []byte {
	if len(b) == 0 {
		return b
	}
	var m map[string]any
	if json.Unmarshal(b, &m) != nil {
		return b
	}
	for k := range m {
		if secretCol.MatchString(k) && m[k] != nil {
			m[k] = "•••"
		}
	}
	out, _ := json.Marshal(m)
	return out
}

func nullJSON(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}

func pgMsg(err error) string {
	s := err.Error()
	if i := strings.Index(s, "pq: "); i >= 0 {
		s = s[i+4:]
	}
	if len(s) > 200 {
		s = s[:200]
	}
	return s
}
