package service

import "testing"

// يحرس شرائح تسعير الكاميرات الي تطلع بعرض الزبون: أي تغيير بالأرقام لازم يكون مقصود.
func TestCameraCableTiers(t *testing.T) {
	if cableTierPrice(25, false) != 0 {
		t.Error("بلا إدخال الكيبل = صفر")
	}
	for m, want := range map[float64]float64{0: 7500, 5: 10000, 9.9: 10000, 10: 12000, 19: 12000, 20: 13000, 30: 14000, 40: 15000, 49.5: 15000, 50: 17000, 200: 17000} {
		if got := cableTierPrice(m, true); got != want {
			t.Errorf("%v متر: %v، المتوقع %v", m, got, want)
		}
	}
}

func TestCameraMultipliers(t *testing.T) {
	for place, want := range map[string]float64{"منزل سكني": 1, "محل تجاري": 0.95, "مدرسة او شركة": 1.3, "مصنع او معمل": 1.4, "غير معروف": 1} {
		if got := cameraPlaceMultiplier(place); got != want {
			t.Errorf("%s: %v", place, got)
		}
	}
	if cameraSystemMultiplier("IP") != 1.2 || cameraSystemMultiplier("ANLOGE") != 1 {
		t.Error("معامل النظام")
	}
	if cameraHeightMultiplier(true) != 1.1 || cameraHeightMultiplier(false) != 1 {
		t.Error("معامل الارتفاع")
	}
}
