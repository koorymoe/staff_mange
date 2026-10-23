package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"staffmange-api/internal/database"
)

// ═══ 🔎 وصول الذكاء لكل النظام — قراءة بس (دازاي 10-23) ═══
// «اريد كلشي ينشاف بالنظام… جاي افكر بمستقبل جبير للنظام». بدل أداة لكل جدول:
// describe_data يعرض الجداول والأعمدة المسموحة، وquery_data ينفّذ SELECT تحت دور قراءة
// (database/ai_readers.go) داخل معاملة READ ONLY ويا مهلة ٥ ثواني وحد ١٠٠ صف.
// الأسرار والهواتف — وعند ماتركس أسماء الزبائن — ممنوعة بصلاحيات Postgres نفسها.

const aiQueryRows = 100

var (
	aiSelectRe   = regexp.MustCompile(`(?is)^\s*(select|with)\b`)
	aiForbidRe   = regexp.MustCompile(`(?i)\b(insert|update|delete|drop|alter|create|grant|revoke|truncate|copy|call|do|set|reset|vacuum|lock|listen|notify|pg_sleep|dblink|lo_import|lo_export|pg_read_file|pg_ls_dir|set_config)\b`)
	aiTableHints = map[string]string{
		"Employee":        "الموظفين: الاسم، الدور (role)، المنصب، الحالة (status: ACTIVE/SUSPENDED…)، سبب الإيقاف، الشهادة، القسم",
		"ExtraTask":       "المهام الإضافية: العنوان، الموظف (assignedToId)، الحالة NEW/IN_PROGRESS/DONE/CANCELLED، الأولوية، dueAt، doneAt",
		"Booking":         "الحجوزات: code، الحالة، النوع، scheduledAt، completedAt، الزبون (customerId)",
		"Mission":         "مهمة تنفيذ الحجز: leaderId (الليدر)، المرحلة، الأوقات",
		"LeaveRequest":    "الإجازات: employeeId، startDate، endDate، status",
		"Complaint":       "الشكاوى",
		"QualityFollowUp": "متابعة الجودة بعد الحجز وتقييم الزبون",
		"LeaderInvoice":   "فواتير الليدر",
		"Notification":    "الإشعارات",
		"Project":         "المشاريع",
		"Customer":        "الزبائن",
		"DisciplineEvent": "الانضباط والمخالفات",
	}
)

type aiQuerier interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func (a *MatrixAgent) aiRole() string {
	if a.plainNames {
		return database.AIRoleIris
	}
	return database.AIRoleMatrix
}

