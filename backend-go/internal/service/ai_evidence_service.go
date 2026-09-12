package service

import (
	"encoding/json"
	"fmt"
	"time"

	"staffmange-api/internal/model"
	"staffmange-api/internal/repository"
)

// ═══ جامع الأدلة ═══
//
// هذا الي وصفه صاحب العمل حرفياً:
//
//   «مو يجي الموظف يسوي توقف العمل… النظام يضل يفكر بالموضوع: يراجع
//    الساعة — إحنا دوامنا ينتهي ١٢ ليلاً ويبدي ٩ صباحاً. أو مادة
//    ناقصة — ليش ناقصة؟ الزبون ما جان طلبها وطلبها، لو إنت ناسيها،
//    لو إنت طلبتها وأبو الكميات ما وفّرها؟ يروح يشوف سلة الزبون…
//    نوب يروح لأبو الكميات… ترجع تشوف شوكت انضافت المادة».
//
// ⚠️ هذا الملف **ما يحكم على أحد**. شغله ينتزع الحقائق بس:
//   طلب مادة؟ متى؟ انوفّرت؟ السلة زادت بعد ما بدأ؟ الدوام خالص؟
// الحكم بملف ثاني، والسبب بالتفصيل برأس schema_ai_core.go.
//
// وليش نجمعها بأنفسنا بدل ما نرمي كلشي للنموذج؟ لأن النموذج بلا
// حقائق يخمّن بثقة — والتخمين هنا يتحول لظلم موظف. الحقائق تنعرض
// للمالك حتى لو النموذج غلط بتفسيرها.

type AiEvidenceService struct {
	db       *repository.AiRepository
	bookings *repository.BookingRepository
}

func NewAiEvidenceService(db *repository.AiRepository, bookings *repository.BookingRepository) *AiEvidenceService {
	return &AiEvidenceService{db: db, bookings: bookings}
}

// CollectForWorkStop يجمع أدلة توقف العمل — المسار الي وصفه صاحب العمل.
//
// ⚠️ كل خطوة ممكن تفشل بلا ما توقف الباقي: نسجّل الفجوة بـgaps ونكمل.
// أدلة ناقصة معلنة أحسن من ولا أدلة — والنموذج لازم يعرف شنو ما شافه
// بدل ما يفترض إنه ماكو.
func (s *AiEvidenceService) CollectForWorkStop(signal model.AiSignal) (*model.AiEvidence, error) {
	ev := model.WorkStopEvidence{}
	gaps := []string{}

	booking, err := s.bookings.FindByID(signal.EntityID)
	if err != nil || booking == nil {
		return nil, fmt.Errorf("الحجز مو موجود")
	}

	if booking.WorkStopReason != nil {
		ev.StopReason = *booking.WorkStopReason
	} else {
		gaps = append(gaps, "ما كتب سبب التوقف")
	}

	// ═══ الساعة ═══
	// «الوقت لا يكفي» تنفحص مقابل نهاية الدوام الحقيقية، مو تنقبل
	// مثل ما هي.
	win, err := s.db.WorkWindow()
	if err != nil {
		gaps = append(gaps, "ما قدرنا نقرا ساعات الدوام")
		win = &model.AiWorkWindow{StartHour: 9, EndHour: 24}
	}
	stoppedAt := signal.OccurredAt
	if booking.WorkStoppedAt != nil {
		stoppedAt = *booking.WorkStoppedAt
	}
	// بغداد = UTC+3. الحساب بالتوقيت المحلي وإلا «الساعة ١١ ليلاً»
	// تنقرا ٨ مساءً وتنقلب النتيجة.
	baghdad := stoppedAt.UTC().Add(3 * time.Hour)
	ev.StoppedAtHour = baghdad.Hour()
	ev.MinutesToShiftEnd = (win.EndHour * 60) - (baghdad.Hour()*60 + baghdad.Minute())
	if ev.MinutesToShiftEnd < 0 {
		ev.MinutesToShiftEnd = 0
	}

	if booking.StartedAt != nil {
		ev.WorkedMinutes = int(stoppedAt.Sub(*booking.StartedAt).Minutes())
		if ev.WorkedMinutes < 0 {
			ev.WorkedMinutes = 0
		}
	} else {
		gaps = append(gaps, "ماكو وقت بداية شغل مسجّل")
	}

	// ═══ خيط المواد: هل طلب من إداري الكميات؟ ═══
	reqs, status, err := s.db.ProcurementSummary(booking.ID, stoppedAt)
	if err != nil {
		gaps = append(gaps, "ما قدرنا نقرا طلبات المواد")
	} else {
		ev.ProcurementRequests = reqs.Total
		ev.RequestedBeforeStop = reqs.BeforeStop > 0
		ev.LastRequestStatus = status
	}

	// ═══ خيط السلة: الزبون طلب زيادة بالموقع؟ ═══
	total, afterStart, err := s.db.CartSummary(booking.ID, booking.StartedAt)
	if err != nil {
		gaps = append(gaps, "ما قدرنا نقرا سلة الزبون")
	} else {
		ev.CartItemsTotal = total
		ev.CartItemsAfterStart = afterStart
	}

	// ═══ سجل الموظف: مرة ولا نمط؟ ═══
	if signal.EmployeeID != nil {
		n, err := s.db.StopCountForEmployee(*signal.EmployeeID, 30)
		if err != nil {
			gaps = append(gaps, "ما قدرنا نقرا سجل توقفات الموظف")
		} else {
			ev.StopsLast30Days = n
		}
	} else {
		gaps = append(gaps, "الإشارة بلا موظف")
	}

	facts, _ := json.Marshal(ev)
	gapsJSON, _ := json.Marshal(gaps)
	return s.db.SaveEvidence(signal.ID, facts, gapsJSON)
}
