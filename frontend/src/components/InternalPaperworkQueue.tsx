import Modal from '../ui/Modal'
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api'
import { parseAmount } from '../utils/money'
import { LineIcon } from './LineIcon'

// ═══ طابور الإداري: ورق الأعمال داخل الشركة (قرار (ع) 10-07) ═══
// «من يكمل الحجز الي داخل الشركة الإداري لازم يسوي الفاتورة والتقرير» —
// الفاتورة يدوية: هو يكتب شنو انعمل وهو يحدد السعر.
//
// (ع) 10-10: «من اضغط ع الحجز اسويله الحساب… ماريد تفرعات» — نافذة وحدة
// (شنو انعمل + السعر المقدّر) بدل صفحة الفاتورة الكاملة.
type Row = Awaited<ReturnType<typeof api.getInternalPaperwork>>[number]

// (ك 11-23) شكل الصف حسب صورة زينا: كود بكبسولة ذهبية، القسم بأيقونة طبقات، التاريخ والوقت،
// وأزرار «سوي الحساب» / «سوي التقرير» تصير خضرة بعلامة ✓ من تخلص. بحث بالقسم أو الكود، وتصفية.
const fmtDone = (s: string) => {
  const [d, t] = s.split(' ')
  const [y, m, dd] = (d ?? '').split('-')
  return { date: y && m && dd ? `${dd}-${m}-${y}` : d, time: t ?? '' }
}

type IpFilter = 'ALL' | 'INVOICE' | 'REPORT'

export default function InternalPaperworkQueue() {
  const [rows, setRows] = useState<Row[] | null>(null)
  const [tick, setTick] = useState(0)
  const [open, setOpen] = useState<Row | null>(null)
  const [q, setQ] = useState('')
  const [filter, setFilter] = useState<IpFilter>('ALL')
  const [filterOpen, setFilterOpen] = useState(false)
  useEffect(() => {
    let alive = true
    api.getInternalPaperwork().then((r) => { if (alive) setRows(r) }).catch(() => { if (alive) setRows([]) })
    return () => { alive = false }
  }, [tick])
  if (!rows || rows.length === 0) return null
  const term = q.trim().toLowerCase()
  const list = rows.filter((r) =>
    (!term || r.code.toLowerCase().includes(term) || (r.department ?? '').toLowerCase().includes(term))
    && (filter === 'ALL' || (filter === 'INVOICE' ? !r.hasInvoice : !r.hasReport)))
  const FILTERS: [IpFilter, string][] = [['ALL', 'الكل'], ['INVOICE', 'تنتظر الحساب'], ['REPORT', 'تنتظر التقرير']]
  return (
    <div dir="rtl" className="ip-wrap">
      <p className="ip-title">🏢 أعمال داخل الشركة تنتظر ورقك ({rows.length})</p>
      <p className="ip-sub">الفني خلّص الشغل — عليك الفاتورة (تكتب شنو انعمل وتحدد السعر) والتقرير.</p>
      <div className="mb-3 flex flex-wrap items-center gap-3">
        <label className="ip-search">
          <LineIcon n="search" className="h-5 w-5 ip-gold" />
          <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="بحث في الأقسام أو الأكواد..." aria-label="بحث في الأقسام أو الأكواد" />
        </label>
        <div className="relative">
          <button type="button" onClick={() => setFilterOpen((o) => !o)} className="ip-filter">
            <LineIcon n="filter" className="h-4 w-4 ip-gold" />التصفية{filter !== 'ALL' && <span className="ip-filter-dot" />}
          </button>
          {filterOpen && (
            <div className="ip-menu">
              {FILTERS.map(([k, t]) => (
                <button key={k} type="button" onClick={() => { setFilter(k); setFilterOpen(false) }} className={filter === k ? 'ip-menu-on' : ''}>{t}</button>
              ))}
            </div>
          )}
        </div>
      </div>
      <div className="space-y-2">
        {list.map((r) => {
          const d = r.doneAt ? fmtDone(r.doneAt) : null
          return (
            <div key={r.id} className="ip-row">
              <span className="ip-code">{r.code}</span>
              <span className="ip-dept"><LineIcon n="layers" className="h-6 w-6 ip-gold" />{r.department ?? '—'}</span>
              {d && <span className="ip-when"><LineIcon n="cal" className="h-5 w-5" /><span className="tabular-nums" dir="ltr">{d.date} {d.time}</span><LineIcon n="clock" className="h-5 w-5" /></span>}
              <span className="ms-auto flex gap-3">
                <span className="ip-slot">{r.hasInvoice
                  ? <span className="ip-btn ip-done">✓ الفاتورة</span>
                  : <button type="button" onClick={() => setOpen(r)} className="ip-btn ip-blue"><LineIcon n="calc" className="h-4 w-4" />سوي الحساب</button>}</span>
                <span className="ip-slot">{r.hasReport
                  ? <span className="ip-btn ip-done">✓ التقرير</span>
                  : <Link to={`/work-reports?bookingId=${r.id}`} className="ip-btn ip-purple"><LineIcon n="file" className="h-4 w-4" />سوي التقرير</Link>}</span>
              </span>
            </div>
          )
        })}
        {list.length === 0 && <p className="ip-sub py-3 text-center">ماكو نتائج</p>}
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
    const amount = parseAmount(price)
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
    <Modal title={<>🧾 حساب الكلفة — {row.code}{row.department && <span className="text-sm font-bold text-slate-500"> · {row.department}</span>}</>} onClose={onClose} size="md" busy={busy} bodyClassName="space-y-3 p-5">
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
          <button type="button" disabled={busy} onClick={save} className="flex-1 rounded-lg bg-[#0f2040] px-4 py-2 text-sm font-bold text-white disabled:opacity-50">حفظ</button>
          <button type="button" onClick={onClose} className="rounded-lg border px-4 py-2 text-sm text-slate-600">إلغاء</button>
        </div>
    </Modal>
  )
}
