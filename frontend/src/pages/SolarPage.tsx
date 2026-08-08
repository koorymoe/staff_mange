import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  api, SOLAR_BRANDS, SOLAR_CATEGORY_LABELS, SOLAR_SPEC_FIELDS,
  SOLAR_WIRING_TYPES, SOLAR_IRON_TYPES, SOLAR_IRON_UNITS,
  type SolarCategory, type SolarComponent, type SolarInstallation,
  type SolarIronLine, type SolarStats, type SolarSystem, type SolarWiringLine,
} from '../api'
import { useSession } from '../session'
import { useNavigate } from 'react-router-dom'
import { matches } from '../utils/search'

// ═══ نظام الطاقة الشمسية ═══
//
// كان نظام منفصل على Google Sheets، وصار جزء من النظام: نفس الموظفين
// ونفس الزبائن ونفس المهارات.
//
// أربع شاشات: لوحة التحكم · المنظومات · المخزن · الزبائن.

const iqd = (n: number) => `${Math.round(n || 0).toLocaleString('en-US')} د.ع`

type Tab = 'dashboard' | 'systems' | 'inventory' | 'customers'

export default function SolarPage() {
  const { employee, permissions } = useSession()
  const canEdit =
    employee?.role === 'ADMIN' || employee?.role === 'OWNER' || permissions.includes('solar_system')

  const [tab, setTab] = useState<Tab>('dashboard')
  const [stats, setStats] = useState<SolarStats | null>(null)
  const [components, setComponents] = useState<SolarComponent[]>([])
  const [systems, setSystems] = useState<SolarSystem[]>([])
  const [installations, setInstallations] = useState<SolarInstallation[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const reload = useCallback(() => {
    return Promise.all([
      api.getSolarStats().then(setStats).catch(() => {}),
      api.getSolarComponents().then(setComponents).catch(() => {}),
      api.getSolarSystems().then(setSystems).catch(() => {}),
      api.getSolarInstallations().then(setInstallations).catch(() => {}),
    ])
  }, [])

  useEffect(() => {
    reload()
      .catch((e) => setError(e instanceof Error ? e.message : 'تعذر التحميل'))
      .finally(() => setLoading(false))
  }, [reload])

  const tabs: { id: Tab; label: string; badge?: number }[] = [
    { id: 'dashboard', label: '📊 لوحة التحكم' },
    { id: 'systems', label: '☀️ المنظومات', badge: systems.length },
    { id: 'inventory', label: '📦 المخزن', badge: stats?.lowStockCount || undefined },
    { id: 'customers', label: '👥 الزبائن', badge: stats?.dueFollowUpCount || undefined },
  ]

  return (
    <div className="space-y-4">
      <div className="rounded-2xl bg-gradient-to-l from-amber-500 to-orange-500 p-5 text-white shadow-lg">
        <h1 className="text-xl font-extrabold">☀️ نظام الطاقة الشمسية</h1>
        <p className="mt-1 text-xs opacity-90">
          كتالوك المنظومات ومخزن المكوّنات وتجهيز المنظومات للزبائن ومتابعتهم
        </p>
      </div>

      <div className="flex flex-wrap gap-2">
        {tabs.map((t) => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`rounded-xl px-4 py-2 text-sm font-bold transition ${
              tab === t.id ? 'bg-[#0f2040] text-white shadow' : 'bg-white text-slate-600 hover:bg-slate-50'
            }`}
          >
            {t.label}
            {t.badge ? (
              <span className="mr-2 rounded-full bg-red-500 px-2 py-0.5 text-[10px] font-black text-white">
                {t.badge}
              </span>
            ) : null}
          </button>
        ))}
      </div>

      {loading && <div className="rounded-2xl bg-white p-8 text-center text-sm text-slate-500">جاري التحميل...</div>}
      {error && <div className="rounded-2xl bg-red-50 p-4 text-sm font-bold text-red-700">{error}</div>}

      {!loading && tab === 'dashboard' && <Dashboard stats={stats} components={components} />}
      {!loading && tab === 'systems' && (
        <Systems systems={systems} components={components} canEdit={canEdit} onChanged={reload} />
      )}
      {!loading && tab === 'inventory' && (
        <Inventory components={components} canEdit={canEdit} onChanged={reload} />
      )}
      {!loading && tab === 'customers' && (
        <Customers rows={installations} canEdit={canEdit} onChanged={reload} />
      )}
    </div>
  )
}

/* ═══════════════ لوحة التحكم ═══════════════ */

