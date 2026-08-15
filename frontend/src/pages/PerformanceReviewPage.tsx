import { useEffect, useState, useCallback } from 'react'
import { api, REVIEW_RATINGS, type BookingAwaitingReview, type CrewReviewState, type ReviewRating } from '../api'

// ═══ تقييم الأداء — لكل حجز ═══
//
// كان التقييم **حر**: الليدر يفتح الشاشة، يختار موظف من قائمة، ويكتب
// سبب. بلا ربط بشغل معيّن.
//
// وهاي تخلي التقييم بلا سياق: بعد أسبوع تقرا «محمد — يحتاج تدريب»
// وما تعرف بأي شغلة، ولا منو كان وياه، ولا شنو صار. والليدر نفسه
// ينسى، فيصير التقييم انطباع عام مو ملاحظة على شغل حقيقي.
//
// «الليدر يكدر يقيّم فريقه لكل حجز يطلعوله، مو مرة وحدة باليوم».
//
// ⚠️ وخليناه **بسيط بقصد**: زرّين وملاحظة اختيارية. التصميم الأصلي
// كان بيه أربع نجوم لكل بند (التزام، سرعة، جودة تنفيذ...) — يعني
// ١٢ ضغطة لتقييم ثلاثة فنيين بحجز واحد. الليدر راجع من شغل ميداني
// وبيده تلفون، والشاشة الي تطلب منه ١٢ ضغطة **ما ينستخدمها** —
// فينتهي بينا لصفر تقييمات بدل تقييمات ناقصة.
export default function PerformanceReviewPage() {
  const [bookings, setBookings] = useState<BookingAwaitingReview[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState<'pending' | 'done' | 'all'>('pending')
  const [search, setSearch] = useState('')
  const [err, setErr] = useState<string | null>(null)

  const load = useCallback(() => {
    api.getMyBookingsForReview()
      .then(setBookings)
      .catch((e) => setErr(e instanceof Error ? e.message : 'تعذر جلب الحجوزات'))
      .finally(() => setLoading(false))
  }, [])
  useEffect(() => { load() }, [load])

  const isDone = (b: BookingAwaitingReview) => b.crew.every((c) => c.rating)
  const pendingCount = bookings.filter((b) => !isDone(b)).length
  const doneCount = bookings.length - pendingCount
  const ratedPeople = bookings.reduce((n, b) => n + b.crew.filter((c) => c.rating).length, 0)

  const shown = bookings
    .filter((b) => filter === 'all' || (filter === 'done' ? isDone(b) : !isDone(b)))
    .filter((b) => {
      const q = search.trim()
      if (!q) return true
      return `${b.code} ${b.customerName}`.includes(q)
    })

  if (loading) return <p className="p-6 text-center text-slate-400">جارٍ التحميل...</p>

  return (
    <div dir="rtl" className="space-y-5">
      {/* ═══ العنوان ═══ */}
      <div className="flex items-center gap-3">
        <span className="flex h-11 w-11 items-center justify-center rounded-xl bg-violet-50 text-xl">⭐</span>
        <div>
          <h1 className="text-2xl font-black text-[#0f2040]">تقييم الأداء</h1>
          <p className="text-xs text-slate-500">قيّم كل واحد طلع وياك — بكل حجز على حدة</p>
        </div>
      </div>

      {err && (
        <div className="rounded-xl border-2 border-red-300 bg-red-50 px-4 py-3 text-sm font-bold text-red-700">{err}</div>
      )}

      {/* ═══ الأرقام ═══ */}
      <div className="grid grid-cols-3 gap-3">
        <Stat icon="📋" label="حجوزات تحتاج تقييم" value={pendingCount} tone="amber" />
        <Stat icon="✅" label="حجوزات انقيّمت" value={doneCount} tone="emerald" />
        <Stat icon="👥" label="تقييمات انسجّلت" value={ratedPeople} tone="violet" />
      </div>

      {/* ═══ الفلاتر ═══ */}
      <div className="flex flex-wrap items-center gap-2">
        {([
          { k: 'pending', label: `بحاجة تقييم (${pendingCount})` },
          { k: 'done', label: `تم تقييمها (${doneCount})` },
          { k: 'all', label: `الكل (${bookings.length})` },
        ] as const).map((t) => (
          <button
            key={t.k}
            onClick={() => setFilter(t.k)}
            className={`rounded-lg px-3 py-1.5 text-xs font-bold transition ${
              filter === t.k ? 'bg-[#2c5aad] text-white shadow-md' : 'border border-slate-300 bg-white text-slate-600 hover:bg-slate-50'
            }`}
          >
            {t.label}
          </button>
        ))}
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="🔍 ابحث برقم الحجز أو اسم الزبون"
          className="min-w-[200px] flex-1 rounded-lg border border-slate-300 px-3 py-1.5 text-xs outline-none focus:border-sky-500"
        />
      </div>

      {/* ═══ الحجوزات ═══ */}
      {shown.length === 0 ? (
        <div className="rounded-2xl border border-slate-200 bg-white p-10 text-center">
          <p className="text-3xl">🎉</p>
          <p className="mt-2 text-sm font-bold text-slate-600">
            {filter === 'pending' ? 'ماكو حجز ينتظر تقييم — كلشي مغطّى' : 'ماكو حجوزات بهذا الفلتر'}
          </p>
          {filter === 'pending' && (
            <p className="mt-1 text-xs text-slate-400">تظهر هنا حجوزاتك المنجزة بآخر ٣٠ يوم الي طلع وياك بيها غيرك</p>
          )}
        </div>
      ) : (
        <div className="space-y-4">
          {shown.map((b) => <BookingReviewCard key={b.bookingId} booking={b} onSaved={load} />)}
        </div>
      )}
    </div>
  )
}

/* ───── بطاقة حجز ───── */

function BookingReviewCard({ booking, onSaved }: { booking: BookingAwaitingReview; onSaved: () => void }) {
  const done = booking.crew.every((c) => c.rating)
  const when = booking.completedAt
    ? new Date(booking.completedAt).toLocaleDateString('ar-IQ', { day: 'numeric', month: 'short', year: 'numeric' })
    : '—'

  return (
    <div className={`overflow-hidden rounded-2xl border-2 bg-white shadow-sm ${done ? 'border-emerald-200' : 'border-amber-300'}`}>
      {/* رأس الحجز */}
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-slate-100 bg-slate-50/60 px-4 py-3">
        <div className="flex flex-wrap items-center gap-3">
          <span className="font-mono text-sm font-black text-[#0f2040]">{booking.code}</span>
          <span className="text-xs font-bold text-slate-700">👤 {booking.customerName}</span>
          {booking.serviceName && <span className="text-xs text-slate-500">🔧 {booking.serviceName}</span>}
          <span className="text-xs text-slate-500">📅 {when}</span>
        </div>
        <span className={`rounded-full px-3 py-1 text-[11px] font-bold ${
          done ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-800'
        }`}>
          {done ? '✅ تم التقييم' : `⏳ باقي ${booking.crew.filter((c) => !c.rating).length} من ${booking.crew.length}`}
        </span>
      </div>

      {/* الكادر */}
      <div className="divide-y divide-slate-100">
        {booking.crew.map((c) => (
          <CrewRow key={c.employeeId} bookingId={booking.bookingId} member={c} onSaved={onSaved} />
        ))}
      </div>
    </div>
  )
}

/* ───── سطر موظف ───── */

function CrewRow({ bookingId, member, onSaved }: {
  bookingId: string; member: CrewReviewState; onSaved: () => void
}) {
  const [note, setNote] = useState(member.reason ?? '')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  const save = async (rating: ReviewRating) => {
    setBusy(true); setErr(null)
    try {
      // الإيجابي والتدريب: الملاحظة اختيارية — الحجز نفسه هو السياق،
      // وإجبار الليدر على كتابة نص بعد شغل ميداني يخليه ما يقيّم.
      //
      // ⚠️ أما المخالفة وخلل الالتزام **يشترطون سبب مكتوب**: هذني
      // بلاغات تروح للإدارة وممكن تنبني عليها غرامة. بلاغ بلا سبب
      // ما يقدر أحد يتصرّف بيه، ويظلم الموظف الي انبلّغ عنه.
      const needsReason = rating === 'MISCONDUCT' || rating === 'COMMITMENT'
      if (needsReason && [...note.trim()].length < 5) {
        setErr('اكتب شنو صار بالضبط — البلاغ يروح للإدارة ولازم يفهمونه')
        setBusy(false)
        return
      }
      const label = REVIEW_RATINGS.find((r) => r.value === rating)?.label ?? ''
      const reason = note.trim() || `${label} — بهذا الحجز`
      await api.createPerformanceReview({ employeeId: member.employeeId, rating, reason, bookingId })
      onSaved()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر حفظ التقييم')
    } finally { setBusy(false) }
  }

  return (
    <div className="px-4 py-3">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2.5">
          <span className="flex h-9 w-9 items-center justify-center rounded-full bg-slate-100 text-sm font-black text-slate-500">
            {member.name.charAt(0)}
          </span>
          <div>
            <p className="text-sm font-bold text-slate-800">{member.name}</p>
            {member.position && <p className="text-[11px] text-slate-500">{member.position}</p>}
          </div>
        </div>

        {/* ═══ أربعة أنواع مو اثنين ═══
            «يحتاج تدريب» غير «مخالفة سلوك»: الأول نقص مهارة علاجه
            دورة، والثاني إجراء إداري. خلطهن يخلي صاحب الأسلوب السيّئ
            ينزل بدورة فنية ما تعالج شي، وناقص المهارة ينحسب مخالف.

            ⚠️ متجاوب: أربع أزرار بصف واحد على شاشة موبايل تصير
            مضغوطة ما تنضغط — فتنكسر ٢×٢. */}
        <div className="grid w-full grid-cols-2 gap-1.5 sm:flex sm:w-auto sm:flex-wrap">
          {REVIEW_RATINGS.map((r) => {
            const active = member.rating === r.value
            const tones: Record<string, { on: string; off: string }> = {
              emerald: { on: 'bg-emerald-600 text-white shadow-md', off: 'border-emerald-300 text-emerald-700 hover:bg-emerald-50' },
              amber:   { on: 'bg-amber-600 text-white shadow-md',   off: 'border-amber-300 text-amber-700 hover:bg-amber-50' },
              red:     { on: 'bg-red-600 text-white shadow-md',     off: 'border-red-300 text-red-700 hover:bg-red-50' },
              orange:  { on: 'bg-orange-600 text-white shadow-md',  off: 'border-orange-300 text-orange-700 hover:bg-orange-50' },
            }
            const t = tones[r.tone]
            return (
              <button
                key={r.value}
                onClick={() => save(r.value)}
                disabled={busy}
                title={r.hint}
                className={`rounded-lg px-3 py-2.5 text-[11px] font-bold transition disabled:opacity-50 sm:px-4 sm:py-2 sm:text-xs ${
                  active ? t.on : `border bg-white ${t.off}`
                }`}
              >
                {r.icon} {r.label}
              </button>
            )
          })}
        </div>
      </div>

      <input
        value={note}
        onChange={(e) => setNote(e.target.value)}
        placeholder="ملاحظة (اختيارية) — تنحفظ مع التقييم"
        className="mt-2 w-full rounded-lg border border-slate-200 px-3 py-1.5 text-xs outline-none focus:border-sky-500"
      />

      {member.rating && (
        <p className="mt-1 text-[10px] text-slate-400">
          ✔️ انسجّل — تگدر تغيّره بضغطة على الزر الثاني
        </p>
      )}
      {err && <p className="mt-1 text-[11px] font-bold text-red-600">{err}</p>}
    </div>
  )
}

/* ───── رقم ───── */

function Stat({ icon, label, value, tone }: {
  icon: string; label: string; value: number; tone: 'amber' | 'emerald' | 'violet'
}) {
  const tones: Record<string, string> = {
    amber: 'text-amber-700 bg-amber-50',
    emerald: 'text-emerald-700 bg-emerald-50',
    violet: 'text-violet-700 bg-violet-50',
  }
  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <div className="flex items-start justify-between gap-2">
        <p className="text-[11px] font-medium text-slate-500">{label}</p>
        <span className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-sm ${tones[tone]}`}>{icon}</span>
      </div>
      <p className={`mt-1 text-2xl font-black ${tones[tone].split(' ')[0]}`}>{value}</p>
    </div>
  )
}
