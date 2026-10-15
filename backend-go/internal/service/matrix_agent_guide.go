package service

import (
	"fmt"
	"strings"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ ماتركس يفهم شاشات النظام — (ع) 10-11 ═══
// «اخذ سكرين من النظام ادزه لماتركس ويكلي هاي ليش». قبل، أدواته أرقام بس:
// ما يعرف شنو يعني «بانتظار موافقة الزبون» ولا يكدر يفتح حجز بكوده.

const systemGuideText = `دليل نظام شركة الأماني (staffmanage):

【محطات الحجز】 بالترتيب:
- «بانتظار التثبيت — بحاجة لتنسيق» (PENDING): الحجز انسجّل والإداري ما ثبّت الموعد ويا الزبون بعد.
- «تم التثبيت — بحاجة لكادر» (CONFIRMED): الموعد ثابت، ينتظر تكليف فني/فريق.
- «مكلّف — بانتظار التنفيذ»: انكلّف كادر وما خلص.
- «تم الإنجاز» (COMPLETED): خلص الشغل.
- «تحتاج إكمال» (PARTIAL): انعمل جزء والباقي يحتاج زيارة ثانية.
- «مرحّل لإدارة أخرى»: محوّل للمشاريع أو للتقنيين.
- «بانتظار قرار الحذف»: أحد طلب حذفه والمدير ما قرر.
- «ما وصلت للتنفيذ» فيها: «زبون ما رد» (قبل/بعد التثبيت — اتصلوا وما رد، waitingKind=NO_ANSWER)، «حجوزات مؤجّلة» (ينتظر موعد جديد)، «ملغية».
- «بانتظار موافقة الزبون» (waitingKind=CUSTOMER_DECISION): الزبون استفسر ورايح يرجّع خبر (سعر، يفكر). زر «الزبون رد» يرجّعه لمحطته، و«مارد» يسجّل محاولة.
- الحجز يطلع بالبحث بس بمحطته؛ إذا تدوّر بمكان غلط يطلع «لگيناه بس مو بهاي المحطة».

【الحجز الداخلي】 (bookingType=INTERNAL): شغل داخل الشركة لقسم. بعد ما يخلص، الإداري يسوّي «الحساب» (شنو انعمل + السعر المقدّر) من طابور «أعمال داخل الشركة تنتظر ورقك» بالتنسيق، ويسوّي التقرير.

【فواتير الليدر】: الليدر يقدّم الفاتورة → المدقّق يدقّق (مطابق/غير مطابق/خطأ بالسعر/مجانية) → المحاسب يعتمد برقم الفاتورة المحاسبية. المجانية والشغل الداخلي ينعتمدن بلا رقم، والداخلي بلا تدقيق.

【المشاريع】: مراحل (جديد، كشف، عرض سعر، البدء بالتنفيذ، مكتمل، مرفوض). المسؤول + منفّذ الكشف + الموجّه إليه. القيمة المالية = السعر المكتوب، وإذا ماكو فصافي عرض السعر. الدفعات تنسجّل وتنتظر تأكيد المحاسب (إلا إذا سجّلها المحاسب). إحصائيات المشاريع تحسب الفلوس من المستلم فعلاً، والمرفوض ما ينحسب بقيمة العقود.

【التقييم】: ماتركس ٦٠٪ + تقييم البشر ٤٠٪، والاعتمادية رقم منفصل. المدير والمالك ما ينقيّمون. ماكو غرامات فلوس.

【الدوام】: بوابة الحضور، التأخير ينحسب من بداية الشفت.`

func (a *MatrixAgent) registerGuideTools() {
	a.AddTool(agentTool{Name: "system_guide", Desc: "دليل شاشات النظام ومعنى كل محطة وحالة (محطات الحجز، بانتظار موافقة الزبون، زبون ما رد، الحجز الداخلي، الفواتير، المشاريع، التقييم). افتحه أول ما يسألك «هاي شنو/ليش» عن شاشة أو صورة.",
		Run: func(map[string]any, *nameMask) (string, error) { return systemGuideText, nil },
	})
	a.AddTool(agentTool{Name: "lookup_code", Desc: "يفتح حجز (كود مثل B6 أو #8697) أو مشروع (PRJ-0042) بكوده: الحالة، المحطة الي هو بيها، المواعيد، الانتظار، الكادر، والفلوس. استعمله لمن تشوف كود بصورة أو رسالة.",
		Props: map[string]any{"code": map[string]any{"type": "string", "description": "كود الحجز أو المشروع"}},
		Req:   []string{"code"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			code := strings.TrimPrefix(strArg(in, "code"), "#")
			if code == "" || a.db == nil {
				return "اكتب الكود", nil
			}
			first, second := a.lookupBooking, a.lookupProject
			if strings.HasPrefix(strings.ToUpper(code), "P") {
				first, second = second, first
			}
			out, err := first(code)
			if err == nil && strings.HasPrefix(out, "ما لگيت") {
				return second(code)
			}
			return out, err
		},
	})
}

func (a *MatrixAgent) lookupBooking(code string) (string, error) {
	var b struct {
		ID         string     `db:"id"`
		Code       string     `db:"code"`
		Status     string     `db:"status"`
		Type       *string    `db:"bookingType"`
		Service    *string    `db:"service"`
		Scheduled  *time.Time `db:"scheduledAt"`
		Confirmed  *time.Time `db:"confirmedAt"`
		Completed  *time.Time `db:"completedAt"`
		Waiting    *time.Time `db:"waitingSince"`
		WaitKind   *string    `db:"waitingKind"`
		WaitNote   *string    `db:"waitingNote"`
		Attempts   int        `db:"contactAttempts"`
		Reschedule bool       `db:"awaitingReschedule"`
		CancelWhy  *string    `db:"cancelReason"`
		Quoted     *float64   `db:"quotedPrice"`
		Collected  *float64   `db:"amountCollected"`
		Dept       *string    `db:"internalDepartment"`
		Crew       *string    `db:"crew"`
		HasInvoice bool       `db:"hasInvoice"`
	}
	err := a.db.Get(&b, `SELECT b.id, b.code, b.status::text AS status, b."bookingType"::text AS "bookingType", s.name AS service,
		b."scheduledAt", b."confirmedAt", b."completedAt", b."waitingSince", b."waitingKind", b."waitingNote",
		b."contactAttempts", b."awaitingReschedule", b."cancelReason", b."quotedPrice", b."amountCollected", b."internalDepartment",
		(SELECT string_agg(e.name, '، ') FROM "BookingAssignment" ba JOIN "Employee" e ON e.id = ba."employeeId" WHERE ba."bookingId" = b.id) AS crew,
		EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id) AS "hasInvoice"
		FROM "Booking" b LEFT JOIN "Service" s ON s.id = b."serviceId"
		WHERE b.code ILIKE $1 ORDER BY b."createdAt" DESC LIMIT 1`, code)
	if err != nil {
		return "ما لگيت حجز بهذا الكود: " + code, nil
	}
	station := "—"
	if locs, _ := repository.NewBookingRepository(a.db).Locate(b.Code); len(locs) > 0 {
		for _, l := range locs {
			if l.Code == b.Code {
				station = l.Station
			}
		}
	}
	t := func(x *time.Time) string {
		if x == nil {
			return "—"
		}
		return bagh(*x).Format("2006-01-02 15:04")
	}
	s := func(x *string) string {
		if x == nil || *x == "" {
			return "—"
		}
		return *x
	}
	f := func(x *float64) string {
		if x == nil {
			return "—"
		}
		return money(*x)
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "حجز %s — الحالة %s، النوع %s، الخدمة %s\nالمحطة الحالية: %s\n", b.Code, b.Status, s(b.Type), s(b.Service), station)
	fmt.Fprintf(&sb, "الموعد %s، انثبّت %s، خلص %s\n", t(b.Scheduled), t(b.Confirmed), t(b.Completed))
	if b.Waiting != nil {
		fmt.Fprintf(&sb, "بالانتظار من %s (النوع %s، ملاحظة: %s، محاولات اتصال %d)\n", t(b.Waiting), s(b.WaitKind), s(b.WaitNote), b.Attempts)
	}
	if b.Reschedule {
		sb.WriteString("ينتظر موعد جديد (مؤجّل)\n")
	}
	if b.CancelWhy != nil {
		fmt.Fprintf(&sb, "سبب الإلغاء: %s\n", *b.CancelWhy)
	}
	if b.Dept != nil {
		fmt.Fprintf(&sb, "داخل الشركة — القسم: %s\n", *b.Dept)
	}
	fmt.Fprintf(&sb, "الكادر: %s\nالسعر المعروض %s، المستلم %s، عنده فاتورة ليدر: %v", s(b.Crew), f(b.Quoted), f(b.Collected), b.HasInvoice)
	return sb.String(), nil
}

func (a *MatrixAgent) lookupProject(code string) (string, error) {
	var p struct {
		Code    string    `db:"code"`
		Name    string    `db:"name"`
		Stage   string    `db:"stage"`
		Price   *string   `db:"price"`
		Quoted  *float64  `db:"quoted"`
		Resp    *string   `db:"resp"`
		Deleg   *string   `db:"deleg"`
		Paid    float64   `db:"paid"`
		Unver   float64   `db:"unver"`
		Created time.Time `db:"createdAt"`
	}
	err := a.db.Get(&p, `SELECT p.code, p.name, p.stage, p.price,
		(SELECT q."netTotal" FROM "Quotation" q WHERE q."projectId" = p.id ORDER BY q."createdAt" DESC LIMIT 1) AS quoted,
		r.name AS resp, d.name AS deleg,
		COALESCE((SELECT SUM(amount) FROM "ProjectPayment" WHERE "projectId" = p.id AND "cancelledAt" IS NULL), 0) AS paid,
		COALESCE((SELECT SUM(amount) FROM "ProjectPayment" WHERE "projectId" = p.id AND "cancelledAt" IS NULL AND "verifiedAt" IS NULL), 0) AS unver,
		p."createdAt"
		FROM "Project" p LEFT JOIN "Employee" r ON r.id = p."responsibleEmployeeId" LEFT JOIN "Employee" d ON d.id = p."delegatedToEmployeeId"
		WHERE p.code ILIKE $1 LIMIT 1`, code)
	if err != nil {
		return "ما لگيت مشروع بهذا الكود: " + code, nil
	}
	s := func(x *string) string {
		if x == nil || *x == "" {
			return "—"
		}
		return *x
	}
	q := "—"
	if p.Quoted != nil {
		q = money(*p.Quoted)
	}
	return fmt.Sprintf("مشروع %s «%s» — المرحلة %s، انضاف %s\nالسعر المكتوب %s، صافي عرض السعر %s\nالمسؤول %s، موجّه إلى %s\nالمستلم %s (منها تنتظر تأكيد المحاسب %s)",
		p.Code, p.Name, p.Stage, bagh(p.Created).Format("2006-01-02"), s(p.Price), q, s(p.Resp), s(p.Deleg), money(p.Paid), money(p.Unver)), nil
}

// ── 💰 ربحية الخدمات وأرقام «شنو لو؟» — (ع) 10-11 ──
// الأرقام من فواتير الليدر المعتمدة/المقدّمة (بلا الملغية والمسحوبة) +
// مصاريف الحجز + أوقات الشغل. ⚠️ الرواتب وإيجار الشركة مو بالنظام —
// الأداة تگولها بصراحة حتى هايكو ما يسمّي الهامش «صافي ربح».
func (a *MatrixAgent) registerEconomicsTools() {
	a.AddTool(agentTool{Name: "service_economics", Desc: "ربحية كل خدمة بآخر N يوم (افتراضي ٩٠): عدد الشغلات، مجموع الفواتير، أجور التنفيذ، ربح المواد، المصاريف، معدل ساعات الشغلة وحجم الفريق، والدخل لكل ساعة عمل. استعملها لسؤال الربحية ولحسابات «شنو لو؟» (زيادة فني، رفع سعر…).",
		Props: map[string]any{"days": map[string]any{"type": "integer", "description": "عدد الأيام (٣٠–٣٦٥)"}},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			days := 90
			if v, ok := in["days"].(float64); ok && v >= 30 && v <= 365 {
				days = int(v)
			}
			rows := []struct {
				Service string   `db:"service"`
				Jobs    int      `db:"jobs"`
				Net     float64  `db:"net"`
				Exec    float64  `db:"exec"`
				MatProf float64  `db:"matprof"`
				Exp     float64  `db:"exp"`
				Hours   *float64 `db:"hours"`
				Crew    *float64 `db:"crew"`
			}{}
			err := a.db.Select(&rows, `
				WITH inv AS (
					SELECT DISTINCT ON (li."bookingId") li."bookingId", li.id, COALESCE(li."netTotal", 0) AS net, COALESCE(li."executionCost", 0) AS exec
					FROM "LeaderInvoice" li
					WHERE li."bookingId" IS NOT NULL AND li.status IN ('SUBMITTED', 'APPROVED') AND li."revokedAt" IS NULL
					ORDER BY li."bookingId", li."createdAt" DESC
				)
				SELECT COALESCE(s.name, 'بلا خدمة') AS service, count(*) AS jobs,
					COALESCE(SUM(inv.net), 0) AS net, COALESCE(SUM(inv.exec), 0) AS exec,
					COALESCE(SUM((SELECT SUM(mi."profitPerUnit" * mi.quantity) FROM "LeaderInvoiceMaterialItem" mi WHERE mi."leaderInvoiceId" = inv.id)), 0) AS matprof,
					COALESCE(SUM((SELECT SUM(e.amount) FROM "Expense" e WHERE e."bookingId" = b.id AND e.status::text <> 'REJECTED')), 0) AS exp,
					AVG((booking_work_minutes(b.id, b."startedAt", b."completedAt", b."autoArrivedAt", b."autoCompletedAt") / 60)) FILTER (WHERE b."startedAt" IS NOT NULL AND b."completedAt" > b."startedAt"
						AND b."completedAt" - b."startedAt" < interval '16 hours') AS hours,
					AVG((SELECT count(*) FROM "BookingAssignment" ba WHERE ba."bookingId" = b.id)) AS crew
				FROM "Booking" b
				LEFT JOIN "Service" s ON s.id = b."serviceId"
				LEFT JOIN inv ON inv."bookingId" = b.id
				WHERE b.status = 'COMPLETED' AND b."completedAt" >= now() - make_interval(days => $1)
				  AND COALESCE(b."bookingType"::text, '') <> 'INTERNAL'
				GROUP BY 1 ORDER BY 3 DESC LIMIT 25`, days)
			if err != nil {
				return "", err
			}
			var sb strings.Builder
			fmt.Fprintf(&sb, "ربحية الخدمات آخر %d يوم (الحجوزات المنجزة، بلا الداخلية):\n", days)
			for _, r := range rows {
				h, c := "—", "—"
				perHour := "—"
				if r.Hours != nil && *r.Hours > 0 {
					h = fmt.Sprintf("%.1f", *r.Hours)
					if r.Crew != nil && *r.Crew > 0 && r.Jobs > 0 {
						perHour = money((r.Net - r.Exp) / (float64(r.Jobs) * *r.Hours * *r.Crew))
					}
				}
				if r.Crew != nil {
					c = fmt.Sprintf("%.1f", *r.Crew)
				}
				fmt.Fprintf(&sb, "- %s: %d شغلة، فواتير %s، أجور تنفيذ %s، ربح مواد %s، مصاريف %s، معدل %s ساعة × %s شخص، الدخل لكل ساعة عمل شخص %s\n",
					r.Service, r.Jobs, money(r.Net), money(r.Exec), money(r.MatProf), money(r.Exp), h, c, perHour)
			}
			sb.WriteString("⚠️ ملاحظة: الرواتب والإيجار والوقود العام مو مسجّلة بالنظام — هذا «هامش مباشر» (أجور التنفيذ + ربح المواد − مصاريف الحجز)، مو صافي ربح. والحجز البلا فاتورة ينحسب بعدّه بس.")
			return sb.String(), nil
		},
	})
}
