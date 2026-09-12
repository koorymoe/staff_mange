import type { Booking } from '../api'

// ═══ هوية الحجز والزبون — مصدر واحد ═══
//
// طلب صاحب العمل: أي فاتورة أو حجز أو تقرير لازم يقول بالضبط لمنو
// ولأي حجز. قبل، كل شاشة تجمع هاي المعلومات بطريقتها — وثنتين منهن
// (الحجوزات والتنسيق) جانن يعرضن كود الزبون **فاضي** لأنهن يقرن
// حقل `code` والسيرفر ما يرسله مع الحجز (يرسل `customerCode` الرقم).
//
// ⚠️ ليش التنسيق بالواجهة مو بالسيرفر؟
// `customerCode` عمود حقيقي وينرسل بكل رد أصلاً، أما `code` المنسّق
// فينتبني بـCustomerResponse الي ما تنستعمل إلا بشاشة الزبائن. لو
// رحنا نعبّي `code` بالسيرفر چان لازم نلمس ١٢ موضع هيدرشن — وأي
// موضع ننساه يرجّع نفس الفراغ. هنا دالة وحدة تغطي الكل.

/** CUST-00042 — نفس صيغة السيرفر بالضبط (FormatCode بـmodel/customer.go) */
export function formatCustomerCode(customer?: { code?: string; customerCode?: number } | null): string {
  if (!customer) return ''
  // لو السيرفر رسل الكود جاهز (شاشة الزبائن) نستعمله مثل ما هو
  if (customer.code) return customer.code
  if (typeof customer.customerCode === 'number' && customer.customerCode > 0) {
    return `CUST-${String(customer.customerCode).padStart(5, '0')}`
  }
  return ''
}

/** ليدر الحجز: المشرف المعيّن، وإلا أول موظف مكلّف مؤشّر «تيم ليدر».
 *
 *  ⚠️ يرجّع null لما ماكو ليدر — **ما نرجّع أول فني** حتى ما تنسب
 *  المسؤولية لواحد ما تحمّلها. الشاشة تعرض «بلا ليدر» صراحةً. */
export function bookingLeaderName(booking?: Partial<Booking> | null): string {
  if (!booking) return ''
  if (booking.projectSupervisor?.name) return booking.projectSupervisor.name
  const fromCrew = (booking.assignments || []).find((a) => a.employee?.isLeader)
  return fromCrew?.employee?.name || ''
}

export type IdentityFields = {
  bookingCode?: string
  customerCode?: string
  customerName?: string
  customerPhone?: string
  address?: string
  leaderName?: string
  serviceName?: string
  scheduledAt?: string | null
}

/** يستخرج كل حقول الهوية من حجز — نقطة الجمع الوحيدة. */
export function identityOf(booking?: Partial<Booking> | null): IdentityFields {
  if (!booking) return {}
  const services = booking.services?.length ? booking.services : booking.service ? [booking.service] : []
  return {
    bookingCode: booking.code,
    customerCode: formatCustomerCode(booking.customer),
    customerName: booking.customer?.name,
    customerPhone: booking.customer?.phone,
    // عنوان الحجز أدق من عنوان الزبون: الزبون ممكن يطلب شغل بمحل ثاني
    address: booking.address || booking.customer?.location || undefined,
    leaderName: bookingLeaderName(booking),
    serviceName: services.map((s) => s.name).join('، ') || undefined,
    scheduledAt: booking.scheduledAt,
  }
}

/** الموعد بتوقيت بغداد بصيغة قصيرة، أو فاضي لو ماكو موعد. */
export function formatWhen(iso?: string | null): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}
