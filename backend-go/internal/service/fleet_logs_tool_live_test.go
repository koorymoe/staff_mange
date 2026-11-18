package service

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func TestFleetLogsToolLive(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("no DATABASE_URL")
	}
	db := sqlx.MustConnect("postgres", url)
	vid := "test-fleet-v1"
	db.MustExec(`DELETE FROM "Vehicle" WHERE id=$1`, vid)
	t.Cleanup(func() { db.MustExec(`DELETE FROM "Vehicle" WHERE id=$1`, vid); db.Close() })
	db.MustExec(`INSERT INTO "Vehicle"(id,name,"plateNumber") VALUES ($1,'كيا اختبار','TST-FLEET-1')`, vid)
	db.MustExec(`INSERT INTO "VehicleLog"(id,"vehicleId",type,"performedAt") VALUES ('test-fleet-l1',$1,'CLEANING',$2)`, vid, time.Now().UTC())
	out, err := fleetLogsTool(db).Run(map[string]any{"type": "CLEANING"}, nil)
	if err != nil || !strings.Contains(out, "TST-FLEET-1") || !strings.Contains(out, "غسيل") {
		t.Fatalf("out=%q err=%v", out, err)
	}
}
