package authz

import "testing"

// (دازاي 11-04) محادثات الواتساب: المالك، أو الصلاحية الممنوحة — ولا الأدمن ولا المراقب بلاها.
func TestWhatsAppChatsRule(t *testing.T) {
	cases := []struct {
		eff, raw string
		perms    []string
		want     bool
	}{
		{"ADMIN", "OWNER", nil, true},
		{"ADMIN", "ADMIN", nil, false},
		{"ADMIN", "MONITOR", []string{"monitoring", "quality_control"}, false},
		{"QUALITY_ENGINEER", "QUALITY_ENGINEER", []string{"quality_control", "coordinator"}, false},
		{"TECHNICIAN", "TECHNICIAN", []string{"whatsapp_chats"}, true},
		{"ADMIN", "ADMIN", []string{"whatsapp_chats"}, true},
	}
	for _, c := range cases {
		if got := Decide(WhatsAppChats, c.eff, c.raw, c.perms); got != c.want {
			t.Errorf("%s/%s %v = %v، المفروض %v", c.eff, c.raw, c.perms, got, c.want)
		}
	}
}
