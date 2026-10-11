package service

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/repository"
)

// ═══ رسالة الصبح + إشادة الأسبوع — أفكار ماتركس (قرار (ع) 10-11) ═══
//
// ١. «رسالة الصبح»: ٨ الصبح، ثلاث أسطر للمالك والمدير — شنو الأهم اليوم
//    وليش، بالرقم. الكود يجمع ويرتّب (مجاني)، وهايكو بس يصيغ (رخيص) إذا
//    مفتاحه شغّال؛ وبلاه تطلع بصياغة الكود.
// ٢. «إشادة الأسبوع»: كل خميس، أكثر الموظفين إنجازاً بلا تأخير ينذكرون
//    بإعلان للكل. حساب بحت بلا ذكاء خارجي (الأسماء ما تطلع).

const (
	morningBriefHour = 8
	kudosWeekday     = time.Thursday
	kudosHour        = 13
)

type MatrixMorningService struct {
	db       *sqlx.DB
	aiRepo   *repository.AiRepository
	notif    *repository.NotificationRepository
	announce *repository.AnnouncementRepository
	agent    *MatrixAgent
}

func NewMatrixMorningService(db *sqlx.DB, aiRepo *repository.AiRepository, notif *repository.NotificationRepository,
	announce *repository.AnnouncementRepository, agent *MatrixAgent) *MatrixMorningService {
	return &MatrixMorningService{db: db, aiRepo: aiRepo, notif: notif, announce: announce, agent: agent}
}

// briefItem بند مرشّح: الوزن × العدد يقرر الترتيب.
type briefItem struct {
	weight float64
	count  float64
	line   string
}

func (s *MatrixMorningService) count(q string, args ...any) float64 {
	var n float64
	if err := s.db.Get(&n, q, args...); err != nil {
		log.Printf("morning brief query: %v", err)
	}
	return n
}

