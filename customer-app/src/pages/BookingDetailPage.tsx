import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { api, techPhotoBlob, type BookingDetail } from '../api'
import { useServices } from '../data'
import ServiceArt from '../components/ServiceArt'
import { Skel, Stars, StatusPill, TopBar } from '../components/ui'
import { fmtDate, fmtTime, iqd } from '../lib'

const STAGES = ['تم الطلب', 'تمت الجدولة', 'قيد التنفيذ', 'تم الانتهاء']
const stageOf = (s: string) => (s === 'COMPLETED' ? 3 : s === 'IN_PROGRESS' || s === 'PARTIAL' ? 2 : s === 'CONFIRMED' || s === 'WAITING' ? 1 : 0)

// ═══ تفاصيل الحجز (شاشة 9) ═══
export default function BookingDetailPage() {
  const { id = '' } = useParams()
  const nav = useNavigate()
  const { data: services } = useServices()
  const [d, setD] = useState<BookingDetail | null>(null)
  const [err, setErr] = useState('')
  const [photo, setPhoto] = useState<string | null>(null)
  const [rating, setRating] = useState(false)

  useEffect(() => { api.booking(id).then(setD).catch((e) => setErr(e.message)) }, [id])
  useEffect(() => {
    if (!d?.tech?.hasPhoto) return
    let url: string | null = null
    techPhotoBlob(id).then((u) => { url = u; setPhoto(u) })
    return () => { if (url) URL.revokeObjectURL(url) }
  }, [d?.tech?.hasPhoto, id])

  if (err) return <div><TopBar title="تفاصيل الحجز" /><p className="m-4 rounded-xl bg-rose-50 p-3 text-sm text-rose-600">{err}</p></div>
  if (!d) return <div><TopBar title="تفاصيل الحجز" /><div className="space-y-3 p-4"><Skel className="h-28" /><Skel className="h-20" /><Skel className="h-32" /></div></div>

  const b = d.booking
  const s = b.serviceId ? services?.find((x) => x.id === b.serviceId) : undefined
  const stage = stageOf(b.status)
  const cancelled = b.status === 'CANCELLED'
  const paid = (b.amountCollected ?? 0) > 0
  const amount = paid ? b.amountCollected : b.quotedPrice
  const ask = (text: string) => nav('/matrix', { state: { prefill: text } })

  return (
    <div className="page-in pb-28">
      <TopBar title="تفاصيل الحجز" />
      <div className="space-y-3 px-4 pt-1">
        <div className="flex items-center justify-between">
          <p className="text-[16px] font-black text-ink" dir="ltr">#{b.code}</p>
          <StatusPill status={b.status} label={d.statusLabel} />
        </div>

        <div className="card flex items-center gap-3 p-3">
          <ServiceArt id={s?.id} name={s?.name ?? d.services} hasImage={s?.hasImage} className="h-[76px] w-[76px] shrink-0 rounded-xl" />
          <div className="min-w-0">
            <p className="text-[17px] font-extrabold text-navy">{d.services || 'خدمة'}</p>
            {d.description && <p className="mt-0.5 line-clamp-2 text-[12px] text-mute">{d.description}</p>}
            {b.fromApp && <p className="mt-1 text-[11px] font-bold text-violet-700">👁️ استلمه ماتركس من التطبيق</p>}
          </div>
        </div>

        <div className="grid grid-cols-3 gap-2">
          {[
            ['📅', 'التاريخ', b.scheduledAt ? fmtDate(b.scheduledAt) : 'يتحدد قريباً'],
            ['🕐', 'الوقت', b.scheduledAt ? fmtTime(b.scheduledAt) : '—'],
            ['📍', 'الموقع', b.address ? b.address.split('—')[0].trim() : '—'],
          ].map(([i, k, v]) => (
            <div key={k} className="card px-2 py-3 text-center">
              <p className="text-[11.5px] text-mute">{i} {k}</p>
              <p className="mt-1 truncate text-[13.5px] font-extrabold text-ink" dir={k === 'الموقع' ? 'rtl' : 'ltr'}>{v}</p>
            </div>
          ))}
        </div>

        {d.tech ? (
          <div className="card p-3">
            <p className="mb-2 text-[14px] font-extrabold text-ink">الفني المسؤول</p>
            <div className="flex items-center gap-3">
              {photo ? <img src={photo} alt="" className="h-[68px] w-[68px] shrink-0 rounded-xl object-cover" />
                : <span className="grid h-[68px] w-[68px] shrink-0 place-items-center rounded-xl bg-navy/8 text-2xl font-black text-navy">{d.tech.name.trim()[0]}</span>}
              <div className="min-w-0 flex-1">
                <p className="truncate text-[15px] font-extrabold text-ink">{d.tech.name}</p>
                {d.tech.rating != null && d.tech.ratingCount > 0
                  ? <p className="mt-0.5 flex items-center gap-1 text-[12px]"><Stars value={d.tech.rating} size={12} /> <b>{d.tech.rating.toFixed(1)}</b> <span className="text-mute">({d.tech.ratingCount} تقييم)</span></p>
                  : <p className="mt-0.5 text-[12px] text-mute">من فريق الأماني الفني</p>}
              </div>
              <button aria-label="مراسلة" onClick={() => ask(`بخصوص حجزي ${b.code}: `)} className="grid h-10 w-10 place-items-center rounded-full bg-blue-50 text-navy">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21 12a8 8 0 0 1-11.6 7.1L4 20l1-4.6A8 8 0 1 1 21 12z" /></svg>
              </button>
            </div>
          </div>
        ) : !cancelled && (
          <div className="card flex items-center gap-3 p-4 text-[13px] text-mute">👷 الفني يتحدد بعد تثبيت الموعد</div>
        )}

        {!cancelled && (
          <div className="card p-4">
            <p className="mb-3 text-[14px] font-extrabold text-ink">مراحل التنفيذ</p>
            <ol className="flex items-start">
              {STAGES.map((t, i) => (
                <li key={t} className="flex flex-1 flex-col items-center">
                  <div className="flex w-full items-center">
                    <span className={`h-0.5 flex-1 ${i === 0 ? 'opacity-0' : i <= stage ? 'bg-navy' : 'bg-line'}`} />
                    <span className={`grid h-6 w-6 place-items-center rounded-full text-[11px] font-black ${i <= stage ? 'bg-navy text-white' : 'border-2 border-line bg-white text-transparent'}`}>✓</span>
                    <span className={`h-0.5 flex-1 ${i === STAGES.length - 1 ? 'opacity-0' : i < stage ? 'bg-navy' : 'bg-line'}`} />
                  </div>
                  <span className={`mt-1.5 text-center text-[10.5px] leading-tight ${i <= stage ? 'font-bold text-navy' : 'text-[#9aa5bd]'}`}>{t}</span>
                </li>
              ))}
            </ol>
          </div>
        )}

        {amount != null && amount > 0 && (
          <div className="card flex items-center justify-between p-4">
            <div><p className="text-[13px] text-mute">{paid ? 'المبلغ المدفوع' : 'المبلغ الإجمالي'}</p><p className="text-[19px] font-black text-navy">{iqd(amount)}</p></div>
            <span className={`rounded-lg px-3 py-1 text-[12px] font-bold ${paid ? 'bg-emerald-50 text-emerald-700' : 'bg-amber-50 text-amber-700'}`}>{paid ? 'مدفوع' : 'غير مدفوع'}</span>
          </div>
        )}

        {d.myRating && (
          <div className="card p-4"><p className="text-[13px] text-mute">تقييمك</p><div className="mt-1"><Stars value={d.myRating.stars} size={18} /></div>{d.myRating.comment && <p className="mt-1 text-[13px] text-ink">{d.myRating.comment}</p>}</div>
        )}
      </div>

      <div className="safe-b fixed inset-x-0 bottom-0 z-30 border-t border-line bg-white px-5 pt-3 pb-3">
        <div className="mx-auto max-w-lg">
          {b.status === 'COMPLETED' && !d.myRating ? (
            <button onClick={() => setRating(true)} className="flex h-[52px] w-full items-center justify-center gap-2 rounded-xl bg-navy-2 text-[16px] font-extrabold text-white">⭐ تقييم الخدمة</button>
          ) : (
            <button onClick={() => ask(`عندي استفسار عن حجزي ${b.code}`)} className="flex h-[52px] w-full items-center justify-center gap-2 rounded-xl bg-navy-2 text-[16px] font-extrabold text-white">💬 استفسار عن الحجز</button>
          )}
        </div>
      </div>
      {rating && <RateSheet id={b.id} onClose={() => setRating(false)} onDone={() => { setRating(false); api.booking(id).then(setD).catch(() => {}) }} />}
    </div>
  )
}

