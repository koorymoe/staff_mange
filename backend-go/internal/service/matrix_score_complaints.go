package service

import (
	"fmt"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ شكاوى الزبائن — مهندس الجودة (دازاي 11-21) ═══
// «تجي شكوى هاي لازم يتصل عليها بسرعة ويشوف شنو المشكلة ويعالجها ويحلها بأسرع وقت ممكن
// سواء كانت تحتاج حجز جديد أو تعويض». بندين:
//   COMPLAINT_CONTACT — أول اتصال: خلال ساعتين = كاملة، خلال يوم = نص، أكثر = صفر
//   COMPLAINT_RESOLVE — الحل: خلال ٣ أيام = كاملة، خلال ٧ = نص، أكثر (أو بعدها مفتوحة) = صفر
// المسؤول: الي اتصل فعلاً، وإلا المكلّف بيها. الشكوى المفتوحة ما تنحسب لحد ما يعدّي حدها.

const (
	ComplaintContactFull = 2 * time.Hour
	ComplaintContactHalf = 24 * time.Hour
	ComplaintResolveFull = 3 * 24 * time.Hour
	ComplaintResolveHalf = 7 * 24 * time.Hour
)

// ComplaintPoints نقاط مدة (وok=false إذا بعده ما خلص ولا عدّى حد النص).
func ComplaintPoints(took *time.Duration, open time.Duration, full, half time.Duration) (int, bool) {
	if took == nil {
		if open > half {
			return 0, true
		}
		return 0, false
	}
	switch {
	case *took <= full:
		return 2, true
	case *took <= half:
		return 1, true
	}
	return 0, true
}

func (s *MatrixScoreService) scoreComplaints(now time.Time, add func(repository.MatrixScoreRow)) {
	var rows []struct {
		ID          string   `db:"id"`
		Contacter   *string  `db:"contacter"`
		Assigned    *string  `db:"assigned"`
		Code        *string  `db:"code"`
		OpenSecs    float64  `db:"openSecs"`
		ContactSecs *float64 `db:"contactSecs"`
		ResolveSecs *float64 `db:"resolveSecs"`
	}
	err := s.repo.DB().Select(&rows, `
		SELECT c.id, c."contactedById" AS contacter, c."assignedToEmployeeId" AS assigned, b.code,
		       EXTRACT(EPOCH FROM (now() AT TIME ZONE 'UTC') - c."createdAt") AS "openSecs",
		       EXTRACT(EPOCH FROM (c."contactedAt" AT TIME ZONE 'UTC') - c."createdAt") AS "contactSecs",
		       EXTRACT(EPOCH FROM c."resolvedAt" - c."createdAt") AS "resolveSecs"
		FROM "Complaint" c LEFT JOIN "Booking" b ON b.id = c."bookingId"
		WHERE c."createdAt" > (now() AT TIME ZONE 'UTC') - make_interval(days => $1) AND c."settledLegacyAt" IS NULL`, scoreLookbackDays)
	if err != nil {
		return
	}
	dur := func(v *float64) *time.Duration {
		if v == nil {
			return nil
		}
		d := time.Duration(*v * float64(time.Second))
		return &d
	}
	for _, c := range rows {
		label := "شكوى"
		if c.Code != nil {
			label += " حجز " + *c.Code
		}
		open := time.Duration(c.OpenSecs * float64(time.Second))
		who := c.Contacter
		if who == nil {
			who = c.Assigned
		}
		if who == nil {
			continue
		}
		if pts, ok := ComplaintPoints(dur(c.ContactSecs), open, ComplaintContactFull, ComplaintContactHalf); ok {
			why := map[int]string{2: "اتصل خلال ساعتين.", 1: "اتصل بس بعد ساعتين (خلال يوم).", 0: "ما اتصل خلال يوم."}[pts]
			add(repository.MatrixScoreRow{EmployeeID: *who, Source: "COMPLAINT", SourceID: c.ID + "@contact", SourceLabel: &label,
				Rule: "COMPLAINT_CONTACT", Points: pts, MaxPoints: 2, At: now, Reason: fmt.Sprintf("%s · الاتصال: %s", label, why)})
		}
		if pts, ok := ComplaintPoints(dur(c.ResolveSecs), open, ComplaintResolveFull, ComplaintResolveHalf); ok {
			why := map[int]string{2: "انحلّت خلال ٣ أيام.", 1: "انحلّت بس بعد ٣ أيام (خلال أسبوع).", 0: "ما انحلّت خلال أسبوع."}[pts]
			add(repository.MatrixScoreRow{EmployeeID: *who, Source: "COMPLAINT", SourceID: c.ID + "@resolve", SourceLabel: &label,
				Rule: "COMPLAINT_RESOLVE", Points: pts, MaxPoints: 2, At: now, Reason: fmt.Sprintf("%s · الحل: %s", label, why)})
		}
	}
}
