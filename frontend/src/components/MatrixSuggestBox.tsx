import { useState } from 'react'
import { api, type MatrixBookingSuggestion } from '../api'
import MatrixNote from './MatrixNote'

// ═══ المرحلة الثانية: ماتركس يقترح والمنسق يقرر ═══
// بالطلب (زر) — ماتركس يقترح موعد وكادر ويشرح ليش. المنسق يطبّقه بضغطة أو
// يختار غيره؛ ماتركس يسجّل شنو صار حتى تنقاس دقته. ماكو تنفيذ تلقائي.

const fmt = (iso: string) => new Date(iso).toLocaleString('ar-IQ', { weekday: 'long', day: 'numeric', month: 'numeric', hour: '2-digit', minute: '2-digit' })
const toLocal = (iso: string) => {
  const d = new Date(iso)
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}T${p(d.getHours())}:${p(d.getMinutes())}`
}
const ROLES = ['TECH_1', 'TECH_2', 'TECH_3'] as const

export default function MatrixSuggestBox({ bookingId, onSchedule, onLeader, onTech }: {
  bookingId: string
  onSchedule: (localValue: string) => Promise<void>
  onLeader: (employeeId: string) => Promise<void>
  onTech: (role: (typeof ROLES)[number], employeeId: string) => Promise<void>
}) {
  const [s, setS] = useState<MatrixBookingSuggestion | null>(null)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const [done, setDone] = useState<Record<string, boolean>>({})

  const load = async () => {
    setBusy(true); setErr('')
    try { setS(await api.getMatrixSuggestion(bookingId)) } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }

  if (!s) {
    return (
      <div className="mt-3">
        <button type="button" disabled={busy} onClick={() => void load()}
          className="rounded-lg border border-violet-300 bg-violet-50 px-3 py-1.5 text-xs font-bold text-violet-800 hover:bg-violet-100 disabled:opacity-50">
          {busy ? 'ماتركس يفكّر…' : '🤖 اقتراح ماتركس للموعد والكادر'}
        </button>
        {err && <span className="ms-2 text-xs text-red-600">{err}</span>}
      </div>
    )
  }

  const applyCrew = async () => {
    if (!s.crew) return
    setBusy(true)
    try {
      if (s.crew.leader) await onLeader(s.crew.leader.id)
      for (let i = 0; i < s.crew.techs.length && i < ROLES.length; i++) await onTech(ROLES[i], s.crew.techs[i].id)
      setDone((d) => ({ ...d, crew: true }))
    } finally { setBusy(false) }
  }

  return (
    <div dir="rtl" className="mt-3 space-y-2 rounded-xl border border-violet-200 bg-violet-50/50 p-3 text-sm">
      <div className="flex items-center justify-between">
        <b className="text-violet-900">🤖 اقتراح ماتركس</b>
        <button type="button" onClick={() => setS(null)} className="text-xs text-slate-500">✖ سكّر</button>
      </div>
      {s.note && <MatrixNote>{s.note}</MatrixNote>}
      {s.schedule ? (
        <div className="rounded-lg bg-white p-2">
          <p className="font-bold">📅 الموعد: {fmt(s.schedule.at)}</p>
          <ul className="mt-1 list-inside list-disc text-xs text-slate-600">{s.scheduleWhy.map((w, i) => <li key={i}>{w}</li>)}</ul>
          <button type="button" disabled={busy || done.schedule} onClick={async () => { setBusy(true); try { await onSchedule(toLocal(s.schedule!.at)); setDone((d) => ({ ...d, schedule: true })) } finally { setBusy(false) } }}
            className="mt-2 rounded-lg bg-violet-700 px-3 py-1 text-xs font-bold text-white disabled:opacity-50">{done.schedule ? '✓ انطبّق' : 'طبّق الموعد'}</button>
        </div>
      ) : s.scheduleWhy.length > 0 && <p className="text-xs text-slate-600">📅 {s.scheduleWhy.join(' ')}</p>}
      {s.crew ? (
        <div className="rounded-lg bg-white p-2">
          <p className="text-xs text-slate-500">👷 الكادر لموعد {fmt(s.crew.for)}:</p>
          {s.crew.leader && <p className="mt-1"><b>الليدر: {s.crew.leader.name}</b> <span className="text-xs text-slate-500">— {s.crew.leader.why.join('، ')}</span></p>}
          {s.crew.techs.map((t) => <p key={t.id}><b>فني: {t.name}</b> <span className="text-xs text-slate-500">— {t.why.join('، ')}</span></p>)}
          {(s.crew.altLeaders.length > 0 || s.crew.altTechs.length > 0) && (
            <p className="mt-1 text-xs text-slate-500">بدائل: {[...s.crew.altLeaders, ...s.crew.altTechs].map((p) => p.name).join('، ')}</p>
          )}
          <ul className="mt-1 list-inside list-disc text-xs text-slate-600">{s.crewWhy.map((w, i) => <li key={i}>{w}</li>)}</ul>
          {(s.crew.leader || s.crew.techs.length > 0) && (
            <button type="button" disabled={busy || done.crew} onClick={() => void applyCrew()}
              className="mt-2 rounded-lg bg-violet-700 px-3 py-1 text-xs font-bold text-white disabled:opacity-50">{done.crew ? '✓ انطبّق' : 'طبّق الكادر'}</button>
          )}
        </div>
      ) : s.crewWhy.length > 0 && <p className="text-xs text-slate-600">👷 {s.crewWhy.join(' ')}</p>}
      <p className="text-[11px] text-slate-400">القرار إلك. ماتركس يسجّل إذا مشيت على اقتراحه لو غيّرته، حتى تنقاس دقته.</p>
    </div>
  )
}
