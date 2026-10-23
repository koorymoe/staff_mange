package whatsapp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestSignatureAndParse(t *testing.T) {
	c := New("tok", "123", "secret", "verify")
	body := []byte(`{"entry":[{"changes":[{"value":{"contacts":[{"profile":{"name":"Ali"},"wa_id":"9647700000000"}],"messages":[{"id":"wamid.1","from":"9647700000000","timestamp":"1761235200","type":"text","text":{"body":"هلا"}}]}}]}]}`)
	m := hmac.New(sha256.New, []byte("secret"))
	m.Write(body)
	good := "sha256=" + hex.EncodeToString(m.Sum(nil))
	if !c.ValidSignature(body, good) {
		t.Fatal("valid signature rejected")
	}
	if c.ValidSignature(body, "sha256=00") || c.ValidSignature(append(body, ' '), good) {
		t.Fatal("bad signature accepted")
	}
	if !c.VerifyToken("verify") || c.VerifyToken("x") {
		t.Fatal("verify token")
	}
	in, err := ParseWebhook(body)
	if err != nil || len(in) != 1 || in[0].Text != "هلا" || in[0].MetaID != "wamid.1" || in[0].Name != "Ali" {
		t.Fatalf("parse: %+v %v", in, err)
	}
	if New("", "", "", "").On() {
		t.Fatal("empty client should be off")
	}
}
