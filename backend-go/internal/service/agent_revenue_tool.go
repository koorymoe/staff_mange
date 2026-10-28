package service

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
)

// ═══ 💵 فلوس كل مسؤول (دازاي 10-28) ═══
// «من اسأله عن الموظفين الي دخلو فلوس… القائد مال كل حجز فقط ينحط اسمه والمبالغ
// الي دخلها، واذا فني فقط بدون ليدر». والمسؤول عن فلوس الحجز:
//   - حجز داخلي ← إداري الحجوزات (الي سوّى الحجز)
//   - جي بي اس / داش كام ← مسؤول الخدمة (الي يكتب فاتورتها)
//   - غيرها ← ليدر الحجز، وإذا ماكو ليدر ← الفني
// المبلغ من فاتورة الحجز (صافي آخر فاتورة)، ومنو دخّلها من الفاتورة نفسها.

var dayRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func (a *MatrixAgent) registerRevenueTool() {
	a.AddTool(agentTool{Name: "revenue_by_responsible",
		Desc: "الفلوس الي دخلها كل مسؤول بفترة (افتراضياً اليوم): لكل حجز منجز مسؤول واحد بس — الليدر (أو الفني إذا ماكو ليدر)، وللحجز الداخلي إداري الحجوزات، وللجي بي اس والداش كام مسؤول الخدمة. يرجّع لكل مسؤول: عدد الحجوزات والمبلغ الي دخّله، والحجوزات المنجزة الي بعدها بلا فاتورة. استعملها لأي سؤال «شكد دخّل كل موظف/ليدر» — لا تحسب الفلوس على كل الكادر.",
		Props: map[string]any{
			"from": map[string]any{"type": "string", "description": "من تاريخ YYYY-MM-DD (فاضي = اليوم)"},
			"to":   map[string]any{"type": "string", "description": "لحد تاريخ YYYY-MM-DD (فاضي = نفس from)"},
		},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			from, to := strArg(in, "from"), strArg(in, "to")
			if !dayRe.MatchString(from) {
				from = bagh(time.Now()).Format("2006-01-02")
			}
			if !dayRe.MatchString(to) {
				to = from
			}
			rows := []struct {
				Code     string  `db:"code"`
				Type     string  `db:"bookingType"`
				Kind     string  `db:"kind"`
				InvBy    *string `db:"invBy"`
				Net      float64 `db:"net"`
				Expected *string `db:"expected"`
			}{}
			err := a.db.Select(&rows, `
				WITH bk AS (
					SELECT b.id, b.code, b."bookingType", COALESCE(s."serviceKind", '') AS kind, b."createdById",
						booking_leader_id(b.id) AS leader,
						(SELECT ba."employeeId" FROM "BookingAssignment" ba WHERE ba."bookingId" = b.id ORDER BY ba."createdAt" LIMIT 1) AS tech
					FROM "Booking" b LEFT JOIN "Service" s ON s.id = b."serviceId"
					WHERE b.status = 'COMPLETED' AND b."settledLegacyAt" IS NULL AND b."archivedAt" IS NULL
					  AND (b."completedAt" AT TIME ZONE 'Asia/Baghdad')::date BETWEEN $1::date AND $2::date
				), inv AS (
					SELECT DISTINCT ON (li."bookingId") li."bookingId", li."employeeId", COALESCE(li."netTotal", 0) AS net
					FROM "LeaderInvoice" li WHERE li."bookingId" IN (SELECT id FROM bk)
					ORDER BY li."bookingId", li."createdAt" DESC
				)
				SELECT bk.code, bk."bookingType", bk.kind, inv."employeeId" AS "invBy", COALESCE(inv.net, 0) AS net,
					CASE WHEN bk."bookingType" = 'INTERNAL' THEN bk."createdById" ELSE COALESCE(bk.leader, bk.tech) END AS expected
				FROM bk LEFT JOIN inv ON inv."bookingId" = bk.id`, from, to)
			if err != nil {
				return "", err
			}
			if len(rows) == 0 {
				return fmt.Sprintf("ماكو حجوزات منجزة من %s لحد %s.", from, to), nil
			}
			type agg struct {
				role          string
				n             int
				total         float64
				codes, nobill []string
			}
			byID := map[string]*agg{}
			ids := []string{}
			get := func(id, role string) *agg {
				if byID[id] == nil {
					byID[id] = &agg{role: role}
					ids = append(ids, id)
				}
				return byID[id]
			}
			var grand float64
			for _, r := range rows {
				role := "ليدر"
				switch {
				case r.Type == "INTERNAL":
					role = "إداري الحجوزات (حجز داخلي)"
				case r.Kind == "GPS" || r.Kind == "DASHCAM":
					role = "مسؤول الخدمة (جي بي اس/داش كام)"
				}
				var id string
				switch {
				case r.InvBy != nil:
					id = *r.InvBy
				case r.Kind == "GPS" || r.Kind == "DASHCAM":
					id = "SERVICE:" + r.Kind
				case r.Expected != nil:
					id = *r.Expected
				default:
					id = "NONE"
				}
				g := get(id, role)
				if r.InvBy != nil {
					g.n++
					g.total += r.Net
					g.codes = append(g.codes, r.Code)
					grand += r.Net
				} else {
					g.nobill = append(g.nobill, r.Code)
				}
			}
			names := map[string]string{}
			leader := map[string]bool{}
			emp := []struct {
				ID     string `db:"id"`
				Name   string `db:"name"`
				Leader bool   `db:"isLeader"`
			}{}
			_ = a.db.Select(&emp, `SELECT id, name, "isLeader" FROM "Employee" WHERE id = ANY($1)`, pq.Array(ids))
			for _, e := range emp {
				names[e.ID], leader[e.ID] = e.Name, e.Leader
			}
			sort.SliceStable(ids, func(i, j int) bool { return byID[ids[i]].total > byID[ids[j]].total })
			var b strings.Builder
			fmt.Fprintf(&b, "من %s لحد %s: %d حجز منجز، المدخول بالفواتير %s.\n", from, to, len(rows), money(grand))
			for _, id := range ids {
				g := byID[id]
				name := names[id]
				switch {
				case id == "NONE":
					name = "بلا مسؤول (ماكو كادر مسجّل)"
				case strings.HasPrefix(id, "SERVICE:"):
					name = "مسؤول الخدمة"
				case name == "":
					name = "موظف غير معروف"
				}
				role := g.role
				if role == "ليدر" && id != "NONE" && !strings.HasPrefix(id, "SERVICE:") && !leader[id] {
					role = "فني (بلا ليدر)"
				}
				if id == "NONE" {
					fmt.Fprintf(&b, "- %s", name)
				} else {
					fmt.Fprintf(&b, "- %s — %s: %d حجز، دخّل %s", name, role, g.n, money(g.total))
				}
				if len(g.codes) > 0 {
					fmt.Fprintf(&b, " (%s)", strings.Join(g.codes, "، "))
				}
				if len(g.nobill) > 0 {
					fmt.Fprintf(&b, "؛ منجز بلا فاتورة بعد: %s", strings.Join(g.nobill, "، "))
				}
				b.WriteString("\n")
			}
			return b.String(), nil
		}})
}
