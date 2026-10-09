import { useCallback, useEffect, useState } from 'react'
import { api, type CarWashVehicle } from '../api'

// ═══ 🧽 غسل السيارات — شاشة عامل الغسل (قرار (ع) 10-09) ═══
// يفتح النظام، يلگه السيارات، والي يغسلها يضغط «تم الغسل». وبعدها أبو
// الكميات يقيّم السيارات بمتابعة السيارات ويشوف منو غسلها ووكت.
const time = (iso: string) => new Date(iso).toLocaleTimeString('ar-IQ', { hour: 'numeric', minute: '2-digit' })

export default function CarWashPage() {
  const [rows, setRows] = useState<CarWashVehicle[] | null>(null)
  const [busy, setBusy] = useState<string | null>(null)
  const [err, setErr] = useState<string | null>(null)

  const load = useCallback(() => {
    api.getCarWash().then((r) => setRows(r.vehicles)).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر التحميل'))
  }, [])
  useEffect(load, [load])

  const toggle = async (v: CarWashVehicle) => {
    if (v.washedAt && !window.confirm(`تلغي «تم الغسل» عن ${v.name}؟`)) return
    setBusy(v.vehicleId); setErr(null)
    try {
      if (v.washedAt) await api.unmarkCarWashed(v.vehicleId)
      else await api.markCarWashed(v.vehicleId)
      load()
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(null) }
  }

  const done = rows?.filter((r) => r.washedAt).length ?? 0
  return (
    <div dir="rtl" className="mx-auto max-w-4xl space-y-4">
      <div className="rounded-2xl bg-gradient-to-l from-cyan-700 to-sky-600 p-5 text-white shadow">
        <h1 className="text-2xl font-extrabold">🧽 غسل السيارات</h1>
        <p className="mt-1 text-sm text-cyan-50">اضغط «تم الغسل» على كل سيارة تغسلها اليوم.</p>
        {rows && <p className="mt-3 inline-block rounded-lg bg-white/20 px-3 py-1.5 text-sm font-bold">غسلت اليوم: {done} من {rows.length}</p>}
      </div>
      {err && <p className="rounded-xl bg-red-50 px-4 py-2 text-sm font-bold text-red-700">{err}</p>}
      {!rows ? <p className="text-slate-400">جاري التحميل…</p> : (
        <div className="grid gap-3 sm:grid-cols-2">
          {rows.map((v) => (
            <div key={v.vehicleId} className={`rounded-2xl border p-4 shadow-sm ${v.washedAt ? 'border-emerald-300 bg-emerald-50' : 'border-slate-200 bg-white'}`}>
              <p className="text-lg font-extrabold text-[#0f2040]">🚙 {v.name}</p>
              {!v.plateNumber.startsWith('بلا-رقم-') && <p className="text-xs text-slate-500">{v.plateNumber}</p>}
              {v.washedAt && <p className="mt-1 text-xs font-bold text-emerald-700">✅ انغسلت {time(v.washedAt)}{v.washedBy ? ` — ${v.washedBy}` : ''}</p>}
              <button disabled={busy === v.vehicleId} onClick={() => void toggle(v)}
                className={`mt-3 w-full rounded-xl py-3 text-base font-extrabold disabled:opacity-50 ${v.washedAt ? 'border border-emerald-300 bg-white text-emerald-700' : 'bg-cyan-600 text-white shadow'}`}>
                {v.washedAt ? '↩︎ تراجع' : '✅ تم الغسل'}
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
