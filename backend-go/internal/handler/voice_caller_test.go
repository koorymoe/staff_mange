package handler

import "testing"

// (دازاي 11-16) زر التجربة بالمنصة ما يدز رقم: إذا المالك حط VOICE_TEST_PHONE تنحسب منه وتكتب حقيقي.
func TestVoiceCaller(t *testing.T) {
	cases := []struct{ phone, test, want string }{
		{"", "", ""},                                      // تجربة بلا رقم تجربة = وضع تجربة
		{"", " 07701234567 ", "07701234567"},              // تجربة ويا رقم تجربة = حقيقي من هالرقم
		{"9647709999999", "07701234567", "9647709999999"}, // مكالمة حقيقية: رقمها هو
	}
	for _, c := range cases {
		if got := voiceCaller(c.phone, c.test); got != c.want {
			t.Fatalf("voiceCaller(%q,%q)=%q، المفروض %q", c.phone, c.test, got, c.want)
		}
	}
}