function Dashboard({ stats, components }: { stats: SolarStats | null; components: SolarComponent[] }) {
  const low = components.filter((c) => c.quantity <= c.minStock)
  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <Stat label="المنظومات" value={stats?.systemCount ?? 0} tone="amber" />
        <Stat label="منظومات مجهزة" value={stats?.processedCount ?? 0} tone="teal" />
        <Stat label="أصناف المخزن" value={stats?.componentCount ?? 0} tone="blue" />
        <Stat label="قيمة المخزن" value={iqd(stats?.inventoryValue ?? 0)} tone="emerald" />
        <Stat label="تنبيهات المخزن" value={stats?.lowStockCount ?? 0} tone="red" />
        <Stat label="نفد مخزونها" value={stats?.outOfStockCount ?? 0} tone="red" />
        <Stat label="زبائن يستحقون اتصال" value={stats?.dueFollowUpCount ?? 0} tone="red" />
        <Stat label="تركيبات هذا الشهر" value={stats?.installedThisMonth ?? 0} tone="blue" />
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <EnergyCalculator />

        <div className="rounded-2xl bg-white p-5 shadow-[0_2px_12px_rgba(0,0,0,0.06)]">
          <h3 className="mb-3 text-sm font-extrabold text-red-700">🔔 تنبيهات المخزن الحرجة</h3>
          {low.length === 0 ? (
            <div className="rounded-lg bg-emerald-50 p-3 text-xs font-bold text-emerald-700">
              ✅ ماكو تنبيهات — المخزن بخير
            </div>
          ) : (
            <div className="max-h-72 space-y-2 overflow-y-auto">
              {low.map((c) => (
                <div
                  key={c.id}
                  className={`flex items-center justify-between rounded-lg p-2.5 text-xs ${
                    c.quantity === 0 ? 'bg-red-50' : 'bg-amber-50'
                  }`}
                >
                  <span className={`font-black ${c.quantity === 0 ? 'text-red-700' : 'text-amber-700'}`}>
                    {c.quantity === 0 ? 'نفد!' : `${c.quantity} باقية`}
                  </span>
                  <div className="text-right">
                    <div className="font-bold text-slate-700">{c.name}</div>
                    <div className="text-[10px] text-slate-400">
                      {SOLAR_CATEGORY_LABELS[c.category]} · الحد الأدنى {c.minStock}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <Stat label="مجموع التسليكات" value={iqd(stats?.totalWiring ?? 0)} tone="violet" small />
        <Stat label="مجموع الحدادة" value={iqd(stats?.totalIron ?? 0)} tone="red" small />
        <Stat label="مجموع التنصيب" value={iqd(stats?.totalInstall ?? 0)} tone="emerald" small />
        <Stat label="مجموع البرمجة" value={iqd(stats?.totalProgram ?? 0)} tone="blue" small />
      </div>
    </div>
  )
}

const TONES: Record<string, string> = {
  amber: 'border-amber-200 bg-amber-50 text-amber-700',
  blue: 'border-blue-200 bg-blue-50 text-blue-700',
  emerald: 'border-emerald-200 bg-emerald-50 text-emerald-700',
  red: 'border-red-200 bg-red-50 text-red-700',
  teal: 'border-teal-200 bg-teal-50 text-teal-700',
  violet: 'border-violet-200 bg-violet-50 text-violet-700',
}

function Stat({ label, value, tone, small }: { label: string; value: number | string; tone: string; small?: boolean }) {
  return (
    <div className={`rounded-2xl border-2 p-4 text-center ${TONES[tone]}`}>
      <div className={`font-black ${small ? 'text-base' : 'text-2xl'}`}>{value}</div>
      <div className="mt-1 text-[11px] font-medium opacity-80">{label}</div>
    </div>
  )
}

/* ═══════════════ حاسبة الطاقة ═══════════════ */

// نفس معادلات النظام القديم بالضبط:
//   الاستهلاك اليومي = الشهري ÷ ٣٠
//   السعة = ceil(اليومي ÷ ٤٫٥)     ← ٤٫٥ ساعة شمس فعّالة بالعراق
//   الألواح = ceil(السعة×١٠٠٠ ÷ ٥٥٠)
//   البطارية = ceil(السعة×١٠٠٠ ÷ ٤٨ ÷ ٠٫٨)
function EnergyCalculator() {
  const [kwh, setKwh] = useState('')
  const result = useMemo(() => {
    const v = parseFloat(kwh)
    if (!v || v <= 0) return null
    const daily = v / 30
    const kw = Math.ceil(daily / 4.5)
    return {
      kw,
      panels: Math.ceil((kw * 1000) / 550),
      batteryAh: Math.ceil((kw * 1000) / 48 / 0.8),
    }
  }, [kwh])

  return (
    <div className="rounded-2xl bg-white p-5 shadow-[0_2px_12px_rgba(0,0,0,0.06)]">
      <h3 className="mb-1 text-sm font-extrabold text-blue-700">⚡ حاسبة الطاقة السريعة</h3>
      <p className="mb-3 text-[11px] text-slate-500">
        من استهلاك الزبون الشهري نحسب سعة المنظومة المناسبة — محسوبة على ٤٫٥ ساعة شمس فعّالة باليوم (العراق).
      </p>
      <label className="mb-1 block text-xs font-medium text-slate-600">الاستهلاك الشهري (كيلو واط ساعة)</label>
      <input
        type="number"
        value={kwh}
        onChange={(e) => setKwh(e.target.value)}
        placeholder="مثال: 500"
        className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-blue-500"
      />
      {result && (
        <div className="mt-3 rounded-xl border-2 border-blue-200 bg-blue-50 p-4">
          <div className="text-center">
            <div className="text-[11px] font-bold text-blue-700">المنظومة المقترحة</div>
            <div className="text-2xl font-black text-blue-900">{result.kw} كيلو واط</div>
          </div>
          <div className="mt-3 grid grid-cols-2 gap-2 text-center">
            <div className="rounded-lg bg-white p-2">
              <div className="text-[10px] text-slate-500">عدد الألواح</div>
              <div className="text-lg font-black text-amber-600">{result.panels}</div>
            </div>
            <div className="rounded-lg bg-white p-2">
              <div className="text-[10px] text-slate-500">حجم البطارية</div>
              <div className="text-lg font-black text-red-600">{result.batteryAh} Ah / 48V</div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

/* ═══════════════ المنظومات ═══════════════ */

function Systems({
  systems, components, canEdit, onChanged,
}: {
  systems: SolarSystem[]
  components: SolarComponent[]
  canEdit: boolean
  onChanged: () => Promise<unknown>
}) {
  const navigate = useNavigate()
  const [brand, setBrand] = useState('الكل')
  const [search, setSearch] = useState('')
  const [editing, setEditing] = useState<SolarSystem | 'new' | null>(null)
  const [processing, setProcessing] = useState<SolarSystem | null>(null)

  const shown = systems
    .filter((s) => brand === 'الكل' || s.brand === brand)
    .filter((s) => {
      return matches([s.brand, s.model, s.capacity, s.panel?.name, s.inverter?.name, s.battery?.name], search)
    })

  return (
    <div className="space-y-4">
      <div className="rounded-2xl bg-white p-4 shadow-sm">
        <div className="flex flex-wrap items-center gap-2">
          {['الكل', ...SOLAR_BRANDS].map((b) => (
            <button
              key={b}
              onClick={() => setBrand(b)}
              className={`rounded-full border-2 px-4 py-1.5 text-xs font-bold ${
                brand === b ? 'border-amber-500 bg-amber-500 text-white' : 'border-slate-200 bg-white text-slate-600'
              }`}
            >
              {b}
            </button>
          ))}
          {canEdit && (
            <button
              onClick={() => setEditing('new')}
              className="mr-auto rounded-xl bg-gradient-to-l from-amber-500 to-orange-500 px-4 py-2 text-sm font-bold text-white"
            >
              + منظومة جديدة
            </button>
          )}
        </div>
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="ابحث بالسعة، الموديل، أو المكوّنات..."
          className="mt-3 w-full rounded-lg border border-slate-300 bg-slate-50 px-4 py-2 text-sm outline-none focus:border-amber-500"
        />
      </div>

      {editing && (
        <SystemForm
          system={editing === 'new' ? null : editing}
          components={components}
          onClose={() => setEditing(null)}
          onSaved={async () => { await onChanged(); setEditing(null) }}
        />
      )}

      {processing && (
        <ProcessForm
          system={processing}
          onClose={() => setProcessing(null)}
          onDone={async () => { await onChanged(); setProcessing(null) }}
        />
      )}

      {shown.length === 0 && (
        <div className="rounded-2xl bg-white p-10 text-center">
          <div className="text-4xl">☀️</div>
          <p className="mt-3 text-sm font-bold text-slate-600">ماكو منظومات بهذي الماركة</p>
        </div>
      )}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {shown.map((s) => {
          const short = s.shortages.length > 0
          return (
            <div
              key={s.id}
              className={`overflow-hidden rounded-2xl border-2 bg-white shadow-sm ${
                short ? 'border-red-400' : 'border-transparent'
              }`}
            >
              <div className={`p-4 text-white ${short ? 'bg-gradient-to-l from-red-700 to-red-500' : 'bg-gradient-to-l from-[#0f2040] to-[#1e293b]'}`}>
                <div className="text-[11px] font-bold uppercase opacity-80">{s.brand}</div>
                <div className="text-2xl font-black">{s.capacity}</div>
                <div className="text-xs opacity-80">{s.model}</div>
                {short && (
                  <div className="mt-2 rounded-lg bg-white/15 p-2 text-[11px] font-bold">
                    ⚠️ المخزن ما يكفي:
                    {s.shortages.map((sh) => (
                      <div key={sh.componentId} className="font-normal">
                        • {sh.name} {sh.missing ? '(مو موجود بالمخزن)' : `— مطلوب ${sh.required} والمتوفر ${sh.available}`}
                      </div>
                    ))}
                  </div>
                )}
              </div>

              <div className="p-4">
                <Row icon="🔆" label="الألواح" value={s.panel ? `${s.panelQty} × ${s.panel.name}` : '—'} />
                <Row icon="⚡" label="الإنفيرتر" value={s.inverter ? `${s.inverterQty} × ${s.inverter.name}` : '—'} />
                <Row icon="🔋" label="البطاريات" value={s.battery ? `${s.batteryQty} × ${s.battery.name}` : '—'} />
                <Row icon="🔌" label="بورد الحماية" value={s.board?.name ?? '—'} />

                <div className="mt-3 space-y-1 border-t border-dashed border-slate-200 pt-3 text-xs">
                  <PriceRow label="الأجهزة" value={s.price.components + s.price.board} />
                  {s.price.wiring > 0 && <PriceRow label="التسليكات" value={s.price.wiring} />}
                  {s.price.iron > 0 && <PriceRow label="الحدادة والتشكيل" value={s.price.iron} />}
                  {s.price.install > 0 && <PriceRow label="التنصيب" value={s.price.install} />}
                  {s.price.program > 0 && <PriceRow label="البرمجة" value={s.price.program} />}
                  {s.price.warranty > 0 && <PriceRow label="الضمان الممتد" value={s.price.warranty} />}
                </div>

                <div className="mt-3 rounded-xl bg-gradient-to-l from-amber-100 to-amber-50 p-3 text-center">
                  <div className="text-[10px] font-bold text-amber-700">الإجمالي (بأسعار المخزن اليوم)</div>
                  <div className="text-xl font-black text-orange-700">{iqd(s.price.total)}</div>
                </div>

                <button
                  onClick={() =>
                    navigate(`/quotations/new?solarSystemId=${s.id}&projectName=${encodeURIComponent(`منظومة طاقة شمسية ${s.capacity} — ${s.brand} ${s.model}`)}`)
                  }
                  className="mt-3 w-full rounded-lg border-2 border-[#1a237e] bg-white px-3 py-2 text-xs font-bold text-[#1a237e]"
                >
                  📄 اطلع عرض سعر رسمي للزبون
                </button>

                {canEdit && (
                  <div className="mt-3 flex flex-wrap gap-2">
                    <button
                      onClick={() => setProcessing(s)}
                      disabled={short}
                      title={short ? 'المخزن ما يكفي' : ''}
                      className="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-bold text-white disabled:cursor-not-allowed disabled:opacity-40"
                    >
                      ✅ جهّزها لزبون
                    </button>
                    <button onClick={() => setEditing(s)} className="rounded-lg border border-slate-300 px-3 py-1.5 text-xs font-bold text-slate-600">
                      ✏️ تعديل
                    </button>
                    <button
                      onClick={async () => {
                        if (!confirm(`تريد تحذف منظومة ${s.brand} ${s.capacity}؟`)) return
                        try { await api.deleteSolarSystem(s.id); await onChanged() }
                        catch (e) { alert(e instanceof Error ? e.message : 'تعذر الحذف') }
                      }}
                      className="rounded-lg border border-red-300 px-3 py-1.5 text-xs font-bold text-red-600"
                    >
                      🗑️ حذف
                    </button>
                  </div>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

function Row({ icon, label, value }: { icon: string; label: string; value: string }) {
  return (
    <div className="mb-1.5 flex items-center justify-between rounded-lg bg-slate-50 px-3 py-2 text-xs">
      <span className="font-bold text-slate-800">{value}</span>
      <span className="font-medium text-slate-500">{icon} {label}</span>
    </div>
  )
}

function PriceRow({ label, value }: { label: string; value: number }) {
  return (
    <div className="flex justify-between">
      <span className="font-bold text-slate-700">{iqd(value)}</span>
      <span className="text-slate-500">{label}</span>
    </div>
  )
}

/* ═══════════════ نموذج المنظومة ═══════════════ */

function SystemForm({
  system, components, onClose, onSaved,
}: {
  system: SolarSystem | null
  components: SolarComponent[]
  onClose: () => void
  onSaved: () => Promise<void>
}) {
  const byCat = (c: SolarCategory) => components.filter((x) => x.category === c)

  const [brand, setBrand] = useState(system?.brand ?? '')
  const [model, setModel] = useState(system?.model ?? '')
  const [capacity, setCapacity] = useState(system?.capacity ?? '')
  const [panelId, setPanelId] = useState(system?.panelId ?? '')
  const [panelQty, setPanelQty] = useState(String(system?.panelQty ?? 1))
  const [inverterId, setInverterId] = useState(system?.inverterId ?? '')
  const [inverterQty, setInverterQty] = useState(String(system?.inverterQty ?? 1))
  const [batteryId, setBatteryId] = useState(system?.batteryId ?? '')
  const [batteryQty, setBatteryQty] = useState(String(system?.batteryQty ?? 1))
  const [boardId, setBoardId] = useState(system?.boardId ?? '')
  const [wiring, setWiring] = useState<SolarWiringLine[]>(system?.wiringDetails ?? [])
  const [iron, setIron] = useState<SolarIronLine[]>(system?.ironDetails ?? [])
  const [installPrice, setInstallPrice] = useState(String(system?.installPrice ?? 0))
  const [programPrice, setProgramPrice] = useState(String(system?.programPrice ?? 0))
  const [warrantyPrice, setWarrantyPrice] = useState(String(system?.warrantyPrice ?? 0))
  const [notes, setNotes] = useState(system?.notes ?? '')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  const wiringTotal = wiring.reduce((s, l) => s + (l.length || 0) * (l.price || 0), 0)
  const ironTotal = iron.reduce((s, l) => s + (l.qty || 0) * (l.price || 0), 0)

  // ═══ اقتراح الأسعار حسب السعة ═══
  // نفس معادلات النظام القديم. يقرا الرقم من نص السعة ويعرف إذا أمبير
  // لو كيلوواط. اقتراح مو إلزام — الأسعار تنعدّل بالإيد بعده.
  const suggestion = useMemo(() => {
    const num = parseFloat(capacity.replace(/[^0-9.]/g, ''))
    if (!num || num <= 0) return null
    const low = capacity.toLowerCase()
    const unit = low.includes('a') ? 'amp' : low.includes('kw') || capacity.includes('كيلو') ? 'kw' : 'unknown'
    if (unit === 'amp') return { unit: 'أمبير', num, iron: num * 70000, program: num * 60000, install: num * 50000 }
    if (unit === 'kw') return { unit: 'كيلو واط', num, iron: num * 70000, program: num * 45000, install: num * 60000 }
    return { unit: '', num, iron: num * 70000, program: num * 50000, install: num * 50000 }
  }, [capacity])

  const componentsTotal = (() => {
    const find = (id: string) => components.find((c) => c.id === id)
    const p = find(panelId), i = find(inverterId), b = find(batteryId), bo = find(boardId)
    return (p ? p.price * (+panelQty || 0) : 0) + (i ? i.price * (+inverterQty || 0) : 0) +
      (b ? b.price * (+batteryQty || 0) : 0) + (bo ? bo.price : 0)
  })()
  const grandTotal = componentsTotal + wiringTotal + ironTotal + (+installPrice || 0) + (+programPrice || 0) + (+warrantyPrice || 0)

  const save = async () => {
    if (!brand || !model.trim() || !capacity.trim()) { setErr('الماركة والموديل والسعة مطلوبات'); return }
    setBusy(true); setErr(null)
    const payload = {
      brand, model: model.trim(), capacity: capacity.trim(),
      panelId: panelId || null, panelQty: +panelQty || 0,
      inverterId: inverterId || null, inverterQty: +inverterQty || 0,
      batteryId: batteryId || null, batteryQty: +batteryQty || 0,
      boardId: boardId || null,
      wiringDetails: wiring.filter((l) => l.length > 0 && l.price > 0),
      ironDetails: iron.filter((l) => l.qty > 0 && l.price > 0),
      installPrice: +installPrice || 0, programPrice: +programPrice || 0, warrantyPrice: +warrantyPrice || 0,
      notes: notes.trim() || null,
    }
    try {
      if (system) await api.updateSolarSystem(system.id, payload)
      else await api.createSolarSystem(payload)
      await onSaved()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally { setBusy(false) }
  }

  const Select = ({ value, onChange, list, label }: { value: string; onChange: (v: string) => void; list: SolarComponent[]; label: string }) => (
    <div>
      <label className="mb-1 block text-xs font-medium text-slate-600">{label}</label>
      <select value={value} onChange={(e) => onChange(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
        <option value="">— اختر —</option>
        {list.map((c) => (
          <option key={c.id} value={c.id}>{c.name} ({iqd(c.price)} · متوفر {c.quantity})</option>
        ))}
      </select>
    </div>
  )

  return (
    <div className="rounded-2xl border-2 border-amber-300 bg-white p-5 shadow-lg">
      <h3 className="mb-4 text-base font-extrabold text-[#0f2040]">
        {system ? '✏️ تعديل المنظومة' : '➕ منظومة جديدة'}
      </h3>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">الماركة *</label>
          <select value={brand} onChange={(e) => setBrand(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
            <option value="">— اختر —</option>
            {SOLAR_BRANDS.map((b) => <option key={b}>{b}</option>)}
          </select>
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">الموديل *</label>
          <input value={model} onChange={(e) => setModel(e.target.value)} placeholder="SUN-5K-SG03LP1" className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">السعة *</label>
          <input value={capacity} onChange={(e) => setCapacity(e.target.value)} placeholder="5 كيلو واط أو 40A" className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
        </div>
      </div>

      {suggestion && (
        <div className="mt-3 rounded-xl border-2 border-sky-200 bg-sky-50 p-3">
          <div className="mb-2 text-xs font-bold text-sky-800">
            💡 اقتراح أسعار حسب السعة ({suggestion.num} {suggestion.unit}) — تقديري وتكدر تعدّله
          </div>
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
            <SuggestBtn label="الحدادة" value={suggestion.iron} onClick={() => setIron([{ type: SOLAR_IRON_TYPES[0], qty: 1, unit: 'مجموعة', price: suggestion.iron }])} />
            <SuggestBtn label="البرمجة" value={suggestion.program} onClick={() => setProgramPrice(String(Math.round(suggestion.program)))} />
            <SuggestBtn label="التنصيب" value={suggestion.install} onClick={() => setInstallPrice(String(Math.round(suggestion.install)))} />
          </div>
        </div>
      )}

      <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
        <div className="flex gap-2">
          <div className="flex-1"><Select value={panelId} onChange={setPanelId} list={byCat('PANEL')} label="الألواح الشمسية" /></div>
          <div className="w-20"><label className="mb-1 block text-xs text-slate-600">العدد</label>
            <input type="number" min={0} value={panelQty} onChange={(e) => setPanelQty(e.target.value)} className="w-full rounded-lg border border-slate-300 px-2 py-2 text-sm" /></div>
        </div>
        <div className="flex gap-2">
          <div className="flex-1"><Select value={inverterId} onChange={setInverterId} list={byCat('INVERTER')} label="الإنفيرتر" /></div>
          <div className="w-20"><label className="mb-1 block text-xs text-slate-600">العدد</label>
            <input type="number" min={0} value={inverterQty} onChange={(e) => setInverterQty(e.target.value)} className="w-full rounded-lg border border-slate-300 px-2 py-2 text-sm" /></div>
        </div>
        <div className="flex gap-2">
          <div className="flex-1"><Select value={batteryId} onChange={setBatteryId} list={byCat('BATTERY')} label="البطاريات" /></div>
          <div className="w-20"><label className="mb-1 block text-xs text-slate-600">العدد</label>
            <input type="number" min={0} value={batteryQty} onChange={(e) => setBatteryQty(e.target.value)} className="w-full rounded-lg border border-slate-300 px-2 py-2 text-sm" /></div>
        </div>
        <Select value={boardId} onChange={setBoardId} list={byCat('BOARD')} label="بورد الحماية" />
      </div>

      {/* التسليكات */}
      <LineEditor
        title="🔌 التسليكات والأطوال"
        tone="violet"
        lines={wiring}
        total={wiringTotal}
        onAdd={() => setWiring([...wiring, { type: SOLAR_WIRING_TYPES[0], length: 0, price: 0 }])}
        onRemove={(i) => setWiring(wiring.filter((_, x) => x !== i))}
        render={(l, i) => (
          <>
            <select value={l.type} onChange={(e) => setWiring(wiring.map((x, xi) => xi === i ? { ...x, type: e.target.value } : x))} className="rounded-lg border border-slate-300 px-2 py-1.5 text-xs">
              {SOLAR_WIRING_TYPES.map((t) => <option key={t}>{t}</option>)}
            </select>
            <input type="number" placeholder="الطول (م)" value={l.length || ''} onChange={(e) => setWiring(wiring.map((x, xi) => xi === i ? { ...x, length: +e.target.value } : x))} className="w-24 rounded-lg border border-slate-300 px-2 py-1.5 text-xs" />
            <input type="number" placeholder="سعر المتر" value={l.price || ''} onChange={(e) => setWiring(wiring.map((x, xi) => xi === i ? { ...x, price: +e.target.value } : x))} className="w-28 rounded-lg border border-slate-300 px-2 py-1.5 text-xs" />
            <span className="text-xs font-bold text-violet-700">{iqd((l.length || 0) * (l.price || 0))}</span>
          </>
        )}
      />

      {/* الحدادة */}
      <LineEditor
        title="🔨 أعمال الحدادة والتشكيل"
        tone="red"
        lines={iron}
        total={ironTotal}
        onAdd={() => setIron([...iron, { type: SOLAR_IRON_TYPES[0], qty: 0, unit: 'متر', price: 0 }])}
        onRemove={(i) => setIron(iron.filter((_, x) => x !== i))}
        render={(l, i) => (
          <>
            <select value={l.type} onChange={(e) => setIron(iron.map((x, xi) => xi === i ? { ...x, type: e.target.value } : x))} className="rounded-lg border border-slate-300 px-2 py-1.5 text-xs">
              {SOLAR_IRON_TYPES.map((t) => <option key={t}>{t}</option>)}
            </select>
            <input type="number" placeholder="العدد" value={l.qty || ''} onChange={(e) => setIron(iron.map((x, xi) => xi === i ? { ...x, qty: +e.target.value } : x))} className="w-20 rounded-lg border border-slate-300 px-2 py-1.5 text-xs" />
            <select value={l.unit} onChange={(e) => setIron(iron.map((x, xi) => xi === i ? { ...x, unit: e.target.value } : x))} className="rounded-lg border border-slate-300 px-2 py-1.5 text-xs">
              {SOLAR_IRON_UNITS.map((u) => <option key={u}>{u}</option>)}
            </select>
            <input type="number" placeholder="سعر الوحدة" value={l.price || ''} onChange={(e) => setIron(iron.map((x, xi) => xi === i ? { ...x, price: +e.target.value } : x))} className="w-28 rounded-lg border border-slate-300 px-2 py-1.5 text-xs" />
            <span className="text-xs font-bold text-red-700">{iqd((l.qty || 0) * (l.price || 0))}</span>
          </>
        )}
      />

      <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
        <NumField label="🔧 سعر التنصيب" value={installPrice} onChange={setInstallPrice} />
        <NumField label="💻 سعر البرمجة" value={programPrice} onChange={setProgramPrice} />
        <NumField label="🛡️ الضمان الممتد (اختياري)" value={warrantyPrice} onChange={setWarrantyPrice} />
      </div>

      <div className="mt-3">
        <label className="mb-1 block text-xs font-medium text-slate-600">ملاحظات</label>
        <textarea value={notes} onChange={(e) => setNotes(e.target.value)} rows={2} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
      </div>

      <div className="mt-4 rounded-xl border-2 border-amber-400 bg-amber-50 p-3 text-center">
        <div className="text-xs font-bold text-amber-800">الإجمالي التقديري للمنظومة</div>
        <div className="text-2xl font-black text-orange-700">{iqd(grandTotal)}</div>
      </div>

      <div className="mt-4 flex items-center gap-2">
        <button onClick={save} disabled={busy} className="rounded-lg bg-gradient-to-l from-amber-500 to-orange-500 px-5 py-2 text-sm font-bold text-white disabled:opacity-50">
          {busy ? 'جاري الحفظ...' : 'حفظ المنظومة'}
        </button>
        <button onClick={onClose} className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-bold text-slate-600">إلغاء</button>
        {err && <span className="whitespace-pre-line text-xs font-bold text-red-600">{err}</span>}
      </div>
    </div>
  )
}

function SuggestBtn({ label, value, onClick }: { label: string; value: number; onClick: () => void }) {
  return (
    <button onClick={onClick} className="rounded-lg border-2 border-sky-300 bg-white p-2 text-center transition hover:bg-sky-100">
      <div className="text-[10px] font-bold text-sky-700">{label}</div>
      <div className="text-sm font-black text-sky-900">{iqd(value)}</div>
      <div className="text-[9px] text-slate-400">اضغط للتطبيق</div>
    </button>
  )
}

function NumField({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  return (
    <div>
      <label className="mb-1 block text-xs font-medium text-slate-600">{label}</label>
      <input type="number" min={0} value={value} onChange={(e) => onChange(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
    </div>
  )
}

function LineEditor<T>({
  title, tone, lines, total, onAdd, onRemove, render,
}: {
  title: string
  tone: 'violet' | 'red'
  lines: T[]
  total: number
  onAdd: () => void
  onRemove: (i: number) => void
  render: (line: T, i: number) => React.ReactNode
}) {
  const bg = tone === 'violet' ? 'border-violet-200 bg-violet-50/50' : 'border-red-200 bg-red-50/50'
  return (
    <div className={`mt-4 rounded-xl border-2 p-3 ${bg}`}>
      <div className="mb-2 flex items-center justify-between">
        <span className="text-xs font-bold text-slate-500">المجموع: {iqd(total)}</span>
        <h4 className="text-sm font-bold text-slate-800">{title}</h4>
      </div>
      {lines.map((l, i) => (
        <div key={i} className="mb-2 flex flex-wrap items-center gap-2">
          {render(l, i)}
          <button onClick={() => onRemove(i)} className="rounded-lg border border-slate-300 bg-white px-2 py-1 text-xs text-red-600">✕</button>
        </div>
      ))}
      <button onClick={onAdd} className="rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-xs font-bold text-slate-600">+ إضافة سطر</button>
    </div>
  )
}

/* ═══════════════ تجهيز منظومة لزبون ═══════════════ */

function ProcessForm({ system, onClose, onDone }: { system: SolarSystem; onClose: () => void; onDone: () => Promise<void> }) {
  const [name, setName] = useState('')
  const [phone, setPhone] = useState('')
  const [address, setAddress] = useState('')
  const [date, setDate] = useState(new Date().toISOString().slice(0, 10))
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  const submit = async () => {
    if (!name.trim() || !phone.trim()) { setErr('اسم الزبون ورقم هاتفه مطلوبين'); return }
    setBusy(true); setErr(null)
    try {
      await api.processSolarSystem(system.id, {
        customerName: name.trim(), customerPhone: phone.trim(),
        customerAddress: address.trim(), installDate: date,
      })
      await onDone()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر التجهيز')
    } finally { setBusy(false) }
  }

  return (
    <div className="rounded-2xl border-2 border-emerald-300 bg-white p-5 shadow-lg">
      <h3 className="mb-1 text-base font-extrabold text-emerald-800">✅ تجهيز منظومة لزبون</h3>
      <p className="mb-4 text-xs text-slate-500">
        {system.brand} · {system.capacity} · {system.model} — راح <b>تنخصم مكوّناتها من المخزن</b>،
        والزبون يدخل دورة متابعة بعد ٣٠ يوم من التركيب.
      </p>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">اسم الزبون الكامل *</label>
          <input value={name} onChange={(e) => setName(e.target.value)} placeholder="الاسم الثلاثي" className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">رقم الهاتف *</label>
          <input value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="07XX XXX XXXX" dir="ltr" className="w-full rounded-lg border border-slate-300 px-3 py-2 text-right text-sm" />
        </div>
        <div className="sm:col-span-2">
          <label className="mb-1 block text-xs font-medium text-slate-600">العنوان</label>
          <input value={address} onChange={(e) => setAddress(e.target.value)} placeholder="المحافظة، المنطقة، أقرب نقطة دالة" className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">تاريخ التركيب *</label>
          <input type="date" value={date} onChange={(e) => setDate(e.target.value)} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
        </div>
      </div>

      <div className="mt-4 flex items-center gap-2">
        <button onClick={submit} disabled={busy} className="rounded-lg bg-emerald-600 px-5 py-2 text-sm font-bold text-white disabled:opacity-50">
          {busy ? 'جاري التجهيز...' : 'أكّد واخصم من المخزن'}
        </button>
        <button onClick={onClose} className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-bold text-slate-600">إلغاء</button>
      </div>
      {err && <div className="mt-3 whitespace-pre-line rounded-lg bg-red-50 p-3 text-xs font-bold text-red-700">{err}</div>}
    </div>
  )
}

/* ═══════════════ المخزن ═══════════════ */

function Inventory({ components, canEdit, onChanged }: { components: SolarComponent[]; canEdit: boolean; onChanged: () => Promise<unknown> }) {
  const [filter, setFilter] = useState<SolarCategory | 'all' | 'low'>('all')
  const [editing, setEditing] = useState<SolarComponent | 'new' | null>(null)

  const shown = components.filter((c) => {
    if (filter === 'all') return true
    if (filter === 'low') return c.quantity <= c.minStock
    return c.category === filter
  })

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-2">
        <FilterBtn active={filter === 'all'} onClick={() => setFilter('all')}>الكل</FilterBtn>
        {(Object.keys(SOLAR_CATEGORY_LABELS) as SolarCategory[]).map((c) => (
          <FilterBtn key={c} active={filter === c} onClick={() => setFilter(c)}>{SOLAR_CATEGORY_LABELS[c]}</FilterBtn>
        ))}
        <FilterBtn active={filter === 'low'} onClick={() => setFilter('low')} danger>⚠️ مخزون منخفض</FilterBtn>
        {canEdit && (
          <button onClick={() => setEditing('new')} className="mr-auto rounded-xl bg-emerald-600 px-4 py-2 text-sm font-bold text-white">
            + مادة جديدة
          </button>
        )}
      </div>

      {editing && (
        <ComponentForm
          component={editing === 'new' ? null : editing}
          onClose={() => setEditing(null)}
          onSaved={async () => { await onChanged(); setEditing(null) }}
        />
      )}

      <div className="overflow-x-auto rounded-2xl bg-white shadow-sm">
        <table className="w-full text-right text-sm">
          <thead>
            <tr className="bg-[#0f2040] text-xs text-white">
              <th className="p-3">المادة</th>
              <th className="p-3">التصنيف</th>
              <th className="p-3">الكمية</th>
              <th className="p-3">الحد الأدنى</th>
              <th className="p-3">السعر</th>
              <th className="p-3">القيمة</th>
              {canEdit && <th className="p-3"></th>}
            </tr>
          </thead>
          <tbody>
            {shown.length === 0 && (
              <tr><td colSpan={7} className="p-8 text-center text-sm text-slate-400">ماكو مواد بهذا التصنيف</td></tr>
            )}
            {shown.map((c) => {
              const out = c.quantity === 0
              const low = !out && c.quantity <= c.minStock
              return (
                <tr key={c.id} className={`border-b border-slate-100 ${out ? 'bg-red-50' : low ? 'bg-amber-50' : ''}`}>
                  <td className="p-3 font-bold text-slate-800">
                    {c.name}
                    {Object.keys(c.specs || {}).length > 0 && (
                      <div className="mt-0.5 text-[10px] font-normal text-slate-400">
                        {Object.entries(c.specs).filter(([, v]) => v).slice(0, 4).map(([k, v]) => `${k}: ${v}`).join(' · ')}
                      </div>
                    )}
                  </td>
                  <td className="p-3 text-xs text-slate-600">{SOLAR_CATEGORY_LABELS[c.category]}</td>
                  <td className="p-3">
                    <span className={`rounded-full px-2.5 py-1 text-xs font-black ${out ? 'bg-red-200 text-red-800' : low ? 'bg-amber-200 text-amber-800' : 'bg-emerald-100 text-emerald-700'}`}>
                      {out ? 'نفد' : c.quantity}
                    </span>
                  </td>
                  <td className="p-3 text-xs text-slate-500">{c.minStock}</td>
                  <td className="p-3 text-xs">{iqd(c.price)}</td>
                  <td className="p-3 text-xs font-bold text-brand-700">{iqd(c.quantity * c.price)}</td>
                  {canEdit && (
                    <td className="p-3">
                      <div className="flex gap-1">
                        <button onClick={() => setEditing(c)} className="rounded-lg border border-slate-300 px-2 py-1 text-xs">✏️</button>
                        <button
                          onClick={async () => {
                            if (!confirm(`تريد تحذف ${c.name}؟`)) return
                            try { await api.deleteSolarComponent(c.id); await onChanged() }
                            catch (e) { alert(e instanceof Error ? e.message : 'تعذر الحذف') }
                          }}
                          className="rounded-lg border border-red-300 px-2 py-1 text-xs text-red-600"
                        >🗑️</button>
                      </div>
                    </td>
                  )}
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function FilterBtn({ children, active, onClick, danger }: { children: React.ReactNode; active: boolean; onClick: () => void; danger?: boolean }) {
  return (
    <button
      onClick={onClick}
      className={`rounded-full border-2 px-4 py-1.5 text-xs font-bold ${
        active
          ? danger ? 'border-red-500 bg-red-500 text-white' : 'border-amber-500 bg-amber-500 text-white'
          : 'border-slate-200 bg-white text-slate-600'
      }`}
    >{children}</button>
  )
}

function ComponentForm({ component, onClose, onSaved }: { component: SolarComponent | null; onClose: () => void; onSaved: () => Promise<void> }) {
  const [name, setName] = useState(component?.name ?? '')
  const [category, setCategory] = useState<SolarCategory>(component?.category ?? 'PANEL')
  const [quantity, setQuantity] = useState(String(component?.quantity ?? 0))
  const [price, setPrice] = useState(String(component?.price ?? 0))
  const [minStock, setMinStock] = useState(String(component?.minStock ?? 5))
  const [specs, setSpecs] = useState<Record<string, string>>(component?.specs ?? {})
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  const save = async () => {
    if (!name.trim()) { setErr('اسم المادة مطلوب'); return }
    setBusy(true); setErr(null)
    const payload = {
      name: name.trim(), category,
      quantity: +quantity || 0, price: +price || 0, minStock: +minStock || 0,
      specs, notes: null,
    }
    try {
      if (component) await api.updateSolarComponent(component.id, payload)
      else await api.createSolarComponent(payload)
      await onSaved()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally { setBusy(false) }
  }

  return (
    <div className="rounded-2xl border-2 border-emerald-300 bg-white p-5 shadow-lg">
      <h3 className="mb-4 text-base font-extrabold text-[#0f2040]">
        {component ? '✏️ تعديل المادة' : '➕ مادة جديدة بالمخزن'}
      </h3>
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-5">
        <div className="sm:col-span-2">
          <label className="mb-1 block text-xs font-medium text-slate-600">اسم المادة *</label>
          <input value={name} onChange={(e) => setName(e.target.value)} placeholder="لوح جينكو 550 واط" className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" />
        </div>
        <div>
          <label className="mb-1 block text-xs font-medium text-slate-600">التصنيف</label>
          <select value={category} onChange={(e) => { setCategory(e.target.value as SolarCategory); setSpecs({}) }} className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
            {(Object.keys(SOLAR_CATEGORY_LABELS) as SolarCategory[]).map((c) => (
              <option key={c} value={c}>{SOLAR_CATEGORY_LABELS[c]}</option>
            ))}
          </select>
        </div>
        <NumField label="الكمية" value={quantity} onChange={setQuantity} />
        <NumField label="السعر (د.ع)" value={price} onChange={setPrice} />
        <NumField label="الحد الأدنى للتنبيه" value={minStock} onChange={setMinStock} />
      </div>

      <div className="mt-4 rounded-xl border-2 border-slate-200 bg-slate-50 p-3">
        <h4 className="mb-2 text-sm font-bold text-slate-700">
          المواصفات الفنية — {SOLAR_CATEGORY_LABELS[category]}
        </h4>
        <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
          {SOLAR_SPEC_FIELDS[category].map((f) => (
            <div key={f}>
              <label className="mb-1 block text-[10px] font-bold text-slate-500">{f}</label>
              <input
                value={specs[f] ?? ''}
                onChange={(e) => setSpecs({ ...specs, [f]: e.target.value })}
                className="w-full rounded-lg border border-slate-300 bg-white px-2 py-1.5 text-xs"
              />
            </div>
          ))}
        </div>
      </div>

      <div className="mt-4 flex items-center gap-2">
        <button onClick={save} disabled={busy} className="rounded-lg bg-emerald-600 px-5 py-2 text-sm font-bold text-white disabled:opacity-50">
          {busy ? 'جاري الحفظ...' : 'حفظ المادة'}
        </button>
        <button onClick={onClose} className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-bold text-slate-600">إلغاء</button>
        {err && <span className="text-xs font-bold text-red-600">{err}</span>}
      </div>
    </div>
  )
}

/* ═══════════════ الزبائن ═══════════════ */

function Customers({ rows, canEdit, onChanged }: { rows: SolarInstallation[]; canEdit: boolean; onChanged: () => Promise<unknown> }) {
  const [filter, setFilter] = useState<'all' | 'due' | 'contacted'>('all')
  const [search, setSearch] = useState('')

  const shown = rows
    .filter((r) => filter === 'all' || (filter === 'due' ? r.dueForFollowUp : r.status === 'CONTACTED'))
    .filter((r) => {
      return matches([r.customer?.name, r.customer?.phone, r.system?.brand, r.system?.capacity], search)
    })

  const fmtDate = (s: string) => new Date(s).toLocaleDateString('ar-IQ', { dateStyle: 'medium' })

  return (
    <div className="space-y-4">
      <div className="rounded-2xl bg-white p-4 shadow-sm">
        <div className="flex flex-wrap gap-2">
          <FilterBtn active={filter === 'all'} onClick={() => setFilter('all')}>الكل ({rows.length})</FilterBtn>
          <FilterBtn active={filter === 'due'} onClick={() => setFilter('due')} danger>
            🔔 يستحقون الاتصال ({rows.filter((r) => r.dueForFollowUp).length})
          </FilterBtn>
          <FilterBtn active={filter === 'contacted'} onClick={() => setFilter('contacted')}>
            تم الاتصال ({rows.filter((r) => r.status === 'CONTACTED').length})
          </FilterBtn>
        </div>
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="ابحث بالاسم أو الهاتف أو المنظومة..."
          className="mt-3 w-full rounded-lg border border-slate-300 bg-slate-50 px-4 py-2 text-sm outline-none focus:border-amber-500"
        />
      </div>

      {shown.length === 0 && (
        <div className="rounded-2xl bg-white p-10 text-center">
          <div className="text-4xl">👥</div>
          <p className="mt-3 text-sm font-bold text-slate-600">ماكو زبائن هنا</p>
          <p className="mt-1 text-xs text-slate-400">الزبائن ينضافون تلقائياً لمن تجهّز منظومة</p>
        </div>
      )}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {shown.map((r) => (
          <div key={r.id} className={`rounded-2xl border-2 bg-white p-4 shadow-sm ${r.dueForFollowUp ? 'border-red-400' : 'border-transparent'}`}>
            <div className="flex items-start justify-between gap-3">
              <div className="text-left">
                {r.dueForFollowUp ? (
                  <span className="rounded-lg bg-red-100 px-2.5 py-1 text-[11px] font-black text-red-700">
                    🔔 يستحق الاتصال ({r.daysOverdue} يوم متأخر)
                  </span>
                ) : r.status === 'CONTACTED' ? (
                  <span className="rounded-lg bg-emerald-100 px-2.5 py-1 text-[11px] font-black text-emerald-700">✅ تم الاتصال</span>
                ) : (
                  <span className="rounded-lg bg-slate-100 px-2.5 py-1 text-[11px] font-bold text-slate-600">⏳ بانتظار موعد المتابعة</span>
                )}
              </div>
              <div>
                <div className="text-base font-black text-[#0f2040]">{r.customer?.name || '—'}</div>
                <div className="text-xs text-slate-500" dir="ltr">{r.customer?.phone || ''}</div>
                {r.customer?.location && <div className="text-[11px] text-slate-400">{r.customer.location}</div>}
              </div>
            </div>

            <div className="mt-3 rounded-xl border border-amber-200 bg-amber-50 p-3">
              <div className="flex items-center justify-between">
                <span className="rounded-md bg-slate-800 px-2 py-0.5 text-[10px] font-bold text-white">{r.system?.brand}</span>
                <span className="text-sm font-black text-amber-700">☀️ {r.system?.capacity}</span>
              </div>
              <div className="mt-1 text-[11px] text-slate-500">{r.system?.model}</div>
              <div className="mt-2 text-sm font-bold text-brand-700">السعر: {iqd(r.totalPrice)}</div>
            </div>

            <div className="mt-3 flex items-center justify-between border-t border-dashed border-slate-200 pt-3 text-xs">
              <div className="text-left">
                <div className="text-[10px] text-slate-400">موعد المتابعة</div>
                <div className="font-bold text-slate-700">{fmtDate(r.followUpAt)}</div>
              </div>
              <div>
                <div className="text-[10px] text-slate-400">تاريخ التركيب</div>
                <div className="font-bold text-slate-700">{fmtDate(r.installDate)}</div>
              </div>
            </div>

            {r.contactedAt && (
              <div className="mt-2 rounded-lg bg-emerald-50 p-2 text-[11px] text-emerald-700">
                اتصل بيه {r.contactedByName || ''} بتاريخ {fmtDate(r.contactedAt)}
                {r.contactNotes ? ` — ${r.contactNotes}` : ''}
              </div>
            )}

            {canEdit && r.status === 'PENDING' && (
              <div className="mt-3 flex gap-2">
                {r.customer?.phone && (
                  <a href={`tel:${r.customer.phone}`} className="flex-1 rounded-lg bg-blue-600 px-3 py-2 text-center text-xs font-bold text-white">
                    📞 اتصل
                  </a>
                )}
                <button
                  onClick={async () => {
                    const notes = prompt('ملاحظة عن الاتصال (اختيارية):') ?? ''
                    try { await api.markSolarContacted(r.id, notes); await onChanged() }
                    catch (e) { alert(e instanceof Error ? e.message : 'تعذر التسجيل') }
                  }}
                  className="flex-1 rounded-lg bg-emerald-600 px-3 py-2 text-xs font-bold text-white"
                >
                  ✅ سجّل إني اتصلت
                </button>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
