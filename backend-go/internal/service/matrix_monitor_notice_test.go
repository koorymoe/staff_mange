package service

import (
	"testing"

	"staffmange-api/internal/model"
)

// (ك) 11-08: بلاغ #22 — تنبيه «بلّغ المراقب» للعلم، وتذكير الموظف نفسه مو للعلم.
func TestIsMonitorNotice(t *testing.T) {
	if !isMonitorNotice(model.AiAction{Kind: model.AiActionShortageOpen, EntityType: "MONITOR"}) {
		t.Error("نقص الجرد المبلّغ للمراقب لازم يكون للعلم")
	}
	if isMonitorNotice(model.AiAction{Kind: model.AiActionShortageOpen, EntityType: "EMPLOYEE"}) {
		t.Error("تذكير أبو الكميات نفسه لازم يبقى ينحسب")
	}
}
