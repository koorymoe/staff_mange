// ═══ أكاديمية المختبر ═══
//
// «هذا المختبر لا تلتزم بي بألوان الشركة — فقط المختبر، أريده هيچ».
//
// ⚠️ **هوية بصرية مستقلة عمداً.** باقي النظام فاتح بألوان الشركة،
// والمختبر داكن بلونه. وهذا مو تناقضاً — هو **فصل مقصود**: الموظف
// لمن يدخل هنا يعرف إنه ترك شغل اليوم ودخل بيئة تدريب. نفس الي
// تسويه الشركات بمنصات تدريبها.
//
// ⚠️⚠️ **كل رقم بهالشاشة محسوب من بيانات حقيقية.** ماكو ولا نسبة
// ولا شريط تقدّم مكتوب بالكود. لوحة تدريب تعرض «٦٨٪» ثابتة تخرب
// ثقة المتدرّب بكل شي ثاني بالشاشة — ولو ماكو بيانات، تنعرض
// الحالة الفارغة صريحة.

import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type SimProject, type SimReviewRow } from '../api'
import SimGate from '../sim/SimGate'
import type { SimAttempt, SimCategory, SimExercise } from '../sim/types'

const DIFFICULTY = ['', 'مبتدئ', 'سهل', 'متوسط', 'متقدم', 'خبير']

/** لون ورمز لكل فئة — يتعرّفون من اسمها.
 *
 *  ⚠️ ما ننحفظ بقاعدة البيانات: الفئة محتوى، والشكل عرض. ربطهم يعني
 *  ترحيلاً كل ما نغيّر لوناً. */
function look(name: string): { icon: string; from: string; to: string; ring: string } {
  if (/شبك|سويچ|راوتر/.test(name)) return { icon: '🌐', from: 'from-sky-500/25', to: 'to-sky-900/10', ring: 'ring-sky-500/30' }
  if (/شمس|طاقة/.test(name)) return { icon: '☀️', from: 'from-amber-500/25', to: 'to-amber-900/10', ring: 'ring-amber-500/30' }
  if (/قفل|أقفال|دخول/.test(name)) return { icon: '🔐', from: 'from-violet-500/25', to: 'to-violet-900/10', ring: 'ring-violet-500/30' }
  if (/حريق|إنذار/.test(name)) return { icon: '🔥', from: 'from-red-500/25', to: 'to-red-900/10', ring: 'ring-red-500/30' }
  if (/صوت|إذاع/.test(name)) return { icon: '🔊', from: 'from-fuchsia-500/25', to: 'to-fuchsia-900/10', ring: 'ring-fuchsia-500/30' }
  if (/gps|تتبّع|تتبع/i.test(name)) return { icon: '🛰️', from: 'from-emerald-500/25', to: 'to-emerald-900/10', ring: 'ring-emerald-500/30' }
  if (/كهرب|دوائر/.test(name)) return { icon: '⚡', from: 'from-yellow-500/25', to: 'to-yellow-900/10', ring: 'ring-yellow-500/30' }
  return { icon: '🧪', from: 'from-slate-500/25', to: 'to-slate-900/10', ring: 'ring-slate-500/30' }
}

export default function SimLab() {
  return <SimGate><Academy /></SimGate>
}

interface CatData {
  cat: SimCategory
  exercises: SimExercise[]
  passed: number
  total: number
}

