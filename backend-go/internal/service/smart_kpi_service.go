package service

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type SmartKpiService struct {
	repo *repository.SmartKpiRepository
}

func NewSmartKpiService(repo *repository.SmartKpiRepository) *SmartKpiService {
	return &SmartKpiService{repo: repo}
}

func getMonthRange(month string) (start, end time.Time, year, mon int) {
	now := time.Now()
	if month != "" {
		parts := strings.Split(month, "-")
		if len(parts) == 2 {
			y, err1 := strconv.Atoi(parts[0])
			m, err2 := strconv.Atoi(parts[1])
			if err1 == nil && err2 == nil {
				year, mon = y, m
			}
		}
	}
	if year == 0 {
		year, mon = now.Year(), int(now.Month())
	}
	start = time.Date(year, time.Month(mon), 1, 0, 0, 0, 0, time.UTC)
	end = time.Date(year, time.Month(mon+1), 1, 0, 0, 0, 0, time.UTC)
	return start, end, year, mon
}

func getWorkingDays(year, month int) int {
	daysInMonth := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()
	workDays := 0
	for d := 1; d <= daysInMonth; d++ {
		day := time.Date(year, time.Month(month), d, 0, 0, 0, 0, time.UTC).Weekday()
		if day != time.Friday {
			workDays++
		}
	}
	return workDays
}

func (s *SmartKpiService) CalculateTechnicianKpi(employeeID, month string) (*model.SmartKpiResult, error) {
	start, end, year, mon := getMonthRange(month)

	employee, err := s.repo.GetEmployeeForKpi(employeeID)
	if err != nil {
		return nil, errors.New("Employee not found")
	}
	if employee.Role != "TECHNICIAN" {
		return nil, errors.New("Employee is not a technician")
	}

	// 1. Completed bookings via assignments
	bCreated, bCompleted, err := s.repo.CompletedBookingTimes(employeeID, start, end)
	if err != nil {
		return nil, err
	}
	completedCount := len(bCreated)
	completedBookingsPoints := completedCount * 10

	// 2. Completion speed
	totalMinutes := 0.0
	speedEntries := 0
	mAssigned, mCompleted, err := s.repo.MissionTimes(employeeID, start, end)
	if err != nil {
		return nil, err
	}
	if len(mAssigned) > 0 {
		for i := range mAssigned {
			totalMinutes += mCompleted[i].Sub(mAssigned[i]).Minutes()
			speedEntries++
		}
	} else {
		for i := range bCreated {
			totalMinutes += bCompleted[i].Sub(bCreated[i]).Minutes()
			speedEntries++
		}
	}

	avgMinutes := 0
	if speedEntries > 0 {
		avgMinutes = int(math.Round(totalMinutes / float64(speedEntries)))
	}
	speedPoints := 0
	if speedEntries > 0 {
		if avgMinutes <= 120 {
			speedPoints = 20
		} else if avgMinutes <= 240 {
			speedPoints = 10
		} else {
			speedPoints = 5
		}
	}

	// 3. Work reports
	fullReports, totalReports, err := s.repo.WorkReportFlags(employeeID, start, end)
	if err != nil {
		return nil, err
	}
	partialReports := totalReports - fullReports
	workReportPoints := fullReports*10 + partialReports*5

	// 4. Attendance
	daysPresent, err := s.repo.AttendanceDaysPresent(employeeID, start, end)
	if err != nil {
		return nil, err
	}
	totalDays := getWorkingDays(year, mon)
	attendanceRatio := 0.0
	if totalDays > 0 {
		attendanceRatio = float64(daysPresent) / float64(totalDays)
	}
	var attendancePoints int
	if attendanceRatio >= 1 {
		attendancePoints = 20
	} else {
		attendancePoints = int(math.Round(attendanceRatio * 20))
	}

	// 5. Complaints
	complaintsCount, err := s.repo.ComplaintsCount(employeeID, start, end)
	if err != nil {
		return nil, err
	}
	complaintsPoints := 15
	if complaintsCount > 0 {
		complaintsPoints = -(complaintsCount * 10)
	}

	// 6. Manual KPI deductions
	manualDeductionPoints, manualDeductionCount, err := s.repo.ManualDeductionPoints(employeeID, start, end)
	if err != nil {
		return nil, err
	}

	// ⚠️ نجمع مو نطرح: التقييم اليدوي مخزون بإشارته (الخصم سالب).
	// الطرح كان يقلب المعنى — الخصم يزيد النقاط والمكافأة تنقصها.
	totalPoints := completedBookingsPoints + speedPoints + workReportPoints + attendancePoints + complaintsPoints + manualDeductionPoints

	return &model.SmartKpiResult{
		EmployeeID:   employee.ID,
		EmployeeName: employee.Name,
		Period:       fmt.Sprintf("%d-%02d", year, mon),
		Breakdown: model.SmartKpiBreakdown{
			CompletedBookings: model.SmartKpiBreakdownEntry{Count: completedCount, Points: completedBookingsPoints},
			CompletionSpeed:   model.CompletionSpeedEntry{AvgMinutes: avgMinutes, Points: speedPoints},
			WorkReports:       model.WorkReportsEntry{Count: totalReports, FullReports: fullReports, Points: workReportPoints},
			Attendance:        model.AttendanceEntry{DaysPresent: daysPresent, TotalDays: totalDays, Points: attendancePoints},
			Complaints:        model.SmartKpiBreakdownEntry{Count: complaintsCount, Points: complaintsPoints},
			ManualDeductions:  model.SmartKpiBreakdownEntry{Count: manualDeductionCount, Points: manualDeductionPoints},
		},
		TotalPoints: totalPoints,
	}, nil
}

func (s *SmartKpiService) Leaderboard(month string) ([]model.SmartKpiResult, error) {
	ids, err := s.repo.ListActiveTechnicianIDs()
	if err != nil {
		return nil, err
	}

	results := make([]model.SmartKpiResult, 0, len(ids))
	for _, id := range ids {
		r, err := s.CalculateTechnicianKpi(id, month)
		if err != nil {
			continue
		}
		results = append(results, *r)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalPoints > results[j].TotalPoints
	})
	return results, nil
}
