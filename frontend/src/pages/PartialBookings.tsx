import { useEffect, useState } from 'react'
import { api, type Booking, type SuggestedCrewMember } from '../api'
import BookingProgressTimeline from '../components/BookingProgressTimeline'

// ═══ حجوزات تحتاج إكمال ═══
//
// الحجز الي انجز جزئياً يرجع لهنا — مو للأرشيف ومو للمكتملة. الإداري
// يشوف وين وصلوا، ويحدد يوم جديد، والنظام يقترحله **نفس الكادر** الي
// طلع بالأيام الفائتة.
//
// ليش الاقتراح مو فرض؟ لأن الكادر الأول يعرف الشغل والزبون والطريق —
// فهو الافتراض الصحيح. بس الإداري إله الحق الكامل يبدّل: ممكن واحد
// منهم بإجازة، أو مكلّف بشغل ثاني، أو الشغل الباقي يحتاج مهارة ثانية.
// لهذا نبيّن حالة كل واحد منهم (متاح لو لا) قبل ما يقرر.
export default function PartialBookings() {
  const [rows, setRows] = useState<Booking[]>([])
  const [loading, setLoading] = useState(true)
  const [crews, setCrews] = useState<Record<string, SuggestedCrewMember[]>>({})
  const [when, setWhen] = useState<Record<string, string>>({})
  const [busy, setBusy] = useState<string | null>(null)
  const [msg, setMsg] = useState<{ id: string; ok: boolean; text: string } | null>(null)

  const load = () => {
    api.getBookings({ status: 'PARTIAL' })
      .then(async (list) => {
        setRows(list)
        // الكادر المقترح لكل حجز — بالتوازي مو واحد ورا الثاني
        const pairs = await Promise.all(
          list.map(async (b) => [b.id, await api.getSuggestedCrew(b.id).catch(() => [])] as const),
        )
        setCrews(Object.fromEntries(pairs))
      })
      .catch(() => setRows([]))
      .finally(() => setLoading(false))
  }
  useEffect(load, [])

  const reschedule = async (b: Booking) => {
    const value = when[b.id]
    if (!value) {
      setMsg({ id: b.id, ok: false, text: 'حدد موعد اليوم الجاي أول' })
      return
    }
    setBusy(b.id); setMsg(null)
    try {
      await api.scheduleContinuation(b.id, value)
      setMsg({ id: b.id, ok: true, text: 'انجدول — الحجز رجع للتنسيق بموعده الجديد' })
      load()
    } catch (e) {
      setMsg({ id: b.id, ok: false, text: e instanceof Error ? e.message : 'تعذر الجدولة' })
    } finally {
      setBusy(null)
    }
  }

  return (
    <div className="space-y-4">
      <div className="rounded-2xl bg-white p-5 shadow-[0_2px_12px_rgba(0,0,0,0.06)]">
        <h1 className="text-xl font-extrabold text-[#0f2040]">🔄 حجوزات تحتاج إكمال</h1>
        <p className="mt-1 text-sm text-slate-500">
          حجوزات انجز منها جزء والباقي يحتاج يوم جديد. شوف وين وصلوا، حدد الموعد،
          والنظام يقترحلك نفس الكادر — وانت تقدر تبدّل.
        </p>
      </div>

      {loading && <p className="text-slate-400">جاري التحميل...</p>}
      {!loading && rows.length === 0 && (
        <p className="rounded-2xl border border-emerald-200 bg-emerald-50 p-4 text-sm font-bold text-emerald-800">
          ✅ ماكو ولا حجز ناقص — كل الحجوزات إما مكتملة أو مجدولة.
        </p>
      )}

      {rows.map((b) => {
        const crew = crews[b.id] || []
        return (
          <div key={b.id} className="rounded-2xl bg-white p-5 shadow-[0_2px_12px_rgba(0,0,0,0.06)]">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div>
                <span className="text-base font-extrabold text-[#0f2040]">{b.code}</span>
                <span className="mr-2 text-sm text-slate-600">{b.customer?.name}</span>
                {b.customer?.phone && <span className="mr-2 text-xs text-slate-400">{b.customer.phone}</span>}
              </div>
              {b.partialCount > 1 && (
                <span className="rounded-full bg-red-50 px-2.5 py-0.5 text-xs font-bold text-red-700">
                  ⚠️ انأجّل {b.partialCount} مرات — راجع التقدير
                </span>
              )}
            </div>
            {b.address && <p className="mt-1 text-xs text-slate-500">📍 {b.address}</p>}

            <BookingProgressTimeline bookingId={b.id} />

            {/* الكادر المقترح */}
            <div className="mt-3 rounded-xl border border-slate-200 p-3">
              <p className="text-sm font-bold text-[#0f2040]">👷 الكادر المقترح (الي اشتغلوا قبل)</p>
              {crew.length === 0 ? (
                <p className="mt-1 text-xs text-slate-400">ماكو كادر مسجّل — كلّف من شاشة التنسيق.</p>
              ) : (
                <div className="mt-2 flex flex-wrap gap-2">
                  {crew.map((c) => (
                    <span
                      key={c.employeeId}
                      className={`rounded-full px-3 py-1 text-xs font-bold ${
                        c.available ? 'bg-emerald-50 text-emerald-800' : 'bg-red-50 text-red-700'
                      }`}
                    >
                      {c.name} · {c.role}
                      {!c.available && ` — ${c.note}`}
                    </span>
                  ))}
                </div>
              )}
              <p className="mt-2 text-[11px] text-slate-400">
                الكادر يبقى مكلّف تلقائياً. لتبديل أي واحد، افتح الحجز من شاشة «تنسيق الحجوزات».
              </p>
            </div>

            {/* الموعد الجديد */}
            <div className="mt-3 flex flex-wrap items-end gap-2">
              <div>
                <label className="block text-xs text-slate-500">موعد الإكمال</label>
                <input
                  type="datetime-local"
                  value={when[b.id] || ''}
                  onChange={(e) => setWhen((p) => ({ ...p, [b.id]: e.target.value }))}
                  className="mt-1 rounded-lg border border-slate-300 px-3 py-2 text-sm"
                />
              </div>
              <button
                onClick={() => reschedule(b)}
                disabled={busy === b.id}
                className="rounded-lg bg-[#0f2040] px-4 py-2 text-sm font-bold text-white disabled:opacity-50"
              >
                {busy === b.id ? 'جاري...' : '📅 جدول الإكمال'}
              </button>
              {msg?.id === b.id && (
                <span className={`text-sm font-bold ${msg.ok ? 'text-emerald-700' : 'text-red-600'}`}>
                  {msg.text}
                </span>
              )}
            </div>
          </div>
        )
      })}
    </div>
  )
}
