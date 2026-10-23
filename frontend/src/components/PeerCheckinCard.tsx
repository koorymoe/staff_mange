import { useEffect, useState } from 'react'
import { api, type MyPeerState, type PeerNote } from '../api'

// ═══ «🤖 ماتركس يسألك» — الفضفضة الأسبوعية (قرار (ع) 10-05) ═══
// كل أسبوع (من الخميس) ماتركس يسأل: شلونك؟ شي مضايقك؟ شلون زملاؤك وياك؟
// إذا قيّمت زميل سيء، ماتركس يسألك شنو صار وليش وشلون تريد ينحل، ويقترح حل،
// وبعدها التقرير يروح للمراقب والمدير. والزميل ما يعرف أبداً.

const Stars = ({ value, onChange, size = 'text-xl' }: { value: number; onChange: (n: number) => void; size?: string }) => (
  <span>
    {[1, 2, 3, 4, 5].map((n) => (
      <button key={n} type="button" aria-label={`${n}`} onClick={() => onChange(n)} className={`${size} ${n <= value ? 'text-amber-500' : 'text-slate-300'}`}>★</button>
    ))}
  </span>
)
const MOODS = ['😣', '😕', '😐', '🙂', '😄']

interface Draft { score: number; kind: 'GOOD' | 'NOTE' | 'PROBLEM'; text: string }

function Deepen({ note, onDone }: { note: PeerNote; onDone: (n: PeerNote) => void }) {
  const [what, setWhat] = useState(note.fWhat ?? '')
  const [why, setWhy] = useState(note.fWhy ?? '')
  const [wish, setWish] = useState(note.fWish ?? '')
  const [other, setOther] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const run = async (fn: () => Promise<PeerNote>) => {
    setBusy(true); setErr('')
    try { onDone(await fn()) } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }
  if (!note.suggestion) {
    return (
      <div className="space-y-2 rounded-xl border border-violet-200 bg-violet-50/60 p-3">
        <p className="text-sm text-violet-900">🤖 گلت إنه أكو مشكلة ويا <b>{note.targetName}</b>. خلّيني أفهم حتى أگدر أساعدك:</p>
        <label className="block text-xs font-bold text-slate-700">شنو صار بالتفصيل؟
          <textarea value={what} onChange={(e) => setWhat(e.target.value)} rows={2} className="mt-1 w-full rounded-lg border border-slate-200 p-2 text-sm font-normal" /></label>
        <label className="block text-xs font-bold text-slate-700">شنو السبب برأيك؟
          <textarea value={why} onChange={(e) => setWhy(e.target.value)} rows={2} className="mt-1 w-full rounded-lg border border-slate-200 p-2 text-sm font-normal" /></label>
        <label className="block text-xs font-bold text-slate-700">شلون تريد ينحل الموضوع؟
          <textarea value={wish} onChange={(e) => setWish(e.target.value)} rows={2} className="mt-1 w-full rounded-lg border border-slate-200 p-2 text-sm font-normal" /></label>
        {err && <p className="text-xs text-red-600">{err}</p>}
        <button type="button" disabled={busy || !what.trim()} onClick={() => void run(() => api.peerFollowup(note.id, what, why, wish))}
          className="rounded-lg bg-violet-700 px-4 py-1.5 text-xs font-bold text-white disabled:opacity-50">{busy ? 'ماتركس يفكّر…' : 'كمّل'}</button>
      </div>
    )
  }
  if (note.accepted == null) {
    return (
      <div className="space-y-2 rounded-xl border border-violet-200 bg-violet-50/60 p-3">
        <p className="text-sm text-violet-900">🤖 <b>اقتراحي:</b> {note.suggestion}</p>
        <p className="text-xs text-slate-600">يناسبك؟</p>
        <input value={other} onChange={(e) => setOther(e.target.value)} placeholder="إذا عندك شي تضيفه (اختياري)" className="w-full rounded-lg border border-slate-200 p-2 text-xs" />
        <div className="flex flex-wrap gap-2">
          <button type="button" disabled={busy} onClick={() => void run(() => api.peerAccept(note.id, true, other))} className="rounded-lg bg-emerald-600 px-4 py-1.5 text-xs font-bold text-white disabled:opacity-50">👍 يناسبني</button>
          <button type="button" disabled={busy} onClick={() => void run(() => api.peerAccept(note.id, false, other))} className="rounded-lg border border-slate-300 bg-white px-4 py-1.5 text-xs font-bold text-slate-700 disabled:opacity-50">لا، أريد المراقب يتدخل</button>
        </div>
        {err && <p className="text-xs text-red-600">{err}</p>}
      </div>
    )
  }
  return <p className="rounded-xl bg-emerald-50 p-2 text-xs text-emerald-800">✅ وصل كلامك عن {note.targetName} للمراقب والمدير ويا اقتراحي. راح أسألك بعد أسبوع شلون صارت الأمور.</p>
}

