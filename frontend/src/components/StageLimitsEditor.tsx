import { useEffect, useState } from 'react'
import { api, type StageLimitRow } from '../api'

// ═══ ⏱️ حدود مراحل المشروع (قرار (ع) 10-10) ═══
// ماتركس يتعلّم الحد من المشاريع القديمة (الوسيط × ١٫٥، أقل شي ٣ أيام). المالك
// يگدر يثبّت حد بإيده، و«رجّع لماتركس» يمسح الحد اليدوي.
export default function StageLimitsEditor() {
  const [rows, setRows] = useState<StageLimitRow[] | null>(null)
  const [draft, setDraft] = useState<Record<string, string>>({})
  const [err, setErr] = useState('')
  useEffect(() => { api.getStageLimits().then(setRows).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر')) }, [])
  const save = (stage: string, days: number | null) =>
    api.setStageLimit(stage, days).then((r) => { setRows(r); setDraft((d) => ({ ...d, [stage]: '' })) }).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر'))
  if (err) return <p className="text-sm text-red-600">{err}</p>
  if (!rows) return null
  return (
    <div dir="rtl" className="rounded-2xl border border-slate-200 bg-white p-4">
      <h3 className="text-lg font-extrabold text-[#0f2040]">⏱️ حدود مراحل المشروع</h3>
      <p className="mb-3 text-xs text-slate-500">ماتركس يتعلّم الحد من المشاريع الي خلصت. إذا تريد تثبّت حد بإيدك اكتبه واحفظ. الي يتجاوز الحد لازم يكتب سبب، والمراقب يشوف.</p>
      <div className="space-y-2">
        {rows.map((r) => (
          <div key={r.stage} className="flex flex-wrap items-center gap-2 rounded-xl bg-slate-50 px-3 py-2 text-sm">
            <b className="min-w-[9rem]">{r.stage}</b>
            <span className="text-xs text-slate-500">🤖 ماتركس: {r.learned} يوم{r.median ? ` (الوسيط ${r.median})` : r.learned === 7 ? ' (ماكو عينات بعد)' : ' (الوسيط أقل من يوم)'}</span>
            <span className={`text-xs font-bold ${r.manual ? 'text-violet-700' : 'text-slate-400'}`}>المعتمد: {r.manual ?? r.learned} يوم{r.manual ? ' (يدوي)' : ''}</span>
            <span className="flex-1" />
            <input type="number" min={1} max={365} value={draft[r.stage] ?? ''} placeholder="يدوي"
              onChange={(e) => setDraft({ ...draft, [r.stage]: e.target.value })}
              className="w-20 rounded-lg border border-slate-300 px-2 py-1 text-sm" />
            <button type="button" disabled={!Number(draft[r.stage])} onClick={() => void save(r.stage, Number(draft[r.stage]))}
              className="rounded-lg bg-[#2c5aad] px-3 py-1 text-xs font-bold text-white disabled:opacity-40">احفظ</button>
            {r.manual != null && <button type="button" onClick={() => void save(r.stage, null)} className="text-xs font-bold text-slate-500 underline">رجّع لماتركس</button>}
          </div>
        ))}
      </div>
    </div>
  )
}
