package model

import (
	"reflect"
	"strings"
	"testing"

	"staffmange-api/internal/models"
)

// (11-15) «عمود بالجدول بلا حقل = SELECT * يفشل كله بالسكوت»: صار فعلاً — عمود seenAt انضاف
// لـBookingAssignment بلا حقل، فكل الحجوزات طلعت بلا كادر وقائمة فواتير الليدر وقعت.
// هالاختبار يتأكد إن كل عمود بالجداول الي تنقرا بـSELECT * إله حقل db بالـstruct.
func TestSelectStarModelsCoverAllColumns(t *testing.T) {
	cases := map[string]any{
		"Employee": Employee{}, "Booking": Booking{}, "Customer": Customer{}, "BookingAssignment": BookingAssignment{},
		"LeaderInvoice": LeaderInvoice{}, "Quotation": Quotation{}, "ExtraTask": ExtraTask{}, "QualityFollowUp": QualityFollowUp{},
		"Complaint": Complaint{}, "Service": Service{}, "Vehicle": Vehicle{}, "AiAction": AiAction{},
	}
	for table, v := range cases {
		tbl, ok := models.Get(table)
		if !ok {
			t.Errorf("%s: مو بالسجل", table)
			continue
		}
		tags := map[string]bool{}
		collectDBTags(reflect.TypeOf(v), tags)
		for _, c := range tbl.Columns {
			if !tags[c.Name] {
				t.Errorf("%s.%s: عمود بلا حقل db بالـstruct — SELECT * راح يفشل", table, c.Name)
			}
		}
	}
}

func collectDBTags(tp reflect.Type, out map[string]bool) {
	for i := 0; i < tp.NumField(); i++ {
		f := tp.Field(i)
		tag := strings.Split(f.Tag.Get("db"), ",")[0]
		if f.Anonymous && f.Type.Kind() == reflect.Struct && tag == "" {
			collectDBTags(f.Type, out)
			continue
		}
		if tag != "" && tag != "-" {
			out[tag] = true
		}
	}
}
