import { useEffect, useMemo, useState } from 'react'
import { api, type Employee, type Service, type TcCurriculum, type TcNeed } from '../api'
import { roleLabel } from '../roleLabels'
import { card } from '../ui/classes'
import { askConfirm } from '../ui/dialogs'
import { swallow } from '../utils/swallow'

// ═══ وحدة التدريب (دازاي 11-12) ═══
// ١) مفردات المنهج شجرة: المسار (مثلاً داش كام) ← تفرع أول ← ثاني ← ثالث.
//    (ك) 11-14 دازاي: الفئة (فني/ليدر/مهندس) مو تبويب فوك — تأشّرها هي على كل سطر بعد ما تبني،
//    وتگدر تختار أكثر من وحدة. السطر الي ما تأشّر عليه ياخذ من الي فوكه.
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
// (ك) 11-14 زينا: كل كبسولة تفتح قائمتها — فني يطلعون الفنيين، ليدر الليدرية، مهندس المهندسين،
// و«أخرى» بحث بالاسم على كل الموظفين.
const inGroup = (e: Employee, g: string) =>
  g === 'LEADER' ? e.isLeader : g === 'TECHNICIAN' ? e.role === 'TECHNICIAN' && !e.isLeader : e.role === g
const groupOf = (e: Employee) => AUDIENCES.find((a) => inGroup(e, a.id))?.id
const LEVEL = ['المسار', 'تفرع أول', 'تفرع ثاني', 'تفرع ثالث']