function RateSheet({ id, onClose, onDone }: { id: string; onClose: () => void; onDone: () => void }) {
  const [stars, setStars] = useState(0)
  const [comment, setComment] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const send = async () => {
    setBusy(true); setErr('')
    try { await api.rate(id, stars, comment); onDone() } catch (e) { setErr(e instanceof Error ? e.message : 'خطأ') } finally { setBusy(false) }
  }
  return (
    <div className="fixed inset-0 z-50 flex items-end bg-black/40" onClick={onClose}>
      <div className="safe-b rise w-full rounded-t-[26px] bg-white p-6" onClick={(e) => e.stopPropagation()}>
        <div className="mx-auto mb-4 h-1.5 w-12 rounded-full bg-line" />
        <p className="text-center text-[18px] font-black text-ink">شلون كانت الخدمة؟</p>
        <div className="mt-4 flex justify-center gap-2" dir="ltr">
          {[1, 2, 3, 4, 5].map((i) => (
            <button key={i} aria-label={`${i} نجوم`} onClick={() => setStars(i)} className="transition active:scale-90">
              <svg width="40" height="40" viewBox="0 0 24 24" fill={i <= stars ? '#f5b335' : '#e5e7eb'}><path d="M12 2l3.1 6.3 6.9 1-5 4.9 1.2 6.8L12 17.8 5.8 21l1.2-6.8-5-4.9 6.9-1z" /></svg>
            </button>
          ))}
        </div>
        <textarea rows={3} value={comment} onChange={(e) => setComment(e.target.value)} maxLength={500} placeholder="اكتب ملاحظتك (اختياري)"
          className="mt-4 w-full rounded-xl border border-line p-3 text-[14px] outline-none focus:border-navy" />
        {err && <p className="mt-2 text-[13px] text-rose-600">{err}</p>}
        <button disabled={!stars || busy} onClick={send} className="mt-3 h-[52px] w-full rounded-xl bg-navy-2 font-extrabold text-white disabled:opacity-40">إرسال التقييم</button>
      </div>
    </div>
  )
}