// Items يجمع كل البنود المرشّحة (للاختبار وللرسالة).
func (s *MatrixMorningService) Items() []briefItem {
	items := []briefItem{}
	add := func(w, n float64, format string, a ...any) {
		if n > 0 {
			items = append(items, briefItem{weight: w, count: n, line: fmt.Sprintf(format, a...)})
		}
	}
	n := s.count(`SELECT count(*) FROM "Booking" WHERE status = 'PENDING' AND "waitingSince" IS NULL AND "archivedAt" IS NULL
		AND "createdAt" < now() - interval '24 hours'`)
	add(3, n, "%.0f حجز بانتظار التثبيت من أكثر من يوم — الزبون ينتظر اتصال", n)

	n = s.count(`SELECT count(*) FROM "Booking" WHERE status <> 'CANCELLED' AND "waitingKind" = 'CUSTOMER_DECISION'
		AND "waitingSince" < now() - interval '3 days'`)
	add(2, n, "%.0f زبون بانتظار قراره من أكثر من ٣ أيام — يحتاجون اتصال متابعة قبل ما يبردون", n)

	n = s.count(`SELECT count(*) FROM "Booking" b WHERE b.status = 'COMPLETED' AND b."completedAt" >= now() - interval '3 days'
		AND b."completedAt" < now() - interval '12 hours'
		AND NOT EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id)`)
	add(3, n, "%.0f حجز خلص بلا فاتورة ليدر — فلوس ما انحسبت", n)

	n = s.count(`SELECT count(*) FROM "LeaderInvoice" WHERE status = 'SUBMITTED' AND "auditedAt" IS NULL AND "createdAt" < now() - interval '24 hours'`)
	add(2, n, "%.0f فاتورة تنتظر التدقيق من أكثر من يوم", n)

	var unv struct {
		N   float64 `db:"n"`
		Sum float64 `db:"s"`
	}
	if err := s.db.Get(&unv, `SELECT count(*) AS n, COALESCE(SUM(amount), 0) AS s FROM "ProjectPayment"
		WHERE "verifiedAt" IS NULL AND "cancelledAt" IS NULL AND "createdAt" < now() - interval '48 hours'`); err != nil {
		log.Printf("morning brief payments: %v", err)
	}
	add(2.5, unv.N, "%.0f دفعة مشاريع (%s) ما أكدها المحاسب من أكثر من يومين", unv.N, money(unv.Sum))

	n = s.count(`SELECT count(*) FROM "MatrixProposal" WHERE status = 'PENDING'`)
	add(1.5, n, "%.0f اقتراح من ماتركس ينتظر قرارك بصندوق القرارات", n)

	n = s.count(`SELECT count(*) FROM "Booking" WHERE status IN ('CONFIRMED', 'PENDING') AND "scheduledAt" < now() - interval '2 hours'
		AND "scheduledAt" > now() - interval '3 days' AND "completedAt" IS NULL AND "waitingSince" IS NULL`)
	add(3, n, "%.0f حجز فات موعده وما خلص — شوف وين واگف", n)

	// نزول مفاجئ: إنجاز هالأسبوع أقل من نص معدّله بالـ٣ أسابيع الي قبله
	drops := []struct {
		Name string  `db:"name"`
		Now  int     `db:"now"`
		Avg  float64 `db:"avg"`
	}{}
	if err := s.db.Select(&drops, `
		SELECT e.name, count(DISTINCT b.id) FILTER (WHERE b."completedAt" >= now() - interval '7 days') AS now,
			count(DISTINCT b.id) FILTER (WHERE b."completedAt" < now() - interval '7 days') / 3.0 AS avg
		FROM "BookingAssignment" ba JOIN "Booking" b ON b.id = ba."bookingId" JOIN "Employee" e ON e.id = ba."employeeId"
		WHERE b.status = 'COMPLETED' AND b."completedAt" >= now() - interval '28 days'
		  AND e.status = 'ACTIVE' AND e.role NOT IN ('ADMIN', 'OWNER')
		GROUP BY e.name
		HAVING count(DISTINCT b.id) FILTER (WHERE b."completedAt" < now() - interval '7 days') >= 12
		   AND count(DISTINCT b.id) FILTER (WHERE b."completedAt" >= now() - interval '7 days')
		       < count(DISTINCT b.id) FILTER (WHERE b."completedAt" < now() - interval '7 days') / 6.0
		ORDER BY 3 DESC LIMIT 2`); err != nil {
		log.Printf("morning brief drops: %v", err)
	}
	if len(drops) > 0 {
		parts := []string{}
		for _, d := range drops {
			parts = append(parts, fmt.Sprintf("%s (%d شغلة هالأسبوع، معدّله %.0f)", d.Name, d.Now, d.Avg))
		}
		add(2, float64(len(drops)), "نزول مفاجئ بالإنجاز: %s — اسأله إذا أكو شي يعيقه", strings.Join(parts, "، "))
	}

	// مشاريع راح تتأخر: بمرحلتها أكثر من المعتاد تاريخياً (الوسيط) وقربت
	// من حدها (٧٥٪+) بس بعد ما عبرته — تنبيه قبل التأخير مو بعده
	risky := []struct {
		Code  string  `db:"code"`
		Stage string  `db:"stage"`
		Days  float64 `db:"days"`
		Limit float64 `db:"lim"`
	}{}
	if err := s.db.Select(&risky, `
		WITH spans AS (
			SELECT l."toStage" AS stage,
				EXTRACT(EPOCH FROM (lead(l."createdAt") OVER (PARTITION BY l."projectId" ORDER BY l."createdAt") - l."createdAt")) / 86400 AS d
			FROM "ProjectStageLog" l
		), med AS (
			SELECT stage, percentile_cont(0.5) WITHIN GROUP (ORDER BY d) AS m, count(*) AS n
			FROM spans WHERE d IS NOT NULL GROUP BY stage
		), cur AS (
			SELECT p.code, p.stage,
				EXTRACT(EPOCH FROM (now() - COALESCE((SELECT max(l."createdAt") FROM "ProjectStageLog" l WHERE l."projectId" = p.id), p."createdAt"))) / 86400 AS days
			FROM "Project" p
			WHERE p.stage NOT LIKE '%مكتمل%' AND p.stage NOT LIKE '%مرفوض%'
		)
		SELECT c.code, c.stage, c.days, sl.days AS lim
		FROM cur c JOIN "ProjectStageLimit" sl ON sl.stage = c.stage
		JOIN med ON med.stage = c.stage AND med.n >= 3
		WHERE c.days > med.m AND c.days >= sl.days * 0.75 AND c.days < sl.days
		ORDER BY c.days / sl.days DESC LIMIT 3`); err != nil {
		log.Printf("morning brief project risk: %v", err)
	}
	if len(risky) > 0 {
		parts := []string{}
		for _, r := range risky {
			parts = append(parts, fmt.Sprintf("%s (%s: %.0f من %.0f يوم)", r.Code, r.Stage, r.Days, r.Limit))
		}
		add(2.5, float64(len(risky)), "مشاريع راح تتأخر إذا ما تتحرك اليوم: %s", strings.Join(parts, "، "))
	}

	sort.SliceStable(items, func(i, j int) bool { return items[i].weight*items[i].count > items[j].weight*items[j].count })
	return items
}

