import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type MatrixProposal } from '../api'
import { GROUP_LABEL, type EyeGroup } from './matrixEyeColors'
import { EV_LABEL, SEV, category, evidenceCards, severity } from './matrixEvidence'

// ═══ اقتراحات ماتركس — هو يقترح، والمدير يقرر ═══
// الموافَق يصير جزء منه (تعليمة، تعديل، أو تذكير للموظف)، والمرفوض
// وسببه يصير درس: ما يتكرر، ويندز لهايكو حتى يتعلّم.

const KIND: Record<MatrixProposal['kind'], { icon: string; label: string; approve: string }> = {
  GUIDE_RULE: { icon: '📚', label: 'تعليمة جديدة', approve: '✅ ضيفها لعقله' },
  PREDICTION: { icon: '🔮', label: 'توقّع', approve: '✅ أرسل التذكير للموظف' },
  RULE_TUNE: { icon: '🛠️', label: 'تعديل تعليمة', approve: '✅ طبّق التعديل' },
}


function ago(at: string) {
  const m = Math.max(0, Math.round((Date.now() - new Date(at).getTime()) / 60000))
  return m < 60 ? `منذ ${m} دقيقة` : m < 1440 ? `منذ ${Math.round(m / 60)} ساعات` : `منذ ${Math.round(m / 1440)} يوم`
}
const initials = (n: string) => n.trim().split(/\s+/).slice(0, 2).map((w) => w[0]).join(' ')

function Stat({ icon, value, label, sub, tone }: { icon: string; value: number | string; label: string; sub: string; tone: string }) {
  return (
    <div className={`flex items-center justify-between gap-2 rounded-xl border p-3 ${tone}`}>
      <div><p className="text-xs font-bold opacity-80">{label}</p><b className="text-2xl">{value}</b><p className="text-[10px] opacity-70">{sub}</p></div>
      <span className="grid h-10 w-10 place-items-center rounded-xl bg-white/70 text-lg">{icon}</span>
    </div>
  )
}

