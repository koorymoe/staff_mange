import { useCallback, useEffect, useState, type ReactNode } from 'react'
import { useLocation } from 'react-router-dom'
import { api, type AttendanceGateState } from '../api'
import { useSession } from '../session'

// ═══ الحضور الإجباري — قرار (ع) 10-06 ═══
// «اول ما يفتح النظام الموظف اريده ينطلب منه تسجيل الحضور… ميعبر للنقطه
// الي بعدها». الكل عدا المالك ومدير النظام والي بإجازة معتمدة (يقرره الخادم).
// إذا الطلب فشل (شبكة) ما نقفل النظام — الحضور مو أهم من شغل الزبون.

const SKIP_KEY = 'attendance-offer-skip'

export default function AttendanceGate({ children }: { children: ReactNode }) {
  const { employee } = useSession()
  const location = useLocation()
  const [gate, setGate] = useState<AttendanceGateState | null>(null)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const exempt = employee?.role === 'ADMIN' || employee?.actualRole === 'OWNER'

  const load = useCallback(() => { if (!exempt) void api.getAttendanceGate().then(setGate).catch(() => setGate(null)) }, [exempt])
  useEffect(load, [load])

  const [skipped, setSkipped] = useState(() => { try { return sessionStorage.getItem(SKIP_KEY) === '1' } catch { return false } })
  const skip = () => { try { sessionStorage.setItem(SKIP_KEY, '1') } catch { /* ما يهم */ } setSkipped(true) }
  const span = gate ? `${gate.evening ? 'المسائي' : 'الصباحي'} من ${gate.startLabel ?? ''} لـ ${gate.endLabel}` : ''

  if (exempt || !gate || location.pathname.startsWith('/attendance')) return <>{children}</>
  if (!gate.required && !(gate.offer && !skipped)) return <>{children}</>

  const checkIn = async () => {
    setBusy(true); setErr('')
    try {
      await api.checkIn()
      window.dispatchEvent(new Event('matrix-refresh'))
      load()
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }
  // برّا وقت دوامه (قرار (ع) 10-07): نسأله بس، ما نقفل.
  if (!gate.required) {
    return (
      <>
        <div dir="rtl" className="mx-auto mb-3 mt-2 flex max-w-3xl flex-wrap items-center justify-between gap-2 rounded-2xl border border-violet-200 bg-violet-50 px-4 py-2.5 text-sm">
          <span className="text-violet-900">🤖 هسه مو وقت دوامك (دوامك {span}). تريد تسجّل حضور؟</span>
          {err && <span className="text-xs text-red-600">{err}</span>}
          <span className="flex gap-2">
            <button type="button" disabled={busy} onClick={() => void checkIn()} className="rounded-xl bg-emerald-600 px-3 py-1.5 text-xs font-bold text-white disabled:opacity-60">✅ سجّل حضوري</button>
            <button type="button" onClick={skip} className="rounded-xl border border-slate-300 bg-white px-3 py-1.5 text-xs font-bold text-slate-600">لا، بس أتصفح</button>
          </span>
        </div>
        {children}
      </>
    )
  }
  const h = new Date().getHours()
  const hello = h < 12 ? 'صباح الخير' : 'مساء الخير'
  return (
    <div dir="rtl" className="flex min-h-[70vh] items-center justify-center p-4">
      <div className="w-full max-w-md rounded-3xl border border-violet-200 bg-white p-6 text-center shadow-xl">
        <div className="text-5xl">🤖</div>
        <p className="mt-2 text-xl font-extrabold text-[#0f2040]">{hello} {employee?.name?.split(' ')[0] ?? ''}</p>
        <p className="mt-1 text-sm text-slate-600">قبل ما تبدي، سجّل حضورك. ماتركس ما يخليك تعبر بدونه.</p>
        <p className="mt-1 text-xs text-slate-400">هسه وقت دوامك — دوامك {span}.</p>
        {err && <p className="mt-2 text-sm text-red-600">{err}</p>}
        <button type="button" disabled={busy} onClick={() => void checkIn()}
          className="mt-5 w-full rounded-2xl bg-gradient-to-l from-emerald-600 to-emerald-700 py-3 text-lg font-extrabold text-white shadow disabled:opacity-60">
          {busy ? '…' : '✅ سجّل حضوري'}
        </button>
      </div>
    </div>
  )
}
