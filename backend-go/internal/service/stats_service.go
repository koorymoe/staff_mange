package service

import (
	"sort"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

type StatsService struct {
	repo *repository.StatsRepository
}

func NewStatsService(repo *repository.StatsRepository) *StatsService {
	return &StatsService{repo: repo}
}

func (s *StatsService) Overview() (*model.StatsOverview, error) {
	totals, err := s.repo.Totals()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	salesStats, err := s.repo.SalesStats(startOfToday, startOfMonth)
	if err != nil {
		return nil, err
	}
	sort.Slice(salesStats, func(i, j int) bool { return salesStats[i].Confirmed > salesStats[j].Confirmed })

	coordinatorStats, err := s.repo.CoordinatorStats(startOfToday, startOfMonth)
	if err != nil {
		return nil, err
	}
	sort.Slice(coordinatorStats, func(i, j int) bool { return coordinatorStats[i].TotalConfirmed > coordinatorStats[j].TotalConfirmed })

	technicianStats, err := s.repo.TechnicianStats()
	if err != nil {
		return nil, err
	}
	sort.Slice(technicianStats, func(i, j int) bool { return technicianStats[i].Completed > technicianStats[j].Completed })

	serviceBreakdown, err := s.repo.ServiceBreakdown()
	if err != nil {
		return nil, err
	}
	sort.Slice(serviceBreakdown, func(i, j int) bool { return serviceBreakdown[i].Count > serviceBreakdown[j].Count })

	roleCounts, err := s.repo.RoleCounts()
	if err != nil {
		return nil, err
	}

	recentBookings, err := s.repo.RecentBookings(8)
	if err != nil {
		return nil, err
	}

	return &model.StatsOverview{
		Totals:           *totals,
		SalesStats:       salesStats,
		CoordinatorStats: coordinatorStats,
		TechnicianStats:  technicianStats,
		ServiceBreakdown: serviceBreakdown,
		RoleCounts:       roleCounts,
		RecentBookings:   recentBookings,
	}, nil
}
