import { useEffect, useState } from 'react'
import { apiRequest } from '../api'

// ⏱ لوحة الالتزام للمنسّق (دازاي 11-06): الحجوزات الي قرب أو فات موعد التزامها ويا الزبون.
type Item = { id: string; code: string; customer: string | null; service: string | null; dueAt: string; hoursLeft: number }

export default function SlaBoardCard() {
  const [rows, setRows] = useState<Item[] | null>(null)
  useEffect(() => {
    apiRequest<Item[]>('/sla/board').then(setRows).catch(() => setRows([]))
  }, [])
  if (!rows || rows.length === 0) return null
  const over = rows.filter((r) => r.hoursLeft < 0) // الترتيب من السيرفر: الأقرب للموعد أول
  return (
    <section className="rounded-2xl border border-amber-200 bg-amber-50 p-3 text-sm" aria-label="مواعيد الالتزام">
      <b>⏱ مواعيد الالتزام ويا الزبون</b>
      {over.length > 0 && <span className="mr-2 rounded bg-rose-600 px-2 py-0.5 text-xs font-bold text-white">{over.length} فات موعدها</span>}
      <ul className="mt-1 space-y-0.5">
        {rows.slice(0, 15).map((r) => (
          <li key={r.id} className={r.hoursLeft < 0 ? 'text-rose-700' : 'text-amber-900'}>
            {r.code} — {r.customer ?? ''} {r.service ? `(${r.service})` : ''}:{' '}
            <b>{r.hoursLeft < 0 ? `فات بـ${Math.round(-r.hoursLeft)} ساعة` : `باقي ${Math.round(r.hoursLeft)} ساعة`}</b>
          </li>
        ))}
      </ul>
    </section>
  )
}
