import { useState } from 'react'
import { apiRequest } from '../api'
import { useApi } from '../utils/useApi'
import { DutyForm } from './WorkGate'

// (11-02 — ٩.٢) الي عليك تقيّمه اليوم — قيّمه قبل ما يخلص اليوم، وإلا باچر
// تنقفل الشاشات لحد ما تعوّض وتبرّر (وينكتب بسجل الإهمال).
type Duty = { day: string; items: { stage: string; bookingId: string; code: string; rateeId: string; rateeName: string; hint: string }[]; labels: Record<string, string> }

export default function TodayRatings() {
  const [reload, setReload] = useState(0)
  const { data } = useApi(() => apiRequest<Duty>('/me/rating-duty'), [reload])
  if (!data || data.items.length === 0) return null
  return (
    <div className="rounded-2xl border border-sky-200 bg-sky-50 p-4">
      <p className="mb-2 text-xs text-sky-800">عليك {data.items.length} تقييم اليوم — إذا ما خلصتهن، باچر النظام يوقفك لحد ما تعوّض وتبرّر.</p>
      <DutyForm duty={data} late={false} onDone={() => setReload((n) => n + 1)} />
    </div>
  )
}
