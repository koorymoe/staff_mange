import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type RemoteTimer } from '../api'
import { useSession } from '../session'
import { pollVisible } from '../utils/poll'

// ═══ عدّاد ساعات البيت بالشريط العلوي — بكل الشاشات ═══
// (ع): «أريد عدّاد، أني من أشتغل أحسب وقتي». يطلع لصاحب صلاحية ساعات البيت
// (لنفسه أو مسؤول): ▶️/⏸️ من أي شاشة، والوقت يمشي. «إنهاء» من صفحة الساعات.
const hms = (sec: number) => `${Math.floor(sec / 3600)}:${String(Math.floor((sec % 3600) / 60)).padStart(2, '0')}:${String(sec % 60).padStart(2, '0')}`

export default function RemoteTimerChip() {
  const { employee, permissions } = useSession()
  const allowed = employee?.role === 'ADMIN' || employee?.actualRole === 'OWNER'
    || permissions.includes('remote_hours_self') || permissions.includes('remote_hours_manage')
  const [timer, setTimer] = useState<RemoteTimer | null>(null)
  const [base, setBase] = useState(0)
  const [tick, setTick] = useState(0)
  const [busy, setBusy] = useState(false)

  const apply = (t: RemoteTimer) => { setTimer(t); setBase(Date.now()); setTick(0) }
  const load = useCallback(() => { api.getMyRemote().then((r) => apply(r.timer)).catch(() => {}) }, [])
  useEffect(() => {
    if (!allowed) return
    load()
    // تزامن كل دقيقة — لو شغّله من جهاز ثاني يبين هنا.
    return pollVisible(load, 60_000)
  }, [allowed, load])
  useEffect(() => {
    if (!timer?.running) return
    const id = window.setInterval(() => setTick(Math.floor((Date.now() - base) / 1000)), 1000)
    return () => window.clearInterval(id)
  }, [timer?.running, base])

  if (!allowed || !timer) return null
  const shown = timer.elapsed + (timer.running ? tick : 0)
  const toggle = async () => {
    setBusy(true)
    try { apply((await api.remoteTimer(timer.running ? 'pause' : 'start')).timer) } catch { /* netErrors يعرض */ } finally { setBusy(false) }
  }
  return (
    <div className={`flex items-center gap-1 rounded-xl px-1.5 py-1 text-xs font-bold ring-1 ${timer.running ? 'bg-emerald-50 text-emerald-700 ring-emerald-200' : 'bg-slate-50 text-slate-600 ring-slate-200'}`}>
      <button type="button" disabled={busy} onClick={toggle} title={timer.running ? 'توقّف' : shown > 0 ? 'كمّل' : 'ابدأ عدّاد ساعات البيت'}
        className="grid h-6 w-6 place-items-center rounded-lg hover:bg-white disabled:opacity-50">{timer.running ? '⏸️' : '▶️'}</button>
      <Link to="/remote-hours" title="ساعات العمل من البيت — إنهاء وحفظ" className="font-mono tabular-nums" dir="ltr">{shown > 0 || timer.running ? hms(shown) : '⏱️'}</Link>
    </div>
  )
}
