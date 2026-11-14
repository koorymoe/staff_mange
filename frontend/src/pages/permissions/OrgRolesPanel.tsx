import { useCallback, useEffect, useState } from 'react'
import { apiRequest } from '../../api'
import { askInput } from '../../ui/dialogs'
import { notify } from '../../ui/notify'
import { roleLabel } from '../../roleLabels'
import { fmtDate } from '../../utils/date'

// ═══ «منين جاية» — الأدوار الجاية من المنصب بالهيكل (دازاي 11-15) ═══
// «شلت منها الصلاحيات وبعدهن الصلاحيات يرجعن»: الموظف بمنصب (مثلاً وحدة المراقبة) ياخذ دور كامل
// يفتح شاشات بدون أي صلاحية — فشيل الصلاحيات ما يأثر. هنا يبين كل دور ومنين جاي، وتكدر تشيله.
type OrgRole = { role: string; assignmentId: string; divisionName: string; unitName: string; position: string; acting: boolean; startedAt: string; revoked: boolean }
const POS: Record<string, string> = { DIRECTOR: 'مدير', DEPUTY: 'نائب', MEMBER: 'عضو' }

export function OrgRolesPanel({ employeeId }: { employeeId: string }) {
  const [rows, setRows] = useState<OrgRole[] | null>(null)
  const [busy, setBusy] = useState('')
  const load = useCallback(() => {
    apiRequest<OrgRole[]>(`/employees/${employeeId}/access/roles`).then(setRows).catch(() => setRows([]))
  }, [employeeId])
  useEffect(load, [load])
  if (!rows || rows.length === 0) return null
  const set = async (role: string, remove: boolean) => {
    const reason = await askInput(remove ? `ليش تشيل دور «${roleLabel(role)}» الجاي من المنصب؟` : 'سبب الإرجاع (اختياري)')
    if (reason === null) return
    setBusy(role)
    try {
      setRows(await apiRequest<OrgRole[]>(`/employees/${employeeId}/access/roles`, { method: 'POST', body: JSON.stringify({ role, remove, reason }) }))
      notify(remove ? 'انشال الدور — الشاشات الي چان يفتحها تنسد' : 'رجع الدور', 'success')
    } catch (e) { notify(e instanceof Error ? e.message : 'تعذر') } finally { setBusy('') }
  }
  return (
    <section className="my-3 rounded-xl border border-amber-200 bg-amber-50/60 p-3 text-sm" dir="rtl">
      <p className="font-extrabold text-amber-900">🏛️ أدوار جاية من المنصب بالهيكل</p>
      <p className="mb-2 text-xs text-amber-800">هاي الأدوار تفتح شاشات كاملة <b>بدون صلاحيات</b> — فشيل الصلاحيات تحت ما يسدّها. إذا ما تريده يشوفها، شيل الدور هنا.</p>
      <ul className="space-y-1.5">
        {rows.map((r) => (
          <li key={r.role + r.assignmentId} className={`flex flex-wrap items-center gap-2 rounded-lg bg-white px-3 py-2 ${r.revoked ? 'opacity-60' : ''}`}>
            <b>{roleLabel(r.role)}</b>
            <span className="text-xs text-slate-600">← {POS[r.position] ?? r.position}{r.acting ? ' (بالوكالة)' : ''} · {r.divisionName}{r.unitName ? ` / ${r.unitName}` : ''} · من {fmtDate(r.startedAt)}</span>
            {r.revoked && <span className="rounded-full bg-red-100 px-2 py-0.5 text-[11px] font-bold text-red-700">مشال</span>}
            <button type="button" disabled={busy === r.role} onClick={() => void set(r.role, !r.revoked)}
              className={`mr-auto rounded-lg px-3 py-1 text-xs font-bold disabled:opacity-50 ${r.revoked ? 'bg-emerald-600 text-white' : 'border border-red-300 text-red-700'}`}>
              {r.revoked ? '↩️ رجّع الدور' : '✖️ شيل الدور'}
            </button>
          </li>
        ))}
      </ul>
    </section>
  )
}
