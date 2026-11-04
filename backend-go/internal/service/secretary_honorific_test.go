package service

import "testing"

func TestSecretaryHonorific(t *testing.T) {
	s := func(v string) *string { return &v }
	cases := map[string]string{"دكتوراه إدارة أعمال": "دكتور", "PhD": "دكتور", "ماجستير": "أستاذ", "بكالوريوس هندسة كهرباء": "مهندس", "اعدادية": "سيدنا"}
	for in, want := range cases {
		if got := SecretaryHonorific(s(in)); got != want {
			t.Errorf("%q: got %q want %q", in, got, want)
		}
	}
	if SecretaryHonorific(nil) != "سيدنا" {
		t.Error("nil")
	}
}
