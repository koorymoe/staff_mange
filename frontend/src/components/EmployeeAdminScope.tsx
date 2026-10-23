import { useEffect, useState } from 'react'
import { api } from '../api'
import { ROLE_LABELS } from '../roleLabels'

// ═══ نطاق المفوَّض بإدارة الموظفين — قرار (ع) 10-11 ═══
// «الموظف الي انطيه صلاحية يضيف موظف، اني احدد شنو الأدوار الي يكدر
// يضيفهم». نفس النطاق يحكم «تعديل كل الموظفين». بلا اختيار = ما يكدر.
export default function EmployeeAdminScope({ employeeId }: { employeeId: string }) {
  const [roles, setRoles] = useState<string[] | null>(null)
  const [saving, setSaving] = useState(false)
  const [msg, setMsg] = useState('')
  useEffect(() => {
    let alive = true
    api.getEmployeeAdminScope(employeeId).then((r) => { if (alive) setRoles(r.roles) }).catch(() => { if (alive) setRoles([]) })
    return () => { alive = false }
  }, [employeeId])
  if (!roles) return null
  const all = Object.keys(ROLE_LABELS).filter((r) => r !== 'ADMIN' && r !== 'OWNER')
  const toggle = (r: string) => setRoles((x) => (x?.includes(r) ? x.filter((y) => y !== r) : [...(x ?? []), r]))
  const save = async () => {
    setSaving(true); setMsg('')
    try { const r = await api.setEmployeeAdminScope(employeeId, roles); setRoles(r.roles); setMsg('✓ انحفظ') }
    catch (e) { setMsg(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setSaving(false) }
  }
  return (
    <div className="mb-6 rounded-xl border border-amber-200 bg-amber-50/60 p-4">
      <p className="font-bold text-amber-900">👥 الأدوار الي يكدر يضيفها ويعدّلها</p>
      <p className="mb-3 text-xs text-amber-800">عنده صلاحية إضافة أو تعديل موظفين — ما مؤشر شي = يضيف ويعدّل كل الأدوار. أشّر أدوار إذا تريد تقيّده بيها بس. المدير والمالك ما ينطون أبداً، والصلاحيات تبقى بيدك.</p>
      <div className="flex flex-wrap gap-2">
        {all.map((r) => (
          <label key={r} className={`cursor-pointer rounded-full border px-3 py-1 text-xs font-bold ${roles.includes(r) ? 'border-amber-500 bg-amber-500 text-white' : 'border-slate-300 bg-white text-slate-600'}`}>
            <input type="checkbox" className="hidden" checked={roles.includes(r)} onChange={() => toggle(r)} />
            {ROLE_LABELS[r as keyof typeof ROLE_LABELS]}
          </label>
        ))}
      </div>
      <div className="mt-3 flex items-center gap-3">
        <button type="button" disabled={saving} onClick={() => void save()} className="rounded-lg bg-amber-600 px-4 py-1.5 text-sm font-bold text-white disabled:opacity-50">حفظ الأدوار</button>
        {msg && <span className="text-xs font-bold text-slate-600">{msg}</span>}
      </div>
    </div>
  )
}
