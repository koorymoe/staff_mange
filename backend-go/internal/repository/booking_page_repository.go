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
	// ⚠️⚠️ **والمتوقّف ينستثنى من محطتَي الاتصال والتثبيت** — العيب الي
	// اشتكوا منه الموظفون:
	//
	// «الحجز الي الزبون مالته ما يرد المفروض ينتقل من بانتظار التثبيت
	//  إلى ما وصلت للتنفيذ. الي جاي يصير: تنتقل نسخة لما وصل للتنفيذ
	//  وتبقى نسخة بانتظار التثبيت، ويرجع الموظف الغافل يتصل ع الزبون
	//  من جديد وما يدري».
	//
	// السبب: التأشير يحط `status='WAITING'` و`waitingSince`، بس هالسلّة
	// چانت تستثني الملغى والمنجز **وبس** — فالحجز يبقى هنا (`confirmedAt`
	// لسه فاضي وماكو كادر) **ويطلع بمحطة «ما وصلت للتنفيذ»** بنفس الوقت.
	// نسختان لنفس الحجز، فالزبون ينتصل عليه مرتين وثلاثة.
	//
	// القاعدة الصح: **الحجز يكون بمحطة وحدة، مو بثنتين.**
	//
	// ⚠️ والمؤجَّل معاه بقرار (ع): «انقله بعد — محطة وحدة بس».
	//
	// ⚠️ وما ينضاع ولا ينحبس: `ResumeFromWaiting` تصفّي `waitingSince`،
	// وتحديد موعد جديد يطفّي `awaitingReschedule` — فالحجز **يرجع
	// لمحطته لحاله** أول ما الزبون يرد أو ينتحدد موعده.
	//
	// ⚠️ ومحصورة بـ«بانتظار التثبيت» و«تم التثبيت» — **مو «مكلّف»**:
	// حجز عليه كادر مجدول لازم يبقى مرئي بطابور الكادر حتى لو الزبون
	// ما رد، وإلا الكادر يطلع لموقع ما أحد يعرف بيه.
	notStuck := ` AND b."waitingSince" IS NULL AND NOT b."awaitingReschedule"`

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
			AND b.status NOT IN ('CANCELLED', 'COMPLETED')` + notDeleting + notStuck
	// ٢ — انثبّت وينتظر موعداً وكادراً
	case "confirmed":
		return `b."confirmedAt" IS NOT NULL AND NOT ` + hasCrewSQL + ` AND NOT ` + startedSQL + `
			AND b.status NOT IN ('CANCELLED', 'COMPLETED')` + notDeleting + notStuck
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
		return `b.status = 'COMPLETED' AND ` + paperworkDoneSQL
	case "done_no_invoice":
		return `b.status = 'COMPLETED' AND NOT ` + hasInvoiceSQL +
			` AND (` + hasReportSQL + ` OR ` + isLegacyImportSQL + `) AND NOT ` + isSurveySQL
	// ⚠️ التاريخي ينستثنى من محطتَي «ناقصه تقرير»: تقريره مستحيل، فبقاؤه
	// هنا يعني طابوراً محد يگدر يفرغه. وحجز الكشف ينستثنى من الثلاثة:
	// ما يحتاج فاتورة ولا تقرير أصلاً.
	case "done_no_report":
		return `b.status = 'COMPLETED' AND ` + hasInvoiceSQL + ` AND NOT ` + hasReportSQL +
			` AND NOT (` + isLegacyImportSQL + `) AND NOT ` + isSurveySQL
	case "done_no_both":
		return `b.status = 'COMPLETED' AND NOT ` + hasInvoiceSQL + ` AND NOT ` + hasReportSQL +
			` AND NOT (` + isLegacyImportSQL + `) AND NOT ` + isSurveySQL
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
// ⚠️ التعريف انتقل لـbooking_delete_filter.go حتى تستعمله كل المصادر
// الثانية (التنسيق، الرئيسية، اليوم، المهام) — قبل چان هنا وحده،
// فالحجز ينختفي من هاي الشاشة ويضل ظاهراً بالباقي.
var deletePendingSQL = BookingDeletePendingSQL("b")

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

// ═══ شغل صار قبل النظام: فاتورة بس، بلا تقرير ═══
//
// حجوزات الاستيراد التاريخي وصلت من النظام القديم وهي منجزة أصلاً.
// مبالغها موجودة بدفتر تدقيق الحسابات فالفاتورة تنكتب وتنعدّ — بس
// **التقرير مستحيل**: منو طلع، وشنو شغّل، وشكد صرف… ولا وحدة منهن
// انوثقت ولا راح تنوثق. فطلب تقرير منهن يخلّي ٢١٩٠ حجز واقفين بطابور
// «ناقصها ورق» للأبد يزاحمون شغلاً حقيقياً يكدر أحد يكمّله.
//
// ⚠️ العلامة كود الحجز: التاريخي 'OLD-…' والعادي 'B<رقم>'
// (booking_service.go) — فما ينخلطون ولا يحتاج عمود جديد.
//
// ⚠️ وليش ما نأشّرهن «مسوّاة إدارياً» (settledLegacyAt) وخلص؟ لأن
// التسوية تقفل الحجز **بلا مبلغ**، والمالك يريد يدقّق شكد انستلم من
// كل حجز. فالفاتورة لازم تنكتب وتنحسب بالإيراد، والناقص هو التقرير بس.
const isLegacyImportSQL = `b.code LIKE 'OLD-%'`

// ═══ زيارة معاينة («كشف»): بلا فاتورة وبلا تقرير أصلاً — قرار عمل ═══
// الإداري يأشّرها وقت التكليف: الكادر ما يسوي فاتورة ولا تقرير عمل
// لهذا الحجز، فبقاؤها بطابور «ناقصها ورق» ظلم — استثناء كامل، مو
// جزئي متل isLegacyImportSQL الي يعفي التقرير بس ويبقي الفاتورة.
const isSurveySQL = `b."bookingType" = 'SURVEY'`

// paperworkDoneSQL — ورق الحجز مكتمل: فاتورة دائماً، وتقرير إلا إذا
// چان شغلاً قبل النظام، أو أصلاً زيارة كشف ما تحتاج ورقاً إطلاقاً.
//
// ⚠️ ينكتب مرة وحدة وينستعمل بكل مكان يسأل «ناقصه ورق؟»: نسختان
// تفترقان بأول تعديل، فتطلع الشاشة رقماً والطابور رقماً ثانياً.
const paperworkDoneSQL = `((` + hasInvoiceSQL + ` AND (` + hasReportSQL + ` OR ` + isLegacyImportSQL + `)) OR ` + isSurveySQL + `)`

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

	// «ما وصلت للتنفيذ»: الملغى والمؤجل والي ما رد.
	// 🔴 وانتظار قرار الزبون **مستثنى**: إله تبويبه الخاص، ولو انعدّ
	// بالاثنين يطلع نفس الحجز بمكانين والمجموع أكبر من الحقيقة.
	var stuck int
	if err := r.db.Get(&stuck, `
		SELECT COUNT(*) FROM "Booking"
		WHERE "archivedAt" IS NULL
		  AND (status = 'CANCELLED' OR "awaitingReschedule"
		       OR ("waitingSince" IS NOT NULL AND COALESCE("waitingKind", 'NO_ANSWER') <> 'CUSTOMER_DECISION'))
	`); err == nil {
		out["stuck"] = stuck
	}

	// «بانتظار موافقة الزبون» — استفسر ورايح يرجعلنا خبر
	var awaitingCustomer int
	if err := r.db.Get(&awaitingCustomer, `
		SELECT COUNT(*) FROM "Booking"
		WHERE "archivedAt" IS NULL AND status <> 'CANCELLED'
		  AND "waitingSince" IS NOT NULL AND "waitingKind" = 'CUSTOMER_DECISION'
	`); err == nil {
		out["awaitingCustomer"] = awaitingCustomer
	}
	return out, nil
}

// ═══ وين الحجز؟ ═══
//
// «يجي الموظف يبحث عن الحجز — وين يبحث؟ بالحجوزات المثبتة، والحجز
// بعده ما متثبت. أريد النظام يساعد الموظف: من يبحث بكود حجز بمكان
// غلط، يكله هذا الحجز بفلان مكان، ابحث عنه هناك».
//
// وهاي أذكى من زيادة سرعة البحث: المشكلة مو إن الموظف ما لگاه، هي
// إنه ما يعرف **وين يدوّر**. عشر محطات، وهو يفتحهن وحدة وحدة —
// وأغلب الوقت ينتهي يظن الحجز انحذف ويسجّله من جديد.
//
// ⚠️ نجاوب بالمحطة الي **الحجز بيها فعلاً** — نحسبها من نفس شروط
// المحطات، مو من تخمين. ولو الحجز مو موجود إطلاقاً، نگول «مو
// موجود» صراحة بدل ما نرجّع فراغ يفسّره الموظف كيف ما جان.
type BookingLocation struct {
	ID       string  `db:"id" json:"id"`
	Code     string  `db:"code" json:"code"`
	Customer *string `db:"customerName" json:"customerName"`
	Station  string  `db:"-" json:"station"`
}

// Locate يلگه الحجز بالكود (أو جزء منه) ويگول بأي محطة هو.
func (r *BookingRepository) Locate(term string) ([]BookingLocation, error) {
	term = strings.TrimSpace(term)
	if term == "" {
		return nil, nil
	}
	// الترتيب مهم: أول محطة يطابقها الحجز هي جوابه، فنمشي بترتيب
	// المسار حتى ما يطلع «منجز» لحجز مطلوب حذفه مثلاً.
	stations := []struct{ key, label string }{
		{"delete_pending", "بانتظار قرار الحذف"},
		{"at_projects", "عند إدارة المشاريع"},
		{"pending", "بانتظار التثبيت — بحاجة لتنسيق"},
		{"confirmed", "تم التثبيت — بحاجة لكادر"},
		{"assigned", "مكلّف — بانتظار التنفيذ"},
		{"done", "تم الإنجاز"},
	}
	sel := []string{}
	for _, st := range stations {
		sel = append(sel, fmt.Sprintf(`(%s) AS %q`, bucketCondition(st.key), st.key))
	}

	rows, err := r.db.Queryx(`
		SELECT b.id, b.code, c.name AS "customerName",
		       b.status::text AS status,
		       (b.status = 'PARTIAL' OR (b."partialCount" > 0 AND b.status = 'CONFIRMED')) AS partial,
		       (b.status = 'CANCELLED' OR b."awaitingReschedule"
		        OR (b."waitingSince" IS NOT NULL AND COALESCE(b."waitingKind", 'NO_ANSWER') <> 'CUSTOMER_DECISION')) AS stuck,
		       (b.status <> 'CANCELLED' AND b."waitingSince" IS NOT NULL
		        AND b."waitingKind" = 'CUSTOMER_DECISION') AS "awaitingCustomer",
		       `+strings.Join(sel, ", ")+`
		FROM "Booking" b
		LEFT JOIN "Customer" c ON c.id = b."customerId"
		WHERE b."archivedAt" IS NULL AND (b.code ILIKE $1 OR c.phone ILIKE $1 OR c.name ILIKE $1)
		ORDER BY b."createdAt" DESC LIMIT 5`, "%"+term+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []BookingLocation{}
	for rows.Next() {
		m := map[string]any{}
		if err := rows.MapScan(m); err != nil {
			return nil, err
		}
		loc := BookingLocation{Station: "بالأرشيف أو خارج المحطات"}
		if v, ok := m["id"].(string); ok {
			loc.ID = v
		}
		if v, ok := m["code"].(string); ok {
			loc.Code = v
		}
		if v, ok := m["customerName"].(string); ok {
			loc.Customer = &v
		}
		// ⚠️ «تحتاج إكمال» و«ما وصلت» يسبقن الباقي بالفحص: حالتهن
		// تغلب (حجز منجز جزئياً مو «مكلّف»، والملغى مو «منجز»).
		if b, _ := m["partial"].(bool); b {
			loc.Station = "تحتاج إكمال"
		} else if b, _ := m["awaitingCustomer"].(bool); b {
			loc.Station = "بانتظار موافقة الزبون"
		} else if b, _ := m["stuck"].(bool); b {
			loc.Station = "ما وصلت للتنفيذ"
		} else {
			for _, st := range stations {
				if v, _ := m[st.key].(bool); v {
					loc.Station = st.label
					break
				}
			}
		}
		out = append(out, loc)
	}
	return out, nil
}
