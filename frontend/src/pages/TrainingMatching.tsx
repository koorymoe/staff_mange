import { useEffect, useMemo, useState } from 'react'
import { api, type Employee, type Service, type TcCurriculum, type TcNeed } from '../api'
import { roleLabel } from '../roleLabels'
import { card } from '../ui/classes'
import { swallow } from '../utils/swallow'

// ═══ وحدة التدريب (دازاي 11-12) ═══
// ١) مفردات المنهج شجرة: الجمهور (فني/ليدر/مهندس) ← المسار (مثلاً داش كام) ← تفرع أول ← ثاني ← ثالث.
// ٢) الماتشينغ بدل بنك الأسئلة: اسم الموظف والأشياء الي يحتاج يتدرب عليها.
//    مشكلة بحجز = إلزامي ومشدّد (موعد، ما ينلغي). نقص مهارة = اقتراح (ينلغي بسبب).
//    مسؤول وحدة التدريب يبدي التدريب، ولمن يخلص يكتب التقرير ويقيّم — التقرير يوصل للمالك.

const btn = 'rounded-lg bg-[#0f2040] px-3 py-1.5 text-xs font-bold text-white hover:bg-[#1e3f7a] disabled:opacity-40'
const input = 'rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-[#2c5aad]'
const msg = (e: unknown) => (e instanceof Error ? e.message : 'تعذّر')

const AUDIENCES = [
  { id: 'TECHNICIAN', label: 'فني' },
  { id: 'LEADER', label: 'ليدر' },
  { id: 'ENGINEER', label: 'مهندس' },
]
const LEVEL = ['المسار', 'تفرع أول', 'تفرع ثاني', 'تفرع ثالث']

