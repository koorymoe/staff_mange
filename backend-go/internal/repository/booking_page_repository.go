package repository

import (
	"fmt"
	"strings"

	"staffmange-api/internal/model"
)

// ═══ الحجوزات صفحة صفحة ═══
//
// «هاي الفلترة هنانه؟ أريدك تسويلي مثلها للحجوزات، حتى لا يضل يحمّل
// السيرفر بتحميل كل الحجوزات — يحمّل جزء جزء».
//
// شاشة الحجوزات كانت تسحب **كل** حجوزات الفلتر بنداء واحد، وتفرزهن
// بالمتصفح على المحطات (بانتظار التثبيت، مثبّت، مكلّف، منجز). يعني
// حتى لو الإداري راح يشوف عشرة، السيرفر يجهّز الآلاف ويمرّرهن بالشبكة
// — ومع تراكم السنين تصير كل فتحة شاشة سحبة ثقيلة على القاعدة
// وعلى تلفون الموظف.
//
// هسه المحطة نفسها تنفلتر **بالسيرفر**، وترجع صفحة وحدة مع العدد
// الكلي (حتى يبان «عرض ١٠ من ٥٠٠»).
//
// ⚠️ شروط المحطات هنا لازم تطابق `inBucket` بالواجهة حرف بحرف. لو
// افترقن، الإداري يشوف عدداً بالعدّاد وقائمة تخالفه — وهذا أسوأ من
// بطء الصفحة. الشروط مكتوبة بنفس ترتيب الواجهة عمداً حتى المقارنة
// بينهن تكون بالعين مباشرة.

// شرط «عليه كادر»: تكليف مسجّل بأي دور.
const hasCrewSQL = `EXISTS (SELECT 1 FROM "BookingAssignment" ba WHERE ba."bookingId" = b.id)`

// شرط «بدا التنفيذ» — نفس `executionStarted` بالواجهة.
const startedSQL = `(b."startedAt" IS NOT NULL OR b."arrivedAt" IS NOT NULL
	OR b.status = 'IN_PROGRESS' OR b.status = 'COMPLETED')`

// bucketCondition يرجّع شرط المحطة، أو نص فاضي للـ«الكل».
func bucketCondition(bucket string) string {
	switch bucket {
	// ١ — انسجّل وما انثبّت بعد: ولا كادر ولا تنفيذ
	case "pending":
		return `b."confirmedAt" IS NULL AND NOT ` + hasCrewSQL + ` AND NOT ` + startedSQL + `
			AND b.status NOT IN ('CANCELLED', 'COMPLETED')`
	// ٢ — انثبّت وينتظر موعداً وكادراً
	case "confirmed":
		return `b."confirmedAt" IS NOT NULL AND NOT ` + hasCrewSQL + ` AND NOT ` + startedSQL + `
			AND b.status NOT IN ('CANCELLED', 'COMPLETED')`
	// ٤ — عليه كادر أو بدا التنفيذ، وما خلص
	// ⚠️ «أو» مو «و»: حجز باشر بيه الليدر بلا تكليف رسمي لازم يبقى مرئي.
	case "assigned":
		return `(` + hasCrewSQL + ` OR ` + startedSQL + `)
			AND b.status NOT IN ('CANCELLED', 'COMPLETED', 'PARTIAL')`
	// ٦ — خلص. المنجز جزئياً **مو** هنا: صارله محطته.
	case "done":
		return `b.status = 'COMPLETED'`
	// تفرّعات المنجز — تنحسب من وجود الفاتورة والتقرير فعلاً
	case "done_full":
		return `b.status = 'COMPLETED' AND ` + hasInvoiceSQL + ` AND ` + hasReportSQL
	case "done_no_invoice":
		return `b.status = 'COMPLETED' AND NOT ` + hasInvoiceSQL + ` AND ` + hasReportSQL
	case "done_no_report":
		return `b.status = 'COMPLETED' AND ` + hasInvoiceSQL + ` AND NOT ` + hasReportSQL
	case "done_no_both":
		return `b.status = 'COMPLETED' AND NOT ` + hasInvoiceSQL + ` AND NOT ` + hasReportSQL
	}
	return ""
}

