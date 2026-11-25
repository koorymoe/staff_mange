import { useEffect, useState } from 'react'
import { api, apiRequest } from '../api'
import { notify } from '../ui/notify'
import { fmtDateTime } from '../utils/date'

// ═══ 🧾 سندات الصرف (ل) 11-25 دازاي ═══
// «إذا صرفت سند لأي مشروع أو حجز أو مشروع خارجي (مو شرط حجز): كلفة السند، ووقت الصرف،
//  والمبلغ لمنو انسلّم لو انحط بالكاشير». الرقم تسلسلي، والسند ما ينحذف — يتلغى بسبب.

type V = { id: string; number: number; targetKind: 'PROJECT' | 'BOOKING' | 'EXTERNAL'; targetLabel: string; purpose: string; amount: number; paidAt: string; recipientKind: 'EMPLOYEE' | 'CASHIER'; recipient: string | null; note: string | null; createdBy: string; cancelledAt: string | null; cancelReason: string | null }
type L = { month: string; items: V[]; total: number; toCashier: number; toStaff: number }
const iqd = (v: number) => Math.round(v).toLocaleString('en-US')
const thisMonth = () => new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' }).slice(0, 7)
const nowLocal = () => new Date(Date.now() - new Date().getTimezoneOffset() * 60000).toISOString().slice(0, 16)
const KIND = { PROJECT: '🏗️ مشروع', BOOKING: '📋 حجز', EXTERNAL: '🌐 مشروع خارجي' } as const
const inp = 'rounded-lg border border-slate-300 p-2 text-sm'
const blank = () => ({ targetKind: 'BOOKING' as V['targetKind'], code: '', externalName: '', purpose: '', amount: '', paidAt: nowLocal(), recipientKind: 'CASHIER' as V['recipientKind'], recipientEmployeeId: '', note: '' })

