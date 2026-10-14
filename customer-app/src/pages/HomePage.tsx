import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, imgUrl, type Account, type Home, type Promo } from '../api'
import Logo from '../components/Logo'
import MatrixEye from '../components/MatrixEye'

const greet = () => {
  const h = new Date().getHours()
  return h < 12 ? 'صباح الخير' : h < 18 ? 'نهارك سعيد' : 'مساء الخير'
}

// ألوان البطاقات حسب الصنف — بلا صورة تطلع خلفية متدرجة
const TONES = ['from-navy to-[#2563eb]', 'from-[#9a3412] to-sun', 'from-[#065f46] to-[#10b981]', 'from-[#4c1d95] to-[#8b5cf6]', 'from-[#0f172a] to-[#334155]']
const ICONS: [RegExp, string][] = [[/gps|جي بي|تتبع/i, '📍'], [/كامير|camera|ip/i, '📹'], [/شمس|طاقة|solar/i, '☀️'], [/شبك|انترنت|network/i, '🌐'],
  [/صوت|sound/i, '🔊'], [/ديكور|decor/i, '🛋️'], [/كهرب/i, '⚡'], [/حريق|انذار/i, '🚨']]
const iconFor = (n: string) => ICONS.find(([re]) => re.test(n))?.[1] ?? '🛠️'

export default function HomePage({ me, onBadge }: { me: Account; onBadge: (n: number) => void }) {
  const nav = useNavigate()
  const [data, setData] = useState<Home | null>(null)
  const [err, setErr] = useState('')
  const [slide, setSlide] = useState(0)
  const [allSvc, setAllSvc] = useState(false)

  useEffect(() => {
    api.home().then((d) => { setData(d); onBadge(d.newCount) }).catch((e) => setErr(e.message))
  }, [onBadge])

  const offers = (data?.items ?? []).filter((p) => p.kind !== 'work')
  const works = (data?.items ?? []).filter((p) => p.kind === 'work')

  // البانر يتبدّل لحاله كل 5 ثواني — «كل شوي يعرض شي جديد»
  useEffect(() => {
    if (offers.length < 2) return
    const t = setInterval(() => setSlide((s) => (s + 1) % offers.length), 5000)
    return () => clearInterval(t)
  }, [offers.length])

  // شاف العروض الجديدة → ينصفّر الرقم بعد ثانيتين على الشاشة
  useEffect(() => {
    if (!data?.newCount) return
    const t = setTimeout(() => { api.seen().then(() => onBadge(0)).catch(() => {}) }, 2500)
    return () => clearTimeout(t)
  }, [data?.newCount, onBadge])

  const ask = (text: string) => nav('/matrix', { state: { prefill: text } })

  return (
    <div className="rise">
      <header className="rounded-b-[2rem] bg-gradient-to-br from-navy to-[#1d4ed8] px-5 pt-8 pb-16 text-white">
        <div className="flex items-center gap-3">
          <Logo size={42} />
          <div className="min-w-0">
            <p className="text-xs text-blue-100">{greet()}</p>
            <p className="truncate text-lg font-black">{me.name.split(' ')[0]} 👋</p>
          </div>
          {data && data.completedJobs > 0 && (
            <span className="mr-auto rounded-2xl bg-white/10 px-3 py-1.5 text-center text-[11px] leading-tight">
              <b className="block text-base">{data.completedJobs.toLocaleString('ar-IQ')}+</b>عمل منجز
            </span>
          )}
        </div>
      </header>

      <div className="-mt-12 space-y-6 px-4">
        {/* ماتركس */}
        <button onClick={() => nav('/matrix')} className="flex w-full items-center gap-3 rounded-3xl bg-white p-4 text-right shadow-lg shadow-navy/10 transition active:scale-[.99]">
          <MatrixEye size={48} />
          <div className="min-w-0 flex-1">
            <p className="font-black text-ink">اسأل ماتركس أو احجز</p>
            <p className="truncate text-xs text-slate-500">مساعدنا الذكي يجاوبك ويسجّل حجزك بدقيقة</p>
          </div>
          <span className="rounded-2xl bg-sun px-3 py-2 text-xs font-black text-white">ابدأ</span>
        </button>

        {err && <p className="rounded-2xl bg-rose-50 p-3 text-sm text-rose-700">{err}</p>}
        {!data && !err && <div className="h-44 animate-pulse rounded-3xl bg-slate-200" />}

        {/* العروض */}
        {offers.length > 0 && (
          <section>
            <SectionTitle title="عروضنا الحالية" />
            <div className="relative h-48 overflow-hidden rounded-3xl shadow-md">
              {offers.map((p, i) => <OfferSlide key={p.id} p={p} i={i} active={i === slide} onAsk={() => ask(`مهتم بالعرض: ${p.title}`)} />)}
              {offers.length > 1 && (
                <div className="absolute bottom-3 left-1/2 flex -translate-x-1/2 gap-1.5">
                  {offers.map((p, i) => <button key={p.id} aria-label={`عرض ${i + 1}`} onClick={() => setSlide(i)} className={`h-1.5 rounded-full transition-all ${i === slide ? 'w-5 bg-white' : 'w-1.5 bg-white/50'}`} />)}
                </div>
              )}
            </div>
          </section>
        )}

        {/* الخدمات */}
        {data && data.services.length > 0 && (
          <section>
            <SectionTitle title="خدماتنا" sub="اضغط على أي خدمة وماتركس يكمل وياك" />
            <div className="grid grid-cols-3 gap-2.5">
              {(allSvc ? data.services : data.services.slice(0, 9)).map((s) => (
                <button key={s.id} onClick={() => ask(`أريد أستفسر عن خدمة ${s.name}`)}
                  className="flex flex-col items-center gap-1.5 rounded-2xl bg-white px-2 py-3.5 text-center shadow-sm transition active:scale-95">
                  <span className="grid h-11 w-11 place-items-center rounded-2xl bg-navy/5 text-2xl">{iconFor(s.name)}</span>
                  <span className="line-clamp-2 text-[12px] leading-tight font-bold text-slate-700">{s.name}</span>
                </button>
              ))}
            </div>
            {data.services.length > 9 && (
              <button onClick={() => setAllSvc((v) => !v)} className="mt-2.5 w-full rounded-2xl bg-white py-2.5 text-sm font-black text-navy shadow-sm">
                {allSvc ? 'عرض أقل' : `عرض كل الخدمات (${data.services.length})`}
              </button>
            )}
          </section>
        )}

        {/* من أعمالنا */}
        {works.length > 0 && (
          <section>
            <SectionTitle title="من أعمالنا" />
            <div className="no-scrollbar -mx-4 flex snap-x gap-3 overflow-x-auto px-4 pb-1">
              {works.map((w) => (
                <article key={w.id} className="w-64 shrink-0 snap-start overflow-hidden rounded-3xl bg-white shadow-sm">
                  {w.hasImage ? <img src={imgUrl(w.id)} alt={w.title} loading="lazy" className="h-36 w-full object-cover" />
                    : <div className="grid h-36 place-items-center bg-gradient-to-br from-slate-100 to-slate-200 text-4xl">🏗️</div>}
                  <div className="p-3">
                    <p className="font-black text-ink">{w.title}</p>
                    {w.body && <p className="mt-0.5 line-clamp-2 text-xs text-slate-500">{w.body}</p>}
                  </div>
                </article>
              ))}
            </div>
          </section>
        )}

        <section className="rounded-3xl bg-gradient-to-l from-amber-50 to-orange-50 p-5">
          <p className="font-black text-ink">ليش الأماني؟ ✨</p>
          <ul className="mt-2 space-y-1.5 text-sm text-slate-600">
            <li>✅ فرق فنية متخصصة وتنفيذ مضمون</li>
            <li>✅ متابعة بعد التنفيذ وخدمة ما بعد البيع</li>
            <li>✅ حجز سهل ومتابعة حالة طلبك من التطبيق</li>
          </ul>
        </section>
      </div>
    </div>
  )
}

