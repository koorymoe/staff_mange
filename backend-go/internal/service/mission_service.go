package service

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type MissionService struct {
	repo *repository.MissionRepository
}

func NewMissionService(repo *repository.MissionRepository) *MissionService {
	return &MissionService{repo: repo}
}

func (s *MissionService) List(stage, leaderID, employeeID string) ([]model.Mission, error) {
	return s.repo.List(stage, leaderID, employeeID)
}

func (s *MissionService) Get(id string) (*model.Mission, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, errors.New("المهمة غير موجودة")
	}
	return m, nil
}

func parseFloatPtr(s *string) *float64 {
	if s == nil || *s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(*s, 64)
	if err != nil {
		return nil
	}
	return &f
}

func parseIntPtr(s *string) *int {
	if s == nil || *s == "" {
		return nil
	}
	i, err := strconv.Atoi(*s)
	if err != nil {
		return nil
	}
	return &i
}

func (s *MissionService) Create(req model.CreateMissionRequest) (*model.Mission, error) {
	if req.BookingID == "" || req.LeaderID == "" {
		return nil, errors.New("bookingId و leaderId مطلوبين")
	}
	count, err := s.repo.CountAll()
	if err != nil {
		return nil, err
	}
	code := fmt.Sprintf("MSN-%04d", count+1)
	memberIDs := req.MemberIDs
	if memberIDs == nil {
		memberIDs = []string{}
	}
	return s.repo.Create(code, req.BookingID, req.LeaderID, memberIDs, parseFloatPtr(req.CustomerLat), parseFloatPtr(req.CustomerLng), req.CustomerAddress)
}

// EnsureForBooking ينفّذ BookingService.MissionStarter.
//
// idempotent بالكامل: تنستدعى بكل تكليف — والتكليف ينعاد كثير
// (تبديل فني، إضافة ثاني للكادر، إعادة تكليف بالغلط). أول مرة تخلق
// المهمة، وبعدها تحدّث الكادر بس.
//
// ⚠️ ما تلمس المرحلة: مهمة وصلت «بالطريق» وانضاف إلها فني ما ترجع
// «تم الإسناد» — الشغل ماشي، بس الكادر توسّع. إرجاعها للبداية يمحي
// توقيتات حقيقية انتسجّلت بالميدان.
func (s *MissionService) EnsureForBooking(
	bookingID, leaderID string, memberIDs []string, address *string, lat, lng *float64,
) error {
	if bookingID == "" || leaderID == "" {
		return errors.New("bookingId و leaderId مطلوبين")
	}
	exists, err := s.repo.ExistsForBooking(bookingID)
	if err != nil {
		return err
	}
	if memberIDs == nil {
		memberIDs = []string{}
	}
	if exists {
		return s.repo.SyncCrew(bookingID, leaderID, memberIDs)
	}
	count, err := s.repo.CountAll()
	if err != nil {
		return err
	}
	_, err = s.repo.Create(fmt.Sprintf("MSN-%04d", count+1), bookingID, leaderID, memberIDs, lat, lng, address)
	return err
}

// BackfillOnce يعوّض الحجوزات الشغّالة الي انكلّفت قبل ما ينربط
// التوليد التلقائي. آمنة بالتكرار — تتخطى الي عندها مهمة أصلاً.
func (s *MissionService) BackfillOnce() {
	n, err := s.repo.BackfillFromAssignments()
	if err != nil {
		log.Printf("[mission] تعذر تعويض المهام الناقصة: %v", err)
		return
	}
	if n > 0 {
		log.Printf("[mission] انخلقت %d مهمة للحجوزات الشغّالة الي جانت بلا مهمة", n)
	}
}