export default function VouchersPage() {
  const [month, setMonth] = useState(thisMonth())
  const [d, setD] = useState<L | null>(null)
  const [err, setErr] = useState('')
  const [f, setF] = useState(blank())
  const [emps, setEmps] = useState<{ id: string; name: string }[]>([])
  const [busy, setBusy] = useState(false)
  const [cancel, setCancel] = useState<Record<string, string>>({})
  const load = (m = month) => apiRequest<L>(`/vouchers?month=${m}`).then(setD).catch((e: Error) => setErr(e.message))
  useEffect(() => { void load(month) }, [month]) // eslint-disable-line react-hooks/exhaustive-deps
  useEffect(() => { api.getEmployees().then((es) => setEmps(es.map((e) => ({ id: e.id, name: e.name })))).catch(() => {}) }, [])
  if (err) return <p role="alert" className="text-rose-700">{err}</p>
  if (!d) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>

  const save = async () => {
    setBusy(true)
    try {
      const v = await apiRequest<V>('/vouchers', { method: 'POST', body: JSON.stringify({ ...f, amount: Number(f.amount), paidAt: new Date(f.paidAt).toISOString() }) })
      notify(`انصرف السند رقم ${v.number} ✅`, 'success'); setF(blank()); void load()
    } catch (e) { notify(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }
  const doCancel = async (v: V) => {
    try { await apiRequest(`/vouchers/${v.id}/cancel`, { method: 'POST', body: JSON.stringify({ reason: cancel[v.id] ?? '' }) }); notify('انلغى السند', 'success'); void load() } catch (e) { notify(e instanceof Error ? e.message : 'تعذر') }
  }

  return (
    <div className="mx-auto max-w-5xl space-y-4" data-testid="vouchers">
      <header className="flex flex-wrap items-center justify-between gap-2">
        <h1 className="text-2xl font-extrabold text-[#0f2040]">🧾 سندات الصرف</h1>
        <input type="month" value={month} onChange={(e) => e.target.value && setMonth(e.target.value)} className={inp} dir="ltr" />
      </header>

      <section className="grid gap-2 rounded-2xl bg-white p-4 shadow-sm sm:grid-cols-2">
        <h2 className="font-bold sm:col-span-2">➕ صرف سند جديد</h2>
        <div className="flex flex-wrap gap-2 sm:col-span-2">
          {(Object.keys(KIND) as V['targetKind'][]).map((k) => (
            <button key={k} type="button" onClick={() => setF({ ...f, targetKind: k })} className={`rounded-full px-3 py-1 text-sm font-bold ${f.targetKind === k ? 'bg-[#0f2040] text-white' : 'bg-slate-100 text-slate-600'}`}>{KIND[k]}</button>
          ))}
        </div>
        {f.targetKind === 'EXTERNAL'
          ? <input value={f.externalName} onChange={(e) => setF({ ...f, externalName: e.target.value })} placeholder="اسم المشروع الخارجي" className={inp} />
          : <input value={f.code} onChange={(e) => setF({ ...f, code: e.target.value })} placeholder={f.targetKind === 'PROJECT' ? 'كود المشروع' : 'كود الحجز (B…)'} className={inp} dir="ltr" />}
        <input value={f.purpose} onChange={(e) => setF({ ...f, purpose: e.target.value })} placeholder="السند لشنو (شراء كيبل، أجور نقل…)" className={inp} />
        <label className="text-xs text-slate-600">المبلغ (كلفة السند)<input type="number" min={0} value={f.amount} onChange={(e) => setF({ ...f, amount: e.target.value })} className={`${inp} mt-1 w-full`} dir="ltr" /></label>
        <label className="text-xs text-slate-600">وقت الصرف<input type="datetime-local" value={f.paidAt} onChange={(e) => setF({ ...f, paidAt: e.target.value })} className={`${inp} mt-1 w-full`} /></label>
        <div className="flex flex-wrap items-center gap-2 sm:col-span-2">
          <span className="text-sm font-bold">المبلغ:</span>
          {([['CASHIER', '🏦 انحط بالكاشير'], ['EMPLOYEE', '👤 انسلّم لموظف']] as const).map(([k, l]) => (
            <button key={k} type="button" onClick={() => setF({ ...f, recipientKind: k })} className={`rounded-full px-3 py-1 text-sm font-bold ${f.recipientKind === k ? 'bg-emerald-700 text-white' : 'bg-slate-100 text-slate-600'}`}>{l}</button>
          ))}
          {f.recipientKind === 'EMPLOYEE' && (
            <select value={f.recipientEmployeeId} onChange={(e) => setF({ ...f, recipientEmployeeId: e.target.value })} className={inp}>
              <option value="">اختار الموظف…</option>
              {emps.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
            </select>
          )}
        </div>
        <input value={f.note} onChange={(e) => setF({ ...f, note: e.target.value })} placeholder="ملاحظة (اختياري)" className={`${inp} sm:col-span-2`} />
        <button type="button" disabled={busy || !f.purpose.trim() || !(Number(f.amount) > 0)} onClick={() => void save()} className="rounded-xl bg-[#0f2040] px-6 py-2.5 font-bold text-white disabled:opacity-40 sm:col-span-2 sm:justify-self-start">{busy ? 'يصرف…' : 'اصرف السند'}</button>
      </section>

      <section className="grid gap-2 sm:grid-cols-3">
        {([['مجموع الشهر', d.total], ['انحط بالكاشير', d.toCashier], ['انسلّم لموظفين', d.toStaff]] as const).map(([k, v]) => (
          <div key={k} className="rounded-xl bg-white p-3 shadow-sm"><p className="text-xs text-slate-500">{k}</p><p className="text-lg font-extrabold" dir="ltr">{iqd(v)}</p></div>
        ))}
      </section>

      <section className="overflow-x-auto rounded-2xl bg-white p-4 shadow-sm">
        <table className="w-full min-w-[820px] text-sm">
          <thead className="bg-slate-50 text-xs text-slate-600"><tr>{['#', 'لـ', 'لشنو', 'المبلغ', 'وقت الصرف', 'المبلغ وين', 'صرفه', ''].map((h) => <th key={h} className="p-2 text-center">{h}</th>)}</tr></thead>
          <tbody className="divide-y">
            {d.items.length === 0 && <tr><td colSpan={8} className="p-4 text-center text-slate-500">ماكو سندات بهالشهر.</td></tr>}
            {d.items.map((v) => (
              <tr key={v.id} className={`text-center ${v.cancelledAt ? 'opacity-50' : ''}`}>
                <td className="p-2 font-bold">{v.number}</td>
                <td className="text-start">{KIND[v.targetKind]} <b>{v.targetLabel}</b></td>
                <td className="text-start">{v.purpose}{v.note && <small className="block text-xs text-slate-500">{v.note}</small>}</td>
                <td dir="ltr" className={v.cancelledAt ? 'line-through' : 'font-bold'}>{iqd(v.amount)}</td>
                <td className="text-xs">{fmtDateTime(v.paidAt)}</td>
                <td>{v.recipientKind === 'CASHIER' ? '🏦 الكاشير' : `👤 ${v.recipient ?? ''}`}</td>
                <td className="text-xs">{v.createdBy}</td>
                <td className="whitespace-nowrap">
                  {v.cancelledAt ? <span className="text-xs text-rose-700">ملغي: {v.cancelReason}</span> : (
                    <span className="flex items-center gap-1">
                      <input value={cancel[v.id] ?? ''} onChange={(e) => setCancel({ ...cancel, [v.id]: e.target.value })} placeholder="سبب الإلغاء" className="w-28 rounded border border-slate-300 p-1 text-xs" />
                      <button type="button" disabled={!(cancel[v.id] ?? '').trim()} onClick={() => void doCancel(v)} className="rounded bg-slate-200 px-2 py-1 text-xs font-bold disabled:opacity-40">إلغاء</button>
                    </span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </div>
  )
}