function Academy() {
  const [cats, setCats] = useState<CatData[]>([])
  const [attempts, setAttempts] = useState<SimAttempt[]>([])
  const [projects, setProjects] = useState<SimProject[]>([])
  const [review, setReview] = useState<SimReviewRow[]>([])
  const [loading, setLoading] = useState(true)
  const [err, setErr] = useState<string | null>(null)
  const [openId, setOpenId] = useState<string | null>(null)
  const [reload, setReload] = useState(0)

  useEffect(() => {
    let alive = true
    void (async () => {
      try {
        const rows = await api.getSimCategories()
        // ⚠️ تمارين كل الفئات بالتوازي: الشاشة تعرض تقدّم **كل** برنامج
        // مو الي انفتح بس، فتحميلها بالتسلسل يعني انتظاراً بلا سبب.
        const withEx = await Promise.all(
          rows.map(async (c) => {
            const ex = await api.getSimExercises(c.id).catch(() => [] as SimExercise[])
            return { cat: c, exercises: ex, passed: ex.filter((e) => e.passed).length, total: ex.length }
          }),
        )
        const [att, prj, rev] = await Promise.all([
          api.getMySimAttempts(50).catch(() => [] as SimAttempt[]),
          api.listSimProjects().catch(() => [] as SimProject[]),
          api.getSimReview().catch(() => [] as SimReviewRow[]),
        ])
        if (!alive) return
        setCats(withEx)
        setAttempts(att)
        setProjects(prj)
        setReview(rev)
      } catch (e) {
        if (alive) setErr(e instanceof Error ? e.message : 'تعذر الجلب')
      } finally {
        if (alive) setLoading(false)
      }
    })()
    return () => { alive = false }
  }, [reload])

  const totals = useMemo(() => {
    const total = cats.reduce((s, c) => s + c.total, 0)
    const passed = cats.reduce((s, c) => s + c.passed, 0)
    return { total, passed, pct: total ? Math.round((passed / total) * 100) : 0 }
  }, [cats])

  const verified = review.filter((r) => r.verified).length
  const published = review.filter((r) => r.status === 'PUBLISHED').length

  return (
    <div dir="rtl" className="-m-3 min-h-screen bg-[#070b14] p-3 text-slate-200 md:-m-6 md:p-6">
      {/* ═══ البانر ═══ */}
      <div className="relative overflow-hidden rounded-3xl bg-gradient-to-l from-[#0d1830] via-[#0b1424] to-[#0a1020] p-6 ring-1 ring-slate-800 md:p-9">
        {/* توهّج خلفي — طبقتان بلا صور. */}
        <div className="pointer-events-none absolute -left-24 -top-24 h-72 w-72 rounded-full bg-sky-500/10 blur-3xl" />
        <div className="pointer-events-none absolute -bottom-32 right-10 h-72 w-72 rounded-full bg-emerald-500/10 blur-3xl" />

        <p className="relative text-[11px] font-bold tracking-wide text-amber-400/90">أكاديمية التدريب المتقدّمة</p>
        <h1 className="relative mt-1.5 text-3xl font-black text-white md:text-4xl">مختبر التدريب والمحاكاة</h1>
        <p className="relative mt-2 max-w-2xl text-sm leading-relaxed text-slate-400">
          تعلّم · درّب · طبّق. الموظف يشتغل بإيده على أجهزة تتصرّف مثل الحقيقية،
          يغلط، ويشوف عاقبة غلطته قبل ما يوصل الميدان.
        </p>

        {/* ⚠️ الشارات تصف **الي موجود فعلاً** — ماكو «شهادات معتمدة»
            لأننا ما نصدر شهادات، والوعد الي ما ينفّذ يخرب المصداقية. */}
        <div className="relative mt-4 flex flex-wrap gap-2">
          {[
            ['🧊', 'محاكاة ثلاثية الأبعاد'],
            ['🖥️', 'سطر أوامر حقيقي'],
            ['📐', 'محرّكات تحسب فعلاً'],
            ['🎯', 'تقييم من السيرفر'],
          ].map(([i, t]) => (
            <span key={t} className="rounded-full bg-white/5 px-3 py-1.5 text-[11.5px] font-bold text-slate-300 ring-1 ring-white/10">
              {i} {t}
            </span>
          ))}
        </div>

        <Link
          to="/simulator-lab/workbench"
          className="relative mt-5 inline-flex items-center gap-2 rounded-xl bg-sky-600 px-5 py-2.5 text-sm font-bold text-white shadow-lg shadow-sky-900/40 transition hover:bg-sky-500"
        >
          ⚡ افتح استوديو المحاكاة
        </Link>
      </div>

      {/* ⚠️ تنبيه دائم — نفس قاعدة بقية المختبر. */}
      <div className="mt-4 rounded-2xl bg-amber-500/10 p-4 text-[12.5px] leading-relaxed text-amber-200 ring-1 ring-amber-500/25">
        <b>🔒 تحت البناء — يشوفه المالك وحده.</b> المحتوى مبني على أعراف منشورة
        عامة مو على كتالوگ موديل بعينه، و<b>ما يوصل أي موظف</b> حتى تجرّبه على جهاز
        حقيقي وتعتمده من لوحة الاعتماد.
        {review.length > 0 && (
          <span className="mt-1 block font-normal">
            الحالة: <b>{verified}</b> معتمد · <b>{published}</b> منشور · <b>{review.length - verified}</b> ينتظر تجربتك.
          </span>
        )}
      </div>

      {err && <p className="mt-4 rounded-xl bg-red-500/10 p-4 text-red-300 ring-1 ring-red-500/25">{err}</p>}
      {loading && <p className="mt-6 text-slate-500">جاري التحميل…</p>}

      {!loading && (
        <div className="mt-4 grid gap-4 xl:grid-cols-[17rem_1fr]">
          {/* ═══ العمود الجانبي ═══ */}
          <div className="space-y-4">
            {/* مسار التعلّم */}
            <div className="rounded-2xl bg-[#0d1421] p-5 ring-1 ring-slate-800">
              <h3 className="text-[13px] font-bold text-slate-200">مسارك بالمختبر</h3>
              <div className="mt-3 flex items-center gap-4">
                <Ring pct={totals.pct} />
                <div className="text-[12px] leading-relaxed text-slate-400">
                  <p><b className="text-lg text-white">{totals.passed}</b> من {totals.total} تمرين</p>
                  <p className="mt-0.5">{attempts.length} محاولة مسجّلة</p>
                </div>
              </div>
              {totals.total === 0 && (
                <p className="mt-3 text-[11px] text-slate-500">ماكو تمارين بعد.</p>
              )}
            </div>

            {/* مستوى المهارات — شريط لكل برنامج */}
            <div className="rounded-2xl bg-[#0d1421] p-5 ring-1 ring-slate-800">
              <h3 className="mb-3 text-[13px] font-bold text-slate-200">مستواك بكل برنامج</h3>
              <div className="space-y-3">
                {cats.map((c) => {
                  const pct = c.total ? Math.round((c.passed / c.total) * 100) : 0
                  const lk = look(c.cat.name)
                  return (
                    <div key={c.cat.id}>
                      <div className="mb-1 flex items-center justify-between text-[11px]">
                        <span className="font-mono tabular-nums text-slate-500">{c.passed}/{c.total}</span>
                        <span className="truncate font-bold text-slate-300">{lk.icon} {c.cat.name.split('—')[0].trim()}</span>
                      </div>
                      <div className="h-1.5 overflow-hidden rounded-full bg-slate-800">
                        <div className="h-full rounded-full bg-gradient-to-l from-sky-400 to-emerald-400 transition-all"
                          style={{ width: `${pct}%` }} />
                      </div>
                    </div>
                  )
                })}
                {cats.length === 0 && <p className="text-[11px] text-slate-500">ماكو برامج بعد.</p>}
              </div>
            </div>

            {/* المخططات المحفوظة */}
            <div className="rounded-2xl bg-[#0d1421] p-5 ring-1 ring-slate-800">
              <div className="mb-2 flex items-center justify-between">
                <span className="text-[10px] text-slate-600">{projects.length}</span>
                <h3 className="text-[13px] font-bold text-slate-200">مخططاتك المحفوظة</h3>
              </div>
              {projects.length === 0 && (
                <p className="text-[11px] leading-relaxed text-slate-500">
                  ماكو مخططات. ابنِ واحداً بالاستوديو واحفظه — يبقى محفوظاً بجلسة الكونسول وياه.
                </p>
              )}
              <div className="space-y-1.5">
                {projects.slice(0, 5).map((pr) => (
                  <Link key={pr.id} to="/simulator-lab/workbench"
                    className="block rounded-lg bg-[#0a101c] px-3 py-2 ring-1 ring-slate-800 transition hover:ring-sky-600">
                    <p className="truncate text-[12px] font-bold text-slate-200">{pr.name}</p>
                    <p className="text-[10px] text-slate-500">{new Date(pr.updatedAt).toLocaleDateString('ar-IQ')}</p>
                  </Link>
                ))}
              </div>
            </div>
          </div>

          {/* ═══ البرامج ═══ */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-[11px] text-slate-500">{cats.length} برنامج</span>
              <h2 className="text-lg font-bold text-white">برامج التدريب</h2>
            </div>

            <div className="grid gap-3 md:grid-cols-2 2xl:grid-cols-3">
              {cats.map((c) => {
                const lk = look(c.cat.name)
                const pct = c.total ? Math.round((c.passed / c.total) * 100) : 0
                const open = openId === c.cat.id
                return (
                  <div key={c.cat.id}
                    className={`overflow-hidden rounded-2xl bg-[#0d1421] ring-1 transition ${open ? 'ring-sky-600' : 'ring-slate-800 hover:ring-slate-700'}`}>
                    <button onClick={() => setOpenId(open ? null : c.cat.id)} className="w-full text-right">
                      {/* الرأس — تدرّج بدل صورة */}
                      <div className={`relative flex h-24 items-center justify-center bg-gradient-to-bl ${lk.from} ${lk.to}`}>
                        <span className="text-4xl opacity-90">{lk.icon}</span>
                        <span className={`absolute left-3 top-3 rounded-full bg-black/40 px-2 py-0.5 text-[10px] font-bold text-slate-300 ring-1 ${lk.ring}`}>
                          {c.total} تمرين
                        </span>
                      </div>

                      <div className="p-4">
                        <h3 className="text-[15px] font-bold text-white">{c.cat.name.split('—')[0].trim()}</h3>
                        {c.cat.description && (
                          <p className="mt-1 line-clamp-2 text-[11.5px] leading-relaxed text-slate-500">{c.cat.description}</p>
                        )}
                        <div className="mt-3 flex items-center gap-2">
                          <span className="font-mono text-[11px] tabular-nums text-slate-500">{c.passed}/{c.total}</span>
                          <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-slate-800">
                            <div className="h-full rounded-full bg-gradient-to-l from-sky-400 to-emerald-400"
                              style={{ width: `${pct}%` }} />
                          </div>
                          <span className="text-[11px] font-bold text-slate-400">{pct}٪</span>
                        </div>
                      </div>
                    </button>

                    {open && (
                      <div className="space-y-1.5 border-t border-slate-800 p-3">
                        {c.exercises.length === 0 && <p className="text-[11px] text-slate-500">ماكو تمارين بهالبرنامج.</p>}
                        {c.exercises.map((e) => (
                          <Link key={e.id} to={`/simulator-lab/exercise/${e.id}`}
                            className="block rounded-xl bg-[#0a101c] p-3 ring-1 ring-slate-800 transition hover:ring-sky-600">
                            <div className="flex items-start justify-between gap-2">
                              <span className="shrink-0 rounded-md bg-slate-800 px-2 py-0.5 text-[10px] font-bold text-slate-400">
                                {DIFFICULTY[e.difficulty] || 'مبتدئ'}
                              </span>
                              <b className="flex-1 text-right text-[13px] text-slate-100">{e.title}</b>
                            </div>
                            {e.brief && <p className="mt-1 text-[11px] leading-relaxed text-slate-500">{e.brief}</p>}
                            <div className="mt-1.5 flex flex-wrap items-center justify-end gap-2 text-[10.5px]">
                              <span className="rounded bg-slate-800/70 px-1.5 py-0.5 font-mono text-slate-500">{e.engineKind}</span>
                              {!e.verified && <span className="font-bold text-amber-400">غير محقّق</span>}
                              {e.bestScore != null && <span className="text-slate-400">أفضل: {e.bestScore}</span>}
                              {e.passed && <span className="font-bold text-emerald-400">✅ نجحته</span>}
                            </div>
                          </Link>
                        ))}
                      </div>
                    )}
                  </div>
                )
              })}
            </div>

            {/* إجراءات سريعة */}
            <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
              {[
                { to: '/simulator-lab/workbench', icon: '🧰', t: 'استوديو المحاكاة', s: 'ابنِ مخططاً وشغّله' },
                { to: '/simulator-lab/workbench', icon: '🖥️', t: 'كونسول الأجهزة', s: 'هيّئ سويچاً بسطر الأوامر' },
              ].map((a, i) => (
                <Link key={i} to={a.to}
                  className="flex items-center gap-3 rounded-2xl bg-[#0d1421] p-4 ring-1 ring-slate-800 transition hover:ring-sky-600">
                  <span className="text-2xl">{a.icon}</span>
                  <span>
                    <span className="block text-[13px] font-bold text-slate-100">{a.t}</span>
                    <span className="block text-[11px] text-slate-500">{a.s}</span>
                  </span>
                </Link>
              ))}
              <ReviewCard rows={review} onDone={() => setReload((n) => n + 1)} />
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

/** حلقة تقدّم — تنرسم بـSVG، والنسبة **محسوبة** مو مكتوبة. */
function Ring({ pct }: { pct: number }) {
  const r = 26
  const c = 2 * Math.PI * r
  return (
    <svg viewBox="0 0 64 64" className="h-16 w-16 shrink-0 -rotate-90">
      <circle cx="32" cy="32" r={r} fill="none" stroke="#1e293b" strokeWidth={7} />
      <circle cx="32" cy="32" r={r} fill="none" stroke="url(#g)" strokeWidth={7} strokeLinecap="round"
        strokeDasharray={`${(c * pct) / 100} ${c}`} />
      <defs>
        <linearGradient id="g" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stopColor="#38bdf8" />
          <stop offset="100%" stopColor="#4ade80" />
        </linearGradient>
      </defs>
      <text x="32" y="32" textAnchor="middle" dominantBaseline="central" className="rotate-90"
        style={{ transformOrigin: '32px 32px' }} fill="#e2e8f0" fontSize="15" fontWeight="bold">{pct}٪</text>
    </svg>
  )
}

/** ═══ بطاقة الاعتماد ═══
 *
 *  ⚠️ **بلا هالإجراء، محتوى المختبر ما يوصل موظفاً أبداً**: استعلامات
 *  المستودع تشترط `verified = TRUE` و`status = 'PUBLISHED'`، وكل
 *  المحتوى ينزرع `FALSE`. */
function ReviewCard({ rows, onDone }: { rows: SimReviewRow[]; onDone: () => void }) {
  const [open, setOpen] = useState(false)
  const [busy, setBusy] = useState<string | null>(null)
  const [note, setNote] = useState<Record<string, string>>({})
  const pending = rows.filter((r) => !r.verified)

  const act = useCallback(async (fn: () => Promise<unknown>, id: string) => {
    setBusy(id)
    try { await fn(); onDone() }
    catch (e) { alert(e instanceof Error ? e.message : 'تعذر التنفيذ') }
    finally { setBusy(null) }
  }, [onDone])

  return (
    <>
      <button onClick={() => setOpen((v) => !v)}
        className="flex items-center gap-3 rounded-2xl bg-[#0d1421] p-4 text-right ring-1 ring-slate-800 transition hover:ring-emerald-600">
        <span className="text-2xl">✅</span>
        <span>
          <span className="block text-[13px] font-bold text-slate-100">اعتماد المحتوى</span>
          <span className="block text-[11px] text-slate-500">{pending.length} ينتظر تجربتك على جهاز حقيقي</span>
        </span>
      </button>

      {open && (
        <div className="col-span-full space-y-2 rounded-2xl bg-[#0d1421] p-4 ring-1 ring-slate-800">
          <p className="rounded-lg bg-amber-500/10 px-3 py-2 text-[11.5px] leading-relaxed text-amber-200">
            ⚠️ لا تعتمد شيئاً حتى <b>يجرّبه فني على جهاز حقيقي بالورشة</b>. الاعتماد ينسجّل
            باسمك وتاريخه. <b>والاعتماد ما يفتح المختبر للموظفين</b> — يبقى مخفياً عن الكل غيرك حتى تقرر أنت.
          </p>
          {rows.map((r) => (
            <div key={r.kind + r.id} className="rounded-xl bg-[#0a101c] p-3 ring-1 ring-slate-800">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="flex gap-1.5">
                  <span className={`rounded px-2 py-0.5 text-[10px] font-bold ${
                    r.verified ? 'bg-emerald-500/15 text-emerald-300' : 'bg-amber-500/15 text-amber-300'}`}>
                    {r.verified ? 'معتمد' : 'غير محقّق'}
                  </span>
                  <span className={`rounded px-2 py-0.5 text-[10px] font-bold ${
                    r.status === 'PUBLISHED' ? 'bg-sky-500/15 text-sky-300' : 'bg-slate-800 text-slate-500'}`}>
                    {r.status === 'PUBLISHED' ? 'منشور' : 'مسودّة'}
                  </span>
                </div>
                <b className="flex-1 text-right text-[13px] text-slate-100">{r.title}</b>
              </div>
              {r.sourceRef && <p className="mt-1 text-[10.5px] leading-relaxed text-slate-500">📖 {r.sourceRef}</p>}
              <div className="mt-2 flex flex-wrap items-center gap-2">
                {!r.verified && (
                  <>
                    <input
                      value={note[r.id] ?? ''} onChange={(e) => setNote((n) => ({ ...n, [r.id]: e.target.value }))}
                      placeholder="شنو جرّبت بالضبط؟"
                      className="min-w-44 flex-1 rounded-lg border border-slate-700 bg-[#070b14] px-2.5 py-1.5 text-[11.5px] text-slate-200 placeholder:text-slate-600"
                    />
                    <button disabled={busy === r.id}
                      onClick={() => act(() => api.setSimVerified(r.kind, r.id, true, note[r.id]), r.id)}
                      className="rounded-lg bg-emerald-600 px-3 py-1.5 text-[11px] font-bold text-white disabled:opacity-50">
                      ✅ اعتمدته
                    </button>
                  </>
                )}
                {r.verified && r.status !== 'PUBLISHED' && (
                  <button disabled={busy === r.id}
                    onClick={() => act(() => api.setSimPublished(r.kind, r.id, true), r.id)}
                    className="rounded-lg bg-sky-600 px-3 py-1.5 text-[11px] font-bold text-white disabled:opacity-50">
                    📢 انشره
                  </button>
                )}
                {r.status === 'PUBLISHED' && (
                  <button disabled={busy === r.id}
                    onClick={() => act(() => api.setSimPublished(r.kind, r.id, false), r.id)}
                    className="rounded-lg bg-slate-800 px-3 py-1.5 text-[11px] font-bold text-slate-300 disabled:opacity-50">
                    رجّعه مسودّة
                  </button>
                )}
                {r.verified && (
                  <button disabled={busy === r.id}
                    onClick={() => act(() => api.setSimVerified(r.kind, r.id, false), r.id)}
                    className="rounded-lg bg-red-500/10 px-3 py-1.5 text-[11px] font-bold text-red-300 ring-1 ring-red-500/25 disabled:opacity-50">
                    اسحب الاعتماد
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </>
  )
}
