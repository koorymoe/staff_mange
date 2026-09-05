import { useEffect, useState } from 'react'
import MultiSelect from './MultiSelect'
import { onEnter } from '../utils/enterKey'
import { api, type Booking, type Service } from '../api'
import { useSession } from '../session'

// ═══ تعديل الحجز بمكان واحد ═══
//
// التعديل جان موزّع: السعر والعنوان بتنسيق الحجوزات، والموعد بمكان
// ثاني، والخدمة **ما بيها تعديل أصلاً** — يعني الزبون يغيّر طلبه
// (يزيد منظومة أو يشيل وحدة) والإداري لازم يلغي الحجز ويسوي غيره.
//
// هذا المكوّن يلمّ الثلاثة سوه ويستعمل بنفس الشكل بشاشة الحجوزات
// وبتنسيق الحجوزات — نفس الحقول ونفس السلوك بالمكانين.
export default function BookingEditPanel({
  booking,
  onSaved,
  onClose,
}: {
  booking: Booking
  onSaved: (b: Booking) => void
  onClose?: () => void
}) {
  // ── تغيير نوع الحجز (المالك ومدير النظام بس) ──
  //
  // النوع ينتحدد وقت الإنشاء، بس بالواقع ينغلط: شغل داخلي ينسجّل حجز
  // عادي، أو صيانة تنسجّل عادي فما تنحسب بإحصاءات الصيانة. وبدون
  // تغيير، الحل الوحيد إلغاء الحجز وإعادة إنشائه — فيضيع تاريخه
  // وتكليفاته وتقاريره.
  //
  // ⚠️ role يتطبّع لـ'ADMIN' للمالك بالجلسة، فهذا الشرط يغطي الاثنين.
  const { employee } = useSession()
  const canChangeType = employee?.role === 'ADMIN'
  const [bookingType, setBookingType] = useState(booking.bookingType)
  const [typeBusy, setTypeBusy] = useState(false)

  const BOOKING_TYPES: { value: 'REGULAR' | 'MAINTENANCE' | 'INTERNAL' | 'SOLAR'; label: string }[] = [
    { value: 'REGULAR', label: '📋 حجز عادي' },
    { value: 'MAINTENANCE', label: '🔧 حجز صيانة' },
    { value: 'INTERNAL', label: '🏢 شغل داخل الشركة' },
    { value: 'SOLAR', label: '☀️ حجز طاقة شمسية' },
  ]

  const changeType = async (value: string) => {
    if (value === booking.bookingType) return
    const label = BOOKING_TYPES.find((t) => t.value === value)?.label || value
    if (!window.confirm(`تحويل الحجز إلى «${label}»؟\n\nنوع الحجز يأثر على الإحصاءات والعمولات وحساب الصيانة، والتغيير ينتسجّل باسمك.`)) return
    setTypeBusy(true)
    try {
      const updated = await api.changeBookingType(booking.id, value as 'REGULAR')
      setBookingType(updated.bookingType)
      onSaved(updated)
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تغيير نوع الحجز')
      setBookingType(booking.bookingType)
    } finally { setTypeBusy(false) }
  }

  const [services, setServices] = useState<Service[]>([])
  const [serviceIds, setServiceIds] = useState<string[]>(
    (booking.services?.length ? booking.services : booking.service ? [booking.service] : []).map((s) => s.id),
  )
  const [price, setPrice] = useState(booking.quotedPrice != null ? String(booking.quotedPrice) : '')
  // datetime-local يريد "YYYY-MM-DDTHH:mm" بتوقيت محلي
  const toLocalInput = (iso: string | null) => {
    if (!iso) return ''
    const d = new Date(iso)
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
  }
  const [when, setWhen] = useState(toLocalInput(booking.scheduledAt))
  const [saving, setSaving] = useState(false)
  const [msg, setMsg] = useState<string | null>(null)

  useEffect(() => { api.getServices().then(setServices).catch(() => {}) }, [])


  const save = async () => {
    if (serviceIds.length === 0) { setMsg('اختار خدمة وحدة على الأقل'); return }
    setSaving(true)
    setMsg(null)
    try {
      let updated = await api.updateBookingDetails(booking.id, {
        serviceIds,
        quotedPrice: price.trim() === '' ? undefined : Number(price),
      })
      // الموعد إله مسار مستقل لأنه ينسجّل بسجل تغييرات المواعيد
      if (when && when !== toLocalInput(booking.scheduledAt)) {
        updated = await api.scheduleBooking(booking.id, when)
      }
      onSaved(updated)
      setMsg('انحفظ ✓')
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="mt-3 rounded-xl border border-brand-200 bg-brand-50/40 p-4">
      <h4 className="mb-3 text-sm font-bold text-[#0f2040]">✏️ تعديل الحجز</h4>

      {/* نوع الحجز — للمالك ومدير النظام بس */}
      {canChangeType && (
        <div className="mb-3 rounded-lg border border-amber-300 bg-amber-50 p-3">
          <label className="mb-1 block text-xs font-bold text-amber-900">
            نوع الحجز <span className="font-normal text-amber-700">(المالك ومدير النظام بس)</span>
          </label>
          <select
            value={bookingType}
            onChange={(e) => { setBookingType(e.target.value); void changeType(e.target.value) }}
            disabled={typeBusy}
            className="w-full rounded-lg border border-amber-300 bg-white px-3 py-2 text-sm outline-none focus:border-amber-500 disabled:opacity-50"
          >
            {BOOKING_TYPES.map((t) => (
              <option key={t.value} value={t.value}>{t.label}</option>
            ))}
          </select>
          <p className="mt-1 text-[11px] text-amber-800">
            يأثر على الإحصاءات والعمولات وحساب الصيانة — والتغيير ينتسجّل باسمك.
          </p>
        </div>
      )}

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <div className="sm:col-span-2">
          <label className="mb-1 block text-xs font-medium text-slate-600">الخدمات المطلوبة</label>
          {/* ═══ قائمة منسدلة مو مربّعات مفتوحة ═══
              «هاي مال الخدمات المطلوبة — كتلك أريدها قائمة منسدلة، ما
              أريدها هيج تاخذ نص الشاشة».
              الخدمات صارن ٣٠+، فالمربّعات المفتوحة تدفن باقي خانات
              التعديل (السعر، الموعد، الملاحظات) تحتها — الإداري يفتح
              «تعديل الحجز» عشان يغيّر الموعد فيلگه جدار خدمات.
              ⚠️ نفس المكوّن الي بالحجز الجديد بالضبط — ما نسوي منتقي
              ثاني يفترق عنه أول تعديل. */}
          <MultiSelect
            options={services.map((sv) => ({ id: sv.id, name: sv.name }))}
            selected={serviceIds}
            onChange={setServiceIds}
            placeholder="اختر الخدمات المطلوبة"
            emptyText="ماكو خدمات"
          />
        </div>

        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">كلفة العمل التقديرية</label>
          <input
            type="number"
            value={price}
            onChange={(e) => setPrice(e.target.value)}
            {...onEnter(save, { disabled: saving })}
            placeholder="غير محددة"
            className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-brand-500"
          />
        </div>

        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">
            الموعد (النهاية تنحسب ساعة بعده تلقائياً)
          </label>
          <input
            type="datetime-local"
            value={when}
            onChange={(e) => setWhen(e.target.value)}
            className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-brand-500"
          />
        </div>
      </div>

      <div className="mt-3 flex items-center gap-2">
        <button
          onClick={save}
          disabled={saving}
          className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-bold text-white disabled:opacity-50"
        >
          {saving ? 'جاري الحفظ...' : 'حفظ التعديل'}
        </button>
        {onClose && (
          <button onClick={onClose} className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-bold text-slate-600">
            إغلاق
          </button>
        )}
        {msg && <span className={`text-xs font-bold ${msg.includes('✓') ? 'text-emerald-600' : 'text-red-600'}`}>{msg}</span>}
      </div>
    </div>
  )
}
