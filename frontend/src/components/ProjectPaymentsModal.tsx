import { useEffect, useState } from 'react'
import { api, type ProjectPaymentsState } from '../api'

// ═══ 💰 دفعات المشروع — قرار (ع) 10-06 ═══
// «من مشروع واحد محصل 500 مليون ليش ماموجوده بالاحصائيات». كل دفعة تنسجّل
// هنا بمبلغها وتاريخها ووصلها، وتدخل بالإيرادات. المحاسب ومدير المشاريع
// ومشرف المشروع يسجّلون؛ المحاسب يأكد ويلغي.

const fmt = (v: number) => v.toLocaleString('en-US') + ' د.ع'
const METHOD: Record<string, string> = { CASH: 'نقد', TRANSFER: 'تحويل', CHEQUE: 'صك' }
const today = () => new Date().toISOString().slice(0, 10)

export default function ProjectPaymentsModal({ projectId, title, onClose }: { projectId: string; title: string; onClose: () => void }) {
  const [st, setSt] = useState<ProjectPaymentsState | null>(null)
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const [amount, setAmount] = useState('')
  const [paidAt, setPaidAt] = useState(today())
  const [method, setMethod] = useState('CASH')
  const [receipt, setReceipt] = useState('')
  const [note, setNote] = useState('')
  const [value, setValue] = useState('')
  useEffect(() => { void api.getProjectPayments(projectId).then((s) => { setSt(s); setValue(s.money.contractValue ? String(s.money.contractValue) : '') }).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر')) }, [projectId])

  const run = async (fn: () => Promise<ProjectPaymentsState>) => {
    setBusy(true); setErr('')
    try { setSt(await fn()); return true } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر'); return false } finally { setBusy(false) }
  }
  const num = (s: string) => Number(s.replace(/[^\d]/g, ''))
  const add = async () => {
    if (!num(amount)) { setErr('اكتب المبلغ'); return }
    if (await run(() => api.addProjectPayment(projectId, { amount: num(amount), paidAt, method, receiptNo: receipt || undefined, note: note || undefined }))) {
      setAmount(''); setReceipt(''); setNote('')
    }
  }
  const m = st?.money
  const left = m?.contractValue != null ? m.contractValue - m.paid : null
  const pct = m?.contractValue ? Math.min(100, Math.round((m.paid / m.contractValue) * 100)) : null

  return (
    <div dir="rtl" className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-3" onClick={onClose}>
      <div className="max-h-[92vh] w-full max-w-xl overflow-y-auto rounded-2xl bg-white p-4 shadow-2xl" onClick={(e) => e.stopPropagation()}>
        <div className="mb-3 flex items-center justify-between">
          <p className="text-lg font-extrabold text-[#0f2040]">💰 دفعات: {title}</p>
          <button type="button" onClick={onClose} className="text-2xl text-slate-400">×</button>
        </div>
        {err && <p className="mb-2 rounded-lg bg-red-50 p-2 text-sm text-red-700">{err}</p>}
        {!st ? <p className="text-sm text-slate-400">…</p> : (
          <div className="space-y-3">
            <div className="grid grid-cols-3 gap-2 text-center">
              <div className="rounded-xl bg-slate-50 p-2"><p className="text-[11px] text-slate-500">قيمة العقد</p><b className="text-sm">{m!.contractValue != null ? fmt(m!.contractValue) : '—'}</b>{m!.valueFromText && <p className="text-[10px] text-amber-600">من نص السعر</p>}</div>
              <div className="rounded-xl bg-emerald-50 p-2"><p className="text-[11px] text-slate-500">المدفوع</p><b className="text-sm text-emerald-700">{fmt(m!.paid)}</b>{m!.unverified > 0 && <p className="text-[10px] text-amber-600">منها {fmt(m!.unverified)} تنتظر تأكيد</p>}</div>
              <div className="rounded-xl bg-amber-50 p-2"><p className="text-[11px] text-slate-500">الباقي</p><b className="text-sm text-amber-700">{left != null ? fmt(Math.max(0, left)) : '—'}</b></div>
            </div>
            {pct != null && <div className="h-2 overflow-hidden rounded-full bg-slate-100"><div className="h-full bg-emerald-500" style={{ width: `${pct}%` }} /></div>}

            {st.canValue && (
              <div className="flex items-end gap-2">
                <label className="flex-1 text-xs text-slate-600">قيمة العقد (رقم)
                  <input inputMode="numeric" value={value} onChange={(e) => setValue(e.target.value)} className="mt-1 w-full rounded-lg border border-slate-300 p-2 text-sm" placeholder="مثلاً 500000000" />
                </label>
                <button type="button" disabled={busy || !num(value)} onClick={() => void run(() => api.setProjectContractValue(projectId, num(value)))} className="rounded-lg border border-slate-300 px-3 py-2 text-xs font-bold disabled:opacity-50">احفظ القيمة</button>
              </div>
            )}

            <div className="space-y-2 rounded-xl border border-emerald-200 bg-emerald-50/40 p-3">
              <p className="text-sm font-extrabold text-emerald-900">+ سجّل دفعة</p>
              <div className="grid grid-cols-2 gap-2">
                <input inputMode="numeric" value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="المبلغ (د.ع) *" className="rounded-lg border border-slate-300 p-2 text-sm" />
                <input type="date" value={paidAt} onChange={(e) => setPaidAt(e.target.value)} className="rounded-lg border border-slate-300 p-2 text-sm" />
                <select value={method} onChange={(e) => setMethod(e.target.value)} className="rounded-lg border border-slate-300 p-2 text-sm">
                  <option value="CASH">نقد</option><option value="TRANSFER">تحويل</option><option value="CHEQUE">صك</option>
                </select>
                <input value={receipt} onChange={(e) => setReceipt(e.target.value)} placeholder="رقم الوصل" className="rounded-lg border border-slate-300 p-2 text-sm" />
              </div>
              <input value={note} onChange={(e) => setNote(e.target.value)} placeholder="ملاحظة (مثلاً: الدفعة الأولى، منو استلم)" className="w-full rounded-lg border border-slate-300 p-2 text-sm" />
              {num(amount) > 0 && <p className="text-xs text-slate-600">المبلغ: <b>{fmt(num(amount))}</b></p>}
              <button type="button" disabled={busy} onClick={() => void add()} className="w-full rounded-lg bg-emerald-700 py-2 text-sm font-bold text-white disabled:opacity-50">سجّل الدفعة</button>
              {!st.canVerify && <p className="text-[11px] text-slate-500">الدفعة تنتظر تأكيد المحاسب، بس تدخل بالإحصائيات من هسه.</p>}
            </div>

            <div className="space-y-1.5">
              <p className="text-xs font-extrabold text-slate-700">سجل الدفعات ({st.payments.length})</p>
              {st.payments.map((p) => (
                <div key={p.id} className={`rounded-xl border p-2 text-sm ${p.cancelledAt ? 'border-slate-200 bg-slate-50 text-slate-400' : 'border-slate-200'}`}>
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <span className={p.cancelledAt ? 'line-through' : ''}><b>{fmt(p.amount)}</b> · {METHOD[p.method]} · {p.paidAt.slice(0, 10)}{p.receiptNo && ` · وصل ${p.receiptNo}`}</span>
                    <span className="flex items-center gap-1 text-xs">
                      {p.cancelledAt ? <span>ملغية: {p.cancelReason}</span> : p.verifiedAt ? <span className="text-emerald-700">✔ تأكدت{p.verifiedBy && ` (${p.verifiedBy})`}</span> : <span className="text-amber-600">⏳ تنتظر تأكيد</span>}
                      {st.canVerify && !p.cancelledAt && !p.verifiedAt && <button type="button" disabled={busy} onClick={() => void run(() => api.verifyProjectPayment(p.id))} className="rounded border border-emerald-400 px-1.5 text-emerald-700">✔ وصلت</button>}
                      {st.canVerify && !p.cancelledAt && <button type="button" disabled={busy} onClick={() => { const r = window.prompt('سبب الإلغاء؟'); if (r) void run(() => api.cancelProjectPayment(p.id, r)) }} className="rounded border border-slate-300 px-1.5">إلغاء</button>}
                    </span>
                  </div>
                  <p className="text-[11px] text-slate-500">سجّلها {p.createdBy ?? '—'}{p.note && ` · ${p.note}`}</p>
                </div>
              ))}
              {st.payments.length === 0 && <p className="rounded-lg bg-amber-50 p-2 text-xs text-amber-800">🤖 ماتركس: ماكو ولا دفعة مسجّلة على هالمشروع.</p>}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
