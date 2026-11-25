package service

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// ═══ 🧩 مؤشر هشاشة النظام (ل) 11-25 دازاي ═══
// «مؤشر هشاشة النظام في حال غياب موظف — واكتشاف إذا تم الاعتماد على موظف واحد لإنجاز نفس العمل،
//  مثل موظف واحد بس يشتغل على الأقفال».
// من الحجوزات المكتملة بآخر N يوم: لكل خدمة كم موظف اشتغلها وحصة أكثر واحد. وكمان من المهارات:
// كم موظف (فعّال) يگدر يسويها. الخطر: واحد بس (حرج) · واحد يسوي ٨٠٪+ (عالي) · اثنين بس (متوسط).
// ولكل موظف: الخدمات الي إذا غاب ماكو غيره يسويها.

type FragilityService struct {
	ServiceID   string  `db:"serviceId" json:"serviceId"`
	ServiceName string  `db:"serviceName" json:"serviceName"`
	Jobs        int     `db:"jobs" json:"jobs"`
	Workers     int     `db:"workers" json:"workers"`
	TopName     *string `db:"topName" json:"topName"`
	TopJobs     int     `db:"topJobs" json:"topJobs"`
	TopShare    float64 `db:"topShare" json:"topShare"`
	Skilled     int     `db:"skilled" json:"skilled"` // يگدرون حسب المهارات
	Risk        string  `db:"-" json:"risk"`          // CRITICAL / HIGH / MEDIUM / OK
}

type FragilityPerson struct {
	EmployeeID string   `json:"employeeId"`
	Name       string   `json:"name"`
	SoleFor    []string `json:"soleFor"` // خدمات هو الوحيد الي سواها
	MainFor    []string `json:"mainFor"` // خدمات هو سوى ٨٠٪+ منها
	OnLeave    bool     `json:"onLeave"` // بإجازة هسه
}

type FragilityReport struct {
	Days     int                `json:"days"`
	Services []FragilityService `json:"services"`
	People   []FragilityPerson  `json:"people"`
	Score    int                `json:"score"` // ٠ (متين) — ١٠٠ (هش)
}

func fragilityRisk(s FragilityService) string {
	switch {
	case s.Workers <= 1 || s.Skilled == 1:
		return "CRITICAL"
	case s.TopShare >= 80:
		return "HIGH"
	case s.Workers == 2:
		return "MEDIUM"
	}
	return "OK"
}