const hasInvoiceSQL = `EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id)`
const hasReportSQL = `EXISTS (SELECT 1 FROM "WorkReport" wr WHERE wr."bookingId" = b.id)`

// BookingPageQuery شنو تطلبه الشاشة.
type BookingPageQuery struct {
	Bucket   string
	Search   string
	Date     string
	Month    string
	Page     int
	PageSize int
}

// ListPaged يرجّع صفحة وحدة من محطة معيّنة، مع العدد الكلي.
//
// ⚠️ العدّ والصفحة بنفس الشروط بالضبط: لو اختلفن، الترقيم يوعد بصفحات
// ما إلها محتوى.
func (r *BookingRepository) ListPaged(q BookingPageQuery) ([]model.Booking, int, error) {
	where := []string{`b."archivedAt" IS NULL`}
	args := []any{}

	if cond := bucketCondition(q.Bucket); cond != "" {
		where = append(where, `(`+cond+`)`)
	}

	// البحث بالسيرفر: كود الحجز، اسم الزبون، كوده، هاتفه.
	// ⚠️ بدونه البحث يشتغل على **الصفحة الحالية بس** — الإداري يدوّر
	// حجز موجود بصفحة ٧ ويطلعله «ماكو نتيجة».
	if s := strings.TrimSpace(q.Search); s != "" {
		args = append(args, "%"+s+"%")
		i := len(args)
		where = append(where, fmt.Sprintf(`(
			b.code ILIKE $%d
			OR EXISTS (SELECT 1 FROM "Customer" c WHERE c.id = b."customerId"
				AND (c.name ILIKE $%d OR c.phone ILIKE $%d OR c."customerCode"::text ILIKE $%d))
		)`, i, i, i, i))
	}

	// التاريخ: نفس منطق «الموعد الفعلي» (الموعد، وإلا تاريخ التسجيل).
	if q.Date != "" {
		args = append(args, q.Date)
		i := len(args)
		where = append(where, fmt.Sprintf(
			`COALESCE(baghdad_date(b."scheduledAt"), baghdad_date(b."createdAt")) = $%d::date`, i))
	} else if q.Month != "" {
		args = append(args, q.Month)
		i := len(args)
		where = append(where, fmt.Sprintf(
			`to_char(COALESCE(b."scheduledAt", b."createdAt"), 'YYYY-MM') = $%d`, i))
	}

	whereSQL := strings.Join(where, " AND ")

	var total int
	if err := r.db.Get(&total, `SELECT COUNT(*) FROM "Booking" b WHERE `+whereSQL, args...); err != nil {
		return nil, 0, err
	}

	pageSize := q.PageSize
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 10
	}
	page := q.Page
	if page < 1 {
		page = 1
	}

	// ⚠️ الترتيب هنا يطابق الواجهة: طابور الانتظار **الأقدم أول**
	// (الي منتظر أكثر أولى)، وباقي المحطات الأحدث أول.
	order := `ORDER BY COALESCE(b."scheduledAt", b."createdAt") DESC`
	if q.Bucket == "pending" {
		order = `ORDER BY COALESCE(b."scheduledAt", b."createdAt") ASC`
	}

	args = append(args, pageSize, (page-1)*pageSize)
	query := fmt.Sprintf(`SELECT b.* FROM "Booking" b WHERE %s %s LIMIT $%d OFFSET $%d`,
		whereSQL, order, len(args)-1, len(args))

	bookings := []model.Booking{}
	if err := r.db.Select(&bookings, query, args...); err != nil {
		return nil, 0, err
	}
	if err := r.hydrateAll(toPointers(bookings)); err != nil {
		return nil, 0, err
	}
	return bookings, total, nil
}
