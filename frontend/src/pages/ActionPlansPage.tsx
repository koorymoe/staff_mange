// (ك) 11-28 زينا: «معالجة الحالات الحرجة» — موظف عنده مشكلة بالشغل (من ماتركس أو المدير أو المالك):
// نفتحله خطة بهدف ومدة، نتابعها، وبالنهاية: تعالجت؟ وشنو الإجراء التالي. داخلية — الموظف ما يشوفها.
// ماكو غرامات فلوس: الإجراء التالي نص يكتبه المسؤول.
import { useEffect, useState } from 'react'
import Modal from '../ui/Modal'
import { LineIcon } from '../components/LineIcon'
import { actionPlansApi, PROBLEMS, SOURCES, type ActionPlan, type APNote, type APProblem, type APSource } from '../actionPlans'
import './ActionPlans.css'

type Filter = 'OPEN' | 'DUE' | 'RESOLVED' | 'NOT_RESOLVED' | 'ALL'
const err = (e: unknown) => (e instanceof Error ? e.message : 'صار خطأ')
const plusDays = (n: number) => {
  const d = new Date(new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' }) + 'T12:00:00')
  d.setDate(d.getDate() + n)
  return d.toISOString().slice(0, 10)
}

export default function ActionPlansPage() {
  const [plans, setPlans] = useState<ActionPlan[] | null>(null)
  const [loadErr, setLoadErr] = useState('')
  const [reload, setReload] = useState(0)
  const [filter, setFilter] = useState<Filter>('OPEN')
  const [q, setQ] = useState('')
  const [creating, setCreating] = useState(false)
  const [openId, setOpenId] = useState<string | null>(null)

  useEffect(() => {
    let alive = true
    actionPlansApi.list().then((d) => { if (alive) { setPlans(d); setLoadErr('') } }).catch((e) => { if (alive) setLoadErr(err(e)) })
    return () => { alive = false }
  }, [reload])

  const all = plans ?? []
  const open = all.filter((p) => p.status === 'OPEN')
  const due = open.filter((p) => p.daysLeft < 0)
  const counts = { OPEN: open.length, DUE: due.length, RESOLVED: all.filter((p) => p.status === 'RESOLVED').length, NOT_RESOLVED: all.filter((p) => p.status === 'NOT_RESOLVED').length, ALL: all.length }
  const rows = all.filter((p) => (filter === 'ALL' ? true : filter === 'DUE' ? p.status === 'OPEN' && p.daysLeft < 0 : p.status === filter))
    .filter((p) => !q.trim() || [p.employeeName, p.target, PROBLEMS[p.problem]].some((v) => v.includes(q.trim())))
  const current = all.find((p) => p.id === openId) ?? null
  const done = () => setReload((n) => n + 1)

  const cards: { k: Filter; l: string; ic: string; c: string }[] = [
    { k: 'OPEN', l: 'خطط مفتوحة', ic: 'clip', c: 'blue' },
    { k: 'DUE', l: 'خلصت مدتها بلا نتيجة', ic: 'warn', c: 'rose' },
    { k: 'RESOLVED', l: 'تعالجت', ic: 'checkCircle', c: 'green' },
    { k: 'NOT_RESOLVED', l: 'ما تعالجت', ic: 'xCircle', c: 'amber' },
  ]

  return (
    <div className="ap-page space-y-4" dir="rtl">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <span className="ap-logo"><LineIcon n="shield" className="h-6 w-6" /></span>
          <div>
            <h1 className="ap-title">معالجة الحالات الحرجة</h1>
            <p className="ap-sub">خطة لكل موظف عنده مشكلة بالشغل: الهدف، والمدة، والمتابعة، وشنو الإجراء التالي. داخلية — الموظف ما يشوفها.</p>
          </div>
        </div>
        <button type="button" className="ap-btn primary" onClick={() => setCreating(true)}><LineIcon n="plus" />خطة جديدة</button>
      </div>

      <div className="ap-cards">
        {cards.map((k) => (
          <button key={k.k} type="button" onClick={() => setFilter(k.k)} className={`ap-card ${k.c} ${filter === k.k ? 'on' : ''}`}>
            <span className="ap-ic"><LineIcon n={k.ic} className="h-5 w-5" /></span>
            <span className="min-w-0"><span className="ap-cl">{k.l}</span><span className="ap-cv">{plans ? counts[k.k] : '…'}</span></span>
          </button>
        ))}
      </div>

      <div className="ap-panel">
        <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
          <label className="ap-search"><LineIcon n="search" /><input value={q} onChange={(e) => setQ(e.target.value)} placeholder="بحث بالموظف أو الهدف..." /></label>
          <button type="button" className={`ap-btn ${filter === 'ALL' ? 'on' : ''}`} onClick={() => setFilter('ALL')}>الكل ({counts.ALL})</button>
        </div>
        {loadErr && <p className="ap-err">{loadErr}</p>}
        {!plans && !loadErr && <p className="ap-sub py-8 text-center">جاري التحميل…</p>}
        {plans && rows.length === 0 && <p className="ap-sub py-8 text-center">ماكو خطط هنا.</p>}
        {rows.length > 0 && (
          <table className="ap-table w-full text-right">
            <thead><tr><th>الموظف</th><th>المشكلة</th><th>الهدف</th><th>المدة</th><th>الحالة</th><th>متابعات</th><th></th></tr></thead>
            <tbody>
              {rows.map((p) => (
                <tr key={p.id}>
                  <td><b>{p.employeeName}</b>{p.position && <span className="ap-small">{p.position}</span>}</td>
                  <td>{PROBLEMS[p.problem]}<span className="ap-small">{SOURCES[p.source]}</span></td>
                  <td className="ap-target">{p.target}</td>
                  <td className="tabular-nums whitespace-nowrap">{p.dueDate}<span className="ap-small">{p.status !== 'OPEN' ? '—' : p.daysLeft < 0 ? `متأخرة ${-p.daysLeft} يوم` : p.daysLeft === 0 ? 'تخلص اليوم' : `باقي ${p.daysLeft} يوم`}</span></td>
                  <td><StatusPill p={p} /></td>
                  <td className="tabular-nums">{p.notes}</td>
                  <td><button type="button" className="ap-btn sm" onClick={() => setOpenId(p.id)}>فتح</button></td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {creating && <CreateModal onClose={() => setCreating(false)} onDone={() => { setCreating(false); setFilter('OPEN'); done() }} />}
      {current && <PlanModal p={current} onClose={() => setOpenId(null)} onChanged={done} />}
    </div>
  )
}

function StatusPill({ p }: { p: ActionPlan }) {
  if (p.status === 'RESOLVED') return <span className="ap-pill ok">✓ تعالجت</span>
  if (p.status === 'NOT_RESOLVED') return <span className="ap-pill bad">✕ ما تعالجت</span>
  if (p.daysLeft < 0) return <span className="ap-pill late">⚠ خلصت المدة</span>
  return <span className="ap-pill run">مفتوحة</span>
}

function CreateModal({ onClose, onDone }: { onClose: () => void; onDone: () => void }) {
  const [emps, setEmps] = useState<{ id: string; name: string; jobTitle: string | null }[]>([])
  const [f, setF] = useState({ employeeId: '', problem: 'SLOW' as APProblem, source: 'MANAGER' as APSource, details: '', target: '', steps: '', dueDate: plusDays(14) })
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState('')
  useEffect(() => {
    let alive = true
    actionPlansApi.employees().then((d) => { if (alive) setEmps(d) }).catch((e) => { if (alive) setMsg(err(e)) })
    return () => { alive = false }
  }, [])
  const save = async () => {
    setBusy(true); setMsg('')
    try { await actionPlansApi.create(f); onDone() } catch (e) { setMsg(err(e)) } finally { setBusy(false) }
  }
  const set = <K extends keyof typeof f>(k: K, v: (typeof f)[K]) => setF((o) => ({ ...o, [k]: v }))
  return (
    <Modal title="خطة معالجة جديدة" onClose={onClose} busy={busy} closeOnBackdrop={false} size="md" bodyClassName="ap-form space-y-3 p-5">
      <label>الموظف
        <select value={f.employeeId} onChange={(e) => set('employeeId', e.target.value)}>
          <option value="">— اختار —</option>
          {emps.map((e) => <option key={e.id} value={e.id}>{e.name}{e.jobTitle ? ` — ${e.jobTitle}` : ''}</option>)}
        </select>
      </label>
      <div className="grid grid-cols-2 gap-3">
        <label>نوع المشكلة
          <select value={f.problem} onChange={(e) => set('problem', e.target.value as APProblem)}>
            {(Object.keys(PROBLEMS) as APProblem[]).map((k) => <option key={k} value={k}>{PROBLEMS[k]}</option>)}
          </select>
        </label>
        <label>منين جاي التقرير
          <select value={f.source} onChange={(e) => set('source', e.target.value as APSource)}>
            {(Object.keys(SOURCES) as APSource[]).map((k) => <option key={k} value={k}>{SOURCES[k]}</option>)}
          </select>
        </label>
      </div>
      <label>شنو المشكلة بالتفصيل<textarea rows={2} value={f.details} onChange={(e) => set('details', e.target.value)} placeholder="مثلاً: تقارير الحجوزات تتأخر يومين" /></label>
      <label>الهدف (التارگت) *<textarea rows={2} value={f.target} onChange={(e) => set('target', e.target.value)} placeholder="مثلاً: كل تقرير ينرفع بنفس يوم الحجز" /></label>
      <label>الخطوات المطلوبة<textarea rows={2} value={f.steps} onChange={(e) => set('steps', e.target.value)} placeholder="شنو لازم يسوي حتى يوصل للهدف" /></label>
      <label>نهاية المدة *
        <input type="date" value={f.dueDate} onChange={(e) => set('dueDate', e.target.value)} />
        <span className="ap-quick">{[7, 14, 30].map((n) => <button key={n} type="button" onClick={() => set('dueDate', plusDays(n))}>{n} يوم</button>)}</span>
      </label>
      {msg && <p className="ap-err">{msg}</p>}
      <div className="flex justify-end gap-2">
        <button type="button" className="ap-btn" onClick={onClose}>إلغاء</button>
        <button type="button" className="ap-btn primary" disabled={busy || !f.employeeId || !f.target.trim()} onClick={save}>حفظ الخطة</button>
      </div>
    </Modal>
  )
}

function PlanModal({ p, onClose, onChanged }: { p: ActionPlan; onClose: () => void; onChanged: () => void }) {
  const [notes, setNotes] = useState<APNote[] | null>(null)
  const [nReload, setNReload] = useState(0)
  const [note, setNote] = useState('')
  const [mode, setMode] = useState<'' | 'extend' | 'close'>('')
  const [ext, setExt] = useState({ dueDate: plusDays(7), reason: '' })
  const [cl, setCl] = useState({ status: 'RESOLVED' as 'RESOLVED' | 'NOT_RESOLVED', outcome: '', nextAction: '' })
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState('')
  useEffect(() => {
    let alive = true
    actionPlansApi.notes(p.id).then((d) => { if (alive) setNotes(d) }).catch((e) => { if (alive) setMsg(err(e)) })
    return () => { alive = false }
  }, [p.id, nReload])
  const run = async (fn: () => Promise<unknown>, after: () => void) => {
    setBusy(true); setMsg('')
    try { await fn(); after(); onChanged(); setNReload((n) => n + 1) } catch (e) { setMsg(err(e)) } finally { setBusy(false) }
  }
  const isOpen = p.status === 'OPEN'
  return (
    <Modal title={<>خطة {p.employeeName}</>} onClose={onClose} busy={busy} size="lg" bodyClassName="ap-form space-y-4 p-5">
      <div className="ap-info">
        <div><span>المشكلة</span><b>{PROBLEMS[p.problem]} · {SOURCES[p.source]}</b></div>
        <div><span>المدة</span><b className="tabular-nums">{p.startDate} ← {p.dueDate}</b></div>
        <div><span>الحالة</span><StatusPill p={p} /></div>
        <div><span>فتحها</span><b>{p.createdBy}</b></div>
      </div>
      {p.details && <div className="ap-box"><span>التفصيل</span><p>{p.details}</p></div>}
      <div className="ap-box target"><span>الهدف</span><p>{p.target}</p></div>
      {p.steps && <div className="ap-box"><span>الخطوات</span><p>{p.steps}</p></div>}
      {!isOpen && (
        <div className={`ap-box ${p.status === 'RESOLVED' ? 'ok' : 'bad'}`}>
          <span>النتيجة — {p.closedBy}</span><p>{p.outcome}</p>
          {p.nextAction && <><span>الإجراء التالي</span><p><b>{p.nextAction}</b></p></>}
        </div>
      )}

      <div>
        <p className="ap-h">المتابعات</p>
        {!notes ? <p className="ap-sub">…</p> : notes.length === 0 ? <p className="ap-sub">ماكو متابعات بعد.</p> : (
          <ul className="ap-notes">{notes.map((n) => <li key={n.id}><p>{n.note}</p><span>{n.by} · {new Date(n.createdAt).toLocaleString('en-GB', { timeZone: 'Asia/Baghdad', dateStyle: 'short', timeStyle: 'short' })}</span></li>)}</ul>
        )}
        {isOpen && (
          <div className="mt-2 flex gap-2">
            <input value={note} onChange={(e) => setNote(e.target.value)} placeholder="اكتب متابعة (شنو تحسّن، شنو بعده…)" />
            <button type="button" className="ap-btn" disabled={busy || !note.trim()} onClick={() => run(() => actionPlansApi.addNote(p.id, note.trim()), () => setNote(''))}>إضافة</button>
          </div>
        )}
      </div>

      {isOpen && mode === '' && (
        <div className="flex flex-wrap justify-end gap-2">
          <button type="button" className="ap-btn" onClick={() => setMode('extend')}><LineIcon n="clock" />تمديد المدة</button>
          <button type="button" className="ap-btn primary" onClick={() => setMode('close')}><LineIcon n="check" />تسجيل النتيجة</button>
        </div>
      )}
      {isOpen && mode === 'extend' && (
        <div className="ap-box space-y-2">
          <label>التاريخ الجديد<input type="date" value={ext.dueDate} onChange={(e) => setExt({ ...ext, dueDate: e.target.value })} /></label>
          <label>سبب التمديد *<input value={ext.reason} onChange={(e) => setExt({ ...ext, reason: e.target.value })} /></label>
          <div className="flex justify-end gap-2">
            <button type="button" className="ap-btn" onClick={() => setMode('')}>رجوع</button>
            <button type="button" className="ap-btn primary" disabled={busy || !ext.reason.trim()} onClick={() => run(() => actionPlansApi.extend(p.id, ext.dueDate, ext.reason.trim()), () => setMode(''))}>تمديد</button>
          </div>
        </div>
      )}
      {isOpen && mode === 'close' && (
        <div className="ap-box space-y-2">
          <div className="flex gap-2">
            <button type="button" className={`ap-btn ${cl.status === 'RESOLVED' ? 'ok' : ''}`} onClick={() => setCl({ ...cl, status: 'RESOLVED' })}>✓ تعالجت</button>
            <button type="button" className={`ap-btn ${cl.status === 'NOT_RESOLVED' ? 'bad' : ''}`} onClick={() => setCl({ ...cl, status: 'NOT_RESOLVED' })}>✕ ما تعالجت</button>
          </div>
          <label>شنو صار *<textarea rows={2} value={cl.outcome} onChange={(e) => setCl({ ...cl, outcome: e.target.value })} /></label>
          <label>الإجراء التالي {cl.status === 'NOT_RESOLVED' ? '*' : '(اختياري)'}<textarea rows={2} value={cl.nextAction} onChange={(e) => setCl({ ...cl, nextAction: e.target.value })} placeholder="مثلاً: تدريب أسبوع ويا الليدر، نقل لوحدة ثانية، إنذار مكتوب…" /></label>
          <div className="flex justify-end gap-2">
            <button type="button" className="ap-btn" onClick={() => setMode('')}>رجوع</button>
            <button type="button" className="ap-btn primary" disabled={busy || !cl.outcome.trim() || (cl.status === 'NOT_RESOLVED' && !cl.nextAction.trim())}
              onClick={() => run(() => actionPlansApi.close(p.id, cl.status, cl.outcome.trim(), cl.nextAction.trim()), () => setMode(''))}>حفظ النتيجة</button>
          </div>
        </div>
      )}
      {msg && <p className="ap-err">{msg}</p>}
    </Modal>
  )
}
