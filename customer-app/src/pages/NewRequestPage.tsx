import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { api, type Account } from '../api'
import { clearCache, useServices } from '../data'
import ServiceArt from '../components/ServiceArt'
import { TopBar } from '../components/ui'
import { CITIES } from '../lib'

const STEPS = ['بيانات الخدمة', 'الموعد والموقع', 'التفاصيل', 'مراجعة']

// ═══ طلب خدمة جديدة (شاشة 7) — 4 خطوات بمؤشر ═══
// الطلب يوصل للحجوزات باسم ماتركس (هو الي يستلم طلبات التطبيق).
export default function NewRequestPage({ me }: { me: Account }) {
  const nav = useNavigate()
  const [params] = useSearchParams()
  const { data: services } = useServices()
  const [step, setStep] = useState(0)
  const [f, setF] = useState({ serviceId: params.get('service') ?? '', city: me.city ?? '', date: '', time: '', address: '', notes: params.get('offer') ? `بخصوص العرض: ${params.get('offer')}` : '' })
  const [photos, setPhotos] = useState<File[]>([])
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const [done, setDone] = useState<{ id: string; code: string } | null>(null)
  const [today] = useState(() => new Date().toLocaleDateString('en-CA'))

  const previews = useMemo(() => photos.map((p) => URL.createObjectURL(p)), [photos])
  useEffect(() => () => previews.forEach((u) => URL.revokeObjectURL(u)), [previews])
  const svc = services?.find((s) => s.id === f.serviceId)
  const set = (k: keyof typeof f) => (e: { target: { value: string } }) => setF((x) => ({ ...x, [k]: e.target.value }))

  const can = [!!f.serviceId, !!f.city && !!f.date, f.address.trim().length >= 4, true][step]
  const next = async () => {
    setErr('')
    if (step < 3) { setStep(step + 1); return }
    setBusy(true)
    try {
      const fd = new FormData()
      Object.entries(f).forEach(([k, v]) => fd.append(k, v))
      photos.forEach((p) => fd.append('photos', p))
      setDone(await api.createBooking(fd)); clearCache()
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر الإرسال') } finally { setBusy(false) }
  }

  if (done) {
    return (
      <div className="page-in flex min-h-full flex-col items-center justify-center bg-white px-6 text-center">
        <div className="grid h-24 w-24 place-items-center rounded-full bg-emerald-50"><span className="text-5xl">✅</span></div>
        <h1 className="mt-5 text-[22px] font-black text-ink">تم إرسال طلبك بنجاح</h1>
        <p className="mt-2 text-[14px] leading-7 text-mute">رقم الطلب <b className="text-navy" dir="ltr">#{done.code}</b><br />ماتركس استلم طلبك، وفريقنا راح يتواصل وياك لتثبيت الموعد.</p>
        <button onClick={() => nav(`/bookings/${done.id}`, { replace: true })} className="mt-7 h-[52px] w-full max-w-sm rounded-xl bg-navy-2 font-extrabold text-white">متابعة الطلب</button>
        <button onClick={() => nav('/', { replace: true })} className="mt-3 text-[14px] font-bold text-mute">الرجوع للرئيسية</button>
      </div>
    )
  }

  const field = 'flex h-[50px] w-full items-center rounded-xl border border-line bg-white px-4 text-[14px] text-ink outline-none focus:border-navy focus:ring-4 focus:ring-navy/10'
  return (
    <div className="page-in min-h-full bg-white pb-28">
      <TopBar title="طلب خدمة جديدة" />
      {/* مؤشر الخطوات */}
      <ol className="flex items-start px-5 pt-2 pb-5">
        {STEPS.map((t, i) => (
          <li key={t} className="flex flex-1 flex-col items-center">
            <div className="flex w-full items-center">
              <span className={`h-0.5 flex-1 ${i === 0 ? 'opacity-0' : i <= step ? 'bg-navy' : 'bg-line'}`} />
              <span className={`grid h-8 w-8 place-items-center rounded-full text-[13px] font-black transition ${i < step ? 'bg-navy text-white' : i === step ? 'bg-navy text-white ring-4 ring-navy/15' : 'bg-[#eef1f7] text-[#9aa5bd]'}`}>{i < step ? '✓' : i + 1}</span>
              <span className={`h-0.5 flex-1 ${i === STEPS.length - 1 ? 'opacity-0' : i < step ? 'bg-navy' : 'bg-line'}`} />
            </div>
            <span className={`mt-1.5 text-[11px] font-bold ${i <= step ? 'text-navy' : 'text-[#9aa5bd]'}`}>{t}</span>
          </li>
        ))}
      </ol>

      <div key={step} className="page-in space-y-4 px-5">
        {step === 0 && (
          <>
            <p className="text-[14px] font-extrabold text-ink">نوع الخدمة</p>
            <div className="grid grid-cols-2 gap-2.5">
              {(services ?? []).map((s) => (
                <button key={s.id} type="button" onClick={() => setF((x) => ({ ...x, serviceId: s.id }))}
                  className={`flex items-center gap-2 rounded-xl border-2 p-2 text-right transition ${f.serviceId === s.id ? 'border-navy bg-navy/5' : 'border-line bg-white'}`}>
                  <ServiceArt id={s.id} name={s.name} hasImage={s.hasImage} className="h-10 w-10 shrink-0 rounded-lg" />
                  <span className="line-clamp-2 text-[12.5px] leading-tight font-bold text-ink">{s.name}</span>
                </button>
              ))}
            </div>
          </>
        )}
        {step === 1 && (
          <>
            <label className="block"><span className="mb-1.5 block text-[14px] font-extrabold text-ink">المحافظة</span>
              <select value={f.city} onChange={set('city')} className={field}><option value="">اختر المحافظة</option>{CITIES.map((c) => <option key={c}>{c}</option>)}</select></label>
            <label className="block"><span className="mb-1.5 block text-[14px] font-extrabold text-ink">التاريخ المفضّل</span>
              <input type="date" min={today} value={f.date} onChange={set('date')} className={field} /></label>
            <label className="block"><span className="mb-1.5 block text-[14px] font-extrabold text-ink">الوقت المفضّل <span className="font-normal text-mute">(اختياري)</span></span>
              <select value={f.time} onChange={set('time')} className={field}><option value="">أي وقت</option><option>صباحاً (9 - 12)</option><option>ظهراً (12 - 3)</option><option>عصراً (3 - 6)</option></select></label>
            <p className="rounded-xl bg-blue-50/70 p-3 text-[12px] leading-6 text-navy">📌 هذا موعد مفضّل — فريقنا يتواصل وياك لتثبيت الموعد النهائي.</p>
          </>
        )}
        {step === 2 && (
          <>
            <label className="block"><span className="mb-1.5 block text-[14px] font-extrabold text-ink">العنوان التفصيلي</span>
              <textarea rows={3} value={f.address} onChange={set('address')} placeholder="المنطقة، الشارع، أقرب نقطة دالة…" className={`${field} h-auto py-3`} /></label>
            <label className="block"><span className="mb-1.5 block text-[14px] font-extrabold text-ink">ملاحظات إضافية</span>
              <textarea rows={3} value={f.notes} onChange={set('notes')} placeholder="أي تفاصيل أخرى؟" className={`${field} h-auto py-3`} /></label>
            <div>
              <span className="mb-1.5 block text-[14px] font-extrabold text-ink">إضافة صور <span className="font-normal text-mute">(اختياري)</span></span>
              <div className="flex gap-2">
                {photos.length < 4 && (
                  <label className="grid h-[72px] w-[72px] shrink-0 cursor-pointer place-items-center rounded-xl border-2 border-dashed border-navy/40 text-3xl text-navy">+
                    <input type="file" accept="image/*" multiple className="hidden" onChange={(e) => { setPhotos((p) => [...p, ...Array.from(e.target.files ?? [])].slice(0, 4)); e.target.value = '' }} />
                  </label>
                )}
                {previews.map((u, i) => (
                  <div key={u} className="relative h-[72px] w-[72px] shrink-0">
                    <img src={u} alt="" className="h-full w-full rounded-xl object-cover" />
                    <button aria-label="حذف الصورة" onClick={() => setPhotos((p) => p.filter((_, j) => j !== i))} className="absolute -top-1.5 -left-1.5 grid h-6 w-6 place-items-center rounded-full bg-rose-500 text-[12px] text-white shadow">×</button>
                  </div>
                ))}
              </div>
            </div>
          </>
        )}
        {step === 3 && (
          <div className="card divide-y divide-line">
            {[['الخدمة', svc?.name ?? '—'], ['المحافظة', f.city], ['الموعد المفضّل', [f.date, f.time].filter(Boolean).join(' · ')], ['العنوان', f.address], ['ملاحظات', f.notes || '—'], ['الصور', photos.length ? `${photos.length} صورة` : '—']].map(([k, v]) => (
              <div key={k} className="flex gap-3 px-4 py-3 text-[13.5px]"><span className="w-24 shrink-0 text-mute">{k}</span><span className="font-bold text-ink">{v}</span></div>
            ))}
          </div>
        )}
        {err && <p className="rounded-xl bg-rose-50 p-3 text-[13px] font-bold text-rose-600">{err}</p>}
      </div>

      <div className="safe-b fixed inset-x-0 bottom-0 z-30 border-t border-line bg-white px-5 pt-3 pb-3">
        <div className="mx-auto flex max-w-lg gap-2">
          {step > 0 && <button onClick={() => setStep(step - 1)} className="h-[52px] rounded-xl border border-line px-5 font-bold text-ink">السابق</button>}
          <button disabled={!can || busy} onClick={next} className="h-[52px] flex-1 rounded-xl bg-navy-2 text-[16px] font-extrabold text-white shadow-[0_10px_20px_-10px_rgba(13,59,140,.7)] disabled:opacity-40">
            {busy ? '…' : step < 3 ? 'التالي' : 'إرسال الطلب'}
          </button>
        </div>
      </div>
    </div>
  )
}