// ── ١) شجرة المنهج ──
export function CurriculumTree({ emps }: { emps: Employee[] }) {
  const [rows, setRows] = useState<TcCurriculum[]>([])
  const [services, setServices] = useState<Service[]>([])
  const [title, setTitle] = useState('')
  const [svc, setSvc] = useState('')
  const [addUnder, setAddUnder] = useState<string | null>(null)
  const [child, setChild] = useState('')
  const [error, setError] = useState('')
  const [pick, setPick] = useState<{ id: string; g: string } | null>(null)
  const [q, setQ] = useState('')
  useEffect(() => {
    api.tcCurriculum().then(setRows).catch(swallow)
    api.getServices().then(setServices).catch(swallow)
  }, [])
  const kids = useMemo(() => {
    const m = new Map<string, TcCurriculum[]>()
    for (const r of rows) if (r.parentId) m.set(r.parentId, [...(m.get(r.parentId) ?? []), r])
    return m
  }, [rows])
  const tracks = rows.filter((r) => !r.parentId)
  // الفئات الفعلية لكل سطر: اختياره هو، وإلا الي فوكه، والمسار القديم بلا اختيار ياخذ دوره.
  const eff = useMemo(() => {
    const byId = new Map(rows.map((r) => [r.id, r]))
    const memo = new Map<string, string[]>()
    const get = (r: TcCurriculum): string[] => {
      const hit = memo.get(r.id)
      if (hit) return hit
      const parent = r.parentId ? byId.get(r.parentId) : undefined
      const v = r.audiences ?? (parent ? get(parent) : [r.role])
      memo.set(r.id, v)
      return v
    }
    for (const r of rows) get(r)
    return memo
  }, [rows])
  // الأشخاص الفعليين لكل سطر: اختياره هو، وإلا الي فوكه.
  const effPeople = useMemo(() => {
    const byId = new Map(rows.map((r) => [r.id, r]))
    const memo = new Map<string, string[]>()
    const get = (r: TcCurriculum): string[] => {
      const hit = memo.get(r.id)
      if (hit) return hit
      const parent = r.parentId ? byId.get(r.parentId) : undefined
      const v = r.assignees ?? (parent ? get(parent) : [])
      memo.set(r.id, v)
      return v
    }
    for (const r of rows) get(r)
    return memo
  }, [rows])
  const empById = useMemo(() => new Map(emps.map((e) => [e.id, e])), [emps])
  // يحفظ الأشخاص، والفئات تنحسب منهم (فني/ليدر/مهندس الي بيها أحد مختار).
  const savePeople = (n: TcCurriculum, ids: string[]) => {
    const groups = new Set(ids.map((id) => { const e = empById.get(id); return e ? groupOf(e) : undefined }))
    run(api.tcSetCurriculumAudiences(n.id, { assignees: ids, audiences: AUDIENCES.map((a) => a.id).filter((g) => groups.has(g)) }))
  }
  const togglePerson = (n: TcCurriculum, id: string) => {
    const cur = effPeople.get(n.id) ?? []
    savePeople(n, cur.includes(id) ? cur.filter((x) => x !== id) : [...cur, id])
  }
  const run = (p: Promise<TcCurriculum[]>, after?: () => void) =>
    p.then((r) => { setRows(r); setError(''); after?.() }).catch((e) => setError(msg(e)))

  const picker = (n: TcCurriculum) => {
    if (pick?.id !== n.id) return null
    const chosen = new Set(effPeople.get(n.id) ?? [])
    const pool = pick.g === 'OTHER'
      ? emps.filter((e) => q.trim() && e.name.includes(q.trim())).slice(0, 30)
      : emps.filter((e) => inGroup(e, pick.g))
    const allOn = pool.length > 0 && pool.every((e) => chosen.has(e.id))
    return (
      <div className="ms-6 mt-1 space-y-2 rounded-xl border border-slate-200 bg-white p-3" data-testid="people-picker">
        <div className="flex flex-wrap items-center gap-2">
          <b className="text-xs text-[#0f2040]">{pick.g === 'OTHER' ? 'أخرى — ابحث بالاسم' : AUDIENCES.find((a) => a.id === pick.g)?.label}</b>
          {pick.g === 'OTHER'
            ? <input autoFocus value={q} onChange={(e) => setQ(e.target.value)} placeholder="اسم الموظف…" className={`${input} flex-1 py-1`} />
            : pool.length > 0 && (
              <button type="button" onClick={() => savePeople(n, allOn ? [...chosen].filter((id) => !pool.some((e) => e.id === id)) : [...new Set([...chosen, ...pool.map((e) => e.id)])])}
                className="rounded-full bg-slate-100 px-2 py-0.5 text-[11px] font-bold text-slate-600">{allOn ? 'شيل الكل' : 'اختيار الكل'}</button>
            )}
          <button type="button" onClick={() => setPick(null)} className="ms-auto text-xs text-slate-500">إغلاق</button>
        </div>
        <div className="flex max-h-48 flex-wrap gap-1.5 overflow-y-auto">
          {pool.map((e) => {
            const on = chosen.has(e.id)
            return (
              <button key={e.id} type="button" onClick={() => togglePerson(n, e.id)} aria-pressed={on}
                className={`rounded-full px-2.5 py-1 text-xs font-bold ${on ? 'bg-[#2c5aad] text-white' : 'bg-slate-100 text-slate-600'}`}>
                {on ? '✓ ' : ''}{e.name}
              </button>
            )
          })}
          {pool.length === 0 && <span className="text-xs text-slate-400">{pick.g === 'OTHER' ? (q.trim() ? 'ماكو أحد بهالاسم' : 'اكتب الاسم') : 'ماكو أحد بهالقائمة'}</span>}
        </div>
      </div>
    )
  }

  const node = (n: TcCurriculum, depth: number): React.ReactNode => (
    <li key={n.id} className="mt-1.5">
      <div className={`flex flex-wrap items-center gap-2 rounded-lg px-3 py-2 text-sm ${depth === 0 ? 'bg-[#0f2040]/5 font-extrabold' : 'bg-slate-50'}`}>
        <span className="text-[10px] font-bold text-slate-400">{LEVEL[depth]}</span>
        <b className={depth === 0 ? 'text-[#0f2040]' : ''}>{n.title}</b>
        {depth === 0 && n.serviceId && <span className="rounded bg-cyan-100 px-1.5 text-[10px] text-cyan-800">{services.find((s) => s.id === n.serviceId)?.name ?? 'خدمة'}</span>}
        <span className="ms-auto flex flex-wrap items-center gap-1" data-testid="aud-chips">
          {[...AUDIENCES, { id: 'OTHER', label: 'أخرى' }].map((a) => {
            const people = (effPeople.get(n.id) ?? []).map((id) => empById.get(id)).filter((e): e is Employee => !!e)
            const count = a.id === 'OTHER' ? people.filter((e) => !groupOf(e)).length : people.filter((e) => inGroup(e, a.id)).length
            const on = a.id === 'OTHER' ? count > 0 : count > 0 || (eff.get(n.id) ?? []).includes(a.id)
            const open = pick?.id === n.id && pick.g === a.id
            return (
              <button key={a.id} type="button" onClick={() => { setPick(open ? null : { id: n.id, g: a.id }); setQ('') }} aria-pressed={on} aria-expanded={open}
                title={n.assignees === null && on ? 'ماخذ من الي فوكه — اضغط حتى تختار لهذا السطر' : ''}
                className={`inline-flex w-[4.5rem] items-center justify-center gap-0.5 rounded-full px-2 py-0.5 text-[11px] font-bold ${on ? (n.assignees === null ? 'bg-[#2c5aad]/50 text-white' : 'bg-[#2c5aad] text-white shadow-sm') : 'bg-slate-100 text-slate-400'} ${open ? 'ring-2 ring-[#2c5aad]/40' : ''}`}>
                {a.label}{count > 0 ? ` ${count}` : ''} <span aria-hidden className={`text-[9px] transition ${open ? 'rotate-180' : ''}`}>▾</span>
              </button>
            )
          })}
          {/* خانة ثابتة العرض حتى الكبسولات تصف وحدة جوّه الثانية */}
          <span className="w-16 text-center">
            {!n.parentId && (eff.get(n.id) ?? []).length === 0 && <span className="text-[10px] font-bold text-amber-600">حدد الفئة</span>}
          </span>
        </span>
        <span className="flex w-36 shrink-0 justify-end gap-2">
          {depth < 3 && <button onClick={() => { setAddUnder(n.id); setChild('') }} className="text-xs font-bold text-[#2c5aad]">+ {LEVEL[depth + 1]}</button>}
          <button onClick={async () => {
            if (kids.has(n.id)) { setError(`«${n.title}» بيه تفرعات — احذفهن أول`); return }
            if (await askConfirm(`حذف «${n.title}»؟`, { danger: true, okLabel: 'حذف' })) run(api.tcDeleteCurriculum(n.id))
          }} className="rounded-full bg-red-100 px-2 py-0.5 text-[11px] font-bold text-red-700">حذف</button>
        </span>
      </div>
      {picker(n)}
      {addUnder === n.id && (
        <div className="ms-6 mt-1 flex gap-2">
          <input autoFocus value={child} onChange={(e) => setChild(e.target.value)} placeholder={`${LEVEL[depth + 1]} تحت «${n.title}»`} className={`${input} flex-1`} />
          <button disabled={!child.trim()} onClick={() => run(api.tcAddCurriculum({ parentId: n.id, title: child }), () => setAddUnder(null))} className={btn}>إضافة</button>
          <button onClick={() => setAddUnder(null)} className="text-xs text-slate-500">إلغاء</button>
        </div>
      )}
      {kids.has(n.id) && <ul className="ms-6 border-s-2 border-slate-200 ps-2">{kids.get(n.id)!.map((k) => node(k, depth + 1))}</ul>}
    </li>
  )

  return (
    <div className={card}>
      <div className="mb-3 flex flex-wrap items-center gap-2">
        <h3 className="font-extrabold text-[#0f2040]">📚 مفردات المنهج</h3>
      </div>
      <p className="mb-2 text-xs text-slate-500">المسار (مثلاً داش كام) ← تفرع أول ← تفرع ثاني ← تفرع ثالث. اربط المسار بخدمته حتى الماتشينغ يعرف منين يدرّب. وبعد ما تبني، اضغط على كبسولة فني أو ليدر أو مهندس وتطلع قائمتهم، واختار واحد أو أكثر أو الكل؛ و«أخرى» تبحث بالاسم على أي موظف. السطر الي ما تختار إله ياخذ من الي فوكه.</p>
      {error && <p className="mb-2 text-sm text-red-600">{error}</p>}
      <ul>{tracks.map((t) => node(t, 0))}</ul>
      {tracks.length === 0 && <p className="text-sm text-slate-500">ماكو مسارات بعد.</p>}
      <div className="mt-3 flex flex-wrap gap-2 border-t border-slate-100 pt-3">
        <input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="عنوان المسار (مثلاً: داش كام)" className={`${input} flex-1`} />
        <select value={svc} onChange={(e) => setSvc(e.target.value)} className={input}>
          <option value="">— الخدمة (اختياري) —</option>
          {services.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
        </select>
        <button disabled={!title.trim()} onClick={() => run(api.tcAddCurriculum({ title, serviceId: svc || undefined }), () => { setTitle(''); setSvc('') })} className={btn}>+ مسار جديد</button>
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
