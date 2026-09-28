package service

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ ماتركس — التنبؤ بالتأخير (قبل ما يصير) ═══
//
// المدة المتوقعة = **وسيط** المدة الفعلية (startedAt → completedAt)
// لحجوزات مكتملة من نفس الخدمة بآخر ٣٦٥ يوم — لنفس الليدر لو عنده ٥
// عيّنات فأكثر، وإلا لكل الخدمة لو بيها ٥ فأكثر، وإلا **ولا كلمة**:
// رقم غلط أسوأ من ماكو رقم.
//
// الوقت المتاح = من الموعد لحد أقرب وحدة من: نهاية الدوام (AiWorkWindow)
// بنفس يوم بغداد، أو موعد الحجز الجاي لنفس الليدر بنفس اليوم.
//
// خطر = المتوقع أكبر من المتاح بـ٣٠ دقيقة فأكثر. تنبيه بس — بلا أي
// غرامة أو نقاط.

const (
	delayMinSamples     = 5
	delayMarginMinutes  = 30.0
	delayLookaheadDays  = 6
	delayMorningHour    = 7 // ٧ الصبح بغداد
	delayDefaultEndHour = 24
)

// DelayRisk حجز متوقع يطوّل أكثر من وقته.
type DelayRisk struct {
	BookingID        string    `json:"bookingId"`
	BookingCode      string    `json:"bookingCode"`
	ScheduledAt      time.Time `json:"scheduledAt"`
	ExpectedMinutes  int       `json:"expectedMinutes"`
	AvailableMinutes int       `json:"availableMinutes"`
	Samples          int       `json:"samples"`
	Basis            string    `json:"basis"`     // LEADER | SERVICE
	LimitedBy        string    `json:"limitedBy"` // SHIFT_END | NEXT_BOOKING
	NextBookingCode  string    `json:"nextBookingCode,omitempty"`
}

type DelayPredictionService struct {
	aiRepo    *repository.AiRepository
	notifRepo *repository.NotificationRepository
}

func NewDelayPredictionService(aiRepo *repository.AiRepository, notifRepo *repository.NotificationRepository) *DelayPredictionService {
	return &DelayPredictionService{aiRepo: aiRepo, notifRepo: notifRepo}
}

// pickExpected يختار التقدير: الليدر أول (≥٥)، بعدين الخدمة (≥٥)، وإلا لا شي.
func pickExpected(stats map[string]repository.DurationStat, serviceID, leaderID string) (float64, int, string, bool) {
	if leaderID != "" {
		if st, ok := stats[serviceID+"|"+leaderID]; ok && st.Samples >= delayMinSamples {
			return st.MedianMinutes, st.Samples, "LEADER", true
		}
	}
	if st, ok := stats[serviceID+"|"]; ok && st.Samples >= delayMinSamples {
		return st.MedianMinutes, st.Samples, "SERVICE", true
	}
	return 0, 0, "", false
}

// availableWindow الدقائق من الموعد لحد نهاية الدوام أو الحجز الجاي.
func availableWindow(start time.Time, endHour int, next *time.Time) (float64, string) {
	local := start.In(debriefLoc)
	shiftEnd := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, debriefLoc).Add(time.Duration(endHour) * time.Hour)
	end, by := shiftEnd, "SHIFT_END"
	if next != nil && next.Before(end) {
		end, by = *next, "NEXT_BOOKING"
	}
	return end.Sub(start).Minutes(), by
}

// isDelayRisk القاعدة نفسها — مفصولة للاختبار.
func isDelayRisk(expected, available float64) bool {
	return expected-available >= delayMarginMinutes
}

