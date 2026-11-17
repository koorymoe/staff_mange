import { useEffect, useState } from 'react'
import { api } from '../api'
import { loadFailed } from '../netErrors'
import { notify } from '../ui/notify'
import { useSession, hasRole, isAdminOrOwner } from '../session'

// ═══ أسعار اشتراك GPS (دازاي 11-17) ═══
// «خليها تدخل بالتجديد»: سعر لكل نوع، وكل طلب تجديد ياخذ سعر نوعه وقت الطلب ويثبت بالموافقة.
// يعدّلها المدير ومدير الـGPS (نفس حارس الخادم gps.put_settings).
const TYPES: { key: string; label: string }[] = [
  { key: 'THREE_MONTHS', label: '٣ أشهر' },
  { key: 'SIX_MONTHS', label: '٦ أشهر' },
  { key: 'YEARLY', label: 'سنوي' },
]

export default function GpsPriceSettings() {
  const { employee } = useSession()
  const canEdit = isAdminOrOwner(employee) || hasRole(employee, 'GPS_ADMIN')
  const [prices, setPrices] = useState<Record<string, string>>({})
  const [saving, setSaving] = useState<string | null>(null)
  useEffect(() => {
    api.getGpsPrices().then((rows) => {
      const m: Record<string, string> = {}
      for (const r of rows) m[r.subscriptionType] = String(r.price)
      setPrices(m)
    }).catch(loadFailed)
  }, [])
  if (!canEdit) return null
  const save = async (key: string) => {
    const v = Number(prices[key])
    if (!Number.isFinite(v) || v < 0) { notify('اكتب سعر صحيح'); return }
    setSaving(key)
    try { await api.setGpsPrice(key, v); notify('انحفظ السعر', 'success') } catch (e) { notify(e instanceof Error ? e.message : 'تعذّر الحفظ') } finally { setSaving(null) }
  }
  return (
    <div className="mb-4 rounded-2xl bg-white p-4 shadow-sm">
      <p className="mb-1 font-bold text-brand-900">💰 أسعار الاشتراك</p>
      <p className="mb-3 text-xs text-slate-500">كل طلب تجديد ياخذ سعر نوعه تلقائياً، ويطلع بالفاتورة.</p>
      <div className="grid gap-3 sm:grid-cols-3">
        {TYPES.map((t) => (
          <label key={t.key} className="flex items-center gap-2 text-sm">
            <span className="w-16 shrink-0 text-slate-600">{t.label}</span>
            <input type="number" min={0} inputMode="numeric" value={prices[t.key] ?? ''} placeholder="د.ع"
              onChange={(e) => setPrices((p) => ({ ...p, [t.key]: e.target.value }))}
              className="w-full rounded-lg border border-slate-300 px-2 py-1.5" />
            <button type="button" onClick={() => void save(t.key)} disabled={saving === t.key || prices[t.key] == null || prices[t.key] === ''}
              className="rounded-lg bg-brand-500 px-3 py-1.5 text-xs font-bold text-white disabled:opacity-50">حفظ</button>
          </label>
        ))}
      </div>
    </div>
  )
}
