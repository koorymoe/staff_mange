import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type PeerVoiceReport, type PeerNote } from '../api'
import MatrixNote from '../components/MatrixNote'

// ═══ «🗣️ صوت الموظفين» — قرار (ع) 10-05 ═══
// المدير والمالك: كل الفضفضات والملاحظات بنصّها ومنو كتبها.
// المراقب: تقارير التعمّق بس (الي ماتركس سأل عنها ورفعها).
// والموظف المقصود ما يشوف هالشاشة أبداً.

const SEV: Record<PeerNote['severity'], { cls: string; label: string }> = {
  URGENT: { cls: 'border-red-300 bg-red-50', label: '🚨 عاجل' },
  ATTENTION: { cls: 'border-amber-200 bg-amber-50', label: '⚠️ يحتاج انتباه' },
  NORMAL: { cls: 'border-slate-200 bg-white', label: '' },
}
const KIND: Record<PeerNote['kind'], string> = { GOOD: '👍 إيجابي', NOTE: '📝 ملاحظة', PROBLEM: '⚠️ مشكلة' }
const d = (s: string) => new Date(s).toLocaleDateString('en-GB')

export default function PeerVoicePage() {
  const nav = useNavigate()
  const [r, setR] = useState<PeerVoiceReport | null>(null)
  const [err, setErr] = useState('')
  useEffect(() => { void api.getPeerVoice().then(setR).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر')) }, [])
  if (err) return <p className="text-red-600">{err}</p>
  if (!r) return <p className="text-slate-400">جاري التحميل…</p>
  const maxMood = 5

  const openIssue = (a: string, b: string | null, title: string, desc: string, noteId?: string) => {
    const q = new URLSearchParams({ a, title, desc })
    if (b) q.set('b', b)
    if (noteId) q.set('note', noteId)
    nav(`/workplace-issues?new=1&${q.toString()}`)
  }

  return (
    <div dir="rtl" className="space-y-4">
      <div>
        <h2 className="text-2xl font-bold text-brand-900">🗣️ صوت الموظفين</h2>
        <p className="text-sm text-slate-500">{r.full ? 'فضفضات الموظفين الأسبوعية وتقييمهم لزملائهم ويا تحليل ماتركس. الموظف المقصود ما يشوف شي.' : 'تقارير ماتركس عن المشاكل الي حچوا عنها الموظفين (بعد ما سألهم وتعمّق).'}</p>
      </div>
      <MatrixNote>{r.insights.join(' ')}</MatrixNote>

      {r.full && r.mood.length > 0 && (
        <div className="rounded-2xl border border-slate-200 bg-white p-3">
          <p className="mb-2 text-xs font-extrabold text-slate-700">مزاج الفريق آخر الأسابيع</p>
          <div className="flex h-24 items-end gap-2">
            {r.mood.map((m) => (
              <div key={m.week} className="flex max-w-16 flex-1 flex-col items-center gap-1" title={`${d(m.week)}: ${m.avg?.toFixed(1) ?? '—'} (${m.count})`}>
                <div className="w-full rounded-t bg-violet-400" style={{ height: `${((m.avg ?? 0) / maxMood) * 80}px` }} />
                <span className="text-[10px] text-slate-500">{m.avg?.toFixed(1) ?? '—'}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {r.tensions.length > 0 && (
        <div className="space-y-2 rounded-2xl border border-red-200 bg-red-50/60 p-3">
          <p className="text-sm font-extrabold text-red-800">⚠️ توتر متبادل (كل واحد قيّم الثاني سيء بآخر أسبوعين)</p>
          {r.tensions.map((t) => (
            <div key={t.aId + t.bId} className="flex flex-wrap items-center justify-between gap-2 rounded-xl bg-white p-2 text-sm">
              <span><b>{t.aName}</b> ⇄ <b>{t.bName}</b> <span className="text-xs text-slate-500">({t.aToB}★ و{t.bToA}★)</span></span>
              <button type="button" onClick={() => openIssue(t.aId, t.bId, `توتر بين ${t.aName} و${t.bName}`, 'ماتركس لگى تقييم سيء متبادل بينهم بالفضفضة الأسبوعية.')}
                className="rounded-lg bg-red-600 px-3 py-1 text-xs font-bold text-white">افتح مشكلة وظيفية</button>
            </div>
          ))}
        </div>
      )}

      {r.scores.length > 0 && (
        <div className="rounded-2xl border border-slate-200 bg-white p-3">
          <p className="mb-2 text-xs font-extrabold text-slate-700">شكد زملاؤه راضين منه (آخر ٤ أسابيع)</p>
          <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
            {r.scores.map((s) => (
              <div key={s.employeeId} className="rounded-lg border border-slate-200 p-2 text-xs">
                <div className="flex justify-between"><b>{s.name}</b><b style={{ color: s.avg >= 4 ? '#059669' : s.avg >= 3 ? '#d97706' : '#dc2626' }}>{s.avg.toFixed(1)}★</b></div>
                <p className="text-slate-500">{s.count} تقييم · {s.problems} مشكلة{s.prevAvg != null && ` · قبلها ${s.prevAvg.toFixed(1)}★`}</p>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="space-y-2">
        <p className="text-sm font-extrabold text-slate-700">{r.full ? 'الملاحظات على الزملاء' : 'تقارير ماتركس'}</p>
        {r.notes.map((n) => (
          <div key={n.id} className={`rounded-2xl border p-3 text-sm ${SEV[n.severity].cls}`}>
            <div className="flex flex-wrap items-center justify-between gap-2">
              <p><b>{n.authorName}</b> عن <b>{n.targetName}</b> · {'★'.repeat(n.score)}{'☆'.repeat(5 - n.score)} · {KIND[n.kind]}</p>
              <span className="text-xs font-bold">{SEV[n.severity].label} <span className="font-normal text-slate-400">{d(n.createdAt)}</span></span>
            </div>
            {n.text && <p className="mt-1 text-slate-700">«{n.text}»</p>}
            {n.fWhat && (
              <div className="mt-2 space-y-0.5 rounded-lg bg-white/70 p-2 text-xs">
                <p><b>شنو صار:</b> {n.fWhat}</p>
                {n.fWhy && <p><b>السبب برأيه:</b> {n.fWhy}</p>}
                {n.fWish && <p><b>شلون يريد ينحل:</b> {n.fWish}</p>}
              </div>
            )}
            {n.suggestion && <p className="mt-2 text-xs text-violet-800">🤖 اقتراح ماتركس: {n.suggestion}</p>}
            {n.accepted != null && <p className="mt-1 text-xs">{n.accepted ? '👍 الموظف وافق على الاقتراح' : '✋ الموظف يريد تدخّل المراقب'}{n.acceptNote && ` — «${n.acceptNote}»`}</p>}
            {n.needsFollowup && (
              <button type="button" onClick={() => openIssue(n.authorId, n.targetId, `مشكلة بين ${n.authorName} و${n.targetName}`, n.fWhat ?? n.text ?? '', n.id)}
                className="mt-2 rounded-lg border border-slate-300 bg-white px-3 py-1 text-xs font-bold">افتح مشكلة وظيفية</button>
            )}
          </div>
        ))}
        {r.notes.length === 0 && <p className="rounded-xl bg-white p-4 text-center text-xs text-slate-400">ماكو شي بهالفترة.</p>}
      </div>

      {r.full && r.checkins.length > 0 && (
        <div className="space-y-2">
          <p className="text-sm font-extrabold text-slate-700">شنو مضايقهم ومقترحاتهم</p>
          {r.checkins.map((c) => (
            <div key={c.id} className={`rounded-xl border p-3 text-sm ${c.urgent ? 'border-red-300 bg-red-50' : 'border-slate-200 bg-white'}`}>
              <p className="text-xs text-slate-500"><b className="text-slate-800">{c.name}</b> · المزاج {c.mood ?? '—'}/5 · {d(c.createdAt)} {c.urgent && '🚨'}</p>
              {c.worry && <p>😟 {c.worry}</p>}
              {c.suggestion && <p>💡 {c.suggestion}</p>}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
