import { useState } from 'react'
import { api, type Booking } from '../api'
import EntityIdentity from './EntityIdentity'

// ═══ «حجوزات مرحّلة إلك» — قرار (ع) 10-07 ═══
// الإداري رحّل الحجز لأن المشكلة ما يعرفون الفنيين يحلّوها. هسه برقبتك:
// تتواصل ويا الزبون، تكتب الكشف، وتعالج بنفسك أو تطلب طاقم. ماتركس يحسب
// وقت كل خطوة عليك.

const fmt = (s?: string | null) => (s ? new Date(s).toLocaleString('en-GB', { timeZone: 'Asia/Baghdad', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }) : '')

function Step({ done, n, title, at }: { done: boolean; n: number; title: string; at?: string | null }) {
  return (
    <span className={`rounded-full px-2 py-0.5 text-[11px] font-bold ${done ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-100 text-slate-500'}`}>
      {done ? '✓' : n}. {title}{done && at ? ` · ${fmt(at)}` : ''}
    </span>
  )
}

function Card({ b, onChange }: { b: Booking; onChange: (b: Booking) => void }) {
  const [text, setText] = useState(b.techDiagnosis ?? '')
  const [fix, setFix] = useState('')
  const [phoneNote, setPhoneNote] = useState('')
  const [visitAt, setVisitAt] = useState('')
  const [mode, setMode] = useState<'' | 'PHONE' | 'VISIT' | 'MOVE'>('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const call = async (f: () => Promise<Booking>) => {
    setBusy(true); setErr('')
    try { onChange(await f()); setMode('') } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }
  const run = (step: 'contacted' | 'diagnosis' | 'crew' | 'resolve' | 'visited', t = '') => call(() => api.techStep(b.id, step, t))
  const phone = b.customer?.phone
  const visit = b.techDecision === 'VISIT'
  const visitPicker = (
    <div className="mt-2 flex flex-wrap items-center gap-2">
      <input type="datetime-local" value={visitAt} onChange={(e) => setVisitAt(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1.5 text-sm" />
      <button disabled={busy || !visitAt} onClick={() => void call(() => api.techDecide(b.id, 'VISIT', '', visitAt))}
        className="rounded-lg bg-sky-600 px-3 py-2 text-sm font-bold text-white disabled:opacity-50">📅 ثبّت موعد الزيارة</button>
    </div>
  )
  return (
    <div className="rounded-xl border-2 border-orange-200 bg-orange-50/40 p-4">
      <EntityIdentity booking={b} variant="full" className="mb-2" />
      <p className="text-sm"><b className="text-orange-900">🛠️ المشكلة حسب الإداري:</b> {b.handoverReason}</p>
      <p className="text-[11px] text-slate-500">رحّله {b.handoverBy?.name ?? '—'} · {fmt(b.handoverAt)}</p>
      <div className="mt-2 flex flex-wrap gap-1">
        <Step n={1} title="تواصلت ويا الزبون" done={!!b.techContactedAt} at={b.techContactedAt} />
        <Step n={2} title={b.techDecision === 'PHONE' ? 'انحلّت بالتلفون' : visit ? 'تحتاج زيارة' : 'القرار'} done={!!b.techDecidedAt} at={b.techDecidedAt} />
        {visit && <Step n={3} title={`الزيارة ${fmt(b.techVisitAt)}`} done={!!b.techVisitedAt} at={b.techVisitedAt} />}
        {b.techDecision !== 'PHONE' && <Step n={visit ? 4 : 3} title="الكشف" done={!!b.techDiagnosedAt} at={b.techDiagnosedAt} />}
        <Step n={visit ? 5 : 4} title="انحلّت" done={b.status === 'COMPLETED'} at={b.completedAt} />
      </div>
      {err && <p className="mt-2 text-xs text-red-600">{err}</p>}

      {/* ١. التواصل */}
      {!b.techContactedAt && (
        <div className="mt-3 flex flex-wrap gap-2">
          {phone && <a href={`tel:${phone}`} className="rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm font-bold">📞 اتصل {phone}</a>}
          <button disabled={busy} onClick={() => void run('contacted')} className="rounded-lg bg-orange-600 px-3 py-2 text-sm font-bold text-white disabled:opacity-50">✅ تواصلت ويا الزبون</button>
        </div>
      )}

      {/* ٢. القرار: بالتلفون لو زيارة */}
      {b.techContactedAt && !b.techDecidedAt && (
        <div className="mt-3 space-y-2">
          <p className="text-sm font-bold text-slate-700">بعد ما حچيت ويا الزبون، شنو صار؟</p>
          <div className="flex flex-wrap gap-2">
            <button onClick={() => setMode('PHONE')} className={`rounded-lg px-3 py-2 text-sm font-bold ${mode === 'PHONE' ? 'bg-emerald-600 text-white' : 'border border-emerald-300 bg-white text-emerald-800'}`}>✅ انحلّت بالتلفون</button>
            <button onClick={() => setMode('VISIT')} className={`rounded-lg px-3 py-2 text-sm font-bold ${mode === 'VISIT' ? 'bg-sky-600 text-white' : 'border border-sky-300 bg-white text-sky-800'}`}>🔍 تحتاج كشف / زيارة</button>
          </div>
          {mode === 'PHONE' && (
            <div>
              <textarea value={phoneNote} onChange={(e) => setPhoneNote(e.target.value)} rows={2} className="w-full rounded-lg border border-slate-300 p-2 text-sm" placeholder="شنو چانت المشكلة وشلون انحلّت؟" />
              <button disabled={busy || phoneNote.trim().length < 10} onClick={() => void call(() => api.techDecide(b.id, 'PHONE', phoneNote))}
                className="mt-1 rounded-lg bg-emerald-600 px-3 py-2 text-sm font-bold text-white disabled:opacity-50">سكّر الحجز — انحلّت</button>
            </div>
          )}
          {mode === 'VISIT' && visitPicker}
        </div>
      )}

      {/* ٣. الزيارة */}
      {visit && !b.techVisitedAt && b.status !== 'COMPLETED' && (
        <div className="mt-3 space-y-2 rounded-lg bg-sky-50 p-3">
          <p className="text-sm"><b>📅 موعد الزيارة:</b> {fmt(b.techVisitAt)}{(b.techVisitMoves ?? 0) > 0 && <span className="text-xs text-amber-700"> · انتغيّر {b.techVisitMoves} مرة</span>}</p>
          <div className="flex flex-wrap gap-2">
            <button disabled={busy} onClick={() => void run('visited')} className="rounded-lg bg-sky-700 px-3 py-2 text-sm font-bold text-white disabled:opacity-50">📍 وصلت للزبون</button>
            <button onClick={() => setMode(mode === 'MOVE' ? '' : 'MOVE')} className="rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm font-bold text-slate-700">🔁 غيّر الموعد</button>
          </div>
          {mode === 'MOVE' && visitPicker}
        </div>
      )}

      {/* ٤. الكشف بعد الوصول */}
      {visit && b.techVisitedAt && !b.techDiagnosedAt && (
        <div className="mt-3">
          <label className="text-sm font-bold text-slate-700">📝 الكشف: شنو المشكلة بالضبط وشنو الحل؟</label>
          <textarea value={text} onChange={(e) => setText(e.target.value)} rows={3} className="mt-1 w-full rounded-lg border border-slate-300 p-2 text-sm" placeholder="مثلاً: الكاميرا الثالثة ما تسجّل — الهارد بيه باد سكتر، يحتاج تبديل" />
          <button disabled={busy || text.trim().length < 10} onClick={() => void run('diagnosis', text)} className="mt-1 rounded-lg bg-orange-600 px-3 py-2 text-sm font-bold text-white disabled:opacity-50">احفظ الكشف</button>
        </div>
      )}

      {/* ٥. العلاج */}
      {b.techDiagnosedAt && b.status !== 'COMPLETED' && (
        <div className="mt-3 space-y-2">
          <p className="rounded-lg bg-white p-2 text-sm"><b>📝 الكشف:</b> {b.techDiagnosis}</p>
          <textarea value={fix} onChange={(e) => setFix(e.target.value)} rows={2} className="w-full rounded-lg border border-slate-300 p-2 text-sm" placeholder="شلون حليتها؟ (أو ملاحظة للطاقم إذا تطلب طاقم)" />
          <div className="flex flex-wrap gap-2">
            <button disabled={busy || !fix.trim()} onClick={() => void run('resolve', fix)} className="rounded-lg bg-emerald-600 px-3 py-2 text-sm font-bold text-white disabled:opacity-50">✅ انحلّت — سكّر الحجز</button>
            <button disabled={busy || !!b.techCrewRequestedAt} onClick={() => void run('crew', fix)} className="rounded-lg border border-sky-300 bg-sky-50 px-3 py-2 text-sm font-bold text-sky-800 disabled:opacity-50">
              {b.techCrewRequestedAt ? `👷 طلبت طاقم ${fmt(b.techCrewRequestedAt)}` : '👷 أحتاج طاقم'}
            </button>
          </div>
          <p className="text-[11px] text-slate-500">حتى لو طلبت طاقم، الحجز يبقى برقبتك لحد ما تسكّره.</p>
        </div>
      )}
    </div>
  )
}

export default function TechHandovers({ bookings, me, onChange }: { bookings: Booking[]; me?: string; onChange: (b: Booking) => void }) {
  const mine = bookings.filter((b) => me && b.handoverToId === me && b.status !== 'COMPLETED' && b.status !== 'CANCELLED')
  if (mine.length === 0) return null
  return (
    <div dir="rtl" className="mb-4">
      <h3 className="mb-1 font-bold text-orange-900">🛠️ حجوزات مرحّلة إلك ({mine.length})</h3>
      <p className="mb-2 text-xs text-slate-500">الإداري رحّلها لأن المشكلة تحتاجك. هسه برقبتك: تواصل ويا الزبون، اكتب الكشف، وعالج. ماتركس يحسب وقت كل خطوة.</p>
      <div className="flex flex-col gap-3">{mine.map((b) => <Card key={b.id} b={b} onChange={onChange} />)}</div>
    </div>
  )
}
