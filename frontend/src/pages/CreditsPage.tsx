import Modal from '../ui/Modal'
import { Fragment, useCallback, useEffect, useState } from 'react'
import { api, type CreditSale } from '../api'
import { dinarRound as iqd } from '../utils/format'
import { LineIcon } from '../components/LineIcon'

import { notify } from '../ui/notify'
// (DRY 11-02) چانت مكتوبة جوّا كل نموذج بهالملف
const inp = 'w-full rounded-xl border border-slate-300 px-3 py-2 text-sm'

// ═══ سجل الآجل — قرار (ع) 10-14 ═══
// حجز أو مشروع خلص والفلوس مؤجلة: دفعات أو مبلغ كامل. الحجز يوصل هنا من
// «التدقيق اليومي» بزر «⏳ الآجل»، والمشروع ينضاف يدوياً. كل دفعة تنسجّل،
// والدين يتسكّر لمن يكتمل، وأكو إكسل شهري.

const d = (s?: string | null) => (s ? new Date(s).toLocaleDateString('en-CA') : '—')

export default function CreditsPage() {
  const [status, setStatus] = useState<'open' | 'settled' | 'all'>('open')
  const [q, setQ] = useState('')
  const [rows, setRows] = useState<CreditSale[] | null>(null)
  const [err, setErr] = useState('')
  const [pay, setPay] = useState<CreditSale | null>(null)
  const [adding, setAdding] = useState(false)
  const [month, setMonth] = useState(() => new Date().toLocaleDateString('en-CA').slice(0, 7))
  const [history, setHistory] = useState<{ id: string; rows: Awaited<ReturnType<typeof api.getCreditPayments>> } | null>(null)

  const load = useCallback(() => {
    api.getCredits(status, q).then(setRows).catch((e) => { setErr(e.message); setRows([]) })
  }, [status, q])
  useEffect(() => { const t = setTimeout(load, 250); return () => clearTimeout(t) }, [load])

  const open = (rows ?? []).filter((r) => !r.settledAt)
  const owed = open.reduce((a, r) => a + r.amount - r.paid, 0)
  const showHistory = async (r: CreditSale) => {
    if (history?.id === r.id) { setHistory(null); return }
    setHistory({ id: r.id, rows: await api.getCreditPayments(r.id).catch(() => []) })
  }

  return (
    <div dir="rtl" className="w-full space-y-4">
      {/* (ك) 11-23 زينا: الآجل مثل الصورة — الألوان بـindex.css (.cr-*) */}
      <div className="cr-head space-y-4 rounded-2xl border p-5">
        <div>
          <h1 className="text-2xl font-black text-[#0f2040]">⏳ الآجل</h1>
          <p className="cr-sub text-sm">شغل خلص والفلوس مؤجلة — دفعات أو مبلغ كامل. كل دفعة تنسجّل هنا وتنضاف لمستلم الحجز.</p>
        </div>
        <div className="flex flex-wrap items-center gap-4">
          <span className="cr-owed flex items-center gap-4 rounded-2xl border px-6 py-2.5">
            <LineIcon n="coins" className="h-9 w-9 shrink-0 text-amber-400" />
            <span className="text-sm">الباقي على الزبائن<b className="block text-xl font-extrabold">{iqd(owed)}</b></span>
          </span>
          <label className="cr-field relative flex items-center gap-3 rounded-xl border px-4 py-2.5">
            <LineIcon n="cal" className="h-5 w-5" />
            <span className="min-w-[8.5rem] text-center" dir="ltr">{new Date(`${month}-01T12:00:00`).toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}</span>
            <input type="month" value={month} onChange={(e) => e.target.value && setMonth(e.target.value)} className="absolute inset-0 cursor-pointer opacity-0" aria-label="الشهر" />
          </label>
          <button onClick={() => api.exportCredits(month).catch((e) => notify(e.message))} className="cr-excel rounded-xl px-7 py-2.5 text-base font-bold text-white">⬇️ إكسل الشهر</button>
          <button onClick={() => setAdding(true)} className="cr-add rounded-xl px-7 py-2.5 text-base font-bold text-white">+ آجل مشروع / يدوي</button>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <label className="cr-field flex min-w-[240px] max-w-2xl flex-1 items-center gap-2 rounded-xl border px-4 py-2.5">
          <LineIcon n="search" className="h-5 w-5 shrink-0" />
          <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="كود الحجز/المشروع، الزبون، الهاتف…" className="w-full bg-transparent text-sm outline-none" />
        </label>
        <div className="ms-auto flex gap-3">
          {([['open', 'المفتوحة'], ['settled', 'المسدّدة'], ['all', 'الكل']] as const).map(([k, l]) => (
            <button key={k} onClick={() => setStatus(k)} className={`cr-tab min-w-[7rem] rounded-xl border px-5 py-2.5 text-sm font-bold ${status === k ? 'cr-tab-on' : ''}`}>{l}</button>
          ))}
        </div>
      </div>
      {err && <p className="rounded-xl bg-rose-50 p-3 text-sm text-rose-700">{err}</p>}

      <div className="cr-table overflow-x-auto rounded-2xl border">
        <table className="w-full min-w-[980px] text-sm">
          <thead className="cr-th text-sm font-bold">
            <tr>{['الحجز / المشروع', 'الزبون', 'المستحق', 'المدفوع', 'الباقي', 'التقدّم', 'آخر دفعة', 'موعد السداد', 'سجّله', ''].map((h, i) => <th key={i} className="p-3 text-center first:text-right">{h || 'الإجراءات'}</th>)}</tr>
          </thead>
          <tbody>
            {rows === null && <tr><td colSpan={10} className="p-6 text-center text-slate-400">جاري التحميل…</td></tr>}
            {rows?.length === 0 && <tr><td colSpan={10} className="p-6 text-center text-slate-400">ماكو آجل {status === 'open' ? 'مفتوح' : ''} — الحجز يدخل هنا من «التدقيق اليومي» بزر «⏳ الآجل».</td></tr>}
            {rows?.map((r) => {
              const left = r.amount - r.paid
              const pct = Math.min(100, Math.round((r.paid / r.amount) * 100))
              const late = !r.settledAt && r.dueDate && new Date(r.dueDate) < new Date()
              return (
                <Fragment key={r.id}>
                  <tr className="cr-row align-middle">
                    <td className="p-3">
                      <p className="font-bold text-[#0f2040]">{r.title}</p>
                      {r.note && <p className="cr-sub text-[11px]">{r.note}</p>}
                    </td>
                    <td className="p-3 text-xs">{r.customerName ?? '—'}{r.phone && <span className="cr-sub mt-0.5 flex items-center gap-1.5"><LineIcon n="phone" className="h-4 w-4 shrink-0" /><span dir="ltr">{r.phone}</span></span>}</td>
                    <td className="p-3 text-center font-bold tabular-nums">{iqd(r.amount)}</td>
                    <td className="cr-paid p-3 text-center font-bold tabular-nums">{iqd(r.paid)}</td>
                    <td className="cr-left p-3 text-center font-bold tabular-nums">{r.settledAt ? <span className="rounded-lg bg-emerald-50 px-2 py-0.5 text-xs text-emerald-700">✓ مسدّد</span> : iqd(left)}</td>
                    <td className="p-3"><div className="cr-bar h-2 w-28 overflow-hidden rounded-full"><div className="h-full bg-emerald-500" style={{ width: `${pct}%` }} /></div><span className="block text-left text-[10px]">{pct}%</span></td>
                    <td className="p-3 text-center text-xs">{d(r.lastPaidAt)}</td>
                    <td className={`p-3 text-center text-xs ${late ? 'font-bold text-rose-600' : ''}`}>{d(r.dueDate)}{late && ' ⚠️ فات'}</td>
                    <td className="p-3 text-xs"><span className="flex items-start gap-1.5"><LineIcon n="user" className="h-4 w-4 shrink-0" />{r.createdByName ?? '—'}</span><span className="mt-0.5 flex items-center gap-1.5"><LineIcon n="cal" className="h-4 w-4 shrink-0" />{d(r.createdAt)}</span></td>
                    <td className="p-3">
                      <div className="flex justify-center gap-1.5">
                        {!r.settledAt && <button onClick={() => setPay(r)} className="cr-pay rounded-lg px-4 py-1.5 text-xs font-bold text-white">💵 دفعة</button>}
                        {r.paid > 0 && <button onClick={() => showHistory(r)} className="rounded-lg border border-slate-300 px-2.5 py-1 text-xs">الدفعات</button>}
                      </div>
                    </td>
                  </tr>
                  {history?.id === r.id && (
                    <tr className="bg-slate-50/70">
                      <td colSpan={10} className="px-6 py-2 text-xs">
                        {history.rows.map((p) => <p key={p.id} className="py-0.5">📅 {p.paidAt.slice(0, 10)} — <b>{iqd(p.amount)}</b> · استلمها {p.byName ?? '—'}{p.note && ` · ${p.note}`}</p>)}
                      </td>
                    </tr>
                  )}
                </Fragment>
              )
            })}
          </tbody>
        </table>
      </div>

      {pay && <PayModal credit={pay} onClose={() => setPay(null)} onDone={() => { setPay(null); setHistory(null); load() }} />}
      {adding && <AddModal onClose={() => setAdding(false)} onDone={() => { setAdding(false); load() }} />}
    </div>
  )
}

