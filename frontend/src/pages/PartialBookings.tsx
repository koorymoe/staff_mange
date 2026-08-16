import { useEffect, useState } from 'react'
import { api, type Booking, type BookingProgressReport, type SuggestedCrewMember } from '../api'

// ═══ حجوزات تحتاج إكمال ═══
//
// الحجز الي انجز جزئياً يرجع لهنا — مو للأرشيف ومو للمكتملة. الإداري
// يشوف وين وصلوا، ويحدد يوم جديد، والنظام يقترحله **نفس الكادر** الي
// طلع بالأيام الفائتة.
//
// ليش الاقتراح مو فرض؟ لأن الكادر الأول يعرف الشغل والزبون والطريق —
// فهو الافتراض الصحيح. بس الإداري إله الحق الكامل يبدّل: ممكن واحد
// منهم بإجازة، أو مكلّف بشغل ثاني، أو الشغل الباقي يحتاج مهارة ثانية.
//
// ═══ الترتيب ═══
// «هذا الترتيب مالتها»: ثلاث عدّادات فوق، وكل حجز ببطاقة تجاوب أربع
// أسئلة بصف واحد — وين وصل الشغل، شنو باقي، شنو انخلص، ومنو يكمّله —
// وتحتهن سطر واحد يحدد الموعد ويجدوله.
//
// ⚠️ «تمت جدولتها» ما تنجاب من حالة `PARTIAL`: أول ما ينحدد موعد
// الإكمال الحجز يرجع `CONFIRMED` (حتى يدخل طابور يومه عادي). فالي
// انجدول ينعرف بـ`partialCount > 0` مع موعد — بدون هاي، العدّاد
// يبقى صفر للأبد ويبان وكأن محد جدول شي.

interface Row {
  booking: Booking
  crew: SuggestedCrewMember[]
  last: BookingProgressReport | null
  days: number
}

/** حلقة التقدّم — النسبة تنقرا بلمحة بدل ما تنقرا رقم بسطر. */
function ProgressRing({ percent }: { percent: number }) {
  const p = Math.max(0, Math.min(100, percent))
  // ⚠️ اللون يتبع النسبة مو ثابت: ٩٠٪ و٣٥٪ ما ينفع يطلعون بنفس
  // الشكل — الإداري يفرز بالنظر منو قريب من الخلاص.
  const color = p >= 80 ? '#16a34a' : p >= 50 ? '#f59e0b' : '#fb923c'
  return (
    <div className="relative h-14 w-14 shrink-0">
      <svg viewBox="0 0 36 36" className="h-14 w-14 -rotate-90">
        <circle cx="18" cy="18" r="15.5" fill="none" stroke="#eef2f7" strokeWidth="3.5" />
        <circle
          cx="18" cy="18" r="15.5" fill="none" stroke={color} strokeWidth="3.5" strokeLinecap="round"
          strokeDasharray={`${(p / 100) * 97.4} 97.4`}
        />
      </svg>
      <div className="absolute inset-0 flex flex-col items-center justify-center leading-none">
        <span className="text-[11px] font-black" style={{ color }}>{p}%</span>
        <span className="text-[8px] text-slate-400">منجز</span>
      </div>
    </div>
  )
}

/** سطر معلومة بالبطاقة — عنوان صغير وتحته المحتوى. */
function Cell({ icon, title, children }: { icon: string; title: string; children: React.ReactNode }) {
  return (
    <div className="min-w-0">
      <p className="mb-1.5 flex items-center gap-1 text-[11px] font-bold text-slate-400">
        <span>{icon}</span> {title}
      </p>
      {children}
    </div>
  )
}

/** يفصل نصاً لأسطر — الكادر يكتب «سوّينا كذا، وكذا» بسطر واحد. */
function lines(text?: string | null): string[] {
  if (!text) return []
  return text.split(/[\n،؛]|\s-\s/).map((s) => s.trim()).filter(Boolean).slice(0, 3)
}