// Risks يرجّع الحجوزات الخطرة بين يومين (بصيغة 2006-01-02، يوم بغداد).
func (s *DelayPredictionService) Risks(from, to string) ([]DelayRisk, error) {
	cands, err := s.aiRepo.UpcomingDelayCandidates(from, to)
	if err != nil {
		return nil, err
	}
	out := []DelayRisk{}
	if len(cands) == 0 {
		return out, nil
	}
	sidSet := map[string]bool{}
	for _, c := range cands {
		sidSet[c.ServiceID] = true
	}
	sids := make([]string, 0, len(sidSet))
	for id := range sidSet {
		sids = append(sids, id)
	}
	rows, err := s.aiRepo.CompletedDurationStats(sids)
	if err != nil {
		return nil, err
	}
	stats := map[string]repository.DurationStat{}
	for _, r := range rows {
		stats[r.ServiceID+"|"+r.LeaderID] = r
	}
	endHour := delayDefaultEndHour
	if w, err := s.aiRepo.WorkWindow(); err == nil && w.EndHour > 0 {
		endHour = w.EndHour
	}
	// الحجز الجاي لنفس الليدر بنفس يوم بغداد (المرشحين مرتّبين بالموعد).
	byLeaderDay := map[string][]repository.DelayCandidate{}
	for _, c := range cands {
		if c.LeaderID == "" {
			continue
		}
		k := c.LeaderID + "|" + c.ScheduledAt.In(debriefLoc).Format("2006-01-02")
		byLeaderDay[k] = append(byLeaderDay[k], c)
	}
	for _, c := range cands {
		expected, samples, basis, ok := pickExpected(stats, c.ServiceID, c.LeaderID)
		if !ok {
			continue
		}
		var next *time.Time
		nextCode := ""
		if c.LeaderID != "" {
			k := c.LeaderID + "|" + c.ScheduledAt.In(debriefLoc).Format("2006-01-02")
			for _, o := range byLeaderDay[k] {
				if o.ID != c.ID && o.ScheduledAt.After(c.ScheduledAt) {
					t := o.ScheduledAt
					next, nextCode = &t, o.Code
					break
				}
			}
		}
		available, by := availableWindow(c.ScheduledAt, endHour, next)
		if !isDelayRisk(expected, available) {
			continue
		}
		r := DelayRisk{
			BookingID: c.ID, BookingCode: c.Code, ScheduledAt: c.ScheduledAt,
			ExpectedMinutes: int(expected + 0.5), AvailableMinutes: int(available + 0.5),
			Samples: samples, Basis: basis, LimitedBy: by,
		}
		if by == "NEXT_BOOKING" {
			r.NextBookingCode = nextCode
		}
		if r.AvailableMinutes < 0 {
			r.AvailableMinutes = 0
		}
		out = append(out, r)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ScheduledAt.Before(out[j].ScheduledAt) })
	return out, nil
}

// Upcoming خطر اليوم والأيام الستة الجاية — لشاشة التنسيق.
func (s *DelayPredictionService) Upcoming() ([]DelayRisk, error) {
	now := time.Now().In(debriefLoc)
	return s.Risks(now.Format("2006-01-02"), now.AddDate(0, 0, delayLookaheadDays).Format("2006-01-02"))
}

// RunMorningIfDue مرة باليوم بعد ٧ الصبح بغداد: تنبيه للمالك ومدير
// النظام ومنو عنده صلاحية coordinator بأكواد حجوزات اليوم الخطرة.
func (s *DelayPredictionService) RunMorningIfDue() error {
	now := time.Now().In(debriefLoc)
	if now.Hour() < delayMorningHour {
		return nil
	}
	today := now.Format("2006-01-02")
	claimed, err := s.aiRepo.ClaimDailyMarker("DAILY_DELAY_RISK_SENT", today)
	if err != nil || !claimed {
		return err
	}
	risks, err := s.Risks(today, today)
	if err != nil || len(risks) == 0 {
		return err
	}
	parts := []string{}
	for _, r := range risks {
		parts = append(parts, fmt.Sprintf("%s (~%s مقابل %s)", r.BookingCode, fmtHours(r.ExpectedMinutes), fmtHours(r.AvailableMinutes)))
	}
	msg := fmt.Sprintf("⏱️ ماتركس — %d حجز اليوم متوقع ياخذ أكثر من وقته المحجوز: %s. راجع الموعد أو الكادر قبل الطلعة.",
		len(risks), strings.Join(parts, "، "))
	return s.notifRepo.CreateForRolesOrPermission([]string{"OWNER", "ADMIN"}, "coordinator", "AI_DELAY_RISK", msg)
}

// fmtHours «٢.٥ ساعة» تقريباً لأقرب نص ساعة.
func fmtHours(minutes int) string {
	h := float64(int(float64(minutes)/30.0+0.5)) / 2.0
	if h == float64(int(h)) {
		return fmt.Sprintf("%d ساعة", int(h))
	}
	return fmt.Sprintf("%.1f ساعة", h)
}
