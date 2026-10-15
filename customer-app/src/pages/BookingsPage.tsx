import { useState } from 'react'
import { Link } from 'react-router-dom'
import type { CBooking } from '../api'
import { useBookings, useServices } from '../data'
import ServiceArt from '../components/ServiceArt'
import { Empty, Skel, StatusPill, TopBar } from '../components/ui'
import { fmtDate, fmtTime } from '../lib'

const TABS: { k: string; label: string; match: (b: CBooking) => boolean }[] = [
  { k: 'active', label: 'الفعّالة', match: (b) => ['PENDING', 'CONFIRMED', 'WAITING'].includes(b.status) },
  { k: 'progress', label: 'قيد التنفيذ', match: (b) => ['IN_PROGRESS', 'PARTIAL'].includes(b.status) },
  { k: 'done', label: 'المكتملة', match: (b) => b.status === 'COMPLETED' },
  { k: 'cancel', label: 'الملغاة', match: (b) => b.status === 'CANCELLED' },
]

// ═══ حجوزاتي (شاشة 8) ═══
export default function BookingsPage() {
  const { data, err } = useBookings()
  const { data: services } = useServices()
  const [tab, setTab] = useState('active')
  const cur = TABS.find((t) => t.k === tab)!
  const list = (data ?? []).filter(cur.match)

  return (
    <div className="page-in">
      <TopBar title="حجوزاتي" back={false} />
      <div className="grid grid-cols-4 gap-1.5 px-4 pt-1 pb-3">
        {TABS.map((t) => (
          <button key={t.k} onClick={() => setTab(t.k)} className={`rounded-lg px-1 py-2 text-[11.5px] font-bold whitespace-nowrap transition ${tab === t.k ? 'bg-navy-2 text-white shadow-md' : 'bg-white text-mute ring-1 ring-line'}`}>
            {t.label} ({(data ?? []).filter(t.match).length})
          </button>
        ))}
      </div>
      <div className="space-y-3 px-4 pb-4">
        {err && <p className="rounded-xl bg-rose-50 p-3 text-sm text-rose-600">{err}</p>}
        {!data && [0, 1, 2].map((i) => <Skel key={i} className="h-28" />)}
        {data && list.length === 0 && (
          <Empty icon="🗓️" title={`ماكو حجوزات ${cur.label}`}>
            <Link to="/request" className="mt-4 inline-block rounded-xl bg-navy-2 px-6 py-2.5 text-[14px] font-extrabold text-white">طلب خدمة</Link>
          </Empty>
        )}
        {list.map((b, i) => {
          const s = b.serviceId ? services?.find((x) => x.id === b.serviceId) : undefined
          const when = b.scheduledAt ?? b.createdAt
          return (
            <Link key={b.id} to={`/bookings/${b.id}`} className="touch-card card rise flex gap-3 p-3" style={{ animationDelay: `${i * 40}ms` }}>
              <div className="min-w-0 flex-1">
                <div className="flex items-start justify-between gap-2">
                  <div className="min-w-0">
                    <p className="text-[11.5px] font-bold text-mute" dir="ltr">#{b.code}</p>
                    <p className="mt-0.5 truncate text-[15px] font-extrabold text-ink">{b.services || 'خدمة'}</p>
                  </div>
                  <StatusPill status={b.status} />
                </div>
                <p className="mt-2 flex items-center gap-1.5 text-[12px] text-mute">
                  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 2" /></svg>
                  {fmtDate(when)}{b.scheduledAt && ` - ${fmtTime(b.scheduledAt)}`}
                </p>
                {b.address && <p className="mt-1 flex items-center gap-1.5 truncate text-[12px] text-mute">
                  <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M12 22s7-6.2 7-12a7 7 0 0 0-14 0c0 5.8 7 12 7 12z" /><circle cx="12" cy="10" r="2.5" /></svg>
                  <span className="truncate">{b.address.split('—')[0].trim()}</span></p>}
              </div>
              <ServiceArt id={s?.id} name={s?.name ?? b.services} hasImage={s?.hasImage} className="h-[84px] w-[84px] shrink-0 self-center rounded-xl" />
            </Link>
          )
        })}
      </div>
    </div>
  )
}
