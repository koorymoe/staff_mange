import { useState } from 'react'
import { apiRequest } from '../api'
import { useApi } from '../utils/useApi'

// ═══ 🧠 تحليل ماتركس للتقرير — ماتركس يقرا التقرير كله مقابل كل شي سجّله النظام ذاك اليوم ═══
// استشاري: ما يغيّر التقييم. own=true للموظف (تقريره هو بس)، وإلا للمراقب.

type Claim = { claim: string; evidence: string; status: 'confirmed' | 'unsupported' | 'contradicted' }
type Analysis = { verdict: string; summary: string; claims: Claim[] | null; missing: string[] | null; notes: string; updatedAt: string }
type Out = { enabled: boolean; analysis: Analysis | null }

const VERDICT: Record<string, [string, string]> = {
  CONSISTENT: ['✅ التقرير يطابق المسجّل', 'bg-emerald-50 text-emerald-700 ring-emerald-200'],
  GAPS: ['⚠️ بيه نواقص أو كلام بلا دليل', 'bg-amber-50 text-amber-700 ring-amber-200'],
  CONTRADICTS: ['❌ يناقض المسجّل بالنظام', 'bg-red-50 text-red-700 ring-red-200'],
  NO_EVIDENCE: ['➖ النظام ما مسجّل شي ينقارن بيه', 'bg-slate-50 text-slate-600 ring-slate-200'],
}
const CLAIM: Record<Claim['status'], [string, string]> = {
  confirmed: ['✓ ثابت', 'text-emerald-700'],
  unsupported: ['؟ بلا دليل', 'text-amber-700'],
  contradicted: ['✗ يناقض', 'text-red-700'],
}

export function ReportAnalysisBox({ reportId, own = false }: { reportId: string; own?: boolean }) {
  const [open, setOpen] = useState(false)
  const url = own ? `/me/daily-reports/${reportId}/analysis` : `/daily-reports/${reportId}/analysis`
  const { data, error } = useApi(() => apiRequest<Out>(url), [url], { enabled: open })
  const a = data?.analysis
  const badge = a ? VERDICT[a.verdict] : undefined
  return (
    <div className="mt-1.5 text-xs">
      <button onClick={() => setOpen((o) => !o)} className={`rounded-full px-2 py-0.5 font-bold ring-1 ${badge ? badge[1] : 'bg-indigo-50 text-indigo-700 ring-indigo-200'}`}>
        🧠 تحليل ماتركس للتقرير{badge && ` — ${badge[0]}`} {open ? '▴' : '▾'}
      </button>
      {open && (
        <div className="mt-2 space-y-1.5 rounded-xl bg-slate-50 p-3 text-slate-700">
          {error && <p className="text-red-600">{error}</p>}
          {!data && !error && <p className="text-slate-400">جاري التحميل…</p>}
          {data && !a && <p className="text-slate-500">{data.enabled ? 'ماتركس بعده ما حلّل هالتقرير — ينحلّل تلقائياً خلال دقايق أو بجولة الليل.' : 'تحليل ماتركس مطفي هسه (مفتاح الذكاء).'}</p>}
          {a && badge && (
            <>
              <p className={`inline-block rounded-full px-2 py-0.5 font-bold ring-1 ${badge[1]}`}>{badge[0]}</p>
              <p className="whitespace-pre-wrap font-bold">{a.summary}</p>
              {(a.claims ?? []).length > 0 && (
                <ul className="space-y-1">
                  {(a.claims ?? []).map((c, i) => {
                    const [l, cls] = CLAIM[c.status] ?? CLAIM.unsupported
                    return <li key={i}><b className={cls}>{l}</b> {c.claim}{c.evidence && <span className="text-slate-500"> — {c.evidence}</span>}</li>
                  })}
                </ul>
              )}
              {(a.missing ?? []).length > 0 && <div><b>سجّله النظام وما ذكره:</b><ul className="mt-1 list-inside list-disc">{(a.missing ?? []).map((m, i) => <li key={i}>{m}</li>)}</ul></div>}
              {a.notes && !own && <p className="text-slate-600">ملاحظة للمراقب: {a.notes}</p>}
              <p className="text-[11px] text-slate-400">استشاري بس — ما يغيّر التقييم.</p>
            </>
          )}
        </div>
      )}
    </div>
  )
}
