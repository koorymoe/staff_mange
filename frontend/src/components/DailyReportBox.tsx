import { useState } from 'react'
import { apiRequest } from '../api'
import { useApi } from '../utils/useApi'
import { todayBaghdad } from '../utils/format'

// (11-02 — تعديلات ٩.١) تقرير إنجاز اليوم — اكتبه قبل ما تطلع. الي ما يكتبه،
// باچر أول ما يفتح النظام تنقفل الشاشات لحد ما يكتبه (ويتأشّر متأخر).
export default function DailyReportBox() {
  const [reload, setReload] = useState(0)
  const { data } = useApi(() => apiRequest<{ todayDone: boolean }>('/me/work-gate'), [reload])
  const [body, setBody] = useState('')
  const [msg, setMsg] = useState('')
  const [busy, setBusy] = useState(false)
  if (!data) return null
  if (data.todayDone) return <p className="rounded-xl bg-emerald-50 p-3 text-sm font-bold text-emerald-700">✅ كتبت تقرير إنجاز اليوم.</p>
  const save = async () => {
    setBusy(true); setMsg('')
    try { await apiRequest('/me/daily-report', { method: 'POST', body: JSON.stringify({ day: todayBaghdad(), body }) }); setBody(''); setReload((n) => n + 1) }
    catch (e) { setMsg(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }
  return (
    <div className="rounded-2xl border border-amber-200 bg-amber-50 p-4">
      <p className="font-bold text-amber-900">📝 تقرير إنجاز اليوم</p>
      <p className="text-xs text-amber-800">اكتبه قبل ما تطلع. إذا ما كتبته، باچر النظام ما يفتحلك إلا بعد ما تكتبه.</p>
      <textarea value={body} onChange={(e) => setBody(e.target.value)} rows={4} className="mt-2 w-full rounded-xl border border-amber-300 bg-white p-2 text-sm" placeholder="شنو سويت اليوم بالتفصيل؟" />
      {msg && <p className="text-xs text-red-600">{msg}</p>}
      <button disabled={busy || body.trim().length < 20} onClick={() => void save()} className="mt-1 rounded-xl bg-amber-600 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">إرسال</button>
    </div>
  )
}
