import { useEffect, useState } from 'react'
import { api, apiRequest, type Employee, type Vehicle } from '../api'
import { roleLabel } from '../roleLabels'

// ═══ «🪪 السواق والبطاقات» — (ع) 10-17 ═══
// ضمن إدارة المركبات: إجازة سوق كل موظف وانتهاؤها، التعهد وتخويل السياقة،
// وبطاقات البنزين (لأي سيارة وبيد منو). الوقود والعطلات والتنظيف بشاشة السيارة.
type Driver = { employeeId: string; name: string; role: string; hasDrivingLicense: boolean; licenseNumber: string | null; licenseExpiry: string | null; authorized: boolean; authorizedVehicles: string | null; pledgeSignedAt: string | null; pledgeNote: string | null }
type Card = { id: string; number: string; provider: string | null; vehicleId: string | null; vehicleName: string | null; holderEmployeeId: string | null; holderName: string | null; monthlyLimit: number | null; active: boolean; notes: string | null }
const day = (s: string | null) => (s ? s.slice(0, 10) : '')

export default function VehicleDriversPage() {
  const [tab, setTab] = useState<'drivers' | 'cards'>('drivers')
  return (
    <div className="space-y-4" dir="rtl">
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#2c5aad] p-5 text-white shadow">
        <h1 className="text-2xl font-black">🪪 السواق وبطاقات البنزين</h1>
        <p className="mt-1 text-sm text-blue-100">إجازات السوق وانتهاؤها، التعهد وتخويل السياقة، وبطاقات البنزين.</p>
      </div>
      <div className="flex gap-2">
        {([['drivers', '🪪 السواق والتعهدات'], ['cards', '⛽ بطاقات البنزين']] as const).map(([k, l]) => (
          <button key={k} type="button" onClick={() => setTab(k)} className={`rounded-xl px-4 py-2 text-sm font-bold ${tab === k ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600 ring-1 ring-slate-200'}`}>{l}</button>
        ))}
      </div>
      {tab === 'drivers' ? <Drivers /> : <Cards />}
    </div>
  )
}

