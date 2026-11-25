package service

import (
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/model"
)

// ═══ 🧰 عهدة الأدوات — الموظف بنفسه (ل) 11-25 دازاي ═══
// «خانة الموظف بنفسه يختار الأداة ووقت أخذ الأداة، وتبقى بعهدته، ويسوي تسليم ويرفع صورة، واني أدققها،
//  أو اني أختار الموظف في حال هو ما گدر».
// مبني فوگ طلبات الأدوات (ToolRequest) — نفس المخزن ونفس الكميات:
//   أخذ بنفسه ← طلب ينوافق عليه فوراً باسمه (selfCheckout) ويطلع من المخزن، ووقت الأخذ الي اختاره.
//   المدير يختار الموظف ← نفس الشي بموافقة المدير.
//   تسليم ← الموظف يرفع صورة (returnRequestedAt) والأداة تبقى بعهدته لحد التدقيق.
//   تدقيق (المالك/المدير) ← صح: ترجع للمخزن · غلط: ترجع لعهدته ويا ملاحظة.

type ToolCustodyItem struct {
	ID                string     `db:"id" json:"id"`
	ToolID            string     `db:"toolId" json:"toolId"`
	ToolName          string     `db:"toolName" json:"toolName"`
	EmployeeID        string     `db:"employeeId" json:"employeeId"`
	EmployeeName      string     `db:"employeeName" json:"employeeName"`
	TakenAt           *time.Time `db:"takenAt" json:"takenAt"`
	SelfCheckout      bool       `db:"selfCheckout" json:"selfCheckout"`
	ByName            *string    `db:"byName" json:"byName"`
	ReturnRequestedAt *time.Time `db:"returnRequestedAt" json:"returnRequestedAt"`
	ReturnPhoto       *string    `db:"returnPhoto" json:"returnPhoto"`
	ReturnNote        *string    `db:"returnNote" json:"returnNote"`
	VerifyNote        *string    `db:"verifyNote" json:"verifyNote"`
	ReturnedAt        *time.Time `db:"returnedAt" json:"returnedAt"`
	VerifiedAt        *time.Time `db:"verifiedAt" json:"verifiedAt"`
	VerifiedBy        *string    `db:"verifiedBy" json:"verifiedBy"`
}

type ToolCustody struct {
	db  *sqlx.DB
	inv *InventoryService
}

func NewToolCustody(db *sqlx.DB, inv *InventoryService) *ToolCustody {
	return &ToolCustody{db: db, inv: inv}
}

const custodyCols = `SELECT r.id, r."toolId", t.name AS "toolName", r."employeeId", e.name AS "employeeName",
	COALESCE(r."takenAt", r."approvedAt" AT TIME ZONE current_setting('TimeZone')) AS "takenAt", r."selfCheckout",
	CASE WHEN r."approvedById" <> r."employeeId" THEN a.name END AS "byName",
	r."returnRequestedAt", r."returnPhoto", r."returnNote", r."verifyNote",
	r."returnedAt" AT TIME ZONE current_setting('TimeZone') AS "returnedAt", r."verifiedAt", v.name AS "verifiedBy"
	FROM "ToolRequest" r JOIN "OnDemandTool" t ON t.id = r."toolId" JOIN "Employee" e ON e.id = r."employeeId"
	LEFT JOIN "Employee" a ON a.id = r."approvedById" LEFT JOIN "Employee" v ON v.id = r."verifiedById"`

func parseTakenAt(s string, now time.Time) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return now, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		if t, err = time.ParseInLocation("2006-01-02T15:04", s, debriefLoc); err != nil {
			return time.Time{}, errors.New("وقت الأخذ غلط")
		}
	}
	if t.After(now.Add(10*time.Minute)) || t.Before(now.AddDate(0, 0, -7)) {
		return time.Time{}, errors.New("وقت الأخذ لازم مو بالمستقبل ولا أقدم من أسبوع")
	}
	return t, nil
}

// Take الموظف ياخذ بنفسه (by == emp)، أو المدير يختاره عنه (by مدير).
func (c *ToolCustody) Take(by, emp, toolID, takenAt, note string, now time.Time) (*ToolCustodyItem, error) {
	at, err := parseTakenAt(takenAt, now)
	if err != nil {
		return nil, err
	}
	var avail int
	if err := c.db.Get(&avail, `SELECT "availableQuantity" FROM "OnDemandTool" WHERE id = $1`, toolID); err != nil {
		return nil, errors.New("الأداة غير موجودة")
	}
	if avail <= 0 {
		return nil, errors.New("ماكو من هاي الأداة بالمخزن هسه — كلها بعهدة غيرك")
	}
	kind := model.ToolRequestKindSpecialized
	n := strings.TrimSpace(note)
	var desc *string
	if n != "" {
		desc = &n
	}
	created, err := c.inv.CreateToolRequest(model.CreateToolRequestRequest{EmployeeID: emp, ToolID: toolID, Reason: model.ToolRequestReasonNeverHad, RequestKind: &kind, Description: desc})
	if err != nil {
		return nil, err
	}
	if _, err := c.inv.ApproveToolRequest(created.ID, model.ApproveToolRequestRequest{ApprovedByID: by}); err != nil {
		_ = c.inv.DeleteToolRequest(created.ID)
		return nil, err
	}
	if _, err := c.db.Exec(`UPDATE "ToolRequest" SET "takenAt" = $2, "selfCheckout" = $3 WHERE id = $1`, created.ID, at, by == emp); err != nil {
		return nil, err
	}
	return c.get(created.ID)
}

