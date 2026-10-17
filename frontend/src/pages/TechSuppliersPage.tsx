import { useEffect, useMemo, useState } from 'react'
import { api, type TechSupplier, type TechSupplierIn } from '../api'
import { useSession } from '../session'

// ═══ 🏪 موردين التقنيين — قرار (ع) 10-06 ═══
// قائمة ثانية غير موردين الشركة. التقنيين ومسؤولي الخدمات يضيفون موردين،
// والمدير يختار لكل تقني شنو الموردين الي يطلعوله. المضاف ما يطلع لأحد —
// حتى لصاحبه — لحد ما المدير يختاره إله.

const empty: TechSupplierIn = { companyName: '', ownerName: '', phone: '', address: '', locationUrl: '', specialty: '', notes: '' }

function AddForm({ onDone }: { onDone: (msg: string) => void }) {
  const [f, setF] = useState<TechSupplierIn>(empty)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const set = (k: keyof TechSupplierIn) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => setF({ ...f, [k]: e.target.value })
  const save = async () => {
    setBusy(true); setErr('')
    try { const r = await api.addTechSupplier(f); setF(empty); onDone(r.message) } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }
  return (
    <div className="space-y-2 rounded-2xl border border-slate-200 bg-white p-4">
      <p className="text-sm font-extrabold text-slate-800">+ أضف مورّد</p>
      <div className="grid gap-2 sm:grid-cols-2">
        <input value={f.companyName} onChange={set('companyName')} placeholder="اسم المورّد / المحل *" className="rounded-lg border border-slate-300 p-2 text-sm" />
        <input value={f.phone} onChange={set('phone')} placeholder="رقم الهاتف *" inputMode="tel" className="rounded-lg border border-slate-300 p-2 text-sm" />
        <input value={f.ownerName} onChange={set('ownerName')} placeholder="اسم صاحبه" className="rounded-lg border border-slate-300 p-2 text-sm" />
        <input value={f.specialty} onChange={set('specialty')} placeholder="شنو يبيع (مثلاً: كيبلات، كاميرات، سويچات)" className="rounded-lg border border-slate-300 p-2 text-sm" />
        <input value={f.address} onChange={set('address')} placeholder="العنوان" className="rounded-lg border border-slate-300 p-2 text-sm" />
        <input value={f.locationUrl} onChange={set('locationUrl')} placeholder="رابط الموقع على الخريطة" className="rounded-lg border border-slate-300 p-2 text-sm" />
      </div>
      <textarea value={f.notes} onChange={set('notes')} rows={2} placeholder="ملاحظات (أسعاره، تعامله، أوقات دوامه…)" className="w-full rounded-lg border border-slate-300 p-2 text-sm" />
      {err && <p className="text-sm text-red-600">{err}</p>}
      <button type="button" disabled={busy || !f.companyName.trim() || !f.phone.trim()} onClick={() => void save()} className="rounded-lg bg-emerald-700 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">أضف</button>
    </div>
  )
}

function Card({ s, children }: { s: TechSupplier; children?: React.ReactNode }) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-3 text-sm">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div>
          <b className="text-base">{s.companyName}</b>
          {s.specialty && <span className="mr-2 rounded-full bg-sky-50 px-2 py-0.5 text-xs text-sky-800">{s.specialty}</span>}
          <p className="text-xs text-slate-500">{s.ownerName && `${s.ownerName} · `}<span dir="ltr" className="select-all font-bold text-slate-700">{s.phone}</span></p>
          {s.address && <p className="text-xs text-slate-600">📍 {s.address}</p>}
          {s.locationUrl && <a href={s.locationUrl} target="_blank" rel="noreferrer" className="text-xs text-sky-700 underline">افتح الموقع</a>}
          {s.notes && <p className="mt-1 text-xs text-slate-600">📝 {s.notes}</p>}
        </div>
        {children}
      </div>
    </div>
  )
}

function TechView() {
  const [rows, setRows] = useState<TechSupplier[] | null>(null)
  const [msg, setMsg] = useState('')
  useEffect(() => { void api.getMyTechSuppliers().then(setRows).catch(() => setRows([])) }, [])
  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <p className="text-sm font-extrabold text-slate-700">موردينك ({rows?.length ?? 0})</p>
        {rows?.length === 0 && <p className="rounded-xl bg-slate-50 p-3 text-sm text-slate-500">بعد ما اختارلك المدير موردين.</p>}
        <div className="grid gap-2 md:grid-cols-2">{rows?.map((s) => <Card key={s.id} s={s} />)}</div>
      </div>
      {msg && <p className="rounded-xl bg-emerald-50 p-3 text-sm font-bold text-emerald-800">{msg}</p>}
      <AddForm onDone={setMsg} />
    </div>
  )
}

