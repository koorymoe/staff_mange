import { useEffect, useState } from 'react'
import { api, type Employee } from '../api'
import { roleLabel } from '../roleLabels'

// ═══ «🔁 تسليم شغله لموظف جديد» — (ع) 10-17 ═══
// «من يطلع أحد الموظفين، نختار موظف ويترحّل كلشي إله»: المنصب والصلاحيات
// والشغل المفتوح ينتقلون للمستلم، والقديم يتوقّف (ما ينحذف).
type Summary = { position: string; permissions: number; bookings: number; missions: number; projects: number; complaints: number; extraTasks: number; deptHead: number }

export default function EmployeeHandoverDialog({ from, employees, onClose, onDone }: { from: Employee; employees: Employee[]; onClose: () => void; onDone: () => void }) {
  const [to, setTo] = useState('')
  const [pre, setPre] = useState<Summary | null>(null)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  useEffect(() => { api.getHandoverPreview(from.id).then(setPre).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر')) }, [from.id])
  const pick = employees.filter((e) => e.id !== from.id && e.status === 'ACTIVE' && e.role !== 'ADMIN')
  const lines = pre ? [
    `المنصب: ${roleLabel(pre.position)}`, `${pre.permissions} صلاحية`, `${pre.bookings} حجز مفتوح`, `${pre.missions} مهمة ميدانية`,
    `${pre.projects} مشروع يشرف عليه`, `${pre.complaints} شكوى مكلّف بيها`, `${pre.extraTasks} مهمة إضافية`,
    ...(pre.deptHead ? ['مسؤولية القسم'] : []),
  ] : []
  const go = async () => {
    if (!to) { setErr('اختار الموظف المستلم'); return }
    setBusy(true); setErr('')
    try { await api.handoverEmployee(from.id, to); onDone() } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }
  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-4" dir="rtl" onClick={onClose}>
      <div className="w-full max-w-md rounded-2xl bg-white p-5 shadow-xl" onClick={(e) => e.stopPropagation()}>
        <h3 className="text-lg font-extrabold text-[#0f2040]">🔁 تسليم شغل {from.name} لموظف جديد</h3>
        <p className="mt-1 text-xs text-slate-500">ينتقل للمستلم كلشي، و{from.name} يتوقّف حسابه — ما ينحذف، سجله وفواتيره تبقى باسمه.</p>
        <select value={to} onChange={(e) => setTo(e.target.value)} className="mt-3 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
          <option value="">— اختار المستلم —</option>
          {pick.map((e) => <option key={e.id} value={e.id}>{e.name} ({roleLabel(e.role)})</option>)}
        </select>
        {pre && (
          <div className="mt-3 rounded-lg bg-slate-50 p-3 text-sm">
            <p className="mb-1 font-bold text-slate-700">راح ينتقل:</p>
            <ul className="list-inside list-disc text-slate-600">{lines.map((l) => <li key={l}>{l}</li>)}</ul>
          </div>
        )}
        {err && <p className="mt-2 text-sm font-bold text-red-600">{err}</p>}
        <div className="mt-4 flex gap-2">
          <button type="button" disabled={busy || !to} onClick={go} className="flex-1 rounded-lg bg-[#0f2040] py-2 text-sm font-bold text-white disabled:opacity-50">{busy ? 'جاري التسليم…' : 'سلّم الشغل'}</button>
          <button type="button" onClick={onClose} className="rounded-lg border border-slate-300 px-4 py-2 text-sm">إلغاء</button>
        </div>
      </div>
    </div>
  )
}