export default function PeerCheckinCard() {
  const [st, setSt] = useState<MyPeerState | null>(null)
  const [mood, setMood] = useState<number | null>(null)
  const [worry, setWorry] = useState('')
  const [sugg, setSugg] = useState('')
  const [drafts, setDrafts] = useState<Record<string, Draft>>({})
  const [extra, setExtra] = useState<string[]>([])
  const [editing, setEditing] = useState(false)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const [fuText, setFuText] = useState<Record<string, string>>({})
  const [fuDone, setFuDone] = useState<Record<string, boolean>>({})

  useEffect(() => { void api.getMyPeer().then(setSt).catch(() => {}) }, [])
  if (!st) return null

  const pendingDeep = st.notes.filter((n) => n.needsFollowup && n.accepted == null)
  const followups = st.followups.filter((f) => !fuDone[f.id])
  const showForm = st.open && (!st.checkin || editing)
  if (!showForm && pendingDeep.length === 0 && followups.length === 0) {
    if (st.checkin && st.open) {
      return (
        <div dir="rtl" className="rounded-2xl border border-violet-100 bg-violet-50/40 p-3 text-xs text-violet-900">
          🤖 شكراً، فضفضة هالأسبوع وصلت. <button type="button" onClick={() => setEditing(true)} className="font-bold underline">تعديل</button>
        </div>
      )
    }
    return null
  }

  const worked = st.colleagues.filter((c) => c.worked)
  const shown = [...worked, ...st.colleagues.filter((c) => extra.includes(c.id))]
  const others = st.colleagues.filter((c) => !c.worked && !extra.includes(c.id))
  const setD = (id: string, patch: Partial<Draft>) => setDrafts((d) => ({ ...d, [id]: { ...(d[id] ?? { score: 0, kind: 'NOTE', text: '' }), ...patch } }))

  const submit = async (nothing: boolean) => {
    setBusy(true); setErr('')
    try {
      const notes = Object.entries(drafts).filter(([, d]) => d.score > 0).map(([targetId, d]) => ({ targetId, score: d.score, kind: d.kind, text: d.text }))
      setSt(await api.submitPeer({ mood, worry, suggestion: sugg, nothing, notes: nothing ? [] : notes }))
      setEditing(false)
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }

  return (
    <div dir="rtl" className="space-y-3 rounded-2xl border border-violet-200 bg-white p-4 shadow-sm">
      <p className="text-base font-extrabold text-violet-900">🤖 ماتركس يسألك</p>

      {followups.map((f) => (
        <div key={f.id} className="space-y-2 rounded-xl border border-amber-200 bg-amber-50 p-3 text-sm">
          <p>🤝 بخصوص «{f.title}» — شلون صارت الأمور ويا زميلك؟</p>
          <input value={fuText[f.id] ?? ''} onChange={(e) => setFuText((t) => ({ ...t, [f.id]: e.target.value }))} placeholder="إذا تريد تضيف شي" className="w-full rounded-lg border border-slate-200 p-2 text-xs" />
          <div className="flex gap-2">
            {[true, false].map((better) => (
              <button key={String(better)} type="button" onClick={async () => { await api.issuePartyFollowup(f.id, better, fuText[f.id] ?? ''); setFuDone((d) => ({ ...d, [f.id]: true })) }}
                className={`rounded-lg px-3 py-1.5 text-xs font-bold ${better ? 'bg-emerald-600 text-white' : 'border border-slate-300 bg-white text-slate-700'}`}>{better ? '👍 تحسّنت' : 'بعدها ما تحسّنت'}</button>
            ))}
          </div>
        </div>
      ))}

      {pendingDeep.map((n) => <Deepen key={n.id} note={n} onDone={(u) => setSt((s) => s && ({ ...s, notes: s.notes.map((x) => (x.id === u.id ? u : x)) }))} />)}

      {showForm && (
        <div className="space-y-3">
          <p className="text-xs text-slate-500">كلامك يوصل للمالك والمدير بس (والمراقب يشوف تقرير المشاكل). <b>الزميل ما يعرف أبداً شنو كتبت عنه.</b></p>
          <div>
            <p className="mb-1 text-sm font-bold">شلونك هالأسبوع؟</p>
            <div className="flex gap-1">
              {MOODS.map((m, i) => (
                <button key={i} type="button" onClick={() => setMood(i + 1)} className={`rounded-xl px-2 py-1 text-2xl ${mood === i + 1 ? 'bg-violet-100 ring-2 ring-violet-400' : 'opacity-60 hover:opacity-100'}`}>{m}</button>
              ))}
            </div>
          </div>
          <textarea value={worry} onChange={(e) => setWorry(e.target.value)} rows={2} placeholder="شي مضايقك بالشغل؟ احچي براحتك (اختياري)" className="w-full rounded-lg border border-slate-200 p-2 text-sm" />
          <textarea value={sugg} onChange={(e) => setSugg(e.target.value)} rows={2} placeholder="عندك اقتراح يحسّن الشغل؟ (اختياري)" className="w-full rounded-lg border border-slate-200 p-2 text-sm" />

          <div>
            <p className="mb-1 text-sm font-bold">شلون زملاؤك وياك هالأسبوع؟ <span className="text-xs font-normal text-slate-500">(قيّم الي تريد بس)</span></p>
            <div className="space-y-2">
              {shown.map((c) => {
                const d = drafts[c.id] ?? { score: 0, kind: 'NOTE', text: '' }
                return (
                  <div key={c.id} className="rounded-xl border border-slate-200 p-2">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="min-w-[7rem] text-sm font-bold">{c.name}</span>
                      <Stars value={d.score} onChange={(n) => setD(c.id, { score: n, kind: n <= 2 ? 'PROBLEM' : n >= 4 ? 'GOOD' : 'NOTE' })} />
                    </div>
                    {d.score > 0 && (
                      <input value={d.text} onChange={(e) => setD(c.id, { text: e.target.value })}
                        placeholder={d.score <= 2 ? 'شنو المشكلة؟ ماتركس راح يسألك عنها أكثر' : 'كلمة عنه (اختياري)'}
                        className="mt-1 w-full rounded-lg border border-slate-200 p-1.5 text-xs" />
                    )}
                  </div>
                )
              })}
              {shown.length === 0 && <p className="text-xs text-slate-400">ما لگيت زملاء اشتغلت وياهم هالأسبوع — تگدر تضيف تحت.</p>}
            </div>
            {others.length > 0 && (
              <select value="" onChange={(e) => e.target.value && setExtra((x) => [...x, e.target.value])} className="mt-2 rounded-lg border border-slate-200 px-2 py-1 text-xs">
                <option value="">+ أضيف زميل ثاني</option>
                {others.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
            )}
          </div>
          {err && <p className="text-xs text-red-600">{err}</p>}
          <div className="flex flex-wrap gap-2">
            <button type="button" disabled={busy} onClick={() => void submit(false)} className="rounded-lg bg-violet-700 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">{busy ? '…' : 'أرسل'}</button>
            <button type="button" disabled={busy} onClick={() => void submit(true)} className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-bold text-slate-600 disabled:opacity-50">ماكو شي، كلشي زين 👍</button>
          </div>
        </div>
      )}
    </div>
  )
}
