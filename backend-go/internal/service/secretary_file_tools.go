package service

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ═══ 📊 كمال يدز إكسل + يحفظ الي تكله (دازاي 10-28) ═══
// «اكسلات هواي بلنظام… كمال كلهن يكدر يدزلياهن» و«من اكله هاي الشغله احفظها يحفظها».
//   - make_excel: نفس قراءة query_data (دور قراءة، معاملة READ ONLY، مهلة) بس لحد
//     ٥٠٠٠ صف، ويطلع ملف xlsx ينحفظ لصاحب المحادثة بس (AiFile، يتنظف بعد ٧ أيام).
//   - remember / forget: تفضيلات دائمة (SecretaryPref) تنقرا بكل محادثة.

const aiExcelRows = 5000

func registerFileTools(a *MatrixAgent, tc *SecretaryTurnCtx) {
	if tc == nil {
		return
	}
	a.AddTool(agentTool{Name: "make_excel",
		Desc: "سوّي ملف إكسل من بيانات النظام ودزّه إله (يطلع بطاقة بيها زر تنزيل). اكتب SELECT واحد (نفس قواعد query_data، جرّبه بـquery_data أول إذا مو متأكد)، وسمّي الأعمدة بالعربي بـAS \"اسم\". لحد ٥٠٠٠ صف. استعملها لأي «دزلي إكسل/جدول/ملف».",
		Props: map[string]any{
			"title": map[string]any{"type": "string", "description": "اسم الملف بالعربي (مثل: حجوزات تشرين الأول)"},
			"sql":   map[string]any{"type": "string", "description": "SELECT واحد بس"},
		},
		Req: []string{"title", "sql"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			title := strings.TrimSpace(strArg(in, "title"))
			if title == "" {
				title = "تقرير"
			}
			q, err := ValidateAIQuery(strArg(in, "sql"))
			if err != nil {
				return "", err
			}
			f := excelize.NewFile()
			defer f.Close()
			sheet := "Sheet1"
			n := 0
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			err = a.aiReadTx(ctx, func(tx aiQuerier) error {
				rows, err := tx.QueryContext(ctx, `SELECT * FROM (`+q+`) x LIMIT `+fmt.Sprint(aiExcelRows))
				if err != nil {
					return err
				}
				defer rows.Close()
				cols, _ := rows.Columns()
				for i, c := range cols {
					cell, _ := excelize.CoordinatesToCellName(i+1, 1)
					_ = f.SetCellValue(sheet, cell, c)
				}
				vals := make([]any, len(cols))
				ptrs := make([]any, len(cols))
				for i := range vals {
					ptrs[i] = &vals[i]
				}
				for rows.Next() {
					if err := rows.Scan(ptrs...); err != nil {
						return err
					}
					n++
					for i, v := range vals {
						cell, _ := excelize.CoordinatesToCellName(i+1, n+1)
						_ = f.SetCellValue(sheet, cell, excelValue(v))
					}
				}
				return rows.Err()
			})
			if err != nil {
				return "", err
			}
			if n == 0 {
				return "الاستعلام ما رجّع ولا صف — ما سوّيت ملف.", nil
			}
			bold, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}, Fill: excelize.Fill{Type: "pattern", Color: []string{"DCE6F1"}, Pattern: 1}})
			last, _ := excelize.CoordinatesToCellName(len(firstRow(f, sheet)), 1)
			_ = f.SetCellStyle(sheet, "A1", last, bold)
			_ = f.SetSheetView(sheet, 0, &excelize.ViewOptions{RightToLeft: boolPtr(true)})
			_ = f.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
			var buf bytes.Buffer
			if err := f.Write(&buf); err != nil {
				return "", err
			}
			_, _ = a.db.Exec(`DELETE FROM "AiFile" WHERE "createdAt" < now() - interval '7 days'`)
			var id string
			if err := a.db.Get(&id, `INSERT INTO "AiFile" ("employeeId", name, data) VALUES ($1, $2, $3) RETURNING id`,
				tc.UserID, title+".xlsx", buf.Bytes()); err != nil {
				return "", err
			}
			link := "/secretary/files/" + id
			tc.Cards = append(tc.Cards, SecretaryCard{Kind: "file", Title: "📊 " + title,
				Lines: []string{fmt.Sprintf("%d صف — اضغط حتى ينزل", n)}, Link: &link})
			return fmt.Sprintf("الملف «%s» جاهز (%d صف) وطلعتله بطاقة تنزيل. گوله بجملة إن الملف جاهز.", title, n), nil
		}})

	a.AddTool(agentTool{Name: "remember",
		Desc: "احفظ شي يريده المالك دايماً (طريقة، تفضيل، معلومة ثابتة) — يبقى بذاكرتك بكل محادثة. استعملها من يگول «احفظها/تذكّر/دايماً سوّي هيچ». اكتبها جملة واضحة ومختصرة.",
		Props: map[string]any{"text": map[string]any{"type": "string"}},
		Req:   []string{"text"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			t := strings.TrimSpace(strArg(in, "text"))
			if len([]rune(t)) < 3 || len([]rune(t)) > 400 {
				return "", fmt.Errorf("اكتبها جملة بين ٣ و٤٠٠ حرف")
			}
			var n int
			_ = a.db.Get(&n, `SELECT count(*) FROM "SecretaryPref" WHERE "employeeId" = $1 AND active`, tc.UserID)
			if n >= 60 {
				return "الذاكرة مليانة (٦٠) — اطلب منه يشيل شي قديم بـforget.", nil
			}
			if _, err := a.db.Exec(`INSERT INTO "SecretaryPref" (id, "employeeId", text) VALUES (gen_random_uuid()::text, $1, $2)`, tc.UserID, t); err != nil {
				return "", err
			}
			return "انحفظت. گوله بجملة قصيرة إنك حفظتها.", nil
		}})

	a.AddTool(agentTool{Name: "forget",
		Desc: "شيل شي محفوظ بذاكرتك (من يگول «انسى/شيلها/لا تسوي هيچ بعد»). اكتب جزء من نصه.",
		Props: map[string]any{"text": map[string]any{"type": "string"}},
		Req:   []string{"text"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			t := strings.TrimSpace(strArg(in, "text"))
			res, err := a.db.Exec(`UPDATE "SecretaryPref" SET active = false WHERE "employeeId" = $1 AND active AND text ILIKE '%' || $2 || '%'`, tc.UserID, t)
			if err != nil {
				return "", err
			}
			k, _ := res.RowsAffected()
			if k == 0 {
				return "ما لگيت شي محفوظ بهالكلام.", nil
			}
			return fmt.Sprintf("شلت %d من الذاكرة.", k), nil
		}})
}

func firstRow(f *excelize.File, sheet string) []string {
	rows, err := f.GetRows(sheet)
	if err != nil || len(rows) == 0 {
		return []string{""}
	}
	return rows[0]
}

func excelValue(v any) any {
	switch x := v.(type) {
	case nil:
		return ""
	case []byte:
		return string(x)
	case time.Time:
		return x.In(debriefLoc).Format("2006-01-02 15:04")
	case sql.NullString:
		return x.String
	}
	return v
}

func boolPtr(b bool) *bool { return &b }