func Fragility(db *sqlx.DB, days int, now time.Time) (*FragilityReport, error) {
	if days <= 0 || days > 365 {
		days = 90
	}
	rep := &FragilityReport{Days: days, Services: []FragilityService{}, People: []FragilityPerson{}}
	if err := db.Select(&rep.Services, `
		WITH done AS (
			SELECT DISTINCT bs."serviceId", a."employeeId", b.id AS bid
			FROM "Booking" b
			JOIN "BookingService" bs ON bs."bookingId" = b.id
			JOIN "BookingAssignment" a ON a."bookingId" = b.id
			JOIN "Employee" e ON e.id = a."employeeId"
			WHERE b.status = 'COMPLETED' AND b."completedAt" >= now() - make_interval(days => $1) AND e.status = 'ACTIVE'
		), per AS (SELECT "serviceId", "employeeId", count(DISTINCT bid) AS n FROM done GROUP BY 1, 2),
		svc AS (SELECT "serviceId", count(DISTINCT bid)::int AS jobs, count(DISTINCT "employeeId")::int AS workers FROM done GROUP BY 1),
		top AS (SELECT DISTINCT ON ("serviceId") "serviceId", "employeeId", n FROM per ORDER BY "serviceId", n DESC)
		SELECT s.id AS "serviceId", s.name AS "serviceName", svc.jobs, svc.workers, e.name AS "topName", top.n::int AS "topJobs",
			round(top.n * 100.0 / NULLIF(svc.jobs, 0))::float8 AS "topShare",
			(SELECT count(DISTINCT es."employeeId") FROM "Skill" k JOIN "EmployeeSkill" es ON es."skillId" = k.id AND es."canPerform"
				JOIN "Employee" x ON x.id = es."employeeId" AND x.status = 'ACTIVE' WHERE k."serviceId" = s.id)::int AS skilled
		FROM svc JOIN "Service" s ON s.id = svc."serviceId" JOIN top ON top."serviceId" = svc."serviceId"
		LEFT JOIN "Employee" e ON e.id = top."employeeId"
		ORDER BY svc.workers, top.n * 1.0 / NULLIF(svc.jobs, 0) DESC, svc.jobs DESC`, days); err != nil {
		return nil, err
	}
	// المهارات: إذا ماكو أحد مسجّل مهارة للخدمة، نعتمد على الحجوزات بس
	people := map[string]*FragilityPerson{}
	type pr struct {
		ServiceID  string  `db:"serviceId"`
		EmployeeID string  `db:"employeeId"`
		Name       string  `db:"name"`
		Share      float64 `db:"share"`
	}
	rows := []pr{}
	_ = db.Select(&rows, `
		WITH done AS (
			SELECT DISTINCT bs."serviceId", a."employeeId", b.id AS bid FROM "Booking" b
			JOIN "BookingService" bs ON bs."bookingId" = b.id JOIN "BookingAssignment" a ON a."bookingId" = b.id
			JOIN "Employee" e ON e.id = a."employeeId"
			WHERE b.status = 'COMPLETED' AND b."completedAt" >= now() - make_interval(days => $1) AND e.status = 'ACTIVE'
		), svc AS (SELECT "serviceId", count(DISTINCT bid) AS jobs FROM done GROUP BY 1)
		SELECT d."serviceId", d."employeeId", e.name, (count(DISTINCT d.bid) * 100.0 / svc.jobs)::float8 AS share
		FROM done d JOIN svc ON svc."serviceId" = d."serviceId" JOIN "Employee" e ON e.id = d."employeeId"
		GROUP BY 1, 2, 3, svc.jobs`, days)
	names := map[string]string{}
	var weighted, total float64
	for i := range rep.Services {
		s := &rep.Services[i]
		if s.Skilled == 0 {
			s.Skilled = -1 // ماكو مهارات مسجّلة
		}
		s.Risk = fragilityRisk(*s)
		if s.Skilled == -1 {
			s.Skilled = 0
		}
		names[s.ServiceID] = s.ServiceName
		w := float64(s.Jobs)
		total += w
		switch s.Risk {
		case "CRITICAL":
			weighted += w
		case "HIGH":
			weighted += 0.7 * w
		case "MEDIUM":
			weighted += 0.35 * w
		}
	}
	if total > 0 {
		rep.Score = int(weighted/total*100 + 0.5)
	}
	for _, r := range rows {
		var svc *FragilityService
		for i := range rep.Services {
			if rep.Services[i].ServiceID == r.ServiceID {
				svc = &rep.Services[i]
			}
		}
		if svc == nil {
			continue
		}
		p := people[r.EmployeeID]
		if p == nil {
			p = &FragilityPerson{EmployeeID: r.EmployeeID, Name: r.Name, SoleFor: []string{}, MainFor: []string{}}
			people[r.EmployeeID] = p
		}
		if svc.Workers == 1 {
			p.SoleFor = append(p.SoleFor, svc.ServiceName)
		} else if r.Share >= 80 {
			p.MainFor = append(p.MainFor, svc.ServiceName)
		}
	}
	onLeave := map[string]bool{}
	ids := []string{}
	_ = db.Select(&ids, `SELECT "employeeId" FROM "LeaveRequest" WHERE status = 'APPROVED' AND "startDate" <= baghdad_today() AND "endDate" >= baghdad_today()`)
	for _, id := range ids {
		onLeave[id] = true
	}
	for _, p := range people {
		if len(p.SoleFor)+len(p.MainFor) == 0 {
			continue
		}
		p.OnLeave = onLeave[p.EmployeeID]
		rep.People = append(rep.People, *p)
	}
	// الأخطر أول: الي هو الوحيد بخدمات أكثر
	for i := 1; i < len(rep.People); i++ {
		for j := i; j > 0 && len(rep.People[j].SoleFor)*10+len(rep.People[j].MainFor) > len(rep.People[j-1].SoleFor)*10+len(rep.People[j-1].MainFor); j-- {
			rep.People[j], rep.People[j-1] = rep.People[j-1], rep.People[j]
		}
	}
	_ = now
	return rep, nil
}