// ── ١) شجرة المنهج ──
export function CurriculumTree() {
  const [rows, setRows] = useState<TcCurriculum[]>([])
  const [services, setServices] = useState<Service[]>([])
  const [aud, setAud] = useState('TECHNICIAN')
  const [title, setTitle] = useState('')
  const [svc, setSvc] = useState('')
  const [addUnder, setAddUnder] = useState<string | null>(null)
  const [child, setChild] = useState('')
  const [error, setError] = useState('')
  useEffect(() => {
    api.tcCurriculum().then(setRows).catch(swallow)
    api.getServices().then(setServices).catch(swallow)
  }, [])
  const kids = useMemo(() => {
    const m = new Map<string, TcCurriculum[]>()
    for (const r of rows) if (r.parentId) m.set(r.parentId, [...(m.get(r.parentId) ?? []), r])
    return m
  }, [rows])
  const tracks = rows.filter((r) => !r.parentId && r.role === aud)
  const run = (p: Promise<TcCurriculum[]>, after?: () => void) =>
    p.then((r) => { setRows(r); setError(''); after?.() }).catch((e) => setError(msg(e)))

  const Node = ({ n, depth }: { n: TcCurriculum; depth: number }) => (
    <li className="mt-1.5">
      <div className={`flex flex-wrap items-center gap-2 rounded-lg px-3 py-2 text-sm ${depth === 0 ? 'bg-[#0f2040]/5 font-extrabold' : 'bg-slate-50'}`}>
        <span className="text-[10px] font-bold text-slate-400">{LEVEL[depth]}</span>
        <b className={depth === 0 ? 'text-[#0f2040]' : ''}>{n.title}</b>
        {depth === 0 && n.serviceId && <span className="rounded bg-cyan-100 px-1.5 text-[10px] text-cyan-800">{services.find((s) => s.id === n.serviceId)?.name ?? 'خدمة'}</span>}
        <span className="ms-auto flex gap-2">
          {depth < 3 && <button onClick={() => { setAddUnder(n.id); setChild('') }} className="text-xs font-bold text-[#2c5aad]">+ {LEVEL[depth + 1]}</button>}
          {!kids.has(n.id) && <button onClick={() => run(api.tcDeleteCurriculum(n.id))} className="text-xs text-slate-400">حذف</button>}
        </span>
      </div>
      {addUnder === n.id && (
        <div className="ms-6 mt-1 flex gap-2">
          <input autoFocus value={child} onChange={(e) => setChild(e.target.value)} placeholder={`${LEVEL[depth + 1]} تحت «${n.title}»`} className={`${input} flex-1`} />
          <button disabled={!child.trim()} onClick={() => run(api.tcAddCurriculum({ parentId: n.id, title: child }), () => setAddUnder(null))} className={btn}>إضافة</button>
          <button onClick={() => setAddUnder(null)} className="text-xs text-slate-500">إلغاء</button>
        </div>
      )}
      {kids.has(n.id) && <ul className="ms-6 border-s-2 border-slate-200 ps-2">{kids.get(n.id)!.map((k) => <Node key={k.id} n={k} depth={depth + 1} />)}</ul>}
    </li>
  )

  return (
    <div className={card}>
      <div className="mb-3 flex flex-wrap items-center gap-2">
        <h3 className="font-extrabold text-[#0f2040]">📚 مفردات المنهج</h3>
        <div className="flex gap-1 rounded-xl bg-slate-100 p-1">
          {AUDIENCES.map((a) => (
            <button key={a.id} onClick={() => setAud(a.id)}
              className={`rounded-lg px-3 py-1 text-xs font-bold ${aud === a.id ? 'bg-[#0f2040] text-white' : 'text-slate-600'}`}>
              {a.label} ({rows.filter((r) => !r.parentId && r.role === a.id).length})
            </button>
          ))}
        </div>
      </div>
      <p className="mb-2 text-xs text-slate-500">المسار (مثلاً داش كام) ← تفرع أول ← تفرع ثاني ← تفرع ثالث. اربط المسار بخدمته حتى الماتشينغ يعرف منين يدرّب.</p>
      {error && <p className="mb-2 text-sm text-red-600">{error}</p>}
      <ul>{tracks.map((t) => <Node key={t.id} n={t} depth={0} />)}</ul>
      {tracks.length === 0 && <p className="text-sm text-slate-500">ماكو مسارات لـ{AUDIENCES.find((a) => a.id === aud)?.label} بعد.</p>}
      <div className="mt-3 flex flex-wrap gap-2 border-t border-slate-100 pt-3">
        <input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="عنوان المسار (مثلاً: داش كام)" className={`${input} flex-1`} />
        <select value={svc} onChange={(e) => setSvc(e.target.value)} className={input}>
          <option value="">— الخدمة (اختياري) —</option>
          {services.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
        </select>
        <button disabled={!title.trim()} onClick={() => run(api.tcAddCurriculum({ role: aud, title, serviceId: svc || undefined }), () => { setTitle(''); setSvc('') })} className={btn}>+ مسار جديد</button>
      </div>
    </div>
  )
}

// ── ٢) الماتشينغ ──
const KIND: Record<TcNeed['kind'], { label: string; cls: string }> = {
  BOOKING_ISSUE: { label: '🔴 إلزامي — مشكلة بحجز', cls: 'bg-red-100 text-red-700' },
  MANUAL: { label: '🔴 إلزامي — من وحدة التدريب', cls: 'bg-red-100 text-red-700' },
  SKILL_GAP: { label: '💡 اقتراح — نقص مهارة', cls: 'bg-amber-100 text-amber-800' },
}
const dt = (s: string | null) => (s ? new Date(s).toLocaleDateString('en-GB') : '—')

function NeedRow({ n, onChange }: { n: TcNeed; onChange: () => void }) {
  const [mode, setMode] = useState<'' | 'done' | 'dismiss'>('')
  const [report, setReport] = useState('')
  const [score, setScore] = useState(0)
  const [note, setNote] = useState('')
  const [error, setError] = useState('')
  const act = (a: 'start' | 'complete' | 'dismiss', b = {}) =>
    api.tcNeedAction(n.id, a, b).then(() => { setMode(''); onChange() }).catch((e) => setError(msg(e)))
  const late = n.kind !== 'SKILL_GAP' && n.dueAt && new Date(n.dueAt) < new Date()
  return (
    <li className="rounded-lg border border-slate-200 p-2.5 text-sm" data-testid="need-row">
      <div className="flex flex-wrap items-center gap-2">
        <span className={`rounded-full px-2 py-0.5 text-[11px] font-bold ${KIND[n.kind].cls}`}>{KIND[n.kind].label}</span>
        <b>{n.topic}</b>
        {n.path && <span className="text-xs text-slate-500">📚 {n.path}</span>}
        {n.status === 'IN_TRAINING' && <span className="rounded-full bg-blue-100 px-2 py-0.5 text-[11px] font-bold text-blue-700">📖 بالتدريب</span>}
        {n.kind !== 'SKILL_GAP' && n.dueAt && <span className={`text-xs ${late ? 'font-bold text-red-600' : 'text-slate-500'}`}>{late ? '⏰ متأخر — ' : ''}الموعد {dt(n.dueAt)}</span>}
      </div>
      {n.reason && <p className="mt-1 text-xs text-slate-600">{n.reason}</p>}
      {error && <p className="mt-1 text-xs text-red-600">{error}</p>}
      {mode === '' && (
        <div className="mt-2 flex flex-wrap gap-2">
          {n.status === 'OPEN' && <button onClick={() => void act('start')} className={btn}>📖 بدي التدريب</button>}
          <button onClick={() => setMode('done')} className="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-bold text-white">✅ خلص — اكتب التقرير</button>
          {n.kind === 'SKILL_GAP' && n.status === 'OPEN' && <button onClick={() => setMode('dismiss')} className="text-xs text-slate-500">ما يحتاج</button>}
        </div>
      )}
      {mode === 'done' && (
        <div className="mt-2 space-y-2 rounded-lg bg-emerald-50 p-2">
          <textarea value={report} onChange={(e) => setReport(e.target.value)} rows={3} placeholder="تقرير التدريب: شنو تدرب، شكد، وشلون صار أداؤه" className={`${input} w-full`} />
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-xs font-bold">تقييمك للموظف:</span>
            {[1, 2, 3, 4, 5].map((s) => (
              <button key={s} type="button" aria-label={`${s} من ٥`} onClick={() => setScore(s)} className={`text-xl ${s <= score ? 'text-amber-500' : 'text-slate-300'}`}>★</button>
            ))}
            <button disabled={!report.trim() || score < 1} onClick={() => void act('complete', { report, score })} className={btn}>أرسل التقرير للمالك</button>
            <button onClick={() => setMode('')} className="text-xs text-slate-500">إلغاء</button>
          </div>
        </div>
      )}
      {mode === 'dismiss' && (
        <div className="mt-2 flex gap-2">
          <input value={note} onChange={(e) => setNote(e.target.value)} placeholder="ليش ما يحتاج؟" className={`${input} flex-1`} />
          <button disabled={!note.trim()} onClick={() => void act('dismiss', { note })} className={btn}>تأكيد</button>
          <button onClick={() => setMode('')} className="text-xs text-slate-500">إلغاء</button>
        </div>
      )}
    </li>
  )
}

export function MatchingTab({ emps }: { emps: Employee[] }) {
  const [rows, setRows] = useState<TcNeed[] | null>(null)
  const [onlyMandatory, setOnlyMandatory] = useState(false)
  const [busy, setBusy] = useState(false)
  const [add, setAdd] = useState({ employeeId: '', topic: '', reason: '' })
  const [error, setError] = useState('')
  const [tick, setTick] = useState(1) // فردي = حدّث من التقييمات والمهارات
  const load = (refresh = false) => { setBusy(true); setTick((t) => (refresh ? (t | 1) + 2 : (t & ~1) + 2)) }
  useEffect(() => {
    api.tcMatching('', (tick & 1) === 1).then(setRows).catch(() => setRows([])).finally(() => setBusy(false))
  }, [tick])
  const groups = useMemo(() => {
    const m = new Map<string, TcNeed[]>()
    for (const r of rows ?? []) if (!onlyMandatory || r.kind !== 'SKILL_GAP') m.set(r.employeeId, [...(m.get(r.employeeId) ?? []), r])
    return [...m.values()].sort((a, b) => b.filter((x) => x.kind !== 'SKILL_GAP').length - a.filter((x) => x.kind !== 'SKILL_GAP').length)
  }, [rows, onlyMandatory])
  const mandatory = (rows ?? []).filter((r) => r.kind !== 'SKILL_GAP').length
  return (
    <div className="space-y-3">
      <div className={`${card} flex flex-wrap items-center gap-3`}>
        <h3 className="font-extrabold text-[#0f2040]">🎯 الماتشينغ — منو يحتاج تدريب وعلى شنو</h3>
        <span className="text-sm">إلزامي: <b className="text-red-600">{mandatory}</b> · اقتراح: <b className="text-amber-700">{(rows?.length ?? 0) - mandatory}</b></span>
        <label className="flex items-center gap-1 text-xs"><input type="checkbox" checked={onlyMandatory} onChange={(e) => setOnlyMandatory(e.target.checked)} /> الإلزامي بس</label>
        <button disabled={busy} onClick={() => load(true)} className={`${btn} ms-auto`}>{busy ? 'يحدّث…' : '🔄 حدّث'}</button>
      </div>
      <p className="text-xs text-slate-500">ماتركس يقرا تقييمات الليدرية بعد الحجوزات (تقييم ١–٢، أو «ما يعرف»، أو «يحتاج تدريب على…») = <b>إلزامي</b> بموعد ٧ أيام وما ينلغي. ويقرا مهارات الموظف بالنظام مقابل الخدمات الي يشتغل عليها = <b>اقتراح</b>.</p>
      {rows && groups.length === 0 && <p className={`${card} text-center text-slate-500`}>ماكو أحد يحتاج تدريب هسه ✅</p>}
      {groups.map((g) => (
        <div key={g[0].employeeId} className={card} data-testid="need-employee">
          <h4 className="mb-2 font-extrabold">{g[0].employeeName} <span className="text-xs font-normal text-slate-500">{roleLabel(g[0].employeeRole)} · {g.length} شي</span></h4>
          <ul className="space-y-2">{g.map((n) => <NeedRow key={n.id} n={n} onChange={() => load()} />)}</ul>
        </div>
      ))}
      <div className={card}>
        <h4 className="mb-2 font-extrabold text-[#0f2040]">✍️ أضف تدريب إلزامي بإيدك</h4>
        {error && <p className="mb-2 text-sm text-red-600">{error}</p>}
        <div className="flex flex-wrap gap-2">
          <select value={add.employeeId} onChange={(e) => setAdd({ ...add, employeeId: e.target.value })} className={input}>
            <option value="">— الموظف —</option>
            {emps.filter((e) => e.role !== 'ADMIN' && e.role !== 'OWNER').map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
          </select>
          <input value={add.topic} onChange={(e) => setAdd({ ...add, topic: e.target.value })} placeholder="يتدرب على شنو" className={`${input} flex-1`} />
          <input value={add.reason} onChange={(e) => setAdd({ ...add, reason: e.target.value })} placeholder="السبب" className={`${input} flex-1`} />
          <button disabled={!add.employeeId || !add.topic.trim()} className={btn}
            onClick={() => api.tcAddNeed(add).then(() => { setAdd({ employeeId: '', topic: '', reason: '' }); setError(''); load() }).catch((e) => setError(msg(e)))}>إضافة</button>
        </div>
      </div>
    </div>
  )
}

// ── ٣) تقارير التدريب (للمالك) ──
export function TrainingReportsTab() {
  const [rows, setRows] = useState<TcNeed[] | null>(null)
  useEffect(() => { api.tcMatching('done').then(setRows).catch(() => setRows([])) }, [])
  return (
    <div className="space-y-3">
      {rows && rows.length === 0 && <p className={`${card} text-center text-slate-500`}>بعد ماكو تدريب مكتمل.</p>}
      {(rows ?? []).map((n) => (
        <div key={n.id} className={card} data-testid="training-report">
          <div className="flex flex-wrap items-center gap-2">
            <b>{n.employeeName}</b>
            <span className={`rounded-full px-2 py-0.5 text-[11px] font-bold ${KIND[n.kind].cls}`}>{KIND[n.kind].label}</span>
            <span className="text-sm">«{n.topic}»</span>
            <span className="ms-auto text-amber-500" aria-label={`التقييم ${n.reportScore} من ٥`}>{'★'.repeat(n.reportScore ?? 0)}<span className="text-slate-300">{'★'.repeat(5 - (n.reportScore ?? 0))}</span></span>
          </div>
          {n.reason && <p className="mt-1 text-xs text-slate-500">السبب: {n.reason}</p>}
          <p className="mt-2 whitespace-pre-wrap rounded-lg bg-slate-50 p-2 text-sm">{n.reportText}</p>
          <p className="mt-1 text-xs text-slate-500">مسؤول التدريب: {n.reportedBy ?? '—'} · بدأ {dt(n.startedAt)} · خلص {dt(n.doneAt)}</p>
        </div>
      ))}
    </div>
  )
}