function Drivers() {
  const [rows, setRows] = useState<Driver[] | null>(null)
  const [edit, setEdit] = useState<Driver | null>(null)
  const [q, setQ] = useState('')
  useEffect(() => { apiRequest<Driver[]>('/vehicle-drivers').then(setRows).catch(() => setRows([])) }, [])
  const [[today, soon]] = useState(() => [new Date().toISOString().slice(0, 10), new Date(Date.now() + 30 * 86400000).toISOString().slice(0, 10)])
  const save = async () => {
    if (!edit) return
    try {
      setRows(await apiRequest<Driver[]>(`/vehicle-drivers/${edit.employeeId}`, { method: 'PUT', body: JSON.stringify({
        licenseNumber: edit.licenseNumber ?? '', licenseExpiry: day(edit.licenseExpiry), authorized: edit.authorized,
        authorizedVehicles: edit.authorizedVehicles ?? '', pledgeSignedAt: day(edit.pledgeSignedAt), pledgeNote: edit.pledgeNote ?? '',
      }) }))
      setEdit(null)
    } catch (e) { alert(e instanceof Error ? e.message : 'تعذر') }
  }
  if (!rows) return <p className="text-slate-400">جاري التحميل…</p>
  const list = rows.filter((r) => !q || r.name.includes(q))
  const expired = rows.filter((r) => r.licenseExpiry && day(r.licenseExpiry) < today).length
  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center gap-2 text-sm">
        <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="ابحث بالاسم" className="rounded-lg border border-slate-300 px-3 py-1.5" />
        <span className="rounded-full bg-emerald-50 px-3 py-1 font-bold text-emerald-800">مخوّلين: {rows.filter((r) => r.authorized).length}</span>
        {expired > 0 && <span className="rounded-full bg-red-50 px-3 py-1 font-bold text-red-700">إجازات منتهية: {expired}</span>}
      </div>
      <div className="overflow-x-auto rounded-2xl bg-white ring-1 ring-slate-200">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 text-xs text-slate-500"><tr><th className="p-2 text-right">الموظف</th><th className="p-2">إجازة السوق</th><th className="p-2">تنتهي</th><th className="p-2">التعهد</th><th className="p-2">التخويل</th><th className="p-2" /></tr></thead>
          <tbody>
            {list.map((r) => {
              const exp = day(r.licenseExpiry)
              const tone = !exp ? '' : exp < today ? 'text-red-700 font-bold' : exp < soon ? 'text-amber-700 font-bold' : ''
              return (
                <tr key={r.employeeId} className="border-t border-slate-100">
                  <td className="p-2"><b>{r.name}</b> <span className="text-xs text-slate-400">{roleLabel(r.role)}</span></td>
                  <td className="p-2 text-center">{r.licenseNumber || (r.hasDrivingLicense ? 'عنده (بلا رقم)' : '—')}</td>
                  <td className={`p-2 text-center tabular-nums ${tone}`}>{exp || '—'}{exp && exp < today ? ' ⛔ منتهية' : exp && exp < soon ? ' ⚠️ قريبة' : ''}</td>
                  <td className="p-2 text-center">{r.pledgeSignedAt ? `✅ ${day(r.pledgeSignedAt)}` : '—'}</td>
                  <td className="p-2 text-center">{r.authorized ? `✅ ${r.authorizedVehicles || 'كل السيارات'}` : '—'}</td>
                  <td className="p-2 text-center"><button type="button" onClick={() => setEdit({ ...r })} className="text-xs font-bold text-brand-700">تعديل</button></td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
      {edit && (
        <div className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-4" onClick={() => setEdit(null)}>
          <div className="w-full max-w-md space-y-2 rounded-2xl bg-white p-5 text-sm" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-extrabold text-[#0f2040]">🪪 {edit.name}</h3>
            <label className="block">رقم إجازة السوق<input value={edit.licenseNumber ?? ''} onChange={(e) => setEdit({ ...edit, licenseNumber: e.target.value })} className="mt-1 w-full rounded-lg border border-slate-300 px-2 py-1.5" /></label>
            <label className="block">تاريخ انتهاء الإجازة<input type="date" value={day(edit.licenseExpiry)} onChange={(e) => setEdit({ ...edit, licenseExpiry: e.target.value })} className="mt-1 w-full rounded-lg border border-slate-300 px-2 py-1.5" /></label>
            <label className="block">تاريخ توقيع التعهد<input type="date" value={day(edit.pledgeSignedAt)} onChange={(e) => setEdit({ ...edit, pledgeSignedAt: e.target.value })} className="mt-1 w-full rounded-lg border border-slate-300 px-2 py-1.5" /></label>
            <label className="block">ملاحظة التعهد<input value={edit.pledgeNote ?? ''} onChange={(e) => setEdit({ ...edit, pledgeNote: e.target.value })} className="mt-1 w-full rounded-lg border border-slate-300 px-2 py-1.5" /></label>
            <label className="flex items-center gap-2 font-bold"><input type="checkbox" checked={edit.authorized} onChange={(e) => setEdit({ ...edit, authorized: e.target.checked })} /> مخوّل يسوق سيارات الشركة</label>
            {edit.authorized && <label className="block">أي سيارات؟ (فارغ = كلها)<input value={edit.authorizedVehicles ?? ''} onChange={(e) => setEdit({ ...edit, authorizedVehicles: e.target.value })} className="mt-1 w-full rounded-lg border border-slate-300 px-2 py-1.5" /></label>}
            <div className="flex gap-2 pt-2">
              <button type="button" onClick={save} className="flex-1 rounded-lg bg-[#0f2040] py-2 font-bold text-white">حفظ</button>
              <button type="button" onClick={() => setEdit(null)} className="rounded-lg border border-slate-300 px-4 py-2">إلغاء</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function Cards() {
  const [rows, setRows] = useState<Card[] | null>(null)
  const [vehicles, setVehicles] = useState<Vehicle[]>([])
  const [emps, setEmps] = useState<Employee[]>([])
  const empty = { id: '', number: '', provider: '', vehicleId: '', holderEmployeeId: '', monthlyLimit: '', active: true, notes: '' }
  const [f, setF] = useState(empty)
  useEffect(() => {
    apiRequest<Card[]>('/fuel-cards').then(setRows).catch(() => setRows([]))
    api.getVehicles().then(setVehicles).catch(() => {})
    api.getEmployees().then((e) => setEmps(e.filter((x) => x.status === 'ACTIVE'))).catch(() => {})
  }, [])
  const save = async (x: typeof empty) => {
    try {
      setRows(await apiRequest<Card[]>('/fuel-cards', { method: 'PUT', body: JSON.stringify({ ...x, monthlyLimit: x.monthlyLimit ? Number(x.monthlyLimit) : null }) }))
      setF(empty)
    } catch (e) { alert(e instanceof Error ? e.message : 'تعذر') }
  }
  if (!rows) return <p className="text-slate-400">جاري التحميل…</p>
  return (
    <div className="space-y-3">
      <div className="grid gap-2 rounded-2xl bg-white p-3 text-sm ring-1 ring-slate-200 sm:grid-cols-6">
        <input value={f.number} onChange={(e) => setF({ ...f, number: e.target.value })} placeholder="رقم البطاقة" className="rounded-lg border border-slate-300 px-2 py-1.5" />
        <input value={f.provider} onChange={(e) => setF({ ...f, provider: e.target.value })} placeholder="الجهة (المحطة/الشركة)" className="rounded-lg border border-slate-300 px-2 py-1.5" />
        <select value={f.vehicleId} onChange={(e) => setF({ ...f, vehicleId: e.target.value })} className="rounded-lg border border-slate-300 px-2 py-1.5"><option value="">السيارة…</option>{vehicles.map((v) => <option key={v.id} value={v.id}>{v.name} {v.plateNumber}</option>)}</select>
        <select value={f.holderEmployeeId} onChange={(e) => setF({ ...f, holderEmployeeId: e.target.value })} className="rounded-lg border border-slate-300 px-2 py-1.5"><option value="">بيد منو…</option>{emps.map((x) => <option key={x.id} value={x.id}>{x.name}</option>)}</select>
        <input type="number" min={0} value={f.monthlyLimit} onChange={(e) => setF({ ...f, monthlyLimit: e.target.value })} placeholder="الحد الشهري د.ع" className="rounded-lg border border-slate-300 px-2 py-1.5" />
        <button type="button" onClick={() => save(f)} className="rounded-lg bg-emerald-600 px-3 py-1.5 font-bold text-white">{f.id ? 'حفظ' : '＋ أضف بطاقة'}</button>
      </div>
      <div className="overflow-x-auto rounded-2xl bg-white ring-1 ring-slate-200">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 text-xs text-slate-500"><tr><th className="p-2 text-right">البطاقة</th><th className="p-2">الجهة</th><th className="p-2">السيارة</th><th className="p-2">بيد</th><th className="p-2">الحد الشهري</th><th className="p-2">الحالة</th><th className="p-2" /></tr></thead>
          <tbody>
            {rows.length === 0 && <tr><td colSpan={7} className="p-6 text-center text-slate-400">ماكو بطاقات مسجّلة.</td></tr>}
            {rows.map((c) => (
              <tr key={c.id} className={`border-t border-slate-100 ${c.active ? '' : 'opacity-50'}`}>
                <td className="p-2 font-bold" dir="ltr">{c.number}</td><td className="p-2 text-center">{c.provider || '—'}</td>
                <td className="p-2 text-center">{c.vehicleName || '—'}</td><td className="p-2 text-center">{c.holderName || '—'}</td>
                <td className="p-2 text-center tabular-nums">{c.monthlyLimit != null ? c.monthlyLimit.toLocaleString('en-US') : '—'}</td>
                <td className="p-2 text-center">{c.active ? 'فعّالة' : 'موقوفة'}</td>
                <td className="p-2 text-center">
                  <button type="button" onClick={() => setF({ id: c.id, number: c.number, provider: c.provider ?? '', vehicleId: c.vehicleId ?? '', holderEmployeeId: c.holderEmployeeId ?? '', monthlyLimit: c.monthlyLimit != null ? String(c.monthlyLimit) : '', active: c.active, notes: c.notes ?? '' })} className="text-xs font-bold text-brand-700">تعديل</button>
                  <button type="button" onClick={() => save({ id: c.id, number: c.number, provider: c.provider ?? '', vehicleId: c.vehicleId ?? '', holderEmployeeId: c.holderEmployeeId ?? '', monthlyLimit: c.monthlyLimit != null ? String(c.monthlyLimit) : '', active: !c.active, notes: c.notes ?? '' })} className="ms-3 text-xs font-bold text-slate-500">{c.active ? 'أوقف' : 'فعّل'}</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
