package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"staffmange-api/internal/model"
)

// ═══ 🎯 أوامر ماتركس على الحجوزات (دازاي 11-16) ═══
// «المفروض ماتركس يوجّه الكوادر ويثبّت الحجوزات وينسّق بيناتهن».
// المدير يكتب: «كلّف كادر لحجز B120»، «ثبّت B118 باچر الساعة عشرة»، «نسّق حجوزات باچر» —
// وماتركس يطلّع بطاقة بالضبط منو ينكلّف وشوكت وليش، وما يتنفّذ شي إلا بـ«نفّذ» (نفس أوامره الباقية).
//
// الاختيار من محرك الاقتراح الموجود (الخبرة بالخدمة، السرعة، التقييمات، منو فاضي ومنو مجاز) —
// والتنفيذ بنفس مسارات المنسّق (Assign / SetSupervisor / Confirm / SetSchedule). ماكو مسار ثاني.
// ⚠️ الأدوات ما ترجّع أسماء للنموذج — بس عدد الكادر وكود الحجز.

const (
	orderKindCrew     = "crew"
	orderKindSchedule = "schedule"
	orderAtLayout     = "2006-01-02T15:04" // بتوقيت بغداد
)

var crewRoles = []string{"TECH_1", "TECH_2", "TECH_3"}

func orderBookingKind(k string) bool { return k == orderKindCrew || k == orderKindSchedule }

// bookingByCode الحجز بكوده (B120) — بلا فرق أحرف.
func (s *MatrixOrderService) bookingByCode(code string) (string, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return "", errors.New("اكتب كود الحجز (مثل B120)")
	}
	var id string
	if err := s.db.Get(&id, `SELECT id FROM "Booking" WHERE upper(code) = $1 AND "archivedAt" IS NULL LIMIT 1`, code); err != nil {
		return "", errors.New("ما لكيت حجز بهالكود: " + code)
	}
	return id, nil
}

// proposeCrew يقترح كادر لحجز: الاقتراح الكامل من المحرك، أو موظف محدد إذا المدير سمّاه.
func (s *MatrixOrderService) proposeCrew(code, employeeID, employeeName string) (OrderAction, error) {
	if s.deps.Suggest == nil || s.deps.Bookings == nil {
		return OrderAction{}, errors.New("تكليف الكادر مو متاح")
	}
	id, err := s.bookingByCode(code)
	if err != nil {
		return OrderAction{}, err
	}
	b, crew, why, err := s.deps.Suggest.CrewFor(id)
	if err != nil {
		if b != nil && b.ScheduledAt == nil {
			return OrderAction{}, errors.New("الحجز " + b.Code + " بلا موعد — ثبّت موعده أول (order_schedule)")
		}
		return OrderAction{}, err
	}
	if b.Status == "COMPLETED" || b.Status == "CANCELLED" {
		return OrderAction{}, errors.New("الحجز " + b.Code + " منتهي")
	}
	a := OrderAction{Kind: orderKindCrew, BookingID: b.ID, BookingCode: b.Code, Techs: map[string]string{},
		At: b.ScheduledAt.In(baghdad).Format(orderAtLayout)}
	if employeeID != "" {
		// موظف سمّاه المدير: الليدر بخانة المشرف، والفني بأول خانة فاضية.
		var isLeader bool
		_ = s.db.Get(&isLeader, `SELECT "isLeader" FROM "Employee" WHERE id = $1`, employeeID)
		if isLeader {
			if b.SupervisorID != nil && *b.SupervisorID != employeeID {
				a.Why = append(a.Why, "⚠️ الحجز بيه ليدر — راح ينبدّل.")
			}
			a.LeaderID, a.LeaderName = employeeID, employeeName
		} else {
			taken := s.deps.Suggest.Repo().AssignmentRoles(b.ID)
			role := ""
			for _, r := range crewRoles {
				if taken[r] == employeeID {
					return OrderAction{}, errors.New("هذا الموظف مكلّف بالحجز أصلاً")
				}
				if role == "" && taken[r] == "" {
					role = r
				}
			}
			if role == "" {
				return OrderAction{}, errors.New("خانات الكادر بالحجز " + b.Code + " كلها مليانة")
			}
			a.Techs[role] = employeeID
			a.TechNames = []string{employeeName}
		}
		a.Why = append(a.Why, "بطلب المدير بالاسم.")
		return a, nil
	}
	if b.SupervisorID != nil || len(b.TechIDs) > 0 {
		return OrderAction{}, errors.New("الحجز " + b.Code + " بيه كادر أصلاً — إذا تريد تضيف واحد سمّيه")
	}
	if crew == nil || (crew.Leader == nil && !b.Solo) {
		return OrderAction{}, errors.New("ماكو ليدر فاضي بموعد الحجز " + b.Code + " — غيّر الموعد أو سمّي الكادر")
	}
	if len(crew.Techs) == 0 && b.Solo {
		return OrderAction{}, errors.New("ماكو فني فاضي بموعد الحجز " + b.Code)
	}
	if crew.Leader != nil {
		a.LeaderID, a.LeaderName = crew.Leader.ID, crew.Leader.Name
		if len(crew.Leader.Why) > 0 {
			a.Why = append(a.Why, "الليدر: "+strings.Join(crew.Leader.Why, "، ")+".")
		}
	}
	for i, t := range crew.Techs {
		if i >= len(crewRoles) {
			break
		}
		a.Techs[crewRoles[i]] = t.ID
		a.TechNames = append(a.TechNames, t.Name)
	}
	a.Why = append(a.Why, why...)
	return a, nil
}

