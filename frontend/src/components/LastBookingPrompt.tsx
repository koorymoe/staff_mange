import { useEffect, useState } from 'react'
import { api, type AttendanceGateState } from '../api'

// ═══ «هذا آخر حجز؟» — قرار (ع) 10-06 ═══
// بعد نهاية الشفت (الصباحي ٤ العصر، المسائي ١٢ بالليل)، كل حجز يخلّصه
// الموظف ماتركس يسأله: هذا آخر حجز؟ «لا» يكمّل — «اي» يسجّل انصرافه.
// الشاشات تطلق الحدث «matrix-booking-done» بعد الإنجاز (تام أو جزئي).

export default function LastBookingPrompt() {
  const [gate, setGate] = useState<AttendanceGateState | null>(null)
  const [step, setStep] = useState<'ask' | 'out' | 'done'>('ask')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  useEffect(() => {
    const on = () => {
      void api.getAttendanceGate().then((g) => { if (g.afterShift && g.open) { setStep('ask'); setGate(g) } }).catch(() => {})
    }
    window.addEventListener('matrix-booking-done', on)
    return () => window.removeEventListener('matrix-booking-done', on)
  }, [])
  if (!gate) return null
  const close = () => setGate(null)
  const checkOut = async () => {
    setBusy(true); setErr('')
    try { await api.checkOut(); setStep('done'); window.dispatchEvent(new Event('matrix-refresh')) } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }
  return (
    <div dir="rtl" className="fixed inset-0 z-[80] flex items-end justify-center bg-black/40 p-3 sm:items-center">
      <div className="w-full max-w-sm rounded-3xl bg-white p-5 text-center shadow-2xl">
        <div className="text-4xl">🤖</div>
        {step === 'ask' && (
          <>
            <p className="mt-2 text-lg font-extrabold text-[#0f2040]">خلص دوامك الساعة {gate.endLabel}</p>
            <p className="mt-1 text-sm text-slate-600">هذا آخر حجز اليوم؟</p>
            <div className="mt-4 grid grid-cols-2 gap-2">
              <button type="button" onClick={close} className="rounded-xl border border-slate-300 py-2.5 font-bold text-slate-700">لا، بعدني</button>
              <button type="button" onClick={() => setStep('out')} className="rounded-xl bg-[#0f2040] py-2.5 font-bold text-white">اي، آخر حجز</button>
            </div>
          </>
        )}
        {step === 'out' && (
          <>
            <p className="mt-2 text-lg font-extrabold text-[#0f2040]">عاشت إيدك 👏</p>
            <p className="mt-1 text-sm text-slate-600">سجّل انصرافك هسه.</p>
            {err && <p className="mt-2 text-sm text-red-600">{err}</p>}
            <button type="button" disabled={busy} onClick={() => void checkOut()} className="mt-4 w-full rounded-xl bg-gradient-to-l from-rose-600 to-rose-700 py-3 font-extrabold text-white disabled:opacity-60">{busy ? '…' : '🚪 سجّل انصرافي'}</button>
          </>
        )}
        {step === 'done' && (
          <>
            <p className="mt-2 text-lg font-extrabold text-emerald-700">✅ انسجّل انصرافك</p>
            <p className="mt-1 text-sm text-slate-600">الله وياك، نشوفك باچر.</p>
            <button type="button" onClick={close} className="mt-4 w-full rounded-xl border border-slate-300 py-2.5 font-bold">تمام</button>
          </>
        )}
      </div>
    </div>
  )
}
