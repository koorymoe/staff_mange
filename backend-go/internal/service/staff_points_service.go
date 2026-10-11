package service

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/repository"
)

// ═══ 🏅 رصيد نقاط الشهر — قرار (ع) 10-11 ═══
//
// «ماريد حدود، اريد تصاعدي بدون لمت»: الرصيد يتراكم بالشهر بلا سقف، وكل
// دور يترتب ويا نفس دوره. الرصيد = الكي بي اي + الحضور + شغل الدور.
// التقييم القديم (ماتركس ٦٠٪ + البشر ٤٠٪) يبقى مثل ما هو — هذا جنبه.

const (
	pointsKpiItem     = 5.0 // كل بند من الثمانية ما انخصم منه
	pointsOnTime      = 1.0 // يوم حضور بوقته
	pointsLate        = 0.5 // يوم حضور متأخر
	pointsBooking     = 2.0
	pointsFree        = 2.0
	pointsInternal    = 2.0
	pointsNoComplaint = 1.0
	pointsLeaderPaper = 1.0
	pointsLeaderRate  = 1.0
	pointsMonitor     = 1.0
	pointsQuality     = 1.0
	pointsOfficeStep  = 1.0
	pointsExtraTask   = 3.0
)

// PointsLine بند من الرصيد: شنو، كم مرة، كم نقطة.
type PointsLine struct {
	Label  string  `json:"label"`
	Count  float64 `json:"count"`
	Each   float64 `json:"each"`
	Points float64 `json:"points"`
}

type StaffPoints struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Role       string       `json:"role"`
	Group      string       `json:"group"`
	GroupLabel string       `json:"groupLabel"`
	Kpi        float64      `json:"kpi"`
	KpiLost    []string     `json:"kpiLost"`
	Attendance float64      `json:"attendance"`
	Work       float64      `json:"work"`
	Total      float64      `json:"total"`
	Lines      []PointsLine `json:"lines"`
}

type StaffPointsService struct {
	db   *sqlx.DB
	repo *repository.MatrixScoreRepository
}

func NewStaffPointsService(db *sqlx.DB, repo *repository.MatrixScoreRepository) *StaffPointsService {
	return &StaffPointsService{db: db, repo: repo}
}

// counts يرجّع employeeId ← عدد من استعلام بعمودين (id, n).
func (s *StaffPointsService) counts(label, q string, month string) map[string]float64 {
	rows := []struct {
		ID string  `db:"id"`
		N  float64 `db:"n"`
	}{}
	if err := s.db.Select(&rows, q, month); err != nil {
		log.Printf("staff points %s: %v", label, err)
	}
	out := map[string]float64{}
	for _, r := range rows {
		out[r.ID] = r.N
	}
	return out
}

// شرط الشهر ببغداد لعمود — $1 = 'YYYY-MM'
func inMonth(col string) string {
	return fmt.Sprintf(`to_char(baghdad_date(%s), 'YYYY-MM') = $1`, col)
}