// proposeSchedule يقترح موعد لحجز: الي گاله المدير، أو أقرب وقت فاضي من جدول الكوادر.
func (s *MatrixOrderService) proposeSchedule(code, at string) (OrderAction, error) {
	if s.deps.Suggest == nil || s.deps.Bookings == nil {
		return OrderAction{}, errors.New("تثبيت الموعد مو متاح")
	}
	id, err := s.bookingByCode(code)
	if err != nil {
		return OrderAction{}, err
	}
	b, err := s.deps.Suggest.Repo().Booking(id)
	if err != nil {
		return OrderAction{}, errors.New("الحجز مو موجود")
	}
	if b.Status == "COMPLETED" || b.Status == "CANCELLED" {
		return OrderAction{}, errors.New("الحجز " + b.Code + " منتهي")
	}
	a := OrderAction{Kind: orderKindSchedule, BookingID: b.ID, BookingCode: b.Code}
	if at = strings.TrimSpace(at); at != "" {
		t, err := time.ParseInLocation(orderAtLayout, at, baghdad)
		if err != nil {
			return OrderAction{}, errors.New("صيغة الموعد YYYY-MM-DDTHH:MM بتوقيت بغداد")
		}
		if t.Before(time.Now()) {
			return OrderAction{}, errors.New("الموعد لازم بالمستقبل")
		}
		a.At = t.Format(orderAtLayout)
		if !s.deps.Suggest.SlotFree(b.ID, t) {
			a.Why = append(a.Why, "⚠️ ماكو ليدر فاضي بهالوقت حسب الجدول (أو برّا الدوام المعتاد) — يتثبّت بطلبك، بس راجع الكادر.")
		}
	} else {
		slots := s.deps.Suggest.FreeSlots(b.ID, 1)
		if len(slots) == 0 {
			return OrderAction{}, errors.New("ماكو وقت فاضي خلال ١٠ أيام للحجز " + b.Code + " — حدد الموعد إنت")
		}
		a.At = slots[0].At.In(baghdad).Format(orderAtLayout)
		a.Why = append(a.Why, fmt.Sprintf("أقرب وقت بيه %d ليدر فاضي ضمن الدوام المعتاد.", slots[0].FreeLeaders))
	}
	if b.Status == "PENDING" {
		a.Why = append(a.Why, "الحجز بعده مو مثبّت — راح يتثبّت بهالموعد.")
	} else if b.ScheduledAt != nil {
		a.Why = append(a.Why, "الموعد الحالي "+b.ScheduledAt.In(baghdad).Format("2006-01-02 15:04")+" راح يتغيّر.")
	}
	return a, nil
}

