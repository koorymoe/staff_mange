import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api, imgUrl, type Account } from '../api'
import { useBookings, useHome, useServices } from '../data'
import ServiceArt from '../components/ServiceArt'
import { Skel, StatusPill } from '../components/ui'
import { fmtDate } from '../lib'

const greet = () => { const h = new Date().getHours(); return h < 12 ? 'صباح الخير' : h < 18 ? 'نهارك سعيد' : 'مساء الخير' }

// ═══ الرئيسية (شاشة 3) ═══
export default function HomePage({ me, onBadge }: { me: Account; onBadge: (n: number) => void }) {
  const nav = useNavigate()
  const home = useHome()
  const services = useServices()
  const bookings = useBookings()
  const [q, setQ] = useState('')
  const [slide, setSlide] = useState(0)

  const offers = useMemo(() => (home.data?.items ?? []).filter((p) => p.kind !== 'work'), [home.data])
  const maxDiscount = Math.max(0, ...offers.map((o) => o.discountPercent ?? 0))
  const last = bookings.data?.[0]
  const svc = services.data ?? []
  const lastSvc = last?.serviceId ? svc.find((s) => s.id === last.serviceId) : undefined

  useEffect(() => { if (home.data) onBadge(home.data.newCount) }, [home.data, onBadge])
  useEffect(() => {
    if (offers.length < 2) return
    const t = setInterval(() => setSlide((s) => (s + 1) % offers.length), 4500)
    return () => clearInterval(t)
  }, [offers.length])
  useEffect(() => {
    if (!home.data?.newCount) return
    const t = setTimeout(() => { api.seen().then(() => onBadge(0)).catch(() => {}) }, 3000)
    return () => clearTimeout(t)
  }, [home.data?.newCount, onBadge])

  const search = (e: React.FormEvent) => { e.preventDefault(); nav(`/services?q=${encodeURIComponent(q)}`) }
  const initials = me.name.trim().split(/\s+/).slice(0, 2).map((w) => w[0]).join('')

  return (
    <div className="page-in pb-4">
      {/* الرأس الكحلي */}
      <header className="safe-t rounded-b-[28px] bg-gradient-to-b from-navy-deep to-navy px-5 pt-5 pb-6 text-white">
        <div className="flex items-center gap-3">
          <Link to="/account" className="grid h-12 w-12 shrink-0 place-items-center rounded-full bg-white/15 text-[17px] font-black ring-2 ring-white/30">{initials}</Link>
          <div className="min-w-0 flex-1">
            <p className="truncate text-[18px] font-extrabold">مرحباً {me.name.split(' ')[0]} 👋</p>
            <p className="text-[12px] text-white/75">{greet()}، مرحباً بك في شركة الأماني</p>
          </div>
          <Link to="/account/edit" className="flex shrink-0 items-center gap-1 rounded-full bg-white/12 px-3 py-1.5 text-[12px] font-bold ring-1 ring-white/25">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M12 22s7-6.2 7-12a7 7 0 0 0-14 0c0 5.8 7 12 7 12z" /><circle cx="12" cy="10" r="2.5" /></svg>
            {me.city || 'المحافظة'}
          </Link>
        </div>
        <form onSubmit={search} className="mt-4 flex h-12 items-center gap-2 rounded-xl bg-white px-4 text-ink shadow-lg">
          <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="ابحث عن خدمة ..." className="min-w-0 flex-1 bg-transparent text-[14px] outline-none placeholder:text-[#a3acbf]" />
          <button aria-label="بحث" className="text-navy"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round"><circle cx="11" cy="11" r="7" /><path d="M21 21l-4.3-4.3" /></svg></button>
        </form>
      </header>

      <div className="space-y-6 px-4 pt-4">
        {/* بانر العروض */}
        {!home.data ? <Skel className="h-40" /> : offers.length > 0 ? (
          <section>
            <div className="relative h-40 overflow-hidden rounded-2xl shadow-[0_10px_24px_-12px_rgba(13,59,140,.6)]">
              {offers.map((p, i) => {
                const s = p.serviceId ? svc.find((x) => x.id === p.serviceId) : undefined
                return (
                  <Link key={p.id} to="/offers" className={`absolute inset-0 transition-opacity duration-700 ${i === slide ? 'opacity-100' : 'pointer-events-none opacity-0'}`}>
                    {p.hasImage ? <img src={imgUrl(p.id)} alt="" className="absolute inset-0 h-full w-full object-cover" />
                      : <ServiceArt name={s?.name ?? p.title} className="absolute inset-y-0 left-0 h-full w-[62%]" />}
                    <div className="absolute inset-0 bg-gradient-to-l from-navy-deep from-35% via-navy-deep/75 via-55% to-transparent to-85%" />
                    <div className="absolute inset-y-0 right-0 flex w-3/5 flex-col justify-center p-4 text-white">
                      <p className="line-clamp-2 text-[17px] leading-snug font-black">{p.title}</p>
                      {p.body && <p className="mt-1 line-clamp-1 text-[11.5px] text-white/80">{p.body}</p>}
                      <span className="mt-2.5 w-fit rounded-lg bg-gold px-4 py-1.5 text-[12px] font-extrabold text-navy-deep">اطلب الآن</span>
                    </div>
                  </Link>
                )
              })}
            </div>
            {offers.length > 1 && (
              <div className="mt-2.5 flex justify-center gap-1.5">
                {offers.map((p, i) => <button key={p.id} aria-label={`عرض ${i + 1}`} onClick={() => setSlide(i)} className={`h-1.5 rounded-full transition-all ${i === slide ? 'w-5 bg-navy' : 'w-1.5 bg-[#c9d1e2]'}`} />)}
              </div>
            )}
          </section>
        ) : null}

        {/* الخدمات الرئيسية */}
        <section>
          <div className="mb-3 flex items-center justify-between">
            <h2 className="text-[17px] font-extrabold text-ink">الخدمات الرئيسية</h2>
            <Link to="/services" className="text-[13px] font-bold text-navy">عرض الكل</Link>
          </div>
          <div className="grid grid-cols-4 gap-2.5">
            {!services.data ? [0, 1, 2, 3].map((i) => <Skel key={i} className="h-24" />) : svc.slice(0, 4).map((s) => (
              <Link key={s.id} to={`/services/${s.id}`} className="card flex flex-col items-center gap-1.5 px-1 py-3 text-center transition active:scale-95">
                <ServiceArt name={s.name} badge className="h-12 w-12" />
                <span className="line-clamp-2 text-[11.5px] leading-tight font-bold text-ink">{s.name}</span>
              </Link>
            ))}
          </div>
        </section>

        {/* عروض خاصة */}
        {maxDiscount > 0 && (
          <Link to="/offers" className="relative flex h-[92px] items-center overflow-hidden rounded-2xl bg-gradient-to-l from-[#8a5a12] via-[#c08a2a] to-[#e9b04a] p-4 text-white shadow-[0_10px_22px_-12px_rgba(185,120,26,.9)]">
            <img src="/img/gift.webp" alt="" className="absolute inset-y-0 left-0 h-full w-[45%] object-cover" />
            <div className="absolute inset-0 bg-gradient-to-l from-[#8a5a12] from-45% via-[#a8741f]/80 via-60% to-transparent" />
            <div className="relative z-10">
              <p className="text-[18px] font-black">عروض خاصة</p>
              <p className="text-[15px] font-extrabold">خصومات تصل إلى {maxDiscount}%</p>
            </div>
          </Link>
        )}

        {/* آخر حجز */}
        <section>
          <h2 className="mb-3 text-[17px] font-extrabold text-ink">آخر حجز لك</h2>
          {!bookings.data ? <Skel className="h-24" /> : last ? (
            <Link to={`/bookings/${last.id}`} className="card flex items-center gap-3 p-3">
              <ServiceArt id={lastSvc?.id} name={lastSvc?.name ?? last.services} hasImage={lastSvc?.hasImage} className="h-16 w-16 shrink-0 rounded-xl" />
              <div className="min-w-0 flex-1">
                <p className="truncate text-[15px] font-extrabold text-ink">{last.services || 'خدمة'}</p>
                <p className="mt-0.5 text-[12px] text-mute">{fmtDate(last.scheduledAt ?? last.createdAt)} · <span dir="ltr">#{last.code}</span></p>
              </div>
              <StatusPill status={last.status} />
            </Link>
          ) : (
            <Link to="/request" className="card flex items-center justify-between p-4">
              <span className="text-[14px] font-bold text-mute">ما عندك حجز بعد</span>
              <span className="rounded-lg bg-navy-2 px-3 py-1.5 text-[12px] font-extrabold text-white">احجز خدمة</span>
            </Link>
          )}
        </section>
      </div>
    </div>
  )
}
