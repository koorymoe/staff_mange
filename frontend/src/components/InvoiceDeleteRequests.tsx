import { useCallback, useEffect, useState } from 'react'
import { api, type InvoiceDeleteRequest } from '../api'

// ═══ 🗑️ طلبات حذف الفواتير (دازاي 10-25) ═══
// المحاسب يطلب ويكتب السبب، والمالك وحده يوافق أو يرفض. قبل الحذف تنحفظ نسخة كاملة.

/** زر المحاسب: «اطلب حذف» ويا نافذة السبب داخل الصفحة (بلا prompt). */
export function RequestInvoiceDelete({ invoiceId, code, onDone }: { invoiceId: string; code: string; onDone: () => void }) {
  const [open, setOpen] = useState(false)
  const [reason, setReason] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const send = async () => {
    setBusy(true); setErr('')
    try { await api.requestInvoiceDelete(invoiceId, reason.trim()); setOpen(false); setReason(''); onDone() }
    catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }
  return (
    <>
      <button type="button" onClick={() => setOpen(true)}
        className="rounded-lg border border-red-200 px-3 py-1.5 text-xs font-bold text-red-600 hover:bg-red-50">🗑️ اطلب حذف</button>
      {open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" dir="rtl" onClick={() => setOpen(false)}>
          <div className="w-full max-w-md rounded-2xl bg-white p-5 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-bold text-[#0f2040]">🗑️ طلب حذف الفاتورة {code}</h3>
            <p className="mt-1 text-xs text-slate-500">الطلب يروح للمالك، وما تنحذف إلا بموافقته. اكتب السبب بوضوح.</p>
            <textarea value={reason} onChange={(e) => setReason(e.target.value)} rows={3} autoFocus
              placeholder="مثلاً: فاتورة مكررة لنفس الحجز بنفس اليوم"
              className="mt-3 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
            {err && <p className="mt-2 rounded-lg bg-red-50 px-3 py-2 text-xs font-bold text-red-700">{err}</p>}
            <div className="mt-4 flex gap-2">
              <button type="button" onClick={() => void send()} disabled={busy || reason.trim().length < 5}
                className="flex-1 rounded-xl bg-red-600 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50">{busy ? '…' : 'أرسل الطلب للمالك'}</button>
              <button type="button" onClick={() => setOpen(false)} className="rounded-xl border border-slate-300 px-4 py-2.5 text-sm font-bold text-slate-600">رجوع</button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}

/** لوحة الطلبات: المالك يحسم، والمحاسب يشوف حالة طلباته. */
export default function InvoiceDeleteRequests({ isOwner, onChanged }: { isOwner: boolean; onChanged: () => void }) {
  const [rows, setRows] = useState<InvoiceDeleteRequest[]>([])
  const [busy, setBusy] = useState('')
  const [err, setErr] = useState('')
  const load = useCallback(() => { api.getInvoiceDeleteRequests('PENDING').then(setRows).catch(() => {}) }, [])
  useEffect(() => { load() }, [load])
  if (rows.length === 0) return null
  const decide = async (id: string, approve: boolean) => {
    setBusy(id); setErr('')
    try { await api.decideInvoiceDelete(id, approve); load(); onChanged() }
    catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy('') }
  }
  return (
    <div className="mt-3 rounded-2xl border border-red-200 bg-red-50 p-4" dir="rtl">
      <h3 className="font-extrabold text-red-800">🗑️ طلبات حذف فواتير ({rows.length}) {isOwner ? '— تنتظر قرارك' : '— بانتظار المالك'}</h3>
      {err && <p className="mt-2 text-xs font-bold text-red-700">{err}</p>}
      <ul className="mt-2 space-y-2">
        {rows.map((r) => (
          <li key={r.id} className="flex flex-wrap items-center gap-2 rounded-xl bg-white p-3 text-sm">
            <b className="font-mono text-[#0f2040]">{r.accountingCode ?? '—'}</b>
            {r.bookingCode && <span className="text-slate-500">حجز {r.bookingCode}</span>}
            {r.netTotal != null && <span className="text-slate-500">{Number(r.netTotal).toLocaleString()} د.ع</span>}
            <span className="text-slate-700">طلبه: {r.requestedByName ?? '—'} — <i>{r.reason}</i></span>
            {isOwner && (
              <span className="ms-auto flex gap-2">
                <button type="button" disabled={!!busy} onClick={() => void decide(r.id, true)}
                  className="rounded-lg bg-red-600 px-3 py-1.5 text-xs font-bold text-white disabled:opacity-50">احذفها</button>
                <button type="button" disabled={!!busy} onClick={() => void decide(r.id, false)}
                  className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-bold text-slate-600 disabled:opacity-50">ارفض</button>
              </span>
            )}
          </li>
        ))}
      </ul>
    </div>
  )
}