// proposeCoordinate «نسّق حجوزات اليوم الفلاني»: كادر مقترح لكل حجز مثبّت بلا كادر بذاك اليوم.
// يرجّع الإجراءات وملخّص للنموذج (بلا أسماء).
func (s *MatrixOrderService) proposeCoordinate(day string) ([]OrderAction, string, error) {
	if s.deps.Actions == nil {
		return nil, "", errors.New("التنسيق مو متاح")
	}
	d, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(day), baghdad)
	if err != nil {
		return nil, "", errors.New("صيغة اليوم YYYY-MM-DD")
	}
	rows, err := s.deps.Actions.UnstaffedOn(d.Format("2006-01-02"))
	if err != nil {
		return nil, "", err
	}
	if len(rows) == 0 {
		return nil, "ماكو حجوزات مثبّتة بلا كادر بهاليوم.", nil
	}
	acts := []OrderAction{}
	var failed []string
	// كل حجز ينقترح وهو يشوف الي انقترح قبله بنفس الأمر — حتى نفس الشخص ما ينحط بحجزين متعارضين.
	busy := map[string][]time.Time{}
	for _, r := range rows {
		a, err := s.proposeCrew(r.Code, "", "")
		if err != nil {
			failed = append(failed, r.Code)
			continue
		}
		at, _ := time.ParseInLocation(orderAtLayout, a.At, baghdad)
		clash := false
		for _, id := range crewIDs(a) {
			for _, t := range busy[id] {
				if diff := at.Sub(t); diff > -4*time.Hour && diff < 4*time.Hour {
					clash = true
				}
			}
		}
		if clash {
			failed = append(failed, r.Code)
			continue
		}
		for _, id := range crewIDs(a) {
			busy[id] = append(busy[id], at)
		}
		acts = append(acts, a)
	}
	msg := fmt.Sprintf("%d حجز بلا كادر: اقترحت كادر لـ%d.", len(rows), len(acts))
	if len(failed) > 0 {
		msg += " ما لكيت كادر فاضي بلا تعارض لـ: " + strings.Join(failed, "، ") + " — يحتاجن المنسّق."
	}
	return acts, msg, nil
}

