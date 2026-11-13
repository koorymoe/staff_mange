package model

import "testing"

func TestLocalPhone(t *testing.T) {
	for in, want := range map[string]string{
		"+9647701234567": "07701234567", "009647701234567": "07701234567", "0770 123 4567": "07701234567",
		"7701234567": "07701234567", "٠٧٧٠١٢٣٤٥٦٧": "07701234567", "07701234567": "07701234567", "abc": "",
	} {
		if got := LocalPhone(in); got != want {
			t.Errorf("%q → %q، المفروض %q", in, got, want)
		}
	}
}
