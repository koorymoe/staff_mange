import { useEffect, useState, useCallback, useMemo } from 'react'
import { matches } from '../utils/search'
import { useSession } from '../session'
import EmptyState from '../components/EmptyState'
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
// ═══ النجوم الثلاث ═══
//
// الحكم لوحده (إيجابي / يحتاج تدريب) يجاوب «زين لو لا؟» بس، وما يگول
// **وين**. فني يوصل متأخر وشغله ممتاز، وفني يوصل بوقته وشغله ناقص —
// الاثنين ينكتبون بنفس الخانة.
//
// فانضافت ثلاث درجات: الالتزام، السرعة، جودة التنفيذ.
//
// ⚠️ وكلهن **اختيارية**. التصميم الأول كان يطلب ١٢ ضغطة لتقييم ثلاثة
// فنيين بحجز واحد — والليدر راجع من شغل ميداني وبيده تلفون، فالشاشة
// الي تطلب منه ١٢ ضغطة ما ينستخدمها، وينتهي بينا لصفر تقييمات بدل
// تقييمات ناقصة. الحكم بضغطة وحدة يكفي، والنجوم زيادة للي عنده وقت.

type Draft = {
  rating: ReviewRating | null
  note: string
  commitment: number | null
  speed: number | null
  quality: number | null
}

const emptyDraft = (c: CrewReviewState): Draft => ({
  rating: c.rating,
  note: c.reason ?? '',
  commitment: c.commitmentScore,
  speed: c.speedScore,
  quality: c.qualityScore,
})

