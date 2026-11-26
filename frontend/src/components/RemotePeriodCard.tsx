import { useEffect, useState } from 'react'
import { api } from '../api'
import type { RemoteHoursConfig, RemoteHoursConfigState } from '../apiTypes'
import { card } from '../ui/classes'
import { notify } from '../ui/notify'

// ═══ (دازاي 11-26) فترة ساعات البيت وموعد رفعها ═══
// «شهرنه يبدي من يوم 26 وينتهي يوم 25»: الأدمن يحدد يوم البداية وشوكت ينرفعن.
// الساعات تنحسب من يوم البداية لحد اليوم الي قبله بالشهر الجاي، والرفع يسكّر
// الفترة الي خلصت بس — الي اشتغل بالفترة الجديدة يبقى مفتوح.

const inp = 'w-20 rounded-lg border border-slate-300 px-2 py-1.5 text-sm'
const day = (v: string) => Math.min(28, Math.max(1, Math.floor(Number(v) || 1)))

export default function RemotePeriodCard({ canEdit }: { canEdit: boolean }) {
  const [st, setSt] = useState<RemoteHoursConfigState | null>(null)
  const [c, setC] = useState<RemoteHoursConfig | null>(null)
  const [busy, setBusy] = useState(false)
  useEffect(() => {
    api.getRemoteConfig().then((x) => { setSt(x); setC(x.config) }).catch(() => {})
  }, [])
  if (!st || !c) return null
  const endDay = c.periodStartDay <= 1 ? 'آخر الشهر' : `يوم ${c.periodStartDay - 1} بالشهر الجاي`
  const save = async () => {
    setBusy(true)
    try {
      const x = await api.saveRemoteConfig(c)
      setSt(x); setC(x.config)
      notify('انحفظت فترة ساعات البيت.')
    } catch (e) {
      notify(e instanceof Error ? e.message : 'تعذّر الحفظ')
    } finally {
      setBusy(false)
    }
  }
  return (
    <section className={card}>
      <h2 className="mb-1 font-extrabold text-[#0f2040]">🗓️ فترة الشهر وموعد رفع الساعات</h2>
      <p className="text-sm text-slate-600">
        الفترة الحالية: من <b dir="ltr">{st.from}</b> إلى <b dir="ltr">{st.to}</b> — وترتفع <b dir="ltr">{st.closeAt}</b>
      </p>
      {canEdit && (
        <div className="mt-3 flex flex-wrap items-end gap-3 text-sm">
          <label className="flex flex-col gap-1">
            <span className="text-xs text-slate-500">الشهر يبدي يوم</span>
            <input type="number" min={1} max={28} className={inp} value={c.periodStartDay} onChange={(e) => setC({ ...c, periodStartDay: day(e.target.value) })} />
          </label>
          <span className="pb-2 text-xs text-slate-500">وينتهي {endDay}</span>
          <label className="flex flex-col gap-1">
            <span className="text-xs text-slate-500">ينرفعن يوم</span>
            <input type="number" min={1} max={28} className={inp} value={c.closeDay} onChange={(e) => setC({ ...c, closeDay: day(e.target.value) })} />
          </label>
          <label className="flex flex-col gap-1">
            <span className="text-xs text-slate-500">الساعة (0-23)</span>
            <input type="number" min={0} max={23} className={inp} value={c.closeHour} onChange={(e) => setC({ ...c, closeHour: Math.min(23, Math.max(0, Math.floor(Number(e.target.value) || 0))) })} />
          </label>
          <button type="button" disabled={busy} onClick={save} className="rounded-lg bg-[#0f2040] px-4 py-1.5 font-bold text-white disabled:opacity-50">
            حفظ
          </button>
        </div>
      )}
    </section>
  )
}
