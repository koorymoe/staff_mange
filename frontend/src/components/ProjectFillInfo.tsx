import { useEffect, useState } from 'react'
import { api, type Employee } from '../api'

// ═══ «أكمل معلومات المشروع» — طلب (ع) 10-10 ═══
// «خيار كدام كل مشروع يخلون المبلغ المالي والمستلم واسم المشرف… وكل
// المعلومات المبهمة تنكتب لمرة وحدة». الدفعة من هنا تروح للمحاسب يأكدها
// (إلا إذا الي يسجّلها محاسب) — نفس مسار الدفعات بالضبط.

export interface FillTarget {
  id: string; code: string; name: string
  priceRaw: string | null; responsibleId: string | null
  paidAmount: number
}

const today = () => new Date().toISOString().slice(0, 10)

export default function ProjectFillInfo({ p, onClose, onSaved }: { p: FillTarget; onClose: () => void; onSaved: () => void }) {
  const [emps, setEmps] = useState<Employee[]>([])
  const [price, setPrice] = useState(p.priceRaw ?? '')
  const [resp, setResp] = useState(p.responsibleId ?? '')
  const [amount, setAmount] = useState('')
  const [paidAt, setPaidAt] = useState(today)
  const [method, setMethod] = useState('CASH')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    api.getEmployees().then((r) => setEmps(r.filter((e) => e.status === 'ACTIVE').sort((a, b) => a.name.localeCompare(b.name, 'ar')))).catch(() => setEmps([]))
  }, [])

  const save = async () => {
    const amt = Number(amount.replace(/[^\d]/g, ''))
    setBusy(true)
    setErr(null)
    try {
      if (price.trim() !== (p.priceRaw ?? '') || resp !== (p.responsibleId ?? '')) {
        await api.fillProjectInfo(p.id, { price: price.trim(), responsibleId: resp })
      }
      if (amt > 0) await api.addProjectPayment(p.id, { amount: amt, paidAt, method })
      onSaved()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally { setBusy(false) }
  }

  const input = 'w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-[var(--color-brand-500)]'
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={onClose}>
      <div className="w-full max-w-md space-y-3 rounded-2xl bg-white p-5 shadow-xl" onClick={(e) => e.stopPropagation()}>
        <p className="text-base font-extrabold text-[var(--color-brand-900)]">✏️ أكمل معلومات «{p.name}» <span className="text-xs text-slate-400">{p.code}</span></p>
        <label className="block text-xs font-bold text-slate-600">💵 القيمة المالية (سعر المشروع)
          <input value={price} onChange={(e) => setPrice(e.target.value)} inputMode="numeric" placeholder="مثلاً 15,000,000" className={`mt-1 ${input}`} />
        </label>
        <label className="block text-xs font-bold text-slate-600">👷 المشرف / المسؤول
          <select value={resp} onChange={(e) => setResp(e.target.value)} className={`mt-1 ${input}`}>
            <option value="">— اختار —</option>
            {emps.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
          </select>
        </label>
        <div className="rounded-xl bg-slate-50 p-3">
          <p className="mb-2 text-xs font-bold text-slate-600">💰 دفعة مستلمة {p.paidAmount > 0 && <span className="font-normal text-slate-400">(المسجّل هسه {p.paidAmount.toLocaleString('en-US')} د.ع — اكتب بس الدفعة الجديدة)</span>}</p>
          <div className="grid grid-cols-2 gap-2">
            <input value={amount} onChange={(e) => setAmount(e.target.value)} inputMode="numeric" placeholder="المبلغ" className={input} />
            <input type="date" value={paidAt} onChange={(e) => setPaidAt(e.target.value)} className={input} />
            <select value={method} onChange={(e) => setMethod(e.target.value)} className={`col-span-2 ${input}`}>
              <option value="CASH">نقد</option><option value="TRANSFER">تحويل</option><option value="CHEQUE">صك</option>
            </select>
          </div>
          <p className="mt-1 text-[10px] text-amber-700">⏳ الدفعة تروح للمحاسب يأكدها.</p>
        </div>
        {err && <p className="text-xs font-bold text-red-700">{err}</p>}
        <div className="flex gap-2">
          <button type="button" disabled={busy} onClick={save} className="flex-1 rounded-lg bg-[var(--color-brand-500)] px-4 py-2 text-sm font-bold text-white disabled:opacity-50">حفظ</button>
          <button type="button" onClick={onClose} className="rounded-lg border px-4 py-2 text-sm text-slate-600">إلغاء</button>
        </div>
      </div>
    </div>
  )
}
