import { useEffect, useMemo, useState } from 'react'
import { api, type InventoryCheck } from '../api'

// ═══ النواقص المفتوحة بكل أعمارها — (ع) 11-04 ═══
// قبل، زر «تم التوفير» چان بجرد اليوم بس، فنقص أغسطس ما يلگاله مكان ينسكّر.
// هسه الإداري يشوف كل المفتوح (الأقدم أول) ويفلتر بالعمر. والنقص الي جرد
// أحدث للموظف نفسه أثبت وجود أدواته ينسكّر لحاله بالخادم.
const AGES = [
  { key: 0, label: 'الكل' },
  { key: 3, label: '+٣ أيام' },
  { key: 7, label: '+أسبوع' },
  { key: 30, label: '+شهر' },
]
// (ل) 11-11 طلب المخازن: أولوية — عاجل (عنده شغل اليوم/باچر)، مهم (+٧ أيام)، عادي. والإداري يغيّرها.
const PRI: Record<string, { label: string; cls: string }> = {
  URGENT: { label: '🔴 عاجل', cls: 'bg-red-100 text-red-700' },
  HIGH: { label: '🟠 مهم', cls: 'bg-amber-100 text-amber-800' },
  NORMAL: { label: '⚪ عادي', cls: 'bg-slate-100 text-slate-600' },
}
const daysOld = (s: string) => Math.floor((Date.now() - new Date(s).getTime()) / 86400000)

export default function OpenShortagesPanel({ canResolve }: { canResolve: boolean }) {
  const [rows, setRows] = useState<InventoryCheck[] | null>(null)
  const [err, setErr] = useState('')
  const [minAge, setMinAge] = useState(0)
  const [pri, setPri] = useState('')
  const [busy, setBusy] = useState<string | null>(null)
  const [reload, setReload] = useState(0)
  useEffect(() => {
    let alive = true
    api.getOpenShortages()
      .then((r) => { if (alive) { setRows(r); setErr('') } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [reload])
  const shown = useMemo(() => (rows ?? []).filter((c) => daysOld(c.checkedAt) >= minAge && (!pri || c.priority === pri)), [rows, minAge, pri])

  const setPriority = async (id: string, p: string) => {
    setBusy(id)
    try { await api.setShortagePriority(id, p); setReload((n) => n + 1) }
    catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') }
    finally { setBusy(null) }
  }
  const resolve = async (id: string) => {
    setBusy(id)
    try { await api.resolveInventoryCheck(id); setReload((n) => n + 1) }
    catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') }
    finally { setBusy(null) }
  }

  return (
    <div className="space-y-3">
      <p className="text-sm text-slate-500">
        كل نقص بالجرد بعده ما انحل — العاجل أول (الفني عنده شغل اليوم أو باچر)، بعده المهم (+٧ أيام)، وبكل مستوى الأقدم أول. إذا الموظف سوّى جرد أحدث وطلعت أدواته موجودة، النقص ينسكّر لحاله.
      </p>
      <div className="flex flex-wrap items-center gap-2">
        {AGES.map((a) => (
          <button key={a.key} type="button" onClick={() => setMinAge(a.key)}
            className={`rounded-full px-3 py-1 text-xs font-bold ${minAge === a.key ? 'bg-brand-700 text-white' : 'bg-slate-100 text-slate-600'}`}>
            {a.label}
          </button>
        ))}
        <span className="mx-1 h-4 w-px bg-slate-300" />
        {[['', 'كل الأولويات'], ['URGENT', PRI.URGENT.label], ['HIGH', PRI.HIGH.label], ['NORMAL', PRI.NORMAL.label]].map(([k, l]) => (
          <button key={k} type="button" onClick={() => setPri(k)}
            className={`rounded-full px-3 py-1 text-xs font-bold ${pri === k ? 'bg-brand-700 text-white' : 'bg-slate-100 text-slate-600'}`}>
            {l}{k && rows ? ` (${rows.filter((c) => c.priority === k).length})` : ''}
          </button>
        ))}
        {rows && <span className="text-xs text-slate-400">{shown.length} من {rows.length}</span>}
      </div>
      {err && <p className="text-sm text-red-600">{err}</p>}
      {rows === null && !err && <p className="text-slate-400">جاري التحميل…</p>}
      {rows && shown.length === 0 && (
        <div className="rounded-xl bg-white p-8 text-center text-slate-400 shadow-sm">ماكو نواقص مفتوحة {minAge ? 'بهالعمر' : ''} 👌</div>
      )}
      {shown.map((c) => {
        const d = daysOld(c.checkedAt)
        return (
          <div key={c.id} className={`rounded-xl border bg-white p-4 shadow-sm ${d >= 30 ? 'border-red-300' : d >= 7 ? 'border-amber-300' : 'border-white'}`}>
            <div className="flex flex-wrap items-center justify-between gap-2">
              <span className="flex flex-wrap items-center gap-2">
                <span className="font-bold text-slate-800">{c.employee?.name || '-'}</span>
                {c.priority && <span className={`rounded-full px-2 py-0.5 text-xs font-bold ${PRI[c.priority].cls}`}>{PRI[c.priority].label}{c.manualPriority ? ' · يدوي' : ''}</span>}
                {c.nextBookingAt && <span className="text-xs text-red-700">عنده حجز {new Date(c.nextBookingAt).toDateString() === new Date().toDateString() ? 'اليوم' : 'باچر'}</span>}
              </span>
              <span className={`rounded-full px-3 py-1 text-xs font-bold ${d >= 30 ? 'bg-red-100 text-red-700' : d >= 7 ? 'bg-amber-100 text-amber-700' : 'bg-slate-100 text-slate-600'}`}>
                {d === 0 ? 'اليوم' : `من ${d} يوم`} · {new Date(c.checkedAt).toLocaleDateString('en-GB')}
              </span>
            </div>
            {c.missingItems && <p className="mt-2 rounded-lg bg-amber-50 px-3 py-2 text-sm text-amber-800">الناقص: {c.missingItems}</p>}
            {canResolve && (
              <div className="mt-2 flex flex-wrap items-center justify-end gap-2">
                <select aria-label="الأولوية" value={c.manualPriority ?? ''} disabled={busy === c.id} onChange={(e) => void setPriority(c.id, e.target.value)}
                  className="rounded-lg border border-slate-300 bg-white px-2 py-1 text-xs">
                  <option value="">أولوية تلقائية</option>
                  <option value="URGENT">عاجل</option>
                  <option value="HIGH">مهم</option>
                  <option value="NORMAL">عادي</option>
                </select>
                <button type="button" onClick={() => void resolve(c.id)} disabled={busy === c.id}
                  className="rounded-lg bg-emerald-700 px-3 py-1.5 text-xs font-bold text-white hover:bg-emerald-800 disabled:opacity-50">
                  {busy === c.id ? 'جاري…' : 'تم الحل ✓'}
                </button>
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}
