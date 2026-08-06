import { useEffect, useState } from 'react'
import { api, type Booking, type Service } from '../api'

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

  const toggleService = (id: string) =>
    setServiceIds((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]))

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

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <div className="sm:col-span-2">
          <label className="mb-1 block text-xs font-medium text-slate-600">الخدمات المطلوبة</label>
          <div className="grid grid-cols-1 gap-1.5 sm:grid-cols-3">
            {services.map((s) => (
              <label
                key={s.id}
                className={`flex cursor-pointer items-center gap-2 rounded-lg border px-2.5 py-1.5 text-xs ${
                  serviceIds.includes(s.id)
                    ? 'border-brand-500 bg-white font-bold text-brand-800'
                    : 'border-slate-200 bg-white text-slate-600'
                }`}
              >
                <input type="checkbox" checked={serviceIds.includes(s.id)} onChange={() => toggleService(s.id)} />
                {s.name}
              </label>
            ))}
          </div>
        </div>

        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">كلفة العمل التقديرية</label>
          <input
            type="number"
            value={price}
            onChange={(e) => setPrice(e.target.value)}
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