func (c *ToolCustody) get(id string) (*ToolCustodyItem, error) {
	var it ToolCustodyItem
	if err := c.db.Get(&it, custodyCols+` WHERE r.id = $1`, id); err != nil {
		return nil, err
	}
	return &it, nil
}

// Mine عهدة الموظف الحالية (الي بعده ما ترجعت).
func (c *ToolCustody) Mine(emp string) ([]ToolCustodyItem, error) {
	rows := []ToolCustodyItem{}
	err := c.db.Select(&rows, custodyCols+` WHERE r."employeeId" = $1 AND r.status = 'APPROVED' AND r."returnedAt" IS NULL ORDER BY r."approvedAt" DESC`, emp)
	return rows, err
}

// All كل العهد المفتوحة (للمدير)، الي تنتظر تدقيق أول.
func (c *ToolCustody) All() ([]ToolCustodyItem, error) {
	rows := []ToolCustodyItem{}
	err := c.db.Select(&rows, custodyCols+` WHERE r.status = 'APPROVED' AND r."returnedAt" IS NULL
		ORDER BY (r."returnRequestedAt" IS NULL), r."returnRequestedAt", e.name LIMIT 500`)
	return rows, err
}

// RequestReturn الموظف يسلّم: صورة (لازم) وملاحظة — تبقى بعهدته لحد التدقيق.
func (c *ToolCustody) RequestReturn(emp, id, photo, note string) error {
	photo = strings.TrimSpace(photo)
	if photo == "" {
		return errors.New("ارفع صورة الأداة وهي مسلّمة")
	}
	res, err := c.db.Exec(`UPDATE "ToolRequest" SET "returnRequestedAt" = now(), "returnPhoto" = $3, "returnNote" = NULLIF(btrim($4), ''), "verifyNote" = NULL
		WHERE id = $1 AND "employeeId" = $2 AND status = 'APPROVED' AND "returnedAt" IS NULL`, id, emp, photo, note)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrDesignHidden
	}
	return nil
}

// Verify المالك/المدير يدقق التسليم: صح ← ترجع للمخزن؛ غلط ← تبقى بعهدته ويا ملاحظة.
func (c *ToolCustody) Verify(by, id string, ok bool, note string) error {
	var pending bool
	if err := c.db.Get(&pending, `SELECT "returnRequestedAt" IS NOT NULL FROM "ToolRequest" WHERE id = $1 AND status = 'APPROVED' AND "returnedAt" IS NULL`, id); err != nil {
		return ErrDesignHidden
	}
	if ok && !pending {
		return errors.New("الموظف بعده ما سلّم (ماكو صورة) — ينتظر تسليمه")
	}
	if !ok {
		if strings.TrimSpace(note) == "" {
			return errors.New("اكتب ليش التسليم مو صحيح حتى يعرف الموظف")
		}
		_, err := c.db.Exec(`UPDATE "ToolRequest" SET "returnRequestedAt" = NULL, "verifyNote" = btrim($2) WHERE id = $1`, id, note)
		return err
	}
	if _, err := c.inv.ReturnToolRequest(id); err != nil {
		return err
	}
	_, err := c.db.Exec(`UPDATE "ToolRequest" SET "verifiedById" = $2, "verifiedAt" = now(), "verifyNote" = NULLIF(btrim($3), '') WHERE id = $1`, id, by, note)
	return err
}

// Tools الأدوات الي بيها كمية (للاختيار).
func (c *ToolCustody) Tools() ([]map[string]any, error) {
	type t struct {
		ID    string `db:"id" json:"id"`
		Name  string `db:"name" json:"name"`
		Avail int    `db:"availableQuantity" json:"available"`
	}
	rows := []t{}
	if err := c.db.Select(&rows, `SELECT id, name, "availableQuantity" FROM "OnDemandTool" WHERE "availableQuantity" > 0 ORDER BY name`); err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{"id": r.ID, "name": r.Name, "available": r.Avail})
	}
	return out, nil
}
