package service

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"staffmange-api/internal/repository"
)

// بلاغ #52 (11-19): المراقب يشوف التدقيق وصندوقه بس، حتى لو عنده صلاحية الفواتير والجودة.
func TestMonitorWorkloadAuditOnlyLive(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	defer db.Close()
	s := &MatrixAutopilotService{actions: repository.NewAiActionRepository(db), aiRepo: repository.NewAiRepository(db)}
	keys := func(role string) map[string]bool {
		out := map[string]bool{}
		for _, w := range s.workload(repository.WatchSubject{ID: "x", Role: role, Perms: []string{"finance", "quality_control"}}) {
			out[w.Key] = true
		}
		return out
	}
	m := keys("MONITOR")
	if !m["AUDIT"] || !m["MONITOR_INBOX"] || m["APPROVE"] || m["QUALITY"] {
		t.Fatalf("monitor workload = %v", m)
	}
	o := keys("SALES") // غير المراقب ياخذها بصلاحيته
	if !o["APPROVE"] || !o["QUALITY"] {
		t.Fatalf("sales with perms = %v", o)
	}
}