// aiReadTx معاملة قراءة بدور الذكاء ومهلة.
func (a *MatrixAgent) aiReadTx(ctx context.Context, fn func(tx aiQuerier) error) error {
	if !database.AIReadersReady {
		return errors.New("وصول القراءة الشامل مو مفعّل بالسيرفر")
	}
	tx, err := a.db.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SET LOCAL statement_timeout = '5s'`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `SET LOCAL ROLE `+a.aiRole()); err != nil {
		return err
	}
	return fn(tx)
}

// ValidateAIQuery SELECT واحد بس.
func ValidateAIQuery(q string) (string, error) {
	q = strings.TrimRight(strings.TrimSpace(q), "; \n\t")
	if !aiSelectRe.MatchString(q) {
		return "", errors.New("بس SELECT أو WITH")
	}
	if strings.Contains(q, ";") {
		return "", errors.New("استعلام واحد بس")
	}
	if aiForbidRe.MatchString(q) {
		return "", errors.New("الاستعلام بيه أمر غير مسموح — قراءة بس")
	}
	return q, nil
}

func (a *MatrixAgent) registerDataTools() {
	a.AddTool(agentTool{Name: "describe_data",
		Desc:  "خريطة بيانات النظام: بلا topic يرجّع أسماء كل الجداول ويا وصف المهمة منها؛ ويا topic (اسم جدول أو جزء منه) يرجّع أعمدته. استعملها قبل query_data إذا ما متأكد من الأسماء.",
		Props: map[string]any{"topic": map[string]any{"type": "string", "description": "اسم جدول (مثلاً ExtraTask) أو فاضي"}},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			topic := strArg(in, "topic")
			var b strings.Builder
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			err := a.aiReadTx(ctx, func(tx aiQuerier) error {
				rows, err := tx.QueryContext(ctx, `SELECT table_name, string_agg(column_name || ':' || data_type, ', ' ORDER BY ordinal_position)
					FROM information_schema.columns WHERE table_schema = 'public' AND ($1 = '' OR table_name ILIKE '%' || $1 || '%')
					GROUP BY table_name ORDER BY table_name`, topic)
				if err != nil {
					return err
				}
				defer rows.Close()
				for rows.Next() {
					var t, c string
					if err := rows.Scan(&t, &c); err != nil {
						return err
					}
					h := aiTableHints[t]
					if topic == "" {
						if h != "" {
							fmt.Fprintf(&b, "%s — %s\n", t, h)
						} else {
							b.WriteString(t + "\n")
						}
						continue
					}
					fmt.Fprintf(&b, "\"%s\": %s\n", t, c)
					if h != "" {
						b.WriteString("  ↳ " + h + "\n")
					}
				}
				return rows.Err()
			})
			if err != nil {
				return "", err
			}
			if b.Len() == 0 {
				return "ما لكيت جدول بهالاسم.", nil
			}
			return "أسماء الجداول والأعمدة حساسة لحالة الأحرف — حطها بين \"علامات تنصيص\".\n" + b.String(), nil
		}})
	a.AddTool(agentTool{Name: "query_data",
		Desc:  "اقرا أي شي من قاعدة بيانات النظام بـSELECT (Postgres). قراءة بس، حد ١٠٠ صف، ٥ ثواني. أسماء الجداول والأعمدة بين \"\" (مثل SELECT t.title, t.status FROM \"ExtraTask\" t JOIN \"Employee\" e ON e.id = t.\"assignedToId\" WHERE e.name ILIKE '%مصطفى%'). للتوقيت: baghdad_today() وbaghdad_date(ts). الأعمدة الممنوعة (أسرار، هواتف…) تطلع خطأ صلاحية — اختار غيرها.",
		Props: map[string]any{"sql": map[string]any{"type": "string"}}, Req: []string{"sql"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			q, err := ValidateAIQuery(strArg(in, "sql"))
			if err != nil {
				return "", err
			}
			var out strings.Builder
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			err = a.aiReadTx(ctx, func(tx aiQuerier) error {
				rows, err := tx.QueryContext(ctx, fmt.Sprintf(`SELECT * FROM (%s) _q LIMIT %d`, q, aiQueryRows+1))
				if err != nil {
					return err
				}
				defer rows.Close()
				cols, _ := rows.Columns()
				n := 0
				for rows.Next() {
					n++
					if n > aiQueryRows {
						out.WriteString("…(أكثر من ١٠٠ صف — ضيّق الاستعلام)\n")
						break
					}
					vals := make([]any, len(cols))
					ptrs := make([]any, len(cols))
					for i := range vals {
						ptrs[i] = &vals[i]
					}
					if err := rows.Scan(ptrs...); err != nil {
						return err
					}
					row := map[string]any{}
					for i, c := range cols {
						switch v := vals[i].(type) {
						case []byte:
							row[c] = string(v)
						case time.Time:
							row[c] = bagh(v).Format("2006-01-02 15:04")
						default:
							row[c] = v
						}
					}
					j, _ := json.Marshal(row)
					out.Write(j)
					out.WriteByte('\n')
				}
				if n == 0 {
					out.WriteString("ماكو نتائج.")
				}
				return rows.Err()
			})
			if err != nil {
				if strings.Contains(err.Error(), "permission denied") {
					return "", errors.New("عمود أو جدول ممنوع (أسرار أو خصوصية) — اختار أعمدة ثانية")
				}
				return "", errors.New("الاستعلام فشل: " + err.Error())
			}
			return out.String(), nil
		}})
}
