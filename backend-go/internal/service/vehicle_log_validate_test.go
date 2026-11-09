package service

import (
	"testing"

	"staffmange-api/internal/model"
)

// (11-07) سجل السيارة ما يقبل كلفة/لترات/عداد سالب — لا بالإضافة ولا بالتعديل.
func TestVehicleLogRejectsNegative(t *testing.T) {
	s := &VehicleService{} // الرفض قبل ما يوصل للقاعدة
	neg, negI := -5.0, -1
	for name, req := range map[string]model.CreateVehicleLogRequest{
		"cost":     {Type: "MAINTENANCE", Cost: &neg},
		"liters":   {Type: "MAINTENANCE", Liters: &neg},
		"odometer": {Type: "MAINTENANCE", Odometer: &negI},
	} {
		if _, err := s.CreateLog("v", req, "e"); err == nil {
			t.Errorf("إضافة %s سالب انقبلت", name)
		}
	}
	if _, err := s.UpdateLog("l", model.UpdateVehicleLogRequest{Cost: &neg}); err == nil {
		t.Error("تعديل كلفة سالبة انقبل")
	}
}