func crewIDs(a OrderAction) []string {
	ids := []string{}
	if a.LeaderID != "" {
		ids = append(ids, a.LeaderID)
	}
	for _, id := range a.Techs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// validateBooking الإجراء لازم بيه حجز، وكادر أو موعد.
func validateBooking(a OrderAction) bool {
	if a.BookingID == "" || a.BookingCode == "" {
		return false
	}
	if a.Kind == orderKindCrew {
		return a.LeaderID != "" || len(a.Techs) > 0
	}
	_, err := time.ParseInLocation(orderAtLayout, a.At, baghdad)
	return err == nil
}

func (s *MatrixOrderService) execBooking(actor string, a OrderAction) (string, string, error) {
	if s.deps.Bookings == nil || s.deps.Suggest == nil {
		return "", "", errors.New("أوامر الحجوزات مو متاحة")
	}
	link := "/bookings/coordination"
	if a.Kind == orderKindSchedule {
		b, err := s.deps.Suggest.Repo().Booking(a.BookingID)
		if err != nil {
			return "", "", errors.New("الحجز مو موجود")
		}
		t, err := time.ParseInLocation(orderAtLayout, a.At, baghdad)
		if err != nil || t.Before(time.Now()) {
			return "", "", errors.New("الموعد فات — اطلب الأمر من جديد")
		}
		when := t.Format("2006-01-02 15:04")
		if b.Status == "PENDING" {
			note := "🎯 ثبّته ماتركس بأمر الإدارة"
			if _, err := s.deps.Bookings.Confirm(a.BookingID, model.ConfirmBookingRequest{ConfirmedByEmployeeID: &actor, AdminNotes: &note, ScheduledAt: &a.At}); err != nil {
				return "", "", err
			}
			return "انثبّت الحجز " + a.BookingCode + " على " + when, link, nil
		}
		if _, err := s.deps.Bookings.SetSchedule(a.BookingID, actor, a.At); err != nil {
			return "", "", err
		}
		return "صار موعد الحجز " + a.BookingCode + " " + when, link, nil
	}
	// الكادر: نفس مسارات المنسّق، وإذا فشل بالنص يرجع الي انسوّى.
	done := []string{}
	undo := func() {
		for _, r := range done {
			_, _ = s.deps.Bookings.Unassign(a.BookingID, r)
		}
	}
	for _, role := range crewRoles {
		emp := a.Techs[role]
		if emp == "" {
			continue
		}
		// editorID فاضي: التكليف قرار ماتركس بموافقة المدير — فحص «عدالة التوزيع» ما يغرّم أحد عليه.
		if _, err := s.deps.Bookings.Assign(a.BookingID, model.AssignBookingRequest{EmployeeID: emp, Role: role}, ""); err != nil {
			undo()
			return "", "", err
		}
		done = append(done, role)
	}
	if a.LeaderID != "" {
		id := a.LeaderID
		if _, err := s.deps.Bookings.SetSupervisor(a.BookingID, &id); err != nil {
			undo()
			return "", "", err
		}
	}
	names := a.TechNames
	if a.LeaderName != "" {
		names = append([]string{"الليدر " + a.LeaderName}, names...)
	}
	if s.deps.Notif != nil {
		msg := fmt.Sprintf("🎯 انكلّفت بحجز %s (موعده %s) — من الإدارة عبر ماتركس. التفاصيل بـ«مهامي».", a.BookingCode, strings.Replace(a.At, "T", " ", 1))
		for _, id := range crewIDs(a) {
			_ = s.deps.Notif.Create(id, "MATRIX_ORDER", msg)
		}
	}
	return "انكلّف بحجز " + a.BookingCode + ": " + strings.Join(names, "، "), link, nil
}

// bookingOrderTools أدوات الحجوزات لفهم الأمر — تسجّل اقتراح بس.
func (s *MatrixOrderService) bookingOrderTools(add func(OrderAction, *nameMask) string,
	emp func(map[string]any, *nameMask) (string, string, error)) []agentTool {
	str := func(in map[string]any, k string) string { v, _ := in[k].(string); return strings.TrimSpace(v) }
	codeProp := map[string]any{"type": "string", "description": "كود الحجز مثل B120"}
	return []agentTool{{
		Name: "order_crew",
		Desc: "يقترح تكليف كادر لحجز مثبّت بموعد: بلا employee = أحسن كادر فاضي من محرك الاقتراح (ليدر + فنيين). ويا employee = هذا الموظف بالذات.",
		Props: map[string]any{"booking": codeProp,
			"employee": map[string]any{"type": "string", "description": "اختياري — رمز موظف#n إذا المدير سمّى واحد"}},
		Req: []string{"booking"},
		Run: func(in map[string]any, names *nameMask) (string, error) {
			id, name := "", ""
			if str(in, "employee") != "" {
				var err error
				if id, name, err = emp(in, names); err != nil {
					return "", err
				}
			}
			a, err := s.proposeCrew(str(in, "booking"), id, name)
			if err != nil {
				return "", err
			}
			add(a, names)
			return fmt.Sprintf("انسجل اقتراح تكليف كادر لحجز %s (%d شخص) — ينتظر موافقة المدير.", a.BookingCode, len(crewIDs(a))), nil
		},
	}, {
		Name:  "order_schedule",
		Desc:  "يقترح تثبيت أو تغيير موعد حجز. at بتوقيت بغداد YYYY-MM-DDTHH:MM إذا المدير حدد وقت؛ بلاه = أقرب وقت فاضي من جدول الكوادر. الحجز الي بعده مو مثبّت يتثبّت بيه.",
		Props: map[string]any{"booking": codeProp, "at": map[string]any{"type": "string"}},
		Req:   []string{"booking"},
		Run: func(in map[string]any, names *nameMask) (string, error) {
			a, err := s.proposeSchedule(str(in, "booking"), str(in, "at"))
			if err != nil {
				return "", err
			}
			add(a, names)
			return "انسجل اقتراح موعد " + strings.Replace(a.At, "T", " ", 1) + " لحجز " + a.BookingCode + " — ينتظر موافقة المدير.", nil
		},
	}, {
		Name:  "order_coordinate",
		Desc:  "«نسّق حجوزات اليوم الفلاني»: يقترح كادر لكل حجز مثبّت بلا كادر بذاك اليوم (بلا تعارض بيناتهن). day بصيغة YYYY-MM-DD.",
		Props: map[string]any{"day": map[string]any{"type": "string"}},
		Req:   []string{"day"},
		Run: func(in map[string]any, names *nameMask) (string, error) {
			acts, msg, err := s.proposeCoordinate(str(in, "day"))
			if err != nil {
				return "", err
			}
			for _, a := range acts {
				add(a, names)
			}
			return msg, nil
		},
	}}
}

// ── للسكرتير: يجهّز نفس البطاقة بلا نموذج ثاني، وينفّذها بعد موافقة المدير ──

// Describe الإجراء بجملة للعرض (بالأسماء — للمدير بس).
func (a OrderAction) Describe() string {
	at := strings.Replace(a.At, "T", " الساعة ", 1)
	if a.Kind == orderKindSchedule {
		return "موعد الحجز " + a.BookingCode + " ← " + at
	}
	names := a.TechNames
	if a.LeaderName != "" {
		names = append([]string{"الليدر " + a.LeaderName}, names...)
	}
	return "حجز " + a.BookingCode + " (" + at + "): " + strings.Join(names, "، ")
}

// CrewAction / ScheduleAction / CoordinateActions نفس اقتراحات الأوامر، للسكرتير.
func (s *MatrixOrderService) CrewAction(code, employeeID, employeeName string) (OrderAction, error) {
	return s.proposeCrew(code, employeeID, employeeName)
}
func (s *MatrixOrderService) ScheduleAction(code, at string) (OrderAction, error) {
	return s.proposeSchedule(code, at)
}
func (s *MatrixOrderService) CoordinateActions(day string) ([]OrderAction, string, error) {
	return s.proposeCoordinate(day)
}

// ProposeActions يخزن أمر بإجراءات جاهزة (بانتظار «نفّذ») — يطلع بقائمة أوامر ماتركس مثل غيره.
func (s *MatrixOrderService) ProposeActions(actor, text, reply string, acts []OrderAction) (*MatrixOrder, error) {
	acts = s.validate(acts)
	if len(acts) == 0 {
		return nil, errors.New("ماكو إجراء صالح للتنفيذ")
	}
	raw, _ := json.Marshal(acts)
	var id string
	if err := s.db.Get(&id, `INSERT INTO "MatrixOrder"(id, "actorId", text, status, actions, reply)
		VALUES (gen_random_uuid()::text, $1, $2, $3, $4, NULLIF($5, '')) RETURNING id`, actor, text, OrderProposed, raw, reply); err != nil {
		return nil, err
	}
	return s.get(id, actor)
}

// LatestProposed آخر أمر للمستخدم بعده ينتظر الموافقة.
func (s *MatrixOrderService) LatestProposed(actor string) (*MatrixOrder, error) {
	var id string
	if err := s.db.Get(&id, `SELECT id FROM "MatrixOrder" WHERE "actorId" = $1 AND status = $2 ORDER BY "createdAt" DESC LIMIT 1`, actor, OrderProposed); err != nil {
		return nil, ErrOrderNotFound
	}
	return s.get(id, actor)
}
