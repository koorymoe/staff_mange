import { useEffect, useMemo, useState } from 'react'
import { api, type Booking, type DecorRate } from '../api'

// ═══ فاتورة الديكور — (ع) 10-16 ═══
// نفس مسار فاتورة الليدر (المحاسب، التدقيق، المراقب)، بس الأجور من «جدول أسعار
// الديكور» × الكمية، والمواد بنود. الحساب الأخير بالسيرفر — هنا تقدير للعرض بس.
export default function DecorInvoiceForm({ booking, onBack, onDone }: { booking: Booking; onBack: () => void; onDone: () => void }) {
  const [rates, setRates] = useState<DecorRate[]>([])
  const [lines, setLines] = useState<{ rateId: string; qty: string; notes: string }[]>([{ rateId: '', qty: '', notes: '' }])
  const [mats, setMats] = useState<{ name: string; quantity: string; unitPrice: string }[]>([])
  const [discount, setDiscount] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  useEffect(() => { api.getDecorRates().then(setRates).catch(() => {}) }, [])
  const byId = useMemo(() => new Map(rates.map((r) => [r.id, r])), [rates])
  const labor = lines.reduce((a, l) => a + (byId.get(l.rateId)?.price ?? 0) * (Number(l.qty) || 0), 0)
  const matTotal = mats.reduce((a, m) => a + (Number(m.quantity) || 0) * (Number(m.unitPrice) || 0), 0)
  const net = Math.max(0, labor + matTotal - (Number(discount) || 0))
  const money = (n: number) => `${Math.round(n).toLocaleString('en-US')} د.ع`

  const submit = async () => {
    const dl = lines.filter((l) => l.rateId && Number(l.qty) > 0)
    if (dl.length === 0) { setErr('ضيف بند واحد على الأقل من جدول الأسعار بكميته'); return }
    setBusy(true); setErr('')
    try {
      await api.createLeaderInvoice({
        bookingId: booking.id,
        customerName: booking.customer?.name, customerPhone: booking.customer?.phone, customerAddress: booking.address ?? undefined,
        systems: [], items: [],
        decorLines: dl.map((l) => ({ rateId: l.rateId, qty: Number(l.qty), notes: l.notes.trim() })),
        materials: mats.filter((m) => m.name.trim() && Number(m.quantity) > 0)
          .map((m) => ({ name: m.name.trim(), quantity: Number(m.quantity), unitPrice: Number(m.unitPrice) || 0 })),
        discountValue: Number(discount) || 0,
      })
      onDone()
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر إرسال الفاتورة') } finally { setBusy(false) }
  }

  return (
    <div className="space-y-4" dir="rtl">
      <div className="rounded-2xl bg-gradient-to-l from-orange-600 to-amber-500 p-5 text-white shadow">
        <h1 className="text-2xl font-black">🪚 فاتورة شغل الديكور — {booking.code}</h1>
        <p className="mt-1 text-sm text-orange-50">{booking.customer?.name} · الأجور من جدول أسعار الديكور × الكمية، والمواد بنود.</p>
      </div>
      <div className="rounded-2xl bg-white p-4 ring-1 ring-slate-200">
        <p className="mb-2 font-extrabold text-[#0f2040]">أجور العمل</p>
        {rates.length === 0 && <p className="text-sm text-amber-700">⚠️ جدول أسعار الديكور فارغ — المدير لازم يضيف البنود أول.</p>}
        {lines.map((l, i) => {
          const rt = byId.get(l.rateId)
          return (
            <div key={i} className="mb-1.5 grid grid-cols-[1fr_6rem_1fr_7rem_2rem] items-center gap-1.5 text-sm">
              <select value={l.rateId} onChange={(e) => setLines(lines.map((x, j) => (j === i ? { ...x, rateId: e.target.value } : x)))} className="rounded-lg border border-slate-300 px-2 py-1.5">
                <option value="">البند…</option>
                {rates.map((r) => <option key={r.id} value={r.id}>{r.serviceName} — {r.name} ({r.price.toLocaleString('en-US')} / {r.unit})</option>)}
              </select>
              <input type="number" min={0} step="0.5" value={l.qty} onChange={(e) => setLines(lines.map((x, j) => (j === i ? { ...x, qty: e.target.value } : x)))} placeholder={rt?.unit ?? 'الكمية'} className="rounded-lg border border-slate-300 px-2 py-1.5" />
              <input value={l.notes} onChange={(e) => setLines(lines.map((x, j) => (j === i ? { ...x, notes: e.target.value } : x)))} placeholder="ملاحظة (اختياري)" className="rounded-lg border border-slate-300 px-2 py-1.5" />
              <span className="text-left tabular-nums text-slate-600">{money((rt?.price ?? 0) * (Number(l.qty) || 0))}</span>
              <button type="button" onClick={() => setLines(lines.filter((_, j) => j !== i))} className="text-red-500">✕</button>
            </div>
          )
        })}
        <button type="button" onClick={() => setLines([...lines, { rateId: '', qty: '', notes: '' }])} className="mt-1 rounded-lg border border-dashed border-slate-400 px-3 py-1 text-xs font-bold text-slate-600">➕ ضيف بند</button>
      </div>
      <div className="rounded-2xl bg-white p-4 ring-1 ring-slate-200">
        <p className="mb-2 font-extrabold text-[#0f2040]">المواد</p>
        {mats.map((m, i) => (
          <div key={i} className="mb-1.5 grid grid-cols-[1fr_6rem_8rem_2rem] gap-1.5 text-sm">
            <input value={m.name} onChange={(e) => setMats(mats.map((x, j) => (j === i ? { ...x, name: e.target.value } : x)))} placeholder="المادة" className="rounded-lg border border-slate-300 px-2 py-1.5" />
            <input type="number" min={0} value={m.quantity} onChange={(e) => setMats(mats.map((x, j) => (j === i ? { ...x, quantity: e.target.value } : x)))} placeholder="الكمية" className="rounded-lg border border-slate-300 px-2 py-1.5" />
            <input type="number" min={0} value={m.unitPrice} onChange={(e) => setMats(mats.map((x, j) => (j === i ? { ...x, unitPrice: e.target.value } : x)))} placeholder="سعر الوحدة" className="rounded-lg border border-slate-300 px-2 py-1.5" />
            <button type="button" onClick={() => setMats(mats.filter((_, j) => j !== i))} className="text-red-500">✕</button>
          </div>
        ))}
        <button type="button" onClick={() => setMats([...mats, { name: '', quantity: '', unitPrice: '' }])} className="rounded-lg border border-dashed border-slate-400 px-3 py-1 text-xs font-bold text-slate-600">➕ ضيف مادة</button>
      </div>
      <div className="flex flex-wrap items-center gap-3 rounded-2xl bg-white p-4 text-sm ring-1 ring-slate-200">
        <span>الأجور: <b>{money(labor)}</b></span><span>المواد: <b>{money(matTotal)}</b></span>
        <label>الخصم: <input type="number" min={0} value={discount} onChange={(e) => setDiscount(e.target.value)} className="w-28 rounded-lg border border-slate-300 px-2 py-1" /></label>
        <span className="font-extrabold text-[#0f2040]">الصافي: {money(net)}</span>
        <span className="text-xs text-slate-400">(الحساب الأخير بالسيرفر)</span>
      </div>
      {err && <p className="text-sm font-bold text-red-600">{err}</p>}
      <div className="flex gap-2">
        <button type="button" disabled={busy} onClick={submit} className="rounded-xl bg-orange-600 px-6 py-2.5 font-extrabold text-white disabled:opacity-50">{busy ? 'جاري الإرسال…' : '📤 أرسل الفاتورة للمحاسب'}</button>
        <button type="button" onClick={onBack} className="rounded-xl border border-slate-300 px-4 py-2.5 text-sm">رجوع</button>
      </div>
    </div>
  )
}
