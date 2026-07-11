package service

import (
	"fmt"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type AttendanceService struct {
	repo *repository.AttendanceRepository
}

func NewAttendanceService(repo *repository.AttendanceRepository) *AttendanceService {
	return &AttendanceService{repo: repo}
}

func (s *AttendanceService) CheckIn(employeeID string) (*model.Attendance, error) {
	return s.repo.CheckIn(employeeID)
}

func (s *AttendanceService) CheckOut(employeeID string) (*model.Attendance, error) {
	return s.repo.CheckOut(employeeID)
}

func (s *AttendanceService) Today() ([]model.Attendance, error) {
	return s.repo.Today()
}

func (s *AttendanceService) Mine(employeeID string) (*model.Attendance, error) {
	return s.repo.FindToday(employeeID)
}

// month is expected in "YYYY-MM" form; defaults to the current month when empty.
func (s *AttendanceService) MonthlyReport(employeeID, month string) (*model.MonthlyAttendanceReport, error) {
	var start time.Time
	var err error
	if month == "" {
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	} else {
		start, err = time.Parse("2006-01", month)
		if err != nil {
			return nil, fmt.Errorf("صيغة الشهر غير صحيحة")
		}
	}
	end := start.AddDate(0, 1, 0)

	records, err := s.repo.ForEmployeeInRange(employeeID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}

	report := &model.MonthlyAttendanceReport{
		EmployeeID: employeeID,
		Month:      start.Format("2006-01"),
		Days:       records,
	}
	for _, rec := range records {
		if rec.CheckOut != nil {
			report.TotalMinutes += int(rec.CheckOut.Sub(rec.CheckIn).Minutes())
		}
	}
	report.DaysPresent = len(records)
	return report, nil
}

func (s *AttendanceService) Correct(id string, req model.SetAttendanceCorrectionRequest) (*model.Attendance, error) {
	return s.repo.Correct(id, req.CheckIn, req.CheckOut)
}
