import { useEffect, useMemo, useState } from 'react'
import {
  api, IT_KIND_LABELS, IT_LOG_LABELS, IT_STATUS_LABELS,
  type Employee, type ItAsset, type ItAssetInput, type ItAssetKind, type ItAssetLog, type ItAssetStatus,
} from '../api'
import { matches } from '../utils/search'

// ═══ جرد أجهزة تقنية المعلومات ═══
// حاسبات ومخدّمات وشبكة وطابعات الشركة: وين كل جهاز، منو يستعمله،
// حالته، ضمانه، وسجل كل تصليح أو صيانة انسوّت عليه.

const STATUS_CLS: Record<ItAssetStatus, string> = {
  ACTIVE: 'bg-emerald-100 text-emerald-700',
  REPAIR: 'bg-amber-100 text-amber-800',
  SPARE: 'bg-sky-100 text-sky-700',
  RETIRED: 'bg-slate-200 text-slate-600',
}

const EMPTY: ItAssetInput = {
  name: '', kind: 'COMPUTER', status: 'ACTIVE', brand: null, model: null, serialNumber: null, ipAddress: null,
  location: null, assignedEmployeeId: null, purchaseDate: null, warrantyUntil: null, notes: null,
}

const day = (v: string | null) => (v ? v.slice(0, 10) : '')
const inp = 'w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500'

