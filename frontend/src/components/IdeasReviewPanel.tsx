import { useCallback, useEffect, useState } from 'react'
import { apiRequest } from '../api'
import { askInput } from '../ui/dialogs'
import { notify } from '../ui/notify'
import { fmtDateTime } from '../utils/date'
import { IDEA_CATEGORIES, IDEA_STATUS, type Idea } from './ideaTypes'

// 💡 أفكار الموظفين (دازاي 11-21 — كايزن): مدير/نائب الإدارة يرد خلال ٤٨ ساعة.
// الخادم يرجّع بس منتسبي إدارته (المالك/المدير الكل). إذا ماكو شي ما يبين.
export default function IdeasReviewPanel() {
  const [items, setItems] = useState<Idea[]>([])
  const [busy, setBusy] = useState<string | null>(null)
  const load = useCallback(() => { apiRequest<Idea[]>('/ideas').then(setItems).catch(() => setItems([])) }, [])
  useEffect(() => { load() }, [load])
  const decide = async (x: Idea, status: string) => {
    const reply = status === 'REJECTED' ? await askInput('ليش لا؟ (الموظف يشوف ردك)') : (await askInput('رد للموظف (اختياري)')) ?? ''
    if (status === 'REJECTED' && !reply) return
    setBusy(x.id)
    try { await apiRequest(`/ideas/${x.id}/decide`, { method: 'POST', body: JSON.stringify({ status, reply }) }); load() }
    catch (e) { notify(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setBusy(null) }
  }
  const open = items.filter((i) => i.status === 'NEW' || i.status === 'TRYING')
  if (items.length === 0) return null
  const btn = 'rounded-lg px-3 py-1.5 text-xs font-bold disabled:opacity-50'
  return (
    <div dir="rtl" className="rounded-2xl border border-amber-200 bg-amber-50/40 p-4">
      <p className="font-extrabold text-amber-900">💡 أفكار الموظفين ({open.length} تنتظر رد)</p>
      <p className="mb-3 text-xs text-amber-800">رد خلال ٤٨ ساعة وإلا تروح للمالك. الي «نطبقها» تنحسب لصاحبها شغل إضافي بتقييمه.</p>
      <ul className="space-y-2">
        {items.map((x) => (
          <li key={x.id} className="rounded-xl bg-white p-3 ring-1 ring-amber-100">
            <p className="text-sm"><b>{x.employee}</b> · {IDEA_CATEGORIES[x.category] ?? x.category}{x.late && <span className="mr-1 rounded bg-red-100 px-1 text-[10px] font-bold text-red-700">عدّت ٤٨ ساعة</span>}</p>
            <p className="mt-1 text-sm text-slate-700">{x.text}</p>
            <p className="text-[11px] text-slate-400">{fmtDateTime(x.createdAt)} · {IDEA_STATUS[x.status] ?? x.status}{x.reply && <> — «{x.reply}»</>}</p>
            {(x.status === 'NEW' || x.status === 'TRYING') && (
              <div className="mt-2 flex flex-wrap gap-2">
                <button type="button" disabled={busy === x.id} onClick={() => void decide(x, 'ACCEPTED')} className={`${btn} bg-emerald-600 text-white hover:bg-emerald-700`}>✅ نطبقها</button>
                {x.status === 'NEW' && <button type="button" disabled={busy === x.id} onClick={() => void decide(x, 'TRYING')} className={`${btn} bg-sky-50 text-sky-700 hover:bg-sky-100`}>🧪 نجربها</button>}
                <button type="button" disabled={busy === x.id} onClick={() => void decide(x, 'REJECTED')} className={`${btn} bg-slate-100 text-slate-700 hover:bg-slate-200`}>❌ لا</button>
              </div>
            )}
          </li>
        ))}
      </ul>
    </div>
  )
}
