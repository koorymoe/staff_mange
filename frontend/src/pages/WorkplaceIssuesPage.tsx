import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { api, type MyPeerState, type WorkplaceIssue, type WorkplaceIssueDetail } from '../api'
import MatrixNote from '../components/MatrixNote'

// ═══ المشاكل الوظيفية — قرار (ع) 10-05 ═══
// المراقب أو المدير يفتح المشكلة، يسجّل إفادة كل طرف، ماتركس يرتّب ويقترح،
// والمدير يقرر ويحدد موعد متابعة — وبعدها ماتركس يسأل الطرفين شلون صارت.
// ماكو أي عقوبة تلقائية.

const STATUS: Record<WorkplaceIssue['status'], { label: string; cls: string }> = {
  OPEN: { label: 'مفتوحة', cls: 'bg-red-100 text-red-800' },
  IN_PROGRESS: { label: 'قيد المعالجة', cls: 'bg-amber-100 text-amber-800' },
  RESOLVED: { label: 'انحلّت', cls: 'bg-emerald-100 text-emerald-800' },
}
const dt = (s: string) => new Date(s).toLocaleString('en-GB', { day: '2-digit', month: '2-digit', year: '2-digit', hour: '2-digit', minute: '2-digit' })
const KIND: Record<string, string> = { STATEMENT: '🗣️ إفادة', NOTE: '📝 ملاحظة', DECISION: '⚖️ قرار', FOLLOWUP: '🤝 متابعة', MATRIX: '🤖 ماتركس' }

function Detail({ id, people }: { id: string; people: { id: string; name: string }[] }) {
  const [d, setD] = useState<WorkplaceIssueDetail | null>(null)
  const [text, setText] = useState('')
  const [kind, setKind] = useState('STATEMENT')
  const [decision, setDecision] = useState('')
  const [follow, setFollow] = useState('')
  const [resolved, setResolved] = useState(false)
  const [busy, setBusy] = useState(false)
  useEffect(() => { void api.getWorkplaceIssue(id).then(setD) }, [id])
  if (!d) return <p className="p-3 text-xs text-slate-400">…</p>
  void people
  const run = async (fn: () => Promise<WorkplaceIssueDetail>) => { setBusy(true); try { setD(await fn()) } finally { setBusy(false) } }
  return (
    <div className="space-y-3 border-t border-slate-100 p-3">
      <p className="text-sm text-slate-700">{d.description}</p>
      <p className="text-xs text-slate-500">السياق: اشتغلوا سوة بـ{d.context.sharedBookings} حجز آخر ٣ أشهر · شكاوى زبائن: {d.partyAName} {d.context.aComplaints}{d.partyBName && ` · ${d.partyBName} ${d.context.bComplaints}`}</p>
      {d.aiSummary && <div className="matrix-note whitespace-pre-wrap">🤖 <b className="mx-tag">ماتركس:</b> {d.aiSummary}</div>}
      <button type="button" disabled={busy} onClick={() => void run(() => api.analyzeIssue(id))} className="text-xs font-bold text-violet-700 underline">🔄 خلّي ماتركس يعيد التحليل</button>
      <div className="space-y-1">
        {d.entries.map((e) => (
          <div key={e.id} className="rounded-lg bg-slate-50 p-2 text-xs"><b>{KIND[e.kind] ?? e.kind}</b> · {e.authorName ?? 'النظام'} · <span className="text-slate-400">{dt(e.createdAt)}</span><p className="mt-0.5 text-sm text-slate-700">{e.text}</p></div>
        ))}
      </div>
      {d.status !== 'RESOLVED' && (
        <>
          <div className="flex flex-wrap gap-2">
            <select value={kind} onChange={(e) => setKind(e.target.value)} className="rounded-lg border border-slate-200 px-2 py-1 text-xs">
              <option value="STATEMENT">🗣️ إفادة طرف</option><option value="NOTE">📝 ملاحظة</option>
            </select>
            <input value={text} onChange={(e) => setText(e.target.value)} placeholder="سجّل إفادة أو ملاحظة (ماتركس يحدّث تحليله)" className="min-w-0 flex-1 rounded-lg border border-slate-200 px-2 py-1 text-sm" />
            <button type="button" disabled={busy || !text.trim()} onClick={() => void run(async () => { const r = await api.addIssueEntry(id, kind, text); setText(''); return r })} className="rounded-lg bg-slate-800 px-3 py-1 text-xs font-bold text-white disabled:opacity-50">سجّل</button>
          </div>
          <div className="space-y-2 rounded-xl border border-slate-200 p-2">
            <p className="text-xs font-bold">⚖️ القرار</p>
            <textarea value={decision} onChange={(e) => setDecision(e.target.value)} rows={2} placeholder="شنو القرار؟" className="w-full rounded-lg border border-slate-200 p-2 text-sm" />
            <div className="flex flex-wrap items-center gap-3 text-xs">
              <label>موعد المتابعة <input type="date" value={follow} onChange={(e) => setFollow(e.target.value)} className="rounded-lg border border-slate-200 px-2 py-1" /></label>
              <label className="flex items-center gap-1"><input type="checkbox" checked={resolved} onChange={(e) => setResolved(e.target.checked)} /> انحلّت</label>
              <button type="button" disabled={busy || !decision.trim()} onClick={() => void run(() => api.decideIssue(id, decision, resolved ? 'RESOLVED' : 'IN_PROGRESS', follow || null))} className="rounded-lg bg-emerald-700 px-3 py-1 font-bold text-white disabled:opacity-50">احفظ القرار</button>
            </div>
            <p className="text-[11px] text-slate-400">بعد موعد المتابعة، ماتركس يسأل الطرفين بفضفضتهم «شلون صارت الأمور ويا زميلك؟».</p>
          </div>
        </>
      )}
    </div>
  )
}

