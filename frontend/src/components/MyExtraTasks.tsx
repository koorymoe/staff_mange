import { useCallback, useEffect, useState } from 'react'
import { api, type ExtraTask } from '../api'

// ═══ مهامي الإضافية — شاشة الموظف ═══
//
// «هاي راح تطلع يم منو؟ يم الموظف الي انتوجّه له العمل».
//
// تنعرض بشاشة «مهامي» فوق مهام الحجوزات — الشغل الموجّه من المدير
// أولوية، ولو انحطت بأسفل الصفحة راح تنتنسى.
//
// ⚠️ المعرّف يجي من التوكن بالسيرفر، ما ينرسل من هنا: بدون هذا أي
// موظف يقرا مهام غيره بتبديل رقم بالرابط.

export default function MyExtraTasks() {
  const [tasks, setTasks] = useState<ExtraTask[]>([])
  const [showDone, setShowDone] = useState(false)
  const [busy, setBusy] = useState<string | null>(null)
  const [doneFor, setDoneFor] = useState<ExtraTask | null>(null)
  const [doneNote, setDoneNote] = useState('')
  const [err, setErr] = useState<string | null>(null)

  // ⚠️ نجيب **الكل** مرة وحدة ونفلتر محلياً، مو نطلب حسب showDone.
  //
  // انمسك بالفحص: لما تكون كل المهام منجزة، طلب «المفتوحة بس» يرجّع
  // فاضي فالصندوق كله يختفي — ووية اختفائه يختفي مربّع «وريني
  // المنجزة»، فالموظف ما عاد يكدر يوصل لتاريخ شغله أبداً.
  //
  // وعدد مهام الموظف صغير (السيرفر يحدّها بـ٣٠٠)، فطلب واحد أخف من
  // طلب جديد كل ما يأشّر المربّع.
  const load = useCallback(() => {
    api.getMyExtraTasks(true).then(setTasks).catch(() => setTasks([]))
  }, [])

  useEffect(load, [load])

  // نأشّر «شافها» أول ما تنعرض — يقطع نقاش «ما وصلني».
  // ⚠️ للجديدة بس: إعادة التأشير على كل تحميل تدوس على الوقت الأول.
  useEffect(() => {
    tasks.filter((t) => t.status === 'NEW' && !t.seenAt)
      .forEach((t) => { void api.markExtraTaskSeen(t.id).catch(() => undefined) })
  }, [tasks])

  const start = async (t: ExtraTask) => {
    setBusy(t.id)
    try { await api.startExtraTask(t.id); load() }
    catch (e) { alert(e instanceof Error ? e.message : 'تعذر البدء') }
    finally { setBusy(null) }
  }

  const complete = async () => {
    if (!doneFor) return
    setErr(null)
    // ⚠️ العدّ بالحروف مو بالبايتات — العربي حرفه بايتين وأكثر.
    if ([...doneNote.trim()].length < 5) {
      setErr('اكتب شنو سويت بالضبط — «تم» ما تكفي')
      return
    }
    setBusy(doneFor.id)
    try {
      await api.completeExtraTask(doneFor.id, doneNote.trim())
      setDoneFor(null); setDoneNote('')
      load()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الإنهاء')
    } finally { setBusy(null) }
  }

  const openTasks = tasks.filter((t) => t.status === 'NEW' || t.status === 'IN_PROGRESS')
  const shown = showDone ? tasks : openTasks
  // الموظف الي ما انوجّهت له ولا مهمة أبداً ما يشوف الصندوق أصلاً —
  // بس الي عنده تاريخ يبقى يوصله حتى لو كله منجز.
  if (tasks.length === 0) return null

  return (
    <div dir="rtl" className="mb-4 rounded-2xl border-2 border-brand-200 bg-brand-50/40 p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h3 className="font-bold text-[#0f2040]">
          📋 مهام موجّهة إلك {openTasks.length > 0 && <span className="rounded-full bg-brand-600 px-2 py-0.5 text-xs text-white">{openTasks.length}</span>}
        </h3>
        <label className="flex items-center gap-1.5 text-xs font-bold text-slate-600">
          <input type="checkbox" checked={showDone} onChange={(e) => setShowDone(e.target.checked)} />
          وريني المنجزة بعد
        </label>
      </div>

      {shown.length === 0 && <p className="mt-2 text-sm text-slate-400">ماكو مهام مفتوحة — أشّر «وريني المنجزة» حتى تشوف الي خلّصته.</p>}

      <div className="mt-3 space-y-2">
        {shown.map((t) => (
          <div key={t.id} className={`rounded-xl border bg-white p-3 ${t.overdue ? 'border-red-300' : 'border-slate-200'}`}>
            <div className="flex flex-wrap items-center gap-2">
              {t.priority === 'URGENT' && <span className="rounded-full bg-red-100 px-2 py-0.5 text-[11px] font-bold text-red-700">🔴 مستعجلة</span>}
              {t.overdue && <span className="rounded-full bg-red-600 px-2 py-0.5 text-[11px] font-bold text-white">⏰ فات موعدها</span>}
              <b className="text-sm text-[#0f2040]">{t.title}</b>
            </div>

            {t.description && <p className="mt-1 text-sm text-slate-600">{t.description}</p>}

            <div className="mt-1.5 flex flex-wrap gap-x-3 gap-y-1 text-[11px] text-slate-500">
              {t.assignedByName && <span>📤 من: <b>{t.assignedByName}</b></span>}
              {t.dueAt && <span>⏳ آخر موعد: {new Date(t.dueAt).toLocaleString('en-GB')}</span>}
            </div>

            {t.status === 'DONE' && t.doneNote && (
              <p className="mt-2 rounded-lg bg-emerald-50 px-3 py-2 text-xs text-emerald-900">✅ {t.doneNote}</p>
            )}
            {t.status === 'CANCELLED' && (
              <p className="mt-2 rounded-lg bg-slate-100 px-3 py-2 text-xs text-slate-700">✖️ انلغت — {t.cancelReason}</p>
            )}

            {(t.status === 'NEW' || t.status === 'IN_PROGRESS') && (
              <div className="mt-2 flex flex-wrap gap-2">
                {t.status === 'NEW' && (
                  <button
                    onClick={() => start(t)}
                    disabled={busy === t.id}
                    className="rounded-lg border border-brand-300 px-3 py-1.5 text-xs font-bold text-brand-700 disabled:opacity-50"
                  >
                    ▶ بديت بيها
                  </button>
                )}
                <button
                  onClick={() => { setDoneFor(t); setDoneNote(''); setErr(null) }}
                  className="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-bold text-white"
                >
                  ✓ خلّصتها
                </button>
              </div>
            )}
          </div>
        ))}
      </div>

      {/* ⚠️ الإنجاز يطلب وصف: «تم» بلا شرح ترجّع المدير يسأل بالتلفون
          — وهاي نفس المكالمة الي بنينا الميزة حتى نلغيها. */}
      {doneFor && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => setDoneFor(null)}>
          <div dir="rtl" className="w-full max-w-md rounded-2xl bg-white p-5" onClick={(e) => e.stopPropagation()}>
            <h3 className="font-bold text-[#0f2040]">✓ إنهاء: {doneFor.title}</h3>
            <label className="mb-1 mt-3 block text-xs font-bold text-slate-500">
              شنو سويت بالضبط؟ <span className="text-red-600">*</span>
            </label>
            <textarea
              value={doneNote}
              onChange={(e) => setDoneNote(e.target.value)}
              rows={3}
              placeholder="مثال: خرّجت ٤٢ فاتورة للشهر ورفعتها بالنظام المحاسبي"
              className="w-full resize-y rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
            />
            {err && <p className="mt-2 rounded-lg bg-red-50 px-3 py-2 text-xs font-bold text-red-700">{err}</p>}
            <div className="mt-3 flex gap-2">
              <button
                onClick={complete}
                disabled={busy !== null}
                className="flex-1 rounded-xl bg-emerald-600 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50"
              >
                {busy ? 'جاري الحفظ...' : 'خلّصتها'}
              </button>
              <button onClick={() => setDoneFor(null)} className="rounded-xl border border-slate-300 px-4 py-2.5 text-sm font-bold text-slate-600">
                رجوع
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
