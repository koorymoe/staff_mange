import { useEffect, useState } from 'react'
import { apiRequest } from '../api'
import { useApi } from '../utils/useApi'

// (11-02 — تعديلات ٩.١) تقرير إنجاز اليوم إجباري. الي ما يكتبه، باچر أول ما يفتح
// النظام تنقفل الشاشات لحد ما يكتبه (ويتأشّر متأخر).
// (دازاي 11-05 «ادمجهن بلقديم»): ماكو نموذج ثاني — «إنجازاتي اليوم» هو التقرير.
// هنا بس الحالة: كتبته الساعة كذا، لو بعدك وزر يوديك.
export default function DailyReportBox({ onWrite }: { onWrite?: () => void }) {
  const [reload, setReload] = useState(0)
  const { data } = useApi(() => apiRequest<{ todayDone: boolean; todayAt: string | null }>('/me/work-gate'), [reload])
  // ينطلق بعد أي حفظ ناجح (api.ts) — فبعد ما يرسل الإنجاز تتحدّث الحالة
  useEffect(() => {
    const r = () => setReload((n) => n + 1)
    window.addEventListener('matrix-refresh', r)
    return () => window.removeEventListener('matrix-refresh', r)
  }, [])
  if (!data) return null
  if (data.todayDone) {
    const at = data.todayAt ? new Date(data.todayAt).toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit', timeZone: 'Asia/Baghdad' }) : ''
    return (
      <p className="rounded-xl bg-emerald-50 p-3 text-sm font-bold text-emerald-700">
        ✅ كتبت تقرير إنجاز اليوم{at && ` — الساعة ${at}`}. تگدر تضيف عليه من «إنجازاتي اليوم».
      </p>
    )
  }
  return (
    <div className="flex flex-wrap items-center justify-between gap-2 rounded-2xl border border-amber-200 bg-amber-50 p-4">
      <div>
        <p className="font-bold text-amber-900">📝 بعدك ما كتبت تقرير إنجاز اليوم</p>
        <p className="text-xs text-amber-800">إجباري — اكتبه (أو سجّله فويس) قبل ما تطلع. إذا ما كتبته، باچر النظام ما يفتحلك إلا بعد ما تكتبه.</p>
      </div>
      {onWrite && (
        <button onClick={onWrite} className="rounded-xl bg-amber-600 px-4 py-2 text-sm font-bold text-white">اكتب التقرير ←</button>
      )}
    </div>
  )
}
