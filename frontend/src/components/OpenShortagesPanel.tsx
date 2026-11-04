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
const daysOld = (s: string) => Math.floor((Date.now() - new Date(s).getTime()) / 86400000)

export default function OpenShortagesPanel({ canResolve }: { canResolve: boolean }) {
  const [rows, setRows] = useState<InventoryCheck[] | null>(null)
  const [err, setErr] = useState('')
  const [minAge, setMinAge] = useState(0)
  const [busy, setBusy] = useState<string | null>(null)
  const [reload, setReload] = useState(0)
  useEffect(() => {
    let alive = true
    api.getOpenShortages()
      .then((r) => { if (alive) { setRows(r); setErr('') } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [reload])
  const shown = useMemo(() => (rows ?? []).filter((c) => daysOld(c.checkedAt) >= minAge), [rows, minAge])

  const resolve = async (id: string) => {
    setBusy(id)
    try { await api.resolveInventoryCheck(id); setReload((n) => n + 1) }
    catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') }
    finally { setBusy(null) }
  }

  return (
    <div className="space-y-3">
      <p className="text-sm text-slate-500">
        كل نقص بالجرد بعده ما انحل — الأقدم أول. إذا الموظف سوّى جرد أحدث وطلعت أدواته موجودة، النقص ينسكّر لحاله.
      </p>
      <div className="flex flex-wrap items-center gap-2">
        {AGES.map((a) => (
          <button key={a.key} type="button" onClick={() => setMinAge(a.key)}
            className={`rounded-full px-3 py-1 text-xs font-bold ${minAge === a.key ? 'bg-brand-700 text-white' : 'bg-slate-100 text-slate-600'}`}>
            {a.label}
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
              <span className="font-bold text-slate-800">{c.employee?.name || '-'}</span>
              <span className={`rounded-full px-3 py-1 text-xs font-bold ${d >= 30 ? 'bg-red-100 text-red-700' : d >= 7 ? 'bg-amber-100 text-amber-700' : 'bg-slate-100 text-slate-600'}`}>
                {d === 0 ? 'اليوم' : `من ${d} يوم`} · {new Date(c.checkedAt).toLocaleDateString('en-GB')}
              </span>
            </div>
            {c.missingItems && <p className="mt-2 rounded-lg bg-amber-50 px-3 py-2 text-sm text-amber-800">الناقص: {c.missingItems}</p>}
            {canResolve && (
              <div className="mt-2 flex justify-end">
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
