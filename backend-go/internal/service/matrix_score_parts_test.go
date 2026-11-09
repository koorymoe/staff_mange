package service

import "testing"

// ratioPart: نسبة الجزء، وماكو جزء إذا ماكو أساس (max=0)، والتفصيل بالأيام (النقاط ÷ ٢).
func TestRatioPart(t *testing.T) {
	if ratioPart("A", "x", 0, 0, "يوم") != nil {
		t.Fatal("max=0 لازم يرجع nil حتى ما ينحسب صفر ظلم")
	}
	p := ratioPart("A", "x", 6, 8, "يوم")
	if p.Pct != 75 || p.Detail != "3 من 4 يوم" {
		t.Fatalf("%+v", p)
	}
}

// periodicStage: الإدارة تقيّم بمرحلة الإدارة، والباقي بمرحلة المراقب — ADMIN/OWNER ما يتقيّمون، بس يقيّمون.
func TestPeriodicStage(t *testing.T) {
	for role, want := range map[string]string{"ADMIN": "ADMIN_PERIODIC", "OWNER": "ADMIN_PERIODIC", "MONITOR": "MONITOR_PERIODIC", "EMPLOYEE": "MONITOR_PERIODIC"} {
		if got := periodicStage(role); got != want {
			t.Errorf("%s: %s، المتوقع %s", role, got, want)
		}
	}
}
