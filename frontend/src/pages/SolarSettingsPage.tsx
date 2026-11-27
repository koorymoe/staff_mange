import { useEffect, useState } from 'react'
import { apiRequest } from '../api'
import { card } from '../ui/classes'
import { notify } from '../ui/notify'

// ═══ ☀️ إعدادات الطاقة الشمسية (دازاي 11-27) ═══
// «منين اخلي الاسعار مال اقتصادي…؟» — جدولين يدويين:
//  ١) استهلاك الأجهزة: منه ينحسب حجم المنظومة الي تناسب الزبون.
//  ٢) أسعار المنظومات حسب الحجم: منها تطلع أسعار اقتصادي/وسط/غالي.

type Appliance = { name: string; minKw: number; maxKw: number }
type Band = { kw: number; economy: number; mid: number; high: number }
type Config = { appliances: Appliance[]; bands: Band[] }

const inp = 'w-full rounded-lg border border-slate-300 px-2 py-1.5 text-sm'
const num = (v: string) => Math.max(0, Number(v) || 0)
const DEFAULTS = ['ثلاجة', 'مجمدة', 'مبردة', 'مكيف', 'غسالة', 'سخان ماء', 'مضخة ماء', 'تلفزيون', 'إنارة']

export default function SolarSettingsPage() {
  const [c, setC] = useState<Config | null>(null)
  const [busy, setBusy] = useState(false)
  useEffect(() => {
    apiRequest<Config>('/solar-inquiries/config')
      .then((x) => setC({
        appliances: x.appliances.length ? x.appliances : DEFAULTS.map((name) => ({ name, minKw: 0, maxKw: 0 })),
        bands: x.bands.length ? x.bands : [3, 5, 10].map((kw) => ({ kw, economy: 0, mid: 0, high: 0 })),
      }))
      .catch((e) => notify(e instanceof Error ? e.message : 'تعذر التحميل'))
  }, [])
  if (!c) return <p className="py-8 text-center text-slate-400">جاري التحميل…</p>
  const setApp = (i: number, p: Partial<Appliance>) => setC({ ...c, appliances: c.appliances.map((a, k) => (k === i ? { ...a, ...p } : a)) })
  const setBand = (i: number, p: Partial<Band>) => setC({ ...c, bands: c.bands.map((b, k) => (k === i ? { ...b, ...p } : b)) })
  const save = async () => {
    setBusy(true)
    try {
      // الجهاز البلا رقم ما ينحفظ (ما نخمّن أرقام)
      const body = { appliances: c.appliances.filter((a) => a.name.trim() && a.maxKw > 0), bands: c.bands.filter((b) => b.kw > 0) }
      const x = await apiRequest<Config>('/solar-inquiries/config', { method: 'PUT', body: JSON.stringify(body) })
      setC(x)
      notify('انحفظت إعدادات الطاقة الشمسية ✅')
    } catch (e) {
      notify(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally {
      setBusy(false)
    }
  }
  return (
    <div className="space-y-4" dir="rtl">
      <section className={card}>
        <h2 className="text-lg font-extrabold text-[#0f2040]">🔌 استهلاك الأجهزة</h2>
        <p className="mb-3 text-sm text-slate-500">كم كيلو واط ياخذ كل جهاز (أقل وأكثر). منه النظام يحسب حجم المنظومة الي تناسب الزبون. الجهاز الي تخليه صفر ما ينحسب.</p>
        <table className="w-full text-sm">
          <thead className="text-xs text-slate-500"><tr><th className="text-right">الجهاز</th><th>أقل (كيلو واط)</th><th>أكثر (كيلو واط)</th><th /></tr></thead>
          <tbody>
            {c.appliances.map((a, i) => (
              <tr key={i}>
                <td className="py-1 pl-2"><input className={inp} value={a.name} onChange={(e) => setApp(i, { name: e.target.value })} /></td>
                <td className="px-1"><input type="number" step="any" min={0} className={inp} value={a.minKw} onChange={(e) => setApp(i, { minKw: num(e.target.value) })} /></td>
                <td className="px-1"><input type="number" step="any" min={0} className={inp} value={a.maxKw} onChange={(e) => setApp(i, { maxKw: num(e.target.value) })} /></td>
                <td><button type="button" className="px-2 text-red-600" onClick={() => setC({ ...c, appliances: c.appliances.filter((_, k) => k !== i) })}>✕</button></td>
              </tr>
            ))}
          </tbody>
        </table>
        <button type="button" className="mt-2 rounded-lg border px-3 py-1 text-sm" onClick={() => setC({ ...c, appliances: [...c.appliances, { name: '', minKw: 0, maxKw: 0 }] })}>＋ جهاز</button>
      </section>

      <section className={card}>
        <h2 className="text-lg font-extrabold text-[#0f2040]">💰 أسعار المنظومات حسب الحجم</h2>
        <p className="mb-3 text-sm text-slate-500">لكل حجم منظومة: سعر الاقتصادي والوسط والغالي (بالدينار). هاي الأسعار تطلع للموظف بالمكالمة. إذا الجدول فاضي، النظام ياخذ الأسعار من المنظومات المسجّلة بشاشة الطاقة الشمسية.</p>
        <table className="w-full text-sm">
          <thead className="text-xs text-slate-500"><tr><th className="text-right">الحجم (كيلو واط)</th><th>🌿 اقتصادي</th><th>⭐ وسط</th><th>💎 غالي</th><th /></tr></thead>
          <tbody>
            {c.bands.map((b, i) => (
              <tr key={i}>
                <td className="py-1 pl-2"><input type="number" step="any" min={0} className={inp} value={b.kw} onChange={(e) => setBand(i, { kw: num(e.target.value) })} /></td>
                <td className="px-1"><input type="number" min={0} className={inp} value={b.economy} onChange={(e) => setBand(i, { economy: num(e.target.value) })} /></td>
                <td className="px-1"><input type="number" min={0} className={inp} value={b.mid} onChange={(e) => setBand(i, { mid: num(e.target.value) })} /></td>
                <td className="px-1"><input type="number" min={0} className={inp} value={b.high} onChange={(e) => setBand(i, { high: num(e.target.value) })} /></td>
                <td><button type="button" className="px-2 text-red-600" onClick={() => setC({ ...c, bands: c.bands.filter((_, k) => k !== i) })}>✕</button></td>
              </tr>
            ))}
          </tbody>
        </table>
        <button type="button" className="mt-2 rounded-lg border px-3 py-1 text-sm" onClick={() => setC({ ...c, bands: [...c.bands, { kw: 0, economy: 0, mid: 0, high: 0 }] })}>＋ حجم</button>
      </section>

      <div className="flex justify-end">
        <button type="button" disabled={busy} onClick={save} className="rounded-lg bg-[#0f2040] px-6 py-2 font-bold text-white disabled:opacity-50">حفظ</button>
      </div>
    </div>
  )
}
