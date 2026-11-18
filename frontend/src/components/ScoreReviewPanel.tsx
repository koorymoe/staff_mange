import { useCallback, useEffect, useState } from 'react'
import { apiRequest } from '../api'
import { fmtDateTime } from '../utils/date'

// ═══ ⚖️ تأكيد التقييمات السلبية (دازاي 11-18) ═══
// مدير الإدارة أو نائبه يأكد تقييم الليدر السيء، ويحدد منو الغلطان بالمشكلة الوظيفية.
// الي ما يتأكد ما ينحسب على الموظف. الخادم يرجّع بس منتسبي إدارته — وإذا ماكو شي ما يبين.
interface Item {
  kind: 'LEADER_RATING' | 'ISSUE'; id: string; employee: string; empId: string
  other: string | null; otherId: string | null; score: number | null; title: string; note: string | null; createdAt: string
}

export default function ScoreReviewPanel() {
  const [items, setItems] = useState<Item[]>([])
  const [busy, setBusy] = useState<string | null>(null)
  const [err, setErr] = useState('')
  const load = useCallback(() => {
    apiRequest<{ items: Item[] }>('/score-reviews').then((d) => setItems(d.items)).catch(() => setItems([]))
  }, [])
  useEffect(() => { load() }, [load])
  const act = async (id: string, path: string, body: object) => {
    setBusy(id); setErr('')
    try { await apiRequest(path, { method: 'POST', body: JSON.stringify(body) }); load() }
    catch (e) { setErr(e instanceof Error ? e.message : 'تعذر الحفظ') }
    finally { setBusy(null) }
  }
  if (items.length === 0) return null
  const btn = 'rounded-lg px-3 py-1.5 text-xs font-bold disabled:opacity-50'
  return (
    <div dir="rtl" className="rounded-2xl border border-amber-200 bg-amber-50/60 p-4">
      <p className="font-extrabold text-amber-900">⚖️ تقييمات سلبية تنتظر تأكيدك ({items.length})</p>
      <p className="mb-3 text-xs text-amber-800">الي ما تأكده ما ينحسب على الموظف. تأكد إنه صدگ غلطان قبل ما توافق.</p>
      {err && <p className="mb-2 text-xs text-red-600">{err}</p>}
      <ul className="space-y-2">
        {items.map((x) => (
          <li key={x.kind + x.id} className="rounded-xl bg-white p-3 ring-1 ring-amber-100">
            {x.kind === 'LEADER_RATING' ? (
              <>
                <p className="text-sm"><b>{x.employee}</b> — قيّمه الليدر <b>{x.other}</b> بـ<b className="text-red-600">{x.score} من ٥</b> · {x.title}</p>
                {x.note && <p className="mt-1 text-xs text-slate-600">«{x.note}»</p>}
                <p className="text-[11px] text-slate-400">{fmtDateTime(x.createdAt)}</p>
                <div className="mt-2 flex flex-wrap gap-2">
                  <button type="button" disabled={busy === x.id} onClick={() => void act(x.id, `/score-reviews/ratings/${x.id}`, { confirm: true })}
                    className={`${btn} bg-red-600 text-white hover:bg-red-700`}>صحيح — ينحسب عليه</button>
                  <button type="button" disabled={busy === x.id} onClick={() => void act(x.id, `/score-reviews/ratings/${x.id}`, { confirm: false })}
                    className={`${btn} bg-slate-100 text-slate-700 hover:bg-slate-200`}>مو صحيح — ما ينحسب</button>
                </div>
              </>
            ) : (
              <>
                <p className="text-sm">مشكلة وظيفية: <b>{x.title}</b> — بين <b>{x.employee}</b>{x.other ? <> و<b>{x.other}</b></> : null}</p>
                {x.note && <p className="mt-1 line-clamp-3 text-xs text-slate-600">{x.note}</p>}
                <p className="text-[11px] text-slate-400">{fmtDateTime(x.createdAt)}</p>
                <div className="mt-2 flex flex-wrap items-center gap-2">
                  <span className="text-xs font-bold text-slate-600">منو الغلطان؟</span>
                  <button type="button" disabled={busy === x.id} onClick={() => void act(x.id, `/score-reviews/issues/${x.id}`, { faultEmployeeId: x.empId })}
                    className={`${btn} bg-red-50 text-red-700 hover:bg-red-100`}>{x.employee}</button>
                  {x.otherId && <button type="button" disabled={busy === x.id} onClick={() => void act(x.id, `/score-reviews/issues/${x.id}`, { faultEmployeeId: x.otherId })}
                    className={`${btn} bg-red-50 text-red-700 hover:bg-red-100`}>{x.other}</button>}
                  <button type="button" disabled={busy === x.id} onClick={() => void act(x.id, `/score-reviews/issues/${x.id}`, { faultEmployeeId: '' })}
                    className={`${btn} bg-slate-100 text-slate-700 hover:bg-slate-200`}>ماكو غلطان</button>
                </div>
              </>
            )}
          </li>
        ))}
      </ul>
    </div>
  )
}
