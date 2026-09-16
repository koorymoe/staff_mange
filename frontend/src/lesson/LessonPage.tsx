// ═══ صفحة الدرس: شرح ← تطبيق ← وين الغلط ═══
//
// ⚠️⚠️ **الشرح والتطبيق مو بنفس الشاشة.** لو انعرضوا سوا، المتدرّب
// يبدي يحرّك قطعاً قبل ما يقرا — ونرجع لنفس المحاكي الي يعلّم
// بالتخمين. الفصل هو الي يجبر ترتيب «افهم بعدين طبّق».
//
// ⚠️ والرجوع للشرح **مفتوح دائماً**: يمنعك ترجع تقرا يعني تحفيظ مو
// تعليم، والفني بالميدان عنده الكتالوگ بجيبه.

import { useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'

import { Bench } from '../lab/LabWorkbench'
import type { LabDoc, SimResult } from '../lab/types'
import { LESSON_BY_ID } from './lessons'
import { stepForError } from './types'

export default function LessonPage() {
  const { id } = useParams()
  const lesson = id ? LESSON_BY_ID[id] : undefined
  const [phase, setPhase] = useState<'learn' | 'practice'>('learn')
  const [focusStep, setFocusStep] = useState<number | null>(null)
  const [hintsOpen, setHintsOpen] = useState(0)
  const [result, setResult] = useState<SimResult | null>(null)
  /** ⚠️ عدد القطع وقت آخر تشغيل — لازم للتفريق بين «سليم» و«فاضي». */
  const [built, setBuilt] = useState(0)

  const startDoc: LabDoc | undefined = useMemo(
    () => (lesson ? { domain: lesson.domain, nodes: [], links: [] } : undefined),
    [lesson],
  )

  if (!lesson) {
    return (
      <div className="p-8 text-center text-slate-400">
        ماكو درس بهذا المعرّف. <Link to="/simulator-lab" className="text-sky-400 hover:underline">رجوع للمختبر</Link>
      </div>
    )
  }

  // ═══ ربط الأخطاء بخطوات الدرس ═══
  //
  // ⚠️ الرسالة الي ما تطابق أي خطوة تبقى **بمجموعتها** — ما ننسبها
  // لأقرب خطوة. نسبة غلط لخطوة ما تخصّه تخلّي المتدرّب يعيد قراءة
  // شرح صحيح وهو يدوّر على غلط بمكان ثانٍ.
  const errs = (result?.messages ?? []).filter((m) => m.kind === 'error')
  const warns = (result?.messages ?? []).filter((m) => m.kind === 'warn')
  const mapped = [...errs, ...warns].map((m) => ({ ...m, step: stepForError(lesson, m.text) }))
  const explained = mapped.filter((m) => m.step !== null)
  const unexplained = mapped.filter((m) => m.step === null)
  // ⚠️ **لوح فاضي مو منظومة سليمة.** بلا هالشرط، أول تشغيل على لوح
  // شبه فاضي يطلع «✅ المنظومة تشتغل» — والمتدرّب يظن خلّص التمرين
  // وهو ما بنى شي. مؤشر أخضر كاذب أسوأ من ماكو مؤشر.
  const tooEmpty = built < 2
  const clean = !!result && !tooEmpty && errs.length === 0

  // ─────────────────────────── الشرح ───────────────────────────
  if (phase === 'learn') {
    return (
      <div className="mx-auto max-w-4xl p-4 md:p-8">
        <Link to="/simulator-lab" className="text-[12px] text-slate-500 hover:text-slate-300">← كل الدروس</Link>

        <div className="mt-3 rounded-3xl bg-gradient-to-l from-[#0d1830] via-[#0b1424] to-[#0a1020] p-6 ring-1 ring-slate-800 md:p-8">
          <div className="flex flex-wrap items-center gap-2 text-[11px]">
            <span className="rounded-full bg-sky-500/15 px-2.5 py-1 font-bold text-sky-300 ring-1 ring-sky-500/30">
              درس · {lesson.minutes} دقيقة
            </span>
            <span className="text-slate-500">{lesson.summary}</span>
          </div>
          <h1 className="mt-3 text-3xl font-extrabold text-white">{lesson.title}</h1>

          {/* ⚠️ «ليش» قبل «شنو»: درس يبدي بتعريف يُقرا وينُسى. */}
          <div className="mt-4 rounded-2xl bg-amber-500/10 p-4 ring-1 ring-amber-500/25">
            <p className="mb-1 text-[11px] font-bold text-amber-300">ليش هذا يهمّك بشغلنا</p>
            <p className="text-[13px] leading-relaxed text-amber-100/90">{lesson.why}</p>
          </div>
        </div>

        <div className="mt-5 space-y-3">
          {lesson.steps.map((s, i) => (
            <div
              key={i}
              id={`step-${i}`}
              className={`rounded-2xl bg-[#0e1626] p-5 ring-1 transition ${
                focusStep === i ? 'ring-2 ring-amber-400' : 'ring-slate-800'}`}
            >
              <div className="mb-2 flex items-start gap-3">
                <span className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-sky-600 text-[12px] font-bold text-white">
                  {i + 1}
                </span>
                <h3 className="text-[15px] font-bold text-slate-100">{s.title}</h3>
              </div>
              <p className="whitespace-pre-line pr-10 text-[13px] leading-[1.9] text-slate-300">{s.body}</p>
              {s.todo && (
                <p className="mt-3 mr-10 rounded-xl bg-emerald-500/10 px-3 py-2 text-[12px] font-bold text-emerald-200 ring-1 ring-emerald-500/25">
                  ✋ بالتطبيق: {s.todo}
                </p>
              )}
            </div>
          ))}
        </div>

        {/* ═══ الأعطال الشائعة ═══
            ⚠️ **بالعَرَض أولاً مو بالسبب**: الفني يوصله عَرَض، ويدوّر
            بذاكرته على «شفت هذا قبل». جدول مرتّب بالأسباب ما ينفع
            وقت البلاغ. */}
        <div className="mt-6 rounded-2xl bg-[#0e1626] p-5 ring-1 ring-slate-800">
          <h3 className="mb-3 text-[14px] font-bold text-slate-100">🔧 شنو تشوفه بالميدان وشنو سببه</h3>
          <div className="space-y-2">
            {lesson.pitfalls.map((p, i) => (
              <div key={i} className="flex flex-wrap items-start gap-2 rounded-xl bg-[#0b1220] p-3 ring-1 ring-slate-800">
                <span className="rounded-lg bg-red-500/10 px-2 py-1 text-[12px] font-bold text-red-300 ring-1 ring-red-500/25">
                  {p.symptom}
                </span>
                <span className="pt-1 text-slate-600">←</span>
                <span className="pt-1 text-[12.5px] text-slate-300">{p.cause}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="mt-6 rounded-2xl bg-[#0e1626] p-5 ring-1 ring-slate-800">
          <p className="mb-1 text-[11px] font-bold text-slate-500">المطلوب بالتطبيق</p>
          <p className="text-[13.5px] font-bold leading-relaxed text-slate-100">{lesson.task}</p>
          <button
            onClick={() => { setPhase('practice'); setFocusStep(null) }}
            className="mt-4 w-full rounded-xl bg-emerald-600 px-6 py-3 text-[14px] font-bold text-white shadow-lg hover:bg-emerald-500"
          >
            قريت الشرح — أبدي التطبيق ←
          </button>
        </div>
      </div>
    )
  }

  // ─────────────────────────── التطبيق ───────────────────────────
  return (
    <div className="p-3 md:p-5">
      <div className="mb-3 flex flex-wrap items-center gap-2 rounded-2xl bg-[#0e1626] px-4 py-3 ring-1 ring-slate-800">
        <button
          onClick={() => setPhase('learn')}
          className="rounded-lg bg-slate-800 px-3 py-1.5 text-[12px] font-bold text-slate-300 hover:bg-slate-700"
        >
          ← ارجع للشرح
        </button>
        <div className="min-w-0 flex-1">
          <p className="text-[10.5px] font-bold text-slate-500">المطلوب</p>
          <p className="truncate text-[13px] font-bold text-slate-100">{lesson.task}</p>
        </div>
        {/* ⚠️ التلميحات **وحدة وحدة** بطلب المتدرّب: عرضها كلها من
            البداية يخلّي التمرين نسخاً، وإخفاؤها كلياً يخلّي الي
            يوگف يوگف. */}
        {hintsOpen < lesson.hints.length && (
          <button
            onClick={() => setHintsOpen((n) => n + 1)}
            className="rounded-lg bg-slate-800 px-3 py-1.5 text-[12px] font-bold text-amber-300 hover:bg-slate-700"
          >
            💡 تلميح ({hintsOpen}/{lesson.hints.length})
          </button>
        )}
      </div>

      {hintsOpen > 0 && (
        <div className="mb-3 space-y-1.5">
          {lesson.hints.slice(0, hintsOpen).map((h, i) => (
            <p key={i} className="rounded-xl bg-amber-500/10 px-4 py-2 text-[12.5px] text-amber-100 ring-1 ring-amber-500/25">
              💡 {h}
            </p>
          ))}
        </div>
      )}

      {/* ═══ وين الغلط ═══
          ⚠️ يظهر **بس بعد التشغيل**: لوح فاضي كل أخطاؤه «ما بنيت شي»،
          وعرضها من البداية يخوّف المتدرّب قبل ما يبدي. */}
      {result && (
        <div className={`mb-3 rounded-2xl p-4 ring-1 ${
          clean ? 'bg-emerald-500/10 ring-emerald-500/30'
            : tooEmpty ? 'bg-slate-500/10 ring-slate-600/40' : 'bg-red-500/10 ring-red-500/30'}`}>
          {tooEmpty ? (
            <p className="text-[13.5px] font-bold text-slate-300">
              اللوح لسه فاضي — حط القطع واربطها، بعدها شغّل.
            </p>
          ) : clean ? (
            <p className="text-[13.5px] font-bold text-emerald-200">
              ✅ ماكو أخطاء خطيرة — المنظومة تشتغل. {warns.length > 0 && `(${warns.length} تحذير باقي)`}
            </p>
          ) : (
            <>
              <p className="mb-2 text-[13px] font-bold text-red-200">
                ⛔ {errs.length} خطأ — وهذا وين طاح كل واحد بالشرح:
              </p>
              <div className="space-y-2">
                {explained.map((m, i) => (
                  <div key={i} className="rounded-xl bg-[#0b1220] p-3 ring-1 ring-slate-800">
                    <p className="text-[12.5px] leading-relaxed text-slate-200">{m.text}</p>
                    <button
                      onClick={() => { setFocusStep(m.step!); setPhase('learn') }}
                      className="mt-2 rounded-lg bg-amber-500/15 px-2.5 py-1 text-[11.5px] font-bold text-amber-300 ring-1 ring-amber-500/30 hover:bg-amber-500/25"
                    >
                      📖 هذا شرحناه بالخطوة {m.step! + 1}: {lesson.steps[m.step!].title} ←
                    </button>
                  </div>
                ))}
                {/* ⚠️ خطأ ما نعرف خطوته يبقى **معروضاً بصراحة** — ما
                    ننسبه لأقرب خطوة حتى تبين القائمة كاملة. */}
                {unexplained.map((m, i) => (
                  <div key={`u${i}`} className="rounded-xl bg-[#0b1220] p-3 ring-1 ring-slate-800">
                    <p className="text-[12.5px] leading-relaxed text-slate-300">{m.text}</p>
                    <p className="mt-1 text-[11px] text-slate-600">هذا برّا نطاق الدرس — بس لازم تصلّحه.</p>
                  </div>
                ))}
              </div>
            </>
          )}
        </div>
      )}

      <Bench embedded startDoc={startDoc} onResult={(d, r) => { setResult(r); setBuilt(d.nodes.length) }} />
    </div>
  )
}
