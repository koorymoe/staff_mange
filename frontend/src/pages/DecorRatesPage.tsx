import { useEffect, useState } from 'react'
import { api, type DecorRate, type Service } from '../api'

// ═══ جدول أسعار الديكور — (ع) 10-16 ═══
// لكل مهنة بنودها بسعرها (متر/قطعة/يوم). فاتورة الديكور تنحسب منه × القياسات —
// ماكو سعر بالإيد، نفس حماية فواتير الشعبة الهندسية.
const UNITS = ['متر', 'متر مربع', 'قطعة', 'كغم', 'يوم']
const empty = { id: '', serviceId: '', name: '', unit: 'متر', price: '', active: true }

export default function DecorRatesPage() {
  const [rows, setRows] = useState<DecorRate[]>([])
  const [services, setServices] = useState<Service[]>([])
  const [f, setF] = useState(empty)
  const [msg, setMsg] = useState('')
  useEffect(() => {
    api.getDecorRates(true).then(setRows).catch(() => {})
    api.getServices().then((s) => setServices(s.filter((x) => x.division === 'DECOR'))).catch(() => {})
  }, [])
  const save = async (x: typeof empty) => {
    setMsg('')
    try {
      setRows(await api.saveDecorRate({ ...x, price: Number(x.price) }))
      setF(empty); setMsg('انحفظ ✓')
    } catch (e) { setMsg(e instanceof Error ? e.message : 'تعذر') }
  }
  return (
    <div className="space-y-4" dir="rtl">
      <div className="rounded-2xl bg-gradient-to-l from-orange-600 to-amber-500 p-5 text-white shadow">
        <h1 className="text-2xl font-black">🪚 جدول أسعار الديكور</h1>
        <p className="mt-1 text-sm text-orange-50">لكل مهنة بنودها بسعرها. فاتورة الديكور تنحسب من هذا الجدول × القياسات الي يكتبها الليدر.</p>
      </div>
      <div className="grid gap-2 rounded-2xl bg-white p-4 ring-1 ring-slate-200 sm:grid-cols-[1fr_1fr_8rem_8rem_auto]">
        <select value={f.serviceId} onChange={(e) => setF({ ...f, serviceId: e.target.value })} className="rounded-lg border border-slate-300 px-2 py-2 text-sm">
          <option value="">المهنة…</option>{services.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
        </select>
        <input value={f.name} onChange={(e) => setF({ ...f, name: e.target.value })} placeholder="البند (مثلاً قاعدة حديد)" className="rounded-lg border border-slate-300 px-2 py-2 text-sm" />
        <select value={f.unit} onChange={(e) => setF({ ...f, unit: e.target.value })} className="rounded-lg border border-slate-300 px-2 py-2 text-sm">{UNITS.map((u) => <option key={u}>{u}</option>)}</select>
        <input type="number" min={0} value={f.price} onChange={(e) => setF({ ...f, price: e.target.value })} placeholder="السعر د.ع" className="rounded-lg border border-slate-300 px-2 py-2 text-sm" />
        <button type="button" onClick={() => save(f)} className="rounded-lg bg-orange-600 px-4 py-2 text-sm font-bold text-white">{f.id ? 'حفظ التعديل' : '＋ أضف'}</button>
      </div>
      {msg && <p className="text-sm font-bold text-slate-600">{msg}</p>}
      <div className="overflow-x-auto rounded-2xl bg-white ring-1 ring-slate-200">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 text-xs text-slate-500"><tr><th className="p-2 text-right">المهنة</th><th className="p-2 text-right">البند</th><th className="p-2">الوحدة</th><th className="p-2">السعر</th><th className="p-2">الحالة</th><th className="p-2" /></tr></thead>
          <tbody>
            {rows.length === 0 && <tr><td colSpan={6} className="p-6 text-center text-slate-400">ماكو بنود بعد — ضيف أول بند فوك.</td></tr>}
            {rows.map((r) => (
              <tr key={r.id} className={`border-t border-slate-100 ${r.active ? '' : 'opacity-50'}`}>
                <td className="p-2">{r.serviceName}</td><td className="p-2 font-bold">{r.name}</td><td className="p-2 text-center">{r.unit}</td>
                <td className="p-2 text-center tabular-nums">{r.price.toLocaleString('en-US')}</td>
                <td className="p-2 text-center">{r.active ? 'فعّال' : 'معطّل'}</td>
                <td className="p-2 text-center">
                  <button type="button" onClick={() => setF({ id: r.id, serviceId: r.serviceId, name: r.name, unit: r.unit, price: String(r.price), active: r.active })} className="text-xs font-bold text-brand-700">تعديل</button>
                  <button type="button" onClick={() => save({ id: r.id, serviceId: r.serviceId, name: r.name, unit: r.unit, price: String(r.price), active: !r.active })} className="ms-3 text-xs font-bold text-slate-500">{r.active ? 'عطّل' : 'فعّل'}</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
