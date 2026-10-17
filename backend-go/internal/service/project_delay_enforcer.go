package service

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/repository"
)

// ═══ «الموظف المتأخر لازم يقدم تقريره» — (ع) 10-17 ═══
// كل مشروع عبر حد مرحلته وصاحبه ما كتب سبب تأخير لهاي المرحلة: ينطلب منه
// تفسير تلقائي (نفس «التبريرات» الي يحكم عليها المراقب) — مرة وحدة لكل
// مرحلة كل ٢٤ ساعة. يكتب: ليش تأخر، وشلون راح يعوّض. بلا أي غرامة.
func StartProjectDelayEnforcer(db *sqlx.DB, report func() (*ProjectsReport, error), notif *repository.NotificationRepository) {
	go func() {
		time.Sleep(2 * time.Minute)
		for {
			if n := enforceProjectDelays(db, report, notif); n > 0 {
				log.Printf("[project-delay] انطلب تفسير تأخير من %d مشروع", n)
			}
			time.Sleep(2 * time.Hour)
		}
	}()
}

func enforceProjectDelays(db *sqlx.DB, report func() (*ProjectsReport, error), notif *repository.NotificationRepository) int {
	rep, err := report()
	if err != nil || rep == nil {
		return 0
	}
	n := 0
	for _, p := range rep.Projects {
		if !p.OverLimit || p.OwnerID == "" || p.Status == ChainNA || projectFinal(p.Stage) {
			continue
		}
		var skip bool
		// كتب سبب لهاي المرحلة، أو انطلب منه تفسير خلال ٢٤ ساعة؟
		_ = db.Get(&skip, `SELECT EXISTS (SELECT 1 FROM "ProjectDelayReason" WHERE "projectId" = $1 AND stage = $2)
			OR EXISTS (SELECT 1 FROM "MonitorJustification" WHERE "employeeId" = $3 AND question LIKE $4 AND "createdAt" > now() - interval '24 hours')`,
			p.ID, p.Stage, p.OwnerID, "%"+p.Code+"%"+p.Stage+"%")
		if skip {
			continue
		}
		q := fmt.Sprintf("مشروع %s (%s) متأخر بمرحلة «%s» — %d يوم والحد %d. اكتب ليش تأخر، وشلون راح تعوّض التأخير.",
			p.Code, p.Name, p.Stage, p.DaysInStage, p.StageLimit)
		if _, err := db.Exec(`INSERT INTO "MonitorJustification" (id, "employeeId", question, "createdAt") VALUES ($1, $2, $3, now())`,
			uuid.NewString(), p.OwnerID, q); err != nil {
			continue
		}
		if notif != nil {
			_ = notif.Create(p.OwnerID, "justification_request", "⏰ "+q)
		}
		n++
	}
	return n
}