function AdminView() {
  const [d, setD] = useState<Awaited<ReturnType<typeof api.getAllTechSuppliers>> | null>(null)
  const [who, setWho] = useState('')
  const [picked, setPicked] = useState<Set<string>>(new Set())
  const [saved, setSaved] = useState('')
  const [q, setQ] = useState('')
  const load = () => { void api.getAllTechSuppliers().then(setD) }
  useEffect(load, [])
  const pick = (id: string) => {
    setWho(id); setSaved('')
    setPicked(new Set((d?.suppliers ?? []).filter((s) => s.assignedTo.includes(id)).map((s) => s.id)))
  }
  const unassigned = useMemo(() => (d?.suppliers ?? []).filter((s) => s.assignedTo.length === 0).length, [d])
  if (!d) return <p className="text-slate-400">…</p>
  const list = d.suppliers.filter((s) => !q || `${s.companyName} ${s.specialty ?? ''} ${s.ownerName ?? ''}`.includes(q))
  const save = async () => {
    await api.setTechSupplierAccess(who, [...picked])
    setSaved('✅ انحفظ — هالموردين يطلعون لهالتقني هسه.'); load()
  }
  const del = async (id: string) => { await api.deleteTechSupplier(id); load() }
  return (
    <div className="space-y-4">
      <div className="rounded-2xl border border-violet-200 bg-violet-50/50 p-4">
        <p className="text-sm font-extrabold text-violet-900">اختار تقني أو مسؤول خدمة، وأشّر الموردين الي يطلعوله</p>
        <div className="mt-2 flex flex-wrap gap-2">
          {d.people.map((p) => (
            <button key={p.id} type="button" onClick={() => pick(p.id)}
              className={`rounded-full px-3 py-1 text-sm ${who === p.id ? 'bg-violet-700 text-white' : 'bg-white text-slate-700 ring-1 ring-slate-200'}`}>
              {p.name} <span className="text-[11px] opacity-70">· {p.kind}</span>
            </button>
          ))}
          {d.people.length === 0 && <p className="text-sm text-slate-500">ماكو تقنيين أو مسؤولي خدمات.</p>}
        </div>
      </div>
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-sm font-extrabold text-slate-700">كل موردين التقنيين ({d.suppliers.length}){unassigned > 0 && <span className="mr-2 rounded-full bg-amber-100 px-2 text-xs text-amber-800">{unassigned} محد مختارله</span>}</p>
        <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="بحث…" className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm" />
      </div>
      <div className="grid gap-2 md:grid-cols-2">
        {list.map((s) => (
          <Card key={s.id} s={s}>
            <div className="flex flex-col items-end gap-1 text-xs">
              {who && (
                <label className="flex cursor-pointer items-center gap-1 rounded-lg bg-violet-50 px-2 py-1 font-bold text-violet-800">
                  <input type="checkbox" checked={picked.has(s.id)} onChange={(e) => { const n = new Set(picked); if (e.target.checked) n.add(s.id); else n.delete(s.id); setPicked(n) }} />
                  يطلعله
                </label>
              )}
              <span className="text-slate-400">أضافه {s.createdBy ?? '—'} · يطلع لـ{s.assignedTo.length}</span>
              <button type="button" onClick={() => void del(s.id)} className="text-red-600 underline">حذف</button>
            </div>
          </Card>
        ))}
      </div>
      {who && (
        <div className="sticky bottom-2 flex items-center justify-between gap-2 rounded-xl bg-[#0f2040] p-3 text-white shadow-lg">
          <span className="text-sm">{picked.size} مورّد مختار لـ{d.people.find((p) => p.id === who)?.name}</span>
          <button type="button" onClick={() => void save()} className="rounded-lg bg-emerald-500 px-4 py-1.5 text-sm font-bold">احفظ</button>
        </div>
      )}
      {saved && <p className="text-sm font-bold text-emerald-700">{saved}</p>}
      <AddForm onDone={() => load()} />
    </div>
  )
}

export default function TechSuppliersPage() {
  const { employee } = useSession()
  const admin = employee?.role === 'ADMIN' || employee?.actualRole === 'OWNER'
  return (
    <div dir="rtl" className="space-y-4">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">🏪 موردين التقنيين</h2>
        <p className="text-sm text-slate-500">{admin ? 'التقنيين ومسؤولي الخدمات يضيفون موردين، وإنت تختار لكل واحد شنو يطلعله.' : 'الموردين الي اختارهم المدير إلك. وإذا عندك مورّد جديد ضيفه، ويطلعلك بعد ما المدير يختاره.'}</p>
      </div>
      {admin ? <AdminView /> : <TechView />}
    </div>
  )
}
