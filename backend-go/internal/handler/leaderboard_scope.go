package handler

import (
	"net/http"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/model"
)

// لوحات الترتيب (ع 11-06): «خليها للمشرف بس، والباقي الاسم والمجموع».
// المشرف يشوف التفاصيل كاملة، والموظف يشوف صفّه كامل وزملاءه بالاسم والمجموع.

func leaderboardFull(r *http.Request) bool {
	return canSeeOperational(middleware.EffectiveRole(r))
}

func limitKpiRows(r *http.Request, rows []model.KpiLeaderboardEntry) []model.KpiLeaderboardEntry {
	if leaderboardFull(r) {
		return rows
	}
	me := middleware.EmployeeIDFromContext(r)
	out := make([]model.KpiLeaderboardEntry, len(rows))
	for i, e := range rows {
		if e.EmployeeID == me {
			out[i] = e
		} else {
			out[i] = e.LimitTo()
		}
	}
	return out
}

func limitSmartKpi(r *http.Request, rows []model.SmartKpiResult) []model.SmartKpiResult {
	if leaderboardFull(r) {
		return rows
	}
	me := middleware.EmployeeIDFromContext(r)
	out := make([]model.SmartKpiResult, len(rows))
	for i, e := range rows {
		if e.EmployeeID == me {
			out[i] = e
		} else {
			out[i] = model.SmartKpiResult{EmployeeID: e.EmployeeID, EmployeeName: e.EmployeeName, Period: e.Period, TotalPoints: e.TotalPoints, Limited: true}
		}
	}
	return out
}
