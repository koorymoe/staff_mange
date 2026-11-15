package service

import "testing"

// (دازاي 11-15) تقرير المبرمج ينطابق ويا سجل تحديثات النظام بيومه.
func TestChangelogMatches(t *testing.T) {
	old := changelogMD
	t.Cleanup(func() { changelogMD = old })
	changelogMD = "- 11-15 الوحدة تكدر تسوي أكثر من خطة بنفس الشهر (زر «خطة ثانية»)\n- 11-10 شي قديم عن الفواتير والليدرية"
	if h := ChangelogMatches("اليوم عدلت شاشة الخطط: الوحدة تسوي خطة ثانية بنفس الشهر", "2026-11-15"); len(h) != 1 {
		t.Fatalf("لازم يطابق: %v", h)
	}
	if h := ChangelogMatches("اليوم عدلت شاشة الخطط: الوحدة تسوي خطة ثانية", "2026-11-12"); len(h) != 0 {
		t.Fatalf("يوم ثاني ما يطابق: %v", h)
	}
	if h := ChangelogMatches("صلحت الفواتير والليدرية", "2026-11-15"); len(h) != 0 {
		t.Fatalf("كلام ما موجود بيومه ما يطابق: %v", h)
	}
}