function SectionTitle({ title, sub }: { title: string; sub?: string }) {
  return (
    <div className="mb-2.5 px-1">
      <h2 className="text-lg font-black text-ink">{title}</h2>
      {sub && <p className="text-xs text-slate-500">{sub}</p>}
    </div>
  )
}

function OfferSlide({ p, i, active, onAsk }: { p: Promo; i: number; active: boolean; onAsk: () => void }) {
  return (
    <div className={`absolute inset-0 transition-opacity duration-700 ${active ? 'opacity-100' : 'pointer-events-none opacity-0'}`}>
      {p.hasImage ? <img src={imgUrl(p.id)} alt="" className="absolute inset-0 h-full w-full object-cover" />
        : <div className={`absolute inset-0 bg-gradient-to-br ${TONES[i % TONES.length]}`} />}
      <div className="absolute inset-0 bg-gradient-to-t from-black/75 via-black/25 to-transparent" />
      <div className="absolute inset-x-0 bottom-0 p-4 pb-7 text-white">
        {p.isNew && <span className="mb-1.5 inline-block rounded-full bg-sun px-2 py-0.5 text-[10px] font-black">جديد</span>}
        <p className="text-lg leading-snug font-black">{p.title}</p>
        {p.body && <p className="mt-0.5 line-clamp-2 text-xs text-white/85">{p.body}</p>}
        <button onClick={onAsk} className="mt-2 rounded-xl bg-white px-3 py-1.5 text-xs font-black text-navy">أريد أعرف أكثر ←</button>
      </div>
    </div>
  )
}
