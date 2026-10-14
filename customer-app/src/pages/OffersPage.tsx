import { useNavigate } from 'react-router-dom'
import { imgUrl } from '../api'
import { useHome, useServices } from '../data'
import ServiceArt from '../components/ServiceArt'
import { Empty, Skel, TopBar } from '../components/ui'

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
        {!data && [0, 1, 2].map((i) => <Skel key={i} className="h-36" />)}
        {data && offers.length === 0 && <Empty icon="🎁" title="ماكو عروض حالياً"><p className="mt-1 text-[13px] text-mute">تابعنا، العروض الجديدة توصلك هنا أول بأول.</p></Empty>}
        {offers.map((p, i) => {
          const s = p.serviceId ? services?.find((x) => x.id === p.serviceId) : undefined
          return (
            <article key={p.id} className="rise relative min-h-[150px] overflow-hidden rounded-2xl bg-navy-deep shadow-[0_10px_22px_-14px_rgba(13,59,140,.8)]" style={{ animationDelay: `${i * 50}ms` }}>
              {p.hasImage ? <img src={imgUrl(p.id)} alt="" className="absolute inset-y-0 left-0 h-full w-3/5 object-cover" />
                : <ServiceArt name={s?.name ?? p.title} className="absolute inset-y-0 left-0 h-full w-[55%]" />}
              <div className="absolute inset-0 bg-gradient-to-l from-navy-deep from-38% via-navy-deep/80 via-52% to-transparent to-75%" />
              <div className="relative mr-0 ml-auto flex min-h-[150px] w-[62%] flex-col justify-center p-4 text-white">
                {p.discountPercent ? <span className="mb-1.5 w-fit rounded-md bg-gold px-2.5 py-0.5 text-[12px] font-black text-navy-deep">{p.discountPercent}% خصم</span>
                  : p.isNew ? <span className="mb-1.5 w-fit rounded-md bg-gold px-2.5 py-0.5 text-[12px] font-black text-navy-deep">جديد</span> : null}
                <p className="line-clamp-2 text-[16px] leading-snug font-black">{p.title}</p>
                {p.body && <p className="mt-0.5 line-clamp-1 text-[11.5px] text-white/80">{p.body}</p>}
                <button onClick={() => order(p.serviceId, p.title)} className="mt-2.5 w-fit rounded-lg bg-white px-5 py-1.5 text-[12.5px] font-extrabold text-navy">اطلب العرض</button>
              </div>
            </article>
          )
        })}
      </div>
    </div>
  )
}
