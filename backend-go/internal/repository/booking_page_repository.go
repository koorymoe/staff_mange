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
	// ⚠️ المطلوب حذفه ينستثنى من **كل** المحطات ويروح لمحطته: بقاؤه
	// بالطابور يعني الإداري يشتغل على حجز يمكن ينحذف.
	notDeleting := ` AND NOT ` + deletePendingSQL
	// ⚠️ والمحبوس عند المشاريع ينستثنى بعد — نفس السبب: طابور الشغل
	// ما يجوز يمتلئ بحجوزات الإداري ما يكدر يلمسها.
	notDeleting += ` AND NOT ` + atProjectsSQL

	switch bucket {
	// محطة مستقلة: ينتظر قرار المراقب
	case "delete_pending":
		return deletePendingSQL
	// محطة مستقلة: عند إدارة المشاريع لحد ما يوصل التنفيذ
	case "at_projects":
		return atProjectsSQL + ` AND NOT ` + deletePendingSQL
	// ١ — انسجّل وما انثبّت بعد: ولا كادر ولا تنفيذ
	case "pending":
		return `b."confirmedAt" IS NULL AND NOT ` + hasCrewSQL + ` AND NOT ` + startedSQL + `
			AND b.status NOT IN ('CANCELLED', 'COMPLETED')` + notDeleting
	// ٢ — انثبّت وينتظر موعداً وكادراً
	case "confirmed":
		return `b."confirmedAt" IS NOT NULL AND NOT ` + hasCrewSQL + ` AND NOT ` + startedSQL + `
			AND b.status NOT IN ('CANCELLED', 'COMPLETED')` + notDeleting
	// ٤ — عليه كادر أو بدا التنفيذ، وما خلص
	// ⚠️ «أو» مو «و»: حجز باشر بيه الليدر بلا تكليف رسمي لازم يبقى مرئي.
	case "assigned":
		return `(` + hasCrewSQL + ` OR ` + startedSQL + `)
			AND b.status NOT IN ('CANCELLED', 'COMPLETED', 'PARTIAL')` + notDeleting
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

// ═══ حجز مطلوب حذفه ═══
// «الحجوزات الي ينحذفن أريدهن يترحّلن بعد، ينتقلن مرحلة مرحلة، ما
// أريد يضلن بمكان واحد».
//
// ⚠️ ينشال من محطته الطبيعية: الإداري ما يجوز يضيّع وقته يدوّر كادراً
// لحجز يمكن ينحذف بعد ساعة. وإذا انرفض الطلب، يرجع لمحطته لحاله لأن
// المحطة تنحسب من حالته مو من علامة ثابتة.
const deletePendingSQL = `EXISTS (SELECT 1 FROM "BookingDeleteRequest" dr
	WHERE dr."bookingId" = b.id AND dr.status = 'PENDING')`

// ═══ محبوس عند إدارة المشاريع ═══
//
// «الحجوزات الي يترحّلن للكادر… أريدهن يترحّلن بعد، ينتقلن مرحلة
// مرحلة، ما أريد يضلن بمكان واحد» — واختار محطة مستقلة.
//
// الحجز الي انرحّل لإدارة المشاريع يبقى **مقفول** على الإداري لحد ما
// المشرف يوصله مرحلة التنفيذ. فبقاؤه بطابور «تم التثبيت» يزاحم شغلاً
// يكدر يلمسه بحجز ما يكدر يلمسه — والإداري يشوف رقماً بالعدّاد ما
// يقدر يشتغل عليه.
//
// ⚠️ ومحطة مستقلة أفضل من إخفائه: الحجز المحبوس شهر عند المشاريع
// لازم ينشاف وينعدّ، وإلا محد ينتبه إنه واقف.
//
// ⚠️ وما يحتاج علامة ترجعه: أول ما يوصل التنفيذ (`projectExecutionAt`)
// الشرط يصير كذباً لحاله، فيرجع لمحطته الطبيعية تلقائياً.
const atProjectsSQL = `(b."transferToProjects" AND b."projectExecutionAt" IS NULL
	AND b.status NOT IN ('CANCELLED', 'COMPLETED'))`

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

	// ═══ الأحدث أول — بكل المحطات ═══
	//
	// «الحجوزات الي بانتظار التثبيت بيهن مشكلة: الحجز القديم يطلع أول
	// واحد والجديد آخر واحد، لازم ينعكسن — احنا نمشي من الأحدث
	// للأقدم».
	//
	// ⚠️ كنت خليت طابور الانتظار **الأقدم أول** باجتهاد مني (الي منتظر
	// أكثر أولى)، وهذا خالف طريقة شغلهم: الحجز الي وصل توّه هو الي
	// ينتظر تواصل، والقديم أغلبه انعالج. صار الترتيب واحداً بكل
	// المحطات — الأحدث أول.
	order := `ORDER BY COALESCE(b."scheduledAt", b."createdAt") DESC`

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

// ═══ عدّادات المحطات ═══
//
// «خل نضيف ملاعيب وترتيبات وزينة للنظام».
//
// الإداري يفتح الشاشة وما يعرف وين متكدّس الشغل إلا لمن يضغط كل
// محطة وحدة وحدة — تسع ضغطات حتى يعرف من وين يبدي. والرقم على
// الخيار يجاوبه بنظرة.
//
// ⚠️ استعلام **واحد** لكل العدّادات مو تسعة: تسع نداءات بكل فتحة
// شاشة (ومع كل تحديث تلقائي) تصير حملاً أثقل من الي شلناه بالترقيم.
// `COUNT(*) FILTER` يخلّي القاعدة تمرّ على الجدول مرة وحدة.
func (r *BookingRepository) StationCounts() (map[string]int, error) {
	buckets := []string{"pending", "confirmed", "assigned", "done", "at_projects", "delete_pending"}
	sel := []string{}
	for _, b := range buckets {
		// نفس شروط المحطات بالضبط — من نفس المصدر، حتى ما يصير العدّاد
		// يگول رقماً والقائمة تعرض غيره.
		sel = append(sel, fmt.Sprintf(`COUNT(*) FILTER (WHERE %s) AS %q`, bucketCondition(b), b))
	}
	row := map[string]any{}
	q := `SELECT ` + strings.Join(sel, ", ") + ` FROM "Booking" b WHERE b."archivedAt" IS NULL`
	rows, err := r.db.Queryx(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.MapScan(row); err != nil {
			return nil, err
		}
	}
	out := map[string]int{}
	for k, v := range row {
		switch n := v.(type) {
		case int64:
			out[k] = int(n)
		case int:
			out[k] = n
		}
	}

	// «تحتاج إكمال» مصدرها الحالة مو شروط المحطات — والمنجدول منها
	// ينعدّ بعد (شوف شاشة الإكمال).
	var partial int
	if err := r.db.Get(&partial, `
		SELECT COUNT(*) FROM "Booking"
		WHERE "archivedAt" IS NULL AND (status = 'PARTIAL' OR ("partialCount" > 0 AND status = 'CONFIRMED'))
	`); err == nil {
		out["partial"] = partial
	}

	// «ما وصلت للتنفيذ»: الملغى والمؤجل والي ما رد
	var stuck int
	if err := r.db.Get(&stuck, `
		SELECT COUNT(*) FROM "Booking"
		WHERE "archivedAt" IS NULL
		  AND (status = 'CANCELLED' OR "waitingSince" IS NOT NULL OR "awaitingReschedule")
	`); err == nil {
		out["stuck"] = stuck
	}
	return out, nil
}
