package handler

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/xuri/excelize/v2"
)

// ═══ التقرير الشهري للمشتريات — قرار (ع) 10-15 ═══
// قديش صرفنا لكل صنف ولكل مورد ولكل مشتري، والمقارنة ويّا الشهر السابق،
// وقديش چان ممكن نوفّر لو اشترينا كل مادة من أرخص مورد (آخر ١٨٠ يوم قبل الشراء).

type purchaseLine struct {
	At           time.Time `db:"at" json:"at"`
	Code         string    `db:"code" json:"code"`
	Product      string    `db:"product" json:"product"`
	Qty          int       `db:"qty" json:"qty"`
	UnitPrice    float64   `db:"unitPrice" json:"unitPrice"`
	Supplier     string    `db:"supplier" json:"supplier"`
	Category     string    `db:"category" json:"category"`
	Buyer        string    `db:"buyer" json:"buyer"`
	BestPrice    *float64  `db:"bestPrice" json:"bestPrice"`
	BestSupplier *string   `db:"bestSupplier" json:"bestSupplier"`
	Saving       float64   `db:"saving" json:"saving"`
}

type purchaseGroup struct {
	Name   string  `json:"name"`
	Total  float64 `json:"total"`
	Count  int     `json:"count"`
	Saving float64 `json:"saving"`
}

type purchaseMonth struct {
	Month      string          `json:"month"`
	Total      float64         `json:"total"`
	Count      int             `json:"count"`
	Requests   int             `json:"requests"`
	PrevTotal  float64         `json:"prevTotal"`
	PrevCount  int             `json:"prevCount"`
	Saving     float64         `json:"saving"`
	Alerts     int             `json:"alerts"`
	Unverified int             `json:"unverified"`
	ByCategory []purchaseGroup `json:"byCategory"`
	BySupplier []purchaseGroup `json:"bySupplier"`
	ByBuyer    []purchaseGroup `json:"byBuyer"`
	Lines      []purchaseLine  `json:"lines"`
}

const purchaseLinesSQL = `SELECT r."fulfilledAt" AS at, r.code, i."productName" AS product, i.quantity AS qty, i."unitPrice",
	COALESCE(s."companyName", '—') AS supplier,
	COALESCE((SELECT sp.name FROM "SupplierSpecialtyLink" l JOIN "SupplierSpecialty" sp ON sp.id = l."specialtyId"
		WHERE l."supplierId" = r."supplierId" ORDER BY sp."order", sp.name LIMIT 1), 'بلا صنف') AS category,
	COALESCE(e.name, '—') AS buyer, b.price AS "bestPrice", b.name AS "bestSupplier",
	GREATEST(COALESCE(i."unitPrice" - b.price, 0), 0) * i.quantity AS saving
	FROM "ProcurementItem" i JOIN "ProcurementRequest" r ON r.id = i."requestId"
	LEFT JOIN "Supplier" s ON s.id = r."supplierId" LEFT JOIN "Employee" e ON e.id = r."fulfilledById"
	LEFT JOIN LATERAL (
		SELECT i2."unitPrice" AS price, s2."companyName" AS name FROM "ProcurementItem" i2
		JOIN "ProcurementRequest" r2 ON r2.id = i2."requestId" JOIN "Supplier" s2 ON s2.id = r2."supplierId"
		WHERE product_key(i2."productName", i2.model) = product_key(i."productName", i.model) AND r2."supplierId" <> r."supplierId"
		  AND i2."unitPrice" > 0 AND r2."fulfilledAt" BETWEEN r."fulfilledAt" - interval '180 days' AND r."fulfilledAt"
		ORDER BY i2."unitPrice" LIMIT 1) b ON true
	WHERE r."fulfilledAt" >= $1 AND r."fulfilledAt" < $2 AND i."unitPrice" > 0
	ORDER BY r."fulfilledAt"`

