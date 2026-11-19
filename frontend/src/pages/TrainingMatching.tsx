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
const LEVEL = ['المسار', 'تفرع أول', 'تفرع ثاني', 'تفرع ثالث']

// ── ١) شجرة المنهج ──
export function CurriculumTree() {
  const [rows, setRows] = useState<TcCurriculum[]>([])
  const [services, setServices] = useState<Service[]>([])
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
  // (ك) 11-18 دازاي: بلا أسماء — كل سطر بس لمنو: فني، ليدر، مهندس، أو الكل.
  // (عمود assignees يبقى بالقاعدة فاضي — شيله يحتاج موافقة دازاي.)
  const saveAud = (n: TcCurriculum, ids: string[]) =>
    run(api.tcSetCurriculumAudiences(n.id, { audiences: AUDIENCES.map((a) => a.id).filter((x) => ids.includes(x)), assignees: [] }))
  const toggleAud = (n: TcCurriculum, id: string) => {
    const cur = eff.get(n.id) ?? []
    saveAud(n, cur.includes(id) ? cur.filter((x) => x !== id) : [...cur, id])
  }
  const run = (p: Promise<TcCurriculum[]>, after?: () => void) =>
    p.then((r) => { setRows(r); setError(''); after?.() }).catch((e) => setError(msg(e)))

  const node = (n: TcCurriculum, depth: number): React.ReactNode => (
    <li key={n.id} className="mt-1.5">
      <div className={`flex items-center gap-2 rounded-lg px-3 py-2 text-sm ${depth === 0 ? 'bg-[#0f2040]/5 font-extrabold' : 'bg-slate-50'}`}>
        <span className="text-[10px] font-bold text-slate-400">{LEVEL[depth]}</span>
        <b className={`min-w-0 truncate ${depth === 0 ? 'text-[#0f2040]' : ''}`} title={n.title}>{n.title}</b>
        {!n.parentId && (eff.get(n.id) ?? []).length === 0 && <span className="shrink-0 text-[10px] font-bold text-amber-600">حدد الفئة</span>}
        {depth === 0 && n.serviceId && <span className="rounded bg-cyan-100 px-1.5 text-[10px] text-cyan-800">{services.find((s) => s.id === n.serviceId)?.name ?? 'خدمة'}</span>}
        <span className="ms-auto flex shrink-0 items-center gap-1" data-testid="aud-chips">
          {AUDIENCES.map((a) => {
            const on = (eff.get(n.id) ?? []).includes(a.id)
            return (
              <button key={a.id} type="button" onClick={() => toggleAud(n, a.id)} aria-pressed={on}
                title={n.audiences === null && on ? 'ماخذ من الي فوكه — اضغط حتى تحدد لهذا السطر' : ''}
                className={`w-14 rounded-full px-1.5 py-0.5 text-[10px] font-bold ${on ? (n.audiences === null ? 'bg-[#2c5aad]/50 text-white' : 'bg-[#2c5aad] text-white shadow-sm') : 'bg-slate-100 text-slate-400'}`}>
                {a.label}
              </button>
            )
          })}
          {(() => {
            const all = AUDIENCES.every((a) => (eff.get(n.id) ?? []).includes(a.id))
            return (
              <button type="button" onClick={() => saveAud(n, all ? [] : AUDIENCES.map((a) => a.id))} aria-pressed={all}
                className={`w-14 rounded-full px-1.5 py-0.5 text-[10px] font-bold ${all ? (n.audiences === null ? 'bg-[#2c5aad]/50 text-white' : 'bg-[#2c5aad] text-white shadow-sm') : 'bg-slate-100 text-slate-400'}`}>
                الكل
              </button>
            )
          })()}
        </span>
        <span className="flex w-32 shrink-0 items-center justify-start gap-2 whitespace-nowrap">
          <span className="w-[4.5rem]">{depth < 3 && <button onClick={() => { setAddUnder(n.id); setChild('') }} className="text-xs font-bold text-[#2c5aad]">+ {LEVEL[depth + 1]}</button>}</span>
          <button onClick={async () => {
            if (kids.has(n.id)) { setError(`«${n.title}» بيه تفرعات — احذفهن أول`); return }
            if (await askConfirm(`حذف «${n.title}»؟`, { danger: true, okLabel: 'حذف' })) run(api.tcDeleteCurriculum(n.id))
          }} className="rounded-full bg-red-100 px-2 py-0.5 text-[11px] font-bold text-red-700">حذف</button>
        </span>
      </div>
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
      <p className="mb-2 text-xs text-slate-500">المسار (مثلاً داش كام) ← تفرع أول ← تفرع ثاني ← تفرع ثالث. اربط المسار بخدمته حتى الماتشينغ يعرف منين يدرّب. وبعد ما تبني، أشّر على كل سطر لمنو: فني، ليدر، مهندس (تگدر أكثر من وحدة) أو «الكل». السطر الي ما تأشّر عليه ياخذ من الي فوكه.</p>
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

// (ك) 11-18 زينا: «بدي التدريب» يفتح مفردات المنهج — تختار عليها شنو راح يتدرب،
// وتطلع بعدين بملفه بإدارة الموظفين («دُرّب على»). المفردات الي تطلع حسب فئته (فني/مهندس…).
function ItemsPicker({ n, tree, onStart, onCancel }: { n: TcNeed; tree: TcCurriculum[]; onStart: (ids: string[]) => void; onCancel: () => void }) {
  const byId = useMemo(() => new Map(tree.map((r) => [r.id, r])), [tree])
  const rootOf = (id: string | null) => { let r = id ? byId.get(id) : undefined; while (r?.parentId) r = byId.get(r.parentId); return r }
  // (11-19) الفني ما يشوف مسارات الليدر (وبالعكس)، والمختار من البداية أول مسار يناسبه
  const mine = n.employeeIsLeader ? 'LEADER' : n.employeeRole
  const aud = (r: TcCurriculum | undefined): string[] => {
    if (!r) return []
    return r.audiences ?? (r.parentId ? aud(byId.get(r.parentId)) : [r.role])
  }
  const fits = (r: TcCurriculum) => !['TECHNICIAN', 'LEADER', 'ENGINEER'].includes(mine) || aud(r).includes(mine)
  const tracks = tree.filter((r) => !r.parentId && fits(r))
  const [track, setTrack] = useState(rootOf(n.curriculumId)?.id ?? tracks[0]?.id ?? '')
  const [picked, setPicked] = useState<string[]>([])
  const list: { r: TcCurriculum; depth: number }[] = []
  const walk = (id: string, depth: number) => {
    for (const c of tree.filter((x) => x.parentId === id).sort((a, b) => a.ord - b.ord)) {
      if (fits(c)) list.push({ r: c, depth })
      walk(c.id, depth + 1)
    }
  }
  if (track) walk(track, 0)
  const toggle = (id: string) => setPicked((p) => (p.includes(id) ? p.filter((x) => x !== id) : [...p, id]))
  return (
    <div className="mt-2 space-y-2 rounded-lg bg-blue-50 p-2" data-testid="items-picker">
      <p className="text-xs font-bold text-[#0f2040]">📚 اختار شنو راح يتدرب عليه من المنهج:</p>
      {tracks.length > 1 && (
        <div className="flex flex-wrap gap-1">
          {tracks.map((t) => (
            <button key={t.id} type="button" onClick={() => { setTrack(t.id); setPicked([]) }}
              className={`rounded-full px-2.5 py-0.5 text-[11px] font-bold ${track === t.id ? 'bg-[#2c5aad] text-white' : 'bg-white text-slate-500'}`}>{t.title}</button>
          ))}
        </div>
      )}
      <div className="max-h-56 space-y-0.5 overflow-y-auto">
        {list.map(({ r, depth }) => (
          <label key={r.id} className="flex cursor-pointer items-center gap-2 rounded px-1 py-0.5 text-xs hover:bg-white" style={{ paddingInlineStart: `${depth * 16 + 4}px` }}>
            <input type="checkbox" checked={picked.includes(r.id)} onChange={() => toggle(r.id)} />
            <span className={depth === 0 ? 'font-bold' : ''}>{r.title}</span>
          </label>
        ))}
        {list.length === 0 && <p className="text-xs text-slate-500">ماكو مفردات بهالمسار لفئته — ضيفها من «مفردات المنهج».</p>}
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <button disabled={list.length > 0 && picked.length === 0} onClick={() => onStart(list.map((x) => x.r.id).filter((id) => picked.includes(id)))} className={btn}>
          📖 بدي التدريب{picked.length ? ` (${picked.length})` : ''}
        </button>
        <button onClick={onCancel} className="text-xs text-slate-500">إلغاء</button>
      </div>
    </div>
  )
}

function NeedRow({ n, tree, onChange }: { n: TcNeed; tree: TcCurriculum[]; onChange: () => void }) {
  const [mode, setMode] = useState<'' | 'start' | 'done' | 'dismiss'>('')
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
      {n.itemTitles.length > 0 && <p className="mt-1 text-xs text-[#2c5aad]">📚 يتدرب على: {n.itemTitles.join('، ')}</p>}
      {error && <p className="mt-1 text-xs text-red-600">{error}</p>}
      {mode === '' && (
        <div className="mt-2 flex flex-wrap gap-2">
          {n.status === 'OPEN' && <button onClick={() => setMode('start')} className={btn}>📖 بدي التدريب</button>}
          <button onClick={() => setMode('done')} className="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-bold text-white">✅ خلص — اكتب التقرير</button>
          {n.kind === 'SKILL_GAP' && n.status === 'OPEN' && <button onClick={() => setMode('dismiss')} className="text-xs text-slate-500">ما يحتاج</button>}
        </div>
      )}
      {mode === 'start' && <ItemsPicker n={n} tree={tree} onStart={(items) => void act('start', { items })} onCancel={() => setMode('')} />}
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
  const [tree, setTree] = useState<TcCurriculum[]>([])
  useEffect(() => { api.tcCurriculum().then(setTree).catch(swallow) }, [])
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
          <ul className="space-y-2">{g.map((n) => <NeedRow key={n.id} n={n} tree={tree} onChange={() => load()} />)}</ul>
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
          {n.itemTitles.length > 0 && <p className="mt-1 text-xs text-[#2c5aad]">📚 اتدرب على: {n.itemTitles.join('، ')}</p>}
          <p className="mt-2 whitespace-pre-wrap rounded-lg bg-slate-50 p-2 text-sm">{n.reportText}</p>
          <p className="mt-1 text-xs text-slate-500">مسؤول التدريب: {n.reportedBy ?? '—'} · بدأ {dt(n.startedAt)} · خلص {dt(n.doneAt)}</p>
        </div>
      ))}
    </div>
  )
}
