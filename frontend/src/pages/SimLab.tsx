import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type SimReviewRow } from '../api'
import SimGate from '../sim/SimGate'
import type { SimCategory, SimExercise, SimLesson } from '../sim/types'

// ═══ مختبر المحاكاة — الفهرس ═══
//
// «أريد محاكيات… الموظف من يجي أنطي كورس يدرسه وبعدها يجي يطبّق هنا».
//
// ⚠️ للمالك وحده بهالمرحلة — شوف `sim/SimGate.tsx`.

const DIFFICULTY = ['', 'مبتدئ', 'سهل', 'متوسط', 'متقدم', 'خبير']

export default function SimLab() {
  return <SimGate><Lab /></SimGate>
}

function Lab() {
  const [cats, setCats] = useState<SimCategory[]>([])
  const [openId, setOpenId] = useState<string | null>(null)
  const [exercises, setExercises] = useState<SimExercise[]>([])
  const [lessons, setLessons] = useState<SimLesson[]>([])
  const [loading, setLoading] = useState(true)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    let alive = true
    void (async () => {
      try {
        const rows = await api.getSimCategories()
        if (alive) { setCats(rows); if (rows.length === 1) setOpenId(rows[0].id) }
      } catch (e) {
        if (alive) setErr(e instanceof Error ? e.message : 'تعذر الجلب')
      } finally {
        if (alive) setLoading(false)
      }
    })()
    return () => { alive = false }
  }, [])

  // ⚠️ التصفير جوّا الدالة اللاتزامنية مو بجسم الـeffect: تعديل الحالة
  // مباشرة بالجسم يولّد رندراً متسلسلاً (وهذا الي تمنعه قاعدة
  // react-hooks). و`alive` تمنع سباق الطلبات لمن يبدّل الفئة بسرعة.
  useEffect(() => {
    let alive = true
    void (async () => {
      if (!openId) {
        if (alive) { setExercises([]); setLessons([]) }
        return
      }
      const [ex, les] = await Promise.all([
        api.getSimExercises(openId).catch(() => []),
        api.getSimLessons(openId).catch(() => []),
      ])
      if (alive) { setExercises(ex); setLessons(les) }
    })()
    return () => { alive = false }
  }, [openId])

  return (
    <div dir="rtl" className="space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Link
          to="/simulator-lab/workbench"
          className="rounded-lg bg-brand-700 px-4 py-2 text-sm font-bold text-white shadow hover:bg-brand-800"
        >
          🧰 افتح مساحة العمل
        </Link>
        <h2 className="text-2xl font-bold text-brand-900">🧪 مختبر المحاكاة</h2>
      </div>
      <p className="text-slate-500">
        تدريب عملي: الموظف يطبّق بإيده — يوصّل، يغلط، ويتعلّم ليش غلط.
      </p>

      {/* ⚠️ تنبيه دائم — المختبر تحت البناء وما يوصل أحد غير المالك. */}
      <div className="rounded-xl bg-amber-50 p-4 text-sm leading-relaxed text-amber-900 ring-1 ring-amber-200">
        <b>🔒 تحت البناء — يشوفه المالك وحده.</b>
        <br />
        المحتوى الحالي <b>غير محقّق ميدانياً</b>: مبني على أعراف منشورة عامة مو على
        كتالوگ موديل بعينه. هدفه إثبات المحرّك مو التدريب — وما يوصل أي موظف
        حتى تعتمده بنفسك.
      </div>

      <ReviewPanel />

      {err && <p className="rounded-lg bg-red-50 p-4 text-red-600">{err}</p>}
      {loading && <p className="text-slate-400">جاري التحميل…</p>}
      {!loading && cats.length === 0 && !err && (
        <p className="rounded-xl border border-white bg-white p-8 text-center text-slate-500 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          ماكو فئات بعد.
        </p>
      )}

      {cats.map((c) => (
        <div key={c.id} className="rounded-2xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <button
            onClick={() => setOpenId((v) => (v === c.id ? null : c.id))}
            className="flex w-full items-center justify-between gap-3 p-5 text-right"
          >
            <span className="text-xs text-slate-400">{openId === c.id ? '▲' : '▼'}</span>
            <span className="flex-1">
              <span className="block text-lg font-bold text-brand-900">{c.name}</span>
              {c.description && <span className="mt-0.5 block text-sm text-slate-500">{c.description}</span>}
            </span>
            <span className="shrink-0 rounded-full bg-brand-50 px-3 py-1 text-xs font-bold text-brand-800">
              {c.exerciseCount} تمرين
            </span>
          </button>

          {openId === c.id && (
            <div className="space-y-4 border-t border-slate-100 p-5">
              {lessons.length > 0 && (
                <div>
                  <h3 className="mb-2 text-sm font-bold text-slate-700">📖 الدروس</h3>
                  {lessons.map((l) => (
                    <div key={l.id} className="rounded-xl bg-slate-50 px-4 py-3 text-sm text-slate-700">
                      {l.title}
                    </div>
                  ))}
                </div>
              )}

              <div>
                <h3 className="mb-2 text-sm font-bold text-slate-700">🔧 التمارين</h3>
                {exercises.length === 0 && <p className="text-sm text-slate-400">ماكو تمارين بهاي الفئة.</p>}
                <div className="grid gap-3 sm:grid-cols-2">
                  {exercises.map((e) => (
                    <Link
                      key={e.id}
                      to={`/simulator-lab/exercise/${e.id}`}
                      className="block rounded-xl border border-slate-200 p-4 transition hover:border-brand-500 hover:shadow-md"
                    >
                      <div className="flex items-start justify-between gap-2">
                        <span className="rounded-full bg-slate-100 px-2 py-0.5 text-[10px] font-bold text-slate-600">
                          {DIFFICULTY[e.difficulty] || 'مبتدئ'}
                        </span>
                        <b className="flex-1 text-right text-[15px] text-brand-900">{e.title}</b>
                      </div>
                      {e.brief && <p className="mt-1.5 text-xs leading-relaxed text-slate-500">{e.brief}</p>}
                      <div className="mt-2 flex items-center justify-end gap-2 text-[11px]">
                        {e.passed && <span className="font-bold text-emerald-700">✅ نجحته</span>}
                        {e.bestScore != null && (
                          <span className="text-slate-500">أفضل نتيجة: {e.bestScore}</span>
                        )}
                        {!e.verified && <span className="font-bold text-amber-700">غير محقّق</span>}
                      </div>
                    </Link>
                  ))}
                </div>
              </div>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}

// ═══ لوحة الاعتماد ═══
//
// ⚠️ **بلا هالشاشة، محتوى المختبر ما يوصل موظفاً أبداً.** استعلامات
// المستودع تشترط `verified = TRUE` و`status = 'PUBLISHED'`، وكل
// المحتوى ينزرع `FALSE`. هذا مقصود — المعلومة الغلط عن أسلاك قفل
// تحرق جهازاً، وعن ستring شمسي تحرق إنفرتر.
//
// ⚠️ والاعتماد **ما يفتح المختبر للموظفين لحاله**: `ownerOnly` بشجرة
// القائمة يبقى. فتحه قرار منفصل يتّخذه صاحب النظام صراحةً.
function ReviewPanel() {
  const [rows, setRows] = useState<SimReviewRow[]>([])
  const [open, setOpen] = useState(false)
  const [busy, setBusy] = useState<string | null>(null)
  const [note, setNote] = useState<Record<string, string>>({})
  const [reload, setReload] = useState(0)

  useEffect(() => {
    let alive = true
    void (async () => {
      try {
        const r = await api.getSimReview()
        if (alive) setRows(r)
      } catch { /* المراجعة مو حرجة — الفهرس يشتغل بدونها */ }
    })()
    return () => { alive = false }
  }, [reload])

  const act = useCallback(async (fn: () => Promise<unknown>, id: string) => {
    setBusy(id)
    try { await fn(); setReload((n) => n + 1) }
    catch (e) { alert(e instanceof Error ? e.message : 'تعذر التنفيذ') }
    finally { setBusy(null) }
  }, [])

  const pending = rows.filter((r) => !r.verified)
  const live = rows.filter((r) => r.verified && r.status === 'PUBLISHED')

  return (
    <div className="rounded-2xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
      <button onClick={() => setOpen((v) => !v)} className="flex w-full items-center justify-between gap-3 p-5 text-right">
        <span className="text-xs text-slate-400">{open ? '▲' : '▼'}</span>
        <span className="flex-1">
          <span className="block text-lg font-bold text-brand-900">✅ اعتماد المحتوى</span>
          <span className="mt-0.5 block text-sm text-slate-500">
            {pending.length} بانتظار الاعتماد · {live.length} منشور للموظفين
          </span>
        </span>
      </button>

      {open && (
        <div className="space-y-3 border-t border-slate-100 p-5">
          <p className="rounded-lg bg-amber-50 px-3 py-2 text-[12px] leading-relaxed text-amber-900">
            ⚠️ لا تعتمد شيئاً حتى <b>يجرّبه فني على جهاز حقيقي بالورشة</b>. معلومة غلط عن
            أسلاك قفل تحرق جهازاً، وعن ستring شمسي تحرق إنفرتر. والاعتماد ينسجّل باسمك وتاريخه.
            <br />
            <b>والاعتماد ما يفتح المختبر للموظفين</b> — المجموعة تبقى مخفية عن الكل غيرك حتى تقرر أنت.
          </p>

          {rows.length === 0 && <p className="text-sm text-slate-400">ماكو محتوى.</p>}

          {rows.map((r) => (
            <div key={r.kind + r.id} className="rounded-xl border border-slate-200 p-3">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="flex items-center gap-1.5">
                  {r.verified
                    ? <span className="rounded-md bg-emerald-50 px-2 py-0.5 text-[10px] font-bold text-emerald-700">معتمد</span>
                    : <span className="rounded-md bg-amber-50 px-2 py-0.5 text-[10px] font-bold text-amber-700">غير محقّق</span>}
                  <span className={`rounded-md px-2 py-0.5 text-[10px] font-bold ${
                    r.status === 'PUBLISHED' ? 'bg-sky-50 text-sky-700' : 'bg-slate-100 text-slate-500'}`}>
                    {r.status === 'PUBLISHED' ? 'منشور' : 'مسودّة'}
                  </span>
                </div>
                <b className="flex-1 text-right text-[14px] text-brand-900">{r.title}</b>
              </div>

              {r.sourceRef && <p className="mt-1 text-[11px] leading-relaxed text-slate-500">📖 {r.sourceRef}</p>}
              {r.localPractice && <p className="mt-0.5 text-[11px] leading-relaxed text-slate-500">🇮🇶 {r.localPractice}</p>}

              <div className="mt-2 flex flex-wrap items-center gap-2">
                {!r.verified && (
                  <>
                    <input
                      value={note[r.id] ?? ''} onChange={(e) => setNote((n) => ({ ...n, [r.id]: e.target.value }))}
                      placeholder="شنو جرّبت بالضبط؟ (اختياري)"
                      className="min-w-52 flex-1 rounded-lg border border-slate-300 px-2.5 py-1.5 text-[12px]"
                    />
                    <button
                      disabled={busy === r.id}
                      onClick={() => act(() => api.setSimVerified(r.kind, r.id, true, note[r.id]), r.id)}
                      className="rounded-lg bg-emerald-600 px-3 py-1.5 text-[11px] font-bold text-white disabled:opacity-50"
                    >
                      ✅ اعتمدته — جرّبته على جهاز حقيقي
                    </button>
                  </>
                )}
                {r.verified && r.status !== 'PUBLISHED' && (
                  <button
                    disabled={busy === r.id}
                    onClick={() => act(() => api.setSimPublished(r.kind, r.id, true), r.id)}
                    className="rounded-lg bg-brand-700 px-3 py-1.5 text-[11px] font-bold text-white disabled:opacity-50"
                  >
                    📢 انشره
                  </button>
                )}
                {r.status === 'PUBLISHED' && (
                  <button
                    disabled={busy === r.id}
                    onClick={() => act(() => api.setSimPublished(r.kind, r.id, false), r.id)}
                    className="rounded-lg bg-slate-100 px-3 py-1.5 text-[11px] font-bold text-slate-600 disabled:opacity-50"
                  >
                    رجّعه مسودّة
                  </button>
                )}
                {r.verified && (
                  <button
                    disabled={busy === r.id}
                    onClick={() => act(() => api.setSimVerified(r.kind, r.id, false), r.id)}
                    className="rounded-lg bg-white px-3 py-1.5 text-[11px] font-bold text-red-600 ring-1 ring-red-200 disabled:opacity-50"
                  >
                    اسحب الاعتماد
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
