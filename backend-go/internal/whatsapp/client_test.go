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

// الفويس والمكالمات (دازاي 11-15): audio يطلع بمعرّف الوسيط، والمكالمة تنقرا بس من تخلص (terminate)
// ومن الزبون — connect ومكالمات الشركة تنتجاهل.
func TestParseVoiceAndCalls(t *testing.T) {
	body := []byte(`{"entry":[{"changes":[{"field":"messages","value":{"messages":[
		{"id":"wamid.a","from":"9647700000000","timestamp":"1761235200","type":"audio","audio":{"id":"MEDIA1","mime_type":"audio/ogg; codecs=opus"}},
		{"id":"wamid.i","from":"9647700000000","timestamp":"1761235201","type":"image","image":{"id":"X"}}]}},
		{"field":"calls","value":{"calls":[
		{"id":"wacid.1","from":"9647700000001","to":"9647800000000","event":"connect","direction":"USER_INITIATED","timestamp":"1761235300"},
		{"id":"wacid.1","from":"9647700000001","to":"9647800000000","event":"terminate","direction":"USER_INITIATED","timestamp":"1761235330","status":"COMPLETED"},
		{"id":"wacid.2","from":"9647700000002","event":"terminate","direction":"USER_INITIATED","timestamp":"1761235400","status":"COMPLETED","start_time":"1761235340","end_time":"1761235400","duration":60},
		{"id":"wacid.3","from":"9647700000003","event":"terminate","direction":"BUSINESS_INITIATED","timestamp":"1761235500","status":"COMPLETED"}]}}]}]}`)
	in, err := ParseWebhook(body)
	if err != nil || len(in) != 4 {
		t.Fatalf("parse: %d %v", len(in), err)
	}
	if in[0].Audio == nil || in[0].Audio.MediaID != "MEDIA1" || in[0].Text != "" {
		t.Fatalf("audio: %+v", in[0])
	}
	if in[1].Audio != nil || in[1].Text == "" {
		t.Fatalf("image should be a text placeholder: %+v", in[1])
	}
	if in[2].Call == nil || in[2].Call.Answered || in[2].From != "9647700000001" || in[2].MetaID != "wacid.1" {
		t.Fatalf("missed call: %+v", in[2])
	}
	if in[3].Call == nil || !in[3].Call.Answered || in[3].Call.Duration != 60 {
		t.Fatalf("answered call: %+v", in[3])
	}
}
