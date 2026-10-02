import { useEffect, useState } from 'react'
import { api, type MatrixProposal } from '../api'
import { GROUP_LABEL, type EyeGroup } from './matrixEyeColors'

// ═══ اقتراحات ماتركس — هو يقترح، والمدير يقرر ═══
// الموافَق يصير جزء منه (تعليمة، تعديل، أو تذكير للموظف)، والمرفوض
// وسببه يصير درس: ما يتكرر، ويندز لهايكو حتى يتعلّم.

const KIND: Record<MatrixProposal['kind'], { icon: string; label: string; approve: string }> = {
  GUIDE_RULE: { icon: '📚', label: 'تعليمة جديدة', approve: '✅ ضيفها لعقله' },
  PREDICTION: { icon: '🔮', label: 'توقّع', approve: '✅ أرسل التذكير للموظف' },
  RULE_TUNE: { icon: '🛠️', label: 'تعديل تعليمة', approve: '✅ طبّق التعديل' },
}

const EV_LABEL: Record<string, string> = {
  done: 'المنجز', left: 'الباقي', hoursElapsed: 'ساعات مضت', hoursRemaining: 'ساعات باقية', pace: 'الوتيرة/ساعة',
  expected: 'المتوقع ينجز', total: 'كل التذكيرات', escalated: 'صعدت', resolved: 'انحلت', rejected: 'رفضتها',
  kind: 'النوع', escalations30d: 'تصعيدات بشهر', group: 'المجموعة', hits: 'مرات الانطباق', lastHitAt: 'آخر انطباق',
}

export default function MatrixProposals({ onChange }: { onChange?: () => void }) {
  const [rows, setRows] = useState<MatrixProposal[] | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [reload, setReload] = useState(0)
  const [busy, setBusy] = useState<string | null>(null)
  const [edit, setEdit] = useState<Record<string, string>>({})
  const [why, setWhy] = useState<string | null>(null)

  useEffect(() => {
    let alive = true
    api.getMatrixProposals('PENDING').then((r) => { if (alive) setRows(r) }).catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر الجلب') })
    return () => { alive = false }
  }, [reload])

  const done = () => { setReload((n) => n + 1); onChange?.() }

  const approve = async (p: MatrixProposal) => {
    setBusy(p.id); setErr(null)
    try {
      const text = edit[p.id]
      const payload = text !== undefined && p.payload ? { ...p.payload, text } : undefined
      await api.approveMatrixProposal(p.id, payload)
      done()
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(null) }
  }
  const reject = async (p: MatrixProposal) => {
    const note = prompt('ليش ترفضه؟ (اختياري — ماتركس يتعلّم من السبب)') ?? null
    if (note === null) return
    setBusy(p.id); setErr(null)
    try { await api.rejectMatrixProposal(p.id, note); done() } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(null) }
  }

  if (rows && rows.length === 0 && !err) {
    return (
      <section className="rounded-2xl border border-slate-200 bg-white p-4 text-sm">
        <h3 className="font-extrabold text-[#0f2040]">💡 اقتراحات ماتركس</h3>
        <p className="mt-1 text-xs text-slate-400">ماكو اقتراحات هسه. ماتركس يتوقع يومياً بعد ١ الظهر، ويراجع نتائجه كل أسبوع.</p>
      </section>
    )
  }

  return (
    <section className="rounded-2xl border border-violet-200 bg-white p-4 text-sm">
      <h3 className="mb-1 font-extrabold text-[#0f2040]">💡 اقتراحات ماتركس <span className="text-xs text-slate-500">({rows?.length ?? '…'})</span></h3>
      <p className="mb-3 text-xs text-slate-500">ماتركس يقترح من الي يشوفه. ولا شي ينفّذ بلا موافقتك، والرفض وسببه يصيرون درس إله.</p>
      {err && <p className="mb-2 rounded-lg bg-red-50 p-2 text-red-600">{err}</p>}
      {!rows ? <p className="text-slate-400">جاري التحميل…</p> : (
        <ul className="space-y-3">
          {rows.map((p) => {
            const k = KIND[p.kind]
            const rule = p.kind === 'GUIDE_RULE' ? (p.payload as { route?: string; match?: string; groups?: string; text?: string; onlyIfPending?: boolean } | null) : null
            return (
              <li key={p.id} className="rounded-xl border border-slate-200 p-3">
                <div className="flex flex-wrap items-center gap-2 text-[11px]">
                  <span className="rounded-full bg-violet-50 px-2 py-0.5 font-bold text-violet-700">{k.icon} {k.label}</span>
                  <span className="text-slate-400">{p.source === 'MODEL' ? '🧠 من هايكو' : '📏 من القواعد'} · {new Date(p.createdAt).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad' })}</span>
                </div>
                <p className="mt-1 font-bold text-slate-800">{p.title}</p>
                {p.rationale && <p className="text-xs text-slate-600">{p.rationale}</p>}
                {rule && (
                  <div className="mt-2 rounded-lg bg-slate-50 p-2">
                    <p className="mb-1 text-[11px] text-slate-500" dir="rtl">
                      على <code dir="ltr">{rule.route}</code>{rule.match && <> · زر: {rule.match}</>} · {rule.groups ? GROUP_LABEL[rule.groups as EyeGroup] ?? rule.groups : 'كل الأدوار'}{rule.onlyIfPending && ' · بس إذا عنده شغل باقي'}
                    </p>
                    <textarea className="w-full rounded-lg border border-slate-300 px-2 py-1 text-sm" rows={2}
                      value={edit[p.id] ?? rule.text ?? ''} onChange={(e) => setEdit({ ...edit, [p.id]: e.target.value })} />
                    <p className="text-[10px] text-slate-400">تگدر تعدّل الكلام قبل الموافقة.</p>
                  </div>
                )}
                {why === p.id && p.evidence && (
                  <dl className="mt-2 grid grid-cols-2 gap-x-3 rounded-lg bg-slate-50 p-2 text-xs text-slate-600 sm:grid-cols-4">
                    {Object.entries(p.evidence).map(([key, v]) => <div key={key}><dt className="font-bold">{EV_LABEL[key] ?? key}</dt><dd>{typeof v === 'object' ? JSON.stringify(v) : String(v)}</dd></div>)}
                  </dl>
                )}
                <div className="mt-2 flex flex-wrap gap-2">
                  <button disabled={busy === p.id} onClick={() => approve(p)} className="rounded-lg bg-emerald-600 px-3 py-1 text-xs font-bold text-white disabled:opacity-50">{k.approve}</button>
                  <button disabled={busy === p.id} onClick={() => reject(p)} className="rounded-lg border border-red-200 px-3 py-1 text-xs font-bold text-red-600 disabled:opacity-50">❌ ارفض</button>
                  {p.evidence && <button onClick={() => setWhy(why === p.id ? null : p.id)} className="text-xs text-brand-700 underline">ليش؟</button>}
                </div>
              </li>
            )
          })}
        </ul>
      )}
    </section>
  )
}