// RunMorningIfDue ٨ الصبح، مرة باليوم.
func (s *MatrixMorningService) RunMorningIfDue() error {
	now := time.Now().In(debriefLoc)
	if now.Hour() < morningBriefHour || now.Hour() >= 12 {
		return nil
	}
	if !AIFeatureOn("MORNING_BRIEF") {
		return nil
	}
	claimed, err := s.aiRepo.ClaimDailyMarker("MATRIX_MORNING_BRIEF", now.Format("2006-01-02"))
	if err != nil || !claimed {
		return err
	}
	items := s.Items()
	var b strings.Builder
	if len(items) == 0 {
		b.WriteString("☀️ صباح الخير — ماتركس فحص كلشي: ماكو شي معلّق يستاهل وقتك اليوم. يوم موفق.")
	} else {
		b.WriteString("☀️ رسالة الصبح من ماتركس — اليوم انتبه على:\n")
		for i, it := range items {
			if i == 3 {
				break
			}
			fmt.Fprintf(&b, "%d. %s\n", i+1, it.line)
		}
	}
	msg := strings.TrimSpace(b.String())
	if s.agent != nil && s.agent.Enabled("TONE") {
		msg = s.agent.Rewrite(nil, msg)
	}
	return notifyOwnerAndAdmin(s.notif, "AI_MORNING_BRIEF", msg)
}

// KudosRow موظف يستاهل إشادة.
type KudosRow struct {
	ID     string `db:"id"`
	Name   string `db:"name"`
	Done   int    `db:"done"`
	Late   int    `db:"late"`
	Praise int    `db:"praise"`
}

// Kudos أكثر ٣ إنجازاً بآخر ٧ أيام (٣ حجوزات على الأقل، وبلا تأخير يذكر).
func (s *MatrixMorningService) Kudos() []KudosRow {
	rows := []KudosRow{}
	err := s.db.Select(&rows, `
		SELECT e.id, e.name, count(DISTINCT b.id) AS done,
			count(DISTINCT b.id) FILTER (WHERE b."scheduledAt" IS NOT NULL AND b."completedAt" > b."scheduledAt" + interval '4 hours') AS late,
			0 AS praise
		FROM "BookingAssignment" ba
		JOIN "Booking" b ON b.id = ba."bookingId"
		JOIN "Employee" e ON e.id = ba."employeeId"
		WHERE b.status = 'COMPLETED' AND b."completedAt" >= now() - interval '7 days'
		  AND e.status = 'ACTIVE' AND e.role NOT IN ('ADMIN', 'OWNER')
		GROUP BY e.id, e.name
		HAVING count(DISTINCT b.id) >= 3
		ORDER BY count(DISTINCT b.id) FILTER (WHERE b."scheduledAt" IS NULL OR b."completedAt" <= b."scheduledAt" + interval '4 hours') DESC,
		         count(DISTINCT b.id) DESC
		LIMIT 3`)
	if err != nil {
		log.Printf("kudos query: %v", err)
	}
	out := rows[:0]
	for _, r := range rows {
		if r.Late*4 <= r.Done { // التأخير ما يتجاوز الربع
			out = append(out, r)
		}
	}
	return out
}

// RunKudosIfDue كل خميس بعد الظهر: إعلان للكل بأسماء الأبطال.
func (s *MatrixMorningService) RunKudosIfDue() error {
	now := time.Now().In(debriefLoc)
	if now.Weekday() != kudosWeekday || now.Hour() < kudosHour || s.announce == nil {
		return nil
	}
	if !AIFeatureOn("KUDOS") {
		return nil
	}
	claimed, err := s.aiRepo.ClaimDailyMarker("MATRIX_WEEKLY_KUDOS", now.Format("2006-01-02"))
	if err != nil || !claimed {
		return err
	}
	rows := s.Kudos()
	if len(rows) == 0 {
		return nil
	}
	var owner string
	if err := s.db.Get(&owner, `SELECT id FROM "Employee" WHERE role = 'OWNER' ORDER BY "createdAt" LIMIT 1`); err != nil {
		return err
	}
	medals := []string{"🥇", "🥈", "🥉"}
	parts := make([]string, 0, len(rows))
	for i, r := range rows {
		parts = append(parts, fmt.Sprintf("%s %s (%d شغلة)", medals[i], r.Name, r.Done))
	}
	body := "🏆 أبطال الأسبوع من ماتركس: " + strings.Join(parts, " · ") + " — شغل نظيف وبوقته، شكراً إلكم 👏"
	if _, err := s.announce.Create(body, owner, 3); err != nil {
		return err
	}
	return notifyOwnerAndAdmin(s.notif, "AI_KUDOS", "🏆 ماتركس نشر إشادة الأسبوع بالإعلانات: "+strings.Join(parts, "، "))
}