function PayModal({ credit, onClose, onDone }: { credit: CreditSale; onClose: () => void; onDone: () => void }) {
  const left = credit.amount - credit.paid
  const [amount, setAmount] = useState('')
  const [paidAt, setPaidAt] = useState(() => new Date().toLocaleDateString('en-CA'))
  const [note, setNote] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const save = async () => {
    setBusy(true); setErr('')
    try { await api.addCreditPayment(credit.id, { amount: Number(amount.replace(/,/g, '')), paidAt, note: note || undefined }); onDone() }
    catch (e) { setErr(e instanceof Error ? e.message : 'خطأ') } finally { setBusy(false) }
  }
  return (
    <Modal title={<>💵 دفعة — {credit.title}</>} onClose={onClose} size="md" busy={busy} bodyClassName="space-y-3 p-5">
        <p className="text-sm text-slate-600">الباقي: <b className="text-violet-700">{iqd(left)}</b></p>
        <div className="flex gap-2">
          <input inputMode="numeric" dir="ltr" placeholder="المبلغ" value={amount} onChange={(e) => setAmount(e.target.value.replace(/[^\d,]/g, ''))} className={inp} />
          <button onClick={() => setAmount(String(Math.round(left)))} className="shrink-0 rounded-xl border border-violet-300 bg-violet-50 px-3 text-xs font-bold text-violet-700">المبلغ كامل</button>
        </div>
        <label className="block text-xs text-slate-500">تاريخ الاستلام<input type="date" value={paidAt} onChange={(e) => setPaidAt(e.target.value)} className={`${inp} mt-1`} /></label>
        <input placeholder="ملاحظة (اختياري)" value={note} onChange={(e) => setNote(e.target.value)} className={inp} />
        {err && <p className="text-sm text-rose-600">{err}</p>}
        <div className="flex gap-2"><button disabled={busy || !Number(amount.replace(/,/g, ''))} onClick={save} className="flex-1 rounded-xl bg-emerald-700 py-2.5 font-bold text-white disabled:opacity-50">تسجيل الدفعة</button><button onClick={onClose} className="rounded-xl border border-slate-300 px-4">إلغاء</button></div>
    </Modal>
  )
}

