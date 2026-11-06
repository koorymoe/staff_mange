import { useState } from 'react'
import { apiRequest } from '../api'

// 👍/👎 على رد الوكيل (دازاي 11-06): الوكيل يشوف تقييمه بمراجعته الذاتية ويتعلم منه.
// path = مسار الحفظ بالسيرفر (رسالة ماتركس أو كمال).
export default function AgentFeedback({ path }: { path: string }) {
  const [state, setState] = useState<'idle' | 'note' | 'sent' | 'error'>('idle')
  const [good, setGood] = useState<boolean | null>(null)
  const [note, setNote] = useState('')
  const save = async (g: boolean, n = '') => {
    try {
      await apiRequest(path, { method: 'POST', body: JSON.stringify({ good: g, note: n }) })
      setGood(g)
      setState('sent')
    } catch {
      setState('error')
    }
  }
  if (state === 'sent')
    return <span className="text-[10px] font-bold text-slate-400">{good ? '👍 شكراً' : '👎 وصل — راح يتعلم منه'}</span>
  return (
    <span className="inline-flex flex-wrap items-center gap-1 text-[10px]">
      <button type="button" aria-label="جواب زين" title="جواب زين" onClick={() => void save(true)} className="rounded px-1 hover:bg-emerald-50">👍</button>
      <button type="button" aria-label="جواب مو زين" title="جواب مو زين" onClick={() => setState('note')} className="rounded px-1 hover:bg-rose-50">👎</button>
      {state === 'error' && <span className="text-rose-500">تعذر الحفظ</span>}
      {state === 'note' && (
        <span className="inline-flex items-center gap-1">
          <input
            autoFocus
            value={note}
            onChange={(e) => setNote(e.target.value)}
            onKeyDown={(e) => { if (e.key === 'Enter') void save(false, note) }}
            placeholder="شنو الغلط؟ (اختياري)"
            className="w-40 rounded border border-rose-200 px-1.5 py-0.5 text-[11px] text-slate-700 outline-none"
          />
          <button type="button" onClick={() => void save(false, note)} className="rounded bg-rose-600 px-2 py-0.5 font-bold text-white">إرسال</button>
        </span>
      )}
    </span>
  )
}
