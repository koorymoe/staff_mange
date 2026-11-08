package service

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/jmoiron/sqlx"

	"staffmange-api/internal/database"
)

// (ع) 11-08: كل إشعار جديد ينبعث لجهاز صاحبه المشترك مرة وحدة، والجهاز الملغي ينشال.
func TestWebPushSendPending_Live(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL غير موجود")
	}
	db := sqlx.MustConnect("postgres", dsn)
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db, "", ""); err != nil {
		t.Fatal(err)
	}
	const emp = "__push_emp"
	clean := func() {
		db.Exec(`DELETE FROM "Notification" WHERE "employeeId" = $1`, emp)     //nolint:errcheck
		db.Exec(`DELETE FROM "PushSubscription" WHERE "employeeId" = $1`, emp) //nolint:errcheck
		db.Exec(`DELETE FROM "Employee" WHERE id = $1`, emp)                   //nolint:errcheck
	}
	clean()
	t.Cleanup(clean)
	db.MustExec(`INSERT INTO "Employee"(id, name) VALUES ($1, 'موظف إشعارات')`, emp)

	var hits atomic.Int32
	gone := false
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.Header.Get("Authorization") == "" || r.Header.Get("Content-Encoding") != "aes128gcm" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if gone {
			w.WriteHeader(http.StatusGone)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	wp, err := NewWebPush(db, "mailto:test@localhost")
	if err != nil {
		t.Fatal(err)
	}
	wp.client = srv.Client()
	again, _ := NewWebPush(db, "mailto:test@localhost")
	if again.PublicKey() != wp.PublicKey() {
		t.Fatal("المفتاح لازم يبقى نفسه بين التشغيلات")
	}
	priv, _ := ecdh.P256().GenerateKey(rand.Reader)
	auth := make([]byte, 16)
	rand.Read(auth) //nolint:errcheck
	b64 := base64.RawURLEncoding.EncodeToString
	if err := wp.Subscribe(emp, srv.URL+"/sub/1", b64(priv.PublicKey().Bytes()), b64(auth), "test"); err != nil {
		t.Fatal(err)
	}
	if err := wp.Subscribe(emp, "http://insecure", "a", "b", ""); err == nil {
		t.Fatal("الاشتراك لازم https")
	}
	db.MustExec(`INSERT INTO "Notification"(id, "employeeId", type, message) VALUES (gen_random_uuid()::text, $1, 'TEST', 'إشعار تجربة')`, emp)
	if n := wp.SendPending(); n != 1 || hits.Load() != 1 {
		t.Fatalf("لازم ينبعث مرة وحدة: sent=%d hits=%d", n, hits.Load())
	}
	if n := wp.SendPending(); n != 0 || hits.Load() != 1 {
		t.Fatal("ما لازم ينبعث مرتين")
	}
	gone = true
	db.MustExec(`INSERT INTO "Notification"(id, "employeeId", type, message) VALUES (gen_random_uuid()::text, $1, 'TEST', 'ثاني')`, emp)
	wp.SendPending()
	var subs int
	db.Get(&subs, `SELECT count(*) FROM "PushSubscription" WHERE "employeeId" = $1`, emp)
	if subs != 0 {
		t.Fatal("الجهاز الي لغى الاشتراك لازم ينشال")
	}
}