func (h *PurchaseControlHandler) buildMonth(month string) (*purchaseMonth, error) {
	loc := baghdadLoc()
	start, err := time.ParseInLocation("2006-01", month, loc)
	if err != nil {
		start, _ = time.ParseInLocation("2006-01", time.Now().In(loc).Format("2006-01"), loc)
	}
	end := start.AddDate(0, 1, 0)
	m := &purchaseMonth{Month: start.Format("2006-01"), Lines: []purchaseLine{}}
	if err := h.db.Select(&m.Lines, purchaseLinesSQL, start, end); err != nil {
		return nil, err
	}
	// المصروف الحقيقي على مستوى الطلب (التكلفة الكلية إذا مسجّلة، وإلا مجموع المواد)
	spend := func(a, b time.Time) (float64, int) {
		var x struct {
			T float64 `db:"t"`
			N int     `db:"n"`
		}
		_ = h.db.Get(&x, `SELECT COALESCE(sum(COALESCE(r."totalCost", (SELECT sum(COALESCE(i."totalPrice", i."unitPrice" * i.quantity)) FROM "ProcurementItem" i WHERE i."requestId" = r.id), 0)), 0) AS t,
			count(*)::int AS n FROM "ProcurementRequest" r WHERE r."fulfilledAt" >= $1 AND r."fulfilledAt" < $2`, a, b)
		return x.T, x.N
	}
	m.Total, m.Requests = spend(start, end)
	// الشهر السابق لنفس اليوم (إذا الشهر الحالي ما خلص) — مقارنة عادلة
	prevEnd := end.AddDate(0, -1, 0)
	if now := time.Now().In(loc); now.Before(end) {
		prevEnd = now.AddDate(0, -1, 0)
	}
	m.PrevTotal, m.PrevCount = spend(start.AddDate(0, -1, 0), prevEnd)
	_ = h.db.Get(&m.Alerts, `SELECT count(*) FROM "ProcurementPriceAlert" WHERE "createdAt" >= $1 AND "createdAt" < $2`, start, end)
	_ = h.db.Get(&m.Unverified, `SELECT count(*) FROM "ProcurementRequest" r WHERE r."fulfilledAt" >= $1 AND r."fulfilledAt" < $2
		AND NOT EXISTS (SELECT 1 FROM "ProcurementVerify" v WHERE v."procurementId" = r.id)`, start, end)
	group := func(key func(purchaseLine) string) []purchaseGroup {
		idx := map[string]int{}
		out := []purchaseGroup{}
		for _, l := range m.Lines {
			k := key(l)
			i, ok := idx[k]
			if !ok {
				i = len(out)
				idx[k] = i
				out = append(out, purchaseGroup{Name: k})
			}
			out[i].Total += l.UnitPrice * float64(l.Qty)
			out[i].Count++
			out[i].Saving += l.Saving
		}
		for i := 1; i < len(out); i++ { // ترتيب تنازلي بالمصروف
			for j := i; j > 0 && out[j].Total > out[j-1].Total; j-- {
				out[j], out[j-1] = out[j-1], out[j]
			}
		}
		return out
	}
	m.ByCategory = group(func(l purchaseLine) string { return l.Category })
	m.BySupplier = group(func(l purchaseLine) string { return l.Supplier })
	m.ByBuyer = group(func(l purchaseLine) string { return l.Buyer })
	m.Count = len(m.Lines)
	for _, l := range m.Lines {
		m.Saving += l.Saving
	}
	return m, nil
}

// GET /api/procurement-monthly?month=YYYY-MM[&format=xlsx]
func (h *PurchaseControlHandler) Monthly(w http.ResponseWriter, r *http.Request) {
	m, err := h.buildMonth(r.URL.Query().Get("month"))
	if err != nil {
		log.Printf("[purchase-report] %v", err)
		WriteError(w, http.StatusInternalServerError, "تعذر إعداد التقرير")
		return
	}
	if r.URL.Query().Get("format") != "xlsx" {
		WriteJSON(w, http.StatusOK, m)
		return
	}
	f := excelize.NewFile()
	head, _ := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "FFFFFF"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"0F2040"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"}})
	rtl := true
	sheet := func(name string, headers []string, data [][]any, widths []float64) {
		if name != "Sheet1" {
			_, _ = f.NewSheet(name)
		}
		_ = f.SetSheetView(name, 0, &excelize.ViewOptions{RightToLeft: &rtl})
		for i, hd := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			_ = f.SetCellValue(name, cell, hd)
			_ = f.SetCellStyle(name, cell, cell, head)
			col, _ := excelize.ColumnNumberToName(i + 1)
			_ = f.SetColWidth(name, col, col, widths[i])
		}
		for ri, row := range data {
			for ci, v := range row {
				cell, _ := excelize.CoordinatesToCellName(ci+1, ri+2)
				_ = f.SetCellValue(name, cell, v)
			}
		}
	}
	_ = f.SetSheetName("Sheet1", "الملخص")
	sheet("الملخص", []string{"البند", "القيمة"}, [][]any{
		{"الشهر", m.Month}, {"المصروف", m.Total}, {"عدد المشتريات", m.Requests}, {"عدد المواد", m.Count},
		{"الشهر السابق (لنفس اليوم)", m.PrevTotal}, {"مشتريات الشهر السابق", m.PrevCount},
		{"چان ممكن نوفّر (لو من الأرخص)", m.Saving}, {"تنبيهات ماتركس", m.Alerts}, {"ما ثبّتها المحاسب", m.Unverified},
	}, []float64{34, 20})
	grp := func(g []purchaseGroup) [][]any {
		out := [][]any{}
		for _, x := range g {
			out = append(out, []any{x.Name, x.Total, x.Count, x.Saving})
		}
		return out
	}
	hd := []string{"الاسم", "المصروف", "مواد", "چان ممكن نوفّر"}
	wd := []float64{30, 16, 10, 18}
	sheet("حسب الصنف", hd, grp(m.ByCategory), wd)
	sheet("حسب المورد", hd, grp(m.BySupplier), wd)
	sheet("حسب المشتري", hd, grp(m.ByBuyer), wd)
	lines := [][]any{}
	for _, l := range m.Lines {
		best, bestSup := "", ""
		if l.BestPrice != nil {
			best = fmt.Sprintf("%.0f", *l.BestPrice)
		}
		if l.BestSupplier != nil {
			bestSup = *l.BestSupplier
		}
		lines = append(lines, []any{l.At.In(baghdadLoc()).Format("2006-01-02"), l.Code, l.Product, l.Qty, l.UnitPrice, l.UnitPrice * float64(l.Qty),
			l.Supplier, l.Category, l.Buyer, best, bestSup, l.Saving})
	}
	sheet("التفاصيل", []string{"التاريخ", "الطلب", "المادة", "الكمية", "سعر الوحدة", "الإجمالي", "المورد", "الصنف", "المشتري", "أرخص سعر", "عند", "الفرق"},
		lines, []float64{12, 12, 30, 8, 13, 14, 22, 16, 18, 13, 22, 13})
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape("المشتريات-"+m.Month+".xlsx"))
	_ = f.Write(w)
}
