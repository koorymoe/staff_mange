package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"staffmange-api/internal/repository"
)

// ═══ «السكرتير» ينفّذ عن المدير: الكادر والمواعيد والتنسيق (دازاي 11-16) ═══
// «كمال يكون سكرتير ك البشر… ويقوم ببعض الاعمال المكلفة اليه».
// المدير يگول بصوته أو كتابة «كلّف كادر لحجز B120» / «نسّق حجوزات باچر»، والسكرتير يجهّز نفس
// أمر ماتركس (matrix_orders_booking.go)، يعرضه ببطاقة، وينفّذه بعد «اي» من المدير **برسالة بعدها** —
// الحراسة بالكود: أمر انجهّز بهذا الدور ما يتنفّذ بنفس الدور.

func registerOrderTools(a *MatrixAgent, orders *MatrixOrderService, repo *repository.SecretaryRepository, tc *SecretaryTurnCtx) {
	if orders == nil || tc == nil {
		return
	}
	link := "/?board=matrix" // مكتب المدير ← ماتركس ← 🎯 أوامر ماتركس
	propose := func(acts []OrderAction, note string) (string, error) {
		o, err := orders.ProposeActions(tc.UserID, tc.Request, note, acts)
		if err != nil {
			return "", err
		}
		if tc.newOrders == nil {
			tc.newOrders = map[string]bool{}
		}
		tc.newOrders[o.ID] = true
		lines := make([]string, 0, len(o.Actions)+1)
		for _, x := range o.Actions {
			lines = append(lines, x.Describe())
			for _, w := range x.Why {
				lines = append(lines, "  · "+w)
			}
		}
		tc.addCard(SecretaryCard{Kind: "task", Title: "🎯 أمر ينتظر موافقتك", Status: "ما انفّذ بعد", Lines: lines, Link: strPtr(link)})
		out := fmt.Sprintf("انجهّز الأمر (orderId=%s) وما انفّذ بعد:\n- %s", o.ID, strings.Join(lines, "\n- "))
		if note != "" {
			out += "\n" + note
		}
		return out + "\nگوله الخلاصة بجملة واسأله «أنفّذ؟» — والتنفيذ بـrun_order بس بعد موافقته برسالة جديدة.", nil
	}
	codeProp := map[string]any{"type": "string", "description": "كود الحجز مثل B120"}

	a.AddTool(agentTool{Name: "assign_crew",
		Desc:  "جهّز تكليف كادر لحجز مثبّت بموعد. بلا employee = أحسن كادر فاضي (ليدر + فنيين) من محرك الاقتراح. ويا employee = هذا الموظف بالذات. ما ينفّذ — يطلّع بطاقة تنتظر موافقته.",
		Props: map[string]any{"booking": codeProp, "employee": map[string]any{"type": "string", "description": "اختياري — اسم الموظف كما گاله"}},
		Req:   []string{"booking"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			id, name := "", ""
			if n := strings.TrimSpace(strArg(in, "employee")); n != "" {
				emps, err := repo.ActiveEmployees()
				if err != nil {
					return "", err
				}
				list := make([]VoiceTaskEmployee, 0, len(emps))
				names := map[string]string{}
				for _, e := range emps {
					list = append(list, VoiceTaskEmployee{ID: e.ID, Name: e.Name})
					names[e.ID] = e.Name
				}
				ids := matchEmployeeBest(n, list)
				if len(ids) != 1 {
					return "ما لگيت موظف واحد واضح بهالاسم — اسأله منو يقصد.", nil
				}
				id, name = ids[0], names[ids[0]]
			}
			act, err := orders.CrewAction(strArg(in, "booking"), id, name)
			if err != nil {
				return "", err
			}
			return propose([]OrderAction{act}, "")
		}})

	a.AddTool(agentTool{Name: "schedule_booking",
		Desc:  "جهّز تثبيت أو تغيير موعد حجز. at بتوقيت بغداد YYYY-MM-DDTHH:MM إذا حدد وقت (الصبح=10:00، الظهر=13:00)؛ بلاه = أقرب وقت فاضي من جدول الكوادر. الحجز الي بعده مو مثبّت يتثبّت بيه. ما ينفّذ — بطاقة تنتظر موافقته.",
		Props: map[string]any{"booking": codeProp, "at": map[string]any{"type": "string"}},
		Req:   []string{"booking"},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			act, err := orders.ScheduleAction(strArg(in, "booking"), strArg(in, "at"))
			if err != nil {
				return "", err
			}
			return propose([]OrderAction{act}, "")
		}})

	a.AddTool(agentTool{Name: "coordinate_day",
		Desc:  "«نسّق حجوزات باچر/اليوم»: جهّز كادر لكل حجز مثبّت بلا كادر بذاك اليوم، بلا تعارض بيناتهن. day بصيغة YYYY-MM-DD (فاضي = باچر). ما ينفّذ — بطاقة تنتظر موافقته.",
		Props: map[string]any{"day": map[string]any{"type": "string"}},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			day := strings.TrimSpace(strArg(in, "day"))
			if day == "" {
				day = time.Now().In(debriefLoc).AddDate(0, 0, 1).Format("2006-01-02")
			}
			acts, msg, err := orders.CoordinateActions(day)
			if err != nil {
				return "", err
			}
			if len(acts) == 0 {
				return msg, nil
			}
			return propose(acts, msg)
		}})

	a.AddTool(agentTool{Name: "run_order",
		Desc:  "نفّذ أمر انجهّز قبل (assign_crew / schedule_booking / coordinate_day) — بس بعد ما المدير وافق صراحة برسالته هاي («اي»، «نفّذ»، «تمام سويها»). orderId من نتيجة التجهيز؛ فاضي = آخر أمر ينتظر.",
		Props: map[string]any{"orderId": map[string]any{"type": "string"}},
		Run: func(in map[string]any, _ *nameMask) (string, error) {
			id := strings.TrimSpace(strArg(in, "orderId"))
			if id == "" {
				o, err := orders.LatestProposed(tc.UserID)
				if err != nil {
					return "ماكو أمر ينتظر الموافقة.", nil
				}
				id = o.ID
			}
			if tc.newOrders[id] {
				return "", errors.New("هذا الأمر انجهّز تواً — اعرضه عليه واستنى موافقته برسالة جديدة قبل التنفيذ")
			}
			o, err := orders.Run(id, tc.UserID, "", nil)
			if err != nil {
				return "ما لگيت أمر ينتظر التنفيذ بهالرقم (يمكن انفّذ أو انلغى).", nil
			}
			lines := []string{}
			for _, x := range o.Actions {
				switch {
				case x.Done:
					lines = append(lines, "✅ "+x.Result)
				case x.Error != "":
					lines = append(lines, "⚠️ "+x.Describe()+" — ما انفّذ: "+x.Error)
				}
			}
			tc.addCard(SecretaryCard{Kind: "task", Title: "🎯 نتيجة التنفيذ", Status: o.Status, Lines: lines, Link: strPtr(link)})
			return strings.Join(lines, "\n"), nil
		}})
}
