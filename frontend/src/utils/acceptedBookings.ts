import type { Booking } from '../api'

// ═══ الحجز المستلَم بس ═══
//
// «مو يتوجّه لي حجز أني ضغطت استلام؟ يطلع هنا الحجز — ماريد كل
// الحجوزات تطلع».
//
// الفني ممكن يكون مكلّف بعشر حجوزات موزّعة على الأسبوع. عرضهن كلهن
// بشاشة الجرد ما ينفع بشيئين:
//   • الجرد يخص الشغل الي طالع له **هسه**، مو شغل بعد أربع أيام.
//   • وكثرة الخيارات تفتح باب يجرد للحجز الغلط.
//
// فالمعيار: الحجز الي **استلمه فعلاً** (ضغط «استلام») — يعني بديت
// شغله أو انطلقت له. الباقي ما يطلع.
export function acceptedBookings(all: Booking[]): Booking[] {
  return all
    .filter((b) => b.status !== 'COMPLETED' && b.status !== 'CANCELLED')
    .filter((b) => b.status === 'IN_PROGRESS' || !!b.startedAt || !!b.arrivedAt)
    .sort((x, y) => (x.scheduledAt || '9999').localeCompare(y.scheduledAt || '9999'))
}