func (s *StaffPointsService) Month(month string) ([]StaffPoints, error) {
	people, err := s.repo.Scorables()
	if err != nil {
		return nil, err
	}

	// ── الكي بي اي: البنود المنخصمة هالشهر لكل موظف ──
	criteria := []string{}
	_ = s.db.Select(&criteria, `SELECT label FROM "KpiCriterion" ORDER BY "createdAt"`)
	deductions := []struct {
		ID     string `db:"id"`
		Reason string `db:"reason"`
	}{}
	if err := s.db.Select(&deductions, `SELECT "employeeId" AS id, COALESCE(reason, '') AS reason FROM "KpiEvaluation"
		WHERE NOT cancelled AND points < 0 AND `+inMonth(`"createdAt"`), month); err != nil {
		log.Printf("staff points kpi: %v", err)
	}
	lost := map[string]map[string]bool{}
	for _, d := range deductions {
		for _, c := range criteria {
			if strings.HasPrefix(strings.TrimSpace(d.Reason), c) {
				if lost[d.ID] == nil {
					lost[d.ID] = map[string]bool{}
				}
				lost[d.ID][c] = true
			}
		}
	}

	// ── الحضور: أول بصمة باليوم مقابل بداية الشفت + ١٠ دقايق سماح ──
	att := []struct {
		ID     string  `db:"id"`
		OnTime float64 `db:"ontime"`
		Late   float64 `db:"late"`
	}{}
	if err := s.db.Select(&att, `
		WITH d AS (
			SELECT a."employeeId" AS id, baghdad_date(a."checkIn") AS day, min(a."checkIn") AS first
			FROM "Attendance" a WHERE a."checkIn" IS NOT NULL AND `+inMonth(`a."checkIn"`)+`
			GROUP BY 1, 2
		)
		SELECT d.id,
			count(*) FILTER (WHERE (d.first AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Baghdad')::time
				<= (COALESCE(NULLIF(e."shiftStart", ''), '08:00')::time + interval '10 minutes')) AS ontime,
			count(*) FILTER (WHERE (d.first AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Baghdad')::time
				> (COALESCE(NULLIF(e."shiftStart", ''), '08:00')::time + interval '10 minutes')) AS late
		FROM d JOIN "Employee" e ON e.id = d.id GROUP BY d.id`, month); err != nil {
		log.Printf("staff points attendance: %v", err)
	}
	attMap := map[string][2]float64{}
	for _, a := range att {
		attMap[a.ID] = [2]float64{a.OnTime, a.Late}
	}

	// ── شغل الحجوزات (الفني والليدر والديكور) ──
	bookingBase := `FROM "BookingAssignment" ba JOIN "Booking" b ON b.id = ba."bookingId"
		WHERE b.status = 'COMPLETED' AND b."archivedAt" IS NULL AND ` + inMonth(`b."completedAt"`)
	booked := s.counts("bookings", `SELECT ba."employeeId" AS id, count(DISTINCT b.id) AS n `+bookingBase+`
		AND COALESCE(b."bookingType"::text, '') <> 'INTERNAL' GROUP BY 1`, month)
	internal := s.counts("internal", `SELECT ba."employeeId" AS id, count(DISTINCT b.id) AS n `+bookingBase+`
		AND b."bookingType"::text = 'INTERNAL' GROUP BY 1`, month)
	free := s.counts("free", `SELECT ba."employeeId" AS id, count(DISTINCT b.id) AS n `+bookingBase+`
		AND EXISTS (SELECT 1 FROM "LeaderInvoice" li WHERE li."bookingId" = b.id AND li."revokedAt" IS NULL
			AND (li."isFree" OR COALESCE(li."auditVerdict", '') = 'FREE')) GROUP BY 1`, month)
	clean := s.counts("noComplaint", `SELECT ba."employeeId" AS id, count(DISTINCT b.id) AS n `+bookingBase+`
		AND NOT EXISTS (SELECT 1 FROM "Complaint" c WHERE c."bookingId" = b.id) GROUP BY 1`, month)

	// ── خطوات ماتركس الكاملة (بوقتها) حسب نوعها ──
	steps := func(label, rules string) map[string]float64 {
		return s.counts(label, `SELECT "employeeId" AS id, count(*) AS n FROM "MatrixScore"
			WHERE "cancelledAt" IS NULL AND points >= "maxPoints" AND "maxPoints" > 0 AND rule IN (`+rules+`)
			AND `+inMonth(`"at"`)+` GROUP BY 1`, month)
	}
	paper := steps("paper", `'PAPER'`)
	rating := steps("rating", `'RATING'`)
	monitor := steps("monitor", `'MONITOR'`)
	quality := steps("quality", `'QUALITY'`)
	office := steps("office", `'CREATE','CONTACT','CONFIRM','HANDOFF','CREW','ACCOUNT','PROC_RESPONSE','TOOL_RESPONSE','VEHICLE_DAILY','SHORTAGE_RESOLVE','PROJECT_WEEK'`)

	// ── المهام الإضافية بوقتها (كل الأدوار) ──
	tasks := s.counts("tasks", `SELECT "assignedToId" AS id, count(*) AS n FROM "ExtraTask"
		WHERE "doneAt" IS NOT NULL AND "cancelledAt" IS NULL AND ("dueAt" IS NULL OR "doneAt" <= "dueAt")
		AND `+inMonth(`"doneAt"`)+` GROUP BY 1`, month)

	out := make([]StaffPoints, 0, len(people))
	for _, p := range people {
		g := ScoreGroup(p)
		sp := StaffPoints{ID: p.ID, Name: p.Name, Role: p.Role, Group: g, GroupLabel: ScoreGroupLabels[g], KpiLost: []string{}, Lines: []PointsLine{}}
		add := func(label string, n, each float64) {
			if n <= 0 {
				return
			}
			sp.Lines = append(sp.Lines, PointsLine{Label: label, Count: n, Each: each, Points: n * each})
			sp.Work += n * each
		}
		// الكي بي اي
		for _, c := range criteria {
			if lost[p.ID][c] {
				sp.KpiLost = append(sp.KpiLost, c)
			} else {
				sp.Kpi += pointsKpiItem
			}
		}
		// الحضور
		a := attMap[p.ID]
		sp.Attendance = a[0]*pointsOnTime + a[1]*pointsLate
		// شغل الدور
		switch g {
		case GroupTechs, GroupLeaders, GroupDecor:
			add("حجز منجز", booked[p.ID], pointsBooking)
			add("صيانة مجانية", free[p.ID], pointsFree)
			add("شغل داخل الشركة", internal[p.ID], pointsInternal)
			add("حجز بلا شكوى", clean[p.ID], pointsNoComplaint)
			if p.IsLeader {
				add("فاتورة وتقرير بوقتهم", paper[p.ID], pointsLeaderPaper)
				add("تقييم الكادر بوقته", rating[p.ID], pointsLeaderRate)
			}
		case GroupMonitors:
			add("حجز دقّقه بوقته", monitor[p.ID], pointsMonitor)
		case GroupQuality:
			add("اتصال جودة بوقته", quality[p.ID], pointsQuality)
		case GroupTechnical:
			add("حجز منجز", booked[p.ID]+internal[p.ID], pointsBooking)
			add("شغلة بوقتها", office[p.ID], pointsOfficeStep)
		case GroupFinance, GroupCoordinator, GroupProcurement, GroupSales:
			add("شغلة بوقتها", office[p.ID], pointsOfficeStep)
		}
		add("مهمة إضافية بوقتها", tasks[p.ID], pointsExtraTask)
		sp.Total = sp.Kpi + sp.Attendance + sp.Work
		out = append(out, sp)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Total > out[j].Total })
	return out, nil
}
