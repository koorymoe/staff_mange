import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api, type SCard } from '../api'
import ServiceArt from '../components/ServiceArt'
import { Skel, Stars } from '../components/ui'
import { iqd } from '../lib'

// ═══ تفاصيل الخدمة (شاشة 6) ═══
export default function ServiceDetailPage() {
  const { id = '' } = useParams()
  const nav = useNavigate()
  const [s, setS] = useState<SCard | null>(null)
  const [err, setErr] = useState('')
  useEffect(() => { api.service(id).then(setS).catch((e) => setErr(e.message)) }, [id])

  const toggleFav = async () => {
    if (!s) return
    const on = !s.favorite
    setS({ ...s, favorite: on })
    await api.fav(s.id, on).catch(() => setS({ ...s, favorite: !on }))
  }
  // المميزات يكتبهن المالك — بلا مميزات مخترعة
  const features = s?.features ?? []

  return (
    <div className="page-in min-h-full bg-white pb-28">
      <div className="relative">
        {s ? <ServiceArt id={s.id} name={s.name} hasImage={s.hasImage} className="h-72 w-full" /> : <Skel className="h-72 rounded-none" />}
        <div className="safe-t absolute inset-x-0 top-0 flex justify-between p-4">
          <button aria-label="رجوع" onClick={() => nav(-1)} className="grid h-10 w-10 place-items-center rounded-full bg-white/90 text-ink shadow-md backdrop-blur">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round"><path d="M9 6l6 6-6 6" /></svg>
          </button>
          {s && (
            <button aria-label="المفضلة" onClick={toggleFav} className="grid h-10 w-10 place-items-center rounded-full bg-white/90 shadow-md backdrop-blur transition active:scale-90">
              <svg width="20" height="20" viewBox="0 0 24 24" fill={s.favorite ? '#e11d48' : 'none'} stroke={s.favorite ? '#e11d48' : '#0f1d3a'} strokeWidth="2"><path d="M12 21s-7-4.4-9.3-9A5.4 5.4 0 0 1 12 6a5.4 5.4 0 0 1 9.3 6c-2.3 4.6-9.3 9-9.3 9z" /></svg>
            </button>
          )}
        </div>
      </div>

      <div className="relative -mt-6 rounded-t-[26px] bg-white px-5 pt-6">
        {err && <p className="rounded-xl bg-rose-50 p-3 text-sm text-rose-600">{err}</p>}
        {!s ? <div className="space-y-3"><Skel className="h-8 w-2/3" /><Skel className="h-16" /><Skel className="h-28" /></div> : (
          <div className="rise">
            <h1 className="text-[24px] font-black text-navy">{s.name}</h1>
            {s.rating != null && s.ratingCount > 0 && (
              <p className="mt-1 flex items-center gap-1.5 text-[13px]"><Stars value={s.rating} /> <b className="text-ink">{s.rating.toFixed(1)}</b> <span className="text-mute">({s.ratingCount} تقييم)</span></p>
            )}
            <p className="mt-3 text-[14px] leading-7 text-[#4b5672]">
              {s.description || `نقدم خدمة ${s.name} بأعلى جودة، بفريق فني متخصص وتنفيذ احترافي ومتابعة بعد التنفيذ.`}
            </p>
            {features.length > 0 && <div className="mt-4 grid grid-cols-2 gap-2.5">
              {features.slice(0, 4).map((f, i) => (
                <div key={i} className="flex items-center gap-2.5 rounded-xl border border-line bg-[#fafbfe] px-3 py-3">
                  <span className="grid h-9 w-9 shrink-0 place-items-center rounded-lg bg-navy/8 text-[18px]">{f.icon || '✔️'}</span>
                  <span className="text-[13px] leading-tight font-bold text-ink">{f.text}</span>
                </div>
              ))}
            </div>}
            {s.priceFrom != null && (
              <div className="mt-5">
                <p className="text-[13px] text-mute">تبدأ الأسعار من</p>
                <p className="text-[28px] leading-tight font-black text-navy">{iqd(s.priceFrom)}</p>
              </div>
            )}
            <Link to="/matrix" state={{ prefill: `عندي سؤال عن خدمة ${s.name}` }} className="mt-4 flex items-center gap-2 text-[13px] font-bold text-navy">
              💬 عندك سؤال؟ اسأل ماتركس
            </Link>
          </div>
        )}
      </div>

      {s && (
        <div className="safe-b fixed inset-x-0 bottom-0 z-30 border-t border-line bg-white px-5 pt-3 pb-3">
          <Link to={`/request?service=${s.id}`} className="mx-auto flex h-[52px] max-w-lg items-center justify-center rounded-xl bg-navy-2 text-[16px] font-extrabold text-white shadow-[0_10px_20px_-10px_rgba(13,59,140,.7)] active:scale-[.98]">
            اطلب الخدمة
          </Link>
        </div>
      )}
    </div>
  )
}