export default function MatrixProposals({ onChange }: { onChange?: () => void }) {
  const [rows, setRows] = useState<MatrixProposal[] | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [reload, setReload] = useState(0)
  const [busy, setBusy] = useState<string | null>(null)
  const [edit, setEdit] = useState<Record<string, string>>({})
  const [why, setWhy] = useState<string | null>(null)

  const [approvedToday, setApprovedToday] = useState(0)
  useEffect(() => {
    const today = new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })
    api.getMatrixProposals('APPROVED').then((r) => setApprovedToday(r.filter((p) => p.decidedAt && new Date(p.decidedAt).toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' }) === today).length)).catch(() => {})
  }, [reload])
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

  const urgent = (rows ?? []).filter((p) => severity(p) === 'URGENT').length
  return (
    <section className="rounded-2xl border border-slate-200 bg-white p-4 text-sm">
      <div className="mb-3 flex items-start gap-2">
        <span className="grid h-10 w-10 place-items-center rounded-xl bg-amber-50 text-xl">💡</span>
        <div>
          <h3 className="text-lg font-extrabold text-[#0f2040]">اقتراحات ماتركس</h3>
          <p className="text-xs text-slate-500">ماتركس يحلّل من الي يشوفه، ويقترح عليك بناءً على أداء الموظفين والمهام المعلّقة. ولا شي ينفّذ بلا موافقتك.</p>
        </div>
      </div>
      <div className="mb-4 grid grid-cols-2 gap-2 lg:grid-cols-4">
        <Stat icon="📄" value={rows?.length ?? '…'} label="إجمالي الاقتراحات" sub="اقتراحات من ماتركس حالياً" tone="border-sky-100 bg-sky-50/60 text-sky-900" />
        <Stat icon="⚠️" value={urgent} label="الاقتراحات العاجلة" sub="تحتاج إلى إجراء سريع" tone="border-red-100 bg-red-50/60 text-red-900" />
        <Stat icon="⏳" value={(rows?.length ?? 0) - urgent} label="بانتظار المراجعة" sub="لم يتم الرد عليها بعد" tone="border-amber-100 bg-amber-50/60 text-amber-900" />
        <Stat icon="✅" value={approvedToday} label="تم قبولها اليوم" sub="تم اتخاذ إجراء عليها" tone="border-emerald-100 bg-emerald-50/60 text-emerald-900" />
      </div>
      {err && <p className="mb-2 rounded-lg bg-red-50 p-2 text-red-600">{err}</p>}
      {!rows ? <p className="text-slate-400">جاري التحميل…</p> : (
        <ul className="space-y-3">
          {rows.map((p) => {
            const k = KIND[p.kind]
            const sev = SEV[severity(p)]
            const rule = p.kind === 'GUIDE_RULE' ? (p.payload as { route?: string; match?: string; groups?: string; text?: string; onlyIfPending?: boolean } | null) : null
            const name = p.employeeName ?? (p.kind === 'GUIDE_RULE' ? 'تعليمة لماتركس' : '')
            return (
              <li key={p.id} className="grid gap-3 rounded-2xl border border-slate-200 p-3 shadow-sm lg:grid-cols-[1fr_auto_auto]">
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2 text-[11px]">
                    <span className="rounded-full bg-violet-50 px-2 py-0.5 font-bold text-violet-700">✨ ماتركس</span>
                    {name && (
                      <span className="flex items-center gap-1.5 font-bold text-[#0f2040]">
                        <span className="grid h-7 w-7 place-items-center rounded-full bg-sky-100 text-[11px] text-sky-800">{p.employeeName ? initials(p.employeeName) : '📚'}</span>
                        {p.employeeId ? <Link to={`/matrix/employee/${p.employeeId}`} className="text-sm hover:underline">{name}</Link> : <span className="text-sm">{name}</span>}
                      </span>
                    )}
                    <span className={`rounded-full px-2 py-0.5 font-bold ring-1 ${sev.c}`}>{sev.icon} {sev.t}</span>
                  </div>
                  <p className="mt-1.5 font-extrabold text-slate-800">{p.title}</p>
                  {p.rationale && <p className="text-xs text-slate-600">{p.rationale}</p>}
                  <p className="mt-1 flex flex-wrap gap-3 text-[11px] text-slate-400">
                    <span>🕒 {ago(p.createdAt)}</span><span>المصدر: {p.source === 'MODEL' ? 'هايكو' : 'نظام ماتركس'}</span><span>🏷️ التصنيف: {category(p)}</span><span>{k.icon} {k.label}</span>
                  </p>
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
                </div>
                <div className="min-w-[220px] rounded-xl bg-slate-50 p-2">
                  <p className="mb-1 text-[11px] font-bold text-slate-500">📊 أدلة من البيانات</p>
                  <div className="grid grid-cols-2 gap-1.5">
                    {evidenceCards(p).map((c) => (
                      <div key={c.label} className="rounded-lg bg-white p-2 text-center ring-1 ring-slate-100">
                        <span className="text-base">{c.icon}</span>
                        <p className="text-[10px] text-slate-500">{c.label}</p>
                        <b className="text-xs text-[#0f2040]">{c.value}</b>
                      </div>
                    ))}
                  </div>
                </div>
                <div className="flex min-w-[170px] flex-col gap-1.5">
                  <button disabled={busy === p.id} onClick={() => approve(p)} className="rounded-lg bg-emerald-600 px-3 py-2 text-xs font-bold text-white disabled:opacity-50">{p.kind === 'PREDICTION' ? '✈️ أرسل التذكير للموظف' : k.approve}</button>
                  <button onClick={() => setWhy(why === p.id ? null : p.id)} className="rounded-lg border border-slate-200 px-3 py-1.5 text-xs font-bold text-slate-700">👁️ {why === p.id ? 'إخفاء التفاصيل' : 'عرض التفاصيل'}</button>
                  <button disabled={busy === p.id} onClick={() => reject(p)} className="rounded-lg border border-red-200 bg-red-50 px-3 py-1.5 text-xs font-bold text-red-600 disabled:opacity-50">✕ رفض</button>
                </div>
              </li>
            )
          })}
        </ul>
      )}
    </section>
  )
}
