import { useCallback, useEffect, useState, type ReactElement } from 'react'
import { api, type Employee } from '../api'
import { navItems, isNavVisible, type NavItem } from '../components/Layout'

// ═══ شوف النظام بعين الموظف ═══
//
// سؤال صاحب العمل حرفياً: «لو أنطيته هاي الصلاحية شراح يطلع عنده من
// شغله؟». قبل، الجواب الوحيد إنك تسجّل دخول بحسابه وتشوف — وهذا ما
// يصير، فالقرار كان يتاخذ بالحدس.
//
// ⚠️ هاي الشاشة تنادي **نفس الدالة** الي ترسم القائمة الحقيقية
// (isNavVisible). لو حسبناها بمنطق مستقل چان انحرفت عن الواقع بأول
// تعديل، وصارت تكذب — وشاشة تكذب أسوأ من ماكو شاشة.
export default function PermissionPreview() {
  const [employees, setEmployees] = useState<Employee[]>([])
  const [selected, setSelected] = useState<string>('')
  const [perms, setPerms] = useState<string[]>([])
  const [gpsServiceId, setGpsServiceId] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api.getEmployees().then((e) => setEmployees(e.filter((x) => x.status === 'ACTIVE'))).catch(() => {})
    api.getServices().then((svc) => {
      const gps = svc.find((s) => s.name.includes('جي بي') || s.name.toUpperCase().includes('GPS'))
      setGpsServiceId(gps?.id || null)
    }).catch(() => {})
  }, [])

  const load = useCallback(async (id: string) => {
    if (!id) { setPerms([]); return }
    setLoading(true)
    setError(null)
    try {
      const rows = await api.getEmployeePermissions(id)
      setPerms(rows.map((p) => p.name))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذر جلب صلاحيات الموظف')
    } finally { setLoading(false) }
  }, [])

  useEffect(() => { void load(selected) }, [selected, load])

  const emp = employees.find((e) => e.id === selected)
  const ctx = { employee: emp, permissions: perms, gpsServiceId }

  // نمشي بنفس شجرة القائمة ونأشّر شنو يظهر وشنو لا — والسبب
  const walk = (items: NavItem[], depth = 0, unitGranted = false): ReactElement[] =>
    items.flatMap((item) => {
      if (item.divider) return []
      const visible = emp ? isNavVisible(item, ctx, unitGranted) : false
      const childGranted =
        unitGranted || (!!item.unitPermission && perms.includes(item.unitPermission))
      const key = (item.to || item.label) + depth
      const rows = [
        <div
          key={key}
          className={`flex items-center justify-between gap-2 rounded-lg px-3 py-1.5 text-sm ${
            visible ? 'bg-emerald-50 text-emerald-900' : 'bg-slate-50 text-slate-400'
          }`}
          style={{ marginInlineStart: depth * 16 }}
        >
          <span className="font-medium">
            {visible ? '✓' : '✕'} {item.label}
          </span>
          <span className="shrink-0 text-[10px] text-slate-500">
            {item.unlockPermission || item.permission || item.unitPermission || (item.roles ? 'بالدور' : '')}
          </span>
        </div>,
      ]
      if (item.children) rows.push(...walk(item.children, depth + 1, childGranted))
      return rows
    })

  const flat = (items: NavItem[]): NavItem[] =>
    items.flatMap((i) => (i.divider ? [] : [i, ...(i.children ? flat(i.children) : [])]))
  const all = flat(navItems)
  const visibleCount = emp ? all.filter((i) => isNavVisible(i, ctx)).length : 0

  return (
    <div dir="rtl" className="space-y-5">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">🔎 شوف النظام بعين الموظف</h2>
        <p className="mt-1 text-slate-500">
          اختار موظف وشوف بالضبط شنو راح تطلعله القائمة — نفس الحساب الي يشتغل بيه النظام فعلاً.
        </p>
      </div>

      {error && <p className="rounded-lg bg-red-50 p-4 text-red-600">{error}</p>}

      <select
        value={selected}
        onChange={(e) => setSelected(e.target.value)}
        className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 sm:max-w-sm"
      >
        <option value="">— اختار موظف —</option>
        {employees.map((e) => (
          <option key={e.id} value={e.id}>{e.name} — {e.jobTitle || e.role}</option>
        ))}
      </select>

      {emp && (
        <>
          <div className="flex flex-wrap gap-3 rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <div>
              <p className="text-xs text-slate-500">الدور</p>
              <p className="font-bold text-[#0f2040]">{emp.role}{emp.isLeader && ' • تيم ليدر'}</p>
            </div>
            <div>
              <p className="text-xs text-slate-500">صلاحياته الممنوحة</p>
              <p className="font-bold text-[#0f2040]">{perms.length}</p>
            </div>
            <div>
              <p className="text-xs text-slate-500">شاشات يشوفها</p>
              <p className="font-bold text-emerald-700">{visibleCount} من {all.length}</p>
            </div>
          </div>

          {perms.length > 0 && (
            <div className="flex flex-wrap gap-1.5">
              {perms.map((p) => (
                <span key={p} className="rounded-full bg-brand-50 px-2.5 py-1 text-[11px] font-bold text-brand-800">{p}</span>
              ))}
            </div>
          )}

          {loading ? (
            <p className="text-slate-400">جاري التحميل...</p>
          ) : (
            <div className="space-y-1 rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              {walk(navItems)}
            </div>
          )}
        </>
      )}
    </div>
  )
}