const STAT_STYLES = {
  all: 'border-slate-200 bg-white text-[#2c5aad]',
  waiting: 'border-amber-200 bg-amber-50/60 text-amber-700',
  scheduled: 'border-emerald-200 bg-emerald-50/60 text-emerald-700',
}

export default function PartialBookings() {
  const [rows, setRows] = useState<Row[]>([])
  const [loading, setLoading] = useState(true)
  const [date, setDate] = useState<Record<string, string>>({})
  const [time, setTime] = useState<Record<string, string>>({})
  const [busy, setBusy] = useState<string | null>(null)
  const [msg, setMsg] = useState<{ id: string; ok: boolean; text: string } | null>(null)
  const [openId, setOpenId] = useState<string | null>(null)

  const load = () => {
    // ⚠️ نجيب الحالتين: `PARTIAL` (ينتظر موعد) و`CONFIRMED` الي عليه
    // إنجاز جزئي سابق (انجدول). الثانية ما تنعرف إلا بـ`partialCount`.
    api.getBookings({ status: ['PARTIAL', 'CONFIRMED'] })
      .then(async (list) => {
        const mine = list.filter((b) => b.status === 'PARTIAL' || (b.partialCount ?? 0) > 0)
        const built = await Promise.all(mine.map(async (b) => {
          const [crew, reports] = await Promise.all([
            api.getSuggestedCrew(b.id).catch(() => [] as SuggestedCrewMember[]),
            api.getBookingProgress(b.id).catch(() => [] as BookingProgressReport[]),
          ])
          const sorted = [...reports].sort((a, c) => c.dayNumber - a.dayNumber)
          return { booking: b, crew, last: sorted[0] ?? null, days: reports.length }
        }))
        // الأقدم أول: الي منتظر أكثر أولى بموعد
        built.sort((a, c) => new Date(a.booking.lastPartialAt || a.booking.createdAt).getTime()
          - new Date(c.booking.lastPartialAt || c.booking.createdAt).getTime())
        setRows(built)
      })
      .catch(() => setRows([]))
      .finally(() => setLoading(false))
  }
  useEffect(load, [])

  const waiting = rows.filter((r) => r.booking.status === 'PARTIAL')
  const scheduled = rows.filter((r) => r.booking.status !== 'PARTIAL')

  const schedule = async (r: Row) => {
    const d = date[r.booking.id]
    const t = time[r.booking.id] || '09:00'
    if (!d) {
      setMsg({ id: r.booking.id, ok: false, text: 'حدد تاريخ الإكمال أول' })
      return
    }
    setBusy(r.booking.id); setMsg(null)
    try {
      await api.scheduleContinuation(r.booking.id, `${d}T${t}`)
      setMsg({ id: r.booking.id, ok: true, text: 'انجدول — الحجز رجع للتنسيق بموعده الجديد' })
      load()
    } catch (e) {
      setMsg({ id: r.booking.id, ok: false, text: e instanceof Error ? e.message : 'تعذر الجدولة' })
    } finally {
      setBusy(null)
    }
  }

  return (
    <div dir="rtl" className="space-y-4">
      {/* ═══ العدّادات ═══ */}
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        {([
          { key: 'all', icon: '🗓️', label: 'كل الحجوزات التي تحتاج إكمال', n: rows.length },
          { key: 'waiting', icon: '🕐', label: 'بانتظار تحديد موعد', n: waiting.length },
          { key: 'scheduled', icon: '📅', label: 'تمت جدولتها', n: scheduled.length },
        ] as const).map((s) => (
          <div key={s.key} className={`flex items-center justify-between rounded-2xl border px-5 py-4 ${STAT_STYLES[s.key]}`}>
            <span className="text-xl opacity-70">{s.icon}</span>
            <div className="text-left">
              <p className="text-[11px] font-bold opacity-80">{s.label}</p>
              <p className="text-2xl font-black leading-tight">{s.n}</p>
            </div>
          </div>
        ))}
      </div>

      {loading && <p className="text-slate-400">جاري التحميل...</p>}
      {!loading && rows.length === 0 && (
        <p className="rounded-2xl border border-emerald-200 bg-emerald-50 p-4 text-sm font-bold text-emerald-800">
          ✅ ماكو ولا حجز ناقص — كل الحجوزات إما مكتملة أو مجدولة.
        </p>
      )}

      {rows.map((r) => {
        const b = r.booking
        const done = lines(r.last?.workDone)
        const rest = lines(r.last?.remainingWork)
        const isScheduled = b.status !== 'PARTIAL'
        return (
          <div key={b.id} className="overflow-hidden rounded-2xl bg-white shadow-[0_2px_12px_rgba(15,32,64,0.06)]">
            {/* ═══ الرأس ═══ */}
            <div className="flex flex-wrap items-center gap-3 px-5 py-3.5">
              <span className="font-mono text-sm font-black text-[#2c5aad]">{b.code}</span>
              <span className="text-sm font-extrabold text-[#0f2040]">{b.customer?.name || 'زبون غير معروف'}</span>

              <div className="flex flex-1 flex-wrap items-center justify-center gap-x-4 gap-y-1 text-[11px] text-slate-500">
                <span>🛠️ {b.services?.map((s) => s.name).join(' + ') || b.service?.name || '—'}</span>
                <span>📍 {b.address || b.customer?.location || '—'}</span>
                <span dir="ltr">📞 {b.customer?.phone || '—'}</span>
              </div>

              {isScheduled && (
                <span className="rounded-full bg-emerald-50 px-2.5 py-1 text-[10px] font-black text-emerald-700">
                  ✓ انجدول
                </span>
              )}
              <ProgressRing percent={r.last?.percentDone ?? 0} />
            </div>

            {/* ═══ الأسئلة الأربعة ═══
                ⚠️ بصف واحد عن قصد: الإداري يقرر «منو يكمّله ومتى» من
                نظرة وحدة، بلا ما يفتح الحجز ولا يدوّر بتقارير. */}
            <div className="grid grid-cols-1 gap-4 border-y border-slate-100 bg-slate-50/40 px-5 py-4 sm:grid-cols-2 lg:grid-cols-4">
              <Cell icon="📍" title="وين وصل الشغل">
                <span className="inline-block rounded-md bg-amber-100 px-2 py-0.5 text-[10px] font-black text-amber-800">
                  {r.days > 1 ? `${r.days} أيام` : `اليوم ${r.days || 1}`}
                </span>
                <p className="mt-1 text-[12px] font-semibold leading-snug text-slate-700">
                  {r.last?.workDone?.split(/[\n،؛]/)[0] || 'ماكو تقرير بعد'}
                </p>
              </Cell>

              <Cell icon="📋" title="الباقي عليكم">
                {rest.length === 0
                  ? <p className="text-[12px] text-slate-400">ما انكتب</p>
                  : rest.map((l, i) => <p key={i} className="text-[12px] leading-snug text-slate-600">{l}</p>)}
              </Cell>

              <Cell icon="✅" title="المنجز (الخلاصة)">
                {done.length === 0
                  ? <p className="text-[12px] text-slate-400">ما انكتب</p>
                  : done.map((l, i) => <p key={i} className="text-[12px] leading-snug text-slate-600">{l}</p>)}
              </Cell>

              <Cell icon="👷" title="الكادر المقترح (الي اشتغلوا قبل)">
                {r.crew.length === 0 ? (
                  <p className="text-[12px] text-slate-400">ماكو كادر مسجّل</p>
                ) : (
                  <div className="flex flex-wrap gap-1.5">
                    {r.crew.map((c) => (
                      <span
                        key={c.employeeId}
                        // ⚠️ اللون يگول متاح لو لا: الإداري ما يجدول
                        // يوماً على واحد بإجازة وبعدين يكتشفها.
                        className={`rounded-lg px-2 py-0.5 text-[11px] font-bold ${
                          c.available ? 'bg-emerald-50 text-emerald-800' : 'bg-red-50 text-red-700'
                        }`}
                        title={c.available ? `${c.role} · اشتغل ${c.daysWorked} يوم` : c.note}
                      >
                        {c.name}
                      </span>
                    ))}
                  </div>
                )}
              </Cell>
            </div>

            {/* ═══ الموعد والجدولة ═══ */}
            <div className="flex flex-wrap items-center gap-2 px-5 py-3">
              <button
                onClick={() => setOpenId(openId === b.id ? null : b.id)}
                className="rounded-lg border border-slate-200 px-3 py-1.5 text-[11px] font-bold text-slate-500 hover:bg-slate-50"
              >
                {openId === b.id ? 'إخفاء التفاصيل' : 'تفاصيل الحجز ›'}
              </button>

              {b.partialCount > 2 && (
                <span className="rounded-full bg-red-50 px-2.5 py-1 text-[10px] font-bold text-red-700">
                  ⚠️ انأجّل {b.partialCount} مرات — راجع التقدير
                </span>
              )}

              <div className="mr-auto flex flex-wrap items-center gap-2">
                <span className="text-[11px] font-bold text-slate-400">موعد الإكمال</span>
                <input
                  type="date"
                  value={date[b.id] || ''}
                  onChange={(e) => setDate((p) => ({ ...p, [b.id]: e.target.value }))}
                  className="rounded-lg border border-slate-300 px-2.5 py-1.5 text-[12px] outline-none focus:border-brand-500"
                />
                <input
                  type="time"
                  value={time[b.id] || '09:00'}
                  onChange={(e) => setTime((p) => ({ ...p, [b.id]: e.target.value }))}
                  className="rounded-lg border border-slate-300 px-2.5 py-1.5 text-[12px] outline-none focus:border-brand-500"
                />
                <button
                  onClick={() => schedule(r)}
                  disabled={busy === b.id}
                  className="rounded-lg bg-[#2c5aad] px-4 py-1.5 text-[12px] font-bold text-white hover:bg-[#0f2040] disabled:opacity-50"
                >
                  {busy === b.id ? 'جاري...' : '📅 جدول الإكمال'}
                </button>
              </div>
            </div>

            {msg?.id === b.id && (
              <p className={`px-5 pb-3 text-[12px] font-bold ${msg.ok ? 'text-emerald-700' : 'text-red-600'}`}>
                {msg.text}
              </p>
            )}

            {/* التفاصيل الكاملة — تنفتح بالطلب مو دايماً: البطاقة
                تجاوب السؤال اليومي، والباقي لمن يحتاجه. */}
            {openId === b.id && (
              <div className="border-t border-slate-100 bg-slate-50/50 px-5 py-4 text-[12px] text-slate-600">
                <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                  <div>
                    <p className="text-[11px] text-slate-400">موعد الحجز الحالي</p>
                    <p className="font-bold text-slate-700">
                      {b.scheduledAt ? new Date(b.scheduledAt).toLocaleString('ar-IQ') : 'غير محدد'}
                    </p>
                  </div>
                  <div>
                    <p className="text-[11px] text-slate-400">آخر إنجاز جزئي</p>
                    <p className="font-bold text-slate-700">
                      {b.lastPartialAt ? new Date(b.lastPartialAt).toLocaleString('ar-IQ') : '—'}
                    </p>
                  </div>
                  <div>
                    <p className="text-[11px] text-slate-400">منو كتب التقرير</p>
                    <p className="font-bold text-slate-700">{r.last?.reportedBy?.name || '—'}</p>
                  </div>
                </div>
                {r.last?.blockers && (
                  <p className="mt-3 rounded-lg bg-red-50 px-3 py-2 text-red-800">
                    🚧 معوقات: <b>{r.last.blockers}</b>
                  </p>
                )}
                {r.last?.materialsUsed && (
                  <p className="mt-2 rounded-lg bg-white px-3 py-2">
                    📦 مواد انستهلكت: <b>{r.last.materialsUsed}</b>
                  </p>
                )}
                <p className="mt-3 text-[11px] text-slate-400">
                  الكادر يبقى مكلّف تلقائياً. لتبديل أي واحد، افتح الحجز من «تنسيق الحجوزات».
                </p>
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}
