package handler

import (
	"net/http"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/middleware"
	"staffmange-api/internal/repository"
)

type SecurityHandler struct {
	db         *sqlx.DB
	loginAudit *repository.LoginAuditRepository
	startedAt  time.Time
}

func NewSecurityHandler(db *sqlx.DB, loginAudit *repository.LoginAuditRepository, startedAt time.Time) *SecurityHandler {
	return &SecurityHandler{db: db, loginAudit: loginAudit, startedAt: startedAt}
}

type SecurityDashboardResponse struct {
	ServerUptimeSeconds  int64   `json:"serverUptimeSeconds"`
	GoroutineCount       int     `json:"goroutineCount"`
	CPUCount             int     `json:"cpuCount"`
	MemoryUsedMB         float64 `json:"memoryUsedMB"`
	FailedLoginsLastHour int     `json:"failedLoginsLastHour"`
	TotalRequests        int64   `json:"totalRequests"`
	RequestsLastMinute   int64   `json:"requestsLastMinute"`
	DiskTotalGB          float64 `json:"diskTotalGB"`
	DiskUsedGB           float64 `json:"diskUsedGB"`
	DiskFreeGB           float64 `json:"diskFreeGB"`
	DBSizeMB             float64 `json:"dbSizeMB"`
	DBConnectionsOpen    int     `json:"dbConnectionsOpen"`
	DBConnectionsInUse   int     `json:"dbConnectionsInUse"`
	OnlineEmployees      int     `json:"onlineEmployees"`
	RecentLogins         any     `json:"recentLogins"`
}

// GET /api/security/dashboard — حصري لمدير النظام/المالك: مؤشرات صحة السيرفر
// الكاملة (ذاكرة/قرص/قاعدة بيانات/ضغط) + سجل محاولات الدخول الأخيرة (ناجحة
// وفاشلة) بعنوان IP والمتصفح/الجهاز. ملاحظة: المتصفح لا يكشف عنوان MAC
// الفعلي لأي جهاز تقنياً — قيد أمان بكل المتصفحات الحديثة.
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
	onlineCount, err := h.loginAudit.OnlineEmployeesCount()
	if err != nil {
		onlineCount = 0
	}

	var dbSizeBytes float64
	_ = h.db.Get(&dbSizeBytes, `SELECT pg_database_size(current_database())`)

	dbStats := h.db.Stats()
	diskTotal, diskUsed, diskFree := diskUsage()

	WriteJSON(w, http.StatusOK, SecurityDashboardResponse{
		ServerUptimeSeconds:  int64(time.Since(h.startedAt).Seconds()),
		GoroutineCount:       runtime.NumGoroutine(),
		CPUCount:             runtime.NumCPU(),
		MemoryUsedMB:         float64(mem.Alloc) / 1024 / 1024,
		FailedLoginsLastHour: failedCount,
		TotalRequests:        middleware.TotalRequests(),
		RequestsLastMinute:   middleware.RequestsLastMinute(),
		DiskTotalGB:          diskTotal,
		DiskUsedGB:           diskUsed,
		DiskFreeGB:           diskFree,
		DBSizeMB:             dbSizeBytes / 1024 / 1024,
		DBConnectionsOpen:    dbStats.OpenConnections,
		DBConnectionsInUse:   dbStats.InUse,
		OnlineEmployees:      onlineCount,
		RecentLogins:         recent,
	})
}

// POST /api/security/free-memory — يطلب من Go يحرر الذاكرة الفاضية للنظام
// (نفس فكرة "تفريغ الكاش")، ويرجع استهلاك الذاكرة الجديد فوراً.
func (h *SecurityHandler) FreeMemory(w http.ResponseWriter, r *http.Request) {
	debug.FreeOSMemory()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	WriteJSON(w, http.StatusOK, map[string]float64{"memoryUsedMB": float64(mem.Alloc) / 1024 / 1024})
}
