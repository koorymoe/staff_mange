package service

import (
	"strings"
	"testing"
)

// الأسماء ما تطلع لهايكو: الاسم الكامل والثنائي والأول (٤ أحرف فأكثر).
func TestRedactNames(t *testing.T) {
	in := "اليوم طلعت ويا علي حسن كاظم لبيت ابو مصطفى، ومصطفى كريم جان مرتاح، وعلي ساعدني"
	out := redactNames(in, []string{"علي حسن كاظم", "مصطفى كريم"}, "محمد جاسم")
	for _, bad := range []string{"علي حسن", "مصطفى كريم"} {
		if strings.Contains(out, bad) {
			t.Fatalf("الاسم %q بعده بالنص: %s", bad, out)
		}
	}
	if !strings.Contains(out, "[موظف]") {
		t.Fatalf("ما انبدّل: %s", out)
	}
	// «علي» ٣ أحرف — ما ينشال لحاله (كلمة قصيرة ممكن تكون عادية).
	if !strings.Contains(out, "وعلي ساعدني") {
		t.Fatalf("انشالت كلمة قصيرة: %s", out)
	}
	// «مصطفى» لحاله ينشال (٥ أحرف) حتى ببداية «ابو مصطفى».
	if strings.Contains(out, "مصطفى") {
		t.Fatalf("مصطفى بعده: %s", out)
	}
}