export default function PerformanceReviewPage({ embedded }: { embedded?: boolean } = {}) {
  const [bookings, setBookings] = useState<BookingAwaitingReview[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState<'all' | 'pending' | 'done'>('all')
  const [search, setSearch] = useState('')
  const [fromDate, setFromDate] = useState('')
  const [toDate, setToDate] = useState('')
  const [service, setService] = useState('')
  const [err, setErr] = useState<string | null>(null)
  const { employee, permissions } = useSession()
  // نفس تفريع الخادم (performance_review_service.go): هذولا يشوفون كل
  // حجوزات الشركة، والليدر يشوف حجوزاته هو. الشاشة لازم تحچي بلسان
  // النطاق الصحيح وإلا التلميحات تصير كذب على المراقب.
  const seesAll =
    employee?.role === 'ADMIN' ||
    employee?.actualRole === 'OWNER' ||
    employee?.role === 'MONITOR' ||
    permissions.includes('performance_review')

  const load = useCallback(() => {
    api.getMyBookingsForReview(fromDate || undefined, toDate || undefined)
      .then(setBookings)
      .catch((e) => setErr(e instanceof Error ? e.message : 'تعذر جلب الحجوزات'))
      .finally(() => setLoading(false))
  }, [fromDate, toDate])
  // تبديل التاريخ يعيد الجلب من الخادم — الافتراضي عنده آخر ٣٠ يوم،
  // وأي فترة يختارها المستخدم تنزل للاستعلام نفسه.
  useEffect(() => { load() }, [load])

  const isDone = (b: BookingAwaitingReview) => b.crew.length > 0 && b.crew.every((c) => c.rating)

  // ── الأرقام الأربعة ──
  const stats = useMemo(() => {
    const now = new Date()
    const startOfDay = new Date(now.getFullYear(), now.getMonth(), now.getDate())
    const weekAgo = new Date(startOfDay.getTime() - 7 * 86400000)
    const doneOn = (b: BookingAwaitingReview, since: Date) =>
      isDone(b) && b.completedAt && new Date(b.completedAt) >= since
    return {
      pending: bookings.filter((b) => !isDone(b)).length,
      doneToday: bookings.filter((b) => doneOn(b, startOfDay)).length,
      waitingPeople: bookings.reduce((n, b) => n + b.crew.filter((c) => !c.rating).length, 0),
      weekReviews: bookings
        .filter((b) => b.completedAt && new Date(b.completedAt) >= weekAgo)
        .reduce((n, b) => n + b.crew.filter((c) => c.rating).length, 0),
    }
  }, [bookings])

  const services = useMemo(
    () => Array.from(new Set(bookings.map((b) => b.serviceName).filter(Boolean) as string[])).sort(),
    [bookings],
  )

  const shown = bookings
    .filter((b) => filter === 'all' || (filter === 'done' ? isDone(b) : !isDone(b)))
    .filter((b) => (service ? b.serviceName === service : true))
    .filter((b) => {
      const q = search.trim()
      if (!q) return true
      return matches([b.code, b.customerName, b.serviceName], q)
    })

  if (loading) return <p className="p-6 text-center text-slate-400">جارٍ التحميل...</p>

  return (
    <div dir="rtl" className="mx-auto max-w-6xl space-y-4">
      {/* ═══ العنوان ═══
          ⚠️ يختفي لمّن الشاشة مضمَّنة بشاشة أخرى (مثلاً «التقييم» أو
          مكتب إدارة الموظفين) — وإلا ترويستان فوگ بعض. */}
      {!embedded && (
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-violet-50 text-lg sm:h-11 sm:w-11 sm:text-xl">⭐</span>
          <div className="min-w-0">
            <h1 className="text-xl font-black text-[#0f2040] sm:text-2xl">تقييم الأداء</h1>
            <p className="text-[11px] text-slate-500 sm:text-xs">قيّم الفريق المرتبط بكل حجز بعد إنجازه</p>
          </div>
        </div>
        <p className="rounded-xl border border-slate-200 bg-white px-3 py-1.5 text-[10px] leading-relaxed text-slate-500 sm:text-[11px]">
          ⓘ يتم تقييم الموظفين ضمن الحجز نفسه وليس كتقييم دائم مستقل
        </p>
      </div>
      )}

      {err && (
        <div className="rounded-xl border-2 border-red-300 bg-red-50 px-4 py-3 text-sm font-bold text-red-700">{err}</div>
      )}

      {/* ═══ الأرقام ═══ */}
      <div className="grid grid-cols-2 gap-2.5 sm:gap-3 lg:grid-cols-4">
        <Stat icon="📋" label="حجوزات بحاجة تقييم" hint="مكتملة بانتظار التقييم" value={stats.pending} tone="amber" />
        <Stat icon="✅" label="تم تقييمها اليوم" hint="حتى الآن" value={stats.doneToday} tone="emerald" />
        <Stat icon="👥" label="موظفون بانتظار التقييم" hint={seesAll ? "بكل الحجوزات" : "موزّعين على حجوزاتك"} value={stats.waitingPeople} tone="violet" />
        <Stat icon="📊" label="تقييمات هذا الأسبوع" hint="خلال آخر ٧ أيام" value={stats.weekReviews} tone="sky" />
      </div>

      {/* ═══ التصفية ═══ */}
      <div className="rounded-2xl border border-slate-200 bg-white p-3 shadow-[0_2px_12px_rgba(15,32,64,0.05)]">
        <div className="grid grid-cols-1 gap-2.5 lg:grid-cols-4">
          <div className="lg:col-span-1">
            <label className="mb-1 block text-[10px] font-bold text-slate-500">بحث</label>
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="🔍 ابحث برقم الحجز أو اسم العميل..."
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-sky-500"
            />
          </div>

          <div>
            <label className="mb-1 block text-[10px] font-bold text-slate-500">تصفية الحالة</label>
            <div className="flex gap-1">
              {([
                { k: 'all', label: `كل الحجوزات` },
                { k: 'pending', label: `بحاجة تقييم` },
                { k: 'done', label: `تم تقييمها` },
              ] as const).map((t) => (
                <button
                  key={t.k}
                  onClick={() => setFilter(t.k)}
                  className={`flex-1 rounded-lg px-2 py-2 text-[10px] font-bold transition sm:text-[11px] ${
                    filter === t.k ? 'bg-[#2c5aad] text-white shadow-md' : 'border border-slate-300 bg-white text-slate-600 hover:bg-slate-50'
                  }`}
                >
                  {t.label}
                </button>
              ))}
            </div>
          </div>

          <div>
            <label className="mb-1 block text-[10px] font-bold text-slate-500">تاريخ الإنجاز</label>
            <div className="flex items-center gap-1">
              <input
                type="date" value={fromDate} onChange={(e) => setFromDate(e.target.value)}
                className="w-full rounded-lg border border-slate-300 px-2 py-2 text-[11px] outline-none focus:border-sky-500"
              />
              <span className="text-[10px] text-slate-400">إلى</span>
              <input
                type="date" value={toDate} onChange={(e) => setToDate(e.target.value)}
                className="w-full rounded-lg border border-slate-300 px-2 py-2 text-[11px] outline-none focus:border-sky-500"
              />
            </div>
          </div>

          <div>
            <label className="mb-1 block text-[10px] font-bold text-slate-500">تصفية الخدمة</label>
            <select
              value={service}
              onChange={(e) => setService(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-sky-500"
            >
              <option value="">كل الخدمات</option>
              {services.map((s) => <option key={s} value={s}>{s}</option>)}
            </select>
          </div>
        </div>
      </div>

      {/* ═══ الحجوزات ═══ */}
      {shown.length === 0 ? (
        <div className="rounded-2xl border border-slate-200 bg-white">
          {/* ⚠️ حالة الفراغ لازم تكول **السبب**. چانت تكول «ماكو حجوزات
              بهذي التصفية» حتى لمن تكون كل المرشّحات فارغة، فالمستخدم
              يدوّر بالمرشّحات والعلّة بمكان ثاني تماماً. */}
          <EmptyState
            icon={bookings.length === 0 ? '🗂️' : '🎉'}
            title={
              bookings.length === 0
                ? 'ماكو حجوزات منجزة بهذي الفترة'
                : filter === 'pending'
                  ? 'ماكو حجز ينتظر تقييم — كلشي مغطّى'
                  : 'ماكو حجز يطابق البحث أو التصفية'
            }
            reason={
              bookings.length === 0
                ? (fromDate || toDate
                    ? 'جرّب توسّع «تاريخ الإنجاز» — الفترة المختارة ماكو بيها حجز منجز.'
                    : seesAll
                      ? 'تُعرض الحجوزات المنجزة بآخر ٣٠ يوم افتراضياً. وسّع «تاريخ الإنجاز» حتى تشوف أقدم.'
                      : 'تظهر هنا حجوزاتك المنجزة بآخر ٣٠ يوم الي طلع وياك بيها غيرك. إذا ما كنت مكلّف بحجز منجز، الصفحة تبقى فارغة.')
                : 'المرشّحات فوك هي الي تخفي الباقي — صفّرها حتى تشوف الكل.'
            }
          />
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
  const done = booking.crew.length > 0 && booking.crew.every((c) => c.rating)
  const when = booking.completedAt
    ? new Date(booking.completedAt).toLocaleDateString('ar-IQ', { day: 'numeric', month: 'long', year: 'numeric' })
    : '—'
  const time = booking.completedAt
    ? new Date(booking.completedAt).toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit' })
    : ''

  // المسوّدات محلية لكل الكادر — الحفظ بضغطة وحدة بآخر البطاقة.
  //
  // ⚠️ ليش مو حفظ فوري بكل ضغطة؟ لأن التقييم الكامل بالتصميم فيه
  // حكم + ثلاث نجوم + ملاحظة: الحفظ الفوري يعني ست نداءات لموظف
  // واحد، وكل وحدة ممكن تفشل بنص الطريق فتترك تقييماً نص محفوظ.
  const [drafts, setDrafts] = useState<Record<string, Draft>>(
    () => Object.fromEntries(booking.crew.map((c) => [c.employeeId, emptyDraft(c)])),
  )
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [okMsg, setOkMsg] = useState(false)
  const [expanded, setExpanded] = useState(!done)

  const patch = (id: string, p: Partial<Draft>) =>
    setDrafts((prev) => ({ ...prev, [id]: { ...prev[id], ...p } }))

  const saveAll = async () => {
    setErr(null); setOkMsg(false)
    const toSave = booking.crew.filter((c) => drafts[c.employeeId]?.rating)
    if (toSave.length === 0) {
      setErr('اختار تقييم لموظف واحد على الأقل')
      return
    }
    // المخالفة وخلل الالتزام **يشترطون سبب مكتوب**: هذني بلاغات
    // تروح للإدارة وممكن تنبني عليها غرامة. بلاغ بلا سبب ما يقدر أحد
    // يتصرّف بيه، ويظلم الموظف الي انبلّغ عنه.
    for (const c of toSave) {
      const d = drafts[c.employeeId]
      const needsReason = d.rating === 'MISCONDUCT' || d.rating === 'COMMITMENT'
      if (needsReason && [...d.note.trim()].length < 5) {
        setErr(`«${c.name}»: اكتب شنو صار بالضبط — البلاغ يروح للإدارة ولازم يفهمونه`)
        return
      }
    }

    setBusy(true)
    try {
      for (const c of toSave) {
        const d = drafts[c.employeeId]
        const label = REVIEW_RATINGS.find((r) => r.value === d.rating)?.label ?? ''
        await api.createPerformanceReview({
          employeeId: c.employeeId,
          rating: d.rating as ReviewRating,
          reason: d.note.trim() || `${label} — بهذا الحجز`,
          bookingId: booking.bookingId,
          commitmentScore: d.commitment,
          speedScore: d.speed,
          qualityScore: d.quality,
        })
      }
      setOkMsg(true)
      onSaved()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر حفظ التقييم')
    } finally { setBusy(false) }
  }

  // متوسطات البطاقة المنجزة
  const avg = (pick: (c: CrewReviewState) => number | null) => {
    const vals = booking.crew.map(pick).filter((v): v is number => v != null)
    if (vals.length === 0) return null
    return (vals.reduce((a, b) => a + b, 0) / vals.length).toFixed(1)
  }
  const avgQuality = avg((c) => c.qualityScore)
  const avgSpeed = avg((c) => c.speedScore)
  const avgCommitment = avg((c) => c.commitmentScore)

  return (
    <div className={`overflow-hidden rounded-2xl border-2 bg-white shadow-sm ${done ? 'border-emerald-200' : 'border-amber-300'}`}>
      {/* ── رأس الحجز ── */}
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-slate-100 bg-slate-50/60 px-3 py-3 sm:px-4">
        <div className="flex flex-wrap items-center gap-2 sm:gap-3">
          <span className="text-[10px] text-slate-400">رقم الحجز</span>
          <span className="font-mono text-sm font-black text-[#0f2040]">{booking.code}</span>
          {booking.serviceName && <span className="text-xs text-slate-600">🔧 {booking.serviceName}</span>}
          <span className="text-xs text-slate-500">📅 {when}{time && ` · ${time}`}</span>
          <span className="rounded-lg bg-emerald-50 px-2 py-0.5 text-[10px] font-bold text-emerald-700">مكتمل</span>
        </div>
        <span className={`rounded-full px-3 py-1 text-[11px] font-bold ${
          done ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-800'
        }`}>
          {done ? '✅ تم التقييم' : `⏳ بانتظار التقييم — باقي ${booking.crew.filter((c) => !c.rating).length} من ${booking.crew.length}`}
        </span>
      </div>

      {/* ── هوية الزبون ──
          ⚠️ مو زينة: الليدر يتذكّر الشغلة من العنوان أسرع بكثير ما
          يتذكرها من كود الحجز، وتقييم بلا ما يتذكر الشغلة تقييم
          مخمّن. */}
      <div className="grid grid-cols-1 gap-2 border-b border-slate-100 px-3 py-2.5 text-[11px] sm:grid-cols-3 sm:px-4">
        <Field icon="👤" label="اسم العميل" value={booking.customerName} />
        <Field icon="📞" label="الجوال" value={booking.customerPhone || '—'} />
        <Field icon="📍" label="العنوان" value={booking.customerAddress || '—'} />
      </div>

      {done && !expanded ? (
        /* ── ملخّص البطاقة المنجزة ──
            انقيّمت خلاص، فما تحتاج تاخذ نص الشاشة. تنطوي على سطر
            متوسطات، والليدر يفتحها لو راد يعدّل. */
        <div className="flex flex-wrap items-center justify-between gap-3 px-3 py-3 sm:px-4">
          <div>
            <p className="text-xs font-bold text-[#0f2040]">📊 ملخص التقييم</p>
            <p className="text-[10px] text-slate-400">تم تقييم {booking.crew.length} من {booking.crew.length} موظف</p>
          </div>
          <div className="flex flex-wrap gap-2">
            <AvgChip label="متوسط جودة التنفيذ" value={avgQuality} />
            <AvgChip label="متوسط السرعة" value={avgSpeed} />
            <AvgChip label="متوسط الالتزام" value={avgCommitment} />
          </div>
          <button
            onClick={() => setExpanded(true)}
            className="rounded-xl border border-slate-300 px-3 py-1.5 text-[11px] font-bold text-slate-600 hover:bg-slate-50"
          >
            👁 عرض تفاصيل التقييم
          </button>
        </div>
      ) : (
        <>
          <div className="px-3 pt-3 sm:px-4">
            <p className="text-xs font-bold text-slate-600">
              👥 الفريق المشارك في هذا الحجز ({booking.crew.length})
            </p>
          </div>

          <div className="grid grid-cols-1 gap-3 p-3 sm:p-4 xl:grid-cols-2">
            {booking.crew.map((c) => (
              <CrewCard
                key={c.employeeId}
                member={c}
                draft={drafts[c.employeeId]}
                onChange={(p) => patch(c.employeeId, p)}
              />
            ))}
          </div>

          {err && <p className="px-4 pb-2 text-[11px] font-bold text-red-600">{err}</p>}
          {okMsg && <p className="px-4 pb-2 text-[11px] font-bold text-emerald-700">✅ انحفظ التقييم</p>}

          <div className="border-t border-slate-100 px-3 pb-3 pt-3 sm:px-4">
            <button
              onClick={saveAll}
              disabled={busy}
              className="w-full rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2.5 text-sm font-bold text-white shadow-md disabled:opacity-50"
            >
              {busy ? 'جاري الحفظ...' : '💾 حفظ تقييم هذا الحجز'}
            </button>
            <p className="mt-1.5 text-center text-[10px] text-slate-400">
              لن يتم مشاركة التقييم مع العميل
            </p>
          </div>
        </>
      )}
    </div>
  )
}

/* ───── بطاقة موظف ───── */

function CrewCard({ member, draft, onChange }: {
  member: CrewReviewState; draft: Draft; onChange: (p: Partial<Draft>) => void
}) {
  return (
    <div className="rounded-2xl border border-slate-200 bg-slate-50/40 p-3">
      <div className="flex items-center gap-2.5 border-b border-slate-200 pb-2.5">
        <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-white text-sm font-black text-slate-500 shadow-sm">
          {member.name.charAt(0)}
        </span>
        <div className="min-w-0">
          <p className="truncate text-sm font-bold text-slate-800">{member.name}</p>
          {member.position && <p className="text-[10px] text-slate-500">{member.position}</p>}
        </div>
        {member.rating && (
          <span className="mr-auto shrink-0 rounded-lg bg-emerald-50 px-2 py-0.5 text-[10px] font-bold text-emerald-700">
            انقيّم
          </span>
        )}
      </div>

      {/* ── النجوم الثلاث (اختيارية) ── */}
      <div className="mt-2.5 space-y-1.5">
        <StarRow label="الالتزام" value={draft.commitment} onPick={(v) => onChange({ commitment: v })} />
        <StarRow label="السرعة" value={draft.speed} onPick={(v) => onChange({ speed: v })} />
        <StarRow label="جودة التنفيذ" value={draft.quality} onPick={(v) => onChange({ quality: v })} />
      </div>

      {/* ═══ أربعة أنواع مو اثنين ═══
          «يحتاج تدريب» غير «مخالفة سلوك»: الأول نقص مهارة علاجه دورة،
          والثاني إجراء إداري. خلطهن يخلي صاحب الأسلوب السيّئ ينزل
          بدورة فنية ما تعالج شي، وناقص المهارة ينحسب مخالف.

          ⚠️ متجاوب: أربع أزرار بصف واحد على شاشة موبايل تصير مضغوطة
          ما تنضغط — فتنكسر ٢×٢. */}
      <p className="mt-2.5 text-[10px] font-bold text-slate-500">تقييم الأداء في هذا الحجز</p>
      <div className="mt-1 grid grid-cols-2 gap-1.5">
        {REVIEW_RATINGS.map((r) => {
          const active = draft.rating === r.value
          const tones: Record<string, { on: string; off: string }> = {
            emerald: { on: 'bg-emerald-700 text-white shadow-md', off: 'border-emerald-300 text-emerald-700 hover:bg-emerald-50' },
            amber:   { on: 'bg-amber-600 text-white shadow-md',   off: 'border-amber-300 text-amber-700 hover:bg-amber-50' },
            red:     { on: 'bg-red-600 text-white shadow-md',     off: 'border-red-300 text-red-700 hover:bg-red-50' },
            orange:  { on: 'bg-orange-600 text-white shadow-md',  off: 'border-orange-300 text-orange-700 hover:bg-orange-50' },
          }
          const t = tones[r.tone]
          return (
            <button
              key={r.value}
              onClick={() => onChange({ rating: active ? null : r.value })}
              title={r.hint}
              className={`rounded-lg px-2 py-2 text-[10px] font-bold transition sm:text-[11px] ${
                active ? t.on : `border bg-white ${t.off}`
              }`}
            >
              {r.icon} {r.label}
            </button>
          )
        })}
      </div>

      <input
        value={draft.note}
        onChange={(e) => onChange({ note: e.target.value })}
        placeholder="ملاحظة مرتبطة بهذا الحجز (اختياري)"
        className="mt-2 w-full rounded-lg border border-slate-200 px-3 py-1.5 text-[11px] outline-none focus:border-sky-500"
      />
      {(draft.rating === 'MISCONDUCT' || draft.rating === 'COMMITMENT') && (
        <p className="mt-1 text-[10px] font-bold text-amber-700">
          ⚠️ هذا بلاغ يروح للإدارة — السبب المكتوب إجباري.
        </p>
      )}
    </div>
  )
}

/* ───── نجوم ─────
   ⚠️ الضغط على نفس النجمة يلغيها: الليدر يضغط بالغلط وما يكون عنده
   طريقة يرجع لـ«ما نطّيت درجة» — فيضطر ينطي رقم ما يقصده، والمتوسط
   يطلع كذب. */
function StarRow({ label, value, onPick }: {
  label: string; value: number | null; onPick: (v: number | null) => void
}) {
  return (
    <div className="flex items-center justify-between gap-2">
      <span className="text-[10px] text-slate-500">{label}</span>
      <div className="flex gap-0.5">
        {[1, 2, 3, 4, 5].map((n) => (
          <button
            key={n}
            onClick={() => onPick(value === n ? null : n)}
            aria-label={`${label} ${n} من ٥`}
            className={`text-sm leading-none transition ${
              value != null && n <= value ? 'text-amber-400' : 'text-slate-300 hover:text-amber-200'
            }`}
          >
            ★
          </button>
        ))}
      </div>
    </div>
  )
}

/* ───── قطع صغيرة ───── */

function Field({ icon, label, value }: { icon: string; label: string; value: string }) {
  return (
    <div className="min-w-0">
      <p className="text-[10px] text-slate-400">{icon} {label}</p>
      <p className="truncate font-bold text-slate-700">{value}</p>
    </div>
  )
}

function AvgChip({ label, value }: { label: string; value: string | null }) {
  // ⚠️ الي ما انطّاه نجوم ما يطلع «0.0/5» — يطلع «—». صفر يعني تقييم
  // سيّئ، وفرق بينه وبين «ما انقيّم» فرق يظلم موظف.
  return (
    <div className="rounded-xl border border-slate-200 bg-white px-2.5 py-1.5 text-center">
      <p className="text-[9px] text-slate-400">{label}</p>
      <p className="text-[11px] font-black text-slate-800">
        {value ? <>{value}<span className="text-[9px] font-normal text-slate-400">/5</span> ⭐</> : '—'}
      </p>
    </div>
  )
}

function Stat({ icon, label, hint, value, tone }: {
  icon: string; label: string; hint: string; value: number; tone: 'amber' | 'emerald' | 'violet' | 'sky'
}) {
  const tones: Record<string, string> = {
    amber: 'text-amber-700 bg-amber-50',
    emerald: 'text-emerald-700 bg-emerald-50',
    violet: 'text-violet-700 bg-violet-50',
    sky: 'text-sky-700 bg-sky-50',
  }
  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-3 shadow-sm sm:p-4">
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <p className="truncate text-[10px] font-medium text-slate-500 sm:text-[11px]">{label}</p>
          <p className={`mt-1 text-2xl font-black sm:text-3xl ${tones[tone].split(' ')[0]}`}>{value}</p>
          <p className="text-[10px] text-slate-400">{hint}</p>
        </div>
        <span className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-sm ${tones[tone]}`}>{icon}</span>
      </div>
    </div>
  )
}
