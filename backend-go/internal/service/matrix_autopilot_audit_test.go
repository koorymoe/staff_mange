package service

import (
	"testing"
	"time"

	"staffmange-api/internal/model"
)

// (دازاي 11-15) الصياغة الجديدة ما تضيف حقائق.
func TestVoiceSafe_RejectsAddedFacts(t *testing.T) {
	orig := "🤖 ماتركس — حجز B1234 اليوم متوقع ياخذ ~3 ساعات. موظف#2 وياك."
	cases := []struct {
		name string
		out  string
		want bool
	}{
		{"rephrase ok", "🤖 انتبه حبي، حجز B1234 اليوم ممكن ياخذ ~3 ساعات، وموظف#2 وياك.", true},
		{"missing fact", "انتبه، حجز B1234 اليوم ممكن يطوّل، وموظف#2 وياك.", false},
		{"added digits", "انتبه، حجز B1234 اليوم ياخذ ~3 ساعات، وموظف#2 وياك، والزبون ينتظر من 45 دقيقة.", false},
		{"added booking code", "حجز B1234 وB9999 اليوم ياخذ ~3 ساعات، موظف#2 وياك.", false},
		{"added project code", "حجز B1234 (مشروع PRJ-0042) ياخذ ~3 ساعات، موظف#2 وياك.", false},
		{"added masked tag", "حجز B1234 ياخذ ~3 ساعات، موظف#2 وموظف#7 وياك.", false},
		{"too long", "حجز B1234 اليوم متوقع ياخذ ~3 ساعات وموظف#2 وياك، فانتبه للوقت زين وخلّي بالك من كل شي بالموقع وتابع الشغل خطوة خطوة ويا الفريق كله لحد ما تخلص.", false},
		{"banned word", "حجز B1234 ياخذ ~3 ساعات، موظف#2 وياك، وإلا غرامة.", false},
		{"empty", "  ", false},
	}
	for _, c := range cases {
		if got := voiceSafe(orig, c.out); got != c.want {
			t.Errorf("%s: voiceSafe=%v want %v", c.name, got, c.want)
		}
	}
}

// (دازاي 11-15) التذكيرين الي چانن ميتين إلهن دورة عصر.
func TestAutopilotAfternoonPass_WiresDeadReminders(t *testing.T) {
	at := func(h int) time.Time { return time.Date(2026, 11, 15, h, 5, 0, 0, debriefLoc) }
	if am, pm := autopilotPassesDue(at(9)); am || pm {
		t.Fatalf("9am: want no pass, got am=%v pm=%v", am, pm)
	}
	if am, pm := autopilotPassesDue(at(10)); !am || pm {
		t.Fatalf("10am: want morning only, got am=%v pm=%v", am, pm)
	}
	if am, pm := autopilotPassesDue(at(14)); !am || !pm {
		t.Fatalf("2pm: want both passes, got am=%v pm=%v", am, pm)
	}
	// شروط الساعة لازم تتحقق بوقت دورة العصر، وإلا يرجعن ميتات.
	if autopilotPMHour < 14 {
		t.Fatalf("PM pass at %d can't satisfy projectIdle (>=13) and vehiclesUnrated (>=14)", autopilotPMHour)
	}
	s := &MatrixAutopilotService{}
	if n := len(s.afternoonSteps(at(14))); n != 2 {
		t.Fatalf("afternoon steps = %d, want 2 (projectIdleToday + vehiclesUnrated)", n)
	}
	// قبل الساعة: الخطوتين ما يلمسن القاعدة ويرجعن صفر.
	if n, err := s.vehiclesUnrated("2026-11-15", at(0), at(10)); n != 0 || err != nil {
		t.Fatalf("vehiclesUnrated at 10am should no-op, got %d %v", n, err)
	}
	if n, err := s.projectIdleToday("2026-11-15", at(0), at(10)); n != 0 || err != nil {
		t.Fatalf("projectIdleToday at 10am should no-op, got %d %v", n, err)
	}
}

// (دازاي 11-15) الحد لكل مستلم مو عام، والعاجل يعبر.
func TestCapDecision_PerRecipient(t *testing.T) {
	emp := "e1"
	normal := model.AiAction{Kind: model.AiActionExtraTaskOverdue, TargetEmployeeID: &emp}
	if ok, _ := capDecision(normal, autopilotPerRecipientCap-1, 999); !ok {
		t.Fatal("7th reminder to recipient must pass even when kind total is huge")
	}
	if ok, lim := capDecision(normal, autopilotPerRecipientCap, 0); ok || lim != 8 {
		t.Fatalf("9th reminder to same recipient must drop (lim=%d)", lim)
	}
	urgent := model.AiAction{Kind: model.AiActionUnstaffedAlert, TargetEmployeeID: &emp}
	if ok, _ := capDecision(urgent, 100, 100); !ok {
		t.Fatal("urgent kinds bypass the cap")
	}
	group := model.AiAction{Kind: model.AiActionLowStock}
	if ok, _ := capDecision(group, 0, autopilotGroupKindCap-1); !ok {
		t.Fatal("group broadcast under cap must pass")
	}
	if ok, _ := capDecision(group, 0, autopilotGroupKindCap); ok {
		t.Fatal("group broadcast over cap must drop")
	}
}
