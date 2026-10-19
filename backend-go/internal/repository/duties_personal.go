package repository

// ═══ الواجبات الشخصية — طلب (ع) 10-19 ═══
// «حالياً كل موظف يگله واجبك اليوم تسجّل حضور… مثلاً يجي على المشرف مال
// المشروع: عندك مشروع مرحّل إلك أنتظر منك تبدي العمل اليوم بيه وتنطيني
// تقرير، وعندك حجز مرحّل للتقنيين لازم تتواصل ويا الزبون. لازم تطلع
// الواجبات بالتفصيل». فوق واجبات المنصب: سطر لكل شغلة بإسمه.

type personalItem struct {
	ID     string `db:"id"`
	Title  string `db:"title"`
	Action string `db:"action"`
	Done   bool   `db:"done"`
	Counts bool   `db:"counts"` // false = معلومة (ما تنحسب بالتقييم)
}

var personalQueries = []struct {
	kind, link string
	sql        string
}{
	// المشاريع الموجّهة إله — الإجراء حسب المرحلة.
	{"project", "/my-projects", `
		SELECT p.id, '📁 مشروع «' || p.name || '» ' || COALESCE(p.code, '') AS title,
		  CASE
		    WHEN p.stage LIKE '5.%' OR p.stage LIKE '6.%' OR p.stage LIKE '7.%' THEN 'ابدي العمل اليوم وارفع تقرير يومي بالي انشغل'
		    WHEN p."customerContactedAt" IS NULL THEN 'تواصل ويا الزبون وسجّل شنو حچيتوا'
		    ELSE 'تابعه ورحّله للمرحلة الجاية (' || p.stage || ')' END AS action,
		  CASE
		    WHEN p.stage LIKE '5.%' OR p.stage LIKE '6.%' OR p.stage LIKE '7.%' THEN EXISTS (
		      SELECT 1 FROM "BookingProgressReport" r WHERE r."bookingId" = p."bookingId"
		        AND r."reportedById" = $1 AND baghdad_date(r."createdAt") = baghdad_today())
		    WHEN p."customerContactedAt" IS NULL THEN false
		    ELSE false END AS done,
		  (p.stage LIKE '5.%' OR p.stage LIKE '6.%' OR p.stage LIKE '7.%' OR p."customerContactedAt" IS NULL) AS counts
		FROM "Project" p
		WHERE p."delegatedToEmployeeId" = $1 AND p.stage NOT LIKE '%مكتمل%' AND p.stage NOT LIKE '%مرفوض%'
		ORDER BY p."updatedAt" DESC LIMIT 10`},
	// الحجز المرحّل إله (حجوزات التقنيين).
	{"handover", "/my-handovers", `
		SELECT b.id, '🔧 حجز ' || b.code || ' — ' || COALESCE(c.name, '') AS title,
		  'مرحّل إلك: تواصل ويا الزبون وشوف المشكلة' AS action,
		  (b."startedAt" IS NOT NULL OR EXISTS (SELECT 1 FROM "BookingStageProof" sp WHERE sp."bookingId" = b.id
		     AND sp."employeeId" = $1 AND baghdad_date(sp."pressedAt") = baghdad_today())) AS done,
		  true AS counts
		FROM "Booking" b LEFT JOIN "Customer" c ON c.id = b."customerId"
		WHERE b."handoverToId" = $1 AND b.status NOT IN ('COMPLETED', 'CANCELLED') AND b."archivedAt" IS NULL
		ORDER BY b."handoverAt" DESC NULLS LAST LIMIT 10`},
	// المهام الإضافية الموجّهة إله.
	{"task", "/my-extra-tasks", `
		SELECT t.id, '📌 مهمة: ' || t.title AS title,
		  CASE WHEN t."dueAt" IS NOT NULL THEN 'موعدها ' || to_char(t."dueAt" AT TIME ZONE 'Asia/Baghdad', 'YYYY-MM-DD HH24:MI')
		       ELSE 'نفّذها وسجّل إنجازها' END AS action,
		  false AS done, true AS counts
		FROM "ExtraTask" t
		WHERE t."assignedToId" = $1 AND t.status IN ('NEW', 'IN_PROGRESS')
		ORDER BY t."dueAt" ASC NULLS LAST LIMIT 10`},
	// الشكاوى المسندة إله وموعد الاتصال بيها حل.
	{"complaint", "/complaints", `
		SELECT k.id, '🎧 شكوى ' || COALESCE(cu.name, '') AS title,
		  CASE WHEN k."noAnswerCount" > 0 THEN 'اتصل بالزبون مرة ثانية (ما رد ' || k."noAnswerCount" || ' من ٣)'
		       ELSE 'اتصل بالزبون وسجّل التواصل والتقييم' END AS action,
		  false AS done, true AS counts
		FROM "Complaint" k LEFT JOIN "Customer" cu ON cu.id = k."customerId"
		WHERE k."assignedToEmployeeId" = $1 AND k.status::text IN ('NEW', 'IN_PROGRESS') AND k."contactedAt" IS NULL
		  AND (k."nextCallAt" IS NULL OR k."nextCallAt" <= now())
		ORDER BY k."createdAt" LIMIT 10`},
}

// personal سطر لكل شغلة موجّهة للموظف — يندمج ويا واجبات منصبه.
func (r *DutiesRepository) personal(employeeID string) []Duty {
	out := []Duty{}
	for _, q := range personalQueries {
		items := []personalItem{}
		if err := r.db.Select(&items, q.sql, employeeID); err != nil {
			continue
		}
		for _, it := range items {
			d := Duty{Key: q.kind + ":" + it.ID, Label: it.Title, Hint: it.Action, Link: q.link}
			if !it.Counts {
				d.Status = "INFO"
			} else {
				p := 1
				d.Status = "MISSED"
				if it.Done {
					p, d.Done, d.Status = 0, 1, "DONE"
				}
				d.Pending = &p
			}
			out = append(out, d)
		}
	}
	return out
}
