package handler

import (
	"net/http"
	"runtime"
	"time"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
)

type SecurityHandler struct {
	loginAudit *repository.LoginAuditRepository
	startedAt  time.Time
}

func NewSecurityHandler(loginAudit *repository.LoginAuditRepository, startedAt time.Time) *SecurityHandler {
	return &SecurityHandler{loginAudit: loginAudit, startedAt: startedAt}
}

type SecurityDashboardResponse struct {
	ServerUptimeSeconds  int64   `json:"serverUptimeSeconds"`
	GoroutineCount       int     `json:"goroutineCount"`
	MemoryUsedMB         float64 `json:"memoryUsedMB"`
	FailedLoginsLastHour int     `json:"failedLoginsLastHour"`
	TotalRequests        int64   `json:"totalRequests"`
	RequestsLastMinute   int64   `json:"requestsLastMinute"`
	RecentLogins         any     `json:"recentLogins"`
}

// GET /api/security/dashboard — حصري لمدير النظام/المالك: مؤشرات صحة السيرفر
// (ذاكرة/uptime) + سجل محاولات الدخول الأخيرة (ناجحة وفاشلة) بعنوان IP
// والمتصفح/الجهاز. ملاحظة: المتصفح لا يكشف عنوان MAC الفعلي لأي جهاز تقنياً.
func (h *SecurityHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	failedCount, err := h.loginAudit.FailedAttemptsCount()
	if err != nil {
		failedCount = 0
	}
	recent, err := h.loginAudit.Recent(100)
	if err != nil {
		recent = nil
	}

	WriteJSON(w, http.StatusOK, SecurityDashboardResponse{
		ServerUptimeSeconds:  int64(time.Since(h.startedAt).Seconds()),
		GoroutineCount:       runtime.NumGoroutine(),
		MemoryUsedMB:         float64(mem.Alloc) / 1024 / 1024,
		FailedLoginsLastHour: failedCount,
		TotalRequests:        middleware.TotalRequests(),
		RequestsLastMinute:   middleware.RequestsLastMinute(),
		RecentLogins:         recent,
	})
}
