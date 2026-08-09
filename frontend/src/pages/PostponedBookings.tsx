import { useCallback, useEffect, useState } from 'react'
import { api, type Booking } from '../api'
import { makeFilter } from '../utils/search'

// ═══ الحجوزات المؤجلة بلا موعد ═══
//
// الحجز الي انأجّل بلا موعد **مقصود** إنه ينزاح من جدول اليوم — وإلا
// الكادر يتحضّر لحجز ماكو. بس هذا يخلق مشكلة ثانية: وين يروح؟ بدون
// هاي الشاشة يضيع، لا بجدول ولا بأي مكان.
//
// هنا يقعدون بطابور، **الأقدم تأجيلاً أول** لأنه الي منتظر أكثر، ومن
// كل سطر الإداري يحدد الموعد مباشرة — وأول ما يحدده الحجز يرجع
// لجدوله الطبيعي ويختفي من هنا.
export default function PostponedBookings() {
  const [rows, setRows] = useState<Booking[]>([])
  const [drafts, setDrafts] = useState<Record<string, string>>({})
  const [busy, setBusy] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      setRows(await api.getPostponedBookings())
      setError(null)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذر جلب الحجوزات المؤجلة')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void load() }, [load])

  const setDate = async (b: Booking) => {
    const when = drafts[b.id]
    if (!when) { setError('حدد الموعد أول'); return }
    setBusy(b.id)
    setError(null)
    try {
      await api.scheduleBooking(b.id, when)
      await load()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذر تحديد الموعد')
    } finally { setBusy(null) }
  }

  // نفس تطبيع البحث العربي المستعمل بكل الشاشات (همزة، تاء مربوطة، أرقام هندية)
  const visible = rows.filter(
    makeFilter(search, (b: Booking) => [b.code, b.customer?.name, b.customer?.phone, b.postponeReason, b.address]))

  const daysSince = (iso: string | null) => {
    if (!iso) return null
    return Math.floor((Date.now() - new Date(iso).getTime()) / 86_400_000)
  }

  return (
    <div dir="rtl" className="space-y-5">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">📅 الحجوزات المؤجلة</h2>
        <p className="mt-1 text-slate-500">
          حجوزات انأجّلت بدون موعد. منزاحة عن جدول اليوم قصداً — حدد لها موعد وترجع لمحلها.
        </p>
      </div>

      {error && <p className="rounded-lg bg-red-50 p-4 text-red-600">{error}</p>}

      <input
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="بحث بالكود أو الزبون أو الهاتف أو سبب التأجيل..."
        className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 sm:max-w-md"
      />

      {loading && <p className="text-slate-400">جاري التحميل...</p>}
      {!loading && rows.length === 0 && (
        <p className="rounded-xl border border-white bg-white p-8 text-center text-slate-400">
          ماكو حجوزات مؤجلة بلا موعد ✓
        </p>
      )}
      {!loading && rows.length > 0 && visible.length === 0 && (
        <p className="text-slate-400">ماكو نتيجة للبحث.</p>
      )}

      <div className="space-y-3">
        {visible.map((b) => {
          const waited = daysSince(b.lastPostponedAt)
          return (
            <div key={b.id} className="rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p className="font-bold text-[#0f2040]">
                    {b.code} — {b.customer?.name || 'بلا زبون'}
                    {b.postponeCount > 2 && (
                      <span className="mr-2 rounded-full bg-red-100 px-2 py-0.5 text-[11px] font-bold text-red-700">
                        تأجّل {b.postponeCount} مرات
                      </span>
                    )}
                  </p>
                  {b.customer?.phone && <p className="text-xs text-slate-500">📞 {b.customer.phone}</p>}
                  {b.postponeReason && <p className="mt-1 text-xs text-slate-600">السبب: {b.postponeReason}</p>}
                  {waited != null && (
                    <p className={`mt-1 text-[11px] font-bold ${waited >= 7 ? 'text-red-600' : 'text-slate-400'}`}>
                      منتظر من {waited} يوم
                    </p>
                  )}
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  <input
                    type="datetime-local"
                    value={drafts[b.id] || ''}
                    onChange={(e) => setDrafts((p) => ({ ...p, [b.id]: e.target.value }))}
                    className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                  />
                  <button
                    onClick={() => setDate(b)}
                    disabled={busy === b.id}
                    className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-bold text-white disabled:opacity-50"
                  >
                    {busy === b.id ? '...' : 'حدد الموعد'}
                  </button>
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
