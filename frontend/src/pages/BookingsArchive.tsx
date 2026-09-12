import { useEffect, useMemo, useState } from 'react'
import { api, type Booking } from '../api'
import { useSession } from '../session'
import BookingLifecycleActions from '../components/BookingLifecycleActions'
import { matches } from '../utils/search'
import LocateHint from '../components/LocateHint'
import BookingCodeChip from '../components/BookingCodeChip'

// ═══ أرشيف الحجوزات ═══
//
// «الحذف» بهذا النظام ما يمحي. قبل كان يمحي فعلاً — الحجز وتاريخه وسبب
// إلغائه يروحون للأبد، وما نكدر نجاوب «شكد حجز انلغى الشهر هذا وليش؟».
//
// هاي الشاشة هي الجواب: كل حجز انحذف، بسببه ومنو حذفه ومتى، وتكدر
// ترجّعه لو الزبون غيّر رأيه.
export default function BookingsArchive() {
  const { employee } = useSession()
  const isAdmin = employee?.role === 'ADMIN' || employee?.role === 'OWNER'

  const [rows, setRows] = useState<Booking[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')

  useEffect(() => {
    api
      .getArchivedBookings(300)
      .then(setRows)
      .catch((e) => setError(e instanceof Error ? e.message : 'تعذر جلب الأرشيف'))
      .finally(() => setLoading(false))
  }, [])

  const filtered = useMemo(() => {
    if (!search.trim()) return rows
    return rows.filter((b) =>
      matches([b.code, b.customer?.name, b.customer?.phone, b.archiveReason, b.address], search),
    )
  }, [rows, search])

  // «شكد انلغى هذا الشهر» — أول سؤال يجي بالبال، فنجاوبه بلا ما ينسأل
  const thisMonth = useMemo(() => {
    const now = new Date()
    return rows.filter((b) => {
      if (!b.archivedAt) return false
      const d = new Date(b.archivedAt)
      return d.getMonth() === now.getMonth() && d.getFullYear() === now.getFullYear()
    }).length
  }, [rows])

  const fmt = (iso: string | null) =>
    iso ? new Date(iso).toLocaleString('ar-IQ', { dateStyle: 'medium', timeStyle: 'short' }) : '—'

  return (
    <div className="space-y-4">
      <div className="rounded-2xl bg-white p-5 shadow-[0_2px_12px_rgba(0,0,0,0.06)]">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h1 className="flex items-center gap-2 text-xl font-extrabold text-[#0f2040]">
              🗄️ أرشيف الحجوزات
            </h1>
            <p className="mt-1 text-xs text-slate-500">
              الحجوزات المحذوفة — محفوظة بكل تفاصيلها وسبب حذفها، ومو ممحية.
            </p>
          </div>
          <div className="flex gap-3">
            <div className="rounded-xl border-2 border-slate-200 bg-slate-50 px-4 py-2 text-center">
              <div className="text-lg font-black text-slate-700">{rows.length}</div>
              <div className="text-[10px] text-slate-500">بالأرشيف</div>
            </div>
            <div className="rounded-xl border-2 border-amber-200 bg-amber-50 px-4 py-2 text-center">
              <div className="text-lg font-black text-amber-700">{thisMonth}</div>
              <div className="text-[10px] text-amber-600">انحذفن هذا الشهر</div>
            </div>
          </div>
        </div>

        <LocateHint query={search} localCount={filtered.length} currentRoute="/bookings-archive" />
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="ابحث بالكود، اسم الزبون، الهاتف، أو سبب الحذف..."
          className="mt-4 w-full rounded-lg border border-slate-300 bg-slate-50 px-4 py-2.5 text-sm outline-none focus:border-brand-500"
        />
      </div>

      {loading && <div className="rounded-2xl bg-white p-8 text-center text-sm text-slate-500">جاري التحميل...</div>}
      {error && <div className="rounded-2xl bg-red-50 p-4 text-sm font-bold text-red-700">{error}</div>}

      {!loading && !error && filtered.length === 0 && (
        <div className="rounded-2xl bg-white p-10 text-center">
          <div className="text-4xl">🗄️</div>
          <p className="mt-3 text-sm font-bold text-slate-600">
            {rows.length === 0 ? 'ماكو ولا حجز بالأرشيف' : 'ماكو نتيجة تطابق البحث'}
          </p>
        </div>
      )}

      <div className="space-y-3">
        {filtered.map((b) => (
          <div key={b.id} className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <div className="flex items-center gap-2">
                  <span className="rounded-lg bg-slate-800 px-2.5 py-1 text-xs font-black text-white"><BookingCodeChip code={b.code} /></span>
                  <span className="text-sm font-bold text-[#0f2040]">{b.customer?.name || '—'}</span>
                  {b.customer?.phone && <span className="text-xs text-slate-500" dir="ltr">{b.customer.phone}</span>}
                </div>
                <div className="mt-1.5 flex flex-wrap gap-x-4 gap-y-1 text-xs text-slate-500">
                  {b.service && <span>الخدمة: {b.service.name}</span>}
                  {b.quotedPrice != null && <span>الكلفة: {b.quotedPrice.toLocaleString()} د.ع</span>}
                  {b.address && <span>{b.address}</span>}
                  {b.postponeCount > 0 && <span className="text-amber-600">تأجل {b.postponeCount} مرة</span>}
                  {b.contactAttempts > 0 && <span>محاولات اتصال: {b.contactAttempts}</span>}
                </div>
              </div>
              <div className="text-left text-xs text-slate-400">{fmt(b.archivedAt)}</div>
            </div>

            <div className="mt-3 rounded-lg border border-red-200 bg-red-50/60 p-2.5">
              <span className="text-xs font-bold text-red-800">سبب الحذف:</span>{' '}
              <span className="text-xs text-red-700">{b.archiveReason || 'غير مذكور'}</span>
            </div>

            {isAdmin && (
              <BookingLifecycleActions
                booking={b}
                canArchive
                onChanged={(u) =>
                  // رجع للعمل → يطلع من هاي الشاشة
                  setRows((prev) => (u.archivedAt ? prev.map((x) => (x.id === u.id ? u : x)) : prev.filter((x) => x.id !== u.id)))
                }
              />
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
