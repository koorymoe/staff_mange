import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type Booking, type WorkPhoto } from '../api'
import MyDirectives from '../components/MyDirectives'

// ═══ 🪚 «شغلي اليوم» — كادر الديكور (دازاي 10-25) ═══
// أغلبهم ما مكملين مدارس: كلمات قليلة، أزرار كبيرة بالترتيب، وكل زر ينفتح بعد
// الي قبله. صورة «قبل» قبل ما يبدي، وصورة «بعد» قبل ما يكول خلصت.

const todayKey = () => new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })
const dayOf = (iso: string | null) => (iso ? new Date(iso).toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' }) : '')
const timeOf = (iso: string | null) => (iso ? new Date(iso).toLocaleTimeString('ar-IQ', { timeZone: 'Asia/Baghdad', hour: 'numeric', minute: '2-digit' }) : '')

export default function DecorTodayPage() {
  const [jobs, setJobs] = useState<Booking[] | null>(null)
  const [err, setErr] = useState('')
  const load = useCallback(() => {
    api.getBookings({ assignedTo: 'me' })
      .then((rows) => setJobs(rows.filter((b) => dayOf(b.scheduledAt) === todayKey() && b.status !== 'CANCELLED')
        .sort((a, b) => (a.scheduledAt ?? '').localeCompare(b.scheduledAt ?? ''))))
      .catch((e) => setErr(e instanceof Error ? e.message : 'تعذر'))
  }, [])
  useEffect(() => { load() }, [load])

  return (
    <div dir="rtl" className="mx-auto max-w-lg space-y-4 pb-24">
      <MyDirectives />
      <div className="rounded-3xl bg-gradient-to-l from-amber-600 to-orange-500 p-5 text-white shadow">
        <h1 className="text-3xl font-black">🪚 شغلي اليوم</h1>
        <p className="mt-1 text-lg">اضغط الأزرار بالترتيب، من فوك لجوّا.</p>
      </div>
      <Link to="/decor-guide" className="block rounded-2xl border-2 border-amber-300 bg-amber-50 p-4 text-center text-xl font-extrabold text-amber-900">
        📖 شنو مطلوب مني؟
      </Link>
      {err && <p className="rounded-2xl bg-red-50 p-4 text-center text-lg text-red-700">{err}</p>}
      {!jobs && !err && <p className="py-10 text-center text-xl text-slate-400">لحظة…</p>}
      {jobs && jobs.length === 0 && (
        <div className="rounded-3xl bg-white p-8 text-center shadow ring-1 ring-slate-200">
          <div className="text-6xl">☕</div>
          <p className="mt-3 text-2xl font-extrabold text-slate-700">ماكو شغل إلك اليوم</p>
          <p className="mt-1 text-lg text-slate-500">إذا المفروض عندك شغل، خابر مسؤولك.</p>
        </div>
      )}
      {jobs?.map((b) => <JobCard key={b.id} b={b} onChange={load} />)}
    </div>
  )
}

function JobCard({ b, onChange }: { b: Booking; onChange: () => void }) {
  const [photos, setPhotos] = useState<WorkPhoto[]>([])
  const [busy, setBusy] = useState('')
  const [err, setErr] = useState('')
  const [phase, setPhase] = useState<'BEFORE' | 'AFTER'>('BEFORE')
  const inputId = `wp-${b.id}`
  useEffect(() => { api.getWorkPhotos(b.id).then(setPhotos).catch(() => {}) }, [b.id])

  const hasBefore = photos.some((p) => p.phase === 'BEFORE')
  const hasAfter = photos.some((p) => p.phase === 'AFTER')
  const done = b.status === 'COMPLETED'
  const crew = b.assignments.map((a) => a.employee.name).join('، ')
  const map = b.locationUrl || (b.mapLatitude && b.mapLongitude ? `https://maps.google.com/?q=${b.mapLatitude},${b.mapLongitude}` : '')

  const run = async (label: string, fn: () => Promise<unknown>) => {
    setBusy(label); setErr('')
    try { await fn(); onChange() } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy('') }
  }
  const pickPhoto = (p: 'BEFORE' | 'AFTER') => { setPhase(p); document.getElementById(inputId)?.click() }
  const onFile = async (input: HTMLInputElement) => {
    const f = input.files?.[0]
    if (!f) return
    setBusy('photo'); setErr('')
    try {
      const p = await api.uploadWorkPhoto(b.id, phase, f)
      setPhotos((ps) => [...ps, p])
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally {
      setBusy('')
      input.value = ''
    }
  }

  // الخطوات بالترتيب — كل وحدة تنفتح بعد الي قبلها.
  const steps: { key: string; icon: string; label: string; doneNow: boolean; enabled: boolean; act: () => void }[] = [
    { key: 'depart', icon: '🚗', label: 'طلعت', doneNow: !!b.departedAt, enabled: true,
      act: () => void run('depart', () => api.departBooking(b.id)) },
    { key: 'arrive', icon: '📍', label: 'وصلت', doneNow: !!b.arrivedAt, enabled: !!b.departedAt,
      act: () => void run('arrive', () => api.markArrived(b.id)) },
    { key: 'before', icon: '📸', label: 'صوّر المكان قبل الشغل', doneNow: hasBefore, enabled: !!b.arrivedAt,
      act: () => pickPhoto('BEFORE') },
    { key: 'start', icon: '▶️', label: 'بديت', doneNow: !!b.startedAt, enabled: !!b.arrivedAt && hasBefore,
      act: () => void run('start', () => api.startBooking(b.id)) },
    { key: 'after', icon: '📸', label: 'صوّر الشغل بعد ما تخلص', doneNow: hasAfter, enabled: !!b.startedAt,
      act: () => pickPhoto('AFTER') },
    { key: 'finish', icon: '✅', label: 'خلصت', doneNow: done, enabled: !!b.startedAt && hasAfter,
      act: () => void run('finish', () => api.completeBooking(b.id, {})) },
  ]
  const next = steps.find((s) => !s.doneNow)

  return (
    <div className="overflow-hidden rounded-3xl bg-white shadow ring-1 ring-slate-200">
      <div className={`p-4 ${done ? 'bg-emerald-50' : 'bg-slate-50'}`}>
        <div className="flex items-center justify-between">
          <span className="text-2xl font-black text-[#0f2040]">⏰ {timeOf(b.scheduledAt) || '—'}</span>
          <span className="text-sm text-slate-400">{b.code}</span>
        </div>
        {b.address && <p className="mt-1 text-lg font-bold text-slate-700">🏠 {b.address}</p>}
        {map && <a href={map} target="_blank" rel="noreferrer" className="mt-2 inline-block rounded-xl bg-sky-600 px-4 py-2 text-lg font-bold text-white">🗺️ افتح الخريطة</a>}
        {crew && <p className="mt-2 text-base text-slate-600">👷 وياك: {crew}</p>}
        {b.crewNotes && <p className="mt-2 rounded-xl bg-amber-50 p-2 text-base text-amber-900">📝 {b.crewNotes}</p>}
      </div>
      {done ? (
        <p className="p-5 text-center text-2xl font-black text-emerald-700">✅ خلصت هالشغلة — عاشت ايدك</p>
      ) : (
        <div className="space-y-2 p-4">
          {steps.map((s) => {
            const isNext = next?.key === s.key
            return (
              <button key={s.key} type="button" disabled={!s.enabled || s.doneNow || !!busy} onClick={s.act}
                className={`flex w-full items-center gap-3 rounded-2xl px-4 py-4 text-right text-xl font-extrabold transition
                  ${s.doneNow ? 'bg-emerald-50 text-emerald-700' : isNext && s.enabled ? 'bg-[#0f2040] text-white shadow-lg' : 'bg-slate-100 text-slate-400'}`}>
                <span className="text-3xl">{s.doneNow ? '✔️' : s.icon}</span>
                <span className="flex-1">{s.label}</span>
                {busy && isNext && <span className="text-base">…</span>}
              </button>
            )
          })}
          {err && <p className="rounded-xl bg-red-50 p-3 text-center text-lg font-bold text-red-700">{err}</p>}
        </div>
      )}
      <input id={inputId} type="file" accept="image/*" capture="environment" className="hidden"
        onChange={(e) => void onFile(e.currentTarget)} />
    </div>
  )
}
