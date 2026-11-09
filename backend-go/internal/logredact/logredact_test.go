package logredact

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

// ٦ز — اختبار يطبع أسرار، ولازم تطلع محجوبة.
func TestRedact(t *testing.T) {
	secrets := []string{
		"eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
		"sk-ant-api03-AbCdEfGhIjKlMnOpQrStUvWxYz0123456789",
		"AIzaSyA1234567890abcdefghijklmnopqrstu",
		"ghp_abcdefghijklmnopqrstuvwxyz0123",
		"Owner-pass-123",
		"s3cr3t-db-pass",
	}
	lines := []string{
		"auth header: Authorization: Bearer " + secrets[0],
		"calling anthropic with key " + secrets[1],
		"google " + secrets[2] + " github " + secrets[3],
		`login body {"username":"owner","password":"` + secrets[4] + `"}`,
		"DATABASE_URL=postgres://staffmange:" + secrets[5] + "@localhost:5432/x",
		"retry ?token=abcdef123456&x=1 api_key=zzz999yyy",
	}
	var buf bytes.Buffer
	l := log.New(Writer{W: &buf}, "", 0)
	for _, s := range lines {
		l.Println(s)
	}
	out := buf.String()
	for _, s := range append(secrets, "abcdef123456", "zzz999yyy") {
		if strings.Contains(out, s) {
			t.Errorf("السر %q طلع بالسجل:\n%s", s, out)
		}
	}
	if !strings.Contains(out, "username") || !strings.Contains(out, "localhost:5432") {
		t.Errorf("الحجب مسح نص مو سر:\n%s", out)
	}
}
