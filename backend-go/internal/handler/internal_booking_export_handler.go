package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/xuri/excelize/v2"

	"staffmange-api/internal/repository"
)

// ═══ إكسل حجوزات داخل الشركة ═══
//
// (ع): «نهاية الشهر احنه نريد إكسل خاص بحجوزات داخل الشركة، يكون بي
// كل الفواتير والمبالغ الي محتاجينها».
type InternalBookingExportHandler struct {
	reports *repository.InternalBookingReportRepository
}

func NewInternalBookingExportHandler(reports *repository.InternalBookingReportRepository) *InternalBookingExportHandler {
	return &InternalBookingExportHandler{reports: reports}
}

var internalExportHeaders = []string{
	"التاريخ", "رقم الحجز", "القسم", "صاحب الطلب", "رقمه", "الخدمات",
	"عدد الأجهزة", "الي سوّى الفاتورة", "رقم الفاتورة", "وصف الشغل",
	"المبلغ", "حالة الفاتورة", "الرقم المحاسبي", "ملاحظة الكوادر",
}

// GET /api/internal-bookings/export?month=YYYY-MM
func (h *InternalBookingExportHandler) ExportMonth(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	rows, err := h.reports.Month(month)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	f := excelize.NewFile()
	sheet := "داخل الشركة"
	f.SetSheetName("Sheet1", sheet)
	for i, head := range internalExportHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, head)
	}
	setExcelHeaderStyle(f, sheet, "N")
	bodyStyle, _ := f.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{Horizontal: "right", WrapText: true}})

	total := 0.0
	// 🔴 عدد **الحجوزات** مو عدد الصفوف: الحجز الي عليه فاتورتان يطلع
	// بصفّين، فـ«٣ حجز» على ٣ صفوف تكون **رقماً غلط**. نعدّ الأكواد
	// المتفرّدة، ونكتب الاثنين بالصف الأخير.
	seenBookings := map[string]bool{}
	for i, row := range rows {
		seenBookings[row.BookingCode] = true
		if row.NetTotal != nil {
			total += *row.NetTotal
		}
		date := ""
		if row.BookingDate != nil {
			date = row.BookingDate.Format("2006-01-02")
		}
		values := []interface{}{
			date,
			row.BookingCode,
			dash(row.Department),
			dash(row.RequesterName),
			dash(row.RequesterPhone),
			dash(row.Services),
			dashInt(row.DeviceCount),
			dash(row.LeaderName),
			dash(row.InvoiceCode),
			dash(row.Work),
			// 🔴 الحجز بلا فاتورة يطلع «—» مو صفراً: صفر يعني «شغل
			// بلا كلفة» وهذا **كذب**، والمطلوب يبيّن إنه ناقص ورق.
			dashMoney(row.NetTotal),
			invoiceStatusLabel(row.InvoiceStatus),
			dash(row.ExternalNumber),
			dash(row.HrNote),
		}
		for c, v := range values {
			cell, _ := excelize.CoordinatesToCellName(c+1, i+2)
			_ = f.SetCellValue(sheet, cell, v)
			_ = f.SetCellStyle(sheet, cell, cell, bodyStyle)
		}
	}

	// صف المجموع — هو الي يجي إله المالك أصلاً.
	sumRow := len(rows) + 3
	sumStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "right"},
	})
	labelCell, _ := excelize.CoordinatesToCellName(9, sumRow)
	totalCell, _ := excelize.CoordinatesToCellName(11, sumRow)
	_ = f.SetCellValue(sheet, labelCell, fmt.Sprintf("مجموع الشهر — %d حجز داخلي بـ%d سطر", len(seenBookings), len(rows)))
	_ = f.SetCellValue(sheet, totalCell, total)
	_ = f.SetCellStyle(sheet, labelCell, totalCell, sumStyle)

	_ = f.SetColWidth(sheet, "A", "B", 14)
	_ = f.SetColWidth(sheet, "C", "E", 20)
	_ = f.SetColWidth(sheet, "F", "F", 26)
	_ = f.SetColWidth(sheet, "G", "I", 16)
	_ = f.SetColWidth(sheet, "J", "J", 40)
	_ = f.SetColWidth(sheet, "K", "N", 18)

	filename := fmt.Sprintf("حجوزات-داخل-الشركة-%s.xlsx", month)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, url.QueryEscape(filename)))
	if err := f.Write(w); err != nil {
		WriteError(w, http.StatusInternalServerError, "تعذر إنشاء ملف الإكسل")
		return
	}
}

// 🔴 «رقم غلط أسوأ من ماكو رقم» — الفاضي يطلع «—» بكل خانة.
func dash(v *string) string {
	if v == nil || *v == "" {
		return "—"
	}
	return *v
}

func dashInt(v *int) interface{} {
	if v == nil {
		return "—"
	}
	return *v
}

func dashMoney(v *float64) interface{} {
	if v == nil {
		return "—"
	}
	return *v
}

func invoiceStatusLabel(v *string) string {
	switch {
	case v == nil || *v == "":
		return "ماكو فاتورة"
	case *v == "APPROVED":
		return "معتمدة"
	default:
		return "مقدَّمة — تنتظر المحاسب"
	}
}
