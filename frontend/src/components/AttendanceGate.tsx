import { useCallback, useEffect, useState, type ReactNode } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { api, type AttendanceGateState } from '../api'
import { useSession } from '../session'

// ═══ الحضور الإجباري — قرار (ع) 10-06 ═══
// «اول ما يفتح النظام الموظف اريده ينطلب منه تسجيل الحضور… ميعبر للنقطه
// الي بعدها». الكل عدا المالك ومدير النظام والي بإجازة معتمدة (يقرره الخادم).
// إذا الطلب فشل (شبكة) ما نقفل النظام — الحضور مو أهم من شغل الزبون.

// «لا» تنحفظ لليوم بس (بتوقيت بغداد) — باچر يرجع يسأله.
const SKIP_KEY = 'attendance-offer-skip:' + new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })

export default function AttendanceGate({ children }: { children: ReactNode }) {
  const { employee } = useSession()
  const location = useLocation()
  const [gate, setGate] = useState<AttendanceGateState | null>(null)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const exempt = employee?.role === 'ADMIN' || employee?.actualRole === 'OWNER'

  const load = useCallback(() => { if (!exempt) void api.getAttendanceGate().then(setGate).catch(() => setGate(null)) }, [exempt])
  useEffect(load, [load])
  // الوقت يمشي: اللي فتح ٦:٣٠ (برّا الدوام) لازم تطلعله البوابة الإجبارية من ٧.
  useEffect(() => {
    if (exempt) return
    const t = window.setInterval(load, 5 * 60 * 1000)
    window.addEventListener('focus', load)
    return () => { window.clearInterval(t); window.removeEventListener('focus', load) }
  }, [exempt, load])

  const [skipped, setSkipped] = useState(() => { try { return sessionStorage.getItem(SKIP_KEY) === '1' } catch { return false } })
  const skip = () => { try { sessionStorage.setItem(SKIP_KEY, '1') } catch { /* ما يهم */ } setSkipped(true) }
  const unskip = () => { try { sessionStorage.removeItem(SKIP_KEY) } catch { /* ما يهم */ } setSkipped(false) }
  const span = gate ? `${gate.evening ? 'المسائي' : 'الصباحي'} من ${gate.startLabel ?? ''} لـ ${gate.endLabel}` : ''

  if (exempt || !gate || location.pathname.startsWith('/attendance')) return <>{children}</>
  if (!gate.required && !gate.offer) return <>{children}</>

  const checkIn = async () => {
    setBusy(true); setErr('')
    try {
      await api.checkIn()
      window.dispatchEvent(new Event('matrix-refresh'))
      load()
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }
  // برّا وقت دوامه (قرار (ع) 10-07): «اليوم متغيّر شفت دوامك؟» — نعم يسجّل
  // حضور، ولا يشوف جدول دوامه بس (كل الصفحات الثانية تنقفل).
  if (!gate.required) {
    if (skipped) {
      return (
        <div dir="rtl" className="flex min-h-[60vh] items-center justify-center p-4">
          <div className="w-full max-w-md rounded-3xl border border-slate-200 bg-white p-6 text-center shadow-xl">
            <div className="text-5xl">🤖</div>
            <p className="mt-2 text-lg font-extrabold text-[#0f2040]">هسه مو وقت دوامك</p>
            <p className="mt-1 text-sm text-slate-600">دوامك {span}. تكدر تشوف جدول دوامك بس.</p>
            <Link to="/attendance" className="mt-4 block w-full rounded-2xl bg-[#0f2040] py-3 font-extrabold text-white">🗓️ جدول دوامي</Link>
            <button type="button" onClick={unskip} className="mt-2 text-xs text-violet-700 underline">لا، شفتي اليوم متغيّر — أريد أسجّل حضور</button>
          </div>
        </div>
      )
    }
    return (
      <div dir="rtl" className="flex min-h-[70vh] items-center justify-center p-4">
        <div className="w-full max-w-md rounded-3xl border border-violet-200 bg-white p-6 text-center shadow-xl">
          <div className="text-5xl">🤖</div>
          <p className="mt-2 text-lg font-extrabold text-[#0f2040]">هسه مو وقت دوامك</p>
          <p className="mt-1 text-sm text-slate-600">دوامك {span}.</p>
          <p className="mt-3 text-base font-bold text-violet-900">اليوم متغيّر شفت دوامك؟</p>
          {err && <p className="mt-2 text-sm text-red-600">{err}</p>}
          <button type="button" disabled={busy} onClick={() => void checkIn()}
            className="mt-4 w-full rounded-2xl bg-gradient-to-l from-emerald-600 to-emerald-700 py-3 text-lg font-extrabold text-white shadow disabled:opacity-60">
            {busy ? '…' : '✅ إي، سجّل حضوري'}
          </button>
          <button type="button" onClick={skip} className="mt-2 w-full rounded-2xl border border-slate-300 py-2.5 font-bold text-slate-600">لا — بس أشوف جدول دوامي</button>
        </div>
      </div>
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