func (s *MissionService) UpdateStage(id string, req model.UpdateMissionStageRequest) (*model.Mission, error) {
	if req.Stage == "" || req.EmployeeID == "" {
		return nil, errors.New("stage و employeeId مطلوبين")
	}
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("المهمة غير موجودة")
	}

	now := time.Now()
	fields := map[string]any{}
	lat := parseFloatPtr(req.Lat)
	lng := parseFloatPtr(req.Lng)

	switch req.Stage {
	case "MATERIALS_PREP":
		// no fields
	case "MATERIALS_READY":
		fields["materialsReadyAt"] = now
	case "EN_ROUTE":
		fields["departedAt"] = now
		if lat != nil {
			fields["departureLat"] = *lat
		}
		if lng != nil {
			fields["departureLng"] = *lng
		}
		if em := parseIntPtr(req.EstimatedMinutes); em != nil {
			fields["estimatedMinutes"] = *em
		}
		if dk := parseFloatPtr(req.DistanceKm); dk != nil {
			fields["distanceKm"] = *dk
		}
	case "ARRIVED":
		fields["arrivedAt"] = now
		if lat != nil {
			fields["arrivalLat"] = *lat
		}
		if lng != nil {
			fields["arrivalLng"] = *lng
		}
		if existing.DepartedAt != nil {
			actual := int(now.Sub(*existing.DepartedAt).Minutes() + 0.5)
			fields["actualMinutes"] = actual
		}
	case "WORK_STARTED":
		fields["workStartedAt"] = now
	case "COMPLETED":
		fields["completedAt"] = now
	case "STOPPED":
		fields["stoppedAt"] = now
		fields["stopReason"] = req.StopReason
	}

	if req.Note != nil && *req.Note != "" {
		fields["notes"] = *req.Note
	}

	if _, err := s.repo.UpdateStage(id, req.Stage, fields); err != nil {
		return nil, err
	}
	if err := s.repo.CreateEvent(id, req.EmployeeID, req.Stage, lat, lng, req.Note); err != nil {
		return nil, err
	}

	return s.repo.FindByID(id)
}

func (s *MissionService) ListForEmployee(employeeID string) ([]model.Mission, error) {
	return s.repo.ListForEmployee(employeeID)
}

func (s *MissionService) MonitorLive() (*model.MissionMonitorResponse, error) {
	missions, err := s.repo.ListActive()
	if err != nil {
		return nil, err
	}

	stats := model.MissionMonitorStats{Total: len(missions)}
	for _, m := range missions {
		switch m.Stage {
		case "ASSIGNED":
			stats.Assigned++
		case "MATERIALS_PREP", "MATERIALS_READY":
			stats.Preparing++
		case "EN_ROUTE":
			stats.EnRoute++
		case "ARRIVED":
			stats.Arrived++
		case "WORK_STARTED":
			stats.Working++
		}
	}

	return &model.MissionMonitorResponse{Missions: missions, Stats: stats}, nil
}

func (s *MissionService) PerformanceReport(from, to *string) ([]model.MissionPerformanceReport, error) {
	missions, err := s.repo.ListForPerformanceReport(from, to)
	if err != nil {
		return nil, err
	}

	empIDs := map[string]bool{}
	for _, m := range missions {
		empIDs[m.LeaderID] = true
		for _, id := range m.MemberIDs {
			empIDs[id] = true
		}
	}
	ids := make([]string, 0, len(empIDs))
	for id := range empIDs {
		ids = append(ids, id)
	}
	briefs, err := s.repo.LoadEmployeeBriefsByIDs(ids)
	if err != nil {
		return nil, err
	}

	report := make([]model.MissionPerformanceReport, 0, len(briefs))
	for _, emp := range briefs {
		var empMissions []model.Mission
		for _, m := range missions {
			if m.LeaderID == emp.ID || containsStr(m.MemberIDs, emp.ID) {
				empMissions = append(empMissions, m)
			}
		}
		completed := 0
		stopped := 0
		onTime := 0
		late := 0
		lateDelaySum := 0
		for _, m := range empMissions {
			switch m.Stage {
			case "COMPLETED":
				completed++
				if m.EstimatedMinutes != nil && m.ActualMinutes != nil {
					if *m.ActualMinutes <= *m.EstimatedMinutes {
						onTime++
					} else {
						late++
						lateDelaySum += *m.ActualMinutes - *m.EstimatedMinutes
					}
				}
			case "STOPPED":
				stopped++
			}
		}
		avgDelay := 0
		if late > 0 {
			avgDelay = int(float64(lateDelaySum)/float64(late) + 0.5)
		}
		compliance := 0
		if completed > 0 {
			compliance = int(float64(onTime)/float64(completed)*100 + 0.5)
		}
		report = append(report, model.MissionPerformanceReport{
			Employee:          emp,
			TotalMissions:     len(empMissions),
			Completed:         completed,
			Stopped:           stopped,
			OnTime:            onTime,
			Late:              late,
			AvgDelayMinutes:   avgDelay,
			CompliancePercent: compliance,
		})
	}

	sortReportByTotalMissionsDesc(report)
	return report, nil
}

func containsStr(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func sortReportByTotalMissionsDesc(report []model.MissionPerformanceReport) {
	for i := 1; i < len(report); i++ {
		for j := i; j > 0 && report[j].TotalMissions > report[j-1].TotalMissions; j-- {
			report[j], report[j-1] = report[j-1], report[j]
		}
	}
}
