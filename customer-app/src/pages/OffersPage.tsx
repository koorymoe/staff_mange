import { useNavigate } from 'react-router-dom'
import { imgUrl } from '../api'
import { useHome, useServices } from '../data'
import ServiceArt from '../components/ServiceArt'
import { Skel, TopBar } from '../components/ui'

const campaigns = [
  { title: 'كاميرات تحمي بيتك 24/7', note: 'تركيب احترافي ومتابعة مستمرة', label: 'اكتشف كاميرات الأماني', service: 'كاميرات المراقبة' },
  { title: 'خلي بيتك أذكى وأسهل', note: 'تحكم بالإضاءة والأجهزة من مكانك', label: 'شوف حلول البيت الذكي', service: 'المنازل الذكية' },
  { title: 'وفر باستهلاك الطاقة', note: 'حلول شمسية تناسب بيتك وعملك', label: 'اكتشف الطاقة الشمسية', service: 'الطاقة الشمسية' },
]

// ═══ العروض (شاشة 5) — بطاقات كبيرة بشارة خصم ذهبية ═══
export default function OffersPage() {
  const nav = useNavigate()
  const { data } = useHome()
  const { data: services } = useServices()
  const offers = (data?.items ?? []).filter((p) => p.kind !== 'work')

  const order = (serviceId: string | null, title: string) => {
    if (serviceId) nav(`/request?service=${serviceId}&offer=${encodeURIComponent(title)}`)
    else nav('/matrix', { state: { prefill: `مهتم بالعرض: ${title}` } })
  }

  return (
    <div className="page-in">
      <TopBar title="العروض" back={false} />
      <div className="space-y-3.5 px-4 pt-1 pb-4">
        <section className="relative overflow-hidden rounded-2xl bg-gradient-to-l from-[#082a66] via-[#123f9c] to-[#1c59b7] px-4 py-4 text-white shadow-[0_14px_28px_-18px_rgba(8,42,102,.9)]">
          <span className="absolute -left-6 -top-8 h-24 w-24 rounded-full bg-gold/20 blur-2xl" />
          <div className="relative flex items-center gap-3">
            <span className="grid h-11 w-11 place-items-center rounded-2xl bg-white/12 text-2xl">🎁</span>
            <div><p className="text-[16px] font-black">عروض تفيدك أكثر</p><p className="mt-0.5 text-[11.5px] text-white/75">كل عرض هنا واضح، وتكدر تطلبه بنفس اللحظة.</p></div>
          </div>
        </section>
        {!data && [0, 1, 2].map((i) => <Skel key={i} className="h-36" />)}
        {data && offers.length === 0 && campaigns.map((p, i) => (
          <article key={p.title} className="touch-card rise relative min-h-[164px] overflow-hidden rounded-[22px] bg-navy-deep shadow-[0_16px_28px_-17px_rgba(13,59,140,.9)]" style={{ animationDelay: `${i * 50}ms` }}>
            <ServiceArt name={p.service} className="absolute inset-y-0 left-0 h-full w-3/5" />
            <div className="absolute inset-0 bg-gradient-to-l from-navy-deep from-38% via-navy-deep/82 via-55% to-transparent" />
            <div className="relative mr-0 ml-auto flex min-h-[164px] w-[62%] flex-col justify-center p-4 text-white"><span className="mb-2 w-fit rounded-full bg-gold px-2.5 py-1 text-[10px] font-black text-navy-deep">شركة الأماني</span><p className="text-[16px] leading-snug font-black">{p.title}</p><p className="mt-1 text-[11px] text-white/80">{p.note}</p><button onClick={() => nav('/services')} className="mt-3 w-fit rounded-xl bg-white px-3 py-2 text-[11.5px] font-extrabold text-navy">{p.label} ←</button></div>
          </article>
        ))}
        {offers.map((p, i) => {
          const s = p.serviceId ? services?.find((x) => x.id === p.serviceId) : undefined
          return (
            <article key={p.id} className="touch-card rise relative min-h-[164px] overflow-hidden rounded-[22px] bg-navy-deep shadow-[0_16px_28px_-17px_rgba(13,59,140,.9)]" style={{ animationDelay: `${i * 50}ms` }}>
              {p.hasImage ? <img src={imgUrl(p.id)} alt="" className="absolute inset-y-0 left-0 h-full w-3/5 object-cover" />
                : <ServiceArt name={s?.name ?? p.title} className="absolute inset-y-0 left-0 h-full w-[55%]" />}
              <div className="absolute inset-0 bg-gradient-to-l from-navy-deep from-38% via-navy-deep/80 via-52% to-transparent to-75%" />
              <div className="relative mr-0 ml-auto flex min-h-[164px] w-[62%] flex-col justify-center p-4 text-white">
                {p.discountPercent ? <span className="mb-1.5 w-fit rounded-full bg-gold px-2.5 py-0.5 text-[12px] font-black text-navy-deep">وفر {p.discountPercent}%</span>
                  : p.isNew ? <span className="mb-1.5 w-fit rounded-md bg-gold px-2.5 py-0.5 text-[12px] font-black text-navy-deep">جديد</span> : null}
                <p className="line-clamp-2 text-[16px] leading-snug font-black">{p.title}</p>
                {p.body && <p className="mt-0.5 line-clamp-1 text-[11.5px] text-white/80">{p.body}</p>}
                <button onClick={() => order(p.serviceId, p.title)} className="mt-2.5 flex w-fit items-center gap-1.5 rounded-xl bg-white px-4 py-2 text-[12.5px] font-extrabold text-navy shadow-lg">اطلب العرض <span>←</span></button>
              </div>
            </article>
          )
        })}
      </div>
    </div>
  )
}
