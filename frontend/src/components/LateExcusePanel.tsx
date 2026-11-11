import { useCallback, useEffect, useState } from 'react'
import { apiRequest } from '../api'
import type { Discipline, LateExcuse } from '../apiTypes'
import { todayBaghdad } from '../utils/format'
import DisciplineBox from './DisciplineBox'

// (ل) 11-11 دازاي: عذر التأخير — الموظف يقدّم، ومسؤوله المباشر يوافق أو يرفض.
// المقبول ما ينحسب عليه بالانضباط. الغياب الكامل يمشي بطلب إجازة.
const STATUS: Record<string, { t: string; cls: string }> = {
  PENDING: { t: 'بانتظار المسؤول', cls: 'bg-amber-100 text-amber-800' },
  APPROVED: { t: 'مقبول ✅', cls: 'bg-emerald-100 text-emerald-700' },
  REJECTED: { t: 'مرفوض', cls: 'bg-red-100 text-red-700' },
}
const day = (s: string) => s.slice(0, 10)

export default function LateExcusePanel() {
  const [mine, setMine] = useState<{ excuses: LateExcuse[]; discipline: Discipline | null; startsOn: string } | null>(null)
  const [inbox, setInbox] = useState<LateExcuse[]>([])
  const [date, setDate] = useState(todayBaghdad())
  const [reason, setReason] = useState('')
  const [minutes, setMinutes] = useState('')
  const [msg, setMsg] = useState<{ ok: boolean; t: string } | null>(null)
  const [busy, setBusy] = useState(false)
  const [notes, setNotes] = useState<Record<string, string>>({})
  const load = useCallback(() => {
    apiRequest<{ excuses: LateExcuse[]; discipline: Discipline | null; startsOn: string }>('/late-excuses/mine').then(setMine).catch(() => setMine(null))
    apiRequest<LateExcuse[]>('/late-excuses/inbox').then(setInbox).catch(() => setInbox([]))
  }, [])
  useEffect(() => { load() }, [load])

  const submit = async () => {
    setBusy(true); setMsg(null)
    try {
      await apiRequest('/late-excuses', { method: 'POST', body: JSON.stringify({ date, reason, minutes: minutes ? Number(minutes) : null }) })
      setReason(''); setMinutes(''); setMsg({ ok: true, t: 'انرسل العذر لمسؤولك ✅' }); load()
    } catch (e) { setMsg({ ok: false, t: (e as Error).message }) } finally { setBusy(false) }
  }
  const decide = async (id: string, approve: boolean) => {
    setBusy(true)
    try {
      await apiRequest(`/late-excuses/${id}/decide`, { method: 'POST', body: JSON.stringify({ approve, note: notes[id] ?? '' }) })
      load()
    } catch (e) { setMsg({ ok: false, t: (e as Error).message }) } finally { setBusy(false) }
  }
  const pending = inbox.filter((x) => x.status === 'PENDING')

  return (
    <section className="space-y-3 rounded-2xl bg-white p-4 shadow-sm">
      <h2 className="text-base font-extrabold text-slate-800">⏰ أعذار التأخير والانضباط</h2>
      {mine && (
        <p className="rounded-xl bg-sky-50 p-2 text-xs text-sky-900">
          من {mine.startsOn} التأخير بلا عذر ينحسب بالتقييم الشهري (١٥٪): كل يوم تأخير بلا عذر −٢، وكل غياب بلا إجازة −٥. إذا راح تتأخر أو تأخرت لسبب، قدّم عذر هنا ومسؤولك يوافق.
        </p>
      )}
      {mine?.discipline && <DisciplineBox d={mine.discipline} title="انضباطي هالشهر" />}
      {msg && <p role={msg.ok ? 'status' : 'alert'} className={`rounded-xl p-2 text-sm ${msg.ok ? 'bg-emerald-50 text-emerald-800' : 'bg-red-50 text-red-700'}`}>{msg.t}</p>}

      <div className="grid gap-2 sm:grid-cols-[auto_auto_1fr_auto] sm:items-end">
        <label className="text-xs">اليوم<input id="lx-date" type="date" value={date} onChange={(e) => setDate(e.target.value)} className="block rounded-lg border border-slate-300 px-2 py-1.5 text-sm" /></label>
        <label className="text-xs">كم دقيقة (اختياري)<input id="lx-min" type="number" min={1} value={minutes} onChange={(e) => setMinutes(e.target.value)} className="block w-28 rounded-lg border border-slate-300 px-2 py-1.5 text-sm" /></label>
        <label className="text-xs">السبب<input id="lx-reason" value={reason} onChange={(e) => setReason(e.target.value)} placeholder="مثلاً: مراجعة دائرة، عطل سيارة، حالة صحية" className="block w-full rounded-lg border border-slate-300 px-2 py-1.5 text-sm" /></label>
        <button type="button" disabled={busy || reason.trim().length < 3} onClick={() => void submit()} className="rounded-xl bg-[#0f2040] px-4 py-2 text-sm font-bold text-white disabled:opacity-50">قدّم العذر</button>
      </div>

      {mine && mine.excuses.length > 0 && (
        <ul className="space-y-1 text-sm">
          {mine.excuses.slice(0, 6).map((x) => (
            <li key={x.id} className="flex flex-wrap items-center gap-2">
              <span className="tabular-nums text-slate-600">{day(x.date)}</span>
              <span className={`rounded-full px-2 py-0.5 text-xs font-bold ${STATUS[x.status].cls}`}>{STATUS[x.status].t}</span>
              <span className="text-slate-700">{x.reason}</span>
              {x.decisionNote && <span className="text-xs text-slate-500">— {x.decidedBy}: {x.decisionNote}</span>}
            </li>
          ))}
        </ul>
      )}

      {inbox.length > 0 && (
        <div className="space-y-2 border-t border-slate-200 pt-3">
          <h3 className="text-sm font-extrabold text-slate-800">📥 أعذار كادرك {pending.length ? <span className="rounded-full bg-amber-100 px-2 text-amber-800">{pending.length} تنتظر</span> : null}</h3>
          {inbox.map((x) => (
            <div key={x.id} className="flex flex-wrap items-center gap-2 rounded-xl bg-slate-50 p-2 text-sm">
              <b>{x.employeeName}</b>
              <span className="tabular-nums text-slate-600">{day(x.date)}{x.minutes ? ` · ${x.minutes} دقيقة` : ''}</span>
              <span className="flex-1 text-slate-700">{x.reason}</span>
              {x.status === 'PENDING' ? (
                <>
                  <input aria-label="ملاحظة" placeholder="ملاحظة (اختياري)" value={notes[x.id] ?? ''} onChange={(e) => setNotes((n) => ({ ...n, [x.id]: e.target.value }))} className="w-40 rounded-lg border border-slate-300 px-2 py-1 text-xs" />
                  <button type="button" disabled={busy} onClick={() => void decide(x.id, true)} className="rounded-lg bg-emerald-700 px-3 py-1 text-xs font-bold text-white">قبول</button>
                  <button type="button" disabled={busy} onClick={() => void decide(x.id, false)} className="rounded-lg bg-red-600 px-3 py-1 text-xs font-bold text-white">رفض</button>
                </>
              ) : (
                <span className={`rounded-full px-2 py-0.5 text-xs font-bold ${STATUS[x.status].cls}`}>{STATUS[x.status].t}{x.decidedBy ? ` · ${x.decidedBy}` : ''}</span>
              )}
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
