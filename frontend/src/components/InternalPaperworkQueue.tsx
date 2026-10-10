import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api'

// ═══ طابور الإداري: ورق الأعمال داخل الشركة (قرار (ع) 10-07) ═══
// «من يكمل الحجز الي داخل الشركة الإداري لازم يسوي الفاتورة والتقرير» —
// الفاتورة يدوية: هو يكتب شنو انعمل وهو يحدد السعر.
//
// (ع) 10-10: «من اضغط ع الحجز اسويله الحساب… ماريد تفرعات» — نافذة وحدة
// (شنو انعمل + السعر المقدّر) بدل صفحة الفاتورة الكاملة.
type Row = Awaited<ReturnType<typeof api.getInternalPaperwork>>[number]

export default function InternalPaperworkQueue() {
  const [rows, setRows] = useState<Row[] | null>(null)
  const [tick, setTick] = useState(0)
  const [open, setOpen] = useState<Row | null>(null)
  useEffect(() => {
    let alive = true
    api.getInternalPaperwork().then((r) => { if (alive) setRows(r) }).catch(() => { if (alive) setRows([]) })
    return () => { alive = false }
  }, [tick])
  if (!rows || rows.length === 0) return null
  return (
    <div dir="rtl" className="rounded-2xl border border-indigo-200 bg-indigo-50/60 p-4">
      <p className="font-extrabold text-indigo-900">🏢 أعمال داخل الشركة تنتظر ورقك ({rows.length})</p>
      <p className="mb-3 text-xs text-indigo-800">الفني خلّص الشغل — عليك الفاتورة (تكتب شنو انعمل وتحدد السعر) والتقرير.</p>
      <div className="space-y-2">
        {rows.map((r) => (
          <div key={r.id} className="flex flex-wrap items-center justify-between gap-2 rounded-xl bg-white px-3 py-2 text-sm">
            <span><b>{r.code}</b>{r.department && <span className="text-slate-500"> · {r.department}</span>}{r.doneAt && <span className="text-xs text-slate-400"> · خلص {r.doneAt}</span>}</span>
            <span className="flex gap-2">
              {r.hasInvoice
                ? <span className="rounded-lg bg-emerald-50 px-2 py-1 text-xs font-bold text-emerald-700">✓ الفاتورة</span>
                : <button type="button" onClick={() => setOpen(r)} className="rounded-lg bg-indigo-600 px-3 py-1.5 text-xs font-bold text-white">🧾 سوّي الحساب</button>}
              {r.hasReport
                ? <span className="rounded-lg bg-emerald-50 px-2 py-1 text-xs font-bold text-emerald-700">✓ التقرير</span>
                : <Link to={`/work-reports?bookingId=${r.id}`} className="rounded-lg border border-indigo-300 bg-white px-3 py-1.5 text-xs font-bold text-indigo-800">📝 سوّي التقرير</Link>}
            </span>
          </div>
        ))}
      </div>
      {open && <QuickCost row={open} onClose={() => setOpen(null)} onSaved={() => { setOpen(null); setTick((t) => t + 1) }} />}
    </div>
  )
}

function QuickCost({ row, onClose, onSaved }: { row: Row; onClose: () => void; onSaved: () => void }) {
  const [work, setWork] = useState('')
  const [price, setPrice] = useState('')
  const [note, setNote] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const save = async () => {
    const amount = Number(price.replace(/[^\d]/g, ''))
    if (work.trim().length < 10) { setErr('اكتب شنو انعمل بالتفصيل (10 أحرف على الأقل)'); return }
    if (!(amount > 0)) { setErr('اكتب السعر المقدّر'); return }
    setBusy(true)
    setErr(null)
    try {
      await api.createInternalInvoice({ bookingId: row.id, work: work.trim(), price: amount, note: note.trim() || undefined })
      onSaved()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally { setBusy(false) }
  }
  const input = 'w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-indigo-500'
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={onClose}>
      <div dir="rtl" className="w-full max-w-md space-y-3 rounded-2xl bg-white p-5 shadow-xl" onClick={(e) => e.stopPropagation()}>
        <p className="text-base font-extrabold text-indigo-900">🧾 حساب الكلفة — {row.code}{row.department && <span className="text-sm font-bold text-slate-500"> · {row.department}</span>}</p>
        <label className="block text-xs font-bold text-slate-600">شنو انعمل بالتفصيل *
          <textarea value={work} onChange={(e) => setWork(e.target.value)} rows={3} placeholder="مثلاً: شد 4 كاميرات بمخزن القسم + تمديد 60 متر" className={`mt-1 ${input}`} />
        </label>
        <label className="block text-xs font-bold text-slate-600">السعر المقدّر (د.ع) *
          <input value={price} onChange={(e) => setPrice(e.target.value)} inputMode="numeric" placeholder="0" className={`mt-1 ${input}`} />
        </label>
        <label className="block text-xs font-bold text-slate-600">ملاحظة (اختيارية)
          <input value={note} onChange={(e) => setNote(e.target.value)} className={`mt-1 ${input}`} />
        </label>
        {err && <p className="text-xs font-bold text-red-700">{err}</p>}
        <div className="flex gap-2">
          <button type="button" disabled={busy} onClick={save} className="flex-1 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">حفظ</button>
          <button type="button" onClick={onClose} className="rounded-lg border px-4 py-2 text-sm text-slate-600">إلغاء</button>
        </div>
      </div>
    </div>
  )
}
