import { useEffect, useState } from 'react'
import { api, type ScoreExplainRow } from '../api'
import { roleLabel } from '../roleLabels'

// ═══ «على شنو ينقيّم» (دازاي 11-18) ═══
// كل موظف: مجموعته، القواعد الي ماتركس يحسبها عليه، منو يقيّمه من البشر، هل تنحسب عليه
// فلوس، وليش. من نفس جداول الحساب بالخادم — مو نص منفصل.
export default function ScoreExplain({ group }: { group: string }) {
  const [rows, setRows] = useState<ScoreExplainRow[] | null>(null)
  const [err, setErr] = useState('')
  const [q, setQ] = useState('')
  const [all, setAll] = useState(false)
  useEffect(() => { api.getScoreExplain().then(setRows).catch((e) => setErr(e instanceof Error ? e.message : 'تعذّر')) }, [])
  if (err) return <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{err}</p>
  if (!rows) return <p className="py-6 text-center text-slate-400">جاري الجلب…</p>
  const shown = rows.filter((r) => (all || r.group === group) && (!q.trim() || r.name.includes(q.trim())))
  return (
    <div className="space-y-3" dir="rtl">
      <div className="flex flex-wrap items-center gap-2">
        <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="دوّر على موظف…" className="min-w-[12rem] flex-1 rounded-xl border border-slate-200 px-3 py-2 text-sm" />
        <label className="flex items-center gap-1 text-xs text-slate-600"><input type="checkbox" checked={all} onChange={(e) => setAll(e.target.checked)} /> كل المجاميع</label>
      </div>
      {shown.length === 0 && <p className="text-sm text-slate-400">ماكو.</p>}
      <div className="grid gap-3 lg:grid-cols-2">
        {shown.map((r) => {
          const [prevRole, since] = r.roleChange?.split('|') ?? []
          return (
            <div key={r.id} className="rounded-2xl border border-slate-200 bg-white p-4 text-sm">
              <div className="mb-1 flex flex-wrap items-center justify-between gap-2">
                <b className="text-base text-[#0f2040]">{r.name}</b>
                <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-bold text-slate-700">{r.groupLabel}</span>
              </div>
              <p className="text-xs text-slate-500">{roleLabel(r.role)}{r.isLeader ? ' · ليدر' : ''}
                {r.aiOnly && <span className="mr-1 rounded-full bg-indigo-50 px-2 text-indigo-700">🤖 ماتركس بس</span>}
                {r.offSystem && <span className="mr-1 rounded-full bg-violet-50 px-2 text-violet-700">💻 شغله برّا النظام</span>}
                {prevRole && <span className="mr-1 rounded-full bg-sky-50 px-2 text-sky-700">🔁 چان {roleLabel(prevRole)} لحد {since}</span>}
              </p>
              <p className="mt-2 rounded-lg bg-amber-50 px-2 py-1.5 text-xs text-amber-900">💡 ليش: {r.why}</p>
              <p className="mt-2 text-xs font-bold text-slate-700">🤖 ماتركس يقيّمه على (٦٠٪):</p>
              <ul className="mr-4 list-disc text-xs text-slate-600">{r.rules.map((x) => <li key={x}>{x}</li>)}</ul>
              <p className="mt-2 text-xs font-bold text-slate-700">👥 منو يقيّمه من البشر (٤٠٪):</p>
              <p className="text-xs text-slate-600">{r.raters.length ? r.raters.join(' · ') : 'ماكو — ينحسب ماتركس بس'}</p>
              <p className="mt-2 text-xs text-slate-600">💰 تغطية الراتب: {r.revenue ? `تنحسب من ${r.revenue}` : 'ما تنحسب عليه (شغله مو مصدر فلوس)'}</p>
            </div>
          )
        })}
      </div>
    </div>
  )
}