function AssetForm({ initial, employees, onSaved, onClose }: {
  initial: ItAsset | null
  employees: Employee[]
  onSaved: () => void
  onClose: () => void
}) {
  const [f, setF] = useState<ItAssetInput>(() => initial ? {
    ...EMPTY, ...initial, purchaseDate: day(initial.purchaseDate) || null, warrantyUntil: day(initial.warrantyUntil) || null,
  } : EMPTY)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const set = <K extends keyof ItAssetInput>(k: K, v: ItAssetInput[K]) => setF((p) => ({ ...p, [k]: v }))
  const txt = (k: keyof ItAssetInput) => ({
    value: (f[k] as string | null) ?? '',
    onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => set(k, (e.target.value || null) as never),
  })

  const save = async () => {
    if (!f.name.trim()) return setErr('اكتب اسم الجهاز')
    setBusy(true); setErr(null)
    try {
      if (initial) await api.updateItAsset(initial.id, f)
      else await api.createItAsset(f)
      onSaved()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally { setBusy(false) }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onClick={() => !busy && onClose()}>
      <div className="max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-2xl bg-white p-5 shadow-xl" onClick={(e) => e.stopPropagation()}>
        <h3 className="text-lg font-bold text-[#0f2040]">{initial ? '✏️ تعديل جهاز' : '➕ جهاز جديد'}</h3>
        <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
          <label className="text-sm font-bold text-slate-700 sm:col-span-2">اسم الجهاز *
            <input className={inp} placeholder="مثال: حاسبة المحاسبة ١" value={f.name} onChange={(e) => set('name', e.target.value)} />
          </label>
          <label className="text-sm font-bold text-slate-700">النوع
            <select className={inp} value={f.kind} onChange={(e) => set('kind', e.target.value as ItAssetKind)}>
              {Object.entries(IT_KIND_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
            </select>
          </label>
          <label className="text-sm font-bold text-slate-700">الحالة
            <select className={inp} value={f.status} onChange={(e) => set('status', e.target.value as ItAssetStatus)}>
              {Object.entries(IT_STATUS_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
            </select>
          </label>
          <label className="text-sm font-bold text-slate-700">الماركة<input className={inp} {...txt('brand')} /></label>
          <label className="text-sm font-bold text-slate-700">الموديل<input className={inp} {...txt('model')} /></label>
          <label className="text-sm font-bold text-slate-700">الرقم التسلسلي<input className={inp} dir="ltr" {...txt('serialNumber')} /></label>
          <label className="text-sm font-bold text-slate-700">عنوان IP<input className={inp} dir="ltr" placeholder="192.168.1.10" {...txt('ipAddress')} /></label>
          <label className="text-sm font-bold text-slate-700">المكان<input className={inp} placeholder="مثال: غرفة الحسابات" {...txt('location')} /></label>
          <label className="text-sm font-bold text-slate-700">يستعمله
            <select className={inp} value={f.assignedEmployeeId ?? ''} onChange={(e) => set('assignedEmployeeId', e.target.value || null)}>
              <option value="">— ماكو —</option>
              {employees.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
            </select>
          </label>
          <label className="text-sm font-bold text-slate-700">تاريخ الشراء<input type="date" className={inp} {...txt('purchaseDate')} /></label>
          <label className="text-sm font-bold text-slate-700">الضمان لغاية<input type="date" className={inp} {...txt('warrantyUntil')} /></label>
          <label className="text-sm font-bold text-slate-700 sm:col-span-2">ملاحظات<textarea rows={2} className={inp} {...txt('notes')} /></label>
        </div>
        {err && <p className="mt-3 rounded-lg bg-red-50 p-2 text-sm font-bold text-red-700">{err}</p>}
        <div className="mt-4 flex gap-2">
          <button onClick={save} disabled={busy} className="flex-1 rounded-lg bg-brand-600 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50">
            {busy ? 'جاري الحفظ…' : 'حفظ'}
          </button>
          <button onClick={onClose} disabled={busy} className="rounded-lg border border-slate-300 px-4 py-2.5 text-sm font-bold text-slate-600">إلغاء</button>
        </div>
      </div>
    </div>
  )
}

function AssetLogs({ asset, onClose }: { asset: ItAsset; onClose: () => void }) {
  const [logs, setLogs] = useState<ItAssetLog[] | null>(null)
  const [kind, setKind] = useState('REPAIR')
  const [note, setNote] = useState('')
  const [cost, setCost] = useState('')
  const [busy, setBusy] = useState(false)
  const [reload, setReload] = useState(0)

  useEffect(() => {
    let alive = true
    api.getItAssetLogs(asset.id).then((l) => { if (alive) setLogs(l) }).catch(() => { if (alive) setLogs([]) })
    return () => { alive = false }
  }, [asset.id, reload])

  const add = async () => {
    if (!note.trim()) return
    setBusy(true)
    try {
      await api.addItAssetLog(asset.id, { kind, note: note.trim(), cost: cost ? Number(cost) : null })
      setNote(''); setCost(''); setReload((n) => n + 1)
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally { setBusy(false) }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onClick={onClose}>
      <div className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-5 shadow-xl" onClick={(e) => e.stopPropagation()}>
        <h3 className="text-lg font-bold text-[#0f2040]">🧾 سجل «{asset.name}»</h3>
        <div className="mt-3 grid grid-cols-3 gap-2">
          <select className={inp} value={kind} onChange={(e) => setKind(e.target.value)}>
            {['REPAIR', 'MAINTENANCE', 'NOTE'].map((k) => <option key={k} value={k}>{IT_LOG_LABELS[k]}</option>)}
          </select>
          <input className={`${inp} col-span-2`} type="number" placeholder="الكلفة (اختياري)" value={cost} onChange={(e) => setCost(e.target.value)} />
          <textarea className={`${inp} col-span-3`} rows={2} placeholder="شنو انسوّى؟ مثال: تبديل هارد SSD" value={note} onChange={(e) => setNote(e.target.value)} />
        </div>
        <button onClick={add} disabled={busy || !note.trim()} className="mt-2 w-full rounded-lg bg-brand-600 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">
          {busy ? 'جاري الحفظ…' : '➕ أضف للسجل'}
        </button>
        <div className="mt-4 space-y-2">
          {logs === null ? <p className="text-sm text-slate-400">جاري التحميل…</p>
            : logs.length === 0 ? <p className="text-sm text-slate-400">ماكو سجل بعد.</p>
              : logs.map((l) => (
                <div key={l.id} className="rounded-xl bg-slate-50 px-3 py-2">
                  <div className="flex items-center justify-between gap-2 text-xs text-slate-500">
                    <span className="font-bold text-slate-700">{IT_LOG_LABELS[l.kind] || l.kind}</span>
                    <span>{new Date(l.createdAt).toLocaleDateString('ar-IQ')} · {l.employeeName || '—'}</span>
                  </div>
                  <p className="mt-1 text-sm text-slate-700">{l.note}</p>
                  {l.cost != null && <p className="text-xs text-slate-500">الكلفة: {l.cost.toLocaleString('en-US')} د.ع</p>}
                </div>
              ))}
        </div>
        <button onClick={onClose} className="mt-4 w-full rounded-lg border border-slate-300 px-4 py-2 text-sm font-bold text-slate-600">إغلاق</button>
      </div>
    </div>
  )
}

export default function ItAssetsPage() {
  const [assets, setAssets] = useState<ItAsset[] | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [employees, setEmployees] = useState<Employee[]>([])
  const [reload, setReload] = useState(0)
  const [search, setSearch] = useState('')
  const [kind, setKind] = useState<'' | ItAssetKind>('')
  const [status, setStatus] = useState<'' | ItAssetStatus>('')
  const [editing, setEditing] = useState<ItAsset | 'new' | null>(null)
  const [logsFor, setLogsFor] = useState<ItAsset | null>(null)

  useEffect(() => {
    let alive = true
    api.getItAssets()
      .then((a) => { if (alive) { setAssets(a); setError(null) } })
      .catch((e) => { if (alive) setError(e instanceof Error ? e.message : 'تعذر جلب الأجهزة') })
    return () => { alive = false }
  }, [reload])

  useEffect(() => {
    api.getEmployees().then((l) => setEmployees(l.filter((e) => e.status === 'ACTIVE'))).catch(() => setEmployees([]))
  }, [])

  const shown = useMemo(() => (assets || []).filter((a) =>
    (!kind || a.kind === kind) && (!status || a.status === status)
    && matches([a.name, a.brand, a.model, a.serialNumber, a.ipAddress, a.location, a.assignedEmployeeName], search),
  ), [assets, kind, status, search])

  const remove = async (a: ItAsset) => {
    if (!window.confirm(`حذف «${a.name}» وكل سجله؟ إذا الجهاز بس انستبعد، الأفضل تغيّر حالته لـ«مستبعد».`)) return
    try { await api.deleteItAsset(a.id); setReload((n) => n + 1) } catch (e) { alert(e instanceof Error ? e.message : 'تعذر الحذف') }
  }

  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">🖥️ جرد أجهزة تقنية المعلومات</h2>
          <p className="text-sm text-slate-500">كل جهاز بالشركة: وين هو، منو يستعمله، حالته، ضمانه، وسجل تصليحه.</p>
        </div>
        <button onClick={() => setEditing('new')} className="rounded-xl bg-brand-600 px-4 py-2.5 text-sm font-bold text-white hover:bg-brand-700">➕ جهاز جديد</button>
      </div>

      <div className="flex flex-wrap gap-2 rounded-2xl border border-slate-200 bg-white p-3">
        <input className="min-w-48 flex-1 rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" placeholder="🔍 بحث بالاسم، التسلسلي، IP، المكان، الموظف…" value={search} onChange={(e) => setSearch(e.target.value)} />
        <select className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" value={kind} onChange={(e) => setKind(e.target.value as '' | ItAssetKind)}>
          <option value="">كل الأنواع</option>
          {Object.entries(IT_KIND_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
        </select>
        <select className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" value={status} onChange={(e) => setStatus(e.target.value as '' | ItAssetStatus)}>
          <option value="">كل الحالات</option>
          {Object.entries(IT_STATUS_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
        </select>
      </div>

      {error && <p className="rounded-lg bg-red-50 p-4 text-red-600">{error}</p>}
      {assets === null && !error ? <p className="text-slate-400">جاري التحميل…</p> : shown.length === 0 ? (
        <p className="rounded-2xl border border-dashed border-slate-300 p-8 text-center text-slate-500">
          {assets?.length ? 'ماكو جهاز يطابق البحث.' : 'ماكو أجهزة بعد — ابدأ بإضافة أول جهاز.'}
        </p>
      ) : (
        <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white">
          <table className="w-full min-w-[820px] text-sm">
            <thead className="bg-slate-50 text-slate-500">
              <tr>
                {['الجهاز', 'النوع', 'الحالة', 'المكان / المستخدم', 'IP / التسلسلي', 'الضمان', ''].map((h) => <th key={h} className="px-3 py-2 text-right font-bold">{h}</th>)}
              </tr>
            </thead>
            <tbody>
              {shown.map((a) => {
                const warranty = a.warrantyUntil ? new Date(a.warrantyUntil) : null
                const daysLeft = warranty ? Math.round((warranty.getTime() - new Date(day(new Date().toISOString())).getTime()) / 86400000) : null
                return (
                  <tr key={a.id} className="border-t border-slate-100 align-top">
                    <td className="px-3 py-2">
                      <p className="font-bold text-slate-800">{a.name}</p>
                      <p className="text-xs text-slate-500">{[a.brand, a.model].filter(Boolean).join(' ') || '—'}</p>
                    </td>
                    <td className="px-3 py-2">{IT_KIND_LABELS[a.kind] || a.kind}</td>
                    <td className="px-3 py-2"><span className={`rounded-full px-2 py-0.5 text-xs font-bold ${STATUS_CLS[a.status]}`}>{IT_STATUS_LABELS[a.status]}</span></td>
                    <td className="px-3 py-2">
                      <p>{a.location || '—'}</p>
                      <p className="text-xs text-slate-500">{a.assignedEmployeeName || ''}</p>
                    </td>
                    <td className="px-3 py-2 font-mono text-xs" dir="ltr">
                      <p>{a.ipAddress || '—'}</p>
                      <p className="text-slate-400">{a.serialNumber || ''}</p>
                    </td>
                    <td className="px-3 py-2 text-xs">
                      {warranty ? (
                        <span className={daysLeft !== null && daysLeft < 0 ? 'text-slate-400' : daysLeft !== null && daysLeft <= 60 ? 'font-bold text-amber-700' : 'text-slate-600'}>
                          {day(a.warrantyUntil)}{daysLeft !== null && daysLeft < 0 ? ' (منتهي)' : daysLeft !== null && daysLeft <= 60 ? ` (${daysLeft} يوم)` : ''}
                        </span>
                      ) : '—'}
                    </td>
                    <td className="px-3 py-2">
                      <div className="flex justify-end gap-1.5 whitespace-nowrap">
                        <button onClick={() => setLogsFor(a)} className="rounded-lg bg-slate-100 px-2 py-1 text-xs font-bold text-slate-700 hover:bg-slate-200">🧾 السجل</button>
                        <button onClick={() => setEditing(a)} className="rounded-lg bg-brand-50 px-2 py-1 text-xs font-bold text-brand-700 hover:bg-brand-100">✏️ تعديل</button>
                        <button onClick={() => remove(a)} className="rounded-lg bg-red-50 px-2 py-1 text-xs font-bold text-red-700 hover:bg-red-100">🗑️</button>
                      </div>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}

      {editing && (
        <AssetForm
          initial={editing === 'new' ? null : editing}
          employees={employees}
          onClose={() => setEditing(null)}
          onSaved={() => { setEditing(null); setReload((n) => n + 1) }}
        />
      )}
      {logsFor && <AssetLogs asset={logsFor} onClose={() => { setLogsFor(null); setReload((n) => n + 1) }} />}
    </div>
  )
}