function AddModal({ onClose, onDone }: { onClose: () => void; onDone: () => void }) {
  const [f, setF] = useState({ projectCode: '', bookingCode: '', title: '', customerName: '', phone: '', amount: '', dueDate: '', note: '' })
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const set = (k: keyof typeof f) => (e: { target: { value: string } }) => setF({ ...f, [k]: e.target.value })
  const save = async () => {
    setBusy(true); setErr('')
    try {
      await api.createCredit({ ...f, amount: Number(f.amount.replace(/,/g, '')), dueDate: f.dueDate || undefined, note: f.note || undefined })
      onDone()
    } catch (e) { setErr(e instanceof Error ? e.message : 'خطأ') } finally { setBusy(false) }
  }
  return (
    <Modal title="+ آجل جديد" onClose={onClose} size="md" busy={busy} bodyClassName="space-y-3 p-5">
        <p className="text-xs text-slate-500">للمشروع اكتب كوده (PRJ-…). الحجز الأفضل يدخل من «التدقيق اليومي»، بس تكدر تضيفه هنا بكوده.</p>
        <div className="grid grid-cols-2 gap-2">
          <input dir="ltr" placeholder="كود المشروع PRJ-0042" value={f.projectCode} onChange={set('projectCode')} className={inp} />
          <input dir="ltr" placeholder="أو كود الحجز B823" value={f.bookingCode} onChange={set('bookingCode')} className={inp} />
        </div>
        <input placeholder="العنوان (اختياري — ينكتب تلقائي من الكود)" value={f.title} onChange={set('title')} className={inp} />
        <div className="grid grid-cols-2 gap-2">
          <input placeholder="اسم الزبون" value={f.customerName} onChange={set('customerName')} className={inp} />
          <input dir="ltr" placeholder="الهاتف" value={f.phone} onChange={set('phone')} className={inp} />
        </div>
        <input inputMode="numeric" dir="ltr" placeholder="المبلغ المستحق *" value={f.amount} onChange={(e) => setF({ ...f, amount: e.target.value.replace(/[^\d,]/g, '') })} className={inp} />
        <label className="block text-xs text-slate-500">موعد السداد (اختياري)<input type="date" value={f.dueDate} onChange={set('dueDate')} className={`${inp} mt-1`} /></label>
        <input placeholder="ملاحظة: مثلاً 3 دفعات شهرية" value={f.note} onChange={set('note')} className={inp} />
        {err && <p className="text-sm text-rose-600">{err}</p>}
        <div className="flex gap-2"><button disabled={busy || !Number(f.amount.replace(/,/g, ''))} onClick={save} className="flex-1 rounded-xl bg-[#0f2040] py-2.5 font-bold text-white disabled:opacity-50">حفظ</button><button onClick={onClose} className="rounded-xl border border-slate-300 px-4">إلغاء</button></div>
    </Modal>
  )
}
