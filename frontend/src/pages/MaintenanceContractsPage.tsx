import { useEffect, useState } from 'react'
import { api, apiRequest, type Customer, type Service } from '../api'

// ═══ 🔄 عقود الصيانة الدورية — (ع) 11-08 ═══
// زبون عنده عقد (كل كم شهر): قبل موعد الزيارة بأسبوع النظام يفتح الحجز لحاله
// ويوصل للإداري ينسّقه. الإيقاف بدل الحذف.
type Contract = {
  id: string; customerId: string; customerName: string | null; customerCode: string | null
  serviceId: string; serviceName: string | null; intervalMonths: number; nextVisitDate: string
  address: string | null; notes: string | null; active: boolean
  lastBookingCode: string | null; lastGeneratedAt: string | null; lastError: string | null
}
const INTERVALS = [1, 2, 3, 4, 6, 12]
const inp = 'w-full rounded-lg border border-slate-200 p-2 text-sm'

export default function MaintenanceContractsPage() {
  const [rows, setRows] = useState<Contract[] | null>(null)
  const [services, setServices] = useState<Service[]>([])
  const [err, setErr] = useState('')
  const [reload, setReload] = useState(0)
  const [show, setShow] = useState(false)
  const [q, setQ] = useState('')
  const [found, setFound] = useState<Customer[]>([])
  const [form, setForm] = useState({ customerId: '', customerLabel: '', serviceId: '', intervalMonths: 3, nextVisitDate: '', address: '', notes: '' })
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    let alive = true
    apiRequest<Contract[]>('/maintenance-contracts')
      .then((r) => { if (alive) setRows(r) })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [reload])
  useEffect(() => { api.getServices().then(setServices).catch(() => {}) }, [])
  useEffect(() => {
    if (q.trim().length < 2) return
    let alive = true
    const t = setTimeout(() => { api.getCustomers({ search: q.trim(), limit: 8 }).then((r) => { if (alive) setFound(r) }).catch(() => {}) }, 300)
    return () => { alive = false; clearTimeout(t) }
  }, [q])

  const create = async () => {
    setBusy(true); setErr('')
    try {
      await apiRequest('/maintenance-contracts', { method: 'POST', body: JSON.stringify({
        customerId: form.customerId, serviceId: form.serviceId, intervalMonths: form.intervalMonths,
        nextVisitDate: form.nextVisitDate, address: form.address, notes: form.notes,
      }) })
      setShow(false); setForm({ customerId: '', customerLabel: '', serviceId: '', intervalMonths: 3, nextVisitDate: '', address: '', notes: '' }); setQ('')
      setReload((n) => n + 1)
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }
  const patch = async (id: string, body: Record<string, unknown>) => {
    setErr('')
    try { await apiRequest(`/maintenance-contracts/${id}`, { method: 'PATCH', body: JSON.stringify(body) }); setReload((n) => n + 1) }
    catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') }
  }

  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex flex-wrap items-end justify-between gap-2">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">🔄 عقود الصيانة الدورية</h2>
          <p className="text-sm text-slate-500">قبل موعد كل زيارة بأسبوع، النظام يفتح الحجز لحاله ويوصلك إشعار حتى تثبّته ويا الزبون.</p>
        </div>
        <button type="button" onClick={() => setShow(!show)} className="rounded-xl bg-[#0f2040] px-4 py-2 text-sm font-bold text-white">+ عقد جديد</button>
      </div>
      {err && <p className="text-sm text-red-600">{err}</p>}
      {show && (
        <div className="space-y-2 rounded-2xl border border-slate-200 bg-white p-4">
          {form.customerId ? (
            <p className="text-sm">الزبون: <b>{form.customerLabel}</b> <button type="button" className="ms-2 text-xs text-slate-500 underline" onClick={() => setForm({ ...form, customerId: '', customerLabel: '' })}>غيّر</button></p>
          ) : (
            <div>
              <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="دوّر الزبون بالاسم أو الهاتف أو الكود" className={inp} />
              {q.trim().length >= 2 && found.length > 0 && (
                <div className="mt-1 max-h-48 overflow-y-auto rounded-lg border border-slate-100">
                  {found.map((c) => (
                    <button key={c.id} type="button" onClick={() => setForm({ ...form, customerId: c.id, customerLabel: `${c.name} (${c.customerCode})`, address: form.address || c.location || '' })}
                      className="block w-full border-b border-slate-50 px-3 py-2 text-right text-sm hover:bg-slate-50">{c.name} <span className="text-xs text-slate-400">· {c.customerCode}</span></button>
                  ))}
                </div>
              )}
            </div>
          )}
          <div className="grid gap-2 sm:grid-cols-3">
            <select value={form.serviceId} onChange={(e) => setForm({ ...form, serviceId: e.target.value })} className={inp}>
              <option value="">الخدمة</option>{services.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
            </select>
            <select value={form.intervalMonths} onChange={(e) => setForm({ ...form, intervalMonths: Number(e.target.value) })} className={inp}>
              {INTERVALS.map((m) => <option key={m} value={m}>كل {m} {m === 1 ? 'شهر' : 'أشهر'}</option>)}
            </select>
            <label className="text-xs text-slate-500">موعد أول زيارة<input type="date" value={form.nextVisitDate} onChange={(e) => setForm({ ...form, nextVisitDate: e.target.value })} className={inp} /></label>
          </div>
          <input value={form.address} onChange={(e) => setForm({ ...form, address: e.target.value })} placeholder="العنوان (اختياري)" className={inp} />
          <textarea value={form.notes} onChange={(e) => setForm({ ...form, notes: e.target.value })} rows={2} placeholder="ملاحظات العقد (شنو تشمل الزيارة…)" className={inp} />
          <button type="button" disabled={busy || !form.customerId || !form.serviceId || !form.nextVisitDate} onClick={() => void create()} className="rounded-lg bg-emerald-700 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">احفظ العقد</button>
        </div>
      )}
      {!rows && !err && <p className="text-slate-400">جاري التحميل…</p>}
      {rows && rows.length === 0 && <p className="rounded-xl bg-white p-8 text-center text-slate-400">ماكو عقود بعد.</p>}
      {rows?.map((c) => (
        <div key={c.id} className={`rounded-2xl border bg-white p-4 text-sm ${c.active ? 'border-slate-200' : 'border-slate-100 opacity-60'}`}>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <span className="font-bold text-slate-800">{c.customerName} <span className="text-xs font-normal text-slate-400">· {c.customerCode}</span></span>
            <span className={`rounded-full px-2 py-0.5 text-xs font-bold ${c.active ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-100 text-slate-500'}`}>{c.active ? 'فعّال' : 'موقوف'}</span>
          </div>
          <p className="mt-1 text-slate-600">{c.serviceName} · كل {c.intervalMonths} {c.intervalMonths === 1 ? 'شهر' : 'أشهر'} · الزيارة الجاية: <b>{c.nextVisitDate}</b></p>
          {c.lastBookingCode && <p className="text-xs text-slate-500">آخر حجز انفتح: {c.lastBookingCode}</p>}
          {c.lastError && <p className="text-xs font-bold text-red-600">⚠️ آخر مرة ما گدر يفتح الحجز: {c.lastError} — افتحه بإيدك.</p>}
          {c.notes && <p className="text-xs text-slate-500">{c.notes}</p>}
          <div className="mt-2 flex flex-wrap items-center gap-2 text-xs">
            <label>الزيارة الجاية <input type="date" defaultValue={c.nextVisitDate} onBlur={(e) => { if (e.target.value && e.target.value !== c.nextVisitDate) void patch(c.id, { nextVisitDate: e.target.value }) }} className="rounded-lg border border-slate-200 px-2 py-1" /></label>
            <button type="button" onClick={() => void patch(c.id, { active: !c.active })} className="rounded-lg border border-slate-200 px-3 py-1 font-bold">{c.active ? '⏸️ أوقف العقد' : '▶️ رجّعه'}</button>
          </div>
        </div>
      ))}
    </div>
  )
}
