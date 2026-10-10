package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/xuri/excelize/v2"

	"staffmange-api/internal/model"
)

// ═══ تصدير متابعة السيارات إكسل — بنفس شكل ملف (ع) (10-09) ═══
// شيت «النظام الشهري» يوم يوم، وشيت «الإحصائيات» بنفس أعمدة الشاشة.

var scoreFill = []string{"EF4444", "FB923C", "FACC15", "84CC16", "059669"}

type trackVehicle struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

// BuildTrackingWorkbook يبني الملف — منفصل عن الطلب حتى ينفحص بالاختبار.
func BuildTrackingWorkbook(from, to string, vehicles []trackVehicle, ratings []model.VehicleDailyRating) (*excelize.File, error) {
	f := excelize.NewFile()
	names := map[string]string{}
	for _, v := range vehicles {
		names[v.ID] = v.Name
	}
	center, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{Horizontal: "center"}})
	right, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{Horizontal: "right", WrapText: true}})
	fills := make([]int, len(scoreFill))
	for i, c := range scoreFill {
		fills[i], _ = f.NewStyle(&excelize.Style{
			Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{c}},
			Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
			Alignment: &excelize.Alignment{Horizontal: "center"},
		})
	}
	set := func(sheet string, col, row int, v any, style int) {
		cell, _ := excelize.CoordinatesToCellName(col, row)
		_ = f.SetCellValue(sheet, cell, v)
		_ = f.SetCellStyle(sheet, cell, cell, style)
	}
	rtl := true
	days := []string{"الأحد", "الاثنين", "الثلاثاء", "الأربعاء", "الخميس", "الجمعة", "السبت"}

	// ── النظام الشهري ──
	daily := "النظام الشهري"
	f.SetSheetName("Sheet1", daily)
	_ = f.SetSheetView(daily, 0, &excelize.ViewOptions{RightToLeft: &rtl})
	head := []string{"التاريخ", "اليوم", "السيارة", "الغسل"}
	for _, it := range trackItems {
		head = append(head, fmt.Sprintf("%s (%d%%)", it.Label, int(it.Weight*100+0.5)))
	}
	head = append(head, "وصف العطل", "التقييم الموزون %")
	for i, h := range head {
		set(daily, i+1, 1, h, 0)
	}
	lastCol, _ := excelize.ColumnNumberToName(len(head))
	setExcelHeaderStyle(f, daily, lastCol)
	for i := range ratings {
		r := &ratings[i]
		row := i + 2
		set(daily, 1, row, r.RatedDate.Format("2006-01-02"), center)
		set(daily, 2, row, days[r.RatedDate.Weekday()], center)
		set(daily, 3, row, names[r.VehicleID], right)
		wash := ""
		if r.Wash != nil {
			wash = "لم يتم"
			if *r.Wash >= 2 {
				wash = "تم"
			}
		}
		set(daily, 4, row, wash, center)
		for j, it := range trackItems {
			if v := it.get(r); v != nil && *v >= 0 && *v <= 4 {
				set(daily, 5+j, row, *v, fills[*v])
			}
		}
		desc := ""
		if r.FaultDescription != nil {
			desc = *r.FaultDescription
		}
		set(daily, 5+len(trackItems), row, desc, right)
		if sc := DailyScore(r); sc != nil {
			set(daily, 6+len(trackItems), row, fmt.Sprintf("%.1f%%", *sc), center)
		}
	}
	_ = f.SetColWidth(daily, "A", "B", 12)
	_ = f.SetColWidth(daily, "C", "C", 24)
	_ = f.SetColWidth(daily, "D", lastCol, 14)
	descCol, _ := excelize.ColumnNumberToName(5 + len(trackItems))
	_ = f.SetColWidth(daily, descCol, descCol, 40)

	// ── الإحصائيات ──
	stats := "الإحصائيات"
	if _, err := f.NewSheet(stats); err != nil {
		return nil, err
	}
	_ = f.SetSheetView(stats, 0, &excelize.ViewOptions{RightToLeft: &rtl})
	sh := []string{"السيارة", "عدد التقييمات", "مجموع النقاط", "الحد الأعلى", "النسبة الخام", "تم الغسل %"}
	for _, it := range trackItems {
		sh = append(sh, it.Label+" %")
	}
	sh = append(sh, "النتيجة الموزونة", "التقدير", "تم الغسل", "لم يتم")
	for i, h := range sh {
		set(stats, i+1, 1, h, 0)
	}
	sLast, _ := excelize.ColumnNumberToName(len(sh))
	setExcelHeaderStyle(f, stats, sLast)
	// سطر الأوزان
	set(stats, 1, 2, "الوزن", center)
	set(stats, 6, 2, fmt.Sprintf("%d%%", 8), center)
	for j, it := range trackItems {
		set(stats, 7+j, 2, fmt.Sprintf("%d%%", int(it.Weight*100+0.5)), center)
	}
	byV := map[string][]model.VehicleDailyRating{}
	for _, r := range ratings {
		byV[r.VehicleID] = append(byV[r.VehicleID], r)
	}
	pct := func(p *float64) string {
		if p == nil {
			return "—"
		}
		return fmt.Sprintf("%.1f%%", *p)
	}
	for i, v := range vehicles {
		st := BuildTrackStats(v.ID, v.Name, byV[v.ID])
		row := i + 3
		vals := []any{st.VehicleName, st.Completed, st.RawSum, st.RawMax, pct(st.RawPct), pct(st.WashPct)}
		for _, it := range trackItems {
			vals = append(vals, pct(st.Items[it.Key]))
		}
		vals = append(vals, pct(st.Weighted), st.Grade, st.WashDone, st.WashNot)
		for c, val := range vals {
			style := center
			if c == 0 {
				style = right
			}
			set(stats, c+1, row, val, style)
		}
	}
	_ = f.SetColWidth(stats, "A", "A", 24)
	_ = f.SetColWidth(stats, "B", sLast, 14)
	_ = f.SetCellValue(stats, fmt.Sprintf("A%d", len(vehicles)+5), fmt.Sprintf("الفترة: %s ← %s", from, to))
	return f, nil
}

// GET /api/vehicles/tracking/export?period=YYYY-MM
func (h *VehicleTrackingHandler) Export(w http.ResponseWriter, r *http.Request) {
	from, to := periodRange(r.URL.Query().Get("period"))
	vehicles := []trackVehicle{}
	if err := h.db.Select(&vehicles, `SELECT id, name FROM "Vehicle" WHERE "isActive" ORDER BY "createdAt", name`); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب السيارات")
		return
	}
	ratings := []model.VehicleDailyRating{}
	if err := h.db.Select(&ratings, `SELECT r.* FROM "VehicleDailyRating" r JOIN "Vehicle" v ON v.id = r."vehicleId" AND v."isActive"
		WHERE r."ratedDate" BETWEEN $1::date AND $2::date ORDER BY r."ratedDate", v."createdAt", v.name`,
		from.Format("2006-01-02"), to.Format("2006-01-02")); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر جلب التقييمات")
		return
	}
	f, err := BuildTrackingWorkbook(from.Format("2006-01-02"), to.Format("2006-01-02"), vehicles, ratings)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر إنشاء ملف الإكسل")
		return
	}
	filename := fmt.Sprintf("متابعة-السيارات-%s.xlsx", from.Format("2006-01"))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename*=UTF-8''%s`, url.PathEscape(filename)))
	_ = f.Write(w)
}
