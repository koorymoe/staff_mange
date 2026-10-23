package service

import (
	"reflect"
	"testing"
)

func TestMatchEmployees(t *testing.T) {
	emps := []VoiceTaskEmployee{
		{ID: "1", Name: "علي حسن"},
		{ID: "2", Name: "مصطفى كريم"},
		{ID: "3", Name: "علاء جبار"},
	}
	cases := map[string][]string{
		"خلي علي حسن يروح للمخزن":    {"1"},
		"گول لمصطفى يجيب الفواتير":   {"2"},
		"ومصطفى كريم ويا علي حسن":    {"1", "2"},
		"روح جيب الكيبلات من المخزن": {},
		"العلاء ما بيه شي":           {}, // «علاء» داخل كلمة ثانية ما تنطابق
	}
	for text, want := range cases {
		if got := matchEmployees(text, emps); !reflect.DeepEqual(got, want) {
			t.Errorf("%q: got %v want %v", text, got, want)
		}
	}
}

func TestParseVoiceTask(t *testing.T) {
	u, err := parseVoiceTask("تفضل:\n{\"title\":\"جرد المخزن\",\"description\":\"جرد الكيبلات\",\"priority\":\"HIGH\",\"dueAt\":\"باچر\"}")
	if err != nil {
		t.Fatal(err)
	}
	if u.Title != "جرد المخزن" || u.Priority != "NORMAL" || u.DueAt != "" {
		t.Fatalf("تنظيف غلط: %+v", u)
	}
	u, err = parseVoiceTask(`{"title":"x","priority":"URGENT","dueAt":"2026-10-20T15:00"}`)
	if err != nil || u.Priority != "URGENT" || u.DueAt != "2026-10-20T15:00" {
		t.Fatalf("got %+v %v", u, err)
	}
	if _, err := parseVoiceTask(`{"title":"  "}`); err == nil {
		t.Fatal("عنوان فاضي لازم يرفض")
	}
}