export default function WorkplaceIssuesPage() {
  const [params, setParams] = useSearchParams()
  const [rows, setRows] = useState<WorkplaceIssue[] | null>(null)
  const [people, setPeople] = useState<MyPeerState['colleagues']>([])
  const [open, setOpen] = useState<string | null>(null)
  const [form, setForm] = useState(() => ({
    show: params.get('new') === '1', title: params.get('title') ?? '', a: params.get('a') ?? '', b: params.get('b') ?? '',
    desc: params.get('desc') ?? '', severity: 'ATTENTION', note: params.get('note'),
  }))
  const [err, setErr] = useState('')
  const [reload, setReload] = useState(0)
  useEffect(() => { void api.getWorkplaceIssues().then(setRows).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر')) }, [reload])
  useEffect(() => { void api.getMyPeer().then((s) => setPeople(s.colleagues)).catch(() => {}) }, [])

  const create = async () => {
    setErr('')
    try {
      const r = await api.openWorkplaceIssue({ title: form.title, partyAId: form.a, partyBId: form.b || null, description: form.desc, severity: form.severity, sourceNoteId: form.note })
      setForm((f) => ({ ...f, show: false, title: '', a: '', b: '', desc: '', note: null }))
      setParams({})
      setOpen(r.id)
      setReload((n) => n + 1)
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') }
  }

  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">🤝 المشاكل الوظيفية</h2>
          <p className="text-sm text-slate-500">مشاكل بين الموظفين: إفادات، تحليل ماتركس واقتراحه، قرار المدير، ومتابعة.</p>
        </div>
        <button type="button" onClick={() => setForm((f) => ({ ...f, show: !f.show }))} className="rounded-xl bg-[#0f2040] px-4 py-2 text-sm font-bold text-white">+ مشكلة جديدة</button>
      </div>
      {err && <p className="text-sm text-red-600">{err}</p>}
      {form.show && (
        <div className="space-y-2 rounded-2xl border border-slate-200 bg-white p-4">
          <input value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} placeholder="العنوان (مثلاً: عركة بالمخزن)" className="w-full rounded-lg border border-slate-200 p-2 text-sm" />
          <div className="grid gap-2 sm:grid-cols-3">
            <select value={form.a} onChange={(e) => setForm({ ...form, a: e.target.value })} className="rounded-lg border border-slate-200 p-2 text-sm">
              <option value="">الطرف الأول</option>{people.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
            </select>
            <select value={form.b} onChange={(e) => setForm({ ...form, b: e.target.value })} className="rounded-lg border border-slate-200 p-2 text-sm">
              <option value="">الطرف الثاني (إذا أكو)</option>{people.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
            </select>
            <select value={form.severity} onChange={(e) => setForm({ ...form, severity: e.target.value })} className="rounded-lg border border-slate-200 p-2 text-sm">
              <option value="NORMAL">عادية</option><option value="ATTENTION">تحتاج انتباه</option><option value="URGENT">🚨 عاجلة</option>
            </select>
          </div>
          <textarea value={form.desc} onChange={(e) => setForm({ ...form, desc: e.target.value })} rows={3} placeholder="شنو صار؟" className="w-full rounded-lg border border-slate-200 p-2 text-sm" />
          <button type="button" disabled={!form.title.trim() || !form.a || !form.desc.trim()} onClick={() => void create()} className="rounded-lg bg-emerald-700 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">افتح</button>
        </div>
      )}
      {rows === null ? <p className="text-slate-400">جاري التحميل…</p> : (
        <div className="space-y-2">
          {rows.length === 0 && <MatrixNote>ماكو مشاكل وظيفية مسجّلة.</MatrixNote>}
          {rows.map((w) => (
            <div key={w.id} className={`rounded-2xl border bg-white ${w.severity === 'URGENT' ? 'border-red-300' : 'border-slate-200'}`}>
              <button type="button" onClick={() => setOpen(open === w.id ? null : w.id)} className="flex w-full flex-wrap items-center justify-between gap-2 p-3 text-right">
                <span className="text-sm font-bold">{w.severity === 'URGENT' && '🚨 '}{w.title} <span className="text-xs font-normal text-slate-500">· {w.partyAName}{w.partyBName && ` ⇄ ${w.partyBName}`}</span></span>
                <span className="flex items-center gap-2 text-xs"><span className={`rounded-full px-2 py-0.5 font-bold ${STATUS[w.status].cls}`}>{STATUS[w.status].label}</span><span className="text-slate-400">{dt(w.createdAt)}</span></span>
              </button>
              {open === w.id && <Detail id={w.id} people={people} />}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
