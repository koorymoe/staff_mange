package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"staffmange-api/internal/repository"
	"staffmange-api/internal/safeguard"
)

// ═══ تذكير معاودة الاتصال بالزبون الي ما رد ═══
//
// حالة «في الانتظار» تعدّ المحاولات، بس ماكو شي يرجع ينبّه الإداري
// «اتصل عليه مرة ثانية». فالحجز يقعد بالطابور بلا حركة — مو ملغي ومو
// شغّال — وبعد شهر أحد يسأل «شنو صار بحجز فلان؟».
//
// الحدود الثلاثة تحت هي أهم شي بهذا الملف. تذكير كل ساعة ينتحوّل
// لضجيج ينتجاهل، وتذكير ينتجاهل أسوأ من ماكو تذكير أصلاً.
const (
	// WaitingReminderMinAgeHours ما نذكّر بنفس الشفت الي أشّر بيه
	// الإداري — توّه اتصل ويعرف.
	WaitingReminderMinAgeHours = 4
	// WaitingReminderGapHours تذكير واحد بالكثير لكل حجز باليوم.
	WaitingReminderGapHours = 24
	// WaitingReminderMax بعد خمس تذكيرات نوقف: زبون ما رد خمس مرات
	// هذا **قرار** لازم ياخذه الإداري (يلغي أو يتابع بطريقة ثانية)،
	// مو النظام يضل ينقّ عليه للأبد.
	WaitingReminderMax = 5
)

type BookingReminderService struct {
	bookings      *repository.BookingRepository
	notifications *repository.NotificationRepository
}

func NewBookingReminderService(
	bookings *repository.BookingRepository,
	notifications *repository.NotificationRepository,
) *BookingReminderService {
	return &BookingReminderService{bookings: bookings, notifications: notifications}
}

// RunWaitingReminderSweep يدزّ **إشعار واحد مجمّع** لكل الحجوزات
// المستحقة، مو إشعار لكل حجز.
//
// عشرة إشعارات بنفس الدقيقة تنمسح كلها بضغطة، فتضيع كلها سوه.
// إشعار واحد يقول «عدك ٣ حجوزات» ينقرا.
func (s *BookingReminderService) RunWaitingReminderSweep() (int, error) {
	rows, err := s.bookings.ListWaitingDueForReminder(
		WaitingReminderMinAgeHours, WaitingReminderGapHours, WaitingReminderMax)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}

	lines := make([]string, 0, len(rows))
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		phone := ""
		if r.CustomerPhone != nil {
			phone = " — " + *r.CustomerPhone
		}
		// عدد المحاولات جوّا النص: الإداري لازم يعرف إنها الخامسة مو
		// الأولى، لأن القرار يختلف.
		lines = append(lines, fmt.Sprintf("• %s: %s%s (محاولة %d)",
			r.Code, r.CustomerName, phone, r.ContactAttempts))
		ids = append(ids, r.ID)
	}

	msg := fmt.Sprintf("📞 عدك %d حجز بالانتظار يحتاج معاودة اتصال:\n%s",
		len(rows), strings.Join(lines, "\n"))
	if err := s.notifications.CreateForRole("HR_COORDINATOR", "waiting_callback_reminder", msg); err != nil {
		return 0, err
	}

	// ⚠️ التأشير بعد نجاح الإشعار مو قبله: لو انعكس الترتيب وفشل
	// الإشعار، الحجوزات تنحسب «انذكّرت» وما ينبّه عليها أحد ٢٤ ساعة.
	if err := s.bookings.MarkWaitingReminded(ids); err != nil {
		return len(rows), err
	}
	return len(rows), nil
}

// StartBackgroundSweeps نفس نمط كنسة الانضباط: goroutine وحدة تنام
// شوي بالبداية (حتى ما تزاحم إقلاع السيرفر) وبعدها كل ساعة.
//
// الفاصل ساعة والحدود فوق هي الي تحدد الإزعاج فعلاً — الكنسة تلگه
// المستحق بس.
// ⚠️ لازم تمر بـsafeguard.Loop مو goroutine عارية: panic هنا كان يقتل
// السيرفر كله ويخلي كل الشاشات تطلع «Failed to fetch».
func (s *BookingReminderService) StartBackgroundSweeps() {
	safeguard.Loop("كنسة تذكير الحجوزات", 3*time.Minute, time.Hour, func() {
		if n, err := s.RunWaitingReminderSweep(); err != nil {
			log.Printf("[reminder] كنسة تذكير المعاودة فشلت: %v", err)
		} else if n > 0 {
			log.Printf("[reminder] انذكّر الإداري بـ%d حجز بالانتظار", n)
		}
	})
}
