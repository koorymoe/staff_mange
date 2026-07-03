package model

import "time"

type StatsTotals struct {
	TotalCustomers    int     `json:"totalCustomers"`
	TotalBookings     int     `json:"totalBookings"`
	PendingBookings   int     `json:"pendingBookings"`
	ConfirmedBookings int     `json:"confirmedBookings"`
	CompletedBookings int     `json:"completedBookings"`
	CancelledBookings int     `json:"cancelledBookings"`
	UrgentPending     int     `json:"urgentPending"`
	TotalRevenue      float64 `json:"totalRevenue"`
	UnverifiedRevenue float64 `json:"unverifiedRevenue"`
}

type SalesStat struct {
	EmployeeID       string `json:"employeeId"`
	Name             string `json:"name"`
	TotalTransferred int    `json:"totalTransferred"`
	Confirmed        int    `json:"confirmed"`
	Today            int    `json:"today"`
	ThisMonth        int    `json:"thisMonth"`
}

type CoordinatorStat struct {
	EmployeeID     string `json:"employeeId"`
	Name           string `json:"name"`
	TotalConfirmed int    `json:"totalConfirmed"`
	Today          int    `json:"today"`
	ThisMonth      int    `json:"thisMonth"`
}

type TechnicianStat struct {
	EmployeeID     string  `json:"employeeId"`
	Name           string  `json:"name"`
	OnDuty         bool    `json:"onDuty"`
	TotalAssigned  int     `json:"totalAssigned"`
	Completed      int     `json:"completed"`
	RevenueHandled float64 `json:"revenueHandled"`
}

type ServiceBreakdownEntry struct {
	ServiceID *string `json:"serviceId"`
	Name      string  `json:"name"`
	Count     int     `json:"count"`
}

type RoleCount struct {
	Role  string `json:"role"`
	Count int    `json:"count"`
}

type RecentBookingEntry struct {
	ID           string    `json:"id"`
	Code         string    `json:"code"`
	Status       string    `json:"status"`
	Priority     string    `json:"priority"`
	CustomerName string    `json:"customerName"`
	ServiceName  *string   `json:"serviceName"`
	CreatedAt    time.Time `json:"createdAt"`
}

type StatsOverview struct {
	Totals           StatsTotals             `json:"totals"`
	SalesStats       []SalesStat             `json:"salesStats"`
	CoordinatorStats []CoordinatorStat       `json:"coordinatorStats"`
	TechnicianStats  []TechnicianStat        `json:"technicianStats"`
	ServiceBreakdown []ServiceBreakdownEntry `json:"serviceBreakdown"`
	RoleCounts       []RoleCount             `json:"roleCounts"`
	RecentBookings   []RecentBookingEntry    `json:"recentBookings"`
}
