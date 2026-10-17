import { useEffect, useMemo, useState } from 'react'
import { api, type PaymentAuditRow } from '../api'

// ═══ سجل دفعات المشاريع — للمالك (ع 10-13) ═══
// «المدير كتب المستلم مرة ثانية فصار ضعف» — هنا يشوف كل دفعة ومنو سجّلها
// ووكتها، والمكرر المحتمل بالأحمر، ويصحّح المبلغ أو يلغي الدفعة.
const money = (n: number) => n.toLocaleString('en-US')
const dt = (s: string) => new Date(s).toLocaleString('ar-IQ', { timeZone: 'Asia/Baghdad', day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })

export default function PaymentAuditModal({ onClose, onChanged }: { onClose: () => void; onChanged: () => void }) {
  const [rows, setRows] = useState<PaymentAuditRow[] | null>(null)
  const [by, setBy] = useState('')
  const [since, setSince] = useState('')
  const [onlyDup, setOnlyDup] = useState(false)
  const [edit, setEdit] = useState<Record<string, string>>({})
  const [busy, setBusy] = useState('')
  const [err, setErr] = useState('')
  const [reload, setReload] = useState(0)

  useEffect(() => { api.getPaymentAudit('', since).then(setRows).catch((e) => { setErr(e instanceof Error ? e.message : 'تعذر'); setRows([]) }) }, [since, reload])

  const people = useMemo(() => {
    const m = new Map<string, { name: string; n: number }>()
    for (const r of rows ?? []) if (r.createdById) m.set(r.createdById, { name: r.createdByName || '—', n: (m.get(r.createdById)?.n ?? 0) + 1 })
    return [...m.entries()]
  }, [rows])

  const groups = useMemo(() => {
    const list = (rows ?? []).filter((r) => (!by || r.createdById === by) && (!onlyDup || r.duplicate))
    const g = new Map<string, PaymentAuditRow[]>()
    for (const r of list) g.set(r.projectId, [...(g.get(r.projectId) ?? []), r])
    // كل دفعات المشروع (حتى الي سجّلها غيره) حتى يشوف الصورة كاملة
    return [...g.keys()].map((pid) => (rows ?? []).filter((r) => r.projectId === pid))
  }, [rows, by, onlyDup])

  const save = async (r: PaymentAuditRow, amount: number) => {
    const reason = amount === 0
      ? window.prompt('سبب إلغاء الدفعة؟', 'دفعة مكررة') : window.prompt('سبب التعديل؟', 'تصحيح المبلغ المستلم')
    if (reason === null) return
    setBusy(r.id); setErr('')
    try { await api.setPaymentAmount(r.id, amount, reason); setEdit((e) => ({ ...e, [r.id]: '' })); setReload((x) => x + 1); onChanged() }
    catch (e) { setErr(e instanceof Error ? e.message : 'تعذر الحفظ') }
    finally { setBusy('') }
  }

  const dupCount = (rows ?? []).filter((r) => r.duplicate).length
  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/40 p-4" dir="rtl" onClick={onClose}>
      <div className="w-full max-w-5xl rounded-2xl bg-white p-5 shadow-2xl" onClick={(e) => e.stopPropagation()}>
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-xl font-extrabold text-[#0f2040]">🧾 سجل دفعات المشاريع</h2>
          <button onClick={onClose} className="text-2xl text-slate-400 hover:text-slate-600">×</button>
        </div>
        <div className="mb-3 flex flex-wrap items-center gap-2 text-sm">
          <select value={by} onChange={(e) => setBy(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-1.5">
            <option value="">سجّلها: الكل</option>
            {people.map(([id, p]) => <option key={id} value={id}>سجّلها: {p.name} ({p.n})</option>)}
          </select>
          <label className="flex items-center gap-1">من تاريخ <input type="date" value={since} onChange={(e) => setSince(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1" /></label>
          <label className="flex items-center gap-1 font-bold text-red-700"><input type="checkbox" checked={onlyDup} onChange={(e) => setOnlyDup(e.target.checked)} /> المكرر المحتمل بس ({dupCount})</label>
        </div>
        {err && <p className="mb-2 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700">{err}</p>}
        {rows === null ? <p className="py-10 text-center text-slate-400">جاري التحميل…</p>
          : groups.length === 0 ? <p className="py-10 text-center text-slate-500">ماكو دفعات بهالفلتر.</p>
          : (
            <div className="space-y-3">
              {groups.map((ps) => {
                const p = ps[0]
                return (
                  <section key={p.projectId} className="rounded-xl border border-slate-200 p-3">
                    <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
                      <p className="font-extrabold text-[#0f2040]">{p.projectName} <span className="text-xs font-normal text-slate-400">{p.projectCode}</span></p>
                      <p className="text-sm">قيمة العقد <b>{p.contractValue != null ? money(p.contractValue) : '—'}</b> · المستلم هسه{' '}
                        <b className={p.contractValue != null && p.received > p.contractValue ? 'text-red-600' : 'text-emerald-700'}>{money(p.received)}</b> د.ع
                        {p.contractValue != null && p.received > p.contractValue && <span className="ms-1 text-xs font-bold text-red-600">⚠️ أكثر من العقد</span>}
                      </p>
                    </div>
                    <table className="w-full text-sm">
                      <thead><tr className="text-[11px] text-slate-500"><th className="text-right">المبلغ</th><th className="text-right">تاريخ الدفعة</th><th className="text-right">سجّلها</th><th className="text-right">الحالة</th><th className="text-right">تصحيح</th></tr></thead>
                      <tbody>
                        {ps.map((r) => (
                          <tr key={r.id} className={`border-t border-slate-100 ${r.cancelledAt ? 'opacity-50' : r.duplicate ? 'bg-red-50' : ''}`}>
                            <td className="py-1.5 font-extrabold tabular-nums">{money(r.amount)} {r.duplicate && !r.cancelledAt && <span className="text-[11px] font-bold text-red-600">🔴 مكرر محتمل</span>}</td>
                            <td className="text-xs">{new Date(r.paidAt).toLocaleDateString('ar-IQ')}</td>
                            <td className="text-xs">{r.createdByName || '—'} · {dt(r.createdAt)}</td>
                            <td className="text-xs">{r.cancelledAt ? `ملغاة — ${r.cancelReason || ''}` : r.verifiedAt ? '✔ مؤكدة' : '⏳ تنتظر المحاسب'}{r.note ? <span className="block text-[10px] text-slate-400">{r.note}</span> : null}</td>
                            <td>
                              {!r.cancelledAt && (
                                <div className="flex items-center gap-1">
                                  <input inputMode="numeric" placeholder="المبلغ الصح" value={edit[r.id] ?? ''} onChange={(e) => setEdit((x) => ({ ...x, [r.id]: e.target.value.replace(/[^0-9]/g, '') }))}
                                    className="w-28 rounded border border-slate-300 px-2 py-1 text-xs" />
                                  <button disabled={busy === r.id || !edit[r.id]} onClick={() => void save(r, Number(edit[r.id]))} className="rounded bg-[#0f2040] px-2 py-1 text-xs font-bold text-white disabled:opacity-40">حفظ</button>
                                  <button disabled={busy === r.id} onClick={() => { if (window.confirm(`تلغي دفعة ${money(r.amount)}؟`)) void save(r, 0) }} className="rounded border border-red-300 px-2 py-1 text-xs font-bold text-red-700">إلغاء</button>
                                </div>
                              )}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </section>
                )
              })}
            </div>
          )}
      </div>
    </div>
  )
}
