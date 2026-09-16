import { useEffect, useState } from 'react'
import { api, type BookingTimeline as Timeline } from '../api'

// ═══ الخط الزمني للحجز ═══
//
// قصة الحجز كاملة بالترتيب + التأخيرات الستة.
//
// ⚠️ القياس الي ما ينطبق **ينختفي كلياً** — ما يطلع «تأخر الفوترة: —».
// حجز ما وصل الفوترة ما إله تأخر فوترة، وعرض الشرطة يخلي القارئ يظن
// إن أكو قياس ناقص أو مكسور.
//
// ⚠️ الحدود **معلنة** بجنب كل رقم: صاحب العمل لازم يعرف على أي أساس
// انتأشّر موظفه «متأخر»، مو رقم مخبّى بالكود.
//
// ⚠️ عرض بس — ماكو غرامة تلقائية تنبني على هذي الأرقام بهاي المرحلة.

const ICONS: Record<string, string> = {
  CREATED: '📝', CONTACTED: '📞', CONFIRMED: '📌', ASSIGNED: '👷',
  SCHEDULE_CHANGE: '📅', STARTED: '🚗', WORK_STOPPED: '⏸️', PARTIAL: '🔄',
  COMPLETED: '🏁', INVOICED: '🧾', INVOICE_APPROVED: '✅', QUALITY: '⭐',
  MONITOR: '👁️', CANCELLED: '✖️', POSTPONED: '⏳', WAITING: '📵',
}

/** يحوّل الدقايق لصيغة يقراها بني آدم: «٣ ساعات» مو «١٨٠ دقيقة». */
function humanMinutes(m: number): string {
  if (m < 60) return `${m} دقيقة`
  const h = Math.floor(m / 60)
  const rem = m % 60
  if (h < 24) return rem ? `${h} ساعة و${rem} دقيقة` : `${h} ساعة`
  const d = Math.floor(h / 24)
  const rh = h % 24
  return rh ? `${d} يوم و${rh} ساعة` : `${d} يوم`
}

export default function BookingTimelineView({ bookingId }: { bookingId: string }) {
  // ⚠️ النتيجة تنحفظ **مع رقم الحجز الي جابها**، و«جاري التحميل»
  // تنشتق منها بدل ما تكون حالة منفصلة. سببين:
  //   • لو المستخدم بدّل الحجز قبل ما يرد الأول، جواب الحجز القديم
  //     يوصل متأخر — وبالمقارنة `loaded.id !== bookingId` ينرفض،
  //     فما يشوف خط زمني لحجز ثاني وهو ما يدري.
  //   • ومؤشر التحميل يطلع فوراً عند التبديل بلا ما نلمس أي حالة
  //     بجسم الـeffect (الي يسبّب دورة رسم زايدة بكل فتحة).
  const [loaded, setLoaded] = useState<{ id: string; tl: Timeline | null } | null>(null)
  const loading = loaded?.id !== bookingId
  const tl = loading ? null : loaded.tl

  useEffect(() => {
    let alive = true
    void (async () => {
      const data = await api.getBookingTimeline(bookingId).catch(() => null)
      if (alive) setLoaded({ id: bookingId, tl: data })
    })()
    return () => { alive = false }
  }, [bookingId])

  if (loading) return <p className="text-xs text-slate-400">جاري تحميل الخط الزمني...</p>
  if (!tl) return null

  // الي ينطبق بس — الباقي ينختفي مو يطلع بشرطة
  const delays = tl.delays.filter((d) => d.minutes !== null)

  return (
    <div dir="rtl" className="mt-3 rounded-xl border border-slate-200 bg-slate-50/60 p-4">
      <h4 className="mb-3 text-sm font-bold text-[#0f2040]">🕒 قصة الحجز</h4>

      <div className="space-y-0">
        {tl.events.map((e, i) => (
          <div key={i} className="flex gap-3">
            {/* العمود: نقطة + خيط. آخر حدث بلا خيط تحته. */}
            <div className="flex flex-col items-center">
              <span className="mt-1 text-sm">{ICONS[e.kind] || '•'}</span>
              {i < tl.events.length - 1 && <span className="my-1 w-px flex-1 bg-slate-300" />}
            </div>
            <div className="flex-1 pb-3">
              <p className="text-sm font-bold text-slate-800">{e.title}</p>
              {e.detail && <p className="text-xs text-slate-600">{e.detail}</p>}
              <p className="text-[11px] text-slate-400">
                {new Date(e.at).toLocaleString('en-GB')}
                {e.actor && <> · {e.actor}</>}
              </p>
            </div>
          </div>
        ))}
        {tl.events.length === 0 && <p className="text-xs text-slate-400">ماكو أحداث مسجّلة.</p>}
      </div>

      {delays.length > 0 && (
        <>
          <div className="my-3 h-px bg-slate-200" />
          <h4 className="mb-2 text-sm font-bold text-[#0f2040]">⏱️ الأوقات</h4>
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
            {delays.map((d) => (
              <div
                key={d.key}
                className={`rounded-lg border px-3 py-2 text-xs ${d.breached ? 'border-red-300 bg-red-50' : 'border-slate-200 bg-white'}`}
              >
                <div className="flex items-center justify-between gap-2">
                  <span className="text-slate-600">{d.label}</span>
                  <b className={d.breached ? 'text-red-700' : 'text-slate-800'}>
                    {humanMinutes(d.minutes!)}
                  </b>
                </div>
                <p className="mt-0.5 text-[10.5px] text-slate-400">
                  {d.owner} · الحد المعلن: {humanMinutes(d.thresholdMinutes)}
                  {d.breached && <span className="mr-1 font-bold text-red-600">— تجاوز</span>}
                </p>
              </div>
            ))}
          </div>
          <p className="mt-2 text-[10.5px] text-slate-400">
            ⚠️ عرض للمتابعة بس — ماكو غرامة تلقائية تنبني على هذي الأرقام.
          </p>
        </>
      )}
    </div>
  )
}
